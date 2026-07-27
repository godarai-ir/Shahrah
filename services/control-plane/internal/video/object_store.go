package video

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var ErrObjectStoreUnavailable = errors.New("video object store unavailable")

type ObjectStore interface {
	PutInput(ctx context.Context, tenantID, jobID, filename, contentType string, body io.Reader) (platform.VideoArtifact, error)
	CopyFromURL(ctx context.Context, tenantID, jobID string, artifact platform.VideoArtifact) (platform.VideoArtifact, error)
	PresignGet(ctx context.Context, objectKey string, ttl time.Duration) (string, error)
	Delete(ctx context.Context, objectKey string) error
}

func (s *S3ObjectStore) PutInput(ctx context.Context, tenantID, jobID, filename, contentType string, body io.Reader) (platform.VideoArtifact, error) {
	if s == nil || s.client == nil {
		return platform.VideoArtifact{}, ErrObjectStoreUnavailable
	}
	extension := path.Ext(path.Base(filename))
	if extension == "" {
		extension = extensionForContentType(contentType)
	}
	objectKey := path.Join(s.keyPrefix, "tenants", tenantID, "videos", jobID, "input"+extension)
	hasher := sha256.New()
	counter := &countingReader{reader: io.TeeReader(body, hasher)}
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket), Key: aws.String(objectKey), Body: counter,
		ContentType: aws.String(firstContentType(contentType)),
		Metadata:    map[string]string{"tenant-id": tenantID, "video-job": jobID, "kind": "input"},
	}
	if s.encryption {
		input.ServerSideEncryption = types.ServerSideEncryptionAes256
	}
	if _, err := s.client.PutObject(ctx, input); err != nil {
		return platform.VideoArtifact{}, err
	}
	return platform.VideoArtifact{
		ID: "input", ObjectKey: objectKey, ContentType: firstContentType(contentType),
		SizeBytes: counter.count, ChecksumSHA256: hex.EncodeToString(hasher.Sum(nil)),
	}, nil
}

type S3ObjectStore struct {
	client     *s3.Client
	presign    *s3.PresignClient
	httpClient *http.Client
	bucket     string
	keyPrefix  string
	allowHTTP  bool
	encryption bool
	defaultTTL time.Duration
	maxBytes   int64
}

func NewS3ObjectStoreFromEnv(ctx context.Context) (*S3ObjectStore, error) {
	bucket := strings.TrimSpace(os.Getenv("VIDEO_OBJECT_STORE_BUCKET"))
	if bucket == "" {
		return nil, ErrObjectStoreUnavailable
	}
	region := envDefault("VIDEO_OBJECT_STORE_REGION", "us-east-1")
	endpoint := strings.TrimSpace(os.Getenv("VIDEO_OBJECT_STORE_ENDPOINT"))
	accessKey := strings.TrimSpace(os.Getenv("VIDEO_OBJECT_STORE_ACCESS_KEY_ID"))
	secretKey := strings.TrimSpace(os.Getenv("VIDEO_OBJECT_STORE_SECRET_ACCESS_KEY"))

	loadOptions := []func(*config.LoadOptions) error{config.WithRegion(region)}
	if accessKey != "" || secretKey != "" {
		if accessKey == "" || secretKey == "" {
			return nil, errors.New("both video object-store access key and secret key are required")
		}
		loadOptions = append(loadOptions, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		))
	}
	awsConfig, err := config.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsConfig, func(options *s3.Options) {
		options.UsePathStyle = envBool("VIDEO_OBJECT_STORE_PATH_STYLE", endpoint != "")
		if endpoint != "" {
			options.BaseEndpoint = aws.String(endpoint)
		}
	})
	ttlSeconds := envInt("VIDEO_OBJECT_STORE_PRESIGN_TTL_SECONDS", 900)
	return &S3ObjectStore{
		client:     client,
		presign:    s3.NewPresignClient(client),
		httpClient: &http.Client{Timeout: time.Duration(envInt("VIDEO_OUTPUT_DOWNLOAD_TIMEOUT_SECONDS", 300)) * time.Second},
		bucket:     bucket,
		keyPrefix:  strings.Trim(strings.TrimSpace(os.Getenv("VIDEO_OBJECT_STORE_PREFIX")), "/"),
		allowHTTP:  envBool("VIDEO_OUTPUT_ALLOW_HTTP", false),
		encryption: envBool("VIDEO_OBJECT_STORE_SSE", true),
		defaultTTL: time.Duration(ttlSeconds) * time.Second,
		maxBytes:   int64(envInt("VIDEO_OUTPUT_MAX_MIB", 2048)) * 1024 * 1024,
	}, nil
}

func (s *S3ObjectStore) CopyFromURL(ctx context.Context, tenantID, jobID string, artifact platform.VideoArtifact) (platform.VideoArtifact, error) {
	if s == nil || s.client == nil {
		return artifact, ErrObjectStoreUnavailable
	}
	source, err := url.Parse(strings.TrimSpace(artifact.ProviderURL))
	if err != nil || source.Host == "" || (source.Scheme != "https" && source.Scheme != "s3" && !(s.allowHTTP && source.Scheme == "http")) {
		return artifact, errors.New("provider output URL must be an allowed absolute URL")
	}

	var sourceBody io.ReadCloser
	sourceContentType := artifact.ContentType
	if source.Scheme == "s3" {
		result, getErr := s.client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(source.Host),
			Key:    aws.String(strings.TrimPrefix(source.Path, "/")),
		})
		if getErr != nil {
			return artifact, fmt.Errorf("provider S3 output download failed: %w", getErr)
		}
		sourceBody = result.Body
		if result.ContentType != nil {
			sourceContentType = *result.ContentType
		}
	} else {
		req, requestErr := http.NewRequestWithContext(ctx, http.MethodGet, source.String(), nil)
		if requestErr != nil {
			return artifact, requestErr
		}
		resp, requestErr := s.httpClient.Do(req)
		if requestErr != nil {
			return artifact, requestErr
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			_ = resp.Body.Close()
			return artifact, fmt.Errorf("provider output download failed with status %d", resp.StatusCode)
		}
		sourceBody = resp.Body
		sourceContentType = firstContentType(artifact.ContentType, resp.Header.Get("Content-Type"))
	}
	defer sourceBody.Close()

	artifactID := strings.TrimSpace(artifact.ID)
	if artifactID == "" {
		artifactID = "output"
	}
	extension := path.Ext(source.Path)
	if extension == "" {
		extension = extensionForContentType(sourceContentType)
	}
	objectKey := path.Join(s.keyPrefix, "tenants", tenantID, "videos", jobID, artifactID+extension)

	tempFile, err := os.CreateTemp("", "shahrah-video-output-*")
	if err != nil {
		return artifact, err
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)
	defer tempFile.Close()

	hasher := sha256.New()
	written, err := io.Copy(io.MultiWriter(tempFile, hasher), io.LimitReader(sourceBody, s.maxBytes+1))
	if err != nil {
		return artifact, err
	}
	if written > s.maxBytes {
		return artifact, fmt.Errorf("provider output exceeds %d bytes", s.maxBytes)
	}
	if _, err := tempFile.Seek(0, io.SeekStart); err != nil {
		return artifact, err
	}

	input := &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(objectKey),
		Body:          tempFile,
		ContentLength: aws.Int64(written),
		ContentType:   aws.String(firstContentType(sourceContentType)),
		Metadata: map[string]string{
			"tenant-id": tenantID,
			"video-job": jobID,
		},
	}
	if s.encryption {
		input.ServerSideEncryption = types.ServerSideEncryptionAes256
	}
	if _, err := s.client.PutObject(ctx, input); err != nil {
		return artifact, err
	}

	artifact.ObjectKey = objectKey
	artifact.ProviderURL = ""
	artifact.ContentType = firstContentType(sourceContentType)
	artifact.SizeBytes = written
	artifact.ChecksumSHA256 = hex.EncodeToString(hasher.Sum(nil))
	if duration, resolution, ok := probeVideoMetadata(ctx, tempPath); ok {
		artifact.Duration = duration
		artifact.Resolution = resolution
	}
	return artifact, nil
}

func probeVideoMetadata(parent context.Context, filename string) (string, string, bool) {
	probePath, err := exec.LookPath("ffprobe")
	if err != nil {
		return "", "", false
	}
	ctx, cancel := context.WithTimeout(parent, 20*time.Second)
	defer cancel()
	output, err := exec.CommandContext(
		ctx, probePath, "-v", "error", "-select_streams", "v:0",
		"-show_entries", "format=duration:stream=width,height",
		"-of", "json", filename,
	).Output()
	if err != nil {
		return "", "", false
	}
	var payload struct {
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
		Streams []struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(output, &payload); err != nil || strings.TrimSpace(payload.Format.Duration) == "" {
		return "", "", false
	}
	duration, err := strconv.ParseFloat(payload.Format.Duration, 64)
	if err != nil || duration <= 0 {
		return "", "", false
	}
	resolution := ""
	if len(payload.Streams) > 0 && payload.Streams[0].Height > 0 {
		resolution = strconv.Itoa(payload.Streams[0].Height) + "p"
	}
	return strconv.FormatFloat(duration, 'f', -1, 64), resolution, true
}

func (s *S3ObjectStore) PresignGet(ctx context.Context, objectKey string, ttl time.Duration) (string, error) {
	if s == nil || s.presign == nil {
		return "", ErrObjectStoreUnavailable
	}
	if ttl <= 0 {
		ttl = s.defaultTTL
	}
	result, err := s.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(strings.TrimSpace(objectKey)),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", err
	}
	return result.URL, nil
}

func (s *S3ObjectStore) Delete(ctx context.Context, objectKey string) error {
	if s == nil || s.client == nil {
		return ErrObjectStoreUnavailable
	}
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(strings.TrimSpace(objectKey)),
	})
	return err
}

type countingReader struct {
	reader io.Reader
	count  int64
}

func (r *countingReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	r.count += int64(n)
	return n, err
}

func extensionForContentType(contentType string) string {
	switch strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0])) {
	case "video/webm":
		return ".webm"
	case "video/quicktime":
		return ".mov"
	default:
		return ".mp4"
	}
}

func firstContentType(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(strings.Split(value, ";")[0])
		}
	}
	return "video/mp4"
}

func envDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envInt(key string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(os.Getenv(key)))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func RetentionDays() int {
	return envInt("VIDEO_OBJECT_RETENTION_DAYS", 7)
}
