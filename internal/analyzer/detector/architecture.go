package detector

import (
	"fmt"
	"strings"

	"github.com/dcsg/archway/internal/provider"
	"golang.org/x/tools/go/packages"
)

func DetectArchitecture(graph provider.DependencyGraph, pkgs []*packages.Package) provider.ArchitectureResult {
	results := []provider.ArchitectureResult{
		detectHexagonal(graph),
		detectClean(graph),
		detectDDD(graph, pkgs),
		detectLayered(graph),
		detectFlat(graph),
	}

	best := provider.ArchitectureResult{Pattern: "unrecognized", Confidence: 0.2, Evidence: []string{"No strong architecture signals found"}}
	for _, result := range results {
		if result.Confidence > best.Confidence {
			best = result
		}
	}

	secondary := []string{}
	for _, result := range results {
		if result.Pattern != best.Pattern && result.Confidence >= 0.5 {
			secondary = append(secondary, fmt.Sprintf("also resembles %s (%.2f)", result.Pattern, result.Confidence))
		}
	}
	best.Evidence = append(best.Evidence, secondary...)
	if best.Confidence < 0.3 {
		best.Evidence = append(best.Evidence, "Consider organizing packages by domain and dependency direction")
	}
	return best
}

func detectHexagonal(graph provider.DependencyGraph) provider.ArchitectureResult {
	hasDomain := hasLayer(graph, "domain")
	hasPorts := hasLayer(graph, "ports")
	hasAdapters := hasLayer(graph, "adapters")
	domainDependsOnAdapters := edgeBetweenLayers(graph, "domain", "adapters")

	evidence := []string{}
	score := 0.0
	if hasDomain {
		score += 0.25
		evidence = append(evidence, "found domain package(s)")
	}
	if hasPorts {
		score += 0.25
		evidence = append(evidence, "found ports package(s)")
	}
	if hasAdapters {
		score += 0.25
		evidence = append(evidence, "found adapters package(s)")
	}
	if !domainDependsOnAdapters && hasDomain && hasAdapters {
		score += 0.25
		evidence = append(evidence, "domain does not depend on adapters")
	}
	if score == 0 {
		score = 0.1
	}
	return provider.ArchitectureResult{Pattern: "hexagonal", Confidence: score, Evidence: evidence}
}

func detectClean(graph provider.DependencyGraph) provider.ArchitectureResult {
	hasUsecase := containsAnyLayer(graph, "application", "usecase")
	hasEntity := hasLayer(graph, "domain")
	hasInfra := containsAnyLayer(graph, "infrastructure", "adapters")
	flowGood := !edgeBetweenLayers(graph, "domain", "infrastructure")

	score := 0.0
	evidence := []string{}
	if hasUsecase {
		score += 0.3
		evidence = append(evidence, "found application/usecase layer")
	}
	if hasEntity {
		score += 0.3
		evidence = append(evidence, "found domain/entity layer")
	}
	if hasInfra {
		score += 0.2
		evidence = append(evidence, "found infrastructure layer")
	}
	if flowGood && hasEntity && hasInfra {
		score += 0.2
		evidence = append(evidence, "dependency direction resembles clean architecture")
	}
	if score == 0 {
		score = 0.1
	}
	return provider.ArchitectureResult{Pattern: "clean", Confidence: score, Evidence: evidence}
}

func detectDDD(graph provider.DependencyGraph, pkgs []*packages.Package) provider.ArchitectureResult {
	hasDomain := hasLayer(graph, "domain")
	hasApplication := hasLayer(graph, "application")
	hasInfra := hasLayer(graph, "infrastructure") || hasLayer(graph, "adapters")

	namingSignals := 0
	for _, pkg := range pkgs {
		if pkg == nil {
			continue
		}
		path := strings.ToLower(pkg.PkgPath)
		if strings.Contains(path, "aggregate") || strings.Contains(path, "valueobject") || strings.Contains(path, "repository") {
			namingSignals++
		}
	}

	score := 0.0
	evidence := []string{}
	if hasDomain {
		score += 0.25
		evidence = append(evidence, "found domain layer")
	}
	if hasApplication {
		score += 0.25
		evidence = append(evidence, "found application layer")
	}
	if hasInfra {
		score += 0.2
		evidence = append(evidence, "found infrastructure/adapters layer")
	}
	if namingSignals > 0 {
		score += 0.3
		evidence = append(evidence, "found DDD naming signals")
	}
	if score == 0 {
		score = 0.1
	}
	return provider.ArchitectureResult{Pattern: "ddd", Confidence: score, Evidence: evidence}
}

func detectLayered(graph provider.DependencyGraph) provider.ArchitectureResult {
	signals := 0
	for _, node := range graph.Nodes {
		p := strings.ToLower(node.Path)
		if strings.Contains(p, "/model") || strings.Contains(p, "/view") || strings.Contains(p, "/controller") || strings.Contains(p, "/handler") || strings.Contains(p, "/service") {
			signals++
		}
	}
	confidence := 0.2
	if signals >= 2 {
		confidence = 0.7
	}
	if signals >= 4 {
		confidence = 0.85
	}
	return provider.ArchitectureResult{Pattern: "layered", Confidence: confidence, Evidence: []string{fmt.Sprintf("detected %d layered naming signals", signals)}}
}

func detectFlat(graph provider.DependencyGraph) provider.ArchitectureResult {
	pkgCount := len(graph.Nodes)
	confidence := 0.1
	if pkgCount <= 2 {
		confidence = 0.9
	} else if pkgCount <= 4 {
		confidence = 0.5
	}
	return provider.ArchitectureResult{Pattern: "flat", Confidence: confidence, Evidence: []string{fmt.Sprintf("project has %d package(s)", pkgCount)}}
}

func hasLayer(graph provider.DependencyGraph, layer string) bool {
	for _, node := range graph.Nodes {
		if strings.EqualFold(node.Layer, layer) {
			return true
		}
	}
	return false
}

func containsAnyLayer(graph provider.DependencyGraph, layers ...string) bool {
	for _, layer := range layers {
		if hasLayer(graph, layer) {
			return true
		}
	}
	return false
}

func edgeBetweenLayers(graph provider.DependencyGraph, fromLayer, toLayer string) bool {
	nodeLayer := map[string]string{}
	for _, node := range graph.Nodes {
		nodeLayer[node.Path] = node.Layer
	}
	for _, edge := range graph.Edges {
		if nodeLayer[edge.From] == fromLayer && nodeLayer[edge.To] == toLayer {
			return true
		}
	}
	return false
}
