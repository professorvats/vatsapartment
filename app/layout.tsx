import type { Metadata } from 'next';
import './globals.css';
import Navigation from '@/components/Navigation';
import FooterWithWhatsApp from '@/components/FooterWithWhatsApp';

export const metadata: Metadata = {
  title: {
    default: 'Vats Apartment | Premium PG near LPU University',
    template: '%s | Vats Apartment',
  },
  description: 'Premium fully furnished PG accommodation near LPU University, Jalandhar. Queen size bed, AC, personal kitchen, attached washroom, Smart TV, WiFi. Starting ₹9,000/month.',
  keywords: 'PG near LPU, hostel near LPU, rooms near LPU, paying guest LPU, Vats Apartment, PG accommodation Jalandhar',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en-IN" className="h-full antialiased">
      <body className="min-h-screen flex flex-col bg-surface text-on-background font-sans">
        <Navigation />
        <main className="flex-1 pt-16 md:pt-20">{children}</main>
        <FooterWithWhatsApp />
      </body>
    </html>
  );
}
