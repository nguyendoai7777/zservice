package hls

import (
	"fmt"
	"os"
	"path/filepath"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type TranscodeOptions struct {
	// InputPath is the source audio file (mp3, wav, flac, ...).
	InputPath string
	// OutDir is where the .m3u8 + .ts segments will be written. Created if missing.
	OutDir string
	// AudioBitrate (e.g. "128k"). Defaults to "128k" when empty.
	AudioBitrate string
	// SegmentDuration seconds per .ts chunk. Defaults to 10 when 0.
	SegmentDuration int
}

type Result struct {
	PlaylistPath string   // absolute path to playlist.m3u8
	SegmentPaths []string // absolute paths to all .ts segments
}

// Transcode runs ffmpeg to convert the input audio into a VOD HLS stream.
// Requires the `ffmpeg` binary to be installed and on PATH.
func Transcode(opts TranscodeOptions) (*Result, error) {
	if opts.InputPath == "" {
		return nil, fmt.Errorf("hls: input path is empty")
	}
	if opts.OutDir == "" {
		return nil, fmt.Errorf("hls: out dir is empty")
	}
	if opts.AudioBitrate == "" {
		opts.AudioBitrate = "128k"
	}
	if opts.SegmentDuration <= 0 {
		opts.SegmentDuration = 10
	}

	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return nil, fmt.Errorf("hls: mkdir %s: %w", opts.OutDir, err)
	}

	playlist := filepath.Join(opts.OutDir, "playlist.m3u8")
	segmentPattern := filepath.Join(opts.OutDir, "seg_%03d.ts")

	err := ffmpeg.Input(opts.InputPath).
		Output(playlist, ffmpeg.KwArgs{
			"c:a":                "aac",
			"b:a":                opts.AudioBitrate,
			"vn":                 "",
			"hls_time":           opts.SegmentDuration,
			"hls_playlist_type":  "vod",
			"hls_segment_filename": segmentPattern,
			"f":                  "hls",
		}).
		OverWriteOutput().
		Run()
	if err != nil {
		return nil, fmt.Errorf("hls: ffmpeg run: %w", err)
	}

	entries, err := os.ReadDir(opts.OutDir)
	if err != nil {
		return nil, fmt.Errorf("hls: read out dir: %w", err)
	}
	var segs []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if filepath.Ext(e.Name()) == ".ts" {
			segs = append(segs, filepath.Join(opts.OutDir, e.Name()))
		}
	}

	return &Result{
		PlaylistPath: playlist,
		SegmentPaths: segs,
	}, nil
}
