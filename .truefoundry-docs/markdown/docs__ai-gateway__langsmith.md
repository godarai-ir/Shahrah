> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# LangSmith

> Learn how to export LLM Gateway traces to LangSmith using OpenTelemetry integration.

This guide provides instructions for integrating [LangSmith](https://smith.langchain.com/) with the TrueFoundry AI Gateway to export OpenTelemetry traces.

## What is LangSmith?

LangSmith is LangChain's observability and evaluation platform for LLM applications. It provides comprehensive tracing, debugging, and monitoring capabilities to help teams build reliable AI applications.

### Key Features of LangSmith

* **[LLM Tracing](https://docs.smith.langchain.com/observability)**: Capture detailed traces of all LLM interactions including inputs, outputs, latency, and token usage with automatic instrumentation
* **[Evaluation & Testing](https://docs.smith.langchain.com/evaluation)**: Run evaluations on your LLM outputs with custom evaluators, datasets, and automated testing pipelines
* **[Prompt Hub](https://docs.smith.langchain.com/prompt_engineering)**: Version control and manage prompts with collaboration features and A/B testing capabilities

## Prerequisites

Before integrating LangSmith with TrueFoundry, ensure you have:

1. **TrueFoundry Account**: Create a [Truefoundry account](https://www.truefoundry.com/register) and follow the instructions in our [Gateway Quick Start Guide](https://docs.truefoundry.com/gateway/quick-start)
2. **LangSmith Account**: Sign up for a [LangSmith account](https://smith.langchain.com/)
3. **LangSmith API Key**: Generate an API key from your LangSmith settings

## Integration Steps

TrueFoundry AI Gateway supports exporting OpenTelemetry traces to LangSmith, allowing you to monitor and analyze your LLM requests in LangSmith's observability platform.

<Steps>
  <Step title="Get Your LangSmith API Key">
    1. Log into your [LangSmith dashboard](https://smith.langchain.com/)
    2. Navigate to **Settings** → **API Keys**
    3. Create a new API key or copy an existing one
  </Step>

  <Step title="Configure OTEL Export in TrueFoundry">
    1. Go to **AI Gateway** → **Controls** → **Settings** in the TrueFoundry dashboard
    2. Scroll down to the **OTEL Config** section and click the edit button
    3. Enable the **OTEL Traces Exporter Configuration** toggle
    4. Select **HTTP Configuration**
    5. Enter the LangSmith traces endpoint: `https://api.smith.langchain.com/otel/v1/traces`
    6. Set **Encoding** to `Proto`

    <Frame>
      <img src="https://mintcdn.com/truefoundry/EA7V9QWyL2-uEDlk/images/2025-12-09_22.45.08.png?fit=max&auto=format&n=EA7V9QWyL2-uEDlk&q=85&s=0a89d30d3091bc7966da5c66812d79e0" alt="TrueFoundry OTEL Traces Exporter Configuration showing HTTP configuration with LangSmith endpoint" width="3058" height="1170" data-path="images/2025-12-09_22.45.08.png" />
    </Frame>
  </Step>

  <Step title="Configure Headers">
    Add the required header for LangSmith authentication:

    | Header      | Value                  |
    | ----------- | ---------------------- |
    | `x-api-key` | Your LangSmith API key |

    Click **Save** to apply your configuration.
  </Step>

  <Step title="Verify the Integration">
    1. Make some requests through the TrueFoundry AI Gateway
    2. Navigate to the **Monitor** section in TrueFoundry to verify traces are being generated
    3. Log into your LangSmith dashboard and navigate to the **Projects** section
    4. Verify that traces from TrueFoundry are appearing in your project

    <Frame>
      <img src="https://mintcdn.com/truefoundry/OS16LuYN8MU__8RG/images/langsmith-img.png?fit=max&auto=format&n=OS16LuYN8MU__8RG&q=85&s=9d97014119fdfd99502a7f048d086bc4" alt="LangSmith dashboard showing traces exported from TrueFoundry AI Gateway" width="2938" height="1664" data-path="images/langsmith-img.png" />
    </Frame>
  </Step>
</Steps>

## Configuration Options

### LangSmith Endpoint

LangSmith uses a single endpoint for OTEL trace ingestion:

| Configuration       | Value                                            |
| ------------------- | ------------------------------------------------ |
| **Traces Endpoint** | `https://api.smith.langchain.com/otel/v1/traces` |
| **Protocol**        | HTTP                                             |
| **Encoding**        | Proto / JSON                                     |
