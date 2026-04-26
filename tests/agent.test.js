import test from 'node:test';
import assert from 'node:assert/strict';
import { CopyAgent } from '../src/agent.js';

test('agent copies normalized text event', async () => {
  const writes = [];
  const agent = new CopyAgent({ writeClipboard: async (text) => writes.push(text) });

  const result = await agent.handleEvent({ platform: 'generic', type: 'copy_text', text: 'agent text', id: '1' });

  assert.equal(result.ok, true);
  assert.equal(result.action, 'copied');
  assert.deepEqual(writes, ['agent text']);
});

test('agent deduplicates event ids', async () => {
  const writes = [];
  const agent = new CopyAgent({ writeClipboard: async (text) => writes.push(text) });
  const event = { platform: 'feishu', type: 'copy_text', text: 'once', id: 'evt' };

  await agent.handleEvent(event);
  const duplicate = await agent.handleEvent(event);

  assert.equal(duplicate.duplicate, true);
  assert.deepEqual(writes, ['once']);
});

test('agent rejects unsupported event types', async () => {
  const agent = new CopyAgent({ writeClipboard: async () => {} });

  await assert.rejects(
    () => agent.handleEvent({ platform: 'generic', type: 'noop' }),
    /unsupported agent event/,
  );
});

test('agent enforces optional actor allowlist', async () => {
  const agent = new CopyAgent({ writeClipboard: async () => {}, allowedActorIds: ['ou_allowed'] });

  await assert.rejects(
    () => agent.handleEvent({ platform: 'feishu', type: 'copy_text', text: 'blocked', actorId: 'ou_blocked' }),
    /actor is not allowed/,
  );
});
