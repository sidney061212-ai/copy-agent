import test from 'node:test';
import assert from 'node:assert/strict';
import { createDedupPolicy, isActorAllowed } from '../../src/core/policy.js';

test('allows all actors when allowlist is empty', () => {
  assert.equal(isActorAllowed('anyone', []), true);
});

test('blocks actor outside allowlist', () => {
  assert.equal(isActorAllowed('b', ['a']), false);
});

test('dedup policy detects repeated ids', () => {
  const policy = createDedupPolicy(2);

  assert.equal(policy.seen('1'), false);
  policy.mark('1');
  assert.equal(policy.seen('1'), true);
  policy.mark('2');
  policy.mark('3');
  assert.equal(policy.seen('1'), false);
});
