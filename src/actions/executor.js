async function executeAction(action, adapters) {
  if (action.type === 'copy_text') {
    await adapters.clipboard.copyText(action.text);
    return { ok: true, action: action.type };
  }

  if (action.type === 'copy_image') {
    await adapters.clipboard.copyImage(action.resource);
    return { ok: true, action: action.type };
  }

  if (action.type === 'save_resource') {
    const path = await adapters.filesystem.saveResource(action.resource, { directoryHint: action.directoryHint });
    return { ok: true, action: action.type, path };
  }

  if (action.type === 'reply') {
    try {
      await adapters.reply.send(action.target, action.text);
      return { ok: true, action: action.type, bestEffort: true };
    } catch (error) {
      return { ok: false, action: action.type, bestEffort: true, error: error.message };
    }
  }

  throw new Error(`unsupported action: ${action.type}`);
}

export async function executePlan(plan, adapters) {
  const results = [];

  for (const action of plan.actions ?? []) {
    results.push(await executeAction(action, adapters));
  }

  return results;
}
