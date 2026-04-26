import test from 'node:test';
import assert from 'node:assert/strict';
import { createCipheriv, createHash } from 'node:crypto';
import { decryptFeishuPayload, parseFeishuRequest, verifyFeishuSignature } from '../../src/platforms/feishu.js';

function sign({ timestamp, nonce, encryptKey, rawBody }) {
  return createHash('sha256').update(timestamp + nonce + encryptKey).update(rawBody).digest('hex');
}

function encryptPayload(payload, encryptKey) {
  const key = createHash('sha256').update(encryptKey).digest();
  const iv = Buffer.from('1234567890abcdef');
  const cipher = createCipheriv('aes-256-cbc', key, iv);
  return Buffer.concat([iv, cipher.update(JSON.stringify(payload), 'utf8'), cipher.final()]).toString('base64');
}

test('verifies feishu raw body signature', () => {
  const rawBody = Buffer.from('{"hello":"world"}');
  const headers = new Headers({
    'x-lark-request-timestamp': '100',
    'x-lark-request-nonce': 'nonce',
    'x-lark-signature': sign({ timestamp: '100', nonce: 'nonce', encryptKey: 'key', rawBody }),
  });

  assert.equal(verifyFeishuSignature({ headers, rawBody, encryptKey: 'key' }), true);
});

test('rejects invalid feishu signature', () => {
  const rawBody = Buffer.from('{"hello":"world"}');
  const headers = new Headers({
    'x-lark-request-timestamp': '100',
    'x-lark-request-nonce': 'nonce',
    'x-lark-signature': 'bad',
  });

  assert.equal(verifyFeishuSignature({ headers, rawBody, encryptKey: 'key' }), false);
});

test('parses feishu text into normalized agent event', async () => {
  const body = {
    token: 'verify-token',
    header: { event_id: 'evt-1' },
    event: { message: { message_type: 'text', content: JSON.stringify({ text: 'copy' }) } },
  };
  const rawBody = Buffer.from(JSON.stringify(body));
  const request = new Request('http://127.0.0.1/feishu', {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: rawBody,
  });

  const parsed = await parseFeishuRequest(request, { verificationToken: 'verify-token' });

  assert.deepEqual(parsed, {
    kind: 'event',
    event: { platform: 'feishu', type: 'copy_text', text: 'copy', id: 'evt-1', actorId: '' },
  });
});

test('decrypts feishu encrypted payload', () => {
  const encrypted = encryptPayload({ type: 'url_verification', challenge: 'abc' }, 'encrypt-key');

  assert.deepEqual(decryptFeishuPayload(encrypted, 'encrypt-key'), { type: 'url_verification', challenge: 'abc' });
});

test('parses encrypted feishu challenge without signature headers', async () => {
  const encrypted = encryptPayload({ type: 'url_verification', challenge: 'encrypted-abc' }, 'encrypt-key');
  const request = new Request('http://127.0.0.1/feishu', {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ encrypt: encrypted }),
  });

  const parsed = await parseFeishuRequest(request, { encryptKey: 'encrypt-key' });

  assert.deepEqual(parsed, { kind: 'challenge', challenge: 'encrypted-abc' });
});
