package proxy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

const maxTunnelStreamErrorBodyBytes = 1 << 20

func writeTunnelStreamHTTPError(c *fiber.Ctx, logger *zap.Logger, stream platform.TunnelStream, onProviderError ...func(statusCode int, body []byte)) error {
	statusCode := stream.StatusCode()
	defer closeStreamBestEffort(logger, stream)

	body, err := readTunnelStreamErrorBody(stream)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Failed to read provider error response", "api_error")
	}

	if statusCode < httpErrorFloor {
		return writeAPIError(c, fiber.StatusBadGateway, "Provider stream returned an invalid status", "api_error")
	}

	if len(onProviderError) > 0 && onProviderError[0] != nil {
		onProviderError[0](statusCode, body)
	}

	return transformProviderError(c, statusCode, body, stream.Headers())
}

func readTunnelStreamErrorBody(stream platform.TunnelStream) ([]byte, error) {
	var body bytes.Buffer

	chunks := stream.Chunks()
	errors := stream.Errors()

	for chunks != nil || errors != nil {
		select {
		case chunk, ok := <-chunks:
			if !ok {
				chunks = nil
				continue
			}

			remaining := maxTunnelStreamErrorBodyBytes - body.Len()
			if remaining > 0 && len(chunk.Data) > 0 {
				data := chunk.Data
				if len(data) > remaining {
					data = data[:remaining]
				}

				_, _ = body.Write(data)
			}

			if chunk.IsDone {
				return body.Bytes(), nil
			}
		case streamErr, ok := <-errors:
			if !ok {
				errors = nil
				continue
			}

			if streamErr != nil {
				return body.Bytes(), fmt.Errorf("read provider error stream: %w", streamErr)
			}
		}
	}

	return body.Bytes(), nil
}

func relayTunnelStream(
	stream platform.TunnelStream,
	writer *bufio.Writer,
	onSuccess func(),
	logger *zap.Logger,
	requestID string,
	operation string,
) {
	for {
		select {
		case chunk, ok := <-stream.Chunks():
			if !ok {
				onSuccess()

				return
			}

			if len(chunk.Data) > 0 {
				if err := writeBufferedBytes(writer, chunk.Data); err != nil {
					logger.Warn("failed to write streaming chunk", zap.Error(err), zap.String("request_id", requestID), zap.String("operation", operation))

					return
				}
			}

			if chunk.IsDone {
				onSuccess()

				return
			}
		case streamErr, ok := <-stream.Errors():
			if ok && streamErr != nil {
				payload, err := json.Marshal(contractsopenai.APIError{
					Error: contractsopenai.ErrorInfo{Message: streamErr.Error(), Type: "streaming_error"},
				})
				if err != nil {
					logger.Warn("failed to encode streaming error payload", zap.Error(err), zap.String("request_id", requestID), zap.String("operation", operation))

					return
				}

				framed := bytes.Join([][]byte{[]byte("data: "), payload, []byte("\n\n")}, nil)
				if err := writeBufferedBytes(writer, framed); err != nil {
					logger.Warn("failed to write streaming error payload", zap.Error(err), zap.String("request_id", requestID), zap.String("operation", operation))
				}
			}

			return
		}
	}
}
