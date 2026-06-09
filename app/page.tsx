'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { MapPin, Video, Dumbbell, ShoppingBag, GraduationCap, Car, ChevronDown, Bed, ShowerHead, Tv, Refrigerator, Wifi, Monitor, DoorOpen, Star, Wind, Flame, Users } from 'lucide-react';
import { PricingDisplay } from '@/components/PricingDisplay';
import { OpenStreetMapLoader } from '@/components/OpenStreetMapLoader';
import { OpenStreetMapWithPin } from '@/components/OpenStreetMapWithPin';

interface Review {
  author: string;
  rating: number;
  text: string;
  time: number;
}

export default function HomePage() {
  const [reviews, setReviews] = useState<Review[]>([]);
  const [rating, setRating] = useState(0);
  const [totalRatings, setTotalRatings] = useState(0);

  useEffect(() => {
    fetch('/api/google-reviews')
      .then((res) => res.json())
      .then((data) => {
        setReviews(data.reviews || []);
        setRating(data.rating || 0);
        setTotalRatings(data.totalRatings || 0);
      })
      .catch(() => {});
  }, []);

  const handleBookNow = () => {
    if (typeof window !== 'undefined' && window.dataLayer) {
      window.dataLayer.push({ event: 'whatsapp_click', source: 'check_availability_button' });
    }
    const message = encodeURIComponent("Hi, I'm interested in booking a room at Vats Apartment. Please share more details.");
    window.open(`https://wa.me/919992937447?text=${message}`, '_blank');
  };

  const handleCallNow = () => {
    if (typeof window !== 'undefined' && window.dataLayer) {
      window.dataLayer.push({ event: 'phone_call_click', source: 'call_now_button' });
    }
  };

  const amenities = [
    { icon: 'bed', name: 'Queen Size Bed' },
    { icon: 'ac', name: 'Air Conditioner' },
    { icon: 'kitchen', name: 'Fridge' },
    { icon: 'wifi', name: 'High-Speed Wifi' },
    { icon: 'tv', name: 'Smart TV' },
    { icon: 'geyser', name: 'Geyser' },
    { icon: 'desk', name: 'Table & Chair' },
    { icon: 'almirah', name: 'Almirah' },
    { icon: 'shower', name: 'Shower' },
  ];

  const features = [
    {
      icon: 'bed',
      title: '1 Bedroom',
      description: 'Spacious private room with a queen size bed. Can be shared between two people if desired.',
    },
    {
      icon: 'countertops',
      title: 'Personal Kitchen',
      description: 'Private kitchen with built-in storage, fridge, and all essential cookware.',
    },
    {
      icon: 'shower',
      title: 'Attached Washroom',
      description: 'Private western-style commode washroom with shower and hot water geyser.',
    },
  ];

  return (
    <div className="w-full flex flex-col pb-section-gap">
      {/* Hero Section */}
      <section className="container-max gutter w-full pt-4 md:pt-6">
        <div className="relative w-full h-[300px] md:h-[400px] lg:h-[500px] rounded-xl overflow-hidden mb-6 md:mb-8">
          <div className="absolute inset-0 bg-gradient-to-t from-black/70 via-black/30 to-transparent flex flex-col justify-end p-3 md:p-6 lg:p-16 z-10">
            <div className="flex flex-wrap gap-2 md:gap-3 mb-2 md:mb-3">
              <span className="bg-secondary-container text-on-secondary-container font-label-caps px-3 md:px-6 py-1.5 md:py-3 rounded-full uppercase text-[12px] md:text-sm">
                Premium Listing
              </span>
              <span className="bg-surface/90 text-on-surface font-label-caps px-3 md:px-6 py-1.5 md:py-3 rounded-full uppercase text-[12px] md:text-sm">
                Available Now
              </span>
            </div>
            <h1 className="text-2xl md:text-4xl lg:text-5xl font-bold tracking-tight text-white mb-1.5 md:mb-3 max-w-3xl leading-tight">
              <span className="block">Premium Modern</span>
              <span className="block">Living near LPU</span>
            </h1>
            <p className="font-body-sm md:text-body-lg text-white/90 max-w-2xl flex items-center gap-1.5 md:gap-2 text-[10px] md:text-sm">
              <MapPin className="w-3 h-3 md:w-5 md:h-5" />
              <span className="line-clamp-2">Near Apna Chai Wala, LPU, Jalandhar, Punjab</span>
            </p>
          </div>
          <div className="w-full h-full bg-gradient-to-br from-primary/20 to-secondary/20" />
        </div>
      </section>

      {/* Bento Grid Details */}
      <section className="container-max gutter w-full grid grid-cols-1 lg:grid-cols-3 gap-4 md:gap-6 mt-6 md:mt-12 mb-section-gap">
        {/* Pricing Card */}
        <div className="col-span-1 lg:col-span-1 bg-surface-container-low rounded-xl border-t-4 border-secondary border border-surface-container flex flex-col justify-between overflow-hidden">
          <div className="p-4 md:p-8">
            <p className="font-label-caps text-outline uppercase tracking-widest text-[10px] md:text-xs mb-1 md:mb-2">Monthly Rent</p>
            <p className="font-label-caps text-secondary uppercase tracking-widest text-[9px] md:text-[11px] mb-2 md:mb-3">Starting from</p>
            <PricingDisplay />
            <div className="mt-3 md:mt-6 space-y-0 font-body-md text-on-surface-variant text-xs md:text-sm">
              <div className="flex justify-between items-center py-2 md:py-3 border-b border-outline-variant/40">
                <span className="text-on-surface-variant text-[10px] md:text-sm">Security Deposit (Refundable)</span>
                <span className="font-semibold text-on-surface text-xs md:text-sm">10 Months Rent</span>
              </div>
              <div className="flex justify-between items-center py-2 md:py-3">
                <span className="text-on-surface-variant text-[10px] md:text-sm">Electricity</span>
                <span className="font-semibold text-on-surface text-xs md:text-sm">₹12 / unit</span>
              </div>
            </div>
          </div>
          <div className="px-4 md:px-8 pb-4 md:pb-8">
            <button
              onClick={handleBookNow}
              className="w-full bg-secondary text-on-secondary font-body-md py-2.5 md:py-4 rounded-lg hover:bg-secondary/90 transition-colors text-xs md:text-base font-semibold"
            >
              Check Availability
            </button>
          </div>
        </div>

        {/* Features Bento */}
        <div className="col-span-1 lg:col-span-2 grid grid-cols-1 md:grid-cols-2 gap-4 md:gap-6">
          {/* Layout Box */}
          <div className="bg-surface-container-lowest rounded-xl p-4 md:p-8 border border-surface-container-highest shadow-sm flex flex-col">
            <h3 className="font-h2 text-on-surface mb-3 md:mb-4 text-sm md:text-lg">Apartment Layout</h3>
            <ul className="space-y-3 md:space-y-5 font-body-md md:font-body-lg text-on-surface-variant flex-grow text-xs md:text-base">
              {features.map((feature) => (
                <li key={feature.title} className="flex items-start gap-2 md:gap-4">
                  <div className="text-secondary mt-0.5 md:mt-1">
                    {feature.title === '1 Bedroom' && <Bed className="w-5 h-5 md:w-8 md:h-8" />}
                    {feature.title === 'Modern Kitchen' && <ShoppingBag className="w-5 h-5 md:w-8 md:h-8" />}
                    {feature.title === 'Full Bathroom' && <ShowerHead className="w-5 h-5 md:w-8 md:h-8" />}
                  </div>
                  <div>
                    <span className="block font-h3 text-on-surface text-xs md:text-base">{feature.title}</span>
                    <span className="font-body-md text-[10px] md:text-sm leading-relaxed">{feature.description}</span>
                  </div>
                </li>
              ))}
            </ul>
          </div>

          {/* Building Extras */}
          <div className="bg-surface-container-lowest rounded-xl p-0 border border-surface-container-highest shadow-sm overflow-hidden flex flex-col relative">
            <div className="w-full h-24 md:h-48 bg-gradient-to-br from-secondary/20 to-tertiary/20" />
            <div className="p-3 md:p-6 flex-grow">
              <h3 className="font-label-caps text-outline uppercase tracking-widest text-[10px] md:text-xs mb-2 md:mb-4">Building Extras</h3>
              <div className="flex flex-wrap gap-1 md:gap-2">
                <span className="inline-flex items-center gap-0.5 md:gap-1.5 bg-surface-container text-on-surface font-caption px-1.5 md:px-4 py-0.5 md:py-1.5 rounded-[10px] md:text-sm text-[10px]">
                  <Video className="w-[10px] h-[10px] md:w-[16px] md:h-[16px]" />
                  24/7 Security
                </span>
                <span className="inline-flex items-center gap-0.5 md:gap-1.5 bg-surface-container text-on-surface font-caption px-1.5 md:px-4 py-0.5 md:py-1.5 rounded-[10px] md:text-sm text-[10px]">
                  <Dumbbell className="w-[10px] h-[10px] md:w-[16px] md:h-[16px]" />
                  Gym 200m
                </span>
                <span className="inline-flex items-center gap-0.5 md:gap-1.5 bg-surface-container text-on-surface font-caption px-1.5 md:px-4 py-0.5 md:py-1.5 rounded-[10px] md:text-sm text-[10px]">
                  <ShoppingBag className="w-[10px] h-[10px] md:w-[16px] md:h-[16px]" />
                  Groceries 200m
                </span>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Booking CTA Section */}
      <section className="container-max gutter w-full mb-section-gap">
        <div className="bg-surface-container-highest rounded-xl p-4 md:p-12 text-center border border-outline-variant">
          <h2 className="font-h1 text-on-surface mb-2 md:mb-4 text-lg md:text-4xl">Ready to Book?</h2>
          <p className="font-body-lg text-on-surface-variant mb-4 md:mb-8 max-w-2xl mx-auto text-xs md:text-base">
            Check room availability and book your premium space near LPU campus today.
          </p>
          <div className="flex flex-col sm:flex-row gap-3 md:gap-4 justify-center">
            <button
              onClick={handleBookNow}
              className="bg-secondary text-on-secondary font-body-md py-2 md:py-4 px-6 md:px-10 rounded-lg hover:bg-secondary/90 transition-colors text-xs md:text-lg"
            >
              Check Availability
            </button>
            <a
              href="tel:+919992937447"
              onClick={handleCallNow}
              className="bg-surface-container text-on-surface font-body-md py-2 md:py-4 px-6 md:px-10 rounded-lg hover:bg-surface-container-highest transition-colors border border-outline text-xs md:text-lg inline-flex items-center justify-center gap-2"
            >
              Call Now
            </a>
          </div>
        </div>
      </section>

      {/* Amenits Section */}
      <section className="container-max gutter w-full mb-section-gap">
        <div className="text-center mb-4 md:mb-12">
          <h2 className="font-h1 text-on-surface mb-2 md:mb-4 text-xl md:text-3xl lg:text-4xl">Fully Furnished for Comfort</h2>
          <p className="font-body-sm md:text-body-lg text-on-surface-variant max-w-2xl mx-auto text-[10px] md:text-base px-4">
            Everything you need to move right in. Queen size bed, AC, personal kitchen with fridge,
            western washroom with shower & geyser, Smart TV, Wi-Fi, and more.
          </p>
        </div>
        <div className="grid grid-cols-3 md:grid-cols-3 lg:grid-cols-3 gap-2 md:gap-4 max-w-4xl mx-auto">
          {amenities.map((amenity) => (
            <div
              key={amenity.name}
              className="bg-surface-container-low p-2 md:p-6 rounded-lg flex flex-col items-center justify-center text-center border border-transparent hover:border-outline-variant transition-colors group"
            >
              <div className="text-xl md:text-4xl text-on-surface-variant mb-1 md:mb-3 group-hover:text-primary transition-colors">
                {amenity.name === 'Smart TV' && <Tv className="w-4 h-4 md:w-8 md:h-8" />}
                {amenity.name === 'Fridge' && <Refrigerator className="w-4 h-4 md:w-8 md:h-8" />}
                {amenity.name === 'High-Speed Wifi' && <Wifi className="w-4 h-4 md:w-8 md:h-8" />}
                {amenity.name === 'Table & Chair' && <Monitor className="w-4 h-4 md:w-8 md:h-8" />}
                {amenity.name === 'Queen Size Bed' && <Bed className="w-4 h-4 md:w-8 md:h-8" />}
                {amenity.name === 'Almirah' && <DoorOpen className="w-4 h-4 md:w-8 md:h-8" />}
                {amenity.name === 'Air Conditioner' && <Wind className="w-4 h-4 md:w-8 md:h-8" />}
                {amenity.name === 'Geyser' && <Flame className="w-4 h-4 md:w-8 md:h-8" />}
                {amenity.name === 'Shower' && <ShowerHead className="w-4 h-4 md:w-8 md:h-8" />}
              </div>
              <span className="font-body-md text-on-surface text-[8px] md:text-sm leading-tight">
                {amenity.name}
              </span>
            </div>
          ))}
        </div>
      </section>

      {/* Google Reviews Section */}
      <section className="container-max gutter w-full mb-section-gap">
        <div className="bg-surface-container-highest rounded-xl p-4 md:p-12 border border-outline-variant">
          <div className="text-center mb-6 md:mb-10">
            <h2 className="font-h1 text-on-surface mb-2 md:mb-3 text-lg md:text-3xl lg:text-4xl">
              What Our Guests Say
            </h2>
            <p className="font-body-md text-on-surface-variant text-[10px] md:text-base max-w-2xl mx-auto">
              Your experience matters to us. Read reviews from our guests and share your own.
            </p>
          </div>

          {/* Rating Summary */}
          <div className="flex flex-col items-center mb-6 md:mb-10">
            <div className="flex items-center gap-1 md:gap-2 mb-2">
              {[1, 2, 3, 4, 5].map((star) => (
                <Star
                  key={star}
                  className={`w-6 h-6 md:w-10 md:h-10 ${
                    star <= Math.round(rating)
                      ? 'text-yellow-400 fill-yellow-400'
                      : 'text-outline-variant'
                  }`}
                />
              ))}
            </div>
            <div className="flex items-center gap-2 md:gap-3">
              <span className="font-h2 text-on-surface text-lg md:text-3xl">{rating > 0 ? rating.toFixed(1) : '—'}</span>
              <span className="font-body-md text-on-surface-variant text-xs md:text-base">
                {totalRatings > 0 ? `Based on ${totalRatings} Google review${totalRatings !== 1 ? 's' : ''}` : 'Google Reviews'}
              </span>
            </div>
          </div>

          {/* Review Cards */}
          {reviews.length > 0 && (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3 md:gap-6 mb-6 md:mb-10">
              {reviews.slice(0, 3).map((review, i) => (
                <div
                  key={i}
                  className="bg-surface-container-low rounded-lg p-4 md:p-6 border border-outline-variant flex flex-col"
                >
                  <div className="flex items-center gap-1 mb-2 md:mb-3">
                    {Array.from({ length: 5 }).map((_, s) => (
                      <Star
                        key={s}
                        className={`w-3 h-3 md:w-4 md:h-4 ${
                          s < review.rating ? 'text-yellow-400 fill-yellow-400' : 'text-outline-variant'
                        }`}
                      />
                    ))}
                  </div>
                  <p className="font-body-md text-on-surface text-[10px] md:text-sm flex-grow line-clamp-4 md:line-clamp-5">
                    &ldquo;{review.text}&rdquo;
                  </p>
                  <p className="font-caption text-on-surface-variant mt-3 md:mt-4 text-[10px] md:text-xs">
                    — {review.author}
                  </p>
                </div>
              ))}
            </div>
          )}

          {/* Leave a Review CTA */}
          <div className="text-center">
            <a
              href="https://g.page/r/CT5Yun5LmQ_3EBM/review"
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-2 md:gap-3 bg-[#4285F4] text-white font-body-md py-2.5 md:py-4 px-6 md:px-10 rounded-lg hover:bg-[#3367D6] transition-colors text-xs md:text-lg font-semibold shadow-md"
            >
              <svg className="w-4 h-4 md:w-6 md:h-6" viewBox="0 0 24 24" fill="currentColor">
                <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z" />
                <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" />
                <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" />
                <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" />
              </svg>
              Leave a Google Review
            </a>
            <p className="font-caption text-on-surface-variant mt-2 md:mt-3 text-[10px] md:text-xs">
              Your feedback helps us improve and helps others find their perfect space.
            </p>
          </div>
        </div>
      </section>

      {/* Location Section */}
      <section className="container-max gutter w-full mb-section-gap">
        <div className="bg-surface-container-highest rounded-xl overflow-hidden flex flex-col md:flex-row">
          <div className="w-full md:w-1/2 p-4 md:p-16 flex flex-col justify-center bg-surface-container-lowest border-r border-surface-container">
            <h3 className="font-label-caps text-outline mb-2 uppercase text-[10px] md:text-sm">Location</h3>
            <p className="font-body-lg text-on-surface-variant mb-4 md:mb-8 text-xs md:text-base">
              Situated near Apna Chai Wala in LPU, Jalandhar. This premium suite offers
              unmatched convenience for students and professionals alike, blending quiet privacy
              with immediate access to essential transit and campus life.
            </p>
            <ul className="space-y-4 md:space-y-6 font-body-md text-on-surface">
              <li className="flex items-center gap-3 md:gap-4">
                <div className="w-10 h-10 md:w-12 md:h-12 rounded-full bg-secondary-container flex items-center justify-center text-on-secondary-container shrink-0">
                  <GraduationCap className="w-5 h-5 md:w-6 md:h-6" />
                </div>
                <div>
                  <span className="block font-h3 text-xs md:text-base">10 Mins from Campus</span>
                  <span className="text-on-surface-variant text-[10px] md:text-sm">Quick daily commute to LPU University</span>
                </div>
              </li>
              <li className="flex items-center gap-3 md:gap-4">
                <div className="w-10 h-10 md:w-12 md:h-12 rounded-full bg-secondary-container flex items-center justify-center text-on-secondary-container shrink-0">
                  <Car className="w-5 h-5 md:w-6 md:h-6" />
                </div>
                <div>
                  <span className="block font-h3 text-xs md:text-base">200m from Auto Stand</span>
                  <span className="text-on-surface-variant text-[10px] md:text-sm">Easy city connectivity</span>
                </div>
              </li>
            </ul>
          </div>
          <div className="w-full md:w-1/2 h-[300px] md:h-auto relative">
            <OpenStreetMapLoader>
              <OpenStreetMapWithPin
                latitude={31.253501}
                longitude={75.694228}
                title="Vats Apartment"
                address="Near Apna Chai Wala, LPU, Jalandhar, Punjab"
                zoom={15}
              />
            </OpenStreetMapLoader>
          </div>
        </div>
      </section>

      {/* FAQ Section */}
      <section className="container-max gutter w-full">
        <div>
          <h2 className="font-h1 text-on-surface mb-6 md:mb-10 text-center text-lg md:text-2xl lg:text-3xl">
            Frequently Asked Questions About PG Near LPU
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3 md:gap-6">
            <details className="group bg-surface-container-low rounded-lg p-3 md:p-6 border border-outline-variant cursor-pointer">
              <summary className="flex justify-between items-center font-h3 text-on-surface list-none text-xs md:text-base">
                How far is the room near LPU University campus?
                <ChevronDown className="transition-transform group-open:rotate-180 w-4 h-4 md:w-5 md:h-5 shrink-0 ml-2" />
              </summary>
              <p className="mt-3 md:mt-4 font-body-md text-on-surface-variant text-[10px] md:text-base">
                Our PG accommodation is just 10 minutes from LPU University campus, making it one
                of the most conveniently located rooms near LPU. You can easily walk or take a quick
                auto ride to reach your classes.
              </p>
            </details>

            <details className="group bg-surface-container-low rounded-lg p-3 md:p-6 border border-outline-variant cursor-pointer">
              <summary className="flex justify-between items-center font-h3 text-on-surface list-none text-xs md:text-base">
                What is the rent for rooms near LPU?
                <ChevronDown className="transition-transform group-open:rotate-180 w-4 h-4 md:w-5 md:h-5 shrink-0 ml-2" />
              </summary>
              <p className="mt-3 md:mt-4 font-body-md text-on-surface-variant text-[10px] md:text-base">
                Our fully furnished rooms near LPU start at ₹9,000/month, which is very competitive
                compared to other PGs near LPU University. This includes 24/7 security, high-speed
                WiFi, and all modern amenities.
              </p>
            </details>

            <details className="group bg-surface-container-low rounded-lg p-3 md:p-6 border border-outline-variant cursor-pointer">
              <summary className="flex justify-between items-center font-h3 text-on-surface list-none text-xs md:text-base">
                Can two people share a room?
                <ChevronDown className="transition-transform group-open:rotate-180 w-4 h-4 md:w-5 md:h-5 shrink-0 ml-2" />
              </summary>
              <p className="mt-3 md:mt-4 font-body-md text-on-surface-variant text-[10px] md:text-base">
                Yes! Our rooms are spacious and can comfortably accommodate two people sharing.
                The room comes with a queen size bed, and we can arrange an additional bed if needed.
                It&apos;s a great way to split the rent and make it even more affordable.
              </p>
            </details>

            <details className="group bg-surface-container-low rounded-lg p-3 md:p-6 border border-outline-variant cursor-pointer">
              <summary className="flex justify-between items-center font-h3 text-on-surface list-none text-xs md:text-base">
                What type of bed and bathroom does each room have?
                <ChevronDown className="transition-transform group-open:rotate-180 w-4 h-4 md:w-5 md:h-5 shrink-0 ml-2" />
              </summary>
              <p className="mt-3 md:mt-4 font-body-md text-on-surface-variant text-[10px] md:text-base">
                Each room features a comfortable queen size bed. Every room has its own attached
                personal washroom with a western-style commode, shower, and hot water geyser —
                no sharing with other tenants.
              </p>
            </details>

            <details className="group bg-surface-container-low rounded-lg p-3 md:p-6 border border-outline-variant cursor-pointer">
              <summary className="flex justify-between items-center font-h3 text-on-surface list-none text-xs md:text-base">
                Does each room have its own kitchen?
                <ChevronDown className="transition-transform group-open:rotate-180 w-4 h-4 md:w-5 md:h-5 shrink-0 ml-2" />
              </summary>
              <p className="mt-3 md:mt-4 font-body-md text-on-surface-variant text-[10px] md:text-base">
                Yes, every room comes with its own personal kitchen equipped with built-in storage,
                a refrigerator, and essential cookware. You don&apos;t need to share kitchen space
                with anyone else.
              </p>
            </details>

            <details className="group bg-surface-container-low rounded-lg p-3 md:p-6 border border-outline-variant cursor-pointer">
              <summary className="flex justify-between items-center font-h3 text-on-surface list-none text-xs md:text-base">
                What amenities are included in the room?
                <ChevronDown className="transition-transform group-open:rotate-180 w-4 h-4 md:w-5 md:h-5 shrink-0 ml-2" />
              </summary>
              <p className="mt-3 md:mt-4 font-body-md text-on-surface-variant text-[10px] md:text-base">
                Our rooms are fully furnished and include: queen size bed, air conditioner (AC),
                fridge, high-speed WiFi, Smart TV, geyser for hot water, shower, western-style
                commode washroom, table & chair, and almirah. Everything you need — just bring
                your bags!
              </p>
            </details>

            <details className="group bg-surface-container-low rounded-lg p-3 md:p-6 border border-outline-variant cursor-pointer">
              <summary className="flex justify-between items-center font-h3 text-on-surface list-none text-xs md:text-base">
                Is this PG better than LPU hostel?
                <ChevronDown className="transition-transform group-open:rotate-180 w-4 h-4 md:w-5 md:h-5 shrink-0 ml-2" />
              </summary>
              <p className="mt-3 md:mt-4 font-body-md text-on-surface-variant text-[10px] md:text-base">
                Many students prefer our PG near LPU over university hostels because of the
                privacy, modern amenities, flexible timings, and homely atmosphere. Plus, we&apos;re
                just 10 minutes from campus, giving you the best of both worlds.
              </p>
            </details>

            <details className="group bg-surface-container-low rounded-lg p-3 md:p-6 border border-outline-variant cursor-pointer">
              <summary className="flex justify-between items-center font-h3 text-on-surface list-none text-xs md:text-base">
                Are there affordable budget PG options near LPU?
                <ChevronDown className="transition-transform group-open:rotate-180 w-4 h-4 md:w-5 md:h-5 shrink-0 ml-2" />
              </summary>
              <p className="mt-3 md:mt-4 font-body-md text-on-surface-variant text-[10px] md:text-base">
                Yes! At ₹9,000/month, we offer one of the most budget-friendly fully furnished
                PG accommodations near LPU University. You get premium amenities at student-friendly
                prices with no hidden charges.
              </p>
            </details>
          </div>
        </div>
      </section>
    </div>
  );
}
