> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Create Message



## OpenAPI

````yaml post /threads/{thread_id}/messages
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
  /threads/{thread_id}/messages:
    servers:
      - url: https://api.portkey.ai/v1
        description: Portkey API Public Endpoint
      - url: SELF_HOSTED_GATEWAY_URL
        description: Self-Hosted Gateway URL
    post:
      tags:
        - Assistants
      summary: Create a message.
      operationId: createMessage
      parameters:
        - in: path
          name: thread_id
          required: true
          schema:
            type: string
          description: >-
            The ID of the
            [thread](https://platform.openai.com/docs/api-reference/threads) to
            create a message for.
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateMessageRequest'
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/MessageObject'
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
            curl https://api.portkey.ai/v1/threads/thread_abc123/messages \
              -H "Content-Type: application/json" \
              -H "x-portkey-api-key: $PORTKEY_API_KEY" \
              -H "x-portkey-virtual-key: $PORTKEY_PROVIDER_VIRTUAL_KEY" \
              -H "OpenAI-Beta: assistants=v2" \
              -d '{
                  "role": "user",
                  "content": "How does AI work? Explain it in simple terms."
                }'
        - lang: python
          source: |
            from portkey_ai import Portkey

            client = Portkey(
              api_key = "PORTKEY_API_KEY",
              virtual_key = "PROVIDER_VIRTUAL_KEY"
            )

            thread_message = client.beta.threads.messages.create(
              "thread_abc123",
              role="user",
              content="How does AI work? Explain it in simple terms.",
            )
            print(thread_message)
        - lang: javascript
          source: |
            import Portkey from 'portkey-ai';

            const client = new Portkey({
              apiKey: 'PORTKEY_API_KEY',
              virtualKey: 'PROVIDER_VIRTUAL_KEY'
            });

            async function main() {
              const threadMessages = await client.beta.threads.messages.create(
                "thread_abc123",
                { role: "user", content: "How does AI work? Explain it in simple terms." }
              );

              console.log(threadMessages);
            }

            main();
          response: |
            {
              "id": "msg_abc123",
              "object": "thread.message",
              "created_at": 1713226573,
              "assistant_id": null,
              "thread_id": "thread_abc123",
              "run_id": null,
              "role": "user",
              "content": [
                {
                  "type": "text",
                  "text": {
                    "value": "How does AI work? Explain it in simple terms.",
                    "annotations": []
                  }
                }
              ],
              "attachments": [],
              "metadata": {}
            }
components:
  schemas:
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
    MessageObject:
      type: object
      title: The message object
      description: >-
        Represents a message within a
        [thread](https://platform.openai.com/docs/api-reference/threads).
      properties:
        id:
          description: The identifier, which can be referenced in API endpoints.
          type: string
        object:
          description: The object type, which is always `thread.message`.
          type: string
          enum:
            - thread.message
        created_at:
          description: The Unix timestamp (in seconds) for when the message was created.
          type: integer
        thread_id:
          description: >-
            The [thread](https://platform.openai.com/docs/api-reference/threads)
            ID that this message belongs to.
          type: string
        status:
          description: >-
            The status of the message, which can be either `in_progress`,
            `incomplete`, or `completed`.
          type: string
          enum:
            - in_progress
            - incomplete
            - completed
        incomplete_details:
          description: >-
            On an incomplete message, details about why the message is
            incomplete.
          type: object
          properties:
            reason:
              type: string
              description: The reason the message is incomplete.
              enum:
                - content_filter
                - max_tokens
                - run_cancelled
                - run_expired
                - run_failed
          nullable: true
          required:
            - reason
        completed_at:
          description: The Unix timestamp (in seconds) for when the message was completed.
          type: integer
          nullable: true
        incomplete_at:
          description: >-
            The Unix timestamp (in seconds) for when the message was marked as
            incomplete.
          type: integer
          nullable: true
        role:
          description: The entity that produced the message. One of `user` or `assistant`.
          type: string
          enum:
            - user
            - assistant
        content:
          description: The content of the message in array of text and/or images.
          type: array
          items:
            oneOf:
              - $ref: '#/components/schemas/MessageContentImageFileObject'
              - $ref: '#/components/schemas/MessageContentImageUrlObject'
              - $ref: '#/components/schemas/MessageContentTextObject'
            x-oaiExpandable: true
        assistant_id:
          description: >-
            If applicable, the ID of the
            [assistant](https://platform.openai.com/docs/api-reference/assistants)
            that authored this message.
          type: string
          nullable: true
        run_id:
          description: >-
            The ID of the
            [run](https://platform.openai.com/docs/api-reference/runs)
            associated with the creation of this message. Value is `null` when
            messages are created manually using the create message or create
            thread endpoints.
          type: string
          nullable: true
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
            A list of files attached to the message, and the tools they were
            added to.
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
      required:
        - id
        - object
        - created_at
        - thread_id
        - status
        - incomplete_details
        - completed_at
        - incomplete_at
        - role
        - content
        - assistant_id
        - run_id
        - attachments
        - metadata
      x-code-samples:
        name: The message object
        beta: true
        example: |
          {
            "id": "msg_abc123",
            "object": "thread.message",
            "created_at": 1698983503,
            "thread_id": "thread_abc123",
            "role": "assistant",
            "content": [
              {
                "type": "text",
                "text": {
                  "value": "Hi! How can I help you today?",
                  "annotations": []
                }
              }
            ],
            "assistant_id": "asst_abc123",
            "run_id": "run_abc123",
            "attachments": [],
            "metadata": {}
          }
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
    MessageContentTextObject:
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
          type: object
          properties:
            value:
              description: The data that makes up the text.
              type: string
            annotations:
              type: array
              items:
                oneOf:
                  - $ref: >-
                      #/components/schemas/MessageContentTextAnnotationsFileCitationObject
                  - $ref: >-
                      #/components/schemas/MessageContentTextAnnotationsFilePathObject
                x-oaiExpandable: true
          required:
            - value
            - annotations
      required:
        - type
        - text
    MessageContentTextAnnotationsFileCitationObject:
      title: File citation
      type: object
      description: >-
        A citation within the message that points to a specific quote from a
        specific File associated with the assistant or the message. Generated
        when the assistant uses the "file_search" tool to search files.
      properties:
        type:
          description: Always `file_citation`.
          type: string
          enum:
            - file_citation
        text:
          description: The text in the message content that needs to be replaced.
          type: string
        file_citation:
          type: object
          properties:
            file_id:
              description: The ID of the specific File the citation is from.
              type: string
            quote:
              description: The specific quote in the file.
              type: string
          required:
            - file_id
            - quote
        start_index:
          type: integer
          minimum: 0
        end_index:
          type: integer
          minimum: 0
      required:
        - type
        - text
        - file_citation
        - start_index
        - end_index
    MessageContentTextAnnotationsFilePathObject:
      title: File path
      type: object
      description: >-
        A URL for the file that's generated when the assistant used the
        `code_interpreter` tool to generate a file.
      properties:
        type:
          description: Always `file_path`.
          type: string
          enum:
            - file_path
        text:
          description: The text in the message content that needs to be replaced.
          type: string
        file_path:
          type: object
          properties:
            file_id:
              description: The ID of the file that was generated.
              type: string
          required:
            - file_id
        start_index:
          type: integer
          minimum: 0
        end_index:
          type: integer
          minimum: 0
      required:
        - type
        - text
        - file_path
        - start_index
        - end_index
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