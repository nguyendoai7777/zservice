package song

import (
	"time"

	"github.com/supabase-community/supabase-go"
)

type SongHandler struct {
	supabase *supabase.Client
}

func NewSongHandler(supabaseClient *supabase.Client) *SongHandler {
	return &SongHandler{supabase: supabaseClient}
}

type AlbumDB struct {
	Name string `json:"name"`
}

type SongDB struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Duration     int       `json:"duration"`
	ListenCount  int       `json:"listen_count"`
	FilePath     string    `json:"file_path"`
	AlbumID      *int64    `json:"album_id"`
	ArtistID     *string   `json:"artist_id"`
	SubArtistIDs []string  `json:"sub_artist_ids"`
	CreatedAt    time.Time `json:"created_at"`
	Albums       *AlbumDB  `json:"albums"`
}

type ArtistDB struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type SubArtist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type SongResponse struct {
	ID          int64       `json:"id"` // Đổi sang int64 đồng bộ client
	Name        string      `json:"name"`
	Duration    int         `json:"duration"`
	ListenCount int         `json:"listenCount"`
	AlbumID     *int64      `json:"albumId"`
	AlbumName   *string     `json:"albumName"`
	ArtistID    *string     `json:"artistId"`
	ArtistName  *string     `json:"artistName"`
	SubArtists  []SubArtist `json:"subArtists"`
	URL         string      `json:"url"`
	CreatedAt   time.Time   `json:"createdAt"`
}
