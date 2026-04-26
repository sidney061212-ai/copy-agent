package feishu

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/copyagent/copyagentd/internal/agent"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
)

type AgentTransport struct {
	appID     string
	appSecret string
	client    *larkws.Client
	apiClient *lark.Client
}

func NewAgentTransport(appID string, appSecret string) *AgentTransport {
	return NewAgentTransportWithClient(appID, appSecret, nil)
}

func NewAgentTransportWithClient(appID string, appSecret string, apiClient *lark.Client) *AgentTransport {
	if apiClient == nil {
		apiClient = lark.NewClient(appID, appSecret, lark.WithLogLevel(larkcore.LogLevelError))
	}
	return &AgentTransport{appID: appID, appSecret: appSecret, apiClient: apiClient}
}

func (transport *AgentTransport) Name() string { return "feishu" }

func (transport *AgentTransport) Start(handler agent.MessageHandler) error {
	eventHandler := dispatcher.NewEventDispatcher("", "").
		OnP2MessageReceiveV1(func(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
			msg, ok := NormalizeAgentMessage(event)
			if !ok {
				return nil
			}
			if handler != nil {
				handler(transport, &msg)
			}
			return nil
		}).
		OnP2MessageReactionCreatedV1(func(context.Context, *larkim.P2MessageReactionCreatedV1) error {
			return nil
		}).
		OnP2MessageReactionDeletedV1(func(context.Context, *larkim.P2MessageReactionDeletedV1) error {
			return nil
		})
	transport.client = larkws.NewClient(
		transport.appID,
		transport.appSecret,
		larkws.WithEventHandler(eventHandler),
		larkws.WithLogLevel(larkcore.LogLevelError),
		larkws.WithDomain(lark.FeishuBaseUrl),
	)
	return transport.client.Start(context.Background())
}

func (transport *AgentTransport) Stop() error { return nil }

func (transport *AgentTransport) Reply(ctx context.Context, replyCtx any, content string) error {
	rctx, ok := replyCtx.(replyContext)
	if !ok {
		return fmt.Errorf("feishu reply requires reply context, got %T", replyCtx)
	}
	if strings.TrimSpace(rctx.messageID) == "" {
		return transport.sendText(ctx, rctx.chatID, content)
	}
	return transport.replyText(ctx, rctx, content)
}

func (transport *AgentTransport) Send(ctx context.Context, replyCtx any, content string) error {
	rctx, ok := replyCtx.(replyContext)
	if !ok {
		return fmt.Errorf("feishu send requires reply context, got %T", replyCtx)
	}
	if strings.TrimSpace(rctx.messageID) != "" && isThreadSessionKey(rctx.sessionKey) {
		return transport.replyText(ctx, rctx, content)
	}
	return transport.sendText(ctx, rctx.chatID, content)
}

func (transport *AgentTransport) Download(ctx context.Context, ref agent.ResourceRef) ([]byte, error) {
	return feishuResourceDownloader{client: transport.apiClient}.DownloadResource(ctx, ref)
}

func (transport *AgentTransport) replyText(ctx context.Context, rctx replyContext, text string) error {
	if transport.apiClient == nil {
		return fmt.Errorf("feishu client is required")
	}
	content, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return err
	}
	body := larkim.NewReplyMessageReqBodyBuilder().MsgType("text").Content(string(content))
	if isThreadSessionKey(rctx.sessionKey) {
		body.ReplyInThread(true)
	}
	resp, err := transport.apiClient.Im.Message.Reply(ctx, larkim.NewReplyMessageReqBuilder().
		MessageId(rctx.messageID).
		Body(body.Build()).
		Build())
	if err != nil {
		return err
	}
	if !resp.Success() {
		return resp.CodeError
	}
	return nil
}

func (transport *AgentTransport) sendText(ctx context.Context, chatID string, text string) error {
	if transport.apiClient == nil {
		return fmt.Errorf("feishu client is required")
	}
	if strings.TrimSpace(chatID) == "" {
		return fmt.Errorf("feishu chat id is required")
	}
	content, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return err
	}
	resp, err := transport.apiClient.Im.Message.Create(ctx, larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(larkim.ReceiveIdTypeChatId).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			ReceiveId(chatID).
			MsgType("text").
			Content(string(content)).
			Build()).
		Build())
	if err != nil {
		return err
	}
	if !resp.Success() {
		return resp.CodeError
	}
	return nil
}
