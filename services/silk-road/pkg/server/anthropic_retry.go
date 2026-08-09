package server

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

const anthropicMaxAttempts = 3

func doAnthropicRequest(
	ctx context.Context,
	client *http.Client,
	request *http.Request,
	retryAllowed bool,
) (*http.Response, error) {
	if !retryAllowed {
		return client.Do(request)
	}

	var lastErr error

	for attempt := 0; attempt < anthropicMaxAttempts; attempt++ {
		current := request

		if attempt > 0 {
			var err error

			current, err = cloneRequestForRetry(request, ctx)
			if err != nil {
				return nil, err
			}
		}

		response, err := client.Do(current)
		if err == nil && !anthropicRetryableStatus(response.StatusCode) {
			return response, nil
		}

		if err == nil && attempt == anthropicMaxAttempts-1 {
			anthropicRetryExhausted.Add(1)
			return response, nil
		}

		if err != nil {
			lastErr = err

			if ctx.Err() != nil || attempt == anthropicMaxAttempts-1 {
				if attempt == anthropicMaxAttempts-1 {
					anthropicRetryExhausted.Add(1)
				}

				return nil, err
			}
		}

		delay := time.Duration(1<<attempt) * 250 * time.Millisecond
		if response != nil {
			delay = anthropicRetryDelay(response.Header, delay, time.Now())
			if _, copyErr := io.Copy(
				io.Discard,
				io.LimitReader(response.Body, 64*1024),
			); copyErr != nil {
				closeBody(response.Body)

				return nil, copyErr
			}

			if closeErr := response.Body.Close(); closeErr != nil {
				return nil, closeErr
			}
		}

		anthropicRetryTotal.Add(1)

		if err := waitForAnthropicRetry(ctx, delay); err != nil {
			return nil, err
		}
	}

	return nil, lastErr
}

func cloneRequestForRetry(
	request *http.Request,
	ctx context.Context,
) (*http.Request, error) {
	cloned := request.Clone(ctx)
	if request.GetBody == nil {
		if request.Body == nil || request.Body == http.NoBody {
			cloned.Body = http.NoBody
			return cloned, nil
		}

		return nil, io.ErrUnexpectedEOF
	}

	body, err := request.GetBody()
	if err != nil {
		return nil, err
	}

	cloned.Body = body

	return cloned, nil
}

func anthropicRetryableStatus(status int) bool {
	return status == http.StatusTooManyRequests ||
		status >= http.StatusInternalServerError
}

func anthropicRetryDelay(
	headers http.Header,
	fallback time.Duration,
	now time.Time,
) time.Duration {
	value := strings.TrimSpace(headers.Get("retry-after"))
	if value == "" {
		return fallback
	}

	if seconds, err := strconv.ParseInt(value, 10, 64); err == nil &&
		seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}

	if retryAt, err := http.ParseTime(value); err == nil {
		if delay := retryAt.Sub(now); delay > 0 {
			return delay
		}

		return 0
	}

	return fallback
}

func waitForAnthropicRetry(
	ctx context.Context,
	delay time.Duration,
) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func anthropicOperationIsIdempotent(
	operation platform.ProviderOperation,
) bool {
	switch operation {
	case platform.ProviderOperationAnthropicCountTokens,
		platform.ProviderOperationAnthropicModelsList,
		platform.ProviderOperationAnthropicModelsGet,
		platform.ProviderOperationAnthropicFilesList,
		platform.ProviderOperationAnthropicFilesGet,
		platform.ProviderOperationAnthropicFilesContent,
		platform.ProviderOperationAnthropicBatchesList,
		platform.ProviderOperationAnthropicBatchesGet,
		platform.ProviderOperationAnthropicBatchesResults:
		return true
	default:
		return false
	}
}
