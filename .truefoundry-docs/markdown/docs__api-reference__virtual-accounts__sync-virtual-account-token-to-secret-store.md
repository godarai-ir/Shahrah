> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Sync virtual account token to secret store

> Syncs the virtual account token to the configured secret store. Returns the updated JWT with sync metadata including timestamp and error (if any).



## OpenAPI

````yaml /openapi.json post /api/svc/v1/virtual-accounts/{id}/sync-to-secret-store
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
  /api/svc/v1/virtual-accounts/{id}/sync-to-secret-store:
    post:
      tags:
        - Virtual Accounts
      summary: Sync virtual account token to secret store
      description: >-
        Syncs the virtual account token to the configured secret store. Returns
        the updated JWT with sync metadata including timestamp and error (if
        any).
      operationId: VirtualAccount.syncVirtualAccountTokenInSecretStore
      parameters:
        - name: id
          required: true
          in: path
          description: serviceaccount id
          schema:
            type: string
      responses:
        '200':
          description: Token synced successfully to secret store
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SyncVirtualAccountTokenResponse'
        '400':
          description: Virtual account does not have sync to secret manager enabled
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '404':
          description: Virtual account not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
      security:
        - jwt: []
components:
  schemas:
    SyncVirtualAccountTokenResponse:
      type: object
      properties:
        syncTokenInSecretStoreInfo:
          description: Sync status including timestamp and error (if any)
          allOf:
            - $ref: '#/components/schemas/SyncTokenInSecretStoreInfo'
      required:
        - syncTokenInSecretStoreInfo
      title: SyncVirtualAccountTokenResponse
      x-fern-type-name: SyncVirtualAccountTokenResponse
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
    SyncTokenInSecretStoreInfo:
      type: object
      properties:
        updatedAt:
          type: string
          description: ISO timestamp of when the token was synced
        error:
          type: string
          description: Error message if sync failed
      required:
        - updatedAt
      title: SyncTokenInSecretStoreInfo
      x-fern-type-name: SyncTokenInSecretStoreInfo
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````