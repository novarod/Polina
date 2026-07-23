//go:build integration

package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/novarod/polina/apps/api/internal/adapters/http/handler"
	httpmw "github.com/novarod/polina/apps/api/internal/adapters/http/middleware"
	"github.com/novarod/polina/apps/api/internal/application/realtime"
	"github.com/novarod/polina/apps/api/internal/application/token"
	"github.com/novarod/polina/apps/api/internal/domain/member"
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
)

const rtSecret = "0123456789abcdef0123456789abcdef"

type rtWorld struct {
	hub      *realtime.Hub
	users    *rtUserRepo
	members  *rtMemberRepo
	missions *rtMissionRepo
	server   *httptest.Server
}

type rtUserRepo struct{ users map[uuid.UUID]ports.User }

func (r *rtUserRepo) Create(_ context.Context, u ports.User) (ports.User, error) { return u, nil }
func (r *rtUserRepo) FindByEmail(_ context.Context, _ string) (ports.User, error) {
	return ports.User{}, apierr.NotFound("user")
}
func (r *rtUserRepo) FindByID(_ context.Context, id uuid.UUID) (ports.User, error) {
	u, ok := r.users[id]
	if !ok {
		return ports.User{}, apierr.NotFound("user")
	}
	return u, nil
}
func (r *rtUserRepo) BumpTokenValidAfter(_ context.Context, _ uuid.UUID) error { return nil }

type rtMemberRepo struct{ byUserOrg map[string]ports.Member }

func rtMemberKey(userID, orgID uuid.UUID) string { return userID.String() + "/" + orgID.String() }

func (r *rtMemberRepo) Create(_ context.Context, m ports.Member) (ports.Member, error) {
	return m, nil
}
func (r *rtMemberRepo) FindByUserAndOrg(_ context.Context, userID, orgID uuid.UUID) (ports.Member, error) {
	m, ok := r.byUserOrg[rtMemberKey(userID, orgID)]
	if !ok {
		return ports.Member{}, apierr.NotFound("member")
	}
	return m, nil
}
func (r *rtMemberRepo) SoftDeleteByOrg(_ context.Context, _ uuid.UUID) error { return nil }

type rtMissionRepo struct{ missions map[uuid.UUID]ports.Mission }

func (r *rtMissionRepo) Create(_ context.Context, m ports.Mission) (ports.Mission, error) {
	return m, nil
}
func (r *rtMissionRepo) FindByID(_ context.Context, id, orgID, workspaceID uuid.UUID) (ports.Mission, error) {
	m, ok := r.missions[id]
	if !ok || m.OrganizationID != orgID || m.WorkspaceID != workspaceID {
		return ports.Mission{}, apierr.NotFound("mission")
	}
	return m, nil
}
func (r *rtMissionRepo) FindByIDForUpdate(ctx context.Context, id, orgID, workspaceID uuid.UUID) (ports.Mission, error) {
	return r.FindByID(ctx, id, orgID, workspaceID)
}
func (r *rtMissionRepo) FindActiveHash(_ context.Context, _, _ uuid.UUID) (string, error) {
	return "", apierr.NotFound("mission")
}
func (r *rtMissionRepo) List(_ context.Context, _, _ uuid.UUID) ([]ports.Mission, error) {
	return nil, nil
}
func (r *rtMissionRepo) UpdateGraph(_ context.Context, _, _, _ uuid.UUID, _ json.RawMessage) (ports.Mission, error) {
	return ports.Mission{}, apierr.NotFound("mission")
}
func (r *rtMissionRepo) Update(_ context.Context, _, _, _ uuid.UUID, _, _ string) (ports.Mission, error) {
	return ports.Mission{}, apierr.NotFound("mission")
}
func (r *rtMissionRepo) SetActiveVersion(_ context.Context, _, _, _ uuid.UUID, _, _ string) (ports.Mission, error) {
	return ports.Mission{}, apierr.NotFound("mission")
}
func (r *rtMissionRepo) SoftDelete(_ context.Context, _, _, _ uuid.UUID) error { return nil }

type rtFrame struct {
	Type      string                  `json:"type"`
	V         int                     `json:"v"`
	Topic     string                  `json:"topic"`
	User      *realtime.User          `json:"user"`
	Users     []realtime.User         `json:"users"`
	UserID    uuid.UUID               `json:"user_id"`
	MissionID uuid.UUID               `json:"mission_id"`
	Count     int                     `json:"count"`
	Missions  []realtime.MissionCount `json:"missions"`
	X         float64                 `json:"x"`
	Y         float64                 `json:"y"`
	Code      string                  `json:"code"`
	Ticket    string                  `json:"ticket"`
}

func newRTWorld(t *testing.T) *rtWorld {
	t.Helper()
	w := &rtWorld{
		hub:      realtime.NewHub(),
		users:    &rtUserRepo{users: make(map[uuid.UUID]ports.User)},
		members:  &rtMemberRepo{byUserOrg: make(map[string]ports.Member)},
		missions: &rtMissionRepo{missions: make(map[uuid.UUID]ports.Mission)},
	}

	e := echo.New()
	e.Use(echomiddleware.ContextTimeoutWithConfig(echomiddleware.ContextTimeoutConfig{
		Skipper: func(c echo.Context) bool { return strings.HasPrefix(c.Path(), "/realtime") },
		Timeout: 15 * time.Second,
	}))
	rtHandler := handler.NewRealtimeHandler(w.hub, w.users, w.members, w.missions, rtSecret, "http://localhost:3000")
	rt := e.Group("/realtime")
	rt.GET("/ticket", rtHandler.Ticket, httpmw.Auth(rtSecret, w.users))
	rt.GET("/ws", rtHandler.Connect, httpmw.AuthOptional(rtSecret, w.users))

	srv := httptest.NewUnstartedServer(e)
	srv.Config.ReadHeaderTimeout = 5 * time.Second
	srv.Config.ReadTimeout = 10 * time.Second
	srv.Config.WriteTimeout = 30 * time.Second
	srv.Start()
	t.Cleanup(func() {
		w.hub.Close(time.Second)
		srv.Close()
	})
	w.server = srv
	return w
}

func (w *rtWorld) addUser(t *testing.T, name string) ports.User {
	t.Helper()
	u := ports.User{ID: uuid.New(), Email: name + "@test.dev", Name: name}
	w.users.users[u.ID] = u
	return u
}

func (w *rtWorld) addMember(userID, orgID uuid.UUID, role member.Role) {
	w.members.byUserOrg[rtMemberKey(userID, orgID)] = ports.Member{
		ID: uuid.New(), UserID: userID, OrganizationID: orgID, Role: role,
	}
}

func (w *rtWorld) addMission(orgID, workspaceID uuid.UUID) ports.Mission {
	m := ports.Mission{ID: uuid.New(), OrganizationID: orgID, WorkspaceID: workspaceID}
	w.missions.missions[m.ID] = m
	return m
}

func (w *rtWorld) wsURL() string {
	return "ws" + strings.TrimPrefix(w.server.URL, "http") + "/realtime/ws"
}

func rtSessionToken(t *testing.T, userID uuid.UUID) string {
	t.Helper()
	claims := &token.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(rtSecret))
	require.NoError(t, err)
	return signed
}

func rtDial(t *testing.T, w *rtWorld, sessionToken string) *websocket.Conn {
	t.Helper()
	header := http.Header{}
	if sessionToken != "" {
		header.Set("Cookie", "session="+sessionToken)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ws, _, err := websocket.Dial(ctx, w.wsURL(), &websocket.DialOptions{HTTPHeader: header})
	require.NoError(t, err)
	t.Cleanup(func() { _ = ws.CloseNow() })
	return ws
}

func rtRead(t *testing.T, ws *websocket.Conn, timeout time.Duration) rtFrame {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_, data, err := ws.Read(ctx)
	require.NoError(t, err)
	var f rtFrame
	require.NoError(t, json.Unmarshal(data, &f))
	return f
}

func rtReadUntil(t *testing.T, ws *websocket.Conn, frameType string) rtFrame {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		f := rtRead(t, ws, 5*time.Second)
		if f.Type == frameType {
			return f
		}
	}
	t.Fatalf("frame %q not received", frameType)
	return rtFrame{}
}

func rtSend(t *testing.T, ws *websocket.Conn, frame map[string]any) {
	t.Helper()
	data, err := json.Marshal(frame)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, ws.Write(ctx, websocket.MessageText, data))
}

func rtSubscribePos(t *testing.T, ws *websocket.Conn, orgID, workspaceID, missionID uuid.UUID) {
	t.Helper()
	rtSend(t, ws, map[string]any{
		"type": "subscribe", "plane": "pos",
		"org_id": orgID.String(), "workspace_id": workspaceID.String(), "mission_id": missionID.String(),
	})
}

func TestRealtimeRejectsInvalidCookieOnUpgrade(t *testing.T) {
	w := newRTWorld(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	header := http.Header{}
	header.Set("Cookie", "session=not-a-jwt")
	ws, resp, err := websocket.Dial(ctx, w.wsURL(), &websocket.DialOptions{HTTPHeader: header})
	require.Error(t, err)
	if ws != nil {
		_ = ws.CloseNow()
	}
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestRealtimeNoCredentialsAndNoTicketCloses1008(t *testing.T) {
	w := newRTWorld(t)
	ws := rtDial(t, w, "")
	rtSend(t, ws, map[string]any{"type": "pos"})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _, err := ws.Read(ctx)
	require.Error(t, err)
	assert.Equal(t, websocket.StatusPolicyViolation, websocket.CloseStatus(err))
}

func TestRealtimeRejectsForeignOrigin(t *testing.T) {
	w := newRTWorld(t)
	user := w.addUser(t, "Origin")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	header := http.Header{}
	header.Set("Cookie", "session="+rtSessionToken(t, user.ID))
	header.Set("Origin", "http://evil.example")
	ws, _, err := websocket.Dial(ctx, w.wsURL(), &websocket.DialOptions{HTTPHeader: header})
	require.Error(t, err)
	if ws != nil {
		_ = ws.CloseNow()
	}
}

func TestRealtimeAcceptsFrontendOrigin(t *testing.T) {
	w := newRTWorld(t)
	user := w.addUser(t, "Origin")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	header := http.Header{}
	header.Set("Cookie", "session="+rtSessionToken(t, user.ID))
	header.Set("Origin", "http://localhost:3000")
	ws, _, err := websocket.Dial(ctx, w.wsURL(), &websocket.DialOptions{HTTPHeader: header})
	require.NoError(t, err)
	defer func() { _ = ws.CloseNow() }()
	hello := rtRead(t, ws, 5*time.Second)
	assert.Equal(t, "hello", hello.Type)
}

func TestRealtimeHelloCarriesIdentityAndVersion(t *testing.T) {
	w := newRTWorld(t)
	user := w.addUser(t, "Ana")
	ws := rtDial(t, w, rtSessionToken(t, user.ID))
	hello := rtRead(t, ws, 5*time.Second)
	assert.Equal(t, "hello", hello.Type)
	assert.Equal(t, 1, hello.V)
	require.NotNil(t, hello.User)
	assert.Equal(t, user.ID, hello.User.ID)
	assert.Equal(t, "Ana", hello.User.Name)
}

func TestRealtimeSubscribeSnapshotAck(t *testing.T) {
	w := newRTWorld(t)
	orgID, workspaceID := uuid.New(), uuid.New()
	user := w.addUser(t, "Ana")
	w.addMember(user.ID, orgID, member.RoleViewer)
	mission := w.addMission(orgID, workspaceID)

	ws := rtDial(t, w, rtSessionToken(t, user.ID))
	rtRead(t, ws, 5*time.Second)
	rtSubscribePos(t, ws, orgID, workspaceID, mission.ID)

	snapshot := rtRead(t, ws, 5*time.Second)
	assert.Equal(t, "snapshot", snapshot.Type)
	assert.Equal(t, realtime.PosTopic(mission.ID), snapshot.Topic)
	require.Len(t, snapshot.Users, 1)
	assert.Equal(t, user.ID, snapshot.Users[0].ID)
}

func TestRealtimeTwoClientsSeeEachOther(t *testing.T) {
	w := newRTWorld(t)
	orgID, workspaceID := uuid.New(), uuid.New()
	ana := w.addUser(t, "Ana")
	bia := w.addUser(t, "Bia")
	w.addMember(ana.ID, orgID, member.RoleViewer)
	w.addMember(bia.ID, orgID, member.RoleDesigner)
	mission := w.addMission(orgID, workspaceID)

	wsAna := rtDial(t, w, rtSessionToken(t, ana.ID))
	rtRead(t, wsAna, 5*time.Second)
	rtSubscribePos(t, wsAna, orgID, workspaceID, mission.ID)
	rtRead(t, wsAna, 5*time.Second)

	wsBia := rtDial(t, w, rtSessionToken(t, bia.ID))
	rtRead(t, wsBia, 5*time.Second)
	rtSubscribePos(t, wsBia, orgID, workspaceID, mission.ID)
	biaSnapshot := rtRead(t, wsBia, 5*time.Second)
	assert.Len(t, biaSnapshot.Users, 2)

	join := rtRead(t, wsAna, 5*time.Second)
	assert.Equal(t, "join", join.Type)
	require.NotNil(t, join.User)
	assert.Equal(t, bia.ID, join.User.ID)
	assert.Equal(t, "Bia", join.User.Name)

	rtSend(t, wsBia, map[string]any{"type": "pos", "mission_id": mission.ID.String(), "x": 42.5, "y": 17.25})
	pos := rtRead(t, wsAna, 5*time.Second)
	assert.Equal(t, "pos", pos.Type)
	assert.Equal(t, bia.ID, pos.UserID)
	assert.Equal(t, 42.5, pos.X)
	assert.Equal(t, 17.25, pos.Y)
}

func TestRealtimeForeignOrgSubscribeReturnsErrorFrameAndStaysOpen(t *testing.T) {
	w := newRTWorld(t)
	orgID, workspaceID := uuid.New(), uuid.New()
	foreignOrg := uuid.New()
	user := w.addUser(t, "Ana")
	w.addMember(user.ID, orgID, member.RoleViewer)
	mission := w.addMission(orgID, workspaceID)

	ws := rtDial(t, w, rtSessionToken(t, user.ID))
	rtRead(t, ws, 5*time.Second)

	rtSubscribePos(t, ws, foreignOrg, workspaceID, mission.ID)
	errFrame := rtRead(t, ws, 5*time.Second)
	assert.Equal(t, "error", errFrame.Type)
	assert.Equal(t, "forbidden", errFrame.Code)

	rtSubscribePos(t, ws, orgID, workspaceID, mission.ID)
	snapshot := rtRead(t, ws, 5*time.Second)
	assert.Equal(t, "snapshot", snapshot.Type)
}

func TestRealtimePosThrottleDropsExcessWithoutKick(t *testing.T) {
	w := newRTWorld(t)
	orgID, workspaceID := uuid.New(), uuid.New()
	sender := w.addUser(t, "Sender")
	receiver := w.addUser(t, "Receiver")
	w.addMember(sender.ID, orgID, member.RoleViewer)
	w.addMember(receiver.ID, orgID, member.RoleViewer)
	mission := w.addMission(orgID, workspaceID)

	wsSender := rtDial(t, w, rtSessionToken(t, sender.ID))
	rtRead(t, wsSender, 5*time.Second)
	rtSubscribePos(t, wsSender, orgID, workspaceID, mission.ID)
	rtRead(t, wsSender, 5*time.Second)

	wsReceiver := rtDial(t, w, rtSessionToken(t, receiver.ID))
	rtRead(t, wsReceiver, 5*time.Second)
	rtSubscribePos(t, wsReceiver, orgID, workspaceID, mission.ID)
	rtRead(t, wsReceiver, 5*time.Second)
	rtRead(t, wsSender, 5*time.Second)

	total := 40
	for i := range total {
		rtSend(t, wsSender, map[string]any{"type": "pos", "mission_id": mission.ID.String(), "x": float64(i), "y": 0})
	}

	received := 0
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		_, _, err := wsReceiver.Read(ctx)
		cancel()
		if err != nil {
			break
		}
		received++
	}
	assert.GreaterOrEqual(t, received, 1)
	assert.Less(t, received, total)

	rtSubscribePos(t, wsSender, orgID, workspaceID, mission.ID)
	snapshot := rtReadUntil(t, wsSender, "snapshot")
	assert.Equal(t, "snapshot", snapshot.Type)
}

func TestRealtimeConnectionSurvivesServerTimeouts(t *testing.T) {
	w := newRTWorld(t)
	orgID, workspaceID := uuid.New(), uuid.New()
	viewer := w.addUser(t, "Viewer")
	mover := w.addUser(t, "Mover")
	w.addMember(viewer.ID, orgID, member.RoleViewer)
	w.addMember(mover.ID, orgID, member.RoleViewer)
	mission := w.addMission(orgID, workspaceID)

	wsViewer := rtDial(t, w, rtSessionToken(t, viewer.ID))
	rtRead(t, wsViewer, 5*time.Second)
	rtSubscribePos(t, wsViewer, orgID, workspaceID, mission.ID)
	rtRead(t, wsViewer, 5*time.Second)

	wsMover := rtDial(t, w, rtSessionToken(t, mover.ID))
	rtRead(t, wsMover, 5*time.Second)
	rtSubscribePos(t, wsMover, orgID, workspaceID, mission.ID)
	rtRead(t, wsMover, 5*time.Second)
	rtRead(t, wsViewer, 5*time.Second)

	deadline := time.Now().Add(12 * time.Second)
	frames := 0
	for time.Now().Before(deadline) {
		rtSend(t, wsMover, map[string]any{"type": "pos", "mission_id": mission.ID.String(), "x": 1, "y": 1})
		pos := rtRead(t, wsViewer, 5*time.Second)
		assert.Equal(t, "pos", pos.Type)
		frames++
		time.Sleep(1500 * time.Millisecond)
	}
	assert.GreaterOrEqual(t, frames, 8)
}

func TestRealtimeTicketFlow(t *testing.T) {
	w := newRTWorld(t)
	user := w.addUser(t, "Ana")
	session := rtSessionToken(t, user.ID)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, w.server.URL+"/realtime/ticket", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+session)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var body struct {
		Ticket string `json:"ticket"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	require.NotEmpty(t, body.Ticket)

	ws := rtDial(t, w, "")
	rtSend(t, ws, map[string]any{"type": "auth", "ticket": body.Ticket})
	hello := rtRead(t, ws, 5*time.Second)
	assert.Equal(t, "hello", hello.Type)
	require.NotNil(t, hello.User)
	assert.Equal(t, user.ID, hello.User.ID)
}

func TestRealtimeExpiredTicketCloses1008(t *testing.T) {
	w := newRTWorld(t)
	user := w.addUser(t, "Ana")
	expired, err := token.NewRealtimeTicket(rtSecret, user.ID, time.Now().Add(-2*time.Minute))
	require.NoError(t, err)

	ws := rtDial(t, w, "")
	rtSend(t, ws, map[string]any{"type": "auth", "ticket": expired})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _, readErr := ws.Read(ctx)
	require.Error(t, readErr)
	assert.Equal(t, websocket.StatusPolicyViolation, websocket.CloseStatus(readErr))
}

func TestRealtimeSessionJWTIsNotAValidTicket(t *testing.T) {
	w := newRTWorld(t)
	user := w.addUser(t, "Ana")
	session := rtSessionToken(t, user.ID)

	ws := rtDial(t, w, "")
	rtSend(t, ws, map[string]any{"type": "auth", "ticket": session})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _, readErr := ws.Read(ctx)
	require.Error(t, readErr)
	assert.Equal(t, websocket.StatusPolicyViolation, websocket.CloseStatus(readErr))
}

func TestRealtimeStatusPlaneBroadcastsEditingCounts(t *testing.T) {
	w := newRTWorld(t)
	orgID, workspaceID := uuid.New(), uuid.New()
	watcher := w.addUser(t, "Watcher")
	editor := w.addUser(t, "Editor")
	w.addMember(watcher.ID, orgID, member.RoleViewer)
	w.addMember(editor.ID, orgID, member.RoleDesigner)
	mission := w.addMission(orgID, workspaceID)

	wsWatcher := rtDial(t, w, rtSessionToken(t, watcher.ID))
	rtRead(t, wsWatcher, 5*time.Second)
	rtSend(t, wsWatcher, map[string]any{"type": "subscribe", "plane": "status", "org_id": orgID.String()})
	snapshot := rtRead(t, wsWatcher, 5*time.Second)
	assert.Equal(t, "snapshot", snapshot.Type)
	assert.Equal(t, realtime.StatusTopic(orgID), snapshot.Topic)
	assert.Empty(t, snapshot.Missions)

	wsEditor := rtDial(t, w, rtSessionToken(t, editor.ID))
	rtRead(t, wsEditor, 5*time.Second)
	rtSend(t, wsEditor, map[string]any{
		"type": "status", "org_id": orgID.String(), "workspace_id": workspaceID.String(),
		"mission_id": mission.ID.String(), "editing": true,
	})

	status := rtRead(t, wsWatcher, 5*time.Second)
	assert.Equal(t, "status", status.Type)
	assert.Equal(t, mission.ID, status.MissionID)
	assert.Equal(t, 1, status.Count)

	require.NoError(t, wsEditor.Close(websocket.StatusNormalClosure, ""))
	status = rtRead(t, wsWatcher, 5*time.Second)
	assert.Equal(t, "status", status.Type)
	assert.Equal(t, 0, status.Count)
}
