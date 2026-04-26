import { extractCopyText } from '../payloads.js';

async function readJson(request) {
  return JSON.parse(await request.text());
}

export async function parseGenericRequest(request, options = {}) {
  const payload = await readJson(request);
  const text = extractCopyText(payload, { maxBytes: options.maxTextBytes });
  const id = typeof payload.id === 'string' ? payload.id : '';

  return {
    kind: 'event',
    event: { platform: 'generic', type: 'copy_text', text, id },
  };
}
