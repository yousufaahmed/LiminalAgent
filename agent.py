"""
LangGraph Agentic AI Base Template

This module implements a basic agentic AI system using LangGraph's StateGraph.
The agent can use tools and maintain conversation state across interactions.
"""

from typing import Annotated, TypedDict, Sequence
from langchain_core.messages import BaseMessage, HumanMessage, AIMessage, ToolMessage
from langchain_openai import ChatOpenAI
from langgraph.graph import StateGraph, END
from langgraph.graph.message import add_messages
from langgraph.prebuilt import ToolNode
from tools import get_available_tools


class AgentState(TypedDict):
    """
    State schema for the agent graph.
    
    Attributes:
        messages: List of conversation messages with automatic reduction
    """
    messages: Annotated[Sequence[BaseMessage], add_messages]


class Agent:
    """
    Main agent class that orchestrates the agentic AI workflow using LangGraph.
    """
    
    def __init__(self, model_name: str = "gpt-4o-mini", temperature: float = 0.7):
        """
        Initialize the agent with a language model and tools.
        
        Args:
            model_name: The OpenAI model to use
            temperature: Sampling temperature for the model
        """
        self.tools = get_available_tools()
        self.llm = ChatOpenAI(model=model_name, temperature=temperature)
        self.llm_with_tools = self.llm.bind_tools(self.tools)
        self.graph = self._build_graph()
        
    def _build_graph(self) -> StateGraph:
        """
        Build the LangGraph StateGraph for the agent workflow.
        
        Returns:
            Compiled StateGraph ready for execution
        """
        # Create the graph
        workflow = StateGraph(AgentState)
        
        # Add nodes
        workflow.add_node("agent", self._call_model)
        workflow.add_node("tools", ToolNode(self.tools))
        
        # Set entry point
        workflow.set_entry_point("agent")
        
        # Add conditional edges
        workflow.add_conditional_edges(
            "agent",
            self._should_continue,
            {
                "continue": "tools",
                "end": END
            }
        )
        
        # Add edge from tools back to agent
        workflow.add_edge("tools", "agent")
        
        return workflow.compile()
    
    def _call_model(self, state: AgentState) -> dict:
        """
        Call the language model with the current state.
        
        Args:
            state: Current agent state
            
        Returns:
            Updated state with model response
        """
        messages = state["messages"]
        response = self.llm_with_tools.invoke(messages)
        return {"messages": [response]}
    
    def _should_continue(self, state: AgentState) -> str:
        """
        Determine whether to continue to tools or end the conversation.
        
        Args:
            state: Current agent state
            
        Returns:
            "continue" if tools should be called, "end" otherwise
        """
        messages = state["messages"]
        last_message = messages[-1]
        
        # If there are tool calls, continue to tools node
        if hasattr(last_message, "tool_calls") and last_message.tool_calls:
            return "continue"
        
        # Otherwise, end the conversation
        return "end"
    
    def run(self, user_input: str, conversation_history: list = None) -> dict:
        """
        Run the agent with a user input.
        
        Args:
            user_input: The user's message
            conversation_history: Optional previous messages
            
        Returns:
            Final state containing all messages
        """
        # Initialize messages
        messages = conversation_history or []
        messages.append(HumanMessage(content=user_input))
        
        # Create initial state
        initial_state = {"messages": messages}
        
        # Run the graph
        final_state = self.graph.invoke(initial_state)
        
        return final_state
    
    async def arun(self, user_input: str, conversation_history: list = None) -> dict:
        """
        Async version of run method.
        
        Args:
            user_input: The user's message
            conversation_history: Optional previous messages
            
        Returns:
            Final state containing all messages
        """
        # Initialize messages
        messages = conversation_history or []
        messages.append(HumanMessage(content=user_input))
        
        # Create initial state
        initial_state = {"messages": messages}
        
        # Run the graph asynchronously
        final_state = await self.graph.ainvoke(initial_state)
        
        return final_state
    
    def stream(self, user_input: str, conversation_history: list = None):
        """
        Stream the agent's response.
        
        Args:
            user_input: The user's message
            conversation_history: Optional previous messages
            
        Yields:
            State updates as they occur
        """
        # Initialize messages
        messages = conversation_history or []
        messages.append(HumanMessage(content=user_input))
        
        # Create initial state
        initial_state = {"messages": messages}
        
        # Stream the graph execution
        for state in self.graph.stream(initial_state):
            yield state
