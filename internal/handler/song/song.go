package song

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/supabase-community/postgrest-go"
)

func (h *SongHandler) GetSongs(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	targetAlbums := c.QueryArray("album")

	for i, alb := range targetAlbums {
		targetAlbums[i] = strings.TrimSpace(alb)
	}

	query := h.supabase.From("songs").
		Select("id, name, duration, listen_count, file_path, album_id, artist_id, sub_artist_ids, created_at, albums!inner(name)", "", false).
		Not("album_id", "is", "null").
		Order("created_at", &postgrest.OrderOpts{Ascending: false})

	if len(targetAlbums) > 0 {
		query = query.In("albums.name", targetAlbums)
	}

	if q != "" {
		escapedQ := q
		escapedQ = strings.ReplaceAll(escapedQ, "%", "\\%")
		escapedQ = strings.ReplaceAll(escapedQ, "_", "\\_")
		query = query.Ilike("name", "%"+escapedQ+"%")
	}

	byteData, _, err := query.Execute()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to query songs",
			"error":   err.Error(),
		})
		return
	}

	var songs []SongDB
	if err := json.Unmarshal(byteData, &songs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to parse songs data",
			"error":   err.Error(),
		})
		return
	}

	if len(songs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"albums":  targetAlbums,
			"q":       q,
			"total":   0,
			"data":    []interface{}{},
		})
		return
	}

	artistIdsMap := make(map[string]bool)
	for _, song := range songs {
		if song.ArtistID != nil && *song.ArtistID != "" {
			artistIdsMap[*song.ArtistID] = true
		}
		for _, subID := range song.SubArtistIDs {
			if subID != "" {
				artistIdsMap[subID] = true
			}
		}
	}

	var artistIds []string
	for id := range artistIdsMap {
		artistIds = append(artistIds, id)
	}

	artistsById := make(map[string]string)
	if len(artistIds) > 0 {
		artistQuery := h.supabase.From("artists").Select("id, name", "", false).In("id", artistIds)
		artistByteData, _, err := artistQuery.Execute()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to query artists",
				"error":   err.Error(),
			})
			return
		}

		var artists []ArtistDB
		if err := json.Unmarshal(artistByteData, &artists); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to parse artists data",
				"error":   err.Error(),
			})
			return
		}

		for _, artist := range artists {
			artistsById[artist.ID] = artist.Name
		}
	}

	responseData := make([]SongResponse, len(songs))
	for i, song := range songs {
		subArtists := make([]SubArtist, len(song.SubArtistIDs))
		for j, id := range song.SubArtistIDs {
			name := id
			if val, ok := artistsById[id]; ok {
				name = val
			}
			subArtists[j] = SubArtist{ID: id, Name: name}
		}

		var artistName *string
		if song.ArtistID != nil {
			if name, ok := artistsById[*song.ArtistID]; ok {
				artistName = &name
			}
		}

		var albumName *string
		if song.Albums != nil {
			albumName = &song.Albums.Name
		}

		responseData[i] = SongResponse{
			ID:          song.ID,
			Name:        song.Name,
			Duration:    song.Duration,
			ListenCount: song.ListenCount,
			AlbumID:     song.AlbumID,
			AlbumName:   albumName,
			ArtistID:    song.ArtistID,
			ArtistName:  artistName,
			SubArtists:  subArtists,
			URL:         song.FilePath,
			CreatedAt:   song.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"albums":  targetAlbums,
		"q":       q,
		"total":   len(responseData),
		"data":    responseData,
	})
}
