import { test, expect, authenticateContext } from '../fixtures/test-base';
import { resetSeq } from '../fixtures/factories';
import { TutorialPage } from '../pages/tutorial.page';
import { create, toJson } from '@bufbuild/protobuf';
import {
  UserSchema,
  GetCurrentUserResponseSchema,
} from '../../src/gen/trendbird/v1/auth_pb';
import type { ApiMock } from '../fixtures/api-mock';
import type { Page } from '@playwright/test';

// ── チュートリアルモード共通セットアップ ──

async function setupTutorialDefaults(apiMock: ApiMock) {
  await apiMock.setupDefaults();

  // GetCurrentUser を tutorialPending: true でオーバーライド
  await apiMock.clearMock('AuthService', 'GetCurrentUser');
  await apiMock.mockRPC('AuthService', 'GetCurrentUser',
    toJson(GetCurrentUserResponseSchema, create(GetCurrentUserResponseSchema, {
      user: create(UserSchema, {
        id: 'tutorial-user',
        name: 'Tutorial User',
        email: 'tutorial@example.com',
        image: '',
        twitterHandle: 'tutorialuser',

        createdAt: '2026-01-01T00:00:00Z',
      }),
      tutorialPending: true,
    })));

  // CompleteTutorial RPC（finish フェーズで呼ばれる）
  await apiMock.mockRPC('AuthService', 'CompleteTutorial', {});
}

type TutorialPhase = 'welcome' | 'topic-setup' | 'dashboard' | 'detail' | 'finish';

async function setTutorialPhase(page: Page, phase: TutorialPhase) {
  await page.addInitScript((p: string) => {
    sessionStorage.setItem('tb_tutorial_pending', '1');
    sessionStorage.setItem('tb_tutorial_state', JSON.stringify({
      phase: p,
      startedAt: Date.now(),
      phaseEnteredAt: Date.now(),
    }));
  }, phase);
}

// ═══════════════════════════════════════════════════════════════
// Group 1: ゲートロジック
// ═══════════════════════════════════════════════════════════════

test.describe('チュートリアル: ゲートロジック', () => {
  test.beforeEach(async ({ context }) => {
    resetSeq();
    await authenticateContext(context);
  });

  test('非新規ユーザーにはチュートリアルが表示されない', async ({ page, apiMock }) => {
    await apiMock.setupDefaults(); // tutorialPending: false (default)
    await page.addInitScript(() => {
      sessionStorage.removeItem('tb_tutorial_pending');
    });

    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    const tutorial = new TutorialPage(page);
    // 2秒待ってポップオーバーが出ないことを確認
    await expect(tutorial.popover).not.toBeVisible({ timeout: 2_000 });
  });

  test('sessionStorage pending キーだけでチュートリアルが起動する', async ({ page, apiMock }) => {
    await apiMock.setupDefaults(); // tutorialPending: false (API側)
    // sessionStorage に pending を設定（removeItem しない）
    await page.addInitScript(() => {
      sessionStorage.setItem('tb_tutorial_pending', '1');
    });

    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    const tutorial = new TutorialPage(page);
    await tutorial.waitForPopover('TrendBird へようこそ！');
  });

  test('stale な完了フラグは tutorialPending=true で自動クリアされる', async ({ page, apiMock }) => {
    await setupTutorialDefaults(apiMock);
    // stale な完了フラグを事前設定
    await page.addInitScript(() => {
      localStorage.setItem('tb_tutorial_completed_tutorial-user', '1');
    });

    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    const tutorial = new TutorialPage(page);
    // stale フラグがクリアされ、チュートリアルが起動する
    await tutorial.waitForPopover('TrendBird へようこそ！');
  });
});

// ═══════════════════════════════════════════════════════════════
// Group 2: Welcome フェーズ
// ═══════════════════════════════════════════════════════════════

test.describe('チュートリアル: Welcome', () => {
  test.beforeEach(async ({ context, apiMock }) => {
    resetSeq();
    await authenticateContext(context);
    await setupTutorialDefaults(apiMock);
  });

  test('Welcome ポップオーバーが表示され、次へで /topics に遷移する', async ({ page }) => {
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    const tutorial = new TutorialPage(page);
    await tutorial.waitForPopover('TrendBird へようこそ！');
    await expect(tutorial.popoverDescription).toContainText('X(Twitter)のトレンドを自動モニタリング');
    await expect(tutorial.nextButton.or(tutorial.doneButton).first()).toBeVisible();

    await tutorial.clickNext();
    await page.waitForURL('**/topics');
  });
});

// ═══════════════════════════════════════════════════════════════
// Group 3: Topic Setup フェーズ
// ═══════════════════════════════════════════════════════════════

test.describe('チュートリアル: Topic Setup', () => {
  test.beforeEach(async ({ context, page, apiMock }) => {
    resetSeq();
    await authenticateContext(context);
    await setupTutorialDefaults(apiMock);
    await setTutorialPhase(page, 'topic-setup');
  });

  test('ジャンル選択 → トピック追加 → サイドバークリックで dashboard に遷移', async ({ page }) => {
    await page.goto('/topics');
    await page.waitForLoadState('networkidle');

    const tutorial = new TutorialPage(page);

    // Step 1: 「ジャンルを選ぼう」ポップオーバー → CTA クリック
    await tutorial.waitForPopover('ジャンルを選ぼう');
    await tutorial.genreSelectCta.click();

    // Step 2: ジャンルモーダル内「テクノロジーを選択」→ ジャンル行クリック
    await tutorial.waitForPopover('テクノロジーを選択');
    await tutorial.tutorialGenreRow.click();

    // Step 3: 「トピックを追加してみよう」→ サジェスト行クリック
    await tutorial.waitForPopover('トピックを追加してみよう');
    await tutorial.tutorialSuggestRow.click();

    // Step 4: サイドバーのダッシュボードアイコン → クリックで遷移
    await tutorial.waitForPopover('ダッシュボードを見てみよう');
    await tutorial.sidebarDashboard.click();

    await page.waitForURL('**/dashboard');
  });
});

// ═══════════════════════════════════════════════════════════════
// Group 4: Dashboard フェーズ
// ═══════════════════════════════════════════════════════════════

test.describe('チュートリアル: Dashboard', () => {
  test.beforeEach(async ({ context, page, apiMock }) => {
    resetSeq();
    await authenticateContext(context);
    await setupTutorialDefaults(apiMock);
    await setTutorialPhase(page, 'dashboard');
  });

  test('ダミーデータ表示 + 4ステップ + カードクリックで detail に遷移', async ({ page }) => {
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    const tutorial = new TutorialPage(page);

    // ダミーデータ検証: ステータスバー
    await expect(page.getByText('2件 盛り上がり中')).toBeVisible();
    await expect(page.getByText('3件 上昇中')).toBeVisible();
    await expect(page.getByText('8件 安定')).toBeVisible();

    // ダミーデータ検証: トピックカード
    await expect(page.getByText('AI・機械学習')).toBeVisible();
    await expect(page.getByText('4.2', { exact: true })).toBeVisible();

    // API データが表示されていないこと
    await expect(page.getByText('Default Topic 1')).not.toBeVisible();

    // Step 1: モニタリング状況
    await tutorial.waitForPopover('モニタリング状況');
    await tutorial.clickNext();

    // Step 2: ジャンルフィルター
    await tutorial.waitForPopover('ジャンルフィルター');
    await tutorial.clickNext();

    // Step 3: トピックカード
    await tutorial.waitForPopover('トピックカード');
    await tutorial.clickNext();

    // Step 4: カードクリック（ボタンなし → カード自体をクリック）
    await tutorial.waitForPopover('カードをクリックしてみよう');
    await tutorial.topicCard.click();

    await page.waitForURL('**/dashboard/tutorial-dummy');
  });
});

// ═══════════════════════════════════════════════════════════════
// Group 5: Detail フェーズ
// ═══════════════════════════════════════════════════════════════

test.describe('チュートリアル: Detail', () => {
  test.beforeEach(async ({ context, apiMock }) => {
    resetSeq();
    await authenticateContext(context);
    await setupTutorialDefaults(apiMock);
  });

  test('ダミー詳細データ表示 + 4ステップ + サイドバークリックで finish に遷移', async ({ page }) => {
    await setTutorialPhase(page, 'detail');

    await page.goto('/dashboard/tutorial-dummy');
    await page.waitForLoadState('networkidle');

    const tutorial = new TutorialPage(page);

    // ダミーデータ検証
    await expect(page.getByText('AI・機械学習').first()).toBeVisible();
    await expect(page.getByText('4.2').first()).toBeVisible();
    await expect(page.getByText('OpenAIの新モデル発表により').first()).toBeVisible();

    // Step 1: 盛り上がり度
    await tutorial.waitForPopover('盛り上がり度');
    await tutorial.clickNext();

    // Step 2: なぜ今バズっているのか
    await tutorial.waitForPopover('なぜ今バズっているのか');
    await tutorial.clickNext();

    // Step 3: AI投稿文生成
    await tutorial.waitForPopover('AI投稿文生成');
    await tutorial.clickNext();

    // Step 4: サイドバー「トピック」クリック（ボタンなし）
    await tutorial.waitForPopover('トピック選択へ進もう');
    await tutorial.sidebarTopics.click();

    await page.waitForURL('**/topics');
  });

  test('チュートリアル外で /dashboard/tutorial-dummy にアクセスすると /dashboard にリダイレクト', async ({ page, apiMock }) => {
    // デフォルトモック（チュートリアルなし）
    await apiMock.clearMock('AuthService', 'GetCurrentUser');
    await apiMock.mockRPC('AuthService', 'GetCurrentUser',
      toJson(GetCurrentUserResponseSchema, create(GetCurrentUserResponseSchema, {
        user: create(UserSchema, {
          id: 'normal-user',
          name: 'Normal User',
          email: 'normal@example.com',
          image: '',
          twitterHandle: 'normaluser',
  
          createdAt: '2026-01-01T00:00:00Z',
        }),
        tutorialPending: false,
      })));
    await page.addInitScript(() => {
      sessionStorage.removeItem('tb_tutorial_pending');
      sessionStorage.removeItem('tb_tutorial_state');
    });

    await page.goto('/dashboard/tutorial-dummy');
    await page.waitForURL('**/dashboard', { timeout: 10_000 });
    expect(page.url()).not.toContain('tutorial-dummy');
  });
});

// ═══════════════════════════════════════════════════════════════
// Group 6: Finish フェーズ
// ═══════════════════════════════════════════════════════════════

test.describe('チュートリアル: Finish', () => {
  test.beforeEach(async ({ context, page, apiMock }) => {
    resetSeq();
    await authenticateContext(context);
    await setupTutorialDefaults(apiMock);
    await setTutorialPhase(page, 'finish');
  });

  test('完了ポップオーバー表示 → CompleteTutorial RPC → /dashboard に遷移', async ({ page, apiMock }) => {
    test.setTimeout(60_000);
    const completeTutorialPromise = page.waitForRequest(
      (req) => req.url().includes('CompleteTutorial'),
    );

    // finish フェーズでは isTutorialTopicSetup=false のため /topics は
    // 通常データで描画される。デフォルトモックのトピックデータがあると
    // クライアントサイドリダイレクトが発生するため、空データで上書きする
    await apiMock.clearMock('TopicService', 'ListTopics');
    await apiMock.mockRPC('TopicService', 'ListTopics', { topics: [] });
    await apiMock.clearMock('TopicService', 'ListUserGenres');
    await apiMock.mockRPC('TopicService', 'ListUserGenres', { genres: [] });

    await page.goto('/topics');
    await page.waitForLoadState('networkidle');

    const tutorial = new TutorialPage(page);
    await tutorial.waitForPopover('セットアップ完了！');
    await expect(tutorial.popoverDescription).toContainText('さっそくダッシュボードでトレンドをチェックしましょう');

    // 「ダッシュボードへ」ボタンをクリック（next/done 両クラスに対応）
    await tutorial.clickNext();

    // CompleteTutorial RPC が呼ばれたことを確認
    await completeTutorialPromise;

    // /dashboard に遷移
    await page.waitForURL('**/dashboard');
  });
});

// ═══════════════════════════════════════════════════════════════
// Group 7: フル統合テスト
// ═══════════════════════════════════════════════════════════════

test.describe('チュートリアル: フル統合', () => {
  test('全フェーズ一気通貫: welcome → topic-setup → dashboard → detail → finish → /dashboard', async ({ context, page, apiMock }) => {
    test.setTimeout(60_000);
    resetSeq();
    await authenticateContext(context);
    await setupTutorialDefaults(apiMock);
    // Disable driver.js animations to avoid timing races between
    // popover visibility (~200ms) and onHighlighted/click handlers (~400ms)
    await page.emulateMedia({ reducedMotion: 'reduce' });

    const tutorial = new TutorialPage(page);

    // ── Phase 1: Welcome ──
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    await tutorial.waitForPopover('TrendBird へようこそ！');
    await tutorial.clickNext();
    await page.waitForURL('**/topics');

    // ── Phase 2: Topic Setup ──
    await tutorial.waitForPopover('ジャンルを選ぼう');
    await tutorial.genreSelectCta.click();

    await tutorial.waitForPopover('テクノロジーを選択');
    await tutorial.tutorialGenreRow.click();

    await tutorial.waitForPopover('トピックを追加してみよう');
    await tutorial.tutorialSuggestRow.click();

    await tutorial.waitForPopover('ダッシュボードを見てみよう');
    await tutorial.sidebarDashboard.click();

    await page.waitForURL('**/dashboard');

    // Wait for TutorialFloatingCta cleanup to advance phase to 'dashboard'
    await page.waitForFunction(
      () => {
        const raw = sessionStorage.getItem('tb_tutorial_state');
        if (!raw) return false;
        try { return JSON.parse(raw).phase === 'dashboard'; } catch { return false; }
      },
      { timeout: 10_000 },
    );

    // ── Phase 3: Dashboard ──
    await tutorial.waitForPopover('モニタリング状況');
    await tutorial.clickNext();

    await tutorial.waitForPopover('ジャンルフィルター');
    await tutorial.clickNext();

    await tutorial.waitForPopover('トピックカード');
    await tutorial.clickNext();

    await tutorial.waitForPopover('カードをクリックしてみよう');
    await tutorial.topicCard.click();

    await page.waitForURL('**/dashboard/tutorial-dummy');

    // ── Phase 4: Detail ──
    await tutorial.waitForPopover('盛り上がり度');
    await tutorial.clickNext();

    await tutorial.waitForPopover('なぜ今バズっているのか');
    await tutorial.clickNext();

    await tutorial.waitForPopover('AI投稿文生成');
    await tutorial.clickNext();

    await tutorial.waitForPopover('トピック選択へ進もう');
    // Clear topics to prevent /topics → /dashboard redirect on finish phase
    await apiMock.clearMock('TopicService', 'ListTopics');
    await apiMock.mockRPC('TopicService', 'ListTopics', { topics: [] });
    await apiMock.clearMock('TopicService', 'ListUserGenres');
    await apiMock.mockRPC('TopicService', 'ListUserGenres', { genres: [] });
    await tutorial.sidebarTopics.click();

    await page.waitForURL('**/topics');

    // Wait for detail phase cleanup to advance phase to 'finish'
    await page.waitForFunction(
      () => {
        const raw = sessionStorage.getItem('tb_tutorial_state');
        if (!raw) return false;
        try { return JSON.parse(raw).phase === 'finish'; } catch { return false; }
      },
      { timeout: 10_000 },
    );

    // ── Phase 5: Finish ──
    await tutorial.waitForPopover('セットアップ完了！');
    await tutorial.clickNext();

    await page.waitForURL('**/dashboard');
  });
});
