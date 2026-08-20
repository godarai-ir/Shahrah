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
	"github.com/redis/go-redis/v9"
)

const ephemeralConversationTTL = 30 * time.Minute

// ConversationMessage is one decrypted runtime message.
type ConversationMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ResolveSession creates a session or validates that the supplied session belongs to the published chatbot.
func (s *Store) ResolveSession(
	ctx context.Context,
	chatbot PublishedChatbot,
	requestedID string,
) (string, error) {
	requestedID = strings.TrimSpace(requestedID)
	if requestedID != "" {
		if _, err := uuid.Parse(requestedID); err != nil {
			return "", ErrNotFound
		}
	}

	if chatbot.Chatbot.RetentionDays == 0 {
		return s.resolveEphemeralSession(ctx, chatbot, requestedID)
	}

	if requestedID == "" {
		return s.CreateSession(ctx, chatbot)
	}

	command, err := s.metadata.Exec(ctx, `
		UPDATE chatbot_sessions
		SET updated_at=now()
		WHERE tenant_id=$1 AND chatbot_id=$2 AND chatbot_version_id=$3 AND session_id=$4
		  AND (expires_at IS NULL OR expires_at > now())`,
		chatbot.Chatbot.TenantID, chatbot.Chatbot.ID, chatbot.Version.VersionID, requestedID)
	if err != nil {
		return "", fmt.Errorf("validate chatbot session: %w", err)
	}

	if command.RowsAffected() == 0 {
		return "", ErrNotFound
	}

	return requestedID, nil
}

// LoadConversation returns at most six previous turns in chronological order.
func (s *Store) LoadConversation(
	ctx context.Context,
	chatbot PublishedChatbot,
	sessionID string,
) ([]ConversationMessage, error) {
	if chatbot.Chatbot.RetentionDays == 0 {
		return s.loadEphemeralConversation(ctx, chatbot, sessionID)
	}

	rows, err := s.metadata.Query(ctx, `
		SELECT role, content_encrypted FROM (
			SELECT role, content_encrypted, created_at
			FROM chatbot_messages
			WHERE tenant_id=$1 AND session_id=$2 AND (expires_at IS NULL OR expires_at > now())
			ORDER BY created_at DESC LIMIT 12
		) recent ORDER BY created_at`, chatbot.Chatbot.TenantID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("load chatbot conversation: %w", err)
	}
	defer rows.Close()

	messages := make([]ConversationMessage, 0, 12)

	for rows.Next() {
		var (
			message   ConversationMessage
			encrypted []byte
		)

		if err := rows.Scan(&message.Role, &encrypted); err != nil {
			return nil, fmt.Errorf("scan chatbot conversation: %w", err)
		}

		content, err := decryptConversation(
			s.conversationKey,
			conversationAAD(chatbot.Chatbot.TenantID, sessionID, message.Role),
			encrypted,
		)
		if err != nil {
			return nil, fmt.Errorf("decrypt chatbot conversation: %w", err)
		}

		message.Content = content
		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chatbot conversation: %w", err)
	}

	return messages, nil
}

// SaveConversationTurn stores one user/assistant turn under the configured retention policy.
func (s *Store) SaveConversationTurn(
	ctx context.Context,
	chatbot PublishedChatbot,
	sessionID, question, answer string,
	refusal bool,
	citations []platform.Citation,
) error {
	if chatbot.Chatbot.RetentionDays == 0 {
		return s.saveEphemeralTurn(ctx, chatbot, sessionID, question, answer)
	}

	questionCiphertext, err := encryptConversation(
		s.conversationKey,
		conversationAAD(chatbot.Chatbot.TenantID, sessionID, "user"),
		question,
	)
	if err != nil {
		return fmt.Errorf("encrypt user message: %w", err)
	}

	answerCiphertext, err := encryptConversation(
		s.conversationKey,
		conversationAAD(chatbot.Chatbot.TenantID, sessionID, "assistant"),
		answer,
	)
	if err != nil {
		return fmt.Errorf("encrypt assistant message: %w", err)
	}

	citationJSON, err := json.Marshal(citations)
	if err != nil {
		return fmt.Errorf("encode conversation citations: %w", err)
	}

	expiresAt := time.Now().UTC().Add(time.Duration(chatbot.Chatbot.RetentionDays) * 24 * time.Hour)

	tx, err := s.metadata.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin save conversation turn: %w", err)
	}
	defer rollback(ctx, tx)

	var exists bool

	err = tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM chatbot_sessions
			WHERE tenant_id=$1 AND chatbot_id=$2 AND chatbot_version_id=$3 AND session_id=$4
			  AND (expires_at IS NULL OR expires_at > now())
		)`, chatbot.Chatbot.TenantID, chatbot.Chatbot.ID, chatbot.Version.VersionID, sessionID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("verify conversation session: %w", err)
	}

	if !exists {
		return ErrNotFound
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO chatbot_messages (tenant_id, session_id, role, content_encrypted, expires_at)
		VALUES ($1,$2,'user',$3,$4)`, chatbot.Chatbot.TenantID, sessionID, questionCiphertext, expiresAt); err != nil {
		return fmt.Errorf("store user message: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO chatbot_messages (tenant_id, session_id, role, content_encrypted, refusal, citations, expires_at)
		VALUES ($1,$2,'assistant',$3,$4,$5,$6)`, chatbot.Chatbot.TenantID, sessionID,
		answerCiphertext, refusal, citationJSON, expiresAt); err != nil {
		return fmt.Errorf("store assistant message: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE chatbot_sessions SET updated_at=now(), expires_at=$3
		WHERE tenant_id=$1 AND session_id=$2`, chatbot.Chatbot.TenantID, sessionID, expiresAt); err != nil {
		return fmt.Errorf("refresh conversation session: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit conversation turn: %w", err)
	}

	return nil
}

func (s *Store) resolveEphemeralSession(ctx context.Context, chatbot PublishedChatbot, requestedID string) (string, error) {
	if requestedID == "" {
		requestedID = uuid.NewString()
		if err := s.redis.Set(ctx, conversationSessionRedisKey(chatbot, requestedID), "1", ephemeralConversationTTL).Err(); err != nil {
			return "", fmt.Errorf("create ephemeral chatbot session: %w", err)
		}

		return requestedID, nil
	}

	exists, err := s.redis.Expire(ctx, conversationSessionRedisKey(chatbot, requestedID), ephemeralConversationTTL).Result()
	if err != nil {
		return "", fmt.Errorf("validate ephemeral chatbot session: %w", err)
	}

	if !exists {
		return "", ErrNotFound
	}

	return requestedID, nil
}

func (s *Store) loadEphemeralConversation(
	ctx context.Context,
	chatbot PublishedChatbot,
	sessionID string,
) ([]ConversationMessage, error) {
	key := conversationRedisKey(chatbot, sessionID)

	values, err := s.redis.LRange(ctx, key, -12, -1).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("load ephemeral conversation: %w", err)
	}

	messages := make([]ConversationMessage, 0, len(values))
	for _, value := range values {
		var message ConversationMessage
		if err := json.Unmarshal([]byte(value), &message); err != nil {
			return nil, fmt.Errorf("decode ephemeral conversation: %w", err)
		}

		messages = append(messages, message)
	}

	if err := s.redis.Expire(ctx, key, ephemeralConversationTTL).Err(); err != nil {
		return nil, fmt.Errorf("refresh ephemeral conversation: %w", err)
	}

	if err := s.redis.Expire(ctx, conversationSessionRedisKey(chatbot, sessionID), ephemeralConversationTTL).Err(); err != nil {
		return nil, fmt.Errorf("refresh ephemeral session: %w", err)
	}

	return messages, nil
}

func (s *Store) saveEphemeralTurn(
	ctx context.Context,
	chatbot PublishedChatbot,
	sessionID, question, answer string,
) error {
	key := conversationRedisKey(chatbot, sessionID)

	userJSON, err := json.Marshal(ConversationMessage{Role: "user", Content: question})
	if err != nil {
		return fmt.Errorf("encode ephemeral user message: %w", err)
	}

	assistantJSON, err := json.Marshal(ConversationMessage{Role: "assistant", Content: answer})
	if err != nil {
		return fmt.Errorf("encode ephemeral assistant message: %w", err)
	}

	_, err = s.redis.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.RPush(ctx, key, userJSON, assistantJSON)
		pipe.LTrim(ctx, key, -12, -1)
		pipe.Expire(ctx, key, ephemeralConversationTTL)
		pipe.Expire(ctx, conversationSessionRedisKey(chatbot, sessionID), ephemeralConversationTTL)

		return nil
	})
	if err != nil {
		return fmt.Errorf("store ephemeral conversation: %w", err)
	}

	return nil
}

func conversationRedisKey(chatbot PublishedChatbot, sessionID string) string {
	return "rag:conversation:" + chatbot.Chatbot.PublicID + ":" + chatbot.Version.VersionID + ":" + sessionID
}

func conversationSessionRedisKey(chatbot PublishedChatbot, sessionID string) string {
	return "rag:session:" + chatbot.Chatbot.PublicID + ":" + chatbot.Version.VersionID + ":" + sessionID
}

func conversationAAD(tenantID, sessionID, role string) []byte {
	return []byte("godarai-managed-rag-message-v1\x00" + tenantID + "\x00" + sessionID + "\x00" + role)
}

// PurgeExpiredConversations permanently removes retained content after its configured window.
func (s *Store) PurgeExpiredConversations(ctx context.Context) error {
	tx, err := s.metadata.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin purge expired conversations: %w", err)
	}
	defer rollback(ctx, tx)

	if _, err := tx.Exec(ctx, `DELETE FROM chatbot_messages WHERE expires_at IS NOT NULL AND expires_at <= now()`); err != nil {
		return fmt.Errorf("purge expired chatbot messages: %w", err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM chatbot_sessions WHERE expires_at IS NOT NULL AND expires_at <= now()`); err != nil {
		return fmt.Errorf("purge expired chatbot sessions: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit expired conversation purge: %w", err)
	}

	return nil
}
