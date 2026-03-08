package graph

import (
	"path"
	"strings"

	"github.com/dcsg/archway/internal/config"
	"github.com/dcsg/archway/internal/provider"
	"golang.org/x/tools/go/packages"
)

func BuildGraph(pkgs []*packages.Package) provider.DependencyGraph {
	nodes := make(map[string]provider.PackageNode)
	edgeSet := make(map[string]provider.DependencyEdge)

	for _, pkg := range pkgs {
		if pkg == nil || pkg.PkgPath == "" {
			continue
		}
		nodes[pkg.PkgPath] = provider.PackageNode{
			Path:       pkg.PkgPath,
			Name:       pkg.Name,
			IsInternal: strings.Contains(pkg.PkgPath, "/internal/"),
			Layer:      guessLayer(pkg.PkgPath),
		}
		for importPath := range pkg.Imports {
			edge := provider.DependencyEdge{From: pkg.PkgPath, To: importPath, ImportType: "import"}
			key := edge.From + "->" + edge.To
			edgeSet[key] = edge
		}
	}

	graph := provider.DependencyGraph{
		Nodes: make([]provider.PackageNode, 0, len(nodes)),
		Edges: make([]provider.DependencyEdge, 0, len(edgeSet)),
	}

	for _, node := range nodes {
		graph.Nodes = append(graph.Nodes, node)
	}
	for _, edge := range edgeSet {
		graph.Edges = append(graph.Edges, edge)
	}
	graph.Cycles = FindCycles(graph)
	return graph
}

func FindCycles(graph provider.DependencyGraph) [][]string {
	adj := map[string][]string{}
	for _, edge := range graph.Edges {
		adj[edge.From] = append(adj[edge.From], edge.To)
	}

	visited := map[string]bool{}
	stack := map[string]bool{}
	pathStack := []string{}
	cycles := [][]string{}

	var dfs func(node string)
	dfs = func(node string) {
		visited[node] = true
		stack[node] = true
		pathStack = append(pathStack, node)

		for _, next := range adj[node] {
			if !visited[next] {
				dfs(next)
				continue
			}
			if stack[next] {
				cycle := []string{}
				for i := len(pathStack) - 1; i >= 0; i-- {
					cycle = append([]string{pathStack[i]}, cycle...)
					if pathStack[i] == next {
						break
					}
				}
				if len(cycle) > 1 {
					cycles = append(cycles, cycle)
				}
			}
		}

		stack[node] = false
		pathStack = pathStack[:len(pathStack)-1]
	}

	for _, node := range graph.Nodes {
		if !visited[node.Path] {
			dfs(node.Path)
		}
	}

	return dedupeCycles(cycles)
}

func LayerViolations(graph provider.DependencyGraph, components []config.Component) []provider.Violation {
	layerByPkg := map[string]string{}
	allowed := map[string]map[string]bool{}
	for _, comp := range components {
		allow := make(map[string]bool, len(comp.MayDependOn))
		for _, target := range comp.MayDependOn {
			allow[target] = true
		}
		allowed[comp.Name] = allow
	}

	for _, node := range graph.Nodes {
		layer := node.Layer
		if layer == "" {
			for _, comp := range components {
				if matchesAnyRule(node.Path, comp.In) {
					layer = comp.Name
					break
				}
			}
		}
		if layer != "" {
			layerByPkg[node.Path] = layer
		}
	}

	violations := []provider.Violation{}
	for _, edge := range graph.Edges {
		srcLayer := layerByPkg[edge.From]
		tgtLayer := layerByPkg[edge.To]
		if srcLayer == "" || tgtLayer == "" || srcLayer == tgtLayer {
			continue
		}
		if allowedLayers, ok := allowed[srcLayer]; ok {
			if allowedLayers[tgtLayer] {
				continue
			}
			violations = append(violations, provider.Violation{
				Rule:     "dependency",
				Message:  srcLayer + " must not depend on " + tgtLayer,
				Source:   edge.From,
				Target:   edge.To,
				Severity: "error",
			})
		}
	}
	return violations
}

// MatchesComponent returns true if the package path matches any of the component's patterns.
func MatchesComponent(pkgPath string, comp config.Component) bool {
	return matchesAnyRule(pkgPath, comp.In)
}

func matchesAnyRule(pkgPath string, patterns []string) bool {
	for _, pattern := range patterns {
		if pattern == "" {
			continue
		}
		if strings.HasSuffix(pattern, "/**") {
			prefix := strings.TrimSuffix(pattern, "/**")
			if strings.Contains(pkgPath, prefix) {
				return true
			}
			continue
		}
		if ok, _ := path.Match(pattern, pkgPath); ok {
			return true
		}
	}
	return false
}

func guessLayer(pkgPath string) string {
	lower := strings.ToLower(pkgPath)
	switch {
	case strings.Contains(lower, "/domain"):
		return "domain"
	case strings.Contains(lower, "/port") || strings.Contains(lower, "/ports"):
		return "ports"
	case strings.Contains(lower, "/adapter") || strings.Contains(lower, "/adapters"):
		return "adapters"
	case strings.Contains(lower, "/application") || strings.Contains(lower, "/usecase"):
		return "application"
	case strings.Contains(lower, "/infrastructure"):
		return "infrastructure"
	default:
		return ""
	}
}

func dedupeCycles(cycles [][]string) [][]string {
	seen := map[string]bool{}
	out := make([][]string, 0, len(cycles))
	for _, cycle := range cycles {
		key := strings.Join(cycle, "->")
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, cycle)
	}
	return out
}
