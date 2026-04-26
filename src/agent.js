import { writeClipboard as defaultWriteClipboard } from './clipboard.js';
import { HttpError } from './errors.js';
import { validateText } from './payloads.js';

function createEventCache(limit = 1_000) {
  const seen = new Set();
  const order = [];

  return {
    has(id) {
      return Boolean(id) && seen.has(id);
    },
    add(id) {
      if (!id || seen.has(id)) {
        return;
      }

      seen.add(id);
      order.push(id);
      while (order.length > limit) {
        const oldest = order.shift();
        seen.delete(oldest);
      }
    },
  };
}

export class CopyAgent {
  constructor(options = {}) {
    this.writeClipboard = options.writeClipboard ?? defaultWriteClipboard;
    this.maxTextBytes = options.maxTextBytes ?? 200_000;
    this.logText = options.logText ?? false;
    this.allowedActorIds = new Set(options.allowedActorIds ?? []);
    this.seenEvents = createEventCache(options.dedupCacheSize ?? 1_000);
  }

  async handleEvent(event) {
    if (!event || event.type !== 'copy_text') {
      throw new HttpError(400, 'unsupported agent event');
    }

    if (this.allowedActorIds.size > 0 && !this.allowedActorIds.has(event.actorId ?? '')) {
      throw new HttpError(403, 'actor is not allowed');
    }

    if (this.seenEvents.has(event.id)) {
      return { ok: true, action: 'skipped', duplicate: true };
    }

    const text = validateText(event.text, { maxBytes: this.maxTextBytes });
    await this.writeClipboard(text);
    this.seenEvents.add(event.id);
    this.logCopy(event.platform ?? 'unknown', text);

    return { ok: true, action: 'copied', bytes: Buffer.byteLength(text, 'utf8') };
  }

  logCopy(platform, text) {
    const suffix = this.logText ? ` text=${JSON.stringify(text)}` : ` bytes=${Buffer.byteLength(text, 'utf8')}`;
    console.error(`[copyagent] copied platform=${platform}${suffix}`);
  }
}
