package slides

import (
	"context"
	"dev/internal/user"
	"dev/internal/utils"
	"net/http"
	"net/url"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	authTimeout   = 5 * time.Second
	pingInterval  = 30 * time.Second
	readTimeout   = 70 * time.Second
	writeTimeout  = 10 * time.Second
	maxMessage    = 4 << 20
	outboxSize    = 256
	closeNotFound = 4404
	closeAuth     = 4401
	closeDeleted  = 4410
	closeRemoved  = 4403
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" || slices.Contains(utils.AllowedOrigins, origin) {
			return true
		}
		parsed, err := url.Parse(origin)
		return err == nil && parsed.Host == r.Host
	},
}

// closing ends a connection once the messages queued before it are sent
type closing struct {
	code   int
	reason string
}

type client struct {
	conn   *websocket.Conn
	outbox chan any
	userId primitive.ObjectID
	ctx    context.Context
	once   sync.Once
	peer   int
	name   string
	role   string
}

// send never blocks the hub: a client that cannot keep up is disconnected
func (c *client) send(message any) {
	select {
	case c.outbox <- message:
	default:
		c.close()
	}
}

func (c *client) close() {
	c.once.Do(func() { c.conn.Close() })
}

// stop is called by the hub once it forgot the client, so nothing is sent after it
func (c *client) stop() {
	close(c.outbox)
}

func (c *client) write() {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()
	for {
		select {
		case message, open := <-c.outbox:
			if !open {
				c.close()
				return
			}
			if end, ok := message.(closing); ok {
				c.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(end.code, end.reason), time.Now().Add(writeTimeout))
				c.close()
				continue
			}
			c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := c.conn.WriteJSON(message); err != nil {
				c.close()
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.close()
			}
		}
	}
}

func refuse(conn *websocket.Conn, code int, reason string) {
	conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, reason), time.Now().Add(writeTimeout))
	conn.Close()
}

// WSDeck godoc
//
//	@Summary		Edit a slide deck live
//	@Description	WebSocket for the owner and the members of a deck. The first message is {"type": "auth", "token": accessToken, "version": n}, n being the version the client has. The server answers {"type": "state", "title", "settings", "peer", "role"} then the updates since n, or {"type": "reset", "version", "doc"} when they are not kept.
//	@Description	Clients send {"type": "push", "version", "updates": [{"clientID", "changes"}]}, {"type": "pull", "version"}, {"type": "title", "title"}, {"type": "settings", "settings"} and {"type": "presence", "version", "ranges": [{"anchor", "head"}]}.
//	@Description	The server sends {"type": "updates", "version", "updates"} to every client, {"type": "pushed", "ok"} to the sender of a push, {"type": "title"}, {"type": "settings"}, {"type": "presence", "peer", "name", "version", "ranges"}, {"type": "joined", "peer", "name"}, {"type": "left", "peer"}, {"type": "role", "role"} when the role of the client changes, {"type": "deleted"} and {"type": "error", "error"}. A refused push is answered with {"type": "reset", "version", "doc", "error"}. Viewers can only pull and send their presence.
//	@Tags			slides
//	@Param			id	path	string	true	"Deck ID"
//	@Success		101
//	@Failure		400	{object}	map[string]string
//	@Router			/slides/{id}/ws [get]
func (controller *Controller) WSDeck(c *gin.Context) {
	id, ok := deckId(c)
	if !ok {
		return
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	conn.SetReadLimit(maxMessage)

	var auth clientMessage
	conn.SetReadDeadline(time.Now().Add(authTimeout))
	if err := conn.ReadJSON(&auth); err != nil || auth.Type != "auth" {
		refuse(conn, closeAuth, "authentication required")
		return
	}
	secret := os.Getenv("ACCESS_TOKEN_SECRET")
	authorized, err := utils.IsAuthorized(auth.Token, secret)
	if err != nil || !authorized {
		refuse(conn, closeAuth, "invalid token")
		return
	}
	userId, err := utils.ExtractID(auth.Token, secret)
	if err != nil {
		refuse(conn, closeAuth, "invalid token")
		return
	}
	if _, _, err := deckAccess(c, userId, id); err != nil {
		refuse(conn, closeNotFound, "deck not found")
		return
	}

	c.Set("x-user-id", userId)
	member := &client{conn: conn, outbox: make(chan any, outboxSize), userId: userId, ctx: c.Copy(), name: "Someone"}
	if account, err := user.GetFullUserById(c, userId); err == nil {
		member.name = account.WebAuthnDisplayName()
	}
	go member.write()

	h := hubs.join(id)
	h.events <- event{kind: joined, client: member, message: auth}
	defer hubs.leave(h, member)

	conn.SetReadDeadline(time.Now().Add(readTimeout))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(readTimeout))
	})
	for {
		var incoming clientMessage
		if err := conn.ReadJSON(&incoming); err != nil {
			return
		}
		conn.SetReadDeadline(time.Now().Add(readTimeout))
		h.events <- event{kind: received, client: member, message: incoming}
	}
}
