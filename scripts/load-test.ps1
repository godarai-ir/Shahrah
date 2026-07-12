[CmdletBinding()]
param(
    [switch]$Smoke,
    [switch]$SkipBuild,
    [switch]$SkipMigrate,
    [double]$MaxHostCPU = 85,
    [int]$FixedRPS = 0,
    [string]$ReportPath = "artifacts/loadtest/report.json"
)

$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot
$baseCompose = Join-Path $repoRoot "infra/compose/local.yml"
$loadCompose = Join-Path $repoRoot "services/control-plane/cmd/gateway-loadtest/compose.load.yml"
$masterKey = "sk-dev-master-key-12345"
$tenantSlug = "loadtest"
$model = "mock:loadtest:mock-model"

function Invoke-Compose {
    param([Parameter(Mandatory = $true)][string[]]$ComposeArgs)
    & docker compose -f $baseCompose -f $loadCompose @ComposeArgs
    if ($LASTEXITCODE -ne 0) {
        throw "docker compose failed: $($ComposeArgs -join ' ')"
    }
}

function Wait-ControlPlane {
    $deadline = (Get-Date).AddMinutes(3)
    while ((Get-Date) -lt $deadline) {
        try {
            $response = Invoke-WebRequest -UseBasicParsing -Uri "http://127.0.0.1:4000/ready" -TimeoutSec 5
            if ($response.StatusCode -eq 200) {
                return
            }
        } catch {
            Start-Sleep -Seconds 2
        }
    }
    throw "control-plane did not become ready within three minutes"
}

function Invoke-ControlPlaneJSON {
    param(
        [string]$Path,
        [string]$Body
    )
    $headers = @{
        Authorization = "Bearer $masterKey"
        "X-GodarAI-Tenant" = $tenantSlug
    }
    Invoke-RestMethod -Method Post -Uri "http://127.0.0.1:4000$Path" -Headers $headers -ContentType "application/json" -Body $Body
}

Push-Location $repoRoot
try {
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        throw "Docker is not installed or is not on PATH"
    }
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        throw "Go is not installed or is not on PATH"
    }

    $services = @("postgres", "redis", "clickhouse", "mailpit", "silk-road", "tunnel-server", "tunnel-client", "control-plane")
    if ($SkipBuild) {
        Invoke-Compose -ComposeArgs (@("up", "-d", "--no-build") + $services)
    } else {
        Invoke-Compose -ComposeArgs (@("up", "-d", "--build") + $services)
    }

    if (-not $SkipMigrate) {
        Invoke-Compose -ComposeArgs @("run", "--rm", "migrate")
    }
    Wait-ControlPlane

    $seedSQL = @'
INSERT INTO tenants (slug, domain, status)
VALUES ('loadtest', 'loadtest.localhost', 'active')
ON CONFLICT (slug) DO UPDATE SET domain = EXCLUDED.domain, status = 'active';

DELETE FROM verification_tokens
WHERE tenant_id = (SELECT tenant_id FROM tenants WHERE slug = 'loadtest')
  AND key_alias = 'loadtest-key';

DELETE FROM model_deployments
WHERE tenant_id = (SELECT tenant_id FROM tenants WHERE slug = 'loadtest')
  AND model_name IN ('loadtest-mock', 'mock:loadtest:mock-model');

DELETE FROM tenant_provider_accounts
WHERE tenant_id = (SELECT tenant_id FROM tenants WHERE slug = 'loadtest')
  AND name = 'loadtest';

INSERT INTO tenant_provider_accounts (tenant_id, provider, name, auth_mode, has_credential)
SELECT tenant_id, 'mock', 'loadtest', 'no_credential', false
FROM tenants WHERE slug = 'loadtest';

INSERT INTO tenant_provider_account_models (account_id, model, enabled, types, pricing_mode)
SELECT account_id, 'mock-model', true, ARRAY['chat']::text[], 'public'
FROM tenant_provider_accounts
WHERE tenant_id = (SELECT tenant_id FROM tenants WHERE slug = 'loadtest')
  AND name = 'loadtest';

INSERT INTO tenant_provider_account_access (account_id, role, principal_type, access_scope)
SELECT account_id, 'user', 'everyone', 'account'
FROM tenant_provider_accounts
WHERE tenant_id = (SELECT tenant_id FROM tenants WHERE slug = 'loadtest')
  AND name = 'loadtest';

INSERT INTO wallets (tenant_id, balance_irr, pending_irr)
SELECT tenant_id, 1000000000000, 0 FROM tenants WHERE slug = 'loadtest'
ON CONFLICT (tenant_id) DO UPDATE SET balance_irr = EXCLUDED.balance_irr, pending_irr = 0, updated_at = NOW();

DELETE FROM tenant_subscriptions
WHERE tenant_id = (SELECT tenant_id FROM tenants WHERE slug = 'loadtest');

INSERT INTO tenant_subscriptions (
    tenant_id, plan_id, version_id, status, billing_interval,
    current_period_start, current_period_end, price_snapshot_json, metadata
)
SELECT t.tenant_id, p.plan_id, pv.version_id, 'active', 'monthly',
       NOW(), NOW() + INTERVAL '30 days', '{}'::jsonb, '{"source":"loadtest"}'::jsonb
FROM tenants t
JOIN commercial_plans p ON p.slug = 'enterprise'
JOIN LATERAL (
    SELECT version_id FROM plan_versions
    WHERE plan_id = p.plan_id AND status = 'published'
    ORDER BY version DESC LIMIT 1
) pv ON true
WHERE t.slug = 'loadtest';
'@
    $seedSQL | & docker compose -f $baseCompose -f $loadCompose exec -T postgres psql -v ON_ERROR_STOP=1 -U postgres -d sharah_db
    if ($LASTEXITCODE -ne 0) {
        throw "failed to seed the load-test tenant"
    }

    Invoke-Compose -ComposeArgs @("restart", "control-plane")
    Wait-ControlPlane

    Invoke-ControlPlaneJSON -Path "/api/v1/models" -Body (@{
        model_name = $model
        provider = "mock"
        model = "mock-model"
        api_key = "mock-no-secret"
        rpm = 10000000
        tpm = 1000000000
        timeout = 10
    } | ConvertTo-Json -Compress) | Out-Null

    $keyResponse = Invoke-ControlPlaneJSON -Path "/api/v1/keys" -Body (@{
        key_alias = "loadtest-key"
        models = @($model)
    } | ConvertTo-Json -Compress)
    $apiKey = $keyResponse.key
    if ([string]::IsNullOrWhiteSpace($apiKey)) {
        throw "control-plane did not return the generated load-test API key"
    }

    $cpu = Get-CimInstance Win32_Processor | Select-Object -First 1
    $computer = Get-CimInstance Win32_ComputerSystem
    $dockerInfo = & docker info --format 'Docker CPUs={{.NCPU}}, memory={{.MemTotal}} bytes'
    $machine = "$($cpu.Name.Trim()); $($cpu.NumberOfCores) cores/$($cpu.NumberOfLogicalProcessors) logical; $([math]::Round($computer.TotalPhysicalMemory / 1GB, 2)) GB RAM; $dockerInfo"

    if (-not [System.IO.Path]::IsPathRooted($ReportPath)) {
        $ReportPath = Join-Path $repoRoot $ReportPath
    }

    $arguments = @(
        "-C", "services/control-plane", "run", "./cmd/gateway-loadtest",
        "--api-key", $apiKey,
        "--tenant", $tenantSlug,
        "--model", $model,
        "--report", $ReportPath,
        "--machine", $machine,
        "--max-host-cpu", $MaxHostCPU.ToString([System.Globalization.CultureInfo]::InvariantCulture)
    )
    if ($Smoke) {
        $arguments += @(
            "--warmup-duration", "10s",
            "--warmup-rps", "5",
            "--ramp-stage-duration", "10s",
            "--ramp-budget", "50s",
            "--ramp-start-rps", "10",
            "--max-ramp-rps", "160",
            "--stable-duration", "20s",
            "--burst-duration", "10s",
            "--max-host-cpu", "100"
        )
    }
    if ($FixedRPS -gt 0) {
        $arguments += @("--fixed-rps", $FixedRPS.ToString([System.Globalization.CultureInfo]::InvariantCulture))
    }

    $env:SHAHRAH_LOADTEST_API_KEY = $apiKey
    & go @arguments
    if ($LASTEXITCODE -ne 0) {
        throw "load-test command failed"
    }
} finally {
    Remove-Item Env:SHAHRAH_LOADTEST_API_KEY -ErrorAction SilentlyContinue
    Pop-Location
}
