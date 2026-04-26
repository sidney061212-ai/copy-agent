package inject

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestParseCommand(t *testing.T) {
	text, ok := ParseCommand("/inject hello")
	if !ok || text != "hello" {
		t.Fatalf("ParseCommand = %q %v", text, ok)
	}
	text, ok = ParseCommand("／inject hello")
	if !ok || text != "hello" {
		t.Fatalf("ParseCommand fullwidth = %q %v", text, ok)
	}
	text, ok = ParseCommand("/inject\nhello\nworld")
	if !ok || text != "hello\nworld" {
		t.Fatalf("ParseCommand multiline = %q %v", text, ok)
	}
	text, ok = ParseCommand("/inject")
	if !ok || text != "" {
		t.Fatalf("ParseCommand empty = %q %v", text, ok)
	}
	if _, ok := ParseCommand("/inject-now hello"); ok {
		t.Fatal("ParseCommand should reject non-command prefix")
	}
}

func TestValidateRequest(t *testing.T) {
	if err := ValidateRequest(Request{Text: "hello", MaxBytes: 10}); err != nil {
		t.Fatalf("ValidateRequest returned error: %v", err)
	}
	if !errors.Is(ValidateRequest(Request{Text: "   "}), ErrEmptyText) {
		t.Fatal("expected ErrEmptyText")
	}
	if !errors.Is(ValidateRequest(Request{Text: strings.Repeat("x", 4), MaxBytes: 3}), ErrTextTooLarge) {
		t.Fatal("expected ErrTextTooLarge")
	}
}

func TestServiceInjectsAndRestoresClipboard(t *testing.T) {
	pasteboard := &fakePasteboard{snapshot: Snapshot{Text: "before", OK: true}}
	keystroker := &fakeKeystroker{}
	service := NewService(fakeInspector{target: Target{AppName: "Terminal", BundleID: "com.apple.Terminal"}}, pasteboard, keystroker)
	result, err := service.InjectText(context.Background(), Request{Text: "hello", AllowedBundleIDs: []string{"com.apple.Terminal"}})
	if err != nil {
		t.Fatalf("InjectText returned error: %v", err)
	}
	if pasteboard.wrote != "hello" || pasteboard.restored.Text != "before" || !keystroker.pasted {
		t.Fatalf("pasteboard=%#v keystroker=%#v", pasteboard, keystroker)
	}
	if !result.RestoredClipboard || result.Target.AppName != "Terminal" {
		t.Fatalf("result = %#v", result)
	}
}

func TestServiceInjectsAndSubmits(t *testing.T) {
	keystroker := &fakeKeystroker{}
	service := NewService(fakeInspector{target: Target{AppName: "Terminal", BundleID: "com.apple.Terminal"}}, &fakePasteboard{snapshot: Snapshot{Text: "before", OK: true}}, keystroker)
	_, err := service.InjectText(context.Background(), Request{Text: "hello", Submit: true, AllowedBundleIDs: []string{"com.apple.Terminal"}})
	if err != nil {
		t.Fatalf("InjectText returned error: %v", err)
	}
	if !keystroker.pasted || !keystroker.submitted {
		t.Fatalf("keystroker=%#v", keystroker)
	}
}

func TestServiceRejectsUnsupportedTargetBeforeClipboardWrite(t *testing.T) {
	pasteboard := &fakePasteboard{}
	service := NewService(fakeInspector{target: Target{AppName: "Notes", BundleID: "com.apple.Notes"}}, pasteboard, &fakeKeystroker{})
	result, err := service.InjectText(context.Background(), Request{Text: "hello", AllowedBundleIDs: []string{"com.apple.Terminal"}})
	if !errors.Is(err, ErrUnsupportedTarget) {
		t.Fatalf("expected ErrUnsupportedTarget, got %v", err)
	}
	if result.Target.AppName != "Notes" {
		t.Fatalf("result target = %#v", result.Target)
	}
	if pasteboard.wrote != "" {
		t.Fatalf("should not write pasteboard, wrote %q", pasteboard.wrote)
	}
}

func TestServiceReturnsRestoreWarning(t *testing.T) {
	pasteboard := &fakePasteboard{snapshot: Snapshot{Text: "before", OK: true}, restoreErr: errors.New("restore failed")}
	service := NewService(fakeInspector{target: Target{AppName: "Terminal", BundleID: "com.apple.Terminal"}}, pasteboard, &fakeKeystroker{})
	result, err := service.InjectText(context.Background(), Request{Text: "hello", AllowedBundleIDs: []string{"com.apple.Terminal"}})
	if err != nil {
		t.Fatalf("InjectText returned error: %v", err)
	}
	if result.Warning == "" || result.RestoredClipboard {
		t.Fatalf("result = %#v", result)
	}
}

type fakeInspector struct {
	target Target
	err    error
}

func (inspector fakeInspector) Frontmost(context.Context) (Target, error) {
	return inspector.target, inspector.err
}

type fakePasteboard struct {
	snapshot    Snapshot
	wrote       string
	restored    Snapshot
	snapshotErr error
	writeErr    error
	restoreErr  error
}

func (pasteboard *fakePasteboard) Snapshot(context.Context) (Snapshot, error) {
	return pasteboard.snapshot, pasteboard.snapshotErr
}
func (pasteboard *fakePasteboard) WriteText(_ context.Context, text string) error {
	pasteboard.wrote = text
	return pasteboard.writeErr
}
func (pasteboard *fakePasteboard) Restore(_ context.Context, snapshot Snapshot) error {
	pasteboard.restored = snapshot
	return pasteboard.restoreErr
}

type fakeKeystroker struct {
	pasted    bool
	submitted bool
	err       error
}

func (keystroker *fakeKeystroker) Paste(context.Context) error {
	keystroker.pasted = true
	return keystroker.err
}

func (keystroker *fakeKeystroker) Submit(context.Context) error {
	keystroker.submitted = true
	return keystroker.err
}
