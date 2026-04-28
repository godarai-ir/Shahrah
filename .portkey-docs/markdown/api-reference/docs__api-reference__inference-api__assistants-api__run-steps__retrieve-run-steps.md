> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Retrieve Run Steps



## OpenAPI

````yaml get /threads/{thread_id}/runs/{run_id}/steps/{step_id}
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
  /threads/{thread_id}/runs/{run_id}/steps/{step_id}:
    servers:
      - url: https://api.portkey.ai/v1
        description: Portkey API Public Endpoint
      - url: SELF_HOSTED_GATEWAY_URL
        description: Self-Hosted Gateway URL
    get:
      tags:
        - Assistants
      summary: Retrieves a run step.
      operationId: getRunStep
      parameters:
        - in: path
          name: thread_id
          required: true
          schema:
            type: string
          description: The ID of the thread to which the run and run step belongs.
        - in: path
          name: run_id
          required: true
          schema:
            type: string
          description: The ID of the run to which the run step belongs.
        - in: path
          name: step_id
          required: true
          schema:
            type: string
          description: The ID of the run step to retrieve.
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/RunStepObject'
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
            curl
            https://api.portkey.ai/v1/threads/thread_abc123/runs/run_abc123/steps/step_abc123
            \
              -H "x-portkey-api-key: $PORTKEY_API_KEY" \
              -H "x-portkey-virtual-key: $PORTKEY_PROVIDER_VIRTUAL_KEY" \
              -H "Content-Type: application/json" \
              -H "OpenAI-Beta: assistants=v2"
        - lang: python
          source: |
            from portkey_ai import Portkey

            client = Portkey(
              api_key = "PORTKEY_API_KEY",
              virtual_key = "PROVIDER_VIRTUAL_KEY"
            )

            run_step = client.beta.threads.runs.steps.retrieve(
                thread_id="thread_abc123",
                run_id="run_abc123",
                step_id="step_abc123"
            )

            print(run_step)
        - lang: javascript
          source: |
            import Portkey from 'portkey-ai';

            const client = new Portkey({
              apiKey: 'PORTKEY_API_KEY',
              virtualKey: 'PROVIDER_VIRTUAL_KEY'
            });

            async function main() {
              const runStep = await client.beta.threads.runs.steps.retrieve(
                "thread_abc123",
                "run_abc123",
                "step_abc123"
              );
              console.log(runStep);
            }

            main();
          response: |
            {
              "id": "step_abc123",
              "object": "thread.run.step",
              "created_at": 1699063291,
              "run_id": "run_abc123",
              "assistant_id": "asst_abc123",
              "thread_id": "thread_abc123",
              "type": "message_creation",
              "status": "completed",
              "cancelled_at": null,
              "completed_at": 1699063291,
              "expired_at": null,
              "failed_at": null,
              "last_error": null,
              "step_details": {
                "type": "message_creation",
                "message_creation": {
                  "message_id": "msg_abc123"
                }
              },
              "usage": {
                "prompt_tokens": 123,
                "completion_tokens": 456,
                "total_tokens": 579
              }
            }
components:
  schemas:
    RunStepObject:
      type: object
      title: Run steps
      description: |
        Represents a step in execution of a run.
      properties:
        id:
          description: >-
            The identifier of the run step, which can be referenced in API
            endpoints.
          type: string
        object:
          description: The object type, which is always `thread.run.step`.
          type: string
          enum:
            - thread.run.step
        created_at:
          description: The Unix timestamp (in seconds) for when the run step was created.
          type: integer
        assistant_id:
          description: >-
            The ID of the
            [assistant](https://platform.openai.com/docs/api-reference/assistants)
            associated with the run step.
          type: string
        thread_id:
          description: >-
            The ID of the
            [thread](https://platform.openai.com/docs/api-reference/threads)
            that was run.
          type: string
        run_id:
          description: >-
            The ID of the
            [run](https://platform.openai.com/docs/api-reference/runs) that this
            run step is a part of.
          type: string
        type:
          description: >-
            The type of run step, which can be either `message_creation` or
            `tool_calls`.
          type: string
          enum:
            - message_creation
            - tool_calls
        status:
          description: >-
            The status of the run step, which can be either `in_progress`,
            `cancelled`, `failed`, `completed`, or `expired`.
          type: string
          enum:
            - in_progress
            - cancelled
            - failed
            - completed
            - expired
        step_details:
          type: object
          description: The details of the run step.
          oneOf:
            - $ref: '#/components/schemas/RunStepDetailsMessageCreationObject'
            - $ref: '#/components/schemas/RunStepDetailsToolCallsObject'
          x-oaiExpandable: true
        last_error:
          type: object
          description: >-
            The last error associated with this run step. Will be `null` if
            there are no errors.
          nullable: true
          properties:
            code:
              type: string
              description: One of `server_error` or `rate_limit_exceeded`.
              enum:
                - server_error
                - rate_limit_exceeded
            message:
              type: string
              description: A human-readable description of the error.
          required:
            - code
            - message
        expired_at:
          description: >-
            The Unix timestamp (in seconds) for when the run step expired. A
            step is considered expired if the parent run is expired.
          type: integer
          nullable: true
        cancelled_at:
          description: The Unix timestamp (in seconds) for when the run step was cancelled.
          type: integer
          nullable: true
        failed_at:
          description: The Unix timestamp (in seconds) for when the run step failed.
          type: integer
          nullable: true
        completed_at:
          description: The Unix timestamp (in seconds) for when the run step completed.
          type: integer
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
        usage:
          $ref: '#/components/schemas/RunStepCompletionUsage'
      required:
        - id
        - object
        - created_at
        - assistant_id
        - thread_id
        - run_id
        - type
        - status
        - step_details
        - last_error
        - expired_at
        - cancelled_at
        - failed_at
        - completed_at
        - metadata
        - usage
      x-code-samples:
        name: The run step object
        beta: true
        example: |
          {
            "id": "step_abc123",
            "object": "thread.run.step",
            "created_at": 1699063291,
            "run_id": "run_abc123",
            "assistant_id": "asst_abc123",
            "thread_id": "thread_abc123",
            "type": "message_creation",
            "status": "completed",
            "cancelled_at": null,
            "completed_at": 1699063291,
            "expired_at": null,
            "failed_at": null,
            "last_error": null,
            "step_details": {
              "type": "message_creation",
              "message_creation": {
                "message_id": "msg_abc123"
              }
            },
            "usage": {
              "prompt_tokens": 123,
              "completion_tokens": 456,
              "total_tokens": 579
            }
          }
    RunStepDetailsMessageCreationObject:
      title: Message creation
      type: object
      description: Details of the message creation by the run step.
      properties:
        type:
          description: Always `message_creation`.
          type: string
          enum:
            - message_creation
        message_creation:
          type: object
          properties:
            message_id:
              type: string
              description: The ID of the message that was created by this run step.
          required:
            - message_id
      required:
        - type
        - message_creation
    RunStepDetailsToolCallsObject:
      title: Tool calls
      type: object
      description: Details of the tool call.
      properties:
        type:
          description: Always `tool_calls`.
          type: string
          enum:
            - tool_calls
        tool_calls:
          type: array
          description: >
            An array of tool calls the run step was involved in. These can be
            associated with one of three types of tools: `code_interpreter`,
            `file_search`, or `function`.
          items:
            oneOf:
              - $ref: '#/components/schemas/RunStepDetailsToolCallsCodeObject'
              - $ref: '#/components/schemas/RunStepDetailsToolCallsFileSearchObject'
              - $ref: '#/components/schemas/RunStepDetailsToolCallsFunctionObject'
            x-oaiExpandable: true
      required:
        - type
        - tool_calls
    RunStepCompletionUsage:
      type: object
      description: >-
        Usage statistics related to the run step. This value will be `null`
        while the run step's status is `in_progress`.
      properties:
        completion_tokens:
          type: integer
          description: Number of completion tokens used over the course of the run step.
        prompt_tokens:
          type: integer
          description: Number of prompt tokens used over the course of the run step.
        total_tokens:
          type: integer
          description: Total number of tokens used (prompt + completion).
      required:
        - prompt_tokens
        - completion_tokens
        - total_tokens
      nullable: true
    RunStepDetailsToolCallsCodeObject:
      title: Code Interpreter tool call
      type: object
      description: Details of the Code Interpreter tool call the run step was involved in.
      properties:
        id:
          type: string
          description: The ID of the tool call.
        type:
          type: string
          description: >-
            The type of tool call. This is always going to be `code_interpreter`
            for this type of tool call.
          enum:
            - code_interpreter
        code_interpreter:
          type: object
          description: The Code Interpreter tool call definition.
          required:
            - input
            - outputs
          properties:
            input:
              type: string
              description: The input to the Code Interpreter tool call.
            outputs:
              type: array
              description: >-
                The outputs from the Code Interpreter tool call. Code
                Interpreter can output one or more items, including text
                (`logs`) or images (`image`). Each of these are represented by a
                different object type.
              items:
                type: object
                oneOf:
                  - $ref: >-
                      #/components/schemas/RunStepDetailsToolCallsCodeOutputLogsObject
                  - $ref: >-
                      #/components/schemas/RunStepDetailsToolCallsCodeOutputImageObject
                x-oaiExpandable: true
      required:
        - id
        - type
        - code_interpreter
    RunStepDetailsToolCallsFileSearchObject:
      title: File search tool call
      type: object
      properties:
        id:
          type: string
          description: The ID of the tool call object.
        type:
          type: string
          description: >-
            The type of tool call. This is always going to be `file_search` for
            this type of tool call.
          enum:
            - file_search
        file_search:
          type: object
          description: For now, this is always going to be an empty object.
          x-oaiTypeLabel: map
      required:
        - id
        - type
        - file_search
    RunStepDetailsToolCallsFunctionObject:
      type: object
      title: Function tool call
      properties:
        id:
          type: string
          description: The ID of the tool call object.
        type:
          type: string
          description: >-
            The type of tool call. This is always going to be `function` for
            this type of tool call.
          enum:
            - function
        function:
          type: object
          description: The definition of the function that was called.
          properties:
            name:
              type: string
              description: The name of the function.
            arguments:
              type: string
              description: The arguments passed to the function.
            output:
              type: string
              description: >-
                The output of the function. This will be `null` if the outputs
                have not been
                [submitted](https://platform.openai.com/docs/api-reference/runs/submitToolOutputs)
                yet.
              nullable: true
          required:
            - name
            - arguments
            - output
      required:
        - id
        - type
        - function
    RunStepDetailsToolCallsCodeOutputLogsObject:
      title: Code Interpreter log output
      type: object
      description: Text output from the Code Interpreter tool call as part of a run step.
      properties:
        type:
          description: Always `logs`.
          type: string
          enum:
            - logs
        logs:
          type: string
          description: The text output from the Code Interpreter tool call.
      required:
        - type
        - logs
    RunStepDetailsToolCallsCodeOutputImageObject:
      title: Code Interpreter image output
      type: object
      properties:
        type:
          description: Always `image`.
          type: string
          enum:
            - image
        image:
          type: object
          properties:
            file_id:
              description: >-
                The [file](https://platform.openai.com/docs/api-reference/files)
                ID of the image.
              type: string
          required:
            - file_id
      required:
        - type
        - image
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