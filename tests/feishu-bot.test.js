import test from 'node:test';
import assert from 'node:assert/strict';
import { createFeishuEventHandler, extractLarkTextEvent } from '../src/feishu-bot.js';

test('extracts text from lark im.message.receive_v1 data', () => {
  const event = extractLarkTextEvent({
    header: { event_id: 'evt-1' },
    sender: { sender_id: { open_id: 'ou_1' } },
    message: {
      message_type: 'text',
      message_id: 'msg-1',
      content: JSON.stringify({ text: 'copy from bot' }),
    },
  });

  assert.deepEqual(event, {
    platform: 'feishu',
    type: 'copy_text',
    text: 'copy from bot',
    id: 'evt-1',
    actorId: 'ou_1',
    messageId: 'msg-1',
  });
});

test('ignores non-text lark messages', () => {
  assert.equal(extractLarkTextEvent({ message: { message_type: 'image', content: '{}' } }), null);
});

test('handler forwards normalized event to agent', async () => {
  const events = [];
  const handler = createFeishuEventHandler({ handleEvent: async (event) => events.push(event) });

  await handler({
    header: { event_id: 'evt-2' },
    sender: { sender_id: { open_id: 'ou_2' } },
    message: { message_type: 'text', content: JSON.stringify({ text: 'hello ws' }) },
  });

  assert.equal(events[0].text, 'hello ws');
});
