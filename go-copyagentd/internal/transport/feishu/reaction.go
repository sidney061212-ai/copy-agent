package feishu

import (
	"context"
	"log"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

const defaultTypingEmoji = "OnIt"

func (transport *AgentTransport) StartTyping(ctx context.Context, replyCtx any) func() {
	rctx, ok := replyCtx.(replyContext)
	if !ok || rctx.messageID == "" {
		return func() {}
	}
	reactionID := transport.addReaction(ctx, rctx.messageID, defaultTypingEmoji)
	return func() {
		go transport.removeReaction(context.Background(), rctx.messageID, reactionID)
	}
}

func (transport *AgentTransport) addReaction(ctx context.Context, messageID string, emojiType string) string {
	if transport.apiClient == nil || messageID == "" || emojiType == "" {
		return ""
	}
	resp, err := transport.apiClient.Im.MessageReaction.Create(ctx,
		larkim.NewCreateMessageReactionReqBuilder().
			MessageId(messageID).
			Body(larkim.NewCreateMessageReactionReqBodyBuilder().
				ReactionType(&larkim.Emoji{EmojiType: &emojiType}).
				Build()).
			Build())
	if err != nil {
		log.Printf("feishu typing reaction failed: message=%s err=%v", messageID, err)
		return ""
	}
	if !resp.Success() {
		log.Printf("feishu typing reaction failed: message=%s code=%d msg=%s", messageID, resp.Code, resp.Msg)
		return ""
	}
	if resp.Data != nil && resp.Data.ReactionId != nil {
		return *resp.Data.ReactionId
	}
	return ""
}

func (transport *AgentTransport) removeReaction(ctx context.Context, messageID string, reactionID string) {
	if transport.apiClient == nil || messageID == "" || reactionID == "" {
		return
	}
	resp, err := transport.apiClient.Im.MessageReaction.Delete(ctx,
		larkim.NewDeleteMessageReactionReqBuilder().
			MessageId(messageID).
			ReactionId(reactionID).
			Build())
	if err != nil {
		log.Printf("feishu typing reaction remove failed: message=%s err=%v", messageID, err)
		return
	}
	if !resp.Success() {
		log.Printf("feishu typing reaction remove failed: message=%s code=%d msg=%s", messageID, resp.Code, resp.Msg)
	}
}
