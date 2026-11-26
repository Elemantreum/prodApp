package schedule

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/websocket"

	http2 "prodApp/internal/platform/http"
)

type TokenVerifier interface {
	Verify(token string) (*http2.Claims, error)
}

type ScheduleHandler struct {
	service  *ScheduleService
	verifier TokenVerifier
}

func NewScheduleHandler(service *ScheduleService, verifier TokenVerifier) *ScheduleHandler {
	return &ScheduleHandler{service: service, verifier: verifier}
}

func (h *ScheduleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server := websocket.Server{
		Handshake: func(cfg *websocket.Config, req *http.Request) error {
			origin, err := url.Parse(req.Header.Get("Origin"))
			if err != nil || origin.Host != req.Host || (origin.Scheme != "https" && origin.Scheme != "http") {
				return errors.New("invalid origin")
			}
			return nil
		},
		Handler: func(conn *websocket.Conn) {
			conn.MaxPayloadBytes = 4096
			_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			var token string
			if websocket.Message.Receive(conn, &token) != nil {
				return
			}
			claims, err := h.verifier.Verify(token)
			if err != nil || claims.ExpiresAt == nil {
				return
			}
			changes, unsubscribe := h.service.Subscribe()
			defer unsubscribe()
			expiry := time.NewTimer(time.Until(claims.ExpiresAt.Time))
			defer expiry.Stop()
			heartbeat := time.NewTicker(25 * time.Second)
			defer heartbeat.Stop()
			for {
				message := "changed"
				select {
				case <-expiry.C:
					return
				case <-heartbeat.C:
					message = "ping"
				case <-changes:
				}
				_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
				if websocket.Message.Send(conn, message) != nil {
					return
				}
			}
		},
	}
	server.ServeHTTP(w, r)
}
