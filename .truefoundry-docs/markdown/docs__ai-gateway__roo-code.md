> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Roo Code

> Learn how to use roo code with TrueFoundry AI Gateway, including setup steps, use cases, and production-ready examples.

This guide provides instructions for integrating [Roo Code](https://github.com/RooCodeInc/Roo-Code) with the Truefoundry AI Gateway.

## What is Roo Code?

Roo Code is an AI-powered VS Code extension that empowers developers to build better software faster with intelligent coding assistance.

### Key Features of Roo Code

1. **Intelligent Code Generation**: Roo Code provides context-aware code completion, function generation, and entire code block creation based on natural language prompts. The extension understands your project context and coding patterns to generate relevant, high-quality code that follows best practices and maintains consistency with your existing codebase.

2. **AI-Powered Debugging and Refactoring**: Advanced debugging assistance helps identify issues, suggest fixes, and explain error messages in plain language. The extension can also refactor existing code for better performance, readability, and maintainability while preserving functionality and following modern coding standards.

3. **Seamless VS Code Integration**: Built specifically for [Visual Studio Code](https://marketplace.visualstudio.com/), Roo Code integrates naturally with your existing development workflow. It works alongside other VS Code extensions, supports multiple programming languages, and provides contextual suggestions without disrupting your coding flow or requiring external tools.

## Prerequisites

Before integrating Roo Code with TrueFoundry, ensure you have:

1. **TrueFoundry Account**: Create a [Truefoundry account](https://www.truefoundry.com/register) and follow the instructions in our [Gateway Quick Start Guide](https://docs.truefoundry.com/gateway/quick-start)
2. **Roo Code Extension**: Install Roo Code from the VS Code Marketplace

## Integration Guide

### Step 1: Install Roo Code Extension

If you haven't already installed Roo Code:

1. **Open VS Code**: Launch Visual Studio Code
2. **Access Extensions**: Click the Extensions icon in the Activity Bar (Ctrl+Shift+X)
3. **Search for Roo Code**: Type "Roo Code" in the search bar
4. **Install Extension**: Click "Install" on the Roo Code extension by RooCode

<Frame caption="Install Roo Code from VS Code Marketplace">
  <img src="https://mintcdn.com/truefoundry/iMid4yIOHvzf4Z4V/images/roo-code-marketplace-install.png?fit=max&auto=format&n=iMid4yIOHvzf4Z4V&q=85&s=507cf8f8f21068ab7a706b8488b4ed5b" alt="VS Code Marketplace showing the Roo Code extension installation page" width="2940" height="1840" data-path="images/roo-code-marketplace-install.png" />
</Frame>

### Step 2: Get TrueFoundry Gateway Configuration

Before configuring Roo Code, gather your TrueFoundry Gateway details:

1. **Navigate to AI Gateway Playground**: Go to your TrueFoundry AI Gateway playground
2. **Access Unified Code Snippet**: Use the unified code snippet
3. **Copy Base URL and Model Name**: You will get both the base URL and model name from the unified code snippet

<Frame>
  <img src="https://mintcdn.com/truefoundry/n3EuZuJ0K8wBFp1G/images/new-code-snippet.png?fit=max&auto=format&n=n3EuZuJ0K8wBFp1G&q=85&s=3634c2dc8c3565fd77ab896d3fd07ed9" alt="TrueFoundry playground showing unified code snippet with base URL and model name" width="2940" height="1664" data-path="images/new-code-snippet.png" />
</Frame>

### Step 3: Configure Roo Code with TrueFoundry Gateway

1. **Open Roo Code Sidebar**: Click the Roo Code icon (🦘) in the VS Code Activity Bar

2. **Access Configuration**: In the welcome screen or settings, look for API provider configuration

3. **Select Custom Provider**: Choose "Custom" or "OpenAI Compatible" from the provider dropdown

4. **Configure TrueFoundry Settings**:
   * **Base URL**: Enter your TrueFoundry Gateway URL (e.g., `{GATEWAY_BASE_URL}`)
   * **Model Name**: Copy the model name from our Gateway playground

<Frame caption="TrueFoundry Configuration in Roo Code">
  <img src="https://mintcdn.com/truefoundry/iMid4yIOHvzf4Z4V/images/roo-code-truefoundry-config.png?fit=max&auto=format&n=iMid4yIOHvzf4Z4V&q=85&s=f54ab60af681ff2787718528468344ad" alt="Roo Code configuration interface with TrueFoundry Gateway settings" width="944" height="1202" data-path="images/roo-code-truefoundry-config.png" />
</Frame>

### Step 4: Complete Configuration

1. **Configure Model**: Enter the model name from the unified code snippet you copied in Step 2
2. **Save Configuration**: Click "Save" or "Let's go!" to apply your settings

**Recommended Models for Roo Code:**

* **`openai-main/gpt-4o`**: Best for complex code generation and reasoning tasks

* **`anthropic-main/claude-3-5-sonnet`**: Excellent for code understanding and following instructions

* Use the exact model name as shown in your TrueFoundry Gateway unified code snippet

### Step 5: Test Your Integration

1. **Test Connection**: Open a code file and try asking Roo Code a question
2. **Verify Functionality**: Ensure responses are generated through TrueFoundry Gateway

<Frame caption="Testing Roo Code with TrueFoundry">
  <img src="https://mintcdn.com/truefoundry/iMid4yIOHvzf4Z4V/images/roo-code-test-integration.png?fit=max&auto=format&n=iMid4yIOHvzf4Z4V&q=85&s=cf13bd02d6c184d7ca4bb4b940accb8a" alt="Roo Code in action with a code generation example using TrueFoundry integration" width="2940" height="1840" data-path="images/roo-code-test-integration.png" />
</Frame>

## Advanced Configuration

### Multiple Model Profiles

Configure different profiles for various use cases:

```json lines theme={"dark"}
{
  "roocode.apiProfiles": [
    {
      "name": "TrueFoundry GPT-4",
      "baseUrl": "{GATEWAY_BASE_URL}",
      "apiKey": "your-truefoundry-token",
      "model": "openai-main/gpt-4o",
      "description": "High-quality code generation"
    },
    {
      "name": "TrueFoundry Fast",
      "baseUrl": "{GATEWAY_BASE_URL}", 
      "apiKey": "your-truefoundry-token",
      "model": "openai-main/gpt-3.5-turbo",
      "description": "Quick responses for simple tasks"
    }
  ]
}
```

### Custom Instructions

Enhance Roo Code's performance with TrueFoundry by setting custom instructions:

```plain Instructions theme={"dark"}
You are an expert software developer working with enterprise-grade code.
Prioritize:
- Security best practices
- Performance optimization
- Clean, maintainable code
- Comprehensive error handling
- Detailed documentation

When suggesting code changes, always consider:
- Code review standards
- Testing requirements
- Scalability implications
```

All your calls from Roo Code will now go through the TrueFoundry Gateway.
