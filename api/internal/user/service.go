package user

import (
	"context"
	"dev/internal/audit"
	"dev/internal/utils"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func ScheduleAccountDeletion() {
	s := gocron.NewScheduler(time.UTC)
	s.Every(1).Day().At("00:00").Do(func() {
		DeleteOldUsers(context.Background())
	})
	s.StartAsync()
}

// SendDeletionEmail godoc
// Sends a notification email to the user when their account is deleted
func SendDeletionEmail(c context.Context, user *User) error {
	_, err := utils.SendEmail(c, utils.Email{
		To:      utils.EmailAddress{Name: user.Name, Email: user.Email},
		Subject: "Your account has been deleted",
		HTMLContent: `
		<h1>Your account has been deleted</h1>
		<p>
			Hi,<br>
			I received a request to delete your account.<br>
			The access to your account has been removed and all your data will be permanently deleted in 30 days.<br>
			If this is an error, please <a href="https://dev.matheo-galuba.com/contact">contact me</a> as soon as possible.
		</p>

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
	`,
	})
	return err
}

// IsRootAdmin tells whether the email is the one of the site owner, who is always an admin
func IsRootAdmin(email string) bool {
	root := os.Getenv("ADMIN_EMAIL")
	return root != "" && strings.EqualFold(email, root)
}

// PromoteRootAdmin makes the site owner an admin. It must only be called for a user whose
// email is proven: at startup for an existing account, then after a login.
func PromoteRootAdmin(ctx context.Context, user *User) error {
	if user == nil || user.IsAdmin() || !IsRootAdmin(user.Email) {
		return nil
	}
	if err := SetUserRole(ctx, user.Id, RoleAdmin); err != nil {
		return err
	}
	user.Role = RoleAdmin
	audit.Record(ctx, audit.Entry{
		Action: audit.AdminRoleChanged,
		UserId: &user.Id,
		Target: &audit.Target{Type: audit.TargetUser, Id: user.Id.Hex()},
		Meta:   audit.Meta{"role": RoleAdmin, "by": "root"},
	})
	return nil
}

func EnsureRootAdmin(ctx context.Context) error {
	email := os.Getenv("ADMIN_EMAIL")
	if email == "" {
		return nil
	}
	user, err := GetFullUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	return PromoteRootAdmin(ctx, user)
}

func notFound(err error) error {
	if errors.Is(err, mongo.ErrNoDocuments) {
		return utils.NotFound("User not found")
	}
	return err
}

func denied(ctx context.Context, requester primitive.ObjectID, id primitive.ObjectID) {
	audit.Record(ctx, audit.Entry{
		Action: audit.AccessDenied,
		UserId: &requester,
		Target: &audit.Target{Type: audit.TargetUser, Id: id.Hex()},
	})
}

// GetOwnUser returns the user only to themselves
func GetOwnUser(ctx context.Context, requester primitive.ObjectID, id primitive.ObjectID) (*User, error) {
	if id != requester {
		denied(ctx, requester, id)
		return nil, utils.Forbidden("Forbidden")
	}
	user, err := GetFullUserById(ctx, id)
	return user, notFound(err)
}

// GetProfile returns the whole user to themselves and a public profile to anyone else
func GetProfile(ctx context.Context, requester primitive.ObjectID, id primitive.ObjectID) (any, error) {
	if id == requester {
		user, err := GetFullUserById(ctx, id)
		return user, notFound(err)
	}
	user, err := GetUserById(ctx, id)
	return user, notFound(err)
}

func UpdateProfile(ctx context.Context, requester primitive.ObjectID, id primitive.ObjectID, changes PatchUserBody) (*User, error) {
	user, err := GetOwnUser(ctx, requester, id)
	if err != nil {
		return nil, err
	}

	changed := []string{}
	if changes.Name != "" && changes.Name != user.Name {
		user.Name = changes.Name
		changed = append(changed, "name")
	}
	if changes.Flavour != "" {
		user.Flavour = changes.Flavour
	}
	if changes.ProfilePicture != "" && changes.ProfilePicture != user.ProfilePicture {
		user.ProfilePicture = changes.ProfilePicture
		changed = append(changed, "profilePicture")
	}

	if _, err := UpdateUser(ctx, id, user); err != nil {
		return nil, err
	}
	if len(changed) > 0 {
		audit.Record(ctx, audit.Entry{Action: audit.UserProfileUpdated, UserId: &id, Meta: audit.Meta{"fields": changed}})
	}
	return user, nil
}

// UnlinkIdentity never removes the email identity, so the user can always log in with a magic link
func UnlinkIdentity(ctx context.Context, requester primitive.ObjectID, id primitive.ObjectID, provider string) (*User, error) {
	if id != requester {
		denied(ctx, requester, id)
		return nil, utils.Forbidden("Forbidden")
	}
	if provider == IdentityProviderEmail {
		return nil, utils.Invalid("The email identity cannot be removed")
	}

	user, err := GetFullUserById(ctx, id)
	if err != nil {
		return nil, notFound(err)
	}
	if !user.RemoveIdentity(provider) {
		return nil, utils.NotFound("Identity not found")
	}
	if _, err := UpdateUser(ctx, id, user); err != nil {
		return nil, err
	}
	audit.Record(ctx, audit.Entry{Action: audit.UserIdentityUnlinked, UserId: &id, Meta: audit.Meta{"provider": provider}})
	return user, nil
}

func ExportData(ctx context.Context, requester primitive.ObjectID, id primitive.ObjectID) (*User, error) {
	user, err := GetOwnUser(ctx, requester, id)
	if err != nil {
		return nil, err
	}
	audit.Record(ctx, audit.Entry{Action: audit.UserDataExported, UserId: &id})
	return user, nil
}

func DeleteAccount(ctx context.Context, requester primitive.ObjectID, id primitive.ObjectID) error {
	user, err := GetOwnUser(ctx, requester, id)
	if err != nil {
		return err
	}

	result, err := DeleteUser(ctx, id)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return utils.NotFound("User not found")
	}
	audit.Record(ctx, audit.Entry{Action: audit.UserDeleted, UserId: &id, Meta: audit.Meta{"by": "user"}})

	// The account is already deleted: a notice that cannot be sent must not report a failure
	if err := SendDeletionEmail(ctx, user); err != nil {
		log.Printf("Failed to send the deletion notice to %s: %v", user.Email, err)
	}
	return nil
}
