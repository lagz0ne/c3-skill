#!/usr/bin/env python3
"""Prepare and run generic protocol-v7 structural-retrieval transactions."""

from __future__ import annotations

import argparse
from dataclasses import dataclass
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import signal
import stat
import subprocess
import sys
import tempfile
from typing import Any, Callable


ACCEPTED_MAIN_SHA256 = "33b6452a0e34d825ef2b895609daa77b91eb129ebdc1e6be8f7f7755fc8b9c6e"
ACCEPTED_MAIN_TEST_SHA256 = "0830a0671aa9eb146902dc6fa56e885126b89a3312ca6c38fd8507fa10481dbf"
ACCEPTED_FIXTURE_SHA256 = "15f57120c6aa9ae07bf4fdacd6ad783afa5e70ed8ebebaff3a42dcf4249e677e"
ACCEPTED_BENCHMARK_SHA256 = "b960525cc42216e6598452946da5fb68735bbf989f311f170cedcfdbe92bf0d5"
ACCEPTED_SCORER_REGION_SHA256 = "c1669d43b13a36eb1454a01a5ea7e444fe67ce28fdae9628da083927f093cbcb"

PRIVACY_POLICY_SCHEMA = "structural-retrieval-privacy-policy.v1"
INPUT_MANIFEST_SCHEMA = "structural-retrieval-protocol-v7-adapter-inputs.v1"
ACTIVATION_SCHEMA = "structural-retrieval-protocol-v7-capture-activation.v1"
AUTHORIZATION_SCHEMA = "structural-retrieval-protocol-v7-baseline-authorization.v1"
RESULT_SCHEMA = "structural-retrieval-protocol-v7-adapter-result.v1"
MAX_POLICY_BYTES = 128 << 10
MAX_POLICY_TERMS = 512
MAX_POLICY_TERM_BYTES = 64 << 10


# The driver is compiled as package main beside the exact accepted evaluator.
# It is never copied into the generated B repository or its bundle.
DRIVER_SOURCE = r'''package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

type protocolV7DriverConfig struct {
	Command                    string `json:"command"`
	SourceRoot                 string `json:"source_root"`
	MainPath                   string `json:"main_path"`
	MainTestPath               string `json:"main_test_path"`
	FixturePath                string `json:"fixture_path"`
	BenchmarkPath              string `json:"benchmark_path"`
	ThresholdCheckinsPath      string `json:"threshold_checkins_path"`
	ThresholdCheckinSeq        int    `json:"threshold_checkin_seq"`
	PrivacyPolicyPath          string `json:"privacy_policy_path"`
	OutputRoot                 string `json:"output_root"`
	StageRoot                  string `json:"stage_root"`
	ScratchRoot                string `json:"scratch_root"`
	ResultPath                 string `json:"result_path"`
	ExpectedMainSHA256         string `json:"expected_main_sha256"`
	ExpectedMainTestSHA256     string `json:"expected_main_test_sha256"`
	ExpectedFixtureSHA256      string `json:"expected_fixture_sha256"`
	ExpectedBenchmarkSHA256    string `json:"expected_benchmark_sha256"`
	ExpectedScorerRegionSHA256 string `json:"expected_scorer_region_sha256"`
	AdapterSHA256              string `json:"adapter_sha256"`
	DriverSHA256               string `json:"driver_sha256"`
	InputManifestSHA256        string `json:"input_manifest_sha256"`
	CaptureAuthorized          bool   `json:"capture_authorized"`
	ActivationProofSHA256      string `json:"activation_proof_sha256"`
	ActivationRecordHash       string `json:"activation_record_hash"`
	ActivationEvidencePath     string `json:"activation_evidence_path"`
	ActivationOriginalPath     string `json:"activation_original_path"`
	ActivationReceiptPath      string `json:"activation_receipt_path"`
	AuthorizationEvidencePath  string `json:"authorization_evidence_path"`
	AuthorizationRecordPath    string `json:"authorization_record_path"`
	AuthorizationRecordSHA256  string `json:"authorization_record_sha256"`
}

type protocolV7AuthorizationRecord struct {
	Seq           int             `json:"seq"`
	RecordedAt    string          `json:"recorded_at"`
	PrevHash      string          `json:"prev_hash"`
	PayloadSHA256 string          `json:"payload_sha256"`
	Payload       json.RawMessage `json:"payload"`
	RecordHash    string          `json:"record_hash"`
}

func TestProtocolV7CaptureDriver(t *testing.T) {
	if err := protocolV7CaptureDriver(); err != nil {
		t.Fatal(err)
	}
}

func protocolV7ReadPrivateEvidence(path string, maxBytes int64) ([]byte, error) {
	data, err := readBoundedStandaloneRegularFile(path, maxBytes)
	if err != nil {
		return nil, errors.New("authorization_invalid")
	}
	info, err := os.Lstat(path)
	if err != nil || info.Mode().Perm() != 0o600 {
		return nil, errors.New("authorization_invalid")
	}
	return data, nil
}

func protocolV7ValidateAuthorizationEvidence(cfg protocolV7DriverConfig, policy privacyPolicy) error {
	if _, err := os.Lstat(cfg.ActivationOriginalPath); err == nil || !errors.Is(err, os.ErrNotExist) {
		return errors.New("activation_invalid")
	}
	activationBytes, err := protocolV7ReadPrivateEvidence(cfg.ActivationEvidencePath, 64<<10)
	if err != nil {
		return err
	}
	receiptBytes, err := protocolV7ReadPrivateEvidence(cfg.ActivationReceiptPath, 64<<10)
	if err != nil || !bytes.Equal(receiptBytes, activationBytes) || shaString(string(activationBytes)) != cfg.ActivationProofSHA256 {
		return errors.New("activation_invalid")
	}
	authorizationBytes, err := protocolV7ReadPrivateEvidence(cfg.AuthorizationEvidencePath, 64<<10)
	if err != nil {
		return err
	}
	authorizationSourceBytes, err := protocolV7ReadPrivateEvidence(cfg.AuthorizationRecordPath, 64<<10)
	if err != nil || !bytes.Equal(authorizationSourceBytes, authorizationBytes) || shaString(string(authorizationBytes)) != cfg.AuthorizationRecordSHA256 {
		return errors.New("authorization_invalid")
	}
	if len(authorizationBytes) < 2 || authorizationBytes[len(authorizationBytes)-1] != '\n' || bytes.Count(authorizationBytes, []byte{'\n'}) != 1 {
		return errors.New("authorization_invalid")
	}
	var record protocolV7AuthorizationRecord
	if err := decodeStrictBytes(authorizationBytes[:len(authorizationBytes)-1], &record); err != nil {
		return errors.New("authorization_invalid")
	}
	parsedTime, err := time.Parse(time.RFC3339, record.RecordedAt)
	if err != nil || parsedTime.UTC().Format("2006-01-02T15:04:05Z") != record.RecordedAt || record.Seq != 1 || record.PrevHash != "GENESIS" {
		return errors.New("authorization_invalid")
	}
	expectedPayload := map[string]any{
		"$schema": "structural-retrieval-protocol-v7-baseline-authorization.v1",
		"activation_path_sha256": shaString(cfg.ActivationOriginalPath),
		"adapter_sha256": cfg.AdapterSHA256,
		"baseline_capture_authorized": true,
		"benchmark_sha256": cfg.ExpectedBenchmarkSHA256,
		"candidate_execution_authorized": false,
		"driver_sha256": cfg.DriverSHA256,
		"effect_claim": false,
		"fixture_sha256": cfg.ExpectedFixtureSHA256,
		"main_sha256": cfg.ExpectedMainSHA256,
		"main_test_sha256": cfg.ExpectedMainTestSHA256,
		"max_capture_count": 1,
		"output_root_sha256": shaString(cfg.OutputRoot),
		"privacy_policy_sha256": policy.SHA256,
		"scorer_region_sha256": cfg.ExpectedScorerRegionSHA256,
		"verdict": "authorized",
	}
	expectedPayloadBytes, _ := json.Marshal(expectedPayload)
	if !bytes.Equal(record.Payload, expectedPayloadBytes) || record.PayloadSHA256 != shaString(string(expectedPayloadBytes)) {
		return errors.New("authorization_invalid")
	}
	withoutHash := map[string]any{
		"payload": json.RawMessage(expectedPayloadBytes),
		"payload_sha256": record.PayloadSHA256,
		"prev_hash": record.PrevHash,
		"recorded_at": record.RecordedAt,
		"seq": record.Seq,
	}
	withoutHashBytes, _ := json.Marshal(withoutHash)
	if record.RecordHash != shaString(string(withoutHashBytes)) || record.RecordHash != cfg.ActivationRecordHash {
		return errors.New("authorization_invalid")
	}
	expectedRecord := map[string]any{
		"payload": json.RawMessage(expectedPayloadBytes),
		"payload_sha256": record.PayloadSHA256,
		"prev_hash": record.PrevHash,
		"record_hash": record.RecordHash,
		"recorded_at": record.RecordedAt,
		"seq": record.Seq,
	}
	expectedRecordBytes, _ := json.Marshal(expectedRecord)
	expectedRecordBytes = append(expectedRecordBytes, '\n')
	if !bytes.Equal(authorizationBytes, expectedRecordBytes) {
		return errors.New("authorization_invalid")
	}
	expectedActivation := map[string]any{
		"$schema": "structural-retrieval-protocol-v7-capture-activation.v1",
		"activation_path_sha256": shaString(cfg.ActivationOriginalPath),
		"activation_record_hash": record.RecordHash,
		"adapter_sha256": cfg.AdapterSHA256,
		"authorization_record_path_sha256": shaString(cfg.AuthorizationRecordPath),
		"benchmark_sha256": cfg.ExpectedBenchmarkSHA256,
		"driver_sha256": cfg.DriverSHA256,
		"fixture_sha256": cfg.ExpectedFixtureSHA256,
		"main_sha256": cfg.ExpectedMainSHA256,
		"main_test_sha256": cfg.ExpectedMainTestSHA256,
		"max_capture_count": 1,
		"output_root_sha256": shaString(cfg.OutputRoot),
		"privacy_policy_sha256": policy.SHA256,
		"scorer_region_sha256": cfg.ExpectedScorerRegionSHA256,
		"verdict": "authorized",
	}
	expectedActivationBytes, _ := json.Marshal(expectedActivation)
	if !bytes.Equal(activationBytes, expectedActivationBytes) {
		return errors.New("activation_invalid")
	}
	return nil
}

func protocolV7CaptureDriver() error {
	configPath := os.Getenv("C3_PROTOCOL_V7_DRIVER_CONFIG")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return errors.New("driver_config_invalid")
	}
	var cfg protocolV7DriverConfig
	if err := decodeStrictBytes(data, &cfg); err != nil {
		return errors.New("driver_config_invalid")
	}
	result, err := protocolV7RunCapture(cfg)
	if err != nil {
		return err
	}
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return errors.New("driver_result_invalid")
	}
	if err := os.WriteFile(cfg.ResultPath, resultBytes, 0o600); err != nil {
		return errors.New("driver_result_invalid")
	}
	return nil
}

func protocolV7RunCapture(cfg protocolV7DriverConfig) (map[string]any, error) {
	if cfg.Command != "prepare-baseline" && cfg.Command != "self-test" && cfg.Command != "capture-baseline" && cfg.Command != "verify-activation" {
		return nil, errors.New("driver_command_invalid")
	}
	requiresActivation := cfg.Command == "capture-baseline" || cfg.Command == "verify-activation"
	if requiresActivation && (!cfg.CaptureAuthorized || !validSHA256(cfg.ActivationProofSHA256) || !validSHA256(cfg.ActivationRecordHash) || !validSHA256(cfg.AuthorizationRecordSHA256) || cfg.ActivationEvidencePath == "" || cfg.ActivationOriginalPath == "" || cfg.ActivationReceiptPath == "" || cfg.AuthorizationEvidencePath == "" || cfg.AuthorizationRecordPath == "") {
		return nil, errors.New("activation_invalid")
	}
	if !requiresActivation && (cfg.CaptureAuthorized || cfg.ActivationProofSHA256 != "" || cfg.ActivationRecordHash != "" || cfg.ActivationEvidencePath != "" || cfg.ActivationOriginalPath != "" || cfg.ActivationReceiptPath != "" || cfg.AuthorizationEvidencePath != "" || cfg.AuthorizationRecordPath != "" || cfg.AuthorizationRecordSHA256 != "") {
		return nil, errors.New("activation_invalid")
	}
	exact := []struct{ path, relative, hash string }{
		{cfg.MainPath, "cli/tools/structural-search-eval-v2/main.go", cfg.ExpectedMainSHA256},
		{cfg.MainTestPath, "cli/tools/structural-search-eval-v2/main_test.go", cfg.ExpectedMainTestSHA256},
		{cfg.FixturePath, "research/eval/structural-retrieval/fixtures.dev.v2.jsonl", cfg.ExpectedFixtureSHA256},
		{cfg.BenchmarkPath, "research/eval/structural-retrieval/benchmark.v2.json", cfg.ExpectedBenchmarkSHA256},
	}
	for _, input := range exact {
		if !registeredPathInside(cfg.SourceRoot, input.path, input.relative) {
			return nil, errors.New("frozen_input_invalid")
		}
		got, err := fileSHA256(input.path)
		if err != nil || got != input.hash {
			return nil, errors.New("frozen_input_invalid")
		}
	}
	region, err := scoringRegionSHA256(cfg.MainPath)
	if err != nil || region != cfg.ExpectedScorerRegionSHA256 {
		return nil, errors.New("frozen_input_invalid")
	}
	policyBytes, err := readBoundedStandaloneRegularFile(cfg.PrivacyPolicyPath, privacyPolicyBytesMax)
	if err != nil {
		return nil, errors.New("privacy_policy_invalid")
	}
	policy, err := decodePrivacyPolicy(bytes.NewReader(policyBytes))
	if err != nil {
		return nil, errors.New("privacy_policy_invalid")
	}
	if requiresActivation {
		if err := protocolV7ValidateAuthorizationEvidence(cfg, policy); err != nil {
			return nil, err
		}
	}
	if cfg.Command == "verify-activation" {
		return map[string]any{
			"$schema": "structural-retrieval-protocol-v7-adapter-result.v1",
			"status": "accepted_disposable_non_study",
			"activation_proof_sha256": cfg.ActivationProofSHA256,
			"activation_record_hash": cfg.ActivationRecordHash,
			"authorization_record_sha256": cfg.AuthorizationRecordSHA256,
		}, nil
	}
	if _, err := os.Lstat(cfg.OutputRoot); err == nil || !errors.Is(err, os.ErrNotExist) {
		return nil, errors.New("output_exists")
	}
	parent := filepath.Dir(cfg.OutputRoot)
	if info, err := os.Stat(parent); err != nil || !info.IsDir() {
		return nil, errors.New("output_parent_invalid")
	}
	stage := cfg.StageRoot
	stageInfo, err := os.Lstat(stage)
	if err != nil || !stageInfo.IsDir() || stageInfo.Mode()&os.ModeSymlink != 0 || stageInfo.Mode().Perm() != 0o700 || filepath.Dir(stage) != parent {
		return nil, errors.New("transaction_failed")
	}
	entries, err := os.ReadDir(stage)
	if err != nil || len(entries) != 0 {
		return nil, errors.New("transaction_failed")
	}
	stageKept := false
	defer func() {
		if !stageKept {
			_ = os.RemoveAll(stage)
		}
	}()
	if _, err := os.Lstat(cfg.ScratchRoot); err == nil || !errors.Is(err, os.ErrNotExist) {
		return nil, errors.New("transaction_failed")
	}
	if err := os.Mkdir(cfg.ScratchRoot, 0o700); err != nil {
		return nil, errors.New("transaction_failed")
	}
	defer os.RemoveAll(cfg.ScratchRoot)
	prep := filepath.Join(cfg.ScratchRoot, "prep")
	if err := os.Mkdir(prep, 0o700); err != nil {
		return nil, errors.New("transaction_failed")
	}
	prepared, err := protocolV7Prepare(cfg, policy, policyBytes, prep)
	if err != nil {
		return nil, err
	}
	manifest := map[string]any{
		"$schema": "structural-retrieval-protocol-v7-preparation.v1",
		"status": "accepted",
		"adapter_sha256": cfg.AdapterSHA256,
		"driver_sha256": cfg.DriverSHA256,
		"input_manifest_sha256": cfg.InputManifestSHA256,
		"main_sha256": cfg.ExpectedMainSHA256,
		"main_test_sha256": cfg.ExpectedMainTestSHA256,
		"fixture_sha256": cfg.ExpectedFixtureSHA256,
		"benchmark_sha256": cfg.ExpectedBenchmarkSHA256,
		"scorer_region_sha256": cfg.ExpectedScorerRegionSHA256,
		"privacy_policy_sha256": policy.SHA256,
		"privacy_term_count": len(policy.DenyTerms),
		"commit": prepared.commit,
		"tree": prepared.tree,
		"source_capsule_sha256": canonicalSHA256(prepared.capsule),
		"runtime_sha256": prepared.runtimeHash,
		"bundle_sha256": prepared.authority.Expected.BundleSHA256,
		"bundle_heads_sha256": prepared.authority.SourceBundleHeadsSHA256,
		"authority_sha256": shaString(string(prepared.authorityBytes)),
	}
	if cfg.Command == "capture-baseline" {
		manifest["activation_proof_sha256"] = cfg.ActivationProofSHA256
		manifest["activation_record_hash"] = cfg.ActivationRecordHash
		manifest["authorization_record_sha256"] = cfg.AuthorizationRecordSHA256
	}
	if cfg.Command == "prepare-baseline" {
		if err := protocolV7VerifyPreparation(prepared, cfg, policy); err != nil {
			return nil, err
		}
		if err := protocolV7WritePreparation(stage, prepared, manifest); err != nil {
			return nil, err
		}
		if err := protocolV7CommitStage(stage, cfg.OutputRoot); err != nil {
			return nil, err
		}
		stageKept = true
		return manifest, nil
	}
	output, outputBytes, err := protocolV7ExecuteAndReplay(stage, prepared, cfg, policy)
	if err != nil {
		return nil, err
	}
	manifest["controller_output_sha256"] = shaString(string(outputBytes))
	manifest["ordered_run_manifest_sha256"] = output.OrderedRunManifestSHA256
	manifest["run_count"] = len(output.Runs)
	manifest["history_sha256"] = output.HistorySHA256
	manifest["privacy_manifest_sha256"] = output.PrivacyManifestSHA256
	if err := protocolV7WriteFreeze(stage, prepared); err != nil {
		return nil, err
	}
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		return nil, errors.New("adapter_manifest_invalid")
	}
	manifestScanner, err := newPrivacyScanner(policy)
	if err != nil || manifestScanner.Scan("adapter_manifest", "capture-manifest.json", manifestBytes) != nil {
		return nil, errors.New("privacy_policy_violation")
	}
	if err := os.WriteFile(filepath.Join(stage, "capture-manifest.json"), manifestBytes, 0o600); err != nil {
		return nil, errors.New("transaction_failed")
	}
	if cfg.Command == "self-test" {
		if err := os.RemoveAll(stage); err != nil {
			return nil, errors.New("self_test_cleanup_failed")
		}
		if _, err := os.Lstat(stage); !errors.Is(err, os.ErrNotExist) {
			return nil, errors.New("self_test_cleanup_failed")
		}
		return map[string]any{
			"$schema": "structural-retrieval-protocol-v7-self-test.v1",
			"status": "accepted_disposable_non_study",
			"adapter_sha256": cfg.AdapterSHA256,
			"driver_sha256": cfg.DriverSHA256,
			"authority_sha256": manifest["authority_sha256"],
			"controller_output_sha256": manifest["controller_output_sha256"],
			"run_count": manifest["run_count"],
			"privacy_policy_sha256": policy.SHA256,
			"privacy_term_count": len(policy.DenyTerms),
		}, nil
	}
	if err := protocolV7CommitStage(stage, cfg.OutputRoot); err != nil {
		return nil, err
	}
	stageKept = true
	return manifest, nil
}

type protocolV7Prepared struct {
	root, runtime, bundle, authorityPath string
	commit, tree, runtimeHash string
	capsule sourceCapsule
	authority controllerAuthorityV4
	authorityBytes []byte
}

func protocolV7Prepare(cfg protocolV7DriverConfig, policy privacyPolicy, policyBytes []byte, root string) (protocolV7Prepared, error) {
	base := filepath.Join(root, "B")
	paths, err := discoverRepositoryBuildInputs(cfg.SourceRoot)
	if err != nil {
		return protocolV7Prepared{}, fmt.Errorf("build_input_discovery_failed: %v", err)
	}
	paths = append(paths,
		"research/eval/structural-retrieval/fixtures.dev.v2.jsonl",
		"research/eval/structural-retrieval/benchmark.v2.json",
		"cli/tools/structural-search-eval-v2/main_test.go",
	)
	sort.Strings(paths)
	previous := ""
	for _, relative := range paths {
		if relative == previous {
			continue
		}
		previous = relative
		source := filepath.Join(cfg.SourceRoot, filepath.FromSlash(relative))
		target := filepath.Join(base, filepath.FromSlash(relative))
		data, err := os.ReadFile(source)
		if err != nil {
			return protocolV7Prepared{}, errors.New("build_input_copy_failed")
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return protocolV7Prepared{}, errors.New("build_input_copy_failed")
		}
		if err := os.WriteFile(target, data, 0o644); err != nil || os.Chmod(target, 0o644) != nil {
			return protocolV7Prepared{}, errors.New("build_input_copy_failed")
		}
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "."}, {"commit", "-q", "--no-gpg-sign", "-m", "portable protocol-v7 baseline"}} {
		if _, err := protocolV7Git(base, args...); err != nil {
			return protocolV7Prepared{}, err
		}
	}
	commitBytes, err := protocolV7Git(base, "rev-parse", "HEAD")
	if err != nil {
		return protocolV7Prepared{}, err
	}
	treeBytes, err := protocolV7Git(base, "rev-parse", "HEAD^{tree}")
	if err != nil {
		return protocolV7Prepared{}, err
	}
	commit := strings.TrimSpace(string(commitBytes))
	tree := strings.TrimSpace(string(treeBytes))
	ref := "refs/c3-eval/commit-pool/protocol-v7-baseline"
	if _, err := protocolV7Git(base, "update-ref", ref, commit); err != nil {
		return protocolV7Prepared{}, err
	}
	bundle := filepath.Join(root, "source.bundle")
	if _, err := protocolV7Git(base, "bundle", "create", bundle, ref); err != nil {
		return protocolV7Prepared{}, err
	}
	runtime := filepath.Join(root, "controller-runtime")
	if err := buildFrozenRuntime(base, runtime); err != nil {
		return protocolV7Prepared{}, errors.New("runtime_build_failed")
	}
	runtimeHash, err := fileSHA256(runtime)
	if err != nil {
		return protocolV7Prepared{}, errors.New("runtime_build_failed")
	}
	capsule, err := captureSourceCapsule(base)
	if err != nil {
		return protocolV7Prepared{}, errors.New("source_capsule_failed")
	}
	benchmarkPath := filepath.Join(base, "research/eval/structural-retrieval/benchmark.v2.json")
	bench, err := loadBenchmark(benchmarkPath)
	if err != nil {
		return protocolV7Prepared{}, errors.New("benchmark_invalid")
	}
	detector, err := newGenericPrivacyDetector()
	if err != nil {
		return protocolV7Prepared{}, errors.New("privacy_detector_invalid")
	}
	environmentHash, err := environmentSHA256(base)
	if err != nil {
		return protocolV7Prepared{}, errors.New("environment_hash_failed")
	}
	moduleHash, err := moduleGraphSHA256(base)
	if err != nil {
		return protocolV7Prepared{}, errors.New("module_hash_failed")
	}
	goHash, err := goExecutableSHA256()
	if err != nil {
		return protocolV7Prepared{}, errors.New("go_identity_failed")
	}
	goVerifyHash, err := goModVerifySHA256(base)
	if err != nil {
		return protocolV7Prepared{}, errors.New("go_verify_failed")
	}
	line, err := protocolV7CheckinLine(cfg.ThresholdCheckinsPath, cfg.ThresholdCheckinSeq)
	if err != nil {
		return protocolV7Prepared{}, err
	}
	headsHash, _, err := gitSHA256(base, "bundle", "list-heads", bundle)
	if err != nil {
		return protocolV7Prepared{}, errors.New("bundle_heads_failed")
	}
	limits := resourceBudget{WallTimeMillis: registeredWallTimeMillis, CPUTimeMillis: 10_000, MaxRSSBytes: 536_870_912, ProcessCount: 16, SQLiteRowCount: 1_000_000, LogicalDumpBytes: 64 << 20, StdoutBytes: 16 << 20, StderrBytes: 1 << 20, CaseCount: 100}
	actionEnvelope := "protocol-v7 generic baseline transaction; no product writes"
	expected := provenance{
		ExperimentID: "protocol-v7-baseline", ArmID: "baseline", Commit: commit, Tree: tree,
		ControllerCommit: commit, ControllerTree: tree, SourceCapsuleSHA256: canonicalSHA256(capsule),
		ControllerSourceCapsuleSHA256: canonicalSHA256(capsule), DiffSHA256: shaString(""),
		FixtureSHA256: cfg.ExpectedFixtureSHA256, BenchmarkSHA256: cfg.ExpectedBenchmarkSHA256, ScorerSHA256: cfg.ExpectedMainSHA256,
		ControllerSHA256: runtimeHash, RuntimeSHA256: runtimeHash, EnvironmentSHA256: environmentHash, ModuleGraphSHA256: moduleHash,
		BudgetSHA256: canonicalSHA256(limits), ActionEnvelopeSHA256: shaString(actionEnvelope), SemanticMode: bench.SemanticMode,
		ContextThresholdAuthoritySHA256: hashThresholdAuthority(bench.ContextThresholdAuthority), BundleSHA256: mustProtocolV7FileHash(bundle),
		PrivacyPolicySHA256: policy.SHA256, PrivacyTermCount: len(policy.DenyTerms), PrivacyDetectorSHA256: detector.DefinitionSHA256,
		GoExecutableSHA256: goHash, GoModVerifySHA256: goVerifyHash, ScanCapsSHA256: canonicalSHA256(protocolV7ScanCaps()),
		SourceBundleHeadsSHA256: headsHash, ProtocolTestSHA256: cfg.ExpectedMainTestSHA256,
	}
	if expected.BundleSHA256 == "" {
		return protocolV7Prepared{}, errors.New("bundle_hash_failed")
	}
	authority := controllerAuthorityV4{
		Schema: controllerAuthorityV4Schema, Mode: "baseline", Expected: expected,
		ControllerSourceCapsule: capsule, RuntimeSourceCapsule: capsule,
		BuildReplay: buildReplayManifest{
			ControllerSHA256: runtimeHash, RuntimeSHA256: runtimeHash, RebuiltControllerSHA256: runtimeHash, RebuiltRuntimeSHA256: runtimeHash,
			ControllerCapsuleRebuildVerified: true, RuntimeCapsuleRebuildVerified: true, BundleVerified: true,
		},
		BudgetLimits: limits, ScanCaps: protocolV7ScanCaps(), PrivacyPolicySHA256: policy.SHA256, PrivacyTermCount: len(policy.DenyTerms),
		PrivacyDetectorSHA256: detector.DefinitionSHA256, GoExecutableSHA256: goHash, GoModVerifySHA256: goVerifyHash,
		SourceBundleHeadsSHA256: headsHash, ProtocolTestSHA256: cfg.ExpectedMainTestSHA256, ActionEnvelope: actionEnvelope,
		CanonicalRowBytesDefinition: canonicalRowBytesDefinition, ContextThresholdAuthorityRecord: string(line),
	}
	authorityBytes, err := json.Marshal(authority)
	if err != nil {
		return protocolV7Prepared{}, errors.New("authority_invalid")
	}
	authorityPath := filepath.Join(root, "controller-authority.v4.json")
	if err := os.WriteFile(authorityPath, authorityBytes, 0o600); err != nil {
		return protocolV7Prepared{}, errors.New("authority_invalid")
	}
	_ = policyBytes
	return protocolV7Prepared{root: base, runtime: runtime, bundle: bundle, authorityPath: authorityPath, commit: commit, tree: tree, runtimeHash: runtimeHash, capsule: capsule, authority: authority, authorityBytes: authorityBytes}, nil
}

func protocolV7VerifyPreparation(prepared protocolV7Prepared, cfg protocolV7DriverConfig, policy privacyPolicy) error {
	scanner, err := newPrivacyScanner(policy)
	if err != nil {
		return errors.New("privacy_policy_invalid")
	}
	fixtures := filepath.Join(prepared.root, "research/eval/structural-retrieval/fixtures.dev.v2.jsonl")
	benchmark := filepath.Join(prepared.root, "research/eval/structural-retrieval/benchmark.v2.json")
	scorer := filepath.Join(prepared.root, "cli/tools/structural-search-eval-v2/main.go")
	bench, err := loadBenchmark(benchmark)
	if err != nil {
		return errors.New("benchmark_invalid")
	}
	if err := verifyControllerAuthorityV4(prepared.authority, prepared.authorityBytes, prepared.runtime, prepared.runtime, prepared.root, prepared.root, prepared.bundle, cfg.PrivacyPolicyPath, fixtures, benchmark, scorer, bench, policy, scanner, parentBaselineFiles{}); err != nil {
		return errors.New("preparation_rejected")
	}
	return nil
}

func protocolV7ExecuteAndReplay(stage string, prepared protocolV7Prepared, cfg protocolV7DriverConfig, policy privacyPolicy) (controllerOutput, []byte, error) {
	parentRoot := filepath.Join(stage, "parent")
	fixturesPath := filepath.Join(prepared.root, "research/eval/structural-retrieval/fixtures.dev.v2.jsonl")
	benchmarkPath := filepath.Join(prepared.root, "research/eval/structural-retrieval/benchmark.v2.json")
	scorerPath := filepath.Join(prepared.root, "cli/tools/structural-search-eval-v2/main.go")
	workParent := filepath.Join(cfg.ScratchRoot, "work-parent")
	if err := os.Mkdir(workParent, 0o700); err != nil {
		return controllerOutput{}, nil, errors.New("controller_transaction_failed")
	}
	workRoot := filepath.Join(workParent, "caller-work-must-remain-absent")
	command := exec.Command(prepared.runtime, "--controller", "--runtime", prepared.runtime, "--fixtures", fixturesPath, "--benchmark", benchmarkPath,
		"--work-root", workRoot, "--authority", prepared.authorityPath, "--controller-source-root", prepared.root, "--runtime-source-root", prepared.root,
		"--bundle", prepared.bundle, "--privacy-policy", cfg.PrivacyPolicyPath, "--scorer-source", scorerPath, "--output-dir", parentRoot)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil || stderr.Len() != 0 {
		return controllerOutput{}, nil, fmt.Errorf("controller_transaction_failed: %s", strings.TrimSpace(stderr.String()))
	}
	var output controllerOutput
	if err := decodeStrictBytes(stdout.Bytes(), &output); err != nil {
		return controllerOutput{}, nil, errors.New("controller_output_invalid")
	}
	bench, err := loadBenchmark(benchmarkPath)
	if err != nil || output.Schema != "structural-retrieval-controller-output.v4" || output.Admitted || output.Admission != "diagnostic_unadmitted" || output.Failure != nil || len(output.Runs) != bench.FixtureCount+2 || output.OrderedRunManifestSHA256 != canonicalSHA256(output.Runs) {
		return controllerOutput{}, nil, errors.New("controller_output_invalid")
	}
	authorityTarget := filepath.Join(parentRoot, "controller-authority.v4.json")
	outputTarget := filepath.Join(parentRoot, "controller-output.v4.json")
	if err := os.WriteFile(authorityTarget, prepared.authorityBytes, 0o600); err != nil || os.WriteFile(outputTarget, stdout.Bytes(), 0o600) != nil {
		return controllerOutput{}, nil, errors.New("parent_seal_failed")
	}
	historyBytes, err := parentRelativeArtifact(parentRoot, output.HistoryPath, prepared.authority.ScanCaps.SingleDurableArtifactBytesMax)
	if err != nil || shaString(string(historyBytes)) != output.HistorySHA256 {
		return controllerOutput{}, nil, errors.New("parent_replay_failed")
	}
	history, err := decodeHistoryBytes(historyBytes)
	if err != nil || len(history) != len(output.Runs) || verifyHistorySchema(history, historyV4Schema) != nil {
		return controllerOutput{}, nil, errors.New("parent_replay_failed")
	}
	fixtures, fixtureHash, err := loadFixtures(fixturesPath)
	if err != nil || fixtureHash != prepared.authority.Expected.FixtureSHA256 {
		return controllerOutput{}, nil, errors.New("parent_replay_failed")
	}
	scanner, err := newPrivacyScanner(policy)
	if err != nil {
		return controllerOutput{}, nil, errors.New("parent_replay_failed")
	}
	known, err := verifyParentRunArtifacts(parentRoot, output, history, prepared.authority, fixtures, bench, scanner)
	if err != nil {
		return controllerOutput{}, nil, errors.New("parent_replay_failed")
	}
	known[output.HistoryPath] = historyBytes
	for index := range output.Runs {
		path := fmt.Sprintf("runtime/%02d.stderr", index+1)
		data, err := parentRelativeArtifact(parentRoot, path, prepared.authority.ScanCaps.SingleDurableArtifactBytesMax)
		if err != nil {
			return controllerOutput{}, nil, errors.New("parent_replay_failed")
		}
		known[path] = data
	}
	if err := verifyParentPrivacyManifest(parentRoot, output, prepared.authority, prepared.authorityBytes, known, prepared.root, fixturesPath, benchmarkPath, scorerPath, scanner); err != nil {
		return controllerOutput{}, nil, errors.New("parent_replay_failed")
	}
	files := parentBaselineFiles{Root: parentRoot, Authority: authorityTarget, Output: outputTarget}
	if err := verifyParentRootCoverage(files, output); err != nil {
		return controllerOutput{}, nil, errors.New("parent_replay_failed")
	}
	if _, err := os.Lstat(workRoot); !errors.Is(err, os.ErrNotExist) {
		return controllerOutput{}, nil, errors.New("controller_work_retained")
	}
	return output, stdout.Bytes(), nil
}

func protocolV7WritePreparation(stage string, prepared protocolV7Prepared, manifest map[string]any) error {
	if err := protocolV7WriteFreeze(stage, prepared); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(stage, "controller-authority.v4.json"), prepared.authorityBytes, 0o600); err != nil {
		return errors.New("transaction_failed")
	}
	data, err := json.Marshal(manifest)
	if err != nil || os.WriteFile(filepath.Join(stage, "preparation-manifest.json"), data, 0o600) != nil {
		return errors.New("transaction_failed")
	}
	return nil
}

func protocolV7WriteFreeze(stage string, prepared protocolV7Prepared) error {
	freeze := filepath.Join(stage, "freeze")
	if err := os.MkdirAll(freeze, 0o700); err != nil {
		return errors.New("transaction_failed")
	}
	for _, item := range []struct{ source, target string }{
		{prepared.bundle, filepath.Join(freeze, "source.bundle")},
		{prepared.runtime, filepath.Join(freeze, "controller-runtime")},
	} {
		data, err := os.ReadFile(item.source)
		if err != nil || os.WriteFile(item.target, data, 0o600) != nil {
			return errors.New("transaction_failed")
		}
	}
	return nil
}

func protocolV7CommitStage(stage, output string) error {
	if err := syncDirectoryTree(stage); err != nil {
		return errors.New("transaction_failed")
	}
	if err := os.Rename(stage, output); err != nil {
		return errors.New("transaction_failed")
	}
	if err := syncDirectory(filepath.Dir(output)); err != nil {
		_ = os.RemoveAll(output)
		return errors.New("transaction_failed")
	}
	return nil
}

func protocolV7CheckinLine(path string, seq int) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.New("threshold_input_invalid")
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var envelope struct{ Seq int `json:"seq"` }
		if err := json.Unmarshal(scanner.Bytes(), &envelope); err != nil {
			return nil, errors.New("threshold_input_invalid")
		}
		if envelope.Seq == seq {
			return append(append([]byte(nil), scanner.Bytes()...), '\n'), nil
		}
	}
	return nil, errors.New("threshold_input_invalid")
}

func protocolV7Git(dir string, args ...string) ([]byte, error) {
	home, err := os.MkdirTemp("", "c3-v7-adapter-git-home-")
	if err != nil {
		return nil, errors.New("git_failed")
	}
	defer os.RemoveAll(home)
	common := []string{"-c", "core.hooksPath=/dev/null", "-c", "core.attributesFile=/dev/null", "-c", "core.fileMode=true", "-c", "commit.gpgSign=false", "-c", "protocol.file.allow=never"}
	command := exec.Command("git", append(common, args...)...)
	command.Dir = dir
	command.Env = []string{
		"PATH=" + os.Getenv("PATH"), "LC_ALL=C", "LANG=C", "TZ=UTC", "HOME=" + home, "XDG_CONFIG_HOME=" + home,
		"GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_OPTIONAL_LOCKS=0", "GIT_NO_REPLACE_OBJECTS=1", "GIT_ALTERNATE_OBJECT_DIRECTORIES=",
		"GIT_AUTHOR_NAME=C3 Eval", "GIT_COMMITTER_NAME=C3 Eval", "GIT_AUTHOR_EMAIL=c3-eval@invalid", "GIT_COMMITTER_EMAIL=c3-eval@invalid",
		"GIT_AUTHOR_DATE=2000-01-01T00:00:00Z", "GIT_COMMITTER_DATE=2000-01-01T00:00:00Z",
	}
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil || stderr.Len() != 0 {
		return nil, errors.New("git_failed")
	}
	return output, nil
}

func mustProtocolV7FileHash(path string) string {
	hash, err := fileSHA256(path)
	if err != nil {
		return ""
	}
	return hash
}
'''
DRIVER_SHA256 = hashlib.sha256(DRIVER_SOURCE.encode()).hexdigest()


class ProtocolV7AdapterError(ValueError):
    """A generic fail-closed adapter error safe for durable output."""


class _DriverInterrupted(BaseException):
    pass


@dataclass(frozen=True)
class PrivacyPolicyBinding:
    sha256: str
    term_count: int
    canonical_bytes: bytes


@dataclass(frozen=True)
class ActivationBinding:
    sha256: str
    record_hash: str
    canonical_bytes: bytes


@dataclass(frozen=True)
class AuthorizationBinding:
    sha256: str
    record_hash: str
    canonical_bytes: bytes


@dataclass
class _ConsumedActivation:
    descriptor: int
    binding: ActivationBinding
    authorization: AuthorizationBinding
    original_path: Path
    consumed_path: Path

    def close(self) -> None:
        if self.descriptor >= 0:
            os.close(self.descriptor)
            self.descriptor = -1


@dataclass(frozen=True)
class CaptureConfig:
    source_root: Path
    main_path: Path
    main_test_path: Path
    fixture_path: Path
    benchmark_path: Path
    threshold_checkins_path: Path
    threshold_checkin_seq: int
    privacy_policy_path: Path
    output_root: Path
    authorization_record_path: Path | None = None
    expected_main_sha256: str = ACCEPTED_MAIN_SHA256
    expected_main_test_sha256: str = ACCEPTED_MAIN_TEST_SHA256
    expected_fixture_sha256: str = ACCEPTED_FIXTURE_SHA256
    expected_benchmark_sha256: str = ACCEPTED_BENCHMARK_SHA256


DriverExecutor = Callable[[str, CaptureConfig, dict[str, Any]], dict[str, Any]]


def _sha256_bytes(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def _sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1 << 20), b""):
            digest.update(block)
    return digest.hexdigest()


def _canonical_json(value: Any) -> bytes:
    return json.dumps(value, sort_keys=True, separators=(",", ":"), ensure_ascii=False).encode()


def _require_regular_no_symlink(path: Path, error_class: str, mode: int | None = None) -> os.stat_result:
    try:
        info = path.lstat()
    except OSError as exc:
        raise ProtocolV7AdapterError(error_class) from exc
    if stat.S_ISLNK(info.st_mode) or not stat.S_ISREG(info.st_mode):
        raise ProtocolV7AdapterError(error_class)
    if mode is not None and stat.S_IMODE(info.st_mode) != mode:
        raise ProtocolV7AdapterError(error_class)
    current = path.resolve(strict=True)
    if current != path.absolute():
        raise ProtocolV7AdapterError(error_class)
    return info


def _read_bounded_regular(path: Path, cap: int, error_class: str, mode: int | None = None) -> bytes:
    before = _require_regular_no_symlink(path, error_class, mode)
    if before.st_size < 0 or before.st_size > cap:
        raise ProtocolV7AdapterError(error_class)
    data = path.read_bytes()
    after = _require_regular_no_symlink(path, error_class, mode)
    if len(data) != before.st_size or (before.st_dev, before.st_ino, before.st_mtime_ns, before.st_size) != (
        after.st_dev,
        after.st_ino,
        after.st_mtime_ns,
        after.st_size,
    ):
        raise ProtocolV7AdapterError(error_class)
    return data


def validate_privacy_policy(path: Path) -> PrivacyPolicyBinding:
    data = _read_bounded_regular(path, MAX_POLICY_BYTES, "privacy_policy_invalid", 0o600)
    try:
        value = json.loads(data)
    except (UnicodeDecodeError, json.JSONDecodeError) as exc:
        raise ProtocolV7AdapterError("privacy_policy_invalid") from exc
    if not isinstance(value, dict) or set(value) != {"$schema", "deny_terms"}:
        raise ProtocolV7AdapterError("privacy_policy_invalid")
    terms = value.get("deny_terms")
    if value.get("$schema") != PRIVACY_POLICY_SCHEMA or not isinstance(terms, list) or not 0 < len(terms) <= MAX_POLICY_TERMS:
        raise ProtocolV7AdapterError("privacy_policy_invalid")
    previous = ""
    total = 0
    for term in terms:
        if not isinstance(term, str) or term != term.strip() or term != term.lower() or len(term) < 4 or term <= previous:
            raise ProtocolV7AdapterError("privacy_policy_invalid")
        if any(ord(character) < 32 or ord(character) == 127 for character in term):
            raise ProtocolV7AdapterError("privacy_policy_invalid")
        total += len(term.encode())
        previous = term
    if total > MAX_POLICY_TERM_BYTES:
        raise ProtocolV7AdapterError("privacy_policy_invalid")
    canonical = json.dumps(
        {"$schema": PRIVACY_POLICY_SCHEMA, "deny_terms": terms},
        separators=(",", ":"),
        ensure_ascii=False,
    ).encode()
    if canonical != data:
        raise ProtocolV7AdapterError("privacy_policy_invalid")
    return PrivacyPolicyBinding(_sha256_bytes(data), len(terms), data)


def _validate_source_path(root: Path, path: Path, relative: str, expected_hash: str) -> str:
    _require_regular_no_symlink(path, "frozen_input_invalid")
    try:
        resolved_root = root.resolve(strict=True)
        resolved_path = path.resolve(strict=True)
    except OSError as exc:
        raise ProtocolV7AdapterError("frozen_input_invalid") from exc
    if resolved_root != root.absolute() or not resolved_root.is_dir() or resolved_path != resolved_root / relative:
        raise ProtocolV7AdapterError("frozen_input_invalid")
    observed = _sha256_file(path)
    if observed != expected_hash:
        raise ProtocolV7AdapterError("frozen_input_invalid")
    return observed


def _adapter_sha256() -> str:
    return _sha256_file(Path(__file__).resolve())


def _validate_output_root(path: Path) -> Path:
    absolute = path.absolute()
    if absolute.exists() or absolute.is_symlink():
        raise ProtocolV7AdapterError("output_exists")
    try:
        parent = absolute.parent.resolve(strict=True)
    except OSError as exc:
        raise ProtocolV7AdapterError("output_parent_invalid") from exc
    if parent != absolute.parent or not parent.is_dir():
        raise ProtocolV7AdapterError("output_parent_invalid")
    return absolute


def _within(path: Path, root: Path) -> bool:
    try:
        path.resolve(strict=True).relative_to(root.resolve(strict=True))
        return True
    except (OSError, ValueError):
        return False


def validate_inputs(config: CaptureConfig) -> dict[str, Any]:
    if config.threshold_checkin_seq <= 0:
        raise ProtocolV7AdapterError("threshold_input_invalid")
    _require_regular_no_symlink(config.threshold_checkins_path, "threshold_input_invalid")
    policy = validate_privacy_policy(config.privacy_policy_path)
    hashes = {
        "main": _validate_source_path(
            config.source_root,
            config.main_path,
            "cli/tools/structural-search-eval-v2/main.go",
            config.expected_main_sha256,
        ),
        "main_test": _validate_source_path(
            config.source_root,
            config.main_test_path,
            "cli/tools/structural-search-eval-v2/main_test.go",
            config.expected_main_test_sha256,
        ),
        "fixture": _validate_source_path(
            config.source_root,
            config.fixture_path,
            "research/eval/structural-retrieval/fixtures.dev.v2.jsonl",
            config.expected_fixture_sha256,
        ),
        "benchmark": _validate_source_path(
            config.source_root,
            config.benchmark_path,
            "research/eval/structural-retrieval/benchmark.v2.json",
            config.expected_benchmark_sha256,
        ),
    }
    source_root = config.source_root.resolve()
    policy_path = config.privacy_policy_path.resolve()
    checkins_path = config.threshold_checkins_path.resolve()
    output_root = _validate_output_root(config.output_root)
    for left, right in (
        (source_root, policy_path),
        (source_root, output_root),
        (policy_path, output_root),
        (checkins_path, output_root),
    ):
        try:
            left.relative_to(right)
            overlap = True
        except ValueError:
            try:
                right.relative_to(left)
                overlap = True
            except ValueError:
                overlap = False
        if overlap:
            raise ProtocolV7AdapterError("governed_path_overlap")
    return {
        "$schema": INPUT_MANIFEST_SCHEMA,
        "status": "accepted",
        "adapter_sha256": _adapter_sha256(),
        "driver_sha256": DRIVER_SHA256,
        "input_sha256": hashes,
        "scorer_region_sha256": ACCEPTED_SCORER_REGION_SHA256,
        "privacy_policy_sha256": policy.sha256,
        "privacy_term_count": policy.term_count,
        "threshold_checkin_seq": config.threshold_checkin_seq,
    }


def _path_sha256(path: Path) -> str:
    return _sha256_bytes(str(path.absolute()).encode())


def _valid_sha256(value: Any) -> bool:
    return isinstance(value, str) and len(value) == 64 and not set(value) - set("0123456789abcdef") and bool(value.strip("0"))


def _authorization_payload(config: CaptureConfig, manifest: dict[str, Any], activation_path: Path) -> dict[str, Any]:
    return {
        "$schema": AUTHORIZATION_SCHEMA,
        "activation_path_sha256": _path_sha256(activation_path),
        "adapter_sha256": manifest["adapter_sha256"],
        "baseline_capture_authorized": True,
        "benchmark_sha256": config.expected_benchmark_sha256,
        "candidate_execution_authorized": False,
        "driver_sha256": DRIVER_SHA256,
        "effect_claim": False,
        "fixture_sha256": config.expected_fixture_sha256,
        "main_sha256": config.expected_main_sha256,
        "main_test_sha256": config.expected_main_test_sha256,
        "max_capture_count": 1,
        "output_root_sha256": _path_sha256(config.output_root),
        "privacy_policy_sha256": manifest["privacy_policy_sha256"],
        "scorer_region_sha256": ACCEPTED_SCORER_REGION_SHA256,
        "verdict": "authorized",
    }


def _validate_authorization_record(
    path: Path,
    config: CaptureConfig,
    manifest: dict[str, Any],
    activation_path: Path,
) -> AuthorizationBinding:
    data = _read_bounded_regular(path, 64 << 10, "authorization_invalid", 0o600)
    if not data.endswith(b"\n") or data.count(b"\n") != 1:
        raise ProtocolV7AdapterError("authorization_invalid")
    try:
        record = json.loads(data[:-1])
    except (UnicodeDecodeError, json.JSONDecodeError) as exc:
        raise ProtocolV7AdapterError("authorization_invalid") from exc
    if not isinstance(record, dict) or set(record) != {
        "payload",
        "payload_sha256",
        "prev_hash",
        "record_hash",
        "recorded_at",
        "seq",
    }:
        raise ProtocolV7AdapterError("authorization_invalid")
    expected_payload = _authorization_payload(config, manifest, activation_path)
    if record.get("payload") != expected_payload:
        raise ProtocolV7AdapterError("authorization_invalid")
    payload_sha256 = _sha256_bytes(_canonical_json(expected_payload))
    recorded_at = record.get("recorded_at")
    if (
        record.get("seq") != 1
        or record.get("prev_hash") != "GENESIS"
        or record.get("payload_sha256") != payload_sha256
        or not isinstance(recorded_at, str)
        or re.fullmatch(r"[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z", recorded_at) is None
    ):
        raise ProtocolV7AdapterError("authorization_invalid")
    without_hash = {key: value for key, value in record.items() if key != "record_hash"}
    record_hash = _sha256_bytes(_canonical_json(without_hash))
    if not _valid_sha256(record.get("record_hash")) or record["record_hash"] != record_hash:
        raise ProtocolV7AdapterError("authorization_invalid")
    expected_bytes = _canonical_json(record) + b"\n"
    if data != expected_bytes:
        raise ProtocolV7AdapterError("authorization_invalid")
    return AuthorizationBinding(_sha256_bytes(data), record_hash, data)


def _validate_activation(
    path: Path,
    config: CaptureConfig,
    manifest: dict[str, Any],
    *,
    bound_path: Path | None = None,
    authorization: AuthorizationBinding | None = None,
) -> ActivationBinding:
    data = _read_bounded_regular(path, 64 << 10, "activation_invalid", 0o600)
    try:
        value = json.loads(data)
    except (UnicodeDecodeError, json.JSONDecodeError) as exc:
        raise ProtocolV7AdapterError("activation_invalid") from exc
    record_hash = value.get("activation_record_hash") if isinstance(value, dict) else None
    if not _valid_sha256(record_hash) or authorization is None or record_hash != authorization.record_hash:
        raise ProtocolV7AdapterError("activation_invalid")
    if config.authorization_record_path is None:
        raise ProtocolV7AdapterError("authorization_invalid")
    expected = {
        "$schema": ACTIVATION_SCHEMA,
        "activation_path_sha256": _path_sha256(bound_path or path),
        "activation_record_hash": record_hash,
        "authorization_record_path_sha256": _path_sha256(config.authorization_record_path),
        "adapter_sha256": manifest["adapter_sha256"],
        "benchmark_sha256": config.expected_benchmark_sha256,
        "driver_sha256": DRIVER_SHA256,
        "fixture_sha256": config.expected_fixture_sha256,
        "main_sha256": config.expected_main_sha256,
        "main_test_sha256": config.expected_main_test_sha256,
        "max_capture_count": 1,
        "output_root_sha256": _path_sha256(config.output_root),
        "privacy_policy_sha256": manifest["privacy_policy_sha256"],
        "scorer_region_sha256": ACCEPTED_SCORER_REGION_SHA256,
        "verdict": "authorized",
    }
    if value != expected or data != _canonical_json(expected):
        raise ProtocolV7AdapterError("activation_invalid")
    return ActivationBinding(_sha256_bytes(data), record_hash, data)


def _fsync_parent(path: Path) -> None:
    descriptor = os.open(path.parent, os.O_RDONLY | getattr(os, "O_DIRECTORY", 0))
    try:
        os.fsync(descriptor)
    finally:
        os.close(descriptor)


def _consume_activation(
    path: Path,
    binding: ActivationBinding,
    authorization: AuthorizationBinding,
) -> _ConsumedActivation:
    consumed = Path(str(path) + ".consumed")
    if consumed.exists() or consumed.is_symlink():
        raise ProtocolV7AdapterError("activation_used")
    descriptor = -1
    try:
        flags = os.O_RDONLY | getattr(os, "O_CLOEXEC", 0) | getattr(os, "O_NOFOLLOW", 0)
        descriptor = os.open(path, flags)
        descriptor_info = os.fstat(descriptor)
        path_info = path.lstat()
        if (
            not stat.S_ISREG(descriptor_info.st_mode)
            or stat.S_IMODE(descriptor_info.st_mode) != 0o600
            or (descriptor_info.st_dev, descriptor_info.st_ino) != (path_info.st_dev, path_info.st_ino)
        ):
            raise ProtocolV7AdapterError("activation_invalid")
        os.lseek(descriptor, 0, os.SEEK_SET)
        descriptor_bytes = os.read(descriptor, (64 << 10) + 1)
        if descriptor_bytes != binding.canonical_bytes:
            raise ProtocolV7AdapterError("activation_invalid")
        os.link(path, consumed, follow_symlinks=False)
        consumed.chmod(0o600)
        consumed_info = consumed.lstat()
        if (
            (consumed_info.st_dev, consumed_info.st_ino) != (descriptor_info.st_dev, descriptor_info.st_ino)
            or _sha256_file(consumed) != binding.sha256
        ):
            raise ProtocolV7AdapterError("activation_invalid")
        os.unlink(path)
        _fsync_parent(consumed)
    except ProtocolV7AdapterError:
        if descriptor >= 0:
            os.close(descriptor)
        raise
    except OSError as exc:
        if descriptor >= 0:
            os.close(descriptor)
        raise ProtocolV7AdapterError("activation_used") from exc
    return _ConsumedActivation(descriptor, binding, authorization, path, consumed)


def capture_baseline(config: CaptureConfig, activation_path: Path, *, executor: DriverExecutor | None = None) -> dict[str, Any]:
    manifest = validate_inputs(config)
    if config.authorization_record_path is None:
        raise ProtocolV7AdapterError("authorization_invalid")
    if _within(activation_path, config.source_root) or activation_path.resolve() == config.privacy_policy_path.resolve():
        raise ProtocolV7AdapterError("activation_invalid")
    authorization = _validate_authorization_record(
        config.authorization_record_path,
        config,
        manifest,
        activation_path,
    )
    binding = _validate_activation(activation_path, config, manifest, authorization=authorization)
    consumed = _consume_activation(activation_path, binding, authorization)
    try:
        if executor is not None:
            return executor("capture-baseline", config, manifest)
        return _run_capture_driver(config, manifest, consumed)
    except BaseException:
        shutil.rmtree(config.output_root, ignore_errors=True)
        raise
    finally:
        consumed.close()


def activation_self_test(config: CaptureConfig, activation_path: Path) -> dict[str, Any]:
    manifest = validate_inputs(config)
    if config.authorization_record_path is None:
        raise ProtocolV7AdapterError("authorization_invalid")
    if _within(activation_path, config.source_root) or activation_path.resolve() == config.privacy_policy_path.resolve():
        raise ProtocolV7AdapterError("activation_invalid")
    authorization = _validate_authorization_record(
        config.authorization_record_path,
        config,
        manifest,
        activation_path,
    )
    binding = _validate_activation(activation_path, config, manifest, authorization=authorization)
    consumed = _consume_activation(activation_path, binding, authorization)
    try:
        result = _run_activation_verifier(config, manifest, consumed)
        if config.output_root.exists() or config.output_root.is_symlink():
            shutil.rmtree(config.output_root, ignore_errors=True)
            raise ProtocolV7AdapterError("self_test_cleanup_failed")
        return result
    finally:
        consumed.close()
        try:
            consumed.consumed_path.unlink()
            _fsync_parent(consumed.consumed_path)
        except FileNotFoundError:
            pass


def self_test(config: CaptureConfig, *, executor: DriverExecutor | None = None) -> dict[str, Any]:
    manifest = validate_inputs(config)
    try:
        result = (executor or run_driver)("self-test", config, manifest)
    except BaseException:
        shutil.rmtree(config.output_root, ignore_errors=True)
        raise
    if config.output_root.exists():
        shutil.rmtree(config.output_root, ignore_errors=True)
        raise ProtocolV7AdapterError("self_test_cleanup_failed")
    return result


def prepare_baseline(config: CaptureConfig, *, executor: DriverExecutor | None = None) -> dict[str, Any]:
    manifest = validate_inputs(config)
    try:
        return (executor or run_driver)("prepare-baseline", config, manifest)
    except BaseException:
        shutil.rmtree(config.output_root, ignore_errors=True)
        raise


def _copy_tree_no_symlinks(source: Path, target: Path) -> None:
    for path in sorted(source.rglob("*")):
        relative = path.relative_to(source)
        destination = target / relative
        info = path.lstat()
        if stat.S_ISLNK(info.st_mode):
            raise ProtocolV7AdapterError("driver_source_invalid")
        if stat.S_ISDIR(info.st_mode):
            destination.mkdir(parents=True, exist_ok=True)
            destination.chmod(0o755)
        elif stat.S_ISREG(info.st_mode):
            destination.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(path, destination)
            destination.chmod(0o644)
        else:
            raise ProtocolV7AdapterError("driver_source_invalid")


def _go_environment() -> dict[str, str]:
    home = Path.home()
    cache = Path(os.environ.get("XDG_CACHE_HOME", str(home / ".cache")))
    if not cache.is_absolute():
        cache = home / ".cache"
    return {
        "PATH": os.environ.get("PATH", ""),
        "HOME": str(home),
        "LC_ALL": "C",
        "LANG": "C",
        "TZ": "UTC",
        "GOENV": "off",
        "GOWORK": "off",
        "GOFLAGS": "",
        "GOTOOLCHAIN": "local",
        "CGO_ENABLED": "0",
        "GOPROXY": "off",
        "GOSUMDB": "off",
        "GONOSUMDB": "",
        "GOPRIVATE": "",
        "GOMODCACHE": str(home / "go" / "pkg" / "mod"),
        "GOCACHE": str(cache / "go-build"),
    }


def _generic_result(value: Any) -> bool:
    encoded = json.dumps(value, sort_keys=True, separators=(",", ":"), ensure_ascii=False).lower()
    forbidden = ("/home/", "/users/", "/root/", "/tmp/", "\\users\\", "private key", "authorization: bearer")
    return not any(item in encoded for item in forbidden)


def _process_session_members(session_id: int) -> set[int]:
    members: set[int] = set()
    proc = Path("/proc")
    if not proc.is_dir():
        return members
    for entry in proc.iterdir():
        if not entry.name.isdigit():
            continue
        try:
            raw = (entry / "stat").read_text(encoding="utf-8")
            fields = raw[raw.rfind(")") + 2 :].split()
            if len(fields) > 3 and int(fields[3]) == session_id:
                members.add(int(entry.name))
        except (OSError, ValueError):
            continue
    return members


def _kill_process_session(process: subprocess.Popen[bytes]) -> None:
    for _ in range(3):
        members = _process_session_members(process.pid)
        for pid in members:
            if pid == os.getpid():
                continue
            try:
                os.kill(pid, signal.SIGKILL)
            except ProcessLookupError:
                pass
        try:
            os.killpg(process.pid, signal.SIGKILL)
        except ProcessLookupError:
            pass
        if not _process_session_members(process.pid):
            break
    try:
        process.wait(timeout=5)
    except (subprocess.TimeoutExpired, ChildProcessError):
        pass


def _run_driver_process(args: list[str], *, cwd: Path, env: dict[str, str], timeout: int) -> subprocess.CompletedProcess[bytes]:
    try:
        process = subprocess.Popen(
            args,
            cwd=cwd,
            env=env,
            stdin=subprocess.DEVNULL,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            start_new_session=True,
        )
    except OSError as exc:
        raise ProtocolV7AdapterError("driver_execution_failed") from exc

    previous_handlers: dict[int, Any] = {}

    def interrupt_handler(signum: int, _frame: Any) -> None:
        raise _DriverInterrupted(signum)

    for signum in (signal.SIGTERM, signal.SIGHUP):
        try:
            previous_handlers[signum] = signal.getsignal(signum)
            signal.signal(signum, interrupt_handler)
        except ValueError:
            for installed, handler in previous_handlers.items():
                signal.signal(installed, handler)
            previous_handlers.clear()
            break
    try:
        stdout, stderr = process.communicate(timeout=timeout)
    except subprocess.TimeoutExpired as exc:
        _kill_process_session(process)
        raise ProtocolV7AdapterError("driver_timeout") from exc
    except _DriverInterrupted as exc:
        _kill_process_session(process)
        raise ProtocolV7AdapterError("driver_interrupted") from exc
    except BaseException:
        _kill_process_session(process)
        raise
    finally:
        for signum, handler in previous_handlers.items():
            signal.signal(signum, handler)
    return subprocess.CompletedProcess(args, process.returncode, stdout, stderr)


def _cleanup_driver_outputs(config: CaptureConfig, stage_root: Path) -> None:
    shutil.rmtree(stage_root, ignore_errors=True)
    shutil.rmtree(config.output_root, ignore_errors=True)


def run_driver(
    command: str,
    config: CaptureConfig,
    input_manifest: dict[str, Any],
    *,
    consumed_activation: Path | None = None,
) -> dict[str, Any]:
    if command == "capture-baseline" or consumed_activation is not None:
        raise ProtocolV7AdapterError("activation_invalid")
    return _run_driver_common(command, config, input_manifest, consumed=None)


def _run_capture_driver(
    config: CaptureConfig,
    input_manifest: dict[str, Any],
    consumed: _ConsumedActivation,
) -> dict[str, Any]:
    return _run_consumed_driver("capture-baseline", config, input_manifest, consumed)


def _run_activation_verifier(
    config: CaptureConfig,
    input_manifest: dict[str, Any],
    consumed: _ConsumedActivation,
) -> dict[str, Any]:
    return _run_consumed_driver("verify-activation", config, input_manifest, consumed)


def _run_consumed_driver(
    command: str,
    config: CaptureConfig,
    input_manifest: dict[str, Any],
    consumed: _ConsumedActivation,
) -> dict[str, Any]:
    if command not in {"capture-baseline", "verify-activation"}:
        raise ProtocolV7AdapterError("activation_invalid")
    observed_manifest = validate_inputs(config)
    if observed_manifest != input_manifest:
        raise ProtocolV7AdapterError("frozen_input_invalid")
    if config.authorization_record_path is None or consumed.descriptor < 0:
        raise ProtocolV7AdapterError("activation_invalid")
    if consumed.original_path.exists() or consumed.original_path.is_symlink():
        raise ProtocolV7AdapterError("activation_invalid")
    descriptor_info = os.fstat(consumed.descriptor)
    receipt_info = consumed.consumed_path.lstat()
    if (descriptor_info.st_dev, descriptor_info.st_ino) != (receipt_info.st_dev, receipt_info.st_ino):
        raise ProtocolV7AdapterError("activation_invalid")
    authorization = _validate_authorization_record(
        config.authorization_record_path,
        config,
        input_manifest,
        consumed.original_path,
    )
    if authorization != consumed.authorization:
        raise ProtocolV7AdapterError("authorization_invalid")
    activation = _validate_activation(
        consumed.consumed_path,
        config,
        input_manifest,
        bound_path=consumed.original_path,
        authorization=authorization,
    )
    if activation != consumed.binding:
        raise ProtocolV7AdapterError("activation_invalid")
    return _run_driver_common(command, config, input_manifest, consumed=consumed)


def _run_driver_common(
    command: str,
    config: CaptureConfig,
    input_manifest: dict[str, Any],
    *,
    consumed: _ConsumedActivation | None,
) -> dict[str, Any]:
    observed_manifest = validate_inputs(config)
    if observed_manifest != input_manifest:
        raise ProtocolV7AdapterError("frozen_input_invalid")
    if command in {"capture-baseline", "verify-activation"} and consumed is None:
        raise ProtocolV7AdapterError("activation_invalid")
    if command not in {"capture-baseline", "verify-activation"} and consumed is not None:
        raise ProtocolV7AdapterError("activation_invalid")
    activation_binding = consumed.binding if consumed else None
    authorization_binding = consumed.authorization if consumed else None
    policy_before = validate_privacy_policy(config.privacy_policy_path)
    original_hashes = dict(input_manifest["input_sha256"])
    with tempfile.TemporaryDirectory(prefix="c3-v7-adapter-driver-") as temporary:
        temporary_root = Path(temporary)
        private_tmp = temporary_root / "tmp"
        private_tmp.mkdir(mode=0o700)
        stage_token = _sha256_bytes((str(temporary_root) + str(config.output_root.absolute())).encode())[:20]
        stage_root = config.output_root.absolute().parent / (".c3-v7-adapter-" + stage_token)
        scratch_root = temporary_root / "scratch"
        driver_root = temporary_root / "driver-source"
        (driver_root / "cli").mkdir(parents=True)
        _copy_tree_no_symlinks(config.source_root / "cli", driver_root / "cli")
        for source, relative in (
            (config.fixture_path, "research/eval/structural-retrieval/fixtures.dev.v2.jsonl"),
            (config.benchmark_path, "research/eval/structural-retrieval/benchmark.v2.json"),
        ):
            destination = driver_root / relative
            destination.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(source, destination)
            destination.chmod(0o644)
        driver_path = driver_root / "cli/tools/structural-search-eval-v2/protocol_v7_capture_driver_test.go"
        driver_path.write_text(DRIVER_SOURCE, encoding="utf-8")
        driver_path.chmod(0o600)
        if _sha256_file(driver_path) != DRIVER_SHA256:
            raise ProtocolV7AdapterError("driver_identity_invalid")

        policy_copy = temporary_root / "privacy-policy.json"
        policy_copy.write_bytes(policy_before.canonical_bytes)
        policy_copy.chmod(0o600)
        activation_evidence_path = temporary_root / "activation-evidence.json"
        authorization_evidence_path = temporary_root / "authorization-evidence.jsonl"
        if consumed is not None:
            os.lseek(consumed.descriptor, 0, os.SEEK_SET)
            live_activation = os.read(consumed.descriptor, (64 << 10) + 1)
            if live_activation != consumed.binding.canonical_bytes:
                raise ProtocolV7AdapterError("activation_invalid")
            activation_evidence_path.write_bytes(live_activation)
            activation_evidence_path.chmod(0o600)
            authorization_evidence_path.write_bytes(consumed.authorization.canonical_bytes)
            authorization_evidence_path.chmod(0o600)
        result_path = temporary_root / "driver-result.json"
        config_path = temporary_root / "driver-config.json"
        driver_config = {
            "command": command,
            "source_root": str(config.source_root.resolve()),
            "main_path": str(config.main_path.resolve()),
            "main_test_path": str(config.main_test_path.resolve()),
            "fixture_path": str(config.fixture_path.resolve()),
            "benchmark_path": str(config.benchmark_path.resolve()),
            "threshold_checkins_path": str(config.threshold_checkins_path.resolve()),
            "threshold_checkin_seq": config.threshold_checkin_seq,
            "privacy_policy_path": str(policy_copy),
            "output_root": str(config.output_root.absolute()),
            "stage_root": str(stage_root),
            "scratch_root": str(scratch_root),
            "result_path": str(result_path),
            "expected_main_sha256": config.expected_main_sha256,
            "expected_main_test_sha256": config.expected_main_test_sha256,
            "expected_fixture_sha256": config.expected_fixture_sha256,
            "expected_benchmark_sha256": config.expected_benchmark_sha256,
            "expected_scorer_region_sha256": ACCEPTED_SCORER_REGION_SHA256,
            "adapter_sha256": input_manifest["adapter_sha256"],
            "driver_sha256": DRIVER_SHA256,
            "input_manifest_sha256": _sha256_bytes(_canonical_json(input_manifest)),
            "capture_authorized": activation_binding is not None,
            "activation_proof_sha256": activation_binding.sha256 if activation_binding else "",
            "activation_record_hash": activation_binding.record_hash if activation_binding else "",
            "activation_evidence_path": str(activation_evidence_path) if consumed else "",
            "activation_original_path": str(consumed.original_path.absolute()) if consumed else "",
            "activation_receipt_path": str(consumed.consumed_path.absolute()) if consumed else "",
            "authorization_evidence_path": str(authorization_evidence_path) if consumed else "",
            "authorization_record_path": str(config.authorization_record_path.absolute()) if consumed and config.authorization_record_path else "",
            "authorization_record_sha256": authorization_binding.sha256 if authorization_binding else "",
        }
        config_path.write_bytes(_canonical_json(driver_config))
        config_path.chmod(0o600)
        environment = _go_environment()
        environment["C3_PROTOCOL_V7_DRIVER_CONFIG"] = str(config_path)
        environment["TMPDIR"] = str(private_tmp)
        if command != "verify-activation":
            try:
                stage_root.mkdir(mode=0o700)
            except OSError as exc:
                raise ProtocolV7AdapterError("transaction_failed") from exc
        try:
            completed = _run_driver_process(
                ["go", "test", "-count=1", "-run", "^TestProtocolV7CaptureDriver$", "./tools/structural-search-eval-v2"],
                cwd=driver_root / "cli",
                env=environment,
                timeout=900,
            )
        except BaseException:
            _cleanup_driver_outputs(config, stage_root)
            raise
        if completed.returncode != 0 or completed.stderr:
            _cleanup_driver_outputs(config, stage_root)
            raise ProtocolV7AdapterError("driver_execution_failed")
        try:
            result = json.loads(result_path.read_bytes())
        except (OSError, json.JSONDecodeError) as exc:
            _cleanup_driver_outputs(config, stage_root)
            raise ProtocolV7AdapterError("driver_result_invalid") from exc
        if not isinstance(result, dict) or result.get("status") not in {"accepted", "accepted_disposable_non_study"} or not _generic_result(result):
            _cleanup_driver_outputs(config, stage_root)
            raise ProtocolV7AdapterError("driver_result_invalid")
        if stage_root.exists() or stage_root.is_symlink():
            _cleanup_driver_outputs(config, stage_root)
            raise ProtocolV7AdapterError("transaction_cleanup_failed")

    policy_after = validate_privacy_policy(config.privacy_policy_path)
    if policy_after != policy_before:
        shutil.rmtree(config.output_root, ignore_errors=True)
        raise ProtocolV7AdapterError("privacy_policy_invalid")
    if consumed is not None:
        if config.authorization_record_path is None:
            shutil.rmtree(config.output_root, ignore_errors=True)
            raise ProtocolV7AdapterError("authorization_invalid")
        authorization_after = _validate_authorization_record(
            config.authorization_record_path,
            config,
            input_manifest,
            consumed.original_path,
        )
        activation_after = _validate_activation(
            consumed.consumed_path,
            config,
            input_manifest,
            bound_path=consumed.original_path,
            authorization=authorization_after,
        )
        if authorization_after != consumed.authorization or activation_after != consumed.binding:
            shutil.rmtree(config.output_root, ignore_errors=True)
            raise ProtocolV7AdapterError("activation_invalid")
    post_hashes = {
        "main": _sha256_file(config.main_path),
        "main_test": _sha256_file(config.main_test_path),
        "fixture": _sha256_file(config.fixture_path),
        "benchmark": _sha256_file(config.benchmark_path),
    }
    if post_hashes != original_hashes:
        shutil.rmtree(config.output_root, ignore_errors=True)
        raise ProtocolV7AdapterError("frozen_input_invalid")
    if command == "self-test" and config.output_root.exists():
        shutil.rmtree(config.output_root, ignore_errors=True)
        raise ProtocolV7AdapterError("self_test_cleanup_failed")
    return result


def _parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("command", choices=("validate-inputs", "prepare-baseline", "self-test", "capture-baseline"))
    parser.add_argument("--source-root", type=Path, required=True)
    parser.add_argument("--main", dest="main_path", type=Path, required=True)
    parser.add_argument("--main-test", dest="main_test_path", type=Path, required=True)
    parser.add_argument("--fixtures", dest="fixture_path", type=Path, required=True)
    parser.add_argument("--benchmark", dest="benchmark_path", type=Path, required=True)
    parser.add_argument("--threshold-checkins", dest="threshold_checkins_path", type=Path, required=True)
    parser.add_argument("--threshold-checkin-seq", type=int, required=True)
    parser.add_argument("--privacy-policy", dest="privacy_policy_path", type=Path, required=True)
    parser.add_argument("--output-root", type=Path, required=True)
    parser.add_argument("--activation", type=Path)
    parser.add_argument("--authorization-record", type=Path)
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = _parse_args(argv)
    config = CaptureConfig(
        source_root=args.source_root,
        main_path=args.main_path,
        main_test_path=args.main_test_path,
        fixture_path=args.fixture_path,
        benchmark_path=args.benchmark_path,
        threshold_checkins_path=args.threshold_checkins_path,
        threshold_checkin_seq=args.threshold_checkin_seq,
        privacy_policy_path=args.privacy_policy_path,
        output_root=args.output_root,
        authorization_record_path=args.authorization_record,
    )
    try:
        if args.command == "validate-inputs":
            result = validate_inputs(config)
        elif args.command == "prepare-baseline":
            result = prepare_baseline(config)
        elif args.command == "self-test":
            result = self_test(config)
        else:
            if args.activation is None:
                raise ProtocolV7AdapterError("activation_invalid")
            result = capture_baseline(config, args.activation)
    except ProtocolV7AdapterError as exc:
        print(json.dumps({"$schema": RESULT_SCHEMA, "status": "rejected", "error_class": str(exc)}, sort_keys=True, separators=(",", ":")))
        return 2
    print(json.dumps(result, sort_keys=True, separators=(",", ":")))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
