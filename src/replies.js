export function formatActionReply(result) {
  if (result.action === 'copied_text') {
    return '✅ 已复制到剪切板';
  }

  if (result.action === 'copied_image') {
    return '✅ 图片已复制到剪切板';
  }

  if (result.action === 'saved_file') {
    return `✅ 文件已保存到：${result.path}`;
  }

  if (result.action === 'ignored') {
    return 'ℹ️ 已忽略：只处理文本、图片和文件';
  }

  if (result.action === 'failed') {
    return `❌ 处理失败：${result.error}`;
  }

  return '✅ 已完成';
}
