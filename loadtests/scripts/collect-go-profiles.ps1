param(
  [string]$BaseUrl = "http://localhost:4000",
  [string]$OutDir = "loadtests/results/profiles",
  [int]$IntervalSeconds = 60,
  [int]$CpuSeconds = 30,
  [int]$Samples = 10
)

$ErrorActionPreference = "Stop"
$base = $BaseUrl.TrimEnd("/")
$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$runDir = Join-Path $OutDir $timestamp
New-Item -ItemType Directory -Force -Path $runDir | Out-Null

function Save-Endpoint {
  param(
    [string]$Path,
    [string]$FileName,
    [switch]$Optional
  )

  $url = "$base$Path"
  $target = Join-Path $runDir $FileName
  try {
    Invoke-WebRequest -Uri $url -OutFile $target -UseBasicParsing | Out-Null
  } catch {
    if ($Optional) {
      Write-Warning "Failed to collect optional endpoint $Path`: $($_.Exception.Message)"
      return
    }

    throw
  }
}

Write-Host "Writing profiles to $runDir"

for ($i = 1; $i -le $Samples; $i++) {
  $sample = "{0:D3}" -f $i
  Write-Host "Collecting sample $sample"

  Save-Endpoint -Path "/metrics" -FileName "$sample-metrics.prom" -Optional
  Save-Endpoint -Path "/debug/pprof/profile?seconds=$CpuSeconds" -FileName "$sample-cpu.pb.gz"
  Save-Endpoint -Path "/debug/pprof/heap" -FileName "$sample-heap.pb.gz"
  Save-Endpoint -Path "/debug/pprof/goroutine" -FileName "$sample-goroutine.pb.gz"
  Save-Endpoint -Path "/debug/pprof/goroutine?debug=1" -FileName "$sample-goroutine.txt"
  Save-Endpoint -Path "/debug/pprof/mutex" -FileName "$sample-mutex.pb.gz"
  Save-Endpoint -Path "/debug/pprof/block" -FileName "$sample-block.pb.gz"
  Save-Endpoint -Path "/debug/pprof/trace?seconds=5" -FileName "$sample-trace.out"

  if ($i -lt $Samples) {
    Start-Sleep -Seconds $IntervalSeconds
  }
}

Write-Host "Done. Inspect with:"
Write-Host "  go tool pprof -http=:0 $base/debug/pprof/profile?seconds=$CpuSeconds"
Write-Host "  go tool pprof -http=:0 $(Join-Path $runDir '001-heap.pb.gz')"
Write-Host "  go tool trace $(Join-Path $runDir '001-trace.out')"
