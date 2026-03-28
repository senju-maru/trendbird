import type { DriveStep } from 'driver.js';

/** Phase A: Welcome — element-less center popover */
export function getWelcomeSteps(): DriveStep[] {
  return [
    {
      popover: {
        title: 'TrendBird へようこそ！',
        description:
          'X(Twitter)のトレンドを自動モニタリングし、盛り上がりをリアルタイムでキャッチ。AI投稿文の生成も可能です。まずは監視したいトピックを設定しましょう！',
        side: 'over',
        align: 'center',
      },
    },
  ];
}
