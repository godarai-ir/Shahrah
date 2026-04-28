> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Get Audit Logs

> Get paginated audit logs filtered by parameters passed in the query.



## OpenAPI

````yaml /openapi.json get /api/svc/v1/audit-logs
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
  /api/svc/v1/audit-logs:
    get:
      tags:
        - Audit Logs
      summary: Get Audit Logs
      description: Get paginated audit logs filtered by parameters passed in the query.
      operationId: ActivityLog.queryActivityLogPaginated
      parameters:
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
        - name: metric
          required: true
          in: query
          description: Either `list` or `count`
          schema:
            $ref: '#/components/schemas/ActivityLogMetric'
        - name: action
          required: false
          in: query
          description: Action
          schema:
            type: string
        - name: resourceId
          required: false
          in: query
          description: Resource Id
          schema:
            type: string
        - name: resourceType
          required: false
          in: query
          description: Resource Type
          schema:
            type: string
        - name: resourceName
          required: false
          in: query
          description: Resource Name
          schema:
            type: string
        - name: resourceFqn
          required: false
          in: query
          description: Resource Fqn
          schema:
            type: string
        - name: clusterId
          required: false
          in: query
          description: Cluster Id
          schema:
            type: string
        - name: workspaceId
          required: false
          in: query
          description: Workspace Id
          schema:
            type: string
        - name: subjectSlug
          required: false
          in: query
          description: Slug of Subjects (comma separated)
          schema:
            type: string
        - name: start
          required: false
          in: query
          description: Start Time (ISO 8601)
          schema:
            type: string
        - name: end
          required: false
          in: query
          description: End Time (ISO 8601)
          schema:
            type: string
      responses:
        '200':
          description: Returns list/count of audit logs according to filter
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListActivityLogResponse'
      security:
        - jwt: []
components:
  schemas:
    ActivityLogMetric:
      type: string
      enum:
        - list
        - count
      title: ActivityLogMetric
      x-fern-type-name: ActivityLogMetric
    ListActivityLogResponse:
      type: object
      properties:
        data:
          description: Array of ActivityLog
          type: array
          items:
            $ref: '#/components/schemas/ActivityLog'
        pagination:
          description: Pagination information
          allOf:
            - $ref: '#/components/schemas/Pagination'
      required:
        - data
        - pagination
      title: ListActivityLogResponse
      x-fern-type-name: ListActivityLogResponse
    ActivityLog:
      type: object
      properties:
        id:
          type: string
        action:
          type: string
          enum:
            - APPLY_ACCOUNT
            - DELETE_ACCOUNT
            - CREATE_DEPLOYMENT
            - CANCEL_DEPLOYMENT
            - DELETE_APPLICATION
            - PAUSE_APPLICATION
            - RESUME_APPLICATION
            - START_NOTEBOOK
            - STOP_NOTEBOOK
            - START_CODESERVER
            - STOP_CODESERVER
            - START_SSH_SERVER
            - STOP_SSH_SERVER
            - START_RSTUDIO
            - STOP_RSTUDIO
            - CREATE_CLUSTER
            - UPDATE_CLUSTER
            - DELETE_CLUSTER
            - CREATE_CLUSTER_INFRA
            - UPDATE_CLUSTER_INFRA
            - CORDON_NODE
            - UNCORDON_NODE
            - TOGGLE_AUTOPILOT
            - TRIGGER_JOB
            - TERMINATE_JOB
            - APPLY_RECOMMENDATION
            - UPGRADE_ADDON
            - APPLY_TEAM
            - DELETE_ROLE
            - APPLY_ROLE
            - DELETE_TEAM
            - APPLY_ML_REPO
            - APPLY_PROVIDER_ACCOUNT
            - DELETE_PROVIDER_ACCOUNT
            - APPLY_PROVIDER_INTEGRATION
            - DELETE_PROVIDER_INTEGRATION
            - CREATE_SECRET_GROUP
            - UPDATE_SECRET_GROUP
            - DELETE_SECRET_GROUP
            - DELETE_SECRET
            - EXECUTE_WORKFLOW
            - TERMINATE_WORKFLOW
            - RECOVER_WORKFLOW
            - APPLY_WORKSPACE
            - DELETE_WORKSPACE
            - APPLY_INTERCEPT
            - ADD_COLLABORATORS
            - UPDATE_COLLABORATORS_ROLE
            - REMOVE_COLLABORATORS
            - CREATE_ENVIRONMENT
            - UPDATE_ENVIRONMENT
            - DELETE_ENVIRONMENT
            - APPLY_ENVIRONMENT
            - CREATE_APPLICATION_SET
            - CREATE_PAT
            - DELETE_PAT
            - CREATE_VIRTUAL_ACCOUNT
            - APPLY_VIRTUAL_ACCOUNT
            - DELETE_VIRTUAL_ACCOUNT
            - REGENERATE_VIRTUAL_ACCOUNT_TOKEN
            - SYNC_VIRTUAL_ACCOUNT_TOKEN_TO_SECRET_STORE
            - CLONE_APPLICATION
            - PROMOTE_APPLICATION
            - PROMOTE_APPLICATION_ROLLOUT
            - TOGGLE_AUTOPILOT_FOR_CLUSTER
            - CREATE_LLM_GATEWAY_RATE_LIMIT
            - UPDATE_LLM_GATEWAY_RATE_LIMIT
            - DELETE_LLM_GATEWAY_RATE_LIMIT
            - ADD_TRUEFOUNDRY_HOSTED_MODEL
            - DELETE_POD
            - DELETE_VCS_INTEGRATION
            - DELETE_JOB_RUN
            - SYNC_METADATA_FOR_CLUSTER
            - APPLY_POLICY
            - DELETE_POLICY
            - APPLY_SETTINGS
            - DELETE_SETTINGS
            - UPDATE_USER_ROLES
            - INVITE_USER
            - REGISTER_USER
            - RESEND_USER_INVITE
            - PRE_REGISTER_USERS
            - ACTIVATE_USER
            - ADD_USERS_TO_ACCOUNT
            - DEACTIVATE_USER
            - DELETE_USER
            - SYNC_USERS_TO_AUTH_SERVER
            - CHANGE_PASSWORD
            - FORGOT_PASSWORD
            - APPLY_ALERT_CONFIG
            - DELETE_ALERT_CONFIG
            - APPLY_ALERT_RULE
            - DELETE_ALERT_RULE
            - CREATE_OR_UPDATE_GATEWAY_CONFIG
            - DELETE_GATEWAY_CONFIG
            - CREATE_TRACING_PROJECT
            - UPDATE_TRACING_PROJECT
            - DELETE_TRACING_PROJECT
            - CREATE_TRACING_APPLICATION
            - UPDATE_TRACING_APPLICATION
            - DELETE_TRACING_APPLICATION
            - DRAIN_NODE
            - DELETE_NODE
            - APPLY_AGENT
            - DELETE_AGENT
            - APPLY_AGENT_APP
            - DELETE_AGENT_APP
            - CREATE_OR_UPDATE_GATEWAY_INSTALLATION
            - DELETE_GATEWAY_INSTALLATION
            - SYNC_APPLICATION
            - APPLY_EXTERNAL_IDENTITY
            - DELETE_EXTERNAL_IDENTITY
            - APPLY_EXTERNAL_IDENTITY_PROVIDER
            - DELETE_EXTERNAL_IDENTITY_PROVIDER
            - SHUTDOWN_CLUSTER
            - RESUME_CLUSTER
            - DELETE_TENANT
            - APPLY_ROLE_BINDING
            - BULK_APPLY_ROLE_BINDING
            - DELETE_ROLE_BINDING
            - APPLY_MCP_SERVER
            - DELETE_MCP_SERVER
        requestPath:
          type: string
        resourceId:
          type: string
        resourceType:
          type: string
          enum:
            - role
            - account
            - cluster
            - workspace
            - environment
            - secret-group
            - deployment
            - application
            - secret
            - docker-registry
            - tenant
            - vcs-integration
            - mlf-project
            - repository
            - provider-integration
            - team
            - service-account
            - llm-gateway
            - policy
            - role-binding
            - settings
            - provider-account
            - user
            - alert-config
            - alert-rule
            - gateway-config
            - gateway-controls
            - tracing-project
            - tracing-application
            - agent
            - agent-app
            - gateway-installation
            - virtual-account
            - sso-based-identity
            - external-identity
            - mcp-server
            - external-identity-provider
        resourceName:
          type: string
        resourceFqn:
          type: string
        tenantName:
          type: string
        clusterId:
          type: string
        workspaceId:
          type: string
        subject:
          $ref: '#/components/schemas/Subject'
        metaData:
          type: object
        createdAt:
          format: date-time
          type: string
        updatedAt:
          format: date-time
          type: string
      required:
        - id
        - action
        - requestPath
        - resourceId
        - resourceType
        - resourceName
        - resourceFqn
        - tenantName
        - clusterId
        - workspaceId
        - subject
        - metaData
        - createdAt
        - updatedAt
      title: ActivityLog
      x-fern-type-name: ActivityLog
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