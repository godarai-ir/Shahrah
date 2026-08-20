package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// PublishedChatbot contains the immutable runtime configuration.
type PublishedChatbot struct {
	Chatbot    platform.Chatbot
	Version    platform.ChatbotPublication
	TenantSlug string
}

// CreateChatbot creates a private-by-default mutable draft.
func (s *Store) CreateChatbot(ctx context.Context, tenantID, actor string, request platform.UpsertChatbotRequest) (*platform.Chatbot, error) {
	publicID := "bot_" + strings.ReplaceAll(uuid.NewString(), "-", "")

	branding, err := json.Marshal(normalizeBranding(request.Branding))
	if err != nil {
		return nil, fmt.Errorf("encode chatbot branding: %w", err)
	}

	var id string

	err = s.metadata.QueryRow(ctx, `
		INSERT INTO chatbots (
			public_id, tenant_id, name, description, system_instruction, answer_model, tone,
			default_locale, answer_language, access_mode, retention_days, knowledge_base_ids,
			allowed_origins, branding, created_by
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12::uuid[],$13::text[],$14,$15)
		RETURNING chatbot_id::text`,
		publicID, tenantID, strings.TrimSpace(request.Name), strings.TrimSpace(request.Description),
		strings.TrimSpace(request.SystemInstruction), strings.TrimSpace(request.AnswerModel), normalizeTone(request.Tone),
		normalizeLocale(request.DefaultLocale), normalizeAnswerLanguage(request.AnswerLanguage),
		normalizeAccessMode(request.AccessMode), request.RetentionDays, request.KnowledgeBaseIDs,
		normalizeOrigins(request.AllowedOrigins), branding, actor).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create chatbot: %w", err)
	}

	return s.GetChatbot(ctx, tenantID, id)
}

// GetChatbot returns a tenant-scoped mutable chatbot draft.
func (s *Store) GetChatbot(ctx context.Context, tenantID, chatbotID string) (*platform.Chatbot, error) {
	chatbotID = strings.TrimSpace(chatbotID)
	if _, err := uuid.Parse(chatbotID); err != nil {
		return nil, ErrNotFound
	}

	return scanChatbot(s.metadata.QueryRow(ctx, `
		SELECT chatbot_id::text, tenant_id::text, public_id, name, description, system_instruction,
		       answer_model, tone, default_locale, answer_language, access_mode, retention_days,
		       array_to_json(knowledge_base_ids),
		       array_to_json(allowed_origins), branding, published_version_id::text,
		       created_at, updated_at, deleted_at
		FROM chatbots
		WHERE tenant_id = $1 AND chatbot_id = $2 AND deleted_at IS NULL`, tenantID, chatbotID))
}

// ListChatbots lists tenant-scoped mutable drafts.
func (s *Store) ListChatbots(ctx context.Context, tenantID string, limit, offset int) (*platform.ChatbotList, error) {
	rows, err := s.metadata.Query(ctx, `
		SELECT chatbot_id::text, tenant_id::text, public_id, name, description, system_instruction,
		       answer_model, tone, default_locale, answer_language, access_mode, retention_days,
		       array_to_json(knowledge_base_ids),
		       array_to_json(allowed_origins), branding, published_version_id::text,
		       created_at, updated_at, deleted_at, count(*) OVER()::int
		FROM chatbots
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list chatbots: %w", err)
	}
	defer rows.Close()

	result := &platform.ChatbotList{Items: []platform.Chatbot{}, Limit: limit, Offset: offset}
	for rows.Next() {
		item, err := scanChatbotWithTotal(rows, &result.Total)
		if err != nil {
			return nil, err
		}

		result.Items = append(result.Items, *item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chatbots: %w", err)
	}

	return result, nil
}

// UpdateChatbot replaces mutable draft configuration without changing a published version.
func (s *Store) UpdateChatbot(ctx context.Context, tenantID, chatbotID string, request platform.UpsertChatbotRequest) (*platform.Chatbot, error) {
	branding, err := json.Marshal(normalizeBranding(request.Branding))
	if err != nil {
		return nil, fmt.Errorf("encode chatbot branding: %w", err)
	}

	command, err := s.metadata.Exec(ctx, `
		UPDATE chatbots SET
			name=$3, description=$4, system_instruction=$5, answer_model=$6, tone=$7,
			default_locale=$8, answer_language=$9, access_mode=$10, retention_days=$11,
			knowledge_base_ids=$12::uuid[], allowed_origins=$13::text[], branding=$14, updated_at=now()
		WHERE tenant_id=$1 AND chatbot_id=$2 AND deleted_at IS NULL`,
		tenantID, chatbotID, strings.TrimSpace(request.Name), strings.TrimSpace(request.Description),
		strings.TrimSpace(request.SystemInstruction), strings.TrimSpace(request.AnswerModel), normalizeTone(request.Tone),
		normalizeLocale(request.DefaultLocale), normalizeAnswerLanguage(request.AnswerLanguage),
		normalizeAccessMode(request.AccessMode), request.RetentionDays, request.KnowledgeBaseIDs,
		normalizeOrigins(request.AllowedOrigins), branding)
	if err != nil {
		return nil, fmt.Errorf("update chatbot: %w", err)
	}

	if command.RowsAffected() == 0 {
		return nil, ErrNotFound
	}

	return s.GetChatbot(ctx, tenantID, chatbotID)
}

// DeleteChatbot unpublishes and soft-deletes a tenant chatbot.
func (s *Store) DeleteChatbot(ctx context.Context, tenantID, chatbotID string) error {
	command, err := s.metadata.Exec(ctx, `
		UPDATE chatbots SET deleted_at=now(), published_version_id=NULL, updated_at=now()
		WHERE tenant_id=$1 AND chatbot_id=$2 AND deleted_at IS NULL`, tenantID, chatbotID)
	if err != nil {
		return fmt.Errorf("delete chatbot: %w", err)
	}

	if command.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// PublishChatbot snapshots the draft and atomically makes the immutable revision active.
func (s *Store) PublishChatbot(ctx context.Context, tenantID, chatbotID string) (*platform.ChatbotPublication, error) {
	tx, err := s.metadata.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin publish chatbot: %w", err)
	}
	defer rollback(ctx, tx)

	chatbot, err := scanChatbot(tx.QueryRow(ctx, `
		SELECT chatbot_id::text, tenant_id::text, public_id, name, description, system_instruction,
		       answer_model, tone, default_locale, answer_language, access_mode, retention_days,
		       array_to_json(knowledge_base_ids),
		       array_to_json(allowed_origins), branding, published_version_id::text,
		       created_at, updated_at, deleted_at
		FROM chatbots WHERE tenant_id=$1 AND chatbot_id=$2 AND deleted_at IS NULL
		FOR UPDATE`, tenantID, chatbotID))
	if err != nil {
		return nil, err
	}

	profileRevision, readyErr := verifyReadyKnowledgeBases(ctx, tx, tenantID, chatbot.KnowledgeBaseIDs)
	if readyErr != nil {
		return nil, readyErr
	}

	configuration, err := json.Marshal(chatbot)
	if err != nil {
		return nil, fmt.Errorf("encode chatbot publication: %w", err)
	}

	var publication platform.ChatbotPublication

	err = tx.QueryRow(ctx, `
		INSERT INTO chatbot_versions (tenant_id, chatbot_id, revision, configuration, rag_profile_revision)
		SELECT $1, $2, COALESCE(max(revision),0)+1, $3, $4
		FROM chatbot_versions WHERE tenant_id=$1 AND chatbot_id=$2
		RETURNING version_id::text, chatbot_id::text, revision, rag_profile_revision, published_at`,
		tenantID, chatbotID, configuration, profileRevision).Scan(
		&publication.VersionID, &publication.ChatbotID, &publication.Revision,
		&publication.RAGProfileRevision, &publication.PublishedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create chatbot publication: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE chatbot_versions SET unpublished_at=now()
		WHERE tenant_id=$1 AND chatbot_id=$2 AND version_id<>$3 AND unpublished_at IS NULL`,
		tenantID, chatbotID, publication.VersionID); err != nil {
		return nil, fmt.Errorf("supersede chatbot publication: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE chatbots SET published_version_id=$3, updated_at=now()
		WHERE tenant_id=$1 AND chatbot_id=$2`, tenantID, chatbotID, publication.VersionID); err != nil {
		return nil, fmt.Errorf("activate chatbot publication: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit chatbot publication: %w", err)
	}

	return &publication, nil
}

// UnpublishChatbot removes runtime admission while retaining immutable history for rollback.
func (s *Store) UnpublishChatbot(ctx context.Context, tenantID, chatbotID string) error {
	command, err := s.metadata.Exec(ctx, `
		WITH current AS (
			UPDATE chatbots SET published_version_id=NULL, updated_at=now()
			WHERE tenant_id=$1 AND chatbot_id=$2 AND deleted_at IS NULL
			RETURNING published_version_id
		)
		UPDATE chatbot_versions SET unpublished_at=now()
		WHERE tenant_id=$1 AND chatbot_id=$2 AND unpublished_at IS NULL`, tenantID, chatbotID)
	if err != nil {
		return fmt.Errorf("unpublish chatbot: %w", err)
	}

	if command.RowsAffected() == 0 {
		if _, getErr := s.GetChatbot(ctx, tenantID, chatbotID); getErr != nil {
			return getErr
		}
	}

	return nil
}

// GetPublishedChatbot resolves only the currently active immutable version.
func (s *Store) GetPublishedChatbot(ctx context.Context, publicID string) (*PublishedChatbot, error) {
	var (
		configuration []byte
		result        PublishedChatbot
	)

	err := s.metadata.QueryRow(ctx, `
		SELECT v.configuration, v.version_id::text, v.chatbot_id::text, v.revision,
		       v.rag_profile_revision, v.published_at, t.slug
		FROM chatbots b
		JOIN chatbot_versions v ON v.tenant_id=b.tenant_id AND v.version_id=b.published_version_id
		JOIN tenants t ON t.tenant_id=b.tenant_id
		WHERE b.public_id=$1 AND b.deleted_at IS NULL`, publicID).Scan(
		&configuration, &result.Version.VersionID, &result.Version.ChatbotID,
		&result.Version.Revision, &result.Version.RAGProfileRevision, &result.Version.PublishedAt,
		&result.TenantSlug,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get published chatbot: %w", err)
	}

	if err := json.Unmarshal(configuration, &result.Chatbot); err != nil {
		return nil, fmt.Errorf("decode published chatbot: %w", err)
	}

	return &result, nil
}

// CreateSession persists a runtime session unless retention is zero.
func (s *Store) CreateSession(ctx context.Context, chatbot PublishedChatbot) (string, error) {
	sessionID := uuid.NewString()

	var expiresAt *time.Time

	if chatbot.Chatbot.RetentionDays > 0 {
		expires := time.Now().UTC().Add(time.Duration(chatbot.Chatbot.RetentionDays) * 24 * time.Hour)
		expiresAt = &expires
	}

	_, err := s.metadata.Exec(ctx, `
		INSERT INTO chatbot_sessions (session_id, tenant_id, chatbot_id, chatbot_version_id, expires_at)
		VALUES ($1,$2,$3,$4,$5)`, sessionID, chatbot.Chatbot.TenantID, chatbot.Chatbot.ID,
		chatbot.Version.VersionID, expiresAt)
	if err != nil {
		return "", fmt.Errorf("create chatbot session: %w", err)
	}

	return sessionID, nil
}

// DeleteSession removes a session only when it belongs to the authorized immutable publication.
func (s *Store) DeleteSession(ctx context.Context, chatbot PublishedChatbot, sessionID string) error {
	redisDeleted, err := s.redis.Del(
		ctx, conversationSessionRedisKey(chatbot, sessionID), conversationRedisKey(chatbot, sessionID),
	).Result()
	if err != nil {
		return fmt.Errorf("delete ephemeral chatbot session: %w", err)
	}

	command, err := s.metadata.Exec(ctx, `
		DELETE FROM chatbot_sessions
		WHERE tenant_id=$1 AND chatbot_id=$2 AND chatbot_version_id=$3 AND session_id=$4`,
		chatbot.Chatbot.TenantID, chatbot.Chatbot.ID, chatbot.Version.VersionID, sessionID)
	if err != nil {
		return fmt.Errorf("delete chatbot session: %w", err)
	}

	if command.RowsAffected() == 0 && redisDeleted == 0 {
		return ErrNotFound
	}

	return nil
}

type rowScanner interface {
	Scan(...any) error
}

func scanChatbot(row rowScanner) (*platform.Chatbot, error) {
	return scanChatbotWithTotal(row, nil)
}

func scanChatbotWithTotal(row rowScanner, total *int) (*platform.Chatbot, error) {
	var (
		item                                         platform.Chatbot
		knowledgeBaseJSON, originsJSON, brandingJSON []byte
	)

	destinations := []any{
		&item.ID, &item.TenantID, &item.PublicID, &item.Name, &item.Description,
		&item.SystemInstruction, &item.AnswerModel, &item.Tone, &item.DefaultLocale,
		&item.AnswerLanguage, &item.AccessMode,
		&item.RetentionDays, &knowledgeBaseJSON, &originsJSON, &brandingJSON,
		&item.PublishedVersionID, &item.CreatedAt, &item.UpdatedAt, &item.DeletedAt,
	}
	if total != nil {
		destinations = append(destinations, total)
	}

	if err := row.Scan(destinations...); errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("scan chatbot: %w", err)
	}

	if err := json.Unmarshal(knowledgeBaseJSON, &item.KnowledgeBaseIDs); err != nil {
		return nil, fmt.Errorf("decode chatbot knowledge bases: %w", err)
	}

	if err := json.Unmarshal(originsJSON, &item.AllowedOrigins); err != nil {
		return nil, fmt.Errorf("decode chatbot allowed origins: %w", err)
	}

	if err := json.Unmarshal(brandingJSON, &item.Branding); err != nil {
		return nil, fmt.Errorf("decode chatbot branding: %w", err)
	}

	return &item, nil
}

func verifyReadyKnowledgeBases(ctx context.Context, tx pgx.Tx, tenantID string, ids []string) (string, error) {
	var (
		total, ready int
		profiles     []string
	)

	err := tx.QueryRow(ctx, `
		SELECT count(DISTINCT kb.knowledge_base_id)::int,
		       count(DISTINCT kb.knowledge_base_id) FILTER (
			   WHERE EXISTS (
				   SELECT 1 FROM knowledge_base_documents d
				   JOIN document_versions v ON v.tenant_id=d.tenant_id AND v.version_id=d.active_version_id
				   WHERE d.tenant_id=kb.tenant_id AND d.knowledge_base_id=kb.knowledge_base_id
				     AND d.deleted_at IS NULL AND v.status='ready'
			   )
		       )::int,
		       array_agg(DISTINCT kb.rag_profile_revision)
		FROM knowledge_bases kb
		WHERE kb.tenant_id=$1 AND kb.knowledge_base_id=ANY($2::uuid[]) AND kb.deleted_at IS NULL`,
		tenantID, ids).Scan(&total, &ready, &profiles)
	if err != nil {
		return "", fmt.Errorf("verify chatbot knowledge bases: %w", err)
	}

	if total != len(ids) || ready != len(ids) {
		return "", errors.New("all chatbot knowledge bases must exist and contain ready documents")
	}

	if len(profiles) != 1 {
		return "", ErrProfileMismatch
	}

	return profiles[0], nil
}

func normalizeAccessMode(value string) string {
	if strings.EqualFold(strings.TrimSpace(value), "public") {
		return "public"
	}

	return "private"
}

func normalizeTone(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "professional"
	}

	return value
}

func normalizeLocale(value string) string {
	if strings.EqualFold(strings.TrimSpace(value), "en-US") {
		return "en-US"
	}

	return "fa-IR"
}

func normalizeAnswerLanguage(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "fa":
		return "fa"
	case "en":
		return "en"
	default:
		return "query"
	}
}

func normalizeBranding(value map[string]any) map[string]any {
	if value == nil {
		return map[string]any{}
	}

	return value
}

func normalizeOrigins(value []string) []string {
	if value == nil {
		return []string{}
	}

	result := make([]string, 0, len(value))
	for _, origin := range value {
		result = append(result, strings.TrimRight(strings.TrimSpace(origin), "/"))
	}

	return result
}
