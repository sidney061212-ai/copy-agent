import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtempSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { createDefaultConfig, loadLocalConfig, saveLocalConfig } from '../src/store.js';

test('creates and loads local config', () => {
  const dir = mkdtempSync(join(tmpdir(), 'copyagent-store-'));
  try {
    const config = createDefaultConfig({ token: 'token', port: 9999 });
    saveLocalConfig(config, dir);

    assert.deepEqual(loadLocalConfig(dir), config);
  } finally {
    rmSync(dir, { recursive: true, force: true });
  }
});
