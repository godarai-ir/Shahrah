# Gateway load test

This command estimates sustainable `/v1/chat/completions` throughput with an open-loop workload. It performs a warm-up, adaptive ramp, refinement after the first failure, sustained verification at an 80% safety target, and a short burst.

Use the repository runner from PowerShell:

```powershell
./scripts/load-test.ps1
```

The default run takes roughly 19 minutes after images are built. A quick validation is available with `./scripts/load-test.ps1 -Smoke`.

The test deliberately sends a unique prompt on every request so the response cache cannot inflate throughput. A stage passes when all of these are true:

- error rate is below 1%;
- p95 latency is at most 500 ms;
- achieved throughput is at least 98% of the requested rate;
- average host CPU is at most 85% when host counters are available.

Reports are written as JSON and Markdown under `artifacts/loadtest/` by default. The API key is passed through an environment variable and is never written into the report.
