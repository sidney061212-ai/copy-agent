const scenario = process.argv[2] || 'baseline';

if (scenario === 'lark-sdk') {
  await import('@larksuiteoapi/node-sdk');
}

if (scenario === 'copyagent-modules') {
  await import('../src/feishu-bot.js');
  await import('../src/actions/feishu-adapters.js');
  await import('../src/runtime/diagnostics.js');
}

console.log(JSON.stringify({ scenario, pid: process.pid, memory: process.memoryUsage() }));
setInterval(() => {}, 60_000);
