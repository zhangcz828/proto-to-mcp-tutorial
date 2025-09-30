#!/usr/bin/env python3
"""
MCP Server for UPM API - Auto-generated from Protocol Buffers

This server provides access to project management operations
through the Model Context Protocol.
"""

import os
import sys
from typing import Any
import json

import httpx
from mcp.server.fastmcp import FastMCP

API_BASE = 'http://localhost:8080'
VERIFY_SSL = False

# Initialize FastMCP
mcp = FastMCP('Bookstore Server')

async def make_api_request(url: str, method: str = "GET", payload: dict = None) -> dict[str, Any] | None:
    """Make a HTTP request to the specified URL."""

    headers = {
        "Content-Type": "application/json",
    }
    # Use SSL verification based on environment variable
    async with httpx.AsyncClient(verify=VERIFY_SSL) as client:
        try:
            if method.upper() == "GET":
                response = await client.get(url, headers=headers, timeout=30.0)
            elif method.upper() == "PUT":
                response = await client.put(url, headers=headers, json=payload, timeout=30.0)
            elif method.upper() == "POST":
                response = await client.post(url, headers=headers, json=payload, timeout=30.0)
            elif method.upper() == "DELETE":
                response = await client.delete(url, headers=headers, timeout=30.0)
            elif method.upper() == "PATCH":
                response = await client.patch(url, headers=headers, json=payload, timeout=30.0)
            else:
                return {"error": f"Unsupported HTTP method: {method}"}
            
            response.raise_for_status()
            
            # Handle DELETE responses that might be empty
            if method.upper() == "DELETE":
                if response.status_code == 200 or response.status_code == 204:
                    return {"success": True, "message": "Resource deleted successfully"}
                
            # Try to parse JSON, return empty dict if no content
            try:
                return response.json()
            except:
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
        
        # Construct the URL
        url = f"{API_BASE}/v1/books/{book_id}"
        
        url = url.replace("{" + "book_id" + "}", str(book_id))
        
        # Prepare payload for non-GET requests
        payload = {}
        
        
        # Make the API request
        result = await make_api_request(url, "GET", payload if payload else None)
        
        
        # Return formatted JSON response
        import json
        return json.dumps(result, indent=2)
        
    except Exception as e:
        # Handle any errors that occur during execution
        import json
        error_result = {
            "error": f"Tool execution failed: {str(e)}",
            "tool_name": "get_book",
            "error_type": type(e).__name__
        }
        return json.dumps(error_result, indent=2)


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
   "book": {
     "bookId": "string", // optional
     "title": "string", // required
     "author": "string", // required
     "pages": int // required
   }
 }
    
    HTTP: POST /v1/books
    
    Parameters:
    - book (object): The book object to create.
    
    Returns:
    - str: JSON formatted response from the API containing the result or error information
    """
    try:
        
        # Construct the URL
        url = f"{API_BASE}/v1/books"
        
        
        # Prepare payload for non-GET requests
        payload = {}
        
        payload["book"] = book
        
        # Make the API request
        result = await make_api_request(url, "POST", payload if payload else None)
        
        
        # Return formatted JSON response
        import json
        return json.dumps(result, indent=2)
        
    except Exception as e:
        # Handle any errors that occur during execution
        import json
        error_result = {
            "error": f"Tool execution failed: {str(e)}",
            "tool_name": "create_book",
            "error_type": type(e).__name__
        }
        return json.dumps(error_result, indent=2)


if __name__ == '__main__':
    # Run the MCP server
    mcp.run()

