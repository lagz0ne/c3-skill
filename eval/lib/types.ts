export interface TestCase {
  name: string;
  fixtures: string;
  command: string;
  goal: string;
  constraints: string[];
  // For expectations-based eval
  expectations?: string[];
  // For alter eval
  eval_type?: "expectations" | "alter";
  change_request?: string; // Original user request for alter eval
}

export interface TraceEntry {
  timestamp: string;
  event: string;
  tool: string;
  input?: Record<string, unknown>;
  result?: unknown;
  blocked?: boolean;
  reason?: string;
}

export interface ExpectationResult {
  text: string;
  pass: boolean;
  reasoning: string;
}

export interface JudgeResponse {
  pass: boolean;
  score: number;
  expectations: ExpectationResult[];
  summary: string;
}

// Re-export alter judge types
export type { AlterJudgeResult, Rating, ConcernStatus, PurposeResult, ConcernResult } from "./alter-judge";

export interface TestResult {
  name: string;
  pass: boolean;
  score: number;
  duration: number;
  trace: TraceEntry[];
  judgment: JudgeResponse;
  error?: string;
}

// Golden Example Types
export interface GoldenMetadata {
  name: string;
  skill: string;
  category: string;
  description: string;
  input_context: {
    project_name: string;
    domain: string;
    features: string[];
    tech_stack: Record<string, string>;
  };
  evaluation_dimensions: {
    structure: string[];
    content_quality: string[];
    references: string[];
    balance: string[];
  };
  created: string;
  source_session: string;
  refinements: string[];
}

export interface DimensionScore {
  dimension: string;
  score: number;
  checks: {
    check: string;
    pass: boolean;
    reasoning: string;
  }[];
}

export interface GoldenJudgeResponse {
  overall_score: number;
  dimensions: DimensionScore[];
  gaps: string[];
  verdict: "pass" | "fail";
  summary: string;
}
