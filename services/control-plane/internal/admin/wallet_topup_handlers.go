package admin

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
)

type createTopUpIntentRequest struct {
	AmountUSD float64                `json:"amount_usd"`
	AmountIRR int64                  `json:"amount_irr"`
	Provider  string                 `json:"provider"`
	ReturnURL string                 `json:"return_url"`
	Metadata  map[string]interface{} `json:"metadata"`
}

type updateTopUpSettingsRequest struct {
	MinAmountUSD float64 `json:"min_amount_usd"`
	MaxAmountUSD float64 `json:"max_amount_usd"`
	TaxPercent   float64 `json:"tax_percent"`
	FeePercent   float64 `json:"fee_percent"`
}

type paymentCreateResult struct {
	Authority  string
	PaymentURL string
	Raw        map[string]interface{}
}

type paymentVerifyResult struct {
	Success      bool
	RefID        string
	GatewayTxnID string
	CardMask     string
	PayerPhone   string
	Raw          map[string]interface{}
	Failure      string
}

type paymentProvider interface {
	CreatePayment(c *fiber.Ctx, intent *platform.WalletTopUpIntent) (*paymentCreateResult, error)
	VerifyPayment(c *fiber.Ctx, authority string) (*paymentVerifyResult, error)
}

type mockProvider struct{}
type zarinpalProvider struct{}
type idpayProvider struct{}
type unsupportedProvider struct{ name string }

func (m *mockProvider) CreatePayment(c *fiber.Ctx, intent *platform.WalletTopUpIntent) (*paymentCreateResult, error) {
	auth := "mock_" + randomToken(12)

	callbackURL := strings.TrimSpace(intent.CallbackURL)
	if callbackURL == "" {
		callbackURL = fmt.Sprintf("%s/api/v1/wallet/topup/callback/mock", resolveExternalBaseURL(c))
	}

	paymentURL := fmt.Sprintf("%s?authority=%s&status=ok", callbackURL, auth)

	return &paymentCreateResult{
		Authority:  auth,
		PaymentURL: paymentURL,
		Raw: map[string]interface{}{
			"provider": "mock",
			"intent":   intent.PublicID,
		},
	}, nil
}

func (m *mockProvider) VerifyPayment(c *fiber.Ctx, authority string) (*paymentVerifyResult, error) {
	status := strings.ToLower(strings.TrimSpace(c.Query("status")))
	if status == "failed" || status == "cancel" || status == "canceled" {
		return &paymentVerifyResult{Success: false, Failure: "payment canceled by user", Raw: map[string]interface{}{"status": status}}, nil
	}

	return &paymentVerifyResult{
		Success:      true,
		RefID:        "mock_ref_" + randomToken(8),
		GatewayTxnID: "mock_txn_" + randomToken(8),
		CardMask:     "6037******1234",
		PayerPhone:   "09*********",
		Raw: map[string]interface{}{
			"authority": authority,
			"status":    "ok",
		},
	}, nil
}

func (u *unsupportedProvider) CreatePayment(c *fiber.Ctx, intent *platform.WalletTopUpIntent) (*paymentCreateResult, error) {
	return nil, fmt.Errorf("provider %s not configured yet", u.name)
}

func (u *unsupportedProvider) VerifyPayment(c *fiber.Ctx, authority string) (*paymentVerifyResult, error) {
	return nil, fmt.Errorf("provider %s not configured yet", u.name)
}

func selectProvider(name string) paymentProvider {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "mock":
		return &mockProvider{}
	case "zarinpal":
		return &zarinpalProvider{}
	case "idpay":
		return &idpayProvider{}
	default:
		return &unsupportedProvider{name: name}
	}
}

func (p *zarinpalProvider) CreatePayment(c *fiber.Ctx, intent *platform.WalletTopUpIntent) (*paymentCreateResult, error) {
	merchantID := strings.TrimSpace(os.Getenv("ZARINPAL_MERCHANT_ID"))
	if merchantID == "" {
		return nil, errors.New("ZARINPAL_MERCHANT_ID is not set")
	}

	baseURL := strings.TrimSpace(os.Getenv("ZARINPAL_BASE_URL"))
	if baseURL == "" {
		baseURL = "https://api.zarinpal.com"
	}

	reqPayload := map[string]interface{}{
		"merchant_id":  merchantID,
		"amount":       int64(intent.PayableIRR),
		"callback_url": intent.CallbackURL,
		"description":  fmt.Sprintf("GodarAI wallet topup %s", intent.PublicID),
		"metadata": map[string]interface{}{
			"mobile": "",
			"email":  "",
		},
	}

	respBody, statusCode, err := postJSON(baseURL+"/pg/v4/payment/request.json", reqPayload, nil)
	if err != nil {
		return nil, err
	}

	if statusCode >= 400 {
		return nil, fmt.Errorf("zarinpal payment request failed with status %d", statusCode)
	}

	var decoded struct {
		Data struct {
			Authority string `json:"authority"`
			Code      int    `json:"code"`
		} `json:"data"`
		Errors interface{} `json:"errors"`
	}
	if err := json.Unmarshal(respBody, &decoded); err != nil {
		return nil, err
	}

	if decoded.Data.Authority == "" {
		return nil, fmt.Errorf("zarinpal authority missing")
	}

	return &paymentCreateResult{
		Authority:  decoded.Data.Authority,
		PaymentURL: "https://www.zarinpal.com/pg/StartPay/" + decoded.Data.Authority,
		Raw: map[string]interface{}{
			"provider":      "zarinpal",
			"request_resp":  json.RawMessage(respBody),
			"request_code":  decoded.Data.Code,
			"request_error": decoded.Errors,
		},
	}, nil
}

func (p *zarinpalProvider) VerifyPayment(c *fiber.Ctx, authority string) (*paymentVerifyResult, error) {
	status := strings.TrimSpace(c.Query("Status"))
	if !strings.EqualFold(status, "OK") {
		return &paymentVerifyResult{
			Success: false,
			Failure: "payment not approved by customer",
			Raw: map[string]interface{}{
				"status": status,
			},
		}, nil
	}

	intent := c.Locals("topup_intent")

	topup := (*platform.WalletTopUpIntent)(nil)
	if loadedTopup, ok := intent.(*platform.WalletTopUpIntent); ok {
		topup = loadedTopup
	}

	if topup == nil {
		return nil, errors.New("topup intent not loaded")
	}

	merchantID := strings.TrimSpace(os.Getenv("ZARINPAL_MERCHANT_ID"))
	if merchantID == "" {
		return nil, errors.New("ZARINPAL_MERCHANT_ID is not set")
	}

	baseURL := strings.TrimSpace(os.Getenv("ZARINPAL_BASE_URL"))
	if baseURL == "" {
		baseURL = "https://api.zarinpal.com"
	}

	verifyPayload := map[string]interface{}{
		"merchant_id": merchantID,
		"amount":      int64(topup.PayableIRR),
		"authority":   authority,
	}

	respBody, statusCode, err := postJSON(baseURL+"/pg/v4/payment/verify.json", verifyPayload, nil)
	if err != nil {
		return nil, err
	}

	if statusCode >= 400 {
		return nil, fmt.Errorf("zarinpal verify failed with status %d", statusCode)
	}

	var decoded struct {
		Data struct {
			Code  int   `json:"code"`
			RefID int64 `json:"ref_id"`
		} `json:"data"`
		Errors interface{} `json:"errors"`
	}
	if err := json.Unmarshal(respBody, &decoded); err != nil {
		return nil, err
	}

	ok := decoded.Data.Code == 100 || decoded.Data.Code == 101

	return &paymentVerifyResult{
		Success:      ok,
		RefID:        fmt.Sprintf("%d", decoded.Data.RefID),
		GatewayTxnID: authority,
		Raw: map[string]interface{}{
			"provider":     "zarinpal",
			"verify_resp":  json.RawMessage(respBody),
			"verify_code":  decoded.Data.Code,
			"verify_error": decoded.Errors,
		},
		Failure: map[bool]string{true: "", false: "verification rejected"}[ok],
	}, nil
}

func (p *idpayProvider) CreatePayment(c *fiber.Ctx, intent *platform.WalletTopUpIntent) (*paymentCreateResult, error) {
	apiKey := strings.TrimSpace(os.Getenv("IDPAY_API_KEY"))
	if apiKey == "" {
		return nil, errors.New("IDPAY_API_KEY is not set")
	}

	baseURL := strings.TrimSpace(os.Getenv("IDPAY_BASE_URL"))
	if baseURL == "" {
		baseURL = "https://api.idpay.ir/v1.1"
	}

	orderID := "topup_" + intent.PublicID
	reqPayload := map[string]interface{}{
		"order_id": orderID,
		"amount":   int64(intent.PayableIRR),
		"name":     "GodarAI User",
		"phone":    "",
		"mail":     "",
		"desc":     "GodarAI wallet top-up",
		"callback": intent.CallbackURL,
	}

	headers := map[string]string{
		"X-API-KEY":    apiKey,
		"Content-Type": "application/json",
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("IDPAY_SANDBOX")), "true") {
		headers["X-SANDBOX"] = "1"
	}

	respBody, statusCode, err := postJSON(baseURL+"/payment", reqPayload, headers)
	if err != nil {
		return nil, err
	}

	if statusCode >= 400 {
		return nil, fmt.Errorf("idpay payment request failed with status %d", statusCode)
	}

	var decoded struct {
		ID      string `json:"id"`
		Link    string `json:"link"`
		OrderID string `json:"order_id"`
	}
	if err := json.Unmarshal(respBody, &decoded); err != nil {
		return nil, err
	}

	if decoded.ID == "" || decoded.Link == "" {
		return nil, errors.New("idpay response missing id/link")
	}

	return &paymentCreateResult{
		Authority:  decoded.ID,
		PaymentURL: decoded.Link,
		Raw: map[string]interface{}{
			"provider":     "idpay",
			"request_resp": json.RawMessage(respBody),
			"order_id":     decoded.OrderID,
		},
	}, nil
}

func (p *idpayProvider) VerifyPayment(c *fiber.Ctx, authority string) (*paymentVerifyResult, error) {
	apiKey := strings.TrimSpace(os.Getenv("IDPAY_API_KEY"))
	if apiKey == "" {
		return nil, errors.New("IDPAY_API_KEY is not set")
	}

	intent := c.Locals("topup_intent")

	topup := (*platform.WalletTopUpIntent)(nil)
	if loadedTopup, ok := intent.(*platform.WalletTopUpIntent); ok {
		topup = loadedTopup
	}

	if topup == nil {
		return nil, errors.New("topup intent not loaded")
	}

	baseURL := strings.TrimSpace(os.Getenv("IDPAY_BASE_URL"))
	if baseURL == "" {
		baseURL = "https://api.idpay.ir/v1.1"
	}

	reqPayload := map[string]interface{}{
		"id":       authority,
		"order_id": "topup_" + topup.PublicID,
	}

	headers := map[string]string{
		"X-API-KEY":    apiKey,
		"Content-Type": "application/json",
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("IDPAY_SANDBOX")), "true") {
		headers["X-SANDBOX"] = "1"
	}

	respBody, statusCode, err := postJSON(baseURL+"/payment/verify", reqPayload, headers)
	if err != nil {
		return nil, err
	}

	if statusCode >= 400 {
		return nil, fmt.Errorf("idpay verify failed with status %d", statusCode)
	}

	var decoded struct {
		Status  int   `json:"status"`
		TrackID int64 `json:"track_id"`
		Payment struct {
			CardNo string `json:"card_no"`
		} `json:"payment"`
		Payer struct {
			Phone string `json:"phone"`
		} `json:"payer"`
	}
	if err := json.Unmarshal(respBody, &decoded); err != nil {
		return nil, err
	}

	ok := decoded.Status == 100

	return &paymentVerifyResult{
		Success:      ok,
		RefID:        fmt.Sprintf("%d", decoded.TrackID),
		GatewayTxnID: authority,
		CardMask:     decoded.Payment.CardNo,
		PayerPhone:   decoded.Payer.Phone,
		Raw: map[string]interface{}{
			"provider":    "idpay",
			"verify_resp": json.RawMessage(respBody),
			"status":      decoded.Status,
		},
		Failure: map[bool]string{true: "", false: "verification rejected"}[ok],
	}, nil
}

func postJSON(url string, payload interface{}, headers map[string]string) ([]byte, int, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Content-Type", "application/json")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 20 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			return
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return respBody, resp.StatusCode, nil
}

func randomToken(n int) string {
	if n <= 0 {
		return ""
	}

	b := make([]byte, (n+1)/2)
	if _, err := rand.Read(b); err != nil {
		return "fallbacktoken"
	}

	return hex.EncodeToString(b)[:n]
}

func calculateIRR(amountUSD, exchangeRate, taxPercent, feePercent float64) (amountIRR, taxIRR, feeIRR, payable float64) {
	amountIRR = amountUSD * exchangeRate
	taxIRR = amountIRR * (taxPercent / 100)
	feeIRR = amountIRR * (feePercent / 100)
	payable = amountIRR + taxIRR + feeIRR

	return
}

func calculatePayableIRR(amountIRR int64, taxPercent, feePercent float64) (taxIRR, feeIRR, payable float64) {
	base := float64(amountIRR)
	taxIRR = base * (taxPercent / 100)
	feeIRR = base * (feePercent / 100)
	payable = base + taxIRR + feeIRR

	return
}

func fallbackUSDtoIRRRate() float64 {
	const defaultRate = 1850000.0

	raw := strings.TrimSpace(os.Getenv("BILLING_SETTINGS_DEFAULT_USD_TO_IRR"))
	if raw == "" {
		return defaultRate
	}

	parsed, err := strconv.ParseFloat(raw, 64)
	if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) || parsed <= 0 {
		return defaultRate
	}

	return parsed
}

func defaultSubscriptionPaymentProvider() string {
	provider := strings.TrimSpace(strings.ToLower(os.Getenv("SUBSCRIPTION_PAYMENT_PROVIDER")))
	if provider == "" {
		provider = "mock"
	}

	return provider
}

func resolveExternalBaseURL(c *fiber.Ctx) string {
	forwardedHost := firstForwardedValue(c.Get("X-Forwarded-Host"))

	forwardedProto := firstForwardedValue(c.Get("X-Forwarded-Proto"))
	if forwardedHost != "" {
		if forwardedProto == "" {
			forwardedProto = "https"
		}

		return fmt.Sprintf("%s://%s", forwardedProto, forwardedHost)
	}

	return c.BaseURL()
}

func firstForwardedValue(raw string) string {
	if raw == "" {
		return ""
	}

	parts := strings.Split(raw, ",")
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value != "" {
			return value
		}
	}

	return ""
}

func resolveUSDtoIRRRate(c *fiber.Ctx, store walletStore) (*platform.ExchangeRate, error) {
	rate, err := store.GetLatestExchangeRate(c.Context())
	if err == nil && rate != nil && rate.USDtoIRR > 0 {
		return rate, nil
	}

	return &platform.ExchangeRate{
		USDtoIRR:  fallbackUSDtoIRRRate(),
		Source:    "static-fallback",
		FetchedAt: time.Now(),
	}, nil
}

func (h *Handlers) CreateWalletTopUpIntent(c *fiber.Ctx) error {
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	var req createTopUpIntentRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if req.AmountIRR <= 0 && req.AmountUSD <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "amount_irr or amount_usd must be greater than 0")
	}

	settings, err := h.wallets.GetWalletTopUpSettings(c.Context())
	if err != nil {
		return err
	}

	rate, err := resolveUSDtoIRRRate(c, h.wallets)
	if err != nil || rate == nil || rate.USDtoIRR <= 0 {
		return fiber.NewError(fiber.StatusBadGateway, "failed to resolve exchange rate")
	}

	amountUSD := req.AmountUSD
	amountIRR := float64(req.AmountIRR)
	var taxIRR, feeIRR, payable float64

	if req.AmountIRR > 0 {
		amountUSD = float64(req.AmountIRR) / rate.USDtoIRR
		taxIRR, feeIRR, payable = calculatePayableIRR(req.AmountIRR, settings.TaxPercent, settings.FeePercent)
	} else {
		amountIRR, taxIRR, feeIRR, payable = calculateIRR(req.AmountUSD, rate.USDtoIRR, settings.TaxPercent, settings.FeePercent)
	}

	if amountUSD < settings.MinAmountUSD || amountUSD > settings.MaxAmountUSD {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("top-up amount must be between %.2f and %.2f USD equivalent", settings.MinAmountUSD, settings.MaxAmountUSD))
	}

	returnURL := strings.TrimSpace(req.ReturnURL)
	if returnURL == "" {
		returnURL = fmt.Sprintf("https://%s.app.sharah.ir/billing/topup/result", currentTenant.Slug)
	}

	providerName := strings.TrimSpace(req.Provider)
	if providerName == "" {
		providerName = "mock"
	}

	intent, err := h.wallets.CreateWalletTopUpIntent(c.Context(), storage.CreateTopUpIntentParams{
		TenantID:     currentTenant.ID,
		Provider:     providerName,
		AmountUSD:    amountUSD,
		AmountIRR:    amountIRR,
		ExchangeRate: rate.USDtoIRR,
		TaxPercent:   settings.TaxPercent,
		FeePercent:   settings.FeePercent,
		TaxAmountIRR: taxIRR,
		FeeAmountIRR: feeIRR,
		PayableIRR:   payable,
		CallbackURL:  fmt.Sprintf("%s/api/v1/wallet/topup/callback/%s", resolveExternalBaseURL(c), strings.ToLower(providerName)),
		ReturnURL:    returnURL,
		Metadata:     req.Metadata,
	})
	if err != nil {
		return err
	}

	provider := selectProvider(providerName)

	payRes, err := provider.CreatePayment(c, intent)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}

	if err := h.wallets.AttachWalletTopUpGatewayData(c.Context(), intent.PublicID, payRes.Authority, payRes.PaymentURL, payRes.Raw); err != nil {
		return err
	}

	intent.Authority = payRes.Authority
	intent.PaymentURL = payRes.PaymentURL
	intent.RawPayload = payRes.Raw

	return c.Status(fiber.StatusCreated).JSON(intent)
}

func (h *Handlers) WalletTopUpCallback(c *fiber.Ctx) error {
	providerName := strings.ToLower(strings.TrimSpace(c.Params("provider")))

	authority := strings.TrimSpace(c.Query("authority"))
	if authority == "" {
		authority = strings.TrimSpace(c.FormValue("authority"))
	}

	if authority == "" {
		return fiber.NewError(fiber.StatusBadRequest, "authority is required")
	}

	intent, err := h.wallets.GetWalletTopUpIntentByAuthority(c.Context(), authority)
	if err == nil && intent != nil {
		c.Locals("topup_intent", intent)
	}

	provider := selectProvider(providerName)

	verifyRes, err := provider.VerifyPayment(c, authority)
	if err != nil {
		if markErr := h.wallets.MarkWalletTopUpFailedByAuthority(c.Context(), authority, err.Error(), map[string]interface{}{"provider": providerName}); markErr != nil {
			return markErr
		}

		return h.redirectTopUpResult(c, authority, "failed", "verify_error")
	}

	if !verifyRes.Success {
		reason := verifyRes.Failure
		if reason == "" {
			reason = "payment failed"
		}

		if markErr := h.wallets.MarkWalletTopUpFailedByAuthority(c.Context(), authority, reason, verifyRes.Raw); markErr != nil {
			return markErr
		}

		return h.redirectTopUpResult(c, authority, "failed", "payment_failed")
	}

	intent, alreadyProcessed, err := h.wallets.CompleteWalletTopUpByAuthority(c.Context(), authority, verifyRes.Raw, verifyRes.RefID, verifyRes.GatewayTxnID, verifyRes.CardMask, verifyRes.PayerPhone)
	if err != nil {
		if errors.Is(err, storage.ErrTopUpIntentNotFound) {
			return h.redirectTopUpResult(c, authority, "failed", "intent_not_found")
		}

		return h.redirectTopUpResult(c, authority, "failed", "commit_failed")
	}

	status := "success"
	if alreadyProcessed {
		status = "success"
	}

	return c.Redirect(fmt.Sprintf("%s?status=%s&intent_id=%s&authority=%s&ref_id=%s", intent.ReturnURL, status, intent.PublicID, authority, intent.RefID), fiber.StatusFound)
}

func (h *Handlers) redirectTopUpResult(c *fiber.Ctx, authority, status, reason string) error {
	intent, err := h.wallets.GetWalletTopUpIntentByAuthority(c.Context(), authority)
	if err != nil || intent == nil || strings.TrimSpace(intent.ReturnURL) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": status, "reason": reason, "authority": authority})
	}

	return c.Redirect(fmt.Sprintf("%s?status=%s&authority=%s&reason=%s", intent.ReturnURL, status, authority, reason), fiber.StatusFound)
}

func (h *Handlers) ListWalletTransactions(c *fiber.Ctx) error {
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	limit := c.QueryInt("limit", 50)

	items, err := h.wallets.ListWalletTransactions(c.Context(), currentTenant.ID, limit)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"data": items})
}

func (h *Handlers) GetWalletTopUpSettings(c *fiber.Ctx) error {
	settings, err := h.wallets.GetWalletTopUpSettings(c.Context())
	if err != nil {
		return err
	}

	return c.JSON(settings)
}

func (h *Handlers) UpdateWalletTopUpSettings(c *fiber.Ctx) error {
	var req updateTopUpSettingsRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if req.MinAmountUSD <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "min_amount_usd must be greater than 0")
	}

	if req.MaxAmountUSD < req.MinAmountUSD {
		return fiber.NewError(fiber.StatusBadRequest, "max_amount_usd must be greater than or equal to min_amount_usd")
	}

	if req.TaxPercent < 0 || req.FeePercent < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "tax_percent and fee_percent must be non-negative")
	}

	settings, err := h.wallets.UpdateWalletTopUpSettings(c.Context(), req.MinAmountUSD, req.MaxAmountUSD, req.TaxPercent, req.FeePercent)
	if err != nil {
		return err
	}

	return c.JSON(settings)
}
