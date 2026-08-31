package memory

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

func Digest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func CanonicalDigest(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return Digest([]byte(fmt.Sprintf("%v", value)))
	}
	var normalized any
	if err := json.Unmarshal(data, &normalized); err != nil {
		return Digest(data)
	}
	canonical, err := json.Marshal(normalized)
	if err != nil {
		return Digest(data)
	}
	return Digest(canonical)
}

func DigestRecord(record EvidenceRecord) string {
	record.RecordDigest = ""
	return CanonicalDigest(record)
}

func DigestReplay(replay ReplayComparison) string {
	replay.ComparisonDigest = ""
	return CanonicalDigest(replay)
}

func DigestLineage(lineage CorpusLineage) string {
	lineage.LineageDigest = ""
	return CanonicalDigest(lineage)
}

func LoadDenominator(path string) (Denominator, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Denominator{}, err
	}
	var denominator Denominator
	if err := json.Unmarshal(data, &denominator); err != nil {
		return Denominator{}, err
	}
	if denominator.Schema != "gooo/counterexample-memory/denominator/v1" || !denominator.Fixed || denominator.Total != 12 || len(denominator.Cells) != 12 {
		return Denominator{}, errors.New("denominator is not the fixed 12-cell contract")
	}
	seenOrdinals := map[int]bool{}
	seenActivities := map[string]bool{}
	for _, cell := range denominator.Cells {
		if cell.Ordinal < 1 || cell.Ordinal > 12 || seenOrdinals[cell.Ordinal] || cell.Activity == "" || seenActivities[cell.Activity] {
			return Denominator{}, errors.New("denominator cells are not unique")
		}
		seenOrdinals[cell.Ordinal] = true
		seenActivities[cell.Activity] = true
	}
	for ordinal := 1; ordinal <= 12; ordinal++ {
		if !seenOrdinals[ordinal] {
			return Denominator{}, errors.New("denominator ordinals are not contiguous")
		}
	}
	return denominator, nil
}

func CompileSource(sourcePath string, source []byte, denominator Denominator) (SemanticIR, error) {
	activities := parseActivities(string(source))
	if len(activities) != denominator.Total {
		return SemanticIR{}, fmt.Errorf("source has %d activities, denominator requires %d", len(activities), denominator.Total)
	}
	byActivity := denominator.CellByActivity()
	seen := make(map[string]bool, len(activities))
	nodes := make([]ActivityIR, 0, len(activities))
	for _, activity := range activities {
		cell, ok := byActivity[activity.name]
		if !ok || seen[activity.name] {
			return SemanticIR{}, fmt.Errorf("activity %q is not uniquely bound to denominator", activity.name)
		}
		seen[activity.name] = true
		nodes = append(nodes, ActivityIR{
			ID:         "gooo://counterexample-memory/activity/" + kebab(activity.name),
			Name:       activity.name,
			SourceLine: activity.line,
			MetricID:   cell.MetricID,
			Artifact:   cell.Artifact,
			Evaluator:  cell.Evaluator,
		})
	}
	for _, cell := range denominator.Cells {
		if !seen[cell.Activity] {
			return SemanticIR{}, fmt.Errorf("denominator activity %q is absent from source", cell.Activity)
		}
	}
	return SemanticIR{Schema: IRSchema, SourcePath: sourcePath, SourceDigest: Digest(source), Nodes: nodes}, nil
}

func LoadMeta(sourcePath, contractPath, irPath string) (Meta, error) {
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		return Meta{}, err
	}
	contract, err := os.ReadFile(contractPath)
	if err != nil {
		return Meta{}, err
	}
	irData, err := os.ReadFile(irPath)
	if err != nil {
		return Meta{}, err
	}
	var ir SemanticIR
	if err := json.Unmarshal(irData, &ir); err != nil {
		return Meta{}, err
	}
	sourceDigest := Digest(source)
	if ir.Schema != IRSchema || ir.SourceDigest != sourceDigest || len(ir.Nodes) != 12 {
		return Meta{}, errors.New("semantic IR is not bound to the current Gooo source")
	}
	return Meta{SourcePath: sourcePath, SourceDigest: sourceDigest, ContractDigest: Digest(contract), IRDigest: Digest(irData), Bindings: BindingsFromIR(ir)}, nil
}

func BindingsFromIR(ir SemanticIR) map[string]ActivityBinding {
	bindings := make(map[string]ActivityBinding, len(ir.Nodes))
	for _, node := range ir.Nodes {
		bindings[node.Name] = ActivityBinding{
			Activity:          node.Name,
			SourcePath:        ir.SourcePath,
			SourceDigest:      ir.SourceDigest,
			IRNode:            node.ID,
			GeneratedArtifact: node.Artifact,
			Evaluator:         node.Evaluator,
			MetricID:          node.MetricID,
		}
	}
	return bindings
}

type parsedActivity struct {
	name string
	line int
}

func parseActivities(source string) []parsedActivity {
	lines := strings.Split(source, "\n")
	activities := make([]parsedActivity, 0)
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "activity ") {
			continue
		}
		rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "activity "))
		open := strings.IndexByte(rest, '(')
		if open <= 0 {
			continue
		}
		activities = append(activities, parsedActivity{name: strings.TrimSpace(rest[:open]), line: index + 1})
	}
	return activities
}

func kebab(value string) string {
	var builder strings.Builder
	for index, r := range value {
		if r >= 'A' && r <= 'Z' {
			if index > 0 {
				builder.WriteByte('-')
			}
			builder.WriteRune(r + ('a' - 'A'))
			continue
		}
		builder.WriteRune(r)
	}
	return builder.String()
}

