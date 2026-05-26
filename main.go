package main

import (
	"context"
	"log"
	"path/filepath"

	"github.com/nguyendoai7777/zservice/internal/config"
	"github.com/nguyendoai7777/zservice/internal/handler"
	"github.com/nguyendoai7777/zservice/internal/handler/song"
	"github.com/nguyendoai7777/zservice/internal/router"
	"github.com/nguyendoai7777/zservice/internal/service/db"
	"github.com/nguyendoai7777/zservice/internal/service/storage"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()

	db.InitSupabase(cfg.DB.SupabaseURL, cfg.DB.SupabaseAnonKey)

	r2, err := storage.NewR2(ctx, cfg.R2)
	if err != nil {
		log.Fatalf("r2: %v", err)
	}

	tmpDir, err := filepath.Abs("tmp")
	if err != nil {
		log.Fatalf("tmp dir: %v", err)
	}

	uploadHandler := handler.NewUploadHandler(r2, db.Supabase, tmpDir)
	songHandler := song.NewSongHandler(db.Supabase)

	engine := router.New(uploadHandler, songHandler)

	addr := ":" + cfg.Port
	log.Printf("listening on %s", addr)
	if err := engine.Run(addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}
