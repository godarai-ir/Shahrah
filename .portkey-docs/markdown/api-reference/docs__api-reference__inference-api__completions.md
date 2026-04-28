> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Completions



## OpenAPI

````yaml post /completions
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
  /completions:
    servers:
      - url: https://api.portkey.ai/v1
        description: Portkey API Public Endpoint
      - url: SELF_HOSTED_GATEWAY_URL
        description: Self-Hosted Gateway URL
    post:
      tags:
        - Completions
      summary: Completions
      operationId: createCompletion
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateCompletionRequest'
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CreateCompletionResponse'
      security:
        - Portkey-Key: []
          Virtual-Key: []
        - Portkey-Key: []
          Provider-Auth: []
          Provider-Name: []
        - Portkey-Key: []
          Config: []
        - Portkey-Key: []
          Provider-Auth: []
          Provider-Name: []
          Custom-Host: []
      x-code-samples:
        - lang: curl
          label: Default
          source: |
            curl https://api.portkey.ai/v1/completions \
              -H "Content-Type: application/json" \
              -H "x-portkey-api-key: $PORTKEY_API_KEY" \
              -H "x-portkey-virtual-key: $PORTKEY_PROVIDER_VIRTUAL_KEY" \
              -d '{
                "model": "gpt-3.5-turbo-instruct",
                "prompt": "Say this is a test",
                "max_tokens": 7,
                "temperature": 0
              }'
        - lang: python
          label: Default
          source: |
            from portkey_ai import Portkey

            portkey = Portkey(
              api_key = "PORTKEY_API_KEY",
              virtual_key = "PROVIDER_VIRTUAL_KEY"
            )

            response = portkey.completions.create(
              model="gpt-3.5-turbo-instruct",
              prompt="Say this is a test",
              max_tokens=7,
              temperature=0
            )

            print(response)
        - lang: javascript
          label: Default
          source: |
            import Portkey from 'portkey-ai';

            const portkey = new Portkey({
              apiKey: 'PORTKEY_API_KEY',
              virtualKey: 'PROVIDER_VIRTUAL_KEY'
            });

            async function main() {
              const response = await portkey.completions.create({
                model: "gpt-3.5-turbo-instruct",
                prompt: "Say this is a test.",
                max_tokens: 7,
                temperature: 0,
              });

              console.log(response);
            }

            main();
        - lang: javascript
          label: Self-Hosted
          source: |
            import Portkey from 'portkey-ai';

            const portkey = new Portkey({
              apiKey: 'PORTKEY_API_KEY',
              virtualKey: 'PROVIDER_VIRTUAL_KEY',
              baseUrl: 'SELF_HOSTED_GATEWAY_URL'
            });

            async function main() {
              const response = await client.completions.create({
                model: "gpt-3.5-turbo-instruct",
                prompt: "Say this is a test.",
                max_tokens: 7,
                temperature: 0,
              });

              console.log(response);
            }

            main();
        - lang: python
          label: Self-Hosted
          source: |
            from portkey_ai import Portkey

            portkey = Portkey(
              api_key = "PORTKEY_API_KEY",
              virtual_key = "PROVIDER_VIRTUAL_KEY",
              base_url = "SELF_HOSTED_GATEWAY_URL"
            )

            response = portkey.completions.create(
              model="gpt-3.5-turbo-instruct",
              prompt="Say this is a test",
              max_tokens=7,
              temperature=0
            )

            print(response)
        - lang: curl
          label: Self-Hosted
          source: |
            curl https://SELF_HOSTED_GATEWAY_URL/completions \
              -H "Content-Type: application/json" \
              -H "x-portkey-api-key: $PORTKEY_API_KEY" \
              -H "x-portkey-virtual-key: $PORTKEY_PROVIDER_VIRTUAL_KEY" \
              -d '{
                "model": "gpt-3.5-turbo-instruct",
                "prompt": "Say this is a test",
                "max_tokens": 7,
                "temperature": 0
              }'
components:
  schemas:
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
  securitySchemes:
    Portkey-Key:
      type: apiKey
      in: header
      name: x-portkey-api-key
    Virtual-Key:
      type: apiKey
      in: header
      name: x-portkey-virtual-key
    Provider-Auth:
      type: http
      scheme: bearer
    Provider-Name:
      type: apiKey
      in: header
      name: x-portkey-provider
    Config:
      type: apiKey
      in: header
      name: x-portkey-config
    Custom-Host:
      type: apiKey
      in: header
      name: x-portkey-custom-host

````