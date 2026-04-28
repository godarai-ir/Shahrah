> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Retrieve Run



## OpenAPI

````yaml get /threads/{thread_id}/runs/{run_id}
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
  /threads/{thread_id}/runs/{run_id}:
    servers:
      - url: https://api.portkey.ai/v1
        description: Portkey API Public Endpoint
      - url: SELF_HOSTED_GATEWAY_URL
        description: Self-Hosted Gateway URL
    get:
      tags:
        - Assistants
      summary: Retrieves a run.
      operationId: getRun
      parameters:
        - in: path
          name: thread_id
          required: true
          schema:
            type: string
          description: >-
            The ID of the
            [thread](https://platform.openai.com/docs/api-reference/threads)
            that was run.
        - in: path
          name: run_id
          required: true
          schema:
            type: string
          description: The ID of the run to retrieve.
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
          source: >
            curl https://api.portkey.ai/v1/threads/thread_abc123/runs/run_abc123
            \
              -H "x-portkey-api-key: $PORTKEY_API_KEY" \
              -H "x-portkey-virtual-key: $PORTKEY_PROVIDER_VIRTUAL_KEY" \
              -H "OpenAI-Beta: assistants=v2"
        - lang: python
          source: |
            from portkey_ai import Portkey

            client = Portkey(
              api_key = "PORTKEY_API_KEY",
              virtual_key = "PROVIDER_VIRTUAL_KEY"
            )

            run = client.beta.threads.runs.retrieve(
              thread_id="thread_abc123",
              run_id="run_abc123"
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
              const run = await client.beta.threads.runs.retrieve(
                "thread_abc123",
                "run_abc123"
              );

              console.log(run);
            }

            main();
          response: |
            {
              "id": "run_abc123",
              "object": "thread.run",
              "created_at": 1699075072,
              "assistant_id": "asst_abc123",
              "thread_id": "thread_abc123",
              "status": "completed",
              "started_at": 1699075072,
              "expires_at": null,
              "cancelled_at": null,
              "failed_at": null,
              "completed_at": 1699075073,
              "last_error": null,
              "model": "gpt-4-turbo",
              "instructions": null,
              "incomplete_details": null,
              "tools": [
                {
                  "type": "code_interpreter"
                }
              ],
              "metadata": {},
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
components:
  schemas:
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