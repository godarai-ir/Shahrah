> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# List Provider Integrations

> Get provider integrations for a tenant with optional filtering by type, fqn or id. Pagination is available based on query parameters.



## OpenAPI

````yaml /openapi.json get /api/svc/v1/provider-integrations
openapi: 3.1.0
info:
  title: TrueFoundry API
  description: API for TrueFoundry
  version: 0.1.0
  contact: {}
servers:
  - url: https://{controlPlaneURL}
    variables:
      controlPlaneURL:
        default: app.truefoundry.com
        description: Control Plane URL
security: []
tags:
  - name: Agent
  - name: Applications
  - name: Apply
  - name: Artifacts
  - name: Audit Logs
  - name: Clusters
  - name: Jobs
  - name: Logs
  - name: MCP Servers
  - name: MLRepos
  - name: Model Deployments
  - name: Models
  - name: Personal Access Tokens
  - name: Prompts
  - name: Provider Integrations
  - name: SCIM v2
  - name: Secret Groups
  - name: Secrets
  - name: Teams
  - name: Traces
  - name: Users
  - name: Virtual Accounts
  - name: Workspaces
paths:
  /api/svc/v1/provider-integrations:
    get:
      tags:
        - Provider Integrations
      summary: List Provider Integrations
      description: >-
        Get provider integrations for a tenant with optional filtering by type,
        fqn or id. Pagination is available based on query parameters.
      operationId: ProviderIntegration.getProviderIntegrations
      parameters:
        - name: limit
          required: false
          in: query
          description: Number of items per page
          schema:
            minimum: 1
            maximum: 1000
            default: 100
            example: 10
            type: integer
        - name: offset
          required: false
          in: query
          description: Number of items to skip
          schema:
            minimum: 0
            default: 0
            example: 0
            type: integer
        - name: type
          required: true
          in: query
          description: Provider Integration Type
          schema:
            type: string
            enum:
              - blob-storage
              - secret-store
              - docker-registry
              - model
              - vcs
              - cluster
              - notification-channel
              - helm-repo
              - jwt-auth
              - mcp-server
              - guardrail-config
        - name: fqn
          required: false
          in: query
          description: Provider Integration FQN
          schema:
            type: string
        - name: id
          required: false
          in: query
          description: Provider Integration Id
          schema:
            type: string
      responses:
        '200':
          description: >-
            Returns Provider Integrations for a tenant And also the response
            includes paginated data.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListProviderIntegrationsResponse'
        '400':
          description: Bad Request. Cannot provide both fqn and id
      security:
        - jwt: []
components:
  schemas:
    ListProviderIntegrationsResponse:
      type: object
      properties:
        data:
          description: Array of Provider Integrations
          type: array
          items:
            $ref: '#/components/schemas/ProviderIntegration'
        pagination:
          description: Pagination information
          allOf:
            - $ref: '#/components/schemas/Pagination'
      required:
        - data
        - pagination
      title: ListProviderIntegrationsResponse
      x-fern-type-name: ListProviderIntegrationsResponse
    ProviderIntegration:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        fqn:
          type: string
        type:
          type: string
          enum:
            - blob-storage
            - secret-store
            - docker-registry
            - model
            - vcs
            - cluster
            - notification-channel
            - helm-repo
            - jwt-auth
            - mcp-server
            - guardrail-config
        manifest:
          type: object
          additionalProperties: true
        metadata:
          type: object
          additionalProperties: true
        providerAccountFqn:
          type: string
        providerAccount:
          $ref: '#/components/schemas/ProviderAccount'
        createdBySubject:
          $ref: '#/components/schemas/Subject'
        createdAt:
          format: date-time
          type: string
        updatedAt:
          format: date-time
          type: string
        createdBy:
          type: string
      required:
        - id
        - name
        - fqn
        - type
        - providerAccountFqn
        - createdBySubject
      title: ProviderIntegration
      x-fern-type-name: ProviderIntegration
    Pagination:
      type: object
      properties:
        total:
          type: integer
          description: Total number of items
          example: 100
        offset:
          type: integer
          description: Number of items skipped
          example: 0
        limit:
          type: integer
          description: Number of items per page
          example: 10
      required:
        - total
      title: Pagination
      x-fern-type-name: Pagination
    ProviderAccount:
      type: object
      properties:
        id:
          type: string
        fqn:
          type: string
        provider:
          type: string
          enum:
            - aws
            - gcp
            - azure
            - dockerhub
            - openai
            - databricks
            - truefoundry
            - jfrog
            - ttl
            - custom
            - quay
            - github
            - gitlab
            - bitbucket
            - ai21
            - cohere
            - anthropic
            - anyscale
            - deepinfra
            - groq
            - mistral-ai
            - nomic
            - palm
            - perplexity-ai
            - together-ai
            - ollama
            - slack
            - aws-bedrock
            - google-vertex
            - google-gemini
            - azure-openai
            - azure-foundry
            - self-hosted-model
            - mcp-server-group
            - pagerduty
            - model-provider-accounts
            - guardrail-config-group
            - openrouter
            - sambanova
            - aws-sagemaker
            - cerebras
            - hashicorp
            - snowflake-cortex
            - cloudera
            - xai
            - baseten
            - virtual-model
            - elevenlabs
            - deepgram
            - cartesia
            - custom-endpoint
        manifest:
          type: object
          additionalProperties: true
        integrations:
          description: List of Provider Integrations
          type: array
          items:
            $ref: '#/components/schemas/ProviderIntegration'
        createdBySubject:
          $ref: '#/components/schemas/Subject'
        accountId:
          type: string
          nullable: false
        createdAt:
          format: date-time
          type: string
        updatedAt:
          format: date-time
          type: string
        createdBy:
          type: string
      required:
        - fqn
        - provider
        - manifest
        - accountId
      title: ProviderAccount
      x-fern-type-name: ProviderAccount
    Subject:
      type: object
      properties:
        subjectId:
          type: string
          description: Subject ID
        subjectType:
          description: Subject type
          allOf:
            - $ref: '#/components/schemas/SubjectType'
        subjectSlug:
          type: string
          description: Subject slug
        subjectDisplayName:
          type: string
          description: Subject display name
        subjectPatName:
          type: string
          description: Subject PAT name
        subjectControllerName:
          type: string
        subjectExternalIdentitySlug:
          type: string
          description: External identity slug (external_identity_id:sub:email)
      required:
        - subjectId
        - subjectType
      title: Subject
      x-fern-type-name: Subject
    SubjectType:
      type: string
      enum:
        - user
        - team
        - serviceaccount
        - virtualaccount
        - external-identity
        - role
      title: SubjectType
      x-fern-type-name: SubjectType
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````