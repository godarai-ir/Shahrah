> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Create Calculator MCP Server

> Learn how to build a simple calculator MCP server that provides basic math operations using FastMCP.

Let's create a simple calculator MCP server that provides basic math operations.

<CodeGroup>
  ```python server.py expandable lines theme={"dark"}
  from fastmcp import FastMCP

  mcp = FastMCP("Calculator MCP Server")

  @mcp.tool
  def add(a: int, b: int) -> int:
      """Add two numbers"""
      return a + b

  @mcp.tool
  def subtract(a: int, b: int) -> int:
      """Subtract two numbers"""
      return a - b

  @mcp.tool
  def multiply(a: int, b: int) -> int:
      """Multiply two numbers"""
      return a * b

  @mcp.tool
  def divide(a: float, b: float) -> float:
      """Divide two numbers"""
      if b == 0:
          raise ValueError("Cannot divide by zero")
      return a / b

  @mcp.tool
  def power(a: int, b: int) -> int:
      """Raise a number to the power of another number"""
      return a ** b

  if __name__ == "__main__":
      mcp.run(transport="http", host="0.0.0.0", port=8000, path="/mcp")
  ```

  ```txt requirements.txt lines theme={"dark"}
  ```
</CodeGroup>

## Run Locally

```bash lines theme={"dark"}
# Install dependencies
pip install -r requirements.txt

# Run the server
python server.py
```

Your MCP server will be available at `http://localhost:8000/mcp`.

<Tip>
  We have already added this code to a Github Repo and deployed it on our live demo link:

  `Repository:` You can find the complete example code at: [Calculator MCP Server](https://github.com/truefoundry/mcp-servers/tree/main/simple-calculator-mcp)

  `Deployment:` You can find the live demo link at: [Calculator MCP Server Deployment](https://platform.live-demo.truefoundry.cloud/deployments/cmcnc0pml0q2x01qs4hiwh6fb?tab=pods)

  `Server Endpoint:` The MCP server can be accessed at this endpoint: [https://calculator-mcp-server.apps.live-demo.truefoundry.cloud](https://calculator-mcp-server.apps.live-demo.truefoundry.cloud)
</Tip>

## Register in TrueFoundry

To add this MCP server to your MCP group, provide the following details:

| Field              | Value                                                                |
| ------------------ | -------------------------------------------------------------------- |
| **Name**           | `calculator-mcp-server`                                              |
| **Endpoint URL**   | `https://calculator-mcp-server.apps.live-demo.truefoundry.cloud/mcp` |
| **Description**    | A simple calculator MCP server                                       |
| **Authentication** | No Auth                                                              |
