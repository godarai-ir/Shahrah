> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Get Cluster

> Get cluster associated with provided id



## OpenAPI

````yaml /openapi.json get /api/svc/v1/clusters/{id}
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
  /api/svc/v1/clusters/{id}:
    get:
      tags:
        - Clusters
      summary: Get Cluster
      description: Get cluster associated with provided id
      operationId: Cluster.get
      parameters:
        - name: id
          required: true
          in: path
          description: Cluster id of the cluster
          schema:
            type: string
      responses:
        '200':
          description: Return the cluster associated with provided id
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetClusterResponse'
        '401':
          description: 'Error: Unauthorized'
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '404':
          description: Invalid cluster id or cluster does not exist
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
      security:
        - jwt: []
components:
  schemas:
    GetClusterResponse:
      type: object
      properties:
        data:
          description: Cluster
          allOf:
            - $ref: '#/components/schemas/Cluster'
      required:
        - data
      title: GetClusterResponse
      x-fern-type-name: GetClusterResponse
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
    Cluster:
      type: object
      properties:
        id:
          type: string
        fqn:
          type: string
        manifest:
          oneOf:
            - $ref: '#/components/schemas/ClusterManifest'
        tenantName:
          type: string
        accountId:
          type: string
          nullable: false
        createdBySubject:
          $ref: '#/components/schemas/Subject'
        createdAt:
          format: date-time
          type: string
        updatedAt:
          format: date-time
          type: string
        createdBy:
          type: string
      required:
        - id
        - fqn
        - manifest
        - tenantName
        - accountId
        - createdAt
        - updatedAt
      title: Cluster
      x-fern-type-name: Cluster
    ClusterManifest:
      type: object
      required:
        - type
        - name
        - cluster_type
        - environment_names
        - collaborators
      properties:
        type:
          description: +value=cluster
          type: string
          enum:
            - cluster
        name:
          description: >-
            +label=Name

            +icon=fa-desktop:#black

            +sort=10

            +message=3 to 35 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,33}[a-z0-9]$
        cluster_type:
          description: |-
            +label=Cluster Type
            +icon=fa-cloud:#7EC8E3
            +sort=20
            +usage=The cluster type of this cluster
          type: string
          enum:
            - aws-eks
            - gcp-gke-standard
            - azure-aks
            - generic
            - civo-talos
        environment_names:
          description: >-
            +label=Environments

            +icon=fa-tags:#black

            +usage=Tag the environment like dev, staging or production. You will
            need to [create
            environments](https://docs.truefoundry.com/docs/key-concepts#creating-environments)
            if you don't have already.

            +sort=30
          type: array
          items:
            type: string
        base_domains:
          description: >-
            +label=Base Domain URLs

            +placeholder=Enter Base Domain URL eg. devtest.mydomain.com or
            *.mydomain.com

            +usage=One or more base URLs, which can be either a wildcard domain
            (resulting in app URLs like
            `<app-name>-<workspace-name>.<base-domain-URL>`) or a non-wildcard
            domain (resulting in app URLs like
            `<base-domain-URL>/<app-name>-<workspace-name>`)

            > Read more about routing
            [here](https://docs.truefoundry.com/docs/routing) and about 'A'
            record and domain mapping
            [here](https://docs.truefoundry.com/docs/setting-up-domain)

            +icon=fa-globe:#68BBE3

            +sort=40

            +message=Domain URL should be valid domain or IP address
          type: array
          items:
            type: string
            pattern: >-
              ^((((\*|[a-zA-Z0-9\-]{1,63})\.)([a-zA-Z0-9\-]{1,63}\.)*([A-Za-z]{1,63}))|(((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)))$
        monitoring:
          description: |-
            +label=Monitoring
            +icon=fa-gear:#68BBE3
            +sort=50
          type: object
          properties:
            loki_url:
              description: +label=Cluster Loki URL
              type: string
            victoria_logs_url:
              description: +label=Cluster VictoriaLogs URL
              type: string
            prometheus_url:
              description: +label=Cluster Prometheus URL
              type: string
            kubecost_url:
              description: +label=Cluster Kubecost URL
              type: string
        default_registry_fqn:
          description: |-
            +label=Cluster Default Registry
            +sort=60
          type: string
        workbench_config:
          description: |-
            +label=Workbench Config
            +sort=70
          type: object
          properties:
            notebook_config:
              $ref: '#/components/schemas/NotebookConfig'
            ssh_server_config:
              $ref: '#/components/schemas/SSHServerConfig'
            default_storage_class:
              description: >-
                +label=Default Storage Class

                +usage=The default storage class for the home directory of
                workbench

                +message=Must not contain any spaces.
              type: string
              pattern: ^[^\s]*$
        spark_config:
          $ref: '#/components/schemas/SparkConfig'
        ingress_controller_config:
          $ref: '#/components/schemas/IngressControllerConfig'
        cluster_integration_fqn:
          description: >-
            +label=Cluster Integration FQN

            +sort=75

            +uiType=IntegrationSelect

            +uiProps={"descriptionKey": "cluster_name","integrationType":
            "cluster"}
          type: string
        workflow_storage_integration_fqn:
          description: >-
            +label=Workflow Storage Integration

            +sort=80

            +uiType=IntegrationSelect

            +uiProps={"descriptionKey": "cluster_name","integrationType":
            "blob-storage"}
          type: string
        supported_nodepools:
          description: >-
            +label=Nodepools

            +usage=Add nodepools that are already created in your cluster.

            When deploying, applications can choose to schedule from these
            nodepools.

            +sort=90
          type: array
          items:
            $ref: '#/components/schemas/Nodepool'
        node_label_keys:
          description: >-
            +label=Node Label Keys

            +usage=The node label keys that this cluster supports.

            Note: You will additionally need to add `truefoundry.com/gpu_type`
            label for GPU-supported node pools.

            For more information, check out [this
            documentation](https://docs.truefoundry.com/docs/generic-control-plane#configuring-node-pools-for-truefoundry).

            +sort=100
          type: object
          required:
            - nodepool_selector_label
          properties:
            nodepool_selector_label:
              description: |-
                +label=Nodepool Selector Label
                +usage=The nodepool selector label that this cluster supports
              type: string
        collaborators:
          description: |-
            +label=Collaborators
            +usage=Collaborators who can access this cluster
            +sort=110
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: ClusterManifest
      x-fern-type-name: ClusterManifest
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
    NotebookConfig:
      type: object
      required:
        - base_domain
      properties:
        base_domain:
          description: >-
            +label=Notebook Base Domain

            +usage=The base domain for the cluster with which you can access
            your Notebooks

            +message=Must not contain any spaces.
          type: string
          pattern: ^[^\s]*$
      title: NotebookConfig
      x-fern-type-name: NotebookConfig
    SSHServerConfig:
      type: object
      required:
        - base_domain
        - port
      properties:
        base_domain:
          description: >-
            +label=SSH Base Domain

            +usage=The base domain for the cluster with which you can access
            your SSH containers

            +sort=1000

            +message=Must not contain any spaces.
          type: string
          pattern: ^[^\s]*$
        port:
          description: >-
            +label=SSH Port

            +usage=The port for the cluster with which you can access your SSH
            containers

            +sort=2000
          type: number
      title: SSHServerConfig
      x-fern-type-name: SSHServerConfig
    SparkConfig:
      type: object
      required:
        - ui_base_domain
      properties:
        ui_base_domain:
          description: >-
            +label=Spark UI Base Domain

            +usage=The base domain for the cluster with which you can access
            your Spark UI

            +message=Must not contain any spaces.
          type: string
          pattern: ^[^\s]*$
      title: SparkConfig
      x-fern-type-name: SparkConfig
    IngressControllerConfig:
      type: object
      required:
        - ingress_class_name
      properties:
        ingress_class_name:
          description: +label=Ingress Class Name
          type: string
          default: nginx
        tls_enabled:
          description: >-
            +label=TLS Enabled

            +usage=Whether TLS is managed by the ingress controller. If enabled,
            the ingress object will have TLS configuration.
          type: boolean
          default: false
      title: IngressControllerConfig
      x-fern-type-name: IngressControllerConfig
    Nodepool:
      description: >-
        +label=Configured Nodepools

        +usage=The nodepools that are already created in your cluster. This will
        be used to schedule your workloads on particular nodepools.
      type: object
      required:
        - name
      properties:
        name:
          type: string
        description:
          type: string
      title: Nodepool
      x-fern-type-name: Nodepool
    Collaborator:
      type: object
      required:
        - subject
        - role_id
      properties:
        subject:
          description: >-
            +label=Subject FQN

            +usage=Fully Qualified Name of the subject. eg: user:email or
            team:teamname
          type: string
        role_id:
          description: |-
            +label=Role ID
            +usage=Role ID for the resource
          type: string
      title: Collaborator
      x-fern-type-name: Collaborator
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
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````