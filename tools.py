"""
Tools Module

Define custom tools that the agent can use.
Each tool should be decorated with @tool and have a clear docstring.
"""

from langchain_core.tools import tool
from datetime import datetime
import math


@tool
def get_current_time() -> str:
    """
    Get the current date and time.
    
    Returns:
        Current date and time in a readable format
    """
    return datetime.now().strftime("%Y-%m-%d %H:%M:%S")


@tool
def calculate(expression: str) -> str:
    """
    Evaluate a mathematical expression safely.
    
    Args:
        expression: A mathematical expression to evaluate (e.g., "2 + 2", "sqrt(16)")
        
    Returns:
        The result of the calculation as a string
    """
    try:
        # Create a safe namespace with math functions
        safe_dict = {
            "abs": abs,
            "round": round,
            "min": min,
            "max": max,
            "sum": sum,
            "pow": pow,
            "sqrt": math.sqrt,
            "sin": math.sin,
            "cos": math.cos,
            "tan": math.tan,
            "pi": math.pi,
            "e": math.e,
        }
        
        # Evaluate the expression
        result = eval(expression, {"__builtins__": {}}, safe_dict)
        return str(result)
    except Exception as e:
        return f"Error calculating expression: {str(e)}"


@tool
def search_knowledge_base(query: str) -> str:
    """
    Search a simulated knowledge base for information.
    
    Args:
        query: The search query
        
    Returns:
        Relevant information from the knowledge base
    """
    # This is a mock implementation - replace with real KB search
    knowledge_base = {
        "python": "Python is a high-level, interpreted programming language known for its simplicity and readability.",
        "langgraph": "LangGraph is a library for building stateful, multi-actor applications with LLMs, using graph-based workflows.",
        "ai": "Artificial Intelligence (AI) refers to the simulation of human intelligence in machines programmed to think and learn.",
    }
    
    query_lower = query.lower()
    for key, value in knowledge_base.items():
        if key in query_lower:
            return value
    
    return f"No information found for query: {query}"


@tool
def text_analysis(text: str) -> str:
    """
    Analyze text and return basic statistics.
    
    Args:
        text: The text to analyze
        
    Returns:
        Analysis results including word count, character count, etc.
    """
    words = text.split()
    word_count = len(words)
    char_count = len(text)
    char_count_no_spaces = len(text.replace(" ", ""))
    sentence_count = text.count(".") + text.count("!") + text.count("?")
    
    analysis = f"""
Text Analysis Results:
- Word count: {word_count}
- Character count: {char_count}
- Characters (no spaces): {char_count_no_spaces}
- Estimated sentences: {sentence_count}
- Average word length: {char_count_no_spaces / max(word_count, 1):.2f} characters
"""
    return analysis.strip()


def get_available_tools() -> list:
    """
    Get all available tools for the agent.
    
    Returns:
        List of tool instances
    """
    return [
        get_current_time,
        calculate,
        search_knowledge_base,
        text_analysis,
    ]
