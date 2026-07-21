package proxy

import (
	"bufio"
	"bytes"
	"encoding/json"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"go.uber.org/zap"
)

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
