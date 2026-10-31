package proxy

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// parseVideoCreateRequest decodes either request shape. The OpenAI shape is
// translated by the provider codec later; the xAI shape is forwarded as-is
// after strict validation.
func parseVideoCreateRequest(c *fiber.Ctx, operation, shape string) (canonicalVideoRequest, map[string]interface{}, error) {
	request := canonicalVideoRequest{Operation: operation}

	if isMultipartContentType(c.Get(fiber.HeaderContentType)) {
		if shape != videoShapeOpenAI || operation != platform.VideoOperationGenerate {
			return request, nil, fmt.Errorf("multipart is only supported for POST /v1/videos; reference source videos by id")
		}

		return parseVideoMultipart(c, request)
	}

	if shape == videoShapeXAI {
		var native map[string]interface{}
		if err := json.Unmarshal(c.Body(), &native); err != nil || native == nil {
			return request, nil, fmt.Errorf("request body must be a JSON object")
		}

		return request, native, nil
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(c.Body(), &fields); err != nil || fields == nil {
		return request, nil, fmt.Errorf("request body must be a JSON object")
	}

	request.Extensions = map[string]json.RawMessage{}

	for field, raw := range fields {
		if err := applyCanonicalVideoField(&request, field, raw); err != nil {
			return request, nil, err
		}
	}

	if operation != platform.VideoOperationGenerate && request.Source == nil {
		return request, nil, fmt.Errorf("video.id is required")
	}

	return request, nil, nil
}

func applyCanonicalVideoField(request *canonicalVideoRequest, field string, raw json.RawMessage) error {
	switch field {
	case "model", "prompt", "size":
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			return fmt.Errorf("%s must be a string", field)
		}

		target := map[string]*string{"model": &request.Model, "prompt": &request.Prompt, "size": &request.Size}[field]
		*target = strings.TrimSpace(value)
	case "seconds":
		seconds, err := platform.ParseVideoSeconds(raw)
		if err != nil {
			return fmt.Errorf("seconds must be a whole number of seconds")
		}

		request.Seconds = seconds
	case "input_reference":
		var reference struct {
			ImageURL string `json:"image_url"`
			FileID   string `json:"file_id"`
		}
		if err := json.Unmarshal(raw, &reference); err != nil {
			return fmt.Errorf("input_reference must be an object with image_url or file_id")
		}

		request.Image = map[string]interface{}{}
		if reference.ImageURL != "" {
			request.Image["url"] = reference.ImageURL
		}

		if reference.FileID != "" {
			request.Image["file_id"] = reference.FileID
		}
	case "video":
		var source struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(raw, &source); err != nil || strings.TrimSpace(source.ID) == "" {
			return fmt.Errorf("video must be an object with id")
		}

		request.Source = map[string]interface{}{"id": strings.TrimSpace(source.ID)}
	default:
		request.Extensions[field] = raw
	}

	return nil
}

func parseVideoMultipart(c *fiber.Ctx, request canonicalVideoRequest) (canonicalVideoRequest, map[string]interface{}, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return request, nil, fmt.Errorf("invalid multipart request body")
	}

	request.Extensions = map[string]json.RawMessage{}

	for field, values := range form.Value {
		if len(values) != 1 {
			return request, nil, fmt.Errorf("%s must be sent once", field)
		}

		raw, _ := json.Marshal(values[0]) //nolint:errcheck // Marshaling a string cannot fail.
		if field == "seconds" {
			raw = json.RawMessage(strconv.Quote(values[0]))
		}

		if err := applyCanonicalVideoField(&request, field, raw); err != nil {
			return request, nil, err
		}
	}

	for field, files := range form.File {
		if field != "input_reference" || len(files) != 1 {
			return request, nil, fmt.Errorf("unsupported file field %q", field)
		}

		file, openErr := files[0].Open()
		if openErr != nil {
			return request, nil, fmt.Errorf("input_reference could not be read")
		}

		data, readErr := io.ReadAll(io.LimitReader(file, maxVideoRequestBytes))
		_ = file.Close() //nolint:errcheck // The upload is fully read; close errors cannot change the request.

		if readErr != nil {
			return request, nil, fmt.Errorf("input_reference could not be read")
		}

		contentType := http.DetectContentType(data)
		if !containsFoldValue([]string{"image/jpeg", "image/png", "image/webp"}, contentType) {
			return request, nil, fmt.Errorf("input_reference must be a JPEG, PNG, or WebP image")
		}

		request.Image = map[string]interface{}{"url": "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(data)}
	}

	return request, nil, nil
}

func checkVideoExtensionFields(extensions map[string]json.RawMessage, profile platform.ModelRuntimeProfile) error {
	for field := range extensions {
		if !containsFoldValue(profile.VideoControls.ExtensionFields, field) {
			return fmt.Errorf("parameter %q is not supported by this video model", field)
		}
	}

	return nil
}

func newVideoPublicID() string {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		panic(fmt.Sprintf("crypto/rand failed: %v", err))
	}

	return "video_" + hex.EncodeToString(buffer)
}

func writeVideoLookupError(c *fiber.Ctx, err error) error {
	if errors.Is(err, errVideoNotFound) {
		return writeAPIErrorDetail(c, fiber.StatusNotFound, "No video found with that id", "invalid_request_error", "video_id", "video_not_found")
	}

	return writeAPIError(c, fiber.StatusServiceUnavailable, "Video jobs are unavailable", "api_error")
}

func videoErrorType(status int) string {
	switch status {
	case fiber.StatusForbidden:
		return "authorization_error"
	case fiber.StatusServiceUnavailable:
		return "routing_error"
	default:
		return "invalid_request_error"
	}
}

func isPendingVideoStatus(status string) bool {
	return status == videoStatusQueued || status == videoStatusInProgress
}

func containsFoldValue(values []string, wanted string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(wanted)) {
			return true
		}
	}

	return false
}

func stringValue(value interface{}) string {
	text, ok := value.(string)
	if !ok {
		return ""
	}

	return strings.TrimSpace(text)
}

func (h *VideosHandler) logWarn(message string, err error, publicID string) {
	if h.logger != nil {
		h.logger.Warn(message, zap.Error(err), zap.String("video_id", publicID))
	}
}
