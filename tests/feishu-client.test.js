import test from 'node:test';
import assert from 'node:assert/strict';
import { FeishuApiClient } from '../src/feishu-client.js';

test('client caches tenant access token', async () => {
  let tokenCalls = 0;
  const client = new FeishuApiClient({
    appId: 'app',
    appSecret: 'secret',
    fetch: async (url) => {
      if (String(url).includes('/tenant_access_token/internal')) {
        tokenCalls += 1;
        return Response.json({ code: 0, tenant_access_token: 'tenant-token', expire: 7200 });
      }
      return Response.json({ code: 0 });
    },
  });

  assert.equal(await client.getTenantAccessToken(), 'tenant-token');
  assert.equal(await client.getTenantAccessToken(), 'tenant-token');
  assert.equal(tokenCalls, 1);
});

test('reply sends text message to message id', async () => {
  const calls = [];
  const client = new FeishuApiClient({
    appId: 'app',
    appSecret: 'secret',
    fetch: async (url, init) => {
      calls.push({ url: String(url), init });
      if (String(url).includes('/tenant_access_token/internal')) {
        return Response.json({ code: 0, tenant_access_token: 'tenant-token', expire: 7200 });
      }
      return Response.json({ code: 0, data: {} });
    },
  });

  await client.replyText('msg-1', 'ok');

  assert.match(calls.at(-1).url, /reply/);
  assert.match(calls.at(-1).init.body, /ok/);
});
