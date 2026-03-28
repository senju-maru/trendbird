import { test, expect, authenticateContext } from '../fixtures/test-base';
import { buildAutoReplyRule, buildReplySentLog, resetSeq } from '../fixtures/factories';
import { expectToast } from '../helpers/assertions';
import { AutoReplyPage } from '../pages/auto-reply.page';

test.describe('自動リプライ', () => {
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
      const autoReply = new AutoReplyPage(page);
      await autoReply.goto();
      await expect(autoReply.emptyState).toBeVisible();
    });

    test('既存ルールが表示される', async ({ page, apiMock }) => {
      const rule = buildAutoReplyRule({
        enabled: true,
        targetTweetText: 'テスト対象ポスト',
        triggerKeywords: ['AI', 'LLM'],
      });
      await apiMock.mockRPC('AutoReplyService', 'ListAutoReplyRules', { rules: [rule] });

      const autoReply = new AutoReplyPage(page);
      await autoReply.goto();
      await expect(page.getByText('ルール 1')).toBeVisible();
      await expect(page.getByText('AI')).toBeVisible();
      await expect(page.getByText('LLM')).toBeVisible();
    });
  });

  test.describe('トグル即時保存', () => {
    test('トグルOFF → APIが呼ばれ、リロード後もOFFのまま', async ({ page, apiMock }) => {
      const rule = buildAutoReplyRule({ enabled: true });
      await apiMock.mockRPC('AutoReplyService', 'ListAutoReplyRules', { rules: [rule] });

      const updatedRule = { ...rule, enabled: false };
      await apiMock.mockRPC('AutoReplyService', 'UpdateAutoReplyRule', { rule: updatedRule });

      const autoReply = new AutoReplyPage(page);
      await autoReply.goto();

      // APIリクエスト監視 + トグルクリックを並行
      const [req] = await Promise.all([
        page.waitForRequest((r) => r.url().includes('AutoReplyService/UpdateAutoReplyRule')),
        autoReply.toggleButton(0).locator('button').click(),
      ]);

      const body = JSON.parse(req.postData() ?? '{}');
      // proto3 JSON は false（デフォルト値）を省略するため、undefined または false を許容
      expect(body.enabled ?? false).toBe(false);

      // リロード後もOFFのまま（ダッシュボード経由で再遷移して auth state を再確立）
      await apiMock.clearMock('AutoReplyService', 'ListAutoReplyRules');
      await apiMock.mockRPC('AutoReplyService', 'ListAutoReplyRules', { rules: [updatedRule] });
      await page.goto('/dashboard');
      await page.waitForLoadState('networkidle');
      await page.locator('a[href="/auto-reply"]').click();
      await page.getByRole('heading', { name: '自動リプライ' }).waitFor({ timeout: 10_000 });
      await expect(page.getByText('ルール 1')).toBeVisible();
    });

    test('トグルON → APIが呼ばれる', async ({ page, apiMock }) => {
      const rule = buildAutoReplyRule({ enabled: false });
      await apiMock.mockRPC('AutoReplyService', 'ListAutoReplyRules', { rules: [rule] });

      const updatedRule = { ...rule, enabled: true };
      await apiMock.mockRPC('AutoReplyService', 'UpdateAutoReplyRule', { rule: updatedRule });

      const autoReply = new AutoReplyPage(page);
      await autoReply.goto();

      const [req] = await Promise.all([
        page.waitForRequest((r) => r.url().includes('AutoReplyService/UpdateAutoReplyRule')),
        autoReply.toggleButton(0).locator('button').click(),
      ]);

      const body = JSON.parse(req.postData() ?? '{}');
      expect(body.enabled).toBe(true);
    });

    test('トグル変更中はdisabledになり連打を防止', async ({ page, apiMock }) => {
      const rule = buildAutoReplyRule({ enabled: true });
      await apiMock.mockRPC('AutoReplyService', 'ListAutoReplyRules', { rules: [rule] });

      // API応答を遅延させる
      await apiMock.clearMock('AutoReplyService', 'UpdateAutoReplyRule');
      await page.route('**/trendbird.v1.AutoReplyService/UpdateAutoReplyRule', async (route) => {
        await new Promise((r) => setTimeout(r, 2000));
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ rule: { ...rule, enabled: false } }),
        });
      });

      const autoReply = new AutoReplyPage(page);
      await autoReply.goto();

      // トグルをクリック
      await autoReply.toggleButton(0).locator('button').click();

      // API応答待ち中にToggleボタンがdisabledになる
      await expect(autoReply.toggleButton(0).locator('button')).toBeDisabled();
    });

    test('トグルAPI失敗時にロールバック＋toast表示', async ({ page, apiMock }) => {
      const rule = buildAutoReplyRule({ enabled: true });
      await apiMock.mockRPC('AutoReplyService', 'ListAutoReplyRules', { rules: [rule] });
      await apiMock.mockRPCError('AutoReplyService', 'UpdateAutoReplyRule', 'internal', 'server error');

      const autoReply = new AutoReplyPage(page);
      await autoReply.goto();

      await autoReply.toggleButton(0).locator('button').click();

      await expectToast(page, '有効/無効の切り替えに失敗しました');
    });

    test('新規ルールのトグルはAPI呼び出しなし', async ({ page }) => {
      const autoReply = new AutoReplyPage(page);
      await autoReply.goto();

      await autoReply.addRuleButton.click();
      await expect(page.getByText('新しいルール')).toBeVisible();

      let apiCalled = false;
      page.on('request', (req) => {
        if (req.url().includes('AutoReplyService/UpdateAutoReplyRule')) {
          apiCalled = true;
        }
      });

      // 新規ルールのトグルをクリック
      await autoReply.toggleButton(0).locator('button').click();

      await page.waitForTimeout(500);
      expect(apiCalled).toBe(false);
    });
  });

  test.describe('ルール作成', () => {
    test('ポストURL＋キーワード＋テンプレートを入力して保存', async ({ page, apiMock }) => {
      const newRule = buildAutoReplyRule({
        targetTweetId: '1234567890',
        targetTweetText: '',
        triggerKeywords: ['テスト'],
        replyTemplate: 'ありがとうございます！',
      });
      await apiMock.mockRPC('AutoReplyService', 'CreateAutoReplyRule', { rule: newRule });

      const autoReply = new AutoReplyPage(page);
      await autoReply.goto();

      await autoReply.addRuleButton.click();

      await autoReply.tweetIdInput.fill('1234567890');

      await autoReply.keywordInput.fill('テスト');
      await autoReply.keywordInput.press('Enter');
      await expect(page.getByText('テスト')).toBeVisible();

      await autoReply.templateInput.fill('ありがとうございます！');

      await autoReply.saveButton('作成する').click();
      await expectToast(page, 'ルールを作成しました');
    });

    test('ポストURLが空で保存ボタン押下 → toast表示', async ({ page }) => {
      const autoReply = new AutoReplyPage(page);
      await autoReply.goto();

      await autoReply.addRuleButton.click();
      await autoReply.saveButton('作成する').click();

      await expectToast(page, '監視対象のポストURLまたはIDを入力してください');
    });
  });

  test.describe('ルール削除', () => {
    test('既存ルールを削除できる', async ({ page, apiMock }) => {
      const rule = buildAutoReplyRule();
      await apiMock.mockRPC('AutoReplyService', 'ListAutoReplyRules', { rules: [rule] });
      await apiMock.mockRPC('AutoReplyService', 'DeleteAutoReplyRule', {});

      const autoReply = new AutoReplyPage(page);
      await autoReply.goto();
      await expect(page.getByText('ルール 1')).toBeVisible();

      await autoReply.deleteButton.click();
      await expectToast(page, 'ルールを削除しました');
    });
  });

  test.describe('保存ボタン', () => {
    test('トグル無効時でも保存ボタンはクリック可能', async ({ page, apiMock }) => {
      const rule = buildAutoReplyRule({ enabled: true });
      await apiMock.mockRPC('AutoReplyService', 'ListAutoReplyRules', { rules: [rule] });
      await apiMock.mockRPC('AutoReplyService', 'UpdateAutoReplyRule', { rule: { ...rule, enabled: false } });

      const autoReply = new AutoReplyPage(page);
      await autoReply.goto();

      // トグルをOFFに
      await autoReply.toggleButton(0).locator('button').click();

      await expect(autoReply.saveButton('保存する')).toBeEnabled();
    });
  });

  test.describe('送信履歴', () => {
    test('履歴タブでログが表示される', async ({ page, apiMock }) => {
      const log = buildReplySentLog({ triggerKeyword: 'AI', replyText: 'テスト返信です' });
      await apiMock.mockRPC('AutoReplyService', 'GetReplySentLogs', { logs: [log] });

      const autoReply = new AutoReplyPage(page);
      await autoReply.goto();

      await autoReply.historyTab.click();
      await expect(page.getByText('AI')).toBeVisible();
      await expect(page.getByText('テスト返信です')).toBeVisible();
    });

    test('履歴がない場合に空状態が表示される', async ({ page, apiMock }) => {
      await apiMock.mockRPC('AutoReplyService', 'GetReplySentLogs', { logs: [] });

      const autoReply = new AutoReplyPage(page);
      await autoReply.goto();

      await autoReply.historyTab.click();
      await expect(page.getByText('まだ送信履歴はありません')).toBeVisible();
    });
  });

  // ─── テンプレート文字数制限 ────────────────────────────────

  test.describe('テンプレート文字数制限', () => {
    test('280文字を超える入力で文字カウンターが上限超過を表示', async ({ page }) => {
      const autoReply = new AutoReplyPage(page);
      await autoReply.goto();

      // 新しいルールを追加
      await autoReply.addRuleButton.click();

      // 280文字超の文字列を入力
      const longText = 'あ'.repeat(290);
      await autoReply.templateInput.fill(longText);

      // 文字カウンターが 290/280 と表示される（上限超過）
      await expect(page.getByText('290/280')).toBeVisible();
    });
  });

  // ─── キーワードバリデーション ──────────────────────────────

  test.describe('キーワードバリデーション', () => {
    test('重複キーワードで toast 表示', async ({ page }) => {
      const autoReply = new AutoReplyPage(page);
      await autoReply.goto();

      // 新しいルールを追加
      await autoReply.addRuleButton.click();

      // キーワード「AI」を追加
      await autoReply.keywordInput.fill('AI');
      await autoReply.keywordInput.press('Enter');
      await expect(page.getByText('AI')).toBeVisible();

      // 同じキーワード「AI」を再追加 → 重複 toast
      await autoReply.keywordInput.fill('AI');
      await autoReply.keywordInput.press('Enter');
      await expectToast(page, '同じキーワードが既に追加されています');
    });

    test('10個のキーワード上限で toast 表示', async ({ page }) => {
      const autoReply = new AutoReplyPage(page);
      await autoReply.goto();

      // 新しいルールを追加
      await autoReply.addRuleButton.click();

      // 10個のキーワードを追加
      for (let i = 0; i < 10; i++) {
        await autoReply.keywordInput.fill(`keyword${i}`);
        await autoReply.keywordInput.press('Enter');
      }

      // 11個目を追加 → 上限 toast
      await autoReply.keywordInput.fill('keyword10');
      await autoReply.keywordInput.press('Enter');
      await expectToast(page, 'キーワードは最大10個までです');
    });
  });
});
