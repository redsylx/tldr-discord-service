package gcs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"cloud.google.com/go/storage"
)

type Reader interface {
	ReadFile(ctx context.Context, path string) ([]byte, error)
	ReadTextLines(ctx context.Context, path string) ([]string, error)
	ReadJSON(ctx context.Context, path string, dst any) error
}

func NewReader(client *storage.Client, bucketName string) Reader {
	return &gcsReader{
		client: client,
		bucket: bucketName,
	}
}

type gcsReader struct {
	client *storage.Client
	bucket string
}

func (r *gcsReader) ReadFile(ctx context.Context, path string) ([]byte, error) {
	obj := r.client.Bucket(r.bucket).Object(path)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("gcs: open object %s/%s: %w", r.bucket, path, err)
	}
	defer reader.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(reader); err != nil {
		return nil, fmt.Errorf("gcs: read object %s/%s: %w", r.bucket, path, err)
	}
	return buf.Bytes(), nil
}

func (r *gcsReader) ReadTextLines(ctx context.Context, path string) ([]string, error) {
	data, err := r.ReadFile(ctx, path)
	if err != nil {
		return nil, err
	}

	if !utf8.Valid(data) {
		return nil, fmt.Errorf("gcs: file %s/%s is not valid UTF-8", r.bucket, path)
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	result := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimRight(line, "\r")
		result = append(result, trimmed)
	}

	if len(result) > 0 && result[len(result)-1] == "" {
		result = result[:len(result)-1]
	}

	return result, nil
}

func (r *gcsReader) ReadJSON(ctx context.Context, path string, dst any) error {
	data, err := r.ReadFile(ctx, path)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("gcs: parse JSON from %s/%s: %w", r.bucket, path, err)
	}
	return nil
}