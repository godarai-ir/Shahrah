# Phase 4 design: IRR sunset

Every tenant has billed in credits since migration 072, so the IRR charging path is dead code. Phase 4 removes it instead of putting it behind `BILLING_IRR_RUNTIME_ENABLED`.

## Product-owner decisions (2026-10-17)

- Remove the IRR path outright, without a flag or a deprecation period.
- Gift codes are removed entirely. Free credits are granted with a support-admin credit adjustment.
- Discount codes are removed for now. A discount on packs or plans would be built separately.
- `x-godarai-cost-irr` is removed now, without a deprecation period.

## What stays

- **Toman payments:** plan subscriptions and credit packs are paid in toman through the gateways or by bank transfer, including the pack bank-transfer review in support-admin.
- **Exchange rates:** the price policy and the plan and pack prices need the USD/IRR rate.
- **`TotalCostIRR`:** the pricing engine's rial equivalent is still computed. Provider candidate ranking compares it, and logs and metrics record it. It no longer charges anything. Replacing it with credits in ranking and logs is a separate change.
- **Billing profile:** the tenant's legal or individual identity for payment receipts.

## PR split

### 4a: runtime

- **Budget service:**
  - Every reservation, commit, and BYOK reservation goes through the credit pool and credit rule windows.
  - Removed: the IRR wallet window, its wallet read and cache, the IRR rule windows, and the IRR rule fields in matching.
  - `isCreditTenant` and the `credit_billing_required` refusal of `GET /v1/budget/me` are gone.
- **Proxy:** the wallet debit fallback in `RequestPipeline.Settle` is removed. A request without a reservation is still charged, because the ledger settles its spend log. The `x-godarai-cost-irr` response header is removed.
- **Workers:** the wallet settlement processor and its config are removed. Its repository method stays until 4b, because the conversion tool still settles the IRR window before converting.
- **Stream telemetry:** the `godarai.telemetry` stream event drops `cost_irr` and keeps `credits`. The tenant-admin playground shows credits from `x-godarai-credits-used` and that event.
- **Billing dependencies:** the wallet store is now only an exchange-rate store (`Dependencies.ExchangeRates`). `BILLING_WALLET_CACHE_TTL_SECONDS` and `WALLET_BUDGET_SETTLEMENT_*` are gone.

### 4b: APIs, UI, and schema

- **Tenant API:** wallet, wallet top-ups and their callbacks, non-pack bank transfers, gift and discount code redemption, IRR reports, and IRR documents.
- **Support-admin:** the wallet pages and adjustments, gift and discount code management, the IRR conversion tool and rollback (credit ↔ IRR), and their APIs.
- **Tenant-admin:** the redirect stubs of the retired pages and the unused IRR helpers.
- **Onboarding:** wallet and signup-credit inserts.
- **Migration:**
  - drops the wallet, top-up, invoice, gift and discount, and conversion tables;
  - drops the IRR budget-rule columns `limit_to` and `block_on_budget_exceed`, and `commercial_signup_settings.signup_credit_irr`;
  - drops `tenants.billing_unit`. The profile field goes with it.
- **Docs:** `financial/overview` is rewritten for credits, and `x-godarai-cost-irr` is removed from the header docs.
