import type { DriveStep } from 'driver.js';

/** Phase C: Dashboard — explain status bar, genre tabs, topic card */
export function getDashboardSteps(): DriveStep[] {
  return [
    {
      element: '[data-tutorial="status-bar"]',
      popover: {
        title: 'モニタリング状況',
        description:
          'スパイク・上昇中・安定の件数をリアルタイムで確認できます。',
        side: 'bottom',
        align: 'center',
      },
    },
    {
      element: '[data-tutorial="genre-tabs"]',
      popover: {
        title: 'ジャンルフィルター',
        description: 'ジャンルごとにトピックを絞り込めます。',
        side: 'bottom',
        align: 'start',
      },
    },
    {
      element: '[data-tutorial="topic-card"]',
      popover: {
        title: 'トピックカード',
        description:
          '盛り上がり度・変化率・ボリューム推移を確認できます。',
        side: 'right',
        align: 'start',
      },
    },
    {
      element: '[data-tutorial="topic-card"]',
      popover: {
        title: 'カードをクリックしてみよう',
        description:
          'クリックすると詳細画面に移動します。実際にクリックしてみましょう！',
        side: 'right',
        align: 'start',
        showButtons: [],
      },
    },
  ];
}
