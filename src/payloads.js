import { HttpError } from './errors.js';

const DEFAULT_MAX_TEXT_BYTES = 200_000;

function byteLength(text) {
  return Buffer.byteLength(text, 'utf8');
}

export function validateText(value, options = {}) {
  const maxBytes = options.maxBytes ?? DEFAULT_MAX_TEXT_BYTES;

  if (typeof value !== 'string') {
    throw new HttpError(400, 'text is required');
  }

  if (value.trim().length === 0) {
    throw new HttpError(400, 'text is required');
  }

  if (byteLength(value) > maxBytes) {
    throw new HttpError(413, `text exceeds ${maxBytes} bytes`);
  }

  return value;
}

export function extractCopyText(payload, options = {}) {
  if (!payload || typeof payload !== 'object') {
    throw new HttpError(400, 'JSON object is required');
  }

  return validateText(payload.text, options);
}

function parseMessageContent(content) {
  if (typeof content !== 'string') {
    throw new HttpError(400, 'message content is required');
  }

  try {
    const parsed = JSON.parse(content);
    return parsed && typeof parsed === 'object' ? parsed : {};
  } catch {
    throw new HttpError(400, 'message content must be JSON');
  }
}

function getEventId(payload) {
  return payload?.header?.event_id
    ?? payload?.event_id
    ?? payload?.uuid
    ?? payload?.event?.message?.message_id
    ?? '';
}

export function extractFeishuEvent(payload, options = {}) {
  if (!payload || typeof payload !== 'object') {
    throw new HttpError(400, 'JSON object is required');
  }

  if (payload.type === 'url_verification' && typeof payload.challenge === 'string') {
    return { kind: 'challenge', challenge: payload.challenge };
  }

  const message = payload.event?.message ?? payload.event;
  if (!message || typeof message !== 'object') {
    throw new HttpError(400, 'message event is required');
  }

  if (message.message_type !== 'text' && message.msg_type !== 'text') {
    throw new HttpError(400, 'only text messages are supported');
  }

  const content = parseMessageContent(message.content ?? message.text?.content ?? '{}');
  const text = validateText(content.text ?? content.content, options);

  return { kind: 'text', text, eventId: getEventId(payload) };
}
