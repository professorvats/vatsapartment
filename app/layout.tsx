import type { Metadata } from 'next';
import Script from 'next/script';
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
      <head>
        <Script id="gtm-head" strategy="afterInteractive">
          {`(function(w,d,s,l,i){w[l]=w[l]||[];w[l].push({'gtm.start':
new Date().getTime(),event:'gtm.js'});var f=d.getElementsByTagName(s)[0],
j=d.createElement(s),dl=l!='dataLayer'?'&l='+l:'';j.async=true;j.src=
'https://www.googletagmanager.com/gtm.js?id='+i+dl;f.parentNode.insertBefore(j,f);
})(window,document,'script','dataLayer','GTM-P6TDL438');`}
        </Script>
      </head>
      <body className="min-h-screen flex flex-col bg-surface text-on-background font-sans">
        <noscript>
          <iframe
            src="https://www.googletagmanager.com/ns.html?id=GTM-P6TDL438"
            height="0"
            width="0"
            style={{ display: 'none', visibility: 'hidden' }}
          />
        </noscript>
        <Navigation />
        <main className="flex-1 pt-16 md:pt-20">{children}</main>
        <FooterWithWhatsApp />
      </body>
    </html>
  );
}
