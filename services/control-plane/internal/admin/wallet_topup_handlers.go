package admin

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
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
type zarinpalProvider struct {
	merchantID string
	baseURL    string
}
type idpayProvider struct {
	apiKey  string
	baseURL string
	sandbox bool
}
type unsupportedProvider struct{ name string }

// PaymentConfig contains immutable wallet and subscription gateway settings.
type PaymentConfig struct {
	FallbackUSDtoIRR float64
	// Production refuses the mock gateway, which approves every payment, and
	// stops it from being the default provider.
	Production                  bool
	DefaultSubscriptionProvider string
	ZarinpalMerchantID          string
	ZarinpalBaseURL             string
	IDPayAPIKey                 string
	IDPayBaseURL                string
	IDPaySandbox                bool
}

func (c PaymentConfig) withDefaults() PaymentConfig {
	if c.FallbackUSDtoIRR <= 0 {
		c.FallbackUSDtoIRR = 1850000
	}

	if strings.TrimSpace(c.DefaultSubscriptionProvider) == "" && !c.Production {
		c.DefaultSubscriptionProvider = "mock"
	}

	c.ZarinpalBaseURL = defaultIfEmpty(strings.TrimSpace(c.ZarinpalBaseURL), "https://api.zarinpal.com")
	c.IDPayBaseURL = defaultIfEmpty(strings.TrimSpace(c.IDPayBaseURL), "https://api.idpay.ir/v1.1")

	return c
}

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

func (h *Handlers) selectPaymentProvider(name string) paymentProvider {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "mock":
		if h.payment.Production {
			return &unsupportedProvider{name: "mock"}
		}

		return &mockProvider{}
	case "zarinpal":
		return &zarinpalProvider{
			merchantID: strings.TrimSpace(h.payment.ZarinpalMerchantID),
			baseURL:    h.payment.ZarinpalBaseURL,
		}
	case "idpay":
		return &idpayProvider{
			apiKey:  strings.TrimSpace(h.payment.IDPayAPIKey),
			baseURL: h.payment.IDPayBaseURL,
			sandbox: h.payment.IDPaySandbox,
		}
	default:
		return &unsupportedProvider{name: name}
	}
}

func (p *zarinpalProvider) CreatePayment(c *fiber.Ctx, intent *platform.WalletTopUpIntent) (*paymentCreateResult, error) {
	if p.merchantID == "" {
		return nil, errors.New("ZARINPAL_MERCHANT_ID is not set")
	}

	reqPayload := map[string]interface{}{
		"merchant_id":  p.merchantID,
		"amount":       int64(intent.PayableIRR),
		"callback_url": intent.CallbackURL,
		"description":  fmt.Sprintf("GodarAI wallet topup %s", intent.PublicID),
		"metadata": map[string]interface{}{
			"mobile": "",
			"email":  "",
		},
	}

	respBody, statusCode, err := postJSON(p.baseURL+"/pg/v4/payment/request.json", reqPayload, nil)
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

	topup := topUpIntentFromContext(c)
	if topup == nil {
		return nil, errors.New("topup intent not loaded")
	}

	if p.merchantID == "" {
		return nil, errors.New("ZARINPAL_MERCHANT_ID is not set")
	}

	verifyPayload := map[string]interface{}{
		"merchant_id": p.merchantID,
		"amount":      int64(topup.PayableIRR),
		"authority":   authority,
	}

	respBody, statusCode, err := postJSON(p.baseURL+"/pg/v4/payment/verify.json", verifyPayload, nil)
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
	if p.apiKey == "" {
		return nil, errors.New("IDPAY_API_KEY is not set")
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
		"X-API-KEY":    p.apiKey,
		"Content-Type": "application/json",
	}
	if p.sandbox {
		headers["X-SANDBOX"] = "1"
	}

	respBody, statusCode, err := postJSON(p.baseURL+"/payment", reqPayload, headers)
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
	if p.apiKey == "" {
		return nil, errors.New("IDPAY_API_KEY is not set")
	}

	topup := topUpIntentFromContext(c)
	if topup == nil {
		return nil, errors.New("topup intent not loaded")
	}

	reqPayload := map[string]interface{}{
		"id":       authority,
		"order_id": "topup_" + topup.PublicID,
	}

	headers := map[string]string{
		"X-API-KEY":    p.apiKey,
		"Content-Type": "application/json",
	}
	if p.sandbox {
		headers["X-SANDBOX"] = "1"
	}

	respBody, statusCode, err := postJSON(p.baseURL+"/payment/verify", reqPayload, headers)
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

func postJSON(endpoint string, payload interface{}, headers map[string]string) ([]byte, int, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
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
	taxIRR, feeIRR, payable = calculatePayableIRR(amountIRR, taxPercent, feePercent)

	return
}

func calculatePayableIRR(amountIRR float64, taxPercent, feePercent float64) (taxIRR, feeIRR, payable float64) {
	taxIRR = amountIRR * (taxPercent / 100)
	feeIRR = amountIRR * (feePercent / 100)
	payable = amountIRR + taxIRR + feeIRR

	return
}

// validatedReturnURL keeps a requested return URL only when it is on this site, so a
// payment cannot send the buyer to another website. An empty value gets the fallback.
func validatedReturnURL(c *fiber.Ctx, requested, fallback string) (string, error) {
	requested = strings.TrimSpace(requested)
	if requested == "" {
		return fallback, nil
	}

	target, err := url.Parse(requested)
	if err != nil {
		return "", fiber.NewError(fiber.StatusBadRequest, "return_url is invalid")
	}

	origin, err := url.Parse(resolveExternalBaseURL(c))
	if err != nil {
		return "", err
	}

	if target.Scheme != origin.Scheme || !strings.EqualFold(target.Host, origin.Host) {
		return "", fiber.NewError(fiber.StatusBadRequest, "return_url must be on this site")
	}

	return requested, nil
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

func resolveUSDtoIRRRate(c *fiber.Ctx, store walletStore, fallback float64) (*platform.ExchangeRate, error) {
	rate, err := store.GetLatestExchangeRate(c.Context())
	if err == nil && rate != nil && rate.USDtoIRR > 0 {
		return rate, nil
	}

	return &platform.ExchangeRate{
		USDtoIRR:  fallback,
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

	rate, err := resolveUSDtoIRRRate(c, h.wallets, h.payment.FallbackUSDtoIRR)
	if err != nil || rate == nil || rate.USDtoIRR <= 0 {
		return fiber.NewError(fiber.StatusBadGateway, "failed to resolve exchange rate")
	}

	amountUSD := req.AmountUSD
	amountIRR := float64(req.AmountIRR)

	var taxIRR, feeIRR, payable float64

	if req.AmountIRR > 0 {
		amountUSD = float64(req.AmountIRR) / rate.USDtoIRR
		taxIRR, feeIRR, payable = calculatePayableIRR(float64(req.AmountIRR), settings.TaxPercent, settings.FeePercent)
	} else {
		amountIRR, taxIRR, feeIRR, payable = calculateIRR(req.AmountUSD, rate.USDtoIRR, settings.TaxPercent, settings.FeePercent)
	}

	if amountUSD < settings.MinAmountUSD || amountUSD > settings.MaxAmountUSD {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("top-up amount must be between %.2f and %.2f USD equivalent", settings.MinAmountUSD, settings.MaxAmountUSD))
	}

	returnURL, err := validatedReturnURL(c, req.ReturnURL, fmt.Sprintf("https://%s.app.sharah.ir/billing/topup/result", currentTenant.Slug))
	if err != nil {
		return err
	}

	providerName := strings.TrimSpace(req.Provider)
	if providerName == "" {
		providerName = h.payment.DefaultSubscriptionProvider
	}

	// Refuse an unavailable provider before a pending intent is written.
	if _, refused := h.selectPaymentProvider(providerName).(*unsupportedProvider); refused {
		return creditPackError(c, billingdomain.ErrCreditPaymentProviderUnavailable)
	}

	intent, err := h.wallets.CreateWalletTopUpIntent(c.Context(), billingdomain.CreateTopUpIntentParams{
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
		return creditPackError(c, err)
	}

	provider := h.selectPaymentProvider(providerName)

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
		setTopUpIntent(c, intent)
	}

	provider := h.selectPaymentProvider(providerName)

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

		if topup := topUpIntentFromContext(c); topup != nil && string(topup.Status) == string(platform.WalletTopUpStatusCompleted) {
			return redirectTopUpSuccess(c, topup)
		}

		if markErr := h.wallets.MarkWalletTopUpFailedByAuthority(c.Context(), authority, reason, verifyRes.Raw); markErr != nil {
			return markErr
		}

		return h.redirectTopUpResult(c, authority, "failed", "payment_failed")
	}

	intent, _, err = h.wallets.CompleteWalletTopUpByAuthority(c.Context(), authority, verifyRes.Raw, verifyRes.RefID, verifyRes.GatewayTxnID, verifyRes.CardMask, verifyRes.PayerPhone)
	if err != nil {
		if errors.Is(err, billingdomain.ErrTopUpIntentNotFound) {
			return h.redirectTopUpResult(c, authority, "failed", "intent_not_found")
		}

		return h.redirectTopUpResult(c, authority, "failed", "commit_failed")
	}

	return redirectTopUpSuccess(c, intent)
}

func redirectTopUpSuccess(c *fiber.Ctx, intent *platform.WalletTopUpIntent) error {
	return c.Redirect(fmt.Sprintf("%s?status=success&intent_id=%s&authority=%s&ref_id=%s", intent.ReturnURL, intent.PublicID, intent.Authority, intent.RefID), fiber.StatusFound)
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
