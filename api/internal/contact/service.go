package contact

import (
	"context"
	"dev/internal/audit"
	"dev/internal/utils"
	"errors"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SendMessage forwards the message to the admin and keeps a copy
func SendMessage(ctx context.Context, message ContactBody) error {
	mailId, err := utils.SendEmail(ctx, utils.Email{
		To:          utils.EmailAddress{Name: "Admin", Email: os.Getenv("ADMIN_EMAIL")},
		ReplyTo:     &utils.EmailAddress{Name: message.Name, Email: message.Email},
		Subject:     "New message from dev.matheo-galuba.com",
		TextContent: "User: " + message.UserId.String() + "\nName: " + message.Name + "\nEmail: " + message.Email + "\nMessage: " + message.Message,
	})
	if err != nil {
		return errors.New("Error while sending email")
	}

	id, err := CreateContact(ctx, Contact{
		UserId:    message.UserId,
		Name:      message.Name,
		Email:     message.Email,
		Message:   message.Message,
		MailId:    mailId,
		CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
	})
	if err != nil {
		return err
	}
	audit.Record(ctx, audit.Entry{Action: audit.ContactSent, Target: &audit.Target{Type: audit.TargetContact, Id: id.Hex()}})
	return nil
}
