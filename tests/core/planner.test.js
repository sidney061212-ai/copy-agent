import test from 'node:test';
import assert from 'node:assert/strict';
import { planEvent } from '../../src/core/planner.js';

test('plans text copy and reply', () => {
  const plan = planEvent({ type: 'text', text: '复制：你好', id: 'evt', actorId: 'u', replyTarget: { messageId: 'm' } });

  assert.deepEqual(plan.actions, [
    { type: 'copy_text', text: '你好' },
    { type: 'reply', text: '✅ 已复制到剪切板', target: { messageId: 'm' } },
  ]);
  assert.equal(plan.requiresApproval, false);
});

test('plans file save and reply', () => {
  const plan = planEvent({
    type: 'file',
    id: 'evt',
    replyTarget: { messageId: 'm' },
    resource: { id: 'file-key', name: 'a.pdf', kind: 'file' },
  });

  assert.deepEqual(plan.actions, [
    { type: 'save_resource', resource: { id: 'file-key', name: 'a.pdf', kind: 'file' }, directoryHint: '' },
    { type: 'reply', text: '✅ 文件已保存', target: { messageId: 'm' } },
  ]);
});

test('plans image save and clipboard copy by default', () => {
  const plan = planEvent({
    type: 'image',
    id: 'evt',
    replyTarget: { messageId: 'm' },
    resource: { id: 'image-key', name: 'i.png', kind: 'image' },
  });

  assert.equal(plan.actions[0].type, 'save_resource');
  assert.equal(plan.actions[1].type, 'copy_image');
  assert.deepEqual(plan.actions[2], { type: 'reply', text: '✅ 图片已复制到剪切板', target: { messageId: 'm' } });
});

test('returns ignored plan for unsupported event', () => {
  const plan = planEvent({ type: 'unknown', replyTarget: { messageId: 'm' } });

  assert.deepEqual(plan.actions, [
    { type: 'reply', text: 'ℹ️ 已忽略：只处理文本、图片和文件', target: { messageId: 'm' } },
  ]);
});

test('returns skipped plan for duplicate event', () => {
  const plan = planEvent({ type: 'text', text: 'hello', policy: { duplicate: true }, replyTarget: { messageId: 'm' } });

  assert.deepEqual(plan.actions, []);
});

test('returns forbidden reply for blocked actor', () => {
  const plan = planEvent({ type: 'text', text: 'hello', policy: { allowed: false, reason: 'actor is not allowed' }, replyTarget: { messageId: 'm' } });

  assert.deepEqual(plan.actions, [
    { type: 'reply', text: '❌ 处理失败：actor is not allowed', target: { messageId: 'm' } },
  ]);
});
