import * as Lark from '@larksuiteoapi/node-sdk';
import { CopyAgent } from './agent.js';
import { executePlan } from './actions/executor.js';
import { createFeishuActionAdapters } from './actions/feishu-adapters.js';
import { extractCopyCommandText } from './commands.js';
import { FeishuApiClient } from './feishu-client.js';
import { handleFeishuMediaEvent } from './media-agent.js';
import { validateText } from './payloads.js';
import { formatActionReply } from './replies.js';
import { planEvent } from './core/planner.js';
import { applyPolicy, createDedupPolicy } from './core/policy.js';
import { normalizeFeishuMessageEvent } from './transports/feishu/normalize.js';

export function extractLarkTextEvent(data, options = {}) {
  const message = data?.message;
  if (!message || message.message_type !== 'text') {
    return null;
  }

  let content;
  try {
    content = JSON.parse(message.content ?? '{}');
  } catch {
    return null;
  }

  const text = validateText(extractCopyCommandText(content.text), { maxBytes: options.maxTextBytes ?? 200_000 });
  const eventId = data?.header?.event_id ?? data?.event_id ?? message.message_id ?? '';
  const actorId = data?.sender?.sender_id?.open_id
    ?? data?.sender?.sender_id?.user_id
    ?? data?.sender?.sender_id?.union_id
    ?? '';

  return {
    platform: 'feishu',
    type: 'copy_text',
    text,
    id: eventId,
    actorId,
    messageId: message.message_id ?? '',
  };
}

export function createFeishuEventHandler(agent, options = {}) {
  const apiClient = options.apiClient;
  const dedupPolicy = options.dedupPolicy ?? createDedupPolicy();
  return async (data) => {
    if (options.useCoreExecutor) {
      const event = normalizeFeishuMessageEvent(data);
      const policy = applyPolicy(event, { allowedActorIds: options.allowedActorIds ?? [], dedupPolicy });
      const plan = planEvent(event, { imageAction: options.imageAction });
      if (!policy.ok) {
        const policyPlan = planEvent({ ...event, policy: { allowed: policy.reason !== 'actor is not allowed', duplicate: policy.duplicate, reason: policy.reason } });
        return executePlan(policyPlan, createFeishuActionAdapters({ apiClient, config: options }));
      }
      return executePlan(plan, createFeishuActionAdapters({ apiClient, config: options }));
    }

    const mediaResult = await handleFeishuMediaEvent({ data, config: options, apiClient: apiClient ?? {} });
    if (mediaResult) {
      return mediaResult;
    }

    const event = extractLarkTextEvent(data, options);
    if (!event) {
      return { ok: true, ignored: true };
    }

    try {
      const result = await agent.handleEvent(event);
      if (apiClient?.replyText && options.replyEnabled !== false) {
        await apiClient.replyText(event.messageId, formatActionReply({ action: 'copied_text' }));
      }
      return result;
    } catch (error) {
      if (apiClient?.replyText && options.replyEnabled !== false) {
        await apiClient.replyText(event.messageId, formatActionReply({ action: 'failed', error: error.message }));
      }
      throw error;
    }
  };
}

export function createFeishuBot(config, options = {}) {
  if (!config.feishuAppId || !config.feishuAppSecret) {
    throw new Error('Feishu bot mode requires app_id and app_secret. Run: copyagent setup --feishu-app-id=... --feishu-app-secret=...');
  }

  const agent = options.agent ?? new CopyAgent({
    maxTextBytes: config.maxTextBytes ?? 200_000,
    logText: config.logText ?? false,
    allowedActorIds: config.allowedActorIds ?? [],
  });
  const apiClient = options.apiClient ?? new FeishuApiClient({
    appId: config.feishuAppId,
    appSecret: config.feishuAppSecret,
  });
  const baseConfig = {
    appId: config.feishuAppId,
    appSecret: config.feishuAppSecret,
  };
  const wsClient = options.wsClient ?? new Lark.WSClient({
    ...baseConfig,
    loggerLevel: Lark.LoggerLevel.info,
  });
  const eventDispatcher = options.eventDispatcher ?? new Lark.EventDispatcher({
    encryptKey: config.feishuEncryptKey || undefined,
  }).register({
    'im.message.receive_v1': createFeishuEventHandler(agent, { ...config, apiClient, useCoreExecutor: true }),
  });

  return {
    start() {
      wsClient.start({ eventDispatcher });
      return wsClient;
    },
  };
}
