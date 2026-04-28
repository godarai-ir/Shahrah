> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Trigger Job

> Trigger Job for provided deploymentId or applicationId



## OpenAPI

````yaml /openapi.json post /api/svc/v1/jobs/trigger
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
  /api/svc/v1/jobs/trigger:
    post:
      tags:
        - Jobs
      summary: Trigger Job
      description: Trigger Job for provided deploymentId or applicationId
      operationId: Job.triggerJob
      parameters: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/TriggerJobRequest'
      responses:
        '200':
          description: >-
            Returns object with message, JobRun Name And Job Details on
            successful creation of Job
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/TriggerJobRunResponse'
        '400':
          description: >-
            Bad Request. This can happen because application not supported for
            job trigger
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '403':
          description: Forbidden. User does not have trigger job access
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '404':
          description: |-
            Not Found. This can happen due to: 
            - Application not found for applicationId 
            - Active deployment not found for applicationId 
            - Deployment not found for deploymentId
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '422':
          description: |-
            Unprocessable Entity. This can happen due to: 
            - Both deploymentId and applicationId are not provided 
            - Params not allowed, since job has no support for params 
            - Command is not string or array of string
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
      security:
        - jwt: []
components:
  schemas:
    TriggerJobRequest:
      type: object
      properties:
        deploymentId:
          type: string
          description: Deployment Id of the job
        applicationId:
          type: string
          description: Application Id of the job
        input:
          description: Job trigger input
          oneOf:
            - $ref: '#/components/schemas/JobTriggerInput'
            - $ref: '#/components/schemas/SparkJobTriggerInput'
        metadata:
          description: Metadata for the job run including job_alias_name
          allOf:
            - $ref: '#/components/schemas/Metadata'
      title: TriggerJobRequest
      x-fern-type-name: TriggerJobRequest
    TriggerJobRunResponse:
      type: object
      properties:
        message:
          type: string
          description: Job triggered
        jobRunName:
          type: string
          description: Name of the job run
        data:
          description: Details of the triggered job run
          allOf:
            - $ref: '#/components/schemas/JobRun'
      required:
        - message
        - jobRunName
        - data
      title: TriggerJobRunResponse
      x-fern-type-name: TriggerJobRunResponse
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
    JobTriggerInput:
      type: object
      properties:
        command:
          oneOf:
            - type: string
            - type: array
              items:
                type: string
          description: Command
        params:
          type: object
          additionalProperties: true
          description: Parameters used to populate template command
      title: JobTriggerInput
      x-fern-type-name: JobTriggerInput
    SparkJobTriggerInput:
      type: object
      properties:
        mainClass:
          type: string
          description: Main Class for Spark Job
        mainApplicationFile:
          type: string
          description: Main Application File for Spark Job
        arguments:
          type: string
          description: Arguments to pass to the main application file
      title: SparkJobTriggerInput
      x-fern-type-name: SparkJobTriggerInput
    Metadata:
      type: object
      properties:
        job_run_name_alias:
          type: string
          description: Job Run Name Alias
      title: Metadata
      x-fern-type-name: Metadata
    JobRun:
      type: object
      properties:
        id:
          type: string
          description: JobRun ID
        name:
          type: string
          description: JobRun Name
        applicationName:
          type: string
          description: Application Name
        deploymentVersion:
          type: string
          description: Deployment Version
        createdAt:
          type: number
          description: Created At
        endTime:
          type: number
          description: End Time of JobRun
        duration:
          type: integer
          description: Duration of JobRun
        command:
          type: string
          description: Command
        totalRetries:
          type: integer
          description: Total Retries
        error:
          type: string
          description: Error
        status:
          description: Status of JobRun
          allOf:
            - $ref: '#/components/schemas/JobRunStatus'
        triggeredBy:
          type: string
          description: Triggered By
        triggeredBySubject:
          description: Triggered By Subject
          allOf:
            - $ref: '#/components/schemas/Subject'
        exitCode:
          type: integer
          description: Exit Code
        sparkUi:
          type: string
          description: Spark UI Url
        applicationId:
          type: string
          description: Application ID
        deploymentId:
          type: string
          description: Deployment ID
        tenantName:
          type: string
          description: Tenant Name
      required:
        - id
        - name
        - applicationName
        - deploymentVersion
        - createdAt
        - command
        - totalRetries
        - status
      title: JobRun
      x-fern-type-name: JobRun
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