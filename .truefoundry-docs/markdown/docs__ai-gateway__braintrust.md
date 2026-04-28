> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Braintrust

> Guide to integrating Braintrust with the TrueFoundry AI Gateway.

This guide provides instructions for integrating [Braintrust](https://www.braintrust.dev/) with the TrueFoundry AI Gateway to export LLM traces for observability and evaluation.

## What is Braintrust?

Braintrust is an LLM engineering platform that helps teams trace LLM calls, evaluate model performance, run experiments, and debug issues in AI applications.

### Key Features of Braintrust

* **[Comprehensive LLM Tracing](https://www.braintrust.dev/docs/start/observability)**: Captures detailed traces of all LLM interactions including input prompts, outputs, token usage, latency, and costs, providing complete visibility into your AI application's behavior
* **[Evaluation and Experiments](https://www.braintrust.dev/docs/start/evaluation)**: Run systematic evaluations with custom scorers and automated experiments to measure and improve model performance
* **[Real-time Analytics](https://www.braintrust.dev/docs/core/logs)**: Built-in analytics dashboard provides real-time insights into model performance, usage patterns, and costs across your entire LLM stack
* **[Prompt Management](https://www.braintrust.dev/docs/core/functions)**: Version control your prompts and test them in the playground before deployment

## Prerequisites

Before integrating Braintrust with TrueFoundry, ensure you have:

1. **TrueFoundry Account**: Create a [TrueFoundry account](https://www.truefoundry.com/register) and follow the instructions in our [Gateway Quick Start Guide](https://docs.truefoundry.com/docs/ai-gateway/quick-start)
2. **Braintrust Account**: Sign up for a [Braintrust account](https://www.braintrust.dev/)
3. **Braintrust API Key**: Generate an API key from your Braintrust account settings
4. **Braintrust Project ID**: Create a project in Braintrust and note the Project ID

## Integration Guide

TrueFoundry AI Gateway supports exporting OpenTelemetry (OTEL) traces to external platforms like Braintrust. This allows you to leverage Braintrust's powerful evaluation and observability features while using TrueFoundry for unified LLM access.

### Step 1: Get Braintrust Credentials

1. Log in to your [Braintrust dashboard](https://www.braintrust.dev/)
2. Navigate to your project's configuration page
3. Copy your **API Key** from the account settings
4. Copy your **Project ID** from the project configuration page (look for the "Copy Project ID" button)

### Step 2: Configure OTEL Export in TrueFoundry

Navigate to the TrueFoundry AI Gateway OTEL configuration:

1. Go to **AI Gateway** → **Controls** → **Settings** in the TrueFoundry dashboard
2. Scroll down to the **OTEL Config** section and click the edit button
3. Enable the **OTEL Traces Exporter Configuration** toggle
4. Select **HTTP Configuration** tab

<Frame caption="TrueFoundry OTEL Config for Braintrust Integration">
  <img src="https://mintcdn.com/truefoundry/znPG50mDYN4MFVK9/images/2025-12-11_20.12.09.png?fit=max&auto=format&n=znPG50mDYN4MFVK9&q=85&s=305bb66281a904150894c6deefe9e24f" alt="TrueFoundry OTEL Config page showing HTTP configuration with Braintrust endpoint and headers" width="3840" height="1978" data-path="images/2025-12-11_20.12.09.png" />
</Frame>

### Step 3: Configure Braintrust Endpoint

Fill in the following configuration:

| Field               | Value                                       |
| ------------------- | ------------------------------------------- |
| **Traces endpoint** | `https://api.braintrust.dev/otel/v1/traces` |
| **Encoding**        | `Proto`                                     |

### Step 4: Add Required Headers

Click **+ Add Headers** and configure the following HTTP headers:

| Header          | Value                              |
| --------------- | ---------------------------------- |
| `Authorization` | `Bearer <YOUR_BRAINTRUST_API_KEY>` |
| `x-bt-parent`   | `project_id:<YOUR_PROJECT_ID>`     |

Replace `<YOUR_BRAINTRUST_API_KEY>` with your actual Braintrust API key and `<YOUR_PROJECT_ID>` with your Braintrust project ID.

<Note>
  The `x-bt-parent` header sets the trace's parent project. You can use prefixes like `project_id:`, `project_name:`, or `experiment_id:` depending on how you want to organize your traces.
</Note>

### Step 5: Save Configuration

Click **Save** to apply the OTEL export configuration. All LLM traces from the TrueFoundry AI Gateway will now be automatically exported to Braintrust.

### Step 6: View Traces in Braintrust

After making LLM requests through TrueFoundry AI Gateway, log in to your Braintrust dashboard to view the traces:

1. Navigate to your project in Braintrust
2. Go to the **Logs** section
3. View detailed traces including:
   * **LLM Calls**: ChatCompletion, AgentResponse, and other LLM operations
   * **Metrics**: Duration, token usage, and performance data
   * **Metadata**: Request details, model information, and custom attributes
   * **Trace Tree**: Hierarchical view of nested spans and operations

<Frame caption="Braintrust Logs Dashboard showing TrueFoundry traces">
  <img src="https://mintcdn.com/truefoundry/znPG50mDYN4MFVK9/images/2025-12-11_20.09.01.png?fit=max&auto=format&n=znPG50mDYN4MFVK9&q=85&s=3501803e28a8cef05885905828535198" alt="Braintrust logs dashboard displaying LLM traces from TrueFoundry AI Gateway with detailed metrics and trace tree" width="3840" height="1976" data-path="images/2025-12-11_20.09.01.png" />
</Frame>

### Self-Hosted Braintrust

If you're self-hosting Braintrust, use your custom API URL instead of `https://api.braintrust.dev`. For example:

```
https://your-braintrust-instance.example.com/otel/v1/traces
```
