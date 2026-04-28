> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# List Secret Groups

> List the secret groups associated with a user along with the associated secrets for each group. Filtered with the options passed in the query fields. Note: This method does not return the secret values of the <em>associatedSecrets</em> in the response. A separate API call to `/v1/secrets/{id}` should be made to fetch the associated secret value.



## OpenAPI

````yaml /openapi.json get /api/svc/v1/secret-groups
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
  /api/svc/v1/secret-groups:
    get:
      tags:
        - Secret Groups
      summary: List Secret Groups
      description: >-
        List the secret groups associated with a user along with the associated
        secrets for each group. Filtered with the options passed in the query
        fields. Note: This method does not return the secret values of the
        <em>associatedSecrets</em> in the response. A separate API call to
        `/v1/secrets/{id}` should be made to fetch the associated secret value.
      operationId: SecretGroup.get
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
        - name: fqn
          required: false
          in: query
          description: Fqn of secret group.
          schema:
            type: string
        - name: search
          required: false
          in: query
          description: >-
            Search query - filters by secret group names that contain the search
            string
          schema:
            type: string
      responses:
        '200':
          description: >-
            Returns all the secret groups associated with a user along with the
            associated secrets for each group.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListSecretGroupResponse'
      security:
        - jwt: []
components:
  schemas:
    ListSecretGroupResponse:
      type: object
      properties:
        data:
          description: Array of Secret Groups
          type: array
          items:
            $ref: '#/components/schemas/SecretGroup'
        pagination:
          description: Pagination information
          allOf:
            - $ref: '#/components/schemas/Pagination'
      required:
        - data
        - pagination
      title: ListSecretGroupResponse
      x-fern-type-name: ListSecretGroupResponse
    SecretGroup:
      type: object
      properties:
        id:
          type: string
        fqn:
          type: string
        tenantName:
          type: string
        createdBySubject:
          $ref: '#/components/schemas/Subject'
        associatedSecrets:
          type: array
          items:
            $ref: '#/components/schemas/Secret'
        integrationId:
          type: string
        manifest:
          type: object
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
        - tenantName
        - createdBySubject
        - associatedSecrets
        - accountId
        - createdAt
        - updatedAt
      title: SecretGroup
      x-fern-type-name: SecretGroup
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
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````