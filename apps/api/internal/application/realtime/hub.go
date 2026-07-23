package realtime

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

const (
	CloseGoingAway       = 1001
	ClosePolicyViolation = 1008
	CloseTryAgainLater   = 1013
)

const (
	sendBuffer   = 64
	posRateLimit = 10
	posRateBurst = 15
)

var (
	ErrHubClosed     = errors.New("realtime: hub closed")
	ErrNotSubscribed = errors.New("realtime: not subscribed to topic")
)

type editingKey struct {
	orgID     uuid.UUID
	missionID uuid.UUID
}

type Conn struct {
	UserID   uuid.UUID
	UserName string
	send     chan []byte
	done     chan struct{}
	once     sync.Once
	code     int
	pos      *rate.Limiter
	topics   map[string]struct{}
	editing  map[editingKey]struct{}
}

func (c *Conn) Kick(code int) {
	c.once.Do(func() {
		c.code = code
		close(c.done)
	})
}

func (c *Conn) Done() <-chan struct{} {
	return c.done
}

func (c *Conn) CloseCode() int {
	return c.code
}

func (c *Conn) Outbox() <-chan []byte {
	return c.send
}

func (c *Conn) AllowPos() bool {
	return c.pos.Allow()
}

func (c *Conn) Send(msg []byte) {
	select {
	case c.send <- msg:
	default:
		c.Kick(CloseTryAgainLater)
	}
}

type presenceEntry struct {
	name  string
	conns int
}

type topicState struct {
	subs     map[*Conn]struct{}
	presence map[uuid.UUID]*presenceEntry
}

func (t *topicState) userList() []User {
	users := make([]User, 0, len(t.presence))
	for id, entry := range t.presence {
		users = append(users, User{ID: id, Name: entry.name})
	}
	return users
}

type Hub struct {
	mu      sync.RWMutex
	closed  bool
	conns   map[*Conn]struct{}
	topics  map[string]*topicState
	editors map[uuid.UUID]map[uuid.UUID]map[uuid.UUID]int
	wg      sync.WaitGroup
}

func NewHub() *Hub {
	return &Hub{
		conns:   make(map[*Conn]struct{}),
		topics:  make(map[string]*topicState),
		editors: make(map[uuid.UUID]map[uuid.UUID]map[uuid.UUID]int),
	}
}

func (h *Hub) Register(userID uuid.UUID, name string) (*Conn, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return nil, ErrHubClosed
	}
	c := &Conn{
		UserID:   userID,
		UserName: name,
		send:     make(chan []byte, sendBuffer),
		done:     make(chan struct{}),
		pos:      rate.NewLimiter(posRateLimit, posRateBurst),
		topics:   make(map[string]struct{}),
		editing:  make(map[editingKey]struct{}),
	}
	h.conns[c] = struct{}{}
	h.wg.Add(1)
	return c, nil
}

func (h *Hub) SubscribePos(c *Conn, missionID uuid.UUID) {
	topic := PosTopic(missionID)
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed || !h.registered(c) {
		return
	}
	t := h.ensureTopic(topic)
	if _, ok := c.topics[topic]; !ok {
		c.topics[topic] = struct{}{}
		t.subs[c] = struct{}{}
		entry, ok := t.presence[c.UserID]
		if !ok {
			entry = &presenceEntry{name: c.UserName}
			t.presence[c.UserID] = entry
		}
		entry.conns++
		if entry.conns == 1 {
			fanOut(t, mustMarshal(presenceFrame{Type: TypeJoin, Topic: topic, User: User{ID: c.UserID, Name: c.UserName}}), c)
		}
	}
	c.Send(mustMarshal(snapshotFrame{Type: TypeSnapshot, Topic: topic, Users: t.userList()}))
}

func (h *Hub) SubscribeStatus(c *Conn, orgID uuid.UUID) {
	topic := StatusTopic(orgID)
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed || !h.registered(c) {
		return
	}
	t := h.ensureTopic(topic)
	if _, ok := c.topics[topic]; !ok {
		c.topics[topic] = struct{}{}
		t.subs[c] = struct{}{}
	}
	c.Send(mustMarshal(statusSnapshotFrame{Type: TypeSnapshot, Topic: topic, Missions: h.missionCounts(orgID)}))
}

func (h *Hub) UnsubscribePos(c *Conn, missionID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.unsubscribeLocked(c, PosTopic(missionID))
}

func (h *Hub) UnsubscribeStatus(c *Conn, orgID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.unsubscribeLocked(c, StatusTopic(orgID))
}

func (h *Hub) PublishPos(c *Conn, missionID uuid.UUID, x, y float64) error {
	topic := PosTopic(missionID)
	h.mu.RLock()
	defer h.mu.RUnlock()
	if _, ok := c.topics[topic]; !ok {
		return ErrNotSubscribed
	}
	t := h.topics[topic]
	if t == nil {
		return ErrNotSubscribed
	}
	fanOut(t, mustMarshal(posFrame{Type: TypePos, Topic: topic, UserID: c.UserID, X: x, Y: y}), c)
	return nil
}

func (h *Hub) SetEditing(c *Conn, orgID, missionID uuid.UUID, editing bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed || !h.registered(c) {
		return
	}
	key := editingKey{orgID: orgID, missionID: missionID}
	if editing {
		if _, ok := c.editing[key]; ok {
			return
		}
		c.editing[key] = struct{}{}
		h.applyEditingLocked(c.UserID, key, 1)
		return
	}
	if _, ok := c.editing[key]; !ok {
		return
	}
	delete(c.editing, key)
	h.applyEditingLocked(c.UserID, key, -1)
}

func (h *Hub) RemoveAll(c *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.registered(c) {
		return
	}
	for topic := range c.topics {
		h.unsubscribeLocked(c, topic)
	}
	for key := range c.editing {
		delete(c.editing, key)
		h.applyEditingLocked(c.UserID, key, -1)
	}
	delete(h.conns, c)
	h.wg.Done()
}

func (h *Hub) Close(timeout time.Duration) {
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return
	}
	h.closed = true
	conns := make([]*Conn, 0, len(h.conns))
	for c := range h.conns {
		conns = append(conns, c)
	}
	h.mu.Unlock()
	for _, c := range conns {
		c.Kick(CloseGoingAway)
	}
	waited := make(chan struct{})
	go func() {
		h.wg.Wait()
		close(waited)
	}()
	select {
	case <-waited:
	case <-time.After(timeout):
	}
}

func (h *Hub) registered(c *Conn) bool {
	_, ok := h.conns[c]
	return ok
}

func (h *Hub) ensureTopic(topic string) *topicState {
	t := h.topics[topic]
	if t == nil {
		t = &topicState{
			subs:     make(map[*Conn]struct{}),
			presence: make(map[uuid.UUID]*presenceEntry),
		}
		h.topics[topic] = t
	}
	return t
}

func (h *Hub) unsubscribeLocked(c *Conn, topic string) {
	t := h.topics[topic]
	if t == nil {
		return
	}
	if _, ok := t.subs[c]; !ok {
		return
	}
	delete(t.subs, c)
	delete(c.topics, topic)
	if entry, ok := t.presence[c.UserID]; ok {
		entry.conns--
		if entry.conns <= 0 {
			delete(t.presence, c.UserID)
			fanOut(t, mustMarshal(presenceFrame{Type: TypeLeave, Topic: topic, User: User{ID: c.UserID, Name: c.UserName}}), nil)
		}
	}
	if len(t.subs) == 0 {
		delete(h.topics, topic)
	}
}

func (h *Hub) applyEditingLocked(userID uuid.UUID, key editingKey, delta int) {
	byMission := h.editors[key.orgID]
	if byMission == nil {
		if delta < 0 {
			return
		}
		byMission = make(map[uuid.UUID]map[uuid.UUID]int)
		h.editors[key.orgID] = byMission
	}
	byUser := byMission[key.missionID]
	if byUser == nil {
		if delta < 0 {
			return
		}
		byUser = make(map[uuid.UUID]int)
		byMission[key.missionID] = byUser
	}
	before := len(byUser)
	byUser[userID] += delta
	if byUser[userID] <= 0 {
		delete(byUser, userID)
	}
	after := len(byUser)
	if after == 0 {
		delete(byMission, key.missionID)
		if len(byMission) == 0 {
			delete(h.editors, key.orgID)
		}
	}
	if before == after {
		return
	}
	topic := StatusTopic(key.orgID)
	if t := h.topics[topic]; t != nil {
		fanOut(t, mustMarshal(statusFrame{Type: TypeStatus, Topic: topic, MissionID: key.missionID, Count: after}), nil)
	}
}

func (h *Hub) missionCounts(orgID uuid.UUID) []MissionCount {
	byMission := h.editors[orgID]
	counts := make([]MissionCount, 0, len(byMission))
	for missionID, byUser := range byMission {
		counts = append(counts, MissionCount{MissionID: missionID, Count: len(byUser)})
	}
	return counts
}

func fanOut(t *topicState, msg []byte, exclude *Conn) {
	for sub := range t.subs {
		if sub == exclude {
			continue
		}
		sub.Send(msg)
	}
}
