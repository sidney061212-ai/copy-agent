import test from 'node:test';
import assert from 'node:assert/strict';
import { executePlan } from '../../src/actions/executor.js';

test('executes copy and reply actions in order', async () => {
  const calls = [];
  await executePlan({
    actions: [
      { type: 'copy_text', text: 'hello' },
      { type: 'reply', text: 'ok', target: { messageId: 'm' } },
    ],
  }, {
    clipboard: { copyText: async (text) => calls.push(['copyText', text]) },
    reply: { send: async (target, text) => calls.push(['reply', target.messageId, text]) },
  });

  assert.deepEqual(calls, [['copyText', 'hello'], ['reply', 'm', 'ok']]);
});

test('reply failures are best-effort', async () => {
  const calls = [];
  const results = await executePlan({
    actions: [
      { type: 'copy_text', text: 'hello' },
      { type: 'reply', text: 'ok', target: { messageId: 'm' } },
    ],
  }, {
    clipboard: { copyText: async (text) => calls.push(['copyText', text]) },
    reply: { send: async () => { throw new Error('reply failed'); } },
  });

  assert.deepEqual(calls, [['copyText', 'hello']]);
  assert.equal(results[1].ok, false);
  assert.equal(results[1].bestEffort, true);
});

test('executes save resource and image copy actions', async () => {
  const calls = [];
  await executePlan({
    actions: [
      { type: 'save_resource', resource: { id: 'r', kind: 'image' }, directoryHint: '' },
      { type: 'copy_image', resource: { id: 'r', kind: 'image' } },
    ],
  }, {
    filesystem: { saveResource: async (resource) => { calls.push(['save', resource.id]); return '/tmp/i.png'; } },
    clipboard: { copyImage: async (resource) => calls.push(['copyImage', resource.id]) },
  });

  assert.deepEqual(calls, [['save', 'r'], ['copyImage', 'r']]);
});
