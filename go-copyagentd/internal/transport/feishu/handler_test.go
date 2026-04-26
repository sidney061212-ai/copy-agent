package feishu

import (
	"context"
	"errors"
	"testing"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func TestNormalizeTextMessage(t *testing.T) {
	content := `{"text":"复制：你好"}`
	event := &larkim.P2MessageReceiveV1{Event: &larkim.P2MessageReceiveV1Data{
		Sender: &larkim.EventSender{SenderId: larkim.NewUserIdBuilder().OpenId("ou_test").Build()},
		Message: larkim.NewEventMessageBuilder().
			MessageId("om_test").
			MessageType("text").
			Content(content).
			Build(),
	}}

	normalized, ok := NormalizeTextMessage(event)
	if !ok {
		t.Fatal("expected text event")
	}
	if normalized.Text != "复制：你好" {
		t.Fatalf("text = %q", normalized.Text)
	}
	if normalized.MessageID != "om_test" {
		t.Fatalf("message id = %q", normalized.MessageID)
	}
	if normalized.ActorID != "ou_test" {
		t.Fatalf("actor id = %q", normalized.ActorID)
	}
}

func TestNormalizeTextMessageRejectsNonText(t *testing.T) {
	event := &larkim.P2MessageReceiveV1{Event: &larkim.P2MessageReceiveV1Data{
		Message: larkim.NewEventMessageBuilder().
			MessageId("om_test").
			MessageType("image").
			Content(`{"image_key":"img"}`).
			Build(),
	}}

	if _, ok := NormalizeTextMessage(event); ok {
		t.Fatal("expected non-text event to be rejected")
	}
}

func TestMessageHandlerCopiesAndReplies(t *testing.T) {
	var copied string
	var replyTarget string
	var replyText string
	handler := NewMessageHandler(MessageHandlerConfig{
		ReplyEnabled: true,
		Clipboard: ClipboardFunc(func(_ context.Context, text string) error {
			copied = text
			return nil
		}),
		Reply: ReplyFunc(func(_ context.Context, messageID string, text string) error {
			replyTarget = messageID
			replyText = text
			return nil
		}),
	})

	event := textEvent("om_test", "复制：你好")
	if err := handler.Handle(context.Background(), event); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if copied != "你好" {
		t.Fatalf("copied = %q", copied)
	}
	if replyTarget != "om_test" {
		t.Fatalf("reply target = %q", replyTarget)
	}
	if replyText != SuccessReplyText {
		t.Fatalf("reply text = %q", replyText)
	}
}

func TestMessageHandlerTreatsReplyFailureAsBestEffort(t *testing.T) {
	copied := false
	handler := NewMessageHandler(MessageHandlerConfig{
		ReplyEnabled: true,
		Clipboard: ClipboardFunc(func(_ context.Context, text string) error {
			copied = true
			return nil
		}),
		Reply: ReplyFunc(func(_ context.Context, messageID string, text string) error {
			return errors.New("reply failed")
		}),
	})

	if err := handler.Handle(context.Background(), textEvent("om_test", "copy: hi")); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if !copied {
		t.Fatal("expected clipboard write")
	}
}

func TestMessageHandlerRejectsBlankText(t *testing.T) {
	called := false
	handler := NewMessageHandler(MessageHandlerConfig{
		Clipboard: ClipboardFunc(func(_ context.Context, text string) error {
			called = true
			return nil
		}),
	})

	if err := handler.Handle(context.Background(), textEvent("om_test", "复制：   ")); err == nil {
		t.Fatal("expected invalid text error")
	}
	if called {
		t.Fatal("blank text should not be copied")
	}
}

func TestMessageHandlerSkipsDisallowedActor(t *testing.T) {
	called := false
	handler := NewMessageHandler(MessageHandlerConfig{
		AllowedActorIDs: []string{"ou_allowed"},
		Clipboard: ClipboardFunc(func(_ context.Context, text string) error {
			called = true
			return nil
		}),
	})

	if err := handler.Handle(context.Background(), textEventWithActor("om_test", "ou_other", "复制：你好")); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if called {
		t.Fatal("disallowed actor should not be copied")
	}
}

func textEvent(messageID string, text string) *larkim.P2MessageReceiveV1 {
	return textEventWithActor(messageID, "ou_test", text)
}

func textEventWithActor(messageID string, actorID string, text string) *larkim.P2MessageReceiveV1 {
	return &larkim.P2MessageReceiveV1{Event: &larkim.P2MessageReceiveV1Data{
		Sender: &larkim.EventSender{SenderId: larkim.NewUserIdBuilder().OpenId(actorID).Build()},
		Message: larkim.NewEventMessageBuilder().
			MessageId(messageID).
			MessageType("text").
			Content(`{"text":"` + text + `"}`).
			Build(),
	}}
}
