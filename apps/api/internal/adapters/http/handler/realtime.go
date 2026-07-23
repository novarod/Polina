package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	httpmw "github.com/novarod/polina/apps/api/internal/adapters/http/middleware"
	"github.com/novarod/polina/apps/api/internal/application/authz"
	"github.com/novarod/polina/apps/api/internal/application/realtime"
	"github.com/novarod/polina/apps/api/internal/application/token"
	"github.com/novarod/polina/apps/api/internal/domain/member"
	"github.com/novarod/polina/apps/api/internal/ports"
)

const (
	realtimeAuthWindow = 5 * time.Second
	realtimeOpTimeout  = 5 * time.Second
	realtimeWriteWait  = 5 * time.Second
	realtimePingEvery  = 10 * time.Second
)

type RealtimeHandler struct {
	hub            *realtime.Hub
	users          ports.UserRepository
	members        ports.MemberRepository
	missions       ports.MissionRepository
	jwtSecret      string
	originPatterns []string
}

func NewRealtimeHandler(hub *realtime.Hub, users ports.UserRepository, members ports.MemberRepository, missions ports.MissionRepository, jwtSecret, frontendURL string) *RealtimeHandler {
	h := &RealtimeHandler{hub: hub, users: users, members: members, missions: missions, jwtSecret: jwtSecret}
	if u, err := url.Parse(frontendURL); err == nil && u.Host != "" {
		h.originPatterns = []string{u.Host}
	}
	return h
}

type realtimeTicketResponse struct {
	Ticket string `json:"ticket"`
}

// @Summary  Issue a short-lived ticket for realtime WebSocket auth
// @Tags     realtime
// @Security BearerAuth
// @Produce  json
// @Success  200  {object}  realtimeTicketResponse
// @Failure  401  {object}  map[string]string
// @Router   /realtime/ticket [get]
func (h *RealtimeHandler) Ticket(c echo.Context) error {
	claims := httpmw.MustGetClaims(c)
	ticket, err := token.NewRealtimeTicket(h.jwtSecret, claims.UserID, time.Now())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}
	return c.JSON(http.StatusOK, realtimeTicketResponse{Ticket: ticket})
}

// @Summary      Realtime presence WebSocket (cursors, avatars, editing status)
// @Description  Upgrades to WebSocket. Authenticates via session cookie or a first `{"type":"auth","ticket":...}` frame.
// @Tags         realtime
// @Success      101  {string}  string  "Switching Protocols"
// @Failure      401  {object}  map[string]string
// @Router       /realtime/ws [get]
func (h *RealtimeHandler) Connect(c echo.Context) error {
	var userID uuid.UUID
	authed := false
	if claims, ok := httpmw.GetClaims(c); ok {
		userID = claims.UserID
		authed = true
	}

	rc := http.NewResponseController(c.Response())
	_ = rc.SetReadDeadline(time.Time{})
	_ = rc.SetWriteDeadline(time.Time{})

	ws, err := websocket.Accept(c.Response(), c.Request(), &websocket.AcceptOptions{OriginPatterns: h.originPatterns})
	if err != nil {
		return nil
	}
	defer func() { _ = ws.CloseNow() }()

	if !authed {
		id, ok := h.awaitTicket(ws)
		if !ok {
			_ = ws.Close(websocket.StatusPolicyViolation, "authentication required")
			return nil
		}
		userID = id
	}

	fctx, fcancel := context.WithTimeout(context.Background(), realtimeOpTimeout)
	user, err := h.users.FindByID(fctx, userID)
	fcancel()
	if err != nil {
		_ = ws.Close(websocket.StatusPolicyViolation, "unknown user")
		return nil
	}

	conn, err := h.hub.Register(user.ID, user.Name)
	if err != nil {
		_ = ws.Close(websocket.StatusGoingAway, "shutting down")
		return nil
	}
	defer h.hub.RemoveAll(conn)

	connCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hctx, hcancel := context.WithTimeout(connCtx, realtimeWriteWait)
	err = ws.Write(hctx, websocket.MessageText, realtime.MarshalHello(realtime.User{ID: user.ID, Name: user.Name}))
	hcancel()
	if err != nil {
		return nil
	}

	go h.writePump(connCtx, cancel, ws, conn)

	for {
		_, data, err := ws.Read(connCtx)
		if err != nil {
			return nil
		}
		h.dispatch(connCtx, conn, data)
	}
}

func (h *RealtimeHandler) awaitTicket(ws *websocket.Conn) (uuid.UUID, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), realtimeAuthWindow)
	defer cancel()
	_, data, err := ws.Read(ctx)
	if err != nil {
		return uuid.Nil, false
	}
	var frame realtime.ClientFrame
	if err := json.Unmarshal(data, &frame); err != nil || frame.Type != realtime.TypeAuth {
		return uuid.Nil, false
	}
	userID, err := token.ParseRealtimeTicket(h.jwtSecret, frame.Ticket)
	if err != nil {
		return uuid.Nil, false
	}
	return userID, true
}

func (h *RealtimeHandler) writePump(ctx context.Context, cancel context.CancelFunc, ws *websocket.Conn, conn *realtime.Conn) {
	defer cancel()
	ticker := time.NewTicker(realtimePingEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-conn.Outbox():
			wctx, wcancel := context.WithTimeout(ctx, realtimeWriteWait)
			err := ws.Write(wctx, websocket.MessageText, msg)
			wcancel()
			if err != nil {
				conn.Kick(realtime.CloseTryAgainLater)
				_ = ws.CloseNow()
				return
			}
		case <-ticker.C:
			pctx, pcancel := context.WithTimeout(ctx, realtimeWriteWait)
			err := ws.Ping(pctx)
			pcancel()
			if err != nil {
				conn.Kick(realtime.CloseTryAgainLater)
				_ = ws.CloseNow()
				return
			}
		case <-conn.Done():
			_ = ws.Close(websocket.StatusCode(conn.CloseCode()), "")
			return
		}
	}
}

func (h *RealtimeHandler) dispatch(ctx context.Context, conn *realtime.Conn, data []byte) {
	var frame realtime.ClientFrame
	if err := json.Unmarshal(data, &frame); err != nil {
		conn.Send(realtime.MarshalError(realtime.ErrCodeInvalidMessage))
		return
	}
	switch frame.Type {
	case realtime.TypeSubscribe:
		h.subscribe(ctx, conn, frame)
	case realtime.TypeUnsubscribe:
		h.unsubscribe(conn, frame)
	case realtime.TypePos:
		h.publishPos(conn, frame)
	case realtime.TypeStatus:
		h.setStatus(ctx, conn, frame)
	default:
		conn.Send(realtime.MarshalError(realtime.ErrCodeUnknownType))
	}
}

func (h *RealtimeHandler) subscribe(ctx context.Context, conn *realtime.Conn, frame realtime.ClientFrame) {
	octx, cancel := context.WithTimeout(ctx, realtimeOpTimeout)
	defer cancel()
	switch frame.Plane {
	case realtime.PlanePos:
		orgID, workspaceID, missionID, ok := parseTenantTriple(frame)
		if !ok {
			conn.Send(realtime.MarshalError(realtime.ErrCodeInvalidMessage))
			return
		}
		if _, err := authz.RequireOrgRole(octx, h.members, conn.UserID, orgID, member.RoleViewer); err != nil {
			conn.Send(realtime.MarshalError(realtime.ErrCodeForbidden))
			return
		}
		if _, err := h.missions.FindByID(octx, missionID, orgID, workspaceID); err != nil {
			conn.Send(realtime.MarshalError(realtime.ErrCodeNotFound))
			return
		}
		h.hub.SubscribePos(conn, missionID)
	case realtime.PlaneStatus:
		orgID, err := uuid.Parse(frame.OrgID)
		if err != nil {
			conn.Send(realtime.MarshalError(realtime.ErrCodeInvalidMessage))
			return
		}
		if _, err := authz.RequireOrgRole(octx, h.members, conn.UserID, orgID, member.RoleViewer); err != nil {
			conn.Send(realtime.MarshalError(realtime.ErrCodeForbidden))
			return
		}
		h.hub.SubscribeStatus(conn, orgID)
	default:
		conn.Send(realtime.MarshalError(realtime.ErrCodeInvalidMessage))
	}
}

func (h *RealtimeHandler) unsubscribe(conn *realtime.Conn, frame realtime.ClientFrame) {
	switch frame.Plane {
	case realtime.PlanePos:
		if missionID, err := uuid.Parse(frame.MissionID); err == nil {
			h.hub.UnsubscribePos(conn, missionID)
			return
		}
	case realtime.PlaneStatus:
		if orgID, err := uuid.Parse(frame.OrgID); err == nil {
			h.hub.UnsubscribeStatus(conn, orgID)
			return
		}
	}
	conn.Send(realtime.MarshalError(realtime.ErrCodeInvalidMessage))
}

func (h *RealtimeHandler) publishPos(conn *realtime.Conn, frame realtime.ClientFrame) {
	missionID, err := uuid.Parse(frame.MissionID)
	if err != nil {
		conn.Send(realtime.MarshalError(realtime.ErrCodeInvalidMessage))
		return
	}
	if !conn.AllowPos() {
		return
	}
	if err := h.hub.PublishPos(conn, missionID, frame.X, frame.Y); err != nil {
		conn.Send(realtime.MarshalError(realtime.ErrCodeNotSubscribed))
	}
}

func (h *RealtimeHandler) setStatus(ctx context.Context, conn *realtime.Conn, frame realtime.ClientFrame) {
	if !frame.Editing {
		orgID, orgErr := uuid.Parse(frame.OrgID)
		missionID, missionErr := uuid.Parse(frame.MissionID)
		if orgErr != nil || missionErr != nil {
			conn.Send(realtime.MarshalError(realtime.ErrCodeInvalidMessage))
			return
		}
		h.hub.SetEditing(conn, orgID, missionID, false)
		return
	}
	octx, cancel := context.WithTimeout(ctx, realtimeOpTimeout)
	defer cancel()
	orgID, workspaceID, missionID, ok := parseTenantTriple(frame)
	if !ok {
		conn.Send(realtime.MarshalError(realtime.ErrCodeInvalidMessage))
		return
	}
	if _, err := authz.RequireOrgRole(octx, h.members, conn.UserID, orgID, member.RoleViewer); err != nil {
		conn.Send(realtime.MarshalError(realtime.ErrCodeForbidden))
		return
	}
	if _, err := h.missions.FindByID(octx, missionID, orgID, workspaceID); err != nil {
		conn.Send(realtime.MarshalError(realtime.ErrCodeNotFound))
		return
	}
	h.hub.SetEditing(conn, orgID, missionID, true)
}

func parseTenantTriple(frame realtime.ClientFrame) (uuid.UUID, uuid.UUID, uuid.UUID, bool) {
	orgID, err := uuid.Parse(frame.OrgID)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}
	workspaceID, err := uuid.Parse(frame.WorkspaceID)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}
	missionID, err := uuid.Parse(frame.MissionID)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}
	return orgID, workspaceID, missionID, true
}
