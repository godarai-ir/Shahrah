package billing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"go.uber.org/zap"
)

const (
	defaultNobitexBaseURL  = "https://api.nobitex.ir"
	defaultNobitexSymbol   = "USDTIRT"
	defaultNobitexInterval = 5 * time.Minute
	defaultNobitexTimeout  = 10 * time.Second
)

type exchangeRateUpdaterStore interface {
	UpsertExchangeRate(ctx context.Context, rate platform.ExchangeRate) (*platform.ExchangeRate, error)
	TryExchangeRateUpdateLock(ctx context.Context) (bool, func(), error)
}

// NobitexExchangeRateConfig controls the Nobitex USDT/IRT exchange-rate updater.
type NobitexExchangeRateConfig struct {
	Enabled   bool
	BaseURL   string
	Symbol    string
	Source    string
	Interval  time.Duration
	Timeout   time.Duration
	MinTrades int
}

// LoadNobitexExchangeRateConfigFromEnv reads Nobitex updater settings.
func LoadNobitexExchangeRateConfigFromEnv(rateSource string) NobitexExchangeRateConfig {
	symbol := strings.ToUpper(strings.TrimSpace(os.Getenv("NOBITEX_EXCHANGE_RATE_SYMBOL")))
	if symbol == "" {
		symbol = defaultNobitexSymbol
	}

	source := strings.TrimSpace(os.Getenv("NOBITEX_EXCHANGE_RATE_SOURCE"))
	if source == "" {
		source = "nobitex:" + symbol
	}

	baseURL := strings.TrimSpace(os.Getenv("NOBITEX_EXCHANGE_RATE_BASE_URL"))
	if baseURL == "" {
		baseURL = defaultNobitexBaseURL
	}

	return NobitexExchangeRateConfig{
		Enabled:   boolEnv("NOBITEX_EXCHANGE_RATE_ENABLED", strings.EqualFold(strings.TrimSpace(rateSource), "nobitex")),
		BaseURL:   baseURL,
		Symbol:    symbol,
		Source:    source,
		Interval:  secondsEnv("NOBITEX_EXCHANGE_RATE_INTERVAL_SECONDS", defaultNobitexInterval),
		Timeout:   secondsEnv("NOBITEX_EXCHANGE_RATE_TIMEOUT_SECONDS", defaultNobitexTimeout),
		MinTrades: intEnv("NOBITEX_EXCHANGE_RATE_MIN_TRADES", 1),
	}
}

// NobitexExchangeRateProcessor periodically fetches USDTIRT market trades and stores a rate snapshot.
type NobitexExchangeRateProcessor struct {
	cfg    NobitexExchangeRateConfig
	store  exchangeRateUpdaterStore
	client *http.Client
	logger *zap.Logger
	cancel context.CancelFunc
	done   chan struct{}
	once   sync.Once
}

// NewNobitexExchangeRateProcessor creates a Nobitex exchange-rate updater.
func NewNobitexExchangeRateProcessor(cfg NobitexExchangeRateConfig, store exchangeRateUpdaterStore, logger *zap.Logger) *NobitexExchangeRateProcessor {
	if logger == nil {
		logger = zap.NewNop()
	}
	cfg = normalizeNobitexExchangeRateConfig(cfg)

	return &NobitexExchangeRateProcessor{
		cfg:    cfg,
		store:  store,
		client: &http.Client{Timeout: cfg.Timeout},
		logger: logger,
		done:   make(chan struct{}),
	}
}

func normalizeNobitexExchangeRateConfig(cfg NobitexExchangeRateConfig) NobitexExchangeRateConfig {
	cfg.BaseURL = strings.TrimSpace(cfg.BaseURL)
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultNobitexBaseURL
	}
	cfg.Symbol = strings.ToUpper(strings.TrimSpace(cfg.Symbol))
	if cfg.Symbol == "" {
		cfg.Symbol = defaultNobitexSymbol
	}
	cfg.Source = strings.TrimSpace(cfg.Source)
	if cfg.Source == "" {
		cfg.Source = "nobitex:" + cfg.Symbol
	}
	if cfg.Interval <= 0 {
		cfg.Interval = defaultNobitexInterval
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = defaultNobitexTimeout
	}
	if cfg.MinTrades <= 0 {
		cfg.MinTrades = 1
	}
	return cfg
}

// Start launches the Nobitex updater loop when enabled.
func (p *NobitexExchangeRateProcessor) Start() {
	if p == nil || !p.cfg.Enabled || p.store == nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel
	go p.run(ctx)
}

// Stop waits for the updater loop to exit.
func (p *NobitexExchangeRateProcessor) Stop() {
	if p == nil {
		return
	}

	p.once.Do(func() {
		if p.cancel != nil {
			p.cancel()
			<-p.done
			return
		}
		close(p.done)
	})
}

func (p *NobitexExchangeRateProcessor) run(ctx context.Context) {
	defer close(p.done)

	p.processWithTimeout(ctx)

	ticker := time.NewTicker(p.cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.processWithTimeout(ctx)
		}
	}
}

func (p *NobitexExchangeRateProcessor) processWithTimeout(ctx context.Context) {
	workCtx, cancel := context.WithTimeout(ctx, p.cfg.Timeout)
	defer cancel()

	p.process(workCtx)
}

func (p *NobitexExchangeRateProcessor) process(ctx context.Context) {
	started := time.Now()
	acquired, release, err := p.store.TryExchangeRateUpdateLock(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		p.logger.Warn("nobitex exchange rate lock failed", zap.Error(err), zap.String("symbol", p.cfg.Symbol))
		return
	}
	if !acquired {
		p.logger.Debug("nobitex exchange rate update skipped because another pod holds the lock", zap.String("symbol", p.cfg.Symbol))
		return
	}
	if release != nil {
		defer release()
	}

	rate, trades, err := p.fetchUSDtoIRR(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		p.logger.Warn("nobitex exchange rate fetch failed", zap.Error(err), zap.String("symbol", p.cfg.Symbol))
		return
	}

	if trades < p.cfg.MinTrades {
		p.logger.Warn("nobitex exchange rate skipped because trade count is below minimum", zap.Int("trades", trades), zap.Int("min_trades", p.cfg.MinTrades), zap.String("symbol", p.cfg.Symbol))
		return
	}

	_, err = p.store.UpsertExchangeRate(ctx, platform.ExchangeRate{
		USDtoIRR: rate,
		Source:   p.cfg.Source,
	})
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		p.logger.Warn("nobitex exchange rate upsert failed", zap.Error(err), zap.String("symbol", p.cfg.Symbol), zap.Float64("usd_to_irr", rate))
		return
	}

	p.logger.Info("nobitex exchange rate updated", zap.String("symbol", p.cfg.Symbol), zap.Float64("usd_to_irr", rate), zap.Int("trades", trades), zap.Duration("elapsed", time.Since(started)))
}

func (p *NobitexExchangeRateProcessor) fetchUSDtoIRR(ctx context.Context) (float64, int, error) {
	endpoint, err := nobitexTradesURL(p.cfg.BaseURL, p.cfg.Symbol)
	if err != nil {
		return 0, 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, 0, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return 0, 0, fmt.Errorf("nobitex returned status %d", resp.StatusCode)
	}

	var payload nobitexTradesResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return 0, 0, err
	}
	if !strings.EqualFold(strings.TrimSpace(payload.Status), "ok") {
		return 0, 0, fmt.Errorf("nobitex returned status %q", payload.Status)
	}

	rate, count, err := vwapFromNobitexTrades(payload.Trades)
	if err != nil {
		return 0, 0, err
	}
	return rate, count, nil
}

func nobitexTradesURL(baseURL, symbol string) (string, error) {
	parsed, err := url.Parse(strings.TrimRight(strings.TrimSpace(baseURL), "/"))
	if err != nil {
		return "", err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("invalid nobitex base url %q", baseURL)
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + "/v2/trades/" + url.PathEscape(strings.ToUpper(strings.TrimSpace(symbol)))
	return parsed.String(), nil
}

type nobitexTradesResponse struct {
	Status string         `json:"status"`
	Trades []nobitexTrade `json:"trades"`
}

type nobitexTrade struct {
	Time   int64  `json:"time"`
	Price  string `json:"price"`
	Volume string `json:"volume"`
	Type   string `json:"type"`
}

func vwapFromNobitexTrades(trades []nobitexTrade) (float64, int, error) {
	if len(trades) == 0 {
		return 0, 0, errors.New("nobitex returned no trades")
	}

	var weightedTotal float64
	var volumeTotal float64
	var simpleTotal float64
	var validTrades int

	for _, trade := range trades {
		price, err := strconv.ParseFloat(strings.TrimSpace(trade.Price), 64)
		if err != nil || price <= 0 {
			continue
		}

		validTrades++
		simpleTotal += price

		volume, err := strconv.ParseFloat(strings.TrimSpace(trade.Volume), 64)
		if err != nil || volume <= 0 {
			continue
		}
		weightedTotal += price * volume
		volumeTotal += volume
	}

	if validTrades == 0 {
		return 0, 0, errors.New("nobitex returned no valid trade prices")
	}
	if volumeTotal > 0 {
		return weightedTotal / volumeTotal, validTrades, nil
	}

	return simpleTotal / float64(validTrades), validTrades, nil
}
