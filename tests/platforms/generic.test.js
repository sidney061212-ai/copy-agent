import test from 'node:test';
import assert from 'node:assert/strict';
import { parseGenericRequest } from '../../src/platforms/generic.js';

test('parses generic copy request into normalized agent event', async () => {
  const request = new Request('http://127.0.0.1/copy', {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ text: 'hello' }),
  });

  const parsed = await parseGenericRequest(request, { maxTextBytes: 1000 });

  assert.deepEqual(parsed, {
    kind: 'event',
    event: { platform: 'generic', type: 'copy_text', text: 'hello', id: '' },
  });
});
