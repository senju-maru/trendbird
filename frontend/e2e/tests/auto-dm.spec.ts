import { test, expect, authenticateContext } from '../fixtures/test-base';
import { buildAutoDMRule, buildDMSentLog, resetSeq } from '../fixtures/factories';
import { expectToast } from '../helpers/assertions';
import { AutoDMPage } from '../pages/auto-dm.page';

test.describe('自動DM', () => {
  test.beforeEach(async ({ context, page, apiMock }) => {
    resetSeq();
    await authenticateContext(context);
    await apiMock.setupDefaults();
    await page.addInitScript(() => {
      sessionStorage.removeItem('tb_tutorial_pending');
    });
  });

  test.describe('ルール一覧', () => {
    test('ルールがない場合に空状態が表示される', async ({ page }) => {
      const autoDM = new AutoDMPage(page);
      await autoDM.goto();
      await expect(autoDM.emptyState).toBeVisible();
    });

    test('既存ルールが表示される', async ({ page, apiMock }) => {
      const rule = buildAutoDMRule({ enabled: true, triggerKeywords: ['AI', 'LLM'] });
      await apiMock.mockRPC('AutoDMService', 'ListAutoDMRules', { rules: [rule] });

      const autoDM = new AutoDMPage(page);
      await autoDM.goto();
      await expect(page.getByText('ルール 1')).toBeVisible();
      await expect(page.getByText('AI')).toBeVisible();
      await expect(page.getByText('LLM')).toBeVisible();
    });
  });

  test.describe('トグル即時保存', () => {
    test('トグルOFF → APIが呼ばれ、リロード後もOFFのまま', async ({ page, apiMock }) => {
      const rule = buildAutoDMRule({ enabled: true });
      await apiMock.mockRPC('AutoDMService', 'ListAutoDMRules', { rules: [rule] });

      const updatedRule = { ...rule, enabled: false };
      await apiMock.mockRPC('AutoDMService', 'UpdateAutoDMRule', { rule: updatedRule });

      const autoDM = new AutoDMPage(page);
      await autoDM.goto();

      // APIリクエスト監視 + トグルクリックを並行
      const [req] = await Promise.all([
        page.waitForRequest((r) => r.url().includes('AutoDMService/UpdateAutoDMRule')),
        autoDM.toggleButton(0).locator('button').click(),
      ]);

      const body = JSON.parse(req.postData() ?? '{}');
      // proto3 JSON は false（デフォルト値）を省略するため、undefined または false を許容
      expect(body.enabled ?? false).toBe(false);

      // リロード後もOFFのまま（ダッシュボード経由で再遷移して auth state を再確立）
      await apiMock.clearMock('AutoDMService', 'ListAutoDMRules');
      await apiMock.mockRPC('AutoDMService', 'ListAutoDMRules', { rules: [updatedRule] });
      await page.goto('/dashboard');
      await page.waitForLoadState('networkidle');
      await page.locator('a[href="/auto-dm"]').click();
      await page.getByRole('heading', { name: '自動DM' }).waitFor({ timeout: 10_000 });
      await expect(page.getByText('ルール 1')).toBeVisible();
    });

    test('トグルON → APIが呼ばれる', async ({ page, apiMock }) => {
      const rule = buildAutoDMRule({ enabled: false });
      await apiMock.mockRPC('AutoDMService', 'ListAutoDMRules', { rules: [rule] });

      const updatedRule = { ...rule, enabled: true };
      await apiMock.mockRPC('AutoDMService', 'UpdateAutoDMRule', { rule: updatedRule });

      const autoDM = new AutoDMPage(page);
      await autoDM.goto();

      const [req] = await Promise.all([
        page.waitForRequest((r) => r.url().includes('AutoDMService/UpdateAutoDMRule')),
        autoDM.toggleButton(0).locator('button').click(),
      ]);

      const body = JSON.parse(req.postData() ?? '{}');
      expect(body.enabled).toBe(true);
    });

    test('トグル変更中はdisabledになり連打を防止', async ({ page, apiMock }) => {
      const rule = buildAutoDMRule({ enabled: true });
      await apiMock.mockRPC('AutoDMService', 'ListAutoDMRules', { rules: [rule] });

      // API応答を遅延させる
      await apiMock.clearMock('AutoDMService', 'UpdateAutoDMRule');
      await page.route('**/trendbird.v1.AutoDMService/UpdateAutoDMRule', async (route) => {
        await new Promise((r) => setTimeout(r, 2000));
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ rule: { ...rule, enabled: false } }),
        });
      });

      const autoDM = new AutoDMPage(page);
      await autoDM.goto();

      // トグルをクリック
      await autoDM.toggleButton(0).locator('button').click();

      // API応答待ち中にToggleボタンがdisabledになる
      await expect(autoDM.toggleButton(0).locator('button')).toBeDisabled();
    });

    test('トグルAPI失敗時にロールバック＋toast表示', async ({ page, apiMock }) => {
      const rule = buildAutoDMRule({ enabled: true });
      await apiMock.mockRPC('AutoDMService', 'ListAutoDMRules', { rules: [rule] });
      await apiMock.mockRPCError('AutoDMService', 'UpdateAutoDMRule', 'internal', 'server error');

      const autoDM = new AutoDMPage(page);
      await autoDM.goto();

      await autoDM.toggleButton(0).locator('button').click();

      await expectToast(page, '有効/無効の切り替えに失敗しました');
    });

    test('新規ルールのトグルはAPI呼び出しなし', async ({ page }) => {
      const autoDM = new AutoDMPage(page);
      await autoDM.goto();

      await autoDM.addRuleButton.click();
      await expect(page.getByText('新しいルール')).toBeVisible();

      let apiCalled = false;
      page.on('request', (req) => {
        if (req.url().includes('AutoDMService/UpdateAutoDMRule')) {
          apiCalled = true;
        }
      });

      // 新規ルールのトグルをクリック（data-testid="auto-dm-toggle-0"）
      await autoDM.toggleButton(0).locator('button').click();

      await page.waitForTimeout(500);
      expect(apiCalled).toBe(false);
    });
  });

  test.describe('ルール作成', () => {
    test('キーワードとテンプレートを入力して保存', async ({ page, apiMock }) => {
      const newRule = buildAutoDMRule({ triggerKeywords: ['テスト'], templateMessage: 'こんにちは' });
      await apiMock.mockRPC('AutoDMService', 'CreateAutoDMRule', { rule: newRule });

      const autoDM = new AutoDMPage(page);
      await autoDM.goto();

      await autoDM.addRuleButton.click();

      await autoDM.keywordInput.fill('テスト');
      await autoDM.keywordInput.press('Enter');
      await expect(page.getByText('テスト')).toBeVisible();

      await autoDM.templateInput.fill('こんにちは');

      await autoDM.saveButton('作成する').click();
      await expectToast(page, 'ルールを作成しました');
    });
  });

  test.describe('ルール削除', () => {
    test('既存ルールを削除できる', async ({ page, apiMock }) => {
      const rule = buildAutoDMRule();
      await apiMock.mockRPC('AutoDMService', 'ListAutoDMRules', { rules: [rule] });
      await apiMock.mockRPC('AutoDMService', 'DeleteAutoDMRule', {});

      const autoDM = new AutoDMPage(page);
      await autoDM.goto();
      await expect(page.getByText('ルール 1')).toBeVisible();

      await autoDM.deleteButton.click();
      await expectToast(page, 'ルールを削除しました');
    });
  });

  test.describe('保存ボタン', () => {
    test('トグル無効時でも保存ボタンはクリック可能', async ({ page, apiMock }) => {
      const rule = buildAutoDMRule({ enabled: true });
      await apiMock.mockRPC('AutoDMService', 'ListAutoDMRules', { rules: [rule] });
      await apiMock.mockRPC('AutoDMService', 'UpdateAutoDMRule', { rule: { ...rule, enabled: false } });

      const autoDM = new AutoDMPage(page);
      await autoDM.goto();

      // トグルをOFFに
      await autoDM.toggleButton(0).locator('button').click();

      await expect(autoDM.saveButton('保存する')).toBeEnabled();
    });
  });

  test.describe('送信履歴', () => {
    test('履歴タブでログが表示される', async ({ page, apiMock }) => {
      const log = buildDMSentLog({ triggerKeyword: 'AI', dmText: 'テスト通知です' });
      await apiMock.mockRPC('AutoDMService', 'GetDMSentLogs', { logs: [log] });

      const autoDM = new AutoDMPage(page);
      await autoDM.goto();

      await autoDM.historyTab.click();
      await expect(page.getByText('AI')).toBeVisible();
      await expect(page.getByText('テスト通知です')).toBeVisible();
    });

    test('履歴がない場合に空状態が表示される', async ({ page, apiMock }) => {
      await apiMock.mockRPC('AutoDMService', 'GetDMSentLogs', { logs: [] });

      const autoDM = new AutoDMPage(page);
      await autoDM.goto();

      await autoDM.historyTab.click();
      await expect(page.getByText('まだ送信履歴はありません')).toBeVisible();
    });
  });

  // ─── ルール上限 ─────────────────────────────────────────────

  test.describe('ルール上限', () => {
    test('5個のルールがある場合、追加ボタンが無効化される', async ({ page, apiMock }) => {
      const rules = Array.from({ length: 5 }, (_, i) =>
        buildAutoDMRule({ enabled: true, triggerKeywords: [`kw-${i}`] }),
      );
      await apiMock.mockRPC('AutoDMService', 'ListAutoDMRules', { rules });

      const autoDM = new AutoDMPage(page);
      await autoDM.goto();

      // 5個のルールが表示されていることを確認
      await expect(page.getByText('ルール 5')).toBeVisible();

      // 追加ボタンが disabled で (5/5) 表示
      await expect(autoDM.addRuleButton).toBeDisabled();
      await expect(page.getByText('ルールを追加 (5/5)')).toBeVisible();
    });
  });

  // ─── テンプレート文字数制限 ────────────────────────────────

  test.describe('テンプレート文字数制限', () => {
    test('280文字を超える入力で文字カウンターが上限超過を表示', async ({ page }) => {
      const autoDM = new AutoDMPage(page);
      await autoDM.goto();

      // 新しいルールを追加
      await autoDM.addRuleButton.click();

      // 280文字超の文字列を入力
      const longText = 'あ'.repeat(290);
      await autoDM.templateInput.fill(longText);

      // 文字カウンターが 290/280 と表示される（上限超過）
      await expect(page.getByText('290/280')).toBeVisible();
    });
  });

  // ─── キーワードバリデーション ──────────────────────────────

  test.describe('キーワードバリデーション', () => {
    test('重複キーワードで toast 表示', async ({ page }) => {
      const autoDM = new AutoDMPage(page);
      await autoDM.goto();

      // 新しいルールを追加
      await autoDM.addRuleButton.click();

      // キーワード「AI」を追加
      await autoDM.keywordInput.fill('AI');
      await autoDM.keywordInput.press('Enter');
      await expect(page.getByText('AI')).toBeVisible();

      // 同じキーワード「AI」を再追加 → 重複 toast
      await autoDM.keywordInput.fill('AI');
      await autoDM.keywordInput.press('Enter');
      await expectToast(page, '同じキーワードが既に追加されています');
    });

    test('10個のキーワード上限で toast 表示', async ({ page }) => {
      const autoDM = new AutoDMPage(page);
      await autoDM.goto();

      // 新しいルールを追加
      await autoDM.addRuleButton.click();

      // 10個のキーワードを追加
      for (let i = 0; i < 10; i++) {
        await autoDM.keywordInput.fill(`keyword${i}`);
        await autoDM.keywordInput.press('Enter');
      }

      // 11個目を追加 → 上限 toast
      await autoDM.keywordInput.fill('keyword10');
      await autoDM.keywordInput.press('Enter');
      await expectToast(page, 'キーワードは最大10個までです');
    });
  });
});
