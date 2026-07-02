package dag_test

import (
	"fmt"
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

func TestValidate_MissingEndNode(t *testing.T) {
	g := dag.Graph{
		Nodes: []dag.Node{
			{ID: "n1", Type: "START"},
			{ID: "n2", Type: "OBJECTIVE"},
		},
		Edges: []dag.Edge{
			{ID: "e1", Source: "n1", Target: "n2"},
			{ID: "e2", Source: "n2", Target: "n1"},
		},
	}
	err := dag.Validate(g)
	if err == nil {
		t.Fatal("expected missing END node error, got nil")
	}
}

func TestValidate_StartToEndDirect(t *testing.T) {
	g := dag.Graph{
		Nodes: []dag.Node{
			{ID: "n1", Type: "START"},
			{ID: "n2", Type: "END"},
		},
		Edges: []dag.Edge{
			{ID: "e1", Source: "n1", Target: "n2"},
		},
	}
	if err := dag.Validate(g); err != nil {
		t.Fatalf("expected valid graph, got: %v", err)
	}
}

func TestValidate_UnreachableNode(t *testing.T) {
	g := dag.Graph{
		Nodes: []dag.Node{
			{ID: "n1", Type: "START"},
			{ID: "n2", Type: "END"},
			{ID: "n3", Type: "OBJECTIVE"},
			{ID: "n4", Type: "END"},
		},
		Edges: []dag.Edge{
			{ID: "e1", Source: "n1", Target: "n2"},
			{ID: "e2", Source: "n3", Target: "n4"},
		},
	}
	err := dag.Validate(g)
	if err == nil {
		t.Fatal("expected unreachable node error, got nil")
	}
}

func TestValidate_NodeLimit(t *testing.T) {
	if err := dag.Validate(chainGraph(dag.MaxNodes)); err != nil {
		t.Fatalf("expected graph with %d nodes to be valid, got: %v", dag.MaxNodes, err)
	}
	if err := dag.Validate(chainGraph(dag.MaxNodes + 1)); err == nil {
		t.Fatalf("expected graph with %d nodes to be invalid", dag.MaxNodes+1)
	}
}

func TestValidate_EdgeLimit(t *testing.T) {
	if err := dag.Validate(parallelEdgesGraph(dag.MaxEdges)); err != nil {
		t.Fatalf("expected graph with %d edges to be valid, got: %v", dag.MaxEdges, err)
	}
	if err := dag.Validate(parallelEdgesGraph(dag.MaxEdges + 1)); err == nil {
		t.Fatalf("expected graph with %d edges to be invalid", dag.MaxEdges+1)
	}
}

// chainGraph builds a valid linear graph START -> OBJECTIVE... -> END with n nodes.
func chainGraph(n int) dag.Graph {
	g := dag.Graph{}
	for i := 0; i < n; i++ {
		typ := "OBJECTIVE"
		switch i {
		case 0:
			typ = "START"
		case n - 1:
			typ = "END"
		}
		g.Nodes = append(g.Nodes, dag.Node{ID: fmt.Sprintf("n%d", i), Type: typ})
		if i > 0 {
			g.Edges = append(g.Edges, dag.Edge{
				ID: fmt.Sprintf("e%d", i), Source: fmt.Sprintf("n%d", i-1), Target: fmt.Sprintf("n%d", i),
			})
		}
	}
	return g
}

// parallelEdgesGraph builds a valid START -> END graph with n parallel edges.
func parallelEdgesGraph(n int) dag.Graph {
	g := dag.Graph{
		Nodes: []dag.Node{
			{ID: "n1", Type: "START"},
			{ID: "n2", Type: "END"},
		},
	}
	for i := 0; i < n; i++ {
		g.Edges = append(g.Edges, dag.Edge{ID: fmt.Sprintf("e%d", i), Source: "n1", Target: "n2"})
	}
	return g
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
