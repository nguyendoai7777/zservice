package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string

	R2 R2Config
	DB DBConfig
}

type R2Config struct {
	AccountID       string
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	PublicDomain    string
}

type DBConfig struct {
	ConnectionString string
	SupabaseURL      string
	SupabaseAnonKey  string
	StoragePath      string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port: getEnv("PORT", "3000"),
		R2: R2Config{
			AccountID:       os.Getenv("CLOUDFLARE_ACCOUNT_ID"),
			Endpoint:        os.Getenv("CLOUDFLARE_S3_API"),
			AccessKeyID:     os.Getenv("CLOUDFLARE_R2_ACCESS_KEY_ID"),
			SecretAccessKey: os.Getenv("CLOUDFLARE_R2_SECRET_ACCESS_KEY"),
			Bucket:          os.Getenv("CLOUDFLARE_R2_BUCKET_NAME"),
			PublicDomain:    os.Getenv("CLOUDFLARE_PUBLIC_DOMAIN"),
		},
		DB: DBConfig{
			ConnectionString: os.Getenv("SUPABASE_DIRECT_CONNECTION_STRING"),
			SupabaseURL:      os.Getenv("SUPABASE_URL"),
			SupabaseAnonKey:  os.Getenv("SUPABASE_PUBLISHABLE_KEY"),
			StoragePath:      os.Getenv("SUPABASE_PATH"),
		},
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	required := map[string]string{
		"CLOUDFLARE_S3_API":               c.R2.Endpoint,
		"CLOUDFLARE_R2_ACCESS_KEY_ID":     c.R2.AccessKeyID,
		"CLOUDFLARE_R2_SECRET_ACCESS_KEY": c.R2.SecretAccessKey,
		"CLOUDFLARE_R2_BUCKET_NAME":       c.R2.Bucket,
		// Thay thế kiểm tra Connection String cũ bằng 2 biến API mới giống Node.js
		"SUPABASE_URL":             c.DB.SupabaseURL,
		"SUPABASE_PUBLISHABLE_KEY": c.DB.SupabaseAnonKey,
	}
	for k, v := range required {
		if v == "" {
			return fmt.Errorf("missing required env: %s", k)
		}
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
