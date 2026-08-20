// Package worker implements durable, leased managed RAG ingestion.
package worker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Delpak/shahrah/services/rag/internal/domain"
	"github.com/Delpak/shahrah/services/rag/internal/store"
	"go.uber.org/zap"
)

const (
	embeddingBatchSize = 64
	maxSourceBytes     = 25 << 20
)

// JobStore is the durable ingestion persistence boundary.
type JobStore interface {
	ClaimPurgeJob(context.Context, string, time.Duration) (*store.PurgeJob, error)
	CompletePurge(context.Context, string, string, string) error
	FailPurge(context.Context, store.PurgeJob, string, string) error
	PurgeTenantVectors(context.Context, string) error
	PurgeKnowledgeBaseVectors(context.Context, string, string) error
	PurgeDocumentVectors(context.Context, string, string) error
	ClaimIngestionJob(context.Context, string, time.Duration) (*store.IngestionJob, error)
	UpdateIngestionStage(context.Context, string, string, string, string, string, time.Duration) error
	RenewIngestionLease(context.Context, string, string, string, time.Duration) error
	ReplaceVersionChunks(context.Context, []store.VectorChunk) error
	CompleteIngestion(context.Context, store.IngestionJob, string, *int, int, string) error
	FailIngestion(context.Context, store.IngestionJob, string, string, error) error
	RecordUsage(context.Context, store.UsageEvent) error
	PurgeExpiredConversations(context.Context) error
}

// ObjectReader opens private source objects.
type ObjectReader interface {
	Get(context.Context, string) (io.ReadCloser, error)
	Delete(context.Context, string) error
}

// TextExtractor extracts embedded structured content from supported files.
type TextExtractor interface {
	Extract(context.Context, string, io.Reader) (domain.SourceDocument, error)
}

// DocumentOCR extracts page-scoped content when a PDF has no usable embedded text.
type DocumentOCR interface {
	OCR(context.Context, string, string, string, []byte) (domain.SourceDocument, int, error)
}

// Embedder creates pinned-dimension vectors through the metered Godar gateway.
type Embedder interface {
	Embed(context.Context, string, string, []string, int) ([][]float32, int, error)
}

// Processor claims and executes resumable ingestion jobs.
type Processor struct {
	store                 JobStore
	objects               ObjectReader
	parser                TextExtractor
	ocr                   DocumentOCR
	embedder              Embedder
	tokenizer             domain.TokenCounter
	logger                *zap.Logger
	owner                 string
	poll                  time.Duration
	lease                 time.Duration
	nextConversationPurge time.Time
}

// New creates an ingestion processor.
func New(
	repository JobStore,
	objects ObjectReader,
	parser TextExtractor,
	ocr DocumentOCR,
	embedder Embedder,
	tokenizer domain.TokenCounter,
	logger *zap.Logger,
	owner string,
) (*Processor, error) {
	if repository == nil || objects == nil || parser == nil || ocr == nil || embedder == nil || tokenizer == nil || logger == nil {
		return nil, errors.New("ingestion dependencies are required")
	}

	if strings.TrimSpace(owner) == "" {
		return nil, errors.New("ingestion worker owner is required")
	}

	return &Processor{
		store: repository, objects: objects, parser: parser, ocr: ocr, embedder: embedder, tokenizer: tokenizer,
		logger: logger, owner: owner, poll: time.Second, lease: 2 * time.Minute,
		nextConversationPurge: time.Now(),
	}, nil
}

// Run processes jobs until context cancellation.
func (p *Processor) Run(ctx context.Context) error {
	ticker := time.NewTicker(p.poll)
	defer ticker.Stop()

	for {
		processed, err := p.processNext(ctx)
		if err != nil {
			p.logger.Error("RAG ingestion iteration failed", zap.Error(err))
		}

		if processed {
			continue
		}

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (p *Processor) processNext(ctx context.Context) (bool, error) {
	if !time.Now().Before(p.nextConversationPurge) {
		if err := p.store.PurgeExpiredConversations(ctx); err != nil {
			return false, err
		}

		p.nextConversationPurge = time.Now().Add(time.Minute)
	}

	purge, claimErr := p.store.ClaimPurgeJob(ctx, p.owner, p.lease)
	if claimErr != nil {
		return false, claimErr
	}

	if purge != nil {
		deleteErr := p.executePurge(ctx, *purge)
		if deleteErr != nil {
			if failErr := p.store.FailPurge(
				context.WithoutCancel(ctx), *purge, p.owner, classifyPurgeError(*purge),
			); failErr != nil {
				if errors.Is(failErr, store.ErrLeaseLost) {
					return true, nil
				}

				return true, errors.Join(deleteErr, failErr)
			}

			return true, nil
		}

		completeErr := p.store.CompletePurge(ctx, purge.ID, purge.TenantID, p.owner)
		if completeErr != nil {
			if errors.Is(completeErr, store.ErrLeaseLost) {
				return true, nil
			}

			return true, completeErr
		}

		return true, nil
	}

	job, err := p.store.ClaimIngestionJob(ctx, p.owner, p.lease)
	if err != nil {
		return false, err
	}

	if job == nil {
		return false, nil
	}

	if err := p.processWithLease(ctx, *job); err != nil {
		if errors.Is(err, store.ErrLeaseLost) {
			p.logger.Warn("RAG ingestion lease lost", zap.String("job_id", job.ID), zap.String("tenant_id", job.TenantID))

			return true, nil
		}

		code := classifyError(err)
		if failErr := p.store.FailIngestion(context.WithoutCancel(ctx), *job, p.owner, code, err); failErr != nil {
			if errors.Is(failErr, store.ErrLeaseLost) {
				return true, nil
			}

			return true, errors.Join(err, failErr)
		}

		p.logger.Warn("RAG ingestion job failed",
			zap.String("job_id", job.ID), zap.String("tenant_id", job.TenantID),
			zap.String("error_code", code), zap.Int("attempt", job.Attempt))

		return true, nil
	}

	return true, nil
}

func (p *Processor) executePurge(ctx context.Context, job store.PurgeJob) error {
	scope, identifier, vectorPurge := store.ParseVectorPurgeObjectKey(job.ObjectKey)
	if vectorPurge {
		switch scope {
		case "tenant":
			return p.store.PurgeTenantVectors(ctx, job.TenantID)
		case "knowledge_base":
			return p.store.PurgeKnowledgeBaseVectors(ctx, job.TenantID, identifier)
		case "document":
			return p.store.PurgeDocumentVectors(ctx, job.TenantID, identifier)
		}
	}

	return p.objects.Delete(ctx, job.ObjectKey)
}

func classifyPurgeError(job store.PurgeJob) string {
	if _, _, vectorPurge := store.ParseVectorPurgeObjectKey(job.ObjectKey); vectorPurge {
		return "vector_delete_failed"
	}

	return "object_delete_failed"
}

func (p *Processor) processWithLease(ctx context.Context, job store.IngestionJob) error {
	processCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	heartbeatErr := make(chan error, 1)
	go func() {
		heartbeatErr <- p.maintainLease(processCtx, cancel, job)
	}()

	processErr := p.process(processCtx, job)

	cancel()

	if err := <-heartbeatErr; err != nil {
		return err
	}

	return processErr
}

func (p *Processor) maintainLease(
	ctx context.Context,
	cancel context.CancelFunc,
	job store.IngestionJob,
) error {
	ticker := time.NewTicker(p.lease / 3)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if ctx.Err() != nil {
				return nil
			}

			if err := p.store.RenewIngestionLease(ctx, job.ID, job.TenantID, p.owner, p.lease); err != nil {
				if ctx.Err() != nil {
					return nil
				}

				cancel()

				return err
			}
		}
	}
}

func (p *Processor) process(ctx context.Context, job store.IngestionJob) error {
	if err := p.updateStage(ctx, job, "scanning"); err != nil {
		return err
	}

	if err := p.updateStage(ctx, job, "parsing"); err != nil {
		return err
	}

	document, err := p.loadDocument(ctx, job)
	if err != nil {
		return err
	}

	if stageErr := p.updateStage(ctx, job, "chunking"); stageErr != nil {
		return stageErr
	}

	chunks, err := domain.ChunkDocument(
		document, job.Profile.ChunkTokens, job.Profile.ChunkOverlapTokens, p.tokenizer,
	)
	if err != nil {
		return fmt.Errorf("chunk document: %w", err)
	}

	if len(chunks) == 0 {
		return errors.New("empty_document")
	}

	if err := p.updateStage(ctx, job, "embedding"); err != nil {
		return err
	}

	vectors := make([][]float32, 0, len(chunks))
	for start := 0; start < len(chunks); start += embeddingBatchSize {
		end := min(start+embeddingBatchSize, len(chunks))

		inputs := make([]string, end-start)
		for index := start; index < end; index++ {
			inputs[index-start] = chunks[index].IndexText
		}

		batch, embeddingTokens, err := p.embedder.Embed(
			ctx, job.TenantSlug, job.Profile.EmbeddingModel, inputs, job.Profile.EmbeddingDimensions,
		)
		if err != nil {
			return fmt.Errorf("embed document chunks: %w", err)
		}

		if err := p.store.RecordUsage(context.WithoutCancel(ctx), store.UsageEvent{
			TenantID: job.TenantID, RequestID: fmt.Sprintf("ingest:%s:%d:%d", job.ID, job.Attempt, start),
			KnowledgeBaseIDs: []string{job.KnowledgeBaseID}, Operation: "ingestion_embedding",
			Status: "completed", EmbeddingTokens: embeddingTokens,
		}); err != nil {
			return fmt.Errorf("record ingestion embedding usage: %w", err)
		}

		if len(batch) != len(inputs) {
			return errors.New("embedding_count_mismatch")
		}

		vectors = append(vectors, batch...)
	}

	if err := p.updateStage(ctx, job, "indexing"); err != nil {
		return err
	}

	vectorChunks := make([]store.VectorChunk, len(chunks))
	for index := range chunks {
		vectorChunks[index] = store.VectorChunk{
			TenantID: job.TenantID, KnowledgeBaseID: job.KnowledgeBaseID,
			DocumentID: job.DocumentID, VersionID: job.VersionID, Title: job.Title,
			Chunk: chunks[index], Embedding: vectors[index],
			NormalizationRevision: job.Profile.Configuration.NormalizationRevision,
		}
	}

	if err := p.store.ReplaceVersionChunks(ctx, vectorChunks); err != nil {
		return err
	}

	return p.store.CompleteIngestion(ctx, job, p.owner, document.PageCount, len(chunks), document.Language)
}

func (p *Processor) updateStage(ctx context.Context, job store.IngestionJob, stage string) error {
	return p.store.UpdateIngestionStage(
		ctx, job.ID, job.TenantID, job.VersionID, stage, p.owner, p.lease,
	)
}

func (p *Processor) loadDocument(
	ctx context.Context,
	job store.IngestionJob,
) (document domain.SourceDocument, err error) {
	if job.SourceType == "text" {
		if job.SourceText == nil {
			return domain.SourceDocument{}, errors.New("text_source_missing")
		}

		return domain.PlainTextDocument(*job.SourceText), nil
	}

	if job.ObjectKey == nil {
		return domain.SourceDocument{}, errors.New("object_key_missing")
	}

	reader, err := p.objects.Get(ctx, *job.ObjectKey)
	if err != nil {
		return domain.SourceDocument{}, err
	}
	defer func() {
		closeErr := reader.Close()
		if closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close private source object: %w", closeErr))
		}
	}()

	source, err := io.ReadAll(io.LimitReader(reader, maxSourceBytes+1))
	if err != nil {
		return domain.SourceDocument{}, fmt.Errorf("read private source object: %w", err)
	}

	if len(source) > maxSourceBytes {
		return domain.SourceDocument{}, errors.New("source object exceeds 25 MiB limit")
	}

	document, extractionErr := p.parser.Extract(ctx, job.MediaType, bytes.NewReader(source))
	if extractionErr == nil {
		return document, nil
	}

	if !requiresOCR(job.MediaType, extractionErr) {
		return domain.SourceDocument{}, extractionErr
	}

	document, pages, ocrErr := p.ocr.OCR(
		ctx, job.TenantSlug, job.Profile.Configuration.OCRModel, job.MediaType, source,
	)

	status := "completed"
	if ocrErr != nil {
		status = "failed"
	}

	usageErr := p.store.RecordUsage(context.WithoutCancel(ctx), store.UsageEvent{
		TenantID: job.TenantID, RequestID: fmt.Sprintf("ingest:%s:%d:ocr", job.ID, job.Attempt),
		KnowledgeBaseIDs: []string{job.KnowledgeBaseID}, Operation: "ingestion_ocr",
		Status: status, OCRPages: pages,
	})
	if ocrErr != nil {
		return domain.SourceDocument{}, errors.Join(fmt.Errorf("OCR unavailable: %w", ocrErr), usageErr)
	}

	if usageErr != nil {
		return domain.SourceDocument{}, fmt.Errorf("record ingestion OCR usage: %w", usageErr)
	}

	return document, nil
}

func requiresOCR(mediaType string, err error) bool {
	if !strings.EqualFold(strings.TrimSpace(mediaType), "application/pdf") {
		return false
	}

	message := strings.ToLower(err.Error())

	return strings.Contains(message, "ocr_required") || strings.Contains(message, "low_extraction_quality")
}

func classifyError(err error) string {
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "ocr_required"):
		return "ocr_required"
	case strings.Contains(message, "ocr"):
		return "ocr_failed"
	case strings.Contains(message, "clam"):
		return "scan_failed"
	case strings.Contains(message, "tika"):
		return "parse_failed"
	case strings.Contains(message, "embed"):
		return "embedding_failed"
	case strings.Contains(message, "vector"), strings.Contains(message, "index"):
		return "indexing_failed"
	default:
		return "ingestion_failed"
	}
}
