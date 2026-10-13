package auth

import (
	"context"
	"dev/internal/audit"
	users "dev/internal/user"
	"dev/internal/utils"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var oauthClient = &http.Client{Timeout: 10 * time.Second}

type OAuthProvider struct {
	Name         string
	AuthorizeURL string
	Scope        string

	tokenURL        string
	clientIdEnv     string
	clientSecretEnv string
	fetchIdentity   func(ctx context.Context, accessToken string) (*users.Identity, error)
}

var oauthProviders = []*OAuthProvider{
	{
		Name:            users.IdentityProviderGithub,
		AuthorizeURL:    "https://github.com/login/oauth/authorize",
		Scope:           "read:user user:email",
		tokenURL:        "https://github.com/login/oauth/access_token",
		clientIdEnv:     "GITHUB_CLIENT_ID",
		clientSecretEnv: "GITHUB_CLIENT_SECRET",
		fetchIdentity:   fetchGithubIdentity,
	},
	{
		Name:            users.IdentityProviderGoogle,
		AuthorizeURL:    "https://accounts.google.com/o/oauth2/v2/auth",
		Scope:           "openid email profile",
		tokenURL:        "https://oauth2.googleapis.com/token",
		clientIdEnv:     "GOOGLE_CLIENT_ID",
		clientSecretEnv: "GOOGLE_CLIENT_SECRET",
		fetchIdentity:   fetchGoogleIdentity,
	},
}

func (provider *OAuthProvider) ClientId() string {
	return os.Getenv(provider.clientIdEnv)
}

func (provider *OAuthProvider) enabled() bool {
	return provider.ClientId() != "" && os.Getenv(provider.clientSecretEnv) != ""
}

func GetOAuthProviders() []*OAuthProvider {
	enabled := []*OAuthProvider{}
	for _, provider := range oauthProviders {
		if provider.enabled() {
			enabled = append(enabled, provider)
		}
	}
	return enabled
}

func GetOAuthProvider(name string) *OAuthProvider {
	for _, provider := range GetOAuthProviders() {
		if provider.Name == name {
			return provider
		}
	}
	return nil
}

func (provider *OAuthProvider) Authenticate(ctx context.Context, code string, redirectURI string) (*users.Identity, error) {
	data := url.Values{}
	data.Set("client_id", provider.ClientId())
	data.Set("client_secret", os.Getenv(provider.clientSecretEnv))
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	if redirectURI != "" {
		data.Set("redirect_uri", redirectURI)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, provider.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	response, err := oauthClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var token struct {
		AccessToken      string `json:"access_token"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.NewDecoder(response.Body).Decode(&token); err != nil {
		return nil, err
	}
	if token.AccessToken == "" {
		return nil, fmt.Errorf("%s refused the authorization code: %s %s", provider.Name, token.Error, token.ErrorDescription)
	}

	identity, err := provider.fetchIdentity(ctx, token.AccessToken)
	if err != nil {
		return nil, err
	}
	identity.Provider = provider.Name
	identity.AccessToken = token.AccessToken
	return identity, nil
}

func getOAuthResource(ctx context.Context, url string, accessToken string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	response, err := oauthClient.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: status %d", url, response.StatusCode)
	}
	return json.NewDecoder(response.Body).Decode(v)
}

func fetchGithubIdentity(ctx context.Context, accessToken string) (*users.Identity, error) {
	var profile struct {
		Id        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		AvatarUrl string `json:"avatar_url"`
		HtmlUrl   string `json:"html_url"`
	}
	if err := getOAuthResource(ctx, "https://api.github.com/user", accessToken, &profile); err != nil {
		return nil, err
	}

	// The email of the public profile is optional and unverified
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := getOAuthResource(ctx, "https://api.github.com/user/emails", accessToken, &emails); err != nil {
		return nil, err
	}
	for _, email := range emails {
		if email.Primary && email.Verified {
			return &users.Identity{
				ProviderId: strconv.Itoa(profile.Id),
				Email:      email.Email,
				Username:   profile.Login,
				Name:       profile.Name,
				AvatarUrl:  profile.AvatarUrl,
				ProfileUrl: profile.HtmlUrl,
			}, nil
		}
	}
	return nil, errors.New("this GitHub account has no verified primary email")
}

func fetchGoogleIdentity(ctx context.Context, accessToken string) (*users.Identity, error) {
	var profile struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := getOAuthResource(ctx, "https://openidconnect.googleapis.com/v1/userinfo", accessToken, &profile); err != nil {
		return nil, err
	}
	if profile.Email == "" || !profile.EmailVerified {
		return nil, errors.New("this Google account has no verified email")
	}

	return &users.Identity{
		ProviderId: profile.Sub,
		Email:      profile.Email,
		Name:       profile.Name,
		AvatarUrl:  profile.Picture,
	}, nil
}

// A passkey is bound to the domain it was created on, hence one relying party per origin
func NewWebAuthn(origin string) (*webauthn.WebAuthn, error) {
	if !slices.Contains(utils.AllowedOrigins, origin) {
		return nil, errors.New("origin not allowed")
	}
	parsed, err := url.Parse(origin)
	if err != nil {
		return nil, err
	}

	return webauthn.New(&webauthn.Config{
		RPID:          parsed.Hostname(),
		RPDisplayName: "Mathéo Galuba - Dev",
		RPOrigins:     []string{origin},
	})
}

// SendVerificationEmail godoc
// Sends a verification email to the user with the provided url
func SendVerificationEmail(c context.Context, user *users.User, url string) error {
	_, err := utils.SendEmail(c, utils.Email{
		To:      utils.EmailAddress{Name: user.Name, Email: user.Email},
		Subject: "Verify your email address",
		HTMLContent: `
		<h1>Verify your email address</h1>
		<p>
			👋 Hi,<br>
			Thanks for signing up to my website!<br>
			Please verify your email address by clicking the link below.
		</p>

		<a href="` + html.EscapeString(url) + `">Verify your email address</a>

		<p>
			Thanks,<br>
			Mathéo
		</p>

		<br>

		<h4>🌱 Why is this email ugly?</h4>
		<p>
			This email is voluntarily ugly because it's lightweight, so the environment impact is reduced..<br>
			By the way, this email is single-use, so you can delete it to avoid keeping it on someone's else hard drive 😎.
		</p>

		<p><small>If you didn't sign up to my website, please ignore this email.</small></p>
	`,
	})
	return err
}

// SignTokenPair godoc
// Signs a refresh token and an access token for the user with the provided id
func SignTokenPair(c context.Context, userId string) (string, string, error) {
	refreshtoken, err := utils.SignRefreshToken(userId, 168)
	if err != nil {
		return "", "", err
	}
	accesstoken, err := utils.SignAccessToken(userId, 1)
	if err != nil {
		return "", "", err
	}
	return refreshtoken, accesstoken, nil
}

func verificationLink(origin string, referer string, token string) (string, error) {
	if !slices.Contains(utils.AllowedOrigins, origin) {
		return "", errors.New("origin not allowed")
	}

	redirect := "/"
	if parsed, err := url.Parse(referer); err == nil && parsed.Path != "" {
		redirect = parsed.RequestURI()
	}
	// "//host" and "/\host" are read by browsers as another host
	if !strings.HasPrefix(redirect, "/") || strings.HasPrefix(redirect, "//") || strings.HasPrefix(redirect, "/\\") {
		redirect = "/"
	}

	query := url.Values{"token": {token}, "redirect": {redirect}}
	return origin + "/verify/email?" + query.Encode(), nil
}

var loginEmailLimiter = utils.NewRateLimiter(3, 15*time.Minute)

func newSession(ctx context.Context, user *users.User) (*Session, error) {
	refreshToken, accessToken, err := SignTokenPair(ctx, user.Id.Hex())
	if err != nil {
		return nil, err
	}
	return &Session{AccessToken: accessToken, RefreshToken: refreshToken, User: user}, nil
}

// RequestLoginLink emails a single-use login link, creating the user on their first visit
func RequestLoginLink(ctx context.Context, email string, origin string, referer string) error {
	if allowed, retryAfter := loginEmailLimiter.Allow(strings.ToLower(email)); !allowed {
		audit.Record(ctx, audit.Entry{Action: audit.AccessRateLimited, Meta: audit.Meta{"limit": "login_email", "email": email}})
		return utils.TooManyRequests(retryAfter)
	}

	user, _ := users.GetFullUserByEmail(ctx, email)
	if user == nil {
		result, err := users.CreateUser(ctx, &users.User{Email: email, Flavour: "mocha"})
		if err != nil {
			return err
		}
		if user, err = users.GetFullUserById(ctx, result.InsertedID.(primitive.ObjectID)); err != nil {
			return err
		}
		audit.Record(ctx, audit.Entry{Action: audit.UserCreated, UserId: &user.Id, Meta: audit.Meta{"via": users.IdentityProviderEmail}})
	}

	token, err := CreateLoginToken(ctx, user.Id)
	if err != nil {
		return err
	}
	link, err := verificationLink(origin, referer, token)
	if err != nil {
		return utils.Invalid(err.Error())
	}
	if err := SendVerificationEmail(ctx, user, link); err != nil {
		return err
	}
	audit.Record(ctx, audit.Entry{Action: audit.AuthLinkRequested, UserId: &user.Id})
	return nil
}

func loginFailed(ctx context.Context, method string, reason string, userId *primitive.ObjectID) {
	audit.Record(ctx, audit.Entry{Action: audit.AuthLoginFailed, UserId: userId, Meta: audit.Meta{"method": method, "reason": reason}})
}

func loggedIn(ctx context.Context, user *users.User, meta audit.Meta) {
	audit.Record(ctx, audit.Entry{Action: audit.AuthLogin, UserId: &user.Id, Meta: meta})
}

// VerifyLoginLink consumes the token of a login link and logs its user in
func VerifyLoginLink(ctx context.Context, token string) (*Session, error) {
	userId, err := ConsumeLoginToken(ctx, token)
	if err != nil {
		loginFailed(ctx, users.IdentityProviderEmail, "invalid_link", nil)
		return nil, utils.Invalid(err.Error())
	}
	user, err := users.GetFullUserById(ctx, userId)
	if err != nil {
		loginFailed(ctx, users.IdentityProviderEmail, "unknown_user", &userId)
		return nil, utils.Invalid(err.Error())
	}

	if err := utils.AddEmailContact(ctx, user.Email); err != nil {
		log.Printf("Failed to add %s to the contact list: %v", user.Email, err)
	}

	user.SetIdentity(users.Identity{
		Provider:   users.IdentityProviderEmail,
		ProviderId: user.Email,
		Email:      user.Email,
	})
	user.LastLogin = time.Now()
	user.LastRefresh = time.Now()
	if _, err = users.UpdateUser(ctx, userId, user); err != nil {
		return nil, err
	}
	loggedIn(ctx, user, audit.Meta{"method": users.IdentityProviderEmail})
	return newSession(ctx, user)
}

// RefreshSession trades a refresh token for a new pair of tokens
func RefreshSession(ctx context.Context, refreshToken string) (*Session, error) {
	secret := os.Getenv("REFRESH_TOKEN_SECRET")
	if _, err := utils.IsAuthorized(refreshToken, secret); err != nil {
		return nil, utils.Invalid(err.Error())
	}
	userId, err := utils.ExtractID(refreshToken, secret)
	if err != nil {
		return nil, utils.Invalid(err.Error())
	}

	newRefreshToken, accessToken, err := SignTokenPair(ctx, userId.Hex())
	if err != nil {
		return nil, err
	}
	return &Session{AccessToken: accessToken, RefreshToken: newRefreshToken}, nil
}

func FindOAuthProvider(name string) (*OAuthProvider, error) {
	provider := GetOAuthProvider(name)
	if provider == nil {
		return nil, utils.NotFound("Unknown identity provider")
	}
	return provider, nil
}

// LoginWithProvider attaches the provider account behind the code to the user having the same
// email, creating the user if needed. With an access token, the account must have the email
// of that user.
func LoginWithProvider(ctx context.Context, provider *OAuthProvider, code string, redirectURI string, accessToken string) (*Session, error) {
	var user *users.User
	if accessToken != "" {
		userId, err := utils.ExtractID(accessToken, os.Getenv("ACCESS_TOKEN_SECRET"))
		if err != nil {
			loginFailed(ctx, provider.Name, "invalid_access_token", nil)
			return nil, utils.Unauthorized(err.Error())
		}
		if user, err = users.GetFullUserById(ctx, userId); err != nil {
			loginFailed(ctx, provider.Name, "unknown_user", &userId)
			return nil, utils.Unauthorized("unauthorized")
		}
	}
	linking := user != nil

	identity, err := provider.Authenticate(ctx, code, redirectURI)
	if err != nil {
		log.Printf("Failed to authenticate with %s: %v", provider.Name, err)
		var userId *primitive.ObjectID
		if user != nil {
			userId = &user.Id
		}
		loginFailed(ctx, provider.Name, "provider_refused", userId)
		return nil, utils.Invalid("Failed to authenticate with " + provider.Name)
	}

	if user != nil {
		if !strings.EqualFold(user.Email, identity.Email) {
			loginFailed(ctx, provider.Name, "email_mismatch", &user.Id)
			return nil, utils.Conflict("This account uses another email address (" + identity.Email + ")")
		}
	} else if user, err = users.GetFullUserByEmail(ctx, identity.Email); err != nil {
		return nil, err
	}

	created := user == nil
	if created {
		result, err := users.CreateUser(ctx, &users.User{Email: identity.Email, Flavour: "mocha"})
		if err != nil {
			return nil, err
		}
		if user, err = users.GetFullUserById(ctx, result.InsertedID.(primitive.ObjectID)); err != nil {
			return nil, err
		}
		audit.Record(ctx, audit.Entry{Action: audit.UserCreated, UserId: &user.Id, Meta: audit.Meta{"via": provider.Name}})
		if err := utils.AddEmailContact(ctx, user.Email); err != nil {
			log.Printf("Failed to add %s to the contact list: %v", user.Email, err)
		}
	}

	if err := users.DetachIdentity(ctx, identity.Provider, identity.ProviderId, user.Id); err != nil {
		return nil, err
	}
	newIdentity := !created && user.GetIdentity(identity.Provider) == nil
	user.SetIdentity(*identity)
	if user.Name == "" {
		user.Name = identity.Name
	}
	if user.ProfilePicture == "" {
		user.ProfilePicture = identity.AvatarUrl
	}
	user.LastLogin = time.Now()
	user.LastRefresh = time.Now()
	if _, err := users.UpdateUser(ctx, user.Id, user); err != nil {
		return nil, err
	}
	if newIdentity {
		audit.Record(ctx, audit.Entry{Action: audit.UserIdentityLinked, UserId: &user.Id, Meta: audit.Meta{"provider": identity.Provider}})
	}
	if !linking {
		loggedIn(ctx, user, audit.Meta{"method": provider.Name})
	}
	return newSession(ctx, user)
}

const passkeyMethod = "passkey"

func WebAuthnFor(origin string) (*webauthn.WebAuthn, error) {
	webAuthn, err := NewWebAuthn(origin)
	if err != nil {
		return nil, utils.Invalid("Passkeys are not available from this origin")
	}
	return webAuthn, nil
}

// BeginPasskeyLogin asks for any passkey of the site: the passkey identifies the user
func BeginPasskeyLogin(ctx context.Context, webAuthn *webauthn.WebAuthn) (*PasskeyBeginResponse, error) {
	options, session, err := webAuthn.BeginDiscoverableLogin(webauthn.WithUserVerification(protocol.VerificationPreferred))
	if err != nil {
		return nil, err
	}
	sessionId, err := CreatePasskeyChallenge(ctx, session)
	if err != nil {
		return nil, err
	}
	return &PasskeyBeginResponse{SessionId: sessionId, Options: options}, nil
}

func FinishPasskeyLogin(ctx context.Context, webAuthn *webauthn.WebAuthn, body PasskeyFinishBody) (*Session, error) {
	session, err := ConsumePasskeyChallenge(ctx, body.SessionId)
	if err != nil {
		loginFailed(ctx, passkeyMethod, "invalid_challenge", nil)
		return nil, utils.Invalid(err.Error())
	}
	assertion, err := protocol.ParseCredentialRequestResponseBytes(body.Credential)
	if err != nil {
		loginFailed(ctx, passkeyMethod, "invalid_response", nil)
		return nil, utils.Invalid("Invalid passkey response")
	}

	findUser := func(rawID, userHandle []byte) (webauthn.User, error) {
		if len(userHandle) != len(primitive.NilObjectID) {
			return nil, errors.New("invalid user handle")
		}
		return users.GetFullUserById(ctx, primitive.ObjectID(userHandle))
	}
	owner, credential, err := webAuthn.ValidatePasskeyLogin(findUser, *session, assertion)
	if err != nil {
		log.Printf("Passkey login refused: %v", err)
		loginFailed(ctx, passkeyMethod, "not_recognized", nil)
		return nil, utils.Unauthorized("This passkey is not recognized")
	}
	user := owner.(*users.User)

	passkey := user.GetPasskey(users.PasskeyId(credential))
	if passkey == nil {
		loginFailed(ctx, passkeyMethod, "not_recognized", &user.Id)
		return nil, utils.Unauthorized("This passkey is not recognized")
	}
	// The stored credential carries the signature counter used to detect cloned authenticators
	if passkey.Credential, err = json.Marshal(credential); err != nil {
		return nil, err
	}
	now := time.Now()
	passkey.LastUsed = &now
	user.LastLogin = now
	user.LastRefresh = now
	if _, err := users.UpdateUser(ctx, user.Id, user); err != nil {
		return nil, err
	}
	loggedIn(ctx, user, audit.Meta{"method": passkeyMethod, "passkeyId": passkey.Id})
	return newSession(ctx, user)
}

func BeginPasskeyRegistration(ctx context.Context, webAuthn *webauthn.WebAuthn, user *users.User) (*PasskeyBeginResponse, error) {
	if len(user.Passkeys) >= users.MaxPasskeys {
		return nil, utils.Invalid("Too many passkeys, remove one first")
	}

	options, session, err := webAuthn.BeginRegistration(user,
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
		webauthn.WithExclusions(webauthn.Credentials(user.WebAuthnCredentials()).CredentialDescriptors()),
	)
	if err != nil {
		return nil, err
	}
	sessionId, err := CreatePasskeyChallenge(ctx, session)
	if err != nil {
		return nil, err
	}
	return &PasskeyBeginResponse{SessionId: sessionId, Options: options}, nil
}

func FinishPasskeyRegistration(ctx context.Context, webAuthn *webauthn.WebAuthn, user *users.User, body PasskeyFinishBody) (*users.User, error) {
	session, err := ConsumePasskeyChallenge(ctx, body.SessionId)
	if err != nil {
		return nil, utils.Invalid(err.Error())
	}
	attestation, err := protocol.ParseCredentialCreationResponseBytes(body.Credential)
	if err != nil {
		return nil, utils.Invalid("Invalid passkey response")
	}
	credential, err := webAuthn.CreateCredential(user, *session, attestation)
	if err != nil {
		log.Printf("Passkey registration refused: %v", err)
		return nil, utils.Invalid("This passkey could not be verified")
	}

	name := body.Name
	if name == "" {
		name = "Passkey"
	}
	passkey, err := users.NewPasskey(name, credential)
	if err != nil {
		return nil, err
	}
	user.Passkeys = append(user.Passkeys, *passkey)
	if _, err := users.UpdateUser(ctx, user.Id, user); err != nil {
		return nil, err
	}
	audit.Record(ctx, audit.Entry{Action: audit.UserPasskeyAdded, UserId: &user.Id, Meta: audit.Meta{"passkeyId": passkey.Id, "name": passkey.Name}})
	return user, nil
}

func RemovePasskey(ctx context.Context, user *users.User, passkeyId string) (*users.User, error) {
	if !user.RemovePasskey(passkeyId) {
		return nil, utils.NotFound("Passkey not found")
	}
	if _, err := users.UpdateUser(ctx, user.Id, user); err != nil {
		return nil, err
	}
	audit.Record(ctx, audit.Entry{Action: audit.UserPasskeyRemoved, UserId: &user.Id, Meta: audit.Meta{"passkeyId": passkeyId}})
	return user, nil
}
