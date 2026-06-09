import { db } from './index';
import { settings } from './schema';

async function seed() {
  console.log('🌱 Seeding database...');

  try {
    // System settings
    const settingsData = [
      {
        key: 'late_fee_percentage',
        value: '10',
        description: 'Late fee percentage per month',
      },
      {
        key: 'grace_period_days',
        value: '5',
        description: 'Grace period before late fees apply',
      },
      {
        key: 'rent_due_day',
        value: '1',
        description: 'Day of month when rent is due',
      },
      {
        key: 'security_deposit_months',
        value: '1',
        description: 'Number of months for security deposit',
      },
      {
        key: 'currency',
        value: 'INR',
        description: 'Currency for payments',
      },
      {
        key: 'reminder_days_before',
        value: '3',
        description: 'Days before due date to send reminder',
      },
      {
        key: 'electricity_rate_per_unit',
        value: '8',
        description: 'Default electricity rate per unit in INR',
      },
      {
        key: 'wifi_ssid',
        value: 'VatsApartment',
        description: 'WiFi network name (shown to tenants)',
      },
      {
        key: 'wifi_password',
        value: 'Vats@2024',
        description: 'WiFi password (shown to tenants)',
      },
    ];

    for (const setting of settingsData) {
      await db.insert(settings).values({
        ...setting,
        updatedAt: new Date(),
      }).onConflictDoNothing();
      console.log(`✓ Set ${setting.key} = ${setting.value}`);
    }

    console.log('\n✅ Database seeded successfully');
  } catch (error) {
    console.error('❌ Error seeding database:', error);
    process.exit(1);
  }
}

seed();
