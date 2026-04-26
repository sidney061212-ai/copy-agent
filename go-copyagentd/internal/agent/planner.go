package agent

import (
	"errors"
	"strings"

	"github.com/copyagent/copyagentd/internal/core"
)

const DirectCopySuccessReplyText = "✅ 已复制到剪切板"
const DirectFileSavedReplyText = "✅ 文件已保存"
const DirectImageCopiedReplyText = "✅ 图片已复制到剪切板"

var ErrEmptyCopyText = errors.New("text is required")
var ErrResourceKeyRequired = errors.New("resource key is required")

func IsExplicitCopyCommand(content string) bool {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return false
	}
	extracted := core.ExtractCopyText(trimmed)
	return extracted != trimmed || !core.ValidText(extracted)
}

type DirectActionType string

const (
	DirectActionCopyText  DirectActionType = "copy_text"
	DirectActionSaveFile  DirectActionType = "save_file"
	DirectActionCopyImage DirectActionType = "copy_image"
	DirectActionReply     DirectActionType = "reply"
)

type DirectAction struct {
	Type        DirectActionType
	Text        string
	FileName    string
	MimeType    string
	Data        []byte
	ResourceRef *ResourceRef
	Path        string
	Reply       string
}

type DirectPlannerConfig struct {
	ImageAction string
}

type DirectPlanner struct {
	imageAction string
}

func NewDirectPlanner(cfg DirectPlannerConfig) *DirectPlanner {
	return &DirectPlanner{imageAction: normalizeImageAction(cfg.ImageAction)}
}

func (planner *DirectPlanner) Plan(msg *Message) ([]DirectAction, error) {
	if msg == nil {
		return nil, nil
	}
	var actions []DirectAction
	if strings.TrimSpace(msg.Content) != "" {
		text := core.ExtractCopyText(msg.Content)
		if !core.ValidText(text) {
			return nil, ErrEmptyCopyText
		}
		actions = append(actions,
			DirectAction{Type: DirectActionCopyText, Text: text},
			DirectAction{Type: DirectActionReply, Reply: DirectCopySuccessReplyText},
		)
	}
	for _, file := range msg.Files {
		resourceRef, err := attachmentResourceRef(msg, "file", file.ID, file.FileName, len(file.Data) == 0)
		if err != nil {
			return nil, err
		}
		actions = append(actions,
			DirectAction{Type: DirectActionSaveFile, FileName: file.FileName, MimeType: file.MimeType, Data: append([]byte(nil), file.Data...), ResourceRef: resourceRef},
			DirectAction{Type: DirectActionReply, Reply: DirectFileSavedReplyText},
		)
	}
	for _, image := range msg.Images {
		resourceRef, err := attachmentResourceRef(msg, "image", image.ID, image.FileName, len(image.Data) == 0)
		if err != nil {
			return nil, err
		}
		actions = append(actions, DirectAction{Type: DirectActionSaveFile, FileName: image.FileName, MimeType: image.MimeType, Data: append([]byte(nil), image.Data...), ResourceRef: resourceRef})
		if planner.imageAction != "save" {
			actions = append(actions,
				DirectAction{Type: DirectActionCopyImage},
				DirectAction{Type: DirectActionReply, Reply: DirectImageCopiedReplyText},
			)
		} else {
			actions = append(actions, DirectAction{Type: DirectActionReply, Reply: DirectFileSavedReplyText})
		}
	}
	return actions, nil
}

func attachmentResourceRef(msg *Message, kind string, key string, fileName string, required bool) (*ResourceRef, error) {
	if strings.TrimSpace(key) == "" {
		if required {
			return nil, ErrResourceKeyRequired
		}
		return nil, nil
	}
	return &ResourceRef{Platform: msg.Platform, MessageID: msg.MessageID, Key: key, Kind: kind, FileName: fileName, ReplyCtx: msg.ReplyCtx}, nil
}

func normalizeImageAction(action string) string {
	if strings.EqualFold(strings.TrimSpace(action), "save") {
		return "save"
	}
	return "clipboard"
}
