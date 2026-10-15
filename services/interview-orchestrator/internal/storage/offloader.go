package storage

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type audioUpload struct {
	sessionID  string
	turnNumber int
	audio      []byte
	mimeType   string
}

// AudioOffloader manages asynchronous offloading of participant audio to MinIO / R2 and recording metadata in Postgres.
type AudioOffloader struct {
	client     *minio.Client
	bucket     string
	pool       *pgxpool.Pool
	uploadChan chan audioUpload
	logger     *slog.Logger
	stopOnce   sync.Once
	doneChan   chan struct{}
}

// NewAudioOffloader creates an offloader.
// Returns a no-op offloader if endpoint is empty.
func NewAudioOffloader(endpoint, accessKey, secretKey, bucket string, useSSL bool, pool *pgxpool.Pool, logger *slog.Logger) *AudioOffloader {
	if logger == nil {
		logger = slog.Default()
	}

	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		logger.Info("audio offloader: MinIO endpoint empty, offloading disabled")
		return &AudioOffloader{
			logger:   logger,
			doneChan: make(chan struct{}),
		}
	}

	// Strip http:// or https:// if provided in endpoint
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		logger.Warn("audio offloader: failed to initialize MinIO client, offloading disabled", "error", err)
		return &AudioOffloader{
			logger:   logger,
			doneChan: make(chan struct{}),
		}
	}

	if bucket == "" {
		bucket = "assessment-audio"
	}

	return &AudioOffloader{
		client:     client,
		bucket:     bucket,
		pool:       pool,
		uploadChan: make(chan audioUpload, 50),
		logger:     logger,
		doneChan:   make(chan struct{}),
	}
}

// Enqueue adds an audio buffer to a bounded channel (size 50).
// If the channel is full or offloading is disabled, drops silently and logs a warning.
// The live assessment turn NEVER waits on this.
func (o *AudioOffloader) Enqueue(sessionID string, turnNum int, audio []byte, mimeType string) {
	if o == nil || o.client == nil || o.uploadChan == nil {
		return
	}
	if len(audio) == 0 {
		return
	}

	if mimeType == "" {
		mimeType = "audio/webm;codecs=opus"
	}

	item := audioUpload{
		sessionID:  sessionID,
		turnNumber: turnNum,
		audio:      audio,
		mimeType:   mimeType,
	}

	select {
	case o.uploadChan <- item:
	default:
		o.logger.Warn("audio offloader: upload channel full, dropping audio frame", "session_id", sessionID, "turn", turnNum)
	}
}

// Start runs the background upload goroutine.
func (o *AudioOffloader) Start() {
	if o == nil || o.client == nil || o.uploadChan == nil {
		return
	}

	go o.worker()
}

func (o *AudioOffloader) worker() {
	defer close(o.doneChan)
	for item := range o.uploadChan {
		o.processItem(item)
	}
}

func (o *AudioOffloader) processItem(item audioUpload) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectName := fmt.Sprintf("%s/%d/participant.webm", item.sessionID, item.turnNumber)
	size := int64(len(item.audio))

	_, err := o.client.PutObject(ctx, o.bucket, objectName, bytes.NewReader(item.audio), size, minio.PutObjectOptions{
		ContentType: item.mimeType,
	})
	if err != nil {
		o.logger.Warn("audio offloader: failed to upload audio to object storage", "error", err, "object", objectName)
		return
	}

	storageURI := fmt.Sprintf("s3://%s/%s", o.bucket, objectName)
	if o.pool != nil {
		_, dbErr := o.pool.Exec(ctx, `
			INSERT INTO audio_recordings (session_id, turn_number, storage_uri, size_bytes)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT DO NOTHING
		`, item.sessionID, item.turnNumber, storageURI, size)
		if dbErr != nil {
			o.logger.Warn("audio offloader: failed to record audio metadata in database", "error", dbErr, "session_id", item.sessionID)
		}
	}
}

// Stop drains and stops the background worker.
func (o *AudioOffloader) Stop() {
	if o == nil || o.uploadChan == nil {
		return
	}

	o.stopOnce.Do(func() {
		close(o.uploadChan)
		<-o.doneChan
	})
}
