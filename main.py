"""
Main Entry Point

Example usage of the LangGraph agentic AI system.
"""

from agent import Agent
from config import Config


def main():
    """Main function demonstrating agent usage."""
    
    print("=" * 60)
    print("LangGraph Agentic AI System")
    print("=" * 60)
    print()
    
    # Initialize the agent
    print("Initializing agent...")
    agent = Agent(
        model_name=Config.MODEL_NAME,
        temperature=Config.TEMPERATURE
    )
    print("Agent initialized successfully!")
    print()
    
    # Conversation history
    conversation_history = []
    
    # Example interactions
    examples = [
        "What time is it right now?",
        "Calculate the square root of 144",
        "Tell me about Python programming",
        "Analyze this text: 'LangGraph makes building agentic AI systems simple and powerful!'",
    ]
    
    print("Running example interactions...")
    print("-" * 60)
    print()
    
    for i, user_input in enumerate(examples, 1):
        print(f"Example {i}:")
        print(f"User: {user_input}")
        
        # Run the agent
        result = agent.run(user_input, conversation_history)
        
        # Get the last message (agent's response)
        agent_response = result["messages"][-1]
        print(f"Agent: {agent_response.content}")
        print()
        
        # Update conversation history (optional - for multi-turn conversations)
        # conversation_history = result["messages"]
    
    print("-" * 60)
    print("Examples completed!")
    print()
    
    # Interactive mode
    print("Starting interactive mode (type 'quit' to exit)...")
    print("=" * 60)
    print()
    
    conversation_history = []
    
    while True:
        try:
            user_input = input("You: ").strip()
            
            if not user_input:
                continue
            
            if user_input.lower() in ["quit", "exit", "bye"]:
                print("Goodbye!")
                break
            
            # Run the agent
            result = agent.run(user_input, conversation_history)
            
            # Get the last message
            agent_response = result["messages"][-1]
            print(f"Agent: {agent_response.content}")
            print()
            
            # Update conversation history for context
            conversation_history = result["messages"]
            
        except KeyboardInterrupt:
            print("\n\nGoodbye!")
            break
        except Exception as e:
            print(f"Error: {e}")
            print()


def streaming_example():
    """Example of streaming agent responses."""
    
    print("=" * 60)
    print("Streaming Example")
    print("=" * 60)
    print()
    
    agent = Agent()
    
    user_input = "Calculate 15 * 23 and then tell me the current time"
    print(f"User: {user_input}")
    print("Agent: ", end="", flush=True)
    
    # Stream the response
    for state in agent.stream(user_input):
        # Print state updates
        if "messages" in state:
            messages = state["messages"]
            if messages:
                last_msg = messages[-1]
                # You can process and display intermediate states here
                pass
    
    print()


if __name__ == "__main__":
    # Run main interactive example
    main()
    
    # Uncomment to run streaming example
    # streaming_example()
