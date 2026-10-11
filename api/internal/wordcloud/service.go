package wordcloud

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"dev/internal/utils"
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"math/big"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/text/unicode/norm"
)

var ErrInvalidParticipant = errors.New("invalid participant token")

// Tokens are "<uuid>.<hmac of session and uuid>", so they cannot be forged or moved to another session.
func participantSignature(sessionId string, participant uuid.UUID) []byte {
	mac := hmac.New(sha256.New, []byte("word-cloud:"+os.Getenv("ACCESS_TOKEN_SECRET")))
	mac.Write([]byte(sessionId + ":" + participant.String()))
	return mac.Sum(nil)
}

func SignParticipant(sessionId string, participant uuid.UUID) string {
	return participant.String() + "." + base64.RawURLEncoding.EncodeToString(participantSignature(sessionId, participant))
}

func VerifyParticipant(sessionId string, token string) (uuid.UUID, error) {
	id, signature, found := strings.Cut(token, ".")
	if !found {
		return uuid.Nil, ErrInvalidParticipant
	}
	participant, err := uuid.FromString(id)
	if err != nil {
		return uuid.Nil, ErrInvalidParticipant
	}
	decoded, err := base64.RawURLEncoding.DecodeString(signature)
	if err != nil || !hmac.Equal(decoded, participantSignature(sessionId, participant)) {
		return uuid.Nil, ErrInvalidParticipant
	}
	return participant, nil
}

// Fingerprint returns a short keyed hash, so values can be compared without being stored.
func Fingerprint(kind string, value string) string {
	mac := hmac.New(sha256.New, []byte("word-cloud-"+kind+":"+os.Getenv("ACCESS_TOKEN_SECRET")))
	mac.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))[:16]
}

// Lists from github.com/LDNOOBW/List-of-Dirty-Naughty-Obscene-and-Otherwise-Bad-Words (CC BY 4.0, see LICENSE there)
//
//go:embed data/profanity/*.txt
var profanityFiles embed.FS

var profanities = func() map[string]bool {
	words := map[string]bool{}
	entries, _ := profanityFiles.ReadDir("data/profanity")
	for _, entry := range entries {
		content, _ := profanityFiles.ReadFile("data/profanity/" + entry.Name())
		for _, line := range strings.Split(string(content), "\n") {
			// Entries with symbols would lose characters when cleaned and become too broad
			lower := strings.Join(strings.Fields(removeDiacritics(strings.ToLower(line))), " ")
			if cleaned := CleanText(line); cleaned != "" && cleaned == lower {
				words[cleaned] = true
			}
		}
	}
	return words
}()

// IsProfane expects a cleaned submission; each of its words is checked too.
func IsProfane(word string) bool {
	if profanities[word] {
		return true
	}
	for _, part := range wordParts(word) {
		if profanities[part] {
			return true
		}
	}
	return false
}

// AFINN lists by Finn Årup Nielsen, github.com/fnielsen/afinn (Apache 2.0, see LICENSE there)
//
//go:embed data/sentiment/*.txt
var afinnFiles embed.FS

// Words listed more than once, in one language or several, get their average score
var afinn = func() map[string]int {
	totals, counts := map[string]int{}, map[string]int{}
	entries, _ := afinnFiles.ReadDir("data/sentiment")
	for _, entry := range entries {
		content, _ := afinnFiles.ReadFile("data/sentiment/" + entry.Name())
		for _, line := range strings.Split(string(content), "\n") {
			word, score, found := strings.Cut(line, "\t")
			value, err := strconv.Atoi(strings.TrimSpace(score))
			if !found || err != nil {
				continue
			}
			if key := CleanText(word); key != "" {
				totals[key] += value
				counts[key]++
			}
		}
	}
	scores := make(map[string]int, len(totals))
	for key, total := range totals {
		scores[key] = int(math.Round(float64(total) / float64(counts[key])))
	}
	return scores
}()

// Sentiment scores a cleaned submission from -5 to 5; ok is false when it is not rated.
// Unknown phrases are rated by the average of their rated words.
func Sentiment(word string) (score int, ok bool) {
	if score, ok := afinn[word]; ok {
		return score, true
	}
	total, rated := 0, 0
	for _, part := range wordParts(word) {
		if score, ok := afinn[part]; ok {
			total += score
			rated++
		}
	}
	if rated == 0 {
		return 0, false
	}
	return int(math.Round(float64(total) / float64(rated))), true
}

var nonWordRegex = regexp.MustCompile(`[^\p{L}\p{N}\s_-]+`)

func removeDiacritics(text string) string {
	return strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Mn, r) {
			return -1
		}
		return r
	}, norm.NFD.String(text))
}

// CleanText normalises a submission: lowercase, without diacritics, keeping letters and digits
// of any script, underscores, hyphens and single spaces between words.
func CleanText(text string) string {
	return strings.Join(strings.Fields(nonWordRegex.ReplaceAllString(removeDiacritics(strings.ToLower(text)), "")), " ")
}

// wordParts splits a cleaned submission into its words, hyphenated parts included.
func wordParts(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool { return r == ' ' || r == '-' || r == '_' })
}

const (
	codeLength = 5
	// No 0/O or 1/I/L, which are easily confused
	codeAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"
	// Per participant rather than per IP: audiences often share one network
	wordCooldown = time.Second
	// Participants are offered more than they need, so their own words can be left out
	recentWords       = 8
	socketAuthTimeout = 5 * time.Second
)

var (
	codeRegex          = regexp.MustCompile("^[a-zA-Z0-9]+$")
	participantLimiter = utils.NewRateLimiter(1, wordCooldown)
)

func GenerateCode() string {
	code := make([]byte, codeLength)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(codeAlphabet))))
		if err != nil {
			panic(err) // crypto/rand does not fail on supported platforms
		}
		code[i] = codeAlphabet[n.Int64()]
	}
	return string(code)
}

func newUniqueCode(c context.Context) (string, error) {
	for range 10 {
		code := GenerateCode()
		inUse, err := IsCodeInUse(c, code)
		if err != nil {
			return "", err
		}
		if !inUse {
			return code, nil
		}
	}
	return "", errors.New("could not generate a unique code")
}

func ValidateCode(code string) error {
	if len(code) != codeLength {
		return fmt.Errorf("code must be %d characters long", codeLength)
	}
	if !codeRegex.MatchString(code) {
		return fmt.Errorf("code must contain only letters and numbers")
	}
	return nil
}

func isOpen(wordCloud *WordCloud) bool {
	return wordCloud.ClosedAt == nil
}

// recentDistinct returns up to count texts, most recent first, without repeats.
func recentDistinct(texts []string, count int) []string {
	recent := []string{}
	for i := len(texts) - 1; i >= 0 && len(recent) < count; i-- {
		if !slices.Contains(recent, texts[i]) {
			recent = append(recent, texts[i])
		}
	}
	return recent
}

func visibleTexts(wordCloud *WordCloud) []string {
	texts := []string{}
	for _, word := range wordCloud.Words {
		if !word.Hidden {
			texts = append(texts, word.Text)
		}
	}
	return texts
}

// broadcastSession sends the session to everyone, and the moderated words to owners.
func broadcastSession(wordCloud *WordCloud) {
	broadcast(wordCloud.Id,
		map[string]any{"type": "session", "session": PublicView(wordCloud), "words": withSentiment(wordCloud, wordCloud.Words), "settings": wordCloud.Settings},
		map[string]any{"type": "session", "session": PublicView(wordCloud)})
}

type socketClient struct {
	conn  *websocket.Conn
	owner bool
	// gorilla/websocket allows one concurrent writer per connection
	write sync.Mutex
}

func (client *socketClient) send(event any) error {
	client.write.Lock()
	defer client.write.Unlock()
	return client.conn.WriteJSON(event)
}

var (
	socketClients = make(map[primitive.ObjectID][]*socketClient)
	socketMutex   = sync.Mutex{}
	upgrader      = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

// broadcast sends ownerEvent to owners and publicEvent to other clients; nil events are skipped.
func broadcast(sessionId primitive.ObjectID, ownerEvent any, publicEvent any) {
	socketMutex.Lock()
	clients := append([]*socketClient(nil), socketClients[sessionId]...)
	socketMutex.Unlock()

	for _, client := range clients {
		event := publicEvent
		if client.owner {
			event = ownerEvent
		}
		if event == nil {
			continue
		}
		if err := client.send(event); err != nil {
			log.Println("Error writing to word cloud socket:", err)
			client.conn.Close()
		}
	}
}

// withSentiment returns a copy: the session is read by the response and by broadcasts at the same time
func withSentiment(wordCloud *WordCloud, words []Word) []Word {
	rated := make([]Word, len(words))
	copy(rated, words)
	for i := range rated {
		if !wordCloud.Settings.Sentiment {
			rated[i].Sentiment = nil
		} else if score, ok := Sentiment(rated[i].Text); ok {
			rated[i].Sentiment = &score
		} else {
			rated[i].Sentiment = nil
		}
	}
	return rated
}

// PublicView is what participants may know of a session
func PublicView(wordCloud *WordCloud) map[string]any {
	session := map[string]any{
		"id":          wordCloud.Id,
		"name":        wordCloud.Name,
		"description": wordCloud.Description,
		"code":        wordCloud.Code,
		"open":        isOpen(wordCloud),
		"settings": map[string]any{
			"showCloudToAudience": wordCloud.Settings.ShowCloudToAudience,
			"allowUpvotes":        wordCloud.Settings.AllowUpvotes,
			"maxWords":            wordCloud.Settings.WordLimit(),
		},
	}
	texts := visibleTexts(wordCloud)
	if wordCloud.Settings.ShowCloudToAudience {
		session["words"] = texts
	}
	if wordCloud.Settings.AllowUpvotes {
		session["recent"] = recentDistinct(texts, recentWords)
	}
	return session
}

func sessionError(err error) error {
	if errors.Is(err, mongo.ErrNoDocuments) {
		return utils.NotFound("word cloud not found")
	}
	return err
}

// GetByCode finds the open session participants join with a code
func GetByCode(ctx context.Context, code string) (*WordCloud, error) {
	if err := ValidateCode(code); err != nil {
		return nil, utils.Invalid(err.Error())
	}
	wordCloud, err := GetWordCloudByCode(ctx, code)
	return wordCloud, sessionError(err)
}

// ListSessions only lists a user's sessions to that user
func ListSessions(ctx context.Context, requester primitive.ObjectID, authenticated bool, userId primitive.ObjectID, status string) ([]*WordCloud, error) {
	if status != "open" && status != "closed" {
		return nil, utils.Invalid("invalid status")
	}
	if !authenticated || userId != requester {
		return nil, utils.Unauthorized("unauthorized")
	}
	return GetWordCloudByUser(ctx, requester, status)
}

func GetSession(ctx context.Context, id primitive.ObjectID) (*WordCloud, error) {
	wordCloud, err := GetWordCloudById(ctx, id)
	return wordCloud, sessionError(err)
}

func GetOwnedSession(ctx context.Context, id primitive.ObjectID, userId primitive.ObjectID) (*WordCloud, error) {
	wordCloud, err := GetSession(ctx, id)
	if err != nil {
		return nil, err
	}
	if wordCloud.UserId != userId {
		return nil, utils.Forbidden("forbidden")
	}
	return wordCloud, nil
}

func IsOwner(wordCloud *WordCloud, userId primitive.ObjectID, authenticated bool) bool {
	return authenticated && userId == wordCloud.UserId
}

// OwnerView completes the session with what only its owner sees
func OwnerView(ctx context.Context, wordCloud *WordCloud) (*WordCloud, error) {
	if err := EnsureWordIds(ctx, wordCloud); err != nil {
		return nil, err
	}
	wordCloud.Words = withSentiment(wordCloud, wordCloud.Words)
	return wordCloud, nil
}

// Join hands out the token a participant signs their words with
func Join(wordCloud *WordCloud) (string, error) {
	if !isOpen(wordCloud) {
		return "", utils.Conflict("word cloud closed")
	}
	return SignParticipant(wordCloud.Id.Hex(), uuid.NewV4()), nil
}

type Submission struct {
	Text      string
	Token     string
	Client    *ClientInfo
	UserAgent string
	IP        string
}

// SubmitWord stores the cleaned word and returns it
func SubmitWord(ctx context.Context, wordCloud *WordCloud, submission Submission) (string, error) {
	participant, err := VerifyParticipant(wordCloud.Id.Hex(), submission.Token)
	if err != nil {
		return "", utils.Unauthorized(err.Error())
	}
	if !isOpen(wordCloud) {
		return "", utils.Conflict("word cloud closed")
	}

	text := CleanText(submission.Text)
	if text == "" {
		return "", utils.Invalid("word is empty")
	}
	submitted := 0
	for _, word := range wordCloud.Words {
		if word.UUID == participant {
			submitted++
		}
	}
	if limit := wordCloud.Settings.WordLimit(); limit > 0 && submitted >= limit {
		return "", utils.Forbidden("word limit reached")
	}
	if allowed, retryAfter := participantLimiter.Allow(wordCloud.Id.Hex() + ":" + participant.String()); !allowed {
		return "", utils.TooManyRequests(retryAfter)
	}

	word := Word{
		Id:        primitive.NewObjectID(),
		Text:      text,
		Hidden:    wordCloud.Settings.FilterProfanity && IsProfane(text),
		UUID:      participant,
		UserAgent: submission.UserAgent,
		CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
		Client:    submission.Client,
		Network:   Fingerprint("network", submission.IP),
	}
	if submission.Client != nil {
		details, _ := json.Marshal(submission.Client)
		word.Device = Fingerprint("device", string(details)+submission.UserAgent)
	}
	if err := AddWordToWordCloud(ctx, wordCloud.Id, &word); err != nil {
		if errors.Is(err, ErrWordNotAdded) {
			return "", utils.Conflict("word already submitted")
		}
		return "", err
	}

	var publicEvent any
	if (wordCloud.Settings.ShowCloudToAudience || wordCloud.Settings.AllowUpvotes) && !word.Hidden {
		publicEvent = map[string]any{"type": "word", "word": map[string]any{"text": word.Text}}
	}
	go broadcast(wordCloud.Id, map[string]any{"type": "word", "word": withSentiment(wordCloud, []Word{word})[0]}, publicEvent)

	return text, nil
}

func CreateSession(ctx context.Context, userId primitive.ObjectID, body PostWordCloudBody) (primitive.ObjectID, error) {
	settings := Settings{FilterProfanity: true, AllowUpvotes: true, Sentiment: true}
	if body.Settings != nil {
		settings = *body.Settings
	}
	return createSession(ctx, userId, body.Name, body.Description, settings)
}

// DuplicateSession starts an empty session with the same name, description and settings
func DuplicateSession(ctx context.Context, wordCloud *WordCloud) (primitive.ObjectID, error) {
	return createSession(ctx, wordCloud.UserId, wordCloud.Name, wordCloud.Description, wordCloud.Settings)
}

func createSession(ctx context.Context, userId primitive.ObjectID, name string, description string, settings Settings) (primitive.ObjectID, error) {
	code, err := newUniqueCode(ctx)
	if err != nil {
		return primitive.NilObjectID, err
	}
	now := primitive.NewDateTimeFromTime(time.Now())
	result, err := CreateWordCloud(ctx, &WordCloud{
		UserId:      userId,
		Name:        name,
		Description: description,
		Code:        code,
		Words:       []Word{},
		Settings:    settings,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		return primitive.NilObjectID, err
	}
	return result.InsertedID.(primitive.ObjectID), nil
}

// ChangeSession applies the given fields. Reopening a session gives it a new code when
// another open session took its code meanwhile.
func ChangeSession(ctx context.Context, wordCloud *WordCloud, body PatchWordCloudBody) (*WordCloud, error) {
	if body.Description != nil && *body.Description != "" && len(*body.Description) < 10 {
		return nil, utils.Invalid("description must be empty or at least 10 characters long")
	}

	set, unset := bson.M{}, bson.M{}
	if body.Name != nil {
		wordCloud.Name = *body.Name
		set["name"] = wordCloud.Name
	}
	if body.Description != nil {
		wordCloud.Description = *body.Description
		set["description"] = wordCloud.Description
	}
	if body.Settings != nil {
		wordCloud.Settings = *body.Settings
		set["settings"] = wordCloud.Settings
	}
	if body.Open != nil && *body.Open != isOpen(wordCloud) {
		if *body.Open {
			inUse, err := IsCodeInUse(ctx, wordCloud.Code)
			if err == nil && inUse {
				wordCloud.Code, err = newUniqueCode(ctx)
			}
			if err != nil {
				return nil, err
			}
			set["code"] = wordCloud.Code
			unset["closedAt"] = ""
			wordCloud.ClosedAt = nil
		} else {
			now := primitive.NewDateTimeFromTime(time.Now())
			set["closedAt"] = now
			wordCloud.ClosedAt = &now
		}
	}

	if err := UpdateWordCloud(ctx, wordCloud.Id, set, unset); err != nil {
		return nil, err
	}
	go broadcastSession(wordCloud)
	return wordCloud, nil
}

func RemoveSession(ctx context.Context, wordCloud *WordCloud) error {
	if err := DeleteWordCloud(ctx, wordCloud.Id); err != nil {
		return err
	}
	event := map[string]any{"type": "session", "session": nil}
	go broadcast(wordCloud.Id, event, event)
	return nil
}

func HideWord(ctx context.Context, wordCloud *WordCloud, wordId primitive.ObjectID, hidden bool) ([]Word, error) {
	found, err := SetWordHidden(ctx, wordCloud.Id, wordId, hidden)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, utils.NotFound("word not found")
	}
	return moderatedWords(ctx, wordCloud.Id)
}

// MergeInto rewrites the given words into one, cleaned like a submission
func MergeInto(ctx context.Context, wordCloud *WordCloud, from []string, into string) ([]Word, error) {
	into = CleanText(into)
	if into == "" {
		return nil, utils.Invalid("word is empty")
	}
	if err := MergeWords(ctx, wordCloud.Id, from, into); err != nil {
		return nil, err
	}
	return moderatedWords(ctx, wordCloud.Id)
}

// moderatedWords returns the session's words after a moderation change and tells every viewer
func moderatedWords(ctx context.Context, sessionId primitive.ObjectID) ([]Word, error) {
	wordCloud, err := GetWordCloudById(ctx, sessionId)
	if err != nil {
		return nil, err
	}
	go broadcastSession(wordCloud)
	return withSentiment(wordCloud, wordCloud.Words), nil
}

func subscribe(sessionId primitive.ObjectID, conn *websocket.Conn, owner bool) *socketClient {
	client := &socketClient{conn: conn, owner: owner}
	socketMutex.Lock()
	socketClients[sessionId] = append(socketClients[sessionId], client)
	socketMutex.Unlock()
	return client
}

func unsubscribe(sessionId primitive.ObjectID, client *socketClient) {
	socketMutex.Lock()
	defer socketMutex.Unlock()
	clients := socketClients[sessionId]
	for i, registered := range clients {
		if registered == client {
			socketClients[sessionId] = append(clients[:i], clients[i+1:]...)
			break
		}
	}
	if len(socketClients[sessionId]) == 0 {
		delete(socketClients, sessionId)
	}
}
