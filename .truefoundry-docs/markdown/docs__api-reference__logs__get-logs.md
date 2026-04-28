> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Get Logs

> Fetch logs for various workload components, including Services, Jobs, Workflows, Job Runs, and Pods. Logs are filtered based on the provided query parameters.



## OpenAPI

````yaml /openapi.json get /api/svc/v1/logs
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
  /api/svc/v1/logs:
    get:
      tags:
        - Logs
      summary: Get Logs
      description: >-
        Fetch logs for various workload components, including Services, Jobs,
        Workflows, Job Runs, and Pods. Logs are filtered based on the provided
        query parameters.
      operationId: Logs.getApplicationLogs
      parameters:
        - name: startTs
          required: false
          in: query
          description: >-
            Start timestamp for querying logs, in nanoseconds from the Unix
            epoch.
          schema:
            format: int64
            type: integer
        - name: endTs
          required: false
          in: query
          description: End timestamp for querying logs, in nanoseconds from the Unix epoch.
          schema:
            format: int64
            type: integer
        - name: limit
          required: false
          in: query
          description: Max number of log lines to fetch
          schema:
            type: integer
        - name: direction
          required: false
          in: query
          description: Direction of sorting logs. Can be `asc` or `desc`
          schema:
            $ref: '#/components/schemas/LogsSortingDirection'
        - name: numLogsToIgnore
          required: false
          in: query
          description: >-
            Number of logs corresponding to the starting timestamp to be
            ignored.
          schema:
            type: integer
        - name: applicationId
          required: false
          in: query
          description: Application ID
          schema:
            type: string
        - name: applicationFqn
          required: false
          in: query
          description: Application FQN
          schema:
            type: string
        - name: deploymentId
          required: false
          in: query
          description: Deployment ID
          schema:
            type: string
        - name: jobRunName
          required: false
          in: query
          description: Name of the Job Run for which to fetch logs.
          schema:
            type: string
        - name: podName
          required: false
          in: query
          description: Name of Pod for which to fetch logs.
          schema:
            type: string
        - name: containerName
          required: false
          in: query
          description: Name of the Container for which to fetch logs.
          schema:
            type: string
        - name: podNames
          required: false
          in: query
          description: List of pod names for which to fetch logs.
          schema:
            type: array
            items:
              type: string
        - name: podNamesRegex
          required: false
          in: query
          description: Regex pattern for pod names to fetch logs.
          schema:
            type: string
        - name: searchFilters
          required: false
          in: query
          description: >-
            JSON string containing array of search filters with string, type and
            operator
          schema:
            type: string
        - name: searchString
          required: false
          in: query
          description: String that needs to be matched
          schema:
            type: string
        - name: searchType
          required: false
          in: query
          description: Query filter type, `regex` `substring` `ignore_case_substring`
          schema:
            $ref: '#/components/schemas/LogsSearchFilterType'
        - name: searchOperator
          required: false
          in: query
          description: Comparison operator for filter. `equal` or `not_equal`
          schema:
            $ref: '#/components/schemas/LogsSearchOperatorType'
      responses:
        '200':
          description: Successfully retrieved logs for the workload
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetLogsResponse'
        '400':
          description: >-
            Bad request. This could happen due to: 

            1. Invalid workspace id 

            2. Only one of podName, podNames, or podNamesRegex can be passed in
            query, not multiple
      security:
        - jwt: []
components:
  schemas:
    LogsSortingDirection:
      type: string
      enum:
        - asc
        - desc
      title: LogsSortingDirection
      x-fern-type-name: LogsSortingDirection
    LogsSearchFilterType:
      type: string
      enum:
        - regex
        - substring
        - ignore_case_substring
      description: Query filter type for logs search
      title: LogsSearchFilterType
      x-fern-type-name: LogsSearchFilterType
    LogsSearchOperatorType:
      type: string
      enum:
        - equal
        - not_equal
      description: Comparison operator logs search
      title: LogsSearchOperatorType
      x-fern-type-name: LogsSearchOperatorType
    GetLogsResponse:
      type: object
      properties:
        data:
          description: List of logs
          type: array
          items:
            $ref: '#/components/schemas/Log'
      required:
        - data
      title: GetLogsResponse
      x-fern-type-name: GetLogsResponse
    Log:
      type: object
      properties:
        job_name:
          type: string
          description: Job Name
        log:
          type: string
          description: Log Content
        stream:
          type: string
          description: Log Stream
        time:
          type: string
          description: Log Time
        containerName:
          type: string
          description: Container Name
      required:
        - log
        - stream
        - time
      title: Log
      x-fern-type-name: Log
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````