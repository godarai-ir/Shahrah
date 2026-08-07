package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/prompts"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ErrPromptNotFound is retained while older consumers migrate to prompts.ErrNotFound.
var ErrPromptNotFound = prompts.ErrNotFound

func (s *PromptRepository) CreatePrompt(ctx context.Context, tenantID string, tenantSlug string, req platform.CreatePromptRequest, actor string) (*platform.PromptVersion, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, mapPromptRepositoryError(err)
	}
	defer rollbackTxBestEffort(ctx, tx)

	slug := prompts.NormalizeSlug(req.Manifest.Name)
	description := strings.TrimSpace(req.Manifest.Description)

	var promptID string

	err = tx.QueryRow(ctx, `
		INSERT INTO prompts (tenant_id, name, slug, description, latest_version, created_by)
		VALUES ($1::uuid, $2, $3, $4, 0, $5)
		RETURNING id::text
	`, tenantID, req.Manifest.Name, slug, description, strings.TrimSpace(actor)).Scan(&promptID)
	if err != nil {
		return nil, mapPromptRepositoryError(err)
	}

	item, err := insertPromptVersion(ctx, tx, tenantID, tenantSlug, promptID, 1, req.Manifest, req.CommitMessage, actor)
	if err != nil {
		return nil, mapPromptRepositoryError(err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE prompts
		SET latest_version = 1, updated_at = NOW(), description = $3
		WHERE tenant_id = $1::uuid AND id = $2::uuid
	`, tenantID, promptID, description); err != nil {
		return nil, mapPromptRepositoryError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, mapPromptRepositoryError(err)
	}

	return item, nil
}

func (s *PromptRepository) CreatePromptVersion(ctx context.Context, tenantID string, tenantSlug string, promptID string, req platform.CreatePromptVersionRequest, actor string) (*platform.PromptVersion, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, mapPromptRepositoryError(err)
	}
	defer rollbackTxBestEffort(ctx, tx)

	var latest int

	queryErr := tx.QueryRow(ctx, `
		SELECT latest_version
		FROM prompts
		WHERE tenant_id = $1::uuid AND id = $2::uuid
		FOR UPDATE
	`, tenantID, promptID).Scan(&latest)
	if queryErr != nil {
		if errors.Is(queryErr, pgx.ErrNoRows) {
			return nil, prompts.ErrNotFound
		}

		return nil, mapPromptRepositoryError(queryErr)
	}

	nextVersion := latest + 1

	item, err := insertPromptVersion(ctx, tx, tenantID, tenantSlug, promptID, nextVersion, req.Manifest, req.CommitMessage, actor)
	if err != nil {
		return nil, mapPromptRepositoryError(err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE prompts
		SET name = $3, slug = $4, description = $5, latest_version = $6, updated_at = NOW()
		WHERE tenant_id = $1::uuid AND id = $2::uuid
	`, tenantID, promptID, req.Manifest.Name, prompts.NormalizeSlug(req.Manifest.Name), strings.TrimSpace(req.Manifest.Description), nextVersion); err != nil {
		return nil, mapPromptRepositoryError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, mapPromptRepositoryError(err)
	}

	return item, nil
}

func insertPromptVersion(ctx context.Context, tx pgx.Tx, tenantID string, tenantSlug string, promptID string, version int, manifest platform.PromptManifest, commitMessage string, actor string) (*platform.PromptVersion, error) {
	messages, err := json.Marshal(manifest.Messages)
	if err != nil {
		return nil, err
	}

	variables, err := marshalJSONObjectString(manifest.Variables)
	if err != nil {
		return nil, err
	}

	parameters, err := marshalJSONObject(manifest.Parameters)
	if err != nil {
		return nil, err
	}

	metadata, err := marshalJSONObject(manifest.Metadata)
	if err != nil {
		return nil, err
	}

	manifestJSON, err := marshalJSONObject(promptManifestToJSONObject(manifest))
	if err != nil {
		return nil, err
	}

	slug := prompts.NormalizeSlug(manifest.Name)
	fqn := prompts.BuildFQN(tenantSlug, slug, version)

	var (
		item                                                               platform.PromptVersion
		messagesRaw, variablesRaw, parametersRaw, metadataRaw, manifestRaw []byte
	)

	err = tx.QueryRow(ctx, `
		INSERT INTO prompt_versions (
			tenant_id, prompt_id, version, fqn, version_alias, name, description, messages,
			variables, model, parameters, metadata, manifest, commit_message, created_by
		)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8::jsonb, $9::jsonb, $10, $11::jsonb, $12::jsonb, $13::jsonb, $14, $15)
		RETURNING id::text, tenant_id::text, prompt_id::text, version, fqn, version_alias, name, description,
		          messages, variables, model, parameters, metadata, manifest, commit_message, created_by, created_at
	`, tenantID, promptID, version, fqn, strings.TrimSpace(manifest.VersionAlias), manifest.Name, strings.TrimSpace(manifest.Description), string(messages),
		string(variables), strings.TrimSpace(manifest.Model), string(parameters), string(metadata), string(manifestJSON), strings.TrimSpace(commitMessage), strings.TrimSpace(actor)).Scan(
		&item.ID, &item.TenantID, &item.PromptID, &item.Version, &item.FQN, &item.VersionAlias, &item.Name, &item.Description,
		&messagesRaw, &variablesRaw, &item.Model, &parametersRaw, &metadataRaw, &manifestRaw, &item.CommitMessage, &item.CreatedBy, &item.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	item.Messages = parsePromptMessages(messagesRaw)
	item.Variables = parseStringMap(variablesRaw)
	item.Parameters = parseJSONObject(parametersRaw)
	item.Metadata = parseJSONObject(metadataRaw)
	item.Manifest = parsePromptManifest(manifestRaw, item)

	return &item, nil
}

func (s *PromptRepository) ListPrompts(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.Prompt, int, error) {
	query = strings.TrimSpace(query)

	if limit <= 0 {
		limit = 50
	}

	if offset < 0 {
		offset = 0
	}

	args := []interface{}{tenantID, limit, offset}
	countArgs := []interface{}{tenantID}
	filter := ""
	countFilter := ""

	if query != "" {
		args = append(args, "%"+query+"%")
		countArgs = append(countArgs, "%"+query+"%")
		filter = " AND (p.name ILIKE $4 OR p.slug ILIKE $4 OR p.description ILIKE $4)"
		countFilter = " AND (p.name ILIKE $2 OR p.slug ILIKE $2 OR p.description ILIKE $2)"
	}

	var total int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM prompts p WHERE p.tenant_id = $1::uuid`+countFilter, countArgs...).Scan(&total); err != nil {
		return nil, 0, mapPromptRepositoryError(err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT p.id::text, p.tenant_id::text, p.name, p.slug, p.description, p.latest_version, p.created_by, p.created_at, p.updated_at,
		       pv.id::text, pv.prompt_id::text, COALESCE(pv.version, 0), COALESCE(pv.fqn, ''), COALESCE(pv.version_alias, ''),
		       COALESCE(pv.name, ''), COALESCE(pv.description, ''), COALESCE(pv.messages, '[]'::jsonb), COALESCE(pv.variables, '{}'::jsonb),
		       COALESCE(pv.model, ''), COALESCE(pv.parameters, '{}'::jsonb), COALESCE(pv.metadata, '{}'::jsonb), COALESCE(pv.manifest, '{}'::jsonb),
		       COALESCE(pv.commit_message, ''), COALESCE(pv.created_by, ''), pv.created_at
		FROM prompts p
		LEFT JOIN LATERAL (
			SELECT *
			FROM prompt_versions
			WHERE tenant_id = p.tenant_id AND prompt_id = p.id
			ORDER BY version DESC
			LIMIT 1
		) pv ON true
		WHERE p.tenant_id = $1::uuid`+filter+`
		ORDER BY p.updated_at DESC
		LIMIT $2 OFFSET $3
	`, args...)
	if err != nil {
		return nil, 0, mapPromptRepositoryError(err)
	}
	defer rows.Close()

	items := []platform.Prompt{}

	for rows.Next() {
		item, err := scanPromptWithLatest(rows)
		if err != nil {
			return nil, 0, mapPromptRepositoryError(err)
		}

		items = append(items, item)
	}

	return items, total, mapPromptRepositoryError(rows.Err())
}

func (s *PromptRepository) GetPrompt(ctx context.Context, tenantID string, promptID string) (*platform.Prompt, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT p.id::text, p.tenant_id::text, p.name, p.slug, p.description, p.latest_version, p.created_by, p.created_at, p.updated_at,
		       pv.id::text, pv.prompt_id::text, COALESCE(pv.version, 0), COALESCE(pv.fqn, ''), COALESCE(pv.version_alias, ''),
		       COALESCE(pv.name, ''), COALESCE(pv.description, ''), COALESCE(pv.messages, '[]'::jsonb), COALESCE(pv.variables, '{}'::jsonb),
		       COALESCE(pv.model, ''), COALESCE(pv.parameters, '{}'::jsonb), COALESCE(pv.metadata, '{}'::jsonb), COALESCE(pv.manifest, '{}'::jsonb),
		       COALESCE(pv.commit_message, ''), COALESCE(pv.created_by, ''), pv.created_at
		FROM prompts p
		LEFT JOIN LATERAL (
			SELECT *
			FROM prompt_versions
			WHERE tenant_id = p.tenant_id AND prompt_id = p.id
			ORDER BY version DESC
			LIMIT 1
		) pv ON true
		WHERE p.tenant_id = $1::uuid AND p.id = $2::uuid
	`, tenantID, promptID)
	if err != nil {
		return nil, mapPromptRepositoryError(err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, prompts.ErrNotFound
	}

	item, err := scanPromptWithLatest(rows)
	if err != nil {
		return nil, mapPromptRepositoryError(err)
	}

	if err := rows.Err(); err != nil {
		return nil, mapPromptRepositoryError(err)
	}

	return &item, nil
}

func (s *PromptRepository) DeletePrompt(ctx context.Context, tenantID string, promptID string) error {
	cmd, err := s.pool.Exec(ctx, `DELETE FROM prompts WHERE tenant_id = $1::uuid AND id = $2::uuid`, tenantID, promptID)
	if err != nil {
		return mapPromptRepositoryError(err)
	}

	if cmd.RowsAffected() == 0 {
		return prompts.ErrNotFound
	}

	return nil
}

func (s *PromptRepository) ListPromptVersions(ctx context.Context, tenantID string, promptID string, limit int, offset int) ([]platform.PromptVersion, int, error) {
	if limit <= 0 {
		limit = 50
	}

	if offset < 0 {
		offset = 0
	}

	var total int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM prompt_versions WHERE tenant_id = $1::uuid AND prompt_id = $2::uuid`, tenantID, promptID).Scan(&total); err != nil {
		return nil, 0, mapPromptRepositoryError(err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id::text, tenant_id::text, prompt_id::text, version, fqn, version_alias, name, description,
		       messages, variables, model, parameters, metadata, manifest, commit_message, created_by, created_at
		FROM prompt_versions
		WHERE tenant_id = $1::uuid AND prompt_id = $2::uuid
		ORDER BY version DESC
		LIMIT $3 OFFSET $4
	`, tenantID, promptID, limit, offset)
	if err != nil {
		return nil, 0, mapPromptRepositoryError(err)
	}
	defer rows.Close()

	items := []platform.PromptVersion{}

	for rows.Next() {
		item, err := scanPromptVersion(rows)
		if err != nil {
			return nil, 0, mapPromptRepositoryError(err)
		}

		items = append(items, item)
	}

	return items, total, mapPromptRepositoryError(rows.Err())
}

func (s *PromptRepository) GetPromptVersion(ctx context.Context, tenantID string, versionID string) (*platform.PromptVersion, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id::text, tenant_id::text, prompt_id::text, version, fqn, version_alias, name, description,
		       messages, variables, model, parameters, metadata, manifest, commit_message, created_by, created_at
		FROM prompt_versions
		WHERE tenant_id = $1::uuid AND id = $2::uuid
	`, tenantID, versionID)

	item, err := scanPromptVersion(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, prompts.ErrNotFound
		}

		return nil, mapPromptRepositoryError(err)
	}

	return &item, nil
}

func (s *PromptRepository) GetPromptVersionByFQN(ctx context.Context, tenantID string, fqn string) (*platform.PromptVersion, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id::text, tenant_id::text, prompt_id::text, version, fqn, version_alias, name, description,
		       messages, variables, model, parameters, metadata, manifest, commit_message, created_by, created_at
		FROM prompt_versions
		WHERE tenant_id = $1::uuid AND fqn = $2
	`, tenantID, strings.TrimSpace(fqn))

	item, err := scanPromptVersion(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, prompts.ErrNotFound
		}

		return nil, mapPromptRepositoryError(err)
	}

	return &item, nil
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanPromptVersion(row rowScanner) (platform.PromptVersion, error) {
	var (
		item                                                               platform.PromptVersion
		messagesRaw, variablesRaw, parametersRaw, metadataRaw, manifestRaw []byte
	)
	if err := row.Scan(
		&item.ID, &item.TenantID, &item.PromptID, &item.Version, &item.FQN, &item.VersionAlias, &item.Name, &item.Description,
		&messagesRaw, &variablesRaw, &item.Model, &parametersRaw, &metadataRaw, &manifestRaw, &item.CommitMessage, &item.CreatedBy, &item.CreatedAt,
	); err != nil {
		return platform.PromptVersion{}, err
	}

	item.Messages = parsePromptMessages(messagesRaw)
	item.Variables = parseStringMap(variablesRaw)
	item.Parameters = parseJSONObject(parametersRaw)
	item.Metadata = parseJSONObject(metadataRaw)
	item.Manifest = parsePromptManifest(manifestRaw, item)

	return item, nil
}

func scanPromptWithLatest(row rowScanner) (platform.Prompt, error) {
	var (
		item                                                               platform.Prompt
		latest                                                             platform.PromptVersion
		messagesRaw, variablesRaw, parametersRaw, metadataRaw, manifestRaw []byte
		latestCreatedAt                                                    interface{}
	)
	if err := row.Scan(
		&item.ID, &item.TenantID, &item.Name, &item.Slug, &item.Description, &item.LatestVersion, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt,
		&latest.ID, &latest.PromptID, &latest.Version, &latest.FQN, &latest.VersionAlias, &latest.Name, &latest.Description,
		&messagesRaw, &variablesRaw, &latest.Model, &parametersRaw, &metadataRaw, &manifestRaw, &latest.CommitMessage, &latest.CreatedBy, &latestCreatedAt,
	); err != nil {
		return platform.Prompt{}, err
	}

	if latest.ID != "" {
		latest.TenantID = item.TenantID
		latest.Messages = parsePromptMessages(messagesRaw)
		latest.Variables = parseStringMap(variablesRaw)
		latest.Parameters = parseJSONObject(parametersRaw)
		latest.Metadata = parseJSONObject(metadataRaw)
		latest.Manifest = parsePromptManifest(manifestRaw, latest)
		item.Latest = &latest
	}

	return item, nil
}

func marshalJSONObjectString(input map[string]string) ([]byte, error) {
	if len(input) == 0 {
		return []byte("{}"), nil
	}

	return json.Marshal(input)
}

func parseStringMap(raw []byte) map[string]string {
	if len(raw) == 0 {
		return map[string]string{}
	}

	var out map[string]string
	if err := json.Unmarshal(raw, &out); err != nil || out == nil {
		return map[string]string{}
	}

	return out
}

func parsePromptMessages(raw []byte) []contractsopenai.ChatMessage {
	if len(raw) == 0 {
		return []contractsopenai.ChatMessage{}
	}

	var out []contractsopenai.ChatMessage
	if err := json.Unmarshal(raw, &out); err != nil {
		return []contractsopenai.ChatMessage{}
	}

	return out
}

func promptManifestToJSONObject(manifest platform.PromptManifest) map[string]interface{} {
	raw := map[string]interface{}{
		"type":          manifest.Type,
		"name":          manifest.Name,
		"description":   manifest.Description,
		"messages":      manifest.Messages,
		"variables":     manifest.Variables,
		"model":         manifest.Model,
		"parameters":    manifest.Parameters,
		"metadata":      manifest.Metadata,
		"version_alias": manifest.VersionAlias,
	}
	if manifest.ModelSettings != nil {
		raw["model_settings"] = manifest.ModelSettings
	}

	if manifest.Logging != nil {
		raw["logging"] = manifest.Logging
	}

	if manifest.Cache != nil {
		raw["cache"] = manifest.Cache
	}

	if manifest.Guardrails != nil {
		raw["guardrails"] = manifest.Guardrails
	}

	if manifest.Structured != nil {
		raw["structured_output"] = manifest.Structured
	}

	return raw
}

func parsePromptManifest(raw []byte, fallback platform.PromptVersion) platform.PromptManifest {
	if len(raw) > 0 {
		var manifest platform.PromptManifest
		if err := json.Unmarshal(raw, &manifest); err == nil {
			if strings.TrimSpace(manifest.Type) == "" {
				manifest.Type = prompts.TypeChatPrompt
			}

			if strings.TrimSpace(manifest.Name) == "" {
				manifest.Name = fallback.Name
			}

			if strings.TrimSpace(manifest.Description) == "" {
				manifest.Description = fallback.Description
			}

			if len(manifest.Messages) == 0 {
				manifest.Messages = fallback.Messages
			}

			if manifest.Variables == nil {
				manifest.Variables = fallback.Variables
			}

			if strings.TrimSpace(manifest.Model) == "" {
				manifest.Model = fallback.Model
			}

			if manifest.Parameters == nil {
				manifest.Parameters = fallback.Parameters
			}

			if manifest.Metadata == nil {
				manifest.Metadata = fallback.Metadata
			}

			if strings.TrimSpace(manifest.VersionAlias) == "" {
				manifest.VersionAlias = fallback.VersionAlias
			}

			return manifest
		}
	}

	return platform.PromptManifest{
		Type:         prompts.TypeChatPrompt,
		Name:         fallback.Name,
		Description:  fallback.Description,
		Messages:     fallback.Messages,
		Variables:    fallback.Variables,
		Model:        fallback.Model,
		Parameters:   fallback.Parameters,
		Metadata:     fallback.Metadata,
		VersionAlias: fallback.VersionAlias,
	}
}

func mapPromptRepositoryError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return prompts.ErrNotFound
	}

	var postgresErr *pgconn.PgError
	if errors.As(err, &postgresErr) {
		switch postgresErr.Code {
		case "23505":
			return fmt.Errorf("%w: %v", prompts.ErrConflict, err)
		case "23514", "23502", "22P02":
			return fmt.Errorf("%w: %v", prompts.ErrInvalidPersistenceData, err)
		case "42P01", "42703":
			return fmt.Errorf("%w: %v", prompts.ErrStorageUnavailable, err)
		}
	}

	return fmt.Errorf("prompt repository: %w", err)
}
