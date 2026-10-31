package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

const xaiVideoPollBodyLimit = 1 << 20

func xaiVideoOperation(operation platform.ProviderOperation) string {
	switch operation {
	case platform.ProviderOperationXAIVideosGenerate:
		return platform.VideoOperationGenerate
	case platform.ProviderOperationXAIVideosEdit:
		return platform.VideoOperationEdit
	case platform.ProviderOperationXAIVideosExtend:
		return platform.VideoOperationExtend
	default:
		return ""
	}
}

// validateXAIVideoRequest re-checks the forwarded body against the pinned
// profile, so the egress never relies on the control plane alone.
func validateXAIVideoRequest(payload platform.TunnelRequest, profile platform.ModelRuntimeProfile) error {
	operation := xaiVideoOperation(payload.Operation)
	if operation == "" {
		return nil
	}

	facts, err := platform.ParseXAIVideoRequest(operation, payload.Body)
	if err != nil {
		return fmt.Errorf("xai_invalid_video_request: %w", err)
	}

	if err := profile.ValidateVideoRequest(facts); err != nil {
		return fmt.Errorf("xai_invalid_video_request: %w", err)
	}

	return nil
}

// dispatchVideoContent polls the job for its temporary signed URL and streams
// the MP4 from it. The URL never leaves the egress, the API key is never sent
// to the media host, and only profile-listed hosts are fetched.
func (a *xaiAdapter) dispatchVideoContent(ctx context.Context, client *http.Client, payload platform.TunnelRequest) (*adapterResponse, error) {
	binding, err := validateXAIRuntimeProfile(payload, strings.TrimSpace(payload.Deployment.Model), "videos")
	if err != nil {
		return nil, err
	}

	poll := payload
	poll.Operation = platform.ProviderOperationXAIVideosGet
	poll.Body = nil

	response, err := a.dispatchWithRetry(ctx, client, poll)
	if err != nil {
		return nil, err
	}

	if response.StatusCode >= http.StatusBadRequest {
		return response, nil
	}

	location, err := xaiVideoContentURL(response.Body)
	_ = response.Body.Close() //nolint:errcheck // The poll body is fully read; close errors cannot change the result.

	if err != nil {
		return xaiVideoContentUnavailable(err.Error()), nil
	}

	target, err := url.Parse(location)
	if err != nil || target.Scheme != "https" || target.User != nil ||
		!containsFold(binding.Profile.VideoControls.ContentHosts, target.Hostname()) {
		host := ""
		if target != nil {
			host = target.Hostname()
		}

		return nil, fmt.Errorf("xai_video_content_host_not_allowed: %s", host)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), http.NoBody)
	if err != nil {
		return nil, err
	}

	noRedirects := *client
	noRedirects.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }

	media, err := noRedirects.Do(request)
	if err != nil {
		return nil, err
	}

	if media.StatusCode >= http.StatusMultipleChoices && media.StatusCode < http.StatusBadRequest {
		_ = media.Body.Close() //nolint:errcheck // The redirect is refused; close errors cannot change the result.

		return nil, fmt.Errorf("xai_video_content_redirect_refused")
	}

	if media.StatusCode >= http.StatusBadRequest {
		_ = media.Body.Close() //nolint:errcheck // The expired asset is reported instead; close errors cannot change the result.

		return xaiVideoContentUnavailable("video content has expired upstream"), nil
	}

	headers := http.Header{}

	for _, name := range []string{"Content-Type", "Content-Length"} {
		if value := media.Header.Get(name); value != "" {
			headers.Set(name, value)
		}
	}

	return &adapterResponse{StatusCode: media.StatusCode, Headers: headers, Body: media.Body}, nil
}

func xaiVideoContentURL(body io.Reader) (string, error) {
	var result struct {
		Status string `json:"status"`
		Video  struct {
			URL               string `json:"url"`
			RespectModeration *bool  `json:"respect_moderation"`
		} `json:"video"`
	}

	if err := json.NewDecoder(io.LimitReader(body, xaiVideoPollBodyLimit)).Decode(&result); err != nil {
		return "", fmt.Errorf("video status is unreadable")
	}

	if result.Status != "done" || (result.Video.RespectModeration != nil && !*result.Video.RespectModeration) ||
		strings.TrimSpace(result.Video.URL) == "" {
		return "", fmt.Errorf("video content is not available")
	}

	return strings.TrimSpace(result.Video.URL), nil
}

func xaiVideoContentUnavailable(message string) *adapterResponse {
	body, _ := json.Marshal(map[string]interface{}{ //nolint:errcheck // Marshaling a string map cannot fail.
		"error": map[string]string{"message": message, "type": "invalid_request_error", "code": "video_content_unavailable"},
	})

	return &adapterResponse{
		StatusCode: http.StatusConflict,
		Headers:    http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader(body)),
	}
}
