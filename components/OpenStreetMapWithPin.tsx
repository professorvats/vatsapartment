'use client';

import { useEffect, useRef, useState } from 'react';
import { ExternalLink, Navigation } from 'lucide-react';

interface OpenStreetMapProps {
  latitude: number;
  longitude: number;
  title?: string;
  address?: string;
  zoom?: number;
  showDirections?: boolean;
}

export function OpenStreetMapWithPin({
  latitude,
  longitude,
  title = "Vats Apartment",
  address = "Near Apna Chai Wala, Lawgate, Jalandhar, Punjab",
  zoom = 16,
  showDirections = true,
}: OpenStreetMapProps) {
  const mapRef = useRef<HTMLDivElement>(null);
  const mapInstanceRef = useRef<L.Map | null>(null);
  const [isPopupOpen, setIsPopupOpen] = useState(true);

  const MAPS_LINK = 'https://maps.app.goo.gl/FtMggqQiCrC6Rnp96';

  const openInGoogleMaps = () => {
    window.open(MAPS_LINK, '_blank');
  };

  const getDirections = () => {
    window.open(MAPS_LINK, '_blank');
  };

  useEffect(() => {
    if (!mapRef.current) return;

    // Dynamically import Leaflet
    import('leaflet').then((L) => {
      // Fix for default marker icons
      delete (L.Icon.Default.prototype as any)._getIconUrl;
      L.Icon.Default.mergeOptions({
        iconRetinaUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-icon-2x.png',
        iconUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-icon.png',
        shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-shadow.png',
      });

      // Create map if not already created
      if (!mapInstanceRef.current) {
        const map = L.map(mapRef.current).setView([latitude, longitude], zoom);

        // Add OpenStreetMap tiles
        L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
          attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
        }).addTo(map);

        // Create custom marker
        const marker = L.marker([latitude, longitude]).addTo(map);

        // Create popup content with clear Google Maps buttons
        const popupContent = `
          <div style="padding: 12px; font-family: system-ui, -apple-system, sans-serif; min-width: 220px;">
            <h3 style="margin: 0 0 8px 0; font-size: 16px; font-weight: 600; color: #1c1b1f;">
              ${title}
            </h3>
            <p style="margin: 0 0 12px 0; font-size: 13px; color: #49454f; line-height: 1.5;">
              ${address.replace(/\n/g, '<br>')}
            </p>
            <div style="display: flex; flex-direction: column; gap: 8px;">
              <a href="https://maps.app.goo.gl/FtMggqQiCrC6Rnp96"
                 target="_blank"
                 rel="noopener noreferrer"
                 style="display: flex; align-items: center; justify-content: center; gap: 8px;
                        padding: 10px 16px; background: #6750A4; color: white;
                        text-decoration: none; border-radius: 6px; font-size: 14px;
                        font-weight: 500; transition: background 0.2s; text-align: center;"
                 onmouseover="this.style.background='#523e85'"
                 onmouseout="this.style.background='#6750A4'">
                Open in Google Maps
              </a>
              ${showDirections ? `
              <a href="https://maps.app.goo.gl/FtMggqQiCrC6Rnp96"
                 target="_blank"
                 rel="noopener noreferrer"
                 style="display: flex; align-items: center; justify-content: center; gap: 8px;
                        padding: 10px 16px; background: #ffffff; color: #6750A4;
                        text-decoration: none; border: 2px solid #6750A4; border-radius: 6px;
                        font-size: 14px; font-weight: 500; transition: background 0.2s; text-align: center;"
                 onmouseover="this.style.background='#f5f5f5'"
                 onmouseout="this.style.background='#ffffff'">
                Get Directions
              </a>
              ` : ''}
            </div>
          </div>
        `;

        marker.bindPopup(popupContent).openPopup();

        // Auto-open popup and keep it open
        setTimeout(() => {
          marker.openPopup();
        }, 500);

        mapInstanceRef.current = map;
      }
    });

    return () => {
      if (mapInstanceRef.current) {
        mapInstanceRef.current.remove();
        mapInstanceRef.current = null;
      }
    };
  }, [latitude, longitude, title, address, zoom, showDirections]);

  return (
    <div className="relative w-full h-full">
      <div
        ref={mapRef}
        className="w-full h-full rounded-lg overflow-hidden"
        style={{ minHeight: '400px' }}
      />
      {/* External button overlay */}
      <div className="absolute bottom-4 right-4 z-40 flex flex-col gap-2">
        <button
          onClick={openInGoogleMaps}
          className="flex items-center gap-2 bg-white hover:bg-gray-50 text-gray-800 px-4 py-3 rounded-lg shadow-lg border border-gray-200 transition-all hover:scale-105 active:scale-95"
          title="Open in Google Maps"
        >
          <ExternalLink className="w-4 h-4" />
          <span className="font-medium text-sm">Open in Google Maps</span>
        </button>
        {showDirections && (
          <button
            onClick={getDirections}
            className="flex items-center gap-2 bg-[#6750A4] hover:bg-[#523e85] text-white px-4 py-3 rounded-lg shadow-lg transition-all hover:scale-105 active:scale-95"
            title="Get Directions"
          >
            <Navigation className="w-4 h-4" />
            <span className="font-medium text-sm">Get Directions</span>
          </button>
        )}
      </div>
    </div>
  );
}