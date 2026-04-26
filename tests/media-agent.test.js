import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtempSync, readFileSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { handleFeishuMediaEvent } from '../src/media-agent.js';

test('saves file event to default directory and replies', async () => {
  const dir = mkdtempSync(join(tmpdir(), 'copyagent-media-'));
  const replies = [];
  try {
    const result = await handleFeishuMediaEvent({
      data: {
        message: {
          message_id: 'msg-1',
          message_type: 'file',
          content: JSON.stringify({ file_key: 'file-key', file_name: 'a.txt' }),
        },
      },
      config: { defaultDownloadDir: dir, replyEnabled: true },
      apiClient: {
        downloadMessageResource: async () => Buffer.from('hello file'),
        replyText: async (messageId, text) => replies.push({ messageId, text }),
      },
    });

    assert.equal(result.action, 'saved_file');
    assert.equal(readFileSync(join(dir, 'a.txt'), 'utf8'), 'hello file');
    assert.match(replies[0].text, /已保存到/);
  } finally {
    rmSync(dir, { recursive: true, force: true });
  }
});

test('returns null for non-media message', async () => {
  const result = await handleFeishuMediaEvent({
    data: { message: { message_type: 'text', content: '{}' } },
    config: {},
    apiClient: {},
  });

  assert.equal(result, null);
});
