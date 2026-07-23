package realtime

import (
	"encoding/json"

	"github.com/google/uuid"
)

const ProtocolVersion = 1

const (
	PlanePos    = "pos"
	PlaneStatus = "status"
)

const (
	TypeAuth        = "auth"
	TypeSubscribe   = "subscribe"
	TypeUnsubscribe = "unsubscribe"
	TypePos         = "pos"
	TypeStatus      = "status"
	TypeHello       = "hello"
	TypeSnapshot    = "snapshot"
	TypeJoin        = "join"
	TypeLeave       = "leave"
	TypeError       = "error"
)

const (
	ErrCodeInvalidMessage = "invalid_message"
	ErrCodeUnknownType    = "unknown_type"
	ErrCodeForbidden      = "forbidden"
	ErrCodeNotFound       = "not_found"
	ErrCodeNotSubscribed  = "not_subscribed"
)

type User struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type MissionCount struct {
	MissionID uuid.UUID `json:"mission_id"`
	Count     int       `json:"count"`
}

type ClientFrame struct {
	Type        string  `json:"type"`
	Plane       string  `json:"plane"`
	OrgID       string  `json:"org_id"`
	WorkspaceID string  `json:"workspace_id"`
	MissionID   string  `json:"mission_id"`
	Editing     bool    `json:"editing"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Ticket      string  `json:"ticket"`
}

type helloFrame struct {
	Type string `json:"type"`
	V    int    `json:"v"`
	User User   `json:"user"`
}

type snapshotFrame struct {
	Type  string `json:"type"`
	Topic string `json:"topic"`
	Users []User `json:"users"`
}

type statusSnapshotFrame struct {
	Type     string         `json:"type"`
	Topic    string         `json:"topic"`
	Missions []MissionCount `json:"missions"`
}

type presenceFrame struct {
	Type  string `json:"type"`
	Topic string `json:"topic"`
	User  User   `json:"user"`
}

type posFrame struct {
	Type   string    `json:"type"`
	Topic  string    `json:"topic"`
	UserID uuid.UUID `json:"user_id"`
	X      float64   `json:"x"`
	Y      float64   `json:"y"`
}

type statusFrame struct {
	Type      string    `json:"type"`
	Topic     string    `json:"topic"`
	MissionID uuid.UUID `json:"mission_id"`
	Count     int       `json:"count"`
}

type errorFrame struct {
	Type string `json:"type"`
	Code string `json:"code"`
}

func PosTopic(missionID uuid.UUID) string {
	return "pos:mission:" + missionID.String()
}

func StatusTopic(orgID uuid.UUID) string {
	return "status:org:" + orgID.String()
}

func MarshalHello(user User) []byte {
	return mustMarshal(helloFrame{Type: TypeHello, V: ProtocolVersion, User: user})
}

func MarshalError(code string) []byte {
	return mustMarshal(errorFrame{Type: TypeError, Code: code})
}

func mustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
