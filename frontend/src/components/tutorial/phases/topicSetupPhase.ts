import type { DriveStep } from 'driver.js';

/** Phase B-1: Highlight the "ジャンルを選ぶ" CTA button */
export function getGenreCTAStep(): DriveStep[] {
  return [
    {
      element: '[data-tutorial="genre-select-cta"]',
      popover: {
        title: 'ジャンルを選ぼう',
        description:
          '「ジャンルを選ぶ」ボタンをクリックしてみましょう！',
        side: 'bottom',
        align: 'center',
      },
    },
  ];
}

/** Phase B-2: Highlight the "テクノロジー" row in the genre modal */
export function getGenreRowStep(): DriveStep[] {
  return [
    {
      element: '[data-tutorial="tutorial-genre-row"]',
      popover: {
        title: 'テクノロジーを選択',
        description:
          '「テクノロジー」をクリックして、ジャンルを追加しましょう！',
        side: 'bottom',
        align: 'center',
      },
    },
  ];
}

/** Phase B-3: Highlight the first suggested topic row */
export function getSuggestRowStep(): DriveStep[] {
  return [
    {
      element: '[data-tutorial="tutorial-suggest-row"]',
      popover: {
        title: 'トピックを追加してみよう',
        description:
          'クリックしてトピックを追加してみましょう！',
        side: 'bottom',
        align: 'start',
      },
    },
  ];
}
