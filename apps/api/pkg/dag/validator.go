package dag

import "fmt"

const (
	MaxNodes = 1000
	MaxEdges = 2000
)

type Node struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type Edge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
}

type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

type ValidationError struct {
	Errors []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("dag validation failed: %v", e.Errors)
}

func Validate(g Graph) error {
	var errs []string

	if len(g.Nodes) > MaxNodes {
		errs = append(errs, fmt.Sprintf("graph exceeds the maximum of %d nodes (got %d)", MaxNodes, len(g.Nodes)))
	}
	if len(g.Edges) > MaxEdges {
		errs = append(errs, fmt.Sprintf("graph exceeds the maximum of %d edges (got %d)", MaxEdges, len(g.Edges)))
	}
	if len(errs) > 0 {
		return &ValidationError{Errors: errs}
	}

	nodeSet := make(map[string]Node, len(g.Nodes))
	for _, n := range g.Nodes {
		nodeSet[n.ID] = n
	}

	adj := make(map[string][]string, len(g.Nodes))
	for _, n := range g.Nodes {
		adj[n.ID] = []string{}
	}
	for _, e := range g.Edges {
		if _, ok := nodeSet[e.Source]; !ok {
			errs = append(errs, fmt.Sprintf("edge %q references unknown source node %q", e.ID, e.Source))
			continue
		}
		if _, ok := nodeSet[e.Target]; !ok {
			errs = append(errs, fmt.Sprintf("edge %q references unknown target node %q", e.ID, e.Target))
			continue
		}
		adj[e.Source] = append(adj[e.Source], e.Target)
	}

	if len(errs) > 0 {
		return &ValidationError{Errors: errs}
	}

	if cycleNode := detectCycle(adj, g.Nodes); cycleNode != "" {
		errs = append(errs, fmt.Sprintf("cycle detected involving node %q", cycleNode))
	}

	starts := nodesOfType(g.Nodes, "START")
	if len(starts) == 0 {
		errs = append(errs, "graph must have exactly one START node")
	} else if len(starts) > 1 {
		errs = append(errs, "graph must have exactly one START node, found multiple")
	}

	if len(nodesOfType(g.Nodes, "END")) == 0 {
		errs = append(errs, "graph must have at least one END node")
	}

	if len(starts) == 1 {
		reachable := reachableFrom(adj, starts[0].ID)
		for _, n := range g.Nodes {
			if !reachable[n.ID] {
				errs = append(errs, fmt.Sprintf("node %q (%s) is not reachable from the START node", n.ID, n.Type))
			}
		}
	}

	for _, n := range g.Nodes {
		if n.Type == "END" {
			continue
		}
		if len(adj[n.ID]) == 0 {
			errs = append(errs, fmt.Sprintf("node %q (%s) has no outgoing edges (dead-end)", n.ID, n.Type))
		}
	}

	if len(errs) > 0 {
		return &ValidationError{Errors: errs}
	}
	return nil
}

func detectCycle(adj map[string][]string, nodes []Node) string {
	color := make(map[string]int, len(nodes))
	var dfs func(id string) string
	dfs = func(id string) string {
		color[id] = 1
		for _, next := range adj[id] {
			if color[next] == 1 {
				return next
			}
			if color[next] == 0 {
				if found := dfs(next); found != "" {
					return found
				}
			}
		}
		color[id] = 2
		return ""
	}
	for _, n := range nodes {
		if color[n.ID] == 0 {
			if found := dfs(n.ID); found != "" {
				return found
			}
		}
	}
	return ""
}

func reachableFrom(adj map[string][]string, start string) map[string]bool {
	visited := map[string]bool{start: true}
	queue := []string{start}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		for _, next := range adj[id] {
			if !visited[next] {
				visited[next] = true
				queue = append(queue, next)
			}
		}
	}
	return visited
}

func nodesOfType(nodes []Node, t string) []Node {
	var result []Node
	for _, n := range nodes {
		if n.Type == t {
			result = append(result, n)
		}
	}
	return result
}
