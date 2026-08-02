package supportfiles

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

const (
	MaxFilesPerMessage   = 5
	MaxFileBytes         = int64(10 * 1024 * 1024)
	MaxMessageBytes      = int64(25 * 1024 * 1024)
	MaxReceiptImageBytes = int64(2 * 1024 * 1024)
)

var (
	ErrNotConfigured  = errors.New("support attachment storage is not configured")
	ErrTooManyFiles   = errors.New("too many support attachments")
	ErrFileTooLarge   = errors.New("support attachment too large")
	ErrTypeNotAllowed = errors.New("support attachment type not allowed")
	ErrScanFailed     = errors.New("support attachment scan failed")
	ErrInfected       = errors.New("support attachment infected")
	ErrInvalidReceipt = errors.New("invalid bank transfer receipt")
)

type ObjectStore interface {
	Put(context.Context, string, io.Reader, int64, string) error
	Get(context.Context, string) (io.ReadCloser, error)
	Delete(context.Context, string) error
}

type ObjectInfo struct {
	Key          string
	LastModified time.Time
}

type ObjectLister interface {
	List(context.Context, string) ([]ObjectInfo, error)
}

type Scanner interface {
	Scan(context.Context, io.Reader) (clean bool, result string, err error)
}

type Service struct {
	store   ObjectStore
	scanner Scanner
}

type ObjectReader struct {
	Body io.ReadCloser
}

type ReceiptUpload struct {
	FileName    string
	ObjectKey   string
	ContentType string
	SizeBytes   int64
}

func New(store ObjectStore, scanner Scanner) *Service {
	return &Service{store: store, scanner: scanner}
}

func NewFromEnv() (*Service, error) {
	endpoint := firstEnv("PRIVATE_FILES_S3_ENDPOINT", "SUPPORT_S3_ENDPOINT")

	bucket := firstEnv("PRIVATE_FILES_S3_BUCKET", "SUPPORT_S3_BUCKET")
	if endpoint == "" || bucket == "" {
		return nil, nil
	}

	accessKey := firstEnv("PRIVATE_FILES_S3_ACCESS_KEY", "SUPPORT_S3_ACCESS_KEY")

	secretKey := firstEnv("PRIVATE_FILES_S3_SECRET_KEY", "SUPPORT_S3_SECRET_KEY")
	if accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("%w: private object-storage credentials are required", ErrNotConfigured)
	}

	region := firstEnv("PRIVATE_FILES_S3_REGION", "SUPPORT_S3_REGION")
	if region == "" {
		region = "us-east-1"
	}

	client := s3.New(s3.Options{
		BaseEndpoint: aws.String(endpoint),
		Region:       region,
		Credentials:  credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		UsePathStyle: boolEnvFallback("PRIVATE_FILES_S3_FORCE_PATH_STYLE", "SUPPORT_S3_FORCE_PATH_STYLE", true),
	})

	var scanner Scanner
	if address := strings.TrimSpace(os.Getenv("SUPPORT_CLAMAV_ADDRESS")); address != "" {
		scanner = &ClamAVScanner{
			Address: address,
			Timeout: durationEnv("SUPPORT_CLAMAV_TIMEOUT", 30*time.Second),
		}
	} else if boolEnv("SUPPORT_ATTACHMENT_ALLOW_UNSCANNED_DEV", false) {
		scanner = DevelopmentScanner{}
	} else {
		scanner = UnavailableScanner{}
	}

	return New(&S3ObjectStore{Client: client, Bucket: bucket}, scanner), nil
}

func (s *Service) UploadFiles(
	ctx context.Context,
	tenantID string,
	scope string,
	files []*multipart.FileHeader,
) ([]platform.SupportTicketAttachmentUpload, error) {
	if len(files) == 0 {
		return nil, nil
	}

	if s == nil || s.store == nil || s.scanner == nil {
		return nil, ErrNotConfigured
	}

	if len(files) > MaxFilesPerMessage {
		return nil, ErrTooManyFiles
	}

	var total int64

	for _, file := range files {
		if file == nil || file.Size <= 0 {
			return nil, ErrTypeNotAllowed
		}

		if file.Size > MaxFileBytes {
			return nil, ErrFileTooLarge
		}

		total += file.Size
		if total > MaxMessageBytes {
			return nil, ErrFileTooLarge
		}
	}

	uploads := make([]platform.SupportTicketAttachmentUpload, 0, len(files))
	for _, header := range files {
		item, err := s.uploadOne(ctx, tenantID, scope, header)
		if err != nil {
			s.DeleteUploads(context.WithoutCancel(ctx), uploads)
			return nil, err
		}

		uploads = append(uploads, item)
	}

	return uploads, nil
}

func (s *Service) uploadOne(
	ctx context.Context,
	tenantID string,
	scope string,
	header *multipart.FileHeader,
) (platform.SupportTicketAttachmentUpload, error) {
	originalFilename := sanitizeFilename(header.Filename)

	detectedType, err := validateFileType(header, originalFilename)
	if err != nil {
		return platform.SupportTicketAttachmentUpload{}, err
	}

	hashValue, err := hashFile(header)
	if err != nil {
		return platform.SupportTicketAttachmentUpload{}, fmt.Errorf("hash support attachment: %w", err)
	}

	attachmentID := uuid.NewString()
	objectKey := fmt.Sprintf(
		"support/%s/%s/%s",
		safePathSegment(tenantID),
		safePathSegment(scope),
		attachmentID,
	)

	// Persist into the private bucket before scanning. The object has no
	// downloadable metadata row yet and is therefore quarantined/fail-closed.
	uploadReader, err := header.Open()
	if err != nil {
		return platform.SupportTicketAttachmentUpload{}, fmt.Errorf("open support attachment for upload: %w", err)
	}

	uploadErr := s.store.Put(ctx, objectKey, uploadReader, header.Size, "application/octet-stream")
	closeErr := uploadReader.Close()

	if uploadErr != nil {
		return platform.SupportTicketAttachmentUpload{}, fmt.Errorf("store quarantined support attachment: %w", uploadErr)
	}

	if closeErr != nil {
		_ = s.store.Delete(context.WithoutCancel(ctx), objectKey) //nolint:errcheck // Failure is intentionally handled by the surrounding best-effort path.
		return platform.SupportTicketAttachmentUpload{}, fmt.Errorf("close support attachment after upload: %w", closeErr)
	}

	scanReader, err := header.Open()
	if err != nil {
		_ = s.store.Delete(context.WithoutCancel(ctx), objectKey) //nolint:errcheck // Failure is intentionally handled by the surrounding best-effort path.
		return platform.SupportTicketAttachmentUpload{}, fmt.Errorf("open support attachment for scan: %w", err)
	}

	clean, scanResult, scanErr := s.scanner.Scan(ctx, io.LimitReader(scanReader, MaxFileBytes+1))
	closeErr = scanReader.Close()

	if scanErr != nil {
		_ = s.store.Delete(context.WithoutCancel(ctx), objectKey) //nolint:errcheck // Failure is intentionally handled by the surrounding best-effort path.
		return platform.SupportTicketAttachmentUpload{}, fmt.Errorf("%w: %v", ErrScanFailed, scanErr)
	}

	if closeErr != nil {
		_ = s.store.Delete(context.WithoutCancel(ctx), objectKey) //nolint:errcheck // Failure is intentionally handled by the surrounding best-effort path.
		return platform.SupportTicketAttachmentUpload{}, fmt.Errorf("close support attachment after scan: %w", closeErr)
	}

	if !clean {
		_ = s.store.Delete(context.WithoutCancel(ctx), objectKey) //nolint:errcheck // Failure is intentionally handled by the surrounding best-effort path.
		return platform.SupportTicketAttachmentUpload{}, fmt.Errorf("%w: %s", ErrInfected, scanResult)
	}

	scannedAt := time.Now().UTC()

	return platform.SupportTicketAttachmentUpload{
		ID:                  attachmentID,
		OriginalFilename:    originalFilename,
		DetectedContentType: detectedType,
		SizeBytes:           header.Size,
		SHA256:              hashValue,
		ObjectKey:           objectKey,
		ScanStatus:          "clean",
		ScanResult:          sanitizeScanResult(scanResult),
		ScannedAt:           &scannedAt,
	}, nil
}

func (s *Service) Get(ctx context.Context, key string) (*ObjectReader, error) {
	if s == nil || s.store == nil {
		return nil, ErrNotConfigured
	}

	body, err := s.store.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	return &ObjectReader{Body: body}, nil
}

func (s *Service) UploadReceiptImage(ctx context.Context, tenantID string, header *multipart.FileHeader) (*ReceiptUpload, error) {
	if s == nil || s.store == nil || s.scanner == nil {
		return nil, ErrNotConfigured
	}
	if header == nil || header.Size <= 0 {
		return nil, ErrInvalidReceipt
	}
	if header.Size > MaxReceiptImageBytes {
		return nil, ErrFileTooLarge
	}

	originalFilename := sanitizeFilename(header.Filename)
	contentType, err := validateFileType(header, originalFilename)
	if err != nil || !isReceiptImageType(contentType) {
		return nil, ErrInvalidReceipt
	}

	extension := receiptExtension(contentType)
	fileName := uuid.NewString() + extension
	objectKey := BankTransferReceiptObjectKey(tenantID, fileName)

	uploadReader, err := header.Open()
	if err != nil {
		return nil, fmt.Errorf("open bank transfer receipt for upload: %w", err)
	}
	uploadErr := s.store.Put(ctx, objectKey, uploadReader, header.Size, contentType)
	closeErr := uploadReader.Close()
	if uploadErr != nil {
		return nil, fmt.Errorf("store bank transfer receipt: %w", uploadErr)
	}
	if closeErr != nil {
		_ = s.store.Delete(context.WithoutCancel(ctx), objectKey) //nolint:errcheck // Compensating cleanup is best-effort.
		return nil, fmt.Errorf("close bank transfer receipt after upload: %w", closeErr)
	}

	scanReader, err := header.Open()
	if err != nil {
		_ = s.store.Delete(context.WithoutCancel(ctx), objectKey) //nolint:errcheck // Compensating cleanup is best-effort.
		return nil, fmt.Errorf("open bank transfer receipt for scan: %w", err)
	}
	clean, scanResult, scanErr := s.scanner.Scan(ctx, io.LimitReader(scanReader, MaxReceiptImageBytes+1))
	closeErr = scanReader.Close()
	if scanErr != nil {
		_ = s.store.Delete(context.WithoutCancel(ctx), objectKey) //nolint:errcheck // Compensating cleanup is best-effort.
		return nil, fmt.Errorf("%w: %v", ErrScanFailed, scanErr)
	}
	if closeErr != nil {
		_ = s.store.Delete(context.WithoutCancel(ctx), objectKey) //nolint:errcheck // Compensating cleanup is best-effort.
		return nil, fmt.Errorf("close bank transfer receipt after scan: %w", closeErr)
	}
	if !clean {
		_ = s.store.Delete(context.WithoutCancel(ctx), objectKey) //nolint:errcheck // Compensating cleanup is best-effort.
		return nil, fmt.Errorf("%w: %s", ErrInfected, scanResult)
	}

	return &ReceiptUpload{
		FileName:    fileName,
		ObjectKey:   objectKey,
		ContentType: contentType,
		SizeBytes:   header.Size,
	}, nil
}

func (s *Service) DeleteObject(ctx context.Context, key string) {
	if s == nil || s.store == nil || strings.TrimSpace(key) == "" {
		return
	}
	_ = s.store.Delete(ctx, key) //nolint:errcheck // Compensating cleanup is best-effort.
}

func BankTransferReceiptObjectKey(tenantID, fileName string) string {
	return fmt.Sprintf("billing/receipts/%s/%s", safePathSegment(tenantID), fileName)
}

func ValidateBankTransferReceiptFileName(fileName string) (string, bool) {
	fileName = strings.TrimSpace(fileName)
	extension := strings.ToLower(filepath.Ext(fileName))
	contentTypes := map[string]string{
		".jpg":  "image/jpeg",
		".png":  "image/png",
		".webp": "image/webp",
		".gif":  "image/gif",
	}
	contentType, ok := contentTypes[extension]
	if !ok {
		return "", false
	}
	if _, err := uuid.Parse(strings.TrimSuffix(fileName, extension)); err != nil {
		return "", false
	}
	return contentType, true
}

func (s *Service) DeleteUploads(ctx context.Context, uploads []platform.SupportTicketAttachmentUpload) {
	if s == nil || s.store == nil {
		return
	}

	for _, item := range uploads {
		_ = s.store.Delete(ctx, item.ObjectKey) //nolint:errcheck // Failure is intentionally handled by the surrounding best-effort path.
	}
}

// CleanupOrphans removes support objects that are older than the safety cutoff
// and have no attachment metadata row. Repeated runs are safe.
func (s *Service) CleanupOrphans(ctx context.Context, referenced map[string]struct{}, olderThan time.Time) (int, error) {
	if s == nil || s.store == nil {
		return 0, ErrNotConfigured
	}

	lister, ok := s.store.(ObjectLister)
	if !ok {
		return 0, fmt.Errorf("%w: object listing is unavailable", ErrNotConfigured)
	}

	objects, err := lister.List(ctx, "support/")
	if err != nil {
		return 0, err
	}

	deleted := 0

	for _, object := range objects {
		if !object.LastModified.Before(olderThan) {
			continue
		}

		if _, exists := referenced[object.Key]; exists {
			continue
		}

		if err := s.store.Delete(ctx, object.Key); err != nil {
			return deleted, err
		}

		deleted++
	}

	return deleted, nil
}

func validateFileType(header *multipart.FileHeader, filename string) (string, error) {
	extension := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	allowed := map[string][]string{
		"pdf":  {"application/pdf"},
		"png":  {"image/png"},
		"jpg":  {"image/jpeg"},
		"jpeg": {"image/jpeg"},
		"webp": {"image/webp"},
		"gif":  {"image/gif"},
		"txt":  {"text/plain"},
		"csv":  {"text/plain", "text/csv", "application/csv"},
		"log":  {"text/plain", "application/octet-stream"},
		"json": {"application/json", "text/plain"},
	}

	allowedTypes, ok := allowed[extension]
	if !ok {
		return "", ErrTypeNotAllowed
	}

	reader, err := header.Open()
	if err != nil {
		return "", err
	}
	defer reader.Close() //nolint:errcheck // Cleanup is best-effort and cannot change the completed operation.

	prefix := make([]byte, 512)

	read, err := io.ReadFull(reader, prefix)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return "", err
	}

	prefix = prefix[:read]
	detected := strings.ToLower(strings.TrimSpace(strings.Split(http.DetectContentType(prefix), ";")[0]))

	if extension == "json" {
		jsonReader, openErr := header.Open()
		if openErr != nil {
			return "", openErr
		}

		content, readErr := io.ReadAll(io.LimitReader(jsonReader, MaxFileBytes+1))
		closeErr := jsonReader.Close()

		if readErr != nil {
			return "", readErr
		}

		if closeErr != nil {
			return "", closeErr
		}

		if int64(len(content)) > MaxFileBytes {
			return "", ErrFileTooLarge
		}

		if !utf8.Valid(content) || !json.Valid(content) {
			return "", ErrTypeNotAllowed
		}

		return "application/json", nil
	}

	if extension == "csv" || extension == "txt" || extension == "log" {
		if !utf8.Valid(prefix) || bytesContainNUL(prefix) {
			return "", ErrTypeNotAllowed
		}

		if extension == "csv" {
			return "text/csv", nil
		}

		return "text/plain", nil
	}

	for _, allowedType := range allowedTypes {
		if detected == allowedType {
			return detected, nil
		}
	}

	return "", ErrTypeNotAllowed
}

func hashFile(header *multipart.FileHeader) (string, error) {
	reader, err := header.Open()
	if err != nil {
		return "", err
	}
	defer reader.Close() //nolint:errcheck // Cleanup is best-effort and cannot change the completed operation.

	hash := sha256.New()

	written, err := io.Copy(hash, io.LimitReader(reader, MaxFileBytes+1))
	if err != nil {
		return "", err
	}

	if written > MaxFileBytes {
		return "", ErrFileTooLarge
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func sanitizeFilename(value string) string {
	value = filepath.Base(strings.TrimSpace(value))

	value = strings.Map(func(r rune) rune {
		switch {
		case r == 0, r == '\r', r == '\n', r == '/', r == '\\':
			return '_'
		case r < 32:
			return -1
		default:
			return r
		}
	}, value)
	if value == "" || value == "." {
		return "attachment"
	}

	runes := []rune(value)
	if len(runes) > 200 {
		value = string(runes[:200])
	}

	return value
}

func safePathSegment(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unknown"
	}

	var builder strings.Builder

	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			builder.WriteRune(r)
		}
	}

	if builder.Len() == 0 {
		return "unknown"
	}

	return builder.String()
}

func sanitizeScanResult(value string) string {
	value = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(value, "\r", " "), "\n", " "))
	if len(value) > 240 {
		return value[:240]
	}

	return value
}

func bytesContainNUL(value []byte) bool {
	for _, item := range value {
		if item == 0 {
			return true
		}
	}

	return false
}

func isReceiptImageType(contentType string) bool {
	switch contentType {
	case "image/jpeg", "image/png", "image/webp", "image/gif":
		return true
	default:
		return false
	}
}

func receiptExtension(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ".bin"
	}
}

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func boolEnvFallback(primary, legacy string, fallback bool) bool {
	if strings.TrimSpace(os.Getenv(primary)) != "" {
		return boolEnv(primary, fallback)
	}
	return boolEnv(legacy, fallback)
}

func boolEnv(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}

	return value
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	value, err := time.ParseDuration(raw)
	if err != nil || value <= 0 {
		return fallback
	}

	return value
}

type S3ObjectStore struct {
	Client *s3.Client
	Bucket string
}

func (s *S3ObjectStore) Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	_, err := s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.Bucket),
		Key:           aws.String(key),
		Body:          body,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	})

	return err
}

func (s *S3ObjectStore) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	result, err := s.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	return result.Body, nil
}

func (s *S3ObjectStore) Delete(ctx context.Context, key string) error {
	_, err := s.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
	})

	return err
}

func (s *S3ObjectStore) List(ctx context.Context, prefix string) ([]ObjectInfo, error) {
	paginator := s3.NewListObjectsV2Paginator(s.Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.Bucket),
		Prefix: aws.String(prefix),
	})
	objects := make([]ObjectInfo, 0)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, item := range page.Contents {
			if item.Key == nil || item.LastModified == nil {
				continue
			}

			objects = append(objects, ObjectInfo{Key: *item.Key, LastModified: *item.LastModified})
		}
	}

	return objects, nil
}

type ClamAVScanner struct {
	Address string
	Timeout time.Duration
}

//nolint:govet // Short-lived branch errors are checked immediately and intentionally scoped with short declarations.
func (s *ClamAVScanner) Scan(ctx context.Context, reader io.Reader) (bool, string, error) {
	dialer := net.Dialer{Timeout: s.Timeout}

	conn, err := dialer.DialContext(ctx, "tcp", s.Address)
	if err != nil {
		return false, "", err
	}
	defer conn.Close() //nolint:errcheck // Cleanup is best-effort and cannot change the completed operation.

	_ = conn.SetDeadline(time.Now().Add(s.Timeout)) //nolint:errcheck // Cleanup is best-effort and cannot change the completed operation.

	if _, err := conn.Write([]byte{'z', 'I', 'N', 'S', 'T', 'R', 'E', 'A', 'M', 0}); err != nil {
		return false, "", err
	}

	buffer := make([]byte, 32*1024)

	var total int64

	for {
		read, readErr := reader.Read(buffer)
		if read > 0 {
			total += int64(read)
			if total > MaxFileBytes {
				return false, "", ErrFileTooLarge
			}

			length := make([]byte, 4)
			binary.BigEndian.PutUint32(length, uint32(read))

			if _, err := conn.Write(length); err != nil {
				return false, "", err
			}

			if _, err := conn.Write(buffer[:read]); err != nil {
				return false, "", err
			}
		}

		if errors.Is(readErr, io.EOF) {
			break
		}

		if readErr != nil {
			return false, "", readErr
		}
	}

	if _, err := conn.Write([]byte{0, 0, 0, 0}); err != nil {
		return false, "", err
	}

	response, err := bufio.NewReader(conn).ReadString(0)
	if err != nil && !errors.Is(err, io.EOF) {
		return false, "", err
	}

	response = strings.TrimSpace(strings.TrimSuffix(response, "\x00"))
	switch {
	case strings.HasSuffix(response, "OK"):
		return true, "clean", nil
	case strings.Contains(response, "FOUND"):
		return false, response, nil
	default:
		return false, response, fmt.Errorf("unexpected ClamAV response: %s", response)
	}
}

type DevelopmentScanner struct{}

func (DevelopmentScanner) Scan(_ context.Context, _ io.Reader) (bool, string, error) {
	return true, "development-bypass", nil
}

type UnavailableScanner struct{}

func (UnavailableScanner) Scan(_ context.Context, _ io.Reader) (bool, string, error) {
	return false, "", ErrNotConfigured
}
