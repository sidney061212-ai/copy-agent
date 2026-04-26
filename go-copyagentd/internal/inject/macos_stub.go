//go:build !darwin

package inject

import (
	"context"
	"errors"
)

func NewDefaultService() *Service {
	service := NewService(unsupportedInspector{}, unsupportedPasteboard{}, unsupportedKeystroker{})
	service.Activator = unsupportedActivator{}
	return service
}

type unsupportedInspector struct{}
type unsupportedPasteboard struct{}
type unsupportedKeystroker struct{}
type unsupportedActivator struct{}

func (unsupportedInspector) Frontmost(context.Context) (Target, error) {
	return Target{}, errors.New("foreground injection is only supported on macOS")
}
func (unsupportedPasteboard) Snapshot(context.Context) (Snapshot, error) {
	return Snapshot{}, errors.New("foreground injection is only supported on macOS")
}
func (unsupportedPasteboard) WriteText(context.Context, string) error {
	return errors.New("foreground injection is only supported on macOS")
}
func (unsupportedPasteboard) Restore(context.Context, Snapshot) error {
	return errors.New("foreground injection is only supported on macOS")
}
func (unsupportedKeystroker) Paste(context.Context) error {
	return errors.New("foreground injection is only supported on macOS")
}
func (unsupportedKeystroker) Submit(context.Context) error {
	return errors.New("foreground injection is only supported on macOS")
}
func (unsupportedActivator) ActivateBundle(context.Context, string) error {
	return errors.New("foreground injection is only supported on macOS")
}
