import { cookies } from 'next/headers';
import { getIronSession } from 'iron-session';
import { sessionOptions, SessionData } from '@/lib/sessions';
import { redirect } from 'next/navigation';

export async function getTenantSession() {
  const cookieStore = await cookies();
  const session = await getIronSession<SessionData>(cookieStore, sessionOptions);

  if (!session.isLoggedIn || session.userType !== 'tenant' || !session.tenant) {
    return null;
  }

  return session.tenant;
}

export async function requireTenantAuth() {
  const tenant = await getTenantSession();

  if (!tenant) {
    redirect('/tenant/login');
  }

  return tenant;
}
