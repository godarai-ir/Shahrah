> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Retrieve Assistant



## OpenAPI

````yaml get /assistants/{assistant_id}
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
  /assistants/{assistant_id}:
    servers:
      - url: https://api.portkey.ai/v1
        description: Portkey API Public Endpoint
      - url: SELF_HOSTED_GATEWAY_URL
        description: Self-Hosted Gateway URL
    get:
      tags:
        - Assistants
      summary: Retrieves an assistant.
      operationId: getAssistant
      parameters:
        - in: path
          name: assistant_id
          required: true
          schema:
            type: string
          description: The ID of the assistant to retrieve.
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/AssistantObject'
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
            curl https://api.portkey.ai/v1/assistants/asst_abc123 \
              -H "Content-Type: application/json" \
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

            my_assistant = client.beta.assistants.retrieve("asst_abc123")
            print(my_assistant)
        - lang: javascript
          source: |
            import Portkey from 'portkey-ai';

            const client = new Portkey({
              apiKey: 'PORTKEY_API_KEY',
              virtualKey: 'PROVIDER_VIRTUAL_KEY'
            });

            async function main() {
              const myAssistant = await client.beta.assistants.retrieve(
                "asst_abc123"
              );

              console.log(myAssistant);
            }

            main();
          response: |
            {
              "id": "asst_abc123",
              "object": "assistant",
              "created_at": 1699009709,
              "name": "HR Helper",
              "description": null,
              "model": "gpt-4-turbo",
              "instructions": "You are an HR bot, and you have access to files to answer employee questions about company policies.",
              "tools": [
                {
                  "type": "file_search"
                }
              ],
              "metadata": {},
              "top_p": 1.0,
              "temperature": 1.0,
              "response_format": "auto"
            }
components:
  schemas:
    AssistantObject:
      type: object
      title: Assistant
      description: Represents an `assistant` that can call the model and use tools.
      properties:
        id:
          description: The identifier, which can be referenced in API endpoints.
          type: string
        object:
          description: The object type, which is always `assistant`.
          type: string
          enum:
            - assistant
        created_at:
          description: The Unix timestamp (in seconds) for when the assistant was created.
          type: integer
        name:
          description: |
            The name of the assistant. The maximum length is 256 characters.
          type: string
          maxLength: 256
          nullable: true
        description:
          description: >
            The description of the assistant. The maximum length is 512
            characters.
          type: string
          maxLength: 512
          nullable: true
        model:
          description: >
            ID of the model to use. You can use the [List
            models](https://platform.openai.com/docs/api-reference/models/list)
            API to see all of your available models, or see our [Model
            overview](https://platform.openai.com/docs/models/overview) for
            descriptions of them.
          type: string
        instructions:
          description: >
            The system instructions that the assistant uses. The maximum length
            is 256,000 characters.
          type: string
          maxLength: 256000
          nullable: true
        tools:
          description: >
            A list of tool enabled on the assistant. There can be a maximum of
            128 tools per assistant. Tools can be of types `code_interpreter`,
            `file_search`, or `function`.
          default: []
          type: array
          maxItems: 128
          items:
            oneOf:
              - $ref: '#/components/schemas/AssistantToolsCode'
              - $ref: '#/components/schemas/AssistantToolsFileSearch'
              - $ref: '#/components/schemas/AssistantToolsFunction'
            x-oaiExpandable: true
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
                    IDs made available to the `code_interpreter`` tool. There
                    can be a maximum of 20 files associated with the tool.
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
        response_format:
          $ref: '#/components/schemas/AssistantsApiResponseFormatOption'
          nullable: true
      required:
        - id
        - object
        - created_at
        - name
        - description
        - model
        - instructions
        - tools
        - metadata
      x-code-samples:
        name: The assistant object
        beta: true
        example: |
          {
            "id": "asst_abc123",
            "object": "assistant",
            "created_at": 1698984975,
            "name": "Math Tutor",
            "description": null,
            "model": "gpt-4-turbo",
            "instructions": "You are a personal math tutor. When asked a question, write and run Python code to answer the question.",
            "tools": [
              {
                "type": "code_interpreter"
              }
            ],
            "metadata": {},
            "top_p": 1.0,
            "temperature": 1.0,
            "response_format": "auto"
          }
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