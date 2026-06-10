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

const GTM_ID = 'GTM-P6TDL438';
const GADS_ID = 'AW-18227145928';

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en-IN" className="h-full antialiased">
      <head>
        <script
          dangerouslySetInnerHTML={{
            __html: `window.dataLayer=window.dataLayer||[];window.dataLayer.push({'gtm.start':new Date().getTime(),event:'gtm.js'});`,
          }}
        />
        <Script id="gtm" strategy="afterInteractive">
          {`(function(w,d,s,l,i){w[l]=w[l]||[];w[l].push({'gtm.start':
new Date().getTime(),event:'gtm.js'});var f=d.getElementsByTagName(s)[0],
j=d.createElement(s),dl=l!='dataLayer'?'&l='+l:'';j.async=true;j.src=
'https://www.googletagmanager.com/gtm.js?id='+i+dl;f.parentNode.insertBefore(j,f);
})(window,document,'script','dataLayer','${GTM_ID}');`}
        </Script>
        <Script
          id="google-ads-gtag"
          strategy="afterInteractive"
          src={`https://www.googletagmanager.com/gtag/js?id=${GADS_ID}`}
        />
        <Script id="google-ads-config" strategy="afterInteractive">
          {`window.dataLayer=window.dataLayer||[];function gtag(){dataLayer.push(arguments);}gtag('js',new Date());gtag('config','${GADS_ID}');`}
        </Script>
      </head>
      <body className="min-h-screen flex flex-col bg-surface text-on-background font-sans">
        <noscript>
          <iframe
            src={`https://www.googletagmanager.com/ns.html?id=${GTM_ID}`}
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
