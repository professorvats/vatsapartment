'use client';

import { useState, useEffect } from 'react';
import { roomsStorage } from '@/lib/rooms-data';

export function PricingDisplay() {
  const [minPrice, setMinPrice] = useState(9000);
  const [availableCount, setAvailableCount] = useState(0);

  useEffect(() => {
    const updatePricing = async () => {
      const available = await roomsStorage.getAvailableRooms();
      setAvailableCount(available.length);
      if (available.length > 0) {
        setMinPrice(Math.min(...available.map(r => r.price)));
      }
    };

    updatePricing();

    const unsubscribe = roomsStorage.subscribe(updatePricing);
    return unsubscribe;
  }, []);

  return (
    <>
      <div className="font-display text-3xl md:text-[40px] leading-tight text-on-surface mb-4 md:mb-6">
        INR {minPrice.toLocaleString()}
        <span className="font-body-md text-on-surface-variant text-sm md:text-base"> /mo</span>
      </div>
      {availableCount > 0 && (
        <div className="text-sm font-body-md text-tertiary mb-4">
          {availableCount} room{availableCount !== 1 ? 's' : ''} available
        </div>
      )}
    </>
  );
}
