import { SessionOptions } from 'iron-session';

export const sessionOptions: SessionOptions = {
  password: process.env.SESSION_SECRET || 'fallback-secret-key-change-in-production-min-32-chars',
  cookieName: 'vatsapartment_session',
  cookieOptions: {
    secure: process.env.NODE_ENV === 'production',
    httpOnly: true,
    sameSite: 'lax',
    maxAge: 60 * 60 * 24 * 30, // 30 days
  },
};

export interface SessionData {
  user?: {
    id: string;
    username: string;
    role: string | null;
  };
  tenant?: {
    id: string;
    name: string;
    username: string;
  };
  isLoggedIn: boolean;
  userType?: 'admin' | 'tenant';
}
