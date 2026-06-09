'use client';

import { usePathname } from 'next/navigation';
import { MessageCircle, Phone, LogIn } from 'lucide-react';
import Image from 'next/image';
import Link from 'next/link';

function pushEvent(event: Record<string, unknown>) {
  if (typeof window !== 'undefined' && window.dataLayer) {
    window.dataLayer.push(event);
  }
}

export default function FooterWithWhatsApp() {
  const pathname = usePathname();

  if (pathname?.startsWith('/management') || pathname?.startsWith('/tenant') || pathname?.startsWith('/login')) {
    return null;
  }

  return (
    <>
      <a
        href="https://wa.me/919992937447?text=Hi,%20I'm%20interested%20in%20Vats%20Apartment.%20Please%20share%20more%20details."
        target="_blank"
        rel="noopener noreferrer"
        className="fixed bottom-6 right-4 md:bottom-8 md:right-8 z-[1001] w-14 h-14 md:w-16 md:h-16 bg-green-500 text-white rounded-full shadow-lg flex items-center justify-center hover:bg-green-600 transition-all hover:scale-105"
        aria-label="Chat on WhatsApp"
        onClick={() => pushEvent({ event: 'whatsapp_click', source: 'floating_button' })}
      >
        <MessageCircle className="w-7 h-7 md:w-8 md:h-8" />
      </a>

      <footer className="bg-surface-container-lowest border-t border-outline-variant w-full py-10 md:py-16 mt-auto">
        <div className="container-max mx-auto margin-mobile md:margin-desktop">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8 md:gap-12 items-start">
            <div>
              <Image src="/logo.png" alt="Vats Apartment" width={200} height={134} className="h-20 w-auto mb-4" priority />
              <p className="text-on-surface-variant text-xs md:text-sm leading-relaxed max-w-xs">
                Premium fully furnished PG accommodation near LPU University, Jalandhar.
              </p>
            </div>

            <div className="flex flex-col space-y-3">
              <h3 className="font-label-caps text-on-surface text-[10px] md:text-xs tracking-widest uppercase">Quick Links</h3>
              <Link href="/book-now" className="text-on-surface-variant text-xs md:text-sm hover:text-primary transition-colors">Book Now</Link>
              <Link href="/referral" className="text-on-surface-variant text-xs md:text-sm hover:text-primary transition-colors">Refer & Earn</Link>
              <Link href="/privacy" className="text-on-surface-variant text-xs md:text-sm hover:text-primary transition-colors">Privacy Policy</Link>
              <Link href="/terms" className="text-on-surface-variant text-xs md:text-sm hover:text-primary transition-colors">Terms</Link>
              <Link href="/login" className="text-on-surface-variant text-xs md:text-sm hover:text-primary transition-colors inline-flex items-center gap-1.5">
                <LogIn className="w-3 h-3" />
                Login
              </Link>
            </div>

            <div className="flex flex-col space-y-3">
              <h3 className="font-label-caps text-on-surface text-[10px] md:text-xs tracking-widest uppercase">Contact</h3>
              <a
                href="tel:+919992937447"
                className="text-on-surface-variant text-xs md:text-sm hover:text-primary transition-colors inline-flex items-center gap-2"
                onClick={() => pushEvent({ event: 'phone_call_click', source: 'footer' })}
              >
                <Phone className="w-3 h-3 md:w-4 md:h-4" />
                +91 9992937447
              </a>
              <a
                href="https://wa.me/919992937447"
                target="_blank"
                rel="noopener noreferrer"
                className="text-success text-xs md:text-sm hover:text-green-600 transition-colors inline-flex items-center gap-2"
                onClick={() => pushEvent({ event: 'whatsapp_click', source: 'footer_link' })}
              >
                <MessageCircle className="w-3 h-3 md:w-4 md:h-4" />
                WhatsApp Us
              </a>
            </div>
          </div>

          <div className="border-t border-outline-variant mt-8 md:mt-12 pt-6 md:pt-8 text-center">
            <p className="text-on-surface-variant text-[10px] md:text-xs">
              &copy; {new Date().getFullYear()} Vats Apartment. All rights reserved.
            </p>
          </div>
        </div>
      </footer>
    </>
  );
}
