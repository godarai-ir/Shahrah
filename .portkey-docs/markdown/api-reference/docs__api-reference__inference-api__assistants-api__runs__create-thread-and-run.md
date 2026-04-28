> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Create thread and Run



## OpenAPI

````yaml post /threads/runs
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
  /threads/runs:
    servers:
      - url: https://api.portkey.ai/v1
        description: Portkey API Public Endpoint
      - url: SELF_HOSTED_GATEWAY_URL
        description: Self-Hosted Gateway URL
    post:
      tags:
        - Assistants
      summary: Create a thread and run it in one request.
      operationId: createThreadAndRun
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateThreadAndRunRequest'
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/RunObject'
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
          source: |
            curl https://api.portkey.ai/v1/threads/runs \
              -H "x-portkey-api-key: $PORTKEY_API_KEY" \
              -H "x-portkey-virtual-key: $PORTKEY_PROVIDER_VIRTUAL_KEY" \
              -H "Content-Type: application/json" \
              -H "OpenAI-Beta: assistants=v2" \
              -d '{
                  "assistant_id": "asst_abc123",
                  "thread": {
                    "messages": [
                      {"role": "user", "content": "Explain deep learning to a 5 year old."}
                    ]
                  }
                }'
        - lang: python
          source: |
            from portkey_ai import Portkey

              client = Portkey(
                api_key = "PORTKEY_API_KEY",
                virtual_key = "PROVIDER_VIRTUAL_KEY"
              )

              run = client.beta.threads.create_and_run(
                assistant_id="asst_abc123",
                thread={
                  "messages": [
                    {"role": "user", "content": "Explain deep learning to a 5 year old."}
                  ]
                }
              )

              print(run)
        - lang: javascript
          source: |
            import Portkey from 'portkey-ai';

            const client = new Portkey({
              apiKey: 'PORTKEY_API_KEY',
              virtualKey: 'PROVIDER_VIRTUAL_KEY'
            });

            async function main() {
              const run = await client.beta.threads.createAndRun({
                assistant_id: "asst_abc123",
                thread: {
                  messages: [
                    { role: "user", content: "Explain deep learning to a 5 year old." },
                  ],
                },
              });

              console.log(run);
            }

            main();
          response: |
            {
              "id": "run_abc123",
              "object": "thread.run",
              "created_at": 1699076792,
              "assistant_id": "asst_abc123",
              "thread_id": "thread_abc123",
              "status": "queued",
              "started_at": null,
              "expires_at": 1699077392,
              "cancelled_at": null,
              "failed_at": null,
              "completed_at": null,
              "required_action": null,
              "last_error": null,
              "model": "gpt-4-turbo",
              "instructions": "You are a helpful assistant.",
              "tools": [],
              "tool_resources": {},
              "metadata": {},
              "temperature": 1.0,
              "top_p": 1.0,
              "max_completion_tokens": null,
              "max_prompt_tokens": null,
              "truncation_strategy": {
                "type": "auto",
                "last_messages": null
              },
              "incomplete_details": null,
              "usage": null,
              "response_format": "auto",
              "tool_choice": "auto",
              "parallel_tool_calls": true
            }
components:
  schemas:
    CreateThreadAndRunRequest:
      type: object
      additionalProperties: false
      properties:
        assistant_id:
          description: >-
            The ID of the
            [assistant](https://platform.openai.com/docs/api-reference/assistants)
            to use to execute this run.
          type: string
        thread:
          $ref: '#/components/schemas/CreateThreadRequest'
          description: If no thread is provided, an empty thread will be created.
        model:
          description: >-
            The ID of the
            [Model](https://platform.openai.com/docs/api-reference/models) to be
            used to execute this run. If a value is provided here, it will
            override the model associated with the assistant. If not, the model
            associated with the assistant will be used.
          example: gpt-4-turbo
          anyOf:
            - type: string
            - type: string
              enum:
                - gpt-4o
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
                - gpt-3.5-turbo-0613
                - gpt-3.5-turbo-1106
                - gpt-3.5-turbo-0125
                - gpt-3.5-turbo-16k-0613
          x-oaiTypeLabel: string
          nullable: true
        instructions:
          description: >-
            Override the default system message of the assistant. This is useful
            for modifying the behavior on a per-run basis.
          type: string
          nullable: true
        tools:
          description: >-
            Override the tools the assistant can use for this run. This is
            useful for modifying the behavior on a per-run basis.
          nullable: true
          type: array
          maxItems: 20
          items:
            oneOf:
              - $ref: '#/components/schemas/AssistantToolsCode'
              - $ref: '#/components/schemas/AssistantToolsFileSearch'
              - $ref: '#/components/schemas/AssistantToolsFunction'
        tool_resources:
          type: object
          description: >
            A set of resources that are used by the assistant's tools. The
            resources are specific to the type of tool. For example, the
            `code_interpreter` tool requires a list of file IDs, while the
            `file_search` tool requires a list of vector store IDs.
          properties:
            code_interpreter:
              type: object
              properties:
                file_ids:
                  type: array
                  description: >
                    A list of
                    [file](https://platform.openai.com/docs/api-reference/files)
                    IDs made available to the `code_interpreter` tool. There can
                    be a maximum of 20 files associated with the tool.
                  default: []
                  maxItems: 20
                  items:
                    type: string
            file_search:
              type: object
              properties:
                vector_store_ids:
                  type: array
                  description: >
                    The ID of the [vector
                    store](https://platform.openai.com/docs/api-reference/vector-stores/object)
                    attached to this assistant. There can be a maximum of 1
                    vector store attached to the assistant.
                  maxItems: 1
                  items:
                    type: string
          nullable: true
        metadata:
          description: >
            Set of 16 key-value pairs that can be attached to an object. This
            can be useful for storing additional information about the object in
            a structured format. Keys can be a maximum of 64 characters long and
            values can be a maxium of 512 characters long.
          type: object
          x-oaiTypeLabel: map
          nullable: true
        temperature:
          description: >
            What sampling temperature to use, between 0 and 2. Higher values
            like 0.8 will make the output more random, while lower values like
            0.2 will make it more focused and deterministic.
          type: number
          minimum: 0
          maximum: 2
          default: 1
          example: 1
          nullable: true
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


            We generally recommend altering this or temperature but not both.
        stream:
          type: boolean
          nullable: true
          description: >
            If `true`, returns a stream of events that happen during the Run as
            server-sent events, terminating when the Run enters a terminal state
            with a `data: [DONE]` message.
        max_prompt_tokens:
          type: integer
          nullable: true
          description: >
            The maximum number of prompt tokens that may be used over the course
            of the run. The run will make a best effort to use only the number
            of prompt tokens specified, across multiple turns of the run. If the
            run exceeds the number of prompt tokens specified, the run will end
            with status `incomplete`. See `incomplete_details` for more info.
          minimum: 256
        max_completion_tokens:
          type: integer
          nullable: true
          description: >
            The maximum number of completion tokens that may be used over the
            course of the run. The run will make a best effort to use only the
            number of completion tokens specified, across multiple turns of the
            run. If the run exceeds the number of completion tokens specified,
            the run will end with status `incomplete`. See `incomplete_details`
            for more info.
          minimum: 256
        truncation_strategy:
          $ref: '#/components/schemas/TruncationObject'
          nullable: true
        tool_choice:
          $ref: '#/components/schemas/AssistantsApiToolChoiceOption'
          nullable: true
        parallel_tool_calls:
          $ref: '#/components/schemas/ParallelToolCalls'
        response_format:
          $ref: '#/components/schemas/AssistantsApiResponseFormatOption'
          nullable: true
      required:
        - thread_id
        - assistant_id
    RunObject:
      type: object
      title: A run on a thread
      description: >-
        Represents an execution run on a
        [thread](https://platform.openai.com/docs/api-reference/threads).
      properties:
        id:
          description: The identifier, which can be referenced in API endpoints.
          type: string
        object:
          description: The object type, which is always `thread.run`.
          type: string
          enum:
            - thread.run
        created_at:
          description: The Unix timestamp (in seconds) for when the run was created.
          type: integer
        thread_id:
          description: >-
            The ID of the
            [thread](https://platform.openai.com/docs/api-reference/threads)
            that was executed on as a part of this run.
          type: string
        assistant_id:
          description: >-
            The ID of the
            [assistant](https://platform.openai.com/docs/api-reference/assistants)
            used for execution of this run.
          type: string
        status:
          description: >-
            The status of the run, which can be either `queued`, `in_progress`,
            `requires_action`, `cancelling`, `cancelled`, `failed`, `completed`,
            `incomplete`, or `expired`.
          type: string
          enum:
            - queued
            - in_progress
            - requires_action
            - cancelling
            - cancelled
            - failed
            - completed
            - incomplete
            - expired
        required_action:
          type: object
          description: >-
            Details on the action required to continue the run. Will be `null`
            if no action is required.
          nullable: true
          properties:
            type:
              description: For now, this is always `submit_tool_outputs`.
              type: string
              enum:
                - submit_tool_outputs
            submit_tool_outputs:
              type: object
              description: Details on the tool outputs needed for this run to continue.
              properties:
                tool_calls:
                  type: array
                  description: A list of the relevant tool calls.
                  items:
                    $ref: '#/components/schemas/RunToolCallObject'
              required:
                - tool_calls
          required:
            - type
            - submit_tool_outputs
        last_error:
          type: object
          description: >-
            The last error associated with this run. Will be `null` if there are
            no errors.
          nullable: true
          properties:
            code:
              type: string
              description: >-
                One of `server_error`, `rate_limit_exceeded`, or
                `invalid_prompt`.
              enum:
                - server_error
                - rate_limit_exceeded
                - invalid_prompt
            message:
              type: string
              description: A human-readable description of the error.
          required:
            - code
            - message
        expires_at:
          description: The Unix timestamp (in seconds) for when the run will expire.
          type: integer
          nullable: true
        started_at:
          description: The Unix timestamp (in seconds) for when the run was started.
          type: integer
          nullable: true
        cancelled_at:
          description: The Unix timestamp (in seconds) for when the run was cancelled.
          type: integer
          nullable: true
        failed_at:
          description: The Unix timestamp (in seconds) for when the run failed.
          type: integer
          nullable: true
        completed_at:
          description: The Unix timestamp (in seconds) for when the run was completed.
          type: integer
          nullable: true
        incomplete_details:
          description: >-
            Details on why the run is incomplete. Will be `null` if the run is
            not incomplete.
          type: object
          nullable: true
          properties:
            reason:
              description: >-
                The reason why the run is incomplete. This will point to which
                specific token limit was reached over the course of the run.
              type: string
              enum:
                - max_completion_tokens
                - max_prompt_tokens
        model:
          description: >-
            The model that the
            [assistant](https://platform.openai.com/docs/api-reference/assistants)
            used for this run.
          type: string
        instructions:
          description: >-
            The instructions that the
            [assistant](https://platform.openai.com/docs/api-reference/assistants)
            used for this run.
          type: string
        tools:
          description: >-
            The list of tools that the
            [assistant](https://platform.openai.com/docs/api-reference/assistants)
            used for this run.
          default: []
          type: array
          maxItems: 20
          items:
            oneOf:
              - $ref: '#/components/schemas/AssistantToolsCode'
              - $ref: '#/components/schemas/AssistantToolsFileSearch'
              - $ref: '#/components/schemas/AssistantToolsFunction'
            x-oaiExpandable: true
        metadata:
          description: >
            Set of 16 key-value pairs that can be attached to an object. This
            can be useful for storing additional information about the object in
            a structured format. Keys can be a maximum of 64 characters long and
            values can be a maxium of 512 characters long.
          type: object
          x-oaiTypeLabel: map
          nullable: true
        usage:
          $ref: '#/components/schemas/RunCompletionUsage'
        temperature:
          description: >-
            The sampling temperature used for this run. If not set, defaults to
            1.
          type: number
          nullable: true
        top_p:
          description: >-
            The nucleus sampling value used for this run. If not set, defaults
            to 1.
          type: number
          nullable: true
        max_prompt_tokens:
          type: integer
          nullable: true
          description: >
            The maximum number of prompt tokens specified to have been used over
            the course of the run.
          minimum: 256
        max_completion_tokens:
          type: integer
          nullable: true
          description: >
            The maximum number of completion tokens specified to have been used
            over the course of the run.
          minimum: 256
        truncation_strategy:
          $ref: '#/components/schemas/TruncationObject'
          nullable: true
        tool_choice:
          $ref: '#/components/schemas/AssistantsApiToolChoiceOption'
          nullable: true
        parallel_tool_calls:
          $ref: '#/components/schemas/ParallelToolCalls'
        response_format:
          $ref: '#/components/schemas/AssistantsApiResponseFormatOption'
          nullable: true
      required:
        - id
        - object
        - created_at
        - thread_id
        - assistant_id
        - status
        - required_action
        - last_error
        - expires_at
        - started_at
        - cancelled_at
        - failed_at
        - completed_at
        - model
        - instructions
        - tools
        - metadata
        - usage
        - incomplete_details
        - max_prompt_tokens
        - max_completion_tokens
        - truncation_strategy
        - tool_choice
        - parallel_tool_calls
        - response_format
      x-code-samples:
        name: The run object
        beta: true
        example: |
          {
            "id": "run_abc123",
            "object": "thread.run",
            "created_at": 1698107661,
            "assistant_id": "asst_abc123",
            "thread_id": "thread_abc123",
            "status": "completed",
            "started_at": 1699073476,
            "expires_at": null,
            "cancelled_at": null,
            "failed_at": null,
            "completed_at": 1699073498,
            "last_error": null,
            "model": "gpt-4-turbo",
            "instructions": null,
            "tools": [{"type": "file_search"}, {"type": "code_interpreter"}],
            "metadata": {},
            "incomplete_details": null,
            "usage": {
              "prompt_tokens": 123,
              "completion_tokens": 456,
              "total_tokens": 579
            },
            "temperature": 1.0,
            "top_p": 1.0,
            "max_prompt_tokens": 1000,
            "max_completion_tokens": 1000,
            "truncation_strategy": {
              "type": "auto",
              "last_messages": null
            },
            "response_format": "auto",
            "tool_choice": "auto",
            "parallel_tool_calls": true
          }
    CreateThreadRequest:
      type: object
      additionalProperties: false
      properties:
        messages:
          description: >-
            A list of
            [messages](https://platform.openai.com/docs/api-reference/messages)
            to start the thread with.
          type: array
          items:
            $ref: '#/components/schemas/CreateMessageRequest'
        tool_resources:
          type: object
          description: >
            A set of resources that are made available to the assistant's tools
            in this thread. The resources are specific to the type of tool. For
            example, the `code_interpreter` tool requires a list of file IDs,
            while the `file_search` tool requires a list of vector store IDs.
          properties:
            code_interpreter:
              type: object
              properties:
                file_ids:
                  type: array
                  description: >
                    A list of
                    [file](https://platform.openai.com/docs/api-reference/files)
                    IDs made available to the `code_interpreter` tool. There can
                    be a maximum of 20 files associated with the tool.
                  default: []
                  maxItems: 20
                  items:
                    type: string
            file_search:
              type: object
              properties:
                vector_store_ids:
                  type: array
                  description: >
                    The [vector
                    store](https://platform.openai.com/docs/api-reference/vector-stores/object)
                    attached to this thread. There can be a maximum of 1 vector
                    store attached to the thread.
                  maxItems: 1
                  items:
                    type: string
                vector_stores:
                  type: array
                  description: >
                    A helper to create a [vector
                    store](https://platform.openai.com/docs/api-reference/vector-stores/object)
                    with file_ids and attach it to this thread. There can be a
                    maximum of 1 vector store attached to the thread.
                  maxItems: 1
                  items:
                    type: object
                    properties:
                      file_ids:
                        type: array
                        description: >
                          A list of
                          [file](https://platform.openai.com/docs/api-reference/files)
                          IDs to add to the vector store. There can be a maximum
                          of 10000 files in a vector store.
                        maxItems: 10000
                        items:
                          type: string
                      chunking_strategy:
                        type: object
                        description: >-
                          The chunking strategy used to chunk the file(s). If
                          not set, will use the `auto` strategy.
                        oneOf:
                          - type: object
                            title: Auto Chunking Strategy
                            description: >-
                              The default strategy. This strategy currently uses
                              a `max_chunk_size_tokens` of `800` and
                              `chunk_overlap_tokens` of `400`.
                            additionalProperties: false
                            properties:
                              type:
                                type: string
                                description: Always `auto`.
                                enum:
                                  - auto
                            required:
                              - type
                          - type: object
                            title: Static Chunking Strategy
                            additionalProperties: false
                            properties:
                              type:
                                type: string
                                description: Always `static`.
                                enum:
                                  - static
                              static:
                                type: object
                                additionalProperties: false
                                properties:
                                  max_chunk_size_tokens:
                                    type: integer
                                    minimum: 100
                                    maximum: 4096
                                    description: >-
                                      The maximum number of tokens in each
                                      chunk. The default value is `800`. The
                                      minimum value is `100` and the maximum
                                      value is `4096`.
                                  chunk_overlap_tokens:
                                    type: integer
                                    description: >
                                      The number of tokens that overlap between
                                      chunks. The default value is `400`.


                                      Note that the overlap must not exceed half
                                      of `max_chunk_size_tokens`.
                                required:
                                  - max_chunk_size_tokens
                                  - chunk_overlap_tokens
                            required:
                              - type
                              - static
                        x-oaiExpandable: true
                      metadata:
                        type: object
                        description: >
                          Set of 16 key-value pairs that can be attached to a
                          vector store. This can be useful for storing
                          additional information about the vector store in a
                          structured format. Keys can be a maximum of 64
                          characters long and values can be a maxium of 512
                          characters long.
                        x-oaiTypeLabel: map
                    x-oaiExpandable: true
              oneOf:
                - required:
                    - vector_store_ids
                - required:
                    - vector_stores
          nullable: true
        metadata:
          description: >
            Set of 16 key-value pairs that can be attached to an object. This
            can be useful for storing additional information about the object in
            a structured format. Keys can be a maximum of 64 characters long and
            values can be a maxium of 512 characters long.
          type: object
          x-oaiTypeLabel: map
          nullable: true
    AssistantToolsCode:
      type: object
      title: Code interpreter tool
      properties:
        type:
          type: string
          description: 'The type of tool being defined: `code_interpreter`'
          enum:
            - code_interpreter
      required:
        - type
    AssistantToolsFileSearch:
      type: object
      title: FileSearch tool
      properties:
        type:
          type: string
          description: 'The type of tool being defined: `file_search`'
          enum:
            - file_search
        file_search:
          type: object
          description: Overrides for the file search tool.
          properties:
            max_num_results:
              type: integer
              minimum: 1
              maximum: 50
              description: >
                The maximum number of results the file search tool should
                output. The default is 20 for gpt-4* models and 5 for
                gpt-3.5-turbo. This number should be between 1 and 50 inclusive.


                Note that the file search tool may output fewer than
                `max_num_results` results. See the [file search tool
                documentation](https://platform.openai.com/docs/assistants/tools/file-search/number-of-chunks-returned)
                for more information.
      required:
        - type
    AssistantToolsFunction:
      type: object
      title: Function tool
      properties:
        type:
          type: string
          description: 'The type of tool being defined: `function`'
          enum:
            - function
        function:
          $ref: '#/components/schemas/FunctionObject'
      required:
        - type
        - function
    TruncationObject:
      type: object
      title: Thread Truncation Controls
      description: >-
        Controls for how a thread will be truncated prior to the run. Use this
        to control the intial context window of the run.
      properties:
        type:
          type: string
          description: >-
            The truncation strategy to use for the thread. The default is
            `auto`. If set to `last_messages`, the thread will be truncated to
            the n most recent messages in the thread. When set to `auto`,
            messages in the middle of the thread will be dropped to fit the
            context length of the model, `max_prompt_tokens`.
          enum:
            - auto
            - last_messages
        last_messages:
          type: integer
          description: >-
            The number of most recent messages from the thread when constructing
            the context for the run.
          minimum: 1
          nullable: true
      required:
        - type
    AssistantsApiToolChoiceOption:
      description: >
        Controls which (if any) tool is called by the model.

        `none` means the model will not call any tools and instead generates a
        message.

        `auto` is the default value and means the model can pick between
        generating a message or calling one or more tools.

        `required` means the model must call one or more tools before responding
        to the user.

        Specifying a particular tool like `{"type": "file_search"}` or `{"type":
        "function", "function": {"name": "my_function"}}` forces the model to
        call that tool.
      oneOf:
        - type: string
          description: >
            `none` means the model will not call any tools and instead generates
            a message. `auto` means the model can pick between generating a
            message or calling one or more tools. `required` means the model
            must call one or more tools before responding to the user.
          enum:
            - none
            - auto
            - required
        - $ref: '#/components/schemas/AssistantsNamedToolChoice'
      x-oaiExpandable: true
    ParallelToolCalls:
      description: >-
        Whether to enable [parallel function
        calling](https://platform.openai.com/docs/guides/function-calling/parallel-function-calling)
        during tool use.
      type: boolean
      default: true
    AssistantsApiResponseFormatOption:
      description: >
        Specifies the format that the model must output. Compatible with
        [GPT-4o](https://platform.openai.com/docs/models/gpt-4o), [GPT-4
        Turbo](https://platform.openai.com/docs/models/gpt-4-turbo-and-gpt-4),
        and all GPT-3.5 Turbo models since `gpt-3.5-turbo-1106`.


        Setting to `{ "type": "json_object" }` enables JSON mode, which
        guarantees the message the model generates is valid JSON.


        **Important:** when using JSON mode, you **must** also instruct the
        model to produce JSON yourself via a system or user message. Without
        this, the model may generate an unending stream of whitespace until the
        generation reaches the token limit, resulting in a long-running and
        seemingly "stuck" request. Also note that the message content may be
        partially cut off if `finish_reason="length"`, which indicates the
        generation exceeded `max_tokens` or the conversation exceeded the max
        context length.
      oneOf:
        - type: string
          description: |
            `auto` is the default value
          enum:
            - none
            - auto
        - $ref: '#/components/schemas/AssistantsApiResponseFormat'
      x-oaiExpandable: true
    RunToolCallObject:
      type: object
      description: Tool call objects
      properties:
        id:
          type: string
          description: >-
            The ID of the tool call. This ID must be referenced when you submit
            the tool outputs in using the [Submit tool outputs to
            run](https://platform.openai.com/docs/api-reference/runs/submitToolOutputs)
            endpoint.
        type:
          type: string
          description: >-
            The type of tool call the output is required for. For now, this is
            always `function`.
          enum:
            - function
        function:
          type: object
          description: The function definition.
          properties:
            name:
              type: string
              description: The name of the function.
            arguments:
              type: string
              description: >-
                The arguments that the model expects you to pass to the
                function.
          required:
            - name
            - arguments
      required:
        - id
        - type
        - function
    RunCompletionUsage:
      type: object
      description: >-
        Usage statistics related to the run. This value will be `null` if the
        run is not in a terminal state (i.e. `in_progress`, `queued`, etc.).
      properties:
        completion_tokens:
          type: integer
          description: Number of completion tokens used over the course of the run.
        prompt_tokens:
          type: integer
          description: Number of prompt tokens used over the course of the run.
        total_tokens:
          type: integer
          description: Total number of tokens used (prompt + completion).
      required:
        - prompt_tokens
        - completion_tokens
        - total_tokens
      nullable: true
    CreateMessageRequest:
      type: object
      additionalProperties: false
      required:
        - role
        - content
      properties:
        role:
          type: string
          enum:
            - user
            - assistant
          description: >
            The role of the entity that is creating the message. Allowed values
            include:

            - `user`: Indicates the message is sent by an actual user and should
            be used in most cases to represent user-generated messages.

            - `assistant`: Indicates the message is generated by the assistant.
            Use this value to insert messages from the assistant into the
            conversation.
        content:
          oneOf:
            - type: string
              description: The text contents of the message.
              title: Text content
            - type: array
              description: >-
                An array of content parts with a defined type, each can be of
                type `text` or images can be passed with `image_url` or
                `image_file`. Image types are only supported on
                [Vision-compatible
                models](https://platform.openai.com/docs/models/overview).
              title: Array of content parts
              items:
                oneOf:
                  - $ref: '#/components/schemas/MessageContentImageFileObject'
                  - $ref: '#/components/schemas/MessageContentImageUrlObject'
                  - $ref: '#/components/schemas/MessageRequestContentTextObject'
                x-oaiExpandable: true
              minItems: 1
          x-oaiExpandable: true
        attachments:
          type: array
          items:
            type: object
            properties:
              file_id:
                type: string
                description: The ID of the file to attach to the message.
              tools:
                description: The tools to add this file to.
                type: array
                items:
                  oneOf:
                    - $ref: '#/components/schemas/AssistantToolsCode'
                    - $ref: '#/components/schemas/AssistantToolsFileSearchTypeOnly'
                  x-oaiExpandable: true
          description: >-
            A list of files attached to the message, and the tools they should
            be added to.
          required:
            - file_id
            - tools
          nullable: true
        metadata:
          description: >
            Set of 16 key-value pairs that can be attached to an object. This
            can be useful for storing additional information about the object in
            a structured format. Keys can be a maximum of 64 characters long and
            values can be a maxium of 512 characters long.
          type: object
          x-oaiTypeLabel: map
          nullable: true
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
    AssistantsNamedToolChoice:
      type: object
      description: >-
        Specifies a tool the model should use. Use to force the model to call a
        specific tool.
      properties:
        type:
          type: string
          enum:
            - function
            - code_interpreter
            - file_search
          description: >-
            The type of the tool. If type is `function`, the function name must
            be set
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
    AssistantsApiResponseFormat:
      type: object
      description: >
        An object describing the expected output of the model. If `json_object`
        only `function` type `tools` are allowed to be passed to the Run. If
        `text` the model can return text or any value needed.
      properties:
        type:
          type: string
          enum:
            - text
            - json_object
          example: json_object
          default: text
          description: Must be one of `text` or `json_object`.
    MessageContentImageFileObject:
      title: Image file
      type: object
      description: >-
        References an image
        [File](https://platform.openai.com/docs/api-reference/files) in the
        content of a message.
      properties:
        type:
          description: Always `image_file`.
          type: string
          enum:
            - image_file
        image_file:
          type: object
          properties:
            file_id:
              description: >-
                The [File](https://platform.openai.com/docs/api-reference/files)
                ID of the image in the message content. Set `purpose="vision"`
                when uploading the File if you need to later display the file
                content.
              type: string
            detail:
              type: string
              description: >-
                Specifies the detail level of the image if specified by the
                user. `low` uses fewer tokens, you can opt in to high resolution
                using `high`.
              enum:
                - auto
                - low
                - high
              default: auto
          required:
            - file_id
      required:
        - type
        - image_file
    MessageContentImageUrlObject:
      title: Image URL
      type: object
      description: References an image URL in the content of a message.
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
              description: >-
                The external URL of the image, must be a supported image types:
                jpeg, jpg, png, gif, webp.
              format: uri
            detail:
              type: string
              description: >-
                Specifies the detail level of the image. `low` uses fewer
                tokens, you can opt in to high resolution using `high`. Default
                value is `auto`
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
    MessageRequestContentTextObject:
      title: Text
      type: object
      description: The text content that is part of a message.
      properties:
        type:
          description: Always `text`.
          type: string
          enum:
            - text
        text:
          type: string
          description: Text content to be sent to the model
      required:
        - type
        - text
    AssistantToolsFileSearchTypeOnly:
      type: object
      title: FileSearch tool
      properties:
        type:
          type: string
          description: 'The type of tool being defined: `file_search`'
          enum:
            - file_search
      required:
        - type
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