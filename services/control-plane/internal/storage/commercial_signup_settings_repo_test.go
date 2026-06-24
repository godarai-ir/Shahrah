package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateCommercialSignupSettingsRejectsInvalidCreditBeforePersistence(t *testing.T) {
	store := &Store{}
	for _, credit := range []int64{-1, 1000000000000000000} {
		_, err := store.UpdateCommercialSignupSettings(context.Background(), credit, "test")
		assert.ErrorIs(t, err, ErrInvalidSignupCredit)
	}
}
