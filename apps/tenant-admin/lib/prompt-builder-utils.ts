const VARIABLE_PATTERN = /\{\{\s*([A-Za-z_][A-Za-z0-9_]*)\s*\}\}/g;

export type PromptEditorMessage = {
  role: "user" | "assistant";
  content: string;
};

export function extractPromptVariableNames(input: Array<string | undefined | null>): string[] {
  const found = new Set<string>();
  for (const item of input) {
    const text = `${item ?? ""}`;
    for (const match of text.matchAll(VARIABLE_PATTERN)) {
      const name = match[1]?.trim();
      if (name) {
        found.add(name);
      }
    }
  }
  return [...found].sort((left, right) => left.localeCompare(right));
}

export function renderPromptText(template: string, values: Record<string, string>) {
  const missing = new Set<string>();
  const rendered = template.replaceAll(VARIABLE_PATTERN, (_match, rawName: string) => {
    const name = rawName.trim();
    const value = values[name];
    if (typeof value !== "string" || value.length === 0) {
      missing.add(name);
      return `{{${name}}}`;
    }
    return value;
  });
  return { rendered, missing: [...missing].sort((left, right) => left.localeCompare(right)) };
}

export function normalizePromptSlug(value: string): string {
  return value.trim().replaceAll(/\s+/g, "-");
}

export function buildPromptUsageSnippet(fqn: string, model?: string) {
  return `from openai import OpenAI

client = OpenAI(
    api_key="YOUR_GODARAI_API_KEY",
    base_url="https://TENANT.sharah.ir/v1"
)

response = client.chat.completions.create(
    model="${model ?? ""}",
    messages=[],
    extra_body={
        "prompt_version_fqn": "${fqn}",
        "prompt_variables": {}
    }
)

print(response.choices[0].message.content)`;
}
