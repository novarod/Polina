package dag_test

import (
	"testing"

	"github.com/novarod/polina/apps/api/pkg/dag"
)

func TestValidate_ValidLinearGraph(t *testing.T) {
	g := dag.Graph{
		Nodes: []dag.Node{
			{ID: "n1", Type: "START"},
			{ID: "n2", Type: "OBJECTIVE"},
			{ID: "n3", Type: "END"},
		},
		Edges: []dag.Edge{
			{ID: "e1", Source: "n1", Target: "n2"},
			{ID: "e2", Source: "n2", Target: "n3"},
		},
	}
	if err := dag.Validate(g); err != nil {
		t.Fatalf("expected valid graph, got: %v", err)
	}
}

func TestValidate_DetectsCycle(t *testing.T) {
	g := dag.Graph{
		Nodes: []dag.Node{
			{ID: "n1", Type: "START"},
			{ID: "n2", Type: "OBJECTIVE"},
			{ID: "n3", Type: "END"},
		},
		Edges: []dag.Edge{
			{ID: "e1", Source: "n1", Target: "n2"},
			{ID: "e2", Source: "n2", Target: "n3"},
			{ID: "e3", Source: "n3", Target: "n2"},
		},
	}
	err := dag.Validate(g)
	if err == nil {
		t.Fatal("expected cycle error, got nil")
	}
}

func TestValidate_DetectsDanglingEdge(t *testing.T) {
	g := dag.Graph{
		Nodes: []dag.Node{
			{ID: "n1", Type: "START"},
			{ID: "n2", Type: "END"},
		},
		Edges: []dag.Edge{
			{ID: "e1", Source: "n1", Target: "n2"},
			{ID: "e2", Source: "n1", Target: "n_missing"},
		},
	}
	err := dag.Validate(g)
	if err == nil {
		t.Fatal("expected dangling edge error, got nil")
	}
}

func TestValidate_DetectsDeadEnd(t *testing.T) {
	g := dag.Graph{
		Nodes: []dag.Node{
			{ID: "n1", Type: "START"},
			{ID: "n2", Type: "OBJECTIVE"},
			{ID: "n3", Type: "END"},
		},
		Edges: []dag.Edge{
			{ID: "e1", Source: "n1", Target: "n3"},
		},
	}
	err := dag.Validate(g)
	if err == nil {
		t.Fatal("expected dead-end error, got nil")
	}
}

func TestValidate_MissingStartNode(t *testing.T) {
	g := dag.Graph{
		Nodes: []dag.Node{
			{ID: "n1", Type: "OBJECTIVE"},
			{ID: "n2", Type: "END"},
		},
		Edges: []dag.Edge{
			{ID: "e1", Source: "n1", Target: "n2"},
		},
	}
	err := dag.Validate(g)
	if err == nil {
		t.Fatal("expected missing START node error, got nil")
	}
}
