"use client";

import {
  buildFallbackManifest,
  type PromptManifest,
  type PromptTranslationFn,
  type PromptVersion,
} from "./prompt-management-shared";

export type PromptDetailTab = "gateway" | "prompt" | "sdk" | "spec";

export type PromptCompareRow = {
  change: "added" | "changed" | "removed" | "same";
  leftLineNumber?: number;
  leftText: string;
  rightLineNumber?: number;
  rightText: string;
};

type DiffOp =
  | { type: "add"; text: string }
  | { type: "remove"; text: string }
  | { type: "same"; text: string };

export function getPromptManifest(version: PromptVersion): PromptManifest {
  return version.manifest ?? buildFallbackManifest(version);
}

export function serializePromptSpec(version: PromptVersion): string {
  return JSON.stringify(getPromptManifest(version), null, 2);
}

export function buildSdkSnippet(version: PromptVersion): string {
  const manifest = getPromptManifest(version);
  const model = manifest.model ?? version.model ?? "";
  const variableNames = Object.keys(manifest.variables ?? {});
  const promptVariables =
    variableNames.length === 0
      ? "{}"
      : `{\n${variableNames.map((name) => `        "${name}": "value"`).join(",\n")}\n    }`;

  return `from openai import OpenAI

client = OpenAI(
    api_key="YOUR_SHAHRAH_API_KEY",
    base_url="https://TENANT.shahrah.ir/v1"
)

response = client.chat.completions.create(
    model="${model}",
    messages=[],
    extra_body={
        "prompt_version_fqn": "${version.fqn}",
        "prompt_variables": ${promptVariables}
    }
)

print(response.choices[0].message.content)`;
}

export function buildGatewaySnippet(version: PromptVersion, stream: boolean): string {
  const manifest = getPromptManifest(version);
  const model = manifest.model ?? version.model ?? "";
  const variableEntries = Object.keys(manifest.variables ?? {}).map((name) => `    "${name}": "value"`);
  const promptVariables = variableEntries.length === 0 ? "  {}" : `  {\n${variableEntries.join(",\n")}\n  }`;

  return `curl --location 'https://TENANT.shahrah.ir/v1/chat/completions' \\
  --header 'Authorization: Bearer YOUR_SHAHRAH_API_KEY' \\
  --header 'Content-Type: application/json' \\
  --data '{
  "model": "${model}",
  "messages": [],
  "stream": ${stream ? "true" : "false"},
  "extra_body": {
    "prompt_version_fqn": "${version.fqn}",
    "prompt_variables": ${promptVariables}
  }
}'`;
}

export function describePromptMessage(version: PromptVersion, t: PromptTranslationFn) {
  const manifest = getPromptManifest(version);
  const systemMessage = manifest.messages.find((message) => message.role === "system");
  const conversationMessages = manifest.messages.filter((message) => message.role === "user" || message.role === "assistant");

  return {
    cache: manifest.cache,
    conversationMessages,
    description: manifest.description ?? version.description ?? "",
    guardrails: manifest.guardrails,
    metadata: manifest.metadata ?? {},
    model: manifest.model ?? version.model ?? "",
    modelSettings: manifest.model_settings ?? {},
    structuredOutput: manifest.structured_output,
    systemPrompt: typeof systemMessage?.content === "string" ? systemMessage.content : "",
    variableNames: Object.keys(manifest.variables ?? {}),
    versionLabel: t("versions.versionValue", { version: version.version }),
  };
}

export function buildPromptDiffRows(leftSpec: string, rightSpec: string): PromptCompareRow[] {
  const leftLines = leftSpec.split("\n");
  const rightLines = rightSpec.split("\n");
  const ops = computeDiffOps(leftLines, rightLines);
  const rows: PromptCompareRow[] = [];
  let leftLineNumber = 1;
  let rightLineNumber = 1;

  for (let index = 0; index < ops.length; index += 1) {
    const op = ops[index];
    if (!op) continue;

    if (op.type === "same") {
      rows.push({
        change: "same",
        leftLineNumber,
        leftText: op.text,
        rightLineNumber,
        rightText: op.text,
      });
      leftLineNumber += 1;
      rightLineNumber += 1;
      continue;
    }

    if (op.type === "remove" || op.type === "add") {
      const removed: string[] = [];
      const added: string[] = [];
      while (ops[index]?.type === "remove" || ops[index]?.type === "add") {
        const current = ops[index];
        if (!current) break;
        if (current.type === "remove") removed.push(current.text);
        if (current.type === "add") added.push(current.text);
        index += 1;
      }
      index -= 1;

      const rowCount = Math.max(removed.length, added.length);
      for (let runIndex = 0; runIndex < rowCount; runIndex += 1) {
        const leftText = removed[runIndex] ?? "";
        const rightText = added[runIndex] ?? "";
        if (leftText && rightText) {
          rows.push({
            change: "changed",
            leftLineNumber,
            leftText,
            rightLineNumber,
            rightText,
          });
          leftLineNumber += 1;
          rightLineNumber += 1;
          continue;
        }
        if (leftText) {
          rows.push({
            change: "removed",
            leftLineNumber,
            leftText,
            rightText: "",
          });
          leftLineNumber += 1;
          continue;
        }
        rows.push({
          change: "added",
          leftText: "",
          rightLineNumber,
          rightText,
        });
        rightLineNumber += 1;
      }
    }
  }

  return rows;
}

function computeDiffOps(leftLines: string[], rightLines: string[]): DiffOp[] {
  const leftCount = leftLines.length;
  const rightCount = rightLines.length;
  const dp = Array.from({ length: leftCount + 1 }, () => Array<number>(rightCount + 1).fill(0));

  for (let leftIndex = leftCount - 1; leftIndex >= 0; leftIndex -= 1) {
    for (let rightIndex = rightCount - 1; rightIndex >= 0; rightIndex -= 1) {
      if (leftLines[leftIndex] === rightLines[rightIndex]) {
        dp[leftIndex][rightIndex] = dp[leftIndex + 1][rightIndex + 1] + 1;
      } else {
        dp[leftIndex][rightIndex] = Math.max(dp[leftIndex + 1][rightIndex], dp[leftIndex][rightIndex + 1]);
      }
    }
  }

  const ops: DiffOp[] = [];
  let leftIndex = 0;
  let rightIndex = 0;

  while (leftIndex < leftCount && rightIndex < rightCount) {
    if (leftLines[leftIndex] === rightLines[rightIndex]) {
      ops.push({ type: "same", text: leftLines[leftIndex] });
      leftIndex += 1;
      rightIndex += 1;
      continue;
    }
    if (dp[leftIndex + 1][rightIndex] >= dp[leftIndex][rightIndex + 1]) {
      ops.push({ type: "remove", text: leftLines[leftIndex] });
      leftIndex += 1;
      continue;
    }
    ops.push({ type: "add", text: rightLines[rightIndex] });
    rightIndex += 1;
  }

  while (leftIndex < leftCount) {
    ops.push({ type: "remove", text: leftLines[leftIndex] });
    leftIndex += 1;
  }

  while (rightIndex < rightCount) {
    ops.push({ type: "add", text: rightLines[rightIndex] });
    rightIndex += 1;
  }

  return ops;
}
