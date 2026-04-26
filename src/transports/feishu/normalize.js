function parseContent(message) {
  try {
    return JSON.parse(message?.content ?? '{}');
  } catch {
    return {};
  }
}

function actorId(data) {
  return data?.sender?.sender_id?.open_id
    ?? data?.sender?.sender_id?.user_id
    ?? data?.sender?.sender_id?.union_id
    ?? '';
}

function eventId(data) {
  return data?.header?.event_id ?? data?.event_id ?? data?.message?.message_id ?? '';
}

function replyTarget(message) {
  return message?.message_id ? { messageId: message.message_id } : undefined;
}

export function normalizeFeishuMessageEvent(data) {
  const message = data?.message ?? {};
  const base = {
    transport: 'feishu',
    id: eventId(data),
    actorId: actorId(data),
    replyTarget: replyTarget(message),
  };

  if (message.message_type === 'text') {
    const content = parseContent(message);
    return { ...base, type: 'text', text: content.text ?? '' };
  }

  if (message.message_type === 'image') {
    const content = parseContent(message);
    return {
      ...base,
      type: 'image',
      resource: {
        id: content.image_key,
        name: content.file_name ?? `${message.message_id || 'image'}.png`,
        kind: 'image',
        messageId: message.message_id,
      },
    };
  }

  if (message.message_type === 'file') {
    const content = parseContent(message);
    return {
      ...base,
      type: 'file',
      resource: {
        id: content.file_key,
        name: content.file_name ?? `${message.message_id || 'file'}`,
        kind: 'file',
        messageId: message.message_id,
      },
    };
  }

  return { ...base, type: 'unknown' };
}
