import { rooms } from '@/db/schema';

// Dynamically import db only on server side
async function getDb() {
  if (typeof window !== 'undefined') return null;
  const { db } = await import('@/db');
  return db;
}

export interface Room {
  id: string;
  floor: number;
  type: string;
  price: number;
  tenant?: string;
  images?: string[];
}

const DEFAULT_ROOMS: Room[] = [
  { id: '1A', floor: 1, type: '2 Bed, 1 Bath', price: 10000 },
  { id: '1B', floor: 1, type: '1 Bed, 1 Bath', price: 9000 },
  { id: '2A', floor: 2, type: '2 Bed, 1 Bath', price: 10000 },
  { id: '2B', floor: 2, type: '1 Bed, 1 Bath', price: 9000 },
  { id: '3A', floor: 3, type: '2 Bed, 1 Bath', price: 10000 },
  { id: '3B', floor: 3, type: '1 Bed, 1 Bath', price: 9000 },
  { id: 'PH', floor: 4, type: 'Penthouse', price: 15000 },
];

const STORAGE_KEY = 'vatsapartment_rooms';

// Helper to parse images from JSON string
function parseImages(imagesJson: string | null): string[] {
  if (!imagesJson) return [];
  try {
    return JSON.parse(imagesJson);
  } catch {
    return [];
  }
}

/**
 * Get all rooms from database
 * Falls back to DEFAULT_ROOMS if database is not accessible (client-side during build)
 */
export async function getRooms(): Promise<Room[]> {
  // If running on client side and database might not be available, use localStorage fallback
  if (typeof window !== 'undefined') {
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored) {
        return JSON.parse(stored);
      }
    } catch {
      // Fall through to default rooms
    }
    return DEFAULT_ROOMS;
  }

  // Server-side: fetch from database
  try {
    const db = await getDb();
    if (!db) return DEFAULT_ROOMS;

    const roomData = await db.select().from(rooms).orderBy(rooms.floor, rooms.id);

    // Get active room assignments to check occupancy
    const { roomAssignments, tenants } = await import('@/db/schema');
    const { eq } = await import('drizzle-orm');

    const activeAssignments = await db
      .select({
        roomId: roomAssignments.roomId,
        tenantId: roomAssignments.tenantId,
      })
      .from(roomAssignments)
      .where(eq(roomAssignments.isActive, true));

    // Get tenant names for active assignments
    const tenantIds = activeAssignments.map(a => a.tenantId);
    const tenantData = tenantIds.length > 0
      ? await db.select().from(tenants).where(eq(tenants.id, tenantIds[0]))
      : [];

    const tenantMap = new Map(tenantData.map(t => [t.id, t.name]));

    // Create a map of room assignments
    const roomOccupancy = new Map<string, string>();
    activeAssignments.forEach(assignment => {
      const tenantName = tenantMap.get(assignment.tenantId);
      if (tenantName) {
        roomOccupancy.set(assignment.roomId, tenantName);
      }
    });

    return roomData.map((room) => ({
      id: room.id,
      floor: room.floor,
      type: room.type,
      price: room.price,
      images: parseImages(room.images),
      tenant: roomOccupancy.get(room.id),
    }));
  } catch (error) {
    console.error('Error fetching rooms from database:', error);
    return DEFAULT_ROOMS;
  }
}

/**
 * Save rooms to database
 * This is used by admin panel to update room information
 */
export async function saveRooms(roomsToUpdate: Room[]): Promise<void> {
  try {
    const db = await getDb();
    if (!db) {
      // Fallback to localStorage only
      if (typeof window !== 'undefined') {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(roomsToUpdate));
        window.dispatchEvent(new Event('rooms-updated'));
      }
      return;
    }

    for (const room of roomsToUpdate) {
      await db
        .insert(rooms)
        .values({
          id: room.id,
          roomNumber: room.id,
          floor: room.floor,
          type: room.type,
          price: room.price,
          images: JSON.stringify(room.images || []),
        })
        .onConflictDoUpdate({
          target: rooms.id,
          set: {
            floor: room.floor,
            type: room.type,
            price: room.price,
            images: JSON.stringify(room.images || []),
            updatedAt: new Date(),
          },
        });
    }

    // Update localStorage as fallback for client-side
    if (typeof window !== 'undefined') {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(roomsToUpdate));
      window.dispatchEvent(new Event('rooms-updated'));
    }
  } catch (error) {
    console.error('Error saving rooms to database:', error);
    throw error;
  }
}

/**
 * Update a single room
 */
export async function updateRoom(roomId: string, updates: Partial<Room>): Promise<Room[]> {
  const currentRooms = await getRooms();
  const index = currentRooms.findIndex((r) => r.id === roomId);

  if (index !== -1) {
    currentRooms[index] = { ...currentRooms[index], ...updates };
    await saveRooms(currentRooms);
  }

  return currentRooms;
}

/**
 * Get available rooms (excluding rooms with active tenant assignments)
 */
export async function getAvailableRooms(): Promise<Room[]> {
  // If running on client side, return all rooms
  if (typeof window !== 'undefined') {
    return getRooms();
  }

  // Server-side: fetch from database and filter out occupied rooms
  try {
    const db = await getDb();
    if (!db) return getRooms();

    const { roomAssignments } = await import('@/db/schema');
    const { eq, and } = await import('drizzle-orm');

    // Get all rooms
    const allRooms = await getRooms();

    // Get active room assignments
    const activeAssignments = await db
      .select({ roomId: roomAssignments.roomId })
      .from(roomAssignments)
      .where(eq(roomAssignments.isActive, true));

    const occupiedRoomIds = new Set(activeAssignments.map(a => a.roomId));

    // Filter out occupied rooms
    return allRooms.filter(room => !occupiedRoomIds.has(room.id));
  } catch (error) {
    console.error('Error fetching available rooms:', error);
    return getRooms();
  }
}

/**
 * Get minimum price across all rooms
 */
export async function getMinPrice(): Promise<number> {
  const allRooms = await getRooms();
  return Math.min(...allRooms.map((r) => r.price));
}

/**
 * Subscribe to room updates (client-side only)
 * For backward compatibility with existing components
 */
export function subscribe(callback: () => void): () => void {
  if (typeof window === 'undefined') return () => {};
  window.addEventListener('rooms-updated', callback);
  return () => window.removeEventListener('rooms-updated', callback);
}

// Legacy export for backward compatibility
export const roomsStorage = {
  getRooms: async () => getRooms(),
  saveRooms: async (rooms: Room[]) => saveRooms(rooms),
  updateRoom: async (roomId: string, updates: Partial<Room>) => updateRoom(roomId, updates),
  subscribe,
  getAvailableRooms: async () => getAvailableRooms(),
  getMinPrice: async () => getMinPrice(),
};
