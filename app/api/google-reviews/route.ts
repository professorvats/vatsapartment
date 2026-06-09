import { NextResponse } from 'next/server';

const PLACE_ID = 'CT5Yun5LmQ_3EBM';
const REVIEW_URL = `https://maps.googleapis.com/maps/api/place/details/json?place_id=${PLACE_ID}&fields=name,rating,user_ratings_total,reviews&key=${process.env.NEXT_PUBLIC_GOOGLE_MAPS_KEY}`;

interface CachedData {
  data: {
    name: string;
    rating: number;
    totalRatings: number;
    reviews: { author: string; rating: number; text: string; time: number }[];
  };
  timestamp: number;
}

let cache: CachedData | null = null;
const CACHE_TTL = 24 * 60 * 60 * 1000; // 24 hours

export async function GET() {
  try {
    if (cache && Date.now() - cache.timestamp < CACHE_TTL) {
      return NextResponse.json(cache.data);
    }

    const response = await fetch(REVIEW_URL);
    const json = await response.json();

    if (json.status !== 'OK' || !json.result) {
      return NextResponse.json({
        name: 'Vats Apartment',
        rating: 0,
        totalRatings: 0,
        reviews: [],
      });
    }

    const data = {
      name: json.result.name,
      rating: json.result.rating || 0,
      totalRatings: json.result.user_ratings_total || 0,
      reviews: (json.result.reviews || []).map((r: any) => ({
        author: r.author_name,
        rating: r.rating,
        text: r.text,
        time: r.time,
      })),
    };

    cache = { data, timestamp: Date.now() };

    return NextResponse.json(data);
  } catch {
    if (cache) {
      return NextResponse.json(cache.data);
    }
    return NextResponse.json({
      name: 'Vats Apartment',
      rating: 0,
      totalRatings: 0,
      reviews: [],
    });
  }
}
