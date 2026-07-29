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
	MaxFilesPerMessage = 5
	MaxFileBytes       = int64(10 * 1024 * 1024)
	MaxMessageBytes    = int64(25 * 1024 * 1024)
)

var (
	ErrNotConfigured  = errors.New("support attachment storage is not configured")
	ErrTooManyFiles   = errors.New("too many support attachments")
	ErrFileTooLarge   = errors.New("support attachment too large")
	ErrTypeNotAllowed = errors.New("support attachment type not allowed")
	ErrScanFailed     = errors.New("support attachment scan failed")
	ErrInfected       = errors.New("support attachment infected")
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

func New(store ObjectStore, scanner Scanner) *Service {
	return &Service{store: store, scanner: scanner}
}

func NewFromEnv() (*Service, error) {
	endpoint := strings.TrimSpace(os.Getenv("SUPPORT_S3_ENDPOINT"))
	bucket := strings.TrimSpace(os.Getenv("SUPPORT_S3_BUCKET"))
	if endpoint == "" || bucket == "" {
		return nil, nil
	}

	accessKey := strings.TrimSpace(os.Getenv("SUPPORT_S3_ACCESS_KEY"))
	secretKey := strings.TrimSpace(os.Getenv("SUPPORT_S3_SECRET_KEY"))
	if accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("%w: SUPPORT_S3_ACCESS_KEY and SUPPORT_S3_SECRET_KEY are required", ErrNotConfigured)
	}

	region := strings.TrimSpace(os.Getenv("SUPPORT_S3_REGION"))
	if region == "" {
		region = "us-east-1"
	}
	client := s3.New(s3.Options{
		BaseEndpoint: aws.String(endpoint),
		Region:       region,
		Credentials:  credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		UsePathStyle: boolEnv("SUPPORT_S3_FORCE_PATH_STYLE", true),
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
		_ = s.store.Delete(context.WithoutCancel(ctx), objectKey)
		return platform.SupportTicketAttachmentUpload{}, fmt.Errorf("close support attachment after upload: %w", closeErr)
	}

	scanReader, err := header.Open()
	if err != nil {
		_ = s.store.Delete(context.WithoutCancel(ctx), objectKey)
		return platform.SupportTicketAttachmentUpload{}, fmt.Errorf("open support attachment for scan: %w", err)
	}
	clean, scanResult, scanErr := s.scanner.Scan(ctx, io.LimitReader(scanReader, MaxFileBytes+1))
	closeErr = scanReader.Close()
	if scanErr != nil {
		_ = s.store.Delete(context.WithoutCancel(ctx), objectKey)
		return platform.SupportTicketAttachmentUpload{}, fmt.Errorf("%w: %v", ErrScanFailed, scanErr)
	}
	if closeErr != nil {
		_ = s.store.Delete(context.WithoutCancel(ctx), objectKey)
		return platform.SupportTicketAttachmentUpload{}, fmt.Errorf("close support attachment after scan: %w", closeErr)
	}
	if !clean {
		_ = s.store.Delete(context.WithoutCancel(ctx), objectKey)
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

func (s *Service) DeleteUploads(ctx context.Context, uploads []platform.SupportTicketAttachmentUpload) {
	if s == nil || s.store == nil {
		return
	}
	for _, item := range uploads {
		_ = s.store.Delete(ctx, item.ObjectKey)
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
	defer reader.Close()

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
	defer reader.Close()

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

func (s *ClamAVScanner) Scan(ctx context.Context, reader io.Reader) (bool, string, error) {
	dialer := net.Dialer{Timeout: s.Timeout}
	conn, err := dialer.DialContext(ctx, "tcp", s.Address)
	if err != nil {
		return false, "", err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(s.Timeout))

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
