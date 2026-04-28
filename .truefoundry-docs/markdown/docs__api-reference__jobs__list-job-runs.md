> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# List Job Runs

> List Job Runs for provided Job Id. Filter the data based on parameters passed in the query



## OpenAPI

````yaml /openapi.json get /api/svc/v1/jobs/{jobId}/runs
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
  /api/svc/v1/jobs/{jobId}/runs:
    get:
      tags:
        - Jobs
      summary: List Job Runs
      description: >-
        List Job Runs for provided Job Id. Filter the data based on parameters
        passed in the query
      operationId: Job.getRunsHistory
      parameters:
        - name: jobId
          required: true
          in: path
          description: Job id of the application
          schema:
            type: string
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
        - name: searchPrefix
          required: false
          in: query
          description: Prefix used to search for job runs by name or identifier
          schema:
            type: string
        - name: sortBy
          required: false
          in: query
          description: Attribute to sort by
          schema:
            $ref: '#/components/schemas/JobRunsSortBy'
        - name: order
          required: false
          in: query
          description: Sorting order
          schema:
            $ref: '#/components/schemas/SortDirection'
        - name: triggeredBy
          required: false
          in: query
          description: Array of subject slugs
          schema:
            type: array
            items:
              type: string
        - name: status
          required: false
          in: query
          description: Status of the job run
          schema:
            type: array
            items:
              $ref: '#/components/schemas/JobRunStatus'
        - name: versionNumbers
          required: false
          in: query
          description: Version number of the deployment
          schema:
            type: array
            items:
              type: number
      responses:
        '200':
          description: >-
            Returns all runs of a Job sorted by creation date in "data" key and
            total count in "totalCount" key
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListJobRunResponse'
        '403':
          description: Forbidden. User does not have access to the workspace
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '404':
          description: Not found. Job with the given id does not exist
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '422':
          description: Unprocessable Entity. Missing or invalid values for offset and limit
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
      security:
        - jwt: []
components:
  schemas:
    JobRunsSortBy:
      type: string
      enum:
        - startTime
        - duration
        - deploymentVersion
        - triggeredBy
        - status
      title: JobRunsSortBy
      x-fern-type-name: JobRunsSortBy
    SortDirection:
      type: string
      enum:
        - asc
        - desc
      title: SortDirection
      x-fern-type-name: SortDirection
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
    ListJobRunResponse:
      type: object
      properties:
        data:
          description: JobRun History
          type: array
          items:
            $ref: '#/components/schemas/JobRun'
        pagination:
          description: Pagination Information
          allOf:
            - $ref: '#/components/schemas/Pagination'
      required:
        - data
        - pagination
      title: ListJobRunResponse
      x-fern-type-name: ListJobRunResponse
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