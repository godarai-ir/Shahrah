> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Get Secret

> Get Secret associated with provided id. The secret value is not returned if the control plane has `DISABLE_SECRET_VALUE_VIEW` set



## OpenAPI

````yaml /openapi.json get /api/svc/v1/secrets/{id}
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
  /api/svc/v1/secrets/{id}:
    get:
      tags:
        - Secrets
      summary: Get Secret
      description: >-
        Get Secret associated with provided id. The secret value is not returned
        if the control plane has `DISABLE_SECRET_VALUE_VIEW` set
      operationId: Secret.getSecretById
      parameters:
        - name: id
          required: true
          in: path
          description: Secret Id of the secret.
          schema:
            type: string
      responses:
        '200':
          description: Returns the Secret associated with provided id
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetSecretResponse'
        '403':
          description: Forbidden. User does not have permission to read the secret.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '404':
          description: Not Found. Secret with the provided id is not found.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
      security:
        - jwt: []
components:
  schemas:
    GetSecretResponse:
      type: object
      properties:
        data:
          description: Secret
          allOf:
            - $ref: '#/components/schemas/Secret'
      required:
        - data
      title: GetSecretResponse
      x-fern-type-name: GetSecretResponse
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
    Secret:
      type: object
      properties:
        id:
          type: string
        fqn:
          type: string
        name:
          type: string
        secretGroupId:
          type: string
        value:
          type: string
        createdBySubject:
          $ref: '#/components/schemas/Subject'
        createdAt:
          format: date-time
          type: string
        updatedAt:
          format: date-time
          type: string
        secretVersions:
          type: array
          items:
            $ref: '#/components/schemas/SecretVersion'
        activeDeploymentsCount:
          type: integer
        createdBy:
          type: string
      required:
        - id
        - fqn
        - name
        - secretGroupId
      title: Secret
      x-fern-type-name: Secret
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
    SecretVersion:
      type: object
      properties:
        id:
          type: string
        fqn:
          type: string
        value:
          type: string
        version:
          type: number
        secret:
          $ref: '#/components/schemas/Secret'
        secretId:
          type: string
        createdAt:
          format: date-time
          type: string
        updatedAt:
          format: date-time
          type: string
      required:
        - id
        - fqn
      title: SecretVersion
      x-fern-type-name: SecretVersion
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