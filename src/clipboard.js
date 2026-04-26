import { spawn } from 'node:child_process';

export function writeClipboard(text) {
  return new Promise((resolve, reject) => {
    const child = spawn('pbcopy', [], { stdio: ['pipe', 'ignore', 'pipe'] });
    const stderr = [];

    child.stderr.on('data', (chunk) => stderr.push(chunk));
    child.on('error', (error) => {
      reject(new Error(`failed to start pbcopy: ${error.message}`));
    });
    child.on('close', (code) => {
      if (code === 0) {
        resolve();
        return;
      }

      const detail = Buffer.concat(stderr).toString('utf8').trim();
      reject(new Error(`pbcopy exited with code ${code}${detail ? `: ${detail}` : ''}`));
    });

    child.stdin.end(text);
  });
}
