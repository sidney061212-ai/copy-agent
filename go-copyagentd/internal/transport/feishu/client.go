package feishu

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	coreevent "github.com/copyagent/copyagentd/internal/event"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
)

type Transport struct {
	client    *larkws.Client
	apiClient *lark.Client
}

func NewTransport(appID, appSecret string) *Transport {
	eventHandler := dispatcher.NewEventDispatcher("", "")
	return &Transport{
		client: larkws.NewClient(
			appID,
			appSecret,
			larkws.WithEventHandler(eventHandler),
			larkws.WithLogLevel(larkcore.LogLevelError),
			larkws.WithDomain(lark.FeishuBaseUrl),
		),
	}
}

func NewTransportWithHandler(appID, appSecret string, cfg MessageHandlerConfig, apiClient *lark.Client) *Transport {
	if apiClient == nil {
		apiClient = lark.NewClient(appID, appSecret, lark.WithLogLevel(larkcore.LogLevelError))
	}
	if cfg.Reply == nil {
		cfg.Reply = feishuReply{client: apiClient}
	}
	if cfg.Downloader == nil {
		cfg.Downloader = feishuDownloader{client: apiClient}
	}
	eventHandler := dispatcher.NewEventDispatcher("", "").OnP2MessageReceiveV1(NewMessageHandler(cfg).Handle)
	return &Transport{
		apiClient: apiClient,
		client: larkws.NewClient(
			appID,
			appSecret,
			larkws.WithEventHandler(eventHandler),
			larkws.WithLogLevel(larkcore.LogLevelError),
			larkws.WithDomain(lark.FeishuBaseUrl),
		),
	}
}

type feishuDownloader struct {
	client *lark.Client
}

func (d feishuDownloader) Download(ctx context.Context, resource coreevent.ResourceMessage) ([]byte, error) {
	if d.client == nil {
		return nil, errors.New("feishu client is required")
	}
	resp, err := d.client.Im.MessageResource.Get(ctx, larkim.NewGetMessageResourceReqBuilder().
		MessageId(resource.MessageID).
		FileKey(resource.Key).
		Type(resource.Kind).
		Build())
	if err != nil {
		return nil, err
	}
	if !resp.Success() {
		return nil, resp.CodeError
	}
	return io.ReadAll(resp.File)
}

func (t *Transport) Start(ctx context.Context) error {
	return t.client.Start(ctx)
}

type feishuReply struct {
	client *lark.Client
}

func (r feishuReply) ReplyText(ctx context.Context, messageID string, text string) error {
	if r.client == nil {
		return errors.New("feishu client is required")
	}
	content, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return err
	}
	resp, err := r.client.Im.Message.Reply(ctx, larkim.NewReplyMessageReqBuilder().
		MessageId(messageID).
		Body(larkim.NewReplyMessageReqBodyBuilder().
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
