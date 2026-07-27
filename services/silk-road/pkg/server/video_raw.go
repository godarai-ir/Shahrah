package server

import (
	"bytes"
	"context"
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

type videoCapableAdapter struct {
	base  providerAdapter
	video *rawVideoAdapter
}

type rawVideoAdapter struct {
	defaultBaseURL string
	authMode       string
	tokenMu        sync.Mutex
	googleToken    string
	googleExpiry   time.Time
}

func newVideoCapableAdapter(base providerAdapter, defaultBaseURL, authMode string) providerAdapter {
	return &videoCapableAdapter{
		base:  base,
		video: &rawVideoAdapter{defaultBaseURL: defaultBaseURL, authMode: authMode},
	}
}

func newVideoOnlyAdapter(defaultBaseURL, authMode string) providerAdapter {
	return newVideoCapableAdapter(nil, defaultBaseURL, authMode)
}

func (a *videoCapableAdapter) Dispatch(ctx context.Context, client *http.Client, payload platform.TunnelRequest) (*adapterResponse, error) {
	if isVideoOperation(payload) {
		return a.video.Dispatch(ctx, client, payload)
	}
	if a.base == nil {
		return nil, fmt.Errorf("provider only supports video operations through this adapter")
	}
	return a.base.Dispatch(ctx, client, payload)
}

func (a *rawVideoAdapter) Dispatch(ctx context.Context, client *http.Client, payload platform.TunnelRequest) (*adapterResponse, error) {
	defaultBaseURL := a.defaultBaseURL
	switch normalizeProvider(payload.Deployment.Provider) {
	case "aws-bedrock", "bedrock":
		if defaultBaseURL == "" {
			region := strings.TrimSpace(payload.Deployment.RegionName)
			if region == "" {
				return nil, fmt.Errorf("aws bedrock video requires a region")
			}
			defaultBaseURL = "https://bedrock-runtime." + region + ".amazonaws.com"
		}
	case "google-vertex", "vertex", "vertex-ai", "vertexai":
		if defaultBaseURL == "" {
			region := strings.TrimSpace(payload.Deployment.RegionName)
			if region == "" {
				return nil, fmt.Errorf("vertex video requires a region")
			}
			defaultBaseURL = "https://" + region + "-aiplatform.googleapis.com"
		}
	}
	target, err := resolveBaseURL(payload.Deployment.APIBase, defaultBaseURL)
	if err != nil {
		return nil, err
	}
	endpointPath := openAICompatiblePath(payload)
	endpoint := endpointPath
	if parsed, parseErr := url.Parse(endpointPath); parseErr != nil || !parsed.IsAbs() {
		endpoint = joinURL(target, endpointPath)
	}
	method := http.MethodPost
	if raw, ok := payload.RequestMetadata["provider_http_method"].(string); ok {
		switch strings.ToUpper(strings.TrimSpace(raw)) {
		case http.MethodGet, http.MethodPost, http.MethodDelete:
			method = strings.ToUpper(strings.TrimSpace(raw))
		default:
			return nil, fmt.Errorf("unsupported video provider method %q", raw)
		}
	}

	var body *bytes.Reader
	if method == http.MethodGet || method == http.MethodDelete {
		body = bytes.NewReader(nil)
	} else {
		body = bytes.NewReader(payload.Body)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}
	if provider := normalizeProvider(payload.Deployment.Provider); provider == "azure-open-ai" || provider == "azure" {
		if version := strings.TrimSpace(payload.Deployment.APIVersion); version != "" {
			query := req.URL.Query()
			query.Set("api-version", version)
			req.URL.RawQuery = query.Encode()
		}
	}
	req.Header.Set("Content-Type", "application/json")
	if requestID := strings.TrimSpace(payload.RequestID); requestID != "" {
		req.Header.Set("Idempotency-Key", requestID)
	}
	apiKey := deploymentAPIKey(payload.Deployment)
	switch {
	case payload.Deployment.ProviderAuth != nil && payload.Deployment.ProviderAuth.Mode == string(platform.ProviderAuthModeAWSSigV4):
		if err := signAWSRequest(req, payload.Body, payload.Deployment); err != nil {
			return nil, err
		}
	case payload.Deployment.ProviderAuth != nil && payload.Deployment.ProviderAuth.Mode == string(platform.ProviderAuthModeGoogleServiceAccount):
		token, tokenErr := a.googleAccessToken(ctx, client, payload.Deployment.ProviderAuth)
		if tokenErr != nil {
			return nil, tokenErr
		}
		req.Header.Set("Authorization", "Bearer "+token)
	case a.authMode == "api-key":
		if apiKey != "" {
			req.Header.Set("api-key", apiKey)
		}
	case a.authMode == "google-api-key":
		if apiKey != "" {
			req.Header.Set("x-goog-api-key", apiKey)
		}
	default:
		if apiKey != "" {
			req.Header.Set("Authorization", "Bearer "+apiKey)
		}
	}
	applyDocumentedProviderHeaders(req, payload.Deployment)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return &adapterResponse{StatusCode: resp.StatusCode, Headers: resp.Header.Clone(), Body: resp.Body}, nil
}

func signAWSRequest(req *http.Request, body []byte, deployment platform.Deployment) error {
	auth := deployment.ProviderAuth
	if auth == nil || strings.TrimSpace(auth.AWSAccessKeyID) == "" || strings.TrimSpace(auth.AWSSecretAccessKey) == "" {
		return fmt.Errorf("aws sigv4 credentials are incomplete")
	}
	region := strings.TrimSpace(deployment.RegionName)
	if region == "" {
		return fmt.Errorf("aws sigv4 region is required")
	}
	now := time.Now().UTC()
	amzDate := now.Format("20060102T150405Z")
	date := now.Format("20060102")
	payloadHashBytes := sha256.Sum256(body)
	payloadHash := hex.EncodeToString(payloadHashBytes[:])
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)
	if token := strings.TrimSpace(auth.AWSSessionToken); token != "" {
		req.Header.Set("X-Amz-Security-Token", token)
	}

	signedHeaders := []string{"content-type", "host", "x-amz-content-sha256", "x-amz-date"}
	if req.Header.Get("X-Amz-Security-Token") != "" {
		signedHeaders = append(signedHeaders, "x-amz-security-token")
	}
	sort.Strings(signedHeaders)
	canonicalHeaders := strings.Builder{}
	for _, name := range signedHeaders {
		value := req.URL.Host
		if req.Host != "" {
			value = req.Host
		}
		if name != "host" {
			value = strings.Join(strings.Fields(req.Header.Get(name)), " ")
		}
		canonicalHeaders.WriteString(name)
		canonicalHeaders.WriteByte(':')
		canonicalHeaders.WriteString(value)
		canonicalHeaders.WriteByte('\n')
	}
	canonicalURI := req.URL.EscapedPath()
	if canonicalURI == "" {
		canonicalURI = "/"
	}
	canonicalRequest := strings.Join([]string{
		req.Method,
		canonicalURI,
		req.URL.Query().Encode(),
		canonicalHeaders.String(),
		strings.Join(signedHeaders, ";"),
		payloadHash,
	}, "\n")
	canonicalHash := sha256.Sum256([]byte(canonicalRequest))
	scope := date + "/" + region + "/bedrock/aws4_request"
	stringToSign := "AWS4-HMAC-SHA256\n" + amzDate + "\n" + scope + "\n" + hex.EncodeToString(canonicalHash[:])
	dateKey := hmacSHA256([]byte("AWS4"+auth.AWSSecretAccessKey), date)
	regionKey := hmacSHA256(dateKey, region)
	serviceKey := hmacSHA256(regionKey, "bedrock")
	signingKey := hmacSHA256(serviceKey, "aws4_request")
	signature := hex.EncodeToString(hmacSHA256(signingKey, stringToSign))
	req.Header.Set("Authorization", "AWS4-HMAC-SHA256 Credential="+auth.AWSAccessKeyID+"/"+scope+
		", SignedHeaders="+strings.Join(signedHeaders, ";")+", Signature="+signature)
	return nil
}

func hmacSHA256(key []byte, value string) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(value))
	return mac.Sum(nil)
}

func (a *rawVideoAdapter) googleAccessToken(ctx context.Context, client *http.Client, auth *platform.ResolvedProviderAuth) (string, error) {
	a.tokenMu.Lock()
	defer a.tokenMu.Unlock()
	if a.googleToken != "" && time.Until(a.googleExpiry) > 2*time.Minute {
		return a.googleToken, nil
	}
	if auth == nil || strings.TrimSpace(auth.GoogleClientEmail) == "" || strings.TrimSpace(auth.GooglePrivateKey) == "" {
		return "", fmt.Errorf("google service-account credentials are incomplete")
	}
	tokenURI := strings.TrimSpace(auth.GoogleTokenURI)
	if tokenURI == "" {
		tokenURI = "https://oauth2.googleapis.com/token"
	}
	block, _ := pem.Decode([]byte(strings.ReplaceAll(auth.GooglePrivateKey, `\n`, "\n")))
	if block == nil {
		return "", fmt.Errorf("google service-account private key is invalid")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		if pkcs1, pkcs1Err := x509.ParsePKCS1PrivateKey(block.Bytes); pkcs1Err == nil {
			key = pkcs1
		} else {
			return "", fmt.Errorf("google service-account private key is invalid")
		}
	}
	privateKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("google service-account private key must be RSA")
	}
	now := time.Now().Unix()
	header, _ := json.Marshal(map[string]interface{}{"alg": "RS256", "typ": "JWT"})
	claims, _ := json.Marshal(map[string]interface{}{
		"iss": auth.GoogleClientEmail, "scope": "https://www.googleapis.com/auth/cloud-platform",
		"aud": tokenURI, "iat": now, "exp": now + 3600,
	})
	unsigned := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(claims)
	digest := sha256.Sum256([]byte(unsigned))
	signed, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
	if err != nil {
		return "", err
	}
	assertion := unsigned + "." + base64.RawURLEncoding.EncodeToString(signed)
	form := url.Values{
		"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
		"assertion":  {assertion},
	}
	tokenReq, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURI, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := client.Do(tokenReq)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", fmt.Errorf("google oauth token request returned status %d", response.StatusCode)
	}
	var token struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err := json.Unmarshal(responseBody, &token); err != nil || strings.TrimSpace(token.AccessToken) == "" {
		return "", fmt.Errorf("google oauth token response is invalid")
	}
	if token.ExpiresIn <= 0 {
		token.ExpiresIn = 3600
	}
	a.googleToken = token.AccessToken
	a.googleExpiry = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	return a.googleToken, nil
}

func isVideoOperation(payload platform.TunnelRequest) bool {
	if payload.RequestMetadata == nil {
		return false
	}
	value, _ := payload.RequestMetadata["provider_operation"].(string)
	return strings.EqualFold(strings.TrimSpace(value), "video")
}
