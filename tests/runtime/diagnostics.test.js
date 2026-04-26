import test from 'node:test';
import assert from 'node:assert/strict';
import { summarizeDoctor, redactObject, parsePsRows } from '../../src/runtime/diagnostics.js';

test('redacts sensitive fields recursively', () => {
  const result = redactObject({ token: 'abc', nested: { feishuAppSecret: 'secret', ok: true } });

  assert.deepEqual(result, { token: '***REDACTED***', nested: { feishuAppSecret: '***REDACTED***', ok: true } });
});

test('summarizes doctor checks', () => {
  const summary = summarizeDoctor([
    { name: 'config', ok: true },
    { name: 'clipboard', ok: false, detail: 'missing' },
  ]);

  assert.equal(summary.ok, false);
  assert.equal(summary.failed.length, 1);
});

test('parses ps output rows', () => {
  const rows = parsePsRows('  PID  %CPU %MEM    RSS COMMAND\n  10   0.0  0.5  12345 node app\n');

  assert.equal(rows[0].pid, 10);
  assert.equal(rows[0].rssKb, 12345);
  assert.match(rows[0].command, /node app/);
});
