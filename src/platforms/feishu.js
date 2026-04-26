import { createDecipheriv, createHash, timingSafeEqual } from 'node:crypto';
import { HttpError } from '../errors.js';
import { extractFeishuEvent } from '../payloads.js';

function safeEqual(left, right) {
  const leftBuffer = Buffer.from(left);
  const rightBuffer = Buffer.from(right);

  if (leftBuffer.length !== rightBuffer.length) {
    return false;
  }

  return timingSafeEqual(leftBuffer, rightBuffer);
}

export function verifyFeishuSignature({ headers, rawBody, encryptKey }) {
  if (!encryptKey) {
    return true;
  }

  const timestamp = headers.get('x-lark-request-timestamp') ?? '';
  const nonce = headers.get('x-lark-request-nonce') ?? '';
  const signature = headers.get('x-lark-signature') ?? '';

  if (!timestamp || !nonce || !signature) {
    return false;
  }

  const computed = createHash('sha256')
    .update(timestamp + nonce + encryptKey)
    .update(rawBody)
    .digest('hex');

  return safeEqual(computed, signature);
}

export function decryptFeishuPayload(encryptedPayload, encryptKey) {
  if (!encryptKey) {
    throw new HttpError(500, 'feishu encrypt key is required');
  }

  try {
    const encrypted = Buffer.from(encryptedPayload, 'base64');
    const iv = encrypted.subarray(0, 16);
    const ciphertext = encrypted.subarray(16);
    const key = createHash('sha256').update(encryptKey).digest();
    const decipher = createDecipheriv('aes-256-cbc', key, iv);
    const decrypted = Buffer.concat([decipher.update(ciphertext), decipher.final()]).toString('utf8');
    return JSON.parse(decrypted);
  } catch {
    throw new HttpError(400, 'invalid encrypted feishu payload');
  }
}

function assertVerificationToken(payload, verificationToken) {
  if (!verificationToken) {
    return;
  }

  const actual = payload?.token ?? payload?.header?.token ?? '';
  if (!safeEqual(String(actual), verificationToken)) {
    throw new HttpError(401, 'invalid feishu verification token');
  }
}

export async function parseFeishuRequest(request, options = {}) {
  const rawBody = Buffer.from(await request.arrayBuffer());
  let payload;

  try {
    payload = JSON.parse(rawBody.toString('utf8'));
  } catch {
    throw new HttpError(400, 'invalid JSON');
  }

  if (typeof payload.encrypt === 'string') {
    payload = decryptFeishuPayload(payload.encrypt, options.encryptKey);
  }

  const parsedEvent = extractFeishuEvent(payload, { maxBytes: options.maxTextBytes });
  if (parsedEvent.kind === 'challenge') {
    return { kind: 'challenge', challenge: parsedEvent.challenge };
  }

  if (!verifyFeishuSignature({ headers: request.headers, rawBody, encryptKey: options.encryptKey })) {
    throw new HttpError(401, 'invalid feishu signature');
  }

  assertVerificationToken(payload, options.verificationToken);

  return {
    kind: 'event',
    event: {
      platform: 'feishu',
      type: 'copy_text',
      text: parsedEvent.text,
      id: parsedEvent.eventId,
      actorId: payload?.event?.sender?.sender_id?.open_id
        ?? payload?.event?.sender?.sender_id?.user_id
        ?? payload?.event?.operator_id?.open_id
        ?? '',
    },
  };
}
