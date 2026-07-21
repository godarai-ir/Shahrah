package storage

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/platformcredentials"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

func (s *Store) CreatePlatformProviderConnection(ctx context.Context, request platform.CreatePlatformProviderConnectionRequest, actor platformcredentials.Actor, idempotencyKey string) (*platform.PlatformProviderConnection, error) {
	configJSON, err := json.Marshal(request.Config)
	if err != nil {
		return nil, fmt.Errorf("encode platform provider connection config: %w", err)
	}

	requestHash, err := platformCredentialRequestHash(request)
	if err != nil {
		return nil, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer s.rollbackPlatformCredentialTx(ctx, tx)

	if replay, found, replayErr := findPlatformCredentialIdempotencyReplay(ctx, tx, "create_connection", idempotencyKey); replayErr != nil {
		return nil, replayErr
	} else if found {
		if !platformCredentialRequestHashesEqual(replay.requestHash, requestHash) {
			return nil, platformcredentials.ErrInvalidRequest
		}

		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			return nil, fmt.Errorf("rollback idempotency replay transaction: %w", rollbackErr)
		}

		return s.GetPlatformProviderConnection(ctx, replay.connectionID)
	}

	connectionID := uuid.NewString()

	_, err = tx.Exec(ctx, `
		INSERT INTO platform_provider_connections
		(connection_id, provider_slug, name, environment, auth_mode, schema_version, config_json, region, priority, weight, is_primary, created_by, updated_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,NULLIF($8,''),$9,$10,$11,$12,$12)`,
		connectionID, request.ProviderSlug, strings.TrimSpace(request.Name), request.Environment, request.AuthMode,
		request.SchemaVersion, configJSON, request.Region, request.Priority, request.Weight, request.IsPrimary, actor.ID)
	if err != nil {
		return nil, err
	}

	for _, scope := range request.Scopes {
		if err := insertPlatformConnectionScope(ctx, tx, connectionID, scope); err != nil {
			return nil, err
		}
	}

	if err := insertPlatformCredentialEvent(ctx, tx, actor, connectionID, "", "create_connection", "success", "", idempotencyKey, requestHash, map[string]interface{}{"provider_slug": request.ProviderSlug, "environment": request.Environment}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return s.GetPlatformProviderConnection(ctx, connectionID)
}

func (s *Store) UpdatePlatformProviderConnection(ctx context.Context, id string, request platform.UpdatePlatformProviderConnectionRequest, actor platformcredentials.Actor) (*platform.PlatformProviderConnection, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer s.rollbackPlatformCredentialTx(ctx, tx)

	configJSON, err := json.Marshal(request.Config)
	if err != nil {
		return nil, fmt.Errorf("encode platform provider connection config: %w", err)
	}

	command, err := tx.Exec(ctx, `
		UPDATE platform_provider_connections SET
		name = CASE WHEN $3::boolean THEN $4 ELSE name END,
		config_json = CASE WHEN $5::boolean THEN $6::jsonb ELSE config_json END,
		region = CASE WHEN $7::boolean THEN NULLIF($8,'') ELSE region END,
		priority = CASE WHEN $9::boolean THEN $10 ELSE priority END,
		weight = CASE WHEN $11::boolean THEN $12 ELSE weight END,
		is_primary = CASE WHEN $13::boolean THEN $14 ELSE is_primary END,
		revision = revision + 1, updated_by = $15, updated_at = now()
		WHERE connection_id=$1 AND revision=$2`, id, request.Revision,
		request.Name != nil, platformCredentialStringValue(request.Name), request.Config != nil, configJSON,
		request.Region != nil, platformCredentialStringValue(request.Region), request.Priority != nil, intValue(request.Priority),
		request.Weight != nil, floatValue(request.Weight), request.IsPrimary != nil, boolValue(request.IsPrimary), actor.ID)
	if err != nil {
		return nil, err
	}

	if command.RowsAffected() == 0 {
		return nil, platformcredentials.ErrRevisionConflict
	}

	if request.Scopes != nil {
		if _, err = tx.Exec(ctx, `DELETE FROM platform_provider_connection_scopes WHERE connection_id=$1`, id); err != nil {
			return nil, err
		}

		for _, scope := range *request.Scopes {
			if err := insertPlatformConnectionScope(ctx, tx, id, scope); err != nil {
				return nil, err
			}
		}
	}

	if err := insertPlatformCredentialEvent(ctx, tx, actor, id, "", "update_connection", "success", "", "", "", nil); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return s.GetPlatformProviderConnection(ctx, id)
}

func (s *Store) ListPlatformProviderConnections(ctx context.Context, provider, environment string) ([]platform.PlatformProviderConnection, error) {
	rows, err := s.pool.Query(ctx, platformConnectionSelect+`
		WHERE ($1='' OR c.provider_slug=$1) AND ($2='' OR c.environment=$2)
		ORDER BY c.provider_slug, c.environment, c.is_primary DESC, c.priority DESC, c.name`, provider, environment)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]platform.PlatformProviderConnection, 0)

	for rows.Next() {
		item, err := scanPlatformConnection(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, *item)
	}

	return items, rows.Err()
}

func (s *Store) GetPlatformProviderConnection(ctx context.Context, id string) (*platform.PlatformProviderConnection, error) {
	connection, err := scanPlatformConnection(s.pool.QueryRow(ctx, platformConnectionSelect+` WHERE c.connection_id=$1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, platformcredentials.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	versionRows, err := s.pool.Query(ctx, `SELECT credential_version_id,version,secret_backend,display_hint,status,valid_from,expires_at,retire_at,created_by,activated_by,created_at,updated_at FROM platform_provider_credential_versions WHERE connection_id=$1 ORDER BY version DESC`, id)
	if err != nil {
		return nil, err
	}
	defer versionRows.Close()

	for versionRows.Next() {
		var (
			version                platform.PlatformProviderCredentialVersion
			createdBy, activatedBy *string
		)
		if scanErr := versionRows.Scan(&version.ID, &version.Version, &version.Backend, &version.DisplayHint, &version.Status, &version.ValidFrom, &version.ExpiresAt, &version.RetireAt, &createdBy, &activatedBy, &version.CreatedAt, &version.UpdatedAt); scanErr != nil {
			return nil, scanErr
		}

		version.Configured = true
		if createdBy != nil {
			version.CreatedBy = *createdBy
		}

		if activatedBy != nil {
			version.ActivatedBy = *activatedBy
		}

		connection.Versions = append(connection.Versions, version)
	}

	scopeRows, err := s.pool.Query(ctx, `SELECT scope_id,model_pattern,deployment_pattern,capabilities,regions,enabled FROM platform_provider_connection_scopes WHERE connection_id=$1 ORDER BY created_at`, id)
	if err != nil {
		return nil, err
	}
	defer scopeRows.Close()

	for scopeRows.Next() {
		var (
			scope                           platform.PlatformProviderConnectionScope
			modelPattern, deploymentPattern *string
		)
		if err := scopeRows.Scan(&scope.ID, &modelPattern, &deploymentPattern, &scope.Capabilities, &scope.Regions, &scope.Enabled); err != nil {
			return nil, err
		}

		if modelPattern != nil {
			scope.ModelPattern = *modelPattern
		}

		if deploymentPattern != nil {
			scope.DeploymentPattern = *deploymentPattern
		}

		connection.Scopes = append(connection.Scopes, scope)
	}

	return connection, nil
}

func (s *Store) CreatePlatformProviderCredentialVersion(
	ctx context.Context,
	connectionID string,
	secrets map[string]string,
	expiresAt *time.Time,
	actor platformcredentials.Actor,
	idempotencyKey string,
	encrypt func(platformcredentials.Binding, map[string]string) (platformcredentials.EncryptedCredential, error),
) (*platform.PlatformProviderCredentialVersion, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer s.rollbackPlatformCredentialTx(ctx, tx)

	var (
		provider, environment, authMode string
		schemaVersion                   int
	)
	if queryErr := tx.QueryRow(ctx, `SELECT provider_slug,environment,auth_mode,schema_version FROM platform_provider_connections WHERE connection_id=$1 FOR UPDATE`, connectionID).Scan(&provider, &environment, &authMode, &schemaVersion); errors.Is(queryErr, pgx.ErrNoRows) {
		return nil, platformcredentials.ErrNotFound
	} else if queryErr != nil {
		return nil, queryErr
	}

	var version int64
	if queryErr := tx.QueryRow(ctx, `SELECT COALESCE(MAX(version),0)+1 FROM platform_provider_credential_versions WHERE connection_id=$1`, connectionID).Scan(&version); queryErr != nil {
		return nil, queryErr
	}

	versionID := uuid.NewString()

	record, err := encrypt(platformcredentials.Binding{ProviderSlug: provider, Environment: environment, ConnectionID: connectionID, CredentialVersionID: versionID, CredentialVersion: version, AuthMode: authMode, SchemaVersion: schemaVersion}, secrets)
	if err != nil {
		return nil, err
	}

	requestHash, err := platformCredentialRequestHash(struct {
		Fingerprint string     `json:"fingerprint"`
		ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	}{Fingerprint: record.Fingerprint, ExpiresAt: expiresAt})
	if err != nil {
		return nil, err
	}

	if replay, found, replayErr := findPlatformCredentialIdempotencyReplay(ctx, tx, "create_credential_version", idempotencyKey); replayErr != nil {
		return nil, replayErr
	} else if found {
		if replay.connectionID != connectionID || replay.credentialVersionID == "" || !platformCredentialRequestHashesEqual(replay.requestHash, requestHash) {
			return nil, platformcredentials.ErrInvalidRequest
		}

		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			return nil, fmt.Errorf("rollback idempotency replay transaction: %w", rollbackErr)
		}

		connection, getErr := s.GetPlatformProviderConnection(ctx, connectionID)
		if getErr != nil {
			return nil, getErr
		}

		for _, item := range connection.Versions {
			if item.ID == replay.credentialVersionID {
				return &item, nil
			}
		}

		return nil, platformcredentials.ErrNotFound
	}

	_, err = tx.Exec(ctx, `INSERT INTO platform_provider_credential_versions
		(credential_version_id,connection_id,version,ciphertext,data_nonce,wrapped_dek,wrap_nonce,kek_version,fingerprint,display_hint,status,expires_at,created_by,idempotency_key,request_hash)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,'draft',$11,$12,NULLIF($13,''),$14)`, versionID, connectionID, version, record.Ciphertext, record.DataNonce, record.WrappedDEK, record.WrapNonce, record.KEKVersion, record.Fingerprint, record.DisplayHint, expiresAt, actor.ID, idempotencyKey, requestHash)
	if err != nil {
		return nil, err
	}

	if _, err = tx.Exec(ctx, `UPDATE platform_provider_connections SET revision=revision+1,updated_by=$2,updated_at=now() WHERE connection_id=$1`, connectionID, actor.ID); err != nil {
		return nil, err
	}

	if eventErr := insertPlatformCredentialEvent(ctx, tx, actor, connectionID, versionID, "create_credential_version", "success", "", idempotencyKey, requestHash, map[string]interface{}{"version": version}); eventErr != nil {
		return nil, eventErr
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return nil, commitErr
	}

	connection, err := s.GetPlatformProviderConnection(ctx, connectionID)
	if err != nil {
		return nil, err
	}

	for _, item := range connection.Versions {
		if item.ID == versionID {
			return &item, nil
		}
	}

	return nil, platformcredentials.ErrNotFound
}

func (s *Store) GetPlatformProviderCredential(ctx context.Context, connectionID string, version int64) (platformcredentials.EncryptedCredential, error) {
	var (
		record     platformcredentials.EncryptedCredential
		configJSON []byte
	)

	err := s.pool.QueryRow(ctx, `SELECT c.provider_slug,c.environment,c.connection_id,v.credential_version_id,v.version,c.auth_mode,c.schema_version,c.config_json,v.ciphertext,v.data_nonce,v.wrapped_dek,v.wrap_nonce,v.kek_version,v.fingerprint,v.display_hint,v.expires_at
		FROM platform_provider_connections c JOIN platform_provider_credential_versions v ON v.connection_id=c.connection_id WHERE c.connection_id=$1 AND v.version=$2`, connectionID, version).Scan(
		&record.ProviderSlug, &record.Environment, &record.ConnectionID, &record.CredentialVersionID,
		&record.CredentialVersion, &record.AuthMode, &record.SchemaVersion, &configJSON,
		&record.Ciphertext, &record.DataNonce, &record.WrappedDEK, &record.WrapNonce,
		&record.KEKVersion, &record.Fingerprint, &record.DisplayHint, &record.ExpiresAt,
	)
	applyPlatformCredentialRuntimeConfig(&record, configJSON)

	if errors.Is(err, pgx.ErrNoRows) {
		return record, platformcredentials.ErrNotFound
	}

	return record, err
}

func (s *Store) SetPlatformProviderCredentialTestResult(ctx context.Context, connectionID string, version int64, success bool, category, summary string, actor platformcredentials.Actor) (*platform.PlatformProviderCredentialTestResult, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer s.rollbackPlatformCredentialTx(ctx, tx)

	status := "failed"

	credentialStatus := "draft"
	if success {
		status, credentialStatus = "success", "validated"
	}

	command, err := tx.Exec(ctx, `UPDATE platform_provider_credential_versions
		SET status=CASE WHEN status='active' THEN 'active' ELSE $3 END,updated_at=now()
		WHERE connection_id=$1 AND version=$2 AND status IN ('draft','validated','active')`, connectionID, version, credentialStatus)
	if err != nil {
		return nil, err
	}

	if command.RowsAffected() == 0 {
		return nil, platformcredentials.ErrNotFound
	}

	_, err = tx.Exec(ctx, `UPDATE platform_provider_connections SET
		status=CASE
			WHEN active_credential_version_id=(SELECT credential_version_id FROM platform_provider_credential_versions WHERE connection_id=$1 AND version=$7)
				THEN CASE WHEN $2 THEN 'active' ELSE 'degraded' END
			WHEN $2 THEN 'draft'
			ELSE 'degraded'
		END,
		last_test_status=$3,last_test_category=$4,last_test_summary=$5,last_tested_at=now(),
		last_success_at=CASE WHEN $2 THEN now() ELSE last_success_at END,
		revision=revision+1,updated_by=$6,updated_at=now()
		WHERE connection_id=$1`, connectionID, success, status, category, truncate(summary, 255), actor.ID, version)
	if err != nil {
		return nil, err
	}

	if err := insertPlatformCredentialEvent(ctx, tx, actor, connectionID, "", "test_credential", map[bool]string{true: "success", false: "failure"}[success], "", "", "", map[string]interface{}{"version": version, "category": category}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &platform.PlatformProviderCredentialTestResult{Success: success, Category: category, Summary: truncate(summary, 255), TestedAt: time.Now().UTC(), Version: version, Validated: success}, nil
}

func (s *Store) ActivatePlatformProviderCredential(ctx context.Context, connectionID string, version, revision int64, actor platformcredentials.Actor, idempotencyKey, reason string) (*platform.PlatformProviderConnection, error) {
	return s.activatePlatformProviderCredential(ctx, connectionID, version, revision, actor, idempotencyKey, reason, "activate")
}

func (s *Store) RollbackPlatformProviderCredential(ctx context.Context, connectionID string, version, revision int64, actor platformcredentials.Actor, reason string) (*platform.PlatformProviderConnection, error) {
	return s.activatePlatformProviderCredential(ctx, connectionID, version, revision, actor, "", reason, "rollback")
}

func (s *Store) activatePlatformProviderCredential(ctx context.Context, connectionID string, version, revision int64, actor platformcredentials.Actor, idempotencyKey, reason, action string) (*platform.PlatformProviderConnection, error) {
	requestHash := ""

	if idempotencyKey != "" {
		var hashErr error

		requestHash, hashErr = platformCredentialRequestHash(struct {
			ConnectionID string `json:"connection_id"`
			Version      int64  `json:"version"`
			Revision     int64  `json:"revision"`
			Reason       string `json:"reason"`
		}{connectionID, version, revision, reason})
		if hashErr != nil {
			return nil, hashErr
		}
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer s.rollbackPlatformCredentialTx(ctx, tx)

	if replay, found, replayErr := findPlatformCredentialIdempotencyReplay(ctx, tx, action, idempotencyKey); replayErr != nil {
		return nil, replayErr
	} else if found {
		if replay.connectionID != connectionID || !platformCredentialRequestHashesEqual(replay.requestHash, requestHash) {
			return nil, platformcredentials.ErrInvalidRequest
		}

		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			return nil, fmt.Errorf("rollback idempotency replay transaction: %w", rollbackErr)
		}

		return s.GetPlatformProviderConnection(ctx, connectionID)
	}

	var (
		currentRevision       int64
		provider, environment string
	)
	if queryErr := tx.QueryRow(ctx, `SELECT revision,provider_slug,environment FROM platform_provider_connections WHERE connection_id=$1 FOR UPDATE`, connectionID).Scan(&currentRevision, &provider, &environment); errors.Is(queryErr, pgx.ErrNoRows) {
		return nil, platformcredentials.ErrNotFound
	} else if queryErr != nil {
		return nil, queryErr
	}

	if currentRevision != revision {
		return nil, platformcredentials.ErrRevisionConflict
	}

	var (
		versionID, status string
		expiresAt         *time.Time
	)
	if queryErr := tx.QueryRow(ctx, `SELECT credential_version_id,status,expires_at FROM platform_provider_credential_versions WHERE connection_id=$1 AND version=$2 FOR UPDATE`, connectionID, version).Scan(&versionID, &status, &expiresAt); errors.Is(queryErr, pgx.ErrNoRows) {
		return nil, platformcredentials.ErrNotFound
	} else if queryErr != nil {
		return nil, queryErr
	}

	if status != "validated" && status != "retiring" && status != "active" {
		return nil, platformcredentials.ErrInvalidRequest
	}

	if expiresAt != nil && !expiresAt.After(time.Now()) {
		return nil, platformcredentials.ErrInvalidRequest
	}

	if _, err = tx.Exec(ctx, `UPDATE platform_provider_credential_versions SET status='retiring',retire_at=COALESCE(retire_at,now()),updated_at=now() WHERE connection_id=$1 AND status='active' AND credential_version_id<>$2`, connectionID, versionID); err != nil {
		return nil, err
	}

	if _, err = tx.Exec(ctx, `UPDATE platform_provider_credential_versions SET status='active',valid_from=COALESCE(valid_from,now()),retire_at=NULL,activated_by=$3,updated_at=now() WHERE connection_id=$1 AND version=$2`, connectionID, version, actor.ID); err != nil {
		return nil, err
	}

	if _, err = tx.Exec(ctx, `UPDATE platform_provider_connections SET is_primary=false,revision=revision+1,updated_at=now() WHERE provider_slug=$1 AND environment=$2 AND connection_id<>$3 AND status='active' AND is_primary=true`, provider, environment, connectionID); err != nil {
		return nil, err
	}

	if _, err = tx.Exec(ctx, `UPDATE platform_provider_connections SET status='active',is_primary=true,active_credential_version_id=$2,last_rotation_at=now(),revision=revision+1,updated_by=$3,updated_at=now() WHERE connection_id=$1`, connectionID, versionID, actor.ID); err != nil {
		return nil, err
	}

	if err := insertPlatformCredentialEvent(ctx, tx, actor, connectionID, versionID, action, "success", reason, idempotencyKey, requestHash, map[string]interface{}{"version": version}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return s.GetPlatformProviderConnection(ctx, connectionID)
}

func (s *Store) DisablePlatformProviderConnection(ctx context.Context, connectionID string, revision int64, actor platformcredentials.Actor, reason string) (*platform.PlatformProviderConnection, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer s.rollbackPlatformCredentialTx(ctx, tx)

	command, err := tx.Exec(ctx, `UPDATE platform_provider_connections SET status='disabled',is_primary=false,active_credential_version_id=NULL,revision=revision+1,updated_by=$3,updated_at=now() WHERE connection_id=$1 AND revision=$2`, connectionID, revision, actor.ID)
	if err != nil {
		return nil, err
	}

	if command.RowsAffected() == 0 {
		return nil, platformcredentials.ErrRevisionConflict
	}

	if _, err = tx.Exec(ctx, `UPDATE platform_provider_credential_versions SET status='retiring',retire_at=COALESCE(retire_at,now()),updated_at=now() WHERE connection_id=$1 AND status='active'`, connectionID); err != nil {
		return nil, err
	}

	if eventErr := insertPlatformCredentialEvent(ctx, tx, actor, connectionID, "", "disable", "success", reason, "", "", nil); eventErr != nil {
		return nil, eventErr
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return nil, commitErr
	}

	return s.GetPlatformProviderConnection(ctx, connectionID)
}

func (s *Store) RevokePlatformProviderCredential(ctx context.Context, connectionID string, version, revision int64, actor platformcredentials.Actor, reason string) (*platform.PlatformProviderConnection, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer s.rollbackPlatformCredentialTx(ctx, tx)

	var (
		versionID string
		isActive  bool
	)

	err = tx.QueryRow(ctx, `
		SELECT v.credential_version_id,
		       COALESCE(c.active_credential_version_id=v.credential_version_id,false)
		FROM platform_provider_connections c
		JOIN platform_provider_credential_versions v ON v.connection_id=c.connection_id
		WHERE c.connection_id=$1 AND c.revision=$2 AND v.version=$3
		FOR UPDATE OF c,v`, connectionID, revision, version).Scan(&versionID, &isActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, platformcredentials.ErrRevisionConflict
	} else if err != nil {
		return nil, err
	}

	if _, err = tx.Exec(ctx, `UPDATE platform_provider_credential_versions SET status='revoked',revoked_by=$3,revoked_at=now(),updated_at=now() WHERE connection_id=$1 AND version=$2`, connectionID, version, actor.ID); err != nil {
		return nil, err
	}

	if isActive {
		if _, err = tx.Exec(ctx, `UPDATE platform_provider_connections SET status='disabled',is_primary=false,active_credential_version_id=NULL,revision=revision+1,updated_by=$2,updated_at=now() WHERE connection_id=$1`, connectionID, actor.ID); err != nil {
			return nil, err
		}
	} else {
		_, err = tx.Exec(ctx, `UPDATE platform_provider_connections SET revision=revision+1,updated_by=$2,updated_at=now() WHERE connection_id=$1`, connectionID, actor.ID)
		if err != nil {
			return nil, err
		}
	}

	if eventErr := insertPlatformCredentialEvent(ctx, tx, actor, connectionID, versionID, "revoke", "success", reason, "", "", map[string]interface{}{"version": version}); eventErr != nil {
		return nil, eventErr
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return nil, commitErr
	}

	return s.GetPlatformProviderConnection(ctx, connectionID)
}

func (s *Store) LoadActivePlatformCredentials(ctx context.Context) ([]platformcredentials.EncryptedCredential, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT c.provider_slug,c.environment,c.connection_id,v.credential_version_id,v.version,
		       c.auth_mode,c.schema_version,c.config_json,v.ciphertext,v.data_nonce,v.wrapped_dek,
		       v.wrap_nonce,v.kek_version,v.fingerprint,v.display_hint,v.expires_at
		FROM platform_provider_connections c
		JOIN platform_provider_credential_versions v
		  ON v.credential_version_id=c.active_credential_version_id AND v.connection_id=c.connection_id
		WHERE c.status='active' AND c.is_primary=true AND v.status='active'
		  AND (v.expires_at IS NULL OR v.expires_at>now())`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]platformcredentials.EncryptedCredential, 0)

	for rows.Next() {
		var (
			item       platformcredentials.EncryptedCredential
			configJSON []byte
		)
		if scanErr := rows.Scan(
			&item.ProviderSlug, &item.Environment, &item.ConnectionID, &item.CredentialVersionID,
			&item.CredentialVersion, &item.AuthMode, &item.SchemaVersion, &configJSON,
			&item.Ciphertext, &item.DataNonce, &item.WrappedDEK, &item.WrapNonce,
			&item.KEKVersion, &item.Fingerprint, &item.DisplayHint, &item.ExpiresAt,
		); scanErr != nil {
			return nil, scanErr
		}

		applyPlatformCredentialRuntimeConfig(&item, configJSON)
		items = append(items, item)
	}

	return items, rows.Err()
}

func applyPlatformCredentialRuntimeConfig(record *platformcredentials.EncryptedCredential, encoded []byte) {
	var config map[string]interface{}
	if json.Unmarshal(encoded, &config) != nil {
		return
	}

	record.BaseURL = firstStringConfig(config, "base_url", "endpoint", "serving_endpoint_url")
	record.APIVersion = stringConfig(config, "api_version")
	record.Organization = stringConfig(config, "organization")
	record.Project = stringConfig(config, "project")
	record.HTTPReferer = stringConfig(config, "http_referer")
	record.AppTitle = stringConfig(config, "app_title")
	record.ClientName = stringConfig(config, "client_name")
}
func firstStringConfig(config map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value := stringConfig(config, key); value != "" {
			return value
		}
	}

	return ""
}
func stringConfig(config map[string]interface{}, key string) string {
	value, ok := config[key].(string)
	if !ok {
		return ""
	}

	return strings.TrimSpace(value)
}

const platformConnectionSelect = `SELECT
	c.connection_id,c.provider_slug,c.name,c.environment,c.auth_mode,c.schema_version,c.config_json,
	c.status,COALESCE(c.region,''),c.priority,c.weight,c.is_primary,
	COALESCE(c.active_credential_version_id::text,''),c.last_test_status,c.last_test_category,
	c.last_test_summary,c.last_tested_at,c.last_success_at,c.last_rotation_at,c.revision,
	c.created_by,c.updated_by,c.created_at,c.updated_at,
	v.credential_version_id,v.version,v.secret_backend,v.display_hint,v.status,v.valid_from,
	v.expires_at,v.retire_at,v.created_by,v.activated_by,v.created_at,v.updated_at
	FROM platform_provider_connections c
	LEFT JOIN platform_provider_credential_versions v
	  ON v.credential_version_id=c.active_credential_version_id`

type platformCredentialRowScanner interface{ Scan(...interface{}) error }

func scanPlatformConnection(row platformCredentialRowScanner) (*platform.PlatformProviderConnection, error) {
	var (
		item                                                                    platform.PlatformProviderConnection
		configJSON                                                              []byte
		lastTestStatus, lastTestCategory, lastTestSummary, createdBy, updatedBy *string
		activeID                                                                *string
		activeVersion                                                           *int64
		backend, hint, status                                                   *string
		validFrom, expiresAt, retireAt, versionCreatedAt, versionUpdatedAt      *time.Time
		versionCreatedBy, activatedBy                                           *string
	)

	err := row.Scan(
		&item.ID, &item.ProviderSlug, &item.Name, &item.Environment, &item.AuthMode,
		&item.SchemaVersion, &configJSON, &item.Status, &item.Region, &item.Priority,
		&item.Weight, &item.IsPrimary, &item.ActiveCredentialVersionID, &lastTestStatus,
		&lastTestCategory, &lastTestSummary, &item.LastTestedAt, &item.LastSuccessAt,
		&item.LastRotationAt, &item.Revision, &createdBy, &updatedBy, &item.CreatedAt,
		&item.UpdatedAt, &activeID, &activeVersion, &backend, &hint, &status, &validFrom,
		&expiresAt, &retireAt, &versionCreatedBy, &activatedBy, &versionCreatedAt,
		&versionUpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(configJSON, &item.Config); err != nil {
		return nil, fmt.Errorf("decode platform provider connection config: %w", err)
	}

	if lastTestStatus != nil {
		item.LastTestStatus = *lastTestStatus
	}

	if lastTestCategory != nil {
		item.LastTestCategory = *lastTestCategory
	}

	if lastTestSummary != nil {
		item.LastTestSummary = *lastTestSummary
	}

	if createdBy != nil {
		item.CreatedBy = *createdBy
	}

	if updatedBy != nil {
		item.UpdatedBy = *updatedBy
	}

	if activeID != nil {
		v := &platform.PlatformProviderCredentialVersion{ID: *activeID, Version: *activeVersion, Backend: *backend, Configured: true, DisplayHint: *hint, Status: *status, ValidFrom: validFrom, ExpiresAt: expiresAt, RetireAt: retireAt, CreatedAt: *versionCreatedAt, UpdatedAt: *versionUpdatedAt}
		if versionCreatedBy != nil {
			v.CreatedBy = *versionCreatedBy
		}

		if activatedBy != nil {
			v.ActivatedBy = *activatedBy
		}

		item.ActiveVersion = v
	}

	return &item, nil
}

func insertPlatformConnectionScope(ctx context.Context, tx pgx.Tx, connectionID string, scope platform.PlatformProviderConnectionScope) error {
	if scope.ModelPattern == "" && scope.DeploymentPattern == "" && len(scope.Capabilities) == 0 && len(scope.Regions) == 0 {
		return platformcredentials.ErrInvalidRequest
	}

	_, err := tx.Exec(ctx, `
		INSERT INTO platform_provider_connection_scopes(
			scope_id,connection_id,model_pattern,deployment_pattern,capabilities,regions,enabled
		) VALUES($1,$2,NULLIF($3,''),NULLIF($4,''),$5,$6,$7)`,
		uuid.NewString(), connectionID, scope.ModelPattern, scope.DeploymentPattern,
		scope.Capabilities, scope.Regions, scope.Enabled,
	)

	return err
}
func insertPlatformCredentialEvent(ctx context.Context, tx pgx.Tx, actor platformcredentials.Actor, connectionID, versionID, action, result, reason, idempotencyKey, requestHash string, metadata map[string]interface{}) error {
	encoded, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("encode platform credential event metadata: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO platform_provider_credential_events(
			event_id,connection_id,credential_version_id,actor_id,actor_role,service_account_id,
			request_id,action,result,reason,metadata,source_ip,idempotency_key,request_hash
		) VALUES(
			$1,$2,NULLIF($3,'')::uuid,NULLIF($4,''),NULLIF($5,''),NULLIF($6,''),
			NULLIF($7,''),$8,$9,NULLIF($10,''),$11,NULLIF($12,''),NULLIF($13,''),NULLIF($14,'')
		)`,
		uuid.NewString(), connectionID, versionID, actor.ID, actor.Role,
		actor.ServiceAccountID, actor.RequestID, action, result, truncate(reason, 255),
		encoded, truncate(actor.SourceIP, 64), idempotencyKey, requestHash,
	)

	return err
}

type platformCredentialIdempotencyReplay struct {
	connectionID        string
	credentialVersionID string
	requestHash         string
}

func findPlatformCredentialIdempotencyReplay(ctx context.Context, tx pgx.Tx, action, idempotencyKey string) (platformCredentialIdempotencyReplay, bool, error) {
	if strings.TrimSpace(idempotencyKey) == "" {
		return platformCredentialIdempotencyReplay{}, false, nil
	}

	var replay platformCredentialIdempotencyReplay

	err := tx.QueryRow(ctx, `SELECT connection_id,COALESCE(credential_version_id::text,''),COALESCE(request_hash,'') FROM platform_provider_credential_events WHERE action=$1 AND idempotency_key=$2`, action, idempotencyKey).Scan(&replay.connectionID, &replay.credentialVersionID, &replay.requestHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return platformCredentialIdempotencyReplay{}, false, nil
	}

	return replay, err == nil, err
}

func platformCredentialRequestHash(value interface{}) (string, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("encode platform credential request hash: %w", err)
	}

	digest := sha256.Sum256(encoded)

	return hex.EncodeToString(digest[:]), nil
}

func (s *Store) rollbackPlatformCredentialTx(ctx context.Context, tx pgx.Tx) {
	err := tx.Rollback(ctx)
	if err != nil && !errors.Is(err, pgx.ErrTxClosed) && s.logger != nil {
		s.logger.Warn("failed to rollback platform credential transaction", zap.Error(err))
	}
}

func platformCredentialRequestHashesEqual(left, right string) bool {
	return len(left) == len(right) && subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
}
func platformCredentialStringValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
func intValue(value *int) int {
	if value == nil {
		return 0
	}

	return *value
}
func floatValue(value *float64) float64 {
	if value == nil {
		return 0
	}

	return *value
}
func boolValue(value *bool) bool {
	if value == nil {
		return false
	}

	return *value
}
func truncate(value string, maxLength int) string {
	if len(value) > maxLength {
		return value[:maxLength]
	}

	return value
}

var _ platformcredentials.Repository = (*Store)(nil)
var _ = fmt.Sprintf
