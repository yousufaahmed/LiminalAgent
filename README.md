# Liminal AI Financial Agent - Hackathon Starter

A complete financial management platform combining AI-powered conversation, real-time banking integration, intelligent receipt scanning with image upload, and advanced visualization. Built with Go backend, React frontend, and TabScanner API integration.

![Application Overview](examples/hackathon-starter/imgs/Screenshot%202026-02-01%20084736.png)

## ✨ Features

### 🏦 Financial Management
- **Real-time Banking Integration**: Connect to Liminal Banking APIs for live transaction data
- **Balance Monitoring**: Real-time account balance checking
- **Transaction History**: Fetch and categorize transaction history
- **Weekly Spending Goals**: Set budget limits with intelligent tracking
- **Spending Categories**: Automatic categorization (Food, Travel, Subscriptions, Shopping, etc.)
- **Smart Budget Analysis**: AI-powered spending pattern recommendations

### 🤖 AI Agent
- **Natural Language Interface**: Conversational banking assistant
- **Context-Aware Assistance**: Multi-turn conversations with memory
- **Intelligent Tool Selection**: Automatic routing to appropriate banking tools
- **Powered by Claude**: Anthropic Claude Sonnet 4 for advanced reasoning
- **Built on nim-go-sdk**: Production-ready orchestration framework
- **WebSocket Communication**: Real-time bidirectional updates

### 📸 Receipt Scanner & Image Upload
- **Custom Camera Button**: One-click image upload from frontend (📷 icon)
- **Image Preview**: Thumbnail preview with "✅ Image Ready" indicator
- **TabScanner API Integration**: Professional OCR service for receipt processing
- **Structured Data Extraction**: Merchant name, items, prices, totals, tax, date
- **Base64 Storage**: In-memory image storage with unique imageIds
- **Automatic Bill Splitting**: AI suggests splitting bills among friends

### 📊 Data Visualization
- **Car-Style Gauge Meter**: Weekly spending goal displayed as semicircular gauge (0-180°)
- **Color-Coded Progress**: Green (<80%), Orange (80-99%), Red (≥100%)
- **Animated Needle**: Smooth CSS transitions for gauge updates
- **Spending Categories Bubble Chart**: Interactive visualization of spending by category
- **2-Decimal Precision**: All monetary values formatted to 2 decimal places
- **Responsive Layout**: Right-aligned "Remaining" section, balanced UI design

![Gauge Meter & Visualization](examples/hackathon-starter/imgs/Screenshot%202026-02-01%20052704.png)

## 🏗️ Complete Project Structure

```
hackathon-starter/
├── main.go                          # Main Go server (4391 lines)
│   ├── uploadedImages map           # In-memory base64 image storage
│   ├── /upload-receipt              # POST endpoint for image uploads
│   ├── /balance endpoint            # GET current balance
│   ├── /transactions endpoint       # GET transaction history
│   ├── /weekly-goal endpoints       # POST/GET weekly spending goals
│   ├── /spending-categories         # GET categorized spending
│   ├── /ws WebSocket                # Real-time agent communication
│   └── Banking Tools:
│       ├── get_balance              # Retrieve current balance
│       ├── get_transactions         # Fetch transaction history
│       ├── categorize_transactions  # Categorize spending by type
│       ├── set_weekly_spending_goal # Set weekly budget limit
│       ├── get_weekly_spending_progress # Track goal progress
│       └── process_receipt_image    # Process receipt via TabScanner API
│
├── frontend/                        # React frontend (665 lines)
│   ├── main.tsx                     # Main app component
│   │   ├── WeeklySpendingGoal       # SVG gauge meter widget
│   │   ├── SpendingCategories       # Bubble chart visualization
│   │   ├── uploadedImage state      # Image preview management
│   │   ├── Camera Button (📷)       # Fixed position upload trigger
│   │   └── Image Preview Component  # Thumbnail with close button
│   │
│   ├── SpendingCategories.tsx       # Spending bubble chart (307 lines)
│   │   ├── D3.js integration        # Force simulation for bubbles
│   │   └── Interactive tooltips     # Hover details for categories
│   │
│   ├── styles.css                   # Global styles (930 lines)
│   │   ├── .gauge-container         # Gauge meter styles
│   │   ├── .gauge-svg               # SVG gauge dimensions
│   │   ├── .spending-details        # Flexbox layout with space-between
│   │   └── .remaining               # Right-aligned remaining section
│   │
│   ├── package.json                 # Frontend dependencies
│   │   ├── react: 18.2.0
│   │   ├── @liminalcash/nim-chat: ^0.1.1
│   │   ├── d3: ^7.8.5               # For bubble chart
│   │   └── vite: 5.0.8              # Build tool
│   │
│   └── vite.config.ts               # Vite configuration
│
├── .env                             # Backend configuration
│   ├── ANTHROPIC_API_KEY            # Claude API key
│   ├── LIMINAL_BASE_URL             # Banking API URL
│   ├── TABSCANNER_API_KEY           # Receipt OCR API key
│   └── PORT=8080                    # Server port
│
├── go.mod                           # Go dependencies
│   ├── github.com/liminalcash/nim-go-sdk v0.3.3
│   ├── github.com/anthropic-ai/anthropic-sdk-go
│   └── github.com/gorilla/websocket
│
└── README.md                        # This file
```

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+**: [Install Go](https://golang.org/doc/install)
- **Node.js 18+**: [Install Node.js](https://nodejs.org/)
- **API Keys**:
  - Anthropic API key: [Get from Anthropic](https://console.anthropic.com/)
  - Liminal API credentials: [Sign up at Liminal](https://liminal.cash)
  - TabScanner API key: [Get from TabScanner](https://tabscanner.com/)

### 1. Backend Setup

```bash
# Navigate to hackathon-starter directory
cd nim-go-sdk/examples/hackathon-starter

# Create .env file
cp .env.example .env

# Edit .env with your API keys
# Required variables:
ANTHROPIC_API_KEY=sk-ant-xxx
LIMINAL_BASE_URL=https://api.liminal.cash
TABSCANNER_API_KEY=your_tabscanner_key
PORT=8080

# Install Go dependencies
go mod download

# Run the backend server
go run main.go
```

**Expected Output:**
```
🚀 Server starting on port 8080...
📡 WebSocket endpoint: /ws
🔧 Banking API connected
✅ Agent initialized with Claude Sonnet 4
💾 Image upload endpoint ready at /upload-receipt
```

### 2. Frontend Setup

```bash
# Navigate to frontend directory
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev
```

**Expected Output:**
```
  VITE v5.0.8  ready in 423 ms

  ➜  Local:   http://localhost:5173/
  ➜  Network: use --host to expose
```

### 3. Access the Application

Open your browser to: **http://localhost:5173**

You should see:
- Chat interface at the bottom
- Weekly spending goal gauge meter at the top
- Spending categories bubble chart
- Camera button (📷) in the bottom right corner

## 💡 Usage Guide

### Using the Financial Agent

**Chat Interface Examples:**

```
User: "What's my current balance?"
Agent: "Your current balance is $1,250.00"

User: "Show me my recent transactions"
Agent: [Displays transaction list with dates, merchants, amounts]

User: "Set a weekly spending goal of $200"
Agent: "✓ Weekly spending goal set to $200.00"

User: "How much have I spent this week?"
Agent: "You've spent $156.34 this week, leaving $43.66 remaining"

User: "Show my spending by category"
Agent: [Displays categorized breakdown with percentages]
```

**Available Commands:**
- Check balance: "What's my balance?", "How much money do I have?"
- View transactions: "Show recent transactions", "What did I spend on?"
- Set goals: "Set weekly goal to $X", "Change my budget to $X"
- Track progress: "How much have I spent?", "Am I on track?"
- Categorize: "Show spending categories", "Where does my money go?"
- Process receipts: "Process this receipt", "Split this bill"

### Uploading and Processing Receipts

![Receipt Upload & Processing](examples/hackathon-starter/imgs/WhatsApp%20Image%202026-02-01%20at%2010.21.11.jpeg)

**Step 1: Upload Receipt Image**
1. Click the camera button (📷) in the bottom right corner
2. Select a receipt image from your device
3. Preview appears with thumbnail and "✅ Image Ready" message

**Step 2: Process Receipt**
1. Open chat interface
2. Type: "Process this receipt" or "Split this bill"
3. Agent calls TabScanner API to extract data
4. Structured receipt data returned:
   - Merchant name
   - Total amount
   - Line items with prices
   - Tax amount
   - Date
   - Currency

**Step 3: Bill Splitting (Optional)**
1. Agent automatically offers to split the bill
2. Specify number of people: "Split it 3 ways"
3. Agent calculates per-person amount including tax

**Example Receipt Processing Flow:**
```
User: [Uploads receipt image via camera button]
User: "Process this receipt"
Agent: "I'll process that receipt for you..."
Agent: "Receipt from Starbucks:
        - Total: $18.50
        - Items: 2x Coffee ($8.00), 1x Pastry ($10.50)
        - Tax: $1.65
        - Date: 2024-01-15
        Would you like me to split this bill?"

User: "Split it 2 ways"
Agent: "Each person owes $9.25"
```

### Viewing Spending Categories

The spending categories bubble chart automatically updates with your transaction data:

1. **Bubble Size**: Represents spending amount (larger = more spent)
2. **Colors**: Different colors for each category (Food, Travel, etc.)
3. **Hover**: Shows exact amount and percentage
4. **Real-time Updates**: Refreshes via WebSocket when new transactions occur

**Categories:**
- 🍔 Food: Restaurants, groceries, delivery
- ✈️ Travel: Flights, hotels, transportation
- 🎬 Entertainment: Movies, concerts, subscriptions
- 🛍️ Shopping: Retail, online purchases
- 🏥 Healthcare: Medical, pharmacy, insurance
- 💡 Utilities: Electricity, water, internet
- 🚗 Transport: Gas, parking, ride-sharing
- 📱 Subscriptions: Streaming, software, memberships

### Understanding the Weekly Goal Gauge

The car-style gauge meter shows your weekly spending progress:

**Gauge Components:**
- **Background Arc**: Light gray semicircle (0-180°)
- **Progress Arc**: Colored based on spending percentage
  - **Green**: 0-79% (on track)
  - **Orange**: 80-99% (approaching limit)
  - **Red**: 100%+ (over budget)
- **Needle**: Animated pointer showing exact position
- **Center Percentage**: Large number showing % of goal used
- **Spent This Week**: Left side - total spent so far (2 decimals)
- **Remaining**: Right side - amount left before hitting goal (2 decimals)

**Example:**
```
Weekly Goal: $200.00
Spent: $156.34
Remaining: $43.66
Gauge: Orange arc at ~140° with needle pointing to 78%
```

## 🛠️ Backend Implementation Details

### Image Upload System

**Endpoint:** `POST /upload-receipt`

**Request:**
```bash
curl -X POST http://localhost:8080/upload-receipt \
  -F "image=@receipt.jpg" \
  -H "Content-Type: multipart/form-data"
```

**Response:**
```json
{
  "imageId": "img_1704124567890",
  "message": "Image uploaded successfully"
}
```

**Implementation:**
```go
// In main.go lines 162-221
var uploadedImages = make(map[string]string) // imageId -> base64 data

http.HandleFunc("/upload-receipt", func(w http.ResponseWriter, r *http.Request) {
    // Parse multipart form (max 10MB)
    r.ParseMultipartForm(10 << 20)
    
    // Read file from form
    file, _, err := r.FormFile("image")
    
    // Convert to base64
    imageData, _ := io.ReadAll(file)
    base64Data := base64.StdEncoding.EncodeToString(imageData)
    
    // Generate unique imageId
    imageId := fmt.Sprintf("img_%d", time.Now().UnixNano()/1000000)
    
    // Store in memory
    uploadedImages[imageId] = base64Data
    
    // Return response
    json.NewEncoder(w).Encode(map[string]string{
        "imageId": imageId,
        "message": "Image uploaded successfully",
    })
})
```

### Receipt Processing Tool

**Tool Name:** `process_receipt_image`

**Parameters:**
- `imageId` (string): ID of uploaded image or "latest" for most recent

**Implementation:**
```go
// In main.go lines 1935-1985
func createReceiptProcessorTool() *core.Tool {
    return tools.NewBuilder().
        WithName("process_receipt_image").
        WithDescription("Process receipt image using TabScanner API").
        WithParameter("imageId", tools.ParamTypeString, "Image ID from upload", true).
        WithHandler(func(ctx context.Context, args map[string]interface{}) (*core.ToolResult, error) {
            imageId := args["imageId"].(string)
            
            // Get latest image if requested
            if imageId == "latest" {
                imageId = findLatestImageId(uploadedImages)
            }
            
            // Retrieve base64 data
            base64Data := uploadedImages[imageId]
            
            // Decode and save temporary file
            imageBytes, _ := base64.StdEncoding.DecodeString(base64Data)
            tmpFile := fmt.Sprintf("receipt-upload-%d.jpeg", time.Now().Unix())
            os.WriteFile(tmpFile, imageBytes, 0644)
            
            // Call TabScanner API
            token := callTabScannerProcess(tmpFile)
            receiptData := pollTabScannerResult(token)
            
            return &core.ToolResult{
                Success: true,
                Data: map[string]interface{}{
                    "merchant": receiptData.Merchant,
                    "total": receiptData.Total,
                    "items": receiptData.LineItems,
                    "tax": receiptData.Tax,
                    "date": receiptData.Date,
                },
            }, nil
        }).
        Build()
}
```

### TabScanner API Integration

**Process Endpoint:**
```go
func callTabScannerProcess(imagePath string) string {
    url := "https://api.tabscanner.com/api/2/process"
    
    // Read image file
    imageData, _ := os.ReadFile(imagePath)
    base64Image := base64.StdEncoding.EncodeToString(imageData)
    
    // Create request
    payload := map[string]interface{}{
        "image": base64Image,
        "documentType": "receipt",
    }
    
    // POST request with API key
    req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
    req.Header.Set("apikey", os.Getenv("TABSCANNER_API_KEY"))
    
    // Get token
    resp, _ := http.DefaultClient.Do(req)
    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    
    return result["token"].(string)
}
```

**Result Polling:**
```go
func pollTabScannerResult(token string) ReceiptData {
    url := fmt.Sprintf("https://api.tabscanner.com/api/result/%s", token)
    maxAttempts := 30
    
    for i := 0; i < maxAttempts; i++ {
        time.Sleep(1 * time.Second)
        
        resp, _ := http.Get(url)
        var result map[string]interface{}
        json.NewDecoder(resp.Body).Decode(&result)
        
        // Status codes: 2 = processing, 3 = done
        if result["status"].(float64) == 3 {
            return parseReceiptData(result["result"])
        }
    }
    
    return ReceiptData{} // Timeout
}
```

### Banking Tools

**Get Balance:**
```go
func createBalanceTool(liminalExec executor.Executor) *core.Tool {
    return tools.NewBuilder().
        WithName("get_balance").
        WithDescription("Get current account balance").
        WithHandler(func(ctx context.Context, args map[string]interface{}) (*core.ToolResult, error) {
            balance, err := liminalExec.GetBalance(ctx)
            return &core.ToolResult{
                Success: true,
                Data: map[string]interface{}{
                    "balance": balance.Available,
                    "currency": "USD",
                },
            }, nil
        }).
        Build()
}
```

**Set Weekly Goal:**
```go
func createWeeklyGoalTool() *core.Tool {
    return tools.NewBuilder().
        WithName("set_weekly_spending_goal").
        WithParameter("amount", tools.ParamTypeNumber, "Goal amount", true).
        WithHandler(func(ctx context.Context, args map[string]interface{}) (*core.ToolResult, error) {
            amount := args["amount"].(float64)
            
            // Store in memory (or database)
            weeklyGoalMutex.Lock()
            currentWeeklyGoal = amount
            weeklyGoalMutex.Unlock()
            
            return &core.ToolResult{
                Success: true,
                Data: map[string]interface{}{
                    "goal": amount,
                    "message": fmt.Sprintf("Weekly goal set to $%.2f", amount),
                },
            }, nil
        }).
        Build()
}
```

## 🎨 Frontend Implementation Details

### Gauge Meter Component

**Location:** `frontend/main.tsx` lines 343-409

**SVG Gauge Visualization:**
```tsx
// Calculate gauge angle (0-180 degrees)
const percentage = (spent_so_far / weekly_goal) * 100
const gaugeAngle = Math.min((percentage / 100) * 180, 180)

// Determine color based on percentage
const gaugeColor = percentage < 80 ? '#10b981' : // Green
                   percentage < 100 ? '#f59e0b' : // Orange
                   '#ef4444' // Red

// Calculate needle endpoint using trigonometry
const needleAngle = gaugeAngle * (Math.PI / 180) - Math.PI
const needleX = 100 + 70 * Math.cos(needleAngle)
const needleY = 100 + 70 * Math.sin(needleAngle)

return (
  <div className="gauge-container">
    <svg viewBox="0 0 200 120" className="gauge-svg">
      {/* Background arc */}
      <path
        d="M 20 100 A 80 80 0 0 1 180 100"
        fill="none"
        stroke="#e5e7eb"
        strokeWidth="20"
        strokeLinecap="round"
      />
      
      {/* Progress arc */}
      <path
        d="M 20 100 A 80 80 0 0 1 180 100"
        fill="none"
        stroke={gaugeColor}
        strokeWidth="20"
        strokeLinecap="round"
        strokeDasharray={`${(gaugeAngle / 180) * 251.2} 251.2`}
        style={{ transition: 'stroke-dasharray 0.5s ease' }}
      />
      
      {/* Needle */}
      <line
        x1="100"
        y1="100"
        x2={needleX}
        y2={needleY}
        stroke="#1f2937"
        strokeWidth="3"
        strokeLinecap="round"
        style={{ transition: 'all 0.5s ease' }}
      />
      
      {/* Center percentage */}
      <text x="100" y="90" textAnchor="middle" fontSize="24" fontWeight="bold">
        {Number(percentage).toFixed(0)}%
      </text>
    </svg>
    
    {/* Spending details */}
    <div className="spending-details">
      <div className="spent">
        <div className="label">Spent This Week</div>
        <div className="amount">{Number(spent_so_far).toFixed(2)} {currency}</div>
      </div>
      <div className="remaining">
        <div className="label">Remaining</div>
        <div className="amount">{Number(remaining).toFixed(2)} {currency}</div>
      </div>
    </div>
  </div>
)
```

### Camera Button & Image Upload

**Location:** `frontend/main.tsx` lines 571-631

**Upload Implementation:**
```tsx
// State for uploaded image preview
const [uploadedImage, setUploadedImage] = React.useState<{
  url: string;
  name: string;
} | null>(null)

// Hidden file input
<input
  type="file"
  accept="image/*"
  id="receipt-upload"
  style={{ display: 'none' }}
  onChange={async (e) => {
    const file = e.target.files?.[0]
    if (!file) return
    
    // Create preview URL
    const previewUrl = URL.createObjectURL(file)
    
    // Upload to backend
    const formData = new FormData()
    formData.append('image', file)
    
    try {
      const response = await fetch('http://localhost:8080/upload-receipt', {
        method: 'POST',
        body: formData,
      })
      
      const data = await response.json()
      
      // Store imageId for agent to use
      (window as any).lastReceiptImageId = data.imageId
      
      // Show preview
      setUploadedImage({
        url: previewUrl,
        name: file.name,
      })
    } catch (error) {
      console.error('Upload failed:', error)
    }
  }}
/>

// Camera button (fixed position)
<button
  onClick={() => document.getElementById('receipt-upload')?.click()}
  style={{
    position: 'fixed',
    bottom: '20px',
    right: '100px',
    width: '50px',
    height: '50px',
    borderRadius: '50%',
    backgroundColor: '#3b82f6',
    border: 'none',
    fontSize: '24px',
    cursor: 'pointer',
    zIndex: 9999,
    boxShadow: '0 4px 6px rgba(0, 0, 0, 0.1)',
  }}
>
  📷
</button>
```

### Image Preview Component

**Location:** `frontend/main.tsx` lines 518-568

```tsx
{uploadedImage && (
  <div style={{
    position: 'fixed',
    bottom: '90px',
    right: '20px',
    backgroundColor: 'white',
    padding: '10px',
    borderRadius: '12px',
    boxShadow: '0 4px 12px rgba(0, 0, 0, 0.15)',
    zIndex: 9999,
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    gap: '8px',
  }}>
    {/* Thumbnail */}
    <img
      src={uploadedImage.url}
      alt="Receipt preview"
      style={{
        width: '60px',
        height: '60px',
        objectFit: 'cover',
        borderRadius: '8px',
        border: '2px solid #10b981',
      }}
    />
    
    {/* Status */}
    <div style={{ fontSize: '14px', fontWeight: 'bold', color: '#10b981' }}>
      ✅ Image Ready
    </div>
    
    {/* Filename */}
    <div style={{ fontSize: '12px', color: '#6b7280', maxWidth: '150px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
      {uploadedImage.name}
    </div>
    
    {/* Instruction */}
    <div style={{ fontSize: '11px', color: '#9ca3af', textAlign: 'center' }}>
      Ask me to process it!
    </div>
    
    {/* Close button */}
    <button
      onClick={() => {
        URL.revokeObjectURL(uploadedImage.url)
        setUploadedImage(null)
      }}
      style={{
        position: 'absolute',
        top: '4px',
        right: '4px',
        background: '#ef4444',
        color: 'white',
        border: 'none',
        borderRadius: '50%',
        width: '20px',
        height: '20px',
        fontSize: '12px',
        cursor: 'pointer',
        lineHeight: '1',
      }}
    >
      ×
    </button>
  </div>
)}
```

### Spending Categories Integration

**Location:** `frontend/SpendingCategories.tsx`

**D3.js Force Simulation:**
```tsx
import * as d3 from 'd3'

// Create force simulation
const simulation = d3.forceSimulation(nodes)
  .force('charge', d3.forceManyBody().strength(5))
  .force('center', d3.forceCenter(width / 2, height / 2))
  .force('collision', d3.forceCollide().radius((d: any) => d.radius + 2))

// Update positions on tick
simulation.on('tick', () => {
  svg.selectAll('circle')
    .attr('cx', (d: any) => d.x)
    .attr('cy', (d: any) => d.y)
})

// Add tooltips
svg.selectAll('circle')
  .on('mouseover', (event, d: any) => {
    tooltip.style('display', 'block')
    tooltip.html(`
      <strong>${d.category}</strong><br/>
      Amount: $${d.value.toFixed(2)}<br/>
      Percentage: ${d.percentage.toFixed(1)}%
    `)
  })
```

## 📋 Complete API Reference

### Backend REST Endpoints

| Method | Endpoint | Description | Request Body | Response |
|--------|----------|-------------|--------------|----------|
| GET | `/balance` | Get current balance | None | `{ "balance": 1250.00, "currency": "USD" }` |
| GET | `/transactions` | Get transaction history | None | `{ "transactions": [...] }` |
| POST | `/weekly-goal` | Set weekly spending goal | `{ "amount": 200.00 }` | `{ "goal": 200.00, "message": "..." }` |
| GET | `/weekly-goal` | Get current weekly goal | None | `{ "goal": 200.00, "spent": 156.34, "remaining": 43.66 }` |
| GET | `/spending-categories` | Get categorized spending | None | `{ "categories": [...] }` |
| POST | `/upload-receipt` | Upload receipt image | `multipart/form-data` with `image` field | `{ "imageId": "img_xxx", "message": "..." }` |
| GET | `/ws` | WebSocket connection | N/A | Bidirectional WebSocket |

### WebSocket Protocol

**Connect:**
```javascript
const ws = new WebSocket('ws://localhost:8080/ws')
```

**Send Message:**
```javascript
ws.send(JSON.stringify({
  type: 'user_message',
  content: 'What is my balance?',
  timestamp: Date.now()
}))
```

**Receive Response:**
```javascript
ws.onmessage = (event) => {
  const data = JSON.parse(event.data)
  // data.type: 'agent_message', 'tool_call', 'error'
  // data.content: message text
  // data.toolName: (if type === 'tool_call')
}
```

### Agent Tools

| Tool Name | Parameters | Description | Example |
|-----------|------------|-------------|---------|
| `get_balance` | None | Retrieve current account balance | "What's my balance?" |
| `get_transactions` | `limit` (optional) | Fetch recent transactions | "Show my last 10 transactions" |
| `categorize_transactions` | None | Categorize spending by type | "Show spending categories" |
| `set_weekly_spending_goal` | `amount` (number) | Set weekly budget limit | "Set goal to $200" |
| `get_weekly_spending_progress` | None | Track goal progress | "How much have I spent?" |
| `process_receipt_image` | `imageId` (string) | Process receipt via TabScanner | "Process this receipt" |

## ⚙️ Configuration

### Environment Variables (.env)

```bash
# Required
ANTHROPIC_API_KEY=sk-ant-xxx          # Your Anthropic API key
LIMINAL_BASE_URL=https://api.liminal.cash  # Liminal Banking API URL
TABSCANNER_API_KEY=your_key_here      # TabScanner API key for receipt OCR

# Optional
PORT=8080                              # Backend server port (default: 8080)
EMAIL_FROM=your-email@outlook.com      # For email notifications (optional)
EMAIL_PASSWORD=your-app-password       # Outlook app password (optional)
LOG_LEVEL=info                        # Logging level: debug, info, warn, error
```

### Frontend Configuration

**WebSocket URL** (in `frontend/main.tsx`):
```typescript
const wsUrl = 'ws://localhost:8080/ws'
```

**Upload Endpoint** (in `frontend/main.tsx`):
```typescript
const uploadUrl = 'http://localhost:8080/upload-receipt'
```

### TabScanner Configuration

**API Endpoints:**
- Process: `https://api.tabscanner.com/api/2/process`
- Result: `https://api.tabscanner.com/api/result/{token}`

**Status Codes:**
- `2`: Processing in progress
- `3`: Processing complete

**Polling Strategy:**
- Max attempts: 30
- Interval: 1 second
- Total timeout: 30 seconds

## 🐛 Troubleshooting

### Backend Issues

**"ANTHROPIC_API_KEY not found"**
```bash
# Create .env file in hackathon-starter directory
echo "ANTHROPIC_API_KEY=your_key_here" > .env
```

**"Port already in use"**
```bash
# Change PORT in .env
PORT=8081
```

**"Failed to connect to Liminal API"**
- Check `LIMINAL_BASE_URL` is correct
- Verify API credentials
- Check internet connection
- Review firewall settings

**"TabScanner API error"**
- Verify `TABSCANNER_API_KEY` is set correctly
- Check TabScanner account has credits
- Ensure image format is supported (JPEG, PNG)
- Verify image size < 10MB

### Frontend Issues

**WebSocket connection failed**
- Ensure backend is running on port 8080
- Check firewall settings
- Verify WebSocket URL in main.tsx
- Check browser console for CORS errors

**Image upload not working**
- Check backend `/upload-receipt` endpoint is running
- Verify CORS headers are set correctly
- Ensure file size < 10MB
- Check browser console for errors
- Verify file input accepts image types

**Gauge meter not updating**
- Check WebSocket connection is active
- Verify weekly goal is set via API
- Check spending data is being sent from backend
- Inspect React component state in DevTools

**Spending categories empty**
- Ensure transactions exist in Liminal account
- Check `/spending-categories` endpoint returns data
- Verify WebSocket is sending category updates
- Check D3.js console errors

### Common Issues

**Camera button not visible**
- Check z-index in styles (should be 9999)
- Verify button is not hidden by other elements
- Inspect element position with DevTools
- Check CSS for `display: none` or `visibility: hidden`

**Decimal formatting issues**
- Ensure `.toFixed(2)` is applied to all monetary values
- Check `Number()` conversion before `.toFixed()`
- Verify currency values are numbers, not strings

**Receipt processing timeout**
- TabScanner may take 10-30 seconds for complex receipts
- Check network connection
- Verify TabScanner API is not rate-limited
- Try with a clearer receipt image

**"Image Ready" preview stuck**
- Click the × button to close preview
- Refresh browser if stuck
- Check `URL.revokeObjectURL()` is called on close
- Verify state update in React DevTools

## 🧪 Testing & Development

### Running Tests

```bash
# Backend tests
cd nim-go-sdk/examples/hackathon-starter
go test ./...

# Frontend tests
cd frontend
npm test
```

### Building for Production

**Backend:**
```bash
# Build binary
go build -o hackathon-agent main.go

# Run binary
./hackathon-agent
```

**Frontend:**
```bash
cd frontend
npm run build
# Output in dist/ directory

# Preview production build
npm run preview
```

### Development Tips

**Backend:**
- Use `go run main.go` for auto-reload during development
- Enable verbose logging: `LOG_LEVEL=debug`
- Use Postman to test REST endpoints
- Monitor WebSocket with browser DevTools

**Frontend:**
- Use `npm run dev` for hot module replacement
- Check browser console for React errors
- Inspect WebSocket messages in Network tab
- Use React DevTools extension for state inspection
- Test responsive design with mobile viewport

**Receipt Processing:**
- Test with various receipt formats (grocery, restaurant, retail)
- Verify OCR accuracy with clear vs. blurry images
- Check error handling for unsupported formats
- Monitor TabScanner API usage and credits

### Local Development Workflow

1. **Start backend:**
   ```bash
   cd nim-go-sdk/examples/hackathon-starter
   go run main.go
   ```

2. **Start frontend:**
   ```bash
   cd frontend
   npm run dev
   ```

3. **Test image upload:**
   - Click camera button (📷)
   - Select test receipt image
   - Verify preview appears

4. **Test receipt processing:**
   - Open chat
   - Type "process this receipt"
   - Verify TabScanner API call and response

5. **Test weekly goal gauge:**
   - Set goal: "Set weekly goal to $200"
   - Verify gauge updates with correct color and percentage
   - Check decimal formatting (2 places)

6. **Test spending categories:**
   - Ensure transactions exist
   - Check bubble chart renders
   - Hover to verify tooltips

## 🎯 Use Cases

### Personal Finance Management
- Track daily spending across categories
- Monitor budget adherence with visual gauge
- Digitize receipts for record-keeping
- Analyze spending patterns over time
- Get AI recommendations for budget optimization

### Expense Tracking
- Upload receipts instantly with one click
- Automatic extraction of merchant, items, amounts
- Split bills among friends automatically
- Export receipt data for reimbursement
- Maintain digital receipt archive

### Budget Planning
- Set realistic weekly spending goals
- Visual feedback with color-coded gauge
- Real-time progress tracking
- Category-based spending analysis
- Adjust goals based on AI insights

### Small Business
- Digitize business receipts
- Track expense categories for tax purposes
- Monitor business spending against budgets
- Generate spending reports
- Maintain compliance documentation

## 🔮 Future Enhancements

### Planned Features
- [ ] Multiple receipt uploads (batch processing)
- [ ] Receipt image history and search
- [ ] Monthly spending trends graph
- [ ] Budget forecast predictions
- [ ] Email notifications for budget alerts
- [ ] Export receipts to PDF
- [ ] Mobile responsive design improvements
- [ ] Dark mode theme
- [ ] User authentication and accounts
- [ ] Persistent database storage (PostgreSQL)
- [ ] Recurring expense detection
- [ ] Bill payment reminders
- [ ] Savings goals tracking
- [ ] Investment portfolio integration

### Technical Improvements
- [ ] Redis caching for faster performance
- [ ] Docker containerization
- [ ] Kubernetes deployment config
- [ ] CI/CD pipeline with GitHub Actions
- [ ] E2E testing with Playwright
- [ ] Load testing with k6
- [ ] Performance monitoring (Prometheus/Grafana)
- [ ] Error tracking (Sentry)
- [ ] Rate limiting middleware
- [ ] Image compression before upload
- [ ] S3 storage for receipt images
- [ ] GraphQL API option
- [ ] Webhook support for real-time updates

## 📚 Architecture Overview

### System Components

```
┌─────────────────────────────────────────────────────────────┐
│                        Frontend (React)                      │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────────┐   │
│  │   Chat UI   │  │Gauge Meter   │  │ Spending Bubbles │   │
│  │  (NimChat)  │  │  (SVG)       │  │    (D3.js)       │   │
│  └──────┬──────┘  └──────┬───────┘  └────────┬─────────┘   │
│         │                 │                    │              │
│         │         ┌───────┴────────────────────┘              │
│         │         │       WebSocket                          │
│  ┌──────▼─────────▼─────────────────────────────────────┐   │
│  │          Camera Button & Image Upload                 │   │
│  │  📷 → FormData → POST /upload-receipt                 │   │
│  └────────────────────────────────────────────────────────┘   │
└───────────────────────┬─────────────────────────────────────┘
                        │ HTTP/WebSocket
┌───────────────────────▼─────────────────────────────────────┐
│                    Go Backend (main.go)                      │
│  ┌───────────────────────────────────────────────────────┐  │
│  │              WebSocket Handler (/ws)                   │  │
│  │  Receives user messages → Routes to Agent             │  │
│  └───────────────────────┬───────────────────────────────┘  │
│                          │                                   │
│  ┌───────────────────────▼───────────────────────────────┐  │
│  │          nim-go-sdk Agent Orchestrator                 │  │
│  │  ┌─────────────────┐       ┌─────────────────────┐    │  │
│  │  │  Claude Sonnet  │       │    Tool Registry    │    │  │
│  │  │  (Anthropic)    │◄─────►│  - get_balance      │    │  │
│  │  │                 │       │  - get_transactions │    │  │
│  │  │  Decides which  │       │  - set_weekly_goal  │    │  │
│  │  │  tools to call  │       │  - process_receipt  │    │  │
│  │  └─────────────────┘       └─────────────────────┘    │  │
│  └───────────────────┬────────────────────────────────────┘  │
│                      │                                        │
│  ┌───────────────────▼───────────────────────────────┐       │
│  │         REST API Endpoints                        │       │
│  │  /balance /transactions /weekly-goal              │       │
│  │  /spending-categories /upload-receipt             │       │
│  └────────┬────────────────────┬──────────────────────┘      │
│           │                    │                              │
│  ┌────────▼────────┐  ┌────────▼──────────┐                 │
│  │ In-Memory Store │  │ uploadedImages    │                 │
│  │ - Weekly Goals  │  │ map[string]string │                 │
│  │ - Notifications │  │ (imageId→base64)  │                 │
│  └─────────────────┘  └───────────────────┘                 │
└─────────┬──────────────────────┬───────────────────────────┘
          │                      │
┌─────────▼────────┐   ┌─────────▼───────────┐
│  Liminal Banking │   │  TabScanner OCR API │
│       API        │   │  - POST /process    │
│ - Get Balance    │   │  - GET /result/{id} │
│ - Transactions   │   │  - Status polling   │
│ - Categories     │   │  - Receipt parsing  │
└──────────────────┘   └─────────────────────┘
```

### Data Flow

**1. User Uploads Receipt:**
```
User clicks 📷 button
→ Browser file picker opens
→ User selects image
→ Frontend creates FormData
→ POST http://localhost:8080/upload-receipt
→ Backend receives multipart form
→ Converts to base64
→ Stores in uploadedImages map with unique imageId
→ Returns { imageId: "img_xxx", message: "success" }
→ Frontend displays preview with thumbnail
```

**2. User Asks to Process Receipt:**
```
User types "process this receipt" in chat
→ WebSocket sends message to backend
→ nim-go-sdk orchestrator receives message
→ Claude analyzes intent: "user wants receipt processed"
→ Claude decides to call process_receipt_image tool
→ Tool retrieves imageId="latest" from uploadedImages
→ Decodes base64 to image bytes
→ Saves temporary file: receipt-upload-{timestamp}.jpeg
→ Calls TabScanner API: POST /api/2/process
→ Receives token: "abc123"
→ Polls GET /api/result/abc123 every 1s
→ Status=2 (processing) → continue polling
→ Status=3 (done) → parse result
→ Extracts: merchant, total, items, tax, date
→ Returns structured data to Claude
→ Claude formats response for user
→ WebSocket sends response to frontend
→ Chat displays receipt details
```

**3. Weekly Goal Gauge Update:**
```
User sets goal: "Set weekly goal to $200"
→ Claude calls set_weekly_spending_goal(amount=200)
→ Backend stores in memory: currentWeeklyGoal = 200
→ Calculates spent_so_far from transactions
→ WebSocket broadcasts update to frontend
→ React component receives { goal: 200, spent: 156.34 }
→ Calculates: percentage = (156.34 / 200) * 100 = 78.17%
→ Calculates: gaugeAngle = 78.17% * 180° = 140.7°
→ Determines color: 78% < 80% → Green
→ SVG gauge updates with new angle
→ Needle rotates smoothly via CSS transition
→ Values update: "Spent: $156.34" "Remaining: $43.66"
```

**4. Spending Categories Visualization:**
```
Backend fetches transactions from Liminal API
→ Categorizes each transaction (Food, Travel, etc.)
→ Aggregates spending per category
→ WebSocket broadcasts to frontend
→ SpendingCategories component receives data
→ D3.js force simulation creates nodes
→ Bubble size based on spending amount
→ Colors assigned per category
→ Force simulation positions bubbles
→ SVG renders interactive chart
→ User hovers → tooltip shows details
```

### Agent Workflow

The AI agent follows this decision-making process:

```
1. Receive User Message
   ↓
2. Claude Analyzes Intent
   - What does the user want?
   - Which tool(s) are needed?
   - Are there missing parameters?
   ↓
3. Tool Selection
   - Balance inquiry → get_balance
   - Transaction list → get_transactions
   - Goal setting → set_weekly_spending_goal
   - Receipt processing → process_receipt_image
   - Spending analysis → categorize_transactions
   ↓
4. Tool Execution
   - Call selected tool(s)
   - Handle errors/retries
   - Aggregate results
   ↓
5. Response Formatting
   - Claude interprets tool results
   - Generates natural language response
   - Includes relevant data/recommendations
   ↓
6. Send to User
   - WebSocket transmits response
   - Frontend displays in chat
   - UI components update if needed
```

### Security & Performance

**Security Measures:**
- API keys stored in `.env` (not committed to git)
- CORS enabled for localhost development
- WebSocket origin validation
- Multipart form size limits (10MB)
- Base64 encoding for image transmission
- In-memory storage (no persistent PII)

**Performance Optimizations:**
- WebSocket for real-time bidirectional communication
- In-memory caching of weekly goals
- D3.js force simulation for efficient bubble layout
- CSS transitions for smooth gauge animations
- Lazy loading of receipt images
- Debounced WebSocket message handling

**Scalability Considerations:**
- Stateless backend design (easy horizontal scaling)
- In-memory storage can be replaced with Redis
- Receipt images can be moved to S3/Cloud Storage
- Database can be added for persistence
- Load balancer can distribute WebSocket connections

## 📞 Support & Resources

### Getting Help

**Troubleshooting Checklist:**
1. Check all API keys are set in `.env`
2. Verify backend is running on port 8080
3. Confirm frontend is running on port 5173
4. Check browser console for errors
5. Inspect WebSocket connection in Network tab
6. Review backend logs for error messages

**Common Questions:**

**Q: Can I use a different AI model?**  
A: Yes, modify `ANTHROPIC_API_KEY` to use different Claude versions, or replace with OpenAI/other providers by updating the nim-go-sdk configuration.

**Q: How do I persist data across restarts?**  
A: Currently uses in-memory storage. Add PostgreSQL or MongoDB by replacing the in-memory maps with database queries.

**Q: Can I deploy this to production?**  
A: Yes, but add:
- HTTPS/TLS for WebSocket (wss://)
- Authentication & authorization
- Database for persistence
- Environment-specific configs
- Rate limiting & error handling
- Monitoring & logging infrastructure

**Q: How accurate is the receipt OCR?**  
A: TabScanner API accuracy depends on:
- Image quality (lighting, resolution)
- Receipt condition (wrinkles, fading)
- Format (structured vs. handwritten)
- Generally 90%+ accuracy for standard receipts

**Q: Can I customize spending categories?**  
A: Yes, modify the categorization logic in `main.go`. Currently uses keyword matching for merchant names (e.g., "Starbucks" → Food).

### Documentation Links

- **Liminal API Docs:** [https://docs.liminal.cash](https://docs.liminal.cash)
- **nim-go-sdk Documentation:** [nim-go-sdk/README.md](../../README.md)
- **Anthropic Claude API:** [https://docs.anthropic.com/](https://docs.anthropic.com/)
- **TabScanner API:** [https://tabscanner.com/api/](https://tabscanner.com/api/)
- **React Documentation:** [https://react.dev/](https://react.dev/)
- **D3.js Documentation:** [https://d3js.org/](https://d3js.org/)
- **Go Documentation:** [https://go.dev/doc/](https://go.dev/doc/)

### Related Projects

- **Liminal Banking SDK:** Core banking API integration
- **nim-go-sdk:** Go SDK for AI agent orchestration
- **TabScanner:** Receipt OCR API service
- **@liminalcash/nim-chat:** React chat component library

## 🙏 Acknowledgments

This project was built for the **Liminal Hackathon** using:

- **Liminal** - Banking API and SDK infrastructure
- **Anthropic** - Claude Sonnet 4 AI model
- **TabScanner** - Receipt OCR API service
- **Go** - Backend server and orchestration
- **React** - Frontend UI framework
- **D3.js** - Data visualization library
- **Vite** - Frontend build tool

Special thanks to the Liminal team for providing excellent documentation and support.

## 📄 License

This project is provided as-is for educational and commercial use. Feel free to use, modify, and distribute as needed.

---

**Built with ❤️ for intelligent financial management**

Happy building! 🚀

## 📞 Contact & Contributing

For issues, questions, or contributions:
- Open an issue on GitHub
- Check existing documentation
- Review troubleshooting section
- Contact project maintainers

**Contribution Areas:**
- Additional spending categories
- More banking tools
- UI/UX improvements
- Documentation enhancements
- Bug fixes
- Test coverage
- Performance optimizations
- Mobile responsiveness
- Accessibility improvements

---

**Version:** 1.0.0  
**Last Updated:** January 2024  
**Tested With:**
- Go 1.21+
- Node.js 18+
- React 18.2.0
- nim-go-sdk v0.3.3
- Claude Sonnet 4
