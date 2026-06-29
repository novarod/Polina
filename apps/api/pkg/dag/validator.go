package dag

import "fmt"

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

	nodeSet := make(map[string]Node, len(g.Nodes))
	for _, n := range g.Nodes {
		nodeSet[n.ID] = n
	}

	adj := make(map[string][]string, len(g.Nodes))
	inDegree := make(map[string]int, len(g.Nodes))
	for _, n := range g.Nodes {
		adj[n.ID] = []string{}
		inDegree[n.ID] = 0
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
		inDegree[e.Target]++
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

func nodesOfType(nodes []Node, t string) []Node {
	var result []Node
	for _, n := range nodes {
		if n.Type == t {
			result = append(result, n)
		}
	}
	return result
}
