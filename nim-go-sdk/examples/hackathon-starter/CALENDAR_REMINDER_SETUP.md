# Google Calendar Reminder Setup

## Overview

The `create_calendar_reminder` tool creates calendar reminders for periodic investment goals. It works in **two modes**:

1. **Local Mode** (default): Stores reminders locally without Google Calendar integration
2. **Google Calendar Mode**: Syncs reminders with your Google Calendar (requires credentials)

## Quick Start (Local Mode)

The tool works out of the box without any setup! Users can create reminders through the chatbot:

```
User: "Set up weekly reminders to invest $25 USDC"
Nim: [Creates 12 weekly reminders and stores them locally]
```

## Google Calendar Integration (Optional)

To sync reminders with Google Calendar, follow these steps:

### 1. Create Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing one
3. Enable the **Google Calendar API**:
   - Navigate to "APIs & Services" > "Library"
   - Search for "Google Calendar API"
   - Click "Enable"

### 2. Create Service Account Credentials

1. Go to "APIs & Services" > "Credentials"
2. Click "Create Credentials" > "Service Account"
3. Fill in the details and click "Create"
4. Grant "Editor" role
5. Click "Done"
6. Click on the created service account
7. Go to "Keys" tab
8. Click "Add Key" > "Create New Key"
9. Choose "JSON" format
10. Download the credentials file

### 3. Configure Environment Variable

Add the path to your credentials file in `.env`:

```bash
GOOGLE_CALENDAR_CREDENTIALS=/path/to/your/credentials.json
```

### 4. Grant Calendar Access

1. Open the downloaded JSON file
2. Copy the `client_email` value (looks like: `xxx@xxx.iam.gserviceaccount.com`)
3. Open [Google Calendar](https://calendar.google.com)
4. Click on "Settings" > "Settings for my calendars" > Select your calendar
5. Scroll to "Share with specific people"
6. Click "Add people"
7. Paste the service account email
8. Grant "Make changes to events" permission
9. Click "Send"

## Tool Usage

### Parameters

- `frequency` (required): "weekly", "bi-weekly", or "monthly"
- `amount` (required): Amount to invest per period
- `currency` (optional): Default is "USDC"
- `start_date` (optional): YYYY-MM-DD format, defaults to next Monday
- `duration` (optional): Number of reminders to create
  - Weekly: 12 (3 months)
  - Bi-weekly: 6 (3 months)
  - Monthly: 3 (3 months)

### Example Conversations

**Weekly Reminders:**
```
User: "Create weekly investment reminders for $50 USDC"
Nim: [Creates 12 weekly reminders starting next Monday]
```

**Custom Schedule:**
```
User: "Set up bi-weekly reminders to invest $100 EURC starting March 1st"
Nim: [Creates 6 bi-weekly reminders starting March 1, 2026]
```

**Monthly Reminders:**
```
User: "I want monthly reminders to invest $200 USDC for 6 months"
Nim: [Creates 6 monthly reminders]
```

## How It Works

### Confirmation Flow

The tool requires user confirmation (like send_money):

1. **User requests reminder**: "Set up weekly reminders for $25"
2. **Nim detects intent**: Uses `create_calendar_reminder` tool
3. **User confirmation required**: Shows preview of reminders
4. **User approves**: Reminders are created
5. **Confirmation message**: Shows next reminder date

### Calendar Event Details

Each reminder includes:
- **Title**: "Weekly/Bi-Weekly/Monthly Investment Reminder"
- **Description**: "Time to invest $X.XX CURRENCY into your savings vault. Stay on track with your financial goals!"
- **Time**: 9:00 AM
- **Reminders**: 
  - Popup notification at event time
  - Email notification 1 hour before

### Storage

- **Local Mode**: Reminders stored in-memory (survives until server restart)
- **Google Calendar Mode**: Reminders synced to your Google Calendar (persistent)

## Troubleshooting

### "GOOGLE_CALENDAR_CREDENTIALS not set"

This is normal! The tool works in local mode without credentials. To enable Google Calendar sync, follow the setup steps above.

### "unable to create Calendar service"

Check that:
1. Credentials file path is correct in `.env`
2. Credentials file is valid JSON
3. Service account has calendar access

### "unable to create event"

Check that:
1. Service account email is shared with your calendar
2. Service account has "Make changes to events" permission
3. Calendar API is enabled in Google Cloud Console

## Integration with LangGraph

The reminder tool integrates with the financial workflow:

```
User: "I want to save"
    ↓
financial_help (checks balance)
    ↓
financial_save (compares vault rates)
    ↓
investment_reminder (offers calendar reminders)
    ↓
[User chooses frequency]
    ↓
create_calendar_reminder (creates events with confirmation)
```

## Security Notes

- **Never commit credentials to git**: Add `credentials.json` to `.gitignore`
- **Use environment variables**: Store paths in `.env` file
- **Limit service account scope**: Only grant calendar access
- **Rotate credentials**: Periodically regenerate service account keys

## Advanced Usage

### Custom Timezone

Edit the `createGoogleCalendarEvents` function in main.go to change timezone:

```go
TimeZone: "America/Los_Angeles",  // Change this
```

### Custom Reminder Times

Edit the `createCalendarEvents` function in main.go:

```go
"time": "14:00",  // Change to 2 PM
```

### Custom Duration Defaults

Edit the tool handler in main.go:

```go
case "weekly":
    params.Duration = 24  // 6 months instead of 3
```

## Future Enhancements

Potential improvements:
- Database storage for reminder persistence
- Email/SMS notifications without Google Calendar
- Reminder management (view, edit, delete)
- Multiple reminder frequencies per user
- Webhook integrations for other calendar services
