package admin

import (
	"bytes"
	crand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const baleSafirSendMessageURL = "https://safir.bale.ai/api/v3/send_message"

var balePhoneNumberPattern = regexp.MustCompile(`^98\d{10}$`)
var baleOTPPattern = regexp.MustCompile(`^\d+$`)

type OTPMobileSender interface {
	SendTenantAdminOTP(to string, tenantSlug string, otpCode string, expiresInMinutes int) error
}

type baleSafirOTPSender struct {
	enabled      bool
	apiAccessKey string
	botID        int64
	configErr    error
	endpoint     string
	client       *http.Client
	requestID    func() string
}

// OTPMobileConfig contains immutable Bale Safir delivery settings.
type OTPMobileConfig struct {
	Enabled      bool
	APIAccessKey string
	BotID        int64
}

type baleSafirOTPRequest struct {
	RequestID   string                  `json:"request_id"`
	BotID       int64                   `json:"bot_id"`
	PhoneNumber string                  `json:"phone_number"`
	MessageData baleSafirOTPMessageData `json:"message_data"`
}

type baleSafirOTPMessageData struct {
	OTPMessage baleSafirOTPMessage `json:"otp_message"`
}

type baleSafirOTPMessage struct {
	OTP string `json:"otp"`
}

type baleSafirOTPResponse struct {
	MessageID string           `json:"message_id"`
	Errors    []baleSafirError `json:"error_data"`
}

type baleSafirError struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

func newBaleSafirOTPSender(cfg OTPMobileConfig) OTPMobileSender {
	var err error
	if cfg.BotID <= 0 {
		err = fmt.Errorf("BALE_SAFIR_BOT_ID must be a positive integer")
	}

	if strings.TrimSpace(cfg.APIAccessKey) == "" {
		err = fmt.Errorf("BALE_SAFIR_API_ACCESS_KEY is required")
	}

	return &baleSafirOTPSender{
		enabled:      cfg.Enabled,
		apiAccessKey: strings.TrimSpace(cfg.APIAccessKey),
		botID:        cfg.BotID,
		configErr:    err,
		endpoint:     baleSafirSendMessageURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		requestID: newBaleSafirRequestID,
	}
}

func (s *baleSafirOTPSender) SendTenantAdminOTP(to string, _ string, otpCode string, _ int) (sendErr error) {
	if !s.enabled {
		return nil
	}

	if s.configErr != nil {
		return fmt.Errorf("invalid Bale Safir OTP configuration: %w", s.configErr)
	}

	phoneNumber, err := normalizeBalePhoneNumber(to)
	if err != nil {
		return err
	}

	if !baleOTPPattern.MatchString(otpCode) {
		return fmt.Errorf("numeric OTP value required for Bale Safir")
	}

	payload, err := json.Marshal(baleSafirOTPRequest{
		RequestID:   s.requestID(),
		BotID:       s.botID,
		PhoneNumber: phoneNumber,
		MessageData: baleSafirOTPMessageData{
			OTPMessage: baleSafirOTPMessage{OTP: otpCode},
		},
	})
	if err != nil {
		return fmt.Errorf("encode Bale Safir OTP request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, s.endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create Bale Safir OTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-access-key", s.apiAccessKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send Bale Safir OTP request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); sendErr == nil && closeErr != nil {
			sendErr = fmt.Errorf("close Bale Safir OTP response body: %w", closeErr)
		}
	}()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		if _, err := io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10)); err != nil {
			return fmt.Errorf("discard Bale Safir OTP error response: %w", err)
		}

		return fmt.Errorf("unexpected Bale Safir OTP HTTP status %d", resp.StatusCode)
	}

	var result baleSafirOTPResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&result); err != nil {
		return fmt.Errorf("decode Bale Safir OTP response: %w", err)
	}

	if len(result.Errors) > 0 {
		providerErr := result.Errors[0]
		return fmt.Errorf("request rejected by Bale Safir (code %d)", providerErr.Code)
	}

	if strings.TrimSpace(result.MessageID) == "" {
		return fmt.Errorf("missing message_id in Bale Safir OTP response")
	}

	return nil
}

func normalizeBalePhoneNumber(value string) (string, error) {
	compact := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "").Replace(strings.TrimSpace(value))
	compact = strings.TrimPrefix(compact, "+")

	switch {
	case strings.HasPrefix(compact, "0098"):
		compact = compact[2:]
	case len(compact) == 11 && strings.HasPrefix(compact, "0"):
		compact = "98" + compact[1:]
	case len(compact) == 10:
		compact = "98" + compact
	}

	if !balePhoneNumberPattern.MatchString(compact) {
		return "", fmt.Errorf("mobile number must contain 10 digits after Iran country code 98")
	}

	return compact, nil
}

func newBaleSafirRequestID() string {
	value := make([]byte, 16)
	if _, err := crand.Read(value); err != nil {
		return fmt.Sprintf("tenant-admin-otp-%d", time.Now().UnixNano())
	}

	return hex.EncodeToString(value)
}
