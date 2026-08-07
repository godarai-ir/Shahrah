package storage

import (
	"context"
	"testing"

	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/stretchr/testify/assert"
)

func TestUpdateCommercialSignupSettingsRejectsInvalidCreditBeforePersistence(t *testing.T) {
	store := composePlanRepository(&planState{})
	for _, credit := range []int64{-1, 1000000000000000000} {
		_, err := store.UpdateCommercialSignupSettings(context.Background(), credit, "test")
		assert.ErrorIs(t, err, billingdomain.ErrInvalidSignupCredit)
	}
}
