package admin

import (
	"context"
	"dev/internal/audit"
	users "dev/internal/user"
	"dev/internal/utils"
	"errors"
	"log"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func ListUsers(ctx context.Context, query UsersQuery) (*UsersPage, error) {
	query.normalize()
	found, total, err := users.ListUsers(ctx, query.Search, query.Page, query.Limit)
	if err != nil {
		return nil, err
	}
	return &UsersPage{Users: found, Total: total, Page: query.Page, Limit: query.Limit}, nil
}

func GetUser(ctx context.Context, id primitive.ObjectID) (*users.User, error) {
	user, err := users.GetFullUserById(ctx, id)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, utils.NotFound("User not found")
	}
	return user, err
}

// protected returns the user unless an admin must not change it: themselves, so that an
// admin always remains, and the site owner
func protected(ctx context.Context, requester primitive.ObjectID, id primitive.ObjectID) (*users.User, error) {
	if id == requester {
		return nil, utils.Invalid("You cannot do this to your own account")
	}
	user, err := GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	if users.IsRootAdmin(user.Email) {
		return nil, utils.Forbidden("The site owner cannot be changed")
	}
	return user, nil
}

func target(id primitive.ObjectID) *audit.Target {
	return &audit.Target{Type: audit.TargetUser, Id: id.Hex()}
}

func ChangeRole(ctx context.Context, requester primitive.ObjectID, id primitive.ObjectID, role string) (*users.User, error) {
	user, err := protected(ctx, requester, id)
	if err != nil {
		return nil, err
	}
	if user.Role == role {
		return user, nil
	}
	if err := users.SetUserRole(ctx, id, role); err != nil {
		return nil, err
	}
	user.Role = role
	audit.Record(ctx, audit.Entry{Action: audit.AdminRoleChanged, UserId: &requester, Target: target(id), Meta: audit.Meta{"role": role}})
	return user, nil
}

func DeleteUser(ctx context.Context, requester primitive.ObjectID, id primitive.ObjectID) error {
	user, err := protected(ctx, requester, id)
	if err != nil {
		return err
	}
	if _, err := users.DeleteUser(ctx, id); err != nil {
		return err
	}
	audit.Record(ctx, audit.Entry{Action: audit.UserDeleted, UserId: &requester, Target: target(id), Meta: audit.Meta{"by": "admin"}})

	if err := users.SendDeletionEmail(ctx, user); err != nil {
		log.Printf("Failed to send the deletion notice to %s: %v", user.Email, err)
	}
	return nil
}

func ListAudit(ctx context.Context, query AuditQuery) (*AuditPage, error) {
	query.normalize()
	filter := audit.Filter{Action: audit.Action(query.Action)}
	if query.UserId != "" {
		userId, err := primitive.ObjectIDFromHex(query.UserId)
		if err != nil {
			return nil, utils.Invalid("Invalid user id")
		}
		filter.UserId = &userId
	}
	entries, total, err := audit.ListEntries(ctx, filter, query.Page, query.Limit)
	if err != nil {
		return nil, err
	}

	ids := []primitive.ObjectID{}
	for _, entry := range entries {
		if entry.UserId != nil {
			ids = append(ids, *entry.UserId)
		}
		if entry.Target != nil && entry.Target.Type == audit.TargetUser {
			if id, err := primitive.ObjectIDFromHex(entry.Target.Id); err == nil {
				ids = append(ids, id)
			}
		}
	}
	mentioned, err := users.GetUsersByIds(ctx, ids)
	if err != nil {
		return nil, err
	}
	summaries := map[string]UserSummary{}
	for _, user := range mentioned {
		summaries[user.Id.Hex()] = UserSummary{Name: user.Name, Email: user.Email}
	}
	return &AuditPage{Entries: entries, Users: summaries, Total: total, Page: query.Page, Limit: query.Limit}, nil
}
