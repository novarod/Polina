package workspace_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wsdomain "github.com/novarod/polina/apps/api/internal/domain/workspace"
)

func TestValidateName(t *testing.T) {
	cases := []struct {
		name string
		in   string
		ok   bool
	}{
		{"empty", "", false},
		{"too short", "a", false},
		{"min ok", "ab", true},
		{"normal", "Cyberpunk Team", true},
		{"max ok", strings.Repeat("x", 255), true},
		{"too long", strings.Repeat("x", 256), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := wsdomain.ValidateName(tc.in)
			if tc.ok {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestValidateDescription(t *testing.T) {
	assert.NoError(t, wsdomain.ValidateDescription(""))
	assert.NoError(t, wsdomain.ValidateDescription(strings.Repeat("x", 1000)))
	require.Error(t, wsdomain.ValidateDescription(strings.Repeat("x", 1001)))
}
