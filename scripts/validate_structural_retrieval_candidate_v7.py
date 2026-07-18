#!/usr/bin/env python3
"""Independently replay a protocol-v7 candidate capability or captured result."""

from __future__ import annotations

import argparse
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import shutil
import stat
import subprocess
import sys
import tempfile
from typing import Any


_ADAPTER_PATH = Path(__file__).with_name("structural_retrieval_candidate_protocol_v7.py")
_SPEC = importlib.util.spec_from_file_location("_c3_candidate_protocol_v7_adapter", _ADAPTER_PATH)
assert _SPEC is not None and _SPEC.loader is not None
adapter = importlib.util.module_from_spec(_SPEC)
sys.modules[_SPEC.name] = adapter
_SPEC.loader.exec_module(adapter)


CANDIDATE_EXECUTION_AUTHORIZED = False
ACCEPTED_MAIN_SHA256 = adapter.ACCEPTED_MAIN_SHA256
ACCEPTED_MAIN_TEST_SHA256 = adapter.ACCEPTED_MAIN_TEST_SHA256
CAPABILITY_FILES = adapter.CAPABILITY_FILES
GATE_HARNESS_SCHEMA = "structural-retrieval-protocol-v7-candidate-gate-harness.v1"
RESULT_SCHEMA = "structural-retrieval-protocol-v7-candidate-validator-result.v1"
MAX_FILES = 2_048
MAX_BYTES = 268_435_456


class CandidateValidatorError(ValueError):
    """A bounded, generic candidate replay failure."""


def canonical_json(value: Any) -> bytes:
    return adapter.canonical_json(value)


def validate_capability_layout(root: Path) -> dict[str, str]:
    try:
        adapter.require_directory(root, 0o700, "capability_layout_invalid")
        return adapter._manifest_snapshot(root, CAPABILITY_FILES)
    except adapter.CandidateAdapterError as exc:
        raise CandidateValidatorError(str(exc)) from exc


def make_test_capability(root: Path) -> None:
    (root / "freeze").mkdir(parents=True, mode=0o700)
    root.chmod(0o700)
    for relative in CAPABILITY_FILES:
        path = root / relative
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_bytes(relative.encode())
        path.chmod(0o600)


def make_test_capability_manifest() -> dict[str, Any]:
    return {
        "$schema": adapter.CAPABILITY_SCHEMA,
        "status": "accepted_unexecuted",
        "effect_claim": False,
        "candidate_execution_authorized": False,
        "registered_variable": adapter.REGISTERED_VARIABLE,
        "candidate_delta_sha256": "1" * 64,
        "candidate_commit": "a" * 40,
        "candidate_tree": "b" * 40,
        "parent_authority_sha256": adapter.ACCEPTED_PARENT_AUTHORITY_SHA256,
        "parent_output_sha256": adapter.ACCEPTED_PARENT_OUTPUT_SHA256,
        "parent_validator_record_hash": adapter.ACCEPTED_PARENT_VALIDATOR_RECORD_HASH,
        "parent_validator_payload_sha256": adapter.ACCEPTED_PARENT_VALIDATOR_PAYLOAD_SHA256,
        "privacy_policy_sha256": "2" * 64,
        "adapter_sha256": "3" * 64,
        "validator_sha256": "4" * 64,
        "files": {
            "candidate-authority.v4.json": "5" * 64,
            "freeze/controller-runtime": "6" * 64,
            "freeze/candidate-runtime": "7" * 64,
            "freeze/source.bundle": "8" * 64,
        },
        "future_output_root_sha256": "9" * 64,
    }


def make_test_capability_hashes(manifest: dict[str, Any]) -> dict[str, str]:
    return dict(manifest["files"])


def make_test_candidate_authority(manifest: dict[str, Any]) -> dict[str, Any]:
    delta = {
        "variable": adapter.REGISTERED_VARIABLE,
        "baseline_commit": "1" * 40, "baseline_tree": "2" * 40,
        "candidate_commit": manifest["candidate_commit"], "candidate_tree": manifest["candidate_tree"],
        "diff_sha256": "3" * 64, "name_status_sha256": "4" * 64,
        "name_status": ["M\tcli/cmd/search.go"], "allowed_paths": ["cli/cmd/search.go"],
        "before_blob_sha256": {"cli/cmd/search.go": "5" * 64},
        "after_blob_sha256": {"cli/cmd/search.go": "6" * 64},
        "bundle_sha256": manifest["files"]["freeze/source.bundle"], "bundle_heads_sha256": "7" * 64,
    }
    delta_hash = adapter.candidate_delta_sha256(delta)
    manifest["candidate_delta_sha256"] = delta_hash
    return {
        "candidate_delta": delta,
        "expected_provenance": {
            "commit": delta["candidate_commit"], "tree": delta["candidate_tree"],
            "controller_commit": delta["baseline_commit"], "controller_tree": delta["baseline_tree"],
            "candidate_delta_sha256": delta_hash, "bundle_sha256": delta["bundle_sha256"],
            "runtime_sha256": manifest["files"]["freeze/candidate-runtime"],
        },
        "runtime_source_capsule": {"head_commit": delta["candidate_commit"], "head_tree": delta["candidate_tree"]},
        "parent_baseline": {
            "authority_sha256": manifest["parent_authority_sha256"],
            "output_sha256": manifest["parent_output_sha256"],
            "validator_record_hash": manifest["parent_validator_record_hash"],
            "validator_payload_sha256": manifest["parent_validator_payload_sha256"],
        },
    }


def validate_capability_manifest(data: bytes, file_hashes: dict[str, str], authority: dict[str, Any]) -> dict[str, Any]:
    try:
        value = adapter.decode_canonical(data, "capability_manifest_invalid")
    except adapter.CandidateAdapterError as exc:
        raise CandidateValidatorError(str(exc)) from exc
    if set(value) != adapter.CAPABILITY_MANIFEST_KEYS or value.get("$schema") != adapter.CAPABILITY_SCHEMA or value.get("status") != "accepted_unexecuted" or value.get("effect_claim") is not False or value.get("candidate_execution_authorized") is not False or value.get("registered_variable") != adapter.REGISTERED_VARIABLE:
        raise CandidateValidatorError("capability_manifest_invalid")
    if file_hashes and value.get("files") != file_hashes:
        raise CandidateValidatorError("capability_manifest_invalid")
    if not adapter.valid_sha256(value.get("candidate_delta_sha256")) or not adapter.valid_oid(value.get("candidate_commit")) or not adapter.valid_oid(value.get("candidate_tree")) or any(not adapter.valid_sha256(item) for item in file_hashes.values()):
        raise CandidateValidatorError("capability_manifest_invalid")
    if value.get("parent_authority_sha256") != adapter.ACCEPTED_PARENT_AUTHORITY_SHA256 or value.get("parent_output_sha256") != adapter.ACCEPTED_PARENT_OUTPUT_SHA256 or value.get("parent_validator_record_hash") != adapter.ACCEPTED_PARENT_VALIDATOR_RECORD_HASH or value.get("parent_validator_payload_sha256") != adapter.ACCEPTED_PARENT_VALIDATOR_PAYLOAD_SHA256:
        raise CandidateValidatorError("capability_manifest_invalid")
    try:
        adapter.validate_manifest_authority_bindings(value, authority)
    except adapter.CandidateAdapterError as exc:
        raise CandidateValidatorError("capability_manifest_invalid") from exc
    return value


def expected_result_artifacts(output: dict[str, Any]) -> set[str]:
    if output.get("$schema") != "structural-retrieval-controller-output.v4" or not isinstance(output.get("runs"), list):
        raise CandidateValidatorError("candidate_output_invalid")

    def relative_artifact(value: Any) -> str:
        if not isinstance(value, str) or not value:
            raise CandidateValidatorError("candidate_output_invalid")
        path = Path(value)
        if path.is_absolute() or path.as_posix() != value or ".." in path.parts or value.endswith("/"):
            raise CandidateValidatorError("candidate_output_invalid")
        return value

    expected = {
        "candidate-capture-manifest.json",
        "candidate/controller-output.v4.json",
        "candidate/" + relative_artifact(output.get("history_path")),
        "candidate/" + relative_artifact(output.get("privacy_manifest_path")),
        "freeze/candidate-authority.v4.json",
        "freeze/capability-manifest.json",
        "freeze/controller-runtime",
        "freeze/candidate-runtime",
        "freeze/source.bundle",
    }
    for index, run in enumerate(output["runs"], start=1):
        if not isinstance(run, dict):
            raise CandidateValidatorError("candidate_output_invalid")
        expected.add("candidate/" + relative_artifact(run.get("result_path")))
        expected.add("candidate/" + relative_artifact(run.get("report_path")))
        expected.add(f"candidate/runtime/{index:02d}.stderr")
    if any(path.endswith("/") or ".." in Path(path).parts or Path(path).is_absolute() for path in expected):
        raise CandidateValidatorError("candidate_output_invalid")
    return expected


def snapshot_result(root: Path, expected: set[str]) -> dict[str, str]:
    try:
        adapter.require_directory(root, 0o700, "candidate_output_invalid")
    except adapter.CandidateAdapterError as exc:
        raise CandidateValidatorError(str(exc)) from exc
    seen: dict[str, str] = {}
    count = 0
    total = 0
    for path in sorted(root.rglob("*")):
        info = path.lstat()
        relative = path.relative_to(root).as_posix()
        if stat.S_ISLNK(info.st_mode):
            raise CandidateValidatorError("candidate_output_invalid")
        if stat.S_ISDIR(info.st_mode):
            if stat.S_IMODE(info.st_mode) != 0o700:
                raise CandidateValidatorError("candidate_output_invalid")
            continue
        if relative not in expected or not stat.S_ISREG(info.st_mode) or stat.S_IMODE(info.st_mode) != 0o600:
            raise CandidateValidatorError("candidate_output_invalid")
        try:
            digest = adapter.sha256_file(path, error="candidate_output_invalid")
        except adapter.CandidateAdapterError as exc:
            raise CandidateValidatorError(str(exc)) from exc
        count += 1
        total += info.st_size
        if count > MAX_FILES or total > MAX_BYTES:
            raise CandidateValidatorError("candidate_output_invalid")
        seen[relative] = digest
    if set(seen) != expected:
        raise CandidateValidatorError("candidate_output_invalid")
    return seen


def validate_capture_manifest_bytes(data: bytes, privacy_terms: tuple[str, ...]) -> dict[str, Any]:
    try:
        value = adapter.decode_canonical(data, "candidate_output_invalid")
    except adapter.CandidateAdapterError as exc:
        raise CandidateValidatorError("candidate_output_invalid") from exc
    if set(value) != adapter.CAPTURE_MANIFEST_KEYS or adapter.privacy_hit(data, privacy_terms):
        raise CandidateValidatorError("candidate_output_invalid")
    return value


def scan_retained_metadata(root: Path, expected: set[str], privacy_terms: tuple[str, ...]) -> None:
    for relative in sorted(expected):
        if not (relative.endswith(".json") or relative.endswith(".jsonl")):
            continue
        try:
            data = adapter.read_regular(root / relative, mode=0o600, cap=adapter.MAX_METADATA_BYTES, error="candidate_output_invalid")
        except adapter.CandidateAdapterError as exc:
            raise CandidateValidatorError("candidate_output_invalid") from exc
        if adapter.privacy_hit(data, privacy_terms):
            raise CandidateValidatorError("candidate_output_invalid")


def validate_gate_harness_result(data: bytes) -> dict[str, Any]:
    try:
        value = adapter.decode_canonical(data, "gate_result_invalid")
    except adapter.CandidateAdapterError as exc:
        raise CandidateValidatorError(str(exc)) from exc
    keys = {
        "$schema", "status", "candidate_status", "microbench_structural_owner_recall_delta",
        "blocking_false_structural_claim_count", "strong_blast_radius_recall_regression", "run_count",
        "run_actuals", "validated_source_main_sha256", "validated_source_test_sha256",
    }
    if set(value) != keys or value.get("$schema") != GATE_HARNESS_SCHEMA or value.get("status") != "accepted" or value.get("candidate_status") not in {"keep", "discard", "crash", "invalid"} or value.get("run_count") != adapter.ACCEPTED_PARENT_RUN_COUNT or value.get("validated_source_main_sha256") != ACCEPTED_MAIN_SHA256 or value.get("validated_source_test_sha256") != ACCEPTED_MAIN_TEST_SHA256:
        raise CandidateValidatorError("gate_result_invalid")
    recall = value.get("microbench_structural_owner_recall_delta")
    false_count = value.get("blocking_false_structural_claim_count")
    regression = value.get("strong_blast_radius_recall_regression")
    if not isinstance(recall, (int, float)) or isinstance(recall, bool) or not isinstance(false_count, int) or isinstance(false_count, bool) or false_count < 0 or regression not in {0, 1}:
        raise CandidateValidatorError("gate_result_invalid")
    if value["candidate_status"] == "keep" and (false_count != 0 or regression != 0):
        raise CandidateValidatorError("gate_result_invalid")
    actuals = value.get("run_actuals")
    budget_keys = {"wall_time_millis", "cpu_time_millis", "max_rss_bytes", "process_count", "sqlite_row_count", "logical_dump_bytes", "stdout_bytes", "stderr_bytes", "case_count"}
    if not isinstance(actuals, list) or len(actuals) != adapter.ACCEPTED_PARENT_RUN_COUNT:
        raise CandidateValidatorError("gate_result_invalid")
    for order, row in enumerate(actuals, start=1):
        if not isinstance(row, dict) or set(row) != {"order", "mode", "case_id", "actual_budget"} or row.get("order") != order or row.get("mode") not in {"isolated", "combined", "scale"} or not isinstance(row.get("case_id"), str):
            raise CandidateValidatorError("gate_result_invalid")
        budget = row.get("actual_budget")
        if not isinstance(budget, dict) or set(budget) != budget_keys or any(not isinstance(number, int) or isinstance(number, bool) or number < 0 for number in budget.values()):
            raise CandidateValidatorError("gate_result_invalid")
    return value


def rejected_result(_detail: str) -> dict[str, Any]:
    return {"$schema": RESULT_SCHEMA, "status": "rejected", "failure_class": "candidate_validation_failed", "effect_claim": False, "candidate_execution_authorized": False}


GATE_HARNESS = r'''package main

import (
    "bytes"
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "path/filepath"
    "reflect"
    "sort"
    "testing"
)

type candidateGateConfig struct {
    Authority string `json:"authority"`; Controller string `json:"controller"`; Runtime string `json:"runtime"`
    BRoot string `json:"b_root"`; CRoot string `json:"c_root"`; Bundle string `json:"bundle"`; Policy string `json:"policy"`
    ParentRoot string `json:"parent_root"`; ParentAuthority string `json:"parent_authority"`; ParentOutput string `json:"parent_output"`; ParentStore string `json:"parent_store"`
    CandidateRoot string `json:"candidate_root"`; Result string `json:"result"`
}

func gateRead(root, relative string, max int64) ([]byte, error) { return parentRelativeArtifact(root, relative, max) }

func gateRootCoverage(root string, output controllerOutput) error {
    allowed:=map[string]bool{output.HistoryPath:true,output.PrivacyManifestPath:true,"controller-output.v4.json":true}
    for index,run:=range output.Runs { allowed[run.ResultPath]=true; allowed[run.ReportPath]=true; allowed[fmt.Sprintf("runtime/%02d.stderr",index+1)]=true }
    seen:=map[string]bool{}; err:=filepath.WalkDir(root,func(path string,entry os.DirEntry,walkErr error)error{ if walkErr!=nil{return walkErr}; if entry.Type()&os.ModeSymlink!=0{return errors.New("symlink")}; if entry.IsDir(){return nil}; info,err:=entry.Info(); if err!=nil || !info.Mode().IsRegular(){return errors.New("non_regular")}; relative,err:=filepath.Rel(root,path); if err!=nil{return err}; relative=filepath.ToSlash(relative); if !allowed[relative] || seen[relative]{return errors.New("extra")}; seen[relative]=true; return nil }); if err!=nil || len(seen)!=len(allowed){return errors.New("coverage")}; return nil
}

func gateReplayCandidate(root string, output controllerOutput, history []historyRecord, fixtures []fixtureCase, bench benchmarkConfig, authority controllerAuthorityV4, authorityBytes []byte, candidateSourceRoot, fixturesPath, benchmarkPath, scorerPath string, scanner *privacyScanner) (evalReport, []map[string]any, error) {
    known := map[string][]byte{}
    actuals := []map[string]any{}
    var combined evalReport
    for index, run := range output.Runs {
        selected, mode, caseID, err := expectedParentRun(fixtures,bench,index); if err != nil || run.Mode!=mode || run.CaseID!=caseID { return evalReport{},nil,errors.New("ordered_run_invalid") }
        resultBytes, err := gateRead(root, run.ResultPath, authority.ScanCaps.SingleDurableArtifactBytesMax); if err != nil || shaString(string(resultBytes)) != run.ResultSHA256 { return evalReport{},nil,errors.New("result_invalid") }
        reportBytes, err := gateRead(root, run.ReportPath, authority.ScanCaps.SingleDurableArtifactBytesMax); if err != nil || shaString(string(reportBytes)) != run.ReportSHA256 { return evalReport{},nil,errors.New("report_invalid") }
        var response armResponse; var durable durableReport
        if decodeStrictBytes(resultBytes, &response) != nil || response.Schema != armResponseSchema || decodeStrictBytes(reportBytes, &durable) != nil || durable.Schema != durableReportV4Schema { return evalReport{},nil,errors.New("report_invalid") }
        if validateLogicalDump(durable.Database) != nil || verifyReportWithProbes(selected,response,durable.Database,mode,durable.DirectProbes,durable.Report) != nil { return evalReport{},nil,errors.New("score_replay_failed") }
        if durable.ResultPath!=run.ResultPath || durable.ResultSHA256!=run.ResultSHA256 || durable.BudgetLimits!=authority.BudgetLimits || durable.ActualBudget!=run.ActualBudget || durable.CanonicalRowBytesDefinition!=authority.CanonicalRowBytesDefinition || !reflect.DeepEqual(durable.ContextThresholdAuthority,bench.ContextThresholdAuthority) || durable.ContextThresholdAuthorityLineSHA256!=shaString(authority.ContextThresholdAuthorityRecord) || durable.BudgetVerdict!=map[bool]string{true:"within_limits",false:"over_limit"}[classifyBudget(authority.BudgetLimits,run.ActualBudget)=="score"] { return evalReport{},nil,errors.New("report_authority_invalid") }
        row := history[index]
        if row.RecordHash!=run.HistoryRecordHash || row.ParentKeep!=authority.ParentBaseline.HistoryTailRecordHash || row.ChangedVariable!="direct_hit_containment_owner_substitution" || !reflect.DeepEqual(row.ChangedPaths,authority.CandidateDelta.AllowedPaths) || row.ResultPath!=run.ResultPath || row.ResultSHA256!=run.ResultSHA256 || row.Budgets!=run.ActualBudget || row.Status!="invalid" || row.Reason!="diagnostic_unadmitted" || row.ErrorClass!="" || row.ErrorSHA256!="" || run.CandidateDeltaSHA256!=authority.Expected.CandidateDeltaSHA256 || run.BundleSHA256!=authority.Expected.BundleSHA256 { return evalReport{},nil,errors.New("history_binding_invalid") }
        wantEvidence := []string{run.ReportPath+"#sha256="+run.ReportSHA256,bench.ContextThresholdAuthority.CheckinRef+"#sha256="+bench.ContextThresholdAuthority.CheckinSHA256}; if !reflect.DeepEqual(row.Evidence,wantEvidence) || durable.Report.Provenance==nil { return evalReport{},nil,errors.New("history_evidence_invalid") }
        want := authority.persistenceAuthority().Expected; want.ExperimentID=fmt.Sprintf("%s-%02d",authority.Expected.ExperimentID,index+1); want.ArmID=mode; if caseID!="" { want.ArmID += ":"+caseID }; want.LogicalDumpSHA256=durable.Database.LogicalSHA256; want.CorpusMode=mode; want.ProjectDirSHA256=row.Provenance.ProjectDirSHA256; want.C3DirSHA256=row.Provenance.C3DirSHA256; want.ContextThresholdAuthorityRecord=[]byte(authority.ContextThresholdAuthorityRecord)
        hp:=row.Provenance; hp.ContextThresholdAuthorityRecord=[]byte(authority.ContextThresholdAuthorityRecord); rp:=*durable.Report.Provenance; rp.ContextThresholdAuthorityRecord=[]byte(authority.ContextThresholdAuthorityRecord); if !reflect.DeepEqual(hp,want) || !reflect.DeepEqual(rp,want) || validateBaselineAdmission(bench,want)!=nil { return evalReport{},nil,errors.New("provenance_invalid") }
        runtimePath:=fmt.Sprintf("runtime/%02d.stderr",index+1); runtimeBytes,err:=gateRead(root,runtimePath,authority.ScanCaps.SingleDurableArtifactBytesMax); if err!=nil { return evalReport{},nil,errors.New("runtime_invalid") }
        if scanner.Scan("runtime_stderr",runtimePath,runtimeBytes)!=nil || scanner.Scan("result",run.ResultPath,resultBytes)!=nil || scanner.Scan("report",run.ReportPath,reportBytes)!=nil { return evalReport{},nil,errors.New("privacy_violation") }
        known[runtimePath]=runtimeBytes; known[run.ResultPath]=resultBytes; known[run.ReportPath]=reportBytes
        actuals=append(actuals,map[string]any{"order":index+1,"mode":mode,"case_id":caseID,"actual_budget":run.ActualBudget})
        if mode==corpusCombined { combined=durable.Report }
    }
    historyBytes,err:=gateRead(root,output.HistoryPath,authority.ScanCaps.SingleDurableArtifactBytesMax); if err!=nil || shaString(string(historyBytes))!=output.HistorySHA256 || scanner.Scan("history",output.HistoryPath,historyBytes)!=nil { return evalReport{},nil,errors.New("history_invalid") }; known[output.HistoryPath]=historyBytes
    manifestBytes,err:=gateRead(root,output.PrivacyManifestPath,authority.ScanCaps.SingleDurableArtifactBytesMax); if err!=nil || shaString(string(manifestBytes))!=output.PrivacyManifestSHA256 { return evalReport{},nil,errors.New("privacy_manifest_invalid") }; var manifest privacyManifest; if decodeStrictBytes(manifestBytes,&manifest)!=nil { return evalReport{},nil,errors.New("privacy_manifest_invalid") }
    entries:=append([]privacyArtifact(nil),scanner.entries...); sort.Slice(entries,func(i,j int)bool{return privacyArtifactLess(entries[i],entries[j])}); var total int64; for _,entry:=range entries { total+=int64(entry.Bytes) }
    if manifest.Schema!="structural-retrieval-privacy-scan.v1" || manifest.PrivacyPolicySHA256!=authority.PrivacyPolicySHA256 || manifest.PrivacyTermCount!=authority.PrivacyTermCount || manifest.DetectorVersion!=privacyDetectorVersion || manifest.DetectorSHA256!=authority.PrivacyDetectorSHA256 || manifest.ScanCaps!=authority.ScanCaps || manifest.Hits!=0 || manifest.SourceObjectCount!=scanner.sourceObjectCount || manifest.SourceObjectBytes!=scanner.sourceObjectBytes || manifest.ArtifactCount!=len(entries) || manifest.TotalScannedBytes!=total || !reflect.DeepEqual(manifest.Artifacts,entries) { return evalReport{},nil,errors.New("privacy_manifest_invalid") }
    if gateRootCoverage(root,output)!=nil { return evalReport{},nil,errors.New("root_coverage_invalid") }
    if combined.Schema=="" { return evalReport{},nil,errors.New("combined_report_missing") }
    _=known; _=authorityBytes; _=candidateSourceRoot; _=fixturesPath; _=benchmarkPath; _=scorerPath
    return combined,actuals,nil
}

func gateCombined(root string, output controllerOutput, fixtures []fixtureCase, bench benchmarkConfig, authority controllerAuthorityV4) (evalReport,error) {
    for _,run:=range output.Runs { if run.Mode!=corpusCombined || run.CaseID!="" { continue }; resultBytes,err:=gateRead(root,run.ResultPath,authority.ScanCaps.SingleDurableArtifactBytesMax); if err!=nil || shaString(string(resultBytes))!=run.ResultSHA256 { return evalReport{},errors.New("result_invalid") }; reportBytes,err:=gateRead(root,run.ReportPath,authority.ScanCaps.SingleDurableArtifactBytesMax); if err!=nil || shaString(string(reportBytes))!=run.ReportSHA256 { return evalReport{},errors.New("report_invalid") }; var response armResponse; var durable durableReport; if decodeStrictBytes(resultBytes,&response)!=nil || decodeStrictBytes(reportBytes,&durable)!=nil || verifyReportWithProbes(fixtures,response,durable.Database,corpusCombined,durable.DirectProbes,durable.Report)!=nil { return evalReport{},errors.New("score_replay_failed") }; return durable.Report,nil }
    return evalReport{},errors.New("combined_report_missing")
}

func TestProtocolV7CandidateGateReplay(t *testing.T) {
    raw, err := os.ReadFile(os.Getenv("C3_PROTOCOL_V7_CANDIDATE_GATE_CONFIG")); if err != nil { t.Fatal("config_invalid") }
    var cfg candidateGateConfig; if decodeStrictBytes(raw, &cfg) != nil { t.Fatal("config_invalid") }
    authorityBytes, err := os.ReadFile(cfg.Authority); if err != nil { t.Fatal("authority_invalid") }
    var authority controllerAuthorityV4; if decodeStrictBytes(authorityBytes, &authority) != nil { t.Fatal("authority_invalid") }
    fixturesPath := cfg.BRoot+"/research/eval/structural-retrieval/fixtures.dev.v2.jsonl"; benchmarkPath := cfg.BRoot+"/research/eval/structural-retrieval/benchmark.v2.json"; scorer := cfg.BRoot+"/cli/tools/structural-search-eval-v2/main.go"
    fixtures, fixtureHash, err := loadFixtures(fixturesPath); if err != nil { t.Fatal("fixture_invalid") }
    bench, err := loadBenchmark(benchmarkPath); if err != nil || bench.FixtureSHA256 != fixtureHash || bench.FixtureCount != len(fixtures) { t.Fatal("benchmark_invalid") }
    policyBytes, err := os.ReadFile(cfg.Policy); if err != nil { t.Fatal("privacy_invalid") }; policy, err := decodePrivacyPolicy(bytes.NewReader(policyBytes)); if err != nil { t.Fatal("privacy_invalid") }; scanner, err := newPrivacyScanner(policy); if err != nil { t.Fatal("privacy_invalid") }
    parentFiles := parentBaselineFiles{Root:cfg.ParentRoot, Authority:cfg.ParentAuthority, Output:cfg.ParentOutput, ValidatorStore:cfg.ParentStore}
    if err := verifyControllerAuthorityV4(authority, authorityBytes, cfg.Runtime, cfg.Controller, cfg.BRoot, cfg.CRoot, cfg.Bundle, cfg.Policy, fixturesPath, benchmarkPath, scorer, bench, policy, scanner, parentFiles); err != nil { t.Fatal(err) }
    candidateOutputBytes, err := os.ReadFile(filepath.Join(cfg.CandidateRoot,"controller-output.v4.json")); if err != nil { t.Fatal("candidate_output_invalid") }; var candidateOutput controllerOutput; if decodeStrictBytes(candidateOutputBytes,&candidateOutput)!=nil || candidateOutput.Schema!="structural-retrieval-controller-output.v4" || candidateOutput.Admitted || candidateOutput.Admission!="diagnostic_unadmitted" || candidateOutput.Failure!=nil || len(candidateOutput.Runs)!=len(fixtures)+2 || candidateOutput.OrderedRunManifestSHA256!=canonicalSHA256(candidateOutput.Runs) { t.Fatal("candidate_output_invalid") }
    historyBytes,err:=os.ReadFile(filepath.Join(cfg.CandidateRoot,candidateOutput.HistoryPath)); if err!=nil || shaString(string(historyBytes))!=candidateOutput.HistorySHA256 { t.Fatal("history_invalid") }; history,err:=decodeHistoryBytes(historyBytes); if err!=nil || len(history)!=len(candidateOutput.Runs) || verifyHistorySchema(history,historyV4Schema)!=nil { t.Fatal("history_invalid") }
    parentOutputBytes, err := os.ReadFile(cfg.ParentOutput); if err != nil { t.Fatal("parent_output_invalid") }; var parentOutput controllerOutput; if decodeStrictBytes(parentOutputBytes,&parentOutput)!=nil { t.Fatal("parent_output_invalid") }
    candidateReport,actuals,err := gateReplayCandidate(cfg.CandidateRoot,candidateOutput,history,fixtures,bench,authority,authorityBytes,cfg.CRoot,fixturesPath,benchmarkPath,scorer,scanner); if err != nil { t.Fatal(err) }
    var parentAuthority controllerAuthorityV4; parentAuthorityBytes, _ := os.ReadFile(cfg.ParentAuthority); if decodeStrictBytes(parentAuthorityBytes,&parentAuthority)!=nil { t.Fatal("parent_authority_invalid") }
    parentReport, err := gateCombined(cfg.ParentRoot,parentOutput,fixtures,bench,parentAuthority); if err != nil { t.Fatal(err) }
    verdict := evaluateGate(parentReport,candidateReport,bench)
    regression := 0; if candidateReport.Metrics.RelationshipRouteRecallAt5 < parentReport.Metrics.RelationshipRouteRecallAt5 || candidateReport.Metrics.RelationshipRouteMRR < parentReport.Metrics.RelationshipRouteMRR || candidateReport.Metrics.WrongLayerMRR < parentReport.Metrics.WrongLayerMRR { regression=1 }
    status := "discard"; if verdict.WouldPassIfRatified && candidateReport.Metrics.ForbiddenStructuralRetrievalCount==0 && regression==0 { status="keep" }
    mainHash:=shaString(string(mustReadFile(scorer))); testHash:=shaString(string(mustReadFile(filepath.Join(filepath.Dir(scorer),"main_test.go"))))
    result := map[string]any{"$schema":"structural-retrieval-protocol-v7-candidate-gate-harness.v1","status":"accepted","candidate_status":status,"microbench_structural_owner_recall_delta":verdict.RecallDelta,"blocking_false_structural_claim_count":candidateReport.Metrics.ForbiddenStructuralRetrievalCount,"strong_blast_radius_recall_regression":regression,"run_count":len(candidateOutput.Runs),"run_actuals":actuals,"validated_source_main_sha256":mainHash,"validated_source_test_sha256":testHash}
    encoded, _ := json.Marshal(result); if err := os.WriteFile(cfg.Result,encoded,0o600); err != nil { t.Fatal("result_invalid") }
}

func mustReadFile(path string) []byte { data, err := os.ReadFile(path); if err != nil { panic(err) }; return data }
'''


def _run_gate_harness(config: adapter.CandidateConfig, candidate_root: Path) -> dict[str, Any]:
    go = adapter._resolve_tool(config.go_executable, "go", adapter.ACCEPTED_GO_EXECUTABLE_SHA256)
    git = adapter._resolve_tool(config.git_executable, "git", adapter.ACCEPTED_GIT_EXECUTABLE_SHA256)
    parent = adapter.validate_parent(config)
    external_before = {
        "policy": adapter.sha256_file(config.privacy_policy, error="governed_input_invalid"),
        "go": adapter.sha256_file(go, error="tool_identity_invalid"),
        "git": adapter.sha256_file(git, error="tool_identity_invalid"),
        "adapter": adapter._adapter_sha256(),
        "validator": adapter._validator_sha256(),
    }
    parent_before = adapter._relative_manifest(config.parent_baseline_root)
    with tempfile.TemporaryDirectory(prefix="c3-v7-candidate-gate-") as raw:
        temporary = Path(raw); temporary.chmod(0o700)
        bundle = candidate_root.parent / "freeze/source.bundle"
        authority_path = candidate_root.parent / "freeze/candidate-authority.v4.json"
        authority = adapter.decode_canonical(adapter.read_regular(authority_path, mode=0o600, cap=adapter.MAX_METADATA_BYTES, error="candidate_authority_invalid"), "candidate_authority_invalid")
        delta = authority["candidate_delta"]
        b_root, c_root = temporary / "B", temporary / "C"
        adapter._checkout_bundle(git, bundle, b_root, adapter.ACCEPTED_PARENT_COMMIT)
        adapter._checkout_bundle(git, bundle, c_root, delta["candidate_commit"])
        harness_root = temporary / "harness"
        # Compile only accepted B. Candidate C remains inert data.
        adapter._copy_module_tree(b_root / "cli", harness_root / "cli")
        harness = harness_root / "cli/tools/structural-search-eval-v2/protocol_v7_candidate_gate_test.go"
        harness.write_text(GATE_HARNESS, encoding="utf-8")
        result_path = temporary / "gate-result.json"
        harness_config = {
            "authority": str(authority_path), "controller": str(candidate_root.parent / "freeze/controller-runtime"), "runtime": str(candidate_root.parent / "freeze/candidate-runtime"),
            "b_root": str(b_root), "c_root": str(c_root), "bundle": str(bundle), "policy": str(config.privacy_policy),
            "parent_root": str(parent.root / "parent"), "parent_authority": str(parent.root / "parent/controller-authority.v4.json"),
            "parent_output": str(parent.root / "parent/controller-output.v4.json"), "parent_store": str(parent.validator_store),
            "candidate_root": str(candidate_root), "result": str(result_path),
        }
        config_path = temporary / "gate-config.json"; config_path.write_bytes(canonical_json(harness_config)); config_path.chmod(0o600)
        tool_root = temporary / "tools"; tool_root.mkdir(mode=0o700)
        for source, name, digest in ((go, "go", adapter.ACCEPTED_GO_EXECUTABLE_SHA256), (git, "git", adapter.ACCEPTED_GIT_EXECUTABLE_SHA256)):
            target = tool_root / name; shutil.copyfile(source, target); target.chmod(0o700)
            if adapter.sha256_file(target, error="tool_identity_invalid") != digest:
                raise CandidateValidatorError("tool_identity_invalid")
        home = Path.home()
        env = {
            "PATH": str(tool_root), "HOME": str(home), "TMPDIR": str(temporary), "GOROOT": str(go.parent.parent), "LC_ALL": "C", "LANG": "C", "TZ": "UTC",
            "GOENV": "off", "GOWORK": "off", "GOFLAGS": "", "GOTOOLCHAIN": "local", "CGO_ENABLED": "0",
            "GOPROXY": "off", "GOSUMDB": "off", "GONOSUMDB": "", "GOPRIVATE": "",
            "GOMODCACHE": str(home / "go/pkg/mod"), "GOCACHE": str(Path(os.environ.get("XDG_CACHE_HOME", str(home / ".cache"))) / "go-build"),
            "GIT_CONFIG_NOSYSTEM": "1", "GIT_CONFIG_GLOBAL": "/dev/null", "GIT_OPTIONAL_LOCKS": "0", "GIT_NO_REPLACE_OBJECTS": "1", "GIT_ALTERNATE_OBJECT_DIRECTORIES": "", "GIT_ATTR_NOSYSTEM": "1",
            "C3_PROTOCOL_V7_CANDIDATE_GATE_CONFIG": str(config_path),
        }
        completed = adapter.run_bounded_process([str(tool_root / "go"), "test", "-count=1", "-run", "^TestProtocolV7CandidateGateReplay$", "./tools/structural-search-eval-v2"], cwd=harness_root / "cli", env=env, timeout=900, output_cap=adapter.PROCESS_OUTPUT_CAP)
        if completed.returncode != 0 or completed.stderr:
            raise CandidateValidatorError("gate_replay_failed")
        accepted = validate_gate_harness_result(adapter.read_regular(result_path, mode=0o600, cap=adapter.MAX_METADATA_BYTES, error="gate_replay_failed"))
    external_after = {
        "policy": adapter.sha256_file(config.privacy_policy, error="governed_input_invalid"),
        "go": adapter.sha256_file(go, error="tool_identity_invalid"),
        "git": adapter.sha256_file(git, error="tool_identity_invalid"),
        "adapter": adapter._adapter_sha256(),
        "validator": adapter._validator_sha256(),
    }
    if external_after != external_before or adapter._relative_manifest(config.parent_baseline_root) != parent_before:
        raise CandidateValidatorError("governed_input_changed")
    return accepted


def validate_capability(config: adapter.CandidateConfig) -> dict[str, Any]:
    before = validate_capability_layout(config.capability_root)
    authority = adapter.decode_canonical(
        adapter.read_regular(config.capability_root / "candidate-authority.v4.json", mode=0o600, cap=adapter.MAX_METADATA_BYTES, error="candidate_authority_invalid"),
        "candidate_authority_invalid",
    )
    manifest = validate_capability_manifest(
        adapter.read_regular(config.capability_root / "capability-manifest.json", mode=0o600, cap=adapter.MAX_METADATA_BYTES, error="capability_manifest_invalid"),
        {key: value for key, value in before.items() if key != "capability-manifest.json"},
        authority,
    )
    try:
        result = adapter.validate_capability(config)
    except adapter.CandidateAdapterError as exc:
        raise CandidateValidatorError(str(exc)) from exc
    after = validate_capability_layout(config.capability_root)
    if before != after or result.get("status") != "accepted_unexecuted":
        raise CandidateValidatorError("capability_changed")
    return {"$schema": RESULT_SCHEMA, "status": "capability_accepted", "effect_claim": False, "candidate_execution_authorized": False, "candidate_delta_sha256": manifest["candidate_delta_sha256"], "capability_manifest_sha256": before["capability-manifest.json"]}


def validate_result(config: adapter.CandidateConfig) -> dict[str, Any]:
    root = config.output_root
    controller_output_path = root / "candidate/controller-output.v4.json"
    try:
        output = adapter.decode_strict_object(adapter.read_regular(controller_output_path, mode=0o600, cap=adapter.MAX_METADATA_BYTES, error="candidate_output_invalid").rstrip(b"\n"), "candidate_output_invalid")
    except adapter.CandidateAdapterError as exc:
        raise CandidateValidatorError(str(exc)) from exc
    expected = expected_result_artifacts(output)
    before = snapshot_result(root, expected)
    _, privacy_terms = adapter.validate_privacy_policy(config.privacy_policy)
    manifest = validate_capture_manifest_bytes(
        adapter.read_regular(root / "candidate-capture-manifest.json", mode=0o600, cap=adapter.MAX_METADATA_BYTES, error="candidate_output_invalid"),
        privacy_terms,
    )
    scan_retained_metadata(root, expected, privacy_terms)
    authority = adapter.decode_canonical(
        adapter.read_regular(root / "freeze/candidate-authority.v4.json", mode=0o600, cap=adapter.MAX_METADATA_BYTES, error="candidate_output_invalid"),
        "candidate_output_invalid",
    )
    capability_manifest = validate_capability_manifest(
        adapter.read_regular(root / "freeze/capability-manifest.json", mode=0o600, cap=adapter.MAX_METADATA_BYTES, error="candidate_output_invalid"),
        {
            "candidate-authority.v4.json": before["freeze/candidate-authority.v4.json"],
            "freeze/controller-runtime": before["freeze/controller-runtime"],
            "freeze/candidate-runtime": before["freeze/candidate-runtime"],
            "freeze/source.bundle": before["freeze/source.bundle"],
        },
        authority,
    )
    delta = authority.get("candidate_delta")
    if not isinstance(delta, dict):
        raise CandidateValidatorError("candidate_output_invalid")
    delta_sha256 = adapter.candidate_delta_sha256(delta)
    if (
        manifest.get("$schema") != adapter.CAPTURE_MANIFEST_SCHEMA
        or manifest.get("status") != "captured_unvalidated"
        or manifest.get("effect_claim") is not False
        or manifest.get("max_capture_count") != 1
        or manifest.get("run_count") != adapter.ACCEPTED_PARENT_RUN_COUNT
        or manifest.get("candidate_execution_authorized") is not True
        or manifest.get("controller_output_sha256") != before["candidate/controller-output.v4.json"]
        or manifest.get("candidate_authority_sha256") != before["freeze/candidate-authority.v4.json"]
        or manifest.get("candidate_runtime_sha256") != before["freeze/candidate-runtime"]
        or manifest.get("bundle_sha256") != before["freeze/source.bundle"]
        or manifest.get("candidate_delta_sha256") != delta_sha256
        or manifest.get("candidate_delta_sha256") != capability_manifest.get("candidate_delta_sha256")
        or manifest.get("capability_manifest_sha256") != before["freeze/capability-manifest.json"]
        or not adapter.valid_sha256(manifest.get("authorization_record_sha256"))
        or not adapter.valid_sha256(manifest.get("activation_proof_sha256"))
        or manifest.get("adapter_sha256") != adapter._adapter_sha256()
        or manifest.get("validator_sha256") != adapter._validator_sha256()
        or manifest.get("main_sha256") != ACCEPTED_MAIN_SHA256
        or manifest.get("main_test_sha256") != ACCEPTED_MAIN_TEST_SHA256
        or manifest.get("scorer_region_sha256") != adapter.ACCEPTED_SCORER_REGION_SHA256
        or manifest.get("parent_authority_sha256") != adapter.ACCEPTED_PARENT_AUTHORITY_SHA256
        or manifest.get("parent_output_sha256") != adapter.ACCEPTED_PARENT_OUTPUT_SHA256
        or manifest.get("parent_validator_record_hash") != adapter.ACCEPTED_PARENT_VALIDATOR_RECORD_HASH
        or manifest.get("parent_cleanup_record_hash") != adapter.ACCEPTED_PARENT_CLEANUP_RECORD_HASH
        or manifest.get("privacy_policy_sha256") != adapter.validate_privacy_policy(config.privacy_policy)[0]
        or capability_manifest.get("future_output_root_sha256") != adapter.sha256_bytes(str(config.output_root.absolute()).encode())
        or capability_manifest.get("adapter_sha256") != manifest.get("adapter_sha256")
        or capability_manifest.get("validator_sha256") != manifest.get("validator_sha256")
    ):
        raise CandidateValidatorError("candidate_output_invalid")
    gate = _run_gate_harness(config, root / "candidate")
    after = snapshot_result(root, expected)
    if before != after:
        raise CandidateValidatorError("candidate_output_changed")
    result = {"$schema": RESULT_SCHEMA, "status": "accepted", "effect_claim": False, "candidate_execution_authorized": False, **{key: value for key, value in gate.items() if key not in {"$schema", "status"}}, "candidate_output_manifest_sha256": before["candidate-capture-manifest.json"]}
    if not adapter.is_generic_result(result):
        raise CandidateValidatorError("result_invalid")
    return result


def _parse(argv: list[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("command", choices=("validate-capability", "validate-result"))
    parser.add_argument("--parent-baseline-root", type=Path, required=True)
    parser.add_argument("--parent-validator-store", type=Path, required=True)
    parser.add_argument("--parent-validator-ref", required=True)
    parser.add_argument("--privacy-policy", type=Path, required=True)
    parser.add_argument("--capability-root", type=Path)
    parser.add_argument("--candidate-output-root", type=Path)
    parser.add_argument("--go-executable", type=Path)
    parser.add_argument("--git-executable", type=Path)
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = _parse(argv)
    capability = args.capability_root or (args.candidate_output_root / "freeze" if args.candidate_output_root else None)
    output = args.candidate_output_root or Path(str(capability) + ".future")
    if capability is None:
        print(canonical_json(rejected_result("missing capability")).decode()); return 1
    config = adapter.CandidateConfig(args.parent_baseline_root, args.parent_validator_store, args.parent_validator_ref, args.privacy_policy, capability, output, go_executable=args.go_executable, git_executable=args.git_executable)
    try:
        result = validate_capability(config) if args.command == "validate-capability" else validate_result(config)
    except (CandidateValidatorError, adapter.CandidateAdapterError):
        print(canonical_json(rejected_result("validation failed")).decode()); return 1
    print(canonical_json(result).decode()); return 0


if __name__ == "__main__":
    raise SystemExit(main())
