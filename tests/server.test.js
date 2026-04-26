import test from 'node:test';
import assert from 'node:assert/strict';
import { createApp } from '../src/server.js';

test('health endpoint returns ok', async () => {
  const app = createApp({ token: 'secret', writeClipboard: async () => {} });
  const response = await app.fetch(new Request('http://127.0.0.1/health'));

  assert.equal(response.status, 200);
  assert.equal((await response.json()).ok, true);
});

test('copy endpoint writes authorized text', async () => {
  const writes = [];
  const app = createApp({ token: 'secret', writeClipboard: async (text) => writes.push(text) });
  const response = await app.fetch(new Request('http://127.0.0.1/copy', {
    method: 'POST',
    headers: { authorization: 'Bearer secret', 'content-type': 'application/json' },
    body: JSON.stringify({ text: 'hello clipboard' }),
  }));

  assert.equal(response.status, 200);
  assert.deepEqual(writes, ['hello clipboard']);
});

test('copy endpoint rejects unauthorized requests without writing', async () => {
  const writes = [];
  const app = createApp({ token: 'secret', writeClipboard: async (text) => writes.push(text) });
  const response = await app.fetch(new Request('http://127.0.0.1/copy', {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ text: 'nope' }),
  }));

  assert.equal(response.status, 401);
  assert.deepEqual(writes, []);
});

test('feishu duplicate event writes once', async () => {
  const writes = [];
  const app = createApp({ token: 'secret', writeClipboard: async (text) => writes.push(text) });
  const requestBody = JSON.stringify({
    header: { event_id: 'same-event' },
    event: { message: { message_type: 'text', content: JSON.stringify({ text: 'once' }) } },
  });

  for (let index = 0; index < 2; index += 1) {
    const response = await app.fetch(new Request('http://127.0.0.1/feishu', {
      method: 'POST',
      headers: { 'x-copyagent-token': 'secret', 'content-type': 'application/json' },
      body: requestBody,
    }));
    assert.equal(response.status, 200);
  }

  assert.deepEqual(writes, ['once']);
});
