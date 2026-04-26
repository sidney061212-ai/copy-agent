import test from 'node:test';
import assert from 'node:assert/strict';
import { normalizeFeishuMessageEvent } from '../../../src/transports/feishu/normalize.js';

test('normalizes feishu text event', () => {
  const event = normalizeFeishuMessageEvent({
    header: { event_id: 'evt-1' },
    sender: { sender_id: { open_id: 'ou_1' } },
    message: { message_id: 'msg-1', message_type: 'text', content: JSON.stringify({ text: '复制：你好' }) },
  });

  assert.deepEqual(event, {
    transport: 'feishu',
    type: 'text',
    id: 'evt-1',
    actorId: 'ou_1',
    text: '复制：你好',
    replyTarget: { messageId: 'msg-1' },
  });
});

test('normalizes feishu file event', () => {
  const event = normalizeFeishuMessageEvent({
    header: { event_id: 'evt-2' },
    message: { message_id: 'msg-2', message_type: 'file', content: JSON.stringify({ file_key: 'key', file_name: 'a.txt' }) },
  });

  assert.equal(event.type, 'file');
  assert.deepEqual(event.resource, { id: 'key', name: 'a.txt', kind: 'file', messageId: 'msg-2' });
});

test('normalizes unsupported event as unknown', () => {
  const event = normalizeFeishuMessageEvent({ message: { message_id: 'm', message_type: 'audio', content: '{}' } });

  assert.equal(event.type, 'unknown');
  assert.deepEqual(event.replyTarget, { messageId: 'm' });
});
