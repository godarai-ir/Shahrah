> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Get Job Run

> Get Job Run for provided jobRunName and jobId



## OpenAPI

````yaml /openapi.json get /api/svc/v1/jobs/{jobId}/runs/{jobRunName}
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
  /api/svc/v1/jobs/{jobId}/runs/{jobRunName}:
    get:
      tags:
        - Jobs
      summary: Get Job Run
      description: Get Job Run for provided jobRunName and jobId
      operationId: Job.getRun
      parameters:
        - name: jobId
          required: true
          in: path
          description: Application Id of JOB
          schema:
            type: string
        - name: jobRunName
          required: true
          in: path
          description: Job run name of the application
          schema:
            type: string
      responses:
        '200':
          description: Return JobRun details of the provided jobRunName
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetJobRunResponse'
        '403':
          description: Forbidden. User does not have access to the workspace
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '404':
          description: Not Found. Job run not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
      security:
        - jwt: []
components:
  schemas:
    GetJobRunResponse:
      type: object
      properties:
        data:
          description: JobRun
          allOf:
            - $ref: '#/components/schemas/JobRun'
      required:
        - data
      title: GetJobRunResponse
      x-fern-type-name: GetJobRunResponse
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