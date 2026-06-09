import { NextRequest, NextResponse } from 'next/server';
import { db } from '@/db';
import { referrals } from '@/db/schema';
import { eq } from 'drizzle-orm';
import crypto from 'crypto';

function generateRaffleNumber(): string {
  const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789';
  let result = '';
  for (let i = 0; i < 3; i++) {
    result += chars[Math.floor(Math.random() * chars.length)];
  }
  result += '-';
  for (let i = 0; i < 4; i++) {
    result += chars[Math.floor(Math.random() * chars.length)];
  }
  return result;
}

function sanitizeHandle(input: string): string {
  return input
    .replace(/[^a-zA-Z0-9]/g, '')
    .toLowerCase()
    .slice(0, 12);
}

function generateReferralCode(email: string, phone: string, name?: string, instagramHandle?: string): string {
  let base = '';

  if (instagramHandle && instagramHandle.trim()) {
    base = sanitizeHandle(instagramHandle);
  } else if (name && name.trim()) {
    base = sanitizeHandle(name);
  } else {
    const emailPart = email.split('@')[0];
    base = sanitizeHandle(emailPart) || sanitizeHandle(phone);
  }

  if (base.length === 0) {
    base = sanitizeHandle(phone);
  }

  const suffix = Math.random().toString(36).slice(2, 6);
  const code = `${base}${suffix}`.toUpperCase().slice(0, 10);
  return code;
}

function generateShortCode(instagramHandle?: string, name?: string): string {
  let base = '';

  if (instagramHandle && instagramHandle.trim()) {
    base = sanitizeHandle(instagramHandle).slice(0, 4);
  } else if (name && name.trim()) {
    const parts = name.trim().split(/\s+/);
    base = parts.map(p => sanitizeHandle(p).slice(0, 2)).join('').slice(0, 4);
  }

  if (base.length < 3) {
    base = crypto.randomBytes(3).toString('hex').slice(0, 4);
  }

  const num = Math.floor(1000 + Math.random() * 9000);
  return `${base}${num}`.toUpperCase();
}

function generateReferralLink(code: string, request: NextRequest): string {
  const baseUrl = process.env.NEXT_PUBLIC_SITE_URL || `${request.nextUrl.protocol}//${request.nextUrl.host}`;
  return `${baseUrl}/referral?ref=${code}`;
}

// GET /api/referrals - Check if referral code exists
async function GET(request: NextRequest) {
  try {
    const searchParams = request.nextUrl.searchParams;
    const code = searchParams.get('code');

    if (!code) {
      return NextResponse.json({ error: 'Code is required' }, { status: 400 });
    }

    const result = await db
      .select({
        referralCode: referrals.referralCode,
        shortCode: referrals.shortCode,
        name: referrals.name,
        raffleNumber: referrals.raffleNumber,
      })
      .from(referrals)
      .where(eq(referrals.referralCode, code.toUpperCase()))
      .limit(1);

    if (result.length === 0) {
      return NextResponse.json({ valid: false }, { status: 404 });
    }

    return NextResponse.json({ valid: true, referral: result[0] });
  } catch (error: unknown) {
    console.error('Error checking referral:', error);
    const err = error as { code?: string; message?: string };
    if (err.code === '42P01' || err.message?.includes('does not exist')) {
      return NextResponse.json({ error: 'Database not ready' }, { status: 503 });
    }
    return NextResponse.json({ error: 'Failed to check referral' }, { status: 500 });
  }
}

// POST /api/referrals - Create a new referral
async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const { email, phone, whatsapp, name, instagramHandle } = body;

    if (!email || !phone) {
      return NextResponse.json(
        { error: 'Email and phone are required' },
        { status: 400 }
      );
    }

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email)) {
      return NextResponse.json(
        { error: 'Invalid email address' },
        { status: 400 }
      );
    }

    const phoneClean = phone.replace(/\D/g, '');
    if (phoneClean.length < 10) {
      return NextResponse.json(
        { error: 'Phone number must be at least 10 digits' },
        { status: 400 }
      );
    }

    const existingEmail = await db
      .select({ id: referrals.id })
      .from(referrals)
      .where(eq(referrals.email, email.toLowerCase().trim()))
      .limit(1);

    if (existingEmail.length > 0) {
      return NextResponse.json(
        { error: 'This email is already registered in our referral program' },
        { status: 409 }
      );
    }

    const referralCode = generateReferralCode(email, phone, name, instagramHandle);
    const raffleNumber = generateRaffleNumber();
    const shortCode = generateShortCode(instagramHandle, name);
    const referralLink = generateReferralLink(referralCode, request);

    const id = crypto.randomUUID();

    const newReferral = await db
      .insert(referrals)
      .values({
        id,
        name: name?.trim() || null,
        email: email.toLowerCase().trim(),
        phone: phoneClean,
        whatsapp: whatsapp?.replace(/\D/g, '') || null,
        instagramHandle: instagramHandle?.trim() || null,
        referralCode,
        raffleNumber,
        shortCode,
        referralLink,
      })
      .returning();

    return NextResponse.json(
      {
        success: true,
        referral: newReferral[0],
      },
      { status: 201 }
    );
  } catch (error: unknown) {
    console.error('Error creating referral:', error);
    const err = error as { code?: string; message?: string };
    if (err.code === '42P01' || err.message?.includes('does not exist')) {
      return NextResponse.json(
        { error: 'Database not ready. Please try again later.' },
        { status: 503 }
      );
    }
    if (err.code === '23505') {
      return NextResponse.json(
        { error: 'A referral with similar details already exists. Please try again.' },
        { status: 409 }
      );
    }
    return NextResponse.json({ error: 'Failed to create referral' }, { status: 500 });
  }
}

export { GET, POST };
