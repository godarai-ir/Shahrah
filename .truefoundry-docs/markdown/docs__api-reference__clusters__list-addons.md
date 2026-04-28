> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# List Addons

> List addons for the provided cluster. Pagination is available based on query parameters.



## OpenAPI

````yaml /openapi.json get /api/svc/v1/clusters/{id}/get-addons
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
  /api/svc/v1/clusters/{id}/get-addons:
    get:
      tags:
        - Clusters
      summary: List Addons
      description: >-
        List addons for the provided cluster. Pagination is available based on
        query parameters.
      operationId: Cluster.getAddOns
      parameters:
        - name: id
          required: true
          in: path
          description: Cluster id of the cluster
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
      responses:
        '200':
          description: Returns a paginated list of addons for the cluster
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListClusterAddonsResponse'
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
    ListClusterAddonsResponse:
      type: object
      properties:
        data:
          description: Array of Cluster Addons
          type: array
          items:
            $ref: '#/components/schemas/AddonComponent'
        pagination:
          description: Pagination information
          allOf:
            - $ref: '#/components/schemas/Pagination'
      required:
        - data
        - pagination
      title: ListClusterAddonsResponse
      x-fern-type-name: ListClusterAddonsResponse
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
    AddonComponent:
      type: object
      properties:
        name:
          allOf:
            - $ref: '#/components/schemas/AddonComponentName'
        appName:
          type: string
        namespace:
          type: string
        applicationId:
          type: string
        description:
          type: string
        path:
          type: string
        addonFolder:
          type: string
        installed:
          type: boolean
        status:
          $ref: '#/components/schemas/AddonComponentStatus'
        version:
          type: string
        manifest:
          type: object
        installationSource:
          type: string
        unsupportedClusterTypes:
          type: array
          items:
            $ref: '#/components/schemas/ClusterType'
        required:
          type: boolean
        knownCRDs:
          type: array
          items:
            type: string
        source:
          $ref: '#/components/schemas/AddOnComponentSource'
        upgrades:
          $ref: '#/components/schemas/UpgradeData'
        labels:
          type: array
          items:
            type: string
        recommendations:
          description: Recommendations
          type: array
          items:
            $ref: '#/components/schemas/Recommendation'
        workspaceId:
          type: string
        metadata:
          type: object
          description: >-
            Addon metadata. autopilotEnabled is true when the backing addon
            application has autopilot enabled; otherwise false.
      required:
        - name
      title: AddonComponent
      x-fern-type-name: AddonComponent
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
    AddonComponentName:
      type: string
      enum:
        - ARGOCD
        - ARGO_ROLLOUT
        - ARGO_WORKFLOW
        - TFY_AGENT
        - ISTIO_BASE
        - ISTIO_DISCOVERY
        - TFY_ISTIO_INGRESS
        - AWS_LOAD_BALANCER_CONTROLLER
        - CERT_MANAGER
        - KARPENTER
        - KARPENTER_CONFIG
        - PROMETHEUS
        - METRICS_SERVER
        - KUBECOST
        - KEDA
        - AWS_EFS_CSI_DRIVER
        - AWS_EBS_CSI_DRIVER
        - TFY_GPU_OPERATOR
        - TFY_INFERENTIA_OPERATOR
        - ELASTI
        - TFY_WORKFLOW_PROPELLER
        - GRAFANA
        - TRUEFOUNDRY
        - TFY_PROMETHEUS_CONFIG
        - SPARK_OPERATOR
        - TFY_LOGS
        - TFY_CERT_MANAGER_CONFIG
      title: AddonComponentName
      x-fern-type-name: AddonComponentName
    AddonComponentStatus:
      type: object
      properties:
        healthStatus:
          type: string
          description: Health status of the addon
        syncStatus:
          type: string
          description: Sync status of the addon
        installed:
          type: boolean
          description: Indicates whether the addon is installed
      required:
        - installed
      title: AddonComponentStatus
      x-fern-type-name: AddonComponentStatus
    ClusterType:
      type: string
      enum:
        - aws-eks
        - gcp-gke-standard
        - azure-aks
        - generic
        - civo-talos
      title: ClusterType
      x-fern-type-name: ClusterType
    AddOnComponentSource:
      type: object
      properties:
        repo_url:
          type: string
          description: Repository URL of the source
        chart:
          type: string
          description: Chart of the source
      required:
        - repo_url
        - chart
      title: AddOnComponentSource
      x-fern-type-name: AddOnComponentSource
    UpgradeData:
      type: object
      properties:
        diff:
          type: array
          items:
            $ref: '#/components/schemas/IChange'
        currentManifest:
          oneOf:
            - $ref: '#/components/schemas/Helm'
        desiredManifest:
          oneOf:
            - $ref: '#/components/schemas/Helm'
        upgradePossible:
          type: boolean
        conflictFields:
          type: array
          items:
            type: string
        hasConflict:
          type: boolean
      title: UpgradeData
      x-fern-type-name: UpgradeData
    Recommendation:
      type: object
      properties:
        id:
          type: string
        clusterId:
          type: string
        applicationId:
          type: string
        deploymentId:
          type: string
        applicationVersion:
          type: integer
        recommendationData:
          type: object
          additionalProperties: true
        recommendationType:
          type: string
        appliedDeploymentId:
          type: string
        expiryTimestamp:
          format: date-time
          type: string
        createdAt:
          format: date-time
          type: string
        updatedAt:
          format: date-time
          type: string
      required:
        - recommendationData
        - recommendationType
        - expiryTimestamp
      title: Recommendation
      x-fern-type-name: Recommendation
    IChange:
      type: object
      properties:
        type:
          allOf:
            - $ref: '#/components/schemas/IChangeOperation'
        key:
          type: string
        value:
          type: object
        oldValue:
          type: object
        changes:
          type: array
          items:
            $ref: '#/components/schemas/IChange'
      required:
        - type
        - key
      title: IChange
      x-fern-type-name: IChange
    Helm:
      type: object
      required:
        - type
        - name
        - source
      properties:
        type:
          description: +value=helm
          type: string
          enum:
            - helm
        name:
          description: >-
            +sort=1

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +usage=Name of the Helm deployment. This will be set as the release
            name of the chart you are deploying.
          type: string
          pattern: ^[a-z](?:[a-z0-9]|-(?!-)){1,30}[a-z0-9]$
        labels:
          description: |-
            +label=Labels
            +usage=Add labels to base argo app
          type: object
          additionalProperties:
            type: string
        source:
          description: |-
            +label=Source helm repository
            +sort=2
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/HelmRepo'
            - $ref: '#/components/schemas/OCIRepo'
            - $ref: '#/components/schemas/GitHelmRepo'
          discriminator:
            propertyName: type
        values:
          description: |-
            +label=Values
            +usage=Values file as block file
          type: object
        kustomize:
          $ref: '#/components/schemas/Kustomize'
        ignoreDifferences:
          type: array
          items:
            type: object
        workspace_fqn:
          description: |-
            +label=Workspace FQN
            +docs=Fully qualified name of the workspace
            +uiType=Hidden
          type: string
      title: Helm
      x-fern-type-name: Helm
    IChangeOperation:
      type: string
      enum:
        - REMOVE
        - ADD
        - UPDATE
      title: IChangeOperation
      x-fern-type-name: IChangeOperation
    HelmRepo:
      type: object
      required:
        - type
        - repo_url
        - chart
        - version
      properties:
        type:
          description: +value=helm-repo
          type: string
          enum:
            - helm-repo
        repo_url:
          description: |-
            +label=Helm repository URL
            +sort=1
            +message=Needs to be a valid URL.
          type: string
          pattern: >-
            ^(((https?|wss):\/\/)?(?:www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}(?:[-a-zA-Z0-9()@:%_\+.~#?&\/=]*))$
        integration_fqn:
          description: >-
            +docs=FQN of the helm repo integration. You can use the FQN of your
            desired helm integration (or add one)

            in the  Integrations
            page[Integrations](https://app.truefoundry.tech/integrations) page

            +label=Helm repo integration

            +sort=2

            +usage=FQN of the helm repo integration. If you can't find your
            integration here,

            add it through the [Integrations](/integrations) page

            +uiType=IntegrationSelect

            +uiProps={"integrationType":"helm-repo"}
          type: string
        chart:
          description: |-
            +label=Chart name
            +sort=3
            +usage=The helm chart name
            +uiType=InputSelect
            +uiProps={"creatable":true, "searchable":true}
          type: string
        version:
          description: |-
            +label=Version
            +sort=4
            +usage=Helm chart version
            +uiType=InputSelect
            +uiProps={"creatable":true, "searchable":true}
          type: string
      title: HelmRepo
      x-fern-type-name: HelmRepo
    OCIRepo:
      description: +label=OCIRepo
      type: object
      required:
        - type
        - oci_chart_url
        - version
      properties:
        type:
          description: +value=oci-repo
          type: string
          enum:
            - oci-repo
        oci_chart_url:
          description: |-
            +label=OCI chart URL
            +message=Need to be a valid URL.
          type: string
          pattern: >-
            ^(((oci):\/\/)?(?:www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}(?:[-a-zA-Z0-9()@:%_\+.~#?&\/=]*))$
        integration_fqn:
          description: >-
            +docs=FQN of the container registry. You can use the FQN of your
            desired container registry (or add one)

            in the  Integrations
            page[Integrations](https://app.truefoundry.tech/integrations) page

            +label=Container Registry

            +usage=FQN of the container registry. If you can't find your
            registry here,

            add it through the [Integrations](/integrations) page

            +uiType=IntegrationSelect

            +uiProps={"integrationType":"docker-registry"}
          type: string
        version:
          description: |-
            +label=Version
            +usage=Helm chart version
          type: string
      title: OCIRepo
      x-fern-type-name: OCIRepo
    GitHelmRepo:
      type: object
      required:
        - type
        - repo_url
        - revision
        - path
      properties:
        type:
          description: +value=git-helm-repo
          type: string
          enum:
            - git-helm-repo
        repo_url:
          description: |-
            TODO: Check this regex and add guidelines
            +label=Git repository URL
            +sort=1
            +message=Needs to be a valid URL.
          type: string
          pattern: >-
            ^(((https?|wss):\/\/)?(?:www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}(?:[-a-zA-Z0-9()@:%_\+.~#?&\/=]*))$
        revision:
          description: |-
            +label=Revision
            +sort=2
            +usage=Branch/Commit SHA/Tag of the git repo.
          type: string
        path:
          description: |-
            +label=Path
            +sort=3
            +usage=Path to the chart.
          type: string
        value_files:
          description: |-
            +label=Value files
            +sort=3
            +usage=Helm values files for overriding values in the helm chart.
            The path is relative to the Path directory defined above
          type: array
          items:
            type: string
      title: GitHelmRepo
      x-fern-type-name: GitHelmRepo
    Kustomize:
      type: object
      properties:
        patch:
          description: >-
            +label=Patch

            +usage=Content of kustomization.yaml to perform kustomize operation.
            Please do not include the `resources` section. It is filled in
            automatically
          type: object
        additions:
          description: >-
            +label=Additional Manifests

            +usage=Additional kubernetes manifests to be included in the
            application
          type: array
          items:
            type: object
      title: Kustomize
      x-fern-type-name: Kustomize
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````