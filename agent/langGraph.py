from langchain_mcp_adapters.client import MultiServerMCPClient
from langgraph.prebuilt import create_react_agent
from langchain_ollama import ChatOllama

import asyncio

model = ChatOllama(model="gpt-oss:20b")

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

            print(response_stream)

            # # Stream the response
            # async for chunk in response_stream:
            #     if "messages" in chunk:
            #         for message in chunk["messages"]:
            #             # Only print the final assistant message to avoid verbose intermediate steps
            #             if message.role == "assistant" and message.content:
            #                 print("Assistant:", message.content)

        except (KeyboardInterrupt, EOFError):
            break
        except Exception as e:
            print(f"An error occurred: {e}")

    print("\nExiting agent.")
    # Properly close the client connections

if __name__ == "__main__":
    asyncio.run(main())


# from langchain_mcp_adapters.client import MultiServerMCPClient, StdioConnection
# from langgraph.prebuilt import create_react_agent
# from langchain_ollama import ChatOllama

# import asyncio


# model = ChatOllama(model="gpt-oss:20b")

# client = MultiServerMCPClient(
#     {
#         "bookstore": {
#             "command": "uv",
#             "args": ["--directory", "./generated/mcp", "run", "mcp_server.py"],
#             "transport": "stdio",
#         }
#     }
# )


# tools = client.get_tools()

# def get_weather(city: str) -> str:  
#     """Get weather for a given city."""
#     return f"It's always sunny in {city}!"

# agent = create_react_agent(
#     model=model,
#     tools=[get_weather],
#     # tools=[tools]
# )


# asyncio.run(
#     agent.ainvoke(
#     {"messages": [{"role": "user", "content": "hello"}]}
# ))
