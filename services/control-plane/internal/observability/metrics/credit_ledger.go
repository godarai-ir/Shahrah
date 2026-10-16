package metrics

import (
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	creditSettlementEntriesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "shahrah_credit_settlement_entries_total",
		Help: "Consume entries written to the credit ledger.",
	})
	creditSettlementKeyConflictsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "shahrah_credit_settlement_key_conflicts_total",
		Help: "Spend logs skipped because their consume key already existed (replays or duplicate request ids).",
	})
	creditSettlementRecoveredTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "shahrah_credit_settlement_recovered_total",
		Help: "Consume entries the recovery sweep wrote for spend logs the cursor had passed over.",
	})
	creditSettlementDebtMCTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "shahrah_credit_settlement_debt_millicredits_total",
		Help: "Consumption that found no spendable lot and was recorded as debt.",
	})
	creditLedgerJobFailuresTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "shahrah_credit_ledger_job_failures_total",
		Help: "Credit ledger background job failures by job.",
	}, []string{"job"})
	creditGrantedMCTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "shahrah_credit_plan_granted_millicredits_total",
		Help: "Plan and trial credits granted by the grant worker.",
	})
	creditExpiredMCTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "shahrah_credit_expired_millicredits_total",
		Help: "Credits expired from lots.",
	})
	creditReconciliationFailuresTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "shahrah_credit_reconciliation_failures_total",
		Help: "Credit reconciliation checks that found a difference, by check.",
	}, []string{"check"})
	creditReconciliationDiffMC = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shahrah_credit_reconciliation_diff_millicredits",
		Help: "Actual minus expected millicredits for each failed check in the latest reconciliation.",
	}, []string{"tenant_id", "check"})
)

// ObserveCreditSettlement records one settlement or sweep pass.
func ObserveCreditSettlement(result credits.SettlementResult, sweep bool) {
	if sweep {
		creditSettlementRecoveredTotal.Add(float64(result.Settled))
	}

	creditSettlementEntriesTotal.Add(float64(result.Settled))
	creditSettlementDebtMCTotal.Add(float64(result.DebtMC))

	if !sweep {
		creditSettlementKeyConflictsTotal.Add(float64(result.KeyConflicts))
	}
}

// ObserveCreditGrants records one plan and trial grant pass.
func ObserveCreditGrants(result credits.GrantResult) {
	creditGrantedMCTotal.Add(float64(result.GrantedMC))
}

// ObserveCreditExpiry records one expiry pass.
func ObserveCreditExpiry(result credits.ExpiryResult) {
	creditExpiredMCTotal.Add(float64(result.ExpiredMC))
}

// ObserveCreditLedgerJobFailure counts a failed credit ledger job.
func ObserveCreditLedgerJobFailure(job string) {
	creditLedgerJobFailuresTotal.WithLabelValues(job).Inc()
}

// ObserveCreditReconciliation replaces the reconciliation gauges with the
// differences found by the latest run.
func ObserveCreditReconciliation(diffs []credits.ReconciliationDiff) {
	creditReconciliationDiffMC.Reset()

	for _, diff := range diffs {
		creditReconciliationFailuresTotal.WithLabelValues(diff.Check).Inc()
		creditReconciliationDiffMC.WithLabelValues(diff.TenantID, diff.Check).Set(float64(diff.ActualMC - diff.ExpectedMC))
	}
}
