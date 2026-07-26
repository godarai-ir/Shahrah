package main

import (
	"testing"
)

func TestQuoteClickHouseIdentifier(t *testing.T) {
	t.Parallel()

	if got, want := quoteClickHouseIdentifier("request_logs_ch"), "`request_logs_ch`"; got != want {
		t.Fatalf("quoteClickHouseIdentifier() = %q, want %q", got, want)
	}
	if got, want := quoteClickHouseIdentifier("odd`name"), "`odd``name`"; got != want {
		t.Fatalf("quoteClickHouseIdentifier() = %q, want %q", got, want)
	}
}

func TestFormatBlockersIsStable(t *testing.T) {
	t.Parallel()

	blockers := map[string]string{
		"wallets":              "wallet_transactions_wallet_id_fkey",
		"tenant_subscriptions": "invoices_subscription_id_fkey",
	}
	want := "tenant_subscriptions (invoices_subscription_id_fkey), wallets (wallet_transactions_wallet_id_fkey)"

	if got := formatBlockers(blockers); got != want {
		t.Fatalf("formatBlockers() = %q, want %q", got, want)
	}
}
