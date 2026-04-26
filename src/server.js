import http from 'node:http';
import { CopyAgent } from './agent.js';
import { isAuthorized } from './auth.js';
import { writeClipboard as defaultWriteClipboard } from './clipboard.js';
import { HttpError } from './errors.js';
import { parseFeishuRequest } from './platforms/feishu.js';
import { parseGenericRequest } from './platforms/generic.js';

function jsonResponse(body, init = {}) {
  return new Response(JSON.stringify(body), {
    status: init.status ?? 200,
    headers: { 'content-type': 'application/json; charset=utf-8', ...(init.headers ?? {}) },
  });
}

function textResponse(text, init = {}) {
  return new Response(text, {
    status: init.status ?? 200,
    headers: { 'content-type': 'text/plain; charset=utf-8', ...(init.headers ?? {}) },
  });
}

function cloneWithBody(request, rawBody) {
  return new Request(request.url, {
    method: request.method,
    headers: request.headers,
    body: rawBody.length ? rawBody : undefined,
  });
}

async function readRawBody(request) {
  return Buffer.from(await request.arrayBuffer());
}

export function createApp(options = {}) {
  const config = {
    token: options.token ?? '',
    maxTextBytes: options.maxTextBytes ?? 200_000,
    logText: options.logText ?? false,
    feishuVerificationToken: options.feishuVerificationToken ?? '',
    feishuEncryptKey: options.feishuEncryptKey ?? '',
    allowedActorIds: options.allowedActorIds ?? [],
  };
  const agent = options.agent ?? new CopyAgent({
    writeClipboard: options.writeClipboard ?? defaultWriteClipboard,
    maxTextBytes: config.maxTextBytes,
    logText: config.logText,
    allowedActorIds: config.allowedActorIds,
  });

  function requireAuthorized(request) {
    if (!isAuthorized(request, config.token)) {
      throw new HttpError(401, 'unauthorized');
    }
  }

  async function handleCopy(request) {
    requireAuthorized(request);
    const parsed = await parseGenericRequest(request, { maxTextBytes: config.maxTextBytes });
    const result = await agent.handleEvent(parsed.event);
    return jsonResponse(result);
  }

  async function handleFeishu(request) {
    const rawBody = await readRawBody(request);
    const parsed = await parseFeishuRequest(cloneWithBody(request, rawBody), {
      maxTextBytes: config.maxTextBytes,
      verificationToken: config.feishuVerificationToken,
      encryptKey: config.feishuEncryptKey,
    });

    if (parsed.kind === 'challenge') {
      return jsonResponse({ challenge: parsed.challenge });
    }

    if (!config.feishuEncryptKey && !config.feishuVerificationToken) {
      requireAuthorized(cloneWithBody(request, rawBody));
    }

    const result = await agent.handleEvent(parsed.event);
    return jsonResponse(result);
  }

  return {
    async fetch(request) {
      try {
        const url = new URL(request.url);

        if (request.method === 'GET' && url.pathname === '/health') {
          return jsonResponse({ ok: true, service: 'copyagent', agent: 'CopyAgent', platform: process.platform });
        }

        if (request.method === 'POST' && url.pathname === '/copy') {
          return await handleCopy(request);
        }

        if (request.method === 'POST' && url.pathname === '/feishu') {
          return await handleFeishu(request);
        }

        return textResponse('not found', { status: 404 });
      } catch (error) {
        if (error instanceof HttpError) {
          return jsonResponse({ ok: false, error: error.message }, { status: error.status });
        }

        console.error('[copyagent] internal error', error);
        return jsonResponse({ ok: false, error: 'internal server error' }, { status: 500 });
      }
    },
  };
}

function requestFromIncoming(incoming, body) {
  const host = incoming.headers.host ?? '127.0.0.1';
  return new Request(`http://${host}${incoming.url}`, {
    method: incoming.method,
    headers: incoming.headers,
    body: body.length ? Buffer.concat(body) : undefined,
  });
}

export function listen(config, options = {}) {
  const app = createApp({ ...config, writeClipboard: options.writeClipboard });
  const server = http.createServer((incoming, outgoing) => {
    const chunks = [];

    incoming.on('data', (chunk) => chunks.push(chunk));
    incoming.on('end', async () => {
      const response = await app.fetch(requestFromIncoming(incoming, chunks));
      outgoing.writeHead(response.status, Object.fromEntries(response.headers));
      outgoing.end(Buffer.from(await response.arrayBuffer()));
    });
    incoming.on('error', () => {
      outgoing.writeHead(400, { 'content-type': 'application/json; charset=utf-8' });
      outgoing.end(JSON.stringify({ ok: false, error: 'request stream error' }));
    });
  });

  return new Promise((resolve, reject) => {
    server.once('error', reject);
    server.listen(config.port, config.host, () => {
      server.off('error', reject);
      resolve(server);
    });
  });
}
