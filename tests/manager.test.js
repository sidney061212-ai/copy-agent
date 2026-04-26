import test from 'node:test';
import assert from 'node:assert/strict';
import { buildPublicUrl, mergeRuntimeConfig } from '../src/manager.js';

test('merges file config into runtime env shape', () => {
  const config = mergeRuntimeConfig({
    host: '127.0.0.1',
    port: 8765,
    token: 'token',
    feishuVerificationToken: 'verify',
    feishuEncryptKey: 'encrypt',
    allowedActorIds: ['a', 'b'],
  });

  assert.equal(config.COPYAGENT_PORT, '8765');
  assert.equal(config.COPYAGENT_ALLOWED_ACTOR_IDS, 'a,b');
});

test('builds copy and feishu URLs', () => {
  const urls = buildPublicUrl({ baseUrl: 'https://example.com/', token: 'token' });

  assert.equal(urls.copy, 'https://example.com/copy?token=token');
  assert.equal(urls.feishu, 'https://example.com/feishu?token=token');
});
