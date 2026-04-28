> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Open WebUI

> Learn how to use open webui with TrueFoundry AI Gateway, including setup steps, use cases, and production-ready examples.

This guide provides instructions for integrating [Open WebUI](https://openwebui.com/) with the Truefoundry AI Gateway.

## What is Open WebUI?

Open WebUI is an open-source LLM application interface that enables users to interact with large language models through a user-friendly web UI. With Truefoundry AI Gateway integration, you can route all Open WebUI LLM requests through TrueFoundry's Gateway for enhanced security, routing, rate limiting, cost management, and access controls while benefiting from all the Gateway's built-in features.

## Prerequisites

Before integrating Open WebUI with TrueFoundry, ensure you have:

1. **TrueFoundry Account**: Create a [Truefoundry account](https://www.truefoundry.com/register) and follow the instructions in our [Gateway Quick Start Guide](https://docs.truefoundry.com/gateway/quick-start)
2. **Open WebUI Instance**: Deploy Open WebUI by following the [official documentation](https://github.com/open-webui/open-webui) for local or cloud deployment

## Configuration Steps

This guide assumes you have an active Open WebUI instance and have obtained your TrueFoundry AI Gateway base URL.

### Step 1: Configure Open WebUI to Use TrueFoundry Gateway

1. Open the Open WebUI settings page.

2. Navigate to the "Admin Panel" section from the top right icon.
   <Frame caption="Open WebUI Navigate to Admin Panel">
     <img src="https://mintcdn.com/truefoundry/n3EuZuJ0K8wBFp1G/images/open-webui-1.jpeg?fit=max&auto=format&n=n3EuZuJ0K8wBFp1G&q=85&s=71360b3de21d86f66a0b469c3c959114" alt="Navigate to the Open WebUI Admin Panel from the top right icon" width="1280" height="831" data-path="images/open-webui-1.jpeg" />
   </Frame>

3. Add a new connection from Settings → Connections → Manage OpenAI API Connections (+).

<Frame caption="Open WebUI Admin Panel">
  <img src="https://mintcdn.com/truefoundry/n3EuZuJ0K8wBFp1G/images/open-webui-2.jpeg?fit=max&auto=format&n=n3EuZuJ0K8wBFp1G&q=85&s=a2faad0b115d40cf2b73a079cbe7eb4e" alt="Open WebUI Admin Panel showing the Connections section" width="1280" height="776" data-path="images/open-webui-2.jpeg" />
</Frame>

4. Add the following details:

* **Base URL and Model ID**: You will get both the base URL and model name from the unified code snippet

<Frame>
  <img src="https://mintcdn.com/truefoundry/n3EuZuJ0K8wBFp1G/images/new-code-snippet.png?fit=max&auto=format&n=n3EuZuJ0K8wBFp1G&q=85&s=3634c2dc8c3565fd77ab896d3fd07ed9" alt="TrueFoundry playground showing unified code snippet with base URL and model name" width="2940" height="1664" data-path="images/new-code-snippet.png" />
</Frame>

* **Model ID**: Add the model IDs from the unified code snippet (Press the + button on the right side to add a model.)

<Frame caption="Connection Form">
  <img src="https://mintcdn.com/truefoundry/n3EuZuJ0K8wBFp1G/images/open-webui-3.jpeg?fit=max&auto=format&n=n3EuZuJ0K8wBFp1G&q=85&s=63b84cb3a781447f7e6e8d349aea6cb2" alt="Open WebUI connection form with TrueFoundry configuration settings" width="1280" height="768" data-path="images/open-webui-3.jpeg" />
</Frame>

5. Click the "Save" button.

### Step 2: Test the Integration

1. Open a new chat in Open WebUI.

2. Select the model you want to use from the dropdown on the top left corner.

3. Send a prompt to verify that requests are routed through TrueFoundry's Gateway.

Your Open WebUI instance is now integrated with TrueFoundry's AI Gateway and ready for secure, compliant LLM usage.
