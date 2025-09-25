#!/usr/bin/env python3
"""
MCP Server for BookstoreService - Auto-generated from Protocol Buffers
"""

from typing import Any, Dict, Optional, List
import httpx
from mcp.server.fastmcp import FastMCP

# Configuration
VERIFY_SSL: bool = False

# Initialize FastMCP
mcp = FastMCP("BookstoreService MCP Server")

async def make_api_request(url: str, method: str = "GET", payload: Optional[dict] = None) -> Dict[str, Any]:
    """Make an HTTP request to the service API."""
    headers = {"Content-Type": "application/json"}
    
    async with httpx.AsyncClient(verify=VERIFY_SSL) as client:
        try:
            if method.upper() == "GET":
                response = await client.get(url, headers=headers, timeout=30.0)
            elif method.upper() == "POST":
                response = await client.post(url, headers=headers, json=payload, timeout=30.0)
            elif method.upper() == "PUT":
                response = await client.put(url, headers=headers, json=payload, timeout=30.0)
            elif method.upper() == "DELETE":
                response = await client.delete(url, headers=headers, timeout=30.0)
            elif method.upper() == "PATCH":
                response = await client.patch(url, headers=headers, json=payload, timeout=30.0)
            else:
                return {"error": f"Unsupported HTTP method: {method}"}

            response.raise_for_status()
            
            try:
                return response.json()
            except Exception:
                return {"success": True}

        except httpx.HTTPStatusError as e:
            return {"error": f"HTTP {e.response.status_code}: {e.response.text}"}
        except Exception as e:
            return {"error": str(e)}

# MCP Tools

@mcp.tool()
async def get_book(book_id: str) -> str:
    """Get a book by ID

    HTTP: GET /v1/books/{book_id}
    
    Parameters:
    - book_id (string): The ID of the book to retrieve

    Returns:
    - str: JSON formatted response from the API containing the result or error information
    """
    try:
        url = "http://localhost:8080/v1/books/{book_id}"
        
        # Replace path parameters
        url = url.replace("{book_id}", str(book_id))
        
        
        # Prepare request body
        payload = None
        if "GET" not in ["GET", "DELETE"]:
            body = {}
            if body:
                payload = body
        
        result = await make_api_request(url, "GET", payload)
        
        import json
        return json.dumps(result, indent=2)
        
    except Exception as e:
        import json
        return json.dumps({"error": f"Tool execution failed: {str(e)}"}, indent=2)


@mcp.tool()
async def create_book(book: dict) -> str:
    """Create a new book in the system.

    INSTRUCTIONS:
    1. For each required field:
        - If the user has not provided a value , prompt the user to supply it (otherwise the request will fail).
    2. For optional fields:
        - If not set by the user, do not set the field in the request and omit them.

    Example payload for creating a book:
    {
        "book": {                // required
            "bookId": "string", // optional
            "title": "string", // required
            "author": "string", // required
            "pages": int // required
        }
    }

    HTTP: POST /v1/books
    
    Parameters:
    - book (object, required): The book object to create.

    Returns:
    - str: JSON formatted response from the API containing the result or error information
    """
    try:
        url = "http://localhost:8080/v1/books"
        
        
        # Prepare request body
        payload = {
            "book": book
        }

        # Make the API request
        result = await make_api_request(url, "POST", payload if payload else None)
        
        import json
        return json.dumps(result, indent=2)
        
    except Exception as e:
        import json
        return json.dumps({"error": f"Tool execution failed: {str(e)}"}, indent=2)


if __name__ == '__main__':
    mcp.run()

