export const REPLIES = Object.freeze({
  textCopied: '✅ 已复制到剪切板',
  imageCopied: '✅ 图片已复制到剪切板',
  fileSaved: '✅ 文件已保存',
  ignored: 'ℹ️ 已忽略：只处理文本、图片和文件',
});

export function failureReply(reason) {
  return `❌ 处理失败：${reason}`;
}

export function replyAction(target, text) {
  if (!target?.messageId) {
    return null;
  }

  return { type: 'reply', text, target };
}
