[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [ValidateNotNullOrEmpty()]
    [string]$OutputDirectory,

    [ValidatePattern('^[A-Za-z0-9._-]+$')]
    [string]$ActiveVersion = 'prod-v1'
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Write-Utf8SecretFile {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Path,

        [Parameter(Mandatory = $true)]
        [string]$Value
    )

    $encoding = [System.Text.UTF8Encoding]::new($false)
    $bytes = $encoding.GetBytes($Value + [Environment]::NewLine)
    $stream = $null

    try {
        $stream = [System.IO.File]::Open(
            $Path,
            [System.IO.FileMode]::CreateNew,
            [System.IO.FileAccess]::Write,
            [System.IO.FileShare]::None
        )
        $stream.Write($bytes, 0, $bytes.Length)
        $stream.Flush($true)
    }
    finally {
        if ($null -ne $stream) {
            $stream.Dispose()
        }
        [Array]::Clear($bytes, 0, $bytes.Length)
    }
}

$repositoryRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..')).TrimEnd(
    [System.IO.Path]::DirectorySeparatorChar,
    [System.IO.Path]::AltDirectorySeparatorChar
)
$resolvedOutputDirectory = [System.IO.Path]::GetFullPath($OutputDirectory).TrimEnd(
    [System.IO.Path]::DirectorySeparatorChar,
    [System.IO.Path]::AltDirectorySeparatorChar
)
$pathComparison = if ($env:OS -eq 'Windows_NT') {
    [System.StringComparison]::OrdinalIgnoreCase
} else {
    [System.StringComparison]::Ordinal
}

$isRepositoryRoot = $resolvedOutputDirectory.Equals($repositoryRoot, $pathComparison)
$isInsideRepository = $resolvedOutputDirectory.StartsWith(
    $repositoryRoot + [System.IO.Path]::DirectorySeparatorChar,
    $pathComparison
)

if ($isRepositoryRoot -or $isInsideRepository) {
    throw 'OutputDirectory must be outside the repository to prevent accidental commits.'
}

if ([System.IO.Directory]::Exists($resolvedOutputDirectory) -or [System.IO.File]::Exists($resolvedOutputDirectory)) {
    throw "OutputDirectory already exists. Choose a new empty path: $resolvedOutputDirectory"
}

$activeVersionPath = Join-Path $resolvedOutputDirectory 'TENANT_PROVIDER_CREDENTIAL_ACTIVE_KEK_VERSION.txt'
$keksPath = Join-Path $resolvedOutputDirectory 'TENANT_PROVIDER_CREDENTIAL_KEKS_JSON.json'
$fingerprintPath = Join-Path $resolvedOutputDirectory 'TENANT_PROVIDER_CREDENTIAL_FINGERPRINT_KEY.txt'
$createdFiles = [System.Collections.Generic.List[string]]::new()
$tenantKek = [byte[]]::new(32)
$tenantFingerprintKey = [byte[]]::new(32)

try {
    [System.IO.Directory]::CreateDirectory($resolvedOutputDirectory) | Out-Null

    [Security.Cryptography.RandomNumberGenerator]::Fill($tenantKek)
    [Security.Cryptography.RandomNumberGenerator]::Fill($tenantFingerprintKey)

    $encodedKeks = [ordered]@{}
    $encodedKeks[$ActiveVersion] = [Convert]::ToBase64String($tenantKek)
    $encodedKeksJson = $encodedKeks | ConvertTo-Json -Compress
    $encodedFingerprintKey = [Convert]::ToBase64String($tenantFingerprintKey)

    Write-Utf8SecretFile -Path $activeVersionPath -Value $ActiveVersion
    $createdFiles.Add($activeVersionPath)
    Write-Utf8SecretFile -Path $keksPath -Value $encodedKeksJson
    $createdFiles.Add($keksPath)
    Write-Utf8SecretFile -Path $fingerprintPath -Value $encodedFingerprintKey
    $createdFiles.Add($fingerprintPath)

    if ($env:OS -eq 'Windows_NT') {
        $currentIdentity = [Security.Principal.WindowsIdentity]::GetCurrent().Name
        $grantRule = '{0}:(OI)(CI)F' -f $currentIdentity
        & icacls $resolvedOutputDirectory /inheritance:r /grant:r $grantRule | Out-Null
        if ($LASTEXITCODE -ne 0) {
            throw 'Failed to restrict the output directory ACL to the current Windows user.'
        }
    }
    else {
        & chmod 700 $resolvedOutputDirectory
        foreach ($createdFile in $createdFiles) {
            & chmod 600 $createdFile
        }
    }
}
catch {
    foreach ($createdFile in $createdFiles) {
        if ([System.IO.File]::Exists($createdFile)) {
            Remove-Item -LiteralPath $createdFile -Force
        }
    }
    if ([System.IO.Directory]::Exists($resolvedOutputDirectory)) {
        Remove-Item -LiteralPath $resolvedOutputDirectory -Force -ErrorAction SilentlyContinue
    }
    throw
}
finally {
    [Array]::Clear($tenantKek, 0, $tenantKek.Length)
    [Array]::Clear($tenantFingerprintKey, 0, $tenantFingerprintKey.Length)
}

Write-Host 'Tenant provider credential encryption values were created in separate files:'
foreach ($createdFile in $createdFiles) {
    Write-Host "- $createdFile"
}
Write-Warning 'These files contain production encryption keys. Move them to an encrypted secret store, then securely remove the local copies.'
