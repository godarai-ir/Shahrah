> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# List Secrets

> List secrets associated with a user filtered with optional parameters passed in the body.



## OpenAPI

````yaml /openapi.json post /api/svc/v1/secrets
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
  /api/svc/v1/secrets:
    post:
      tags:
        - Secrets
      summary: List Secrets
      description: >-
        List secrets associated with a user filtered with optional parameters
        passed in the body.
      operationId: Secret.getSecrets
      parameters: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ListSecretsRequest'
      responses:
        '200':
          description: >-
            Returns the secrets associated with a user filtered with optional
            parameters passed in the body.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListSecretsResponse'
        '403':
          description: >-
            Forbidden. This can happen due to: 

            - withValue is true and the control plane has
            `DISABLE_SECRET_VALUE_VIEW` set.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '404':
          description: >-
            Not Found. This can happen due to: 

            - The user does not have access to the secrets with the fqns
            provided in the body. 

            - The secrets with the fqns provided in the body do not exist
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
      security:
        - jwt: []
components:
  schemas:
    ListSecretsRequest:
      type: object
      properties:
        limit:
          type: integer
          minimum: 1
          maximum: 1000
          default: 100
          description: Number of items per page
          example: 10
        offset:
          type: integer
          minimum: 0
          default: 0
          description: Number of items to skip
          example: 0
        secretFqns:
          description: Array of FQNs
          type: array
          items:
            type: string
        secretGroupId:
          type: string
          description: Secret Group Id of the secret gourp.
        withValue:
          type: boolean
          description: >-
            Whether to include the secret values in the response. Defaults to
            false.
          default: false
      title: ListSecretsRequest
      x-fern-type-name: ListSecretsRequest
    ListSecretsResponse:
      type: object
      properties:
        data:
          description: Array of Secrets
          type: array
          items:
            $ref: '#/components/schemas/Secret'
        pagination:
          description: Pagination Information
          allOf:
            - $ref: '#/components/schemas/Pagination'
      required:
        - data
        - pagination
      title: ListSecretsResponse
      x-fern-type-name: ListSecretsResponse
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