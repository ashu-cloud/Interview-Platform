package storage

import (
	"log/slog"
	"os"
	"testing"

	"github.com/minio/minio-go/v7"
)

func TestAudioOffloaderNoopWhenEndpointEmpty(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	offloader := NewAudioOffloader("", "", "", "assessment-audio", false, nil, logger)
	if offloader == nil {
		t.Fatal("expected non-nil offloader")
	}

	// Must not panic and must safely discard
	offloader.Enqueue("session-1", 1, []byte("fake-audio"), "audio/webm")
	offloader.Start()
	offloader.Stop()
}

func TestAudioOffloaderEnqueueBounded(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	offloader := &AudioOffloader{
		client:     &minio.Client{},
		uploadChan: make(chan audioUpload, 2),
		logger:     logger,
	}

	// Should not block even if channel overflows
	offloader.Enqueue("session-1", 1, []byte("a"), "audio/webm")
	offloader.Enqueue("session-1", 2, []byte("b"), "audio/webm")
	offloader.Enqueue("session-1", 3, []byte("c"), "audio/webm") // dropped without blocking

	if len(offloader.uploadChan) != 2 {
		t.Fatalf("channel len = %d, want 2", len(offloader.uploadChan))
	}
}
