'use client';

import { useEffect, useState } from 'react';
import 'leaflet/dist/leaflet.css';

export function OpenStreetMapLoader({ children }: { children: React.ReactNode }) {
  const [isLoaded, setIsLoaded] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // OpenStreetMap doesn't require any API key or external script loading
    // It just needs the Leaflet CSS which we've imported above
    setIsLoaded(true);
  }, []);

  if (error) {
    return (
      <div className="w-full h-full flex items-center justify-center bg-surface-container-lowest rounded-lg">
        <p className="text-on-surface-variant">Map unavailable.</p>
      </div>
    );
  }

  if (!isLoaded) {
    return (
      <div className="w-full h-full flex items-center justify-center bg-surface-container-lowest rounded-lg">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
      </div>
    );
  }

  return <>{children}</>;
}