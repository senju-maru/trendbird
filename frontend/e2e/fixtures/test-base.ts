import { test as base, BrowserContext } from '@playwright/test';
import { ApiMock } from './api-mock';

type Fixtures = {
  apiMock: ApiMock;
};

export const test = base.extend<Fixtures>({
  apiMock: async ({ page }, use) => {
    const mock = new ApiMock(page);
    await use(mock);
  },
});

export { expect } from '@playwright/test';

/** 認証済みコンテキスト用のヘルパー */
export async function authenticateContext(context: BrowserContext) {
  await context.addCookies([
    {
      name: 'tb_jwt',
      value: 'test-jwt-token-for-e2e',
      domain: 'localhost',
      path: '/',
      httpOnly: true,
      secure: false,
      sameSite: 'Lax',
    },
  ]);
}
