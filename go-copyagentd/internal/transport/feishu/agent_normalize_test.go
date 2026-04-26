package feishu

import (
	"testing"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func TestNormalizeAgentMessageText(t *testing.T) {
	event := agentTextEvent("om_text", "oc_chat", "", "", "ou_test", "复制：你好")
	msg, ok := NormalizeAgentMessage(event)
	if !ok {
		t.Fatal("expected agent message")
	}
	if msg.Platform != "feishu" || msg.MessageID != "om_text" || msg.UserID != "ou_test" || msg.ChatID != "oc_chat" {
		t.Fatalf("unexpected identity: %#v", msg)
	}
	if msg.Content != "复制：你好" {
		t.Fatalf("content = %q", msg.Content)
	}
	rctx, ok := msg.ReplyCtx.(replyContext)
	if !ok || rctx.messageID != "om_text" || rctx.chatID != "oc_chat" || rctx.sessionKey != "feishu:oc_chat:ou_test" {
		t.Fatalf("reply ctx = %#v", msg.ReplyCtx)
	}
	if msg.SessionKey != "feishu:oc_chat:ou_test" {
		t.Fatalf("session key = %q", msg.SessionKey)
	}
}

func TestNormalizeAgentMessageWithoutMessageIDHasNoReplyCtx(t *testing.T) {
	event := agentTextEvent("", "oc_chat", "", "", "ou_test", "copy hi")
	msg, ok := NormalizeAgentMessage(event)
	if !ok {
		t.Fatal("expected agent message")
	}
	if msg.ReplyCtx != nil {
		t.Fatalf("reply ctx = %#v", msg.ReplyCtx)
	}
}

func TestNormalizeAgentMessageThreadSessionKey(t *testing.T) {
	event := agentTextEvent("om_text", "oc_chat", "omt_root", "omt_thread", "ou_test", "copy hi")
	msg, ok := NormalizeAgentMessage(event)
	if !ok {
		t.Fatal("expected agent message")
	}
	if msg.SessionKey != "feishu:oc_chat:thread:omt_thread" {
		t.Fatalf("session key = %q", msg.SessionKey)
	}
}

func TestNormalizeAgentMessageImageAsResourceAttachment(t *testing.T) {
	event := agentResourceEvent("om_img", "oc_chat", "image", `{"image_key":"img_key","file_name":"photo.png"}`)
	msg, ok := NormalizeAgentMessage(event)
	if !ok {
		t.Fatal("expected agent message")
	}
	if len(msg.Images) != 1 || msg.Images[0].ID != "img_key" || msg.Images[0].FileName != "photo.png" {
		t.Fatalf("images = %#v", msg.Images)
	}
}

func TestNormalizeAgentMessageFileAsResourceAttachment(t *testing.T) {
	event := agentResourceEvent("om_file", "oc_chat", "file", `{"file_key":"file_key","file_name":"report.txt"}`)
	msg, ok := NormalizeAgentMessage(event)
	if !ok {
		t.Fatal("expected agent message")
	}
	if len(msg.Files) != 1 || msg.Files[0].ID != "file_key" || msg.Files[0].FileName != "report.txt" {
		t.Fatalf("files = %#v", msg.Files)
	}
}

func agentTextEvent(messageID string, chatID string, rootID string, threadID string, actorID string, text string) *larkim.P2MessageReceiveV1 {
	builder := larkim.NewEventMessageBuilder().
		MessageId(messageID).
		ChatId(chatID).
		MessageType("text").
		Content(`{"text":"` + text + `"}`)
	if rootID != "" {
		builder.RootId(rootID)
	}
	if threadID != "" {
		builder.ThreadId(threadID)
	}
	return &larkim.P2MessageReceiveV1{Event: &larkim.P2MessageReceiveV1Data{
		Sender:  &larkim.EventSender{SenderId: larkim.NewUserIdBuilder().OpenId(actorID).Build()},
		Message: builder.Build(),
	}}
}

func agentResourceEvent(messageID string, chatID string, messageType string, content string) *larkim.P2MessageReceiveV1 {
	return &larkim.P2MessageReceiveV1{Event: &larkim.P2MessageReceiveV1Data{
		Sender: &larkim.EventSender{SenderId: larkim.NewUserIdBuilder().OpenId("ou_test").Build()},
		Message: larkim.NewEventMessageBuilder().
			MessageId(messageID).
			ChatId(chatID).
			MessageType(messageType).
			Content(content).
			Build(),
	}}
}
