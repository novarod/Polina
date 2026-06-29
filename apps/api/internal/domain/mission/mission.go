package mission

import (
	"encoding/json"
	"unicode/utf8"

	"github.com/novarod/polina/apps/api/pkg/apierr"
	"github.com/novarod/polina/apps/api/pkg/dag"
)

// Status is the mission lifecycle state (matches the mission_status enum).
type Status string

const (
	StatusDraft    Status = "DRAFT"
	StatusApproved Status = "APPROVED"
)

const (
	nameMin = 2
	nameMax = 255
	descMax = 1000
)

func ValidateName(name string) error {
	n := utf8.RuneCountInString(name)
	if n < nameMin || n > nameMax {
		return apierr.Validation("name", "name must be between 2 and 255 characters")
	}
	return nil
}

func ValidateDescription(description string) error {
	if utf8.RuneCountInString(description) > descMax {
		return apierr.Validation("description", "description must be at most 1000 characters")
	}
	return nil
}

// ValidateGraph parses the raw graph JSON and runs the structural DAG checks
// (single START, no cycles, no dead-ends, no orphan edges). Returns a 422
// validation error on malformed JSON or any structural problem.
func ValidateGraph(raw []byte) error {
	var g dag.Graph
	if err := json.Unmarshal(raw, &g); err != nil {
		return apierr.Validation("graph", "graph must be valid JSON with nodes and edges")
	}
	if err := dag.Validate(g); err != nil {
		return apierr.Validation("graph", err.Error())
	}
	return nil
}
