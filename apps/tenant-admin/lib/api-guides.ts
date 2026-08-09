export function buildBashCurlGuideCommand(input: {
  baseUrl: string;
  endpoint: string;
  payload: Record<string, unknown>;
}): string {
  if (input.endpoint === "/v1/audio/transcriptions") {
    const model = escapeShellSingleQuoted(String(input.payload.model ?? ""));
    const language = escapeShellSingleQuoted(String(input.payload.language ?? "fa"));
    const responseFormat = escapeShellSingleQuoted(String(input.payload.response_format ?? "json"));
    return [
      `curl --location '${input.baseUrl}${input.endpoint}' \\`,
      "  --header 'Authorization: Bearer <PAT_TOKEN>' \\",
      `  --form 'model=${model}' \\`,
      `  --form 'language=${language}' \\`,
      `  --form 'response_format=${responseFormat}' \\`,
      "  --form 'file=@sample.wav;type=audio/wav'",
    ].join("\n");
  }

  const payload = JSON.stringify(input.payload, null, 2).replaceAll("'", "'\\''");
  if (input.endpoint === "/v1/audio/speech") {
    return [
      `curl --location '${input.baseUrl}${input.endpoint}' \\`,
      "  --header 'Authorization: Bearer <PAT_TOKEN>' \\",
      "  --header 'Content-Type: application/json' \\",
      `  --data '${payload}' \\`,
      "  --output speech.mp3",
    ].join("\n");
  }

  return [
    `curl --location '${input.baseUrl}${input.endpoint}' \\`,
    "  --header 'Authorization: Bearer <PAT_TOKEN>' \\",
    "  --header 'Content-Type: application/json' \\",
    `  --data '${payload}'`,
  ].join("\n");
}

function escapeShellSingleQuoted(value: string): string {
  return value.replaceAll("'", "'\\''");
}
