'use client';

import { usePathname } from 'next/navigation';
import { Menu, X } from 'lucide-react';
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

          <button
            onClick={() => setMobileOpen(!mobileOpen)}
            className="md:hidden p-2 text-on-surface"
            aria-label="Toggle menu"
          >
            {mobileOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
          </button>
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
    </>
  );
}
