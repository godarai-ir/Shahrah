package proxy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
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

func TestOpenAIResponseStreamStateTracksFragmentedFramesAndSequence(t *testing.T) {
	t.Parallel()

	state := &openAIResponseStreamState{}
	state.Observe([]byte("event: response.output_text.delta\r\ndata: {\"type\":\"response.output_text.delta\","))
	state.Observe([]byte("\"sequence_number\":41}\r\n\r\n"))

	if state.NextSequenceNumber() != 42 {
		t.Fatalf("expected next sequence number 42, got %d", state.NextSequenceNumber())
	}

	if state.TerminalEventType() != "" {
		t.Fatalf("unexpected terminal event %q", state.TerminalEventType())
	}
}

func TestOpenAIResponseStreamStateRecognizesLargeTerminalFrameWithoutBufferingItAll(t *testing.T) {
	t.Parallel()

	state := &openAIResponseStreamState{}
	prefix := "event: response.completed\n" +
		"data: {\"type\":\"response.completed\",\"response\":{\"output_text\":\""
	state.Observe(append([]byte(prefix), bytes.Repeat([]byte("x"), 300*1024)...))
	state.Observe([]byte("\"},\"sequence_number\":87}\n\n"))

	if state.TerminalEventType() != "response.completed" {
		t.Fatalf("expected large terminal frame to be recognized, got %q", state.TerminalEventType())
	}

	if state.NextSequenceNumber() != 88 {
		t.Fatalf("expected next sequence number 88, got %d", state.NextSequenceNumber())
	}
}

func TestOpenAIResponseStreamStateAcceptsEverySSELineEnding(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		separator string
		boundary  string
	}{
		{name: "LF", separator: "\n", boundary: "\n\n"},
		{name: "CRLF", separator: "\r\n", boundary: "\r\n\r\n"},
		{name: "CR", separator: "\r", boundary: "\r\r"},
		{name: "mixed CRLF LF", separator: "\r\n", boundary: "\r\n\n"},
		{name: "mixed LF CRLF", separator: "\n", boundary: "\n\r\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			state := &openAIResponseStreamState{}
			frame := "event: response.completed" + test.separator +
				"data: {\"type\":\"response.completed\",\"sequence_number\":3}" + test.boundary
			state.Observe([]byte(frame))

			if state.TerminalEventType() != "response.completed" {
				t.Fatalf("expected terminal event with %s line endings, got %q", test.name, state.TerminalEventType())
			}
		})
	}
}

func TestOpenAIResponseStreamStateJoinsMultipleSSEDataLines(t *testing.T) {
	t.Parallel()

	state := &openAIResponseStreamState{}
	state.Observe([]byte("event: response.completed\n" +
		"data: {\"type\":\"response.completed\",\n" +
		"data: \"sequence_number\":12}\n\n"))

	if state.TerminalEventType() != "response.completed" || state.NextSequenceNumber() != 13 {
		t.Fatalf("expected multi-line SSE data to be parsed, state=%+v", state)
	}
}

func TestOpenAIResponseStreamStateRecognizesEveryTerminalEvent(t *testing.T) {
	t.Parallel()

	for _, eventType := range contractsopenai.ResponsesTerminalStreamingEventTypes() {
		t.Run(eventType, func(t *testing.T) {
			t.Parallel()

			state := &openAIResponseStreamState{}
			state.Observe([]byte("event: " + eventType + "\ndata: {\"type\":\"" + eventType + "\",\"sequence_number\":9}\n\n"))

			if state.TerminalEventType() != eventType {
				t.Fatalf("expected terminal event %q, got %q", eventType, state.TerminalEventType())
			}

			if state.NextSequenceNumber() != 10 {
				t.Fatalf("expected next sequence number 10, got %d", state.NextSequenceNumber())
			}
		})
	}
}

func TestOpenAIResponseStreamTerminalClassifierMatchesEntireContractInventory(t *testing.T) {
	t.Parallel()

	terminal := make(map[string]struct{}, len(contractsopenai.ResponsesTerminalStreamingEventTypes()))
	for _, eventType := range contractsopenai.ResponsesTerminalStreamingEventTypes() {
		terminal[eventType] = struct{}{}
	}

	for _, eventType := range contractsopenai.ResponsesStreamingEventTypes() {
		_, expected := terminal[eventType]
		if actual := isTerminalOpenAIResponseEvent(eventType); actual != expected {
			t.Errorf("terminal classification for %q: expected %t, got %t", eventType, expected, actual)
		}
	}
}

func TestFrameOpenAIResponseStreamErrorUsesSuppliedSequenceNumber(t *testing.T) {
	t.Parallel()

	framed, err := frameOpenAIResponseStreamError(assertiveError("stream interrupted"), 17)
	if err != nil {
		t.Fatalf("frame streaming error: %v", err)
	}

	data := strings.TrimSpace(strings.TrimPrefix(strings.SplitN(string(framed), "\n", 2)[1], "data: "))

	var event openAIResponseErrorEvent
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		t.Fatalf("decode streaming error: %v", err)
	}

	if event.SequenceNumber != 17 {
		t.Fatalf("expected sequence number 17, got %d", event.SequenceNumber)
	}
}

type assertiveError string

func (err assertiveError) Error() string {
	return string(err)
}
