package platform

import "time"

const (
	SupportTicketStatusOpen               = "open"
	SupportTicketStatusInProgress         = "in_progress"
	SupportTicketStatusWaitingForCustomer = "waiting_for_customer"
	SupportTicketStatusResolved           = "resolved"
	SupportTicketStatusClosed             = "closed"
)

const (
	SupportTicketPriorityLow    = "low"
	SupportTicketPriorityNormal = "normal"
	SupportTicketPriorityHigh   = "high"
	SupportTicketPriorityUrgent = "urgent"
)

const (
	SupportTicketCategoryTechnical      = "technical"
	SupportTicketCategoryBilling        = "billing"
	SupportTicketCategoryAccount        = "account"
	SupportTicketCategoryFeatureRequest = "feature_request"
	SupportTicketCategoryOther          = "other"
)

const (
	SupportTicketAuthorTenantUser   = "tenant_user"
	SupportTicketAuthorSupportAgent = "support_agent"
	SupportTicketAuthorSystem       = "system"
)

type SupportTicketActor struct {
	Type      string
	ID        string
	Role      string
	RequestID string
}

type SupportTicketAttachment struct {
	ID                  string     `json:"id"`
	TicketID            string     `json:"ticket_id"`
	MessageID           string     `json:"message_id"`
	OriginalFilename    string     `json:"original_filename"`
	DetectedContentType string     `json:"detected_content_type"`
	SizeBytes           int64      `json:"size_bytes"`
	SHA256              string     `json:"sha256"`
	ScanStatus          string     `json:"scan_status"`
	CreatedAt           time.Time  `json:"created_at"`
	ScannedAt           *time.Time `json:"scanned_at,omitempty"`
}

// SupportTicketAttachmentUpload is trusted metadata produced by the control-plane
// attachment service after object storage and malware scanning succeed.
type SupportTicketAttachmentUpload struct {
	ID                  string
	OriginalFilename    string
	DetectedContentType string
	SizeBytes           int64
	SHA256              string
	ObjectKey           string
	ScanStatus          string
	ScanResult          string
	ScannedAt           *time.Time
}

type SupportTicketMessage struct {
	ID          string                    `json:"id"`
	TicketID    string                    `json:"ticket_id"`
	AuthorType  string                    `json:"author_type"`
	AuthorID    string                    `json:"author_id"`
	AuthorLabel string                    `json:"author_label,omitempty"`
	Body        string                    `json:"body"`
	Attachments []SupportTicketAttachment `json:"attachments"`
	CreatedAt   time.Time                 `json:"created_at"`
}

type SupportTicketRating struct {
	Score       int       `json:"score"`
	Comment     string    `json:"comment,omitempty"`
	SubmittedBy string    `json:"submitted_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SupportTicket struct {
	ID                     string               `json:"id"`
	PublicID               string               `json:"public_id"`
	TenantID               string               `json:"tenant_id"`
	TenantSlug             string               `json:"tenant_slug,omitempty"`
	RequesterUserID        string               `json:"requester_user_id"`
	RequesterDisplayName   string               `json:"requester_display_name,omitempty"`
	RequesterEmail         string               `json:"requester_email,omitempty"`
	Subject                string               `json:"subject"`
	Category               string               `json:"category"`
	Priority               string               `json:"priority"`
	Status                 string               `json:"status"`
	AssigneeID             string               `json:"assignee_id,omitempty"`
	LastMessageAuthorType  string               `json:"last_message_author_type"`
	LastMessageAt          time.Time            `json:"last_message_at"`
	FirstSupportResponseAt *time.Time           `json:"first_support_response_at,omitempty"`
	ResolvedAt             *time.Time           `json:"resolved_at,omitempty"`
	ClosedAt               *time.Time           `json:"closed_at,omitempty"`
	CreatedAt              time.Time            `json:"created_at"`
	UpdatedAt              time.Time            `json:"updated_at"`
	Version                int64                `json:"version"`
	Rating                 *SupportTicketRating `json:"rating,omitempty"`
}

type SupportTicketListParams struct {
	TenantID        string
	RequesterUserID string
	ReadAll         bool
	Queue           string
	Status          string
	Category        string
	Priority        string
	AssigneeID      string
	Query           string
	UpdatedFrom     *time.Time
	UpdatedTo       *time.Time
	Limit           int
	Offset          int
}

type SupportTicketListResponse struct {
	Data       []SupportTicket `json:"data"`
	NextCursor string          `json:"next_cursor,omitempty"`
}

type SupportTicketCounts struct {
	NeedsSupportReply  int      `json:"needs_support_reply"`
	Open               int      `json:"open"`
	InProgress         int      `json:"in_progress"`
	WaitingForCustomer int      `json:"waiting_for_customer"`
	Resolved           int      `json:"resolved"`
	Closed             int      `json:"closed"`
	All                int      `json:"all"`
	AverageRating      *float64 `json:"average_rating,omitempty"`
}

type SupportTicketDetail struct {
	Ticket   SupportTicket          `json:"ticket"`
	Messages []SupportTicketMessage `json:"messages"`
}

type CreateSupportTicketRequest struct {
	Subject        string `json:"subject"`
	Category       string `json:"category"`
	Body           string `json:"body"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
	RequestHash    string `json:"-"`
}

type AddSupportTicketMessageRequest struct {
	Body           string `json:"body"`
	NextStatus     string `json:"next_status,omitempty"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
	RequestHash    string `json:"-"`
}

type TriageSupportTicketRequest struct {
	Status         *string `json:"status,omitempty"`
	Priority       *string `json:"priority,omitempty"`
	AssigneeID     *string `json:"assignee_id,omitempty"`
	Reason         string  `json:"reason,omitempty"`
	Version        int64   `json:"version"`
	IdempotencyKey string  `json:"idempotency_key,omitempty"`
}

type RateSupportTicketRequest struct {
	Score          int    `json:"score"`
	Comment        string `json:"comment,omitempty"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

type ReopenSupportTicketRequest struct {
	Reason         string `json:"reason"`
	Version        int64  `json:"version"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

type SupportTicketEvent struct {
	ID        string                 `json:"id"`
	TicketID  string                 `json:"ticket_id"`
	TenantID  string                 `json:"tenant_id"`
	Type      string                 `json:"type"`
	ActorType string                 `json:"actor_type"`
	ActorID   string                 `json:"actor_id,omitempty"`
	ActorRole string                 `json:"actor_role,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

type SupportTicketAttachmentObject struct {
	SupportTicketAttachment
	TenantID  string
	ObjectKey string
}
