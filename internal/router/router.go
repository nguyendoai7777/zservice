package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nguyendoai7777/zservice/internal/handler"
	"github.com/nguyendoai7777/zservice/internal/handler/song"
)

func New(upload *handler.UploadHandler, song *song.SongHandler) *gin.Engine {
	r := gin.Default()
	r.MaxMultipartMemory = 64 << 20 // 64 MiB in-memory before spooling to disk

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	api := r.Group("/api")
	{
		api.POST("/upload", upload.Upload)
		api.GET("/songs", song.GetSongs) // FIX: Đổi từ song.List thành song.GetSongs
	}

	return r
}
