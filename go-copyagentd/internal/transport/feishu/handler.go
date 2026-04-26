package feishu

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/copyagent/copyagentd/internal/core"
	coreevent "github.com/copyagent/copyagentd/internal/event"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

const SuccessReplyText = "✅ 已复制到剪切板"
const FileSavedReplyText = "✅ 文件已保存"
const ImageCopiedReplyText = "✅ 图片已复制到剪切板"

type ClipboardWriter interface {
	WriteText(ctx context.Context, text string) error
}

type ClipboardFunc func(ctx context.Context, text string) error

func (fn ClipboardFunc) WriteText(ctx context.Context, text string) error {
	return fn(ctx, text)
}

type Replier interface {
	ReplyText(ctx context.Context, messageID string, text string) error
}

type ReplyFunc func(ctx context.Context, messageID string, text string) error

func (fn ReplyFunc) ReplyText(ctx context.Context, messageID string, text string) error {
	return fn(ctx, messageID, text)
}

type ResourceDownloader interface {
	Download(ctx context.Context, resource coreevent.ResourceMessage) ([]byte, error)
}

type DownloadFunc func(ctx context.Context, resource coreevent.ResourceMessage) ([]byte, error)

func (fn DownloadFunc) Download(ctx context.Context, resource coreevent.ResourceMessage) ([]byte, error) {
	return fn(ctx, resource)
}

type ImageClipboard interface {
	WritePNGFile(ctx context.Context, path string) error
}

type ImageClipboardFunc func(ctx context.Context, path string) error

func (fn ImageClipboardFunc) WritePNGFile(ctx context.Context, path string) error {
	return fn(ctx, path)
}

type MessageHandlerConfig struct {
	ReplyEnabled       bool
	AllowedActorIDs    []string
	DefaultDownloadDir string
	ImageAction        string
	Clipboard          ClipboardWriter
	ImageClipboard     ImageClipboard
	Downloader         ResourceDownloader
	Reply              Replier
}

type MessageHandler struct {
	replyEnabled       bool
	allowedActorIDs    map[string]struct{}
	defaultDownloadDir string
	imageAction        string
	clipboard          ClipboardWriter
	imageClipboard     ImageClipboard
	downloader         ResourceDownloader
	reply              Replier
}

func NewMessageHandler(cfg MessageHandlerConfig) *MessageHandler {
	return &MessageHandler{
		replyEnabled:       cfg.ReplyEnabled,
		allowedActorIDs:    allowedActorIDs(cfg.AllowedActorIDs),
		defaultDownloadDir: cfg.DefaultDownloadDir,
		imageAction:        normalizedImageAction(cfg.ImageAction),
		clipboard:          cfg.Clipboard,
		imageClipboard:     cfg.ImageClipboard,
		downloader:         cfg.Downloader,
		reply:              cfg.Reply,
	}
}

func (h *MessageHandler) Handle(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	if resource, ok := NormalizeResourceMessage(event); ok {
		return h.handleResource(ctx, resource)
	}
	normalized, ok := NormalizeTextMessage(event)
	if !ok {
		log.Printf("feishu message skipped: unsupported or empty message")
		return nil
	}
	if !h.actorAllowed(normalized.ActorID) {
		log.Printf("feishu message skipped: actor not allowed actor=%s message=%s", normalized.ActorID, normalized.MessageID)
		return nil
	}
	text := core.ExtractCopyText(normalized.Text)
	if !core.ValidText(text) {
		log.Printf("feishu message skipped: empty copy text message=%s actor=%s", normalized.MessageID, normalized.ActorID)
		return errors.New("text is required")
	}
	if h.clipboard == nil {
		return errors.New("clipboard writer is required")
	}
	if err := h.clipboard.WriteText(ctx, text); err != nil {
		return err
	}
	log.Printf("feishu message copied: message=%s actor=%s bytes=%d", normalized.MessageID, normalized.ActorID, len([]byte(text)))
	if h.replyEnabled && h.reply != nil && normalized.MessageID != "" {
		if err := h.reply.ReplyText(ctx, normalized.MessageID, SuccessReplyText); err != nil {
			log.Printf("feishu reply failed: message=%s err=%v", normalized.MessageID, err)
		} else {
			log.Printf("feishu reply sent: message=%s", normalized.MessageID)
		}
	}
	return nil
}

func (h *MessageHandler) handleResource(ctx context.Context, resource coreevent.ResourceMessage) error {
	if !h.actorAllowed(resource.ActorID) {
		log.Printf("feishu resource skipped: actor not allowed actor=%s message=%s", resource.ActorID, resource.MessageID)
		return nil
	}
	if strings.TrimSpace(resource.Key) == "" {
		return errors.New("resource key is required")
	}
	if h.downloader == nil {
		return errors.New("resource downloader is required")
	}
	data, err := h.downloader.Download(ctx, resource)
	if err != nil {
		return err
	}
	path, err := saveResourceFile(h.defaultDownloadDir, resource.FileName, data)
	if err != nil {
		return err
	}
	log.Printf("feishu resource saved: message=%s actor=%s kind=%s bytes=%d path=%s", resource.MessageID, resource.ActorID, resource.Kind, len(data), path)
	replyText := FileSavedReplyText
	if resource.Kind == "image" && h.imageAction != "save" {
		if h.imageClipboard == nil {
			return errors.New("image clipboard is required")
		}
		if err := h.imageClipboard.WritePNGFile(ctx, path); err != nil {
			return err
		}
		log.Printf("feishu image copied: message=%s actor=%s path=%s", resource.MessageID, resource.ActorID, path)
		replyText = ImageCopiedReplyText
	}
	if h.replyEnabled && h.reply != nil && resource.MessageID != "" {
		if err := h.reply.ReplyText(ctx, resource.MessageID, replyText); err != nil {
			log.Printf("feishu reply failed: message=%s err=%v", resource.MessageID, err)
		}
	}
	return nil
}

func normalizedImageAction(action string) string {
	if strings.EqualFold(strings.TrimSpace(action), "save") {
		return "save"
	}
	return "clipboard"
}

func (h *MessageHandler) actorAllowed(actorID string) bool {
	if len(h.allowedActorIDs) == 0 {
		return true
	}
	_, ok := h.allowedActorIDs[actorID]
	return ok
}

func allowedActorIDs(actorIDs []string) map[string]struct{} {
	if len(actorIDs) == 0 {
		return nil
	}
	allowed := make(map[string]struct{}, len(actorIDs))
	for _, actorID := range actorIDs {
		trimmed := strings.TrimSpace(actorID)
		if trimmed != "" {
			allowed[trimmed] = struct{}{}
		}
	}
	return allowed
}

func NormalizeTextMessage(event *larkim.P2MessageReceiveV1) (coreevent.TextMessage, bool) {
	if event == nil || event.Event == nil || event.Event.Message == nil {
		return coreevent.TextMessage{}, false
	}
	message := event.Event.Message
	if value(message.MessageType) != "text" {
		return coreevent.TextMessage{}, false
	}
	text, ok := parseTextContent(value(message.Content))
	if !ok {
		return coreevent.TextMessage{}, false
	}
	return coreevent.TextMessage{
		ActorID:   normalizeActorID(event.Event.Sender),
		MessageID: value(message.MessageId),
		Text:      text,
	}, true
}

func NormalizeResourceMessage(event *larkim.P2MessageReceiveV1) (coreevent.ResourceMessage, bool) {
	if event == nil || event.Event == nil || event.Event.Message == nil {
		return coreevent.ResourceMessage{}, false
	}
	message := event.Event.Message
	messageType := value(message.MessageType)
	if messageType != "image" && messageType != "file" {
		return coreevent.ResourceMessage{}, false
	}
	resource, ok := parseResourceContent(messageType, value(message.Content), value(message.MessageId))
	if !ok {
		return coreevent.ResourceMessage{}, false
	}
	resource.ActorID = normalizeActorID(event.Event.Sender)
	resource.MessageID = value(message.MessageId)
	return resource, true
}

func parseTextContent(content string) (string, bool) {
	var payload struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(content), &payload); err != nil {
		return "", false
	}
	return payload.Text, strings.TrimSpace(payload.Text) != ""
}

func parseResourceContent(kind string, content string, messageID string) (coreevent.ResourceMessage, bool) {
	var payload struct {
		ImageKey string `json:"image_key"`
		FileKey  string `json:"file_key"`
		FileName string `json:"file_name"`
	}
	if err := json.Unmarshal([]byte(content), &payload); err != nil {
		return coreevent.ResourceMessage{}, false
	}
	fallback := messageID
	if fallback == "" {
		fallback = kind
	}
	if kind == "image" {
		fallback += ".png"
	}
	key := payload.FileKey
	if kind == "image" {
		key = payload.ImageKey
	}
	return coreevent.ResourceMessage{Kind: kind, Key: key, FileName: safeFileName(payload.FileName, fallback)}, true
}

func saveResourceFile(directory string, fileName string, data []byte) (string, error) {
	if directory == "" {
		directory = "~/Downloads/copyagent"
	}
	directory = expandHome(directory)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return "", err
	}
	path := uniquePath(filepath.Join(directory, safeFileName(fileName, "file")))
	return path, os.WriteFile(path, data, 0o600)
}

func uniquePath(path string) string {
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

func safeFileName(name string, fallback string) string {
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

func expandHome(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, strings.TrimPrefix(path, "~/"))
		}
	}
	return path
}

func normalizeActorID(sender *larkim.EventSender) string {
	if sender == nil || sender.SenderId == nil {
		return ""
	}
	if sender.SenderId.OpenId != nil && *sender.SenderId.OpenId != "" {
		return *sender.SenderId.OpenId
	}
	if sender.SenderId.UserId != nil && *sender.SenderId.UserId != "" {
		return *sender.SenderId.UserId
	}
	return value(sender.SenderId.UnionId)
}

func value(text *string) string {
	if text == nil {
		return ""
	}
	return *text
}
