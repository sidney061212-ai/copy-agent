package feishu

import (
	"context"
	"errors"
	"io"

	"github.com/copyagent/copyagentd/internal/agent"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type feishuResourceDownloader struct {
	client *lark.Client
}

func (downloader feishuResourceDownloader) DownloadResource(ctx context.Context, ref agent.ResourceRef) ([]byte, error) {
	if downloader.client == nil {
		return nil, errors.New("feishu client is required")
	}
	resp, err := downloader.client.Im.MessageResource.Get(ctx, larkim.NewGetMessageResourceReqBuilder().
		MessageId(ref.MessageID).
		FileKey(ref.Key).
		Type(ref.Kind).
		Build())
	if err != nil {
		return nil, err
	}
	if !resp.Success() {
		return nil, resp.CodeError
	}
	return io.ReadAll(resp.File)
}
