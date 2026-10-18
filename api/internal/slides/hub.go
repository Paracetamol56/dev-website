package slides

import (
	"context"
	"dev/internal/audit"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin/binding"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	maxContentLength = 500_000
	keptUpdates      = 500
	saveDelay        = 2 * time.Second
	retryDelay       = 10 * time.Second
)

// Range is a selection range, in positions of the document at the version sent with it
type Range struct {
	Anchor int `json:"anchor"`
	Head   int `json:"head"`
}

const maxRanges = 100

// clientMessage is anything a client sends; Type tells which fields are set
type clientMessage struct {
	Type     string    `json:"type"`
	Token    string    `json:"token"`
	Version  int64     `json:"version"`
	Updates  []Update  `json:"updates"`
	Title    string    `json:"title"`
	Settings *Settings `json:"settings"`
	Ranges   []Range   `json:"ranges"`
}

type eventKind int

const (
	joined eventKind = iota
	left
	received
	flush
	deleted
	persist
	accessChanged
)

type event struct {
	kind    eventKind
	client  *client
	message clientMessage
	done    chan struct{}
	members []Member
}

// hub holds the document of an open deck. Only its goroutine touches its fields, so every
// change is applied in the order it arrived.
type hub struct {
	id      primitive.ObjectID
	events  chan event
	members int
	closed  chan struct{}

	deck    *Deck
	failed  bool
	doc     []uint16
	version int64
	base    int64
	updates []Update
	clients map[*client]bool
	dirty   bool
	timer   *time.Timer
	removed bool
	peers   int
}

type registry struct {
	mutex   sync.Mutex
	hubs    map[primitive.ObjectID]*hub
	closing map[primitive.ObjectID]chan struct{}
}

var hubs = &registry{hubs: map[primitive.ObjectID]*hub{}, closing: map[primitive.ObjectID]chan struct{}{}}

// join returns the hub of a deck, waiting for a closing one to save first
func (registry *registry) join(id primitive.ObjectID) *hub {
	for {
		registry.mutex.Lock()
		if existing := registry.hubs[id]; existing != nil {
			existing.members++
			registry.mutex.Unlock()
			return existing
		}
		if wait := registry.closing[id]; wait != nil {
			registry.mutex.Unlock()
			<-wait
			continue
		}
		created := &hub{id: id, events: make(chan event, 64), members: 1, clients: map[*client]bool{}}
		registry.hubs[id] = created
		registry.mutex.Unlock()
		go created.run()
		return created
	}
}

func (registry *registry) leave(h *hub, c *client) {
	registry.mutex.Lock()
	h.members--
	if h.members == 0 {
		delete(registry.hubs, h.id)
		h.closed = make(chan struct{})
		registry.closing[h.id] = h.closed
	}
	registry.mutex.Unlock()
	h.events <- event{kind: left, client: c}
}

func (registry *registry) finished(h *hub) {
	registry.mutex.Lock()
	delete(registry.closing, h.id)
	registry.mutex.Unlock()
	close(h.closed)
}

// remove disconnects the clients of a deck that was just deleted
func (registry *registry) remove(id primitive.ObjectID) {
	registry.mutex.Lock()
	h := registry.hubs[id]
	registry.mutex.Unlock()
	if h != nil {
		h.events <- event{kind: deleted}
	}
}

// access gives an open deck its new members, so it updates the role of its clients
func (registry *registry) access(id primitive.ObjectID, members []Member) {
	registry.mutex.Lock()
	h := registry.hubs[id]
	registry.mutex.Unlock()
	if h != nil {
		h.events <- event{kind: accessChanged, members: members}
	}
}

// SaveAll saves the open decks, for a server that is about to stop
func SaveAll(ctx context.Context) {
	hubs.mutex.Lock()
	open := make([]*hub, 0, len(hubs.hubs))
	for _, h := range hubs.hubs {
		open = append(open, h)
	}
	hubs.mutex.Unlock()

	for _, h := range open {
		done := make(chan struct{})
		h.events <- event{kind: persist, done: done}
		select {
		case <-done:
		case <-time.After(5 * time.Second):
		case <-ctx.Done():
			return
		}
	}
}

func (h *hub) run() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	deck, err := GetDeckById(ctx, h.id)
	cancel()
	if err != nil {
		h.failed = true
	} else {
		h.deck = deck
		h.doc = encodeText(deck.Content)
		h.version = deck.Revision
		h.base = deck.Revision
	}

	for e := range h.events {
		switch e.kind {
		case joined:
			h.join(e.client, e.message.Version)
		case received:
			if h.clients[e.client] {
				h.receive(e.client, e.message)
			}
		case flush:
			h.save()
		case persist:
			h.save()
			close(e.done)
		case accessChanged:
			h.updateAccess(e.members)
		case deleted:
			h.removed = true
			h.broadcast(message("deleted", nil))
			for c := range h.clients {
				c.send(closing{closeDeleted, "deck deleted"})
			}
		case left:
			if h.clients[e.client] {
				delete(h.clients, e.client)
				h.broadcast(message("left", map[string]any{"peer": e.client.peer}))
			}
			e.client.stop()
			if len(h.clients) == 0 && h.membersGone() {
				h.save()
				if h.timer != nil {
					h.timer.Stop()
				}
				hubs.finished(h)
				return
			}
		}
	}
}

func (h *hub) membersGone() bool {
	hubs.mutex.Lock()
	defer hubs.mutex.Unlock()
	return h.members == 0 && h.closed != nil
}

func message(kind string, fields map[string]any) map[string]any {
	result := map[string]any{"type": kind}
	for key, value := range fields {
		result[key] = value
	}
	return result
}

func (h *hub) broadcast(event map[string]any) {
	for c := range h.clients {
		c.send(event)
	}
}

func (h *hub) join(c *client, version int64) {
	if !h.failed {
		c.role = access(h.deck, c.userId)
	}
	if h.failed || h.removed || c.role == "" {
		c.send(message("error", map[string]any{"error": "Deck not found"}))
		c.send(closing{closeNotFound, "deck not found"})
		return
	}
	h.peers++
	c.peer = h.peers
	h.broadcast(message("joined", map[string]any{"peer": c.peer, "name": c.name}))
	h.clients[c] = true
	c.send(message("state", map[string]any{"title": h.deck.Title, "settings": h.deck.Settings, "peer": c.peer, "role": c.role}))
	h.catchUp(c, version)
}

// catchUp sends the updates a client misses, or the whole document when they are not kept
func (h *hub) catchUp(c *client, version int64) {
	if version < h.base || version > h.version {
		c.send(message("reset", map[string]any{"version": h.version, "doc": decodeText(h.doc)}))
		return
	}
	missing := append([]Update{}, h.updates[version-h.base:]...)
	c.send(message("updates", map[string]any{"version": version, "updates": missing}))
}

// updateAccess disconnects the clients who are not members anymore, and tells the others their role
func (h *hub) updateAccess(members []Member) {
	if h.failed {
		return
	}
	h.deck.Members = members
	for c := range h.clients {
		role := access(h.deck, c.userId)
		switch {
		case role == "":
			delete(h.clients, c)
			h.broadcast(message("left", map[string]any{"peer": c.peer}))
			c.send(closing{closeRemoved, "access removed"})
		case role != c.role:
			c.role = role
			c.send(message("role", map[string]any{"role": role}))
		}
	}
}

func (h *hub) receive(c *client, incoming clientMessage) {
	if c.role == RoleViewer {
		switch incoming.Type {
		case "push":
			h.refuse(c, "You can only view this deck, your last changes were not kept")
			return
		case "title", "settings":
			c.send(message("error", map[string]any{"error": "You can only view this deck"}))
			return
		}
	}
	switch incoming.Type {
	case "push":
		h.push(c, incoming.Version, incoming.Updates)
	case "pull":
		h.catchUp(c, incoming.Version)
	case "title":
		h.rename(c, incoming.Title)
	case "settings":
		h.configure(c, incoming.Settings)
	case "presence":
		h.relayPresence(c, incoming.Version, incoming.Ranges)
	default:
		c.send(message("error", map[string]any{"error": "Unknown message"}))
	}
}

func (h *hub) push(c *client, version int64, updates []Update) {
	if version != h.version || len(updates) == 0 {
		c.send(message("pushed", map[string]any{"ok": false}))
		return
	}
	doc := h.doc
	for _, update := range updates {
		var err error
		if update.ClientId == "" || len(update.ClientId) > 64 {
			err = ErrInvalidChanges
		} else {
			doc, err = applyChanges(doc, update.Changes)
		}
		if err != nil {
			h.refuse(c, "These changes do not match the document, they were not kept")
			return
		}
	}
	if len(doc) > maxContentLength {
		h.refuse(c, "The deck is too long, your last changes were not kept")
		return
	}

	h.broadcast(message("updates", map[string]any{"version": h.version, "updates": updates}))
	h.doc = doc
	h.version += int64(len(updates))
	h.updates = append(h.updates, updates...)
	if excess := len(h.updates) - keptUpdates; excess > 0 {
		h.updates = append([]Update(nil), h.updates[excess:]...)
		h.base += int64(excess)
	}
	c.send(message("pushed", map[string]any{"ok": true}))
	h.schedule(saveDelay)
}

// refuse sends the server's document back: pushing the same changes again would fail again
func (h *hub) refuse(c *client, reason string) {
	c.send(message("reset", map[string]any{"version": h.version, "doc": decodeText(h.doc), "error": reason}))
}

// relayPresence passes a selection on to the other clients; it is never stored
func (h *hub) relayPresence(c *client, version int64, ranges []Range) {
	if len(ranges) == 0 || len(ranges) > maxRanges || version < 0 {
		c.send(message("error", map[string]any{"error": "Invalid selection"}))
		return
	}
	for _, r := range ranges {
		if r.Anchor < 0 || r.Head < 0 || r.Anchor > maxContentLength || r.Head > maxContentLength {
			c.send(message("error", map[string]any{"error": "Invalid selection"}))
			return
		}
	}
	presence := message("presence", map[string]any{"peer": c.peer, "name": c.name, "version": version, "ranges": ranges})
	for other := range h.clients {
		if other != c {
			other.send(presence)
		}
	}
}

func (h *hub) rename(c *client, title string) {
	title = strings.TrimSpace(title)
	if title == "" || len([]rune(title)) > 100 {
		c.send(message("error", map[string]any{"error": "The title must be 1 to 100 characters long"}))
		return
	}
	if title == h.deck.Title {
		return
	}
	if err := h.store(bson.M{"title": title}); err != nil {
		c.send(message("error", map[string]any{"error": "The deck could not be renamed"}))
		return
	}
	audit.Record(c.ctx, audit.Entry{Action: audit.SlidesRenamed, UserId: &c.userId, Target: target(h.id), Meta: audit.Meta{"from": h.deck.Title, "to": title}})
	h.deck.Title = title
	h.broadcast(message("title", map[string]any{"title": title}))
}

func (h *hub) configure(c *client, settings *Settings) {
	if settings == nil || binding.Validator.ValidateStruct(settings) != nil {
		c.send(message("error", map[string]any{"error": "These settings are not valid"}))
		return
	}
	if err := h.store(bson.M{"settings": *settings}); err != nil {
		c.send(message("error", map[string]any{"error": "The settings could not be saved"}))
		return
	}
	h.deck.Settings = *settings
	h.broadcast(message("settings", map[string]any{"settings": *settings}))
}

func (h *hub) store(set bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	set["updatedAt"] = primitive.NewDateTimeFromTime(time.Now())
	return SetDeckFields(ctx, h.id, set)
}

func (h *hub) schedule(delay time.Duration) {
	h.dirty = true
	if h.timer != nil {
		h.timer.Stop()
	}
	events := h.events
	h.timer = time.AfterFunc(delay, func() { events <- event{kind: flush} })
}

func (h *hub) save() {
	if !h.dirty || h.removed {
		return
	}
	content := decodeText(h.doc)
	if err := h.store(bson.M{"content": content, "outline": BuildOutline(content), "revision": h.version}); err != nil {
		log.Printf("Failed to save slide deck %s: %v", h.id.Hex(), err)
		h.schedule(retryDelay)
		return
	}
	h.dirty = false
}
