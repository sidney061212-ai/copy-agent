import test from 'node:test';
import assert from 'node:assert/strict';
import { extractCopyText, extractFeishuEvent } from '../src/payloads.js';

test('extracts generic text payload', () => {
  assert.equal(extractCopyText({ text: 'hello' }), 'hello');
});

test('rejects empty generic payload', () => {
  assert.throws(() => extractCopyText({ text: '   ' }), /text is required/);
});

test('extracts feishu challenge', () => {
  const event = extractFeishuEvent({ type: 'url_verification', challenge: 'abc' });

  assert.deepEqual(event, { kind: 'challenge', challenge: 'abc' });
});

test('extracts feishu text message v2 payload', () => {
  const event = extractFeishuEvent({
    header: { event_id: 'evt-1' },
    event: {
      message: {
        message_type: 'text',
        content: JSON.stringify({ text: 'copy me' }),
      },
    },
  });

  assert.deepEqual(event, { kind: 'text', text: 'copy me', eventId: 'evt-1' });
});

test('rejects unsupported feishu message type', () => {
  assert.throws(() => extractFeishuEvent({
    event: { message: { message_type: 'image', content: '{}' } },
  }), /only text messages/);
});
