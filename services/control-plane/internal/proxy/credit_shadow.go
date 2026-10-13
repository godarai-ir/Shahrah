package proxy

import (
	"context"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"go.uber.org/zap"
)

// CreditShadowSpendLogger attaches shadow credit computation to every spend
// log before forwarding it. It never changes cost, IRR, or enforcement data and
// never fails a spend write because of a credit computation problem.
type CreditShadowSpendLogger struct {
	next      SpendLogger
	rateCards billing.CreditRateCardProvider
	logger    *zap.Logger
}

// NewCreditShadowSpendLogger wraps next with shadow credit computation.
func NewCreditShadowSpendLogger(
	next SpendLogger,
	rateCards billing.CreditRateCardProvider,
	logger *zap.Logger,
) *CreditShadowSpendLogger {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &CreditShadowSpendLogger{next: next, rateCards: rateCards, logger: logger}
}

// LogSpend computes credits for the spend log and forwards it.
func (l *CreditShadowSpendLogger) LogSpend(ctx context.Context, log platform.SpendLog) error {
	if l == nil || l.next == nil {
		return nil
	}

	if log.Credits == nil {
		credit, err := computeSpendLogCredits(log, l.currentRateCard())
		if err != nil {
			l.logger.Warn("shadow credit computation failed", zap.Error(err), zap.String("request_id", log.RequestID))
		} else {
			log.Credits = credit
		}
	}

	return l.next.LogSpend(ctx, log)
}

func (l *CreditShadowSpendLogger) currentRateCard() credits.RateCard {
	if l.rateCards == nil {
		return credits.IdentityRateCard()
	}

	return l.rateCards.CurrentRateCard()
}

// computeSpendLogCredits derives the credit charge and labels for one spend log.
func computeSpendLogCredits(log platform.SpendLog, card credits.RateCard) (*platform.SpendLogCredits, error) {
	breakdown := platform.UsageCostBreakdown{Model: log.Model, TotalCostUSD: log.CostUSD}
	if log.UsageBreakdown != nil {
		breakdown = *log.UsageBreakdown
		if strings.TrimSpace(breakdown.Model) == "" {
			breakdown.Model = log.Model
		}
	}

	charge, err := credits.FromBreakdown(breakdown, card)
	if err != nil {
		return nil, err
	}

	return &platform.SpendLogCredits{
		Millicredits:    int64(charge.Millicredits),
		RateCardVersion: charge.RateCardVersion,
		Product:         creditProduct(log),
		BillingMode:     metadataText(log.RequestMetadata, "billing_mode"),
		Source:          string(charge.Source),
	}, nil
}

// creditProduct labels the GodarAI surface that produced a spend log.
func creditProduct(log platform.SpendLog) string {
	requestType := ""
	if log.UsageBreakdown != nil {
		requestType = log.UsageBreakdown.RequestType
	}

	return credits.Product(log.RequestMetadata, requestType)
}

func metadataText(metadata map[string]interface{}, key string) string {
	value, _ := metadataString(metadata, key)

	return value
}
