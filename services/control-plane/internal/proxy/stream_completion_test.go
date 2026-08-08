package proxy

import "testing"

func TestStreamCompletionCaptureRequiresDoneFrame(t *testing.T) {
	t.Parallel()

	capture := &streamCompletionCapture{}
	capture.Add([]byte(": keep-alive\n\n"))
	capture.Add([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"partial\"}}]}\n\n"))
	capture.Finish()

	if capture.Completed() {
		t.Fatal("expected a stream without data: [DONE] to remain incomplete")
	}
}

func TestStreamCompletionCaptureHandlesFragmentedDoneFrame(t *testing.T) {
	t.Parallel()

	capture := &streamCompletionCapture{}
	capture.Add([]byte("data: [DO"))
	capture.Add([]byte("NE]\n"))
	capture.Add([]byte("\n"))

	if !capture.Completed() {
		t.Fatal("expected a fragmented data: [DONE] frame to complete the stream")
	}
}
