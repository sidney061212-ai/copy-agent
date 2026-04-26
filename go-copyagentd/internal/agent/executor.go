package agent

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var ErrClipboardWriterRequired = errors.New("clipboard writer is required")
var ErrImageClipboardRequired = errors.New("image clipboard is required")
var ErrResourceDownloaderRequired = errors.New("resource downloader is required")

type TextClipboardWriter interface {
	WriteText(ctx context.Context, text string) error
}

type TextClipboardWriterFunc func(ctx context.Context, text string) error

func (fn TextClipboardWriterFunc) WriteText(ctx context.Context, text string) error {
	return fn(ctx, text)
}

type ImageClipboardWriter interface {
	WritePNGFile(ctx context.Context, path string) error
}

type ImageClipboardWriterFunc func(ctx context.Context, path string) error

func (fn ImageClipboardWriterFunc) WritePNGFile(ctx context.Context, path string) error {
	return fn(ctx, path)
}

type DirectExecutorConfig struct {
	DefaultDownloadDir string
	ReplyEnabled       bool
	Clipboard          TextClipboardWriter
	ImageClipboard     ImageClipboardWriter
}

type DirectExecutor struct {
	defaultDownloadDir string
	replyEnabled       bool
	clipboard          TextClipboardWriter
	imageClipboard     ImageClipboardWriter
}

func NewDirectExecutor(cfg DirectExecutorConfig) *DirectExecutor {
	return &DirectExecutor{
		defaultDownloadDir: cfg.DefaultDownloadDir,
		replyEnabled:       cfg.ReplyEnabled,
		clipboard:          cfg.Clipboard,
		imageClipboard:     cfg.ImageClipboard,
	}
}

func (executor *DirectExecutor) Execute(ctx context.Context, transport Transport, msg *Message, actions []DirectAction) error {
	var lastSavedPath string
	for _, action := range actions {
		switch action.Type {
		case DirectActionCopyText:
			if executor.clipboard == nil {
				return ErrClipboardWriterRequired
			}
			if err := executor.clipboard.WriteText(ctx, action.Text); err != nil {
				return err
			}
		case DirectActionSaveFile:
			data := action.Data
			if len(data) == 0 && action.ResourceRef != nil {
				if transport == nil {
					return ErrResourceDownloaderRequired
				}
				downloader, ok := transport.(ResourceCapable)
				if !ok {
					return ErrResourceDownloaderRequired
				}
				downloaded, err := downloader.Download(ctx, *action.ResourceRef)
				if err != nil {
					return err
				}
				data = downloaded
			}
			path, err := saveDirectFile(executor.defaultDownloadDir, action.FileName, data)
			if err != nil {
				return err
			}
			lastSavedPath = path
		case DirectActionCopyImage:
			if executor.imageClipboard == nil {
				return ErrImageClipboardRequired
			}
			if err := executor.imageClipboard.WritePNGFile(ctx, lastSavedPath); err != nil {
				return err
			}
		case DirectActionReply:
			if executor.replyEnabled && transport != nil && msg != nil && msg.ReplyCtx != nil {
				if replier, ok := transport.(ReplyCapable); ok {
					if err := replier.Reply(ctx, msg.ReplyCtx, action.Reply); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func saveDirectFile(directory string, fileName string, data []byte) (string, error) {
	if directory == "" {
		directory = "~/Downloads/copyagent"
	}
	directory = expandDirectHome(directory)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return "", err
	}
	path := uniqueDirectPath(filepath.Join(directory, safeDirectFileName(fileName, "file")))
	return path, os.WriteFile(path, data, 0o600)
}

func uniqueDirectPath(path string) string {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return path
	}
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	for index := 1; ; index++ {
		candidate := base + "-" + strconv.Itoa(index) + ext
		if _, err := os.Stat(candidate); errors.Is(err, os.ErrNotExist) {
			return candidate
		}
	}
}

func safeDirectFileName(name string, fallback string) string {
	cleaned := strings.TrimSpace(filepath.Base(name))
	cleaned = strings.Map(func(r rune) rune {
		if r < 32 {
			return -1
		}
		return r
	}, cleaned)
	if cleaned == "." || cleaned == string(filepath.Separator) || cleaned == "" {
		return fallback
	}
	return cleaned
}

func expandDirectHome(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, strings.TrimPrefix(path, "~/"))
		}
	}
	return path
}
