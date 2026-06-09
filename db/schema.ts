import { pgTable, text, integer, serial, boolean, timestamp, doublePrecision } from 'drizzle-orm/pg-core';
import { sql } from 'drizzle-orm';

// Users table - Admin authentication
export const users = pgTable('users', {
  id: text('id').primaryKey(),
  username: text('username').notNull().unique(),
  passwordHash: text('password_hash').notNull(),
  role: text('role').default('admin'),
  createdAt: timestamp('created_at').default(sql`CURRENT_TIMESTAMP`),
  updatedAt: timestamp('updated_at').default(sql`CURRENT_TIMESTAMP`),
});

// Rooms table
export const rooms = pgTable('rooms', {
  id: text('id').primaryKey(),
  roomNumber: text('room_number').notNull(), // Room number like "101", "102", etc.
  floor: integer('floor').notNull(),
  type: text('type').notNull(),
  price: doublePrecision('price').notNull(),
  images: text('images'), // JSON string: ["url1", "url2"]
  createdAt: timestamp('created_at').default(sql`CURRENT_TIMESTAMP`),
  updatedAt: timestamp('updated_at').default(sql`CURRENT_TIMESTAMP`),
});

// Tenants table
export const tenants = pgTable('tenants', {
  id: text('id').primaryKey(),
  name: text('name').notNull(),
  email: text('email').unique(),
  phone: text('phone').notNull().unique(),
  aadhaarNumber: text('aadhaar_number'),
  alternateContact: text('alternate_contact'),
  emergencyContact: text('emergency_contact'),
  address: text('address'),
  status: text('status').default('active'), // active, inactive, moved_out
  agreementDocument: text('agreement_document'), // URL to rental agreement PDF
  documentPhoto: text('document_photo'), // URL to ID document photo
  collegeId: text('college_id'), // URL to college ID document (optional)
  checkInDate: timestamp('check_in_date'), // Date when tenant moved in
  securityDeposit: doublePrecision('security_deposit'), // Security deposit amount
  securityLockInPeriod: integer('security_lock_in_period'), // Lock-in period in months
  createdAt: timestamp('created_at').default(sql`CURRENT_TIMESTAMP`),
  updatedAt: timestamp('updated_at').default(sql`CURRENT_TIMESTAMP`),
});

// Payments table
export const payments = pgTable('payments', {
  id: text('id').primaryKey(),
  tenantId: text('tenant_id').notNull().references(() => tenants.id),
  amount: doublePrecision('amount').notNull(),
  paymentDate: timestamp('payment_date').notNull(),
  paymentMethod: text('payment_method'), // cash, upi, bank_transfer
  transactionId: text('transaction_id').unique(),
  status: text('status').default('pending'), // pending, completed, failed, refunded
  lateFee: doublePrecision('late_fee').default(0),
  monthCovered: text('month_covered'), // YYYY-MM format
  notes: text('notes'),
  receiptUrl: text('receipt_url'), // URL to receipt PDF
  createdAt: timestamp('created_at').default(sql`CURRENT_TIMESTAMP`),
  updatedAt: timestamp('updated_at').default(sql`CURRENT_TIMESTAMP`),
});

// Payment reminders table
export const paymentReminders = pgTable('payment_reminders', {
  id: text('id').primaryKey(),
  tenantId: text('tenant_id').notNull().references(() => tenants.id),
  dueDate: timestamp('due_date').notNull(),
  reminderDate: timestamp('reminder_date').notNull(),
  status: text('status').default('pending'), // pending, sent, paid
  reminderMethod: text('reminder_method'), // whatsapp, email, sms
  sentAt: timestamp('sent_at'),
  message: text('message'), // Stored message content
  createdAt: timestamp('created_at').default(sql`CURRENT_TIMESTAMP`),
});

// Settings table
export const settings = pgTable('settings', {
  key: text('key').primaryKey(),
  value: text('value').notNull(),
  description: text('description'),
  updatedAt: timestamp('updated_at').default(sql`CURRENT_TIMESTAMP`),
});

// Electricity bills table
export const electricityBills = pgTable('electricity_bills', {
  id: text('id').primaryKey(),
  tenantId: text('tenant_id').notNull().references(() => tenants.id),
  roomId: text('room_id').notNull().references(() => rooms.id),
  meterId: text('meter_id').references(() => meters.id),
  billingPeriod: text('billing_period').notNull(), // YYYY-MM format
  previousReading: integer('previous_reading').notNull(),
  currentReading: integer('current_reading').notNull(),
  unitsConsumed: integer('units_consumed').notNull(),
  ratePerUnit: doublePrecision('rate_per_unit').notNull(),
  totalAmount: doublePrecision('total_amount').notNull(),
  dueDate: timestamp('due_date').notNull(),
  paidDate: timestamp('paid_date'),
  status: text('status').default('pending'), // pending, paid, overdue
  meterReadingPhoto: text('meter_reading_photo'), // URL to photo
  notes: text('notes'),
  createdAt: timestamp('created_at').default(sql`CURRENT_TIMESTAMP`),
  paidAt: timestamp('paid_at'),
});

// Tenant portal credentials table - for tenant self-service portal access
export const tenantPortalCredentials = pgTable('tenant_portal_credentials', {
  id: text('id').primaryKey(),
  tenantId: text('tenant_id').notNull().unique().references(() => tenants.id),
  username: text('username').notNull().unique(),
  passwordHash: text('password_hash').notNull(),
  lastLogin: timestamp('last_login'),
  isActive: boolean('is_active').default(true),
  createdAt: timestamp('created_at').default(sql`CURRENT_TIMESTAMP`),
  updatedAt: timestamp('updated_at').default(sql`CURRENT_TIMESTAMP`),
});

// Tenant payments table - for self-reported payments that need admin verification
export const tenantPayments = pgTable('tenant_payments', {
  id: text('id').primaryKey(),
  tenantId: text('tenant_id').notNull().references(() => tenants.id),
  roomId: text('room_id').notNull().references(() => rooms.id),
  amount: doublePrecision('amount').notNull(),
  paymentMethod: text('payment_method').notNull(), // 'upi', 'bank_transfer'
  transactionId: text('transaction_id'),
  paymentDate: timestamp('payment_date').notNull().default(sql`CURRENT_TIMESTAMP`),
  status: text('status').default('pending'), // 'pending', 'verified', 'rejected'
  monthCovered: text('month_covered'), // YYYY-MM format
  screenshotUrl: text('screenshot_url'), // URL to payment screenshot
  notes: text('notes'),
  verifiedBy: text('verified_by').references(() => users.id),
  verifiedAt: timestamp('verified_at'),
  rejectionReason: text('rejection_reason'),
  createdAt: timestamp('created_at').default(sql`CURRENT_TIMESTAMP`),
  updatedAt: timestamp('updated_at').default(sql`CURRENT_TIMESTAMP`),
});

// Bookings table - Room bookings with multiple people
export const bookings = pgTable('bookings', {
  id: text('id').primaryKey(),
  roomId: text('room_id').notNull().references(() => rooms.id),
  rentAmount: doublePrecision('rent_amount').notNull(), // Total rent for the booking
  securityDeposit: doublePrecision('security_deposit'), // Security deposit amount
  securityLockInPeriod: integer('security_lock_in_period'), // Lock-in period in months
  checkInDate: timestamp('check_in_date').notNull(), // Check-in date
  checkOutDate: timestamp('check_out_date'), // Check-out date (null if active)
  status: text('status').default('active'), // active, completed, cancelled
  notes: text('notes'), // Additional notes about the booking
  createdAt: timestamp('created_at').default(sql`CURRENT_TIMESTAMP`),
  updatedAt: timestamp('updated_at').default(sql`CURRENT_TIMESTAMP`),
});

// People table - Individuals associated with bookings
export const people = pgTable('people', {
  id: text('id').primaryKey(),
  bookingId: text('booking_id').notNull().references(() => bookings.id),
  name: text('name').notNull(),
  phone: text('phone').notNull(),
  email: text('email'),
  aadhaarNumber: text('aadhaar_number'),
  documentPhoto: text('document_photo'), // URL to ID document photo
  collegeId: text('college_id'), // URL to college ID document (optional)
  isPrimary: boolean('is_primary').default(false), // Primary contact for the booking
  createdAt: timestamp('created_at').default(sql`CURRENT_TIMESTAMP`),
  updatedAt: timestamp('updated_at').default(sql`CURRENT_TIMESTAMP`),
});

// Room assignments table - Links tenants to rooms with rent details
export const roomAssignments = pgTable('room_assignments', {
  id: text('id').primaryKey(),
  tenantId: text('tenant_id').notNull().references(() => tenants.id),
  roomId: text('room_id').notNull().references(() => rooms.id),
  rentAmount: doublePrecision('rent_amount').notNull(), // Rent amount for this tenant
  startDate: timestamp('start_date').notNull(), // When tenant moved into this room
  endDate: timestamp('end_date'), // When tenant moved out (null if current)
  isActive: boolean('is_active').default(true), // Whether this assignment is currently active
  sharePercentage: doublePrecision('share_percentage'), // For shared rooms - percentage of rent this tenant pays
  notes: text('notes'), // Additional notes about the assignment
  createdAt: timestamp('created_at').default(sql`CURRENT_TIMESTAMP`),
  updatedAt: timestamp('updated_at').default(sql`CURRENT_TIMESTAMP`),
});

// Meters table - Support for multiple meter types per room
export const meters = pgTable('meters', {
  id: text('id').primaryKey(),
  roomId: text('room_id').notNull().references(() => rooms.id),
  meterType: text('meter_type').notNull(), // electric, inverter, solar, water, gas
  meterNumber: text('meter_number').notNull(), // Unique meter number/identifier
  meterName: text('meter_name'), // Custom name for the meter (e.g., "Main Electric Meter")
  initialReading: integer('initial_reading').default(0), // Initial meter reading
  currentReading: integer('current_reading').default(0), // Latest meter reading
  installationDate: timestamp('installation_date'), // When meter was installed
  isActive: boolean('is_active').default(true), // Whether meter is currently in use
  notes: text('notes'), // Additional notes about the meter
  createdAt: timestamp('created_at').default(sql`CURRENT_TIMESTAMP`),
  updatedAt: timestamp('updated_at').default(sql`CURRENT_TIMESTAMP`),
});

// Meter readings table - History of meter readings with AI extraction support
export const meterReadings = pgTable('meter_readings', {
  id: text('id').primaryKey(),
  meterId: text('meter_id').notNull().references(() => meters.id),
  reading: integer('reading').notNull(),
  readingDate: timestamp('reading_date').notNull(),
  photoUrl: text('photo_url'),
  extractedByAI: boolean('extracted_by_ai').default(false),
  aiProvider: text('ai_provider'), // deepseek, claude, openai
  aiConfidence: text('ai_confidence'), // high, medium, low
  billingPeriod: text('billing_period'), // YYYY-MM
  notes: text('notes'),
  createdAt: timestamp('created_at').default(sql`CURRENT_TIMESTAMP`),
});

// Referrals table - Referral program for marketing
export const referrals = pgTable('referrals', {
  id: text('id').primaryKey(),
  name: text('name'),
  email: text('email').notNull(),
  phone: text('phone').notNull(),
  whatsapp: text('whatsapp'),
  instagramHandle: text('instagram_handle'),
  referralCode: text('referral_code').notNull().unique(),
  raffleNumber: text('raffle_number').notNull().unique(),
  shortCode: text('short_code').notNull().unique(),
  referralLink: text('referral_link').notNull(),
  successfulReferrals: integer('successful_referrals').default(0),
  totalEarned: doublePrecision('total_earned').default(0),
  createdAt: timestamp('created_at').default(sql`CURRENT_TIMESTAMP`),
  updatedAt: timestamp('updated_at').default(sql`CURRENT_TIMESTAMP`),
});

// Referral usage tracking
export const referralUsages = pgTable('referral_usages', {
  id: text('id').primaryKey(),
  referralId: text('referral_id').notNull().references(() => referrals.id),
  bookingId: text('booking_id'),
  referrerName: text('referrer_name'),
  referrerPhone: text('referrer_phone'),
  discountApplied: doublePrecision('discount_applied').default(2000),
  rewardEarned: doublePrecision('reward_earned').default(4000),
  status: text('status').default('pending'),
  createdAt: timestamp('created_at').default(sql`CURRENT_TIMESTAMP`),
  updatedAt: timestamp('updated_at').default(sql`CURRENT_TIMESTAMP`),
});

// Password reset tokens table - For tenant password setup/reset
export const passwordResetTokens = pgTable('password_reset_tokens', {
  id: text('id').primaryKey(),
  tenantId: text('tenant_id').notNull().references(() => tenants.id),
  token: text('token').notNull().unique(),
  expiresAt: timestamp('expires_at').notNull(),
  usedAt: timestamp('used_at'),
  createdAt: timestamp('created_at').default(sql`CURRENT_TIMESTAMP`),
});
