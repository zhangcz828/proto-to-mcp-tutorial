from langgraph.prebuilt import create_react_agent
from langchain_ollama import ChatOllama
from pydantic import BaseModel


model = ChatOllama(model="gpt-oss:20b")

class ResponseFormat(BaseModel):
    """Response format for the agent."""
    result: str

def tool() -> None:
    """Testing tool."""
    ...

agent = create_react_agent(
    model,
    tools=[tool],
    response_format=ResponseFormat,
)

# Visualize the graph
# For Jupyter or GUI environments:
agent.get_graph().draw_mermaid_png()

# To save PNG to file:
png_data = agent.get_graph().draw_mermaid_png()
with open("graph.png", "wb") as f:
    f.write(png_data)

# For terminal/ASCII output:
agent.get_graph().draw_ascii()