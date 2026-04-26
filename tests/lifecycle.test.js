import test from 'node:test';
import assert from 'node:assert/strict';
import { renderLaunchdPlist } from '../src/lifecycle.js';

test('renders launchd plist for copyagent agent', () => {
  const plist = renderLaunchdPlist({
    nodePath: '/node',
    cliPath: '/copyagent/src/cli.js',
    logPath: '/tmp/copyagent.log',
    env: { COPYAGENT_TOKEN: 'token', COPYAGENT_PORT: 8765 },
  });

  assert.match(plist, /local\.copyagent/);
  assert.match(plist, /COPYAGENT_TOKEN/);
  assert.match(plist, /copyagent\/src\/cli\.js/);
});
