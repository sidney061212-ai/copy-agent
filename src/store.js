import { existsSync, mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { homedir } from 'node:os';
import { join } from 'node:path';
import { generateToken } from './config.js';

export function configDir(home = homedir()) {
  return join(home, '.copyagent');
}

export function configPath(dir = configDir()) {
  return join(dir, 'config.json');
}

export function createDefaultConfig(overrides = {}) {
  return {
    host: '127.0.0.1',
    port: 8765,
    token: generateToken(),
    baseUrl: '',
    mode: 'feishu-bot',
    feishuAppId: '',
    feishuAppSecret: '',
    feishuVerificationToken: '',
    feishuEncryptKey: '',
    allowedActorIds: [],
    defaultDownloadDir: '~/Downloads/copyagent',
    imageAction: 'clipboard',
    replyEnabled: true,
    ...overrides,
  };
}

export function loadLocalConfig(dir = configDir()) {
  const path = configPath(dir);
  if (!existsSync(path)) {
    return null;
  }

  return JSON.parse(readFileSync(path, 'utf8'));
}

export function saveLocalConfig(config, dir = configDir()) {
  mkdirSync(dir, { recursive: true, mode: 0o700 });
  writeFileSync(configPath(dir), `${JSON.stringify(config, null, 2)}\n`, { mode: 0o600 });
}

export function ensureLocalConfig(overrides = {}, dir = configDir()) {
  const existing = loadLocalConfig(dir);
  const next = existing ? { ...existing, ...overrides } : createDefaultConfig(overrides);
  saveLocalConfig(next, dir);
  return next;
}
