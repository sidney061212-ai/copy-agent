package agent

import "testing"

func TestMessageHasStableSessionKeyFallback(t *testing.T) {
	msg := Message{Platform: "feishu", UserID: "ou_123"}
	if got := msg.EffectiveSessionKey(); got != "feishu:ou_123" {
		t.Fatalf("EffectiveSessionKey() = %q", got)
	}
}

func TestMessagePrefersExplicitSessionKey(t *testing.T) {
	msg := Message{SessionKey: "feishu:chat:user", Platform: "feishu", UserID: "ou_123"}
	if got := msg.EffectiveSessionKey(); got != "feishu:chat:user" {
		t.Fatalf("EffectiveSessionKey() = %q", got)
	}
}
