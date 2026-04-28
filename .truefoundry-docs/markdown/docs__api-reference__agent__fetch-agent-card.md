> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Fetch Agent Card

> Fetch agent card JSON from a source URL.



## OpenAPI

````yaml /openapi.json post /api/svc/v1/agents/proxy/agent-card
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
  /api/svc/v1/agents/proxy/agent-card:
    post:
      tags:
        - Agent
      summary: Fetch Agent Card
      description: Fetch agent card JSON from a source URL.
      operationId: Agent.fetchAgentCardProxy
      parameters: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/FetchAgentCardDto'
      responses:
        '200':
          description: Agent card fetched successfully
      security:
        - jwt: []
components:
  schemas:
    FetchAgentCardDto:
      type: object
      properties:
        sourceURL:
          type: string
          description: URL to fetch agent card JSON from
          example: https://example.com/.well-known/agent-card.json
      required:
        - sourceURL
      title: FetchAgentCardDto
      x-fern-type-name: FetchAgentCardDto
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````