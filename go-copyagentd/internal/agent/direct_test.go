package agent

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type mockTextClipboard struct {
	texts []string
	err   error
}

func (clipboard *mockTextClipboard) WriteText(_ context.Context, text string) error {
	clipboard.texts = append(clipboard.texts, text)
	return clipboard.err
}

type mockImageClipboard struct {
	paths []string
	err   error
}

func (clipboard *mockImageClipboard) WritePNGFile(_ context.Context, path string) error {
	clipboard.paths = append(clipboard.paths, path)
	return clipboard.err
}

type mockReplyTransport struct {
	name    string
	replies []string
	ctxs    []any
	data    []byte
	refs    []ResourceRef
	err     error
}

func (transport *mockReplyTransport) Name() string { return transport.name }
func (transport *mockReplyTransport) Start(MessageHandler) error {
	return nil
}
func (transport *mockReplyTransport) Stop() error { return nil }
func (transport *mockReplyTransport) Reply(_ context.Context, replyCtx any, content string) error {
	transport.ctxs = append(transport.ctxs, replyCtx)
	transport.replies = append(transport.replies, content)
	return transport.err
}
func (transport *mockReplyTransport) Download(_ context.Context, ref ResourceRef) ([]byte, error) {
	transport.refs = append(transport.refs, ref)
	return transport.data, transport.err
}

func TestDirectPolicyAllowsAllowedUserAndDedupesMessageID(t *testing.T) {
	policy := NewDirectPolicy(DirectPolicyConfig{AllowedUserIDs: []string{"ou_allowed"}, MaxTextBytes: 20})
	msg := &Message{Platform: "feishu", MessageID: "om_1", UserID: "ou_allowed", Content: "copy hi"}
	allowed, err := policy.Allow(msg)
	if err != nil || !allowed {
		t.Fatalf("first allow = %v, %v", allowed, err)
	}
	policy.Complete(msg, true)
	allowed, err = policy.Allow(msg)
	if err != nil || allowed {
		t.Fatalf("duplicate allow = %v, %v", allowed, err)
	}
}

func TestDirectPolicyAllowsRetryAfterFailure(t *testing.T) {
	policy := NewDirectPolicy(DirectPolicyConfig{})
	msg := &Message{Platform: "feishu", MessageID: "om_1", Content: "copy hi"}
	allowed, err := policy.Allow(msg)
	if err != nil || !allowed {
		t.Fatalf("first allow = %v, %v", allowed, err)
	}
	policy.Complete(msg, false)
	allowed, err = policy.Allow(msg)
	if err != nil || !allowed {
		t.Fatalf("retry allow = %v, %v", allowed, err)
	}
}

func TestDirectPolicySkipsDisallowedUser(t *testing.T) {
	policy := NewDirectPolicy(DirectPolicyConfig{AllowedUserIDs: []string{"ou_allowed"}})
	allowed, err := policy.Allow(&Message{UserID: "ou_other"})
	if err != nil || allowed {
		t.Fatalf("expected skip, got allowed=%v err=%v", allowed, err)
	}
}

func TestDirectPolicyRejectsOversizedText(t *testing.T) {
	policy := NewDirectPolicy(DirectPolicyConfig{MaxTextBytes: 3})
	allowed, err := policy.Allow(&Message{Content: "hello"})
	if allowed || !errors.Is(err, ErrMessageTextTooLarge) {
		t.Fatalf("expected text too large, allowed=%v err=%v", allowed, err)
	}
}

func TestDirectPlannerPlansCopyTextAndReply(t *testing.T) {
	planner := NewDirectPlanner(DirectPlannerConfig{})
	actions, err := planner.Plan(&Message{Content: "复制：你好"})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if len(actions) != 2 || actions[0].Type != DirectActionCopyText || actions[0].Text != "你好" || actions[1].Reply != DirectCopySuccessReplyText {
		t.Fatalf("unexpected actions: %#v", actions)
	}
}

func TestDirectPlannerRejectsBlankCopyCommand(t *testing.T) {
	planner := NewDirectPlanner(DirectPlannerConfig{})
	_, err := planner.Plan(&Message{Content: "复制：   "})
	if !errors.Is(err, ErrEmptyCopyText) {
		t.Fatalf("expected ErrEmptyCopyText, got %v", err)
	}
}

func TestDirectPlannerRejectsResourceAttachmentWithoutDataOrKey(t *testing.T) {
	planner := NewDirectPlanner(DirectPlannerConfig{})
	_, err := planner.Plan(&Message{Files: []FileAttachment{{FileName: "report.txt"}}})
	if !errors.Is(err, ErrResourceKeyRequired) {
		t.Fatalf("expected ErrResourceKeyRequired, got %v", err)
	}
}

func TestDirectPlannerPlansFilesAndImages(t *testing.T) {
	planner := NewDirectPlanner(DirectPlannerConfig{})
	actions, err := planner.Plan(&Message{
		Files:  []FileAttachment{{FileName: "report.txt", Data: []byte("file")}},
		Images: []ImageAttachment{{FileName: "photo.png", Data: []byte("png")}},
	})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	wantTypes := []DirectActionType{DirectActionSaveFile, DirectActionReply, DirectActionSaveFile, DirectActionCopyImage, DirectActionReply}
	if len(actions) != len(wantTypes) {
		t.Fatalf("action count = %d, actions=%#v", len(actions), actions)
	}
	for i, want := range wantTypes {
		if actions[i].Type != want {
			t.Fatalf("action[%d] = %s, want %s", i, actions[i].Type, want)
		}
	}
}

func TestDirectPlannerSaveModeSkipsImageClipboard(t *testing.T) {
	planner := NewDirectPlanner(DirectPlannerConfig{ImageAction: "save"})
	actions, err := planner.Plan(&Message{Images: []ImageAttachment{{FileName: "photo.png", Data: []byte("png")}}})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if len(actions) != 2 || actions[0].Type != DirectActionSaveFile || actions[1].Reply != DirectFileSavedReplyText {
		t.Fatalf("unexpected actions: %#v", actions)
	}
}

func TestDirectExecutorCopiesTextAndReplies(t *testing.T) {
	clipboard := &mockTextClipboard{}
	transport := &mockReplyTransport{name: "feishu"}
	executor := NewDirectExecutor(DirectExecutorConfig{ReplyEnabled: true, Clipboard: clipboard})
	err := executor.Execute(context.Background(), transport, &Message{ReplyCtx: "om_1"}, []DirectAction{
		{Type: DirectActionCopyText, Text: "hello"},
		{Type: DirectActionReply, Reply: DirectCopySuccessReplyText},
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if len(clipboard.texts) != 1 || clipboard.texts[0] != "hello" {
		t.Fatalf("clipboard texts = %#v", clipboard.texts)
	}
	if len(transport.replies) != 1 || transport.replies[0] != DirectCopySuccessReplyText || transport.ctxs[0] != "om_1" {
		t.Fatalf("replies = %#v ctxs=%#v", transport.replies, transport.ctxs)
	}
}

func TestDirectExecutorSavesWithoutOverwriteAndCopiesImage(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "photo.png"), []byte("old"), 0o600); err != nil {
		t.Fatalf("seed: %v", err)
	}
	imageClipboard := &mockImageClipboard{}
	executor := NewDirectExecutor(DirectExecutorConfig{DefaultDownloadDir: dir, ImageClipboard: imageClipboard})
	err := executor.Execute(context.Background(), nil, nil, []DirectAction{
		{Type: DirectActionSaveFile, FileName: "photo.png", Data: []byte("new")},
		{Type: DirectActionCopyImage},
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	path := filepath.Join(dir, "photo-1.png")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read saved image: %v", err)
	}
	if string(data) != "new" {
		t.Fatalf("saved data = %q", data)
	}
	if len(imageClipboard.paths) != 1 || imageClipboard.paths[0] != path {
		t.Fatalf("image clipboard paths = %#v", imageClipboard.paths)
	}
}

func TestDirectExecutorDownloadsResourceAttachments(t *testing.T) {
	dir := t.TempDir()
	transport := &mockReplyTransport{name: "feishu", data: []byte("downloaded")}
	executor := NewDirectExecutor(DirectExecutorConfig{DefaultDownloadDir: dir})
	resourceRef := &ResourceRef{Platform: "feishu", MessageID: "om_file", Key: "file_key", Kind: "file", FileName: "report.txt"}
	err := executor.Execute(context.Background(), transport, nil, []DirectAction{{Type: DirectActionSaveFile, FileName: "report.txt", ResourceRef: resourceRef}})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "report.txt"))
	if err != nil {
		t.Fatalf("read saved file: %v", err)
	}
	if string(data) != "downloaded" {
		t.Fatalf("saved data = %q", data)
	}
	if len(transport.refs) != 1 || transport.refs[0].Key != "file_key" {
		t.Fatalf("refs = %#v", transport.refs)
	}
}

func TestDirectHandlerRunsThroughEngine(t *testing.T) {
	clipboard := &mockTextClipboard{}
	transport := &emittingTransport{name: "feishu", message: Message{Platform: "feishu", MessageID: "om_1", UserID: "ou_allowed", Content: "copy hi"}}
	handler := NewDirectHandler(DirectHandlerConfig{
		Policy:   DirectPolicyConfig{AllowedUserIDs: []string{"ou_allowed"}},
		Executor: DirectExecutorConfig{Clipboard: clipboard},
	})
	engine := NewDirectEngine("direct", []Transport{transport}, handler)
	if err := engine.Start(context.Background()); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	if len(clipboard.texts) != 1 || clipboard.texts[0] != "hi" {
		t.Fatalf("clipboard texts = %#v", clipboard.texts)
	}
}
