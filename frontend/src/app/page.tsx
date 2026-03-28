import { AmbientBackground, MarketingHeader, Footer } from '@/components/layout';
import {
  HeroSection,
  SocialProofBar,
  FeaturesSection,
  HowItWorksSection,
  FaqSection,
  FinalCtaSection,
} from '@/components/landing';

export default function LandingPage() {
  return (
    <div className="relative flex min-h-screen flex-col">
      <AmbientBackground />
      <MarketingHeader />

      <main className="relative z-10 flex-1 pt-16">
        <HeroSection />
        <SocialProofBar />
        <FeaturesSection />
        <HowItWorksSection />
        <FaqSection />
        <FinalCtaSection />
      </main>

      <Footer />
    </div>
  );
}
