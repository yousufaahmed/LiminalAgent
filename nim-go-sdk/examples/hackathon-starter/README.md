# Liminal Agentic AI - Financial Management Platform

A comprehensive AI-powered financial management platform combining intelligent budget tracking, spending analysis, receipt scanning, and conversational banking powered by Liminal's banking APIs and AI agents.

## 🌟 Features

### 💰 Financial Management
- **Real-time Banking Integration**: Connect to Liminal banking APIs for live transaction data
- **Budget Tracking**: Set and monitor weekly spending goals
- **Spending Categories**: Automatic transaction categorization (Food, Travel, Subscriptions, etc.)
- **Transaction Analysis**: Smart merchant detection and spending pattern analysis
- **Balance Monitoring**: Real-time account balance tracking

### 🤖 AI Agent
- **Conversational Banking**: Natural language financial assistant
- **Intelligent Insights**: AI-powered spending analysis and recommendations
- **Multi-turn Conversations**: Contextual understanding across interactions
- **Custom Tools**: Extensible tool system for banking operations

### 📸 Receipt Scanner
- **Camera Integration**: Capture receipts directly from your camera
- **Advanced OCR**: Donut model for structured receipt understanding
- **Image Preprocessing**: CLAHE, denoising, and adaptive thresholding for better accuracy
- **Structured Extraction**: Automatic parsing of items, prices, and totals

### 📊 Data Visualization
- **Spending Categories Bubble Chart**: Visual representation of spending by category
- **Interactive Dashboard**: Real-time updates via WebSocket
- **Responsive UI**: Modern React-based interface with smooth animations

## 🏗️ Project Structure

```
LiminalAgenticAI/
├── nim-go-sdk/                       # Go backend service
│   └── examples/hackathon-starter/
│       ├── main.go                   # Main server with banking tools
│       ├── .env                      # Backend configuration
│       └── frontend/                 # React frontend
│           ├── main.tsx              # App entry point
│           ├── SpendingCategories.tsx # Spending visualization
│           ├── styles.css            # Global styles
│           └── package.json          # Frontend dependencies
│
├── receipt.py                        # Receipt scanner with OCR
├── agent.py                          # LangGraph agent (optional)
├── tools.py                          # Custom agent tools
├── config.py                         # Configuration management
├── requirements.txt                  # Python dependencies
└── README.md                         # This file
```

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+** (for backend)
- **Node.js 18+** (for frontend)
- **Python 3.9+** (for receipt scanner)
- **Anthropic API Key** (for AI agent)
- **Liminal API Access** (for banking features)

### 1. Backend Setup (Go)

```bash
cd nim-go-sdk/examples/hackathon-starter

# Create .env file
cat > .env << EOL
ANTHROPIC_API_KEY=your_anthropic_key_here
LIMINAL_BASE_URL=https://api.liminal.cash
PORT=8080
EOL

# Install dependencies
go mod download

# Run backend
go run main.go
```

Backend will start on `http://localhost:8080`

### 2. Frontend Setup (React)

```bash
cd nim-go-sdk/examples/hackathon-starter/frontend

# Install dependencies
npm install

# Run development server
npm run dev
```

Frontend will start on `http://localhost:5173`

### 3. Receipt Scanner Setup (Python)

```bash
# Create virtual environment
python -m venv .venv
source .venv/bin/activate  # On Windows: .venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt

# Run receipt scanner
python receipt.py
```

## 📋 Detailed Setup

### Backend Configuration (.env)

```bash
# Required
ANTHROPIC_API_KEY=sk-ant-xxx          # Your Anthropic API key
LIMINAL_BASE_URL=https://api.liminal.cash

# Optional
PORT=8080                              # Backend server port
EMAIL_FROM=your-email@outlook.com      # For notifications (optional)
EMAIL_PASSWORD=your-app-password       # Outlook app password (optional)
```

### Frontend Configuration

The frontend automatically connects to the backend WebSocket at `ws://localhost:8080/ws`.

To change the backend URL, modify [main.tsx](nim-go-sdk/examples/hackathon-starter/frontend/main.tsx):

```typescript
const wsUrl = 'ws://your-backend:8080/ws';
```

### Python Dependencies

Key packages for receipt scanner:
- `transformers` - Donut OCR model
- `torch` - PyTorch for model inference
- `opencv-python` - Image processing
- `Pillow` - Image enhancement
- `protobuf` - Model serialization
- `sentencepiece` - Tokenization

## 💡 Usage Guide

### Using the Financial Agent

**Chat Interface:**
```
User: "What's my current balance?"
Agent: "Your balance is $1,250.00"

User: "Set a weekly spending goal of $200"
Agent: "✓ Weekly spending goal set to $200.00"

User: "Show my spending by category"
Agent: [Displays categorized spending breakdown]
```

**Available Commands:**
- Check balance
- View recent transactions
- Set weekly spending goals
- Get spending progress
- Analyze spending by category
- View spending trends

### Using the Receipt Scanner

**Option 1: Camera Capture**
```bash
python receipt.py
# Choose option 1
# Press SPACE to capture
# Press ESC to exit
```

**Option 2: Load from File**
```bash
python receipt.py
# Choose option 2
# Enter image path: /path/to/receipt.jpg
```

**Output:**
- Extracted text display in terminal
- Structured data saved to `receipt_output.json`
- Parsed items, quantities, and prices

### Viewing Spending Categories

1. Open frontend: `http://localhost:5173`
2. View the spending categories bubble chart
3. Hover over bubbles for detailed breakdown
4. Categories automatically update from backend

## 🛠️ Backend Tools

### Banking Tools
- `get_balance` - Retrieve current account balance
- `get_transactions` - Fetch transaction history
- `categorize_transactions` - Categorize spending by type

### Budget Tools
- `set_weekly_spending_goal` - Set weekly budget limit
- `get_weekly_spending_progress` - Track goal progress

### Custom Tool Implementation

Add new tools in [main.go](nim-go-sdk/examples/hackathon-starter/main.go):

```go
func createMyCustomTool(liminalExec executor.Executor) *core.Tool {
    return tools.NewBuilder().
        WithName("my_custom_tool").
        WithDescription("What this tool does").
        WithParameter("param_name", tools.ParamTypeString, "Parameter description", true).
        WithHandler(func(ctx context.Context, args map[string]interface{}) (*core.ToolResult, error) {
            // Your implementation here
            return &core.ToolResult{
                Success: true,
                Data: map[string]interface{}{
                    "result": "value",
                },
            }, nil
        }).
        Build()
}

// Register in main():
agent.RegisterTool(createMyCustomTool(liminalExec))
```

## 🎨 Frontend Customization

### Adding New Components

Create in `frontend/` directory:

```typescript
import { useEffect, useState } from 'react';

export function MyComponent({ wsUrl }: { wsUrl: string }) {
  const [data, setData] = useState<any>(null);
  
  useEffect(() => {
    const ws = new WebSocket(wsUrl);
    
    ws.onmessage = (event) => {
      // Handle data
    };
    
    return () => ws.close();
  }, [wsUrl]);
  
  return <div>{/* Your UI */}</div>;
}
```

### Styling

Global styles in [styles.css](nim-go-sdk/examples/hackathon-starter/frontend/styles.css).

Component-specific styles using inline styles or CSS modules.

## 📊 Receipt Scanner Details

### Image Preprocessing Pipeline

1. **CLAHE**: Contrast Limited Adaptive Histogram Equalization
2. **Denoising**: Fast Non-Local Means Denoising
3. **Adaptive Thresholding**: Gaussian adaptive threshold
4. **Sharpening**: PIL ImageEnhance sharpness boost
5. **Contrast Enhancement**: PIL ImageEnhance contrast adjustment

### Supported Receipt Formats

- Store receipts (groceries, retail)
- Restaurant bills
- Service receipts
- Invoice documents

### Model Information

**Donut (Document Understanding Transformer)**
- Model: `naver-clova-ix/donut-base-finetuned-cord-v2`
- Task: Structured document parsing
- Output: JSON with items, prices, store info

## 🔒 Security Considerations

### API Keys
- Store all API keys in `.env` files
- Never commit `.env` files to version control
- Use environment variables in production

### Camera Permissions
- macOS: Grant Terminal camera access in System Settings → Privacy & Security → Camera
- Windows: Check Windows Security camera permissions

### Banking API
- Liminal API credentials should be secured
- Use HTTPS in production
- Implement rate limiting

## 🐛 Troubleshooting

### Backend Issues

**"ANTHROPIC_API_KEY not found"**
```bash
# Create .env file in hackathon-starter directory
echo "ANTHROPIC_API_KEY=your_key" > .env
```

**Port already in use**
```bash
# Change PORT in .env
PORT=8081
```

### Frontend Issues

**WebSocket connection failed**
- Ensure backend is running on port 8080
- Check firewall settings
- Verify WebSocket URL in main.tsx

### Receipt Scanner Issues

**"No module named 'cv2'"**
```bash
pip install opencv-python
```

**Camera not opening**
- Grant camera permissions to Terminal
- Try different camera indices (code tries 0, 1, 2)
- Use file input option instead

**Poor OCR accuracy**
- Ensure good lighting
- Hold receipt flat and steady
- Use high-resolution camera
- Clean receipt (no wrinkles)

**Model loading slow**
- First run downloads ~500MB model weights
- Subsequent runs load from cache
- Consider using GPU for faster inference

## 📚 Architecture

### System Flow

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│   Frontend  │ ◄─WS──► │  Go Backend  │ ◄─API─► │   Liminal   │
│   (React)   │         │   (Agent)    │         │   Banking   │
└─────────────┘         └──────────────┘         └─────────────┘
                              │
                              │
                        ┌─────▼──────┐
                        │  Anthropic │
                        │    Claude  │
                        └────────────┘

┌─────────────┐         ┌──────────────┐
│   Camera    │ ───────►│   Receipt    │
│             │         │   Scanner    │
└─────────────┘         └──────────────┘
                              │
                        ┌─────▼──────┐
                        │   Donut    │
                        │    OCR     │
                        └────────────┘
```

### Backend Agent Flow

1. User message received via WebSocket
2. Agent processes with Anthropic Claude
3. Agent decides which tools to call
4. Tools execute (banking API, calculations, etc.)
5. Results returned to agent
6. Agent formulates response
7. Response sent to frontend

### Data Storage

- **In-Memory**: Weekly goals, notification tracking
- **JSON Output**: Receipt scanner results
- **No Database**: Stateless design for simplicity

## 🧪 Development

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
go build -o hackathon-agent main.go
./hackathon-agent
```

**Frontend:**
```bash
cd frontend
npm run build
# Output in dist/ directory
```

### Development Tips

- Use `go run main.go` for hot-reloading backend
- Use `npm run dev` for hot-reloading frontend
- Check browser console for WebSocket connection issues
- Use verbose logging for debugging

## 🎯 Use Cases

### Personal Finance Management
- Track daily spending
- Monitor budget adherence
- Analyze spending patterns
- Receipt organization

### Small Business
- Expense tracking
- Receipt digitization
- Category-based reporting
- Budget forecasting

### Financial Education
- Learn banking API integration
- Understand AI agents
- Practice full-stack development
- Explore OCR technology

## 🔮 Future Enhancements

### Planned Features
- [ ] Persistent database storage
- [ ] User authentication
- [ ] Multiple account support
- [ ] Email notifications for budget alerts
- [ ] PDF receipt export
- [ ] Mobile app
- [ ] Historical spending trends
- [ ] Budget recommendations
- [ ] Recurring expense detection
- [ ] Bill payment reminders

### Technical Improvements
- [ ] GraphQL API
- [ ] Redis caching
- [ ] Docker containerization
- [ ] Kubernetes deployment
- [ ] CI/CD pipeline
- [ ] E2E testing
- [ ] Performance monitoring
- [ ] Error tracking (Sentry)

## 📖 API Reference

### Backend WebSocket API

**Connect:**
```javascript
const ws = new WebSocket('ws://localhost:8080/ws');
```

**Send Message:**
```javascript
ws.send(JSON.stringify({
  type: 'user_message',
  content: 'What is my balance?'
}));
```

**Receive Response:**
```javascript
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log(data.content);
};
```

### Receipt Scanner API

**Command Line:**
```bash
python receipt.py
```

**Programmatic:**
```python
from receipt import ReceiptScanner

scanner = ReceiptScanner()
receipt_data = scanner.process_receipt(pil_image)
print(receipt_data)
```

## 🤝 Contributing

Contributions welcome! Areas for contribution:

- Additional spending categories
- More banking tools
- UI/UX improvements
- Documentation enhancements
- Bug fixes
- Test coverage
- Performance optimizations

## 📄 License

This project is provided as-is for educational and commercial use.

## 🔗 Resources

### Documentation
- [Liminal API Docs](https://docs.liminal.cash)
- [nim-go-sdk Documentation](nim-go-sdk/README.md)
- [LangGraph Documentation](https://langchain-ai.github.io/langgraph/)
- [Anthropic Claude API](https://docs.anthropic.com/)

### Related Projects
- [Donut OCR Model](https://huggingface.co/naver-clova-ix/donut-base-finetuned-cord-v2)
- [OpenCV Documentation](https://docs.opencv.org/)
- [React Documentation](https://react.dev/)
- [Go Documentation](https://go.dev/doc/)

## 🙏 Acknowledgments

- **Liminal** - Banking API and SDK
- **Anthropic** - Claude AI model
- **Naver Clova** - Donut OCR model
- **Hugging Face** - Model hosting and transformers library

## 📞 Support

For issues and questions:
- Check the [Troubleshooting](#-troubleshooting) section
- Review [nim-go-sdk examples](nim-go-sdk/examples/)
- Open an issue on GitHub

---

**Built with ❤️ for the Liminal Hackathon**

Happy building! 🚀

# Run a single interaction
result = agent.run("What time is it?")
print(result["messages"][-1].content)

# Multi-turn conversation
conversation_history = []
result1 = agent.run("Calculate 10 + 5", conversation_history)
conversation_history = result1["messages"]

result2 = agent.run("Now multiply that by 2", conversation_history)
print(result2["messages"][-1].content)
```

### Async Usage

```python
import asyncio
from agent import Agent

async def main():
    agent = Agent()
    result = await agent.arun("Tell me about Python")
    print(result["messages"][-1].content)

asyncio.run(main())
```

### Streaming Responses

```python
from agent import Agent

agent = Agent()

for state in agent.stream("Calculate 15 * 23"):
    if "messages" in state:
        # Process state updates
        print(state)
```

## Creating Custom Tools

Add new tools in [tools.py](tools.py):

```python
from langchain_core.tools import tool

@tool
def my_custom_tool(input_param: str) -> str:
    """
    Description of what your tool does.
    
    Args:
        input_param: Description of the parameter
        
    Returns:
        Description of the return value
    """
    # Your tool implementation
    return f"Processed: {input_param}"

# Add to get_available_tools() function
def get_available_tools() -> list:
    return [
        get_current_time,
        calculate,
        search_knowledge_base,
        text_analysis,
        my_custom_tool,  # Add your tool here
    ]
```

## Customizing the Agent

### Change the LLM Model

Modify [config.py](config.py) or set environment variables:

```python
MODEL_NAME=gpt-4
TEMPERATURE=0.5
```

### Modify the Graph Structure

Edit the `_build_graph()` method in [agent.py](agent.py) to:
- Add new nodes
- Change edge conditions
- Implement custom routing logic

### Add Memory/Persistence

Extend the `AgentState` TypedDict to include additional state:

```python
class AgentState(TypedDict):
    messages: Annotated[Sequence[BaseMessage], add_messages]
    user_context: dict  # Add custom state
    iteration_count: int
```

## Architecture

The agent uses LangGraph's StateGraph with the following flow:

```
Entry → Agent (LLM) → Decision
                        ├─→ Tools (if tool calls needed) → Agent
                        └─→ End (if response complete)
```

1. **Agent Node**: Calls the LLM with current state
2. **Decision**: Checks if tools need to be called
3. **Tools Node**: Executes requested tools
4. **Loop**: Returns to agent with tool results
5. **End**: Returns final response

## Configuration Options

| Variable | Default | Description |
|----------|---------|-------------|
| `OPENAI_API_KEY` | Required | Your OpenAI API key |
| `MODEL_NAME` | `gpt-4o-mini` | OpenAI model to use |
| `TEMPERATURE` | `0.7` | Sampling temperature (0-2) |
| `MAX_ITERATIONS` | `10` | Max agent iterations |
| `VERBOSE` | `true` | Enable verbose logging |

## Troubleshooting

### "OPENAI_API_KEY not found" Error

Make sure you've created a `.env` file with your API key:
```
OPENAI_API_KEY=sk-...
```

### Import Errors

Ensure all dependencies are installed:
```bash
pip install -r requirements.txt
```

### Tool Not Working

Verify the tool is:
1. Decorated with `@tool`
2. Has a clear docstring
3. Added to `get_available_tools()` list

## Advanced Features

### Adding Checkpointing

For persistence across sessions, integrate LangGraph's checkpointing:

```python
from langgraph.checkpoint.memory import MemorySaver

memory = MemorySaver()
self.graph = workflow.compile(checkpointer=memory)
```

### Human-in-the-Loop

Add approval nodes for sensitive operations:

```python
workflow.add_node("human_approval", human_approval_node)
workflow.add_edge("tools", "human_approval")
workflow.add_edge("human_approval", "agent")
```

## Examples

See [main.py](main.py) for complete examples including:
- Basic single-turn interactions
- Multi-turn conversations with context
- Streaming responses
- Error handling

## Contributing

Feel free to extend this template with:
- Additional tools
- Different LLM providers
- Enhanced state management
- Custom routing logic
- Memory systems

## License

This template is provided as-is for educational and commercial use.

## Resources

- [LangGraph Documentation](https://langchain-ai.github.io/langgraph/)
- [LangChain Documentation](https://python.langchain.com/)
- [OpenAI API Documentation](https://platform.openai.com/docs/)

## Next Steps

1. Add your custom tools
2. Integrate with your data sources
3. Implement domain-specific logic
4. Add error handling and retry logic
5. Deploy to production environment

Happy building! 🚀
