> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Download Logs

> Download logs associated with the specified workload, including Jobs, Services, Job Runs, Pods, or Workflows. Logs are filtered based on the provided query parameters.



## OpenAPI

````yaml /openapi.json get /api/svc/v1/logs/{workspaceId}/download
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
  /api/svc/v1/logs/{workspaceId}/download:
    get:
      tags:
        - Logs
      summary: Download Logs
      description: >-
        Download logs associated with the specified workload, including Jobs,
        Services, Job Runs, Pods, or Workflows. Logs are filtered based on the
        provided query parameters.
      operationId: Logs.downloadLogs
      parameters:
        - name: workspaceId
          required: true
          in: path
          schema:
            type: string
        - name: startTs
          required: false
          in: query
          description: >-
            Start timestamp for querying logs, in nanoseconds from the Unix
            epoch.
          schema:
            example: '1635467890123456789'
            type: string
        - name: endTs
          required: false
          in: query
          description: End timestamp for querying logs, in nanoseconds from the Unix epoch.
          schema:
            example: '1635467891123456789'
            type: string
        - name: limit
          required: false
          in: query
          description: Max number of log lines to fetch
          schema:
            type: string
        - name: direction
          required: false
          in: query
          description: Direction of sorting logs. Can be `asc` or `desc`
          schema:
            type: string
        - name: getLogsQuery
          required: false
          in: query
          description: Query params to fetch logs
          schema:
            $ref: '#/components/schemas/GetLogsQuery'
        - name: filterQuery
          required: false
          in: query
          description: Query to filter logs
          schema:
            $ref: '#/components/schemas/LogsFilterQuery'
        - name: numLogsToIgnore
          required: false
          in: query
          description: >-
            Number of logs corresponding to the starting timestamp to be
            ignored.
          schema:
            type: number
      responses:
        '200':
          description: Successfully retrieved and streamed logs as a downloadable file
          content:
            application/json:
              schema:
                type: string
                format: binary
        '400':
          description: Bad Request. Invalid workspace ID or incorrect query parameters
      security:
        - jwt: []
components:
  schemas:
    GetLogsQuery:
      type: object
      properties:
        applicationId:
          type: string
          description: Application ID
          example: application-12345
        deploymentId:
          type: string
          description: Deployment ID
          example: deployment-67890
        jobRunName:
          type: string
          description: Name of the Job Run for which to fetch logs.
          example: example-job-5147f0c8daaa
        podName:
          type: string
          description: Name of Pod for which to fetch logs.
          example: example-pod-1879544112
        containerName:
          type: string
          description: Name of the Container for which to fetch logs.
          example: example-container
        podNames:
          type: string
          description: List of pod names (comma-separated) for which to fetch logs.
          example:
            - pod-1
        podNamesRegex:
          type: string
          description: Regex pattern for pod names to fetch logs.
          example: example-pod-.*
      title: GetLogsQuery
      x-fern-type-name: GetLogsQuery
    LogsFilterQuery:
      type: object
      properties:
        matchString:
          type: string
          description: String that needs to be matched
        type:
          description: query filter type, `regex` or `substring`
          allOf:
            - $ref: '#/components/schemas/LogsFilterQueryType'
        operator:
          description: Comparison operator for filter. `equal` or `not_equal`
          allOf:
            - $ref: '#/components/schemas/LogsSearchOperatorType'
      required:
        - matchString
        - type
        - operator
      title: LogsFilterQuery
      x-fern-type-name: LogsFilterQuery
    LogsFilterQueryType:
      type: string
      enum:
        - regex
        - substring
        - ignore_case_substring
      description: query filter type, `regex` or `substring`
      title: LogsFilterQueryType
      x-fern-type-name: LogsFilterQueryType
    LogsSearchOperatorType:
      type: string
      enum:
        - equal
        - not_equal
      description: Comparison operator logs search
      title: LogsSearchOperatorType
      x-fern-type-name: LogsSearchOperatorType
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````