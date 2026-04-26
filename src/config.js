import { randomBytes } from 'node:crypto';

const DEFAULT_HOST = '127.0.0.1';
const DEFAULT_PORT = 8765;
const DEFAULT_MAX_TEXT_BYTES = 200_000;

function parseInteger(value, fallback) {
  if (value === undefined || value === '') {
    return fallback;
  }

  const parsed = Number.parseInt(value, 10);
  if (!Number.isInteger(parsed) || parsed <= 0) {
    throw new Error(`invalid positive integer: ${value}`);
  }

  return parsed;
}

function parseList(value) {
  if (!value) {
    return [];
  }

  return value.split(',').map((item) => item.trim()).filter(Boolean);
}

export function readConfig(env = process.env) {
  return {
    host: env.COPYAGENT_HOST || DEFAULT_HOST,
    port: parseInteger(env.COPYAGENT_PORT, DEFAULT_PORT),
    token: env.COPYAGENT_TOKEN || '',
    maxTextBytes: parseInteger(env.COPYAGENT_MAX_TEXT_BYTES, DEFAULT_MAX_TEXT_BYTES),
    logText: env.COPYAGENT_LOG_TEXT === 'true',
    feishuVerificationToken: env.COPYAGENT_FEISHU_VERIFICATION_TOKEN || '',
    feishuEncryptKey: env.COPYAGENT_FEISHU_ENCRYPT_KEY || '',
    feishuAppId: env.COPYAGENT_FEISHU_APP_ID || '',
    feishuAppSecret: env.COPYAGENT_FEISHU_APP_SECRET || '',
    mode: env.COPYAGENT_MODE || 'feishu-bot',
    allowedActorIds: parseList(env.COPYAGENT_ALLOWED_ACTOR_IDS),
    defaultDownloadDir: env.COPYAGENT_DEFAULT_DOWNLOAD_DIR || '~/Downloads/copyagent',
    imageAction: env.COPYAGENT_IMAGE_ACTION || 'clipboard',
    replyEnabled: env.COPYAGENT_REPLY_ENABLED !== 'false',
  };
}

export function requireServerConfig(env = process.env) {
  const config = readConfig(env);
  if (!config.token) {
    throw new Error('COPYAGENT_TOKEN is required for server mode');
  }

  return config;
}

export function generateToken() {
  return randomBytes(32).toString('base64url');
}
