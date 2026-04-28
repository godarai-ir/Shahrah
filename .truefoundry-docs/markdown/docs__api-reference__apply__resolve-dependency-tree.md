> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Resolve Dependency Tree

> Resolves the dependency tree for the given manifests.



## OpenAPI

````yaml /openapi.json post /api/svc/v1/dependency-tree
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
  /api/svc/v1/dependency-tree:
    post:
      tags:
        - Apply
      summary: Resolve Dependency Tree
      description: Resolves the dependency tree for the given manifests.
      operationId: TfyApply.dependencyTree
      parameters: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/DependencyTreeRequestDto'
      responses:
        '200':
          description: The dependency tree has been successfully resolved.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DependencyTreeResponseDto'
      security:
        - jwt: []
components:
  schemas:
    DependencyTreeRequestDto:
      type: object
      properties:
        manifests:
          description: >-
            Array of manifests of the resources to be resolved for the
            dependency tree
          type: array
          items:
            type: object
      required:
        - manifests
      title: DependencyTreeRequestDto
      x-fern-type-name: DependencyTreeRequestDto
    DependencyTreeResponseDto:
      type: object
      properties:
        data:
          description: List of resources with their dependencies
          type: array
          items:
            $ref: '#/components/schemas/Resource'
      required:
        - data
      title: DependencyTreeResponseDto
      x-fern-type-name: DependencyTreeResponseDto
    Resource:
      type: object
      properties:
        resourceType:
          type: string
          description: The type of the manifest
        resourceName:
          type: string
          description: The name of the resource
        dependencies:
          description: The dependencies of the resource
          type: array
          items:
            type: string
      required:
        - resourceType
        - resourceName
        - dependencies
      title: Resource
      x-fern-type-name: Resource
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````