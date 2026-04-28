> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Check user registration

> Check if a user is registered with the platform



## OpenAPI

````yaml /openapi.json post /api/svc/v1/users/is-registered
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
  /api/svc/v1/users/is-registered:
    post:
      tags:
        - Users
      summary: Check user registration
      description: Check if a user is registered with the platform
      operationId: User.checkUserRegistered
      parameters: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CheckUserRegisteredDto'
      responses:
        '201':
          description: ''
      security:
        - jwt: []
components:
  schemas:
    CheckUserRegisteredDto:
      type: object
      properties:
        email:
          type: string
          description: Email of the user to check
      required:
        - email
      title: CheckUserRegisteredDto
      x-fern-type-name: CheckUserRegisteredDto
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````