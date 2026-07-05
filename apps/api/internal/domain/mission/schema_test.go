package mission_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	missiondomain "github.com/novarod/polina/apps/api/internal/domain/mission"
)

const contractsDir = "../../../../../packages/contracts"

func compileSchema(t *testing.T) *jsonschema.Schema {
	t.Helper()
	sch, err := jsonschema.NewCompiler().Compile(filepath.Join(contractsDir, "schema", "contract.schema.json"))
	require.NoError(t, err)
	return sch
}

func fixturePaths(t *testing.T, kind string) []string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(contractsDir, "fixtures", kind, "*.json"))
	require.NoError(t, err)
	require.NotEmpty(t, paths, "no %s fixtures found", kind)
	return paths
}

func TestSchema_CompiledContractValidates(t *testing.T) {
	c, err := missiondomain.Compile("0b8e7f3a-9c41-4d2a-b6f0-2f4a5f1c9e77", json.RawMessage(linearGraph))
	require.NoError(t, err)
	c.Hash, err = c.ContentHash()
	require.NoError(t, err)

	data, err := json.Marshal(c)
	require.NoError(t, err)
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	require.NoError(t, err)

	assert.NoError(t, compileSchema(t).Validate(doc), "a real published contract must validate against the committed schema")
}

func TestSchema_ValidFixtures(t *testing.T) {
	sch := compileSchema(t)
	for _, path := range fixturePaths(t, "valid") {
		t.Run(filepath.Base(path), func(t *testing.T) {
			b, err := os.ReadFile(filepath.Clean(path))
			require.NoError(t, err)

			doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(b))
			require.NoError(t, err)
			assert.NoError(t, sch.Validate(doc), "valid fixture must pass the schema")

			var c missiondomain.Contract
			require.NoError(t, json.Unmarshal(b, &c))
			h, err := c.ContentHash()
			require.NoError(t, err)
			assert.Equal(t, h, c.Hash, "fixture hash must match pkg/hash.Mission recomputation")
		})
	}
}

func TestSchema_InvalidFixtures(t *testing.T) {
	sch := compileSchema(t)
	for _, path := range fixturePaths(t, "invalid") {
		t.Run(filepath.Base(path), func(t *testing.T) {
			b, err := os.ReadFile(filepath.Clean(path))
			require.NoError(t, err)

			doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(b))
			require.NoError(t, err)
			assert.Error(t, sch.Validate(doc), "invalid fixture must fail the schema")
		})
	}
}
