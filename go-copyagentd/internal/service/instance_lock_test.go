//go:build !windows

package service

import "testing"

func TestAcquireInstanceLockSuccess(t *testing.T) {
	lock, err := AcquireInstanceLock(t.TempDir() + "/config.json")
	if err != nil {
		t.Fatalf("acquire lock: %v", err)
	}
	if lock == nil || lock.Path() == "" {
		t.Fatal("expected a non-empty lock path")
	}
	lock.Release()
}

func TestAcquireInstanceLockAlreadyLocked(t *testing.T) {
	configPath := t.TempDir() + "/config.json"
	first, err := AcquireInstanceLock(configPath)
	if err != nil {
		t.Fatalf("first acquire lock: %v", err)
	}
	defer first.Release()

	if _, err := AcquireInstanceLock(configPath); err == nil {
		t.Fatal("expected second lock attempt to fail")
	}
}
