> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Get Deployment Specifications

> Fetches deployment specifications for a model version or a HuggingFace model URL.



## OpenAPI

````yaml /openapi.json get /api/svc/v1/model-catalogues/deployment-specs
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
  /api/svc/v1/model-catalogues/deployment-specs:
    get:
      tags:
        - Model Deployments
      summary: Get Deployment Specifications
      description: >-
        Fetches deployment specifications for a model version or a HuggingFace
        model URL.
      operationId: ModelCatalogue.getDeploymentSpecs
      parameters:
        - name: workspaceId
          required: true
          in: query
          description: Workspace Id
          schema:
            type: string
        - name: modelVersionFqn
          required: false
          in: query
          description: FQN of the Model Version
          schema:
            type: string
        - name: huggingfaceHubUrl
          required: false
          in: query
          description: huggingface hub url
          schema:
            type: string
        - name: huggingfaceHubTokenSecretFqn
          required: false
          in: query
          description: Huggingface hub token secret
          schema:
            type: string
        - name: pipelineTagOverride
          required: false
          in: query
          description: >-
            The pipeline tag of the model for which deployment spec is being
            generated
          schema:
            type: string
      responses:
        '200':
          description: >-
            Returns deployment specifications for the model version or
            HuggingFace model URL
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListModelDeploymentSpecResponse'
        '400':
          description: |-
            Bad Request. This can occur due to: 
            1. Neither Model Version FQN nor HuggingFace URL provided. 
            2. Model version is not deployable. 
            3. Framework is not supported for deployment.
        '404':
          description: Not Found. Workspace with the provided ID not found.
      security:
        - jwt: []
components:
  schemas:
    ListModelDeploymentSpecResponse:
      type: object
      properties:
        data:
          description: Array of ModelDeploymentSpec
          type: array
          items:
            $ref: '#/components/schemas/ModelServerOption'
      required:
        - data
      title: ListModelDeploymentSpecResponse
      x-fern-type-name: ListModelDeploymentSpecResponse
    ModelServerOption:
      type: object
      properties:
        name:
          type: string
          enum:
            - vLLM
            - SGLang
            - TRT-LLM
            - NIM
            - TEI
            - Infinity
            - Triton
            - MMS
            - FastAPI
            - TGI
        displayName:
          type: string
        description:
          type: string
        openAPIDescription:
          type: string
        deployments:
          type: array
          items:
            $ref: '#/components/schemas/DeploymentConfig'
      required:
        - name
        - displayName
        - description
        - openAPIDescription
        - deployments
      title: ModelServerOption
      x-fern-type-name: ModelServerOption
    DeploymentConfig:
      type: object
      properties:
        name:
          type: string
        spec:
          type: object
          additionalProperties: true
        isAvailableInWorkspace:
          type: boolean
        cost:
          type: object
          additionalProperties: true
        isTRTLLMEngineAvailable:
          type: boolean
      required:
        - name
        - spec
        - isAvailableInWorkspace
      title: DeploymentConfig
      x-fern-type-name: DeploymentConfig
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````