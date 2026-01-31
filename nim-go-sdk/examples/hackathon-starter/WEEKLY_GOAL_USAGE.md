# Weekly Spending Goal Feature

## Overview
The weekly spending goal feature allows users to set spending limits with **user confirmation required** (like WRITE OPERATIONS) and track their progress throughout the week with a visual progress bar in the frontend. Uses actual transaction dates to calculate weekly spending accurately.

## Backend Tools

### 1. `spend_weekly_goal` (WRITE - Requires Confirmation)

**Purpose:** Set or update weekly spending goals

**Requires Confirmation:** ✅ Yes (like send_money, deposit_savings)

#### Features
- **Set Goals**: Extract amount and currency from natural language
- **Requires User Confirmation**: User must approve before goal is set
- **Track Progress**: Calculate spending from transaction history
- **Date-Aware**: Uses actual transaction timestamps to filter weekly spending
- **Weekly Reset**: Goals are based on Monday-Sunday weeks
- **Multi-Currency**: Support for USD, LIL, USDC, and other currencies

### 2. `get_weekly_spending_progress` (READ - No Confirmation)

**Purpose:** Check current weekly spending progress

Nim: I'll set your weekly spending goal to 5 LIL. 
     Please confirm to proceed.
     
     [User clicks Confirm]

Nim: Perfect! I've set your weekly spending goal to 5.00 LIL...
```

The tool extracts:
- Amount: `5`
- Currency: `LIL`

Other examples:
- "I want to limit my spending to 100 USD per week"
- "Set weekly budget of 50 USDC"
- "My weekly spending goal is 200 dollars"

### Checking Progress (No Confirmation)quires Confirmation)
```
User: "Set a weekly spend of 5 LIL this week"
```
The tool extracts:
- Amount: `5`
- Currency: `LIL`

Other examples:
- "I want to limit my spending to 100 USD per week"

Nim: [Immediately shows progress without confirmation]
     You've spent 2.35 LIL out of your 5.00 LIL weekly goal...
```

## Tool Parameters

### spend_weekly_goal
```json
{
  "amount": 5.0,           // The weekly spending limit
```
1. Frontend connects to WebSocket
2. Sends: { type: 'message', content: 'get_weekly_spending_progress' }
3. Backend executes tool (no confirmation needed)
4. Returns: { type: 'tool_result', data: { goal_amount: 5.0, ... } }
5. Frontend parses and displays progress bar
```

## Response Data Format
L",       // Currency code (USD, LIL, USDC, etc.)
  "action": "set"          // Always "set" for this tool
}
```

### get_weekly_spending_progress
```json
{
  // No parameters required
}
```

## How Transaction Date Filtering Works

###Confirmation Flow

### Setting a Goal (Requires Confirmation)
```
User: "Set weekly budget of 100 USD"
  ↓
Nim: "I'll set your weekly spending goal to $100.00 USD.
      Week: Monday, Jan 27 - Monday, Feb 3
      Current spending: $24.50
      
      Please confirm to proceed."
  ↓
User: [Clicks Confirm Button]
  ↓
Nim: "✅ Goal set! You've spent $24.50 so far..."
```

### Checking Progress (No Confirmation)
```
User: "How much have I spent?"
  ↓
Nim: [Immediately shows data]
     "You've spent $67.80 of your $100.00 USD weekly goal..."
```

##  Before (Inaccurate)
- Counted ALL "send" transactions regardless of date
- Could include spending from previous weeks or months

### After (Accurate) ✅
```go
// Parse transaction timestamp
var txTime time.Time
if timestamp, ok := tx["timestamp"].(string); ok {
    txTime, _ = time.Parse(time.RFC3339, timestamp)
} else if createdAt, ok := tx["created_at"].(string); ok {
    txTime, _ = time.Parse(time.RFC3339, createdAt)
} else if date, ok := tx["date"].(string); ok {
    txTime, _ = time.Parse("2006-01-02", date)
}

// Only count transactions from THIS WEEK
if !txTime.IsZero() && txTime.After(weekStart) && txTime.Before(weekEnd) {
    if txCurrency == currency || txCurrency == "" {
        weeklySpending += amount
    }
}
```

**Handles multiple date formats:**
- RFC3339: `"2026-01-31T12:00:00Z"`
- Date only: `"2026-01-31"`
- Checks `timestamp`, `created_at`, and `date` fields

## Frontend Progress Bar

### Fixed Issues
1. ✅ **Loading Forever**: Now uses dedicated read-only tool
2. ✅ **Better Error Handling**: Shows "No goal set" if no response
3. ✅ **Console Logging**: Debug WebSocket messages
4. ✅ **Auto-timeout**: Falls back after 5 seconds

### WebSocket Message Flow
{
  "amount": 5.0,           // The weekly spending limit
  "currency": "LIL",       // Currency code (USD, LIL, USDC, etc.)
  "action": "set"          // "set" to create/update, "get" to check progress
}
```

### Response Data
```json
{
  "goal_amount": 5.0,
  "currency": "LIL",
  "week_start": "Monday, Jan 27",
  "week_end": "Monday, Feb 3",
  "spent_so_far": 2.35,
  "remaining": 2.65,
  "percentage": 47.0,
  "on_track": true,
  "days_left": 4
}
```

## Frontend Progress Bar

### Features
- **Real-time Updates**: Connects via WebSocket to fetch latest data
- **Visual Progress**: Animated progress bar with percentage
- **Status Indicators**:
  - 🟢 **Green (On Track)**: Spending pace is normal for the week
  - 🟠 **Orange (Warning)**: Spending faster than expected
  - 🔴 **Red (Over Budget)**: Exceeded weekly limit
- **Responsive Design**: Works on mobile and desktop

### UI Components

1. **Goal Summary**
   - Weekly limit amount and currency
   - Days remaining in the week

2. **Progress Bar**
   - Visual percentage indicator
   - Color-coded based on status
   - Animated transitions

3. **Spending Details**
   - Amount spent so far
   - Remaining budget
   - Negative values shown in red if over budget

4. **Status Badge**
   - ✅ On Track
   - ⚠️ Watch Spending
   - ⚠️ Over Budget

### No Goal State
When no goal is set, shows a helpful message:
```
💰 Weekly Spending Goal
No goal set yet
Ask Nim: "Set a weekly spend of 5 LIL"
```

## How It Works

### Backend Logic
1. **Goal Storage**: Uses in-memory map (consider database for production)
2. **Week Calculation**: Monday = start of week, Sunday = end
3. **Transaction Analysis**: 
   - Fetches up to 100 recent transactions
   - Filters for "send" type (spending)
   - Matches currency if specified
   - Calculates total for current week
4. **Progress Calculation**:
   - Percentage = (spent / goal) × 100
   - On track check: compares actual vs. expected daily pace
  Expected output:
# ✅ Added custom spending analyzer and weekly goal tools
```

### Set a Goal (with confirmation)
```
User: Set a weekly spend of 10 USD
Nim: [Shows confirmation dialog]
User: [Confirms]
Nim: ✅ Goal set!
```

### Check Progress (no confirmation)
```
User: What's my weekly spending?
Nim: [Immediately shows data]
```

### Frontend
```bash
cd frontend
npm run dev

# Open http://localhost:5173
# Progress bar should appear after setting goal
```

## Troubleshooting

### Progress bar loading forever?an 27 - Monday, Feb 3).

     Current Status:
     • Spent so far: $24.50
     • Remaining: $75.50
     • Progress: 24.5%
     • Status: ✅ On track!

     You have 5 days left in the week. Keep up the great work!

---

Later in the week...

User: Am I still on track with my budget?

Nim: Let me check your weekly spending progress...

     You're doing great! Here's your status:
     • Goal: $100.00 USD
     • Spent: $67.80
     • Remaining: $32.20
     • Progress: 67.8%
     • Days left: 2

     ⚠️ Watch your spending - you're ahead of the expected pace,
     but still within budget!
```

## Customization

### Changing Week Start Day
Edit `getWeekStart()` function in main.go:
```go
// Change from Monday to Sunday
daysToSunday := int(t.Weekday())
sunday := t.AddDate(0, 0, -daysToSunday)
```

### Adjusting Progress Bar Colors
Edit styles.css:
```css
.progress-fill.on-track {
  background: linear-gradient(135deg, #your-color 0%, #your-color-dark 100%);
}
```

### Adding Notifications
Extend the tool to send alerts:
- When reaching 50%, 75%, 90% of budget
- Daily spending summaries
- End-of-week reports

## Production Considerations

1. **Persistence**: Replace in-memory map with database
2. **Multi-User**: Current implementation uses UserID as key
3. **Time Zones**: Consider user's local timezone
4. **Currency Conversion**: Add real-time exchange rates
5. **Historical Data**: Store past weeks for trends
6. **Notifications**: Push alerts for budget milestones
7. **Authentication**: Ensure JWT token validation

## Testing

### Backend
```bash
# Start the server
go run main.go

# Test via Nim chat or API
```

### Frontend
```bash
cd frontend
npm run dev

# Open http://localhost:5173
# Set a goal via Nim chat
# Progress bar should appear and update
```

## Troubleshooting

**Progress bar not showing?**
- Check WebSocket connection in browser console
- Verify backend is running on correct port
- Ensure goal is set: ask "What's my weekly spending goal?"

**Incorrect spending amounts?**
- Check transaction currency matches goal currency
- Verify transaction history is being fetched
- Confirm "send" type transactions are counted

**Colors not working?**
- Clear browser cache
- Check CSS is loading properly
- Verify percentage calculation is correct

## Future Enhancements

- [ ] Monthly/yearly goal options
- [ ] Category-specific budgets (food, transport, etc.)
- [ ] Savings goal integration
- [ ] Spending predictions based on history
- [ ] Shared goals with family/friends
- [ ] Gamification with achievements
- [ ] Export spending reports
- [ ] Smart recommendations for staying on track
