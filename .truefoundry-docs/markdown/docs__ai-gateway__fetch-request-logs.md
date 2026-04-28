> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# API Access to Logs

> Learn how to fetch Gateway request logs using the spans query API for different use cases.

The LLM Gateway delivers detailed request logs through tracing. Retrieve these logs using the [Query Spans API](/docs/truefoundry_sdk/traces#query-spans).

<Note>
  To better understand tracing concepts like traces, spans, and how they work together, see our [tracing overview](/docs/tracing/overview).
</Note>

## Quickstart

<Tabs>
  <Tab title="Using TrueFoundry SDK">
    ### Setup the TrueFoundry SDK

    To start querying request logs, install and configure the **TrueFoundry SDK and CLI**.\
    Follow the [CLI Setup Guide](/docs/setup-cli) for installation instructions and authentication steps.

    Once setup is complete, you can use the SDK to query tracing data programmatically.

    ### Fetch using TrueFoundry SDK

    Each request to the LLM Gateway generates a **trace**—a timeline of everything that happened, from the incoming request to guardrails to the model call and any external APIs.
    Let's pull the latest Gateway traces and see real data quickly.

    Fetch the latest LLM Gateway request logs:

    <Note>
      You can get the `tracing_project_fqn` from the `Fetch via API` button on the Request Logs page

      <img src="https://mintcdn.com/truefoundry/E4NNBYXjH9Zfztxa/images/request-logs-fetch-via-api.png?fit=max&auto=format&n=E4NNBYXjH9Zfztxa&q=85&s=08df5e487e0151284f08c7aed16a6faf" alt="" width="3006" height="614" data-path="images/request-logs-fetch-via-api.png" />
    </Note>

    ```python lines theme={"dark"}
    from truefoundry import client

    # Fetch LLM Gateway request logs
    spans = client.traces.query_spans(
        data_routing_destination="default",
        start_time="2025-01-21T00:00:00.000Z",
    )

    for span in spans:
        print(span.span_name, span.span_attributes.get('tfy.span_type'))
    ```
  </Tab>

  <Tab title="Using HTTP API">
    ### Authentication

    To start querying request logs, you need to authenticate with your TrueFoundry API key. You can use either a Personal Access Token **(PAT)** or Virtual Account Token **(VAT)**.

    <Accordion title="Get your API key">
      To generate an API key:

      1. **Personal Access Token (PAT)**: Go to Access → Personal Access Tokens in your TrueFoundry dashboard
      2. **Virtual Account Token (VAT)**: Go to Access → Virtual Account Tokens (requires admin permissions)

      For detailed authentication setup, see our [Authentication guide](/docs/ai-gateway/authentication).
    </Accordion>

    ### Fetch using HTTP API

    Each request to the LLM Gateway generates a **trace**—a timeline of everything that happened, from the incoming request to guardrails to the model call and any external APIs.
    Let's pull the latest Gateway traces and see real data quickly.

    Fetch the latest LLM Gateway request logs:

    <Note>
      You can get the `tracing_project_fqn` from the `Fetch via API` button on the Request Logs page

      <img src="https://mintcdn.com/truefoundry/E4NNBYXjH9Zfztxa/images/request-logs-fetch-via-api.png?fit=max&auto=format&n=E4NNBYXjH9Zfztxa&q=85&s=08df5e487e0151284f08c7aed16a6faf" alt="" width="3006" height="614" data-path="images/request-logs-fetch-via-api.png" />
    </Note>

    ```python lines theme={"dark"}
    import requests

    page_token = None
    done = False
    while not done:
        # Make API request
        response = requests.post(
            "https://{control_plane_url}/api/svc/v1/spans/query",
            headers={
                "Authorization": "Bearer YOUR_API_TOKEN",
                "Content-Type": "application/json"
            },
            json={
                "dataRoutingDestination": "default",
                "startTime": "2025-10-08T20:16:54",
                "pageToken": page_token
            }
        )
        
        response.raise_for_status()
        data = response.json()

        for span in data['data']:
            print(span['spanName'], span['spanAttributes'].get('tfy.span_type', ''))
        
        page_token = data['pagination'].get("nextPageToken")
        done = page_token is None

    print("Fetch spans completed!")
    ```
  </Tab>
</Tabs>

## Deep Dive: Inspect a Single Trace

Now that you've run a basic query, inspect one request end-to-end. The examples below fetch all spans for a specific `trace_id`, so you can see the full hierarchy (root request, guardrail processing, model span, and outbound HTTP calls).

<CodeGroup>
  ```python TrueFoundry SDK lines theme={"dark"}
  from truefoundry import client
  from truefoundry_sdk import SortDirection

  # Fetch all spans for a specific trace ID with guardrail processing
  trace_id = "019a047ee43577009b6bb5b6ab9477d2"
  spans = client.traces.query_spans(
      data_routing_destination="default",
      start_time="2025-10-21T00:00:00.000Z",
      end_time="2025-10-21T23:59:59.999Z",
      trace_ids=["019a047ee43577009b6bb5b6ab9477d2"],
      limit=200,
      sort_direction=SortDirection.DESC
  )

  for span in spans:
      print(span)
  ```

  ```python HTTP API theme={"dark"}
  import requests

  # Fetch all spans for a specific trace ID with guardrail processing
  trace_id = "019a047ee43577009b6bb5b6ab9477d2"
  page_token = None
  done = False

  while not done:
      # Make API request
      response = requests.post(
          "https://{control_plane_url}/api/svc/v1/spans/query",
          headers={
              "Authorization": "Bearer YOUR_API_TOKEN",
              "Content-Type": "application/json"
          },
          json={
              "dataRoutingDestination": "default",
              "startTime": "2025-10-21T00:00:00.000Z",
              "endTime": "2025-10-21T23:59:59.999Z",
              "traceIds": ["019a047ee43577009b6bb5b6ab9477d2"],
              "limit": 200,
              "sortDirection": "desc",
              "pageToken": page_token
          }
      )
      
      response.raise_for_status()
      data = response.json()

      for span in data['data']:
          print(span)
      
      page_token = data['pagination'].get("nextPageToken")
      done = page_token is None

  print("Fetch spans for trace completed!")
  ```
</CodeGroup>

The following trace contains **5 spans** that form a hierarchical relationship, demonstrating a request flow with PII redaction guardrail processing:

<Frame caption="Chat completion request with guardrail processing span hierarchy">
  <img src="https://mintcdn.com/truefoundry/T91LoM0hMaH4kV7M/images/gateway-tracing-overview.png?fit=max&auto=format&n=T91LoM0hMaH4kV7M&q=85&s=d7cb97bf79ed29a9d505fa599d35da68" width="1753" height="311" data-path="images/gateway-tracing-overview.png" />
</Frame>

### Span Hierarchy Breakdown

The following example demonstrates a complete trace with 5 spans that form a hierarchical relationship. Each span represents a different phase of the request processing, from the initial chat completion request through guardrail processing to the final model inference and network calls.

<AccordionGroup>
  <Accordion title="ChatCompletion Span (Root Span)">
    The ChatCompletion span with ID `bddb6503c0eeb940` serves as the root span with no parent span ID (empty parent\_span\_id).
    This span represents the complete chat completion request lifecycle from the client's perspective, capturing the total time from when the request enters the gateway until the response is sent back to the client.
    With a duration of 7.09 seconds, it provides the overall performance measurement for the entire request flow.
    The span includes the `tfy.triggered_guardrail_fqns` attribute showing which guardrails were triggered during processing.

    <Expandable title="View ChatCompletion Span Details" defaultOpen="true">
      ```shell theme={"dark"}
      TraceSpan(
          span_id='bddb6503c0eeb940',
          trace_id='019a047ee43577009b6bb5b6ab9477d2',
          parent_span_id='',
          service_name='tfy-llm-gateway',
          span_name='ChatCompletion: openai-main/gpt-5',
          timestamp='2025-10-21T02:00:16.181Z',
          duration_ns=7087251623.0,
          status_code='Unset',
          span_attributes={
              'tfy.model.fqn': 'truefoundry:openai:openai-main:model:gpt-5',
              'tfy.model.request_type': 'ChatCompletion',
              'tfy.error_message': '',
              'tfy.model.name': 'openai-main/gpt-5',
              'tfy.model.id': 'cme6zartb0v0701pjeqk9fulg',
              'tfy.triggered_guardrail_fqns': ['truefoundry:guardrail-config-group:pii-guardrails:guardrail-config:pii-redaction'],
              'tfy.model.metric.inter_token_latency_in_ms': 0,
              'tfy.model.metric.time_to_first_token_in_ms': 7022.75,
              'tfy.model.metric.latency_in_ms': 6600.23,
              'tfy.model.metric.input_tokens': 12,
              'tfy.model.metric.cost_in_usd': 0.00481501,
              'tfy.model.metric.output_tokens': 480,
              'tfy.input': '{"model":"openai-main/gpt-5","messages":[{"role":"user","content":"I am sateesh. Hi"}],"stream":true}',
              'tfy.output': '{"id":"chatcmpl-CSvxYlQKenRQWspGHQaCLRcz1MbZU","usage":{"prompt_tokens":12,"completion_tokens":480,"cache_read_tokens":0,"cache_write_tokens":0,"reasoning_tokens":448,"total_tokens":492},"object":"chat.completion","model":"gpt-5-2025-08-07","created":1761012016,"choices":[{"finish_reason":"stop","index":0,"logprobs":null,"message":{"role":"assistant","content":"Hi! Nice to meet you. What would you like me to call you, and how can I help today?","refusal":null}}]}',
              'tfy.span_type': 'ChatCompletion'
          },
          events=[],
          created_by_subject=Subject(
              subject_id='sateesh@truefoundry.com',
              subject_type=<SubjectType.USER: 'user'>,
              subject_slug='sateesh@truefoundry.com',
              subject_display_name=None
          )
      )
      ```
    </Expandable>

    <AccordionGroup>
      <Accordion title="Guardrail Span">
        The Guardrail span with ID `d46e8d5202edc22c` has a parent span "ChatCompletion Span (Root Span)" with ID `bddb6503c0eeb940`, representing the PII redaction guardrail processing.
        This span shows the guardrail configuration used.

        <Expandable title="View Guardrail Span Details" defaultOpen="true">
          ```shell theme={"dark"}
          TraceSpan(
              span_id='d46e8d5202edc22c',
              trace_id='019a047ee43577009b6bb5b6ab9477d2',
              parent_span_id='bddb6503c0eeb940',
              service_name='tfy-llm-gateway',
              span_name='Guardrail: pii-guardrails/pii-redaction',
              timestamp='2025-10-21T02:00:16.183Z',
              duration_ns=483181280.0,
              status_code='Ok',
              span_attributes={
                  'tfy.guardrail.fqn': 'truefoundry:guardrail-config-group:pii-guardrails:guardrail-config:pii-redaction',
                  'tfy.guardrail.name': 'pii-guardrails/pii-redaction',
                  'tfy.guardrail.id': 'cmd4avynw05xa01mo1oli6bsj',
                  'tfy.guardrail.result': 'mutation',
                  'tfy.guardrail.applied_on_entity.type': 'model',
                  'tfy.guardrail.applied_on_entity.id': 'cme6zartb0v0701pjeqk9fulg',
                  'tfy.guardrail.applied_on_entity.name': 'openai-main/gpt-5',
                  'tfy.guardrail.applied_on_entity.fqn': 'truefoundry:openai:openai-main:model:gpt-5',
                  'tfy.guardrail.applied_on_entity.scope': 'input',
                  'tfy.guardrail.metric.latency_in_ms': 482.98,
                  'tfy.span_type': 'Guardrail'
              },
              events=[],
              created_by_subject=Subject(
                  subject_id='sateesh@truefoundry.com',
                  subject_type=<SubjectType.USER: 'user'>,
                  subject_slug='sateesh@truefoundry.com',
                  subject_display_name=None
              )
          )
          ```
        </Expandable>

        <Accordion title="Guardrail Network Call Span">
          The Guardrail Network Call span with ID `fb07005b3c28a98b` has a parent span "Guardrail Span" with ID `d46e8d5202edc22c`, representing the actual HTTP communication with the external guardrail service (AWS Bedrock). This span captures the network latency and external guardrail service processing time. With a duration of 0.48 seconds, it shows the time spent on the actual guardrail API call.

          <Expandable title="View Guardrail Network Call Span Details" defaultOpen="true">
            ```shell theme={"dark"}
            TraceSpan(
                span_id='fb07005b3c28a98b',
                trace_id='019a047ee43577009b6bb5b6ab9477d2',
                parent_span_id='d46e8d5202edc22c',
                service_name='tfy-llm-gateway',
                span_name='POST https://bedrock-runtime.eu-central-1.amazonaws.com/guardrail/o8g71agqe4cl/version/1/apply',
                timestamp='2025-10-21T02:00:16.184Z',
                duration_ns=481350372.0,
                status_code='Unset',
                span_attributes={
                    'url.path': '/guardrail/o8g71agqe4cl/version/1/apply',
                    'http.request.method': 'POST',
                    'url.scheme': 'https',
                    'network.peer.address': '18.194.132.160',
                    'url.full': 'https://bedrock-runtime.eu-central-1.amazonaws.com/guardrail/o8g71agqe4cl/version/1/apply',
                    'http.request.method_original': 'POST',
                    'user_agent.original': 'node',
                    'url.query': '',
                    'http.request.url': 'https://bedrock-runtime.eu-central-1.amazonaws.com/guardrail/o8g71agqe4cl/version/1/apply',
                    'server.address': 'bedrock-runtime.eu-central-1.amazonaws.com',
                    'server.port': 443,
                    'http.response.status_code': 200,
                    'network.peer.port': 443,
                },
                created_by_subject=Subject(
                    subject_id='sateesh@truefoundry.com',
                    subject_type=<SubjectType.USER: 'user'>,
                    subject_slug='sateesh@truefoundry.com',
                    subject_display_name=None
                )
            )
            ```
          </Expandable>
        </Accordion>
      </Accordion>

      <Accordion title="Model Span">
        The Model span with ID `de09be32ba8e0c37` has a parent span "ChatCompletion Span (Root Span)" with ID `bddb6503c0eeb940`, making it a sibling to the Guardrail span.
        This span contains all the detailed model metrics and represents the LLM model inference processing within the gateway.
        Notice how the input content has been redacted from `I am sateesh. Hi` to `I am {NAME}. Hi`, demonstrating the PII redaction working.

        <Expandable title="View Model Span Details" defaultOpen="true">
          ```shell theme={"dark"}
          TraceSpan(
              span_id='de09be32ba8e0c37',
              trace_id='019a047ee43577009b6bb5b6ab9477d2',
              parent_span_id='bddb6503c0eeb940',
              service_name='tfy-llm-gateway',
              span_name='Model: openai-main/gpt-5',
              timestamp='2025-10-21T02:00:16.667Z',
              duration_ns=6600714193.0,
              status_code='Ok',
              span_attributes={
                  'tfy.model.fqn': 'truefoundry:openai:openai-main:model:gpt-5',
                  'tfy.model.id': 'cme6zartb0v0701pjeqk9fulg',
                  'tfy.model.name': 'openai-main/gpt-5',
                  'tfy.model.request_url': 'https://api.openai.com/v1/chat/completions',
                  'tfy.model.request_type': 'ChatCompletion',
                  'tfy.error_message': '',
                  'tfy.model.metric.cost_in_usd': 0.00481501,
                  'tfy.model.metric.inter_token_latency_in_ms': 0,
                  'tfy.model.metric.time_to_first_token_in_ms': 7022.75,
                  'tfy.model.metric.input_tokens': 12,
                  'tfy.model.metric.latency_in_ms': 6600.23,
                  'tfy.model.metric.output_tokens': 480,
                  'tfy.model.metric.cache_read_input_tokens': 0,
                  'tfy.model.metric.cache_creation_input_tokens': 0,
                  'tfy.model.streaming': True,
                  'tfy.input': '{"model":"openai-main/gpt-5","messages":[{"role":"user","content":"I am {NAME}. Hi"}],"stream":true}',
                  'tfy.output': '{"id":"chatcmpl-CSvxYlQKenRQWspGHQaCLRcz1MbZU","usage":{"prompt_tokens":12,"completion_tokens":480,"cache_read_tokens":0,"cache_write_tokens":0,"reasoning_tokens":448,"total_tokens":492},"object":"chat.completion","model":"gpt-5-2025-08-07","created":1761012016,"choices":[{"finish_reason":"stop","index":0,"logprobs":null,"message":{"role":"assistant","content":"Hi! Nice to meet you. What would you like me to call you, and how can I help today?","refusal":null}}]}',
                  'tfy.span_type': 'Model'
              },
              created_by_subject=Subject(
                  subject_id='sateesh@truefoundry.com',
                  subject_type=<SubjectType.USER: 'user'>,
                  subject_slug='sateesh@truefoundry.com',
                  subject_display_name=None
              )
          )
          ```
        </Expandable>

        <Accordion title="Model Network Call Span">
          The Model Network Call span with ID `95794dcfbaad832a` has a parent span "Model Span" with ID `de09be32ba8e0c37`, representing the actual HTTP communication with the external provider (OpenAI). This span captures pure network latency and external provider processing time. With a duration of 6.60 seconds, it shows the time spent on the actual API call to the external service.

          <Expandable title="View Model Network Call Span Details" defaultOpen="true">
            ```shell theme={"dark"}
            TraceSpan(
                span_id='95794dcfbaad832a',
                trace_id='019a047ee43577009b6bb5b6ab9477d2',
                parent_span_id='de09be32ba8e0c37',
                service_name='tfy-llm-gateway',
                span_name='POST https://api.openai.com/v1/chat/completions',
                timestamp='2025-10-21T02:00:16.667Z',
                duration_ns=6599577503.0,
                status_code='Unset',
                span_attributes={
                    'url.path': '/v1/chat/completions',
                    'http.request.method': 'POST',
                    'url.scheme': 'https',
                    'network.peer.address': '162.159.140.245',
                    'url.full': 'https://api.openai.com/v1/chat/completions',
                    'http.request.method_original': 'POST',
                    'user_agent.original': 'node',
                    'url.query': '',
                    'http.request.url': 'https://api.openai.com/v1/chat/completions',
                    'server.address': 'api.openai.com',
                    'server.port': 443,
                    'http.response.status_code': 200,
                    'network.peer.port': 443,
                },
                created_by_subject=Subject(
                    subject_id='sateesh@truefoundry.com',
                    subject_type=<SubjectType.USER: 'user'>,
                    subject_slug='sateesh@truefoundry.com',
                    subject_display_name=None
                )
            )
            ```
          </Expandable>
        </Accordion>
      </Accordion>
    </AccordionGroup>
  </Accordion>
</AccordionGroup>

## Filter Request Logs

While fetching all Gateway request logs is useful for general monitoring,
you'll often want to filter logs based on specific criteria such as user identity, model names, etc.
You can achieve this using the `filters` parameter in the `query_spans` method.

The API supports the following common filter types:

* **Span fields filtering**: Filter logs by span fields such as `spanName`, `traceId`, `spanId`, etc. See the [Query Spans documentation](/docs/truefoundry_sdk/traces#query-spans) to understand the supported options for `spanFieldName` and `operator`.
* **Span attributes filtering**: Filter logs by span attributes, e.g., using `tfy.model.name` for model name. See [Attributes](/docs/ai-gateway/fetch-request-logs#understanding-span-attributes) section to understand the supported options for `spanAttributeKey`
* **Gateway request metadata filtering**: Filter logs based on [Custom Metadata](/docs/ai-gateway/log-custom-metadata) keys and values that you passed to Gateway requests.

The following example demonstrates how to retrieve request logs for OpenAI models submitted by a specific user, where custom metadata matches certain values:

<CodeGroup>
  ```python TrueFoundry SDK lines theme={"dark"}
  from truefoundry import client
  from truefoundry_sdk import SpanFieldFilter, SpanAttributeFilter, GatewayRequestMetadataFilter
  from truefoundry_sdk.types.span_field_filter_span_field_name import SpanFieldFilterSpanFieldName
  from truefoundry_sdk.types.span_field_filter_operator import SpanFieldFilterOperator
  from truefoundry_sdk.types.span_attribute_filter_operator import SpanAttributeFilterOperator
  from truefoundry_sdk.types.gateway_request_metadata_filter_operator import GatewayRequestMetadataFilterOperator

  # Fetch LLM Gateway request logs with filters
  spans = client.traces.query_spans(
      data_routing_destination="default",
      start_time="2025-01-21T00:00:00.000Z",
      filters=[
          SpanFieldFilter(span_field_name=SpanFieldFilterSpanFieldName.CREATED_BY_SUBJECT_SLUG, operator=SpanFieldFilterOperator.EQUAL, value="user@example.com"),
          SpanAttributeFilter(span_attribute_key="tfy.model.name", operator=SpanAttributeFilterOperator.STRING_CONTAINS, value="openai"),
          GatewayRequestMetadataFilter(gateway_request_metadata_key="foo", operator=GatewayRequestMetadataFilterOperator.IN, value=["bar1", "bar2"]),
      ]
  )

  for span in spans:
      print(span.span_name, span.span_attributes.get('tfy.span_type'))
  ```

  ```python HTTP API lines theme={"dark"}
  import requests

  page_token = None
  done = False
  while not done:
      # Make API request
      response = requests.post(
          "https://{control_plane_url}/api/svc/v1/spans/query",
          headers={
              "Authorization": "Bearer YOUR_API_TOKEN",
              "Content-Type": "application/json"
          },
          json={
              "dataRoutingDestination": "default",
              "startTime": "2025-10-08T20:16:54",
              "filters": [
                  {"spanFieldName": "createdBySubjectSlug", "operator": "EQUAL", "value": "user@example.com"},
                  {"spanAttributeKey": "tfy.model.name", "operator": "STRING_CONTAINS", "value": "openai"},
                  {"gatewayRequestMetadataKey": "foo", "operator": "IN", "value": ["bar1", "bar2"]}
              ],
              "pageToken": page_token
          }
      )
      
      response.raise_for_status()
      data = response.json()

      for span in data['data']:
          print(span['spanName'], span['spanAttributes'].get('tfy.span_type', ''))
      
      page_token = data['pagination'].get("nextPageToken")
      done = page_token is None

  print("Fetch spans completed!")
  ```
</CodeGroup>

## Common Use Cases

<AccordionGroup>
  <Accordion title="Fetch all spans for a time interval">
    Define the time range for your query. Use ISO 8601 format for timestamps.

    <CodeGroup>
      ```python TrueFoundry SDK lines theme={"dark"}
      from truefoundry import client
      from truefoundry_sdk import SortDirection

      spans = client.traces.query_spans(
          data_routing_destination="default",
          start_time="2025-10-08T00:00:00.000Z",
          end_time="2025-10-08T23:59:59.999Z",
          limit=200,
          sort_direction=SortDirection.DESC
      )

      # Process all spans across all pages
      for span in spans:
          print(span.span_name, span.duration, span.span_attributes.get("tfy.span_type"))
      ```

      ```python HTTP API lines theme={"dark"}
      import requests

      page_token = None
      done = False

      while not done:
          # Make API request
          response = requests.post(
              "https://{control_plane_url}/api/svc/v1/spans/query",
              headers={
                  "Authorization": "Bearer YOUR_API_TOKEN",
                  "Content-Type": "application/json"
              },
              json={
                  "dataRoutingDestination": "default",
                  "startTime": "2025-10-08T00:00:00.000Z",
                  "endTime": "2025-10-08T23:59:59.999Z",
                  "limit": 200,
                  "sortDirection": "desc",
                  "pageToken": page_token
              }
          )
          
          response.raise_for_status()
          data = response.json()

          for span in data['data']:
              print(span['spanName'], span['duration'], span['spanAttributes'].get("tfy.span_type"))
          
          page_token = data['pagination'].get("nextPageToken")
          done = page_token is None

      print("Fetch spans completed!")
      ```
    </CodeGroup>
  </Accordion>

  <Accordion title="Fetch all Root spans for a time interval">
    A root span is the top-level span in a trace hierarchy that has no parent span.

    Define the time range for your query. Use ISO 8601 format for timestamps.

    <CodeGroup>
      ```python TrueFoundry SDK lines theme={"dark"}
      from truefoundry import client
      from truefoundry_sdk import SortDirection

      spans = client.traces.query_spans(
          data_routing_destination="default",
          start_time="2025-10-08T00:00:00.000Z",
          end_time="2025-10-08T23:59:59.999Z",
          parent_span_ids=[""],
          limit=200,
          sort_direction=SortDirection.DESC
      )

      # Process all spans across all pages
      for span in spans:
          print(span.span_name, span.duration, span.span_attributes.get("tfy.span_type"))
      ```

      ```python HTTP API lines theme={"dark"}
      import requests

      page_token = None
      done = False

      while not done:
          # Make API request
          response = requests.post(
              "https://{control_plane_url}/api/svc/v1/spans/query",
              headers={
                  "Authorization": "Bearer YOUR_API_TOKEN",
                  "Content-Type": "application/json"
              },
              json={
                  "dataRoutingDestination": "default",
                  "startTime": "2025-10-08T00:00:00.000Z",
                  "endTime": "2025-10-08T23:59:59.999Z",
                  "parentSpanIds": [""],
                  "limit": 200,
                  "sortDirection": "desc",
                  "pageToken": page_token
              }
          )
          
          response.raise_for_status()
          data = response.json()

          for span in data['data']:
              print(span['spanName'], span['duration'], span['spanAttributes'].get("tfy.span_type"))
          
          page_token = data['pagination'].get("nextPageToken")
          done = page_token is None

      print("Fetch root spans completed!")
      ```
    </CodeGroup>
  </Accordion>

  <Accordion title="Fetch all spans for virtual accounts">
    Define the time range for your query. Use ISO 8601 format for timestamps.

    <CodeGroup>
      ```python TrueFoundry SDK lines theme={"dark"}
      from truefoundry import client
      from truefoundry_sdk import SortDirection

      spans = client.traces.query_spans(
          data_routing_destination="default",
          start_time="2025-10-08T00:00:00.000Z",
          end_time="2025-10-08T23:59:59.999Z",
          created_by_subject_types=["virtualaccount"],
          limit=200,
          sort_direction=SortDirection.DESC
      )

      # Process all spans across all pages
      for span in spans:
          print(span.span_name, span.duration, span.span_attributes.get("tfy.span_type"))
      ```

      ```python HTTP API lines theme={"dark"}
      import requests

      page_token = None
      done = False

      while not done:
          # Make API request
          response = requests.post(
              "https://{control_plane_url}/api/svc/v1/spans/query",
              headers={
                  "Authorization": "Bearer YOUR_API_TOKEN",
                  "Content-Type": "application/json"
              },
              json={
                  "dataRoutingDestination": "default",
                  "startTime": "2025-10-08T00:00:00.000Z",
                  "endTime": "2025-10-08T23:59:59.999Z",
                  "createdBySubjectTypes": ["virtualaccount"],
                  "limit": 200,
                  "sortDirection": "desc",
                  "pageToken": page_token
              }
          )
          
          response.raise_for_status()
          data = response.json()

          for span in data['data']:
              print(span['spanName'], span['duration'], span['spanAttributes'].get("tfy.span_type"))
          
          page_token = data['pagination'].get("nextPageToken")
          done = page_token is None

      print("Fetch virtual account spans completed!")
      ```
    </CodeGroup>
  </Accordion>

  <Accordion title="Fetch all spans for a virtual account `exampleaccount`">
    Define the time range for your query. Use ISO 8601 format for timestamps.

    <CodeGroup>
      ```python TrueFoundry SDK lines theme={"dark"}
      from truefoundry import client
      from truefoundry_sdk import SortDirection

      spans = client.traces.query_spans(
          data_routing_destination="default",
          start_time="2025-10-08T00:00:00.000Z",
          end_time="2025-10-08T23:59:59.999Z",
          created_by_subject_slugs=["exampleaccount"],
          limit=200,
          sort_direction=SortDirection.DESC
      )

      # Process all spans across all pages
      for span in spans:
          print(span.span_name, span.duration, span.span_attributes.get("tfy.span_type"))
      ```

      ```python HTTP API lines theme={"dark"}
      import requests

      page_token = None
      done = False

      while not done:
          # Make API request
          response = requests.post(
              "https://{control_plane_url}/api/svc/v1/spans/query",
              headers={
                  "Authorization": "Bearer YOUR_API_TOKEN",
                  "Content-Type": "application/json"
              },
              json={
                  "dataRoutingDestination": "default",
                  "startTime": "2025-10-08T00:00:00.000Z",
                  "endTime": "2025-10-08T23:59:59.999Z",
                  "createdBySubjectSlugs": ["exampleaccount"],
                  "limit": 200,
                  "sortDirection": "desc",
                  "pageToken": page_token
              }
          )
          
          response.raise_for_status()
          data = response.json()

          for span in data['data']:
              print(span['spanName'], span['duration'], span['spanAttributes'].get("tfy.span_type"))
          
          page_token = data['pagination'].get("nextPageToken")
          done = page_token is None

      print("Fetch spans for virtual account completed!")
      ```
    </CodeGroup>
  </Accordion>

  <Accordion title="Fetch all spans for users">
    Define the time range for your query. Use ISO 8601 format for timestamps.

    <CodeGroup>
      ```python TrueFoundry SDK lines theme={"dark"}
      from truefoundry import client
      from truefoundry_sdk import SortDirection

      spans = client.traces.query_spans(
          data_routing_destination="default",
          start_time="2025-10-08T00:00:00.000Z",
          end_time="2025-10-08T23:59:59.999Z",
          created_by_subject_types=["user"],
          limit=200,
          sort_direction=SortDirection.DESC
      )

      # Process all spans across all pages
      for span in spans:
          print(span.span_name, span.duration, span.span_attributes.get("tfy.span_type"))
      ```

      ```python HTTP API lines theme={"dark"}
      import requests

      page_token = None
      done = False

      while not done:
          # Make API request
          response = requests.post(
              "https://{control_plane_url}/api/svc/v1/spans/query",
              headers={
                  "Authorization": "Bearer YOUR_API_TOKEN",
                  "Content-Type": "application/json"
              },
              json={
                  "dataRoutingDestination": "default",
                  "startTime": "2025-10-08T00:00:00.000Z",
                  "endTime": "2025-10-08T23:59:59.999Z",
                  "createdBySubjectTypes": ["user"],
                  "limit": 200,
                  "sortDirection": "desc",
                  "pageToken": page_token
              }
          )
          
          response.raise_for_status()
          data = response.json()

          for span in data['data']:
              print(span['spanName'], span['duration'], span['spanAttributes'].get("tfy.span_type"))
          
          page_token = data['pagination'].get("nextPageToken")
          done = page_token is None

      print("Fetch user spans completed!")
      ```
    </CodeGroup>
  </Accordion>

  <Accordion title="Fetch all spans for a user with email example@email.com">
    Define the time range for your query. Use ISO 8601 format for timestamps.

    <CodeGroup>
      ```python TrueFoundry SDK lines theme={"dark"}
      from truefoundry import client
      from truefoundry_sdk import SortDirection

      spans = client.traces.query_spans(
          data_routing_destination="default",
          start_time="2025-10-08T00:00:00.000Z",
          end_time="2025-10-08T23:59:59.999Z",
          created_by_subject_slugs=["example@email.com"],
          limit=200,
          sort_direction=SortDirection.DESC
      )

      # Process all spans across all pages
      for span in spans:
          print(span.span_name, span.duration, span.span_attributes.get("tfy.span_type"))
      ```

      ```python HTTP API lines theme={"dark"}
      import requests

      page_token = None
      done = False

      while not done:
          # Make API request
          response = requests.post(
              "https://{control_plane_url}/api/svc/v1/spans/query",
              headers={
                  "Authorization": "Bearer YOUR_API_TOKEN",
                  "Content-Type": "application/json"
              },
              json={
                  "dataRoutingDestination": "default",
                  "startTime": "2025-10-08T00:00:00.000Z",
                  "endTime": "2025-10-08T23:59:59.999Z",
                  "createdBySubjectSlugs": ["example@email.com"],
                  "limit": 200,
                  "sortDirection": "desc",
                  "pageToken": page_token
              }
          )
          
          response.raise_for_status()
          data = response.json()

          for span in data['data']:
              print(span['spanName'], span['duration'], span['spanAttributes'].get("tfy.span_type"))
          
          page_token = data['pagination'].get("nextPageToken")
          done = page_token is None

      print("Fetch spans for user completed!")
      ```
    </CodeGroup>
  </Accordion>

  <Accordion title="Filter by Specific Trace ID">
    Fetch spans of a specific traceId

    <CodeGroup>
      ```python TrueFoundry SDK lines theme={"dark"}
      from truefoundry import client
      from truefoundry_sdk import SortDirection

      spans = client.traces.query_spans(
          data_routing_destination="default",
          start_time="2025-10-08T00:00:00.000Z",
          end_time="2025-10-08T23:59:59.999Z",
          trace_ids=[
              "0199c25e124a70989b0455584fbbf7b7"
          ],
          limit=200,
          sort_direction=SortDirection.DESC
      )

      # Process all spans across all pages
      for span in spans:
          print(span.span_name, span.duration, span.span_attributes.get("tfy.span_type"))
      ```

      ```python HTTP API lines theme={"dark"}
      import requests

      page_token = None
      done = False

      while not done:
          # Make API request
          response = requests.post(
              "https://{control_plane_url}/api/svc/v1/spans/query",
              headers={
                  "Authorization": "Bearer YOUR_API_TOKEN",
                  "Content-Type": "application/json"
              },
              json={
                  "dataRoutingDestination": "default",
                  "startTime": "2025-10-08T00:00:00.000Z",
                  "endTime": "2025-10-08T23:59:59.999Z",
                  "traceIds": ["0199c25e124a70989b0455584fbbf7b7"],
                  "limit": 200,
                  "sortDirection": "desc",
                  "pageToken": page_token
              }
          )
          
          response.raise_for_status()
          data = response.json()

          for span in data['data']:
              print(span['spanName'], span['duration'], span['spanAttributes'].get("tfy.span_type"))
          
          page_token = data['pagination'].get("nextPageToken")
          done = page_token is None

      print("Fetch spans by trace ID completed!")
      ```
    </CodeGroup>
  </Accordion>

  <Accordion title="Fetch all model spans">
    Filter spans by model spanType using the `tfy.span_type` span attribute filter with value `Model`.

    <CodeGroup>
      ```python TrueFoundry SDK lines theme={"dark"}
      from truefoundry import client
      from truefoundry_sdk import SortDirection, SpanAttributeFilter
      from truefoundry_sdk.types.span_attribute_filter_operator import SpanAttributeFilterOperator

      spans = client.traces.query_spans(
          data_routing_destination="default",
          start_time="2025-10-08T00:00:00.000Z",
          end_time="2025-10-08T23:59:59.999Z",
          filters=[
              SpanAttributeFilter(span_attribute_key="tfy.span_type", operator=SpanAttributeFilterOperator.EQUAL, value="Model"),
          ],
          limit=200,
          sort_direction=SortDirection.DESC
      )

      # Process all spans across all pages
      for span in spans:
          print(span.span_name, span.span_attributes.get("tfy.span_type"))
      ```

      ```python HTTP API lines theme={"dark"}
      import requests

      page_token = None
      done = False

      while not done:
          # Make API request
          response = requests.post(
              "https://{control_plane_url}/api/svc/v1/spans/query",
              headers={
                  "Authorization": "Bearer YOUR_API_TOKEN",
                  "Content-Type": "application/json"
              },
              json={
                  "dataRoutingDestination": "default",
                  "startTime": "2025-10-08T00:00:00.000Z",
                  "endTime": "2025-10-08T23:59:59.999Z",
                  "filters": [
                      {"spanAttributeKey": "tfy.span_type", "operator": "EQUAL", "value": "Model"},
                  ],
                  "limit": 200,
                  "sortDirection": "desc",
                  "pageToken": page_token
              }
          )
          
          response.raise_for_status()
          data = response.json()

          for span in data['data']:
              print(span['spanName'], span['spanAttributes'].get("tfy.span_type"))
          
          page_token = data['pagination'].get("nextPageToken")
          done = page_token is None

      print("Fetch spans by model spanType completed!")
      ```
    </CodeGroup>
  </Accordion>

  <Accordion title="Fetch requests with Gateway request metadata">
    Filter spans based on custom metadata keys and values that were passed to Gateway requests using the `X-TFY-METADATA` header.

    <CodeGroup>
      ```python TrueFoundry SDK lines theme={"dark"}
      from truefoundry import client
      from truefoundry_sdk import SortDirection, GatewayRequestMetadataFilter
      from truefoundry_sdk.types.gateway_request_metadata_filter_operator import GatewayRequestMetadataFilterOperator

      spans = client.traces.query_spans(
          data_routing_destination="default",
          start_time="2025-10-08T00:00:00.000Z",
          end_time="2025-10-08T23:59:59.999Z",
          filters=[
              GatewayRequestMetadataFilter(gateway_request_metadata_key="application", operator=GatewayRequestMetadataFilterOperator.EQUAL, value="booking-bot"),
              GatewayRequestMetadataFilter(gateway_request_metadata_key="environment", operator=GatewayRequestMetadataFilterOperator.IN, value=["staging", "production"]),
          ],
          limit=200,
          sort_direction=SortDirection.DESC
      )

      # Process all spans across all pages
      for span in spans:
          print(span.span_name, span.span_attributes.get("tfy.span_type"))
      ```

      ```python HTTP API lines theme={"dark"}
      import requests

      page_token = None
      done = False

      while not done:
          # Make API request
          response = requests.post(
              "https://{control_plane_url}/api/svc/v1/spans/query",
              headers={
                  "Authorization": "Bearer YOUR_API_TOKEN",
                  "Content-Type": "application/json"
              },
              json={
                  "dataRoutingDestination": "default",
                  "startTime": "2025-10-08T00:00:00.000Z",
                  "endTime": "2025-10-08T23:59:59.999Z",
                  "filters": [
                      {"gatewayRequestMetadataKey": "application", "operator": "EQUAL", "value": "booking-bot"},
                      {"gatewayRequestMetadataKey": "environment", "operator": "IN", "value": ["staging", "production"]},
                  ],
                  "limit": 200,
                  "sortDirection": "desc",
                  "pageToken": page_token
              }
          )
          
          response.raise_for_status()
          data = response.json()

          for span in data['data']:
              print(span['spanName'], span['spanAttributes'].get("tfy.span_type"))
          
          page_token = data['pagination'].get("nextPageToken")
          done = page_token is None

      print("Fetch spans with Gateway request metadata completed!")
      ```
    </CodeGroup>
  </Accordion>

  <Accordion title="Fetch spans with MCP in span name">
    Filter spans that have `MCP` in the span name.

    <CodeGroup>
      ```python TrueFoundry SDK lines theme={"dark"}
      from truefoundry import client
      from truefoundry_sdk import SortDirection, SpanFieldFilter
      from truefoundry_sdk.types.span_field_filter_span_field_name import SpanFieldFilterSpanFieldName
      from truefoundry_sdk.types.span_field_filter_operator import SpanFieldFilterOperator

      spans = client.traces.query_spans(
          data_routing_destination="default",
          start_time="2025-10-08T00:00:00.000Z",
          end_time="2025-10-08T23:59:59.999Z",
          filters=[
              SpanFieldFilter(span_field_name=SpanFieldFilterSpanFieldName.SPAN_NAME, operator=SpanFieldFilterOperator.STRING_CONTAINS, value="MCP"),
          ],
          limit=200,
          sort_direction=SortDirection.DESC
      )

      # Process all spans across all pages
      for span in spans:
          print(span.span_name, span.span_attributes.get("tfy.span_type"))
      ```

      ```python HTTP API lines theme={"dark"}
      import requests

      page_token = None
      done = False

      while not done:
          # Make API request
          response = requests.post(
              "https://{control_plane_url}/api/svc/v1/spans/query",
              headers={
                  "Authorization": "Bearer YOUR_API_TOKEN",
                  "Content-Type": "application/json"
              },
              json={
                  "dataRoutingDestination": "default",
                  "startTime": "2025-10-08T00:00:00.000Z",
                  "endTime": "2025-10-08T23:59:59.999Z",
                  "filters": [
                      {"spanFieldName": "spanName", "operator": "STRING_CONTAINS", "value": "MCP"},
                  ],
                  "limit": 200,
                  "sortDirection": "desc",
                  "pageToken": page_token
              }
          )
          
          response.raise_for_status()
          data = response.json()

          for span in data['data']:
              print(span['spanName'], span['spanAttributes'].get("tfy.span_type"))
          
          page_token = data['pagination'].get("nextPageToken")
          done = page_token is None

      print("Fetch MCP spans completed!")
      ```
    </CodeGroup>
  </Accordion>

  <Accordion title="Filter model spans by model name">
    Fetch model spans by filtering with `tfy.span_type` span attribute filter with value `Model`, and then further filter by model name using the `tfy.model.name` span attribute filter.

    <CodeGroup>
      ```python TrueFoundry SDK lines theme={"dark"}
      from truefoundry import client
      from truefoundry_sdk import SortDirection, SpanAttributeFilter
      from truefoundry_sdk.types.span_attribute_filter_operator import SpanAttributeFilterOperator

      spans = client.traces.query_spans(
          data_routing_destination="default",
          start_time="2025-10-08T00:00:00.000Z",
          end_time="2025-10-08T23:59:59.999Z",
          filters=[
              SpanAttributeFilter(span_attribute_key="tfy.model.name", operator=SpanAttributeFilterOperator.EQUAL, value="openai-main/gpt-4"),
              SpanAttributeFilter(span_attribute_key="tfy.span_type", operator=SpanAttributeFilterOperator.EQUAL, value="Model"),
          ],
          limit=200,
          sort_direction=SortDirection.DESC
      )

      # Process all spans across all pages
      for span in spans:
          print(span.span_name, span.span_attributes.get("tfy.model.name"))
      ```

      ```python HTTP API lines theme={"dark"}
      import requests

      page_token = None
      done = False

      while not done:
          # Make API request
          response = requests.post(
              "https://{control_plane_url}/api/svc/v1/spans/query",
              headers={
                  "Authorization": "Bearer YOUR_API_TOKEN",
                  "Content-Type": "application/json"
              },
              json={
                  "dataRoutingDestination": "default",
                  "startTime": "2025-10-08T00:00:00.000Z",
                  "endTime": "2025-10-08T23:59:59.999Z",
                  "filters": [
                      {"spanAttributeKey": "tfy.model.name", "operator": "EQUAL", "value": "openai-main/gpt-4"},
                      {"spanAttributeKey": "tfy.span_type", "operator": "EQUAL", "value": "Model"},
                  ],
                  "limit": 200,
                  "sortDirection": "desc",
                  "pageToken": page_token
              }
          )
          
          response.raise_for_status()
          data = response.json()

          for span in data['data']:
              print(span['spanName'], span['spanAttributes'].get("tfy.model.name"))
          
          page_token = data['pagination'].get("nextPageToken")
          done = page_token is None

      print("Fetch spans by model name completed!")
      ```
    </CodeGroup>
  </Accordion>
</AccordionGroup>

## Understanding Span Attributes

Each span you query from LLM Gateway captures key request and model details. Recognizing these attributes helps you analyze and debug usage effectively.

### Core Span Attributes

| Attribute                      | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `tfy.span_type`                | Type of span, with possible values:<br />• `"ChatCompletion"` - Complete chat request lifecycle<br />• `"Completion"` - Text completion requests without chat context<br />• `"MCP"` - Model Context Protocol server interactions and tool calls<br />• `"Rerank"` - Document reranking operations for search relevance<br />• `"Embedding"` - Vector embedding generation operations<br />• `"Model"` - Actual LLM model inference processing<br />• `"AgentResponse"` - Multi-tool agent orchestration workflows<br />• `"Guardrail"` - Safety, compliance, and content validation checks |
| `tfy.input`                    | Complete input data sent to the model, mcp\_server, guardrail, etc..                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| `tfy.output`                   | Complete output response from the model, mcp\_server, guardrail, etc..                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| `tfy.input_short_hand`         | Abbreviated version of the input for display purposes                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `tfy.error_message`            | Error message if the request failed                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| `tfy.prompt_version_fqn`       | FQN of the prompt version used (if applicable)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| `tfy.prompt_variables`         | Variables used in prompt templating                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| `tfy.triggered_guardrail_fqns` | List of guardrails that were triggered during the request                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |

### Request Context Attributes

| Attribute                              | Description                                                              |
| -------------------------------------- | ------------------------------------------------------------------------ |
| `tfy.request.model_name`               | Name of the model that was requested                                     |
| `tfy.request.created_by_subject`       | Subject (user/service account) that made the request                     |
| `tfy.request.created_by_subject_teams` | Teams associated with the requesting subject                             |
| `tfy.request.metadata`                 | Additional metadata associated with the request (e.g., `{'foo': 'bar'}`) |
| `tfy.request.conversation_id`          | Unique identifier for the conversation (if part of a chat)               |

### Model Attributes

| Attribute                                      | Description                                                                                                                           |
| ---------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| `tfy.model.id`                                 | Unique identifier of the model                                                                                                        |
| `tfy.model.name`                               | Display name of the model                                                                                                             |
| `tfy.model.fqn`                                | Fully qualified name of the model                                                                                                     |
| `tfy.model.request_url`                        | URL endpoint used for the model request                                                                                               |
| `tfy.model.streaming`                          | Whether the request used streaming mode                                                                                               |
| `tfy.model.request_type`                       | Type of request (e.g., `"ChatCompletion", "Completion", "Embedding", "Rerank", "AgentResponse", "MCPGateway", "CreateModelResponse"`) |
| `tfy.model.metric.cache_read_input_tokens`     | Number of input tokens served from the cache, billed at a lower cache read rate instead of the standard input rate                    |
| `tfy.model.metric.cache_creation_input_tokens` | Number of input tokens written to the cache, billed at a higher cache write rate to cover the cost of storage                         |

### Model Performance Metrics

| Attribute                                    | Description                                       |
| -------------------------------------------- | ------------------------------------------------- |
| `tfy.model.metric.time_to_first_token_in_ms` | Time taken to receive the first token (streaming) |
| `tfy.model.metric.latency_in_ms`             | Total request latency in milliseconds             |
| `tfy.model.metric.input_tokens`              | Number of tokens in the model input               |
| `tfy.model.metric.output_tokens`             | Number of tokens in the model output              |
| `tfy.model.metric.cost_in_usd`               | Cost of the request in USD                        |
| `tfy.model.metric.inter_token_latency_in_ms` | Average latency between tokens (streaming)        |

### Load Balancing Attributes

| Attribute                      | Description                                                                |
| ------------------------------ | -------------------------------------------------------------------------- |
| `applied_loadbalance_rule_ids` | IDs of load balancing rules that were applied (e.g., `['gpt-4-dev-load']`) |

### Budget Control Attributes

| Attribute                 | Description                                                                                        |
| ------------------------- | -------------------------------------------------------------------------------------------------- |
| `applied_budget_rule_ids` | IDs of budget rules that were applied to this request (e.g., `['virtualaccount1-monthly-budget']`) |

### Rate Limiting Attributes

| Attribute                    | Description                                                                                    |
| ---------------------------- | ---------------------------------------------------------------------------------------------- |
| `applied_ratelimit_rule_ids` | IDs of all rate limiting rules that were applied (e.g., `['virtualaccount1-daily-ratelimit']`) |

### MCP (Model Context Protocol) Server Attributes

| Attribute                                     | Description                                    |
| --------------------------------------------- | ---------------------------------------------- |
| `tfy.mcp_server.id`                           | Unique identifier of the MCP server            |
| `tfy.mcp_server.name`                         | Display name of the MCP server                 |
| `tfy.mcp_server.url`                          | URL endpoint of the MCP server                 |
| `tfy.mcp_server.fqn`                          | Fully qualified name of the MCP server         |
| `tfy.mcp_server.server_name`                  | Internal name of the MCP server                |
| `tfy.mcp_server.method`                       | MCP method that was called                     |
| `tfy.mcp_server.primitive_name`               | Name of the MCP primitive used                 |
| `tfy.mcp_server.error_code`                   | Error code if the MCP call failed              |
| `tfy.mcp_server.is_tool_call_execution_error` | Whether the error was from tool call execution |

### MCP Server Metrics

| Attribute                               | Description                                    |
| --------------------------------------- | ---------------------------------------------- |
| `tfy.mcp_server.metric.latency_in_ms`   | Latency of the MCP server call in milliseconds |
| `tfy.mcp_server.metric.number_of_tools` | Number of tools available in the MCP server    |

### Guardrail Attributes

| Attribute              | Description                                                          |
| ---------------------- | -------------------------------------------------------------------- |
| `tfy.guardrail.id`     | Unique identifier of the guardrail                                   |
| `tfy.guardrail.name`   | Display name of the guardrail                                        |
| `tfy.guardrail.fqn`    | Fully qualified name of the guardrail                                |
| `tfy.guardrail.result` | Result of the guardrail check (e.g., `'pass'`, `'mutate'`, `'flag'`) |

### Guardrail Applied Entity Attributes

| Attribute                               | Description                                 |
| --------------------------------------- | ------------------------------------------- |
| `tfy.guardrail.applied_on_entity.type`  | Type of entity the guardrail was applied to |
| `tfy.guardrail.applied_on_entity.id`    | ID of the entity                            |
| `tfy.guardrail.applied_on_entity.name`  | Name of the entity                          |
| `tfy.guardrail.applied_on_entity.fqn`   | FQN of the entity                           |
| `tfy.guardrail.applied_on_entity.scope` | Scope of the entity                         |

### Guardrail Metrics

| Attribute                            | Description                                        |
| ------------------------------------ | -------------------------------------------------- |
| `tfy.guardrail.metric.latency_in_ms` | Time taken for the guardrail check in milliseconds |

### HTTP Response Attributes

| Attribute                   | Description                      |
| --------------------------- | -------------------------------- |
| `http.response.status_code` | HTTP status code of the response |

## References

* [Query Spans API](/docs/truefoundry_sdk/traces#query-spans)
* [`query_spans` method in TrueFoundry SDK](/docs/truefoundry_sdk/traces#query-spans)
