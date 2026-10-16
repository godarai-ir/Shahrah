package credits

// SettlementResult summarizes one settlement pass over spend logs.
type SettlementResult struct {
	// Scanned is the number of spend logs the cursor moved past (0 when the pass
	// stopped at a failing row).
	Scanned int
	// Settled is the number of new consume entries.
	Settled int
	// KeyConflicts counts rows whose consume key already existed: replays after a
	// crash, the recovery sweep, or a duplicate (request_id, product) pair.
	KeyConflicts int
	// DebtMC is the consumption that found no spendable lot.
	DebtMC int64
	// LastSpendLogID is the cursor after the pass.
	LastSpendLogID int64
	// Locked is false when another worker held the settlement lock.
	Locked bool
}

// ExpiryResult summarizes one expiry pass.
type ExpiryResult struct {
	ExpiredLots int
	ExpiredMC   int64
	TenantIDs   []string
}

// GrantResult summarizes one plan and trial grant pass.
type GrantResult struct {
	Granted    int
	GrantedMC  int64
	Duplicates int
	// OverageMC is debt invoiced as overage before new plan slices.
	OverageMC int64
}

// ReconciliationDiff is one failed reconciliation check.
type ReconciliationDiff struct {
	TenantID   string
	Check      string
	ExpectedMC int64
	ActualMC   int64
}

// Reconciliation checks reported in ReconciliationDiff.Check.
const (
	CheckLedgerSum        = "ledger_sum"
	CheckAvailable        = "available"
	CheckDebt             = "debt"
	CheckLotProjection    = "lot_projection"
	CheckEntryAllocations = "entry_allocations"
	CheckDefaultPartition = "default_partition"
)
