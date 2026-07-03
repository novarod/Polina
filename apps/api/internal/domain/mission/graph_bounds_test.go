package mission_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	missiondomain "github.com/novarod/polina/apps/api/internal/domain/mission"
	"github.com/novarod/polina/apps/api/pkg/apierr"
	"github.com/novarod/polina/apps/api/pkg/dag"
)

func TestValidateGraph_TooManyNodes_422(t *testing.T) {
	var b strings.Builder
	b.WriteString(`{"nodes":[`)
	for i := 0; i <= dag.MaxNodes; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, `{"id":"n%d","type":"OBJECTIVE"}`, i)
	}
	b.WriteString(`],"edges":[]}`)

	err := missiondomain.ValidateGraph([]byte(b.String()))
	require.Error(t, err)
	var appErr *apierr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusUnprocessableEntity, appErr.Code)
	assert.Contains(t, appErr.Message, fmt.Sprintf("exceeds the maximum of %d nodes", dag.MaxNodes))
}
