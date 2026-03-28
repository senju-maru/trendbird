import type { DriveStep } from 'driver.js';

/** Phase E: Detail — explain hero, context, AI generation */
export function getDetailSteps(): DriveStep[] {
  return [
    {
      element: '[data-tutorial="detail-hero"]',
      popover: {
        title: '盛り上がり度',
        description:
          'z-scoreで盛り上がり度をリアルタイムで確認。ボリュームの推移も一目で分かります',
        side: 'bottom',
        align: 'center',
      },
    },
    {
      element: '[data-tutorial="detail-context"]',
      popover: {
        title: 'なぜ今バズっているのか',
        description: 'AIがスパイクの理由を自動分析してお知らせします',
        side: 'bottom',
        align: 'center',
      },
    },
    {
      element: '[data-tutorial="detail-ai"]',
      popover: {
        title: 'AI投稿文生成',
        description:
          'バズに乗った投稿文をAIが自動生成します',
        side: 'left',
        align: 'start',
      },
    },
    {
      element: '[data-tutorial="sidebar-topics"]',
      popover: {
        title: 'トピック選択へ進もう',
        description:
          'サイドバーの「トピック」をクリックして、気になるトピックを登録しましょう！',
        side: 'right',
        align: 'center',
        showButtons: [],
      },
    },
  ];
}
