package platform

import (
	"encoding/json"
	"time"
)

const (
	VideoGenerationSubmitting    = "submitting"
	VideoGenerationQueued        = "queued"
	VideoGenerationRunning       = "running"
	VideoGenerationSucceeded     = "succeeded"
	VideoGenerationFailed        = "failed"
	VideoGenerationCancelled     = "cancelled"
	VideoGenerationBillingReview = "billing_review"

	VideoDeliveryPending = "pending"
	VideoDeliveryCopying = "copying"
	VideoDeliveryReady   = "ready"
	VideoDeliveryFailed  = "failed"
	VideoDeliveryDeleted = "deleted"

	VideoSettlementPending   = "pending"
	VideoSettlementCommitted = "committed"
	VideoSettlementReleased  = "released"
	VideoSettlementReview    = "review"
)

// VideoCreateRequest is the provider-independent video generation request.
type VideoCreateRequest struct {
	Model           string                 `json:"model"`
	Prompt          string                 `json:"prompt"`
	InputMediaID    string                 `json:"input_media_id,omitempty"`
	Seconds         int                    `json:"seconds"`
	Size            string                 `json:"size,omitempty"`
	Resolution      string                 `json:"resolution,omitempty"`
	AspectRatio     string                 `json:"aspect_ratio,omitempty"`
	Count           int                    `json:"n,omitempty"`
	ServiceTier     string                 `json:"service_tier,omitempty"`
	Seed            *int64                 `json:"seed,omitempty"`
	ProviderOptions map[string]interface{} `json:"provider_options,omitempty"`
}

// VideoArtifact is one durably managed generated output.
type VideoArtifact struct {
	ID             string `json:"id"`
	ObjectKey      string `json:"object_key,omitempty"`
	ProviderURL    string `json:"provider_url,omitempty"`
	ContentType    string `json:"content_type,omitempty"`
	SizeBytes      int64  `json:"size_bytes,omitempty"`
	ChecksumSHA256 string `json:"checksum_sha256,omitempty"`
	Duration       string `json:"duration_seconds,omitempty"`
	Resolution     string `json:"resolution,omitempty"`
}

// MarshalJSON prevents provider-temporary output URLs and internal object keys
// from crossing the public API boundary while allowing the persistence layer to
// serialize VideoArtifact values directly for delivery retries.
func (j VideoGenerationJob) MarshalJSON() ([]byte, error) {
	type jobAlias VideoGenerationJob
	safe := jobAlias(j)
	safe.Request.InputMediaID = ""
	safe.Artifacts = append([]VideoArtifact(nil), j.Artifacts...)
	for i := range safe.Artifacts {
		safe.Artifacts[i].ProviderURL = ""
		safe.Artifacts[i].ObjectKey = ""
	}
	return json.Marshal(safe)
}

// VideoGenerationJob is the tenant-safe public and persistence projection.
type VideoGenerationJob struct {
	ID                string                 `json:"id"`
	TenantID          string                 `json:"-"`
	RequestID         string                 `json:"request_id"`
	IdempotencyKey    string                 `json:"-"`
	TokenHash         string                 `json:"-"`
	ActorID           string                 `json:"-"`
	OwnerPrincipal    string                 `json:"-"`
	Model             string                 `json:"model"`
	Provider          string                 `json:"provider"`
	ProviderModel     string                 `json:"provider_model"`
	DeploymentID      string                 `json:"-"`
	ProviderAccountID string                 `json:"-"`
	Region            string                 `json:"region,omitempty"`
	BillingMode       string                 `json:"-"`
	PricingSource     string                 `json:"pricing_source,omitempty"`
	ProviderJobID     string                 `json:"-"`
	GenerationStatus  string                 `json:"generation_status"`
	DeliveryStatus    string                 `json:"delivery_status"`
	SettlementStatus  string                 `json:"settlement_status"`
	Request           VideoCreateRequest     `json:"request"`
	RequestedUsage    []UsageMeasure         `json:"requested_usage"`
	ActualUsage       []UsageMeasure         `json:"actual_usage,omitempty"`
	PricingSnapshot   *ModelPrice            `json:"-"`
	EstimatedRating   *RatedUsage            `json:"estimated_rating,omitempty"`
	FinalRating       *RatedUsage            `json:"final_rating,omitempty"`
	ExchangeRate      string                 `json:"-"`
	BudgetReservation *BudgetReservation     `json:"-"`
	ProviderResponse  map[string]interface{} `json:"-"`
	Artifacts         []VideoArtifact        `json:"artifacts,omitempty"`
	LastError         string                 `json:"last_error,omitempty"`
	NextPollAt        *time.Time             `json:"-"`
	LeaseOwner        string                 `json:"-"`
	LeaseExpiresAt    *time.Time             `json:"-"`
	AttemptCount      int                    `json:"-"`
	Version           int64                  `json:"-"`
	CompletedAt       *time.Time             `json:"completed_at,omitempty"`
	CancelledAt       *time.Time             `json:"cancelled_at,omitempty"`
	ExpiresAt         time.Time              `json:"expires_at"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

type VideoJobList struct {
	Data       []VideoGenerationJob `json:"data"`
	NextCursor string               `json:"next_cursor,omitempty"`
}
