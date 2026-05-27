package storage

import (
	"context"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/nguyendoai7777/zservice/internal/config"
)

type R2 struct {
	client       *s3.Client
	bucket       string
	publicDomain string
}

func NewR2(ctx context.Context, cfg config.R2Config) (*R2, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion("auto"),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true
	})

	return &R2{
		client:       client,
		bucket:       cfg.Bucket,
		publicDomain: strings.TrimRight(cfg.PublicDomain, "/"),
	}, nil
}

func (r *R2) PutObject(ctx context.Context, key string, body io.Reader, contentType string) error {
	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("put object %s: %w", key, err)
	}
	return nil
}

func (r *R2) PutFile(ctx context.Context, key, localPath string) error {
	f, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open %s: %w", localPath, err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)

	ct := mime.TypeByExtension(filepath.Ext(localPath))
	if ct == "" {
		ct = "application/octet-stream"
	}

	switch strings.ToLower(filepath.Ext(localPath)) {
	case ".m3u8":
		ct = "application/vnd.apple.mpegurl"
	case ".ts":
		ct = "video/mp2t"
	}

	return r.PutObject(ctx, key, f, ct)
}

func (r *R2) PutDir(ctx context.Context, localDir, keyPrefix string, concurrency int) error {
	if concurrency <= 0 {
		concurrency = 4
	}

	type job struct {
		key   string
		local string
	}
	jobs := make(chan job)
	errCh := make(chan error, 1)

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				if err := r.PutFile(ctx, j.key, j.local); err != nil {
					select {
					case errCh <- err:
					default:
					}
					return
				}
			}
		}()
	}

	walkErr := filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(localDir, path)
		if err != nil {
			return err
		}
		key := strings.TrimRight(keyPrefix, "/") + "/" + filepath.ToSlash(rel)
		select {
		case jobs <- job{key: key, local: path}:
		case <-ctx.Done():
			return ctx.Err()
		}
		return nil
	})
	close(jobs)
	wg.Wait()

	if walkErr != nil {
		return walkErr
	}
	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

func (r *R2) PublicURL(key string) string {
	if r.publicDomain == "" {
		return ""
	}
	return r.publicDomain + "/" + strings.TrimLeft(key, "/")
}
