'use client';

import { useState, useEffect } from 'react';
import { ChevronDown } from 'lucide-react';
import { DatePicker } from '@/components/ui/date-picker';

// Skeleton loader component for room cards
function RoomCardSkeleton() {
  return (
    <div className="animate-pulse bg-surface-container-low rounded-lg p-4 md:p-5 h-24 border border-outline-variant">
      <div className="flex flex-col md:flex-row md:justify-between md:items-start w-full gap-2 md:gap-3">
        <div className="flex flex-col gap-2">
          <div className="h-5 w-20 bg-surface-variant rounded" />
          <div className="h-4 w-24 bg-surface-variant/70 rounded" />
        </div>
        <div className="h-6 w-16 bg-surface-variant rounded-full" />
      </div>
    </div>
  );
}

interface Room {
  id: string;
  roomNumber: string;
  floor: number;
  type: string;
  price: number;
}

interface Booking {
  id: string;
  roomId: string;
  status: string;
}

interface RoomAssignment {
  id: string;
  roomId: string;
  isActive: boolean;
}

// Complete building structure - all possible room slots per floor
const BUILDING_STRUCTURE: Record<number, { id: string; type: string; price: number }[]> = {
  1: [
    { id: '1A', type: '2 Bed, 1 Bath', price: 10000 },
    { id: '1B', type: '1 Bed, 1 Bath', price: 9000 },
  ],
  2: [
    { id: '2A', type: '2 Bed, 1 Bath', price: 10000 },
    { id: '2B', type: '1 Bed, 1 Bath', price: 9000 },
  ],
  3: [
    { id: '3A', type: '2 Bed, 1 Bath', price: 10000 },
    { id: '3B', type: '1 Bed, 1 Bath', price: 9000 },
  ],
  4: [
    { id: 'PH', type: 'Penthouse', price: 15000 },
  ],
};

export default function BookNowPage() {
  const [rooms, setRooms] = useState<Room[]>([]);
  const [bookings, setBookings] = useState<Booking[]>([]);
  const [roomAssignments, setRoomAssignments] = useState<RoomAssignment[]>([]);
  const [loading, setLoading] = useState(true);
  const [initialLoadComplete, setInitialLoadComplete] = useState(false);
  const [isUpdating, setIsUpdating] = useState(false);
  const [formData, setFormData] = useState({
    room: '',
    date: undefined as Date | undefined,
    name: '',
    phone: '',
  });

  useEffect(() => {
    // Load cached data first if available
    const loadCachedData = () => {
      try {
        const cachedRooms = localStorage.getItem('cached_rooms');
        const cachedBookings = localStorage.getItem('cached_bookings');
        const cachedAssignments = localStorage.getItem('cached_assignments');

        if (cachedRooms) {
          const rooms = JSON.parse(cachedRooms);
          setRooms(rooms);
        }
        if (cachedBookings) setBookings(JSON.parse(cachedBookings));
        if (cachedAssignments) setRoomAssignments(JSON.parse(cachedAssignments));

        // If we have cached rooms, show them immediately
        if (cachedRooms) {
          setLoading(false);
        }
      } catch (error) {
        console.error('Error loading cached data:', error);
      }
    };

    loadCachedData();

    // Then fetch fresh data
    setIsUpdating(true);
    Promise.all([fetchRooms(), fetchBookings(), fetchRoomAssignments()])
      .then(() => {
        setInitialLoadComplete(true);
        setIsUpdating(false);
      });
  }, []);

  const fetchRooms = async () => {
    try {
      const response = await fetch('/api/rooms');
      const data = await response.json();
      const rooms = data.rooms || [];
      setRooms(rooms);

      // Cache the data
      if (rooms.length > 0) {
        localStorage.setItem('cached_rooms', JSON.stringify(rooms));
      }
    } catch (error) {
      console.error('Error fetching rooms:', error);
    }
  };

  const fetchBookings = async () => {
    try {
      const response = await fetch('/api/booking-management?active=true');
      const data = await response.json();
      const bookings = data.bookings || [];
      setBookings(bookings);

      // Cache the data
      localStorage.setItem('cached_bookings', JSON.stringify(bookings));
    } catch (error) {
      console.error('Error fetching bookings:', error);
    }
  };

  const fetchRoomAssignments = async () => {
    try {
      const response = await fetch('/api/room-assignments?active=true');
      const data = await response.json();
      const assignments = data.assignments || [];
      setRoomAssignments(assignments);

      // Cache the data
      localStorage.setItem('cached_assignments', JSON.stringify(assignments));
    } catch (error) {
      console.error('Error fetching room assignments:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    const selectedRoom = rooms.find((r) => r.id === formData.room);
    const roomPrice = selectedRoom?.price || 9000;

    // Send to Notion API
    try {
      const response = await fetch('/api/bookings', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          room: formData.room,
          roomName: selectedRoom?.id,
          roomType: selectedRoom?.type,
          price: roomPrice,
          date: formData.date?.toISOString(),
          name: formData.name,
          phone: formData.phone,
        }),
      });

      const result = await response.json();
    } catch (error) {
      console.error('Error sending booking to API:', error);
    }

    // Still open WhatsApp
    const message = encodeURIComponent(
      `Hi, I'd like to book a room at Vats Apartment.

Room: ${selectedRoom?.id}
Type: ${selectedRoom?.type}
Starting Date: ${formData.date ? formData.date.toLocaleDateString() : 'Not specified'}

My Details:
Name: ${formData.name}
Phone: ${formData.phone}

Please confirm availability and next steps.`
    );

    window.open(`https://wa.me/919992937447?text=${message}`, '_blank');
  };

  const handleQuickBook = (roomId: string) => {
    setFormData((prev) => ({ ...prev, room: roomId }));
  };

  const getFloorLabel = (floor: number) => {
    if (floor === 4) return 'ROOFTOP';
    return `FLOOR ${floor}`;
  };

  const getRoomName = (room: Room) => {
    if (room.floor === 4) return 'Rooftop - ' + room.type;
    return `${room.id} - ${room.type}`;
  };

  const isRoomBooked = (roomId: string, roomType: string) => {
    // Check if room has ANY active booking or assignment
    const hasActiveBooking = bookings.some(booking => booking.roomId === roomId && booking.status === 'active');
    const hasActiveAssignment = roomAssignments.some(assignment => assignment.roomId === roomId && assignment.isActive === true);

    // Room is unavailable if it has any active booking or assignment
    return hasActiveBooking || hasActiveAssignment;
  };

  if (loading) {
    return (
      <>
        <div className="container-max mx-auto px-4 md:px-6 lg:px-12 mt-4 md:mt-6 mb-6 md:mb-8">
          <section className="w-full h-[140px] md:h-[220px] rounded-xl overflow-hidden bg-gradient-to-br from-primary/20 to-secondary/20 flex items-center justify-center px-4">
            <div className="text-center">
              <h1 className="text-2xl md:text-5xl font-bold tracking-tight text-on-background">Book Your Space</h1>
            </div>
          </section>
        </div>
        <main className="container-max mx-auto gutter pb-24 md:pb-16">
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-3 md:gap-6 items-start">
            {/* Building Visual Skeleton */}
            <div className="lg:col-span-7 bg-surface-container-lowest border border-outline-variant rounded-lg p-4 md:p-8 shadow-sm">
              <div className="h-6 md:h-8 bg-surface-variant rounded animate-pulse mb-4 md:mb-6 border-b border-surface-variant pb-3 md:pb-4" />

              <div className="flex flex-col-reverse gap-3 md:gap-6">
                {[1, 2, 3, 4].map((floorNum) => {
                  const floorSlots = BUILDING_STRUCTURE[floorNum] || [];
                  return (
                    <div key={floorNum} className="flex flex-col gap-2">
                      <div className={`grid gap-2 md:gap-4 ${floorNum === 4 ? 'grid-cols-1' : 'grid-cols-2'}`}>
                        {floorSlots.map((slot) => (
                          <RoomCardSkeleton key={slot.id} />
                        ))}
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>

            {/* Booking Form Skeleton */}
            <div className="lg:col-span-5 bg-surface-container-lowest border border-outline-variant rounded-lg p-4 md:p-8 shadow-sm lg:sticky lg:top-24">
              <div className="space-y-4 md:space-y-5">
                {/* Room Selection Skeleton */}
                <div>
                  <div className="h-3 md:h-4 bg-surface-variant rounded animate-pulse mb-2 w-24 md:w-32" />
                  <div className="h-10 md:h-12 bg-surface-container-low border border-outline-variant rounded-lg animate-pulse" />
                </div>

                {/* Starting Date Skeleton */}
                <div>
                  <div className="h-3 md:h-4 bg-surface-variant rounded animate-pulse mb-2 w-24 md:w-32" />
                  <div className="h-10 md:h-12 bg-surface-container-low border border-outline-variant rounded-lg animate-pulse" />
                </div>

                <div className="border-t border-surface-variant my-4 md:my-5" />

                {/* Contact Details Skeleton */}
                <div className="space-y-3 md:space-y-4">
                  <div>
                    <div className="h-3 md:h-4 bg-surface-variant rounded animate-pulse mb-2 w-20 md:w-24" />
                    <div className="h-10 md:h-12 bg-surface-container-low border border-outline-variant rounded-lg animate-pulse" />
                  </div>
                  <div>
                    <div className="h-3 md:h-4 bg-surface-variant rounded animate-pulse mb-2 w-24 md:w-32" />
                    <div className="h-10 md:h-12 bg-surface-container-low border border-outline-variant rounded-lg animate-pulse" />
                  </div>
                </div>

                {/* Submit Button Skeleton */}
                <div className="pt-2">
                  <div className="h-10 md:h-12 bg-primary/50 rounded-lg animate-pulse" />
                  <div className="h-3 md:h-4 bg-surface-variant/50 rounded animate-pulse mt-3 mx-auto w-40 md:w-48" />
                </div>
              </div>
            </div>
          </div>
        </main>
      </>
    );
  }

  return (
    <>
      {/* Thin Hero Section */}
      <div className="container-max mx-auto px-4 md:px-6 lg:px-12 mt-4 md:mt-6 mb-6 md:mb-8">
        <section className="w-full h-[140px] md:h-[220px] rounded-xl overflow-hidden bg-gradient-to-br from-primary/20 to-secondary/20 flex items-center justify-center px-4">
          <div className="text-center">
            <h1 className="text-2xl md:text-5xl font-bold tracking-tight text-on-background mb-2 md:mb-3">Book Your Space</h1>
            <p className="text-xs md:text-base text-on-surface-variant max-w-2xl mx-auto">
              Select your preferred room and starting date. We'll confirm availability instantly.
            </p>
          </div>
        </section>
      </div>

      <main className="container-max mx-auto gutter pb-24 md:pb-16">
        {/* Main Layout */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-3 md:gap-6 items-start">
        {/* Building Visual */}
        <div className="lg:col-span-7 bg-surface-container-lowest border border-outline-variant rounded-lg p-4 md:p-8 shadow-sm">
          <div className="flex items-center justify-between mb-4 md:mb-6 border-b border-surface-variant pb-3 md:pb-4">
            <h2 className="font-h2 text-primary-container text-sm md:text-base">
              Building Layout
            </h2>
            {isUpdating && (
              <div className="flex items-center gap-2 text-xs md:text-sm text-on-surface-variant">
                <div className="animate-spin w-3 h-3 md:w-4 md:h-4 border-2 border-outline border-t-transparent rounded-full" />
                <span className="text-[10px] md:text-sm">Updating availability...</span>
              </div>
            )}
          </div>

          <div className="flex flex-col-reverse gap-3 md:gap-6">
            {[1, 2, 3, 4].map((floorNum) => {
              const floorSlots = BUILDING_STRUCTURE[floorNum] || [];

              return (
                <div key={floorNum} className="flex flex-col gap-2 relative">
                  <div className="absolute -left-16 top-1/2 -translate-y-1/2 font-label-caps text-outline rotate-[-90deg] hidden sm:block whitespace-nowrap origin-center">
                    {getFloorLabel(floorNum)}
                  </div>
                  <div className={`grid gap-2 md:gap-4 ${floorNum === 4 ? 'grid-cols-1' : 'grid-cols-2'}`}>
                    {floorSlots.map((slot) => {
                      const room = rooms.find((r) => r.id === slot.id);
                      const booked = room ? isRoomBooked(room.id, room.type) : true;
                      const displayPrice = room?.price || slot.price;
                      const displayType = room?.type || slot.type;
                      const roomExists = !!room;

                      return (
                        <button
                          key={slot.id}
                          onClick={() => !booked && roomExists && handleQuickBook(slot.id)}
                          disabled={booked || !roomExists}
                          className={`group relative rounded-lg p-3 md:p-5 text-left h-auto overflow-hidden transition-all ${
                            formData.room === slot.id
                              ? 'border-2 border-secondary bg-success-container/40 shadow-lg ring-2 ring-secondary/30'
                              : booked
                                ? 'bg-error-container/10 border-2 border-error/30 cursor-not-allowed'
                                : 'bg-secondary-container border-2 border-secondary/30 hover:border-secondary cursor-pointer'
                          }`}
                        >
                          {/* Status Badge - Top Right */}
                          <div className="absolute top-2 md:top-3 right-2 md:right-3 z-20">
                            <div className={`flex items-center gap-0.5 md:gap-1 px-1.5 md:px-3 py-0.5 md:py-1 rounded-md text-[8px] md:text-xs font-bold ${
                              booked
                                ? 'bg-error text-on-error shadow-sm'
                                : formData.room === slot.id
                                  ? 'bg-black/30 border border-black/30 text-on-secondary shadow-sm'
                                  : 'bg-secondary-container border border-secondary/40 text-on-surface shadow-sm'
                            }`}>
                              {booked ? (
                                <>
                                  <span className="w-1.5 h-1.5 md:w-2 md:h-2 rounded-full bg-current"></span>
                                  <span className="hidden md:inline">UNAVAILABLE</span>
                                  <span className="md:hidden">NA</span>
                                </>
                              ) : formData.room === slot.id ? (
                                <>
                                  <span className="w-1.5 h-1.5 md:w-2 md:h-2 rounded-full bg-current"></span>
                                  <span className="hidden md:inline">SELECTED</span>
                                  <span className="md:hidden">OK</span>
                                </>
                              ) : (
                                <>
                                  <span className="w-1.5 h-1.5 md:w-2 md:h-2 rounded-full bg-current"></span>
                                  <span className="hidden md:inline">AVAILABLE</span>
                                  <span className="md:hidden">OK</span>
                                </>
                              )}
                            </div>
                          </div>

                          {/* Room Info - Left Side */}
                          <div className="relative z-10 pr-16 md:pr-20">
                            <div>
                              <span className={`font-body-md md:font-h3 block text-xs md:text-lg font-bold ${
                                formData.room === slot.id
                                  ? 'text-white'
                                  : booked
                                    ? 'text-error'
                                    : 'text-on-surface'
                              }`}>
                                {slot.id}
                              </span>
                              <span className={`font-caption text-[9px] md:text-sm block ${
                                formData.room === slot.id
                                  ? 'text-white/80'
                                  : booked
                                    ? 'text-error/70'
                                    : 'text-on-surface-variant'
                              }`}>
                                {displayType}
                              </span>
                              {!booked && (
                                <div className={`mt-0.5 text-[8px] md:text-sm font-semibold ${
                                  formData.room === slot.id ? 'text-white/90' : 'text-on-surface'
                                }`}>
                                  ₹{displayPrice.toLocaleString()}/mo
                                </div>
                              )}
                            </div>
                          </div>
                        </button>
                      );
                    })}
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* Booking Form */}
        <div className="lg:col-span-5 bg-surface-container-lowest border border-outline-variant rounded-lg p-4 md:p-8 shadow-sm lg:sticky lg:top-24">
          <form onSubmit={handleSubmit} className="space-y-4 md:space-y-5">
            {/* Room Selection */}
            <div>
              <label className="block font-label-caps text-on-surface-variant mb-2 text-[10px] md:text-sm" htmlFor="room">
                SELECTED ROOM
              </label>
              <div className="relative">
                <select
                  id="room"
                  name="room"
                  value={formData.room}
                  onChange={handleInputChange}
                  className="block w-full pl-3 md:pl-4 pr-8 md:pr-10 py-2.5 md:py-4 text-sm md:text-lg font-body-md text-primary-container bg-surface-container-low border border-outline-variant focus:outline-none focus:ring-2 focus:ring-primary-container/50 focus:border-primary-container sm:text-sm rounded-lg appearance-none"
                  required
                >
                  <option value="" disabled>Select a room</option>
                  {Object.values(BUILDING_STRUCTURE).flat().map((slot) => {
                    const room = rooms.find((r) => r.id === slot.id);
                    const booked = room ? isRoomBooked(room.id, room.type) : true;
                    const displayPrice = room?.price || slot.price;
                    return (
                      <option
                        key={slot.id}
                        value={slot.id}
                        disabled={booked}
                      >
                        {slot.id} - {room?.type || slot.type} - ₹{displayPrice.toLocaleString()}/month {booked ? '(Unavailable)' : ''}
                      </option>
                    );
                  })}
                </select>
                <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center px-3 md:px-4 text-on-surface-variant">
                  <ChevronDown className="w-4 h-4 md:w-5 md:h-5" />
                </div>
              </div>
            </div>

            {/* Starting Date */}
            <div>
              <label className="block font-label-caps text-on-surface-variant mb-2 text-[10px] md:text-sm" htmlFor="date">
                STARTING DATE
              </label>
              <DatePicker
                value={formData.date}
                onChange={(date) => setFormData((prev) => ({ ...prev, date }))}
                placeholder="Select your starting date"
              />
            </div>

            <div className="border-t border-surface-variant my-4 md:my-5" />

            {/* Contact Details */}
            <div className="space-y-3 md:space-y-4">
              <div>
                <label className="block font-label-caps text-on-surface-variant mb-2 text-[10px] md:text-sm" htmlFor="name">
                  FULL NAME
                </label>
                <input
                  id="name"
                  name="name"
                  type="text"
                  placeholder="Enter your full name"
                  value={formData.name}
                  onChange={handleInputChange}
                  required
                  className="block w-full px-3 md:px-4 py-2.5 md:py-4 text-sm md:text-lg font-body-md text-primary-container bg-surface-container-low border border-outline-variant focus:outline-none focus:ring-2 focus:ring-primary-container/50 focus:border-primary-container sm:text-sm rounded-lg placeholder:text-outline"
                />
              </div>
              <div>
                <label className="block font-label-caps text-on-surface-variant mb-2 text-[10px] md:text-sm" htmlFor="phone">
                  PHONE NUMBER
                </label>
                <input
                  id="phone"
                  name="phone"
                  type="tel"
                  placeholder="+91 XXXXX XXXXX"
                  value={formData.phone}
                  onChange={handleInputChange}
                  required
                  className="block w-full px-3 md:px-4 py-2.5 md:py-4 text-sm md:text-lg font-body-md text-primary-container bg-surface-container-low border border-outline-variant focus:outline-none focus:ring-2 focus:ring-primary-container/50 focus:border-primary-container sm:text-sm rounded-lg placeholder:text-outline"
                />
              </div>
            </div>

            {/* Submit Button */}
            <div className="pt-2">
              <button
                type="submit"
                className="w-full flex justify-center py-3 md:py-4 px-6 border border-transparent rounded-lg shadow-sm text-xs md:text-lg font-medium text-white bg-primary hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary"
              >
                Confirm Booking
              </button>
              <p className="mt-3 text-center font-caption text-on-surface-variant text-[10px] md:text-xs">
                By confirming, you agree to our{' '}
                <a href="/terms" className="underline hover:text-primary-container">
                  Booking Terms
                </a>
                .
              </p>
            </div>
          </form>
        </div>
      </div>
    </main>
    </>
  );
}
