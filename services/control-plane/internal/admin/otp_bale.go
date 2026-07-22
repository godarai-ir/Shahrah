package admin

import (
	"bytes"
	crand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
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

func newBaleSafirOTPSenderFromEnv() OTPMobileSender {
	apiAccessKey := strings.TrimSpace(os.Getenv("BALE_SAFIR_API_ACCESS_KEY"))
	botIDRaw := strings.TrimSpace(os.Getenv("BALE_SAFIR_BOT_ID"))
	enabledRaw := strings.TrimSpace(os.Getenv("BALE_OTP_ENABLED"))
	enabled := apiAccessKey != "" || botIDRaw != ""
	if enabledRaw != "" {
		enabled = strings.EqualFold(enabledRaw, "true")
	}

	botID, err := strconv.ParseInt(botIDRaw, 10, 64)
	if err != nil || botID <= 0 {
		err = fmt.Errorf("BALE_SAFIR_BOT_ID must be a positive integer")
	}
	if apiAccessKey == "" {
		err = fmt.Errorf("BALE_SAFIR_API_ACCESS_KEY is required")
	}

	return &baleSafirOTPSender{
		enabled:      enabled,
		apiAccessKey: apiAccessKey,
		botID:        botID,
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

func (s *baleSafirOTPSender) SendTenantAdminOTP(to string, _ string, otpCode string, _ int) error {
	if !s.enabled {
		return nil
	}
	if s.configErr != nil {
		return fmt.Errorf("Bale Safir OTP configuration is invalid: %w", s.configErr)
	}

	phoneNumber, err := normalizeBalePhoneNumber(to)
	if err != nil {
		return err
	}
	if !baleOTPPattern.MatchString(otpCode) {
		return fmt.Errorf("Bale Safir only supports numeric OTP values")
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
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
		return fmt.Errorf("Bale Safir OTP request returned HTTP %d", resp.StatusCode)
	}

	var result baleSafirOTPResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&result); err != nil {
		return fmt.Errorf("decode Bale Safir OTP response: %w", err)
	}
	if len(result.Errors) > 0 {
		providerErr := result.Errors[0]
		return fmt.Errorf("Bale Safir rejected OTP request (code %d)", providerErr.Code)
	}
	if strings.TrimSpace(result.MessageID) == "" {
		return fmt.Errorf("Bale Safir OTP response did not include message_id")
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
