package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"go.uber.org/zap"
)

type Exporter interface {
	InsertModelMetricEvent(ctx context.Context, ev platform.ModelMetricEvent) error
}

type Processor struct {
	cfg      Config
	queue    *Queue
	exporter Exporter
	logger   *zap.Logger
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

func NewProcessor(cfg Config, queue *Queue, exporter Exporter, logger *zap.Logger) *Processor {
	return &Processor{
		cfg:      cfg,
		queue:    queue,
		exporter: exporter,
		logger:   logger,
		stopCh:   make(chan struct{}),
	}
}

func (p *Processor) Start() {
	p.wg.Add(1)
	go p.loop()
}

func (p *Processor) Stop(ctx context.Context) error {
	close(p.stopCh)

	done := make(chan struct{})
	go func() {
		defer close(done)

		p.wg.Wait()
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *Processor) loop() {
	defer p.wg.Done()

	ticker := time.NewTicker(p.cfg.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopCh:
			p.flush()
			return
		case <-ticker.C:
			p.flush()
		}
	}
}

func (p *Processor) flush() {
	batch := p.queue.DequeueBatch(p.cfg.BatchSize)
	if len(batch) == 0 {
		return
	}

	for _, ev := range batch {
		if err := p.insertWithRetry(ev); err != nil {
			p.logger.Warn("model metrics insert failed; writing to dlq", zap.Error(err))

			if dlqErr := p.appendDLQ(ev); dlqErr != nil {
				p.logger.Warn("model metrics dlq write failed", zap.Error(dlqErr))
			}
		}
	}
}

func (p *Processor) insertWithRetry(ev platform.ModelMetricEvent) error {
	var lastErr error

	for attempt := 1; attempt <= p.cfg.RetryMaxAttempts; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), p.cfg.ExportTimeout)
		err := p.exporter.InsertModelMetricEvent(ctx, ev)

		cancel()

		if err == nil {
			return nil
		}

		lastErr = err

		time.Sleep(backoffDelay(attempt, p.cfg.RetryBaseDelay, p.cfg.RetryMaxDelay))
	}

	return lastErr
}

func (p *Processor) appendDLQ(ev platform.ModelMetricEvent) error {
	if p.cfg.DLQPath == "" {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(p.cfg.DLQPath), 0o755); err != nil {
		return err
	}

	f, err := os.OpenFile(p.cfg.DLQPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer closeDLQFile(f)

	b, err := json.Marshal(ev)
	if err != nil {
		return err
	}

	if _, err := f.Write(append(b, '\n')); err != nil {
		return err
	}

	return nil
}

func backoffDelay(attempt int, base, max time.Duration) time.Duration {
	if attempt < 1 {
		attempt = 1
	}

	exp := float64(base) * math.Pow(2, float64(attempt-1))

	d := time.Duration(exp)
	if d > max {
		d = max
	}

	jitter := time.Duration(rand.Int63n(int64(d / 5)))

	return d + jitter
}

func closeDLQFile(f *os.File) {
	if f == nil {
		return
	}

	if err := f.Close(); err != nil {
		return
	}
}

func (p *Processor) BuildEvent(input BuildInput) platform.ModelMetricEvent {
	errorType := "none"
	if input.StatusCode >= 500 {
		errorType = "provider_error"
	} else if input.StatusCode == 429 {
		errorType = "rate_limit_error"
	} else if input.StatusCode == 402 {
		errorType = "budget_exceeded"
	} else if input.StatusCode == 401 || input.StatusCode == 403 {
		errorType = "auth_error"
	} else if input.StatusCode >= 400 {
		errorType = "client_error"
	}

	if input.ErrorType != "" {
		errorType = input.ErrorType
	}

	return platform.ModelMetricEvent{
		TS:                   input.TS.UTC(),
		TenantID:             input.TenantID,
		RequestID:            input.RequestID,
		ModelName:            input.ModelName,
		VirtualModelName:     input.VirtualModelName,
		ProviderName:         input.ProviderName,
		ProviderModelName:    input.ProviderModelName,
		RequestType:          input.RequestType,
		UserID:               input.UserID,
		UserEmail:            input.UserEmail,
		VirtualAccountID:     input.VirtualAccountID,
		VirtualAccountName:   input.VirtualAccountName,
		TeamIDs:              input.TeamIDs,
		TeamNames:            input.TeamNames,
		StatusCode:           input.StatusCode,
		ErrorType:            errorType,
		LatencyMS:            input.LatencyMS,
		TTFTMS:               input.TTFTMS,
		ITLMS:                input.ITLMS,
		TPOTMS:               input.TPOTMS,
		InputTokens:          input.InputTokens,
		OutputTokens:         input.OutputTokens,
		CostUSD:              input.CostUSD,
		CacheMode:            normalizedCacheValue(input.CacheMode),
		CacheStatus:          normalizedCacheValue(input.CacheStatus),
		CacheMatch:           normalizedCacheValue(input.CacheMatch),
		CacheSimilarity:      input.CacheSimilarity,
		CacheOriginRequestID: input.CacheOriginRequestID,
		AvoidedInputTokens:   input.AvoidedInputTokens,
		AvoidedOutputTokens:  input.AvoidedOutputTokens,
		AvoidedCostUSD:       input.AvoidedCostUSD,
		AvoidedCostIRR:       input.AvoidedCostIRR,
		EmbeddingCostUSD:     input.EmbeddingCostUSD,
		PricingVersion:       input.PricingVersion,
		Metadata:             sanitizeMetadata(input.Metadata),
	}
}

type BuildInput struct {
	TS                   time.Time
	TenantID             string
	RequestID            string
	ModelName            string
	VirtualModelName     string
	ProviderName         string
	ProviderModelName    string
	RequestType          string
	UserID               string
	UserEmail            string
	VirtualAccountID     string
	VirtualAccountName   string
	TeamIDs              []string
	TeamNames            []string
	StatusCode           int
	ErrorType            string
	LatencyMS            int
	TTFTMS               *int
	ITLMS                *int
	TPOTMS               *int
	InputTokens          int
	OutputTokens         int
	CostUSD              float64
	CacheMode            string
	CacheStatus          string
	CacheMatch           string
	CacheSimilarity      *float64
	CacheOriginRequestID string
	AvoidedInputTokens   int
	AvoidedOutputTokens  int
	AvoidedCostUSD       float64
	AvoidedCostIRR       float64
	EmbeddingCostUSD     float64
	PricingVersion       string
	Metadata             map[string]interface{}
}

func normalizedCacheValue(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "none"
	}

	return value
}

func sanitizeMetadata(in map[string]interface{}) map[string]interface{} {
	if len(in) == 0 {
		return map[string]interface{}{}
	}

	allowed := map[string]struct{}{
		"environment": {},
		"feature":     {},
		"endpoint":    {},
		"key_type":    {},
		"org_unit_id": {},
		"team_id":     {},
		"tenant_slug": {},
	}
	out := make(map[string]interface{}, len(in))
	count := 0

	for k, v := range in {
		if _, ok := allowed[k]; !ok {
			continue
		}

		if len(k) > 64 {
			continue
		}

		switch t := v.(type) {
		case string:
			if len(t) > 256 {
				out[k] = t[:256]
			} else {
				out[k] = t
			}
		default:
			out[k] = fmt.Sprintf("%v", t)
		}

		count++
		if count >= 16 {
			break
		}
	}

	return out
}
