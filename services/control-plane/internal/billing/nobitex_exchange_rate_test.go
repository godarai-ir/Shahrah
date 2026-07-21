package billing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"go.uber.org/zap"
)

func TestLoadNobitexExchangeRateConfigUsesReachableDefaultBaseURL(t *testing.T) {
	t.Setenv("NOBITEX_EXCHANGE_RATE_BASE_URL", "")

	cfg := LoadNobitexExchangeRateConfigFromEnv("nobitex")

	if !cfg.Enabled {
		t.Fatal("expected Nobitex exchange rate updater to be enabled")
	}

	if cfg.BaseURL != "https://apiv2.nobitex.ir" {
		t.Fatalf("unexpected default base URL: %s", cfg.BaseURL)
	}

	endpoint, err := nobitexTradesURL(cfg.BaseURL, cfg.Symbol)
	if err != nil {
		t.Fatalf("nobitexTradesURL returned error: %v", err)
	}

	if !strings.HasPrefix(endpoint, "https://apiv2.nobitex.ir/v2/trades/") {
		t.Fatalf("unexpected endpoint: %s", endpoint)
	}
}

func TestNobitexExchangeRateProcessorFetchesVWAP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/trades/USDTIRT" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"status": "ok",
			"trades": [
				{"time": 1, "price": "1883600", "volume": "4", "type": "buy"},
				{"time": 2, "price": "1883500", "volume": "1", "type": "sell"}
			]
		}`))
	}))
	defer server.Close()

	processor := NewNobitexExchangeRateProcessor(NobitexExchangeRateConfig{
		Enabled:  true,
		BaseURL:  server.URL,
		Symbol:   "USDTIRT",
		Interval: time.Hour,
		Timeout:  time.Second,
	}, &nobitexRateStoreStub{}, zap.NewNop())

	rate, trades, err := processor.fetchUSDtoIRR(context.Background())
	if err != nil {
		t.Fatalf("fetchUSDtoIRR returned error: %v", err)
	}

	if trades != 2 {
		t.Fatalf("expected 2 valid trades, got %d", trades)
	}

	expected := ((1883600.0 * 4.0) + 1883500.0) / 5.0
	if rate != expected {
		t.Fatalf("expected vwap %f, got %f", expected, rate)
	}
}

func TestNobitexExchangeRateProcessorProcessUpsertsRate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"status": "ok",
			"trades": [
				{"time": 1, "price": "1883600", "volume": "2", "type": "buy"},
				{"time": 2, "price": "1883400", "volume": "2", "type": "sell"}
			]
		}`))
	}))
	defer server.Close()

	store := &nobitexRateStoreStub{lockAcquired: true}
	processor := NewNobitexExchangeRateProcessor(NobitexExchangeRateConfig{
		Enabled:   true,
		BaseURL:   server.URL,
		Symbol:    "USDTIRT",
		Source:    "nobitex:USDTIRT",
		Interval:  time.Hour,
		Timeout:   time.Second,
		MinTrades: 2,
	}, store, zap.NewNop())

	processor.process(context.Background())

	if store.calls != 1 {
		t.Fatalf("expected one upsert call, got %d", store.calls)
	}

	if store.rate.Source != "nobitex:USDTIRT" {
		t.Fatalf("unexpected source: %s", store.rate.Source)
	}

	if store.rate.USDtoIRR != 1883500 {
		t.Fatalf("expected rate 1883500, got %f", store.rate.USDtoIRR)
	}
}

func TestNobitexExchangeRateProcessorSkipsWhenLockUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("processor should not call Nobitex when lock is unavailable")
	}))
	defer server.Close()

	store := &nobitexRateStoreStub{lockAcquired: false}
	processor := NewNobitexExchangeRateProcessor(NobitexExchangeRateConfig{
		Enabled: true,
		BaseURL: server.URL,
		Symbol:  "USDTIRT",
		Timeout: time.Second,
	}, store, zap.NewNop())

	processor.process(context.Background())

	if store.lockCalls != 1 {
		t.Fatalf("expected one lock attempt, got %d", store.lockCalls)
	}

	if store.calls != 0 {
		t.Fatalf("expected no upsert calls, got %d", store.calls)
	}
}

func TestVWAPFromNobitexTradesFallsBackToSimpleAverageWithoutVolumes(t *testing.T) {
	rate, trades, err := vwapFromNobitexTrades([]nobitexTrade{
		{Price: "100", Volume: "0"},
		{Price: "200", Volume: ""},
	})
	if err != nil {
		t.Fatalf("vwapFromNobitexTrades returned error: %v", err)
	}

	if trades != 2 {
		t.Fatalf("expected 2 valid trades, got %d", trades)
	}

	if rate != 150 {
		t.Fatalf("expected simple average 150, got %f", rate)
	}
}

type nobitexRateStoreStub struct {
	calls        int
	lockCalls    int
	lockAcquired bool
	rate         platform.ExchangeRate
}

func (s *nobitexRateStoreStub) UpsertExchangeRate(_ context.Context, rate platform.ExchangeRate) (*platform.ExchangeRate, error) {
	s.calls++
	s.rate = rate

	return &rate, nil
}

func (s *nobitexRateStoreStub) TryExchangeRateUpdateLock(context.Context) (bool, func(), error) {
	s.lockCalls++
	return s.lockAcquired, func() {}, nil
}
