export function isActorAllowed(actorId, allowedActorIds = []) {
  if (!allowedActorIds || allowedActorIds.length === 0) {
    return true;
  }

  return allowedActorIds.includes(actorId ?? '');
}

export function createDedupPolicy(limit = 1_000) {
  const seenIds = new Set();
  const order = [];

  return {
    seen(id) {
      return Boolean(id) && seenIds.has(id);
    },
    mark(id) {
      if (!id || seenIds.has(id)) {
        return;
      }

      seenIds.add(id);
      order.push(id);
      while (order.length > limit) {
        const oldest = order.shift();
        seenIds.delete(oldest);
      }
    },
  };
}

export function applyPolicy(event, options = {}) {
  if (!isActorAllowed(event.actorId, options.allowedActorIds ?? [])) {
    return { ok: false, reason: 'actor is not allowed' };
  }

  const dedup = options.dedupPolicy;
  if (dedup?.seen(event.id)) {
    return { ok: false, reason: 'duplicate event', duplicate: true };
  }

  dedup?.mark(event.id);
  return { ok: true };
}
