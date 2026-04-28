> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# List Virtual Accounts

> List virtual accounts for the tenant.



## OpenAPI

````yaml /openapi.json get /api/svc/v1/virtual-accounts
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
  /api/svc/v1/virtual-accounts:
    get:
      tags:
        - Virtual Accounts
      summary: List Virtual Accounts
      description: List virtual accounts for the tenant.
      operationId: VirtualAccount.getAllVirtualAccountsForTenant
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
        - name: nameSearchQuery
          required: false
          in: query
          description: Return virtual accounts with names that contain this string
          schema:
            type: string
        - name: ownedByTeams
          required: false
          in: query
          description: Return virtual accounts owned by these teams
          schema:
            type: array
            items:
              type: string
        - name: isExpired
          required: false
          in: query
          description: >-
            Filter virtual accounts by expiration status. true = expired, false
            = not expired
          schema:
            type: boolean
        - name: filter
          required: false
          in: query
          description: >-
            JSON string: structured filter tree (AND/OR groups, column leaves on
            `name`, json_map leaves on manifest.tags).
          schema:
            type: string
      responses:
        '200':
          description: Return all virtual accounts for the tenant
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListVirtualAccountResponse'
      security:
        - jwt: []
components:
  schemas:
    ListVirtualAccountResponse:
      type: object
      properties:
        data:
          description: Array of Virtual Accounts
          type: array
          items:
            $ref: '#/components/schemas/VirtualAccount'
        pagination:
          description: Pagination Information
          allOf:
            - $ref: '#/components/schemas/Pagination'
      required:
        - data
        - pagination
      title: ListVirtualAccountResponse
      x-fern-type-name: ListVirtualAccountResponse
    VirtualAccount:
      type: object
      properties:
        id:
          type: string
        type:
          type: string
        tenantName:
          type: string
        manifest:
          oneOf:
            - $ref: '#/components/schemas/VirtualAccountManifest'
            - $ref: '#/components/schemas/PersonalAccessTokenManifest'
        jwtId:
          type: string
        createdBySubject:
          $ref: '#/components/schemas/Subject'
        createdAt:
          format: date-time
          type: string
        updatedAt:
          format: date-time
          type: string
        isExpired:
          type: boolean
        jwts:
          type: array
          items:
            $ref: '#/components/schemas/Jwt'
        accountId:
          type: string
          nullable: false
        metadata:
          type: object
          additionalProperties: true
        roleIds:
          type: array
          items:
            type: string
        rolesWithResource:
          type: array
          items:
            $ref: '#/components/schemas/RoleWithResource'
        createdBy:
          type: string
        nextScheduledRotation:
          type: string
      required:
        - id
        - type
        - tenantName
        - createdBySubject
        - createdAt
        - updatedAt
        - accountId
      title: VirtualAccount
      x-fern-type-name: VirtualAccount
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
    VirtualAccountManifest:
      type: object
      required:
        - name
        - type
        - permissions
      properties:
        name:
          description: >-
            +label=Name

            +sort=1

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +usage=Virtual Account Name
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        type:
          description: +value=virtual-account
          type: string
          enum:
            - virtual-account
        permissions:
          description: |-
            +label=Permissions
            +sort=2
            +uiType=ServiceAccountPermissions
          type: array
          items:
            $ref: '#/components/schemas/Permissions'
        expiration_date:
          description: >-
            +label=Expiration Date (UTC)

            +message=Expiration date of the virtual account

            +usage=Expiration Date of the Virtual Account (should be in the
            format yyyy-mm-dd)

            +sort=3

            +uiType=DatePicker
          type: string
          pattern: ^\d{4}-\d{2}-\d{2}$
        auto_rotate:
          $ref: '#/components/schemas/AutoRotate'
        notification_target:
          $ref: '#/components/schemas/NotificationTarget'
        secret_store_config:
          $ref: '#/components/schemas/SecretStoreConfig'
        ownedBy:
          $ref: '#/components/schemas/VirtualAccountOwnedBy'
        tags:
          description: >-
            +label=Tags

            +usage=Key-value pairs to categorize this Virtual Account (e.g., by
            owner or environment).

            +sort=8
          type: object
      title: VirtualAccountManifest
      x-fern-type-name: VirtualAccountManifest
    PersonalAccessTokenManifest:
      type: object
      required:
        - name
        - type
        - user_fqn
      properties:
        name:
          description: >-
            +label=Name

            +sort=1

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +usage=Personal Access Token Name
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        type:
          description: +value=pat
          type: string
          enum:
            - pat
        expiration_date:
          description: >-
            +label=Expiration Date (UTC)

            +sort=2

            +message=Expiration date of the personal access token

            +usage=Expiration Date of the Personal Access Token (should be in
            the ISO 8601 format)

            +uiType=DatePicker
          type: string
          pattern: ^\S+$
        user_fqn:
          description: |-
            +label=User FQN
            +sort=3
            +usage=The fully qualified name of the user
          type: string
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: PersonalAccessTokenManifest
      x-fern-type-name: PersonalAccessTokenManifest
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
    Jwt:
      type: object
      properties:
        id:
          type: string
        subjectType:
          type: string
        subjectId:
          type: string
        metadata:
          type: object
          additionalProperties: true
        expiry:
          format: date-time
          type: string
        createdAt:
          format: date-time
          type: string
        updatedAt:
          format: date-time
          type: string
      required:
        - id
        - subjectType
        - subjectId
        - expiry
        - createdAt
        - updatedAt
      title: Jwt
      x-fern-type-name: Jwt
    RoleWithResource:
      type: object
      properties:
        roleId:
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
        resourceId:
          type: string
      required:
        - roleId
        - resourceType
        - resourceId
      title: RoleWithResource
      x-fern-type-name: RoleWithResource
    Permissions:
      type: object
      required:
        - resource_fqn
        - resource_type
        - role_id
      properties:
        resource_fqn:
          description: |-
            +label= Resource FQN
            +usage=The fully qualified name of the resource
          type: string
        resource_type:
          description: |-
            +label=Resource Type
            +usage=The type of the resource (cluster, workspace, etc.)
          type: string
        role_id:
          description: >-
            +label=Role ID

            +usage=The role id of the role to be assigned to the service account
            for that resource
          type: string
      title: Permissions
      x-fern-type-name: Permissions
    AutoRotate:
      description: |-
        +label=Enable Auto Rotation
        +sort=4
        +usage=Enable Auto Rotation to automatically rotate the token
        +message=Enable Auto Rotation to automatically rotate the token
        +uiProps={"disableEdit":true}
      type: object
      required:
        - auto_rotate_interval
        - grace_period
      properties:
        auto_rotate_interval:
          description: >-
            +label=Rotation Interval in days

            +sort=1

            +usage=Rotation Interval in days after which the token will be
            rotated. Minimum value is 30.
          type: integer
          minimum: 30
          default: 360
        grace_period:
          description: >-
            +label=Grace Period in days

            +sort=2

            +usage=Grace Period in days for which the token will be valid after
            rotation interval. Minimum value is 1.
          type: integer
          minimum: 1
          default: 30
      title: AutoRotate
      x-fern-type-name: AutoRotate
    NotificationTarget:
      description: +discriminator=type
      type: object
      oneOf:
        - $ref: '#/components/schemas/Email'
        - $ref: '#/components/schemas/SlackWebhook'
        - $ref: '#/components/schemas/SlackBot'
        - $ref: '#/components/schemas/PagerDuty'
      discriminator:
        propertyName: type
      title: NotificationTarget
      x-fern-type-name: NotificationTarget
    SecretStoreConfig:
      description: |-
        +label=Sync to Secret Manager
        +sort=6
        +usage=Enable to store the virtual account token in a secret store
        +message=Enable to store the virtual account token in a secret store
      type: object
      required:
        - integration_fqn
        - secret_path
      properties:
        integration_fqn:
          description: |-
            +label=Secret Store
            +usage=The secret store to store the virtual account token in
            +sort=1
            +uiType=IntegrationSelect
          type: string
        secret_path:
          description: >-
            +label=Secret Path

            +usage=The path where the virtual account token will be stored in
            the secret store (example: secret/path/to/virtual-account-token)

            +message=The path should only contain alphanumeric characters,
            hyphens, and slashes and it should start with a slash if giving the
            path containing slashes in between

            +sort=2
          type: string
          pattern: ^/?[A-Za-z0-9/-]*[A-Za-z0-9]$
      title: SecretStoreConfig
      x-fern-type-name: SecretStoreConfig
    VirtualAccountOwnedBy:
      description: |-
        +label=Owner
        +usage=Names of the owners that own the virtual account
      type: object
      required:
        - team
      properties:
        account:
          description: |-
            +label=Account Name
            +sort=1
            +usage=The name of the account that owns this resource
            +placeholder=Search Account
            +uiType=Hidden
            +uiProps={"maxAllowedAccounts": 1}
          type: string
        team:
          description: |-
            +label=Team Name
            +sort=2
            +usage=The name of the team that owns this resource
            +placeholder=Search Team
            +uiType=UserSelect
            +uiProps={"optionTypes": ["teams"], "maxAllowedUsers": 1}
          type: string
      title: VirtualAccountOwnedBy
      x-fern-type-name: VirtualAccountOwnedBy
    OwnedBy:
      description: |-
        +label=Owner
        +usage=Names of the owners that own this resource
      type: object
      required:
        - account
      properties:
        account:
          description: |-
            +label=Account Name
            +sort=1
            +usage=The name of the account that owns this resource
            +placeholder=Search Account
            +uiType=AccountSelect
            +uiProps={"maxAllowedAccounts": 1}
          type: string
      title: OwnedBy
      x-fern-type-name: OwnedBy
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
    Email:
      description: +label=Email
      type: object
      required:
        - type
        - notification_channel
        - to_emails
      properties:
        type:
          description: +value=email
          type: string
          enum:
            - email
        notification_channel:
          description: |-
            +label=Notification Channel
            +usage=Specify the notification channel to send alerts to
            +uiType=AlertNotificationChannel
            +uiProps={"integrationType":"notification-channel"}
            +sort=660
          type: string
          minLength: 1
        to_emails:
          description: >-
            +label=To Emails

            +usage=List of recipients' email addresses if the notification
            channel is Email.

            +docs=Specify the emails to send alerts to

            +sort=665
          type: array
          items:
            type: string
            minLength: 1
      title: Email
      x-fern-type-name: Email
    SlackWebhook:
      description: +label=Slack Webhook
      type: object
      required:
        - type
        - notification_channel
      properties:
        type:
          description: +value=slack-webhook
          type: string
          enum:
            - slack-webhook
        notification_channel:
          description: |-
            +label=Notification Channel
            +usage=Specify the notification channel to send alerts to
            +uiType=AlertNotificationChannel
            +uiProps={"integrationType":"notification-channel"}
            +sort=660
          type: string
          minLength: 1
      title: SlackWebhook
      x-fern-type-name: SlackWebhook
    SlackBot:
      type: object
      required:
        - type
        - notification_channel
        - channels
      properties:
        type:
          description: +value=slack-bot
          type: string
          enum:
            - slack-bot
        notification_channel:
          description: |-
            +label=Notification Channel
            +usage=Specify the notification channel to send alerts to
            +uiType=AlertNotificationChannel
            +uiProps={"integrationType":"notification-channel"}
            +sort=660
          type: string
          minLength: 1
        channels:
          description: >-
            +label=Slack Channel Names

            +usage=List of channels to send messages to.

            +message=3 to 81 lower case characters long alphanumeric word, may
            contain - or _ in between, must start with #

            +sort=665
          type: array
          items:
            type: string
            pattern: ^#[a-z0-9\-_]{2,80}$
      title: SlackBot
      x-fern-type-name: SlackBot
    PagerDuty:
      type: object
      required:
        - type
        - notification_channel
      properties:
        type:
          description: +value=pagerduty
          type: string
          enum:
            - pagerduty
        notification_channel:
          description: |-
            +label=Notification Channel
            +usage=Specify the notification channel to send alerts to
            +uiType=AlertNotificationChannel
            +uiProps={"integrationType":"notification-channel"}
            +sort=660
          type: string
          minLength: 1
      title: PagerDuty
      x-fern-type-name: PagerDuty
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````