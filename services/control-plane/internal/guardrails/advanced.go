package guardrails

import (
	"bytes"
	"context"
	"crypto"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"mime"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode"
)

type regexReplaceGuardrail struct {
	name        string
	enforcement Enforcement
	pattern     *regexp.Regexp
	replacement string
}

func newRegexReplaceGuardrail(name string, enforcement Enforcement, pattern string, replacement string, caseSensitive bool) Guardrail {
	compiled, err := regexpFromConfig(pattern, caseSensitive)
	if err != nil {
		compiled = regexp.MustCompile(`a\Ab`)
	}

	return &regexReplaceGuardrail{name: name, enforcement: enforcement, pattern: compiled, replacement: replacement}
}

func (g *regexReplaceGuardrail) Name() string             { return g.name }
func (g *regexReplaceGuardrail) Operation() Operation     { return OperationMutate }
func (g *regexReplaceGuardrail) Enforcement() Enforcement { return g.enforcement }

func (g *regexReplaceGuardrail) Evaluate(input EvaluationInput) (EvaluationResult, error) {
	out := append([]string(nil), input.Texts...)
	mutated := false

	for i, text := range out {
		next := g.pattern.ReplaceAllString(text, g.replacement)
		if next != text {
			out[i] = next
			mutated = true
		}
	}

	if !mutated {
		return EvaluationResult{Texts: out}, nil
	}

	return EvaluationResult{Texts: out, Mutated: true, Reason: "Regex replacement applied"}, nil
}

type countGuardrail struct {
	name        string
	unit        string
	enforcement Enforcement
	min         int
	max         int
}

func newCountGuardrail(name string, unit string, enforcement Enforcement, min int, max int) Guardrail {
	return &countGuardrail{name: name, unit: unit, enforcement: enforcement, min: min, max: max}
}

func (g *countGuardrail) Name() string             { return g.name }
func (g *countGuardrail) Operation() Operation     { return OperationValidate }
func (g *countGuardrail) Enforcement() Enforcement { return g.enforcement }

func (g *countGuardrail) Evaluate(input EvaluationInput) (EvaluationResult, error) {
	for _, text := range input.Texts {
		var count int

		switch g.unit {
		case "sentence":
			count = sentenceCount(text)
		case "word":
			count = wordCount(text)
		default:
			count = len([]rune(text))
		}

		if (g.min >= 0 && count < g.min) || (g.max >= 0 && count > g.max) {
			return EvaluationResult{Texts: input.Texts, Blocked: true, Reason: fmt.Sprintf("%s count is outside allowed range", g.unit)}, nil
		}
	}

	return EvaluationResult{Texts: input.Texts}, nil
}

func sentenceCount(text string) int {
	count := 0
	seen := false

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			seen = true
		}

		if strings.ContainsRune(".!?؟۔", r) && seen {
			count++
			seen = false
		}
	}

	if seen {
		count++
	}

	return count
}

func wordCount(text string) int {
	return len(strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-'
	}))
}

type jsonSchemaGuardrail struct {
	name        string
	enforcement Enforcement
	schema      map[string]interface{}
}

func newJSONSchemaGuardrail(name string, enforcement Enforcement, schema map[string]interface{}) Guardrail {
	return &jsonSchemaGuardrail{name: name, enforcement: enforcement, schema: schema}
}

func (g *jsonSchemaGuardrail) Name() string             { return g.name }
func (g *jsonSchemaGuardrail) Operation() Operation     { return OperationValidate }
func (g *jsonSchemaGuardrail) Enforcement() Enforcement { return g.enforcement }

func (g *jsonSchemaGuardrail) Evaluate(input EvaluationInput) (EvaluationResult, error) {
	for _, text := range input.Texts {
		var value interface{}

		if !isJSONText(text, &value) {
			return blockedEvaluation(input.Texts, "Response is not valid JSON"), nil
		}

		if !validateMinimalJSONSchema(value, g.schema) {
			return EvaluationResult{Texts: input.Texts, Blocked: true, Reason: "JSON schema validation failed"}, nil
		}
	}

	return EvaluationResult{Texts: input.Texts}, nil
}

func validateMinimalJSONSchema(value interface{}, schema map[string]interface{}) bool {
	if len(schema) == 0 {
		return true
	}

	schemaType, hasSchemaType := schema["type"].(string)
	if hasSchemaType && schemaType != "" && !jsonValueMatchesType(value, schemaType) {
		return false
	}

	var obj map[string]interface{}
	if objectValue, ok := value.(map[string]interface{}); ok {
		obj = objectValue
	}

	for _, key := range stringSliceFromConfig(schema, "required") {
		if obj == nil {
			return false
		}

		if _, ok := obj[key]; !ok {
			return false
		}
	}

	props := mapFromConfig(schema, "properties")
	for key, rawProp := range props {
		prop, propOK := rawProp.(map[string]interface{})
		if !propOK || prop == nil || obj == nil {
			continue
		}

		if child, ok := obj[key]; ok {
			childType, hasChildType := prop["type"].(string)
			if hasChildType && childType != "" && !jsonValueMatchesType(child, childType) {
				return false
			}
		}
	}

	return true
}

func jsonValueMatchesType(value interface{}, schemaType string) bool {
	switch schemaType {
	case "object":
		_, ok := value.(map[string]interface{})
		return ok
	case "array":
		_, ok := value.([]interface{})
		return ok
	case "string":
		_, ok := value.(string)
		return ok
	case "number":
		_, ok := value.(float64)
		return ok
	case "integer":
		n, ok := value.(float64)
		return ok && n == float64(int64(n))
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "null":
		return value == nil
	default:
		return true
	}
}

type jsonKeysGuardrail struct {
	name        string
	enforcement Enforcement
	keys        []string
	mode        string
}

func newJSONKeysGuardrail(name string, enforcement Enforcement, keys []string, mode string) Guardrail {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "" {
		mode = "all"
	}

	return &jsonKeysGuardrail{name: name, enforcement: enforcement, keys: keys, mode: mode}
}

func (g *jsonKeysGuardrail) Name() string             { return g.name }
func (g *jsonKeysGuardrail) Operation() Operation     { return OperationValidate }
func (g *jsonKeysGuardrail) Enforcement() Enforcement { return g.enforcement }

func (g *jsonKeysGuardrail) Evaluate(input EvaluationInput) (EvaluationResult, error) {
	for _, text := range input.Texts {
		var obj map[string]interface{}

		if !isJSONText(text, &obj) {
			return blockedEvaluation(input.Texts, "Response is not valid JSON"), nil
		}

		matches := 0

		for _, key := range g.keys {
			if _, ok := obj[key]; ok {
				matches++
			}
		}

		var pass bool

		switch g.mode {
		case "any":
			pass = matches > 0
		case "none":
			pass = matches == 0
		default:
			pass = matches == len(g.keys)
		}

		if !pass {
			return EvaluationResult{Texts: input.Texts, Blocked: true, Reason: "JSON keys validation failed"}, nil
		}
	}

	return EvaluationResult{Texts: input.Texts}, nil
}

type containsCodeGuardrail struct {
	name        string
	enforcement Enforcement
	languages   []string
	mustContain bool
}

func newContainsCodeGuardrail(name string, enforcement Enforcement, languages []string, mustContain bool) Guardrail {
	return &containsCodeGuardrail{name: name, enforcement: enforcement, languages: languages, mustContain: mustContain}
}

func (g *containsCodeGuardrail) Name() string             { return g.name }
func (g *containsCodeGuardrail) Operation() Operation     { return OperationValidate }
func (g *containsCodeGuardrail) Enforcement() Enforcement { return g.enforcement }

func (g *containsCodeGuardrail) Evaluate(input EvaluationInput) (EvaluationResult, error) {
	for _, text := range input.Texts {
		found := containsCode(text, g.languages)
		if found != g.mustContain {
			return EvaluationResult{Texts: input.Texts, Blocked: true, Reason: "Code content validation failed"}, nil
		}
	}

	return EvaluationResult{Texts: input.Texts}, nil
}

func containsCode(text string, languages []string) bool {
	lower := strings.ToLower(text)
	if strings.Contains(lower, "```") {
		if len(languages) == 0 {
			return true
		}

		for _, lang := range languages {
			if strings.Contains(lower, "```"+strings.ToLower(strings.TrimSpace(lang))) {
				return true
			}
		}
	}

	checks := map[string][]string{
		"sql":        {"select ", " from ", "insert into ", "update ", "delete from "},
		"python":     {"def ", "import ", "print(", "lambda "},
		"typescript": {"const ", "let ", "interface ", "type ", "=>"},
		"javascript": {"function ", "const ", "let ", "=>"},
		"go":         {"func ", "package ", ":= "},
		"shell":      {"#!/bin/", "curl ", "export "},
	}

	if len(languages) == 0 {
		languages = []string{"sql", "python", "typescript", "javascript", "go", "shell"}
	}

	for _, lang := range languages {
		for _, marker := range checks[strings.ToLower(strings.TrimSpace(lang))] {
			if strings.Contains(lower, marker) {
				return true
			}
		}
	}

	return false
}

type notNullGuardrail struct {
	name        string
	enforcement Enforcement
}

func newNotNullGuardrail(name string, enforcement Enforcement) Guardrail {
	return &notNullGuardrail{name: name, enforcement: enforcement}
}

func (g *notNullGuardrail) Name() string             { return g.name }
func (g *notNullGuardrail) Operation() Operation     { return OperationValidate }
func (g *notNullGuardrail) Enforcement() Enforcement { return g.enforcement }

func (g *notNullGuardrail) Evaluate(input EvaluationInput) (EvaluationResult, error) {
	if len(input.Texts) == 0 {
		return EvaluationResult{Texts: input.Texts, Blocked: true, Reason: "Response content is empty"}, nil
	}

	for _, text := range input.Texts {
		trimmed := strings.TrimSpace(text)
		if trimmed == "" || strings.EqualFold(trimmed, "null") || strings.EqualFold(trimmed, "undefined") {
			return EvaluationResult{Texts: input.Texts, Blocked: true, Reason: "Response content is empty"}, nil
		}
	}

	return EvaluationResult{Texts: input.Texts}, nil
}

type jwtValidationGuardrail struct {
	name           string
	enforcement    Enforcement
	tokenPattern   *regexp.Regexp
	issuer         string
	audience       string
	sharedSecret   string
	inlineJWKS     map[string]interface{}
	jwksURI        string
	introspection  string
	introspectAuth string
}

func newJWTValidationGuardrail(name string, enforcement Enforcement, config map[string]interface{}) Guardrail {
	pattern := stringFromConfig(config, "token_regex", `(?i)bearer\s+([A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+)`)

	compiled, err := regexp.Compile(pattern)
	if err != nil {
		compiled = regexp.MustCompile(`(?i)bearer\s+([A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+)`)
	}

	return &jwtValidationGuardrail{
		name:           name,
		enforcement:    enforcement,
		tokenPattern:   compiled,
		issuer:         stringFromConfig(config, "issuer", ""),
		audience:       stringFromConfig(config, "audience", ""),
		sharedSecret:   stringFromConfig(config, "shared_secret", ""),
		inlineJWKS:     mapFromConfig(config, "inline_jwks"),
		jwksURI:        stringFromConfig(config, "jwks_uri", ""),
		introspection:  stringFromConfig(config, "introspection_url", ""),
		introspectAuth: stringFromConfig(config, "introspection_auth_header", ""),
	}
}

func (g *jwtValidationGuardrail) Name() string             { return g.name }
func (g *jwtValidationGuardrail) Operation() Operation     { return OperationValidate }
func (g *jwtValidationGuardrail) Enforcement() Enforcement { return g.enforcement }

func (g *jwtValidationGuardrail) Evaluate(input EvaluationInput) (EvaluationResult, error) {
	for _, text := range input.Texts {
		match := g.tokenPattern.FindStringSubmatch(text)
		if len(match) < 2 {
			return EvaluationResult{Texts: input.Texts, Blocked: true, Reason: "JWT token is missing"}, nil
		}

		if !g.isValidToken(match[1]) {
			return blockedEvaluation(input.Texts, "JWT validation failed"), nil
		}
	}

	return EvaluationResult{Texts: input.Texts}, nil
}

func (g *jwtValidationGuardrail) validateToken(token string) error {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return errors.New("invalid JWT format")
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return err
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return err
	}

	var (
		header map[string]interface{}
		claims map[string]interface{}
	)

	headerErr := json.Unmarshal(headerBytes, &header)
	if headerErr != nil {
		return headerErr
	}

	claimsErr := json.Unmarshal(payloadBytes, &claims)
	if claimsErr != nil {
		return claimsErr
	}

	if !validateJWTClaims(claims, g.issuer, g.audience) {
		return errors.New("invalid JWT claims")
	}

	if g.introspection != "" {
		return introspectToken(g.introspection, g.introspectAuth, token)
	}

	alg := ""
	if headerAlg, ok := header["alg"].(string); ok {
		alg = headerAlg
	}

	kid := ""
	if headerKID, ok := header["kid"].(string); ok {
		kid = headerKID
	}

	signed := []byte(parts[0] + "." + parts[1])

	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return err
	}

	if strings.HasPrefix(alg, "HS") && g.sharedSecret != "" {
		return verifyHMACJWT(alg, []byte(g.sharedSecret), signed, sig)
	}

	if strings.HasPrefix(alg, "RS") {
		jwks := g.inlineJWKS
		if len(jwks) == 0 && g.jwksURI != "" {
			jwks, err = fetchJWKS(g.jwksURI)
			if err != nil {
				return err
			}
		}

		key, err := rsaPublicKeyFromJWKS(jwks, kid)
		if err != nil {
			return err
		}

		hash := crypto.SHA256

		switch alg {
		case "RS384":
			hash = crypto.SHA384
		case "RS512":
			hash = crypto.SHA512
		}

		h := hash.New()
		_, _ = h.Write(signed)

		return rsa.VerifyPKCS1v15(key, hash, h.Sum(nil), sig)
	}

	return errors.New("unsupported JWT algorithm")
}

func validateJWTClaims(claims map[string]interface{}, issuer string, audience string) bool {
	now := float64(time.Now().Unix())
	if exp, ok := claims["exp"].(float64); ok && now >= exp {
		return false
	}

	if nbf, ok := claims["nbf"].(float64); ok && now < nbf {
		return false
	}

	if issuer != "" && claims["iss"] != issuer {
		return false
	}

	if audience == "" {
		return true
	}

	switch aud := claims["aud"].(type) {
	case string:
		return aud == audience
	case []interface{}:
		for _, item := range aud {
			if item == audience {
				return true
			}
		}
	}

	return false
}

func verifyHMACJWT(alg string, secret []byte, signed []byte, sig []byte) error {
	var mac hashWriter

	switch alg {
	case "HS256":
		mac = hmac.New(sha256.New, secret)
	case "HS384":
		mac = hmac.New(sha512.New384, secret)
	case "HS512":
		mac = hmac.New(sha512.New, secret)
	default:
		return errors.New("unsupported HMAC algorithm")
	}

	_, writeErr := mac.Write(signed)
	if writeErr != nil {
		return writeErr
	}

	if !hmac.Equal(sig, mac.Sum(nil)) {
		return errors.New("invalid JWT signature")
	}

	return nil
}

type hashWriter interface {
	Write([]byte) (int, error)
	Sum([]byte) []byte
}

func fetchJWKS(rawURL string) (map[string]interface{}, error) {
	client := http.Client{Timeout: 3 * time.Second}

	resp, err := client.Get(rawURL)
	if err != nil {
		return nil, err
	}
	defer closeReadCloser(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("JWKS fetch failed")
	}

	var jwks map[string]interface{}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 256*1024)).Decode(&jwks); err != nil {
		return nil, err
	}

	return jwks, nil
}

func rsaPublicKeyFromJWKS(jwks map[string]interface{}, kid string) (*rsa.PublicKey, error) {
	keys, keysOK := jwks["keys"].([]interface{})
	if !keysOK {
		return nil, errors.New("JWKS payload missing keys")
	}

	for _, raw := range keys {
		key, keyOK := raw.(map[string]interface{})
		if !keyOK || key == nil {
			continue
		}

		if kid != "" && key["kid"] != kid {
			continue
		}

		if certs, ok := key["x5c"].([]interface{}); ok && len(certs) > 0 {
			if certText, ok := certs[0].(string); ok {
				der, err := base64.StdEncoding.DecodeString(certText)
				if err == nil {
					cert, err := x509.ParseCertificate(der)
					if err == nil {
						if pub, ok := cert.PublicKey.(*rsa.PublicKey); ok {
							return pub, nil
						}
					}
				}
			}
		}

		nText := ""
		if value, ok := key["n"].(string); ok {
			nText = value
		}

		eText := ""
		if value, ok := key["e"].(string); ok {
			eText = value
		}

		nBytes, nErr := base64.RawURLEncoding.DecodeString(nText)

		eBytes, eErr := base64.RawURLEncoding.DecodeString(eText)
		if nErr != nil || eErr != nil {
			continue
		}

		e := 0
		for _, b := range eBytes {
			e = e<<8 + int(b)
		}

		return &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: e}, nil
	}

	return nil, errors.New("matching JWKS key not found")
}

func introspectToken(endpoint string, authHeader string, token string) error {
	client := http.Client{Timeout: 3 * time.Second}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint, strings.NewReader(url.Values{"token": {token}}.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("content-type", "application/x-www-form-urlencoded")

	if strings.TrimSpace(authHeader) != "" {
		req.Header.Set("authorization", authHeader)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer closeReadCloser(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New("introspection request failed")
	}

	var payload struct {
		Active bool `json:"active"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64*1024)).Decode(&payload); err != nil {
		return err
	}

	if !payload.Active {
		return errors.New("token is inactive")
	}

	return nil
}

type inlineImageURLsGuardrail struct {
	name           string
	enforcement    Enforcement
	maxBytes       int64
	allowPrivate   bool
	allowedHosts   []string
	urlPattern     *regexp.Regexp
	requestTimeout time.Duration
}

func newInlineImageURLsGuardrail(name string, enforcement Enforcement, config map[string]interface{}) Guardrail {
	maxBytes := int64(intFromConfig(config, "max_bytes", 1024*1024))
	if maxBytes <= 0 {
		maxBytes = 1024 * 1024
	}

	return &inlineImageURLsGuardrail{
		name:           name,
		enforcement:    enforcement,
		maxBytes:       maxBytes,
		allowPrivate:   boolFromConfig(config, "allow_private_networks", false),
		allowedHosts:   stringSliceFromConfig(config, "allowed_hosts"),
		urlPattern:     regexp.MustCompile(`https?://[^\s"'<>]+`),
		requestTimeout: time.Duration(intFromConfig(config, "timeout_ms", 3000)) * time.Millisecond,
	}
}

func (g *inlineImageURLsGuardrail) Name() string             { return g.name }
func (g *inlineImageURLsGuardrail) Operation() Operation     { return OperationMutate }
func (g *inlineImageURLsGuardrail) Enforcement() Enforcement { return g.enforcement }

func (g *inlineImageURLsGuardrail) Evaluate(input EvaluationInput) (EvaluationResult, error) {
	out := append([]string(nil), input.Texts...)
	mutated := false

	for i, text := range out {
		next, changed, err := g.inlineImages(text)
		if err != nil {
			return EvaluationResult{Texts: out}, err
		}

		if changed {
			out[i] = next
			mutated = true
		}
	}

	if !mutated {
		return EvaluationResult{Texts: out}, nil
	}

	return EvaluationResult{Texts: out, Mutated: true, Reason: "Image URLs converted to inline data"}, nil
}

func (g *inlineImageURLsGuardrail) inlineImages(text string) (string, bool, error) {
	changed := false

	var firstErr error

	out := g.urlPattern.ReplaceAllStringFunc(text, func(raw string) string {
		if firstErr != nil {
			return raw
		}

		dataURL, err := g.fetchImageDataURL(raw)
		if err != nil {
			firstErr = err
			return raw
		}

		changed = true

		return dataURL
	})

	return out, changed, firstErr
}

func (g *inlineImageURLsGuardrail) fetchImageDataURL(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", errors.New("invalid image URL")
	}

	if parsed.User != nil {
		return "", errors.New("image URL credentials are not allowed")
	}

	if !g.hostAllowed(parsed.Hostname()) {
		return "", errors.New("image URL host is not allowed")
	}

	if !g.allowPrivate {
		privateHostErr := rejectPrivateHost(parsed.Hostname())
		if privateHostErr != nil {
			return "", privateHostErr
		}
	}

	client := http.Client{
		Timeout: g.requestTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if req.URL == nil || req.URL.User != nil || (req.URL.Scheme != "http" && req.URL.Scheme != "https") {
				return errors.New("invalid image URL redirect")
			}

			if !g.hostAllowed(req.URL.Hostname()) {
				return errors.New("image URL redirect host is not allowed")
			}

			if !g.allowPrivate {
				return rejectPrivateHost(req.URL.Hostname())
			}

			return nil
		},
	}

	resp, err := client.Get(raw)
	if err != nil {
		return "", err
	}
	defer closeReadCloser(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errors.New("image fetch failed")
	}

	contentType := resp.Header.Get("content-type")

	mediaType, _, parseErr := mime.ParseMediaType(contentType)
	if parseErr != nil {
		return "", errors.New("invalid image content type")
	}

	if !strings.HasPrefix(mediaType, "image/") {
		return "", errors.New("URL did not return an image")
	}

	var buf bytes.Buffer

	_, err = io.Copy(&buf, io.LimitReader(resp.Body, g.maxBytes+1))
	if err != nil {
		return "", err
	}

	if int64(buf.Len()) > g.maxBytes {
		return "", errors.New("image exceeds max size")
	}

	return "data:" + mediaType + ";base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func (g *inlineImageURLsGuardrail) hostAllowed(hostname string) bool {
	if len(g.allowedHosts) == 0 {
		return true
	}

	for _, host := range g.allowedHosts {
		if strings.EqualFold(hostname, strings.TrimSpace(host)) {
			return true
		}
	}

	return false
}

func rejectPrivateHost(host string) error {
	ips, err := net.LookupIP(host)
	if err != nil {
		return err
	}

	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
			return errors.New("private image URL hosts are not allowed")
		}
	}

	return nil
}

func blockedEvaluation(texts []string, reason string) EvaluationResult {
	return EvaluationResult{Texts: texts, Blocked: true, Reason: reason}
}

func isJSONText(text string, target interface{}) bool {
	return json.Unmarshal([]byte(text), target) == nil
}

func (g *jwtValidationGuardrail) isValidToken(token string) bool {
	return g.validateToken(token) == nil
}

func closeReadCloser(closer io.Closer) {
	if closer == nil {
		return
	}

	if err := closer.Close(); err != nil {
		return
	}
}
