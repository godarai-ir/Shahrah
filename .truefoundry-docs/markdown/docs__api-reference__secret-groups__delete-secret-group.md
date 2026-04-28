> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Delete Secret Group

> Deletes the secret group, its associated secrets and secret versions of those secrets.



## OpenAPI

````yaml /openapi.json delete /api/svc/v1/secret-groups/{id}
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
  /api/svc/v1/secret-groups/{id}:
    delete:
      tags:
        - Secret Groups
      summary: Delete Secret Group
      description: >-
        Deletes the secret group, its associated secrets and secret versions of
        those secrets.
      operationId: SecretGroup.delete
      parameters:
        - name: id
          required: true
          in: path
          description: Secret Id of the secret group.
          schema:
            type: string
      responses:
        '200':
          description: Deletes Secret Group.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DeleteSecretGroupResponse'
        '403':
          description: Forbidden. User does not have permission to delete the secret group.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '404':
          description: Not Found. Secret Group with the provided id is not found.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
      security:
        - jwt: []
components:
  schemas:
    DeleteSecretGroupResponse:
      type: object
      properties: {}
      title: DeleteSecretGroupResponse
      x-fern-type-name: DeleteSecretGroupResponse
    HttpError:
      type: object
      properties:
        statusCode:
          type: integer
          description: HTTP Status Code
        message:
          type: string
          description: Error Message
        code:
          oneOf:
            - type: integer
            - type: string
          description: Error code
        details:
          type: array
          description: Error details
          items:
            type: object
      required:
        - statusCode
        - message
      title: HttpError
      x-fern-type-name: HttpError
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````