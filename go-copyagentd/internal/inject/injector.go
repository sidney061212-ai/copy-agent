package inject

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	DefaultMaxBytes = 16 * 1024
)

var (
	ErrEmptyText         = errors.New("inject text is required")
	ErrInvalidUTF8       = errors.New("inject text must be valid utf-8")
	ErrTextTooLarge      = errors.New("inject text is too large")
	ErrUnsupportedTarget = errors.New("frontmost app is not allowed")
)

type Request struct {
	Text             string
	Submit           bool
	TargetBundleID   string
	AllowedBundleIDs []string
	MaxBytes         int
}

type Target struct {
	AppName     string
	BundleID    string
	WindowTitle string
}

type Result struct {
	Target              Target
	RestoredClipboard   bool
	ClipboardRestoreErr error
	Warning             string
}

type Snapshot struct {
	Text string
	OK   bool
}

type FrontmostInspector interface {
	Frontmost(context.Context) (Target, error)
}

type Activator interface {
	ActivateBundle(context.Context, string) error
}

type Pasteboard interface {
	Snapshot(context.Context) (Snapshot, error)
	WriteText(context.Context, string) error
	Restore(context.Context, Snapshot) error
}

type Keystroker interface {
	Paste(context.Context) error
	Submit(context.Context) error
}

type Executor interface {
	InjectText(context.Context, Request) (Result, error)
	Status(context.Context) (Target, error)
	Activate(context.Context, string) (Target, error)
}

type Service struct {
	Inspector  FrontmostInspector
	Pasteboard Pasteboard
	Keystroker Keystroker
	Activator  Activator
}

func NewService(inspector FrontmostInspector, pasteboard Pasteboard, keystroker Keystroker) *Service {
	return &Service{Inspector: inspector, Pasteboard: pasteboard, Keystroker: keystroker}
}

func (service *Service) InjectText(ctx context.Context, req Request) (Result, error) {
	if err := ValidateRequest(req); err != nil {
		return Result{}, err
	}
	if service == nil || service.Inspector == nil || service.Pasteboard == nil || service.Keystroker == nil {
		return Result{}, errors.New("inject service is not configured")
	}
	target, err := service.Inspector.Frontmost(ctx)
	if err != nil {
		return Result{}, err
	}
	if strings.TrimSpace(req.TargetBundleID) != "" && !strings.EqualFold(target.BundleID, strings.TrimSpace(req.TargetBundleID)) {
		if service.Activator == nil {
			return Result{Target: target}, ErrUnsupportedTarget
		}
		if err := service.Activator.ActivateBundle(ctx, req.TargetBundleID); err != nil {
			return Result{Target: target}, err
		}
		target, err = service.Inspector.Frontmost(ctx)
		if err != nil {
			return Result{}, err
		}
		if !strings.EqualFold(target.BundleID, strings.TrimSpace(req.TargetBundleID)) {
			return Result{Target: target}, ErrUnsupportedTarget
		}
	}
	if !BundleAllowed(target.BundleID, req.AllowedBundleIDs) {
		return Result{Target: target}, ErrUnsupportedTarget
	}
	snapshot, snapshotErr := service.Pasteboard.Snapshot(ctx)
	if snapshotErr != nil {
		return Result{}, fmt.Errorf("snapshot pasteboard: %w", snapshotErr)
	}
	if err := service.Pasteboard.WriteText(ctx, req.Text); err != nil {
		return Result{Target: target}, fmt.Errorf("write pasteboard: %w", err)
	}
	if err := service.Keystroker.Paste(ctx); err != nil {
		_ = service.Pasteboard.Restore(ctx, snapshot)
		return Result{Target: target}, fmt.Errorf("paste keystroke: %w", err)
	}
	if req.Submit {
		if err := service.Keystroker.Submit(ctx); err != nil {
			_ = service.Pasteboard.Restore(ctx, snapshot)
			return Result{Target: target}, fmt.Errorf("submit keystroke: %w", err)
		}
	}
	result := Result{Target: target}
	if snapshot.OK {
		if err := service.Pasteboard.Restore(ctx, snapshot); err != nil {
			result.ClipboardRestoreErr = err
			result.Warning = "恢复原剪切板失败：" + err.Error()
		} else {
			result.RestoredClipboard = true
		}
	}
	return result, nil
}

func (service *Service) Status(ctx context.Context) (Target, error) {
	if service == nil || service.Inspector == nil {
		return Target{}, errors.New("inject service is not configured")
	}
	return service.Inspector.Frontmost(ctx)
}

func (service *Service) Activate(ctx context.Context, bundleID string) (Target, error) {
	if strings.TrimSpace(bundleID) == "" {
		return Target{}, ErrUnsupportedTarget
	}
	if service == nil || service.Inspector == nil || service.Activator == nil {
		return Target{}, errors.New("inject service activation is not configured")
	}
	if !BundleAllowed(bundleID, DefaultAllowedBundleIDs()) {
		return Target{BundleID: bundleID}, ErrUnsupportedTarget
	}
	if err := service.Activator.ActivateBundle(ctx, bundleID); err != nil {
		return Target{}, err
	}
	target, err := service.Inspector.Frontmost(ctx)
	if err != nil {
		return Target{}, err
	}
	if !strings.EqualFold(target.BundleID, strings.TrimSpace(bundleID)) {
		return target, ErrUnsupportedTarget
	}
	return target, nil
}

func ValidateRequest(req Request) error {
	text := strings.TrimSpace(req.Text)
	if text == "" {
		return ErrEmptyText
	}
	if !utf8.ValidString(req.Text) {
		return ErrInvalidUTF8
	}
	maxBytes := req.MaxBytes
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBytes
	}
	if len([]byte(req.Text)) > maxBytes {
		return ErrTextTooLarge
	}
	return nil
}

func ParseCommand(content string) (string, bool) {
	trimmed := strings.TrimSpace(content)
	for _, command := range []string{"/inject", "／inject"} {
		if trimmed == command {
			return "", true
		}
		for _, prefix := range []string{command + " ", command + "\n", command + "\t"} {
			if strings.HasPrefix(trimmed, prefix) {
				return strings.TrimSpace(strings.TrimPrefix(trimmed, command)), true
			}
		}
	}
	return "", false
}

func BundleAllowed(bundleID string, allowed []string) bool {
	bundleID = strings.TrimSpace(bundleID)
	if bundleID == "" {
		return false
	}
	for _, item := range allowed {
		if strings.EqualFold(bundleID, strings.TrimSpace(item)) {
			return true
		}
	}
	return false
}

func DefaultAllowedBundleIDs() []string {
	return []string{
		"com.openai.codex",
		"com.apple.Terminal",
		"com.googlecode.iterm2",
		"dev.warp.Warp-Stable",
		"com.anthropic.claude",
		"com.microsoft.VSCode",
		"com.todesktop.230313mzl4w4u92",
		"com.jetbrains.intellij",
		"com.jetbrains.pycharm",
		"com.jetbrains.goland",
	}
}

type NamedTarget struct {
	Name     string
	BundleID string
}

func DefaultNamedTargets() []NamedTarget {
	return []NamedTarget{
		{Name: "codex", BundleID: "com.openai.codex"},
		{Name: "claude", BundleID: "com.anthropic.claude"},
		{Name: "terminal", BundleID: "com.apple.Terminal"},
		{Name: "iterm", BundleID: "com.googlecode.iterm2"},
		{Name: "warp", BundleID: "dev.warp.Warp-Stable"},
		{Name: "vscode", BundleID: "com.microsoft.VSCode"},
		{Name: "cursor", BundleID: "com.todesktop.230313mzl4w4u92"},
	}
}

func ResolveNamedTarget(name string) (NamedTarget, bool) {
	clean := strings.ToLower(strings.TrimSpace(name))
	for _, target := range DefaultNamedTargets() {
		if clean == target.Name || clean == strings.ToLower(target.BundleID) {
			return target, true
		}
	}
	return NamedTarget{}, false
}
