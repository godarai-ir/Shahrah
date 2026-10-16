package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	creditBudgetRefusalsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "shahrah_credit_budget_refusals_total",
		Help: "Credit requests refused because a budget was exhausted, by the refusing scope.",
	}, []string{"scope"})
	creditBudgetAlertsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "shahrah_credit_budget_alerts_total",
		Help: "Budget-rule alert thresholds crossed, counted once per rule, scope value, window, and threshold.",
	}, []string{"threshold", "on_exceed"})
	creditBudgetOverageMCTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "shahrah_credit_budget_commit_overage_millicredits_total",
		Help: "Millicredits committed beyond a window's allowance because the usage was already served.",
	})
)

// ObserveCreditBudgetRefusal counts a credit request refused at reservation.
func ObserveCreditBudgetRefusal(scope string) {
	creditBudgetRefusalsTotal.WithLabelValues(scope).Inc()
}

// ObserveCreditBudgetAlert counts one crossed alert threshold.
func ObserveCreditBudgetAlert(threshold int, onExceed string) {
	creditBudgetAlertsTotal.WithLabelValues(strconv.Itoa(threshold), onExceed).Inc()
}

// ObserveCreditBudgetOverage counts spend committed beyond an allowance.
func ObserveCreditBudgetOverage(amountMC int64) {
	if amountMC > 0 {
		creditBudgetOverageMCTotal.Add(float64(amountMC))
	}
}
