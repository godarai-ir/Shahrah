// Package files owns private RAG objects, antivirus scanning, and safe file validation.
package files

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

const (
	// MaxFileBytes is the public MVP upload limit.
	MaxFileBytes     = int64(25 * 1024 * 1024)
	maxExpandedBytes = uint64(100 * 1024 * 1024)
)

var (
	// ErrFileRejected indicates invalid size, MIME, archive, or text encoding.
	ErrFileRejected = errors.New("RAG file rejected")
	// ErrInfected indicates that ClamAV found malware.
	ErrInfected = errors.New("RAG file contains malware")
)

// Config configures the dedicated private RAG bucket and fail-closed scanner.
type Config struct {
	Endpoint       string
	Bucket         string
	AccessKey      string
	SecretKey      string
	Region         string
	ForcePathStyle bool
	ClamAVAddress  string
	ClamAVTimeout  time.Duration
}

// Service stores files in a tenant-prefixed private bucket.
type Service struct {
	client        *s3.Client
	bucket        string
	clamAVAddress string
	clamAVTimeout time.Duration
}

// Upload contains validated object metadata safe to persist.
type Upload struct {
	Title     string
	ObjectKey string
	MediaType string
	SHA256    string
	SizeBytes int64
}

// New creates a fail-closed private file service.
func New(config Config) (*Service, error) {
	if strings.TrimSpace(config.Endpoint) == "" || strings.TrimSpace(config.Bucket) == "" ||
		strings.TrimSpace(config.AccessKey) == "" || strings.TrimSpace(config.SecretKey) == "" ||
		strings.TrimSpace(config.ClamAVAddress) == "" {
		return nil, errors.New("RAG object storage and ClamAV configuration are required")
	}

	if config.Region == "" {
		config.Region = "us-east-1"
	}

	if config.ClamAVTimeout <= 0 {
		config.ClamAVTimeout = 30 * time.Second
	}

	client := s3.New(s3.Options{
		BaseEndpoint: aws.String(config.Endpoint),
		Region:       config.Region,
		Credentials:  credentials.NewStaticCredentialsProvider(config.AccessKey, config.SecretKey, ""),
		UsePathStyle: config.ForcePathStyle,
	})

	return &Service{
		client: client, bucket: config.Bucket,
		clamAVAddress: config.ClamAVAddress, clamAVTimeout: config.ClamAVTimeout,
	}, nil
}

// ValidateAndPut scans and stores one supported file.
func (s *Service) ValidateAndPut(ctx context.Context, tenantID, filename string, content []byte) (*Upload, error) {
	if len(content) == 0 || int64(len(content)) > MaxFileBytes {
		return nil, ErrFileRejected
	}

	title := sanitizeFilename(filename)

	mediaType, err := validateType(title, content)
	if err != nil {
		return nil, err
	}

	clean, err := scanClamAV(ctx, s.clamAVAddress, s.clamAVTimeout, content)
	if err != nil {
		return nil, fmt.Errorf("scan RAG file: %w", err)
	}

	if !clean {
		return nil, ErrInfected
	}

	hash := sha256.Sum256(content)
	objectKey := fmt.Sprintf("rag/%s/%s", safeSegment(tenantID), uuid.NewString())

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(objectKey),
		Body:          bytes.NewReader(content),
		ContentLength: aws.Int64(int64(len(content))),
		ContentType:   aws.String(mediaType),
	})
	if err != nil {
		return nil, fmt.Errorf("put private RAG object: %w", err)
	}

	return &Upload{
		Title: title, ObjectKey: objectKey, MediaType: mediaType,
		SHA256: hex.EncodeToString(hash[:]), SizeBytes: int64(len(content)),
	}, nil
}

// Get opens a private object for worker parsing.
func (s *Service) Get(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket), Key: aws.String(objectKey),
	})
	if err != nil {
		return nil, fmt.Errorf("get private RAG object: %w", err)
	}

	return result.Body, nil
}

// Delete permanently removes a private object.
func (s *Service) Delete(ctx context.Context, objectKey string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket), Key: aws.String(objectKey),
	})
	if err != nil {
		return fmt.Errorf("delete private RAG object: %w", err)
	}

	return nil
}

func validateType(filename string, content []byte) (string, error) {
	extension := strings.ToLower(filepath.Ext(filename))
	detected := strings.ToLower(strings.TrimSpace(strings.Split(http.DetectContentType(content[:min(512, len(content))]), ";")[0]))

	switch extension {
	case ".pdf":
		if detected != "application/pdf" {
			return "", ErrFileRejected
		}

		return "application/pdf", nil
	case ".docx":
		if detected != "application/zip" && detected != "application/octet-stream" {
			return "", ErrFileRejected
		}

		if err := validateDOCXArchive(content); err != nil {
			return "", err
		}

		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document", nil
	case ".txt", ".html", ".htm":
		if !utf8.Valid(content) || bytes.IndexByte(content, 0) >= 0 {
			return "", ErrFileRejected
		}

		if extension == ".txt" {
			return "text/plain", nil
		}

		return "text/html", nil
	default:
		return "", ErrFileRejected
	}
}

func validateDOCXArchive(content []byte) error {
	reader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return ErrFileRejected
	}

	var expanded uint64

	foundDocument := false

	for _, file := range reader.File {
		expanded += file.UncompressedSize64
		if expanded > maxExpandedBytes {
			return ErrFileRejected
		}

		if file.Name == "word/document.xml" {
			foundDocument = true
		}
	}

	if !foundDocument {
		return ErrFileRejected
	}

	return nil
}

func scanClamAV(ctx context.Context, address string, timeout time.Duration, content []byte) (bool, error) {
	dialer := net.Dialer{Timeout: timeout}

	connection, dialErr := dialer.DialContext(ctx, "tcp", address)
	if dialErr != nil {
		return false, dialErr
	}

	defer closeConnection(connection)

	if err := connection.SetDeadline(time.Now().Add(timeout)); err != nil {
		return false, err
	}

	if err := writeFull(connection, []byte{'z', 'I', 'N', 'S', 'T', 'R', 'E', 'A', 'M', 0}); err != nil {
		return false, err
	}

	for offset := 0; offset < len(content); {
		end := min(offset+32*1024, len(content))
		length := make([]byte, 4)
		binary.BigEndian.PutUint32(length, uint32(end-offset))

		if err := writeFull(connection, length); err != nil {
			return false, err
		}

		if err := writeFull(connection, content[offset:end]); err != nil {
			return false, err
		}

		offset = end
	}

	if err := writeFull(connection, []byte{0, 0, 0, 0}); err != nil {
		return false, err
	}

	response, err := bufio.NewReader(connection).ReadString(0)
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}

	response = strings.TrimSpace(strings.TrimSuffix(response, "\x00"))
	if strings.HasSuffix(response, "OK") {
		return true, nil
	}

	if strings.Contains(response, "FOUND") {
		return false, nil
	}

	return false, fmt.Errorf("unexpected ClamAV response")
}

func writeFull(writer io.Writer, payload []byte) error {
	for len(payload) > 0 {
		written, err := writer.Write(payload)
		if err != nil {
			return err
		}

		if written <= 0 || written > len(payload) {
			return io.ErrShortWrite
		}

		payload = payload[written:]
	}

	return nil
}

func closeConnection(connection net.Conn) {
	if err := connection.Close(); err != nil {
		slog.Warn("managed RAG scanner connection close failed", "error_type", "connection_close")
	}
}

func sanitizeFilename(value string) string {
	value = filepath.Base(strings.TrimSpace(value))

	value = strings.Map(func(r rune) rune {
		if r == 0 || r == '\r' || r == '\n' || r == '/' || r == '\\' || r < 32 {
			return -1
		}

		return r
	}, value)
	if value == "" || value == "." {
		return "document"
	}

	return string([]rune(value)[:min(200, len([]rune(value)))])
}

func safeSegment(value string) string {
	var builder strings.Builder

	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			builder.WriteRune(r)
		}
	}

	if builder.Len() == 0 {
		return "unknown"
	}

	return builder.String()
}
