package realtime_test

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/novarod/polina/apps/api/internal/application/realtime"
)

type frame struct {
	Type      string                  `json:"type"`
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
}

func drain(t *testing.T, c *realtime.Conn) []frame {
	t.Helper()
	frames := make([]frame, 0)
	for {
		select {
		case msg := <-c.Outbox():
			var f frame
			require.NoError(t, json.Unmarshal(msg, &f))
			frames = append(frames, f)
		default:
			return frames
		}
	}
}

func register(t *testing.T, h *realtime.Hub, name string) *realtime.Conn {
	t.Helper()
	c, err := h.Register(uuid.New(), name)
	require.NoError(t, err)
	return c
}

func TestSubscribePosSnapshotAndJoin(t *testing.T) {
	h := realtime.NewHub()
	missionID := uuid.New()
	alice := register(t, h, "Alice")
	bob := register(t, h, "Bob")

	h.SubscribePos(alice, missionID)
	aliceFrames := drain(t, alice)
	require.Len(t, aliceFrames, 1)
	assert.Equal(t, realtime.TypeSnapshot, aliceFrames[0].Type)
	assert.Equal(t, realtime.PosTopic(missionID), aliceFrames[0].Topic)
	require.Len(t, aliceFrames[0].Users, 1)
	assert.Equal(t, "Alice", aliceFrames[0].Users[0].Name)

	h.SubscribePos(bob, missionID)
	bobFrames := drain(t, bob)
	require.Len(t, bobFrames, 1)
	assert.Equal(t, realtime.TypeSnapshot, bobFrames[0].Type)
	assert.Len(t, bobFrames[0].Users, 2)

	aliceFrames = drain(t, alice)
	require.Len(t, aliceFrames, 1)
	assert.Equal(t, realtime.TypeJoin, aliceFrames[0].Type)
	assert.Equal(t, "Bob", aliceFrames[0].User.Name)
}

func TestPresenceDedupePerUser(t *testing.T) {
	h := realtime.NewHub()
	missionID := uuid.New()
	watcher := register(t, h, "Watcher")
	h.SubscribePos(watcher, missionID)
	drain(t, watcher)

	userID := uuid.New()
	tab1, err := h.Register(userID, "Dup")
	require.NoError(t, err)
	tab2, err := h.Register(userID, "Dup")
	require.NoError(t, err)

	h.SubscribePos(tab1, missionID)
	joins := drain(t, watcher)
	require.Len(t, joins, 1)
	assert.Equal(t, realtime.TypeJoin, joins[0].Type)

	h.SubscribePos(tab2, missionID)
	assert.Empty(t, drain(t, watcher))

	snapshot := drain(t, tab2)
	require.Len(t, snapshot, 1)
	assert.Len(t, snapshot[0].Users, 2)

	h.RemoveAll(tab2)
	assert.Empty(t, drain(t, watcher))

	h.RemoveAll(tab1)
	leaves := drain(t, watcher)
	require.Len(t, leaves, 1)
	assert.Equal(t, realtime.TypeLeave, leaves[0].Type)
	assert.Equal(t, userID, leaves[0].User.ID)
}

func TestPublishPosFanOutExcludesSender(t *testing.T) {
	h := realtime.NewHub()
	missionID := uuid.New()
	sender := register(t, h, "Sender")
	receiver := register(t, h, "Receiver")
	h.SubscribePos(sender, missionID)
	h.SubscribePos(receiver, missionID)
	drain(t, sender)
	drain(t, receiver)

	require.NoError(t, h.PublishPos(sender, missionID, 10, 20))

	assert.Empty(t, drain(t, sender))
	frames := drain(t, receiver)
	require.Len(t, frames, 1)
	assert.Equal(t, realtime.TypePos, frames[0].Type)
	assert.Equal(t, sender.UserID, frames[0].UserID)
	assert.Equal(t, float64(10), frames[0].X)
	assert.Equal(t, float64(20), frames[0].Y)
}

func TestPublishPosRequiresSubscription(t *testing.T) {
	h := realtime.NewHub()
	c := register(t, h, "Loner")
	assert.ErrorIs(t, h.PublishPos(c, uuid.New(), 1, 1), realtime.ErrNotSubscribed)
}

func TestSlowConsumerIsKicked(t *testing.T) {
	h := realtime.NewHub()
	missionID := uuid.New()
	sender := register(t, h, "Sender")
	slow := register(t, h, "Slow")
	h.SubscribePos(sender, missionID)
	h.SubscribePos(slow, missionID)
	drain(t, sender)

	for range 80 {
		require.NoError(t, h.PublishPos(sender, missionID, 1, 1))
	}

	select {
	case <-slow.Done():
		assert.Equal(t, realtime.CloseTryAgainLater, slow.CloseCode())
	default:
		t.Fatal("slow consumer was not kicked")
	}

	assert.Empty(t, drain(t, sender))
}

func TestConcurrentBroadcastAndChurn(t *testing.T) {
	h := realtime.NewHub()
	missionID := uuid.New()
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c := register(t, h, "Churner")
			h.SubscribePos(c, missionID)
			for range 50 {
				_ = h.PublishPos(c, missionID, 1, 2)
				for {
					select {
					case <-c.Outbox():
						continue
					default:
					}
					break
				}
			}
			h.RemoveAll(c)
		}()
	}
	wg.Wait()
	h.Close(time.Second)
}

func TestStatusEditingDedupeAndBroadcast(t *testing.T) {
	h := realtime.NewHub()
	orgID := uuid.New()
	missionID := uuid.New()
	watcher := register(t, h, "Watcher")
	h.SubscribeStatus(watcher, orgID)
	snapshot := drain(t, watcher)
	require.Len(t, snapshot, 1)
	assert.Equal(t, realtime.TypeSnapshot, snapshot[0].Type)
	assert.Empty(t, snapshot[0].Missions)

	userID := uuid.New()
	tab1, err := h.Register(userID, "Editor")
	require.NoError(t, err)
	tab2, err := h.Register(userID, "Editor")
	require.NoError(t, err)

	h.SetEditing(tab1, orgID, missionID, true)
	frames := drain(t, watcher)
	require.Len(t, frames, 1)
	assert.Equal(t, realtime.TypeStatus, frames[0].Type)
	assert.Equal(t, missionID, frames[0].MissionID)
	assert.Equal(t, 1, frames[0].Count)

	h.SetEditing(tab2, orgID, missionID, true)
	assert.Empty(t, drain(t, watcher))

	late := register(t, h, "Late")
	h.SubscribeStatus(late, orgID)
	lateSnapshot := drain(t, late)
	require.Len(t, lateSnapshot, 1)
	require.Len(t, lateSnapshot[0].Missions, 1)
	assert.Equal(t, 1, lateSnapshot[0].Missions[0].Count)

	h.RemoveAll(tab1)
	assert.Empty(t, drain(t, watcher))

	h.SetEditing(tab2, orgID, missionID, false)
	frames = drain(t, watcher)
	require.Len(t, frames, 1)
	assert.Equal(t, 0, frames[0].Count)
}

func TestPosThrottleBurstThenDrop(t *testing.T) {
	h := realtime.NewHub()
	c := register(t, h, "Fast")
	allowed := 0
	for range 30 {
		if c.AllowPos() {
			allowed++
		}
	}
	assert.Equal(t, 15, allowed)
}

func TestCloseKicksAllAndRejectsNewConns(t *testing.T) {
	h := realtime.NewHub()
	conns := make([]*realtime.Conn, 0, 3)
	for range 3 {
		c := register(t, h, "Closing")
		conns = append(conns, c)
		go func() {
			<-c.Done()
			h.RemoveAll(c)
		}()
	}

	start := time.Now()
	h.Close(2 * time.Second)
	assert.Less(t, time.Since(start), 2*time.Second)

	for _, c := range conns {
		select {
		case <-c.Done():
			assert.Equal(t, realtime.CloseGoingAway, c.CloseCode())
		default:
			t.Fatal("conn was not kicked on close")
		}
	}

	_, err := h.Register(uuid.New(), "TooLate")
	assert.ErrorIs(t, err, realtime.ErrHubClosed)
}

func TestKickIsIdempotent(t *testing.T) {
	h := realtime.NewHub()
	c := register(t, h, "Kicked")
	c.Kick(realtime.CloseTryAgainLater)
	c.Kick(realtime.CloseGoingAway)
	assert.Equal(t, realtime.CloseTryAgainLater, c.CloseCode())
	h.RemoveAll(c)
	h.RemoveAll(c)
}
