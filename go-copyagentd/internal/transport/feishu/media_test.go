package feishu

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	coreevent "github.com/copyagent/copyagentd/internal/event"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func TestNormalizeResourceMessage(t *testing.T) {
	event := resourceEvent("om_img", "image", `{"image_key":"img_key","file_name":"photo.png"}`)

	resource, ok := NormalizeResourceMessage(event)
	if !ok {
		t.Fatal("expected resource event")
	}
	if resource.Kind != "image" {
		t.Fatalf("kind = %q", resource.Kind)
	}
	if resource.Key != "img_key" {
		t.Fatalf("key = %q", resource.Key)
	}
	if resource.FileName != "photo.png" {
		t.Fatalf("file name = %q", resource.FileName)
	}
	if resource.MessageID != "om_img" {
		t.Fatalf("message id = %q", resource.MessageID)
	}
}

func TestMediaHandlerSavesWithoutOverwrite(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "report.txt"), []byte("old"), 0o600); err != nil {
		t.Fatalf("seed: %v", err)
	}
	var replyTarget string
	handler := NewMessageHandler(MessageHandlerConfig{
		DefaultDownloadDir: dir,
		ReplyEnabled:       true,
		Downloader: DownloadFunc(func(_ context.Context, resource coreevent.ResourceMessage) ([]byte, error) {
			return []byte("new"), nil
		}),
		Reply: ReplyFunc(func(_ context.Context, messageID string, text string) error {
			replyTarget = messageID
			return nil
		}),
	})

	if err := handler.Handle(context.Background(), resourceEvent("om_file", "file", `{"file_key":"file_key","file_name":"report.txt"}`)); err != nil {
		t.Fatalf("handle: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "report-1.txt"))
	if err != nil {
		t.Fatalf("read saved file: %v", err)
	}
	if string(data) != "new" {
		t.Fatalf("saved data = %q", data)
	}
	if replyTarget != "om_file" {
		t.Fatalf("reply target = %q", replyTarget)
	}
}

func TestImageHandlerCopiesImageByDefault(t *testing.T) {
	dir := t.TempDir()
	var copiedPath string
	var replyText string
	handler := NewMessageHandler(MessageHandlerConfig{
		DefaultDownloadDir: dir,
		ReplyEnabled:       true,
		Downloader: DownloadFunc(func(_ context.Context, resource coreevent.ResourceMessage) ([]byte, error) {
			return []byte("png"), nil
		}),
		ImageClipboard: ImageClipboardFunc(func(_ context.Context, path string) error {
			copiedPath = path
			return nil
		}),
		Reply: ReplyFunc(func(_ context.Context, messageID string, text string) error {
			replyText = text
			return nil
		}),
	})

	if err := handler.Handle(context.Background(), resourceEvent("om_img", "image", `{"image_key":"img_key","file_name":"photo.png"}`)); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if copiedPath != filepath.Join(dir, "photo.png") {
		t.Fatalf("copied path = %q", copiedPath)
	}
	if replyText != ImageCopiedReplyText {
		t.Fatalf("reply text = %q", replyText)
	}
}

func TestImageHandlerSaveModeSkipsClipboard(t *testing.T) {
	called := false
	handler := NewMessageHandler(MessageHandlerConfig{
		DefaultDownloadDir: t.TempDir(),
		ImageAction:        "save",
		Downloader: DownloadFunc(func(_ context.Context, resource coreevent.ResourceMessage) ([]byte, error) {
			return []byte("png"), nil
		}),
		ImageClipboard: ImageClipboardFunc(func(_ context.Context, path string) error {
			called = true
			return nil
		}),
	})

	if err := handler.Handle(context.Background(), resourceEvent("om_img", "image", `{"image_key":"img_key","file_name":"photo.png"}`)); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if called {
		t.Fatal("save mode should not copy image to clipboard")
	}
}

func TestImageHandlerRequiresClipboardInClipboardMode(t *testing.T) {
	handler := NewMessageHandler(MessageHandlerConfig{
		DefaultDownloadDir: t.TempDir(),
		ImageAction:        "clipboard",
		Downloader: DownloadFunc(func(_ context.Context, resource coreevent.ResourceMessage) ([]byte, error) {
			return []byte("png"), nil
		}),
	})

	if err := handler.Handle(context.Background(), resourceEvent("om_img", "image", `{"image_key":"img_key","file_name":"photo.png"}`)); err == nil {
		t.Fatal("expected image clipboard error")
	}
}

func TestMediaHandlerRejectsMissingKey(t *testing.T) {
	called := false
	handler := NewMessageHandler(MessageHandlerConfig{
		DefaultDownloadDir: t.TempDir(),
		Downloader: DownloadFunc(func(_ context.Context, resource coreevent.ResourceMessage) ([]byte, error) {
			called = true
			return []byte("new"), nil
		}),
	})

	if err := handler.Handle(context.Background(), resourceEvent("om_file", "file", `{"file_name":"report.txt"}`)); err == nil {
		t.Fatal("expected missing resource key error")
	}
	if called {
		t.Fatal("missing key should not download")
	}
}

func resourceEvent(messageID string, messageType string, content string) *larkim.P2MessageReceiveV1 {
	return &larkim.P2MessageReceiveV1{Event: &larkim.P2MessageReceiveV1Data{
		Sender: &larkim.EventSender{SenderId: larkim.NewUserIdBuilder().OpenId("ou_test").Build()},
		Message: larkim.NewEventMessageBuilder().
			MessageId(messageID).
			MessageType(messageType).
			Content(content).
			Build(),
	}}
}
