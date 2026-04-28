> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Prompt Completions

> Execute your saved prompt templates on Portkey


Portkey Prompts API completely <Tooltip tip="Chat Completions or Completions">follows OpenAI's format</Tooltip>  for both *requests* and *responses*, making it a drop-in replacement existing for your existing **[Chat](/api-reference/inference-api/chat)** or **[Completions](/api-reference/inference-api/completions)** calls.

# Features

<AccordionGroup>
  <Accordion icon="input-pipe" title="Send Variables">
    Create your Propmt Template on [Portkey UI](/product/prompt-library/prompt-templates), define variables, and pass them with this API:

    <Tabs>
      <Tab title="String Variables">
        <CodeGroup>
          ```sh cURL theme={"system"}
          curl -X POST "https://api.portkey.ai/v1/prompts/YOUR_PROMPT_ID/completions" \
            -H "Content-Type: application/json" \
            -H "x-portkey-api-key: $PORTKEY_API_KEY" \
            -d '{
              "variables": {
                "joke_topic": "elections",
                "humor_level": "10"
              }
            }'
          ```

          ```py Python theme={"system"}
          from portkey_ai import Portkey

          client = Portkey(
              api_key="PORTKEY_API_KEY"
          )

          completion = client.prompts.completions.create(
              prompt_id="YOUR_PROMPT_ID",
              variables={
                  "joke_topic": "elections",
                  "humor_level": "10"
              }
          )
          ```

          ```js JavaScript theme={"system"}
          import Portkey from 'portkey-ai';

          const portkey = new Portkey({
            apiKey: 'PORTKEY_API_KEY'
          });

          const completion = await portkey.prompts.completions.create({
            promptId: "YOUR_PROMPT_ID",
            variables: {
              "joke_topic": "elections",
              "humor_level": "10"
            }
          });
          ```
        </CodeGroup>
      </Tab>

      <Tab title="JSON Variables">
        <Note>
          When passing JSON data with variables, `stringify` the value before sending.
        </Note>

        <CodeGroup>
          ```sh cURL theme={"system"}
          curl -X POST "https://api.portkey.ai/v1/prompts/YOUR_PROMPT_ID/completions" \
            -H "Content-Type: application/json" \
            -H "x-portkey-api-key: $PORTKEY_API_KEY" \
            -d '{
              "variables": {
                "user_data": "{\"name\":\"John\",\"preferences\":{\"topic\":\"AI\",\"format\":\"brief\"}}"
              }
            }'
          ```

          ```python Python theme={"system"}
          import json

          user_data = json.dumps({
              "name": "John",
              "preferences": {
                  "topic": "AI",
                  "format": "brief"
              }
          })

          completion = client.prompts.completions.create(
              prompt_id="YOUR_PROMPT_ID",
              variables={
                  "user_data": user_data
              }
          )
          ```

          ```javascript JavaScript theme={"system"}
          const userData = JSON.stringify({
            name: "John",
            preferences: {
              topic: "AI",
              format: "brief"
            }
          });

          const completion = await portkey.prompts.completions.create({
            promptId: "YOUR_PROMPT_ID",
            variables: {
              user_data: userData
            }
          });
          ```
        </CodeGroup>
      </Tab>
    </Tabs>
  </Accordion>

  <Accordion icon="pen-to-square" title="Override Prompt Settings">
    You can override any model hyperparameter saved in the prompt template by sending its new value at the time of making a request:

    <CodeGroup>
      ```sh cURL theme={"system"}
      curl -X POST "https://api.portkey.ai/v1/prompts/YOUR_PROMPT_ID/completions" \
        -H "Content-Type: application/json" \
        -H "x-portkey-api-key: $PORTKEY_API_KEY" \
        -d '{
          "variables": {
            "user_input": "Hello world"
          },
          "temperature": 0.7,
          "max_tokens": 250,
          "presence_penalty": 0.2
        }'
      ```

      ```python Python theme={"system"}
      completion = client.prompts.completions.create(
          prompt_id="YOUR_PROMPT_ID",
          variables={
              "user_input": "Hello world"
          },
          temperature=0.7,
          max_tokens=250,
          presence_penalty=0.2
      )
      ```

      ```javascript JavaScript theme={"system"}
      const completion = await portkey.prompts.completions.create({
        promptId: "YOUR_PROMPT_ID",
        variables: {
          user_input: "Hello world"
        },
        temperature: 0.7,
        max_tokens: 250,
        presence_penalty: 0.2
      });
      ```
    </CodeGroup>
  </Accordion>

  <Accordion icon="code-compare" title="Call Specific Prompt Version">
    Passing the `{promptId}` always calls the `Published` version of your prompt.

    But, you can also call a specific template version by appending its version number, like `{promptId@12}`:

    **Version Tags**:

    * `@latest`: Calls the <Tooltip tip="May not be the same as the Published version">most recent version</Tooltip>
    * `@{NUMBER}` (like `@12`): Calls the specified version number
    * `No Suffix`: Here, Portkey defaults to the `Published` version

    <CodeGroup>
      ```curl cURL {1} theme={"system"}
      curl -X POST "https://api.portkey.ai/v1/prompts/PROMPT_ID@12/completions" \
        -H "Content-Type: application/json" \
        -H "x-portkey-api-key: $PORTKEY_API_KEY" \
        -d '{
          "variables": {
            "user_input": "Hello world"
          }
        }'
      ```

      ```python Python {2} theme={"system"}
      completion = client.prompts.completions.create(
          prompt_id="PROMPT_ID@12", # PROMPT_ID@latest will call the latest version
          variables={
              "user_input": "Hello world"
          }
      )
      ```

      ```javascript JavaScript {2} theme={"system"}
      const completion = await portkey.prompts.completions.create({
        promptId: "PROMPT_ID@12", // PROMPT_ID@latest will call the latest version
        variables: {
          user_input: "Hello world"
        }
      });
      ```
    </CodeGroup>
  </Accordion>

  <Accordion icon="water" title="Streaming">
    Prompts API also supports streaming responses, and completely follows the OpenAI schema.

    * Set `stream:True` explicitly in your request to enable streaming

    <CodeGroup>
      ```sh cURL {8} theme={"system"}
      curl -X POST "https://api.portkey.ai/v1/prompts/YOUR_PROMPT_ID/completions" \
        -H "Content-Type: application/json" \
        -H "x-portkey-api-key: $PORTKEY_API_KEY" \
        -d '{
          "variables": {
            "user_input": "Hello world"
          },
          "stream": true
          "max_tokens": 250,
          "presence_penalty": 0.2
        }'
      ```

      ```python Python {4} theme={"system"}
      completion = client.prompts.completions.create(
          prompt_id="YOUR_PROMPT_ID",
          variables={"user_input": "Hello"},
          stream=True
      )

      for chunk in completion:
          print(chunk.choices[0].delta)
      ```

      ```javascript JavaScript {6} theme={"system"}
      const completion = await portkey.prompts.completions.create({
        promptId: "YOUR_PROMPT_ID",
        variables: {
          user_input: "Hello"
        },
        stream: true
      });

      for await (const chunk of completion) {
        console.log(chunk.choices[0].delta.content);
      }
      ```
    </CodeGroup>
  </Accordion>
</AccordionGroup>


## OpenAPI

````yaml post /prompts/{promptId}/completions
openapi: 3.0.0
info:
  title: Portkey API
  description: >-
    The Portkey REST API. Please see https://portkey.ai/docs/api-reference for
    more details.
  version: 2.0.0
  termsOfService: https://portkey.ai/terms
  contact:
    name: Portkey Developer Forum
    url: https://portkey.wiki/community
  license:
    name: MIT
    url: https://github.com/Portkey-AI/portkey-openapi/blob/master/LICENSE
servers:
  - url: https://api.portkey.ai/v1
    description: Portkey API Public Endpoint
security:
  - Portkey-Key: []
tags:
  - name: Assistants
    description: Build Assistants that can call models and use tools.
  - name: Audio
    description: Turn audio into text or text into audio.
  - name: Chat
    description: >-
      Given a list of messages comprising a conversation, the model will return
      a response.
  - name: Realtime
    description: WebSocket proxy for provider Realtime APIs
  - name: Collections
    description: Create, List, Retrieve, Update, and Delete collections of prompts.
  - name: Labels
    description: Create, List, Retrieve, Update, and Delete labels.
  - name: Prompt Collections
    description: Create, List, Retrieve, Update, and Delete prompt collections.
  - name: PromptPartials
    description: Create, List, Retrieve, Update, and Delete prompt partials.
  - name: Prompts
    description: >-
      Given a prompt template ID and variables, will run the saved prompt
      template and return a response.
  - name: Guardrails
    description: Create, List, Retrieve, Update, and Delete prompt Guardrails.
  - name: Completions
    description: >-
      Given a prompt, the model will return one or more predicted completions,
      and can also return the probabilities of alternative tokens at each
      position.
  - name: Embeddings
    description: >-
      Get a vector representation of a given input that can be easily consumed
      by machine learning models and algorithms.
  - name: Rerank
    description: >-
      Rerank a list of documents based on their relevance to a query. Supported
      providers include Cohere, Voyage, Jina, Pinecone, Bedrock, and Azure AI.
  - name: Fine-tuning
    description: Manage fine-tuning jobs to tailor a model to your specific training data.
  - name: Batch
    description: Create large batches of API requests to run asynchronously.
  - name: Files
    description: >-
      Files are used to upload documents that can be used with features like
      Assistants and Fine-tuning.
  - name: Images
    description: Given a prompt and/or an input image, the model will generate a new image.
  - name: Models
    description: List and describe the various models available in the API.
  - name: Moderations
    description: >-
      Given a input text, outputs if the model classifies it as potentially
      harmful.
  - name: Configs
    description: Create, List, Retrieve, and Update your Portkey Configs.
  - name: Feedback
    description: Send and Update any feedback.
  - name: Logs
    description: Custom Logger to add external logs to Portkey.
  - name: Integrations
    description: Create, List, Retrieve, Update, and Delete your Portkey Integrations.
  - name: Integrations > Workspaces
    description: Manage workspace access for your Portkey Integrations.
  - name: Integrations > Models
    description: Manage model access for your Portkey Integrations.
  - name: Providers
    description: Create, List, Retrieve, Update, and Delete your Portkey Providers.
  - name: Virtual-keys
    description: Create, List, Retrieve, Update, and Delete your Portkey Virtual keys.
  - name: Users
    description: Create and manage users.
  - name: User-invites
    description: Create and manage user invites.
  - name: Workspaces
    description: Create and manage workspaces.
  - name: Workspaces > Members
    description: Create and manage workspace members.
  - name: MCP Integrations
    description: Create, List, Retrieve, Update, and Delete MCP Integrations.
  - name: MCP Integrations > Workspaces
    description: Manage workspace access for MCP Integrations.
  - name: MCP Integrations > Capabilities
    description: List and manage capabilities for MCP Integrations.
  - name: MCP Integrations > Metadata
    description: Get MCP Integration metadata and sync info.
  - name: MCP Servers
    description: >-
      Create, List, Retrieve, Update, and Delete MCP Servers (workspace
      instances of MCP Integrations).
  - name: MCP Servers > Capabilities
    description: List and manage capabilities for MCP Servers.
  - name: MCP Servers > User Access
    description: List and manage user access for MCP Servers.
  - name: MCP Servers > Connections
    description: List and manage user connections for MCP Servers.
  - name: Api-Keys
    description: Create, List, Retrieve, Update, and Delete your Portkey API keys.
  - name: Logs Export
    description: Exports logs service.
  - name: Audit Logs
    description: Get audit logs for your Portkey account.
  - name: Analytics
    description: >-
      Get analytics over different data points like requests, costs, tokens,
      etc.
  - name: Analytics > Graphs
    description: Get data points for graphical representation.
  - name: Analytics > Summary
    description: Get overall summary for the selected time bucket.
  - name: Analytics > Groups
    description: Get grouped metrics for the selected time bucket.
  - name: Usage Limits Policies
    description: Manage usage limits policies to control total usage over time
  - name: Rate Limits Policies
    description: Manage rate limits policies to control request or token rates
  - name: Model Pricing
    description: Model pricing configurations for 2300+ LLMs across 40+ providers
  - name: Secret-References
    description: >-
      Create, List, Retrieve, Update, and Delete secret references to external
      secret managers.
paths:
  /prompts/{promptId}/completions:
    servers:
      - url: https://api.portkey.ai/v1
        description: Portkey API Public Endpoint
      - url: SELF_HOSTED_GATEWAY_URL
        description: Self-Hosted Gateway URL
    post:
      tags:
        - Prompts
      summary: Prompts Completions
      description: |
        Execute your saved prompt templates on Portkey
      operationId: createPromptCompletion
      parameters:
        - in: path
          name: promptId
          required: true
          schema:
            type: string
          description: The unique identifier of the prompt template to use
      requestBody:
        required: true
        content:
          application/json:
            schema:
              allOf:
                - type: object
                  required:
                    - variables
                  description: >
                    Note: Although hyperparameters are shown grouped here (like
                    messages, max_completion_tokens, temperature, etc.), they
                    should only be passed at the root level, alongside
                    'variables' and 'stream'. The `max_tokens` parameter is
                    deprecated — use `max_completion_tokens` instead.
                  properties:
                    variables:
                      type: object
                      description: Variables to substitute in the prompt template
                    stream:
                      type: boolean
                      default: false
                      description: >-
                        Default: False. Set to True if you want to stream the
                        response
                    hyperparameters:
                      oneOf:
                        - $ref: '#/components/schemas/CreateChatCompletionRequest'
                          title: Chat Completions
                        - $ref: '#/components/schemas/CreateCompletionRequest'
                          title: Completions
                      description: >
                        **Note**: All hyperparameters are optional. Pass them at
                        the root level, and not nested under `hyperparameters`.
                        Their grouping here is for educational purposes only.
      responses:
        '200':
          description: Successful completion response
          content:
            application/json:
              schema:
                type: object
                properties:
                  status:
                    type: string
                    description: Response status
                  headers:
                    type: object
                    description: Response headers
                  body:
                    oneOf:
                      - $ref: '#/components/schemas/CreateChatCompletionResponse'
                        title: Chat Completions
                      - $ref: '#/components/schemas/CreateCompletionResponse'
                        title: Completions
      x-code-samples:
        - lang: cURL
          label: Default
          source: >
            curl -X POST
            "https://api.portkey.ai/v1/prompts/YOUR_PROMPT_ID/completions" \
              -H "Content-Type: application/json" \
              -H "x-portkey-api-key: $PORTKEY_API_KEY" \
              -d '{
                "variables": {
                  "user_input": "Hello world"
                },
                "max_completion_tokens": 250,
                "presence_penalty": 0.2
              }'
        - lang: Python
          label: Default
          source: |
            from portkey_ai import Portkey

            client = Portkey(
                api_key="PORTKEY_API_KEY"
            )

            completion = client.prompts.completions.create(
                prompt_id="YOUR_PROMPT_ID",
                variables={
                    "user_input": "Hello world"
                },
                max_completion_tokens=250,
                presence_penalty=0.2
            )

            print(completion)
        - lang: JavaScript
          label: Default
          source: |
            import Portkey from 'portkey-ai';

            const portkey = new Portkey({
              apiKey: 'PORTKEY_API_KEY'
            });

            const completion = await portkey.prompts.completions.create({
              promptId: "YOUR_PROMPT_ID",
              variables: {
                user_input: "Hello world"
              },
              max_completion_tokens: 250,
              presence_penalty: 0.2
            });

            console.log(completion);
        - lang: curl
          label: Self-Hosted
          source: >
            curl -X POST
            "SELF_HOSTED_GATEWAY_URL/prompts/YOUR_PROMPT_ID/completions" \
              -H "Content-Type: application/json" \
              -H "x-portkey-api-key: $PORTKEY_API_KEY" \
              -d '{
                "variables": {
                  "user_input": "Hello world"
                },
                "max_completion_tokens": 250,
                "presence_penalty": 0.2
              }'
        - lang: python
          label: Self-Hosted
          source: |
            from portkey_ai import Portkey

            client = Portkey(
                api_key="PORTKEY_API_KEY",
                base_url="SELF_HOSTED_GATEWAY_URL"
            )

            completion = client.prompts.completions.create(
                prompt_id="YOUR_PROMPT_ID",
                variables={
                    "user_input": "Hello world"
                },
                max_completion_tokens=250,
                presence_penalty=0.2
            )

            print(completion)
        - lang: javascript
          label: Self-Hosted
          source: |
            import Portkey from 'portkey-ai';

            const portkey = new Portkey({
              apiKey: 'PORTKEY_API_KEY',
              baseURL: 'SELF_HOSTED_GATEWAY_URL'
            });

            const completion = await portkey.prompts.completions.create({
              promptId: "YOUR_PROMPT_ID",
              variables: {
                user_input: "Hello world"
              },
              max_completion_tokens: 250,
              presence_penalty: 0.2
            });

            console.log(completion);
components:
  schemas:
    CreateChatCompletionRequest:
      type: object
      properties:
        messages:
          description: >-
            A list of messages comprising the conversation so far. [Example
            Python
            code](https://cookbook.openai.com/examples/how_to_format_inputs_to_chatgpt_models).
          type: array
          minItems: 1
          items:
            $ref: '#/components/schemas/ChatCompletionRequestMessage'
        model:
          description: >-
            ID of the model to use. See the [model endpoint
            compatibility](https://platform.openai.com/docs/models/model-endpoint-compatibility)
            table for details on which models work with the Chat API.
          example: gpt-5
          anyOf:
            - type: string
            - type: string
              enum:
                - gpt-5
                - gpt-5-mini
                - gpt-5-nano
                - o4-mini
                - o3
                - o3-mini
                - o1
                - o1-mini
                - gpt-4o
                - gpt-4o-mini
                - gpt-4o-2024-05-13
                - gpt-4-turbo
                - gpt-4-turbo-2024-04-09
                - gpt-4-0125-preview
                - gpt-4-turbo-preview
                - gpt-4-1106-preview
                - gpt-4-vision-preview
                - gpt-4
                - gpt-4-0314
                - gpt-4-0613
                - gpt-4-32k
                - gpt-4-32k-0314
                - gpt-4-32k-0613
                - gpt-3.5-turbo
                - gpt-3.5-turbo-16k
                - gpt-3.5-turbo-0301
                - gpt-3.5-turbo-0613
                - gpt-3.5-turbo-1106
                - gpt-3.5-turbo-0125
                - gpt-3.5-turbo-16k-0613
          x-oaiTypeLabel: string
        frequency_penalty:
          type: number
          default: 0
          minimum: -2
          maximum: 2
          nullable: true
          description: >
            Number between -2.0 and 2.0. Positive values penalize new tokens
            based on their existing frequency in the text so far, decreasing the
            model's likelihood to repeat the same line verbatim.


            [See more information about frequency and presence
            penalties.](https://platform.openai.com/docs/guides/text-generation/parameter-details)
        logit_bias:
          type: object
          x-oaiTypeLabel: map
          default: null
          nullable: true
          additionalProperties:
            type: integer
          description: >
            Modify the likelihood of specified tokens appearing in the
            completion.


            Accepts a JSON object that maps tokens (specified by their token ID
            in the tokenizer) to an associated bias value from -100 to 100.
            Mathematically, the bias is added to the logits generated by the
            model prior to sampling. The exact effect will vary per model, but
            values between -1 and 1 should decrease or increase likelihood of
            selection; values like -100 or 100 should result in a ban or
            exclusive selection of the relevant token.
        logprobs:
          description: >-
            Whether to return log probabilities of the output tokens or not. If
            true, returns the log probabilities of each output token returned in
            the `content` of `message`.
          type: boolean
          default: false
          nullable: true
        top_logprobs:
          description: >-
            An integer between 0 and 20 specifying the number of most likely
            tokens to return at each token position, each with an associated log
            probability. `logprobs` must be set to `true` if this parameter is
            used.
          type: integer
          minimum: 0
          maximum: 20
          nullable: true
        max_tokens:
          description: >
            Deprecated in favor of `max_completion_tokens`.


            The maximum number of
            [tokens](https://platform.openai.com/tokenizer?view=bpe) that can be
            generated in the chat completion. This value can be used to control
            [costs](https://openai.com/api/pricing/) for text generated via API.


            **Supported models:** GPT-4o, GPT-4o-mini, GPT-4, GPT-4 Turbo,
            GPT-3.5 Turbo.


            **Not supported:** o-series reasoning models (o1, o3, o3-mini,
            o4-mini) and GPT-5 series — use `max_completion_tokens` instead.
          type: integer
          nullable: true
          deprecated: true
        max_completion_tokens:
          description: >
            An upper bound for the number of tokens that can be generated for a
            completion, including visible output tokens and [reasoning
            tokens](https://platform.openai.com/docs/guides/reasoning).


            **Supported models:** GPT-5 series, o-series reasoning models (o1,
            o3, o3-mini, o4-mini) — required. Also supported on GPT-4o,
            GPT-4o-mini, GPT-4, GPT-4 Turbo, GPT-3.5 Turbo as a replacement for
            `max_tokens`.
          type: integer
          nullable: true
        'n':
          type: integer
          minimum: 1
          maximum: 128
          default: 1
          example: 1
          nullable: true
          description: >-
            How many chat completion choices to generate for each input message.
            Note that you will be charged based on the number of generated
            tokens across all of the choices. Keep `n` as `1` to minimize costs.
        presence_penalty:
          type: number
          default: 0
          minimum: -2
          maximum: 2
          nullable: true
          description: >
            Number between -2.0 and 2.0. Positive values penalize new tokens
            based on whether they appear in the text so far, increasing the
            model's likelihood to talk about new topics.


            [See more information about frequency and presence
            penalties.](https://platform.openai.com/docs/guides/text-generation/parameter-details)
        response_format:
          type: object
          description: >
            An object specifying the format that the model must output.


            Setting to `{ "type": "json_schema", "json_schema": {...} }`enables
            Structured Outputs which ensures the model will match your

            supplied JSON schema. Works across all the providers that support
            this functionality. [OpenAI & Azure
            OpenAI](/integrations/llms/openai/structured-outputs), [Gemini &
            Vertex AI](/integrations/llms/vertex-ai/controlled-generations).


            Setting to `{ "type": "json_object" }` enables the older JSON mode,
            which ensures the message the model generates is valid JSON.


            Using `json_schema` is preferred for models that support it.
          oneOf:
            - $ref: '#/components/schemas/ResponseFormatText'
            - $ref: '#/components/schemas/ResponseFormatJsonSchema'
            - $ref: '#/components/schemas/ResponseFormatJsonObject'
        seed:
          type: integer
          minimum: -9223372036854776000
          maximum: 9223372036854776000
          nullable: true
          description: >
            This feature is in Beta.

            If specified, our system will make a best effort to sample
            deterministically, such that repeated requests with the same `seed`
            and parameters should return the same result.

            Determinism is not guaranteed, and you should refer to the
            `system_fingerprint` response parameter to monitor changes in the
            backend.
          x-code-samples:
            beta: true
        stop:
          description: |
            Up to 4 sequences where the API will stop generating further tokens.
          default: null
          oneOf:
            - type: string
              nullable: true
            - type: array
              minItems: 1
              maxItems: 4
              items:
                type: string
        stream:
          description: >
            If set, partial message deltas will be sent, like in ChatGPT. Tokens
            will be sent as data-only [server-sent
            events](https://developer.mozilla.org/en-UShttps://platform.openai.com/docs/Web/API/Server-sent_events/Using_server-sent_events#Event_stream_format)
            as they become available, with the stream terminated by a `data:
            [DONE]` message. [Example Python
            code](https://cookbook.openai.com/examples/how_to_stream_completions).
          type: boolean
          nullable: true
          default: false
        stream_options:
          $ref: '#/components/schemas/ChatCompletionStreamOptions'
        thinking:
          type: object
          nullable: true
          description: >
            View the thinking/reasoning tokens as part of your response.
            Thinking models produce a long internal chain of thought before
            generating a response. Supported only for specific Claude models on
            Anthropic, Google Vertex AI, and AWS Bedrock.  Requires setting
            `strict_openai_compliance = false` in your API call.
          properties:
            type:
              type: string
              enum:
                - enabled
                - disabled
              description: Enables or disables the thinking mode capability.
              default: disabled
            budget_tokens:
              type: integer
              description: >
                The maximum number of tokens to allocate for the thinking
                process.

                A higher token budget allows for more thorough reasoning but may
                increase overall response time.
              minimum: 1
              example: 2030
          required:
            - type
          example:
            type: enabled
            budget_tokens: 2030
        temperature:
          type: number
          minimum: 0
          maximum: 2
          default: 1
          example: 1
          nullable: true
          description: >
            What sampling temperature to use, between 0 and 2. Higher values
            like 0.8 will make the output more random, while lower values like
            0.2 will make it more focused and deterministic.


            We generally recommend altering this or `top_p` but not both.
        top_p:
          type: number
          minimum: 0
          maximum: 1
          default: 1
          example: 1
          nullable: true
          description: >
            An alternative to sampling with temperature, called nucleus
            sampling, where the model considers the results of the tokens with
            top_p probability mass. So 0.1 means only the tokens comprising the
            top 10% probability mass are considered.


            We generally recommend altering this or `temperature` but not both.
        tools:
          type: array
          description: >
            A list of tools the model may call. Currently, only functions are
            supported as a tool. Use this to provide a list of functions the
            model may generate JSON inputs for. A max of 128 functions are
            supported.
          items:
            $ref: '#/components/schemas/ChatCompletionTool'
        tool_choice:
          $ref: '#/components/schemas/ChatCompletionToolChoiceOption'
        parallel_tool_calls:
          $ref: '#/components/schemas/ParallelToolCalls'
        user:
          type: string
          example: user-1234
          description: >
            A unique identifier representing your end-user, which can help
            OpenAI to monitor and detect abuse. [Learn
            more](https://platform.openai.com/docs/guides/safety-best-practices/end-user-ids).
        function_call:
          deprecated: true
          description: >
            Deprecated in favor of `tool_choice`.


            Controls which (if any) function is called by the model.

            `none` means the model will not call a function and instead
            generates a message.

            `auto` means the model can pick between generating a message or
            calling a function.

            Specifying a particular function via `{"name": "my_function"}`
            forces the model to call that function.


            `none` is the default when no functions are present. `auto` is the
            default if functions are present.
          oneOf:
            - type: string
              description: >
                `none` means the model will not call a function and instead
                generates a message. `auto` means the model can pick between
                generating a message or calling a function.
              enum:
                - none
                - auto
            - $ref: '#/components/schemas/ChatCompletionFunctionCallOption'
          x-oaiExpandable: true
        functions:
          deprecated: true
          description: |
            Deprecated in favor of `tools`.

            A list of functions the model may generate JSON inputs for.
          type: array
          minItems: 1
          maxItems: 128
          items:
            $ref: '#/components/schemas/ChatCompletionFunctions'
      required:
        - model
        - messages
    CreateCompletionRequest:
      type: object
      properties:
        model:
          description: >
            ID of the model to use. You can use the [List
            models](https://platform.openai.com/docs/api-reference/models/list)
            API to see all of your available models, or see our [Model
            overview](https://platform.openai.com/docs/models/overview) for
            descriptions of them.
          anyOf:
            - type: string
            - type: string
              enum:
                - gpt-3.5-turbo-instruct
                - davinci-002
                - babbage-002
          x-oaiTypeLabel: string
        prompt:
          description: >
            The prompt(s) to generate completions for, encoded as a string,
            array of strings, array of tokens, or array of token arrays.


            Note that <|endoftext|> is the document separator that the model
            sees during training, so if a prompt is not specified the model will
            generate as if from the beginning of a new document.
          default: <|endoftext|>
          nullable: true
          oneOf:
            - type: string
              default: ''
              example: This is a test.
            - type: array
              items:
                type: string
                default: ''
                example: This is a test.
            - type: array
              minItems: 1
              items:
                type: integer
              example: '[1212, 318, 257, 1332, 13]'
            - type: array
              minItems: 1
              items:
                type: array
                minItems: 1
                items:
                  type: integer
              example: '[[1212, 318, 257, 1332, 13]]'
        best_of:
          type: integer
          default: 1
          minimum: 0
          maximum: 20
          nullable: true
          description: >
            Generates `best_of` completions server-side and returns the "best"
            (the one with the highest log probability per token). Results cannot
            be streamed.


            When used with `n`, `best_of` controls the number of candidate
            completions and `n` specifies how many to return – `best_of` must be
            greater than `n`.


            **Note:** Because this parameter generates many completions, it can
            quickly consume your token quota. Use carefully and ensure that you
            have reasonable settings for `max_tokens` and `stop`.
        echo:
          type: boolean
          default: false
          nullable: true
          description: |
            Echo back the prompt in addition to the completion
        frequency_penalty:
          type: number
          default: 0
          minimum: -2
          maximum: 2
          nullable: true
          description: >
            Number between -2.0 and 2.0. Positive values penalize new tokens
            based on their existing frequency in the text so far, decreasing the
            model's likelihood to repeat the same line verbatim.


            [See more information about frequency and presence
            penalties.](https://platform.openai.com/docs/guides/text-generation/parameter-details)
        logit_bias:
          type: object
          x-oaiTypeLabel: map
          default: null
          nullable: true
          additionalProperties:
            type: integer
          description: >
            Modify the likelihood of specified tokens appearing in the
            completion.


            Accepts a JSON object that maps tokens (specified by their token ID
            in the GPT tokenizer) to an associated bias value from -100 to 100.
            You can use this [tokenizer
            tool](https://platform.openai.com/tokenizer?view=bpe) to convert
            text to token IDs. Mathematically, the bias is added to the logits
            generated by the model prior to sampling. The exact effect will vary
            per model, but values between -1 and 1 should decrease or increase
            likelihood of selection; values like -100 or 100 should result in a
            ban or exclusive selection of the relevant token.


            As an example, you can pass `{"50256": -100}` to prevent the
            <|endoftext|> token from being generated.
        logprobs:
          type: integer
          minimum: 0
          maximum: 5
          default: null
          nullable: true
          description: >
            Include the log probabilities on the `logprobs` most likely output
            tokens, as well the chosen tokens. For example, if `logprobs` is 5,
            the API will return a list of the 5 most likely tokens. The API will
            always return the `logprob` of the sampled token, so there may be up
            to `logprobs+1` elements in the response.


            The maximum value for `logprobs` is 5.
        max_tokens:
          type: integer
          minimum: 0
          default: 16
          example: 16
          nullable: true
          description: >
            The maximum number of
            [tokens](https://platform.openai.com/tokenizer?view=bpe) that can be
            generated in the completion.


            The token count of your prompt plus `max_tokens` cannot exceed the
            model's context length. [Example Python
            code](https://cookbook.openai.com/examples/how_to_count_tokens_with_tiktoken)
            for counting tokens.
        'n':
          type: integer
          minimum: 1
          maximum: 128
          default: 1
          example: 1
          nullable: true
          description: >
            How many completions to generate for each prompt.


            **Note:** Because this parameter generates many completions, it can
            quickly consume your token quota. Use carefully and ensure that you
            have reasonable settings for `max_tokens` and `stop`.
        presence_penalty:
          type: number
          default: 0
          minimum: -2
          maximum: 2
          nullable: true
          description: >
            Number between -2.0 and 2.0. Positive values penalize new tokens
            based on whether they appear in the text so far, increasing the
            model's likelihood to talk about new topics.


            [See more information about frequency and presence
            penalties.](https://platform.openai.com/docs/guides/text-generation/parameter-details)
        seed:
          type: integer
          minimum: -9223372036854776000
          maximum: 9223372036854776000
          nullable: true
          description: >
            If specified, our system will make a best effort to sample
            deterministically, such that repeated requests with the same `seed`
            and parameters should return the same result.


            Determinism is not guaranteed, and you should refer to the
            `system_fingerprint` response parameter to monitor changes in the
            backend.
        stop:
          description: >
            Up to 4 sequences where the API will stop generating further tokens.
            The returned text will not contain the stop sequence.
          default: null
          nullable: true
          oneOf:
            - type: string
              default: <|endoftext|>
              example: |+

              nullable: true
            - type: array
              minItems: 1
              maxItems: 4
              items:
                type: string
                example: '["\n"]'
        stream:
          description: >
            Whether to stream back partial progress. If set, tokens will be sent
            as data-only [server-sent
            events](https://developer.mozilla.org/en-UShttps://platform.openai.com/docs/Web/API/Server-sent_events/Using_server-sent_events#Event_stream_format)
            as they become available, with the stream terminated by a `data:
            [DONE]` message. [Example Python
            code](https://cookbook.openai.com/examples/how_to_stream_completions).
          type: boolean
          nullable: true
          default: false
        stream_options:
          $ref: '#/components/schemas/ChatCompletionStreamOptions'
        suffix:
          description: |
            The suffix that comes after a completion of inserted text.

            This parameter is only supported for `gpt-3.5-turbo-instruct`.
          default: null
          nullable: true
          type: string
          example: test.
        temperature:
          type: number
          minimum: 0
          maximum: 2
          default: 1
          example: 1
          nullable: true
          description: >
            What sampling temperature to use, between 0 and 2. Higher values
            like 0.8 will make the output more random, while lower values like
            0.2 will make it more focused and deterministic.


            We generally recommend altering this or `top_p` but not both.
        top_p:
          type: number
          minimum: 0
          maximum: 1
          default: 1
          example: 1
          nullable: true
          description: >
            An alternative to sampling with temperature, called nucleus
            sampling, where the model considers the results of the tokens with
            top_p probability mass. So 0.1 means only the tokens comprising the
            top 10% probability mass are considered.


            We generally recommend altering this or `temperature` but not both.
        user:
          type: string
          example: user-1234
          description: >
            A unique identifier representing your end-user, which can help
            OpenAI to monitor and detect abuse. [Learn
            more](https://platform.openai.com/docs/guides/safety-best-practices/end-user-ids).
      required:
        - model
        - prompt
    CreateChatCompletionResponse:
      type: object
      description: >-
        Represents a chat completion response returned by model, based on the
        provided input.
      properties:
        id:
          type: string
          description: A unique identifier for the chat completion.
        choices:
          type: array
          description: >-
            A list of chat completion choices. Can be more than one if `n` is
            greater than 1.
          items:
            type: object
            required:
              - finish_reason
              - index
              - message
              - logprobs
            properties:
              finish_reason:
                type: string
                description: >
                  The reason the model stopped generating tokens. This will be
                  `stop` if the model hit a natural stop point or a provided
                  stop sequence,

                  `length` if the maximum number of tokens specified in the
                  request was reached,

                  `content_filter` if content was omitted due to a flag from our
                  content filters,

                  `tool_calls` if the model called a tool, or `function_call`
                  (deprecated) if the model called a function.
                enum:
                  - stop
                  - length
                  - tool_calls
                  - content_filter
                  - function_call
              index:
                type: integer
                description: The index of the choice in the list of choices.
              message:
                $ref: '#/components/schemas/ChatCompletionResponseMessage'
              logprobs:
                description: Log probability information for the choice.
                type: object
                nullable: true
                properties:
                  content:
                    description: >-
                      A list of message content tokens with log probability
                      information.
                    type: array
                    items:
                      $ref: '#/components/schemas/ChatCompletionTokenLogprob'
                    nullable: true
                required:
                  - content
        created:
          type: integer
          description: >-
            The Unix timestamp (in seconds) of when the chat completion was
            created.
        model:
          type: string
          description: The model used for the chat completion.
        system_fingerprint:
          type: string
          description: >
            This fingerprint represents the backend configuration that the model
            runs with.


            Can be used in conjunction with the `seed` request parameter to
            understand when backend changes have been made that might impact
            determinism.
        object:
          type: string
          description: The object type, which is always `chat.completion`.
          enum:
            - chat.completion
        usage:
          $ref: '#/components/schemas/CompletionUsage'
      required:
        - choices
        - created
        - id
        - model
        - object
    CreateCompletionResponse:
      type: object
      description: >
        Represents a completion response from the API. Note: both the streamed
        and non-streamed response objects share the same shape (unlike the chat
        endpoint).
      properties:
        id:
          type: string
          description: A unique identifier for the completion.
        choices:
          type: array
          description: >-
            The list of completion choices the model generated for the input
            prompt.
          items:
            type: object
            required:
              - finish_reason
              - index
              - logprobs
              - text
            properties:
              finish_reason:
                type: string
                description: >
                  The reason the model stopped generating tokens. This will be
                  `stop` if the model hit a natural stop point or a provided
                  stop sequence,

                  `length` if the maximum number of tokens specified in the
                  request was reached,

                  or `content_filter` if content was omitted due to a flag from
                  our content filters.
                enum:
                  - stop
                  - length
                  - content_filter
              index:
                type: integer
              logprobs:
                type: object
                nullable: true
                properties:
                  text_offset:
                    type: array
                    items:
                      type: integer
                  token_logprobs:
                    type: array
                    items:
                      type: number
                  tokens:
                    type: array
                    items:
                      type: string
                  top_logprobs:
                    type: array
                    items:
                      type: object
                      additionalProperties:
                        type: number
              text:
                type: string
        created:
          type: integer
          description: The Unix timestamp (in seconds) of when the completion was created.
        model:
          type: string
          description: The model used for completion.
        system_fingerprint:
          type: string
          description: >
            This fingerprint represents the backend configuration that the model
            runs with.


            Can be used in conjunction with the `seed` request parameter to
            understand when backend changes have been made that might impact
            determinism.
        object:
          type: string
          description: The object type, which is always "text_completion"
          enum:
            - text_completion
        usage:
          $ref: '#/components/schemas/CompletionUsage'
      required:
        - id
        - object
        - created
        - model
        - choices
      x-code-samples:
        name: The completion object
        legacy: true
        example: |
          {
            "id": "cmpl-uqkvlQyYK7bGYrRHQ0eXlWi7",
            "object": "text_completion",
            "created": 1589478378,
            "model": "gpt-4-turbo",
            "choices": [
              {
                "text": "\n\nThis is indeed a test",
                "index": 0,
                "logprobs": null,
                "finish_reason": "length"
              }
            ],
            "usage": {
              "prompt_tokens": 5,
              "completion_tokens": 7,
              "total_tokens": 12
            }
          }
    ChatCompletionRequestMessage:
      oneOf:
        - $ref: '#/components/schemas/ChatCompletionRequestSystemMessage'
        - $ref: '#/components/schemas/ChatCompletionRequestDeveloperMessage'
        - $ref: '#/components/schemas/ChatCompletionRequestUserMessage'
        - $ref: '#/components/schemas/ChatCompletionRequestAssistantMessage'
        - $ref: '#/components/schemas/ChatCompletionRequestToolMessage'
        - $ref: '#/components/schemas/ChatCompletionRequestFunctionMessage'
      x-oaiExpandable: true
    ResponseFormatText:
      type: object
      title: Text
      description: |
        Default response format. Used to generate text responses.
      properties:
        type:
          type: string
          description: The type of response format being defined. Always `text`.
          enum:
            - text
          x-stainless-const: true
      required:
        - type
    ResponseFormatJsonSchema:
      type: object
      title: JSON schema
      description: |
        JSON Schema response format. Used to generate structured JSON responses.
        Learn more about [Structured Outputs](/docs/guides/structured-outputs).
      properties:
        type:
          type: string
          description: The type of response format being defined. Always `json_schema`.
          enum:
            - json_schema
          x-stainless-const: true
        json_schema:
          type: object
          title: JSON schema
          description: |
            Structured Outputs configuration options, including a JSON Schema.
          properties:
            description:
              type: string
              description: >
                A description of what the response format is for, used by the
                model to

                determine how to respond in the format.
            name:
              type: string
              description: >
                The name of the response format. Must be a-z, A-Z, 0-9, or
                contain

                underscores and dashes, with a maximum length of 64.
            schema:
              $ref: '#/components/schemas/ResponseFormatJsonSchemaSchema'
            strict:
              type: boolean
              nullable: true
              default: false
              description: >
                Whether to enable strict schema adherence when generating the
                output.

                If set to true, the model will always follow the exact schema
                defined

                in the `schema` field. Only a subset of JSON Schema is supported
                when

                `strict` is `true`. To learn more, read the [Structured Outputs

                guide](/docs/guides/structured-outputs).
          required:
            - name
      required:
        - type
        - json_schema
    ResponseFormatJsonObject:
      type: object
      title: JSON object
      description: >
        JSON object response format. An older method of generating JSON
        responses.

        Using `json_schema` is recommended for models that support it. Note that
        the

        model will not generate JSON without a system or user message
        instructing it

        to do so.
      properties:
        type:
          type: string
          description: The type of response format being defined. Always `json_object`.
          enum:
            - json_object
          x-stainless-const: true
      required:
        - type
    ChatCompletionStreamOptions:
      description: >
        Options for streaming response. Only set this when you set `stream:
        true`.
      type: object
      nullable: true
      default: null
      properties:
        include_usage:
          type: boolean
          description: >
            If set, an additional chunk will be streamed before the `data:
            [DONE]` message. The `usage` field on this chunk shows the token
            usage statistics for the entire request, and the `choices` field
            will always be an empty array. All other chunks will also include a
            `usage` field, but with a null value.
    ChatCompletionTool:
      type: object
      properties:
        type:
          type: string
          enum:
            - function
          description: The type of the tool. Currently, only `function` is supported.
        function:
          $ref: '#/components/schemas/FunctionObject'
      required:
        - type
        - function
    ChatCompletionToolChoiceOption:
      description: >
        Controls which (if any) tool is called by the model.

        `none` means the model will not call any tool and instead generates a
        message.

        `auto` means the model can pick between generating a message or calling
        one or more tools.

        `required` means the model must call one or more tools.

        Specifying a particular tool via `{"type": "function", "function":
        {"name": "my_function"}}` forces the model to call that tool.


        `none` is the default when no tools are present. `auto` is the default
        if tools are present.
      oneOf:
        - type: string
          description: >
            `none` means the model will not call any tool and instead generates
            a message. `auto` means the model can pick between generating a
            message or calling one or more tools. `required` means the model
            must call one or more tools.
          enum:
            - none
            - auto
            - required
        - $ref: '#/components/schemas/ChatCompletionNamedToolChoice'
      x-oaiExpandable: true
    ParallelToolCalls:
      description: >-
        Whether to enable [parallel function
        calling](https://platform.openai.com/docs/guides/function-calling/parallel-function-calling)
        during tool use.
      type: boolean
      default: true
    ChatCompletionFunctionCallOption:
      type: object
      description: >
        Specifying a particular function via `{"name": "my_function"}` forces
        the model to call that function.
      properties:
        name:
          type: string
          description: The name of the function to call.
      required:
        - name
    ChatCompletionFunctions:
      type: object
      deprecated: true
      properties:
        description:
          type: string
          description: >-
            A description of what the function does, used by the model to choose
            when and how to call the function.
        name:
          type: string
          description: >-
            The name of the function to be called. Must be a-z, A-Z, 0-9, or
            contain underscores and dashes, with a maximum length of 64.
        parameters:
          $ref: '#/components/schemas/FunctionParameters'
      required:
        - name
    ChatCompletionResponseMessage:
      type: object
      description: A chat completion message generated by the model.
      properties:
        content:
          type: string
          description: The contents of the message.
          nullable: true
        tool_calls:
          $ref: '#/components/schemas/ChatCompletionMessageToolCalls'
        role:
          type: string
          enum:
            - assistant
          description: The role of the author of this message.
        function_call:
          type: object
          deprecated: true
          description: >-
            Deprecated and replaced by `tool_calls`. The name and arguments of a
            function that should be called, as generated by the model.
          properties:
            arguments:
              type: string
              description: >-
                The arguments to call the function with, as generated by the
                model in JSON format. Note that the model does not always
                generate valid JSON, and may hallucinate parameters not defined
                by your function schema. Validate the arguments in your code
                before calling your function.
            name:
              type: string
              description: The name of the function to call.
          required:
            - name
            - arguments
        content_blocks:
          nullable: true
          type: array
          description: >-
            The content blocks of the message. This is only present for certain
            providers with strict-open-ai-compliance flag set to false
          items:
            $ref: '#/components/schemas/ChatCompletionMessageContentBlock'
            type: object
      required:
        - role
        - content
    ChatCompletionTokenLogprob:
      type: object
      properties:
        token:
          description: The token.
          type: string
        logprob:
          description: >-
            The log probability of this token, if it is within the top 20 most
            likely tokens. Otherwise, the value `-9999.0` is used to signify
            that the token is very unlikely.
          type: number
        bytes:
          description: >-
            A list of integers representing the UTF-8 bytes representation of
            the token. Useful in instances where characters are represented by
            multiple tokens and their byte representations must be combined to
            generate the correct text representation. Can be `null` if there is
            no bytes representation for the token.
          type: array
          items:
            type: integer
          nullable: true
        top_logprobs:
          description: >-
            List of the most likely tokens and their log probability, at this
            token position. In rare cases, there may be fewer than the number of
            requested `top_logprobs` returned.
          type: array
          items:
            type: object
            properties:
              token:
                description: The token.
                type: string
              logprob:
                description: >-
                  The log probability of this token, if it is within the top 20
                  most likely tokens. Otherwise, the value `-9999.0` is used to
                  signify that the token is very unlikely.
                type: number
              bytes:
                description: >-
                  A list of integers representing the UTF-8 bytes representation
                  of the token. Useful in instances where characters are
                  represented by multiple tokens and their byte representations
                  must be combined to generate the correct text representation.
                  Can be `null` if there is no bytes representation for the
                  token.
                type: array
                items:
                  type: integer
                nullable: true
            required:
              - token
              - logprob
              - bytes
      required:
        - token
        - logprob
        - bytes
        - top_logprobs
    CompletionUsage:
      type: object
      description: Usage statistics for the completion request.
      properties:
        completion_tokens:
          type: integer
          description: Number of tokens in the generated completion.
        prompt_tokens:
          type: integer
          description: Number of tokens in the prompt.
        total_tokens:
          type: integer
          description: Total number of tokens used in the request (prompt + completion).
        completion_tokens_details:
          type: object
          nullable: true
          description: Breakdown of tokens used in a completion.
          properties:
            reasoning_tokens:
              type: integer
              description: Tokens generated by the model for reasoning.
            accepted_prediction_tokens:
              type: integer
              description: >-
                When using Predicted Outputs, the number of tokens in the
                prediction that appeared in the completion.
            rejected_prediction_tokens:
              type: integer
              description: >-
                When using Predicted Outputs, the number of tokens in the
                prediction that did not appear in the completion.
        prompt_tokens_details:
          type: object
          nullable: true
          description: Breakdown of tokens used in the prompt.
          properties:
            cached_tokens:
              type: integer
              description: Cached tokens present in the prompt.
      required:
        - prompt_tokens
        - completion_tokens
        - total_tokens
    ChatCompletionRequestSystemMessage:
      type: object
      title: System message
      properties:
        content:
          description: The contents of the system message.
          type: string
        role:
          type: string
          enum:
            - system
          description: The role of the messages author, in this case `system`.
        name:
          type: string
          description: >-
            An optional name for the participant. Provides the model information
            to differentiate between participants of the same role.
      required:
        - content
        - role
    ChatCompletionRequestDeveloperMessage:
      type: object
      title: Developer message
      description: >-
        New role by OpenAI for select models. Must be explicitly used for models
        that support it. When used with incompatible models or providers,
        Portkey automatically converts it to a system role.
      properties:
        content:
          description: The contents of the Developer message.
          type: string
        role:
          type: string
          enum:
            - developer
          description: The role of the messages author, in this case `Developer`.
        name:
          type: string
          description: >-
            An optional name for the participant. Provides the model information
            to differentiate between participants of the same role.
      required:
        - content
        - role
    ChatCompletionRequestUserMessage:
      type: object
      title: User message
      properties:
        content:
          description: |
            The contents of the user message.
          oneOf:
            - type: string
              description: The text contents of the message.
              title: Text content
            - type: array
              description: >-
                An array of content parts with a defined type, each can be of
                type `text` or `image_url` when passing in images. You can pass
                multiple images by adding multiple `image_url` content parts.
                Image input is only supported when using the
                `gpt-4-visual-preview` model.
              title: Array of content parts
              items:
                $ref: '#/components/schemas/ChatCompletionRequestMessageContentPart'
              minItems: 1
          x-oaiExpandable: true
        role:
          type: string
          enum:
            - user
          description: The role of the messages author, in this case `user`.
        name:
          type: string
          description: >-
            An optional name for the participant. Provides the model information
            to differentiate between participants of the same role.
      required:
        - content
        - role
    ChatCompletionRequestAssistantMessage:
      type: object
      title: Assistant message
      properties:
        content:
          nullable: true
          type: string
          description: >
            The contents of the assistant message. Required unless `tool_calls`
            or `function_call` is specified.
        role:
          type: string
          enum:
            - assistant
          description: The role of the messages author, in this case `assistant`.
        name:
          type: string
          description: >-
            An optional name for the participant. Provides the model information
            to differentiate between participants of the same role.
        tool_calls:
          $ref: '#/components/schemas/ChatCompletionMessageToolCalls'
        function_call:
          type: object
          deprecated: true
          description: >-
            Deprecated and replaced by `tool_calls`. The name and arguments of a
            function that should be called, as generated by the model.
          nullable: true
          properties:
            arguments:
              type: string
              description: >-
                The arguments to call the function with, as generated by the
                model in JSON format. Note that the model does not always
                generate valid JSON, and may hallucinate parameters not defined
                by your function schema. Validate the arguments in your code
                before calling your function.
            name:
              type: string
              description: The name of the function to call.
          required:
            - arguments
            - name
      required:
        - role
    ChatCompletionRequestToolMessage:
      type: object
      title: Tool message
      properties:
        role:
          type: string
          enum:
            - tool
          description: The role of the messages author, in this case `tool`.
        content:
          type: string
          description: The contents of the tool message.
        tool_call_id:
          type: string
          description: Tool call that this message is responding to.
      required:
        - role
        - content
        - tool_call_id
    ChatCompletionRequestFunctionMessage:
      type: object
      title: Function message
      deprecated: true
      properties:
        role:
          type: string
          enum:
            - function
          description: The role of the messages author, in this case `function`.
        content:
          nullable: true
          type: string
          description: The contents of the function message.
        name:
          type: string
          description: The name of the function to call.
      required:
        - role
        - content
        - name
    ResponseFormatJsonSchemaSchema:
      type: object
      title: JSON schema
      description: |
        The schema for the response format, described as a JSON Schema object.
        Learn how to build JSON schemas [here](https://json-schema.org/).
      additionalProperties: true
    FunctionObject:
      type: object
      properties:
        description:
          type: string
          description: >-
            A description of what the function does, used by the model to choose
            when and how to call the function.
        name:
          type: string
          description: >-
            The name of the function to be called. Must be a-z, A-Z, 0-9, or
            contain underscores and dashes, with a maximum length of 64.
        parameters:
          $ref: '#/components/schemas/FunctionParameters'
        strict:
          type: boolean
          nullable: true
          default: false
          description: >-
            Whether to enable strict schema adherence when generating the
            function call. If set to true, the model will follow the exact
            schema defined in the `parameters` field. Only a subset of JSON
            Schema is supported when `strict` is `true`. Learn more about
            Structured Outputs in the [function calling
            guide](docs/guides/function-calling).
      required:
        - name
    ChatCompletionNamedToolChoice:
      type: object
      description: >-
        Specifies a tool the model should use. Use to force the model to call a
        specific function.
      properties:
        type:
          type: string
          enum:
            - function
          description: The type of the tool. Currently, only `function` is supported.
        function:
          type: object
          properties:
            name:
              type: string
              description: The name of the function to call.
          required:
            - name
      required:
        - type
        - function
    FunctionParameters:
      type: object
      description: >-
        The parameters the functions accepts, described as a JSON Schema object.
        See the
        [guide](https://platform.openai.com/docs/guides/function-calling) for
        examples, and the [JSON Schema
        reference](https://json-schema.org/understanding-json-schema/) for
        documentation about the format. 


        Omitting `parameters` defines a function with an empty parameter list.
      additionalProperties: true
    ChatCompletionMessageToolCalls:
      type: array
      description: The tool calls generated by the model, such as function calls.
      items:
        $ref: '#/components/schemas/ChatCompletionMessageToolCall'
    ChatCompletionMessageContentBlock:
      type: object
      description: A block of content in a chat completion message.
      oneOf:
        - $ref: '#/components/schemas/ChatCompletionRequestMessageContentPartText'
        - $ref: '#/components/schemas/ChatCompletionMessageContentPartThinking'
        - $ref: >-
            #/components/schemas/ChatCompletionMessageContentPartRedactedThinking
    ChatCompletionRequestMessageContentPart:
      oneOf:
        - $ref: '#/components/schemas/ChatCompletionRequestMessageContentPartText'
        - $ref: '#/components/schemas/ChatCompletionRequestMessageContentPartImage'
      x-oaiExpandable: true
    ChatCompletionMessageToolCall:
      type: object
      properties:
        id:
          type: string
          description: The ID of the tool call.
        type:
          type: string
          enum:
            - function
          description: The type of the tool. Currently, only `function` is supported.
        function:
          type: object
          description: The function that the model called.
          properties:
            name:
              type: string
              description: The name of the function to call.
            arguments:
              type: string
              description: >-
                The arguments to call the function with, as generated by the
                model in JSON format. Note that the model does not always
                generate valid JSON, and may hallucinate parameters not defined
                by your function schema. Validate the arguments in your code
                before calling your function.
          required:
            - name
            - arguments
      required:
        - id
        - type
        - function
    ChatCompletionRequestMessageContentPartText:
      type: object
      title: Text content part
      properties:
        type:
          type: string
          enum:
            - text
          description: The type of the content part.
        text:
          type: string
          description: The text content.
      required:
        - type
        - text
    ChatCompletionMessageContentPartThinking:
      type: object
      title: Thinking content part
      properties:
        type:
          type: string
          enum:
            - thinking
          description: The type of the content part.
        thinking:
          type: string
          description: The thinking content.
      required:
        - type
        - thinking
    ChatCompletionMessageContentPartRedactedThinking:
      type: object
      title: Redacted thinking content part
      properties:
        type:
          type: string
          enum:
            - redacted_thinking
          description: The type of the content part.
        data:
          type: string
          description: The redacted thinking content.
      required:
        - type
        - data
    ChatCompletionRequestMessageContentPartImage:
      type: object
      title: Image content part
      properties:
        type:
          type: string
          enum:
            - image_url
          description: The type of the content part.
        image_url:
          type: object
          properties:
            url:
              type: string
              description: Either a URL of the image or the base64 encoded image data.
              format: uri
            detail:
              type: string
              description: >-
                Specifies the detail level of the image. Learn more in the
                [Vision
                guide](https://platform.openai.com/docs/guides/vision/low-or-high-fidelity-image-understanding).
              enum:
                - auto
                - low
                - high
              default: auto
          required:
            - url
      required:
        - type
        - image_url
  securitySchemes:
    Portkey-Key:
      type: apiKey
      in: header
      name: x-portkey-api-key

````