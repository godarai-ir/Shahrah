package storage

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	supportdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/support"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const supportTicketSelect = `
	SELECT st.ticket_id::text,
	       st.public_id,
	       st.tenant_id::text,
	       COALESCE(t.slug, ''),
	       st.requester_user_id::text,
	       COALESCE(NULLIF(u.user_alias, ''), NULLIF(u.email, ''), st.requester_user_id::text),
	       COALESCE(u.email, ''),
	       st.subject,
	       st.category,
	       st.priority,
	       st.status,
	       COALESCE(st.assignee_id, ''),
	       st.last_message_author_type,
	       st.last_message_at,
	       st.first_support_response_at,
	       st.resolved_at,
	       st.closed_at,
	       st.created_at,
	       st.updated_at,
	       st.version,
	       str.score,
	       COALESCE(str.comment, ''),
	       COALESCE(str.requester_user_id::text, ''),
	       str.created_at,
	       str.updated_at
	FROM support_tickets st
	JOIN tenants t ON t.tenant_id = st.tenant_id
	JOIN users u ON u.user_id = st.requester_user_id
	LEFT JOIN support_ticket_ratings str ON str.ticket_id = st.ticket_id
`

type supportTicketRowScanner interface {
	Scan(dest ...interface{}) error
}

func scanSupportTicket(row supportTicketRowScanner) (*platform.SupportTicket, error) {
	var (
		item            platform.SupportTicket
		firstResponse   sql.NullTime
		resolvedAt      sql.NullTime
		closedAt        sql.NullTime
		ratingScore     sql.NullInt32
		ratingComment   string
		ratingSubmitter string
		ratingCreatedAt sql.NullTime
		ratingUpdatedAt sql.NullTime
	)

	if err := row.Scan(
		&item.ID,
		&item.PublicID,
		&item.TenantID,
		&item.TenantSlug,
		&item.RequesterUserID,
		&item.RequesterDisplayName,
		&item.RequesterEmail,
		&item.Subject,
		&item.Category,
		&item.Priority,
		&item.Status,
		&item.AssigneeID,
		&item.LastMessageAuthorType,
		&item.LastMessageAt,
		&firstResponse,
		&resolvedAt,
		&closedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.Version,
		&ratingScore,
		&ratingComment,
		&ratingSubmitter,
		&ratingCreatedAt,
		&ratingUpdatedAt,
	); err != nil {
		return nil, err
	}

	if firstResponse.Valid {
		value := firstResponse.Time
		item.FirstSupportResponseAt = &value
	}

	if resolvedAt.Valid {
		value := resolvedAt.Time
		item.ResolvedAt = &value
	}

	if closedAt.Valid {
		value := closedAt.Time
		item.ClosedAt = &value
	}

	if ratingScore.Valid {
		item.Rating = &platform.SupportTicketRating{
			Score:       int(ratingScore.Int32),
			Comment:     ratingComment,
			SubmittedBy: ratingSubmitter,
		}
		if ratingCreatedAt.Valid {
			item.Rating.CreatedAt = ratingCreatedAt.Time
		}

		if ratingUpdatedAt.Valid {
			item.Rating.UpdatedAt = ratingUpdatedAt.Time
		}
	}

	return &item, nil
}

func (s *SupportRepository) ListSupportTickets(ctx context.Context, params platform.SupportTicketListParams) ([]platform.SupportTicket, int, error) {
	if params.Limit <= 0 {
		params.Limit = 25
	}

	if params.Limit > 100 {
		params.Limit = 100
	}

	if params.Offset < 0 {
		params.Offset = 0
	}

	where, args := supportTicketWhere(params, true)

	var total int

	countQuery := `SELECT COUNT(1) FROM support_tickets st JOIN tenants t ON t.tenant_id = st.tenant_id JOIN users u ON u.user_id = st.requester_user_id ` + where
	if err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	orderBy := " ORDER BY st.last_message_at DESC, st.ticket_id DESC"
	if strings.EqualFold(strings.TrimSpace(params.Queue), "needs_support_reply") {
		orderBy = " ORDER BY st.last_message_at ASC, st.ticket_id ASC"
	}

	listArgs := append(append([]interface{}{}, args...), params.Limit, params.Offset)
	query := supportTicketSelect + where + orderBy + fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)

	rows, err := s.pool.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]platform.SupportTicket, 0, params.Limit)

	for rows.Next() {
		item, err := scanSupportTicket(rows)
		if err != nil {
			return nil, 0, err
		}

		items = append(items, *item)
	}

	return items, total, rows.Err()
}

func (s *SupportRepository) CountSupportTickets(ctx context.Context, params platform.SupportTicketListParams) (*platform.SupportTicketCounts, error) {
	params.Queue = ""
	params.Status = ""
	where, args := supportTicketWhere(params, false)

	query := `
		SELECT
			COUNT(1) FILTER (WHERE st.status = 'open' AND st.last_message_author_type = 'tenant_user'),
			COUNT(1) FILTER (WHERE st.status = 'open'),
			COUNT(1) FILTER (WHERE st.status = 'in_progress'),
			COUNT(1) FILTER (WHERE st.status = 'waiting_for_customer'),
			COUNT(1) FILTER (WHERE st.status = 'resolved'),
			COUNT(1) FILTER (WHERE st.status = 'closed'),
			COUNT(1),
			AVG(str.score::double precision)
		FROM support_tickets st
		JOIN tenants t ON t.tenant_id = st.tenant_id
		JOIN users u ON u.user_id = st.requester_user_id
		LEFT JOIN support_ticket_ratings str ON str.ticket_id = st.ticket_id
	` + where

	var (
		out           platform.SupportTicketCounts
		averageRating sql.NullFloat64
	)
	if err := s.pool.QueryRow(ctx, query, args...).Scan(
		&out.NeedsSupportReply,
		&out.Open,
		&out.InProgress,
		&out.WaitingForCustomer,
		&out.Resolved,
		&out.Closed,
		&out.All,
		&averageRating,
	); err != nil {
		return nil, err
	}

	if averageRating.Valid {
		out.AverageRating = &averageRating.Float64
	}

	return &out, nil
}

func supportTicketWhere(params platform.SupportTicketListParams, includeStatus bool) (string, []interface{}) {
	conditions := []string{"1 = 1"}
	args := make([]interface{}, 0, 12)
	add := func(condition string, value interface{}) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(condition, len(args)))
	}

	if value := strings.TrimSpace(params.TenantID); value != "" {
		add("st.tenant_id::text = $%d", value)
	}

	if !params.ReadAll && strings.TrimSpace(params.RequesterUserID) != "" {
		add("st.requester_user_id::text = $%d", strings.TrimSpace(params.RequesterUserID))
	}

	if includeStatus {
		if strings.EqualFold(strings.TrimSpace(params.Queue), "needs_support_reply") {
			conditions = append(conditions, "st.status = 'open'", "st.last_message_author_type = 'tenant_user'")
		} else if value := strings.TrimSpace(params.Status); value != "" {
			add("st.status = $%d", value)
		}
	}

	if value := strings.TrimSpace(params.Category); value != "" {
		add("st.category = $%d", value)
	}

	if value := strings.TrimSpace(params.Priority); value != "" {
		add("st.priority = $%d", value)
	}

	if value := strings.TrimSpace(params.AssigneeID); value != "" {
		if value == "unassigned" {
			conditions = append(conditions, "st.assignee_id IS NULL")
		} else {
			add("st.assignee_id = $%d", value)
		}
	}

	if value := strings.TrimSpace(params.Query); value != "" {
		args = append(args, value)
		placeholder := fmt.Sprintf("$%d", len(args))
		conditions = append(conditions,
			"(LOWER(st.public_id) LIKE LOWER('%' || "+placeholder+" || '%')"+
				" OR LOWER(st.subject) LIKE LOWER('%' || "+placeholder+" || '%')"+
				" OR LOWER(t.slug) LIKE LOWER('%' || "+placeholder+" || '%')"+
				" OR LOWER(COALESCE(u.email, '')) LIKE LOWER('%' || "+placeholder+" || '%'))",
		)
	}

	if params.UpdatedFrom != nil {
		add("st.last_message_at >= $%d", params.UpdatedFrom.UTC())
	}

	if params.UpdatedTo != nil {
		add("st.last_message_at <= $%d", params.UpdatedTo.UTC())
	}

	return " WHERE " + strings.Join(conditions, " AND "), args
}

func (s *SupportRepository) GetSupportTicket(ctx context.Context, tenantID, ticketRef, requesterUserID string, readAll bool) (*platform.SupportTicket, error) {
	conditions := []string{"(st.ticket_id::text = $1 OR UPPER(st.public_id) = UPPER($1))"}

	args := []interface{}{strings.TrimSpace(ticketRef)}
	if strings.TrimSpace(tenantID) != "" {
		args = append(args, strings.TrimSpace(tenantID))
		conditions = append(conditions, fmt.Sprintf("st.tenant_id::text = $%d", len(args)))
	}

	if !readAll && strings.TrimSpace(requesterUserID) != "" {
		args = append(args, strings.TrimSpace(requesterUserID))
		conditions = append(conditions, fmt.Sprintf("st.requester_user_id::text = $%d", len(args)))
	}

	item, err := scanSupportTicket(s.pool.QueryRow(ctx, supportTicketSelect+" WHERE "+strings.Join(conditions, " AND ")+" LIMIT 1", args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, supportdomain.ErrSupportTicketNotFound
	}

	return item, err
}

//nolint:govet // Short-lived branch errors are checked immediately and intentionally scoped with short declarations.
func (s *SupportRepository) CreateSupportTicket(
	ctx context.Context,
	tenantID string,
	requesterUserID string,
	req platform.CreateSupportTicketRequest,
	attachments []platform.SupportTicketAttachmentUpload,
	actor platform.SupportTicketActor,
	maxTicketsPerHour int,
) (*platform.SupportTicketDetail, bool, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, false, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	var existingID, existingHash string

	err = tx.QueryRow(ctx, `
		SELECT ticket_id::text, request_hash
		FROM support_tickets
		WHERE tenant_id = $1::uuid AND requester_user_id = $2::uuid AND idempotency_key = $3
		LIMIT 1
	`, tenantID, requesterUserID, req.IdempotencyKey).Scan(&existingID, &existingHash)
	if err == nil {
		if existingHash != req.RequestHash {
			return nil, false, supportdomain.ErrSupportTicketIdempotency
		}

		if err := tx.Commit(ctx); err != nil {
			return nil, false, err
		}

		detail, loadErr := s.GetSupportTicketDetail(ctx, tenantID, existingID, requesterUserID, false)

		return detail, true, loadErr
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, false, err
	}

	if maxTicketsPerHour > 0 {
		var recent int
		if err := tx.QueryRow(ctx, `
			SELECT COUNT(1)
			FROM support_tickets
			WHERE tenant_id = $1::uuid AND requester_user_id = $2::uuid AND created_at >= NOW() - INTERVAL '1 hour'
		`, tenantID, requesterUserID).Scan(&recent); err != nil {
			return nil, false, err
		}

		if recent >= maxTicketsPerHour {
			return nil, false, supportdomain.ErrSupportTicketRateLimited
		}
	}

	ticketID := uuid.NewString()
	messageID := uuid.NewString()

	publicID, err := newSupportTicketPublicID()
	if err != nil {
		return nil, false, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO support_tickets (
			ticket_id, public_id, tenant_id, requester_user_id, subject, category,
			priority, status, last_message_author_type, last_message_at,
			idempotency_key, request_hash, version, created_at, updated_at
		)
		VALUES ($1::uuid, $2, $3::uuid, $4::uuid, $5, $6, 'normal', 'open',
		        'tenant_user', NOW(), $7, $8, 1, NOW(), NOW())
	`, ticketID, publicID, tenantID, requesterUserID, strings.TrimSpace(req.Subject), req.Category, req.IdempotencyKey, req.RequestHash); err != nil {
		return nil, false, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO support_ticket_messages (
			message_id, ticket_id, author_type, author_id, body, idempotency_key, request_hash, created_at
		)
		VALUES ($1::uuid, $2::uuid, 'tenant_user', $3, $4, $5, $6, NOW())
	`, messageID, ticketID, requesterUserID, req.Body, req.IdempotencyKey, req.RequestHash); err != nil {
		return nil, false, err
	}

	if err := insertSupportTicketAttachmentsTx(ctx, tx, tenantID, ticketID, messageID, attachments); err != nil {
		return nil, false, err
	}

	if err := insertSupportTicketEventTx(ctx, tx, ticketID, tenantID, "ticket.created", actor, map[string]interface{}{
		"category": req.Category,
		"priority": platform.SupportTicketPriorityNormal,
	}); err != nil {
		return nil, false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, false, err
	}

	detail, err := s.GetSupportTicketDetail(ctx, tenantID, ticketID, requesterUserID, false)

	return detail, false, err
}

func (s *SupportRepository) GetSupportTicketDetail(ctx context.Context, tenantID, ticketRef, requesterUserID string, readAll bool) (*platform.SupportTicketDetail, error) {
	ticket, err := s.GetSupportTicket(ctx, tenantID, ticketRef, requesterUserID, readAll)
	if err != nil {
		return nil, err
	}

	messages, err := s.ListSupportTicketMessages(ctx, ticket.ID, 100, 0)
	if err != nil {
		return nil, err
	}

	return &platform.SupportTicketDetail{Ticket: *ticket, Messages: messages}, nil
}

//nolint:govet // Short-lived branch errors are checked immediately and intentionally scoped with short declarations.
func (s *SupportRepository) ListSupportTicketMessages(ctx context.Context, ticketID string, limit, offset int) ([]platform.SupportTicketMessage, error) {
	if limit <= 0 {
		limit = 50
	}

	if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	rows, err := s.pool.Query(ctx, `
		SELECT stm.message_id::text,
		       stm.ticket_id::text,
		       stm.author_type,
		       stm.author_id,
		       CASE
		         WHEN stm.author_type = 'tenant_user'
		         THEN COALESCE(NULLIF(u.user_alias, ''), NULLIF(u.email, ''), stm.author_id)
		         WHEN stm.author_type = 'support_agent' THEN 'GodarAI Support'
		         ELSE 'System'
		       END,
		       stm.body,
		       stm.created_at
		FROM support_ticket_messages stm
		LEFT JOIN users u ON stm.author_type = 'tenant_user' AND u.user_id::text = stm.author_id
		WHERE stm.ticket_id = $1::uuid
		ORDER BY stm.created_at ASC, stm.message_id ASC
		LIMIT $2 OFFSET $3
	`, ticketID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]platform.SupportTicketMessage, 0, limit)
	messageIndex := make(map[string]int, limit)

	for rows.Next() {
		var item platform.SupportTicketMessage
		if err := rows.Scan(&item.ID, &item.TicketID, &item.AuthorType, &item.AuthorID, &item.AuthorLabel, &item.Body, &item.CreatedAt); err != nil {
			return nil, err
		}

		item.Attachments = []platform.SupportTicketAttachment{}
		messageIndex[item.ID] = len(messages)
		messages = append(messages, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(messages) == 0 {
		return messages, nil
	}

	ids := make([]uuid.UUID, 0, len(messages))
	for _, message := range messages {
		if parsed, parseErr := uuid.Parse(message.ID); parseErr == nil {
			ids = append(ids, parsed)
		}
	}

	attachmentRows, err := s.pool.Query(ctx, `
		SELECT attachment_id::text, ticket_id::text, message_id::text, original_filename,
		       detected_content_type, size_bytes, sha256, scan_status, created_at, scanned_at
		FROM support_ticket_attachments
		WHERE message_id = ANY($1::uuid[])
		ORDER BY created_at, attachment_id
	`, ids)
	if err != nil {
		return nil, err
	}
	defer attachmentRows.Close()

	for attachmentRows.Next() {
		var (
			item      platform.SupportTicketAttachment
			scannedAt sql.NullTime
		)
		if err := attachmentRows.Scan(
			&item.ID,
			&item.TicketID,
			&item.MessageID,
			&item.OriginalFilename,
			&item.DetectedContentType,
			&item.SizeBytes,
			&item.SHA256,
			&item.ScanStatus,
			&item.CreatedAt,
			&scannedAt,
		); err != nil {
			return nil, err
		}

		if scannedAt.Valid {
			value := scannedAt.Time
			item.ScannedAt = &value
		}

		if index, ok := messageIndex[item.MessageID]; ok {
			messages[index].Attachments = append(messages[index].Attachments, item)
		}
	}

	return messages, attachmentRows.Err()
}

//nolint:govet // Short-lived branch errors are checked immediately and intentionally scoped with short declarations.
func (s *SupportRepository) AddSupportTicketMessage(
	ctx context.Context,
	tenantID string,
	ticketRef string,
	req platform.AddSupportTicketMessageRequest,
	attachments []platform.SupportTicketAttachmentUpload,
	actor platform.SupportTicketActor,
	readAll bool,
	maxMessagesPerHour int,
) (*platform.SupportTicketDetail, bool, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, false, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	ticket, err := lockSupportTicketTx(ctx, tx, tenantID, ticketRef, actor.ID, readAll)
	if err != nil {
		return nil, false, err
	}

	var existingMessageID, existingHash string

	err = tx.QueryRow(ctx, `
		SELECT message_id::text, request_hash
		FROM support_ticket_messages
		WHERE ticket_id = $1::uuid AND author_type = $2 AND author_id = $3 AND idempotency_key = $4
		LIMIT 1
	`, ticket.ID, actor.Type, actor.ID, req.IdempotencyKey).Scan(&existingMessageID, &existingHash)
	if err == nil {
		if existingHash != req.RequestHash {
			return nil, false, supportdomain.ErrSupportTicketIdempotency
		}

		if err := tx.Commit(ctx); err != nil {
			return nil, false, err
		}

		detail, loadErr := s.GetSupportTicketDetail(ctx, tenantID, ticket.ID, actor.ID, readAll)

		return detail, true, loadErr
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, false, err
	}

	now := time.Now().UTC()

	if ticket.Status == platform.SupportTicketStatusClosed {
		return nil, false, supportdomain.ErrSupportTicketClosed
	}

	if ticket.Status == platform.SupportTicketStatusResolved && ticket.ResolvedAt != nil && now.Sub(ticket.ResolvedAt.UTC()) > 7*24*time.Hour {
		return nil, false, supportdomain.ErrSupportTicketClosed
	}

	if maxMessagesPerHour > 0 {
		var recent int
		if err := tx.QueryRow(ctx, `
			SELECT COUNT(1)
			FROM support_ticket_messages
			WHERE ticket_id = $1::uuid AND author_type = $2 AND author_id = $3
			  AND created_at >= NOW() - INTERVAL '1 hour'
		`, ticket.ID, actor.Type, actor.ID).Scan(&recent); err != nil {
			return nil, false, err
		}

		if recent >= maxMessagesPerHour {
			return nil, false, supportdomain.ErrSupportTicketRateLimited
		}
	}

	nextStatus := ticket.Status
	if actor.Type == platform.SupportTicketAuthorTenantUser {
		if ticket.Status == platform.SupportTicketStatusWaitingForCustomer || ticket.Status == platform.SupportTicketStatusResolved {
			nextStatus = platform.SupportTicketStatusOpen
		}
	} else if strings.TrimSpace(req.NextStatus) != "" {
		nextStatus = strings.TrimSpace(req.NextStatus)
	} else {
		nextStatus = platform.SupportTicketStatusInProgress
	}

	if !supportdomain.CanTransitionTicket(ticket.Status, nextStatus, actor.Type == platform.SupportTicketAuthorSupportAgent) {
		return nil, false, supportdomain.ErrSupportTicketInvalidTransition
	}

	messageID := uuid.NewString()
	if _, err := tx.Exec(ctx, `
		INSERT INTO support_ticket_messages (
			message_id, ticket_id, author_type, author_id, body, idempotency_key, request_hash, created_at
		)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, NOW())
	`, messageID, ticket.ID, actor.Type, actor.ID, req.Body, req.IdempotencyKey, req.RequestHash); err != nil {
		return nil, false, err
	}

	if err := insertSupportTicketAttachmentsTx(ctx, tx, ticket.TenantID, ticket.ID, messageID, attachments); err != nil {
		return nil, false, err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE support_tickets
		SET status = $2::varchar,
		    last_message_author_type = $3::varchar,
		    last_message_at = NOW(),
		    first_support_response_at = CASE
		      WHEN $3::varchar = 'support_agent' THEN COALESCE(first_support_response_at, NOW())
		      ELSE first_support_response_at
		    END,
		    resolved_at = CASE WHEN $2::varchar = 'resolved' THEN NOW() WHEN $2::varchar <> 'resolved' THEN NULL ELSE resolved_at END,
		    closed_at = CASE WHEN $2::varchar = 'closed' THEN NOW() WHEN $2::varchar <> 'closed' THEN NULL ELSE closed_at END,
		    version = version + 1,
		    updated_at = NOW()
		WHERE ticket_id = $1::uuid
	`, ticket.ID, nextStatus, actor.Type); err != nil {
		return nil, false, err
	}

	if err := insertSupportTicketEventTx(ctx, tx, ticket.ID, ticket.TenantID, "ticket.message_added", actor, map[string]interface{}{
		"attachment_count": len(attachments),
		"status_before":    ticket.Status,
		"status_after":     nextStatus,
	}); err != nil {
		return nil, false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, false, err
	}

	detail, err := s.GetSupportTicketDetail(ctx, tenantID, ticket.ID, actor.ID, readAll)

	return detail, false, err
}

func (s *SupportRepository) TriageSupportTicket(ctx context.Context, ticketRef string, req platform.TriageSupportTicketRequest, actor platform.SupportTicketActor) (*platform.SupportTicket, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	ticket, err := lockSupportTicketTx(ctx, tx, "", ticketRef, "", true)
	if err != nil {
		return nil, err
	}

	requestHash := supportMutationRequestHash(req)

	replay, err := supportMutationReplayTx(ctx, tx, ticket.ID, actor, "triage", req.IdempotencyKey, requestHash)
	if err != nil {
		return nil, err
	}

	if replay {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}

		return s.GetSupportTicket(ctx, "", ticket.ID, "", true)
	}

	if req.Version <= 0 || ticket.Version != req.Version {
		return nil, supportdomain.ErrSupportTicketStaleVersion
	}

	nextStatus := ticket.Status
	if req.Status != nil {
		nextStatus = strings.TrimSpace(*req.Status)
		if !supportdomain.CanTransitionTicket(ticket.Status, nextStatus, true) {
			return nil, supportdomain.ErrSupportTicketInvalidTransition
		}

		if ticket.Status == platform.SupportTicketStatusClosed && nextStatus == platform.SupportTicketStatusOpen {
			return nil, supportdomain.ErrSupportTicketInvalidTransition
		}
	}

	nextPriority := ticket.Priority
	if req.Priority != nil {
		nextPriority = strings.TrimSpace(*req.Priority)
	}

	nextAssignee := ticket.AssigneeID
	if req.AssigneeID != nil {
		nextAssignee = strings.TrimSpace(*req.AssigneeID)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE support_tickets
		SET status = $2::varchar,
		    priority = $3::varchar,
		    assignee_id = NULLIF($4, ''),
		    resolved_at = CASE WHEN $2::varchar = 'resolved' AND status <> 'resolved' THEN NOW() WHEN $2::varchar <> 'resolved' THEN NULL ELSE resolved_at END,
		    closed_at = CASE WHEN $2::varchar = 'closed' AND status <> 'closed' THEN NOW() WHEN $2::varchar <> 'closed' THEN NULL ELSE closed_at END,
		    version = version + 1,
		    updated_at = NOW()
		WHERE ticket_id = $1::uuid
	`, ticket.ID, nextStatus, nextPriority, nextAssignee); err != nil {
		return nil, err
	}

	if err := insertSupportTicketEventTx(ctx, tx, ticket.ID, ticket.TenantID, "ticket.triaged", actor, map[string]interface{}{
		"status_before":    ticket.Status,
		"status_after":     nextStatus,
		"priority_before":  ticket.Priority,
		"priority_after":   nextPriority,
		"assignee_changed": ticket.AssigneeID != nextAssignee,
		"reason":           strings.TrimSpace(req.Reason),
	}); err != nil {
		return nil, err
	}

	if err := recordSupportMutationTx(ctx, tx, ticket.ID, actor, "triage", req.IdempotencyKey, requestHash); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return s.GetSupportTicket(ctx, "", ticket.ID, "", true)
}

func (s *SupportRepository) ReopenSupportTicket(ctx context.Context, ticketRef string, req platform.ReopenSupportTicketRequest, actor platform.SupportTicketActor) (*platform.SupportTicket, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	ticket, err := lockSupportTicketTx(ctx, tx, "", ticketRef, "", true)
	if err != nil {
		return nil, err
	}

	requestHash := supportMutationRequestHash(req)

	replay, err := supportMutationReplayTx(ctx, tx, ticket.ID, actor, "reopen", req.IdempotencyKey, requestHash)
	if err != nil {
		return nil, err
	}

	if replay {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}

		return s.GetSupportTicket(ctx, "", ticket.ID, "", true)
	}

	if ticket.Version != req.Version {
		return nil, supportdomain.ErrSupportTicketStaleVersion
	}

	if ticket.Status != platform.SupportTicketStatusClosed || strings.TrimSpace(req.Reason) == "" {
		return nil, supportdomain.ErrSupportTicketInvalidTransition
	}

	if _, err := tx.Exec(ctx, `
		UPDATE support_tickets
		SET status = 'open', closed_at = NULL, resolved_at = NULL,
		    version = version + 1, updated_at = NOW()
		WHERE ticket_id = $1::uuid
	`, ticket.ID); err != nil {
		return nil, err
	}

	if err := insertSupportTicketEventTx(ctx, tx, ticket.ID, ticket.TenantID, "ticket.reopened", actor, map[string]interface{}{
		"reason": strings.TrimSpace(req.Reason),
	}); err != nil {
		return nil, err
	}

	if err := recordSupportMutationTx(ctx, tx, ticket.ID, actor, "reopen", req.IdempotencyKey, requestHash); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return s.GetSupportTicket(ctx, "", ticket.ID, "", true)
}

func (s *SupportRepository) RateSupportTicket(ctx context.Context, tenantID, ticketRef, requesterUserID string, req platform.RateSupportTicketRequest, actor platform.SupportTicketActor) (*platform.SupportTicket, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	ticket, err := lockSupportTicketTx(ctx, tx, tenantID, ticketRef, requesterUserID, false)
	if err != nil {
		return nil, err
	}

	requestHash := supportMutationRequestHash(req)

	replay, err := supportMutationReplayTx(ctx, tx, ticket.ID, actor, "rating", req.IdempotencyKey, requestHash)
	if err != nil {
		return nil, err
	}

	if replay {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}

		return s.GetSupportTicket(ctx, tenantID, ticket.ID, requesterUserID, false)
	}

	var ratingExists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM support_ticket_ratings WHERE ticket_id = $1::uuid)`, ticket.ID).Scan(&ratingExists); err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	if ticket.Status != platform.SupportTicketStatusResolved && ticket.Status != platform.SupportTicketStatusClosed {
		return nil, supportdomain.ErrSupportTicketRatingNotAllowed
	}

	if ticket.Status == platform.SupportTicketStatusClosed && (ticket.ClosedAt == nil || now.Sub(ticket.ClosedAt.UTC()) > 30*24*time.Hour) {
		return nil, supportdomain.ErrSupportTicketRatingNotAllowed
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO support_ticket_ratings (
			ticket_id, tenant_id, requester_user_id, score, comment, created_at, updated_at
		)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4, NULLIF($5, ''), NOW(), NOW())
		ON CONFLICT (ticket_id)
		DO UPDATE SET score = EXCLUDED.score, comment = EXCLUDED.comment, updated_at = NOW()
	`, ticket.ID, ticket.TenantID, requesterUserID, req.Score, strings.TrimSpace(req.Comment)); err != nil {
		return nil, err
	}

	if ticket.Status == platform.SupportTicketStatusResolved {
		if _, err := tx.Exec(ctx, `
			UPDATE support_tickets
			SET status = 'closed', closed_at = NOW(), version = version + 1, updated_at = NOW()
			WHERE ticket_id = $1::uuid
		`, ticket.ID); err != nil {
			return nil, err
		}
	} else {
		if _, err := tx.Exec(ctx, `
			UPDATE support_tickets SET version = version + 1, updated_at = NOW() WHERE ticket_id = $1::uuid
		`, ticket.ID); err != nil {
			return nil, err
		}
	}

	ratingEvent := "ticket.rating_submitted"
	if ratingExists {
		ratingEvent = "ticket.rating_updated"
	}

	if err := insertSupportTicketEventTx(ctx, tx, ticket.ID, ticket.TenantID, ratingEvent, actor, map[string]interface{}{
		"score": req.Score,
	}); err != nil {
		return nil, err
	}

	if err := recordSupportMutationTx(ctx, tx, ticket.ID, actor, "rating", req.IdempotencyKey, requestHash); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return s.GetSupportTicket(ctx, tenantID, ticket.ID, requesterUserID, false)
}

func (s *SupportRepository) ListSupportTicketAttachmentObjectKeys(ctx context.Context) (map[string]struct{}, error) {
	rows, err := s.pool.Query(ctx, `SELECT object_key FROM support_ticket_attachments`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := make(map[string]struct{})

	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}

		keys[key] = struct{}{}
	}

	return keys, rows.Err()
}

func (s *SupportRepository) ListSupportTicketEvents(ctx context.Context, ticketID string, limit, offset int) ([]platform.SupportTicketEvent, error) {
	if limit <= 0 {
		limit = 50
	}

	if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	rows, err := s.pool.Query(ctx, `
		SELECT event_id::text, ticket_id::text, tenant_id::text, event_type, actor_type,
		       COALESCE(actor_id, ''), COALESCE(actor_role, ''), COALESCE(request_id, ''),
		       payload, created_at
		FROM support_ticket_events
		WHERE ticket_id = $1::uuid
		ORDER BY created_at DESC, event_id DESC
		LIMIT $2 OFFSET $3
	`, ticketID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]platform.SupportTicketEvent, 0, limit)

	for rows.Next() {
		var (
			item platform.SupportTicketEvent
			raw  []byte
		)
		if err := rows.Scan(
			&item.ID,
			&item.TicketID,
			&item.TenantID,
			&item.Type,
			&item.ActorType,
			&item.ActorID,
			&item.ActorRole,
			&item.RequestID,
			&raw,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}

		item.Payload = map[string]interface{}{}
		_ = json.Unmarshal(raw, &item.Payload) //nolint:errcheck // Invalid optional response metadata is treated as absent.
		items = append(items, item)
	}

	return items, rows.Err()
}

func (s *SupportRepository) GetSupportTicketAttachment(ctx context.Context, tenantID, attachmentID, requesterUserID string, readAll bool) (*platform.SupportTicketAttachmentObject, error) {
	args := []interface{}{strings.TrimSpace(attachmentID)}
	conditions := []string{"sta.attachment_id::text = $1"}

	if strings.TrimSpace(tenantID) != "" {
		args = append(args, strings.TrimSpace(tenantID))
		conditions = append(conditions, fmt.Sprintf("sta.tenant_id::text = $%d", len(args)))
	}

	if !readAll && strings.TrimSpace(requesterUserID) != "" {
		args = append(args, strings.TrimSpace(requesterUserID))
		conditions = append(conditions, fmt.Sprintf("st.requester_user_id::text = $%d", len(args)))
	}

	var (
		item      platform.SupportTicketAttachmentObject
		scannedAt sql.NullTime
	)

	err := s.pool.QueryRow(ctx, `
		SELECT sta.attachment_id::text, sta.ticket_id::text, sta.message_id::text,
		       sta.original_filename, sta.detected_content_type, sta.size_bytes, sta.sha256,
		       sta.scan_status, sta.created_at, sta.scanned_at, sta.tenant_id::text, sta.object_key
		FROM support_ticket_attachments sta
		JOIN support_tickets st ON st.ticket_id = sta.ticket_id
		WHERE `+strings.Join(conditions, " AND ")+`
		LIMIT 1
	`, args...).Scan(
		&item.ID,
		&item.TicketID,
		&item.MessageID,
		&item.OriginalFilename,
		&item.DetectedContentType,
		&item.SizeBytes,
		&item.SHA256,
		&item.ScanStatus,
		&item.CreatedAt,
		&scannedAt,
		&item.TenantID,
		&item.ObjectKey,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, supportdomain.ErrSupportTicketNotFound
	}

	if err != nil {
		return nil, err
	}

	if scannedAt.Valid {
		value := scannedAt.Time
		item.ScannedAt = &value
	}

	return &item, nil
}

func (s *SupportRepository) RecordSupportTicketAttachmentDownload(ctx context.Context, attachment *platform.SupportTicketAttachmentObject, actor platform.SupportTicketActor, outcome string) error {
	if attachment == nil {
		return nil
	}

	raw, _ := json.Marshal(map[string]interface{}{ //nolint:errcheck // The value is composed exclusively of JSON-safe fields.
		"attachment_id": attachment.ID,
		"outcome":       strings.TrimSpace(outcome),
	})
	_, err := s.pool.Exec(ctx, `
		INSERT INTO support_ticket_events (
			ticket_id, tenant_id, event_type, actor_type, actor_id, actor_role, request_id, payload, created_at
		)
		VALUES ($1::uuid, $2::uuid, 'ticket.attachment_downloaded', $3, NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''), $7, NOW())
	`, attachment.TicketID, attachment.TenantID, actor.Type, actor.ID, actor.Role, actor.RequestID, raw)

	return err
}

func (s *SupportRepository) CloseResolvedSupportTickets(ctx context.Context, before time.Time, batchSize int) (int, error) {
	if batchSize <= 0 {
		batchSize = 100
	}

	if batchSize > 1000 {
		batchSize = 1000
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	rows, err := tx.Query(ctx, `
		SELECT ticket_id::text, tenant_id::text
		FROM support_tickets
		WHERE status = 'resolved' AND resolved_at <= $1
		ORDER BY resolved_at, ticket_id
		LIMIT $2
		FOR UPDATE SKIP LOCKED
	`, before.UTC(), batchSize)
	if err != nil {
		return 0, err
	}

	type target struct{ ticketID, tenantID string }

	targets := make([]target, 0, batchSize)

	for rows.Next() {
		var item target
		if err := rows.Scan(&item.ticketID, &item.tenantID); err != nil {
			rows.Close()
			return 0, err
		}

		targets = append(targets, item)
	}

	rows.Close()

	if err := rows.Err(); err != nil {
		return 0, err
	}

	systemActor := platform.SupportTicketActor{Type: platform.SupportTicketAuthorSystem, ID: "support-ticket-closer", Role: "system"}

	for _, item := range targets {
		if _, err := tx.Exec(ctx, `
			UPDATE support_tickets
			SET status = 'closed', closed_at = NOW(), version = version + 1, updated_at = NOW()
			WHERE ticket_id = $1::uuid
		`, item.ticketID); err != nil {
			return 0, err
		}

		if err := insertSupportTicketEventTx(ctx, tx, item.ticketID, item.tenantID, "ticket.auto_closed", systemActor, map[string]interface{}{}); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return len(targets), nil
}

func lockSupportTicketTx(ctx context.Context, tx pgx.Tx, tenantID, ticketRef, requesterUserID string, readAll bool) (*platform.SupportTicket, error) {
	conditions := []string{"(st.ticket_id::text = $1 OR UPPER(st.public_id) = UPPER($1))"}

	args := []interface{}{strings.TrimSpace(ticketRef)}
	if strings.TrimSpace(tenantID) != "" {
		args = append(args, strings.TrimSpace(tenantID))
		conditions = append(conditions, fmt.Sprintf("st.tenant_id::text = $%d", len(args)))
	}

	if !readAll && strings.TrimSpace(requesterUserID) != "" {
		args = append(args, strings.TrimSpace(requesterUserID))
		conditions = append(conditions, fmt.Sprintf("st.requester_user_id::text = $%d", len(args)))
	}

	var (
		item          platform.SupportTicket
		firstResponse sql.NullTime
		resolvedAt    sql.NullTime
		closedAt      sql.NullTime
	)

	err := tx.QueryRow(ctx, `
		SELECT st.ticket_id::text, st.public_id, st.tenant_id::text, st.requester_user_id::text,
		       st.subject, st.category, st.priority, st.status, COALESCE(st.assignee_id, ''),
		       st.last_message_author_type, st.last_message_at, st.first_support_response_at,
		       st.resolved_at, st.closed_at, st.created_at, st.updated_at, st.version
		FROM support_tickets st
		WHERE `+strings.Join(conditions, " AND ")+`
		LIMIT 1
		FOR UPDATE
	`, args...).Scan(
		&item.ID,
		&item.PublicID,
		&item.TenantID,
		&item.RequesterUserID,
		&item.Subject,
		&item.Category,
		&item.Priority,
		&item.Status,
		&item.AssigneeID,
		&item.LastMessageAuthorType,
		&item.LastMessageAt,
		&firstResponse,
		&resolvedAt,
		&closedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.Version,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, supportdomain.ErrSupportTicketNotFound
	}

	if err != nil {
		return nil, err
	}

	if firstResponse.Valid {
		value := firstResponse.Time
		item.FirstSupportResponseAt = &value
	}

	if resolvedAt.Valid {
		value := resolvedAt.Time
		item.ResolvedAt = &value
	}

	if closedAt.Valid {
		value := closedAt.Time
		item.ClosedAt = &value
	}

	return &item, nil
}

func insertSupportTicketAttachmentsTx(
	ctx context.Context,
	tx pgx.Tx,
	tenantID string,
	ticketID string,
	messageID string,
	attachments []platform.SupportTicketAttachmentUpload,
) error {
	for _, item := range attachments {
		if _, err := tx.Exec(ctx, `
			INSERT INTO support_ticket_attachments (
				attachment_id, message_id, ticket_id, tenant_id, original_filename,
				detected_content_type, size_bytes, sha256, object_key, scan_status,
				scan_result, created_at, scanned_at
			)
			VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, $5, $6, $7, $8, $9, $10, NULLIF($11, ''), NOW(), $12)
		`, item.ID, messageID, ticketID, tenantID, item.OriginalFilename, item.DetectedContentType,
			item.SizeBytes, item.SHA256, item.ObjectKey, item.ScanStatus, item.ScanResult, item.ScannedAt); err != nil {
			return err
		}
	}

	return nil
}

func insertSupportTicketEventTx(
	ctx context.Context,
	tx pgx.Tx,
	ticketID string,
	tenantID string,
	eventType string,
	actor platform.SupportTicketActor,
	payload map[string]interface{},
) error {
	raw, _ := json.Marshal(payload) //nolint:errcheck // The value is composed exclusively of JSON-safe fields.
	_, err := tx.Exec(ctx, `
		INSERT INTO support_ticket_events (
			ticket_id, tenant_id, event_type, actor_type, actor_id, actor_role, request_id, payload, created_at
		)
		VALUES ($1::uuid, $2::uuid, $3, $4, NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''), $8, NOW())
	`, ticketID, tenantID, eventType, actor.Type, actor.ID, actor.Role, actor.RequestID, raw)

	return err
}

func supportMutationRequestHash(value interface{}) string {
	raw, _ := json.Marshal(value) //nolint:errcheck // The value is composed exclusively of JSON-safe fields.
	sum := sha256.Sum256(raw)

	return fmt.Sprintf("%x", sum[:])
}

func supportMutationReplayTx(
	ctx context.Context,
	tx pgx.Tx,
	ticketID string,
	actor platform.SupportTicketActor,
	operation string,
	idempotencyKey string,
	requestHash string,
) (bool, error) {
	var existingHash string

	err := tx.QueryRow(ctx, `
		SELECT request_hash
		FROM support_ticket_mutations
		WHERE ticket_id = $1::uuid AND actor_type = $2 AND actor_id = $3
		  AND operation = $4 AND idempotency_key = $5
		LIMIT 1
	`, ticketID, actor.Type, actor.ID, operation, idempotencyKey).Scan(&existingHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	if existingHash != requestHash {
		return false, supportdomain.ErrSupportTicketIdempotency
	}

	return true, nil
}

func recordSupportMutationTx(
	ctx context.Context,
	tx pgx.Tx,
	ticketID string,
	actor platform.SupportTicketActor,
	operation string,
	idempotencyKey string,
	requestHash string,
) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO support_ticket_mutations (
			ticket_id, actor_type, actor_id, operation, idempotency_key, request_hash, created_at
		)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, NOW())
	`, ticketID, actor.Type, actor.ID, operation, idempotencyKey, requestHash)

	return err
}

func newSupportTicketPublicID() (string, error) {
	raw := make([]byte, 8)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}

	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw)

	return "TKT-" + encoded, nil
}
