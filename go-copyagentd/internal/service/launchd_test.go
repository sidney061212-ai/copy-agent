package service

import (
	"strings"
	"testing"
)

func TestLaunchdPlistContainsNoSecrets(t *testing.T) {
	plist := LaunchdPlist(Config{
		BinaryPath: "/tmp/copyagentd",
		ConfigPath: "/tmp/config.json",
		WorkDir:    "/tmp",
		LogFile:    "/tmp/copyagentd.log",
		LogMaxSize: DefaultLogMaxSize,
		EnvPATH:    "/usr/bin:/bin",
	})
	if !strings.HasPrefix(plist, "<?xml") {
		t.Fatalf("plist should start with xml declaration, got %q", plist[:min(len(plist), 8)])
	}

	for _, forbidden := range []string{"feishuAppSecret", "token", "feishuEncryptKey", "app_secret"} {
		if strings.Contains(plist, forbidden) {
			t.Fatalf("plist contains secret-like key %q", forbidden)
		}
	}
	for _, expected := range []string{"com.copyagent.copyagentd", "/tmp/copyagentd", "feishu-serve", "/tmp/copyagentd.log", "/tmp/config.json", "/tmp"} {
		if !strings.Contains(plist, expected) {
			t.Fatalf("plist missing %q", expected)
		}
	}
	for _, expected := range []string{"EnvironmentVariables", "LANG", "LC_ALL", "UTF-8"} {
		if !strings.Contains(plist, expected) {
			t.Fatalf("plist missing UTF-8 environment %q", expected)
		}
	}
	for _, expected := range []string{LogFileEnv, LogMaxSizeEnv, ConfigPathEnv, "<key>SuccessfulExit</key>", "<string>/dev/null</string>"} {
		if !strings.Contains(plist, expected) {
			t.Fatalf("plist missing daemon env/launchd guard %q", expected)
		}
	}
	if strings.Contains(plist, "<key>KeepAlive</key>\n  <true/>") {
		t.Fatal("plist must not use boolean KeepAlive true")
	}
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
