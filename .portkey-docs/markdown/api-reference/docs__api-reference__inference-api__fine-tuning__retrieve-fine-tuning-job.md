> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Retrieve Fine-tuning Job



## OpenAPI

````yaml get /fine_tuning/jobs/{fine_tuning_job_id}
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
  /fine_tuning/jobs/{fine_tuning_job_id}:
    servers:
      - url: https://api.portkey.ai/v1
        description: Portkey API Public Endpoint
      - url: SELF_HOSTED_GATEWAY_URL
        description: Self-Hosted Gateway URL
    get:
      tags:
        - Fine-tuning
      summary: >
        Get info about a fine-tuning job.


        [Learn more about
        fine-tuning](https://platform.openai.com/docs/guides/fine-tuning)
      operationId: retrieveFineTuningJob
      parameters:
        - in: path
          name: fine_tuning_job_id
          required: true
          schema:
            type: string
            example: ft-AF1WoRqd3aJAHsqc9NY7iL8F
          description: |
            The ID of the fine-tuning job.
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/FineTuningJob'
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
          source: >
            curl
            https://api.portkey.ai/v1/fine_tuning/jobs/ft-AF1WoRqd3aJAHsqc9NY7iL8F
            \
              -H "x-portkey-api-key: $PORTKEY_API_KEY" \
              -H "x-portkey-virtual-key: $PORTKEY_PROVIDER_VIRTUAL_KEY"
        - lang: python
          label: Default
          source: |
            from portkey_ai import Portkey

            client = Portkey(
              api_key = "PORTKEY_API_KEY",
              virtual_key = "PROVIDER_VIRTUAL_KEY"
            )

            client.fine_tuning.jobs.retrieve("ftjob-abc123")
        - lang: javascript
          label: Default
          source: |
            import Portkey from 'portkey-ai';

            const client = new Portkey({
              apiKey: 'PORTKEY_API_KEY',
              virtualKey: 'PROVIDER_VIRTUAL_KEY'
            });

            async function main() {
              const fineTune = await client.fineTuning.jobs.retrieve("ftjob-abc123");

              console.log(fineTune);
            }

            main();
        - lang: curl
          label: Self-hosted
          source: >
            curl
            SELF_HOSTED_GATEWAY_URL/fine_tuning/jobs/ft-AF1WoRqd3aJAHsqc9NY7iL8F
            \
              -H "x-portkey-api-key: $PORTKEY_API_KEY" \
              -H "x-portkey-virtual-key: $PORTKEY_PROVIDER_VIRTUAL_KEY"
        - lang: python
          label: Self-hosted
          source: |
            from portkey_ai import Portkey

            client = Portkey(
              api_key = "PORTKEY_API_KEY",
              base_url = "SELF_HOSTED_GATEWAY_URL",
              virtual_key = "PROVIDER_VIRTUAL_KEY"
            )

            client.fine_tuning.jobs.retrieve("ftjob-abc123")
        - lang: javascript
          label: Self-hosted
          source: |
            import Portkey from 'portkey-ai';

            const client = new Portkey({
              apiKey: 'PORTKEY_API_KEY',
              baseUrl: 'SELF_HOSTED_GATEWAY_URL',
              virtualKey: 'PROVIDER_VIRTUAL_KEY'
            });

            async function main() {
              const fineTune = await client.fineTuning.jobs.retrieve("ftjob-abc123");

              console.log(fineTune);
            }

            main();
components:
  schemas:
    FineTuningJob:
      type: object
      title: FineTuningJob
      description: >
        The `fine_tuning.job` object represents a fine-tuning job that has been
        created through the API.
      properties:
        id:
          type: string
          description: The object identifier, which can be referenced in the API endpoints.
        created_at:
          type: integer
          description: >-
            The Unix timestamp (in seconds) for when the fine-tuning job was
            created.
        error:
          type: object
          nullable: true
          description: >-
            For fine-tuning jobs that have `failed`, this will contain more
            information on the cause of the failure.
          properties:
            code:
              type: string
              description: A machine-readable error code.
            message:
              type: string
              description: A human-readable error message.
            param:
              type: string
              description: >-
                The parameter that was invalid, usually `training_file` or
                `validation_file`. This field will be null if the failure was
                not parameter-specific.
              nullable: true
          required:
            - code
            - message
            - param
        fine_tuned_model:
          type: string
          nullable: true
          description: >-
            The name of the fine-tuned model that is being created. The value
            will be null if the fine-tuning job is still running.
        finished_at:
          type: integer
          nullable: true
          description: >-
            The Unix timestamp (in seconds) for when the fine-tuning job was
            finished. The value will be null if the fine-tuning job is still
            running.
        hyperparameters:
          type: object
          description: >-
            The hyperparameters used for the fine-tuning job. See the
            [fine-tuning
            guide](https://platform.openai.com/docs/guides/fine-tuning) for more
            details.
          properties:
            n_epochs:
              oneOf:
                - type: string
                  enum:
                    - auto
                - type: integer
                  minimum: 1
                  maximum: 50
              default: auto
              description: >-
                The number of epochs to train the model for. An epoch refers to
                one full cycle through the training dataset.

                "auto" decides the optimal number of epochs based on the size of
                the dataset. If setting the number manually, we support any
                number between 1 and 50 epochs.
          required:
            - n_epochs
        model:
          type: string
          description: The base model that is being fine-tuned.
        object:
          type: string
          description: The object type, which is always "fine_tuning.job".
          enum:
            - fine_tuning.job
        organization_id:
          type: string
          description: The organization that owns the fine-tuning job.
        result_files:
          type: array
          description: >-
            The compiled results file ID(s) for the fine-tuning job. You can
            retrieve the results with the [Files
            API](https://platform.openai.com/docs/api-reference/files/retrieve-contents).
          items:
            type: string
            example: file-abc123
        status:
          type: string
          description: >-
            The current status of the fine-tuning job, which can be either
            `validating_files`, `queued`, `running`, `succeeded`, `failed`, or
            `cancelled`.
          enum:
            - validating_files
            - queued
            - running
            - succeeded
            - failed
            - cancelled
        trained_tokens:
          type: integer
          nullable: true
          description: >-
            The total number of billable tokens processed by this fine-tuning
            job. The value will be null if the fine-tuning job is still running.
        training_file:
          type: string
          description: >-
            The file ID used for training. You can retrieve the training data
            with the [Files
            API](https://platform.openai.com/docs/api-reference/files/retrieve-contents).
        validation_file:
          type: string
          nullable: true
          description: >-
            The file ID used for validation. You can retrieve the validation
            results with the [Files
            API](https://platform.openai.com/docs/api-reference/files/retrieve-contents).
        integrations:
          type: array
          nullable: true
          description: A list of integrations to enable for this fine-tuning job.
          maxItems: 5
          items:
            oneOf:
              - $ref: '#/components/schemas/FineTuningIntegration'
            x-oaiExpandable: true
        seed:
          type: integer
          description: The seed used for the fine-tuning job.
        estimated_finish:
          type: integer
          nullable: true
          description: >-
            The Unix timestamp (in seconds) for when the fine-tuning job is
            estimated to finish. The value will be null if the fine-tuning job
            is not running.
      required:
        - created_at
        - error
        - finished_at
        - fine_tuned_model
        - hyperparameters
        - id
        - model
        - object
        - organization_id
        - result_files
        - status
        - trained_tokens
        - training_file
        - validation_file
        - seed
    FineTuningIntegration:
      type: object
      title: Fine-Tuning Job Integration
      required:
        - type
        - wandb
      properties:
        type:
          type: string
          description: The type of the integration being enabled for the fine-tuning job
          enum:
            - wandb
        wandb:
          type: object
          description: >
            The settings for your integration with Weights and Biases. This
            payload specifies the project that

            metrics will be sent to. Optionally, you can set an explicit display
            name for your run, add tags

            to your run, and set a default entity (team, username, etc) to be
            associated with your run.
          required:
            - project
          properties:
            project:
              description: |
                The name of the project that the new run will be created under.
              type: string
              example: my-wandb-project
            name:
              description: >
                A display name to set for the run. If not set, we will use the
                Job ID as the name.
              nullable: true
              type: string
            entity:
              description: >
                The entity to use for the run. This allows you to set the team
                or username of the WandB user that you would

                like associated with the run. If not set, the default entity for
                the registered WandB API key is used.
              nullable: true
              type: string
            tags:
              description: >
                A list of tags to be attached to the newly created run. These
                tags are passed through directly to WandB. Some

                default tags are generated by OpenAI: "openai/finetune",
                "openai/{base-model}", "openai/{ftjob-abcdef}".
              type: array
              items:
                type: string
                example: custom-tag
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