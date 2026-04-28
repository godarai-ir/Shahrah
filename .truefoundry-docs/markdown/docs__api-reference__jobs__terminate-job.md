> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Terminate Job

> Terminate Job for provided deploymentId and jobRunName



## OpenAPI

````yaml /openapi.json post /api/svc/v1/jobs/terminate
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
  /api/svc/v1/jobs/terminate:
    post:
      tags:
        - Jobs
      summary: Terminate Job
      description: Terminate Job for provided deploymentId and jobRunName
      operationId: Job.terminateJob
      parameters:
        - name: deploymentId
          required: true
          in: query
          description: Deployment Id of the Deployment
          schema:
            type: string
        - name: jobRunName
          required: true
          in: query
          description: Job Run name
          schema:
            type: string
      responses:
        '200':
          description: >-
            Returns object with message and JobRun Status on successful
            termination of Job
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/TerminateJobResponse'
        '403':
          description: Forbidden. User does not have terminate job access
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '404':
          description: |-
            Not Found. This can happen due to: 
            - Deployment not found for deploymentId 
            - Job run not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '417':
          description: Expectation Failed. Could not terminate job
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '422':
          description: Unprocessable Entity. Invalid deploymentId or jobRunName
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
      security:
        - jwt: []
components:
  schemas:
    TerminateJobResponse:
      type: object
      properties:
        message:
          type: string
          description: Terminate Job Message
        jobRunStatus:
          description: Job run status
          allOf:
            - $ref: '#/components/schemas/JobRunStatus'
      required:
        - message
        - jobRunStatus
      title: TerminateJobResponse
      x-fern-type-name: TerminateJobResponse
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
    JobRunStatus:
      type: string
      enum:
        - CREATED
        - SCHEDULED
        - RUNNING
        - FINISHED
        - FAILED
        - TERMINATION_REQUESTED
        - TERMINATING
        - TERMINATED
        - UNKNOWN
      title: JobRunStatus
      x-fern-type-name: JobRunStatus
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````