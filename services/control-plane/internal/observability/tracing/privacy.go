package tracing

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

type CapturePolicy string

const (
	CapturePolicyNone            CapturePolicy = "none"
	CapturePolicyMetadataOnly    CapturePolicy = "metadata-only"
	CapturePolicySampledRedacted CapturePolicy = "sampled-redacted"
	CapturePolicyFullSecure      CapturePolicy = "full-secure"
)

const (
	payloadRequestKey  = "request_payload"
	payloadResponseKey = "response_payload"
)

func ParseCapturePolicy(raw string) CapturePolicy {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(CapturePolicyNone):
		return CapturePolicyNone
	case string(CapturePolicySampledRedacted):
		return CapturePolicySampledRedacted
	case string(CapturePolicyFullSecure):
		return CapturePolicyFullSecure
	default:
		return CapturePolicyMetadataOnly
	}
}

type PrivacyManager struct {
	policy      CapturePolicy
	maxPayload  int
	piiHashSalt string
	blobStore   *encryptedBlobStore
	secureReady bool
	resolver    TenantPolicyResolver
	cacheTTL    time.Duration
	cacheMu     sync.RWMutex
	cache       map[string]tenantPolicyCacheItem
}

type tenantPolicyCacheItem struct {
	policy    CapturePolicy
	expiresAt time.Time
}

type TenantPolicyResolver interface {
	ResolveTenantPolicy(tenantID string) (CapturePolicy, bool)
}

func NewPrivacyManager(cfg Config) *PrivacyManager {
	m := &PrivacyManager{
		policy:      cfg.CapturePolicy,
		maxPayload:  cfg.PayloadMaxBytes,
		piiHashSalt: cfg.PIIHashSalt,
		cacheTTL:    90 * time.Second,
		cache:       map[string]tenantPolicyCacheItem{},
	}
	if m.maxPayload <= 0 {
		m.maxPayload = 16 * 1024
	}

	if cfg.CapturePolicy != CapturePolicyFullSecure {
		return m
	}

	store, err := newEncryptedBlobStore(cfg.PayloadBlobDir, cfg.PayloadEncryptKey)
	if err != nil {
		// Fail open: fallback to redacted capture if secure mode can't initialize.
		m.policy = CapturePolicySampledRedacted
		return m
	}

	m.blobStore = store
	m.secureReady = true

	return m
}

func (m *PrivacyManager) SetTenantPolicyResolver(resolver TenantPolicyResolver) {
	m.resolver = resolver
}

func (m *PrivacyManager) ResolvePolicyForTenant(tenantID string) CapturePolicy {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" || m.resolver == nil {
		return m.policy
	}

	now := time.Now()

	m.cacheMu.RLock()
	cached, ok := m.cache[tenantID]
	m.cacheMu.RUnlock()

	if ok && now.Before(cached.expiresAt) {
		return cached.policy
	}

	resolved, found := m.resolver.ResolveTenantPolicy(tenantID)
	if !found {
		resolved = m.policy
	}

	m.cacheMu.Lock()
	m.cache[tenantID] = tenantPolicyCacheItem{
		policy:    resolved,
		expiresAt: now.Add(m.cacheTTL),
	}
	m.cacheMu.Unlock()

	return resolved
}

func (m *PrivacyManager) Apply(traceID, spanID, spanName string, attrs map[string]interface{}, effectivePolicy ...CapturePolicy) map[string]interface{} {
	if len(attrs) == 0 {
		return attrs
	}

	out := cloneMap(attrs)
	requestPayload, hasReq := out[payloadRequestKey]
	responsePayload, hasResp := out[payloadResponseKey]
	delete(out, payloadRequestKey)
	delete(out, payloadResponseKey)

	policy := m.policy
	if len(effectivePolicy) > 0 {
		policy = effectivePolicy[0]
	}

	switch policy {
	case CapturePolicyNone:
		// Drop request/response payloads entirely.
	case CapturePolicyMetadataOnly:
		if hasReq {
			out["request_payload_meta"] = payloadMeta(requestPayload, m.maxPayload)
		}

		if hasResp {
			out["response_payload_meta"] = payloadMeta(responsePayload, m.maxPayload)
		}
	case CapturePolicySampledRedacted:
		if hasReq {
			out["request_payload_meta"] = payloadMeta(requestPayload, m.maxPayload)
			out["request_payload_redacted"] = m.redactValue("request_payload", requestPayload)
		}

		if hasResp {
			out["response_payload_meta"] = payloadMeta(responsePayload, m.maxPayload)
			out["response_payload_redacted"] = m.redactValue("response_payload", responsePayload)
		}
	case CapturePolicyFullSecure:
		if hasReq {
			out["request_payload_meta"] = payloadMeta(requestPayload, m.maxPayload)

			ref, err := m.storePayload(traceID, spanID, spanName, "request", requestPayload)
			if err == nil {
				out["request_payload_blob_ref"] = ref.BlobRef
				out["request_payload_sha256"] = ref.SHA256
				out["request_payload_bytes"] = ref.Bytes
				out["request_payload_encrypted"] = true
			} else {
				out["request_payload_redacted"] = m.redactValue("request_payload", requestPayload)
				out["privacy_warning"] = "request payload secure storage unavailable"
			}
		}

		if hasResp {
			out["response_payload_meta"] = payloadMeta(responsePayload, m.maxPayload)

			ref, err := m.storePayload(traceID, spanID, spanName, "response", responsePayload)
			if err == nil {
				out["response_payload_blob_ref"] = ref.BlobRef
				out["response_payload_sha256"] = ref.SHA256
				out["response_payload_bytes"] = ref.Bytes
				out["response_payload_encrypted"] = true
			} else {
				out["response_payload_redacted"] = m.redactValue("response_payload", responsePayload)
				out["privacy_warning"] = "response payload secure storage unavailable"
			}
		}
	}

	return m.sanitizeMap(out)
}

func (m *PrivacyManager) storePayload(traceID, spanID, spanName, kind string, payload interface{}) (*blobRef, error) {
	if m.blobStore == nil {
		return nil, fmt.Errorf("blob store unavailable")
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	if len(raw) > m.maxPayload {
		raw = raw[:m.maxPayload]
	}

	return m.blobStore.store(traceID, spanID, spanName, kind, raw)
}

func (m *PrivacyManager) sanitizeMap(in map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = m.sanitizeValue(k, v)
	}

	return out
}

func (m *PrivacyManager) sanitizeValue(key string, value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		return m.sanitizeMap(v)
	case []interface{}:
		out := make([]interface{}, 0, len(v))
		for _, item := range v {
			out = append(out, m.sanitizeValue(key, item))
		}

		return out
	case string:
		if looksSensitiveKey(key) || looksSensitiveValue(v) {
			return m.maskedString(v)
		}

		return truncateString(v, 1024)
	default:
		return value
	}
}

func (m *PrivacyManager) redactValue(key string, value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(v))
		for k, item := range v {
			out[k] = m.redactValue(k, item)
		}

		return out
	case []interface{}:
		out := make([]interface{}, 0, len(v))
		for _, item := range v {
			out = append(out, m.redactValue(key, item))
		}

		return out
	case string:
		return m.maskedString(v)
	default:
		if looksSensitiveKey(key) {
			return m.maskedString(fmt.Sprintf("%v", value))
		}

		return value
	}
}

func (m *PrivacyManager) maskedString(raw string) string {
	trimmed := truncateString(strings.TrimSpace(raw), 256)
	sum := sha256.Sum256([]byte(m.piiHashSalt + "|" + raw))

	return fmt.Sprintf("[redacted sha256=%s len=%d preview=%q]", hex.EncodeToString(sum[:])[:12], len(raw), trimmed)
}

func payloadMeta(payload interface{}, max int) map[string]interface{} {
	raw, err := json.Marshal(payload)
	if err != nil {
		return map[string]interface{}{"present": true, "marshal_error": true}
	}

	sha := sha256.Sum256(raw)

	return map[string]interface{}{
		"present":       true,
		"bytes":         len(raw),
		"truncated":     len(raw) > max,
		"sha256_prefix": hex.EncodeToString(sha[:])[:16],
	}
}

func cloneMap(in map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = v
	}

	return out
}

func truncateString(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}

	return s[:max]
}

var sensitiveValueRegex = regexp.MustCompile(`(?i)(bearer\s+[a-z0-9\-\._]+|sk-[a-z0-9\-_]+|[A-Z0-9._%+\-]+@[A-Z0-9.\-]+\.[A-Z]{2,})`)

func looksSensitiveValue(v string) bool {
	return sensitiveValueRegex.MatchString(v)
}

func looksSensitiveKey(key string) bool {
	k := strings.ToLower(strings.TrimSpace(key))
	if k == "" {
		return false
	}

	sensitiveParts := []string{
		"password", "secret", "token", "api_key", "authorization", "auth",
		"email", "mobile", "phone", "national", "ssn", "prompt", "content",
		"message", "request_body", "response_body",
	}
	for _, part := range sensitiveParts {
		if strings.Contains(k, part) {
			return true
		}
	}

	return false
}

type encryptedBlobStore struct {
	baseDir string
	gcm     cipher.AEAD
}

type blobRef struct {
	BlobRef string
	SHA256  string
	Bytes   int
}

func newEncryptedBlobStore(baseDir, key string) (*encryptedBlobStore, error) {
	decodedKey, err := decodeKeyMaterial(key)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(decodedKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &encryptedBlobStore{
		baseDir: strings.TrimSpace(baseDir),
		gcm:     gcm,
	}, nil
}

func (s *encryptedBlobStore) store(traceID, spanID, spanName, kind string, plain []byte) (*blobRef, error) {
	if strings.TrimSpace(s.baseDir) == "" {
		return nil, fmt.Errorf("empty blob dir")
	}

	if err := os.MkdirAll(s.baseDir, 0o700); err != nil {
		return nil, err
	}

	nonce := make([]byte, s.gcm.NonceSize(), s.gcm.NonceSize()+len(plain)+s.gcm.Overhead())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	aad := []byte(traceID + "|" + spanID + "|" + spanName + "|" + kind)
	ciphertext := s.gcm.Seal(nil, nonce, plain, aad)
	sum := sha256.Sum256(plain)
	blobID := fmt.Sprintf("%s_%s_%s_%d", sanitizeFilePart(traceID), sanitizeFilePart(spanID), kind, time.Now().UnixNano())
	path := filepath.Join(s.baseDir, blobID+".bin")

	content := append(nonce, ciphertext...)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		return nil, err
	}

	return &blobRef{
		BlobRef: path,
		SHA256:  hex.EncodeToString(sum[:]),
		Bytes:   len(plain),
	}, nil
}

func sanitizeFilePart(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, string(filepath.Separator), "_")

	s = strings.ReplaceAll(s, ":", "_")
	if s == "" {
		return "empty"
	}

	return s
}

func decodeKeyMaterial(raw string) ([]byte, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("missing TRACE_PAYLOAD_ENCRYPTION_KEY")
	}

	if decoded, err := base64.StdEncoding.DecodeString(raw); err == nil {
		if len(decoded) == 32 {
			return decoded, nil
		}
	}

	if len(raw) == 32 {
		return []byte(raw), nil
	}

	return nil, fmt.Errorf("TRACE_PAYLOAD_ENCRYPTION_KEY must be 32-byte raw or base64")
}
