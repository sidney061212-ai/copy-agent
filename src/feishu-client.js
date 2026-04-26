export class FeishuApiClient {
  constructor(options) {
    this.appId = options.appId;
    this.appSecret = options.appSecret;
    this.fetch = options.fetch ?? globalThis.fetch;
    this.baseUrl = options.baseUrl ?? 'https://open.feishu.cn/open-apis';
    this.cachedToken = null;
    this.tokenExpiresAt = 0;
  }

  async requestJson(path, init = {}) {
    const response = await this.fetch(`${this.baseUrl}${path}`, init);
    if (!response.ok) {
      throw new Error(`Feishu API HTTP ${response.status}`);
    }

    const data = await response.json();
    if (data.code !== 0) {
      throw new Error(data.msg || `Feishu API error ${data.code}`);
    }

    return data;
  }

  async getTenantAccessToken() {
    const now = Date.now();
    if (this.cachedToken && now < this.tokenExpiresAt) {
      return this.cachedToken;
    }

    const data = await this.requestJson('/auth/v3/tenant_access_token/internal', {
      method: 'POST',
      headers: { 'content-type': 'application/json; charset=utf-8' },
      body: JSON.stringify({ app_id: this.appId, app_secret: this.appSecret }),
    });

    this.cachedToken = data.tenant_access_token;
    this.tokenExpiresAt = now + Math.max(60, (data.expire ?? 7200) - 300) * 1000;
    return this.cachedToken;
  }

  async authorizedFetch(path, init = {}) {
    const token = await this.getTenantAccessToken();
    const headers = new Headers(init.headers ?? {});
    headers.set('authorization', `Bearer ${token}`);
    return this.fetch(`${this.baseUrl}${path}`, { ...init, headers });
  }

  async replyText(messageId, text) {
    if (!messageId) {
      return null;
    }

    return this.requestJson(`/im/v1/messages/${encodeURIComponent(messageId)}/reply`, {
      method: 'POST',
      headers: {
        authorization: `Bearer ${await this.getTenantAccessToken()}`,
        'content-type': 'application/json; charset=utf-8',
      },
      body: JSON.stringify({ msg_type: 'text', content: JSON.stringify({ text }) }),
    });
  }

  async downloadMessageResource(messageId, fileKey, type) {
    const response = await this.authorizedFetch(`/im/v1/messages/${encodeURIComponent(messageId)}/resources/${encodeURIComponent(fileKey)}?type=${encodeURIComponent(type)}`);
    if (!response.ok) {
      throw new Error(`Feishu resource download HTTP ${response.status}`);
    }

    return Buffer.from(await response.arrayBuffer());
  }
}
