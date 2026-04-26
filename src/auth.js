import { timingSafeEqual } from 'node:crypto';

function safeEqual(left, right) {
  const leftBuffer = Buffer.from(left);
  const rightBuffer = Buffer.from(right);

  if (leftBuffer.length !== rightBuffer.length) {
    return false;
  }

  return timingSafeEqual(leftBuffer, rightBuffer);
}

export function extractToken(request) {
  const authorization = request.headers.get('authorization') ?? '';
  if (authorization.toLowerCase().startsWith('bearer ')) {
    return authorization.slice(7).trim();
  }

  const headerToken = request.headers.get('x-copyagent-token');
  if (headerToken) {
    return headerToken.trim();
  }

  return new URL(request.url).searchParams.get('token')?.trim() ?? '';
}

export function isAuthorized(request, expectedToken) {
  if (!expectedToken) {
    return false;
  }

  const actualToken = extractToken(request);
  if (!actualToken) {
    return false;
  }

  return safeEqual(actualToken, expectedToken);
}
