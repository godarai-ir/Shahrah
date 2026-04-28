> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Export OpenTelemetry Data

> Learn how to export OpenTelemetry traces and metrics with TrueFoundry AI Gateway for comprehensive observability.

AI Gateway is **OpenTelemetry (OTEL) compliant**, making it easy to integrate with modern observability tools and platforms. **Tracing** and **metrics** can each be exported from the TrueFoundry dashboard to external OTEL backends. For Prometheus scraping of `/metrics`, labels, and the full metric catalog on self-hosted gateways, see [Prometheus & Grafana integration](/docs/ai-gateway/prometheus-grafana-integration).

***

## Tracing

<Callout icon="lightbulb" color="#FFC107" iconType="regular">The **OTEL Traces Exporter** sends only trace data to your OTEL platform, not metrics. For OpenTelemetry metrics export, see [Metrics](#metrics) below.</Callout>

OpenTelemetry tracing allows you to capture detailed traces of requests as they flow through the AI Gateway. This enables debugging, performance analysis, and end-to-end visibility. Traces storage is managed by TrueFoundry.

Optionally, traces can also be configured to be exported to any existing OTEL compatible observability platform. This will not affect the storage of traces managed by TrueFoundry.

<Frame caption="AI Gateway - OpenTelemetry Tracing Flow">
  <img src="https://mintcdn.com/truefoundry/JASo0tW1KD8H_FG8/images/gateway-traces-export-flow.png?fit=max&auto=format&n=JASo0tW1KD8H_FG8&q=85&s=a35942e49eb7a97e007ef24058b3de7f" alt="OpenTelemetry traces storage and export flow in TrueFoundry" width="3415" height="410" data-path="images/gateway-traces-export-flow.png" />
</Frame>

### Export Traces to OTEL Compatible Observability Platforms

To export spans to your existing OTEL Platform:

<Frame caption="AI Gateway - OpenTelemetry Tracing Export Config Form">
  <img
    src="https://mintcdn.com/truefoundry/f8fMxCx2dWRgKr74/images/gateway-otel-config-traces.png?fit=max&auto=format&n=f8fMxCx2dWRgKr74&q=85&s=854525003e069802031e586a083c9d0a"
    alt="OpenTelemetry Config form for exporting spans to existing OTEL platform
request with spans"
    width="3024"
    height="1726"
    data-path="images/gateway-otel-config-traces.png"
  />
</Frame>

1. Go to **AI Gateway** → **Controls** → **Settings** in the TrueFoundry dashboard
2. Scroll down to the **OTEL Config** section and click the edit button
3. Enable the **OTEL Traces Exporter Configuration** toggle
4. Select either **HTTP Configuration** or **gRPC Configuration** based on your OTEL platform's endpoint
5. Enter the complete traces endpoint for eg., the traces endpoint for Grafana Cloud is `https://otlp-gateway-prod-ap-south-1.grafana.net/otlp/v1/traces`
6. Select the encoding (only for HTTP Configuration) as either `proto` or `json`
7. Optionally, enter the headers required to export spans to the OTEL provider for eg., authorization, region, etc..

***

## Metrics

<Callout icon="lightbulb" color="#FFC107" iconType="regular">The **OTEL Metrics Exporter** sends only metrics to your OTEL platform, not traces. For OpenTelemetry traces export, see [Tracing](#tracing) above.</Callout>

OpenTelemetry metrics export lets you send gateway metrics to an OTLP-compatible observability platform. Configure this in the same **OTEL Config** editor as traces; it does not replace metrics stored and shown inside TrueFoundry.

<Frame caption="AI Gateway - OpenTelemetry Metrics Flow">
  <img src="https://mintcdn.com/truefoundry/f8fMxCx2dWRgKr74/images/gateway-metrics-export-flow.png?fit=max&auto=format&n=f8fMxCx2dWRgKr74&q=85&s=03f6d29a5cbab1e052095d850b051006" alt="OpenTelemetry metrics storage and export flow in TrueFoundry" width="2306" height="284" data-path="images/gateway-metrics-export-flow.png" />
</Frame>

### Export Metrics to OTEL Compatible Observability Platforms

To export metrics to your existing OTEL platform:

<Frame caption="AI Gateway - OpenTelemetry Metrics Export Config Form">
  <img src="https://mintcdn.com/truefoundry/f8fMxCx2dWRgKr74/images/gateway-otel-config-metrics.png?fit=max&auto=format&n=f8fMxCx2dWRgKr74&q=85&s=a1f0700e7d45f18189274bab5b9c32a3" alt="OpenTelemetry Config form for exporting metrics to an OTEL platform" width="3024" height="1726" data-path="images/gateway-otel-config-metrics.png" />
</Frame>

1. Go to **AI Gateway** → **Controls** → **Settings** in the TrueFoundry dashboard
2. Scroll down to the **OTEL Config** section and click the edit button
3. Enable the **OTEL Metrics Exporter Configuration** toggle
4. Select either **HTTP Configuration** or **gRPC Configuration** based on your OTEL platform’s metrics endpoint
5. Enter the complete **metrics** endpoint—for example, Datadog uses `https://otlp.datadoghq.com/v1/metrics`
6. Select the encoding (only for HTTP Configuration) as either `proto` or `json`
7. Optionally, enter the headers required to export metrics to the OTEL provider—for example, API keys or vendor-specific headers
8. Click **Save**

#### Example: Exporting Metrics to Datadog

Datadog supports OTLP metrics ingest natively. Use the following configuration:

**Step 1 — Set the environment variable on your `tfy-llm-gateway` deployment**

Datadog's OTLP ingest endpoint only accepts **delta-encoded metrics**. Set this environment variable in your `tfy-llm-gateway` deployment before configuring the exporter:

```
OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE=Delta
```

**Step 2 — Configure the OTEL Metrics Exporter in TrueFoundry**

In the **OTEL Config** editor, enable **OTEL Metrics Exporter Configuration** and set:

| Field            | Value                                   |
| ---------------- | --------------------------------------- |
| Protocol         | HTTP Configuration                      |
| Metrics Endpoint | `https://otlp.datadoghq.com/v1/metrics` |
| Encoding         | `proto`                                 |

Add the following headers:

| Header                  | Value                                     |
| ----------------------- | ----------------------------------------- |
| `dd-api-key`            | `<your-datadog-api-key>`                  |
| `dd-otel-metric-config` | `{"histograms":{"mode":"distributions"}}` |

The `dd-otel-metric-config` header enables the **histogram translator**, which converts OTLP histogram metrics into Datadog distribution metrics. This is required to query percentile aggregations such as `p50:ai_gateway_request_processing_ms{*}`, `p95:ai_gateway_request_processing_ms{*}`, and similar metrics in the Datadog Metrics Explorer.

<Callout icon="lightbulb" color="#FFC107" iconType="regular">
  For sites other than US1, use the region-specific endpoint—for example, `https://otlp.datadoghq.eu/v1/metrics` for the EU site. See the [Datadog OTLP metrics ingest documentation](https://docs.datadoghq.com/opentelemetry/setup/otlp_ingest/metrics/?tab=javascript) for the full list of endpoints and additional `dd-otel-metric-config` options.
</Callout>

***

For scraping `/metrics` with Prometheus, OTLP endpoint patterns, and a full list of metric names and labels, see [Prometheus & Grafana integration](/docs/ai-gateway/prometheus-grafana-integration).
