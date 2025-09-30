from langchain_mcp_adapters.client import MultiServerMCPClient
from langgraph.prebuilt import create_react_agent
from langchain_ollama import ChatOllama
import asyncio

model = ChatOllama(model="gpt-oss:20b")

import asyncio

client = MultiServerMCPClient(
    {
        "bookstore": {
            "command": "uv",
            "args": ["--directory", "./generated/mcp", "run", "mcp_server.py"],
            "transport": "stdio",
        }
    }
)


async def main():
    tools = await client.get_tools()

    agent = create_react_agent(
        model=model,
        tools=tools,
    )

    print("Agent is ready. Type 'exit' or 'quit' to end the session.")
    while True:
        try:
            user_input = input("You: ")
            if user_input.lower() in ["exit", "quit"]:
                break

            # Use the same input format as the original example
            response_stream = await agent.ainvoke({"messages": [("user", user_input)]})

            print("Output: ", response_stream["messages"][-1].content)

        except (KeyboardInterrupt, EOFError):
            break
        except Exception as e:
            print(f"An error occurred: {e}")

    print("\nExiting agent.")
    # Properly close the client connections


if __name__ == "__main__":
    asyncio.run(main())
