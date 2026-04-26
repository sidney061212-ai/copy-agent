import test from 'node:test';
import assert from 'node:assert/strict';
import { extractCopyCommandText } from '../src/commands.js';

test('copies raw text by default', () => {
  assert.equal(extractCopyCommandText('hello'), 'hello');
});

test('extracts Chinese copy command after colon', () => {
  assert.equal(extractCopyCommandText('复制：你好世界'), '你好世界');
});

test('extracts English copy command after colon', () => {
  assert.equal(extractCopyCommandText('copy: hello'), 'hello');
});

test('ignores bot mention prefix before command', () => {
  assert.equal(extractCopyCommandText('@copyagent 复制：中文'), '中文');
});
