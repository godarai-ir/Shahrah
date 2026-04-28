> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# TrueFoundry AI Gateway Playground

> Experiment with AI models, configure settings, set up guardrails, connect MCP servers, and create reusable prompts in the Playground.

TrueFoundry AI Gateway Playground provides a comprehensive environment to experiment with AI models, fine-tune configurations, configure safety measures, and create reusable prompts. This guide will help you get started quickly with all the essential features.

## How to Access the Playground

To access the AI Gateway Playground:

1. Log into your TrueFoundry dashboard
2. Navigate to the `AI Gateway` section in the main menu
3. Select `Playground` from the dropdown options

<img src="https://mintcdn.com/truefoundry/jw406UAsc7ErYUq8/images/Screenshot2025-08-11at10.18.32PM-min.png?fit=max&auto=format&n=jw406UAsc7ErYUq8&q=85&s=b3caea9c4c29ea0a00af600ad7866867" alt="TrueFoundry AI Gateway Playground navigation" width="3600" height="2016" data-path="images/Screenshot2025-08-11at10.18.32PM-min.png" />

## Playground Interface Overview

The Playground interface combines simplicity with powerful functionality, organized into three main sections:

1. **Top Bar**: Contains Recent History, prompt management, and code generation options
2. **Left Sidebar**: Access to model selection, guardrails, MCP servers, and configuration options
3. **Main Chat Area**: Where you interact with the model and view responses

   <img src="https://mintcdn.com/truefoundry/jw406UAsc7ErYUq8/images/Screenshot2025-08-12at2.23.42PM.png?fit=max&auto=format&n=jw406UAsc7ErYUq8&q=85&s=8e8f7db0332bc504981388366ca09859" alt="Screenshot2025 08 12at2 23 42PM Pn" width="3284" height="2008" data-path="images/Screenshot2025-08-12at2.23.42PM.png" />

### Key Components

* **Model Selection**: Browse and select from over 1,000 models across 15+ providers
* **Interactive Chat Interface**: Send prompts and view model responses in real-time
* **Configuration Controls**: Adjust model parameters and settings
* **Guardrails**: Configure safety measures for inputs and outputs
* **MCP Server Integration**: Connect to MCP Servers for enhanced capabilities
* **Prompt Library**: Save and load prompt templates with their configurations

## Features

<AccordionGroup>
  <Accordion title="Configure Model Settings">
    Fine-tune your model's behavior by adjusting various parameters:

    | Parameter           | Description                                    |
    | ------------------- | ---------------------------------------------- |
    | **Temperature**     | Controls randomness in responses               |
    | **Response Format** | Choose between streaming or complete responses |
    | **Maximum Tokens**  | Sets the length limit for model responses      |
    | **Stop Sequence**   | Defines where the model should stop generating |
    | **Request Logging** | Enables or disables logging of requests        |

    To access these settings, click on the settings icon next to the model selection dropdown. Changes take effect immediately for your next interaction.

    <img src="https://mintcdn.com/truefoundry/jw406UAsc7ErYUq8/images/Screenshot2025-08-11at11.31.42PM-min.png?fit=max&auto=format&n=jw406UAsc7ErYUq8&q=85&s=0e482d5b702c8d4f33fd13a2f357734f" alt="Model configuration settings panel" width="3582" height="2002" data-path="images/Screenshot2025-08-11at11.31.42PM-min.png" />
  </Accordion>

  <Accordion title="Test out Models">
    ### Selecting a Model

    To select a model for your session:

    1. Open the `model dropdown` menu from the left sidebar
    2. Browse through available models or use the search function
    3. Select your desired model

    Models are organized by type (`chat, embedding, rerank, image`). You can filter models based on these categories or search for specific model names.

    <img src="https://mintcdn.com/truefoundry/jw406UAsc7ErYUq8/images/Screenshot2025-08-12at10.42.18AM-min.png?fit=max&auto=format&n=jw406UAsc7ErYUq8&q=85&s=3dad37e2fe0cfe6dad87dd5b5f28da2a" alt="Model selection interface" width="3600" height="2016" data-path="images/Screenshot2025-08-12at10.42.18AM-min.png" />

    ### Interacting with Models

    Once you've selected a model, you can begin interacting with it:

    1. Type your prompt in the text input field
    2. Click the `Run` button to send your prompt to the model
    3. View the model's response in the chat interface

    For multi-turn conversations, simply continue typing in the input field and clicking `Run`. The Playground maintains the conversation context between turns, allowing for natural dialogue

    <img src="https://mintcdn.com/truefoundry/jw406UAsc7ErYUq8/images/Screenshot2025-08-11at11.31.31PM-min.png?fit=max&auto=format&n=jw406UAsc7ErYUq8&q=85&s=f342fd2bca601545ca0306a2862f2391" alt="Basic model interaction in the Playground" width="3592" height="2016" data-path="images/Screenshot2025-08-11at11.31.31PM-min.png" />
  </Accordion>

  <Accordion title="Measure Gateway Latency">
    The Playground provides detailed performance metrics for each request, helping you understand where time is spent during model interactions:

    | Stage               | Description                                                                | Example Time |
    | ------------------- | -------------------------------------------------------------------------- | ------------ |
    | Frontend to Gateway | Time for request to travel from browser to AI Gateway                      | 127.9 ms     |
    | Gateway Processing  | Time spent on rate limiting, budget checks, load balancing, and guardrails | 2.9 ms       |
    | Gateway to Model    | Time for request to reach the AI model provider                            | 468.2 ms     |
    | Model Processing    | Time taken by the model to generate a response                             | 467 ms       |

    This breakdown helps identify bottlenecks and optimize your AI implementation.

    <img src="https://mintcdn.com/truefoundry/qq95eE7VF4xYUrTK/images/Screenshot2025-10-07at7.35.08PM.png?fit=max&auto=format&n=qq95eE7VF4xYUrTK&q=85&s=3ea4750334bf84b86fbf1f5f2fd0453b" alt="Latency breakdown visualization" width="3600" height="2010" data-path="images/Screenshot2025-10-07at7.35.08PM.png" />

    For more detailed timing information, you can inspect the `server-timing` header in your browser's developer tools:

    <img src="https://mintcdn.com/truefoundry/qq95eE7VF4xYUrTK/images/Screenshot2025-10-07at7.34.57PM.png?fit=max&auto=format&n=qq95eE7VF4xYUrTK&q=85&s=c335069f090a4051ab39656010568378" alt="Server timing header in developer tools" width="874" height="460" data-path="images/Screenshot2025-10-07at7.34.57PM.png" />
  </Accordion>

  <Accordion title="Accessing Chat History">
    Access your previous conversations easily:

    1. Click on the `Recent` button in the top bar
    2. Browse through your previous chat sessions
    3. Select any conversation to resume from where you left off

    This feature maintains continuity between sessions and allows you to reference or build upon previous interactions with the model.

    <img src="https://mintcdn.com/truefoundry/jw406UAsc7ErYUq8/images/Screenshot2025-08-12at10.28.56AM-min.png?fit=max&auto=format&n=jw406UAsc7ErYUq8&q=85&s=466d5f36daffb3c879f6d29809942f44" alt="Recent chat history interface" width="3600" height="2012" data-path="images/Screenshot2025-08-12at10.28.56AM-min.png" />
  </Accordion>

  <Accordion title="View Code Snippets to integrate in code">
    ### Generating Code Snippets

    The Playground makes it easy to integrate your configured model into your applications:

    1. Click the `Code` button located in the top bar
    2. Browse through the available code snippets for different libraries and frameworks
    3. Copy the generated code directly into your application

    Available code snippets include: `OpenAI SDK`, `Langchain`, `Langgraph`, `Google ADK`, `Stream API`, `Rest API`, `Go-OpenAI`, `Rust-OpenAI`, `cURL`, `NodeJS`, `llamaindex`, `Langchain4j`

    The generated code includes your current model selection and parameter settings, making it ready to use in your application.

    <img src="https://mintcdn.com/truefoundry/jw406UAsc7ErYUq8/images/Screenshot2025-08-11at10.20.44PM-min.png?fit=max&auto=format&n=jw406UAsc7ErYUq8&q=85&s=72ba1c4df6c4872e86ea74205c1ac558" alt="Code button location in the Playground interface" width="3598" height="2008" data-path="images/Screenshot2025-08-11at10.20.44PM-min.png" />

    <img src="https://mintcdn.com/truefoundry/jw406UAsc7ErYUq8/images/Screenshot2025-08-11at10.21.16PM-min.png?fit=max&auto=format&n=jw406UAsc7ErYUq8&q=85&s=b71ee8f1739eaf1df9d28aae543923f9" alt="Code snippet generation interface with language options" width="2522" height="2008" data-path="images/Screenshot2025-08-11at10.21.16PM-min.png" />
  </Accordion>

  <Accordion title="Test, Edit and Save Prompts">
    ### Loading Saved Prompts

    To use previously created prompts:

    1. Click the `Your Prompts` button in the top bar
    2. Browse through your saved prompts or prompts shared with you
    3. Select the desired prompt to load its complete configuration
    4. The prompt is now ready to use

           <img src="https://mintcdn.com/truefoundry/jw406UAsc7ErYUq8/images/Screenshot2025-08-11at10.21.43PM-min.png?fit=max&auto=format&n=jw406UAsc7ErYUq8&q=85&s=92fedd379314c05d5d442cfcd56041a8" alt="Interface for loading saved prompts" width="3600" height="2010" data-path="images/Screenshot2025-08-11at10.21.43PM-min.png" />

    ### Creating and Saving Prompts

    When you've created a configuration you want to preserve:

    1. Set up your model, parameters, guardrails, and MCP servers as desired
    2. Click the `Save Prompt` button
    3. Choose the repository you want to save your prompt into
    4. Name your prompt and add a commit message
    5. Add optional tags to categorize your prompt
    6. Click `Save` to store your prompt

    Saved prompts are versioned for tracking changes and can be used as templates for new projects. This feature is particularly useful for standardizing AI interactions across your organization.

    <img src="https://mintcdn.com/truefoundry/jw406UAsc7ErYUq8/images/Screenshot2025-08-11at10.24.04PM-min.png?fit=max&auto=format&n=jw406UAsc7ErYUq8&q=85&s=0ceea560e6d68d2ea586ba8489118325" alt="Interface for creating and saving new prompts" width="3600" height="2016" data-path="images/Screenshot2025-08-11at10.24.04PM-min.png" />
  </Accordion>

  <Accordion title="Test MCP Servers">
    MCP (Model Control Protocol) Servers allow your AI models to perform complex operations and interact with external systems.

    ### Setting Up MCP Servers

    To integrate MCP Servers with your Playground session:

    1. Click on `MCP Server` in the left bar of Playground
    2. Choose the MCP Servers and tools you want to use
    3. Connect with the MCP Server if required by entering credentials
    4. The MCP Server is now ready to use

    Once connected, your model can perform operations like managing calendars, retrieving data from databases, or controlling external systems, all through natural language requests.

    <img src="https://mintcdn.com/truefoundry/jw406UAsc7ErYUq8/images/Screenshot2025-08-11at10.22.56PM-min.png?fit=max&auto=format&n=jw406UAsc7ErYUq8&q=85&s=6a6b8a9e65072ac0cc8fc43b81614a73" alt="Interface for adding MCP Servers to the Playground" width="3598" height="2008" data-path="images/Screenshot2025-08-11at10.22.56PM-min.png" />

    ### Using MCP Servers in Conversations

    Once you've connected an MCP Server, you can use it in your conversations:

    1. Send a natural language request like `Clear my calendar for tomorrow`
    2. The model interprets the request and uses the MCP Server to access your calendar
    3. The operation is performed and confirmation is provided

           <img src="https://mintcdn.com/truefoundry/jw406UAsc7ErYUq8/images/Screenshot2025-08-11at11.26.20PM.png?fit=max&auto=format&n=jw406UAsc7ErYUq8&q=85&s=f65a5ed79b3365a82079836445dc8c2d" alt="Example of calendar management using MCP Server integration" width="3600" height="2008" data-path="images/Screenshot2025-08-11at11.26.20PM.png" />
  </Accordion>

  <Accordion title="Configure Guardrails">
    Guardrails help control AI model behavior by filtering or modifying inputs and outputs.

    ### Types and Modes of Guardrails

    The Playground supports two types of guardrails:

    1. **Input Guardrails**: Applied to user prompts before they reach the model
    2. **Output Guardrails**: Applied to model responses before they're shown to the user

    Each guardrail operates in one of two modes:

    1. **Validation Mode**: Blocks requests or responses that fail to meet criteria
    2. **Mutation Mode**: Automatically modifies requests or responses to comply with requirements

    <Note>
      Output Guardrails will only work for non-streaming requests. Input guardrails work on both streaming and non-streaming responses.
    </Note>

    ### Setting Up Guardrails

    To configure guardrails in the Playground:

    1. Select either `Input` or `Output guardrails` from the left bar in Playground
    2. Choose from the dropdown of available guardrails
    3. You can select multiple guardrails for each Input and Output
    4. Send a request to see the guardrails in action

    ### Guardrail Examples

    #### Input Validation Example

    When an input guardrail detects prohibited content, it prevents the request from reaching the model and displays an error message:

    <img src="https://mintcdn.com/truefoundry/jw406UAsc7ErYUq8/images/Screenshot2025-08-11at10.39.35PM-min.png?fit=max&auto=format&n=jw406UAsc7ErYUq8&q=85&s=1837ee4f0e77f28b1ea5e31663c206cb" alt="Example of input guardrail blocking inappropriate content" width="3600" height="2008" data-path="images/Screenshot2025-08-11at10.39.35PM-min.png" />

    #### Output Modification Example

    Output guardrails can automatically transform model responses to meet specific requirements:

    <img src="https://mintcdn.com/truefoundry/jw406UAsc7ErYUq8/images/Screenshot2025-08-11at10.50.20PM-min.png?fit=max&auto=format&n=jw406UAsc7ErYUq8&q=85&s=a355bda5501a0d016c2eabceabfc35bc" alt="Example of output guardrail modifying model response" width="3600" height="2014" data-path="images/Screenshot2025-08-11at10.50.20PM-min.png" />
  </Accordion>
</AccordionGroup>
