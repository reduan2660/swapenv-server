package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/reduan2660/swapenv-server/internal/session"
	"time"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type ShareHandler struct {
	Sessions *session.Manager
}

// WS /share
func (h *ShareHandler) Share(c echo.Context) error {
	orgID := c.Get("org_id").(uuid.UUID)

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	sess := h.Sessions.Create(orgID, conn)
	defer h.Sessions.Delete(sess.Code)

	// send session code to sharer
	conn.WriteJSON(map[string]string{"type": "waiting", "code": sess.Code})

	for {
		// wait for receiver to join
		for sess.ReceiverConn == nil {
			time.Sleep(100 * time.Millisecond)
		}

		conn.WriteJSON(map[string]string{"type": "ready"})

		// relay: receiver -> sharer (public key)
		_, pubKey, err := sess.ReceiverConn.ReadMessage()
		if err != nil {
			sess.ReceiverConn = nil
			conn.WriteJSON(map[string]string{"type": "waiting", "code": sess.Code})
			continue
		}
		conn.WriteMessage(websocket.TextMessage, pubKey)

		// relay: sharer -> receiver (encrypted payload)
		_, payload, err := conn.ReadMessage()
		if err != nil {
			return nil
		}
		sess.ReceiverConn.WriteMessage(websocket.TextMessage, payload)

		// signal receiver that transfer is complete
		close(sess.Done)

		// reset for next receiver
		sess.ReceiverConn = nil
		sess.Done = make(chan struct{})
		conn.WriteJSON(map[string]string{"type": "waiting", "code": sess.Code})
	}
}

// WS /receive
func (h *ShareHandler) Receive(c echo.Context) error {
	orgID := c.Get("org_id").(uuid.UUID)

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	sessions := h.Sessions.GetByOrg(orgID)

	if len(sessions) == 0 {
		conn.WriteJSON(map[string]string{"type": "error", "message": "no active share sessions"})
		return nil
	}

	var sess *session.Session
	if len(sessions) == 1 {
		sess = sessions[0]
		conn.WriteJSON(map[string]string{"type": "connected", "code": sess.Code})
	} else {
		codes := make([]string, len(sessions))
		for i, s := range sessions {
			codes[i] = s.Code
		}
		conn.WriteJSON(map[string]any{"type": "choose", "codes": codes})

		// wait for choice
		var choice struct {
			Code string `json:"code"`
		}
		if err := conn.ReadJSON(&choice); err != nil {
			return nil
		}

		sess = h.Sessions.Get(choice.Code)
		if sess == nil || sess.OrgID != orgID {
			conn.WriteJSON(map[string]string{"type": "error", "message": "invalid session"})
			return nil
		}
		conn.WriteJSON(map[string]string{"type": "connected", "code": sess.Code})
	}
	sess.ReceiverConn = conn

	// Share handler takes over from here
	// wait until transfer is complete
	<-sess.Done

	return nil

}
