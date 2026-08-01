package admin

import (
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestValidateCreateSupportTicketRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		request   platform.CreateSupportTicketRequest
		fileCount int
		wantError bool
	}{
		{"valid text", platform.CreateSupportTicketRequest{Subject: "API request fails", Category: "technical", Body: "Steps to reproduce"}, 0, false},
		{"valid attachment only", platform.CreateSupportTicketRequest{Subject: "Review attached log", Category: "technical"}, 1, false},
		{"short subject", platform.CreateSupportTicketRequest{Subject: "bad", Category: "technical", Body: "details"}, 0, true},
		{"unknown category", platform.CreateSupportTicketRequest{Subject: "Account problem", Category: "unknown", Body: "details"}, 0, true},
		{"empty message", platform.CreateSupportTicketRequest{Subject: "Account problem", Category: "account"}, 0, true},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := validateCreateSupportTicketRequest(test.request, test.fileCount)
			if (err != nil) != test.wantError {
				t.Fatalf("validateCreateSupportTicketRequest() error = %v wantError=%v", err, test.wantError)
			}
		})
	}
}

func TestOffsetCursorRoundTripAndRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	cursor := encodeOffsetCursor(75)

	offset, err := decodeOffsetCursor(cursor)
	if err != nil || offset != 75 {
		t.Fatalf("cursor round trip got offset=%d err=%v", offset, err)
	}

	if _, err := decodeOffsetCursor("not-a-cursor"); err == nil {
		t.Fatal("expected invalid cursor to be rejected")
	}
}
