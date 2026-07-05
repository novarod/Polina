package mission

import (
	"encoding/json"

	"github.com/novarod/polina/apps/api/pkg/apierr"
	"github.com/novarod/polina/apps/api/pkg/hash"
)

type ContractNode struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
	Next []string        `json:"next"`
}

type Contract struct {
	MissionID string                  `json:"mission_id"`
	Version   int                     `json:"version"`
	Hash      string                  `json:"hash" jsonschema:"pattern=^[0-9a-f]{64}$"`
	StartNode string                  `json:"start_node"`
	Nodes     map[string]ContractNode `json:"nodes"`
}

type compileGraph struct {
	Nodes []struct {
		ID   string          `json:"id"`
		Type string          `json:"type"`
		Data json.RawMessage `json:"data,omitempty"`
	} `json:"nodes"`
	Edges []struct {
		Source string `json:"source"`
		Target string `json:"target"`
	} `json:"edges"`
}

func Compile(missionID string, raw json.RawMessage) (Contract, error) {
	if err := ValidateGraph(raw); err != nil {
		return Contract{}, err
	}
	var g compileGraph
	if err := json.Unmarshal(raw, &g); err != nil {
		return Contract{}, apierr.Validation("graph", "graph must be valid JSON with nodes and edges")
	}

	next := make(map[string][]string, len(g.Nodes))
	for _, e := range g.Edges {
		next[e.Source] = append(next[e.Source], e.Target)
	}

	nodes := make(map[string]ContractNode, len(g.Nodes))
	var startNode string
	for _, n := range g.Nodes {
		out := next[n.ID]
		if out == nil {
			out = []string{}
		}
		nodes[n.ID] = ContractNode{Type: n.Type, Data: n.Data, Next: out}
		if n.Type == "START" {
			startNode = n.ID
		}
	}

	return Contract{MissionID: missionID, StartNode: startNode, Nodes: nodes}, nil
}

func (c Contract) ContentHash() (string, error) {
	return hash.Mission(struct {
		MissionID string                  `json:"mission_id"`
		StartNode string                  `json:"start_node"`
		Nodes     map[string]ContractNode `json:"nodes"`
	}{MissionID: c.MissionID, StartNode: c.StartNode, Nodes: c.Nodes})
}
