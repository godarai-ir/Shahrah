import type { LucideIcon } from "lucide-react";
import {
  BadgeCheck,
  Ban,
  Braces,
  BrainCircuit,
  CircleDashed,
  CodeXml,
  FileLock2,
  Fingerprint,
  ImageDown,
  KeyRound,
  ListTree,
  Pilcrow,
  Regex,
  Replace,
  Type,
  WholeWord,
} from "lucide-react";

import type { GuardrailOperation, GuardrailType } from "./types";

export type GuardrailCategory = "semantic" | "data" | "content" | "length" | "structure" | "access";
export type GuardrailOps = GuardrailOperation | "both";
export type GuardrailCountUnit = "sentence" | "word" | "character";

export type GuardrailCatalogEntry = {
  type: GuardrailType;
  category: GuardrailCategory;
  icon: LucideIcon;
  ops: GuardrailOps;
  recommended?: boolean;
  unit?: GuardrailCountUnit;
};

export const GUARDRAIL_CATEGORIES: GuardrailCategory[] = ["semantic", "data", "content", "length", "structure", "access"];

/** The single source of guardrail type metadata. Array order is display order. */
export const GUARDRAIL_CATALOG: GuardrailCatalogEntry[] = [
  { type: "semantic_policy", category: "semantic", icon: BrainCircuit, ops: "validate", recommended: true },
  { type: "secrets_detection", category: "data", icon: KeyRound, ops: "both", recommended: true },
  { type: "basic_pii", category: "data", icon: Fingerprint, ops: "both", recommended: true },
  { type: "document_confidentiality", category: "data", icon: FileLock2, ops: "validate" },
  { type: "regex", category: "data", icon: Regex, ops: "both" },
  { type: "regex_replace", category: "data", icon: Replace, ops: "mutate" },
  { type: "keyword_blocklist", category: "content", icon: Ban, ops: "validate", recommended: true },
  { type: "contains_code", category: "content", icon: CodeXml, ops: "validate" },
  { type: "not_null", category: "content", icon: CircleDashed, ops: "validate" },
  { type: "sentence_count", category: "length", icon: Pilcrow, ops: "validate", unit: "sentence" },
  { type: "word_count", category: "length", icon: WholeWord, ops: "validate", unit: "word" },
  { type: "character_count", category: "length", icon: Type, ops: "validate", unit: "character" },
  { type: "json_schema", category: "structure", icon: Braces, ops: "validate", recommended: true },
  { type: "json_keys", category: "structure", icon: ListTree, ops: "validate" },
  { type: "jwt_validation", category: "access", icon: BadgeCheck, ops: "validate" },
  { type: "inline_image_urls", category: "access", icon: ImageDown, ops: "mutate" },
];

const catalogByType = new Map(GUARDRAIL_CATALOG.map((entry) => [entry.type, entry]));

export function catalogEntry(type: GuardrailType): GuardrailCatalogEntry {
  const entry = catalogByType.get(type);
  if (!entry) throw new Error(`Unknown guardrail type: ${type}`);
  return entry;
}

export function isValidateOnlyType(type: GuardrailType) {
  return catalogEntry(type).ops === "validate";
}

export function isMutateOnlyType(type: GuardrailType) {
  return catalogEntry(type).ops === "mutate";
}

export function operationForType(type: GuardrailType, operation: GuardrailOperation): GuardrailOperation {
  const ops = catalogEntry(type).ops;
  return ops === "both" ? operation : ops;
}

/** The operation preselected when a type is picked: `mutate` for flexible types, except `regex`. */
export function defaultOperationForType(type: GuardrailType): GuardrailOperation {
  const ops = catalogEntry(type).ops;
  if (ops !== "both") return ops;
  return type === "regex" ? "validate" : "mutate";
}
