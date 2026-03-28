'use client';

import dynamic from 'next/dynamic';

const OnboardingTutorial = dynamic(
  () => import('./OnboardingTutorial').then((m) => ({ default: m.OnboardingTutorial })),
  { ssr: false },
);

export function OnboardingTutorialLoader() {
  return <OnboardingTutorial />;
}
