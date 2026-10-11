package proxy

import (
	"context"
	"errors"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/shopspring/decimal"
)

type recordingSpendLogger struct {
	logs []platform.SpendLog
	err  error
}

func (r *recordingSpendLogger) LogSpend(_ context.Context, log platform.SpendLog) error {
	r.logs = append(r.logs, log)

	return r.err
}

type staticRateCards struct{ card credits.RateCard }

func (s staticRateCards) CurrentRateCard() credits.RateCard { return s.card }

func TestCreditShadowSpendLoggerAttachesCredits(t *testing.T) {
	t.Parallel()

	next := &recordingSpendLogger{}
	card := credits.RateCard{Version: 2, Multipliers: map[string]decimal.Decimal{}}
	logger := NewCreditShadowSpendLogger(next, staticRateCards{card: card}, nil)

	breakdown := platform.UsageCostBreakdown{
		Model:        "anthropic/claude-sonnet-5",
		RequestType:  "chat_completions",
		TotalCostUSD: 0.009,
		TotalCostIRR: 2430,
		Components: []platform.UsageCostComponent{
			{Dimension: "input_tokens", CostUSD: "0.004"},
			{Dimension: "output_tokens", CostUSD: "0.005"},
		},
	}
	original := platform.SpendLog{
		RequestID:       "req-1",
		Model:           "anthropic/claude-sonnet-5",
		CostUSD:         0.009,
		CostIRR:         2430,
		UsageBreakdown:  &breakdown,
		RequestMetadata: map[string]interface{}{"billing_mode": platform.BillingModeWalletBilled},
	}

	if err := logger.LogSpend(context.Background(), original); err != nil {
		t.Fatalf("LogSpend: %v", err)
	}

	if len(next.logs) != 1 {
		t.Fatalf("forwarded %d logs, want 1", len(next.logs))
	}

	got := next.logs[0]
	if got.CostUSD != original.CostUSD || got.CostIRR != original.CostIRR {
		t.Fatalf("cost changed: got USD %v IRR %v", got.CostUSD, got.CostIRR)
	}

	want := platform.SpendLogCredits{
		Millicredits:    4500,
		RateCardVersion: 2,
		Product:         platform.CreditProductGateway,
		BillingMode:     platform.BillingModeWalletBilled,
		Source:          string(credits.SourceComponents),
	}
	if got.Credits == nil || *got.Credits != want {
		t.Fatalf("credits = %+v, want %+v", got.Credits, want)
	}
}

func TestCreditShadowSpendLoggerForwardsOnComputationError(t *testing.T) {
	t.Parallel()

	next := &recordingSpendLogger{}
	logger := NewCreditShadowSpendLogger(next, nil, nil)
	breakdown := platform.UsageCostBreakdown{
		Components: []platform.UsageCostComponent{{Dimension: "input_tokens", CostUSD: "bad"}},
	}

	if err := logger.LogSpend(context.Background(), platform.SpendLog{RequestID: "req-2", UsageBreakdown: &breakdown}); err != nil {
		t.Fatalf("LogSpend: %v", err)
	}

	if len(next.logs) != 1 || next.logs[0].Credits != nil {
		t.Fatalf("logs = %+v, want one forwarded log without credits", next.logs)
	}
}

func TestCreditShadowSpendLoggerPropagatesWriteError(t *testing.T) {
	t.Parallel()

	writeErr := errors.New("write failed")
	logger := NewCreditShadowSpendLogger(&recordingSpendLogger{err: writeErr}, nil, nil)

	if err := logger.LogSpend(context.Background(), platform.SpendLog{}); !errors.Is(err, writeErr) {
		t.Fatalf("LogSpend error = %v, want %v", err, writeErr)
	}
}

func TestCreditShadowCacheHitIsNoCost(t *testing.T) {
	t.Parallel()

	credit, err := computeSpendLogCredits(platform.SpendLog{Model: "m"}, credits.IdentityRateCard())
	if err != nil {
		t.Fatalf("computeSpendLogCredits: %v", err)
	}

	if credit.Millicredits != 0 || credit.Source != string(credits.SourceNoCost) {
		t.Fatalf("credit = %+v, want zero no_cost", credit)
	}
}

func TestCreditProduct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		log  platform.SpendLog
		want string
	}{
		{name: "default gateway", log: platform.SpendLog{}, want: platform.CreditProductGateway},
		{
			name: "explicit product wins",
			log: platform.SpendLog{RequestMetadata: map[string]interface{}{
				credits.ProductMetadataKey: platform.CreditProductLibreChat,
				"playground_source":        "tenant_admin_playground",
			}},
			want: platform.CreditProductLibreChat,
		},
		{
			name: "playground",
			log:  platform.SpendLog{RequestMetadata: map[string]interface{}{"playground_source": "tenant_admin_playground"}},
			want: platform.CreditProductPlayground,
		},
		{
			name: "embeddings",
			log:  platform.SpendLog{UsageBreakdown: &platform.UsageCostBreakdown{RequestType: "embeddings"}},
			want: platform.CreditProductEmbeddings,
		},
		{
			name: "ocr",
			log:  platform.SpendLog{UsageBreakdown: &platform.UsageCostBreakdown{RequestType: "OCR"}},
			want: platform.CreditProductOCR,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := creditProduct(tt.log); got != tt.want {
				t.Fatalf("creditProduct = %s, want %s", got, tt.want)
			}
		})
	}
}
