interface BookingData {
  room: string;
  roomName: string;
  roomType: string;
  price: number;
  date?: string;
  name: string;
  phone: string;
  timestamp: string;
}

interface NotionPageResponse {
  success: boolean;
  data?: any;
  error?: string;
}

export async function createNotionPage(bookingData: BookingData): Promise<NotionPageResponse> {
  const token = process.env.NOTION_TOKEN;
  const databaseId = process.env.NOTION_DATABASE_ID;

  if (!token) {
    return { success: false, error: 'Notion token not configured' };
  }

  if (!databaseId) {
    return { success: false, error: 'Notion database ID not configured' };
  }

  try {
    const response = await fetch('https://api.notion.com/v1/pages', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
        'Notion-Version': '2022-06-28',
      },
      body: JSON.stringify({
        parent: {
          database_id: databaseId,
        },
        properties: {
          'Booking Name': {
            title: [
              {
                text: {
                  content: `${bookingData.roomName} - ${bookingData.name}`,
                },
              },
            ],
          },
          'Student Name': {
            rich_text: [
              {
                text: {
                  content: bookingData.name,
                },
              },
            ],
          },
          'Phone Number': {
            phone_number: bookingData.phone,
          },
          'Room Number': {
            rich_text: [
              {
                text: {
                  content: bookingData.roomName,
                },
              },
            ],
          },
          'Check-in Date': {
            date: bookingData.date ? { start: bookingData.date } : null,
          },
          'Status': {
            status: {
              name: 'Pending',
            },
          },
          'Monthly Rent': {
            number: bookingData.price,
          },
          'Security Deposit': {
            number: bookingData.price, // Same as monthly rent
          },
        },
      }),
    });

    if (!response.ok) {
      const error = await response.json();
      console.error('Notion API error:', error);
      return { success: false, error: JSON.stringify(error) };
    }

    const data = await response.json();
    return { success: true, data };
  } catch (error) {
    console.error('Error creating Notion page:', error);
    return { success: false, error: error instanceof Error ? error.message : 'Unknown error' };
  }
}

export async function testNotionConnection(): Promise<NotionPageResponse> {
  const token = process.env.NOTION_TOKEN;
  const databaseId = process.env.NOTION_DATABASE_ID;

  if (!token) {
    return { success: false, error: 'Notion token not configured' };
  }

  if (!databaseId) {
    return { success: false, error: 'Notion database ID not configured' };
  }

  try {
    const response = await fetch(`https://api.notion.com/v1/databases/${databaseId}`, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Notion-Version': '2022-06-28',
      },
    });

    if (!response.ok) {
      const error = await response.json();
      return { success: false, error: JSON.stringify(error) };
    }

    const data = await response.json();
    return { success: true, data };
  } catch (error) {
    console.error('Error testing Notion connection:', error);
    return { success: false, error: error instanceof Error ? error.message : 'Unknown error' };
  }
}
