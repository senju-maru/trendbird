import { Page, Locator } from '@playwright/test';

type TabId = 'profile' | 'notifications' | 'twitter' | 'account';

const TAB_LABELS: Record<TabId, string> = {
  profile: 'プロフィール',
  notifications: '通知',
  twitter: 'X連携',
  account: 'アカウント',
};

export class SettingsPage {
  constructor(private page: Page) {}

  // ── Navigation ──

  async goto(tab?: TabId) {
    const url = tab ? `/settings?tab=${tab}` : '/settings';
    await this.page.goto(url);
    await this.page.waitForLoadState('networkidle');
  }

  async switchTab(tab: TabId) {
    // TabsTrigger は role="tab" を持たないカスタム <button>
    await this.page.getByRole('button', { name: TAB_LABELS[tab], exact: true }).click();
  }

  // ── Profile tab ──

  get heading(): Locator {
    return this.page.getByRole('heading', { name: '設定' });
  }

  get emailInput(): Locator {
    return this.page.getByPlaceholder('email@example.com');
  }

  get saveButton(): Locator {
    return this.page.getByRole('button', { name: '保存する' });
  }

  // ── Notifications tab ──

  notificationToggle(label: string): Locator {
    // Toggle DOM: <div(root)> → <div(label)> → <div>{label}</div> + <button>
    // getByText → label div → .. → label container → .. → root → button
    return this.page.getByText(label, { exact: true }).locator('../..').locator('button');
  }

  // ── Twitter tab ──

  get connectButton(): Locator {
    return this.page.getByRole('button', { name: 'Xアカウントを連携' });
  }

  get reconnectButton(): Locator {
    return this.page.getByRole('button', { name: '再接続' });
  }

  get disconnectButton(): Locator {
    return this.page.getByRole('button', { name: '連携解除' });
  }

  // ── Account tab ──

  get logoutButton(): Locator {
    return this.page.getByRole('main').getByRole('button', { name: 'ログアウト' });
  }

  get deleteAccountButton(): Locator {
    return this.page.getByRole('button', { name: 'アカウントを削除' });
  }
}
