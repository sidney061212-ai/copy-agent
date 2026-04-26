import { mkdirSync, writeFileSync } from 'node:fs';
import { basename, join } from 'node:path';
import { spawn } from 'node:child_process';
import { writeClipboard } from '../clipboard.js';
import { resolveSaveDirectory } from '../media-rules.js';

function safeFileName(name, fallback) {
  const cleaned = basename(String(name || fallback)).replace(/[\u0000-\u001f]/g, '').trim();
  return cleaned || fallback;
}

function copyImageFileToClipboard(path) {
  return new Promise((resolve, reject) => {
    const script = `set the clipboard to (read (POSIX file ${JSON.stringify(path)}) as «class PNGf»)`;
    const child = spawn('osascript', ['-e', script], { stdio: ['ignore', 'ignore', 'pipe'] });
    const stderr = [];
    child.stderr.on('data', (chunk) => stderr.push(chunk));
    child.on('error', reject);
    child.on('close', (code) => {
      if (code === 0) {
        resolve();
        return;
      }
      reject(new Error(Buffer.concat(stderr).toString('utf8').trim() || `osascript exited ${code}`));
    });
  });
}

export function createFeishuActionAdapters({ apiClient, config }) {
  const savedResources = new Map();

  return {
    clipboard: {
      copyText: writeClipboard,
      async copyImage(resource) {
        const path = savedResources.get(resource.id);
        if (!path) {
          throw new Error('image resource must be saved before copy');
        }
        await copyImageFileToClipboard(path);
      },
    },
    filesystem: {
      async saveResource(resource, options = {}) {
        if (!resource?.id) {
          throw new Error('missing resource id');
        }

        const buffer = await apiClient.downloadMessageResource(resource.messageId, resource.id, resource.kind);
        const directory = resolveSaveDirectory({
          defaultDownloadDir: config.defaultDownloadDir || '~/Downloads/copyagent',
          hint: options.directoryHint || '',
        });
        mkdirSync(directory, { recursive: true });
        const path = join(directory, safeFileName(resource.name, resource.kind === 'image' ? 'image.png' : 'file'));
        writeFileSync(path, buffer);
        savedResources.set(resource.id, path);
        return path;
      },
    },
    reply: {
      async send(target, text) {
        await apiClient.replyText(target.messageId, text);
      },
    },
  };
}
