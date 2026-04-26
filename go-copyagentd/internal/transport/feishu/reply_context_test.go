package feishu

import "testing"

func TestReconstructReplyContextChatSession(t *testing.T) {
	ctx, err := ReconstructReplyContext("feishu:oc_chat:ou_user")
	if err != nil {
		t.Fatalf("ReconstructReplyContext returned error: %v", err)
	}
	rctx, ok := ctx.(replyContext)
	if !ok || rctx.chatID != "oc_chat" || rctx.messageID != "" || rctx.sessionKey != "feishu:oc_chat:ou_user" {
		t.Fatalf("reply context = %#v", ctx)
	}
}

func TestReconstructReplyContextThreadSession(t *testing.T) {
	ctx, err := ReconstructReplyContext("feishu:oc_chat:thread:omt_thread")
	if err != nil {
		t.Fatalf("ReconstructReplyContext returned error: %v", err)
	}
	rctx, ok := ctx.(replyContext)
	if !ok || rctx.chatID != "oc_chat" || rctx.messageID != "omt_thread" || rctx.sessionKey != "feishu:oc_chat:thread:omt_thread" {
		t.Fatalf("reply context = %#v", ctx)
	}
}

func TestReconstructReplyContextRejectsInvalid(t *testing.T) {
	if _, err := ReconstructReplyContext("slack:C:U"); err == nil {
		t.Fatal("expected invalid session key error")
	}
}
