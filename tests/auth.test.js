import test from 'node:test';
import assert from 'node:assert/strict';
import { isAuthorized } from '../src/auth.js';

test('accepts bearer token', () => {
  const request = new Request('http://127.0.0.1/copy', {
    headers: { authorization: 'Bearer secret' },
  });

  assert.equal(isAuthorized(request, 'secret'), true);
});

test('accepts x-copyagent-token header', () => {
  const request = new Request('http://127.0.0.1/copy', {
    headers: { 'x-copyagent-token': 'secret' },
  });

  assert.equal(isAuthorized(request, 'secret'), true);
});

test('rejects missing or wrong tokens', () => {
  const missing = new Request('http://127.0.0.1/copy');
  const wrong = new Request('http://127.0.0.1/copy?token=nope');

  assert.equal(isAuthorized(missing, 'secret'), false);
  assert.equal(isAuthorized(wrong, 'secret'), false);
});
