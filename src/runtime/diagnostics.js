import { accessSync, constants, existsSync, mkdirSync } from 'node:fs';
import { spawnSync } from 'node:child_process';
import { launchAgentPath } from '../lifecycle.js';
import { loadLocalConfig } from '../store.js';
import { expandHome } from '../media-rules.js';

const SECRET_KEYS = new Set([
  'token',
  'appSecret',
  'feishuAppSecret',
  'feishuEncryptKey',
  'secret',
  'apiKey',
]);

export function redactObject(value) {
  if (Array.isArray(value)) {
    return value.map(redactObject);
  }

  if (!value || typeof value !== 'object') {
    return value;
  }

  return Object.fromEntries(Object.entries(value).map(([key, entry]) => [
    key,
    SECRET_KEYS.has(key) ? '***REDACTED***' : redactObject(entry),
  ]));
}

export function summarizeDoctor(checks) {
  const failed = checks.filter((check) => !check.ok);
  return { ok: failed.length === 0, failed, checks };
}

export function parsePsRows(output) {
  return output
    .trim()
    .split('\n')
    .slice(1)
    .map((line) => line.trim())
    .filter(Boolean)
    .map((line) => {
      const match = line.match(/^(\d+)\s+([\d.]+)\s+([\d.]+)\s+(\d+)\s+(.+)$/);
      if (!match) {
        return null;
      }

      return {
        pid: Number.parseInt(match[1], 10),
        cpu: Number.parseFloat(match[2]),
        mem: Number.parseFloat(match[3]),
        rssKb: Number.parseInt(match[4], 10),
        command: match[5],
      };
    })
    .filter(Boolean);
}

function checkPathWritable(path) {
  try {
    accessSync(path, constants.W_OK);
    return true;
  } catch {
    return false;
  }
}

function launchctlRunning() {
  const result = spawnSync('launchctl', ['print', `gui/${process.getuid()}/local.copyagent`], { encoding: 'utf8' });
  return result.status === 0;
}

function clipboardAvailable() {
  const result = spawnSync('which', ['pbcopy'], { encoding: 'utf8' });
  return result.status === 0;
}

export function collectDoctorChecks() {
  const config = loadLocalConfig();
  const downloadDir = config?.defaultDownloadDir ? expandHome(config.defaultDownloadDir) : '';

  return [
    { name: 'config', ok: Boolean(config), detail: config ? 'configured' : 'missing ~/.copyagent/config.json' },
    { name: 'mode', ok: config?.mode === 'feishu-bot', detail: config?.mode ?? 'unknown' },
    { name: 'feishu_app_id', ok: Boolean(config?.feishuAppId), detail: config?.feishuAppId ? 'configured' : 'missing' },
    { name: 'feishu_app_secret', ok: Boolean(config?.feishuAppSecret), detail: config?.feishuAppSecret ? 'configured' : 'missing' },
    { name: 'launchd_plist', ok: existsSync(launchAgentPath()), detail: launchAgentPath() },
    { name: 'launchd_running', ok: launchctlRunning(), detail: 'local.copyagent' },
    { name: 'clipboard', ok: clipboardAvailable(), detail: 'pbcopy' },
    { name: 'download_dir', ok: ensureWritableDir(downloadDir), detail: downloadDir || 'not configured' },
  ];
}

function ensureWritableDir(path) {
  if (!path) {
    return false;
  }

  try {
    mkdirSync(path, { recursive: true });
    return existsSync(path) && checkPathWritable(path);
  } catch {
    return false;
  }
}

export function collectProfile() {
  const result = spawnSync('ps', ['axo', 'pid,%cpu,%mem,rss,command'], { encoding: 'utf8' });
  const rows = parsePsRows(result.stdout);
  const interesting = rows.filter((row) => /copyagentd|copyagent\.app|copyagent\/src\/cli\.js|\/copyagent(?:\s|$)/.test(row.command));
  return interesting.map((row) => ({ ...row, rssMb: Math.round((row.rssKb / 1024) * 10) / 10 }));
}
