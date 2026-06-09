'use client';

import { useState, useEffect } from 'react';
import { Gift, Copy, Check, Share2, User, Mail, Phone, MessageCircle, Instagram, Ticket, Link2, Tag, ChevronRight, Sparkles, IndianRupee, ArrowLeft } from 'lucide-react';

interface ReferralResult {
  id: string;
  name: string | null;
  email: string;
  phone: string;
  whatsapp: string | null;
  instagramHandle: string | null;
  referralCode: string;
  raffleNumber: string;
  shortCode: string;
  referralLink: string;
  createdAt: string;
}

interface ReferrerInfo {
  referralCode: string;
  name: string | null;
  shortCode: string | null;
}

type CopiedField = 'code' | 'link' | 'short' | null;

export default function ReferralPage() {
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    phone: '',
    whatsapp: '',
    instagramHandle: '',
  });

  const [isSubmitting, setIsSubmitting] = useState(false);
  const [result, setResult] = useState<ReferralResult | null>(null);
  const [error, setError] = useState('');
  const [copiedField, setCopiedField] = useState<CopiedField>(null);
  const [referrer, setReferrer] = useState<ReferrerInfo | null>(null);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const refCode = params.get('ref');
    if (refCode) {
      fetch(`/api/referrals?code=${encodeURIComponent(refCode)}`)
        .then((res) => res.json())
        .then((data) => {
          if (data.valid) {
            setReferrer(data.referral);
          }
        })
        .catch(() => { /* ignore */ });
    }
  }, []);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const copyToClipboard = async (text: string, field: CopiedField) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedField(field);
      setTimeout(() => setCopiedField(null), 2000);
    } catch {
      const textarea = document.createElement('textarea');
      textarea.value = text;
      document.body.appendChild(textarea);
      textarea.select();
      document.execCommand('copy');
      document.body.removeChild(textarea);
      setCopiedField(field);
      setTimeout(() => setCopiedField(null), 2000);
    }
  };

  const shareReferral = async () => {
    if (!result) return;
    const shareText = `${result.name || 'My friend'} told me about Vats Apartment near LPU — get ₹2000 off your first booking! Use code: ${result.referralCode}\n\nLink: ${result.referralLink}`;
    if (navigator.share) {
      try {
        await navigator.share({ title: 'Vats Apartment Referral', text: shareText, url: result.referralLink });
      } catch { /* user cancelled */ }
    } else {
      copyToClipboard(shareText, 'link');
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsSubmitting(true);

    try {
      const res = await fetch('/api/referrals', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData),
      });

      const data = await res.json();

      if (!res.ok) {
        setError(data.error || 'Something went wrong. Please try again.');
        return;
      }

      setResult(data.referral);
    } catch {
      setError('Network error. Please check your connection and try again.');
    } finally {
      setIsSubmitting(false);
    }
  };

  const inputClass =
    'w-full bg-transparent border-0 border-b border-outline-variant focus:border-primary focus:ring-0 py-2 md:py-4 px-0 font-body-md text-on-surface transition-colors rounded-none placeholder:text-outline text-xs md:text-lg';

  const labelClass = 'font-label-caps text-on-surface-variant text-[10px] md:text-xl';

  if (referrer && !result) {
    return (
      <main className="container-max mx-auto margin-mobile md:margin-desktop pt-6 md:pt-10 pb-20 md:pb-24 min-h-screen flex flex-col items-center">
        <div className="w-full max-w-2xl">
          <div className="bg-surface-container-lowest rounded-xl border border-outline-variant shadow-sm p-6 md:p-10 space-y-6 md:space-y-8">
            <div className="text-center">
              <div className="inline-flex items-center justify-center w-16 h-16 md:w-20 md:h-20 rounded-full bg-primary/10 mb-4">
                <MessageCircle className="w-8 h-8 md:w-10 md:h-10 text-success" />
              </div>
              <h1 className="text-2xl md:text-4xl font-bold text-on-background mb-2">
                {referrer.name
                  ? `${referrer.name} Recommended Vats Apartment`
                  : 'You Were Referred to Vats Apartment'}
              </h1>
              <p className="text-on-surface-variant text-xs md:text-base max-w-md mx-auto">
                Get <span className="font-semibold text-primary">₹2,000 off</span> your first booking using this referral!
              </p>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <RewardCard
                icon={<Gift className="w-5 h-5 md:w-6 md:h-6 text-primary" />}
                title="You Save"
                amount="₹2,000"
                description="on your first booking"
              />
              <RewardCard
                icon={<IndianRupee className="w-5 h-5 md:w-6 md:h-6 text-success" />}
                title="Referrer Earns"
                amount="₹4,000"
                description="when you book successfully"
              />
            </div>

            <button
              onClick={() => {
                const referrerLabel = referrer.name || 'my friend';
                const message = `Hi, I want to book a room at Vats Apartment. ${referrerLabel} told me about it. Please share available rooms and pricing details.`;
                const waUrl = `https://wa.me/919992937447?text=${encodeURIComponent(message)}`;
                window.open(waUrl, '_blank');
              }}
              className="w-full flex items-center justify-center gap-3 py-4 md:py-5 px-6 rounded-lg bg-success text-on-secondary font-semibold text-sm md:text-xl transition-colors hover:bg-success/90"
            >
              <MessageCircle className="w-5 h-5 md:w-6 md:h-6" />
              Book Now via WhatsApp
            </button>

            <button
              onClick={() => setReferrer(null)}
              className="w-full text-center text-primary font-medium text-xs md:text-base hover:underline flex items-center justify-center gap-1"
            >
              <ArrowLeft className="w-3 h-3 md:w-4 md:h-4" />
              I want to create my own referral code
            </button>
          </div>
        </div>
      </main>
    );
  }

  if (result) {
    return (
      <main className="container-max mx-auto margin-mobile md:margin-desktop pt-6 md:pt-10 pb-20 md:pb-24 min-h-screen flex flex-col items-center">
        <div className="w-full max-w-2xl">
          <div className="bg-surface-container-lowest rounded-xl border border-outline-variant shadow-sm p-6 md:p-10 space-y-6 md:space-y-8">
            <div className="text-center">
              <div className="inline-flex items-center justify-center w-16 h-16 md:w-20 md:h-20 rounded-full bg-success/10 mb-4">
                <Sparkles className="w-8 h-8 md:w-10 md:h-10 text-success" />
              </div>
              <h1 className="text-2xl md:text-4xl font-bold text-on-background mb-2">Your Referral Kit is Ready!</h1>
              <p className="text-on-surface-variant text-xs md:text-base">
                Share your unique codes and start earning
              </p>
            </div>

            <div className="space-y-4 md:space-y-6">
              <CopyCard
                icon={<Ticket className="w-4 h-4 md:w-5 md:h-5 text-primary" />}
                label="Raffle Number"
                value={result.raffleNumber}
                copied={copiedField === 'code'}
                onCopy={() => copyToClipboard(result.raffleNumber, 'code')}
              />
              <CopyCard
                icon={<Tag className="w-4 h-4 md:w-5 md:h-5 text-primary" />}
                label="Promo Code"
                value={result.referralCode}
                copied={copiedField === 'short'}
                onCopy={() => copyToClipboard(result.referralCode, 'short')}
              />
              <CopyCard
                icon={<Link2 className="w-4 h-4 md:w-5 md:h-5 text-primary" />}
                label="Short Code"
                value={result.shortCode}
                copied={copiedField === 'link'}
                onCopy={() => copyToClipboard(result.shortCode, 'link')}
              />
              <CopyCard
                icon={<Link2 className="w-4 h-4 md:w-5 md:h-5 text-tertiary" />}
                label="Referral Link"
                value={result.referralLink}
                copied={copiedField === null && copiedField === 'link'}
                onCopy={() => copyToClipboard(result.referralLink, 'link')}
                isLink
              />
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <button
                onClick={() => {
                  const refName = result.name || 'my friend';
                  const message = `Hi, I want to book a room at Vats Apartment. ${refName} told me about it. Please share available rooms and pricing details.`;
                  const waUrl = `https://wa.me/919992937447?text=${encodeURIComponent(`${message}\n\nReferral code: ${result.referralCode}`)}`;
                  window.open(waUrl, '_blank');
                }}
                className="flex items-center justify-center gap-2 py-3 md:py-4 px-6 rounded-lg bg-success text-on-secondary font-medium text-xs md:text-lg transition-colors hover:bg-success/90"
              >
                <MessageCircle className="w-4 h-4 md:w-5 md:h-5" />
                Share on WhatsApp
              </button>
              <button
                onClick={shareReferral}
                className="flex items-center justify-center gap-2 py-3 md:py-4 px-6 rounded-lg bg-secondary text-on-secondary font-medium text-xs md:text-lg transition-colors hover:bg-secondary/90"
              >
                <Share2 className="w-4 h-4 md:w-5 md:h-5" />
                Share
              </button>
            </div>

            <button
              onClick={() => {
                setResult(null);
                setFormData({ name: '', email: '', phone: '', whatsapp: '', instagramHandle: '' });
              }}
              className="w-full text-center text-primary font-medium text-xs md:text-base hover:underline"
            >
              Generate another referral
            </button>
          </div>

          <div className="mt-6 md:mt-8 grid grid-cols-1 sm:grid-cols-2 gap-4">
            <RewardCard
              icon={<IndianRupee className="w-5 h-5 md:w-6 md:h-6 text-success" />}
              title="You Earn"
              amount="₹4,000"
              description="for every successful booking"
            />
            <RewardCard
              icon={<Gift className="w-5 h-5 md:w-6 md:h-6 text-primary" />}
              title="They Save"
              amount="₹2,000"
              description="on their first booking"
            />
          </div>
        </div>
      </main>
    );
  }

  return (
    <main className="container-max mx-auto margin-mobile md:margin-desktop pt-6 md:pt-10 pb-20 md:pb-24 min-h-screen flex flex-col">
      <div className="mb-6 md:mb-16 max-w-3xl">
        <div className="inline-flex items-center gap-2 bg-primary/10 text-primary px-3 py-1 md:px-4 md:py-1.5 rounded-full text-[10px] md:text-sm font-medium mb-4">
          <Gift className="w-3 h-3 md:w-4 md:h-4" />
          Referral Program
        </div>
        <h1 className="text-2xl md:text-5xl lg:text-6xl font-bold tracking-tight text-on-background mb-2 md:mb-4">
          Refer & Earn
        </h1>
        <p className="font-body-lg text-on-surface-variant text-xs md:text-base max-w-xl">
          Share Vats Apartment with friends. You earn <span className="font-semibold text-success">₹4,000</span> for
          every successful booking, and your friend gets <span className="font-semibold text-primary">₹2,000 off</span> their stay.
        </p>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 md:gap-gutter mb-8 md:mb-12 max-w-2xl">
        <RewardCard
          icon={<IndianRupee className="w-5 h-5 md:w-6 md:h-6 text-success" />}
          title="You Earn"
          amount="₹4,000"
          description="for every successful booking"
        />
        <RewardCard
          icon={<Gift className="w-5 h-5 md:w-6 md:h-6 text-primary" />}
          title="They Save"
          amount="₹2,000"
          description="on their first booking"
        />
      </div>

      <div className="max-w-2xl w-full">
        <h2 className="text-lg md:text-2xl font-bold text-on-background mb-2">Get Your Referral Code</h2>
        <p className="text-on-surface-variant text-xs md:text-base mb-6 md:mb-8">
          Fill in your details below and we&apos;ll generate a unique referral code, link, and raffle number for you.
        </p>

        <form onSubmit={handleSubmit} className="space-y-4 md:space-y-8">
          <div className="flex flex-col space-y-2">
            <label className={labelClass} htmlFor="name">
              <span className="flex items-center gap-1.5">
                <User className="w-3 h-3 md:w-4 md:h-4" />
                Full Name
              </span>
            </label>
            <input
              id="name"
              name="name"
              type="text"
              placeholder="Enter your full name"
              value={formData.name}
              onChange={handleInputChange}
              className={inputClass}
            />
          </div>

          <div className="flex flex-col space-y-2">
            <label className={labelClass} htmlFor="email">
              <span className="flex items-center gap-1.5">
                <Mail className="w-3 h-3 md:w-4 md:h-4" />
                Email Address *
              </span>
            </label>
            <input
              id="email"
              name="email"
              type="email"
              placeholder="you@example.com"
              value={formData.email}
              onChange={handleInputChange}
              required
              className={inputClass}
            />
          </div>

          <div className="flex flex-col space-y-2">
            <label className={labelClass} htmlFor="phone">
              <span className="flex items-center gap-1.5">
                <Phone className="w-3 h-3 md:w-4 md:h-4" />
                Phone Number *
              </span>
            </label>
            <input
              id="phone"
              name="phone"
              type="tel"
              placeholder="+91 9992937447"
              value={formData.phone}
              onChange={handleInputChange}
              required
              className={inputClass}
            />
          </div>

          <div className="flex flex-col space-y-2">
            <label className={labelClass} htmlFor="whatsapp">
              <span className="flex items-center gap-1.5">
                <MessageCircle className="w-3 h-3 md:w-4 md:h-4" />
                WhatsApp Number
              </span>
            </label>
            <input
              id="whatsapp"
              name="whatsapp"
              type="tel"
              placeholder="+91 9992937447 (optional)"
              value={formData.whatsapp}
              onChange={handleInputChange}
              className={inputClass}
            />
          </div>

          <div className="flex flex-col space-y-2">
            <label className={labelClass} htmlFor="instagramHandle">
              <span className="flex items-center gap-1.5">
                <Instagram className="w-3 h-3 md:w-4 md:h-4" />
                Instagram Handle
              </span>
            </label>
            <input
              id="instagramHandle"
              name="instagramHandle"
              type="text"
              placeholder="@your_handle (used for your referral code)"
              value={formData.instagramHandle}
              onChange={handleInputChange}
              className={inputClass}
            />
          </div>

          {error && (
            <div className="bg-error/10 text-error text-xs md:text-sm p-3 md:p-4 rounded-lg border border-error/20">
              {error}
            </div>
          )}

          <div className="pt-4">
            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full flex items-center justify-center gap-2 py-3 md:py-4 px-6 md:px-10 rounded-lg shadow-sm text-xs md:text-lg font-medium text-white bg-primary hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {isSubmitting ? (
                <>
                  <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                  Generating...
                </>
              ) : (
                <>
                  <ChevronRight className="w-4 h-4 md:w-5 md:h-5" />
                  Generate My Referral Code
                </>
              )}
            </button>
          </div>
        </form>
      </div>

      <div className="max-w-2xl w-full mt-12 md:mt-16">
        <h2 className="text-lg md:text-2xl font-bold text-on-background mb-6 md:mb-8">How It Works</h2>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 md:gap-gutter">
          {[
            {
              step: '1',
              title: 'Get Your Code',
              description: 'Fill in your details and get a unique referral code and link.',
            },
            {
              step: '2',
              title: 'Share With Friends',
              description: 'Share your code or link with friends looking for accommodation.',
            },
            {
              step: '3',
              title: 'Earn Rewards',
              description: 'Earn ₹4,000 per successful booking. They save ₹2,000!',
            },
          ].map((item) => (
            <div
              key={item.step}
              className="bg-surface-container-lowest rounded-xl border border-outline-variant p-4 md:p-6"
            >
              <div className="w-8 h-8 md:w-10 md:h-10 rounded-full bg-primary text-on-secondary flex items-center justify-center font-bold text-xs md:text-base mb-3 md:mb-4">
                {item.step}
              </div>
              <h3 className="font-bold text-on-background text-xs md:text-base mb-1 md:mb-2">{item.title}</h3>
              <p className="text-on-surface-variant text-[10px] md:text-sm leading-relaxed">{item.description}</p>
            </div>
          ))}
        </div>
      </div>
    </main>
  );
}

function CopyCard({
  icon,
  label,
  value,
  copied,
  onCopy,
  isLink = false,
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
  copied: boolean;
  onCopy: () => void;
  isLink?: boolean;
}) {
  return (
    <div className="bg-surface-container-low rounded-lg border border-outline-variant p-3 md:p-4 flex items-center justify-between gap-3">
      <div className="flex items-center gap-2 md:gap-3 min-w-0">
        <div className="shrink-0">{icon}</div>
        <div className="min-w-0">
          <p className="text-[10px] md:text-xs text-on-surface-variant font-medium uppercase tracking-wide">
            {label}
          </p>
          <p className={`font-mono font-semibold text-on-surface text-xs md:text-base truncate ${isLink ? 'text-primary' : ''}`}>
            {value}
          </p>
        </div>
      </div>
      <button
        onClick={onCopy}
        className="shrink-0 flex items-center gap-1 px-2 py-1 md:px-3 md:py-1.5 rounded-md bg-surface-container-highest text-on-surface-variant hover:text-on-surface text-[10px] md:text-xs transition-colors"
      >
        {copied ? (
          <>
            <Check className="w-3 h-3 md:w-4 md:h-4 text-success" />
            <span className="text-success">Copied!</span>
          </>
        ) : (
          <>
            <Copy className="w-3 h-3 md:w-4 md:h-4" />
            Copy
          </>
        )}
      </button>
    </div>
  );
}

function RewardCard({
  icon,
  title,
  amount,
  description,
}: {
  icon: React.ReactNode;
  title: string;
  amount: string;
  description: string;
}) {
  return (
    <div className="bg-surface-container-lowest rounded-xl border border-outline-variant shadow-sm p-4 md:p-6 flex items-center gap-3 md:gap-4">
      <div className="shrink-0">{icon}</div>
      <div>
        <p className="text-[10px] md:text-xs text-on-surface-variant font-medium uppercase tracking-wide">
          {title}
        </p>
        <p className="text-xl md:text-3xl font-bold text-on-background">{amount}</p>
        <p className="text-[10px] md:text-sm text-on-surface-variant">{description}</p>
      </div>
    </div>
  );
}
