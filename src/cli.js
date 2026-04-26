#!/usr/bin/env node
import { readFileSync } from 'node:fs';
import { homedir } from 'node:os';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { writeClipboard } from './clipboard.js';
import { generateToken, readConfig, requireServerConfig } from './config.js';
import { createFeishuBot } from './feishu-bot.js';
import { installLaunchAgent, launchAgentPath, renderLaunchdPlist, uninstallLaunchAgent } from './lifecycle.js';
import { buildPublicUrl, installAgent, printAgentStatus, setupAgent, toServerConfig, uninstallAgent, updateConfigFile, writePlistPreview } from './manager.js';
import { collectDoctorChecks, collectProfile, summarizeDoctor } from './runtime/diagnostics.js';
import { listen } from './server.js';
import { loadLocalConfig } from './store.js';

const __dirname = dirname(fileURLToPath(import.meta.url));
const packageJson = JSON.parse(readFileSync(join(__dirname, '..', 'package.json'), 'utf8'));

function printHelp() {
  console.log(`copyagent ${packageJson.version}

Usage:
  copyagent serve              Start agent foreground
  copyagent setup              Create ~/.copyagent/config.json
  copyagent start              Install and start background agent
  copyagent stop               Stop and remove background agent
  copyagent status             Show agent status
  copyagent doctor             Run local health checks
  copyagent profile            Show memory/CPU profile
  copyagent url [base-url]     Show or set webhook URLs
  copyagent copy <text>         Copy text directly
  echo text | copyagent copy    Copy stdin directly
  copyagent token              Generate a random token
  copyagent plist              Print a launchd plist
  copyagent install            Install and start launchd agent
  copyagent uninstall          Stop and remove launchd agent
  copyagent help               Show help

Environment:
  COPYAGENT_TOKEN              Required for server mode
  COPYAGENT_HOST               Default: 127.0.0.1
  COPYAGENT_PORT               Default: 8765
  COPYAGENT_MAX_TEXT_BYTES     Default: 200000
  COPYAGENT_LOG_TEXT           Default: false
  COPYAGENT_FEISHU_APP_ID     Feishu custom app App ID
  COPYAGENT_FEISHU_APP_SECRET Feishu custom app App Secret
  COPYAGENT_MODE              Default: feishu-bot
`);
}

async function readStdin() {
  if (process.stdin.isTTY) {
    return '';
  }

  const chunks = [];
  for await (const chunk of process.stdin) {
    chunks.push(Buffer.from(chunk));
  }

  return Buffer.concat(chunks).toString('utf8');
}

function buildLaunchdOptions() {
  const nodePath = process.execPath;
  const cliPath = fileURLToPath(import.meta.url);
  const home = homedir();
  const config = readConfig();

  return {
    nodePath,
    cliPath,
    logPath: `${home}/Library/Logs/copyagent.log`,
    env: {
      COPYAGENT_HOST: config.host,
      COPYAGENT_PORT: config.port,
      COPYAGENT_TOKEN: config.token || 'replace-with-token-from-copyagent-token',
      COPYAGENT_MAX_TEXT_BYTES: config.maxTextBytes,
      COPYAGENT_LOG_TEXT: String(config.logText),
      COPYAGENT_FEISHU_VERIFICATION_TOKEN: config.feishuVerificationToken,
      COPYAGENT_FEISHU_ENCRYPT_KEY: config.feishuEncryptKey,
      COPYAGENT_ALLOWED_ACTOR_IDS: config.allowedActorIds.join(','),
    },
  };
}

function parseOption(args, name) {
  const prefix = `--${name}=`;
  const value = args.find((arg) => arg.startsWith(prefix));
  return value ? value.slice(prefix.length) : '';
}

function printSetup(config) {
  const urls = buildPublicUrl(config);
  const configPath = join(homedir(), '.copyagent', 'config.json');
  console.log(`copyagent configured

Config: ${configPath}
Mode:   ${config.mode || 'feishu-bot'}
Bot:    ${config.feishuAppId ? 'configured' : 'missing App ID/Secret'}
Local:  http://${config.host}:${config.port} (HTTP fallback)
Copy:   ${urls.copy || '(HTTP fallback disabled until you set copyagent url)'}
Feishu: ${urls.feishu || '(not needed in feishu-bot mode)'}
`);
}

async function main() {
  const [command = 'help', ...args] = process.argv.slice(2);

  if (command === 'help' || command === '--help' || command === '-h') {
    printHelp();
    return;
  }

  if (command === 'token') {
    console.log(generateToken());
    return;
  }

  if (command === 'setup') {
    const config = setupAgent({
      baseUrl: parseOption(args, 'base-url') || undefined,
      feishuAppId: parseOption(args, 'feishu-app-id') || undefined,
      feishuAppSecret: parseOption(args, 'feishu-app-secret') || undefined,
      mode: parseOption(args, 'mode') || undefined,
      feishuVerificationToken: parseOption(args, 'feishu-token') || undefined,
      feishuEncryptKey: parseOption(args, 'feishu-encrypt-key') || undefined,
      allowedActorIds: parseOption(args, 'allow-actors') ? parseOption(args, 'allow-actors').split(',').map((item) => item.trim()).filter(Boolean) : undefined,
    });
    printSetup(config);
    return;
  }

  if (command === 'start') {
    const { config, plistPath } = installAgent();
    console.log(`copyagent started
Plist: ${plistPath}
Local: http://${config.host}:${config.port}
`);
    return;
  }

  if (command === 'stop') {
    const plistPath = uninstallAgent();
    console.log(`copyagent stopped: ${plistPath}`);
    return;
  }

  if (command === 'status') {
    const status = printAgentStatus();
    console.log(JSON.stringify({
      configured: status.configured,
      installed: status.installed,
      running: status.launchctlOk,
      configDir: status.configDir,
      plistPath: status.plistPath,
      urls: status.urls,
    }, null, 2));
    return;
  }

  if (command === 'doctor') {
    const summary = summarizeDoctor(collectDoctorChecks());
    for (const check of summary.checks) {
      console.log(`${check.ok ? '✅' : '❌'} ${check.name}: ${check.detail ?? ''}`);
    }
    if (!summary.ok) {
      process.exitCode = 1;
    }
    return;
  }

  if (command === 'profile') {
    const rows = collectProfile();
    if (rows.length === 0) {
      console.log('No copy-agent related processes found.');
      return;
    }
    for (const row of rows) {
      console.log(`${row.pid}\t${row.rssMb} MB\tCPU ${row.cpu}%\t${row.command}`);
    }
    return;
  }

  if (command === 'url') {
    const baseUrl = args[0];
    const config = baseUrl ? updateConfigFile({ baseUrl }) : loadLocalConfig();
    if (!config) {
      throw new Error('copyagent is not configured; run copyagent setup first');
    }
    console.log(JSON.stringify(buildPublicUrl(config), null, 2));
    return;
  }

  if (command === 'plist') {
    console.log(loadLocalConfig() ? writePlistPreview() : renderLaunchdPlist(buildLaunchdOptions()));
    return;
  }

  if (command === 'install') {
    const config = requireServerConfig();
    const plistPath = installLaunchAgent({ ...buildLaunchdOptions(), env: { ...buildLaunchdOptions().env, COPYAGENT_TOKEN: config.token } });
    console.error(`[copyagent] installed launchd agent: ${plistPath}`);
    return;
  }

  if (command === 'uninstall') {
    const plistPath = uninstallLaunchAgent(launchAgentPath());
    console.error(`[copyagent] uninstalled launchd agent: ${plistPath}`);
    return;
  }

  if (command === 'copy') {
    const text = args.length > 0 ? args.join(' ') : await readStdin();
    if (!text.trim()) {
      throw new Error('text is required');
    }

    await writeClipboard(text);
    console.error(`[copyagent] copied bytes=${Buffer.byteLength(text, 'utf8')}`);
    return;
  }

  if (command === 'serve') {
    const config = loadLocalConfig() ? toServerConfig(loadLocalConfig()) : requireServerConfig();
    if ((config.mode ?? 'feishu-bot') === 'feishu-bot') {
      createFeishuBot(config).start();
      console.error('[copyagent] feishu bot long-connection agent started');
      await new Promise(() => {});
      return;
    }

    await listen(config);
    console.error(`[copyagent] listening on http://${config.host}:${config.port}`);
    return;
  }

  throw new Error(`unknown command: ${command}`);
}

main().catch((error) => {
  console.error(`[copyagent] ${error.message}`);
  process.exitCode = 1;
});
