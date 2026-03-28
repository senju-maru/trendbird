/**
 * Proto Schema ベースの API モッククラス。
 * setupDefaults() は全レスポンスを Response Schema の create() + toJson() で構築する。
 * proto フィールド名・構造の変更は TypeScript コンパイルエラーとして即座に検知される。
 */
import { Page } from '@playwright/test';
import { create, toJson } from '@bufbuild/protobuf';
import {
  UserSchema,
  GetCurrentUserResponseSchema,
} from '../../src/gen/trendbird/v1/auth_pb';
import {
  TopicSchema,
  TopicStatus,
  GenreSchema,
  ListTopicsResponseSchema,
  ListGenresResponseSchema,
  ListUserGenresResponseSchema,
} from '../../src/gen/trendbird/v1/topic_pb';
import {
  PostStatsSchema,
  ListDraftsResponseSchema,
  ListPostHistoryResponseSchema,
  GetPostStatsResponseSchema,
} from '../../src/gen/trendbird/v1/post_pb';
import {
  DashboardStatsSchema,
  GetActivitiesResponseSchema,
  GetStatsResponseSchema,
} from '../../src/gen/trendbird/v1/dashboard_pb';
import {
  ListNotificationsResponseSchema,
} from '../../src/gen/trendbird/v1/notification_pb';
import {
  NotificationSettingsSchema,
  GetNotificationSettingsResponseSchema,
} from '../../src/gen/trendbird/v1/settings_pb';
import {
  TwitterConnectionInfoSchema,
  TwitterConnectionStatus,
  GetConnectionInfoResponseSchema,
} from '../../src/gen/trendbird/v1/twitter_pb';
import {
  ListAutoDMRulesResponseSchema,
  GetDMSentLogsResponseSchema,
} from '../../src/gen/trendbird/v1/auto_dm_pb';

/**
 * setupDefaults 用の固定初期化データ。
 * ファクトリのシーケンスを消費しないよう、固定値で定義する。
 * create() に渡すことで proto Schema による型チェックが行われる。
 */
const DEFAULT_USER_INIT = {
  id: 'default-user',
  name: 'Default User',
  email: 'default@example.com',
  image: '',
  twitterHandle: 'defaultuser',
  createdAt: '2026-01-01T00:00:00Z',
};

const DEFAULT_TOPIC_INITS = [
  {
    id: 'default-topic-1',
    name: 'Default Topic 1',
    keywords: ['keyword-1'],
    genre: 'technology',
    status: TopicStatus.SPIKE,
    changePercent: 0.5,
    zScore: 3.5,
    currentVolume: 500,
    baselineVolume: 100,
    sparklineData: [],
    weeklySparklineData: [],
    spikeHistory: [],
    context: '最新のAI技術が注目されています',
    contextSummary: '最新のAI技術が注目されています',
    spikeStartedAt: '2026-01-01T00:00:00Z',
    notificationEnabled: true,
    createdAt: '2026-01-01T00:00:00Z',
  },
  {
    id: 'default-topic-2',
    name: 'Default Topic 2',
    keywords: ['keyword-2'],
    genre: 'technology',
    status: TopicStatus.STABLE,
    changePercent: 0.05,
    zScore: 1.2,
    currentVolume: 100,
    baselineVolume: 80,
    sparklineData: [],
    weeklySparklineData: [],
    spikeHistory: [],
    notificationEnabled: true,
    createdAt: '2026-01-01T00:00:00Z',
  },
];

const DEFAULT_GENRE_INITS = [
  { id: 'default-genre-1', slug: 'technology', label: 'Technology', description: '', sortOrder: 0 },
  { id: 'default-genre-2', slug: 'business', label: 'Business', description: '', sortOrder: 1 },
];

export class ApiMock {
  constructor(private page: Page) {}

  /** デフォルトハンドラを一括登録（Happy Path）— proto Response Schema ベースで型安全 */
  async setupDefaults(userOverrides?: Record<string, unknown>) {
    const user = create(UserSchema, { ...DEFAULT_USER_INIT, ...userOverrides });
    const topics = DEFAULT_TOPIC_INITS.map((init) => create(TopicSchema, init));
    const genres = DEFAULT_GENRE_INITS.map((init) => create(GenreSchema, init));

    await Promise.all([
      // AuthService
      this.mockRPC('AuthService', 'GetCurrentUser',
        toJson(GetCurrentUserResponseSchema,
          create(GetCurrentUserResponseSchema, { user, tutorialPending: false }))),

      // TopicService
      this.mockRPC('TopicService', 'ListTopics',
        toJson(ListTopicsResponseSchema,
          create(ListTopicsResponseSchema, { topics }))),
      this.mockRPC('TopicService', 'ListGenres',
        toJson(ListGenresResponseSchema,
          create(ListGenresResponseSchema, { genres }))),
      this.mockRPC('TopicService', 'ListUserGenres',
        toJson(ListUserGenresResponseSchema,
          create(ListUserGenresResponseSchema, { genres: ['technology'] }))),

      // PostService
      this.mockRPC('PostService', 'ListDrafts',
        toJson(ListDraftsResponseSchema,
          create(ListDraftsResponseSchema, {}))),
      this.mockRPC('PostService', 'ListPostHistory',
        toJson(ListPostHistoryResponseSchema,
          create(ListPostHistoryResponseSchema, {}))),
      this.mockRPC('PostService', 'GetPostStats',
        toJson(GetPostStatsResponseSchema,
          create(GetPostStatsResponseSchema, {
            stats: create(PostStatsSchema, {}),
          }))),

      // DashboardService
      this.mockRPC('DashboardService', 'GetActivities',
        toJson(GetActivitiesResponseSchema,
          create(GetActivitiesResponseSchema, {}))),
      this.mockRPC('DashboardService', 'GetStats',
        toJson(GetStatsResponseSchema,
          create(GetStatsResponseSchema, {
            stats: create(DashboardStatsSchema, {}),
          }))),

      // NotificationService
      this.mockRPC('NotificationService', 'ListNotifications',
        toJson(ListNotificationsResponseSchema,
          create(ListNotificationsResponseSchema, {}))),

      // SettingsService
      this.mockRPC('SettingsService', 'GetNotificationSettings',
        toJson(GetNotificationSettingsResponseSchema,
          create(GetNotificationSettingsResponseSchema, {
            settings: create(NotificationSettingsSchema, {
              spikeEnabled: true,
              risingEnabled: true,
            }),
          }))),

      // TwitterService
      this.mockRPC('TwitterService', 'GetConnectionInfo',
        toJson(GetConnectionInfoResponseSchema,
          create(GetConnectionInfoResponseSchema, {
            info: create(TwitterConnectionInfoSchema, {
              status: TwitterConnectionStatus.DISCONNECTED,
            }),
          }))),

      // AutoDMService
      this.mockRPC('AutoDMService', 'ListAutoDMRules',
        toJson(ListAutoDMRulesResponseSchema,
          create(ListAutoDMRulesResponseSchema, {}))),
      this.mockRPC('AutoDMService', 'GetDMSentLogs',
        toJson(GetDMSentLogsResponseSchema,
          create(GetDMSentLogsResponseSchema, {}))),
    ]);
  }

  /** 個別 RPC モック */
  async mockRPC(
    service: string,
    method: string,
    body: unknown,
    status = 200,
    headers?: Record<string, string>,
  ) {
    await this.page.route(
      `**/trendbird.v1.${service}/${method}`,
      (route) => {
        route.fulfill({
          status,
          contentType: 'application/json',
          body: JSON.stringify(body),
          ...(headers ? { headers: { 'Content-Type': 'application/json', ...headers } } : {}),
        });
      },
    );
  }

  /** Connect RPC エラーレスポンス */
  async mockRPCError(
    service: string,
    method: string,
    code: string,
    message: string,
  ) {
    const statusMap: Record<string, number> = {
      not_found: 404,
      unauthenticated: 401,
      permission_denied: 403,
      resource_exhausted: 429,
      invalid_argument: 400,
      internal: 500,
      already_exists: 409,
      failed_precondition: 400,
    };
    await this.page.route(
      `**/trendbird.v1.${service}/${method}`,
      (route) => {
        route.fulfill({
          status: statusMap[code] ?? 500,
          contentType: 'application/json',
          body: JSON.stringify({ code, message }),
        });
      },
    );
  }

  /** 特定 RPC のオーバーライドをクリア */
  async clearMock(service: string, method: string) {
    await this.page.unroute(`**/trendbird.v1.${service}/${method}`);
  }
}
