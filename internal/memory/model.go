package memory

import (
	"encoding/json"
)

const (
	Schema               = "gooo/counterexample-memory/evaluation/v1"
	InputSchema          = "gooo/counterexample-memory/case/v1"
	IRSchema             = "gooo/counterexample-memory/ir/v1"
	CorpusSchema         = "gooo/counterexample-memory/evidence-record/v1"
	ToolDigest           = "sha256:6c5f0b8dd7cfbb9c5ecaf2a97d1d99bcf5c643a8af80a23bd0b4f6d0ddf4b0d1"
	StatusDetected       = "DETECTED"
	StatusResolved       = "RESOLVED_WITH_EVIDENCE"
	StatusUnknown        = "UNKNOWN"
	StatusRefuted        = "REFUTED"
	StateClosed          = "CLOSED"
	StateUnknown         = "UNKNOWN"
	StateRefuted         = "REFUTED"
	UnknownDirectMissing = "DIRECT_MISSING"
	UnknownDependency    = "DEPENDENCY_BLOCKED"
)

var Precedence = []string{StateRefuted, StateUnknown, StateClosed}

type CorpusLineage struct {
	CorpusID           string `json:"corpus_id"`
	ParentCorpusDigest string `json:"parent_corpus_digest"`
	SourceVersion      string `json:"source_version"`
	LineageDigest      string `json:"lineage_digest"`
}

type ReplayComparison struct {
	PreviousVersion        string `json:"previous_version"`
	CandidateVersion       string `json:"candidate_version"`
	InputDigest            string `json:"input_digest"`
	PreviousOutcomeDigest  string `json:"previous_outcome_digest"`
	CandidateOutcomeDigest string `json:"candidate_outcome_digest"`
	Difference             string `json:"difference"`
	ComparisonDigest       string `json:"comparison_digest"`
}

type ResolutionEvidence struct {
	EvidenceID     string `json:"evidence_id"`
	EvidenceDigest string `json:"evidence_digest"`
	Basis          string `json:"basis"`
}

type EvidenceConflict struct {
	LeftDigest     string `json:"left_digest"`
	RightDigest    string `json:"right_digest"`
	ConflictReason string `json:"conflict_reason"`
}

type EvidenceRecord struct {
	Schema                       string              `json:"schema"`
	RecordID                     string              `json:"record_id"`
	Ordinal                      int                 `json:"ordinal"`
	Kind                         string              `json:"kind"`
	Subject                      string              `json:"subject"`
	SourceVersion                string              `json:"source_version"`
	ObservedCounterexampleDigest string              `json:"observed_counterexample_digest"`
	ExpectedResolutionEvidence  *ResolutionEvidence `json:"expected_resolution_evidence"`
	CorpusLineage                CorpusLineage       `json:"corpus_lineage"`
	Replay                       ReplayComparison    `json:"replay"`
	CausalFrontier               []string            `json:"causal_frontier"`
	InheritsFrom                 []string            `json:"inherits_from"`
	EvidenceConflict             *EvidenceConflict   `json:"evidence_conflict"`
	DenominatorDigest            string              `json:"denominator_digest"`
	PreviousRecordDigest         string              `json:"previous_record_digest"`
	RecordDigest                 string              `json:"record_digest"`
}

type UnknownTuple struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type ExternalRelease struct {
	Repository  string `json:"repository"`
	Tag         string `json:"tag"`
	CommitSHA   string `json:"commit_sha"`
	Asset       string `json:"asset"`
	AssetDigest string `json:"asset_digest"`
}

type CaseInput struct {
	Schema                     string              `json:"schema"`
	CaseID                     string              `json:"case_id"`
	Subject                    string              `json:"subject"`
	Kind                       string              `json:"kind"`
	EvidenceRecordID           string              `json:"evidence_record_id"`
	EvidenceRecordDigest       string              `json:"evidence_record_digest"`
	CandidateVersion           string              `json:"candidate_version"`
	CandidateMemoryRecordID    string              `json:"candidate_memory_record_id"`
	CandidateClaim             string              `json:"candidate_claim"`
	CandidateObservationDigest string              `json:"candidate_observation_digest"`
	Replay                     *ReplayComparison    `json:"replay"`
	CausalFrontier             []string             `json:"causal_frontier"`
	ResolutionEvidence         *ResolutionEvidence `json:"resolution_evidence"`
	InheritsFrom               []string             `json:"inherits_from"`
	Unknown                    *UnknownTuple        `json:"unknown"`
	CandidateDenominatorDigest string               `json:"candidate_denominator_digest"`
	ExternalRelease            *ExternalRelease    `json:"external_release"`
	FixtureExpectedStatus      string               `json:"fixture_expected_status"`
}

type ActivityBinding struct {
	Activity          string `json:"activity"`
	SourcePath        string `json:"source_path"`
	SourceDigest      string `json:"source_digest"`
	IRNode            string `json:"ir_node"`
	GeneratedArtifact string `json:"generated_artifact"`
	Evaluator         string `json:"evaluator"`
	MetricID          string `json:"metric_id"`
}

type CaseResult struct {
	Schema                 string               `json:"schema"`
	CaseID                 string               `json:"case_id"`
	Kind                   string               `json:"kind"`
	Subject                string               `json:"subject"`
	Status                 string               `json:"status"`
	State                  string               `json:"state"`
	Retention              string               `json:"retention"`
	Reason                 string               `json:"reason"`
	Unknown                *UnknownTuple        `json:"unknown"`
	EvidenceRecordID       string               `json:"evidence_record_id"`
	EvidenceRecordDigest   string               `json:"evidence_record_digest"`
	CausalFrontier         []string             `json:"causal_frontier"`
	Replay                 ReplayComparison     `json:"replay"`
	ResolutionEvidence     *ResolutionEvidence `json:"resolution_evidence"`
	InheritsFrom           []string             `json:"inherits_from"`
	ExternalRelease        *ExternalRelease    `json:"external_release"`
	Activities             []ActivityBinding    `json:"activities"`
}

type CaseSummary struct {
	Total     int `json:"total"`
	Retained  int `json:"retained"`
	Forgotten int `json:"forgotten"`
	Resolved  int `json:"resolved"`
	Unknown   int `json:"unknown"`
}

type EvaluationReport struct {
	Schema                        string            `json:"schema"`
	Decision                      string            `json:"decision"`
	State                         string            `json:"state"`
	Precedence                    []string          `json:"precedence"`
	SourcePath                    string            `json:"source_path"`
	SourceDigest                  string            `json:"source_digest"`
	ContractDigest                string            `json:"contract_digest"`
	SemanticIRDigest              string            `json:"semantic_ir_digest"`
	CorpusPath                    string            `json:"corpus_path"`
	CorpusDigest                  string            `json:"corpus_digest"`
	CorpusRecordCount             int               `json:"corpus_record_count"`
	AppendOnly                    bool              `json:"append_only"`
	Cases                         []CaseResult      `json:"cases"`
	Summary                       CaseSummary       `json:"summary"`
	ActivityBindings              []ActivityBinding `json:"activity_bindings"`
	ExternalReleaseInputsOptional bool              `json:"external_release_inputs_optional"`
	CrossProjectRequiredGates     int               `json:"cross_project_required_gates"`
}
type Denominator struct {
	Schema           string    `json:"schema"`
	DenominatorID    string    `json:"denominator_id"`
	CandidateID      string    `json:"candidate_id"`
	Total            int       `json:"total"`
	Fixed            bool      `json:"fixed"`
	Proofs           []Balance `json:"proofs"`
	IndicatorClasses []Balance `json:"indicator_classes"`
	Cells            []Cell    `json:"cells"`
}

type Balance struct {
	Choice string `json:"choice"`
	Class  string `json:"class"`
	Total  int    `json:"total"`
}

type Cell struct {
	Ordinal        int    `json:"ordinal"`
	ID             string `json:"id"`
	Activity       string `json:"activity"`
	ProofChoice    string `json:"proof_choice"`
	IndicatorClass string `json:"indicator_class"`
	MetricID       string `json:"metric_id"`
	MetricPath     string `json:"metric_path"`
	Artifact       string `json:"artifact"`
	Evaluator      string `json:"evaluator"`
}

type SemanticIR struct {
	Schema       string       `json:"schema"`
	SourcePath   string       `json:"source_path"`
	SourceDigest string       `json:"source_digest"`
	Nodes        []ActivityIR `json:"nodes"`
}

type ActivityIR struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	SourceLine int    `json:"source_line"`
	MetricID   string `json:"metric_id"`
	Artifact   string `json:"artifact"`
	Evaluator  string `json:"evaluator"`
}

type Meta struct {
	SourcePath     string
	SourceDigest   string
	ContractDigest string
	IRDigest       string
	Bindings       map[string]ActivityBinding
}

func (d Denominator) CellByActivity() map[string]Cell {
	result := make(map[string]Cell, len(d.Cells))
	for _, cell := range d.Cells {
		result[cell.Activity] = cell
	}
	return result
}

func (r EvaluationReport) JSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}
