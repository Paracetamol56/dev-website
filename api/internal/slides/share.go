package slides

import (
	"context"
	"crypto/rand"
	"dev/internal/audit"
	"dev/internal/user"
	"dev/internal/utils"
	"encoding/base64"
	"errors"
	"fmt"
	"html"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const defaultSite = "https://dev.matheo-galuba.com"

var emailLimiter = utils.NewRateLimiter(10, time.Hour)

func validLinkRole(role string) error {
	if role != RoleEditor && role != RoleViewer {
		return utils.Invalid("The role must be editor or viewer")
	}
	return nil
}

func newToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func JoinURL(site string, token string) string {
	return strings.TrimRight(site, "/") + "/tools/slides/join/" + token
}

// CreateLink turns a share link on, or replaces its token so the old link stops working
func CreateLink(ctx context.Context, userId primitive.ObjectID, id primitive.ObjectID, role string) (*Links, error) {
	if err := validLinkRole(role); err != nil {
		return nil, err
	}
	deck, err := ownedDeck(ctx, userId, id)
	if err != nil {
		return nil, err
	}
	token, err := newToken()
	if err != nil {
		return nil, err
	}
	if err := SetLink(ctx, id, role, &token); err != nil {
		return nil, err
	}
	audit.Record(ctx, audit.Entry{Action: audit.SlidesLinkCreated, UserId: &userId, Target: target(id), Meta: audit.Meta{"role": role}})
	links := deck.Links
	if role == RoleEditor {
		links.Editor = &token
	} else {
		links.Viewer = &token
	}
	return &links, nil
}

// RevokeLink turns a share link off; the members who joined through it stay
func RevokeLink(ctx context.Context, userId primitive.ObjectID, id primitive.ObjectID, role string) error {
	if err := validLinkRole(role); err != nil {
		return err
	}
	if _, err := ownedDeck(ctx, userId, id); err != nil {
		return err
	}
	if err := SetLink(ctx, id, role, nil); err != nil {
		return err
	}
	audit.Record(ctx, audit.Entry{Action: audit.SlidesLinkRevoked, UserId: &userId, Target: target(id), Meta: audit.Meta{"role": role}})
	return nil
}

// JoinDeck makes the user a member of the deck of a link. An editor link upgrades a viewer,
// a viewer link never downgrades anyone.
func JoinDeck(ctx context.Context, userId primitive.ObjectID, token string) (*JoinedDeck, error) {
	var deck *Deck
	role := ""
	for _, candidate := range []string{RoleEditor, RoleViewer} {
		found, err := GetDeckByLink(ctx, candidate, token)
		if errors.Is(err, mongo.ErrNoDocuments) {
			continue
		}
		if err != nil {
			return nil, err
		}
		deck, role = found, candidate
		break
	}
	if deck == nil {
		return nil, utils.NotFound("This link does not work anymore")
	}

	current := access(deck, userId)
	switch {
	case current == RoleOwner || current == RoleEditor || current == role:
		return &JoinedDeck{Id: deck.Id, Role: current}, nil
	case current == RoleViewer:
		if _, err := SetMemberRole(ctx, deck.Id, userId, RoleEditor); err != nil {
			return nil, err
		}
	default:
		member := Member{UserId: userId, Role: role, JoinedAt: primitive.NewDateTimeFromTime(time.Now())}
		if err := AddMember(ctx, deck.Id, member); err != nil {
			return nil, err
		}
	}
	audit.Record(ctx, audit.Entry{Action: audit.SlidesJoined, UserId: &userId, Target: target(deck.Id), Meta: audit.Meta{"role": role}})
	notifyAccess(ctx, deck.Id)
	return &JoinedDeck{Id: deck.Id, Role: role}, nil
}

func ChangeMemberRole(ctx context.Context, userId primitive.ObjectID, id primitive.ObjectID, memberId primitive.ObjectID, role string) error {
	if _, err := ownedDeck(ctx, userId, id); err != nil {
		return err
	}
	found, err := SetMemberRole(ctx, id, memberId, role)
	if err != nil {
		return err
	}
	if !found {
		return utils.NotFound("Member not found")
	}
	audit.Record(ctx, audit.Entry{Action: audit.SlidesMemberRoleChanged, UserId: &userId, Target: target(id), Meta: audit.Meta{"member": memberId.Hex(), "role": role}})
	notifyAccess(ctx, id)
	return nil
}

// RemoveMember takes a member out of a deck: the owner can remove anyone, a member can leave
func RemoveMember(ctx context.Context, userId primitive.ObjectID, id primitive.ObjectID, memberId primitive.ObjectID) error {
	_, role, err := deckAccess(ctx, userId, id)
	if err != nil {
		return err
	}
	if role != RoleOwner && userId != memberId {
		return utils.Forbidden("Only the owner of the deck can remove other members")
	}
	removed, err := DeleteMember(ctx, id, memberId)
	if err != nil {
		return err
	}
	if !removed {
		return utils.NotFound("Member not found")
	}
	audit.Record(ctx, audit.Entry{Action: audit.SlidesMemberRemoved, UserId: &userId, Target: target(id), Meta: audit.Meta{"member": memberId.Hex()}})
	notifyAccess(ctx, id)
	return nil
}

// EmailLink sends a share link, turning it on first if needed. site is where the link points.
func EmailLink(ctx context.Context, userId primitive.ObjectID, id primitive.ObjectID, role string, address string, site string) error {
	if err := validLinkRole(role); err != nil {
		return err
	}
	deck, err := ownedDeck(ctx, userId, id)
	if err != nil {
		return err
	}
	if allowed, retryAfter := emailLimiter.Allow(userId.Hex()); !allowed {
		audit.Record(ctx, audit.Entry{Action: audit.AccessRateLimited, UserId: &userId, Target: target(id), Meta: audit.Meta{"limit": "slides.email"}})
		return utils.TooManyRequests(retryAfter)
	}
	token := deck.Links.Viewer
	if role == RoleEditor {
		token = deck.Links.Editor
	}
	if token == nil {
		links, err := CreateLink(ctx, userId, id, role)
		if err != nil {
			return err
		}
		token = links.Viewer
		if role == RoleEditor {
			token = links.Editor
		}
	}

	sender := "Someone"
	if account, err := user.GetFullUserById(ctx, userId); err == nil {
		sender = account.WebAuthnDisplayName()
	}
	action := "view"
	if role == RoleEditor {
		action = "edit"
	}
	link := JoinURL(site, *token)
	if _, err := utils.SendEmail(ctx, utils.Email{
		To:      utils.EmailAddress{Email: address},
		Subject: fmt.Sprintf("%s shared a slide deck with you", sender),
		HTMLContent: fmt.Sprintf(`
		<h1>%s shared a slide deck with you</h1>
		<p>
			Hi,<br>
			%s invited you to %s the slide deck <strong>%s</strong>.<br>
			<a href="%s">Open the deck</a>, you will be asked to log in first.
		</p>

		<p>
			If you did not expect this email, you can ignore it.
		</p>
	`, html.EscapeString(sender), html.EscapeString(sender), action, html.EscapeString(deck.Title), html.EscapeString(link)),
		TextContent: fmt.Sprintf("%s invited you to %s the slide deck \"%s\": %s", sender, action, deck.Title, link),
	}); err != nil {
		return errors.New("the email could not be sent")
	}
	domain := address[strings.LastIndex(address, "@")+1:]
	audit.Record(ctx, audit.Entry{Action: audit.SlidesLinkEmailed, UserId: &userId, Target: target(id), Meta: audit.Meta{"role": role, "domain": domain}})
	return nil
}

// notifyAccess tells the open deck that its members changed, so it can update its clients
func notifyAccess(ctx context.Context, id primitive.ObjectID) {
	deck, err := GetDeckById(ctx, id)
	if err != nil {
		return
	}
	hubs.access(id, deck.Members)
}
