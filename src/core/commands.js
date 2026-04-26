const COMMAND_PATTERN = /^(?:@\S+\s+)?(?:复制|拷贝|copy|cp)\s*[:：]\s*/iu;

export function extractCopyText(text) {
  if (typeof text !== 'string') {
    return text;
  }

  return text.replace(COMMAND_PATTERN, '');
}
