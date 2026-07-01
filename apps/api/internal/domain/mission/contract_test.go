package mission_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	missiondomain "github.com/novarod/polina/apps/api/internal/domain/mission"
)

const linearGraph = `{"nodes":[{"id":"n1","type":"START"},{"id":"n2","type":"OBJECTIVE","data":{"reward":10}},{"id":"n3","type":"END"}],"edges":[{"id":"e1","source":"n1","target":"n2"},{"id":"e2","source":"n2","target":"n3"}]}`
const brokenGraph = `{"nodes":[{"id":"n1","type":"OBJECTIVE"}],"edges":[]}`

func TestCompile_LinearGraph(t *testing.T) {
	c, err := missiondomain.Compile("mission-1", json.RawMessage(linearGraph))
	require.NoError(t, err)

	assert.Equal(t, "mission-1", c.MissionID)
	assert.Equal(t, "n1", c.StartNode)
	require.Len(t, c.Nodes, 3)
	assert.Equal(t, []string{"n2"}, c.Nodes["n1"].Next)
	assert.Equal(t, []string{"n3"}, c.Nodes["n2"].Next)
	assert.Equal(t, []string{}, c.Nodes["n3"].Next, "END node has no outgoing edges")
	assert.JSONEq(t, `{"reward":10}`, string(c.Nodes["n2"].Data))
}

func TestCompile_InvalidGraph_Error(t *testing.T) {
	// No START, dead-end → validation fails, no contract.
	_, err := missiondomain.Compile("mission-1", json.RawMessage(brokenGraph))
	require.Error(t, err)
}

func TestContentHash_Deterministic(t *testing.T) {
	c1, err := missiondomain.Compile("mission-1", json.RawMessage(linearGraph))
	require.NoError(t, err)
	c2, err := missiondomain.Compile("mission-1", json.RawMessage(linearGraph))
	require.NoError(t, err)

	h1, err := c1.ContentHash()
	require.NoError(t, err)
	h2, err := c2.ContentHash()
	require.NoError(t, err)

	assert.Equal(t, h1, h2, "same graph in same mission must hash identically")
	assert.Len(t, h1, 64, "sha-256 hex")
}

func TestContentHash_DiffersByGraph(t *testing.T) {
	other := `{"nodes":[{"id":"n1","type":"START"},{"id":"n2","type":"END"}],"edges":[{"id":"e1","source":"n1","target":"n2"}]}`
	c1, err := missiondomain.Compile("mission-1", json.RawMessage(linearGraph))
	require.NoError(t, err)
	c2, err := missiondomain.Compile("mission-1", json.RawMessage(other))
	require.NoError(t, err)

	h1, _ := c1.ContentHash()
	h2, _ := c2.ContentHash()
	assert.NotEqual(t, h1, h2, "different graphs must hash differently")
}

func TestContentHash_DiffersByMission(t *testing.T) {
	c1, err := missiondomain.Compile("mission-1", json.RawMessage(linearGraph))
	require.NoError(t, err)
	c2, err := missiondomain.Compile("mission-2", json.RawMessage(linearGraph))
	require.NoError(t, err)

	h1, _ := c1.ContentHash()
	h2, _ := c2.ContentHash()
	assert.NotEqual(t, h1, h2, "identical graph in a different mission must hash differently (global UNIQUE)")
}
