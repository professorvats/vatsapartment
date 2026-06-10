'use client';

import { usePathname } from 'next/navigation';
import { Home, CalendarCheck, MapPin, Phone, LogIn, MessageCircle } from 'lucide-react';
import { useState } from 'react';
import Image from 'next/image';
import Link from 'next/link';

const NAV_LINKS = [
  { href: '/', label: 'Home' },
  { href: '/book-now', label: 'Book Now' },
  { href: '/referral', label: 'Refer & Earn' },
  { href: '/blog', label: 'Blog' },
  { href: '/room-showcase', label: '3D Tour' },
  { href: '/location', label: 'Location' },
  { href: '/contact-us', label: 'Contact' },
];

const MOBILE_TABS = [
  { href: '/', label: 'Home', icon: Home },
  { href: '/book-now', label: 'Book', icon: CalendarCheck },
  { href: '/location', label: 'Location', icon: MapPin },
  { href: '/contact-us', label: 'Contact', icon: Phone },
  { href: '/login', label: 'Login', icon: LogIn },
];

function pushEvent(event: Record<string, unknown>) {
  if (typeof window !== 'undefined' && window.dataLayer) {
    window.dataLayer.push(event);
  }
}

export default function Navigation() {
  const pathname = usePathname();
  const [mobileOpen, setMobileOpen] = useState(false);

  if (pathname?.startsWith('/management') || pathname?.startsWith('/tenant') || pathname?.startsWith('/login')) {
    return null;
  }

  return (
    <>
      <header className="fixed top-0 left-0 right-0 z-50 bg-surface/80 backdrop-blur-lg border-b border-outline-variant/50">
        <div className="container-max mx-auto margin-mobile md:margin-desktop h-16 md:h-20 flex items-center justify-between">
          <Link href="/" className="flex items-center">
            <Image src="/logo.png" alt="Vats Apartment" width={120} height={80} className="h-10 md:h-12 w-auto" priority />
          </Link>

          <nav className="hidden md:flex items-center gap-8">
            {NAV_LINKS.map((link) => (
              <Link
                key={link.href}
                href={link.href}
                className={`font-label-caps text-xs tracking-wide transition-colors ${
                  pathname === link.href
                    ? 'text-secondary font-semibold'
                    : 'text-on-surface-variant hover:text-on-surface'
                }`}
              >
                {link.label}
              </Link>
            ))}
            <Link
              href="/login"
              className="bg-primary text-on-secondary px-4 py-2 rounded-lg text-xs font-semibold hover:bg-primary/90 transition-colors"
            >
              Login
            </Link>
          </nav>
        </div>
      </header>

      {mobileOpen && (
        <div className="fixed inset-0 top-16 z-40 bg-surface md:hidden">
          <nav className="flex flex-col p-4 space-y-1">
            {NAV_LINKS.map((link) => (
              <Link
                key={link.href}
                href={link.href}
                onClick={() => setMobileOpen(false)}
                className={`font-label-caps text-sm py-3 px-4 rounded-lg transition-colors ${
                  pathname === link.href
                    ? 'bg-secondary/10 text-secondary font-semibold'
                    : 'text-on-surface-variant hover:bg-surface-container'
                }`}
              >
                {link.label}
              </Link>
            ))}
            <Link
              href="/login"
              onClick={() => setMobileOpen(false)}
              className="bg-primary text-on-secondary py-3 px-4 rounded-lg text-sm font-semibold text-center mt-4 hover:bg-primary/90 transition-colors"
            >
              Login
            </Link>
          </nav>
        </div>
      )}

      <nav className="fixed bottom-0 left-0 right-0 z-50 bg-surface/95 backdrop-blur-lg border-t border-outline-variant/50 md:hidden">
        <div className="flex items-end justify-around px-1 pb-[env(safe-area-inset-bottom)] h-16">
          {MOBILE_TABS.slice(0, 2).map((tab) => {
            const Icon = tab.icon;
            const active = pathname === tab.href;
            return (
              <Link
                key={tab.href}
                href={tab.href}
                className={`flex flex-col items-center justify-center gap-0.5 min-w-[48px] py-1.5 transition-colors ${
                  active ? 'text-secondary' : 'text-on-surface-variant'
                }`}
              >
                <Icon className="w-5 h-5" />
                <span className="text-[10px] font-medium">{tab.label}</span>
              </Link>
            );
          })}

          <a
            href="https://wa.me/919992937447?text=Hi,%20I'm%20interested%20in%20Vats%20Apartment.%20Please%20share%20more%20details."
            target="_blank"
            rel="noopener noreferrer"
            className="flex flex-col items-center justify-center -mt-5"
            onClick={() => pushEvent({ event: 'whatsapp_click', source: 'mobile_bottom_bar' })}
          >
            <div className="w-14 h-14 bg-green-500 rounded-full shadow-lg flex items-center justify-center hover:bg-green-600 transition-all hover:scale-105 border-4 border-surface">
              <MessageCircle className="w-7 h-7 text-white" />
            </div>
            <span className="text-[10px] font-medium text-green-600 mt-0.5">WhatsApp</span>
          </a>

          {MOBILE_TABS.slice(2).map((tab) => {
            const Icon = tab.icon;
            const active = pathname === tab.href;
            return (
              <Link
                key={tab.href}
                href={tab.href}
                className={`flex flex-col items-center justify-center gap-0.5 min-w-[48px] py-1.5 transition-colors ${
                  active ? 'text-secondary' : 'text-on-surface-variant'
                }`}
              >
                <Icon className="w-5 h-5" />
                <span className="text-[10px] font-medium">{tab.label}</span>
              </Link>
            );
          })}
        </div>
      </nav>
    </>
  );
}
