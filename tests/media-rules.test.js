import test from 'node:test';
import assert from 'node:assert/strict';
import { normalizePathAlias, resolveSaveDirectory } from '../src/media-rules.js';

test('normalizes Chinese directory aliases', () => {
  assert.equal(normalizePathAlias('桌面'), 'Desktop');
  assert.equal(normalizePathAlias('下载'), 'Downloads');
});

test('uses default directory when no hint is provided', () => {
  assert.match(resolveSaveDirectory({ defaultDownloadDir: '~/Downloads/copyagent' }), /Downloads\/copyagent$/);
});

test('resolves save command directory hint', () => {
  assert.match(resolveSaveDirectory({ defaultDownloadDir: '~/Downloads/copyagent', hint: '保存到：桌面' }), /Desktop$/);
});
