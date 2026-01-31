# LangGraph Agentic AI Template

A production-ready base template for building agentic AI systems using Python and LangGraph. This template provides a clean, modular structure with tool integration, state management, and conversation handling.

## Features

- 🤖 **LangGraph Integration**: Built on LangGraph's StateGraph for reliable agent workflows
- 🛠️ **Tool System**: Extensible tool framework with example implementations
- 💬 **Conversation Management**: Maintains context across multi-turn interactions
- ⚡ **Async Support**: Both synchronous and asynchronous execution modes
- 🌊 **Streaming**: Stream agent responses in real-time
- ⚙️ **Configuration**: Centralized configuration with environment variable support
- 📝 **Type Safety**: Full type hints for better IDE support

## Project Structure

```
LiminalAgenticTool/
├── agent.py              # Main agent implementation with LangGraph
├── tools.py              # Custom tool definitions
├── config.py             # Configuration management
├── main.py               # Example usage and interactive mode
├── requirements.txt      # Python dependencies
├── .env.example          # Example environment variables
└── .gitignore           # Git ignore patterns
```

## Setup

### 1. Clone or Download the Template

```bash
cd LiminalAgenticTool
```

### 2. Create a Virtual Environment (Recommended)

```bash
# Windows
python -m venv venv
venv\Scripts\activate

# macOS/Linux
python -m venv venv
source venv/bin/activate
```

### 3. Install Dependencies

```bash
pip install -r requirements.txt
```

### 4. Configure Environment Variables

Create a `.env` file in the project root:

```bash
cp .env.example .env
```

Edit `.env` and add your OpenAI API key:

```
OPENAI_API_KEY=your_actual_api_key_here
MODEL_NAME=gpt-4o-mini
TEMPERATURE=0.7
```

## Usage

### Basic Usage

Run the example script with predefined interactions:

```bash
python main.py
```

This will:
1. Run several example interactions
2. Start an interactive mode where you can chat with the agent

### Interactive Mode

The agent will respond to your inputs and can use tools like:
- Getting current time
- Performing calculations
- Searching knowledge base
- Analyzing text

Type `quit`, `exit`, or `bye` to exit.

### Using the Agent in Your Code

```python
from agent import Agent

# Initialize the agent
agent = Agent(model_name="gpt-4o-mini", temperature=0.7)

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
