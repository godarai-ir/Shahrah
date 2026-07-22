package admin

import "strings"

// TenantAdminOTPNotifier sends tenant-admin OTPs through the delivery provider
// associated with the requested authentication channel.
type TenantAdminOTPNotifier interface {
	SendTenantAdminOTP(channel string, to string, tenantSlug string, otpCode string, expiresInMinutes int) error
}

type tenantAdminOTPNotifier struct {
	email  OTPEmailSender
	mobile OTPMobileSender
}

func NewTenantAdminOTPNotifierFromEnv() TenantAdminOTPNotifier {
	return &tenantAdminOTPNotifier{
		email:  NewOTPEmailSenderFromEnv(),
		mobile: newBaleSafirOTPSenderFromEnv(),
	}
}

func (n *tenantAdminOTPNotifier) SendTenantAdminOTP(channel string, to string, tenantSlug string, otpCode string, expiresInMinutes int) error {
	switch strings.ToLower(strings.TrimSpace(channel)) {
	case "email":
		if n.email == nil {
			return nil
		}

		return n.email.SendTenantAdminOTP(to, tenantSlug, otpCode, expiresInMinutes)
	case "mobile":
		if n.mobile == nil {
			return nil
		}

		return n.mobile.SendTenantAdminOTP(to, tenantSlug, otpCode, expiresInMinutes)
	default:
		return nil
	}
}
