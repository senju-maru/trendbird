import { AmbientBackground, MarketingHeader } from '@/components/layout';

export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div
      className="relative flex min-h-screen flex-col"
      style={{ backgroundColor: '#e4eaf1', color: '#2c3e50' }}
    >
      <AmbientBackground />
      <MarketingHeader />

      <main className="relative z-10 flex flex-1 items-center justify-center px-4 pt-16">
        <div className="w-full max-w-md">
          {children}
        </div>
      </main>
    </div>
  );
}
