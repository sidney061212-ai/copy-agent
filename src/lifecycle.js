import { existsSync, mkdirSync, writeFileSync, unlinkSync } from 'node:fs';
import { homedir } from 'node:os';
import { dirname, join } from 'node:path';
import { spawnSync } from 'node:child_process';

export const LAUNCHD_LABEL = 'local.copyagent';

export function launchAgentPath(home = homedir()) {
  return join(home, 'Library', 'LaunchAgents', `${LAUNCHD_LABEL}.plist`);
}

export function renderLaunchdPlist(options) {
  const env = options.env ?? {};
  const envEntries = Object.entries(env)
    .filter(([, value]) => value !== undefined && value !== '')
    .map(([key, value]) => `    <key>${key}</key>\n    <string>${String(value)}</string>`)
    .join('\n');

  return `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>${LAUNCHD_LABEL}</string>
  <key>ProgramArguments</key>
  <array>
    <string>${options.nodePath}</string>
    <string>${options.cliPath}</string>
    <string>serve</string>
  </array>
  <key>EnvironmentVariables</key>
  <dict>
    <key>LANG</key>
    <string>zh_CN.UTF-8</string>
    <key>LC_ALL</key>
    <string>zh_CN.UTF-8</string>
${envEntries}
  </dict>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
  <key>StandardOutPath</key>
  <string>${options.logPath}</string>
  <key>StandardErrorPath</key>
  <string>${options.logPath}</string>
</dict>
</plist>`;
}

function runLaunchctl(args) {
  const result = spawnSync('launchctl', args, { encoding: 'utf8' });
  if (result.status !== 0) {
    throw new Error(result.stderr.trim() || result.stdout.trim() || `launchctl ${args.join(' ')} failed`);
  }
}

export function installLaunchAgent(options) {
  const plistPath = options.plistPath ?? launchAgentPath();
  mkdirSync(dirname(plistPath), { recursive: true });
  writeFileSync(plistPath, renderLaunchdPlist(options), { mode: 0o600 });
  spawnSync('launchctl', ['unload', plistPath], { encoding: 'utf8' });
  runLaunchctl(['load', plistPath]);
  runLaunchctl(['start', LAUNCHD_LABEL]);
  return plistPath;
}

export function uninstallLaunchAgent(plistPath = launchAgentPath()) {
  if (existsSync(plistPath)) {
    spawnSync('launchctl', ['unload', plistPath], { encoding: 'utf8' });
    unlinkSync(plistPath);
  }

  return plistPath;
}
