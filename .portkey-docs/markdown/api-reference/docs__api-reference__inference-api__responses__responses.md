> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Create a Response



## OpenAPI

````yaml post /responses
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
  /responses:
    servers:
      - url: https://api.portkey.ai/v1
        description: Portkey API Public Endpoint
      - url: SELF_HOSTED_GATEWAY_URL
        description: Self-Hosted Gateway URL
    post:
      tags:
        - Responses
      summary: |
        Creates a model response
      operationId: createResponse
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateResponse'
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Response'
            text/event-stream:
              schema:
                $ref: '#/components/schemas/ResponseStreamEvent'
components:
  schemas:
    CreateResponse:
      allOf:
        - $ref: '#/components/schemas/CreateModelResponseProperties'
        - $ref: '#/components/schemas/ResponseProperties'
        - type: object
          properties:
            input:
              description: >
                Text, image, or file inputs to the model, used to generate a
                response.


                Learn more:

                - [Text inputs and
                outputs](https://platform.openai.com/docs/guides/text?api-mode=responses)

                - [Image
                inputs](https://platform.openai.com/docs/guides/images-vision?api-mode=responses)

                - [File
                inputs](https://platform.openai.com/docs/guides/pdf-files?api-mode=responses)

                - [Conversation
                state](https://platform.openai.com/docs/guides/conversation-state?api-mode=responses)

                - [Function
                calling](https://platform.openai.com/docs/guides/function-calling?api-mode=responses)
              x-oaiExpandable: true
              oneOf:
                - type: string
                  title: Text input
                  description: >
                    A text input to the model, equivalent to a text input with
                    the 

                    `user` role.
                - type: array
                  title: Input item list
                  description: |
                    A list of one or many input items to the model, containing 
                    different content types.
                  items:
                    $ref: '#/components/schemas/InputItem'
                    x-oaiExpandable: true
            include:
              type: array
              description: >
                Specify additional output data to include in the model response.
                Currently

                supported values are:

                - `file_search_call.results`: Include the search results of
                  the file search tool call.
                - `message.input_image.image_url`: Include image urls from the
                input message.

                - `computer_call_output.output.image_url`: Include image urls
                from the computer call output.
              items:
                $ref: '#/components/schemas/Includable'
                x-oaiExpandable: true
              nullable: true
            parallel_tool_calls:
              type: boolean
              description: |
                Whether to allow the model to run tool calls in parallel.
              default: true
              nullable: true
            store:
              type: boolean
              description: >
                Whether to store the generated model response for later
                retrieval via

                API.
              default: true
              nullable: true
            stream:
              description: >
                If set to true, the model response data will be streamed to the
                client

                as it is generated using [server-sent
                events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events#Event_stream_format).

                See the [Streaming section
                below](/docs/api-reference/responses-streaming)

                for more information.
              type: boolean
              nullable: true
              default: false
          required:
            - model
            - input
    Response:
      allOf:
        - $ref: '#/components/schemas/ModelResponseProperties'
        - $ref: '#/components/schemas/ResponseProperties'
        - type: object
          properties:
            id:
              type: string
              description: |
                Unique identifier for this Response.
            object:
              type: string
              description: |
                The object type of this resource - always set to `response`.
              enum:
                - response
              x-stainless-const: true
            status:
              type: string
              description: >
                The status of the response generation. One of `completed`,
                `failed`, 

                `in_progress`, or `incomplete`.
              enum:
                - completed
                - failed
                - in_progress
                - incomplete
            created_at:
              type: number
              description: |
                Unix timestamp (in seconds) of when this Response was created.
            error:
              $ref: '#/components/schemas/ResponseError'
            incomplete_details:
              type: object
              nullable: true
              description: |
                Details about why the response is incomplete.
              properties:
                reason:
                  type: string
                  description: The reason why the response is incomplete.
                  enum:
                    - max_output_tokens
                    - content_filter
            output:
              type: array
              x-oaiExpandable: true
              description: >
                An array of content items generated by the model.


                - The length and order of items in the `output` array is
                dependent
                  on the model's response.
                - Rather than accessing the first item in the `output` array
                and 
                  assuming it's an `assistant` message with the content generated by
                  the model, you might consider using the `output_text` property where
                  supported in SDKs.
              items:
                $ref: '#/components/schemas/OutputItem'
                x-oaiExpandable: true
            output_text:
              type: string
              nullable: true
              description: >
                SDK-only convenience property that contains the aggregated text
                output 

                from all `output_text` items in the `output` array, if any are
                present. 

                Supported in the Python and JavaScript SDKs.
              x-oaiSupportedSDKs:
                - python
                - javascript
            usage:
              $ref: '#/components/schemas/ResponseUsage'
            parallel_tool_calls:
              type: boolean
              description: |
                Whether to allow the model to run tool calls in parallel.
              default: true
          required:
            - id
            - object
            - created_at
            - error
            - incomplete_details
            - instructions
            - model
            - tools
            - output
            - parallel_tool_calls
            - metadata
            - tool_choice
            - temperature
            - top_p
    ResponseStreamEvent:
      anyOf:
        - $ref: '#/components/schemas/ResponseAudioDeltaEvent'
        - $ref: '#/components/schemas/ResponseAudioDoneEvent'
        - $ref: '#/components/schemas/ResponseAudioTranscriptDeltaEvent'
        - $ref: '#/components/schemas/ResponseAudioTranscriptDoneEvent'
        - $ref: '#/components/schemas/ResponseCodeInterpreterCallCodeDeltaEvent'
        - $ref: '#/components/schemas/ResponseCodeInterpreterCallCodeDoneEvent'
        - $ref: '#/components/schemas/ResponseCodeInterpreterCallCompletedEvent'
        - $ref: '#/components/schemas/ResponseCodeInterpreterCallInProgressEvent'
        - $ref: '#/components/schemas/ResponseCodeInterpreterCallInterpretingEvent'
        - $ref: '#/components/schemas/ResponseCompletedEvent'
        - $ref: '#/components/schemas/ResponseContentPartAddedEvent'
        - $ref: '#/components/schemas/ResponseContentPartDoneEvent'
        - $ref: '#/components/schemas/ResponseCreatedEvent'
        - $ref: '#/components/schemas/ResponseErrorEvent'
        - $ref: '#/components/schemas/ResponseFileSearchCallCompletedEvent'
        - $ref: '#/components/schemas/ResponseFileSearchCallInProgressEvent'
        - $ref: '#/components/schemas/ResponseFileSearchCallSearchingEvent'
        - $ref: '#/components/schemas/ResponseFunctionCallArgumentsDeltaEvent'
        - $ref: '#/components/schemas/ResponseFunctionCallArgumentsDoneEvent'
        - $ref: '#/components/schemas/ResponseInProgressEvent'
        - $ref: '#/components/schemas/ResponseFailedEvent'
        - $ref: '#/components/schemas/ResponseIncompleteEvent'
        - $ref: '#/components/schemas/ResponseOutputItemAddedEvent'
        - $ref: '#/components/schemas/ResponseOutputItemDoneEvent'
        - $ref: '#/components/schemas/ResponseRefusalDeltaEvent'
        - $ref: '#/components/schemas/ResponseRefusalDoneEvent'
        - $ref: '#/components/schemas/ResponseTextAnnotationDeltaEvent'
        - $ref: '#/components/schemas/ResponseTextDeltaEvent'
        - $ref: '#/components/schemas/ResponseTextDoneEvent'
        - $ref: '#/components/schemas/ResponseWebSearchCallCompletedEvent'
        - $ref: '#/components/schemas/ResponseWebSearchCallInProgressEvent'
        - $ref: '#/components/schemas/ResponseWebSearchCallSearchingEvent'
      discriminator:
        propertyName: type
    CreateModelResponseProperties:
      allOf:
        - $ref: '#/components/schemas/ModelResponseProperties'
    ResponseProperties:
      type: object
      properties:
        previous_response_id:
          type: string
          description: |
            The unique ID of the previous response to the model. Use this to
            create multi-turn conversations. Learn more about 
            [conversation state](/docs/guides/conversation-state).
          nullable: true
        model:
          $ref: '#/components/schemas/ModelIdsResponses'
          description: >
            Model ID used to generate the response, like `gpt-4o` or `o1`.
            OpenAI

            offers a wide range of models with different capabilities,
            performance

            characteristics, and price points. Refer to the [model
            guide](/docs/models)

            to browse and compare available models.
        reasoning:
          $ref: '#/components/schemas/Reasoning'
          nullable: true
        max_output_tokens:
          description: >
            An upper bound for the number of tokens that can be generated for a
            response, including visible output tokens and [reasoning
            tokens](/docs/guides/reasoning).
          type: integer
          nullable: true
        instructions:
          type: string
          description: >
            Inserts a system (or developer) message as the first item in the
            model's context.


            When using along with `previous_response_id`, the instructions from
            a previous

            response will be not be carried over to the next response. This
            makes it simple

            to swap out system (or developer) messages in new responses.
          nullable: true
        text:
          type: object
          description: >
            Configuration options for a text response from the model. Can be
            plain

            text or structured JSON data. Learn more:

            - [Text inputs and outputs](/docs/guides/text)

            - [Structured Outputs](/docs/guides/structured-outputs)
          properties:
            format:
              $ref: '#/components/schemas/TextResponseFormatConfiguration'
        tools:
          type: array
          description: >
            An array of tools the model may call while generating a response.
            You 

            can specify which tool to use by setting the `tool_choice`
            parameter.


            The two categories of tools you can provide the model are:


            - **Built-in tools**: Tools that are provided by OpenAI that extend
            the
              model's capabilities, like [web search](/docs/guides/tools-web-search)
              or [file search](/docs/guides/tools-file-search). Learn more about
              [built-in tools](/docs/guides/tools).
            - **Function calls (custom tools)**: Functions that are defined by
            you,
              enabling the model to call your own code. Learn more about
              [function calling](/docs/guides/function-calling).
          items:
            $ref: '#/components/schemas/Tool'
        tool_choice:
          description: >
            How the model should select which tool (or tools) to use when
            generating

            a response. See the `tools` parameter to see how to specify which
            tools

            the model can call.
          x-oaiExpandable: true
          oneOf:
            - $ref: '#/components/schemas/ToolChoiceOptions'
            - $ref: '#/components/schemas/ToolChoiceTypes'
            - $ref: '#/components/schemas/ToolChoiceFunction'
        truncation:
          type: string
          description: >
            The truncation strategy to use for the model response.

            - `auto`: If the context of this response and previous ones exceeds
              the model's context window size, the model will truncate the 
              response to fit the context window by dropping input items in the
              middle of the conversation. 
            - `disabled` (default): If a model response will exceed the context
            window 
              size for a model, the request will fail with a 400 error.
          enum:
            - auto
            - disabled
          nullable: true
          default: disabled
    InputItem:
      oneOf:
        - $ref: '#/components/schemas/EasyInputMessage'
        - $ref: '#/components/schemas/Item'
          type: object
          title: Item
          description: |
            An item representing part of the context for the response to be 
            generated by the model. Can contain text, images, and audio inputs,
            as well as previous assistant responses and tool call outputs.
        - $ref: '#/components/schemas/ItemReference'
      discriminator:
        propertyName: type
    Includable:
      type: string
      description: >
        Specify additional output data to include in the model response.
        Currently

        supported values are:

        - `file_search_call.results`: Include the search results of
          the file search tool call.
        - `message.input_image.image_url`: Include image urls from the input
        message.

        - `computer_call_output.output.image_url`: Include image urls from the
        computer call output.
      enum:
        - file_search_call.results
        - message.input_image.image_url
        - computer_call_output.output.image_url
    ModelResponseProperties:
      type: object
      properties:
        metadata:
          $ref: '#/components/schemas/Metadata'
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
            sampling,

            where the model considers the results of the tokens with top_p
            probability

            mass. So 0.1 means only the tokens comprising the top 10%
            probability mass

            are considered.


            We generally recommend altering this or `temperature` but not both.
        user:
          type: string
          example: user-1234
          description: >
            A unique identifier representing your end-user, which can help
            OpenAI to monitor and detect abuse. [Learn
            more](/docs/guides/safety-best-practices#end-user-ids).
    ResponseError:
      type: object
      description: |
        An error object returned when the model fails to generate a Response.
      nullable: true
      properties:
        code:
          $ref: '#/components/schemas/ResponseErrorCode'
        message:
          type: string
          description: |
            A human-readable description of the error.
      required:
        - code
        - message
    OutputItem:
      anyOf:
        - $ref: '#/components/schemas/OutputMessage'
        - $ref: '#/components/schemas/FileSearchToolCall'
        - $ref: '#/components/schemas/FunctionToolCall'
        - $ref: '#/components/schemas/WebSearchToolCall'
        - $ref: '#/components/schemas/ComputerToolCall'
        - $ref: '#/components/schemas/ReasoningItem'
      x-oaiExpandable: true
      discriminator:
        propertyName: type
    ResponseUsage:
      type: object
      description: |
        Represents token usage details including input tokens, output tokens,
        a breakdown of output tokens, and the total tokens used.
      properties:
        input_tokens:
          type: integer
          description: The number of input tokens.
        input_tokens_details:
          type: object
          description: A detailed breakdown of the input tokens.
          properties:
            cached_tokens:
              type: integer
              description: |
                The number of tokens that were retrieved from the cache. 
                [More on prompt caching](/docs/guides/prompt-caching).
          required:
            - cached_tokens
        output_tokens:
          type: integer
          description: The number of output tokens.
        output_tokens_details:
          type: object
          description: A detailed breakdown of the output tokens.
          properties:
            reasoning_tokens:
              type: integer
              description: The number of reasoning tokens.
          required:
            - reasoning_tokens
        total_tokens:
          type: integer
          description: The total number of tokens used.
      required:
        - input_tokens
        - input_tokens_details
        - output_tokens
        - output_tokens_details
        - total_tokens
    ResponseAudioDeltaEvent:
      type: object
      description: Emitted when there is a partial audio response.
      properties:
        type:
          type: string
          description: |
            The type of the event. Always `response.audio.delta`.
          enum:
            - response.audio.delta
          x-stainless-const: true
        delta:
          type: string
          description: |
            A chunk of Base64 encoded response audio bytes.
      required:
        - type
        - delta
    ResponseAudioDoneEvent:
      type: object
      description: Emitted when the audio response is complete.
      properties:
        type:
          type: string
          description: |
            The type of the event. Always `response.audio.done`.
          enum:
            - response.audio.done
          x-stainless-const: true
      required:
        - type
        - response_id
    ResponseAudioTranscriptDeltaEvent:
      type: object
      description: Emitted when there is a partial transcript of audio.
      properties:
        type:
          type: string
          description: |
            The type of the event. Always `response.audio.transcript.delta`.
          enum:
            - response.audio.transcript.delta
          x-stainless-const: true
        delta:
          type: string
          description: |
            The partial transcript of the audio response.
      required:
        - type
        - response_id
        - delta
    ResponseAudioTranscriptDoneEvent:
      type: object
      description: Emitted when the full audio transcript is completed.
      properties:
        type:
          type: string
          description: |
            The type of the event. Always `response.audio.transcript.done`.
          enum:
            - response.audio.transcript.done
          x-stainless-const: true
      required:
        - type
        - response_id
    ResponseCodeInterpreterCallCodeDeltaEvent:
      type: object
      description: Emitted when a partial code snippet is added by the code interpreter.
      properties:
        type:
          type: string
          description: >
            The type of the event. Always
            `response.code_interpreter_call.code.delta`.
          enum:
            - response.code_interpreter_call.code.delta
          x-stainless-const: true
        output_index:
          type: integer
          description: >
            The index of the output item that the code interpreter call is in
            progress.
        delta:
          type: string
          description: |
            The partial code snippet added by the code interpreter.
      required:
        - type
        - response_id
        - output_index
        - delta
    ResponseCodeInterpreterCallCodeDoneEvent:
      type: object
      description: Emitted when code snippet output is finalized by the code interpreter.
      properties:
        type:
          type: string
          description: >
            The type of the event. Always
            `response.code_interpreter_call.code.done`.
          enum:
            - response.code_interpreter_call.code.done
          x-stainless-const: true
        output_index:
          type: integer
          description: >
            The index of the output item that the code interpreter call is in
            progress.
        code:
          type: string
          description: |
            The final code snippet output by the code interpreter.
      required:
        - type
        - response_id
        - output_index
        - code
    ResponseCodeInterpreterCallCompletedEvent:
      type: object
      description: Emitted when the code interpreter call is completed.
      properties:
        type:
          type: string
          description: >
            The type of the event. Always
            `response.code_interpreter_call.completed`.
          enum:
            - response.code_interpreter_call.completed
          x-stainless-const: true
        output_index:
          type: integer
          description: >
            The index of the output item that the code interpreter call is in
            progress.
        code_interpreter_call:
          $ref: '#/components/schemas/CodeInterpreterToolCall'
      required:
        - type
        - response_id
        - output_index
        - code_interpreter_call
    ResponseCodeInterpreterCallInProgressEvent:
      type: object
      description: Emitted when a code interpreter call is in progress.
      properties:
        type:
          type: string
          description: >
            The type of the event. Always
            `response.code_interpreter_call.in_progress`.
          enum:
            - response.code_interpreter_call.in_progress
          x-stainless-const: true
        output_index:
          type: integer
          description: >
            The index of the output item that the code interpreter call is in
            progress.
        code_interpreter_call:
          $ref: '#/components/schemas/CodeInterpreterToolCall'
      required:
        - type
        - response_id
        - output_index
        - code_interpreter_call
    ResponseCodeInterpreterCallInterpretingEvent:
      type: object
      description: >-
        Emitted when the code interpreter is actively interpreting the code
        snippet.
      properties:
        type:
          type: string
          description: >
            The type of the event. Always
            `response.code_interpreter_call.interpreting`.
          enum:
            - response.code_interpreter_call.interpreting
          x-stainless-const: true
        output_index:
          type: integer
          description: >
            The index of the output item that the code interpreter call is in
            progress.
        code_interpreter_call:
          $ref: '#/components/schemas/CodeInterpreterToolCall'
      required:
        - type
        - response_id
        - output_index
        - code_interpreter_call
    ResponseCompletedEvent:
      type: object
      description: Emitted when the model response is complete.
      properties:
        type:
          type: string
          description: |
            The type of the event. Always `response.completed`.
          enum:
            - response.completed
          x-stainless-const: true
        response:
          $ref: '#/components/schemas/Response'
          description: |
            Properties of the completed response.
      required:
        - type
        - response
    ResponseContentPartAddedEvent:
      type: object
      description: Emitted when a new content part is added.
      properties:
        type:
          type: string
          description: |
            The type of the event. Always `response.content_part.added`.
          enum:
            - response.content_part.added
          x-stainless-const: true
        item_id:
          type: string
          description: |
            The ID of the output item that the content part was added to.
        output_index:
          type: integer
          description: |
            The index of the output item that the content part was added to.
        content_index:
          type: integer
          description: |
            The index of the content part that was added.
        part:
          $ref: '#/components/schemas/OutputContent'
          x-oaiExpandable: true
          description: |
            The content part that was added.
      required:
        - type
        - item_id
        - output_index
        - content_index
        - part
    ResponseContentPartDoneEvent:
      type: object
      description: Emitted when a content part is done.
      properties:
        type:
          type: string
          description: |
            The type of the event. Always `response.content_part.done`.
          enum:
            - response.content_part.done
          x-stainless-const: true
        item_id:
          type: string
          description: |
            The ID of the output item that the content part was added to.
        output_index:
          type: integer
          description: |
            The index of the output item that the content part was added to.
        content_index:
          type: integer
          description: |
            The index of the content part that is done.
        part:
          $ref: '#/components/schemas/OutputContent'
          x-oaiExpandable: true
          description: |
            The content part that is done.
      required:
        - type
        - item_id
        - output_index
        - content_index
        - part
    ResponseCreatedEvent:
      type: object
      description: |
        An event that is emitted when a response is created.
      properties:
        type:
          type: string
          description: |
            The type of the event. Always `response.created`.
          enum:
            - response.created
          x-stainless-const: true
        response:
          $ref: '#/components/schemas/Response'
          description: |
            The response that was created.
      required:
        - type
        - response
    ResponseErrorEvent:
      type: object
      description: Emitted when an error occurs.
      properties:
        type:
          type: string
          description: |
            The type of the event. Always `error`.
          enum:
            - error
          x-stainless-const: true
        code:
          type: string
          description: |
            The error code.
          nullable: true
        message:
          type: string
          description: |
            The error message.
        param:
          type: string
          description: |
            The error parameter.
          nullable: true
      required:
        - type
        - code
        - message
        - param
    ResponseFileSearchCallCompletedEvent:
      type: object
      description: Emitted when a file search call is completed (results found).
      properties:
        type:
          type: string
          description: |
            The type of the event. Always `response.file_search_call.completed`.
          enum:
            - response.file_search_call.completed
          x-stainless-const: true
        output_index:
          type: integer
          description: |
            The index of the output item that the file search call is initiated.
        item_id:
          type: string
          description: |
            The ID of the output item that the file search call is initiated.
      required:
        - type
        - output_index
        - item_id
    ResponseFileSearchCallInProgressEvent:
      type: object
      description: Emitted when a file search call is initiated.
      properties:
        type:
          type: string
          description: >
            The type of the event. Always
            `response.file_search_call.in_progress`.
          enum:
            - response.file_search_call.in_progress
          x-stainless-const: true
        output_index:
          type: integer
          description: |
            The index of the output item that the file search call is initiated.
        item_id:
          type: string
          description: |
            The ID of the output item that the file search call is initiated.
      required:
        - type
        - output_index
        - item_id
    ResponseFileSearchCallSearchingEvent:
      type: object
      description: Emitted when a file search is currently searching.
      properties:
        type:
          type: string
          description: |
            The type of the event. Always `response.file_search_call.searching`.
          enum:
            - response.file_search_call.searching
          x-stainless-const: true
        output_index:
          type: integer
          description: |
            The index of the output item that the file search call is searching.
        item_id:
          type: string
          description: |
            The ID of the output item that the file search call is initiated.
      required:
        - type
        - output_index
        - item_id
    ResponseFunctionCallArgumentsDeltaEvent:
      type: object
      description: Emitted when there is a partial function-call arguments delta.
      properties:
        type:
          type: string
          description: >
            The type of the event. Always
            `response.function_call_arguments.delta`.
          enum:
            - response.function_call_arguments.delta
          x-stainless-const: true
        item_id:
          type: string
          description: >
            The ID of the output item that the function-call arguments delta is
            added to.
        output_index:
          type: integer
          description: >
            The index of the output item that the function-call arguments delta
            is added to.
        delta:
          type: string
          description: |
            The function-call arguments delta that is added.
      required:
        - type
        - item_id
        - output_index
        - delta
    ResponseFunctionCallArgumentsDoneEvent:
      type: object
      description: Emitted when function-call arguments are finalized.
      properties:
        type:
          type: string
          enum:
            - response.function_call_arguments.done
          x-stainless-const: true
        item_id:
          type: string
          description: The ID of the item.
        output_index:
          type: integer
          description: The index of the output item.
        arguments:
          type: string
          description: The function-call arguments.
      required:
        - type
        - item_id
        - output_index
        - arguments
    ResponseInProgressEvent:
      type: object
      description: Emitted when the response is in progress.
      properties:
        type:
          type: string
          description: |
            The type of the event. Always `response.in_progress`.
          enum:
            - response.in_progress
          x-stainless-const: true
        response:
          $ref: '#/components/schemas/Response'
          description: |
            The response that is in progress.
      required:
        - type
        - response
    ResponseFailedEvent:
      type: object
      description: |
        An event that is emitted when a response fails.
      properties:
        type:
          type: string
          description: |
            The type of the event. Always `response.failed`.
          enum:
            - response.failed
          x-stainless-const: true
        response:
          $ref: '#/components/schemas/Response'
          description: |
            The response that failed.
      required:
        - type
        - response
    ResponseIncompleteEvent:
      type: object
      description: |
        An event that is emitted when a response finishes as incomplete.
      properties:
        type:
          type: string
          description: |
            The type of the event. Always `response.incomplete`.
          enum:
            - response.incomplete
          x-stainless-const: true
        response:
          $ref: '#/components/schemas/Response'
          description: |
            The response that was incomplete.
      required:
        - type
        - response
    ResponseOutputItemAddedEvent:
      type: object
      description: Emitted when a new output item is added.
      properties:
        type:
          type: string
          description: |
            The type of the event. Always `response.output_item.added`.
          enum:
            - response.output_item.added
          x-stainless-const: true
        output_index:
          type: integer
          description: |
            The index of the output item that was added.
        item:
          $ref: '#/components/schemas/OutputItem'
          x-oaiExpandable: true
          description: |
            The output item that was added.
      required:
        - type
        - output_index
        - item
    ResponseOutputItemDoneEvent:
      type: object
      description: Emitted when an output item is marked done.
      properties:
        type:
          type: string
          description: |
            The type of the event. Always `response.output_item.done`.
          enum:
            - response.output_item.done
          x-stainless-const: true
        output_index:
          type: integer
          description: |
            The index of the output item that was marked done.
        item:
          $ref: '#/components/schemas/OutputItem'
          x-oaiExpandable: true
          description: |
            The output item that was marked done.
      required:
        - type
        - output_index
        - item
    ResponseRefusalDeltaEvent:
      type: object
      description: Emitted when there is a partial refusal text.
      properties:
        type:
          type: string
          description: |
            The type of the event. Always `response.refusal.delta`.
          enum:
            - response.refusal.delta
          x-stainless-const: true
        item_id:
          type: string
          description: |
            The ID of the output item that the refusal text is added to.
        output_index:
          type: integer
          description: |
            The index of the output item that the refusal text is added to.
        content_index:
          type: integer
          description: |
            The index of the content part that the refusal text is added to.
        delta:
          type: string
          description: |
            The refusal text that is added.
      required:
        - type
        - item_id
        - output_index
        - content_index
        - delta
    ResponseRefusalDoneEvent:
      type: object
      description: Emitted when refusal text is finalized.
      properties:
        type:
          type: string
          description: |
            The type of the event. Always `response.refusal.done`.
          enum:
            - response.refusal.done
          x-stainless-const: true
        item_id:
          type: string
          description: |
            The ID of the output item that the refusal text is finalized.
        output_index:
          type: integer
          description: |
            The index of the output item that the refusal text is finalized.
        content_index:
          type: integer
          description: |
            The index of the content part that the refusal text is finalized.
        refusal:
          type: string
          description: |
            The refusal text that is finalized.
      required:
        - type
        - item_id
        - output_index
        - content_index
        - refusal
    ResponseTextAnnotationDeltaEvent:
      type: object
      description: Emitted when a text annotation is added.
      properties:
        type:
          type: string
          description: >
            The type of the event. Always
            `response.output_text.annotation.added`.
          enum:
            - response.output_text.annotation.added
          x-stainless-const: true
        item_id:
          type: string
          description: |
            The ID of the output item that the text annotation was added to.
        output_index:
          type: integer
          description: |
            The index of the output item that the text annotation was added to.
        content_index:
          type: integer
          description: |
            The index of the content part that the text annotation was added to.
        annotation_index:
          type: integer
          description: |
            The index of the annotation that was added.
        annotation:
          $ref: '#/components/schemas/Annotation'
      required:
        - type
        - item_id
        - output_index
        - content_index
        - annotation_index
        - annotation
    ResponseTextDeltaEvent:
      type: object
      description: Emitted when there is an additional text delta.
      properties:
        type:
          type: string
          description: |
            The type of the event. Always `response.output_text.delta`.
          enum:
            - response.output_text.delta
          x-stainless-const: true
        item_id:
          type: string
          description: |
            The ID of the output item that the text delta was added to.
        output_index:
          type: integer
          description: |
            The index of the output item that the text delta was added to.
        content_index:
          type: integer
          description: |
            The index of the content part that the text delta was added to.
        delta:
          type: string
          description: |
            The text delta that was added.
      required:
        - type
        - item_id
        - output_index
        - content_index
        - delta
    ResponseTextDoneEvent:
      type: object
      description: Emitted when text content is finalized.
      properties:
        type:
          type: string
          description: |
            The type of the event. Always `response.output_text.done`.
          enum:
            - response.output_text.done
          x-stainless-const: true
        item_id:
          type: string
          description: |
            The ID of the output item that the text content is finalized.
        output_index:
          type: integer
          description: |
            The index of the output item that the text content is finalized.
        content_index:
          type: integer
          description: |
            The index of the content part that the text content is finalized.
        text:
          type: string
          description: |
            The text content that is finalized.
      required:
        - type
        - item_id
        - output_index
        - content_index
        - text
    ResponseWebSearchCallCompletedEvent:
      type: object
      description: Emitted when a web search call is completed.
      properties:
        type:
          type: string
          description: |
            The type of the event. Always `response.web_search_call.completed`.
          enum:
            - response.web_search_call.completed
          x-stainless-const: true
        output_index:
          type: integer
          description: >
            The index of the output item that the web search call is associated
            with.
        item_id:
          type: string
          description: |
            Unique ID for the output item associated with the web search call.
      required:
        - type
        - output_index
        - item_id
    ResponseWebSearchCallInProgressEvent:
      type: object
      description: Emitted when a web search call is initiated.
      properties:
        type:
          type: string
          description: >
            The type of the event. Always
            `response.web_search_call.in_progress`.
          enum:
            - response.web_search_call.in_progress
          x-stainless-const: true
        output_index:
          type: integer
          description: >
            The index of the output item that the web search call is associated
            with.
        item_id:
          type: string
          description: |
            Unique ID for the output item associated with the web search call.
      required:
        - type
        - output_index
        - item_id
    ResponseWebSearchCallSearchingEvent:
      type: object
      description: Emitted when a web search call is executing.
      properties:
        type:
          type: string
          description: |
            The type of the event. Always `response.web_search_call.searching`.
          enum:
            - response.web_search_call.searching
          x-stainless-const: true
        output_index:
          type: integer
          description: >
            The index of the output item that the web search call is associated
            with.
        item_id:
          type: string
          description: |
            Unique ID for the output item associated with the web search call.
      required:
        - type
        - output_index
        - item_id
    ModelIdsResponses:
      example: gpt-4o
      anyOf:
        - type: string
          enum:
            - o1-pro
            - o1-pro-2025-03-19
            - computer-use-preview
            - computer-use-preview-2025-03-11
    Reasoning:
      type: object
      description: |
        **o-series models only**

        Configuration options for 
        [reasoning models](https://platform.openai.com/docs/guides/reasoning).
      title: Reasoning
      x-oaiExpandable: true
      properties:
        effort:
          $ref: '#/components/schemas/ReasoningEffort'
        generate_summary:
          type: string
          description: >
            **computer_use_preview only**


            A summary of the reasoning performed by the model. This can be

            useful for debugging and understanding the model's reasoning
            process.

            One of `concise` or `detailed`.
          enum:
            - concise
            - detailed
          nullable: true
    TextResponseFormatConfiguration:
      description: >
        An object specifying the format that the model must output.


        Configuring `{ "type": "json_schema" }` enables Structured Outputs, 

        which ensures the model will match your supplied JSON schema. Learn more
        in the 

        [Structured Outputs guide](/docs/guides/structured-outputs).


        The default format is `{ "type": "text" }` with no additional options.


        **Not recommended for gpt-4o and newer models:**


        Setting to `{ "type": "json_object" }` enables the older JSON mode,
        which

        ensures the message the model generates is valid JSON. Using
        `json_schema`

        is preferred for models that support it.
      oneOf:
        - $ref: '#/components/schemas/ResponseFormatText'
        - $ref: '#/components/schemas/TextResponseFormatJsonSchema'
        - $ref: '#/components/schemas/ResponseFormatJsonObject'
      x-oaiExpandable: true
    Tool:
      oneOf:
        - $ref: '#/components/schemas/FileSearchTool'
        - $ref: '#/components/schemas/FunctionTool'
        - $ref: '#/components/schemas/ComputerTool'
        - $ref: '#/components/schemas/WebSearchTool'
      x-oaiExpandable: true
    ToolChoiceOptions:
      type: string
      title: Tool choice mode
      description: >
        Controls which (if any) tool is called by the model.


        `none` means the model will not call any tool and instead generates a
        message.


        `auto` means the model can pick between generating a message or calling
        one or

        more tools.


        `required` means the model must call one or more tools.
      enum:
        - none
        - auto
        - required
    ToolChoiceTypes:
      type: object
      title: Hosted tool
      description: >
        Indicates that the model should use a built-in tool to generate a
        response.

        [Learn more about built-in tools](/docs/guides/tools).
      properties:
        type:
          type: string
          description: |
            The type of hosted tool the model should to use. Learn more about
            [built-in tools](/docs/guides/tools).

            Allowed values are:
            - `file_search`
            - `web_search_preview`
            - `computer_use_preview`
          enum:
            - file_search
            - web_search_preview
            - computer_use_preview
            - web_search_preview_2025_03_11
      required:
        - type
    ToolChoiceFunction:
      type: object
      title: Function tool
      description: |
        Use this option to force the model to call a specific function.
      properties:
        type:
          type: string
          enum:
            - function
          description: For function calling, the type is always `function`.
          x-stainless-const: true
        name:
          type: string
          description: The name of the function to call.
      required:
        - type
        - name
    EasyInputMessage:
      type: object
      title: Input message
      description: >
        A message input to the model with a role indicating instruction
        following

        hierarchy. Instructions given with the `developer` or `system` role take

        precedence over instructions given with the `user` role. Messages with
        the

        `assistant` role are presumed to have been generated by the model in
        previous

        interactions.
      properties:
        role:
          type: string
          description: >
            The role of the message input. One of `user`, `assistant`, `system`,
            or

            `developer`.
          enum:
            - user
            - assistant
            - system
            - developer
        content:
          description: >
            Text, image, or audio input to the model, used to generate a
            response.

            Can also contain previous assistant responses.
          x-oaiExpandable: true
          oneOf:
            - type: string
              title: Text input
              description: |
                A text input to the model.
            - $ref: '#/components/schemas/InputMessageContentList'
        type:
          type: string
          description: |
            The type of the message input. Always `message`.
          enum:
            - message
          x-stainless-const: true
      required:
        - role
        - content
    Item:
      type: object
      description: |
        Content item used to generate a response.
      oneOf:
        - $ref: '#/components/schemas/InputMessage'
        - $ref: '#/components/schemas/OutputMessage'
        - $ref: '#/components/schemas/FileSearchToolCall'
        - $ref: '#/components/schemas/ComputerToolCall'
        - $ref: '#/components/schemas/ComputerToolCallOutput'
        - $ref: '#/components/schemas/WebSearchToolCall'
        - $ref: '#/components/schemas/FunctionToolCall'
        - $ref: '#/components/schemas/FunctionToolCallOutput'
        - $ref: '#/components/schemas/ReasoningItem'
      x-oaiExpandable: true
      discriminator:
        propertyName: type
    ItemReference:
      type: object
      title: Item reference
      description: |
        An internal identifier for an item to reference.
      properties:
        id:
          type: string
          description: |
            The ID of the item to reference.
        type:
          type: string
          description: |
            The type of item to reference. Always `item_reference`.
          enum:
            - item_reference
          x-stainless-const: true
      required:
        - id
        - type
    Metadata:
      type: object
      description: >
        Set of 16 key-value pairs that can be attached to an object. This can be

        useful for storing additional information about the object in a
        structured

        format, and querying for objects via API or the dashboard. 


        Keys are strings with a maximum length of 64 characters. Values are
        strings

        with a maximum length of 512 characters.
      additionalProperties:
        type: string
      x-oaiTypeLabel: map
      nullable: true
    ResponseErrorCode:
      type: string
      description: |
        The error code for the response.
      enum:
        - server_error
        - rate_limit_exceeded
        - invalid_prompt
        - vector_store_timeout
        - invalid_image
        - invalid_image_format
        - invalid_base64_image
        - invalid_image_url
        - image_too_large
        - image_too_small
        - image_parse_error
        - image_content_policy_violation
        - invalid_image_mode
        - image_file_too_large
        - unsupported_image_media_type
        - empty_image_file
        - failed_to_download_image
        - image_file_not_found
    OutputMessage:
      type: object
      title: Output message
      description: |
        An output message from the model.
      properties:
        id:
          type: string
          description: |
            The unique ID of the output message.
        type:
          type: string
          description: |
            The type of the output message. Always `message`.
          enum:
            - message
          x-stainless-const: true
        role:
          type: string
          description: |
            The role of the output message. Always `assistant`.
          enum:
            - assistant
          x-stainless-const: true
        content:
          type: array
          description: |
            The content of the output message.
          x-oaiExpandable: true
          items:
            $ref: '#/components/schemas/OutputContent'
            x-oaiExpandable: true
        status:
          type: string
          description: >
            The status of the message input. One of `in_progress`, `completed`,
            or

            `incomplete`. Populated when input items are returned via API.
          enum:
            - in_progress
            - completed
            - incomplete
      required:
        - id
        - type
        - role
        - content
        - status
    FileSearchToolCall:
      type: object
      title: File search tool call
      description: >
        The results of a file search tool call. See the 

        [file search guide](/docs/guides/tools-file-search) for more
        information.
      properties:
        id:
          type: string
          description: |
            The unique ID of the file search tool call.
        type:
          type: string
          enum:
            - file_search_call
          description: |
            The type of the file search tool call. Always `file_search_call`.
          x-stainless-const: true
        status:
          type: string
          description: |
            The status of the file search tool call. One of `in_progress`, 
            `searching`, `incomplete` or `failed`,
          enum:
            - in_progress
            - searching
            - completed
            - incomplete
            - failed
        queries:
          type: array
          items:
            type: string
          description: |
            The queries used to search for files.
        results:
          type: array
          description: |
            The results of the file search tool call.
          items:
            type: object
            properties:
              file_id:
                type: string
                description: |
                  The unique ID of the file.
              text:
                type: string
                description: |
                  The text that was retrieved from the file.
              filename:
                type: string
                description: |
                  The name of the file.
              attributes:
                $ref: '#/components/schemas/VectorStoreFileAttributes'
              score:
                type: number
                format: float
                description: |
                  The relevance score of the file - a value between 0 and 1.
          nullable: true
      required:
        - id
        - type
        - status
        - queries
    FunctionToolCall:
      type: object
      title: Function tool call
      description: >
        A tool call to run a function. See the 

        [function calling guide](/docs/guides/function-calling) for more
        information.
      properties:
        id:
          type: string
          description: |
            The unique ID of the function tool call.
        type:
          type: string
          enum:
            - function_call
          description: |
            The type of the function tool call. Always `function_call`.
          x-stainless-const: true
        call_id:
          type: string
          description: |
            The unique ID of the function tool call generated by the model.
        name:
          type: string
          description: |
            The name of the function to run.
        arguments:
          type: string
          description: |
            A JSON string of the arguments to pass to the function.
        status:
          type: string
          description: |
            The status of the item. One of `in_progress`, `completed`, or
            `incomplete`. Populated when items are returned via API.
          enum:
            - in_progress
            - completed
            - incomplete
      required:
        - type
        - call_id
        - name
        - arguments
    WebSearchToolCall:
      type: object
      title: Web search tool call
      description: |
        The results of a web search tool call. See the 
        [web search guide](/docs/guides/tools-web-search) for more information.
      properties:
        id:
          type: string
          description: |
            The unique ID of the web search tool call.
        type:
          type: string
          enum:
            - web_search_call
          description: |
            The type of the web search tool call. Always `web_search_call`.
          x-stainless-const: true
        status:
          type: string
          description: |
            The status of the web search tool call.
          enum:
            - in_progress
            - searching
            - completed
            - failed
      required:
        - id
        - type
        - status
    ComputerToolCall:
      type: object
      title: Computer tool call
      description: >
        A tool call to a computer use tool. See the 

        [computer use guide](/docs/guides/tools-computer-use) for more
        information.
      properties:
        type:
          type: string
          description: The type of the computer call. Always `computer_call`.
          enum:
            - computer_call
          default: computer_call
        id:
          type: string
          description: The unique ID of the computer call.
        call_id:
          type: string
          description: |
            An identifier used when responding to the tool call with output.
        action:
          $ref: '#/components/schemas/ComputerAction'
          x-oaiExpandable: true
        pending_safety_checks:
          type: array
          x-oaiExpandable: true
          items:
            $ref: '#/components/schemas/ComputerToolCallSafetyCheck'
          description: |
            The pending safety checks for the computer call.
        status:
          type: string
          description: |
            The status of the item. One of `in_progress`, `completed`, or
            `incomplete`. Populated when items are returned via API.
          enum:
            - in_progress
            - completed
            - incomplete
      required:
        - type
        - id
        - action
        - call_id
        - pending_safety_checks
        - status
    ReasoningItem:
      type: object
      description: >
        A description of the chain of thought used by a reasoning model while
        generating

        a response.
      title: Reasoning
      x-oaiExpandable: true
      properties:
        type:
          type: string
          description: |
            The type of the object. Always `reasoning`.
          enum:
            - reasoning
          x-stainless-const: true
        id:
          type: string
          description: |
            The unique identifier of the reasoning content.
        summary:
          type: array
          description: |
            Reasoning text contents.
          items:
            type: object
            properties:
              type:
                type: string
                description: |
                  The type of the object. Always `summary_text`.
                enum:
                  - summary_text
                x-stainless-const: true
              text:
                type: string
                description: >
                  A short summary of the reasoning used by the model when
                  generating

                  the response.
            required:
              - type
              - text
        status:
          type: string
          description: |
            The status of the item. One of `in_progress`, `completed`, or
            `incomplete`. Populated when items are returned via API.
          enum:
            - in_progress
            - completed
            - incomplete
      required:
        - id
        - summary
        - type
    CodeInterpreterToolCall:
      type: object
      title: Code interpreter tool call
      description: |
        A tool call to run code.
      properties:
        id:
          type: string
          description: |
            The unique ID of the code interpreter tool call.
        type:
          type: string
          enum:
            - code_interpreter_call
          description: >
            The type of the code interpreter tool call. Always
            `code_interpreter_call`.
          x-stainless-const: true
        code:
          type: string
          description: |
            The code to run.
        status:
          type: string
          enum:
            - in_progress
            - interpreting
            - completed
          description: |
            The status of the code interpreter tool call.
        results:
          type: array
          items:
            $ref: '#/components/schemas/CodeInterpreterToolOutput'
            x-oaiExpandable: true
          description: |
            The results of the code interpreter tool call.
      required:
        - id
        - type
        - code
        - status
        - results
    OutputContent:
      oneOf:
        - $ref: '#/components/schemas/OutputText'
        - $ref: '#/components/schemas/Refusal'
    Annotation:
      oneOf:
        - $ref: '#/components/schemas/FileCitation'
        - $ref: '#/components/schemas/UrlCitation'
        - $ref: '#/components/schemas/FilePath'
    ReasoningEffort:
      type: string
      enum:
        - low
        - medium
        - high
      default: medium
      nullable: true
      description: |
        **o-series models only** 

        Constrains effort on reasoning for 
        [reasoning models](https://platform.openai.com/docs/guides/reasoning).
        Currently supported values are `low`, `medium`, and `high`. Reducing
        reasoning effort can result in faster responses and fewer tokens used
        on reasoning in a response.
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
    TextResponseFormatJsonSchema:
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
        description:
          type: string
          description: >
            A description of what the response format is for, used by the model
            to

            determine how to respond in the format.
        name:
          type: string
          description: |
            The name of the response format. Must be a-z, A-Z, 0-9, or contain
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
        - type
        - schema
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
    FileSearchTool:
      type: object
      title: File search
      description: |
        A tool that searches for relevant content from uploaded files.
        Learn more about the [file search tool](/docs/guides/tools-file-search).
      properties:
        type:
          type: string
          enum:
            - file_search
          description: |
            The type of the file search tool. Always `file_search`.
          x-stainless-const: true
        vector_store_ids:
          type: array
          items:
            type: string
          description: |
            The IDs of the vector stores to search.
        max_num_results:
          type: integer
          description: >
            The maximum number of results to return. This number should be
            between 1 

            and 50 inclusive.
        filters:
          description: A filter to apply based on file attributes.
          oneOf:
            - $ref: '#/components/schemas/ComparisonFilter'
            - $ref: '#/components/schemas/CompoundFilter'
          x-oaiExpandable: true
        ranking_options:
          description: Ranking options for search.
          type: object
          additionalProperties: false
          properties:
            ranker:
              type: string
              description: The ranker to use for the file search.
              enum:
                - auto
                - default-2024-11-15
              default: auto
            score_threshold:
              type: number
              description: >
                The score threshold for the file search, a number between 0 and
                1.

                Numbers closer to 1 will attempt to return only the most
                relevant

                results, but may return fewer results.
              minimum: 0
              maximum: 1
              default: 0
      required:
        - type
        - vector_store_ids
    FunctionTool:
      type: object
      title: Function
      description: >
        Defines a function in your own code the model can choose to call. Learn
        more

        about [function calling](/docs/guides/function-calling).
      properties:
        type:
          type: string
          enum:
            - function
          description: |
            The type of the function tool. Always `function`.
          x-stainless-const: true
        name:
          type: string
          description: |
            The name of the function to call.
        description:
          type: string
          nullable: true
          description: >
            A description of the function. Used by the model to determine
            whether

            or not to call the function.
        parameters:
          type: object
          description: |
            A JSON schema object describing the parameters of the function.
          additionalProperties: true
        strict:
          type: boolean
          description: |
            Whether to enforce strict parameter validation. Default `true`.
      required:
        - type
        - name
        - parameters
        - strict
    ComputerTool:
      type: object
      title: Computer use
      description: |
        A tool that controls a virtual computer. Learn more about the 
        [computer tool](/docs/guides/tools-computer-use).
      properties:
        type:
          type: string
          enum:
            - computer_use_preview
          description: |
            The type of the computer use tool. Always `computer_use_preview`.
          x-stainless-const: true
        display_width:
          type: number
          description: |
            The width of the computer display.
        display_height:
          type: number
          description: |
            The height of the computer display.
        environment:
          type: string
          description: |
            The type of computer environment to control.
          enum:
            - mac
            - windows
            - ubuntu
            - browser
      required:
        - type
        - display_width
        - display_height
        - environment
    WebSearchTool:
      type: object
      title: Web search
      description: |
        This tool searches the web for relevant results to use in a response.
        Learn more about the [web search tool](/docs/guides/tools-web-search).
      properties:
        type:
          type: string
          enum:
            - web_search_preview
            - web_search_preview_2025_03_11
          description: |
            The type of the web search tool. One of:
            - `web_search_preview`
            - `web_search_preview_2025_03_11`
        user_location:
          allOf:
            - $ref: '#/components/schemas/WebSearchLocation'
            - type: object
              properties:
                type:
                  type: string
                  description: |
                    The type of location approximation. Always `approximate`.
                  enum:
                    - approximate
                  x-stainless-const: true
              required:
                - type
          nullable: true
        search_context_size:
          $ref: '#/components/schemas/WebSearchContextSize'
      required:
        - type
    InputMessageContentList:
      type: array
      title: Input item content list
      description: >
        A list of one or many input items to the model, containing different
        content 

        types.
      x-oaiExpandable: true
      items:
        $ref: '#/components/schemas/InputContent'
        x-oaiExpandable: true
    InputMessage:
      type: object
      title: Input message
      description: >
        A message input to the model with a role indicating instruction
        following

        hierarchy. Instructions given with the `developer` or `system` role take

        precedence over instructions given with the `user` role.
      properties:
        type:
          type: string
          description: |
            The type of the message input. Always set to `message`.
          enum:
            - message
          x-stainless-const: true
        role:
          type: string
          description: >
            The role of the message input. One of `user`, `system`, or
            `developer`.
          enum:
            - user
            - system
            - developer
        status:
          type: string
          description: |
            The status of item. One of `in_progress`, `completed`, or
            `incomplete`. Populated when items are returned via API.
          enum:
            - in_progress
            - completed
            - incomplete
        content:
          $ref: '#/components/schemas/InputMessageContentList'
      required:
        - role
        - content
    ComputerToolCallOutput:
      type: object
      title: Computer tool call output
      description: |
        The output of a computer tool call.
      properties:
        type:
          type: string
          description: >
            The type of the computer tool call output. Always
            `computer_call_output`.
          enum:
            - computer_call_output
          default: computer_call_output
          x-stainless-const: true
        id:
          type: string
          description: |
            The ID of the computer tool call output.
        call_id:
          type: string
          description: |
            The ID of the computer tool call that produced the output.
        acknowledged_safety_checks:
          type: array
          x-oaiExpandable: true
          description: >
            The safety checks reported by the API that have been acknowledged by
            the 

            developer.
          items:
            $ref: '#/components/schemas/ComputerToolCallSafetyCheck'
        output:
          $ref: '#/components/schemas/ComputerScreenshotImage'
        status:
          type: string
          description: >
            The status of the message input. One of `in_progress`, `completed`,
            or

            `incomplete`. Populated when input items are returned via API.
          enum:
            - in_progress
            - completed
            - incomplete
      required:
        - type
        - call_id
        - output
    FunctionToolCallOutput:
      type: object
      title: Function tool call output
      description: |
        The output of a function tool call.
      properties:
        id:
          type: string
          description: >
            The unique ID of the function tool call output. Populated when this
            item

            is returned via API.
        type:
          type: string
          enum:
            - function_call_output
          description: >
            The type of the function tool call output. Always
            `function_call_output`.
          x-stainless-const: true
        call_id:
          type: string
          description: |
            The unique ID of the function tool call generated by the model.
        output:
          type: string
          description: |
            A JSON string of the output of the function tool call.
        status:
          type: string
          description: |
            The status of the item. One of `in_progress`, `completed`, or
            `incomplete`. Populated when items are returned via API.
          enum:
            - in_progress
            - completed
            - incomplete
      required:
        - type
        - call_id
        - output
    VectorStoreFileAttributes:
      type: object
      description: >
        Set of 16 key-value pairs that can be attached to an object. This can
        be 

        useful for storing additional information about the object in a
        structured 

        format, and querying for objects via API or the dashboard. Keys are
        strings 

        with a maximum length of 64 characters. Values are strings with a
        maximum 

        length of 512 characters, booleans, or numbers.
      maxProperties: 16
      additionalProperties:
        oneOf:
          - type: string
            maxLength: 512
          - type: number
          - type: boolean
      x-oaiTypeLabel: map
      nullable: true
    ComputerAction:
      oneOf:
        - $ref: '#/components/schemas/Click'
        - $ref: '#/components/schemas/DoubleClick'
        - $ref: '#/components/schemas/Drag'
        - $ref: '#/components/schemas/KeyPress'
        - $ref: '#/components/schemas/Move'
        - $ref: '#/components/schemas/Screenshot'
        - $ref: '#/components/schemas/Scroll'
        - $ref: '#/components/schemas/Type'
        - $ref: '#/components/schemas/Wait'
    ComputerToolCallSafetyCheck:
      type: object
      description: |
        A pending safety check for the computer call.
      properties:
        id:
          type: string
          description: The ID of the pending safety check.
        code:
          type: string
          description: The type of the pending safety check.
        message:
          type: string
          description: Details about the pending safety check.
      required:
        - id
        - code
        - message
    CodeInterpreterToolOutput:
      oneOf:
        - $ref: '#/components/schemas/CodeInterpreterTextOutput'
        - $ref: '#/components/schemas/CodeInterpreterFileOutput'
    OutputText:
      type: object
      title: Output text
      description: |
        A text output from the model.
      properties:
        type:
          type: string
          description: |
            The type of the output text. Always `output_text`.
          enum:
            - output_text
          x-stainless-const: true
        text:
          type: string
          description: |
            The text output from the model.
        annotations:
          type: array
          description: |
            The annotations of the text output.
          items:
            $ref: '#/components/schemas/Annotation'
            x-oaiExpandable: true
      required:
        - type
        - text
        - annotations
    Refusal:
      type: object
      title: Refusal
      description: |
        A refusal from the model.
      properties:
        type:
          type: string
          description: |
            The type of the refusal. Always `refusal`.
          enum:
            - refusal
          x-stainless-const: true
        refusal:
          type: string
          description: |
            The refusal explanationfrom the model.
      required:
        - type
        - refusal
    FileCitation:
      type: object
      title: File citation
      description: |
        A citation to a file.
      properties:
        type:
          type: string
          description: |
            The type of the file citation. Always `file_citation`.
          enum:
            - file_citation
          x-stainless-const: true
        index:
          type: integer
          description: |
            The index of the file in the list of files.
        file_id:
          type: string
          description: |
            The ID of the file.
      required:
        - type
        - index
        - file_id
    UrlCitation:
      type: object
      title: URL citation
      description: |
        A citation for a web resource used to generate a model response.
      properties:
        url:
          type: string
          description: |
            The URL of the web resource.
        title:
          type: string
          description: |
            The title of the web resource.
        type:
          type: string
          description: |
            The type of the URL citation. Always `url_citation`.
          enum:
            - url_citation
          x-stainless-const: true
        start_index:
          type: integer
          description: |
            The index of the first character of the URL citation in the message.
        end_index:
          type: integer
          description: |
            The index of the last character of the URL citation in the message.
      required:
        - url
        - title
        - type
        - start_index
        - end_index
    FilePath:
      type: object
      title: File path
      description: |
        A path to a file.
      properties:
        type:
          type: string
          description: |
            The type of the file path. Always `file_path`.
          enum:
            - file_path
          x-stainless-const: true
        file_id:
          type: string
          description: |
            The ID of the file.
        index:
          type: integer
          description: |
            The index of the file in the list of files.
      required:
        - type
        - file_id
        - index
    ResponseFormatJsonSchemaSchema:
      type: object
      title: JSON schema
      description: |
        The schema for the response format, described as a JSON Schema object.
        Learn how to build JSON schemas [here](https://json-schema.org/).
      additionalProperties: true
    ComparisonFilter:
      type: object
      additionalProperties: false
      title: Comparison Filter
      description: >
        A filter used to compare a specified attribute key to a given value
        using a defined comparison operation.
      properties:
        type:
          type: string
          default: eq
          enum:
            - eq
            - ne
            - gt
            - gte
            - lt
            - lte
          description: >
            Specifies the comparison operator: `eq`, `ne`, `gt`, `gte`, `lt`,
            `lte`.

            - `eq`: equals

            - `ne`: not equal

            - `gt`: greater than

            - `gte`: greater than or equal

            - `lt`: less than

            - `lte`: less than or equal
        key:
          type: string
          description: The key to compare against the value.
        value:
          oneOf:
            - type: string
            - type: number
            - type: boolean
          description: >-
            The value to compare against the attribute key; supports string,
            number, or boolean types.
      required:
        - type
        - key
        - value
    CompoundFilter:
      type: object
      additionalProperties: false
      title: Compound Filter
      description: Combine multiple filters using `and` or `or`.
      properties:
        type:
          type: string
          description: 'Type of operation: `and` or `or`.'
          enum:
            - and
            - or
        filters:
          type: array
          description: >-
            Array of filters to combine. Items can be `ComparisonFilter` or
            `CompoundFilter`.
          items:
            oneOf:
              - $ref: '#/components/schemas/ComparisonFilter'
              - type: object
                additionalProperties: true
      required:
        - type
        - filters
    WebSearchLocation:
      type: object
      title: Web search location
      description: Approximate location parameters for the search.
      properties:
        country:
          type: string
          description: >
            The two-letter 

            [ISO country code](https://en.wikipedia.org/wiki/ISO_3166-1) of the
            user,

            e.g. `US`.
        region:
          type: string
          description: |
            Free text input for the region of the user, e.g. `California`.
        city:
          type: string
          description: |
            Free text input for the city of the user, e.g. `San Francisco`.
        timezone:
          type: string
          description: >
            The [IANA
            timezone](https://timeapi.io/documentation/iana-timezones) 

            of the user, e.g. `America/Los_Angeles`.
    WebSearchContextSize:
      type: string
      description: >
        High level guidance for the amount of context window space to use for
        the 

        search. One of `low`, `medium`, or `high`. `medium` is the default.
      enum:
        - low
        - medium
        - high
      default: medium
    InputContent:
      oneOf:
        - $ref: '#/components/schemas/InputText'
        - $ref: '#/components/schemas/InputImage'
        - $ref: '#/components/schemas/InputFile'
      x-oaiExpandable: true
    ComputerScreenshotImage:
      type: object
      description: |
        A computer screenshot image used with the computer use tool.
      properties:
        type:
          type: string
          enum:
            - computer_screenshot
          default: computer_screenshot
          description: >
            Specifies the event type. For a computer screenshot, this property
            is 

            always set to `computer_screenshot`.
          x-stainless-const: true
        image_url:
          type: string
          description: The URL of the screenshot image.
        file_id:
          type: string
          description: The identifier of an uploaded file that contains the screenshot.
      required:
        - type
    Click:
      type: object
      title: Click
      description: |
        A click action.
      properties:
        type:
          type: string
          enum:
            - click
          default: click
          description: |
            Specifies the event type. For a click action, this property is 
            always set to `click`.
          x-stainless-const: true
        button:
          type: string
          enum:
            - left
            - right
            - wheel
            - back
            - forward
          description: >
            Indicates which mouse button was pressed during the click. One of
            `left`, `right`, `wheel`, `back`, or `forward`.
        x:
          type: integer
          description: |
            The x-coordinate where the click occurred.
        'y':
          type: integer
          description: |
            The y-coordinate where the click occurred.
      required:
        - type
        - button
        - x
        - 'y'
    DoubleClick:
      type: object
      title: DoubleClick
      description: |
        A double click action.
      properties:
        type:
          type: string
          enum:
            - double_click
          default: double_click
          description: >
            Specifies the event type. For a double click action, this property
            is 

            always set to `double_click`.
          x-stainless-const: true
        x:
          type: integer
          description: |
            The x-coordinate where the double click occurred.
        'y':
          type: integer
          description: |
            The y-coordinate where the double click occurred.
      required:
        - type
        - x
        - 'y'
    Drag:
      type: object
      title: Drag
      description: |
        A drag action.
      properties:
        type:
          type: string
          enum:
            - drag
          default: drag
          description: |
            Specifies the event type. For a drag action, this property is 
            always set to `drag`.
          x-stainless-const: true
        path:
          type: array
          description: >
            An array of coordinates representing the path of the drag action.
            Coordinates will appear as an array

            of objects, eg

            ```

            [
              { x: 100, y: 200 },
              { x: 200, y: 300 }
            ]

            ```
          x-oaiExpandable: true
          items:
            $ref: '#/components/schemas/Coordinate'
            title: Drag path coordinates
            x-oaiExpandable: true
            description: |
              A series of x/y coordinate pairs in the drag path.
      required:
        - type
        - path
    KeyPress:
      type: object
      title: KeyPress
      description: |
        A collection of keypresses the model would like to perform.
      properties:
        type:
          type: string
          enum:
            - keypress
          default: keypress
          description: |
            Specifies the event type. For a keypress action, this property is 
            always set to `keypress`.
          x-stainless-const: true
        keys:
          type: array
          items:
            type: string
            description: |
              One of the keys the model is requesting to be pressed.
          description: >
            The combination of keys the model is requesting to be pressed. This
            is an

            array of strings, each representing a key.
      required:
        - type
        - keys
    Move:
      type: object
      title: Move
      description: |
        A mouse move action.
      properties:
        type:
          type: string
          enum:
            - move
          default: move
          description: |
            Specifies the event type. For a move action, this property is 
            always set to `move`.
          x-stainless-const: true
        x:
          type: integer
          description: |
            The x-coordinate to move to.
        'y':
          type: integer
          description: |
            The y-coordinate to move to.
      required:
        - type
        - x
        - 'y'
    Screenshot:
      type: object
      title: Screenshot
      description: |
        A screenshot action.
      properties:
        type:
          type: string
          enum:
            - screenshot
          default: screenshot
          description: |
            Specifies the event type. For a screenshot action, this property is 
            always set to `screenshot`.
          x-stainless-const: true
      required:
        - type
    Scroll:
      type: object
      title: Scroll
      description: |
        A scroll action.
      properties:
        type:
          type: string
          enum:
            - scroll
          default: scroll
          description: |
            Specifies the event type. For a scroll action, this property is 
            always set to `scroll`.
          x-stainless-const: true
        x:
          type: integer
          description: |
            The x-coordinate where the scroll occurred.
        'y':
          type: integer
          description: |
            The y-coordinate where the scroll occurred.
        scroll_x:
          type: integer
          description: |
            The horizontal scroll distance.
        scroll_y:
          type: integer
          description: |
            The vertical scroll distance.
      required:
        - type
        - x
        - 'y'
        - scroll_x
        - scroll_y
    Type:
      type: object
      title: Type
      description: |
        An action to type in text.
      properties:
        type:
          type: string
          enum:
            - type
          default: type
          description: |
            Specifies the event type. For a type action, this property is 
            always set to `type`.
          x-stainless-const: true
        text:
          type: string
          description: |
            The text to type.
      required:
        - type
        - text
    Wait:
      type: object
      title: Wait
      description: |
        A wait action.
      properties:
        type:
          type: string
          enum:
            - wait
          default: wait
          description: |
            Specifies the event type. For a wait action, this property is 
            always set to `wait`.
          x-stainless-const: true
      required:
        - type
    CodeInterpreterTextOutput:
      type: object
      title: Code interpreter text output
      description: |
        The output of a code interpreter tool call that is text.
      properties:
        type:
          type: string
          enum:
            - logs
          description: |
            The type of the code interpreter text output. Always `logs`.
          x-stainless-const: true
        logs:
          type: string
          description: |
            The logs of the code interpreter tool call.
      required:
        - type
        - logs
    CodeInterpreterFileOutput:
      type: object
      title: Code interpreter file output
      description: |
        The output of a code interpreter tool call that is a file.
      properties:
        type:
          type: string
          enum:
            - files
          description: |
            The type of the code interpreter file output. Always `files`.
          x-stainless-const: true
        files:
          type: array
          items:
            type: object
            properties:
              mime_type:
                type: string
                description: |
                  The MIME type of the file.
              file_id:
                type: string
                description: |
                  The ID of the file.
            required:
              - mime_type
              - file_id
      required:
        - type
        - files
    InputText:
      type: object
      title: Text input
      description: |
        A text input to the model.
      properties:
        type:
          type: string
          description: |
            The type of the input item. Always `input_text`.
          enum:
            - input_text
          x-stainless-const: true
        text:
          type: string
          description: |
            The text input to the model.
      required:
        - type
        - text
    InputImage:
      type: object
      title: Image input
      description: >
        An image input to the model. Learn about [image
        inputs](/docs/guides/vision).
      properties:
        type:
          type: string
          description: |
            The type of the input item. Always `input_image`.
          enum:
            - input_image
          x-stainless-const: true
        image_url:
          type: string
          description: >
            The URL of the image to be sent to the model. A fully qualified URL
            or

            base64 encoded image in a data URL.
          nullable: true
        file_id:
          type: string
          description: |
            The ID of the file to be sent to the model.
          nullable: true
        detail:
          type: string
          description: >
            The detail level of the image to be sent to the model. One of
            `high`,

            `low`, or `auto`. Defaults to `auto`.
          enum:
            - high
            - low
            - auto
          default: auto
      required:
        - type
        - detail
    InputFile:
      type: object
      title: File input
      description: |
        A file input to the model.
      properties:
        type:
          type: string
          description: |
            The type of the input item. Always `input_file`.
          enum:
            - input_file
          x-stainless-const: true
        file_id:
          type: string
          description: |
            The ID of the file to be sent to the model.
        filename:
          type: string
          description: |
            The name of the file to be sent to the model.
        file_data:
          type: string
          description: |
            The content of the file to be sent to the model.
      required:
        - type
    Coordinate:
      type: object
      title: Coordinate
      description: |
        An x/y coordinate pair, e.g. `{ x: 100, y: 200 }`.
      properties:
        x:
          type: integer
          description: |
            The x-coordinate.
        'y':
          type: integer
          description: |
            The y-coordinate.
      required:
        - x
        - 'y'
  securitySchemes:
    Portkey-Key:
      type: apiKey
      in: header
      name: x-portkey-api-key

````