> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# List Input Items



## OpenAPI

````yaml get /responses/{response_id}/input_items
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
  /responses/{response_id}/input_items:
    servers:
      - url: https://api.portkey.ai/v1
        description: Portkey API Public Endpoint
      - url: SELF_HOSTED_GATEWAY_URL
        description: Self-Hosted Gateway URL
    get:
      tags:
        - Responses
      summary: Returns a list of input items for a given response.
      operationId: listInputItems
      parameters:
        - in: path
          name: response_id
          required: true
          schema:
            type: string
          description: The ID of the response to retrieve input items for.
        - name: limit
          in: query
          description: >
            A limit on the number of objects to be returned. Limit can range
            between

            1 and 100, and the default is 20.
          required: false
          schema:
            type: integer
            default: 20
        - in: query
          name: order
          schema:
            type: string
            enum:
              - asc
              - desc
          description: |
            The order to return the input items in. Default is `asc`.
            - `asc`: Return the input items in ascending order.
            - `desc`: Return the input items in descending order.
        - in: query
          name: after
          schema:
            type: string
          description: |
            An item ID to list items after, used in pagination.
        - in: query
          name: before
          schema:
            type: string
          description: |
            An item ID to list items before, used in pagination.
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ResponseItemList'
components:
  schemas:
    ResponseItemList:
      type: object
      description: A list of Response items.
      properties:
        object:
          type: string
          description: The type of object returned, must be `list`.
          enum:
            - list
          x-stainless-const: true
        data:
          type: array
          description: A list of items used to generate this response.
          items:
            $ref: '#/components/schemas/ItemResource'
        has_more:
          type: boolean
          description: Whether there are more items available.
        first_id:
          type: string
          description: The ID of the first item in the list.
        last_id:
          type: string
          description: The ID of the last item in the list.
      required:
        - object
        - data
        - has_more
        - first_id
        - last_id
      x-oaiExpandable: true
    ItemResource:
      description: |
        Content item used to generate a response.
      oneOf:
        - $ref: '#/components/schemas/InputMessageResource'
        - $ref: '#/components/schemas/OutputMessage'
        - $ref: '#/components/schemas/FileSearchToolCall'
        - $ref: '#/components/schemas/ComputerToolCall'
        - $ref: '#/components/schemas/ComputerToolCallOutputResource'
        - $ref: '#/components/schemas/WebSearchToolCall'
        - $ref: '#/components/schemas/FunctionToolCallResource'
        - $ref: '#/components/schemas/FunctionToolCallOutputResource'
      x-oaiExpandable: true
      discriminator:
        propertyName: type
    InputMessageResource:
      allOf:
        - $ref: '#/components/schemas/InputMessage'
        - type: object
          properties:
            id:
              type: string
              description: |
                The unique ID of the message input.
          required:
            - id
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
    ComputerToolCallOutputResource:
      allOf:
        - $ref: '#/components/schemas/ComputerToolCallOutput'
        - type: object
          properties:
            id:
              type: string
              description: |
                The unique ID of the computer call tool output.
          required:
            - id
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
    FunctionToolCallResource:
      allOf:
        - $ref: '#/components/schemas/FunctionToolCall'
        - type: object
          properties:
            id:
              type: string
              description: |
                The unique ID of the function tool call.
          required:
            - id
    FunctionToolCallOutputResource:
      allOf:
        - $ref: '#/components/schemas/FunctionToolCallOutput'
        - type: object
          properties:
            id:
              type: string
              description: |
                The unique ID of the function call tool output.
          required:
            - id
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
    OutputContent:
      oneOf:
        - $ref: '#/components/schemas/OutputText'
        - $ref: '#/components/schemas/Refusal'
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
    InputContent:
      oneOf:
        - $ref: '#/components/schemas/InputText'
        - $ref: '#/components/schemas/InputImage'
        - $ref: '#/components/schemas/InputFile'
      x-oaiExpandable: true
    Annotation:
      oneOf:
        - $ref: '#/components/schemas/FileCitation'
        - $ref: '#/components/schemas/UrlCitation'
        - $ref: '#/components/schemas/FilePath'
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
  securitySchemes:
    Portkey-Key:
      type: apiKey
      in: header
      name: x-portkey-api-key

````