import { mkdirSync, writeFileSync } from 'node:fs';
import { basename, join } from 'node:path';
import { spawn } from 'node:child_process';
import { resolveSaveDirectory } from './media-rules.js';
import { formatActionReply } from './replies.js';

function parseContent(message) {
  try {
    return JSON.parse(message?.content ?? '{}');
  } catch {
    return {};
  }
}

function safeFileName(name, fallback) {
  const cleaned = basename(String(name || fallback)).replace(/[\u0000-\u001f]/g, '').trim();
  return cleaned || fallback;
}

function getResourceInfo(message) {
  const content = parseContent(message);
  if (message.message_type === 'image') {
    return {
      type: 'image',
      key: content.image_key,
      fileName: safeFileName(content.file_name, `${message.message_id || 'image'}.png`),
    };
  }

  if (message.message_type === 'file') {
    return {
      type: 'file',
      key: content.file_key,
      fileName: safeFileName(content.file_name, `${message.message_id || 'file'}`),
    };
  }

  return null;
}

export function copyImageToClipboard(buffer) {
  return new Promise((resolve, reject) => {
    const script = 'set the clipboard to (read (POSIX file (system attribute "COPYAGENT_IMAGE_PATH")) as «class PNGf»)';
    reject(new Error('copyImageToClipboard requires file path wrapper'));
  });
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

async function maybeReply(apiClient, config, messageId, result) {
  if (config.replyEnabled === false || !apiClient.replyText) {
    return;
  }

  await apiClient.replyText(messageId, formatActionReply(result));
}

export async function handleFeishuMediaEvent({ data, config, apiClient }) {
  const message = data?.message;
  const resource = getResourceInfo(message);
  if (!resource) {
    return null;
  }

  try {
    if (!resource.key) {
      throw new Error('missing file key');
    }

    const buffer = await apiClient.downloadMessageResource(message.message_id, resource.key, resource.type);
    const directory = resolveSaveDirectory({
      defaultDownloadDir: config.defaultDownloadDir || '~/Downloads/copyagent',
      hint: config.saveHint || '',
    });
    mkdirSync(directory, { recursive: true });
    const path = join(directory, resource.fileName);
    writeFileSync(path, buffer);

    if (resource.type === 'image' && config.imageAction !== 'save') {
      await copyImageFileToClipboard(path);
      const result = { action: 'copied_image', path };
      await maybeReply(apiClient, config, message.message_id, result);
      return result;
    }

    const result = { action: 'saved_file', path };
    await maybeReply(apiClient, config, message.message_id, result);
    return result;
  } catch (error) {
    const result = { action: 'failed', error: error.message };
    await maybeReply(apiClient, config, message?.message_id, result);
    return result;
  }
}
