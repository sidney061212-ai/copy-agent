import { spawnSync } from 'node:child_process';
import { existsSync, readFileSync, writeFileSync } from 'node:fs';
import { homedir } from 'node:os';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { installLaunchAgent, launchAgentPath, LAUNCHD_LABEL, renderLaunchdPlist, uninstallLaunchAgent } from './lifecycle.js';
import { configDir, ensureLocalConfig, loadLocalConfig, saveLocalConfig } from './store.js';

export function mergeRuntimeConfig(config) {
  return {
    COPYAGENT_HOST: String(config.host),
    COPYAGENT_PORT: String(config.port),
    COPYAGENT_TOKEN: String(config.token),
    COPYAGENT_FEISHU_VERIFICATION_TOKEN: config.feishuVerificationToken || '',
    COPYAGENT_FEISHU_ENCRYPT_KEY: config.feishuEncryptKey || '',
    COPYAGENT_FEISHU_APP_ID: config.feishuAppId || '',
    COPYAGENT_FEISHU_APP_SECRET: config.feishuAppSecret || '',
    COPYAGENT_MODE: config.mode || 'feishu-bot',
    COPYAGENT_ALLOWED_ACTOR_IDS: (config.allowedActorIds ?? []).join(','),
    COPYAGENT_DEFAULT_DOWNLOAD_DIR: config.defaultDownloadDir || '~/Downloads/copyagent',
    COPYAGENT_IMAGE_ACTION: config.imageAction || 'clipboard',
    COPYAGENT_REPLY_ENABLED: String(config.replyEnabled !== false),
  };
}

export function redactConfig(config) {
  return {
    ...config,
    token: config.token ? '***REDACTED***' : '',
    feishuAppSecret: config.feishuAppSecret ? '***REDACTED***' : '',
    feishuEncryptKey: config.feishuEncryptKey ? '***REDACTED***' : '',
  };
}

export function toServerConfig(config) {
  return {
    host: config.host,
    port: config.port,
    token: config.token,
    feishuVerificationToken: config.feishuVerificationToken || '',
    feishuEncryptKey: config.feishuEncryptKey || '',
    feishuAppId: config.feishuAppId || '',
    feishuAppSecret: config.feishuAppSecret || '',
    mode: config.mode || 'feishu-bot',
    allowedActorIds: config.allowedActorIds ?? [],
    defaultDownloadDir: config.defaultDownloadDir || '~/Downloads/copyagent',
    imageAction: config.imageAction || 'clipboard',
    replyEnabled: config.replyEnabled !== false,
  };
}

export function buildPublicUrl(config) {
  if (!config.baseUrl) {
    return { copy: '', feishu: '' };
  }

  const base = config.baseUrl.replace(/\/+$/, '');
  const token = encodeURIComponent(config.token);
  return {
    copy: `${base}/copy?token=${token}`,
    feishu: `${base}/feishu?token=${token}`,
  };
}

export function defaultLaunchdOptions(config) {
  const cliPath = fileURLToPath(new URL('./cli.js', import.meta.url));
  return {
    nodePath: process.execPath,
    cliPath,
    logPath: join(homedir(), 'Library', 'Logs', 'copyagent.log'),
    env: mergeRuntimeConfig(config),
  };
}

export function setupAgent(overrides = {}) {
  return ensureLocalConfig(overrides);
}

export function installAgent() {
  const config = ensureLocalConfig();
  const plistPath = installLaunchAgent(defaultLaunchdOptions(config));
  return { config, plistPath };
}

export function uninstallAgent() {
  return uninstallLaunchAgent(launchAgentPath());
}

export function printAgentStatus() {
  const config = loadLocalConfig();
  const plistPath = launchAgentPath();
  const result = spawnSync('launchctl', ['print', `gui/${process.getuid()}/${LAUNCHD_LABEL}`], { encoding: 'utf8' });
  return {
    configured: Boolean(config),
    configDir: configDir(),
    plistPath,
    installed: existsSync(plistPath),
    launchctlOk: result.status === 0,
    launchctlOutput: result.status === 0 ? result.stdout : result.stderr,
    urls: config ? buildPublicUrl(config) : { copy: '', feishu: '' },
  };
}

export function writePlistPreview(config = ensureLocalConfig()) {
  return renderLaunchdPlist(defaultLaunchdOptions(config));
}

export function updateConfigFile(updates) {
  const config = ensureLocalConfig();
  const next = { ...config, ...updates };
  saveLocalConfig(next);
  return next;
}
