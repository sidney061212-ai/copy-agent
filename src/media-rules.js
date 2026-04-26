import { homedir } from 'node:os';
import { join, resolve } from 'node:path';

const ALIASES = new Map([
  ['桌面', 'Desktop'],
  ['desktop', 'Desktop'],
  ['下载', 'Downloads'],
  ['downloads', 'Downloads'],
  ['文档', 'Documents'],
  ['documents', 'Documents'],
]);

export function expandHome(path) {
  if (!path) {
    return path;
  }

  if (path === '~') {
    return homedir();
  }

  if (path.startsWith('~/')) {
    return join(homedir(), path.slice(2));
  }

  return path;
}

export function normalizePathAlias(value = '') {
  const trimmed = value.trim().replace(/^保存到\s*[:：]\s*/u, '').replace(/^save\s*[:：]\s*/iu, '');
  return ALIASES.get(trimmed.toLowerCase()) ?? ALIASES.get(trimmed) ?? trimmed;
}

export function resolveSaveDirectory({ defaultDownloadDir, hint = '' }) {
  const normalized = normalizePathAlias(hint);
  if (!normalized) {
    return resolve(expandHome(defaultDownloadDir));
  }

  if (normalized.startsWith('/') || normalized.startsWith('~/')) {
    return resolve(expandHome(normalized));
  }

  return resolve(join(homedir(), normalized));
}
