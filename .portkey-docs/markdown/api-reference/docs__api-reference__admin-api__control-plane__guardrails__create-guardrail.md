> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Create Guardrail

> Creates a new guardrail with specified checks and actions



## OpenAPI

````yaml post /guardrails
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
  /guardrails:
    post:
      tags:
        - Guardrails
      summary: Create a new guardrail
      description: Creates a new guardrail with specified checks and actions
      operationId: createGuardrail
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateGuardrailRequest'
            examples:
              jwt_authentication:
                summary: '[BASIC] JWT Token Validation'
                value:
                  name: JWT Authentication Guard
                  organisation_id: 550e8400-e29b-41d4-a716-446655440001
                  checks:
                    - id: default.jwt
                      parameters:
                        jwksUri: https://example.com/.well-known/jwks.json
                        headerKey: Authorization
                        algorithms:
                          - RS256
                        cacheMaxAge: 86400
                        clockTolerance: 5
                        maxTokenAge: 1d
                  actions:
                    onFail: block
                    message: Invalid JWT token
              model_whitelist:
                summary: '[BASIC] Model Whitelist Control'
                value:
                  name: Allowed Models Only
                  workspace_id: 550e8400-e29b-41d4-a716-446655440000
                  checks:
                    - id: default.modelWhitelist
                      parameters:
                        models:
                          - gpt-4
                          - gpt-3.5-turbo
                          - claude-3-sonnet
                          - claude-3-haiku
                  actions:
                    onFail: block
                    message: Model not in approved whitelist
              allowed_request_types:
                summary: '[BASIC] Restrict to Chat and Embed Only'
                value:
                  name: Restrict to Chat and Embed Only
                  checks:
                    - id: default.allowedRequestTypes
                      parameters:
                        allowedTypes:
                          - chatComplete
                          - embed
                  actions:
                    on_fail: deny
                    on_success: continue
              case_validation:
                summary: '[BASIC] Case Validation Checks'
                value:
                  name: Text Case Validation
                  workspace_id: 550e8400-e29b-41d4-a716-446655440000
                  checks:
                    - id: default.isAllLowerCase
                    - id: default.alluppercase
                      parameters:
                        not: true
                  actions:
                    onFail: log
                    message: Text case validation failed
              content_regex:
                summary: '[BASIC] Regex Pattern Matching'
                value:
                  name: Content Pattern Validation
                  workspace_id: 550e8400-e29b-41d4-a716-446655440000
                  checks:
                    - id: default.regexMatch
                      parameters:
                        rule: \b(inappropriate|banned|harmful|offensive)\b
                        not: false
                    - id: default.endsWith
                      parameters:
                        suffix: .
                        not: false
                  actions:
                    onFail: block
                    message: Content violates pattern rules
              length_controls:
                summary: '[BASIC] Content Length Controls'
                value:
                  name: Content Length Validation
                  workspace_id: 550e8400-e29b-41d4-a716-446655440000
                  checks:
                    - id: default.wordCount
                      parameters:
                        minWords: 5
                        maxWords: 500
                        not: false
                    - id: default.sentenceCount
                      parameters:
                        minSentences: 1
                        maxSentences: 20
                        not: false
                    - id: default.characterCount
                      parameters:
                        minCharacters: 10
                        maxCharacters: 4000
                        not: false
                  actions:
                    onFail: block
                    message: Content length out of bounds
              json_validation:
                summary: '[BASIC] JSON Structure Validation'
                value:
                  name: JSON Response Validation
                  workspace_id: 550e8400-e29b-41d4-a716-446655440000
                  checks:
                    - id: default.jsonSchema
                      parameters:
                        schema:
                          type: object
                          properties:
                            result:
                              type: string
                            confidence:
                              type: number
                              minimum: 0
                              maximum: 1
                            metadata:
                              type: object
                          required:
                            - result
                        not: false
                    - id: default.jsonKeys
                      parameters:
                        keys:
                          - result
                          - timestamp
                          - id
                        operator: all
                  actions:
                    onFail: block
                    message: Response does not match expected format
              content_analysis:
                summary: '[BASIC] Content Analysis Checks'
                value:
                  name: Content Quality Checks
                  workspace_id: 550e8400-e29b-41d4-a716-446655440000
                  checks:
                    - id: default.contains
                      parameters:
                        words:
                          - please
                          - thank you
                          - help
                        operator: any
                    - id: default.validUrls
                      parameters:
                        onlyDNS: true
                        not: false
                    - id: default.containsCode
                      parameters:
                        format: SQL
                        not: true
                  actions:
                    onFail: warn
                    message: Content quality check failed
              webhook_integration:
                summary: '[BASIC] Custom Webhook Validation'
                value:
                  name: External Validation Service
                  workspace_id: 550e8400-e29b-41d4-a716-446655440000
                  checks:
                    - id: default.webhook
                      parameters:
                        webhookURL: https://api.example.com/validate-content
                        headers:
                          Authorization: Bearer token123
                          Content-Type: application/json
                          X-API-Version: v1
                        timeout: 5000
                        failOnError: true
                  actions:
                    onFail: block
                    message: External validation failed
              metadata_validation:
                summary: '[BASIC] Required Metadata Keys'
                value:
                  name: Metadata Requirement Check
                  workspace_id: 550e8400-e29b-41d4-a716-446655440000
                  checks:
                    - id: default.requiredMetadataKeys
                      parameters:
                        metadataKeys:
                          - user_id
                          - session_id
                          - request_type
                        operator: all
                  actions:
                    onFail: block
                    message: Required metadata missing
              portkey_moderation:
                summary: '[PRO] OpenAI Content Moderation'
                value:
                  name: Advanced Content Moderation
                  workspace_id: 550e8400-e29b-41d4-a716-446655440000
                  checks:
                    - id: portkey.moderateContent
                      parameters:
                        categories:
                          - hate/threatening
                          - harassment/threatening
                          - self-harm/intent
                          - sexual/minors
                          - violence/graphic
                        timeout: 5000
                  actions:
                    onFail: block
                    message: Content flagged by moderation system
              portkey_language:
                summary: '[PRO] Language Detection & Validation'
                value:
                  name: Multi-Language Support
                  workspace_id: 550e8400-e29b-41d4-a716-446655440000
                  checks:
                    - id: portkey.language
                      parameters:
                        language: eng_Latn
                        not: false
                        timeout: 5000
                  actions:
                    onFail: block
                    message: Content not in expected language
              portkey_pii:
                summary: '[PRO] Advanced PII Detection'
                value:
                  name: PII Protection System
                  workspace_id: 550e8400-e29b-41d4-a716-446655440000
                  checks:
                    - id: portkey.pii
                      parameters:
                        redact: true
                        categories:
                          - EMAIL_ADDRESS
                          - PHONE_NUMBER
                          - SSN
                          - CREDIT_CARD
                          - NAME
                        timeout: 5000
                  actions:
                    onFail: block
                    message: PII detected and redacted
              portkey_gibberish:
                summary: '[PRO] Gibberish Detection'
                value:
                  name: Content Quality Filter
                  workspace_id: 550e8400-e29b-41d4-a716-446655440000
                  checks:
                    - id: portkey.gibberish
                      parameters:
                        timeout: 5000
                  actions:
                    onFail: block
                    message: Content appears to be gibberish
              sydelabs_security:
                summary: '[PARTNER] SydeLabs AI Security'
                value:
                  name: AI Security Suite
                  workspace_id: 550e8400-e29b-41d4-a716-446655440000
                  checks:
                    - id: sydelabs.sydeguard
                      parameters:
                        prompt_injection_threshold: 0.5
                        toxicity_threshold: 0.3
                        evasion_threshold: 0.6
                        timeout: 5000
                  actions:
                    onFail: block
                    message: AI security check failed
              aporia_validation:
                summary: '[PARTNER] Aporia Project Validation'
                value:
                  name: Aporia Policy Enforcement
                  workspace_id: 550e8400-e29b-41d4-a716-446655440000
                  checks:
                    - id: aporia.validateProject
                      parameters:
                        projectID: proj_abc123
                        timeout: 5000
                  actions:
                    onFail: block
                    message: Aporia validation failed
              pillar_scanning:
                summary: '[PARTNER] Pillar Security Scanning'
                value:
                  name: Comprehensive Security Scan
                  workspace_id: 550e8400-e29b-41d4-a716-446655440000
                  checks:
                    - id: pillar.scanPrompt
                      parameters:
                        scanners:
                          - prompt_injection
                          - pii
                          - secrets
                          - toxic_language
                          - invisible_characters
                        timeout: 5000
                    - id: pillar.scanResponse
                      parameters:
                        scanners:
                          - pii
                          - secrets
                          - toxic_language
                        timeout: 5000
                  actions:
                    onFail: block
                    message: Security scan detected issues
              patronus_comprehensive:
                summary: '[PARTNER] Patronus AI Complete Suite'
                value:
                  name: Patronus Content Analysis
                  workspace_id: 550e8400-e29b-41d4-a716-446655440000
                  checks:
                    - id: patronus.pii
                      parameters:
                        redact: true
                        timeout: 5000
                    - id: patronus.toxicity
                      parameters:
                        timeout: 5000
                    - id: patronus.noGenderBias
                      parameters:
                        timeout: 15000
                    - id: patronus.isHelpful
                      parameters:
                        timeout: 15000
                    - id: patronus.custom
                      parameters:
                        profile: system:is-concise
                        timeout: 15000
                  actions:
                    onFail: block
                    message: Content failed Patronus analysis
              azure_content_safety:
                summary: '[PARTNER] Azure Content Safety Suite'
                value:
                  name: Microsoft Azure Safety
                  workspace_id: 550e8400-e29b-41d4-a716-446655440000
                  checks:
                    - id: azure.contentSafety
                      parameters:
                        blocklistNames:
                          - company_blocklist
                        apiVersion: '2024-09-01'
                        severity: 2
                        categories:
                          - Hate
                          - SelfHarm
                          - Sexual
                          - Violence
                        timeout: 5000
                    - id: azure.pii
                      parameters:
                        domain: phi
                        apiVersion: '2024-11-01'
                        modelVersion: latest
                        redact: true
                        timeout: 5000
                  actions:
                    onFail: block
                    message: Azure safety checks failed
              mistral_moderation:
                summary: '[PARTNER] Mistral Content Moderation'
                value:
                  name: Mistral AI Moderation
                  workspace_id: 550e8400-e29b-41d4-a716-446655440000
                  checks:
                    - id: mistral.moderateContent
                      parameters:
                        categories:
                          - sexual
                          - hate_and_discrimination
                          - violence_and_threats
                          - selfharm
                          - pii
                        timeout: 5000
                  actions:
                    onFail: block
                    message: Mistral moderation flagged content
              bedrock_enterprise:
                summary: '[PARTNER] AWS Bedrock Guardrails'
                value:
                  name: Enterprise AWS Security
                  workspace_id: 550e8400-e29b-41d4-a716-446655440000
                  checks:
                    - id: bedrock.guard
                      parameters:
                        guardrailVersion: DRAFT
                        guardrailId: gdrail123abc
                        redact: true
                        timeout: 5000
                  actions:
                    onFail: block
                    message: AWS Bedrock guardrail violation
              promptfoo_testing:
                summary: '[PARTNER] Promptfoo Security Testing'
                value:
                  name: Security Testing Suite
                  workspace_id: 550e8400-e29b-41d4-a716-446655440000
                  checks:
                    - id: promptfoo.guard
                      parameters:
                        timeout: 5000
                    - id: promptfoo.pii
                      parameters:
                        redact: true
                        timeout: 5000
                    - id: promptfoo.harm
                      parameters:
                        timeout: 5000
                  actions:
                    onFail: block
                    message: Promptfoo security tests failed
              acuvity_comprehensive:
                summary: '[PARTNER] Acuvity Multi-Vector Security'
                value:
                  name: Complete Security Analysis
                  workspace_id: 550e8400-e29b-41d4-a716-446655440000
                  checks:
                    - id: acuvity.scan
                      parameters:
                        prompt_injection: true
                        prompt_injection_threshold: 0.5
                        toxic: true
                        toxic_threshold: 0.3
                        jail_break: true
                        jail_break_threshold: 0.6
                        malicious_url: true
                        biased: true
                        harmful: true
                        language: true
                        language_values: english
                        pii: true
                        pii_redact: true
                        pii_categories:
                          - email_address
                          - ssn
                          - credit_card
                        secrets: true
                        secrets_redact: true
                        secrets_categories:
                          - aws_secret_key
                          - openai
                          - github
                        timeout: 5000
                  actions:
                    onFail: block
                    message: Comprehensive security scan failed
              lasso_classification:
                summary: '[PARTNER] Lasso Security Classification'
                value:
                  name: Content Classification
                  workspace_id: 550e8400-e29b-41d4-a716-446655440000
                  checks:
                    - id: lasso.classify
                      parameters:
                        timeout: 5000
                  actions:
                    onFail: block
                    message: Lasso classification failed
              panw_prisma:
                summary: '[PARTNER] PANW Prisma AIRS Enterprise'
                value:
                  name: Enterprise Security Runtime
                  workspace_id: 550e8400-e29b-41d4-a716-446655440000
                  checks:
                    - id: panw-prisma-airs.intercept
                      parameters:
                        profile_name: enterprise_profile
                        ai_model: gpt-4
                        app_user: api_user_123
                  actions:
                    onFail: block
                    message: Prisma AIRS blocked request
      responses:
        '200':
          description: Guardrail created successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CreateGuardrailResponse'
        '400':
          description: Bad request - validation failed
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '403':
          description: Forbidden - insufficient permissions or guardrail not allowed
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '500':
          description: Internal server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
components:
  schemas:
    CreateGuardrailRequest:
      type: object
      required:
        - name
        - checks
        - actions
      properties:
        name:
          type: string
          description: Name of the guardrail
          example: Content Safety Check
        workspace_id:
          type: string
          format: uuid
          description: >-
            Workspace UUID (required if organisation_id not provided and not
            using API key)
        organisation_id:
          type: string
          format: uuid
          description: >-
            Organisation UUID (required if workspace_id not provided and not
            using API key)
        checks:
          type: array
          description: Array of guardrail checks to apply
          items:
            $ref: '#/components/schemas/GuardrailCheck'
          minItems: 1
        actions:
          $ref: '#/components/schemas/GuardrailActions'
    CreateGuardrailResponse:
      type: object
      required:
        - id
        - slug
        - version_id
      properties:
        id:
          type: string
          description: Unique identifier of the created guardrail
        slug:
          type: string
          description: URL-friendly slug for the guardrail
        version_id:
          type: string
          description: Version identifier for the guardrail configuration
    ErrorResponse:
      type: object
      properties:
        error:
          $ref: '#/components/schemas/Error'
      required:
        - error
    GuardrailCheck:
      type: object
      required:
        - id
      properties:
        id:
          type: string
          description: Identifier of the guardrail check type
          enum:
            - default.jwt
            - default.modelWhitelist
            - default.isAllLowerCase
            - default.regexMatch
            - default.sentenceCount
            - default.wordCount
            - default.characterCount
            - default.jsonSchema
            - default.jsonKeys
            - default.contains
            - default.validUrls
            - default.containsCode
            - default.webhook
            - default.endsWith
            - default.alluppercase
            - default.requiredMetadataKeys
            - default.allowedRequestTypes
            - portkey.moderateContent
            - portkey.language
            - portkey.pii
            - portkey.gibberish
            - sydelabs.sydeguard
            - aporia.validateProject
            - pillar.scanPrompt
            - pillar.scanResponse
            - patronus.phi
            - patronus.pii
            - patronus.isConcise
            - patronus.isHelpful
            - patronus.isPolite
            - patronus.noApologies
            - patronus.noGenderBias
            - patronus.noRacialBias
            - patronus.retrievalAnswerRelevance
            - patronus.toxicity
            - patronus.custom
            - mistral.moderateContent
            - bedrock.guard
            - promptfoo.guard
            - promptfoo.pii
            - promptfoo.harm
            - acuvity.scan
            - lasso.classify
            - azure.contentSafety
            - azure.pii
            - panw-prisma-airs.intercept
        parameters:
          oneOf:
            - $ref: '#/components/schemas/JWTParameters'
            - $ref: '#/components/schemas/ModelWhitelistParameters'
            - $ref: '#/components/schemas/RegexMatchParameters'
            - $ref: '#/components/schemas/SentenceCountParameters'
            - $ref: '#/components/schemas/WordCountParameters'
            - $ref: '#/components/schemas/CharacterCountParameters'
            - $ref: '#/components/schemas/JSONSchemaParameters'
            - $ref: '#/components/schemas/JSONKeysParameters'
            - $ref: '#/components/schemas/ContainsParameters'
            - $ref: '#/components/schemas/ValidUrlsParameters'
            - $ref: '#/components/schemas/ContainsCodeParameters'
            - $ref: '#/components/schemas/WebhookParameters'
            - $ref: '#/components/schemas/EndsWithParameters'
            - $ref: '#/components/schemas/UppercaseParameters'
            - $ref: '#/components/schemas/RequiredMetadataKeysParameters'
            - $ref: '#/components/schemas/AllowedRequestTypesParameters'
            - $ref: '#/components/schemas/SydeGuardParameters'
            - $ref: '#/components/schemas/AporiaParameters'
            - $ref: '#/components/schemas/PillarScanParameters'
            - $ref: '#/components/schemas/PatronusParameters'
            - $ref: '#/components/schemas/PatronusCustomParameters'
            - $ref: '#/components/schemas/PortkeyModerationParameters'
            - $ref: '#/components/schemas/PortkeyLanguageParameters'
            - $ref: '#/components/schemas/PortkeyPIIParameters'
            - $ref: '#/components/schemas/MistralModerationParameters'
            - $ref: '#/components/schemas/BedrockGuardParameters'
            - $ref: '#/components/schemas/PromptfooParameters'
            - $ref: '#/components/schemas/AcuvityScanParameters'
            - $ref: '#/components/schemas/AzureContentSafetyParameters'
            - $ref: '#/components/schemas/AzurePIIParameters'
            - $ref: '#/components/schemas/PANWPrismaParameters'
            - $ref: '#/components/schemas/BasicParameters'
          description: Configuration parameters specific to the check type
        name:
          type: string
          description: Custom name for this specific check instance
        is_enabled:
          type: boolean
          description: Whether this check is enabled
          default: true
    GuardrailActions:
      type: object
      description: Actions to take when guardrail checks fail or pass
      properties:
        deny:
          type: boolean
          description: Whether to deny the request when guardrail check fails
          default: false
        async:
          type: boolean
          description: Whether the guardrail check should be performed asynchronously
          default: false
        on_success:
          type: object
          description: Actions to take when guardrail check passes
          properties:
            feedback:
              type: object
              description: Feedback configuration for successful checks
              properties:
                value:
                  type: number
                  description: Feedback value for successful checks
                  default: 5
                weight:
                  type: number
                  description: Weight of the feedback
                  default: 1
                metadata:
                  type: string
                  description: Additional metadata for the feedback
                  default: ''
              required:
                - value
                - weight
                - metadata
          required:
            - feedback
        on_fail:
          type: object
          description: Actions to take when guardrail check fails
          properties:
            feedback:
              type: object
              description: Feedback configuration for failed checks
              properties:
                value:
                  type: number
                  description: Feedback value for failed checks
                  default: -5
                weight:
                  type: number
                  description: Weight of the feedback
                  default: 1
                metadata:
                  type: string
                  description: Additional metadata for the feedback
                  default: ''
              required:
                - value
                - weight
                - metadata
          required:
            - feedback
      required:
        - deny
        - async
        - on_success
        - on_fail
    Error:
      type: object
      properties:
        code:
          type: string
          nullable: true
        message:
          type: string
          nullable: false
        param:
          type: string
          nullable: true
        type:
          type: string
          nullable: false
      required:
        - type
        - message
        - param
        - code
    JWTParameters:
      title: JWT Parameters
      type: object
      required:
        - jwksUri
        - headerKey
      properties:
        jwksUri:
          type: string
          format: uri
          description: JWKS URI of the JWT token
        headerKey:
          type: string
          description: Header key to check for the JWT token
        cacheMaxAge:
          type: number
          description: Cache max age in seconds
          default: 86400
        clockTolerance:
          type: number
          description: Clock tolerance in seconds
          default: 5
        maxTokenAge:
          type: string
          description: Max token age
          default: 1d
        algorithms:
          type: array
          items:
            type: string
          description: Algorithms to check for the JWT token
          default:
            - RS256
    ModelWhitelistParameters:
      title: Model Whitelist Parameters
      type: object
      required:
        - models
      properties:
        models:
          type: array
          items:
            type: string
          description: List of allowed models
    RegexMatchParameters:
      title: Regex Match Parameters
      type: object
      required:
        - rule
      properties:
        rule:
          type: string
          description: Regex pattern to match
        not:
          type: boolean
          description: If true, the check will fail when the regex pattern matches
          default: false
    SentenceCountParameters:
      title: Sentence Count Parameters
      type: object
      properties:
        minSentences:
          type: number
          description: Minimum number of sentences to allow
          default: 0
        maxSentences:
          type: number
          description: Maximum number of sentences to allow
          default: 99999
        not:
          type: boolean
          description: If true, the verdict will be inverted
          default: false
    WordCountParameters:
      title: Word Count Parameters
      type: object
      properties:
        minWords:
          type: number
          description: Minimum number of words to allow
          default: 0
        maxWords:
          type: number
          description: Maximum number of words to allow
          default: 99999
        not:
          type: boolean
          description: If true, the verdict will be inverted
          default: false
    CharacterCountParameters:
      title: Character Count Parameters
      type: object
      properties:
        minCharacters:
          type: number
          description: Minimum number of characters to allow
          default: 0
        maxCharacters:
          type: number
          description: Maximum number of characters to allow
          default: 9999999
        not:
          type: boolean
          description: If true, the verdict will be inverted
          default: false
    JSONSchemaParameters:
      title: JSON Schema Parameters
      type: object
      required:
        - schema
      properties:
        schema:
          type: object
          additionalProperties: true
          description: JSON schema to validate against
        not:
          type: boolean
          description: If true, the verdict will be inverted
          default: false
    JSONKeysParameters:
      title: JSON Keys Parameters
      type: object
      required:
        - keys
        - operator
      properties:
        keys:
          type: array
          items:
            type: string
          description: Keys to check for in JSON
        operator:
          type: string
          enum:
            - any
            - all
            - none
          description: Operator to use for key checking
          default: any
    ContainsParameters:
      title: Contains Parameters
      type: object
      required:
        - words
        - operator
      properties:
        words:
          type: array
          items:
            type: string
          description: Words or phrases to check for
        operator:
          type: string
          enum:
            - any
            - all
            - none
          description: Operator to use for word checking
          default: any
    ValidUrlsParameters:
      title: Valid URLs Parameters
      type: object
      properties:
        onlyDNS:
          type: boolean
          description: Only check if URL domains resolve (10x faster)
          default: false
        not:
          type: boolean
          description: If true, the verdict will be inverted
          default: false
    ContainsCodeParameters:
      title: Contains Code Parameters
      type: object
      required:
        - format
      properties:
        format:
          type: string
          enum:
            - SQL
            - Python
            - TypeScript
            - JavaScript
            - Java
            - C#
            - C++
            - C
            - Ruby
            - PHP
            - Swift
            - Kotlin
            - Go
            - Rust
            - Scala
            - R
            - Perl
            - Shell
            - HTML
            - CSS
            - XML
            - JSON
            - YAML
            - Markdown
            - Dockerfile
          description: Code format to check for
        not:
          type: boolean
          description: If true, the verdict will be inverted
          default: false
    WebhookParameters:
      title: Webhook Parameters
      type: object
      required:
        - webhookURL
      properties:
        webhookURL:
          type: string
          format: uri
          description: Webhook URL to call
        headers:
          type: object
          additionalProperties: true
          description: Headers to send with the request
        timeout:
          type: number
          description: Timeout in milliseconds
          default: 3000
        failOnError:
          type: boolean
          description: Fail if webhook returns non-200 status or times out
          default: false
    EndsWithParameters:
      title: Ends With Parameters
      type: object
      required:
        - suffix
      properties:
        suffix:
          type: string
          description: Suffix to check for
        not:
          type: boolean
          description: If true, the verdict will be inverted
          default: false
    UppercaseParameters:
      title: Uppercase Parameters
      type: object
      properties:
        not:
          type: boolean
          description: If true, the verdict will be inverted
          default: false
    RequiredMetadataKeysParameters:
      title: Required Metadata Keys Parameters
      type: object
      required:
        - metadataKeys
        - operator
      properties:
        metadataKeys:
          type: array
          items:
            type: string
          description: Metadata keys to check for
        operator:
          type: string
          enum:
            - all
            - any
            - none
          description: Operator to use for key checking
          default: all
    AllowedRequestTypesParameters:
      title: Allowed Request Types Parameters
      type: object
      description: >-
        Parameters for default.allowedRequestTypes check. Restrict which request
        types are allowed or blocked. You can use either or both; when both are
        specified, blocked types take precedence.
      properties:
        allowedTypes:
          type: array
          items:
            type: string
            enum:
              - complete
              - chatComplete
              - embed
              - rerank
              - moderate
              - proxy
              - imageGenerate
              - createSpeech
              - createTranscription
              - createTranslation
              - realtime
              - uploadFile
              - listFiles
              - retrieveFile
              - deleteFile
              - retrieveFileContent
              - createBatch
              - retrieveBatch
              - cancelBatch
              - listBatches
              - getBatchOutput
              - listFinetunes
              - createFinetune
              - retrieveFinetune
              - cancelFinetune
              - createModelResponse
              - getModelResponse
              - deleteModelResponse
              - listResponseInputItems
              - messages
          description: >-
            Request types to allow (allowlist). When set, only these request
            types are permitted.
        blockedTypes:
          type: array
          items:
            type: string
            enum:
              - complete
              - chatComplete
              - embed
              - rerank
              - moderate
              - proxy
              - imageGenerate
              - createSpeech
              - createTranscription
              - createTranslation
              - realtime
              - uploadFile
              - listFiles
              - retrieveFile
              - deleteFile
              - retrieveFileContent
              - createBatch
              - retrieveBatch
              - cancelBatch
              - listBatches
              - getBatchOutput
              - listFinetunes
              - createFinetune
              - retrieveFinetune
              - cancelFinetune
              - createModelResponse
              - getModelResponse
              - deleteModelResponse
              - listResponseInputItems
              - messages
          description: >-
            Request types to block (blocklist). When set, these request types
            are denied.
    SydeGuardParameters:
      title: SydeGuard Parameters
      type: object
      properties:
        prompt_injection_threshold:
          type: number
          minimum: 0
          maximum: 1
          multipleOf: 0.01
          description: Threshold for prompt injection risk score (0-1)
          default: 0.5
        toxicity_threshold:
          type: number
          minimum: 0
          maximum: 1
          multipleOf: 0.01
          description: Threshold for toxicity risk score (0-1)
          default: 0.5
        evasion_threshold:
          type: number
          minimum: 0
          maximum: 1
          multipleOf: 0.01
          description: Threshold for evasion risk score (0-1)
          default: 0.5
        timeout:
          type: number
          description: Timeout in milliseconds
          default: 5000
    AporiaParameters:
      title: Aporia Parameters
      type: object
      required:
        - projectID
      properties:
        projectID:
          type: string
          description: Aporia Project ID to validate
        timeout:
          type: number
          description: Timeout in milliseconds
          default: 5000
    PillarScanParameters:
      title: Pillar Scan Parameters
      type: object
      required:
        - scanners
      properties:
        scanners:
          type: array
          items:
            type: string
            enum:
              - prompt_injection
              - pii
              - secrets
              - toxic_language
              - invisible_characters
          description: Scanners to use for content analysis
        timeout:
          type: number
          description: Timeout in milliseconds
          default: 5000
    PatronusParameters:
      title: Patronus Parameters
      type: object
      properties:
        redact:
          type: boolean
          description: Whether to redact detected content
          default: false
        timeout:
          type: number
          description: Timeout in milliseconds
          default: 5000
    PatronusCustomParameters:
      title: Patronus Custom Parameters
      type: object
      required:
        - profile
      properties:
        profile:
          type: string
          description: Custom evaluator profile name (e.g., system:is-concise)
        timeout:
          type: number
          description: Timeout in milliseconds
          default: 15000
    PortkeyModerationParameters:
      title: Portkey Moderation Parameters
      type: object
      required:
        - categories
      properties:
        categories:
          type: array
          items:
            type: string
            enum:
              - hate
              - hate/threatening
              - harassment
              - harassment/threatening
              - self-harm
              - self-harm/intent
              - self-harm/instructions
              - sexual
              - sexual/minors
              - violence
              - violence/graphic
          description: Categories that should NOT be allowed
        timeout:
          type: number
          description: Timeout in milliseconds
          default: 5000
    PortkeyLanguageParameters:
      title: Portkey Language Parameters
      type: object
      properties:
        language:
          type: string
          enum:
            - eng_Latn
            - zho_Hans
            - spa_Latn
            - ara_Arab
            - por_Latn
            - ind_Latn
            - fra_Latn
            - jpn_Jpan
            - rus_Cyrl
            - deu_Latn
            - kor_Hang
            - tur_Latn
            - ita_Latn
            - pes_Arab
            - pol_Latn
            - vie_Latn
            - nld_Latn
            - hin_Deva
            - tha_Thai
            - heb_Hebr
            - ben_Beng
            - swe_Latn
            - ces_Latn
            - ron_Latn
            - ell_Grek
            - ukr_Cyrl
            - dan_Latn
            - fin_Latn
            - nor_Latn
            - hun_Latn
            - cat_Latn
            - bul_Cyrl
            - msa_Latn
            - hrv_Latn
            - arb_Latn
            - slk_Latn
            - lit_Latn
            - lav_Latn
            - srp_Cyrl
            - slv_Latn
            - est_Latn
            - urd_Arab
            - fil_Latn
            - aze_Latn
            - tam_Taml
            - tel_Telu
            - mar_Deva
            - kan_Knda
            - fas_Arab
          description: Language that should be allowed in content
        not:
          type: boolean
          description: If true, the verdict will be inverted
          default: false
        timeout:
          type: number
          description: Timeout in milliseconds
          default: 5000
    PortkeyPIIParameters:
      title: Portkey PII Parameters
      type: object
      required:
        - categories
      properties:
        redact:
          type: boolean
          description: Whether to redact detected PII
          default: false
        categories:
          type: array
          items:
            type: string
            enum:
              - EMAIL_ADDRESS
              - PHONE_NUMBER
              - LOCATION_ADDRESS
              - NAME
              - IP_ADDRESS
              - CREDIT_CARD
              - SSN
          description: Types of PII that should NOT be allowed
        timeout:
          type: number
          description: Timeout in milliseconds
          default: 5000
    MistralModerationParameters:
      title: Mistral Moderation Parameters
      type: object
      required:
        - categories
      properties:
        categories:
          type: array
          items:
            type: string
            enum:
              - sexual
              - hate_and_discrimination
              - violence_and_threats
              - dangerous_and_criminal_content
              - selfharm
              - health
              - financial
              - law
              - pii
          description: Categories that should NOT be allowed
        timeout:
          type: number
          description: Timeout in milliseconds
          default: 5000
    BedrockGuardParameters:
      title: Bedrock Guard Parameters
      type: object
      required:
        - guardrailVersion
        - guardrailId
      properties:
        guardrailVersion:
          type: string
          description: Version of the guardrail to use
        guardrailId:
          type: string
          description: ID of the guardrail
        redact:
          type: boolean
          description: Whether to redact detected PII
          default: false
        timeout:
          type: number
          description: Timeout in milliseconds
          default: 5000
    PromptfooParameters:
      title: Promptfoo Parameters
      type: object
      properties:
        redact:
          type: boolean
          description: Whether to redact detected content
          default: false
        timeout:
          type: number
          description: Timeout in milliseconds
          default: 5000
    AcuvityScanParameters:
      title: Acuvity Scan Parameters
      type: object
      properties:
        prompt_injection:
          type: boolean
          description: Enable prompt injection detection
          default: true
        prompt_injection_threshold:
          type: number
          minimum: 0
          maximum: 1
          multipleOf: 0.01
          description: Threshold for prompt injection detection
          default: 0.5
        toxic:
          type: boolean
          description: Enable toxicity detection
          default: true
        toxic_threshold:
          type: number
          minimum: 0
          maximum: 1
          multipleOf: 0.01
          description: Threshold for toxicity detection
          default: 0.5
        jail_break:
          type: boolean
          description: Enable jailbreak detection
          default: true
        jail_break_threshold:
          type: number
          minimum: 0
          maximum: 1
          multipleOf: 0.01
          description: Threshold for jailbreak detection
          default: 0.5
        malicious_url:
          type: boolean
          description: Enable malicious URL detection
          default: true
        malicious_url_threshold:
          type: number
          minimum: 0
          maximum: 1
          multipleOf: 0.01
          description: Threshold for malicious URL detection
          default: 0.5
        biased:
          type: boolean
          description: Enable bias detection
          default: true
        biased_threshold:
          type: number
          minimum: 0
          maximum: 1
          multipleOf: 0.01
          description: Threshold for bias detection
          default: 0.5
        harmful:
          type: boolean
          description: Enable harmful content detection
          default: true
        harmful_threshold:
          type: number
          minimum: 0
          maximum: 1
          multipleOf: 0.01
          description: Threshold for harmful content detection
          default: 0.5
        language:
          type: boolean
          description: Enable language check
          default: true
        language_values:
          type: string
          enum:
            - english
            - chinese
            - spanish
            - french
            - german
            - japanese
            - gibberish
          description: Language to check
          default: english
        pii:
          type: boolean
          description: Enable PII detection
          default: true
        pii_redact:
          type: boolean
          description: Enable PII redaction
          default: false
        pii_categories:
          type: array
          items:
            type: string
            enum:
              - email_address
              - ssn
              - person
              - aba_routing_number
              - address
              - bank_account
              - bitcoin_wallet
              - credit_card
              - driver_license
              - itin_number
              - location
              - medical_license
              - money_amount
              - passport_number
              - phone_number
          description: PII categories to detect
        secrets:
          type: boolean
          description: Enable secrets detection
          default: true
        secrets_redact:
          type: boolean
          description: Enable secrets redaction
          default: false
        secrets_categories:
          type: array
          items:
            type: string
            enum:
              - credentials
              - aws_secret_key
              - github
              - openai
              - stripe
              - jwt
              - private_key
          description: Secret categories to detect
        timeout:
          type: number
          description: Timeout in milliseconds
          default: 5000
    AzureContentSafetyParameters:
      title: Azure Content Safety Parameters
      type: object
      properties:
        blocklistNames:
          type: array
          items:
            type: string
          description: Array of blocklist names to check against
          default: []
        apiVersion:
          type: string
          description: API version for the Content Safety API
          default: '2024-09-01'
        severity:
          type: number
          description: Severity threshold for the Content Safety API
          default: 2
        categories:
          type: array
          items:
            type: string
            enum:
              - Hate
              - SelfHarm
              - Sexual
              - Violence
          description: Categories to check against
        timeout:
          type: number
          description: Timeout in milliseconds
          default: 5000
    AzurePIIParameters:
      title: Azure PII Parameters
      type: object
      properties:
        domain:
          type: string
          enum:
            - none
            - phi
          description: Domain to check against
          default: none
        apiVersion:
          type: string
          description: API version for the Content Safety API
          default: '2024-11-01'
        modelVersion:
          type: string
          description: Version of the PII detection model to use
          default: latest
        redact:
          type: boolean
          description: Whether to redact detected PII
          default: false
        timeout:
          type: number
          description: Timeout in milliseconds
          default: 5000
    PANWPrismaParameters:
      title: PANW Prisma Parameters
      type: object
      required:
        - profile_name
      properties:
        profile_name:
          type: string
          description: Prisma profile name
        ai_model:
          type: string
          description: AI model identifier
        app_user:
          type: string
          description: Application user identifier
    BasicParameters:
      title: Basic Parameters
      type: object
      description: Basic parameters with no specific requirements
      additionalProperties: true
  securitySchemes:
    Portkey-Key:
      type: apiKey
      in: header
      name: x-portkey-api-key

````