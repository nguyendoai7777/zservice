package handler

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/supabase-community/supabase-go"

	"github.com/nguyendoai7777/zservice/internal/service/hls"
	"github.com/nguyendoai7777/zservice/internal/service/storage"
)

type UploadHandler struct {
	R2     *storage.R2
	SB     *supabase.Client
	TmpDir string
}

func NewUploadHandler(r2 *storage.R2, supabaseClient *supabase.Client, tmpDir string) *UploadHandler {
	return &UploadHandler{R2: r2, SB: supabaseClient, TmpDir: tmpDir}
}

type uploadResponse struct {
	ID          string `json:"id"`
	PlaylistKey string `json:"playlistKey"`
	PlaylistURL string `json:"playlistUrl,omitempty"`
	DurationMS  int64  `json:"durationMs"`
}

func (h *UploadHandler) Upload(c *gin.Context) {
	start := time.Now()

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing form field 'file'"})
		return
	}

	id := uuid.NewString()
	workDir := filepath.Join(h.TmpDir, id)
	hlsDir := filepath.Join(workDir, "hls")
	if err := os.MkdirAll(hlsDir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("mkdir: %v", err)})
		return
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {

		}
	}(workDir)

	srcPath := filepath.Join(workDir, "source"+filepath.Ext(fileHeader.Filename))
	if err := c.SaveUploadedFile(fileHeader, srcPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("save upload: %v", err)})
		return
	}

	if _, err := hls.Transcode(hls.TranscodeOptions{
		InputPath:       srcPath,
		OutDir:          hlsDir,
		AudioBitrate:    "128k",
		SegmentDuration: 10,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("transcode: %v", err)})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()

	keyPrefix := "tracks/" + id
	if err := h.R2.PutDir(ctx, hlsDir, keyPrefix, 8); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("upload r2: %v", err)})
		return
	}

	playlistKey := keyPrefix + "/playlist.m3u8"
	c.JSON(http.StatusOK, uploadResponse{
		ID:          id,
		PlaylistKey: playlistKey,
		PlaylistURL: h.R2.PublicURL(playlistKey),
		DurationMS:  time.Since(start).Milliseconds(),
	})
}
