> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Get Finetuning Specifications

> Fetches finetuning specifications for a model version or a HuggingFace model URL



## OpenAPI

````yaml /openapi.json get /api/svc/v1/model-catalogues/finetune-specs
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
  /api/svc/v1/model-catalogues/finetune-specs:
    get:
      tags:
        - Model Deployments
      summary: Get Finetuning Specifications
      description: >-
        Fetches finetuning specifications for a model version or a HuggingFace
        model URL
      operationId: ModelCatalogue.getFinetuneSpecs
      parameters:
        - name: huggingfaceHubUrl
          required: true
          in: query
          description: huggingface hub url
          schema:
            type: string
        - name: workspaceId
          required: true
          in: query
          description: workspace id
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
            The pipeline tag of the model for which finetuning spec is being
            generated
          schema:
            type: string
      responses:
        '200':
          description: >-
            Returns finetuning specifications for the model version or
            HuggingFace model URL
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListModelFinetuningSpecResponse'
        '400':
          description: |-
            Bad Request. This can occur due to: 
            1. HuggingFace URL is not provided. 
            2. Model type is not supported for finetuning.
      security:
        - jwt: []
components:
  schemas:
    ListModelFinetuningSpecResponse:
      type: object
      properties:
        data:
          description: Array of ModelFinetuneSpec
          type: array
          items:
            $ref: '#/components/schemas/ModelFinetuneSpec'
      required:
        - data
      title: ListModelFinetuningSpecResponse
      x-fern-type-name: ListModelFinetuningSpecResponse
    ModelFinetuneSpec:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        finetuneType:
          type: string
          enum:
            - qlora
            - lora
            - full
        metadata:
          type: object
          additionalProperties: true
        spec:
          type: object
          additionalProperties: true
        isAvailableInWorkspace:
          type: boolean
      required:
        - id
        - name
        - finetuneType
        - metadata
        - spec
        - isAvailableInWorkspace
      title: ModelFinetuneSpec
      x-fern-type-name: ModelFinetuneSpec
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````