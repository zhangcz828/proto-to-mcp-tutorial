from griptape.structures import Agent
from griptape.tools.mcp.tool import MCPTool
from griptape.tools.mcp.sessions import StdioConnection
from griptape.drivers.prompt.ollama import OllamaPromptDriver
from griptape.utils.chat import Chat

import logging

book_mcp_conn: StdioConnection = {
    "transport": "stdio",
    "command": "uv",
    "args": ["--directory", "./generated/mcp", "run", "mcp_server.py"],
}

book_hcp_tools = MCPTool(name="bookstore", connection=book_mcp_conn)

agent = Agent(
    prompt_driver=OllamaPromptDriver(model="qwen3:8b"),
    tools=[book_hcp_tools]
)

Chat(agent, logger_level=logging.DEBUG).start()
