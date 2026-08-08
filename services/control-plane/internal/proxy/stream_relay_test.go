package proxy

import (
	"bufio"
	"bytes"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"go.uber.org/zap"
)

func TestRelayTunnelStreamDrainsChunksAfterErrorChannelCloses(t *testing.T) {
	stream := newFakeTunnelStream(map[string]string{"content-type": "text/event-stream"}, 2, 0)
	close(stream.errors)

	go func() {
		time.Sleep(10 * time.Millisecond)

		stream.chunks <- platform.TunnelChunk{Data: []byte("data: payload\n\n")}

		stream.chunks <- platform.TunnelChunk{IsDone: true}

		close(stream.chunks)
	}()

	var output bytes.Buffer

	writer := bufio.NewWriter(&output)
	succeeded := false

	relayTunnelStream(stream, writer, func() { succeeded = true }, zap.NewNop(), "request-test", "responses")

	if !succeeded {
		t.Fatal("expected stream relay to complete successfully")
	}

	if output.String() != "data: payload\n\n" {
		t.Fatalf("expected buffered chunk to be relayed, got %q", output.String())
	}
}
