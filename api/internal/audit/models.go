package audit

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Retention is how long an entry is kept before the database expires it
const Retention = 180 * 24 * time.Hour

type Action string

const (
	AuthLinkRequested Action = "auth.link_requested"
	AuthLogin         Action = "auth.login"
	AuthLoginFailed   Action = "auth.login_failed"

	UserCreated          Action = "user.created"
	UserIdentityLinked   Action = "user.identity_linked"
	UserIdentityUnlinked Action = "user.identity_unlinked"
	UserPasskeyAdded     Action = "user.passkey_added"
	UserPasskeyRemoved   Action = "user.passkey_removed"
	UserProfileUpdated   Action = "user.profile_updated"
	UserDataExported     Action = "user.data_exported"
	UserDeleted          Action = "user.deleted"

	WordCloudCreated     Action = "wordcloud.created"
	WordCloudDuplicated  Action = "wordcloud.duplicated"
	WordCloudUpdated     Action = "wordcloud.updated"
	WordCloudDeleted     Action = "wordcloud.deleted"
	WordCloudWordHidden  Action = "wordcloud.word_hidden"
	WordCloudWordsMerged Action = "wordcloud.words_merged"

	ContactSent Action = "contact.sent"

	AccessDenied      Action = "access.denied"
	AccessRateLimited Action = "access.rate_limited"
)

const (
	TargetUser      = "user"
	TargetWordCloud = "wordcloud"
	TargetContact   = "contact"
)

type Meta = map[string]any

type Target struct {
	Type string `json:"type" bson:"type"`
	Id   string `json:"id" bson:"id"`
}

type Entry struct {
	Id        primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	Action    Action              `json:"action" bson:"action"`
	CreatedAt time.Time           `json:"createdAt" bson:"createdAt"`
	UserId    *primitive.ObjectID `json:"userId,omitempty" bson:"userId,omitempty"`
	Ip        string              `json:"ip,omitempty" bson:"ip,omitempty"`
	UserAgent string              `json:"userAgent,omitempty" bson:"userAgent,omitempty"`
	RequestId string              `json:"requestId,omitempty" bson:"requestId,omitempty"`
	Target    *Target             `json:"target,omitempty" bson:"target,omitempty"`
	Meta      Meta                `json:"meta,omitempty" bson:"meta,omitempty"`
}
