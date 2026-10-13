package user

import (
	"context"
	"dev/internal/utils"
	"errors"
	"log"
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

func notFound(err error) error {
	if errors.Is(err, mongo.ErrNoDocuments) {
		return utils.NotFound("User not found")
	}
	return err
}

// GetOwnUser returns the user only to themselves
func GetOwnUser(ctx context.Context, requester primitive.ObjectID, id primitive.ObjectID) (*User, error) {
	if id != requester {
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

	if changes.Name != "" {
		user.Name = changes.Name
	}
	if changes.Flavour != "" {
		user.Flavour = changes.Flavour
	}
	if changes.ProfilePicture != "" {
		user.ProfilePicture = changes.ProfilePicture
	}

	if _, err := UpdateUser(ctx, id, user); err != nil {
		return nil, err
	}
	return user, nil
}

// UnlinkIdentity never removes the email identity, so the user can always log in with a magic link
func UnlinkIdentity(ctx context.Context, requester primitive.ObjectID, id primitive.ObjectID, provider string) (*User, error) {
	if id != requester {
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

	// The account is already deleted: a notice that cannot be sent must not report a failure
	if err := SendDeletionEmail(ctx, user); err != nil {
		log.Printf("Failed to send the deletion notice to %s: %v", user.Email, err)
	}
	return nil
}
