import test from 'node:test';
import assert from 'node:assert/strict';
import { formatActionReply } from '../src/replies.js';

test('formats text copied reply', () => {
  assert.equal(formatActionReply({ action: 'copied_text' }), '✅ 已复制到剪切板');
});

test('formats file saved reply', () => {
  assert.equal(formatActionReply({ action: 'saved_file', path: '/tmp/a.pdf' }), '✅ 文件已保存到：/tmp/a.pdf');
});

test('formats failure reply', () => {
  assert.equal(formatActionReply({ action: 'failed', error: 'boom' }), '❌ 处理失败：boom');
});
