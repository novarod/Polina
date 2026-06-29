package mission_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	missiondomain "github.com/novarod/polina/apps/api/internal/domain/mission"
)

func TestValidateName(t *testing.T) {
	assert.Error(t, missiondomain.ValidateName(""))
	assert.Error(t, missiondomain.ValidateName("a"))
	assert.NoError(t, missiondomain.ValidateName("Old Country"))
	assert.NoError(t, missiondomain.ValidateName(strings.Repeat("x", 255)))
	assert.Error(t, missiondomain.ValidateName(strings.Repeat("x", 256)))
}

func TestValidateDescription(t *testing.T) {
	assert.NoError(t, missiondomain.ValidateDescription(""))
	assert.NoError(t, missiondomain.ValidateDescription(strings.Repeat("x", 1000)))
	assert.Error(t, missiondomain.ValidateDescription(strings.Repeat("x", 1001)))
}

func TestValidateGraph(t *testing.T) {
	valid := `{"nodes":[{"id":"n1","type":"START"},{"id":"n2","type":"OBJECTIVE"},{"id":"n3","type":"END"}],
		"edges":[{"id":"e1","source":"n1","target":"n2"},{"id":"e2","source":"n2","target":"n3"}]}`
	cyclic := `{"nodes":[{"id":"n1","type":"START"},{"id":"n2","type":"OBJECTIVE"},{"id":"n3","type":"END"}],
		"edges":[{"id":"e1","source":"n1","target":"n2"},{"id":"e2","source":"n2","target":"n3"},{"id":"e3","source":"n3","target":"n2"}]}`
	orphanEdge := `{"nodes":[{"id":"n1","type":"START"},{"id":"n2","type":"END"}],
		"edges":[{"id":"e1","source":"n1","target":"n2"},{"id":"e2","source":"n1","target":"ghost"}]}`
	noStart := `{"nodes":[{"id":"n1","type":"OBJECTIVE"},{"id":"n2","type":"END"}],
		"edges":[{"id":"e1","source":"n1","target":"n2"}]}`

	assert.NoError(t, missiondomain.ValidateGraph([]byte(valid)))

	for name, raw := range map[string]string{
		"cyclic":      cyclic,
		"orphan edge": orphanEdge,
		"no START":    noStart,
		"malformed":   `{not json`,
	} {
		t.Run(name, func(t *testing.T) {
			require.Error(t, missiondomain.ValidateGraph([]byte(raw)))
		})
	}
}
