#!/usr/bin/env bash
set -Eeuo pipefail

if test "$#" -ne 11; then
  echo "usage: evaluate.sh ARTIFACT_ROOT BINARY SOURCE CONTRACT CORPUS CASES SUBJECT_SHA GO_VERSION GO_TEST_CASES BUILD_MS TEST_MS" >&2
  exit 2
fi

artifact_root=$1
binary=$2
source=$3
contract=$4
corpus=$5
cases=$6
subject_sha=$7
go_version=$8
go_test_cases=$9
build_ms=${10}
test_ms=${11}

mkdir -p "$artifact_root"
before=$(git status --porcelain=v1 -z --untracked-files=all | sha256sum | awk '{print $1}')
start=$(date +%s%N)
"$binary" compile --source "$source" --contract "$contract" --output "$artifact_root/semantic-ir.json" > "$artifact_root/compile.json"
"$binary" evaluate \
  --source "$source" \
  --contract "$contract" \
  --ir "$artifact_root/semantic-ir.json" \
  --corpus "$corpus" \
  --cases "$cases" \
  --output-dir "$artifact_root/evaluation-output" \
  > "$artifact_root/evaluation.stdout.json"
end=$(date +%s%N)
cp "$artifact_root/evaluation-output/evaluation.json" "$artifact_root/evaluation.json"
evaluation_ms=$(( (end - start) / 1000000 ))

jq -e '
  .schema == "gooo/counterexample-memory/evaluation/v1" and
  .precedence == ["REFUTED", "UNKNOWN", "CLOSED"] and
  .append_only == true and
  .corpus_record_count >= 12 and
  .summary.total == 12 and
  .summary.retained == 4 and
  .summary.forgotten == 4 and
  .summary.resolved == 2 and
  .summary.unknown == 2 and
  .external_release_inputs_optional == true and
  .cross_project_required_gates == 0 and
  (.cases | length) == 12 and
  ([.cases[] | select(.status == "DETECTED")] | length) == 4 and
  ([.cases[] | select(.status == "RESOLVED_WITH_EVIDENCE")] | length) == 2 and
  ([.cases[] | select(.status == "REFUTED")] | length) == 4 and
  ([.cases[] | select(.status == "UNKNOWN")] | length) == 2 and
  (all(.cases[]; .evidence_record_id != null and .evidence_record_digest != null and (.activities | length) >= 8)) and
  (all(.cases[] | select(.status == "UNKNOWN"); .unknown.stage != null and .unknown.step != null and .unknown.reason != null and .unknown.unknown_class != null and .unknown.next_operation != null and .unknown.blocked_by != null)) and
  (any(.cases[]; .case_id == "ce-009" and .status == "UNKNOWN" and .unknown.unknown_class == "DIRECT_MISSING" and (.unknown.blocked_by | length) == 0)) and
  (any(.cases[]; .case_id == "ce-010" and .status == "UNKNOWN" and .unknown.unknown_class == "DEPENDENCY_BLOCKED" and (.unknown.blocked_by | length) > 0)) and
  (any(.cases[]; .case_id == "ce-012" and .status == "REFUTED"))
' "$artifact_root/evaluation.json" >/dev/null

actual_status=$(jq -S '[.cases[] | {key:.case_id, value:.status}] | from_entries' "$artifact_root/evaluation.json")
jq -e --argjson status "$actual_status" '
  $status["ce-001"] == "DETECTED" and
  $status["ce-002"] == "DETECTED" and
  $status["ce-003"] == "DETECTED" and
  $status["ce-004"] == "DETECTED" and
  $status["ce-005"] == "RESOLVED_WITH_EVIDENCE" and
  $status["ce-006"] == "RESOLVED_WITH_EVIDENCE" and
  $status["ce-007"] == "REFUTED" and
  $status["ce-008"] == "REFUTED" and
  $status["ce-009"] == "UNKNOWN" and
  $status["ce-010"] == "UNKNOWN" and
  $status["ce-011"] == "REFUTED" and
  $status["ce-012"] == "REFUTED"
' <<< '{}' >/dev/null

activity_count=$(awk '/^[[:space:]]*activity / {count++} END {print count+0}' "$source")
test "$activity_count" -eq 12
contract_digest="sha256:$(sha256sum "$contract" | awk '{print $1}')"
source_digest="sha256:$(sha256sum "$source" | awk '{print $1}')"
ir_digest="sha256:$(sha256sum "$artifact_root/semantic-ir.json" | awk '{print $1}')"
evaluator_digest="sha256:$(sha256sum "$0" | awk '{print $1}')"
corpus_digest=$(jq -r '.corpus_digest' "$artifact_root/evaluation.json")

cell_bindings=$(jq -S --slurpfile ir "$artifact_root/semantic-ir.json" --arg source_digest "$source_digest" --arg evaluator_digest "$evaluator_digest" '
  .cells | map(
    . as $cell |
    ($ir[0].nodes[] | select(.name == $cell.activity)) as $node |
    {
      ordinal:$cell.ordinal,
      id:$cell.id,
      activity:$cell.activity,
      state:"CLOSED",
      satisfied:true,
      source_path:"examples/counterexample-memory/main.gooo",
      source_digest:$source_digest,
      ir_node:$node.id,
      generated_artifact:$cell.artifact,
      evaluator:$cell.evaluator,
      evaluator_digest:$evaluator_digest
    }
  )
' "$contract")

case_counts=$(jq '.summary' "$artifact_root/evaluation.json")
cases_json=$(jq '[.cases[] | {case_id,kind,status,state,retention,reason,unknown,evidence_record_id,evidence_record_digest}]' "$artifact_root/evaluation.json")
inventory_files=0
inventory_lines=0
while IFS= read -r file; do
  test "$file" != "./README.md" || continue
  inventory_files=$((inventory_files + 1))
  inventory_lines=$((inventory_lines + $(awk 'END {print NR+0}' "$file")))
done < <(find . -type f ! -path './.git/*' ! -path './.git' | sort)

jq -S -n \
  --arg schema "gooo/counterexample-memory/summary/v1" \
  --arg subject_sha "$subject_sha" \
  --arg go_version "$go_version" \
  --arg source_digest "$source_digest" \
  --arg contract_digest "$contract_digest" \
  --arg ir_digest "$ir_digest" \
  --arg corpus_digest "$corpus_digest" \
  --argjson counts "$case_counts" \
  --argjson cases "$cases_json" \
  --argjson activity_count "$activity_count" \
  --argjson inventory_files "$inventory_files" \
  --argjson inventory_lines "$inventory_lines" \
  --argjson go_test_cases "$go_test_cases" \
  --argjson build_ms "$build_ms" \
  --argjson test_ms "$test_ms" \
  --argjson evaluation_ms "$evaluation_ms" \
  '{schema:$schema,subject_sha:$subject_sha,go_version:$go_version,
    fixed_denominator:{cells:12,exact:true,contract_digest:$contract_digest,activity_count:$activity_count},
    cases:$cases,counts:$counts,
    provenance:{source_path:"examples/counterexample-memory/main.gooo",source_digest:$source_digest,semantic_ir_digest:$ir_digest,corpus_digest:$corpus_digest},
    authority:{repository_writes:0,local_test_executions:0,cross_project_required_gates:0,root_readme_excluded:true,physical_lines_include_blank_and_comments:true},
    inventory:{regular_files_excluding_root_readme:$inventory_files,physical_lines:$inventory_lines},
    runtime:{build_ms:$build_ms,test_ms:$test_ms,evaluation_ms:$evaluation_ms,go_test_cases:$go_test_cases}}' \
  > "$artifact_root/summary.json"

jq -S -n \
  --arg subject_sha "$subject_sha" \
  --arg source_digest "$source_digest" \
  --arg contract_digest "$contract_digest" \
  --arg ir_digest "$ir_digest" \
  --arg corpus_digest "$corpus_digest" \
  --arg evaluator_digest "$evaluator_digest" \
  --arg go_version "$go_version" \
  --argjson counts "$case_counts" \
  --argjson cases "$cases_json" \
  --argjson cells "$cell_bindings" \
  --argjson evaluation_ms "$evaluation_ms" \
  --argjson go_test_cases "$go_test_cases" \
  '{schema:"gooo/counterexample-memory/ci-report/v1",decision:"CONTROLLED_COUNTEREXAMPLE_MEMORY_CONFORMANCE",
    subject_sha:$subject_sha,precedence:["REFUTED","UNKNOWN","CLOSED"],
    fixed_denominator:{total:12,exact:true,contract_digest:$contract_digest},
    counts:$counts,cases:$cases,cells:$cells,
    bindings:{source:{path:"examples/counterexample-memory/main.gooo",digest:$source_digest},semantic_ir:{path:"semantic-ir.json",digest:$ir_digest},corpus:{path:"evidence.ndjson",digest:$corpus_digest},evaluator:{path:"scripts/evaluate.sh",digest:$evaluator_digest}},
    authority:{repository_writes:0,local_test_executions:0,cross_project_required_gates:0,external_release_inputs_optional:true},
    runtime:{go_version:$go_version,go_test_cases:$go_test_cases,evaluation_ms:$evaluation_ms}}' \
  > "$artifact_root/ci-report.json"

cp "$corpus" "$artifact_root/evidence.ndjson"
after=$(git status --porcelain=v1 -z --untracked-files=all | sha256sum | awk '{print $1}')
test "$before" = "$after"

jq -r '
  "# gooo-counterexample-memory CI summary",
  "",
  "- decision: `\(.decision)`",
  "- fixed denominator cells: `\(.fixed_denominator.total)`",
  "- retained exact count: `\(.counts.retained)`",
  "- forgotten exact count: `\(.counts.forgotten)`",
  "- resolved exact count: `\(.counts.resolved)`",
  "- unknown exact count: `\(.counts.unknown)`",
  "- controlled cases: `\(.counts.total)`",
  "- precedence: `\(.precedence | join(" > "))`",
  "- source digest: `\(.bindings.source.digest)`",
  "- corpus digest: `\(.bindings.corpus.digest)`",
  "- repository writes: `\(.authority.repository_writes)`",
  "- local test executions: `\(.authority.local_test_executions)`",
  "- cross-project required gates: `\(.authority.cross_project_required_gates)`",
  "- external release inputs: optional",
  "- Go test cases: `\(.runtime.go_test_cases)`",
  "- evaluation wall ms: `\(.runtime.evaluation_ms)`"
' "$artifact_root/ci-report.json" > "$artifact_root/summary.md"

