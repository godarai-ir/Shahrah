# Anthropic live readiness matrix

This command runs the credential-backed release matrix against the GodarAI
Anthropic-native and OpenAI-compatible APIs. Run it once for an API-key-backed
provider account and once for a WIF-backed provider account.

The command reads secrets only from environment variables and writes a
sanitized JSON report with status codes and request IDs. It never writes the
virtual key or Anthropic Admin key.

```powershell
$env:GODARAI_BASE_URL = "https://staging.example.com"
$env:GODARAI_VIRTUAL_KEY = "<staging virtual key>"
$env:GODARAI_ANTHROPIC_PROVIDER_ACCOUNT = "<provider account id>"
$env:GODARAI_ANTHROPIC_AUTH_LABEL = "api_key"

go run ./cmd/anthropic-readiness `
  -run-resources `
  -test-300k `
  -report anthropic-api-key-readiness.json
```

Repeat with the WIF provider account and
`GODARAI_ANTHROPIC_AUTH_LABEL=wif`. To capture the organization and Workspace
Rate Limits API snapshot, provide the separate Admin API credential:

```powershell
$env:ANTHROPIC_ADMIN_KEY = "<admin api key>"
$env:ANTHROPIC_ADMIN_WORKSPACE_ID = "<workspace id>"
```

The resources flags create a small file and Message Batch and can incur
provider charges. The command intentionally fails when any check fails. Live
reports are release evidence and must not be committed if organizational
metadata is considered sensitive.

The local 500 MB bounded-memory gate is separate:

```powershell
$env:ANTHROPIC_READINESS_LOAD_TEST = "1"
Push-Location ../silk-road
go test ./pkg/server `
  -run TestAnthropic500MBUploadUsesBoundedMemory `
  -count=1 -v
Pop-Location
```
