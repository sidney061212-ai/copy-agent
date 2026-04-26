import { spawn, spawnSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';

const __dirname = dirname(fileURLToPath(import.meta.url));
const childPath = join(__dirname, 'memory-child.js');

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function ps(pid) {
  const result = spawnSync('ps', ['-o', 'pid,%cpu,%mem,rss,vsz,command', '-p', String(pid)], { encoding: 'utf8' });
  return result.stdout.trim();
}

async function measureScenario(scenario) {
  const child = spawn(process.execPath, [childPath, scenario], { stdio: ['ignore', 'pipe', 'pipe'] });
  const firstLine = await new Promise((resolve, reject) => {
    child.stdout.once('data', (chunk) => resolve(String(chunk).trim().split('\n')[0]));
    child.once('error', reject);
  });
  await sleep(1500);
  const output = { scenario, initial: JSON.parse(firstLine), ps: ps(child.pid) };
  child.kill();
  return output;
}

const scenarios = ['baseline', 'lark-sdk', 'copyagent-modules'];
const results = [];
for (const scenario of scenarios) {
  results.push(await measureScenario(scenario));
}

const processList = spawnSync('ps', ['axo', 'pid,%cpu,%mem,rss,vsz,command'], { encoding: 'utf8' }).stdout
  .split('\n')
  .filter((line) => /copyagentd|copyagent\.app|copyagent\/src\/cli\.js|\/copyagent(?:\s|$)/.test(line));

console.log(JSON.stringify({ results, processList }, null, 2));
