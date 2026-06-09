// Helper wrapper for API routes with session
import { cookies } from 'next/headers';
import { getIronSession } from 'iron-session';
import { sessionOptions, SessionData } from '@/lib/sessions';

export async function getSession(): Promise<SessionData> {
  const cookieStore = await cookies();
  return await getIronSession(cookieStore, sessionOptions);
}

export async function saveSession(sessionData: SessionData): Promise<void> {
  const cookieStore = await cookies();
  const session = await getIronSession(cookieStore, sessionOptions);
  Object.assign(session, sessionData);
  await session.save();
}
