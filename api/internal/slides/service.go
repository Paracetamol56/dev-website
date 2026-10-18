package slides

import (
	"context"
	"dev/internal/audit"
	"dev/internal/user"
	"dev/internal/utils"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func target(id primitive.ObjectID) *audit.Target {
	return &audit.Target{Type: audit.TargetSlides, Id: id.Hex()}
}

func ListDecks(ctx context.Context, userId primitive.ObjectID, query DecksQuery) (*DeckPage, error) {
	if query.Page == 0 {
		query.Page = 1
	}
	if query.Limit == 0 {
		query.Limit = DefaultLimit
	}
	decks, total, err := GetDecks(ctx, userId, query.Search, query.Page, query.Limit)
	if err != nil {
		return nil, err
	}
	owners := []primitive.ObjectID{}
	for i := range decks {
		decks[i].Role = roleIn(decks[i].UserId, decks[i].Members, userId)
		if decks[i].Role != RoleOwner {
			owners = append(owners, decks[i].UserId)
		}
	}
	names := displayNames(ctx, owners)
	for i := range decks {
		if decks[i].Role != RoleOwner {
			decks[i].OwnerName = names[decks[i].UserId]
		}
	}
	return &DeckPage{Items: decks, Total: total, Page: query.Page, Limit: query.Limit}, nil
}

func roleIn(owner primitive.ObjectID, members []Member, userId primitive.ObjectID) string {
	if owner == userId {
		return RoleOwner
	}
	for _, member := range members {
		if member.UserId == userId {
			return member.Role
		}
	}
	return ""
}

func access(deck *Deck, userId primitive.ObjectID) string {
	return roleIn(deck.UserId, deck.Members, userId)
}

func displayNames(ctx context.Context, ids []primitive.ObjectID) map[primitive.ObjectID]string {
	names := map[primitive.ObjectID]string{}
	if len(ids) == 0 {
		return names
	}
	users, err := user.GetUsersByIds(ctx, ids)
	if err != nil {
		return names
	}
	for _, account := range users {
		names[account.Id] = account.WebAuthnDisplayName()
	}
	return names
}

// deckAccess gives a deck with the role of the user in it, and answers not found to anyone else
func deckAccess(ctx context.Context, userId primitive.ObjectID, id primitive.ObjectID) (*Deck, string, error) {
	deck, err := GetDeckById(ctx, id)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, "", utils.NotFound("Deck not found")
	}
	if err != nil {
		return nil, "", err
	}
	role := access(deck, userId)
	if role == "" {
		return nil, "", utils.NotFound("Deck not found")
	}
	return deck, role, nil
}

func ownedDeck(ctx context.Context, userId primitive.ObjectID, id primitive.ObjectID) (*Deck, error) {
	deck, role, err := deckAccess(ctx, userId, id)
	if err != nil {
		return nil, err
	}
	if role != RoleOwner {
		return nil, utils.Forbidden("Only the owner of the deck can do this")
	}
	return deck, nil
}

func GetDeck(ctx context.Context, userId primitive.ObjectID, id primitive.ObjectID) (*DeckView, error) {
	deck, role, err := deckAccess(ctx, userId, id)
	if err != nil {
		return nil, err
	}
	return viewOf(ctx, deck, role), nil
}

func viewOf(ctx context.Context, deck *Deck, role string) *DeckView {
	view := &DeckView{Deck: deck, Role: role}
	if role != RoleOwner {
		return view
	}
	view.Links = &deck.Links
	view.Members = []MemberView{}
	ids := []primitive.ObjectID{}
	for _, member := range deck.Members {
		ids = append(ids, member.UserId)
	}
	names := displayNames(ctx, ids)
	for _, member := range deck.Members {
		name, ok := names[member.UserId]
		if !ok {
			name = "Deleted account"
		}
		view.Members = append(view.Members, MemberView{Member: member, Name: name})
	}
	return view
}

func insertDeck(ctx context.Context, userId primitive.ObjectID, title string, content string, settings Settings) (*Deck, error) {
	now := primitive.NewDateTimeFromTime(time.Now())
	deck := &Deck{
		UserId:    userId,
		Title:     title,
		Content:   content,
		Outline:   BuildOutline(content),
		Settings:  settings,
		CreatedAt: now,
		UpdatedAt: now,
	}
	result, err := CreateDeck(ctx, deck)
	if err != nil {
		return nil, err
	}
	deck.Id = result.InsertedID.(primitive.ObjectID)
	return deck, nil
}

func AddDeck(ctx context.Context, userId primitive.ObjectID, body PostDeckBody) (*Deck, error) {
	content := defaultContent
	if body.Content != nil {
		content = *body.Content
	}
	deck, err := insertDeck(ctx, userId, body.Title, content, DefaultSettings())
	if err != nil {
		return nil, err
	}
	audit.Record(ctx, audit.Entry{Action: audit.SlidesCreated, UserId: &userId, Target: target(deck.Id), Meta: audit.Meta{"title": deck.Title}})
	return deck, nil
}

func DuplicateDeck(ctx context.Context, userId primitive.ObjectID, id primitive.ObjectID) (*Deck, error) {
	original, role, err := deckAccess(ctx, userId, id)
	if err != nil {
		return nil, err
	}
	if role == RoleViewer {
		return nil, utils.Forbidden("Viewers cannot duplicate this deck")
	}
	title := original.Title + " (copy)"
	if len([]rune(title)) > 100 {
		title = original.Title
	}
	deck, err := insertDeck(ctx, userId, title, original.Content, original.Settings)
	if err != nil {
		return nil, err
	}
	audit.Record(ctx, audit.Entry{Action: audit.SlidesDuplicated, UserId: &userId, Target: target(deck.Id), Meta: audit.Meta{"from": id.Hex()}})
	return deck, nil
}

func RemoveDeck(ctx context.Context, userId primitive.ObjectID, id primitive.ObjectID) error {
	if _, err := ownedDeck(ctx, userId, id); err != nil {
		return err
	}
	result, err := DeleteDeck(ctx, userId, id)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return utils.NotFound("Deck not found")
	}
	audit.Record(ctx, audit.Entry{Action: audit.SlidesDeleted, UserId: &userId, Target: target(id)})
	hubs.remove(id)
	return nil
}

// BackfillOutlines gives an outline to the decks saved before decks had one
func BackfillOutlines(ctx context.Context) (int, error) {
	decks, err := GetDecksWithoutOutline(ctx)
	if err != nil {
		return 0, err
	}
	for _, deck := range decks {
		if err := SetDeckOutline(ctx, deck.Id, BuildOutline(deck.Content)); err != nil {
			return 0, err
		}
	}
	return len(decks), nil
}

func BackfillSettings(ctx context.Context) (int64, error) {
	return SetMissingSettings(ctx, DefaultSettings())
}
