package responsecache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	basecache "github.com/Delpak/shahrah/services/control-plane/internal/cache"
)

const schemaVersion = "v2"

var namespacePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$`)

type Config struct {
	Enabled               bool
	KeyPrefix             string
	DefaultTTL            time.Duration
	MinTTL                time.Duration
	MaxTTL                time.Duration
	MaxResponseBytes      int
	MaxSemanticCandidates int
}

type RequestConfig struct {
	Type                string  `json:"type"`
	TTL                 int     `json:"ttl,omitempty"`
	Namespace           string  `json:"namespace,omitempty"`
	SimilarityThreshold float64 `json:"similarity_threshold,omitempty"`
}

type Scope struct {
	TenantID     string
	KeyID        string
	Namespace    string
	Model        string
	ModelVersion string
}

type Entry struct {
	Version       string            `json:"version"`
	Response      json.RawMessage   `json:"response"`
	TraceID       string            `json:"trace_id"`
	CreatedAt     time.Time         `json:"created_at"`
	ExpiresAt     time.Time         `json:"expires_at"`
	Provider      string            `json:"provider,omitempty"`
	ProviderModel string            `json:"provider_model,omitempty"`
	ExactHash     string            `json:"exact_hash"`
	ParameterHash string            `json:"parameter_hash,omitempty"`
	Embedding     []float32         `json:"embedding,omitempty"`
	EmbeddingID   EmbeddingIdentity `json:"embedding_identity,omitempty"`
	CostSnapshot  CacheCostSnapshot `json:"cost_snapshot"`
}

// CacheCostSnapshot is the immutable, auditable counterfactual used to report
// the provider cost avoided by each subsequent cache hit.
type CacheCostSnapshot struct {
	InputTokens    int       `json:"input_tokens"`
	OutputTokens   int       `json:"output_tokens"`
	CostUSD        float64   `json:"cost_usd"`
	CostIRR        float64   `json:"cost_irr"`
	ExchangeRate   float64   `json:"exchange_rate"`
	PricingVersion string    `json:"pricing_version"`
	CalculatedAt   time.Time `json:"calculated_at"`
}

type Result struct {
	Entry              *Entry
	Status             string
	Similarity         float64
	Semantic           bool
	CacheType          string
	EmbeddingLatencyMS int64
}

type Pending struct {
	config        RequestConfig
	exactKey      string
	semanticKey   string
	exactHash     string
	parameterHash string
	embedding     []float32
}

type Service struct {
	store    basecache.Cache
	embedder Embedder
	cfg      Config
	flightMu sync.Mutex
	flights  map[string]chan struct{}
}

type distributedStore interface {
	DistributedGet(ctx context.Context, key string) (string, error)
	DistributedSet(ctx context.Context, key string, value interface{}, ttl time.Duration) error
}

func NewService(store basecache.Cache, embedder Embedder, cfg Config) (*Service, error) {
	if store == nil {
		return nil, errors.New("response cache store is required")
	}

	if cfg.KeyPrefix == "" {
		cfg.KeyPrefix = "shahrah:response-cache"
	}

	if cfg.DefaultTTL <= 0 {
		cfg.DefaultTTL = 10 * time.Minute
	}

	if cfg.MinTTL <= 0 {
		cfg.MinTTL = time.Second
	}

	if cfg.MaxTTL < cfg.MinTTL {
		cfg.MaxTTL = 24 * time.Hour
	}

	if cfg.MaxResponseBytes <= 0 {
		cfg.MaxResponseBytes = 1 << 20
	}

	if cfg.MaxSemanticCandidates <= 0 {
		cfg.MaxSemanticCandidates = 500
	}

	return &Service{store: store, embedder: embedder, cfg: cfg, flights: make(map[string]chan struct{})}, nil
}

func (s *Service) Enabled() bool { return s != nil && s.cfg.Enabled }

func (s *Service) ParseHeader(raw string) (RequestConfig, error) {
	var cfg RequestConfig

	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("invalid x-godarai-cache-config: %w", err)
	}

	var trailing interface{}
	if err := decoder.Decode(&trailing); err != io.EOF {
		return cfg, errors.New("invalid x-godarai-cache-config: multiple JSON values")
	}

	cfg.Type = strings.ToLower(strings.TrimSpace(cfg.Type))
	if cfg.Type != "simple" && cfg.Type != "smart" {
		return cfg, errors.New("cache type must be simple or smart")
	}

	if cfg.TTL == 0 {
		cfg.TTL = int(s.cfg.DefaultTTL.Seconds())
	}

	ttl := time.Duration(cfg.TTL) * time.Second
	if ttl < s.cfg.MinTTL || ttl > s.cfg.MaxTTL {
		return cfg, fmt.Errorf("cache ttl must be between %d and %d seconds", int(s.cfg.MinTTL.Seconds()), int(s.cfg.MaxTTL.Seconds()))
	}

	cfg.Namespace = strings.TrimSpace(cfg.Namespace)
	if cfg.Namespace == "" {
		cfg.Namespace = "default"
	}

	if !namespacePattern.MatchString(cfg.Namespace) {
		return cfg, errors.New("cache namespace must be 1-128 characters using letters, digits, dot, underscore, colon, or hyphen")
	}

	if cfg.Type == "simple" && cfg.SimilarityThreshold != 0 {
		return cfg, errors.New("similarity_threshold is only valid for smart caching")
	}

	if cfg.Type == "smart" {
		if cfg.SimilarityThreshold == 0 {
			cfg.SimilarityThreshold = 0.9
		}

		if cfg.SimilarityThreshold < 0.8 || cfg.SimilarityThreshold > 1 {
			return cfg, errors.New("similarity_threshold must be between 0.8 and 1.0")
		}
	}

	return cfg, nil
}

func (s *Service) Lookup(ctx context.Context, scope Scope, req contractsopenai.ChatCompletionRequest, cfg RequestConfig) (Result, *Pending, error) {
	if cfg.Type == "smart" {
		if req.Stream || len(req.Tools) > 0 || req.ToolChoice != nil || len(req.Messages) == 0 || s.embedder == nil {
			return Result{Status: "bypass", CacheType: cfg.Type}, nil, nil
		}

		last := req.Messages[len(req.Messages)-1]
		if strings.ToLower(strings.TrimSpace(last.Role)) != "user" || contractsopenai.ChatContentHasImage(last.Content) || contractsopenai.ChatContentHasFile(last.Content) || contractsopenai.ChatContentText(last.Content) == "" {
			return Result{Status: "bypass", CacheType: cfg.Type}, nil, nil
		}
	}

	scope.Namespace = cfg.Namespace

	scopeHash, err := hashJSON(scope)
	if err != nil {
		return Result{}, nil, err
	}

	exactHash, err := hashJSON(struct {
		Version string
		Request contractsopenai.ChatCompletionRequest
	}{schemaVersion, req})
	if err != nil {
		return Result{}, nil, err
	}

	exactKey := fmt.Sprintf("%s:%s:exact:%s:%s", s.cfg.KeyPrefix, schemaVersion, scopeHash, exactHash)

	pending := &Pending{config: cfg, exactKey: exactKey, exactHash: exactHash}
	if entry, ok := s.getEntry(ctx, exactKey); ok {
		return Result{Entry: entry, Status: "hit", Similarity: 1, CacheType: cfg.Type}, pending, nil
	}

	if cfg.Type != "smart" {
		return Result{Status: "miss", CacheType: cfg.Type}, pending, nil
	}

	last := req.Messages[len(req.Messages)-1]
	text := contractsopenai.ChatContentText(last.Content)
	parameterReq := req
	parameterReq.Messages = append([]contractsopenai.ChatMessage(nil), req.Messages...)
	parameterReq.Messages[len(parameterReq.Messages)-1].Content = "__semantic_final_user_text__"

	parameterHash, err := hashJSON(struct {
		Version string
		Request contractsopenai.ChatCompletionRequest
	}{schemaVersion, parameterReq})
	if err != nil {
		return Result{}, pending, err
	}

	embeddingIDHash, err := hashJSON(s.embedder.Identity())
	if err != nil {
		return Result{}, pending, err
	}

	pending.parameterHash = parameterHash
	pending.semanticKey = fmt.Sprintf("%s:%s:semantic:%s:%s:%s", s.cfg.KeyPrefix, schemaVersion, scopeHash, parameterHash, embeddingIDHash)
	embeddingStart := time.Now()
	vectors, err := s.embedder.Embed(ctx, []string{text})
	embeddingLatencyMS := time.Since(embeddingStart).Milliseconds()

	if err != nil || len(vectors) != 1 {
		if err == nil {
			err = errors.New("embedder returned an unexpected vector count")
		}

		return Result{Status: "error", CacheType: cfg.Type, EmbeddingLatencyMS: embeddingLatencyMS}, pending, err
	}

	pending.embedding = vectors[0]
	keys := s.getIndex(ctx, pending.semanticKey)
	bestScore := -2.0

	var best *Entry

	for _, key := range keys {
		entry, ok := s.getEntry(ctx, key)
		if !ok || entry.ParameterHash != parameterHash || entry.EmbeddingID != s.embedder.Identity() {
			continue
		}

		score, scoreErr := CosineSimilarity(pending.embedding, entry.Embedding)
		if scoreErr == nil && score > bestScore {
			bestScore, best = score, entry
		}
	}

	if best != nil && bestScore >= cfg.SimilarityThreshold {
		return Result{Entry: best, Status: "hit", Similarity: bestScore, Semantic: true, CacheType: cfg.Type, EmbeddingLatencyMS: embeddingLatencyMS}, pending, nil
	}

	return Result{Status: "miss", CacheType: cfg.Type, EmbeddingLatencyMS: embeddingLatencyMS}, pending, nil
}

// AcquireMiss serializes provider work for one exact cache key. The returned
// release function must be called. Waiting respects context cancellation.
func (s *Service) AcquireMiss(ctx context.Context, pending *Pending) (func(), bool, error) {
	if pending == nil || pending.exactKey == "" {
		return func() {}, false, nil
	}

	waited := false

	for {
		s.flightMu.Lock()

		wait, exists := s.flights[pending.exactKey]
		if !exists {
			wait = make(chan struct{})
			s.flights[pending.exactKey] = wait
			s.flightMu.Unlock()

			var once sync.Once

			return func() {
				once.Do(func() {
					s.flightMu.Lock()
					delete(s.flights, pending.exactKey)
					close(wait)
					s.flightMu.Unlock()
				})
			}, waited, nil
		}
		s.flightMu.Unlock()

		select {
		case <-ctx.Done():
			return nil, waited, ctx.Err()
		case <-wait:
			waited = true
		}
	}
}

func (s *Service) Store(ctx context.Context, pending *Pending, response []byte, traceID, provider, providerModel string, cost platform.UsageCostBreakdown, pricingVersion string) error {
	if pending == nil || len(response) == 0 || len(response) > s.cfg.MaxResponseBytes {
		return errors.New("response is empty, oversized, or cache context is missing")
	}

	ttl := time.Duration(pending.config.TTL) * time.Second
	now := time.Now().UTC()

	entry := Entry{
		Version: schemaVersion, Response: append(json.RawMessage(nil), response...), TraceID: traceID,
		CreatedAt: now, ExpiresAt: now.Add(ttl), Provider: provider, ProviderModel: providerModel,
		ExactHash: pending.exactHash, ParameterHash: pending.parameterHash, Embedding: pending.embedding,
		CostSnapshot: CacheCostSnapshot{
			InputTokens: cost.InputTokens, OutputTokens: cost.OutputTokens, CostUSD: cost.TotalCostUSD,
			CostIRR: cost.TotalCostIRR, ExchangeRate: cost.ExchangeRate,
			PricingVersion: pricingVersion, CalculatedAt: now,
		},
	}
	if len(pending.embedding) > 0 && s.embedder != nil {
		entry.EmbeddingID = s.embedder.Identity()
	}

	raw, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	s.store.Set(ctx, pending.exactKey, string(raw), ttl)

	if pending.semanticKey != "" && len(pending.embedding) > 0 {
		keys := s.getIndex(ctx, pending.semanticKey)
		filtered := make([]string, 0, len(keys)+1)

		filtered = append(filtered, pending.exactKey)
		for _, key := range keys {
			if key != pending.exactKey && len(filtered) < s.cfg.MaxSemanticCandidates {
				filtered = append(filtered, key)
			}
		}

		indexRaw, marshalErr := json.Marshal(filtered)
		if marshalErr != nil {
			return marshalErr
		}

		if shared, ok := s.store.(distributedStore); ok {
			if err := shared.DistributedSet(ctx, pending.semanticKey, string(indexRaw), s.cfg.MaxTTL); err != nil {
				return err
			}
		} else {
			s.store.Set(ctx, pending.semanticKey, string(indexRaw), s.cfg.MaxTTL)
		}
	}

	return nil
}

func (s *Service) getEntry(ctx context.Context, key string) (*Entry, bool) {
	value, ok := s.store.Get(ctx, key)
	if !ok {
		return nil, false
	}

	raw, ok := cacheString(value)
	if !ok {
		return nil, false
	}

	var entry Entry
	if json.Unmarshal([]byte(raw), &entry) != nil || entry.Version != schemaVersion || time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return &entry, true
}

func (s *Service) getIndex(ctx context.Context, key string) []string {
	var raw string

	if shared, ok := s.store.(distributedStore); ok {
		value, err := shared.DistributedGet(ctx, key)
		if err != nil {
			return nil
		}

		raw = value
	} else {
		value, found := s.store.Get(ctx, key)
		if !found {
			return nil
		}

		var ok bool

		raw, ok = cacheString(value)
		if !ok {
			return nil
		}
	}

	var keys []string
	if json.Unmarshal([]byte(raw), &keys) != nil {
		return nil
	}

	if len(keys) > s.cfg.MaxSemanticCandidates {
		keys = keys[:s.cfg.MaxSemanticCandidates]
	}

	return keys
}

func cacheString(value interface{}) (string, bool) {
	switch v := value.(type) {
	case string:
		return v, true
	case []byte:
		return string(v), true
	default:
		return "", false
	}
}

func hashJSON(value interface{}) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(raw)

	return hex.EncodeToString(hash[:]), nil
}
