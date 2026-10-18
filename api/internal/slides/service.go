package slides

import (
	"context"
	"dev/internal/audit"
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
	decks, total, err := GetDecksByUserId(ctx, userId, query.Search, query.Page, query.Limit)
	if err != nil {
		return nil, err
	}
	return &DeckPage{Items: decks, Total: total, Page: query.Page, Limit: query.Limit}, nil
}

// GetDeck answers not found for the decks of other users too
func GetDeck(ctx context.Context, userId primitive.ObjectID, id primitive.ObjectID) (*Deck, error) {
	deck, err := GetDeckByUserIdAndId(ctx, userId, id)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, utils.NotFound("Deck not found")
	}
	return deck, err
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
	original, err := GetDeck(ctx, userId, id)
	if err != nil {
		return nil, err
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
