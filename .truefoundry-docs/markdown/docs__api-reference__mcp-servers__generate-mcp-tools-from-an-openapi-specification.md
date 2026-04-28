> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Generate MCP tools from an OpenAPI specification

> Returns MCP tool definitions from an OpenAPI specification.



## OpenAPI

````yaml /openapi.json post /api/svc/v1/mcp-servers/openapi/generate-mcp-tools
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
  /api/svc/v1/mcp-servers/openapi/generate-mcp-tools:
    post:
      tags:
        - MCP Servers
      summary: Generate MCP tools from an OpenAPI specification
      description: Returns MCP tool definitions from an OpenAPI specification.
      operationId: MCPServer.generateOpenAPITools
      parameters: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/GenerateOpenAPIToolsRequestDto'
      responses:
        '200':
          description: MCP tools generated successfully from the OpenAPI specification
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GenerateOpenAPIToolsResponseDto'
        '400':
          description: Failed to generate MCP tools from OpenAPI specification
      security:
        - jwt: []
        - jwt: []
components:
  schemas:
    GenerateOpenAPIToolsRequestDto:
      type: object
      properties:
        openapi_spec_source:
          description: The OpenAPI specification source
          oneOf:
            - $ref: '#/components/schemas/OpenAPISpecSource'
      required:
        - openapi_spec_source
      title: GenerateOpenAPIToolsRequestDto
      x-fern-type-name: GenerateOpenAPIToolsRequestDto
    GenerateOpenAPIToolsResponseDto:
      type: object
      properties:
        tools:
          description: List of MCP tools generated from the OpenAPI specification
          type: array
          items:
            type: object
        base_urls:
          description: >-
            Base URLs of the servers where requests can be sent (e.g.,
            https://myapp.example.org)
          type: array
          items:
            type: string
        errors:
          description: Validation errors from the OpenAPI specification
          type: array
          items:
            type: object
      required:
        - tools
        - base_urls
        - errors
      title: GenerateOpenAPIToolsResponseDto
      x-fern-type-name: GenerateOpenAPIToolsResponseDto
    OpenAPISpecSource:
      type: object
      oneOf:
        - $ref: '#/components/schemas/RemoteSpecSource'
        - $ref: '#/components/schemas/UploadSpecSource'
        - $ref: '#/components/schemas/InlineSpecSource'
      title: OpenAPISpecSource
      x-fern-type-name: OpenAPISpecSource
    RemoteSpecSource:
      description: >-
        +label=Fetch from URL

        +usage=Fetch the OpenAPI specification from a remote HTTP/HTTPS
        endpoint.
      type: object
      required:
        - type
        - url
      properties:
        type:
          description: +value=remote
          type: string
          enum:
            - remote
        url:
          description: |-
            +label=OpenAPI Spec URL
            +usage=URL to fetch the OpenAPI specification (JSON or YAML format).
            +placeholder=https://myapp.example.org/openapi.json
            +message=Must be a valid HTTP/HTTPS URL
            +sort=10
          type: string
          pattern: ^https?:\/\/[^\s]+$
      title: RemoteSpecSource
      x-fern-type-name: RemoteSpecSource
    UploadSpecSource:
      description: |-
        +label=Upload Spec
        +usage=Fetch the stored/uploaded OpenAPI specification.
      type: object
      required:
        - type
        - url
      properties:
        type:
          description: +value=upload
          type: string
          enum:
            - upload
        url:
          description: >-
            +label=OpenAPI Spec Path

            +usage=Cloud storage path where the OpenAPI specification is stored
            (S3, GCS, or Azure Blob).

            +message=Must be a valid storage path: s3://bucket/...,
            gs://bucket/..., or
            https://<account>.blob.core.windows.net/<container>/...

            +sort=10
          type: string
          pattern: >-
            ^(s3://[^/]+|gs://[^/]+|https://[^.]+\.blob\.core\.windows\.net/[^/]+)/tfy-openapi-specs/(?:\d{4}-\d{2}/|[a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+/)[a-f0-9-]+\.json$
      title: UploadSpecSource
      x-fern-type-name: UploadSpecSource
    InlineSpecSource:
      description: >-
        +label=Inline Spec

        +usage=Inline your OpenAPI specification directly. The spec will be
        stored securely.
      type: object
      required:
        - type
        - openapi_spec
      properties:
        type:
          description: +value=inline
          type: string
          enum:
            - inline
        openapi_spec:
          description: >-
            +label=OpenAPI Spec

            +usage=The OpenAPI specification as a JSON object or a JSON/YAML
            string.

            +message=Must be a valid OpenAPI spec (object or string)

            +sort=10

            +uiType=CodeEditor
          oneOf:
            - type: object
            - type: string
      title: InlineSpecSource
      x-fern-type-name: InlineSpecSource
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````