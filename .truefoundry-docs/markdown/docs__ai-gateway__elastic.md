> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Elastic

> Learn how to export LLM Gateway traces to Elastic Cloud using OpenTelemetry integration.

This guide provides instructions for integrating [Elastic Cloud](https://cloud.elastic.co/) with the TrueFoundry AI Gateway to export OpenTelemetry traces.

## What is Elastic Cloud?

Elastic Cloud is a fully managed cloud service that provides the Elastic Stack (Elasticsearch, Kibana, and more) for search, observability, and security solutions. It offers powerful capabilities for analyzing logs, metrics, and traces at scale.

### Key Features of Elastic Cloud

* **[Full-Stack Observability](https://www.elastic.co/observability)**: Unified visibility into logs, metrics, traces, and APM data in a single platform
* **[OpenTelemetry Native](https://www.elastic.co/docs/solutions/observability/get-started/quickstart-elastic-cloud-otel-endpoint)**: Native OTLP endpoint support with Elastic Cloud Managed OTLP Endpoint for seamless OpenTelemetry data ingestion
* **[AI-Powered Analysis](https://www.elastic.co/elasticsearch/elasticsearch-relevance-engine)**: Leverage machine learning and AI capabilities for anomaly detection and intelligent insights

## Prerequisites

Before integrating Elastic Cloud with TrueFoundry, ensure you have:

1. **TrueFoundry Account**: Create a [TrueFoundry account](https://www.truefoundry.com/register) and follow the instructions in our [Gateway Quick Start Guide](https://docs.truefoundry.com/docs/ai-gateway/quick-start)
2. **Elastic Cloud Account**: Sign up for an [Elastic Cloud account](https://cloud.elastic.co/)
3. **Elastic Cloud Deployment**: Create an Elastic Cloud Hosted deployment (version 9.2 or later) or Serverless Observability project
4. **Elastic API Key**: Generate an API key from your Elastic Cloud deployment

## Integration Guide

TrueFoundry AI Gateway supports exporting OpenTelemetry (OTEL) traces to external platforms like Elastic Cloud. This allows you to leverage Elastic's powerful observability features while using TrueFoundry for unified LLM access.

### Step 1: Get Your Elastic Cloud Endpoint and API Key

1. Log in to your [Elastic Cloud console](https://cloud.elastic.co/)
2. Open your deployment or Serverless project
3. Navigate to **Add data** → **Applications** → **OpenTelemetry**
4. Copy the **endpoint** URL (e.g., `https://<your-motlp-endpoint>`)
5. Copy the **API key** value from the authentication headers section

<Note>
  The Elastic Cloud Managed OTLP Endpoint requires Elastic Cloud Hosted deployment version 9.2 or later. For Serverless projects, the endpoint is available by default.
</Note>

### Step 2: Configure OTEL Export in TrueFoundry

Navigate to the TrueFoundry AI Gateway OTEL configuration:

1. Go to **AI Gateway** → **Controls** → **Settings** in the TrueFoundry dashboard
2. Scroll down to the **OTEL Config** section
3. Click the edit button to configure the OTEL exporter

<Frame caption="TrueFoundry AI Gateway Settings - OTEL Config">
  <img src="https://mintcdn.com/truefoundry/hSeQIrH-xOufB9Aw/images/otel-config.png?fit=max&auto=format&n=hSeQIrH-xOufB9Aw&q=85&s=c739def6e0f1bbc7a317e6d9c8499b50" alt="TrueFoundry AI Gateway Controls Settings page showing OTEL Config section" width="1647" height="873" data-path="images/otel-config.png" />
</Frame>

### Step 3: Configure Elastic Cloud Endpoint

Enable the OTEL Traces Exporter and fill in the following configuration:

| Field               | Value                                         |
| ------------------- | --------------------------------------------- |
| **Config Type**     | `http`                                        |
| **Traces endpoint** | `https://<your-motlp-endpoint>` (from Step 1) |
| **Encoding**        | `Json` or `Proto`                             |

<Note>
  Replace `<your-motlp-endpoint>` with the actual endpoint URL copied from your Elastic Cloud deployment. The endpoint is unique to each deployment or Serverless project.
</Note>

### Step 4: Add Required Headers

Click **+ Add Headers** and configure the following HTTP header:

| Header          | Value                   |
| --------------- | ----------------------- |
| `Authorization` | `ApiKey <your-api-key>` |

<Warning>
  Make sure to include the `ApiKey` prefix before your API key value. The format must be `ApiKey <your-api-key>`, not just the API key alone.
</Warning>

### Step 5: Save Configuration

Click **Save** to apply the OTEL export configuration. All LLM traces from the TrueFoundry AI Gateway will now be automatically exported to Elastic Cloud.

### Step 6: View Traces in Elastic Cloud

After making LLM requests through TrueFoundry AI Gateway, log in to your Elastic Cloud deployment to view the traces:

1. Navigate to **Observability** → **APM** → **Services** in Kibana
2. Look for the `tfy-llm-gateway` service to view your LLM Gateway traces
3. Explore the traces including:
   * **Distributed Tracing**: End-to-end trace visualization with the Traces UI
   * **Service Map**: Visual representation of service dependencies
   * **Transaction Details**: Individual request information with timing and metadata

## Configuration Options

### Supported Encoding Formats

Elastic Cloud Managed OTLP Endpoint supports both JSON and Protobuf encoding:

| Encoding | Description                                        |
| -------- | -------------------------------------------------- |
| `Json`   | Human-readable format, easier for debugging        |
| `Proto`  | Binary format, more efficient for high-volume data |

### Additional Resource Attributes

You can optionally configure **Additional Resource Attributes** to append custom attributes to every exported trace. This is useful for adding environment-specific metadata or organizational tags that help with filtering and grouping in Elastic.

### Troubleshooting

#### API Key Prefix Not Found

If you see an error like `"Unauthenticated desc = ApiKey prefix not found"`, ensure your Authorization header is formatted correctly as `ApiKey <your-api-key>` with the `ApiKey` prefix.

#### Too Many Requests (429)

If you encounter HTTP 429 errors, your deployment may be hitting ingest rate limits. Consider upgrading your Elastic Cloud plan or optimizing your trace sampling rate.
