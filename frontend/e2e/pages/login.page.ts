import { Page } from '@playwright/test';

export class LoginPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto('/login');
  }

  get heading() {
    return this.page.getByRole('heading', { name: 'TrendBird' });
  }

  get subtitle() {
    return this.page.getByText('ログイン');
  }

  get xLoginButton() {
    return this.page.getByRole('button', { name: /X（Twitter）でログイン/ });
  }

  get loggedOutMessage() {
    return this.page.getByText('ログアウトしました');
  }

  get guidanceText() {
    return this.page.getByText('Xアカウントでログインしてすぐに始められます');
  }
}
