# Z.ai readiness runner

This command sends a bounded, billable smoke matrix through GodarAI for the
four curated Z.ai models. It never receives the upstream Z.ai key.

Required environment variables:

- `GODARAI_ZAI_ALLOW_LIVE=true`
- `GODARAI_BASE_URL`
- `GODARAI_VIRTUAL_KEY`
- `GODARAI_ZAI_PROVIDER_ACCOUNT`
- `GODARAI_ZAI_MAX_SPEND_USD` (explicitly approved; minimum USD 0.02)
- HTTPS fixture URLs in `GODARAI_ZAI_IMAGE_URL`, `GODARAI_ZAI_VIDEO_URL`,
  `GODARAI_ZAI_DOCUMENT_URL`, and `GODARAI_ZAI_OCR_FILE_URL`

Run from `services/control-plane`:

```powershell
go run ./cmd/zai-readiness -report zai-readiness-report.json
```

The report contains only model IDs, statuses and usage. A passing run does not
set `production_ready`; release owners must approve the conflict gates,
billing reconciliation, privacy, load, observability, canary, and rollback.
