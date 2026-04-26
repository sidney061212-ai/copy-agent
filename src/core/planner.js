import { extractCopyText } from './commands.js';
import { failureReply, REPLIES, replyAction } from './replies.js';

function compactActions(actions) {
  return actions.filter(Boolean);
}

function planText(event) {
  return compactActions([
    { type: 'copy_text', text: extractCopyText(event.text) },
    replyAction(event.replyTarget, REPLIES.textCopied),
  ]);
}

function planFile(event) {
  return compactActions([
    { type: 'save_resource', resource: event.resource, directoryHint: event.directoryHint ?? '' },
    replyAction(event.replyTarget, REPLIES.fileSaved),
  ]);
}

function planImage(event, options) {
  const imageAction = options.imageAction ?? 'clipboard';
  const actions = [
    { type: 'save_resource', resource: event.resource, directoryHint: event.directoryHint ?? '' },
  ];

  if (imageAction !== 'save') {
    actions.push({ type: 'copy_image', resource: event.resource });
    actions.push(replyAction(event.replyTarget, REPLIES.imageCopied));
  } else {
    actions.push(replyAction(event.replyTarget, REPLIES.fileSaved));
  }

  return compactActions(actions);
}

export function planEvent(event, options = {}) {
  if (event.policy?.duplicate) {
    return { actions: [], requiresApproval: false };
  }

  if (event.policy?.allowed === false) {
    return {
      actions: compactActions([replyAction(event.replyTarget, failureReply(event.policy.reason ?? 'not allowed'))]),
      requiresApproval: false,
    };
  }

  if (event.type === 'text') {
    return { actions: planText(event), requiresApproval: false };
  }

  if (event.type === 'file') {
    return { actions: planFile(event), requiresApproval: false };
  }

  if (event.type === 'image') {
    return { actions: planImage(event, options), requiresApproval: false };
  }

  return {
    actions: compactActions([replyAction(event.replyTarget, REPLIES.ignored)]),
    requiresApproval: false,
  };
}
