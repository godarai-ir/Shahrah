> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# List Associated Active Deployments

> List the active deployments that are associated with a secret.



## OpenAPI

````yaml /openapi.json get /api/svc/v1/secrets/{id}/deployments
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
  /api/svc/v1/secrets/{id}/deployments:
    get:
      tags:
        - Secrets
      summary: List Associated Active Deployments
      description: List the active deployments that are associated with a secret.
      operationId: Secret.getAssociatedActiveDeployments
      parameters:
        - name: id
          required: true
          in: path
          description: Secret Id of the secret.
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
          description: Returns the active deployments that are associated with a secret.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetAssociatedDeploymentsResponse'
        '403':
          description: Forbidden. User does not have permission to read the secret.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '404':
          description: Not Found. Secret with the provided id is not found.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
      security:
        - jwt: []
components:
  schemas:
    GetAssociatedDeploymentsResponse:
      type: object
      properties:
        data:
          description: Associated Deployments
          type: array
          items:
            $ref: '#/components/schemas/Deployment'
        pagination:
          description: Pagination Information
          allOf:
            - $ref: '#/components/schemas/Pagination'
      required:
        - data
        - pagination
      title: GetAssociatedDeploymentsResponse
      x-fern-type-name: GetAssociatedDeploymentsResponse
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
    Deployment:
      type: object
      properties:
        id:
          type: string
        version:
          type: number
        fqn:
          type: string
        applicationId:
          type: string
        manifest:
          oneOf:
            - $ref: '#/components/schemas/Service'
            - $ref: '#/components/schemas/AsyncService'
            - $ref: '#/components/schemas/Job'
            - $ref: '#/components/schemas/Notebook'
            - $ref: '#/components/schemas/Codeserver'
            - $ref: '#/components/schemas/SSHServer'
            - $ref: '#/components/schemas/RStudio'
            - $ref: '#/components/schemas/Helm'
            - $ref: '#/components/schemas/Volume'
            - $ref: '#/components/schemas/ApplicationSet'
            - $ref: '#/components/schemas/Workflow'
            - $ref: '#/components/schemas/SparkJob'
            - $ref: '#/components/schemas/Intercept'
        application:
          $ref: '#/components/schemas/Application'
        createdBySubject:
          $ref: '#/components/schemas/Subject'
        createdAt:
          format: date-time
          type: string
        updatedAt:
          format: date-time
          type: string
        deploymentBuilds:
          type: array
          items:
            $ref: '#/components/schemas/BuildInfo'
        deploymentStatuses:
          type: array
          items:
            $ref: '#/components/schemas/DeploymentStatus'
        currentStatusId:
          type: string
        currentStatus:
          $ref: '#/components/schemas/DeploymentStatus'
        appliedRecommendations:
          description: Applied recommendations for this deployment
          type: array
          items:
            $ref: '#/components/schemas/Recommendation'
        createdBy:
          type: string
      required:
        - manifest
        - createdBySubject
      title: Deployment
      x-fern-type-name: Deployment
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
    Service:
      description: +docs=Describes the configuration for the service
      type: object
      properties:
        type:
          description: +value=service
          type: string
          enum:
            - service
        replicas:
          description: >-
            +label=Replicas

            +usage=Deploy multiple instances of your pods to distribute incoming
            traffic across them, ensuring effective load balancing.

            +icon=fa-clone

            +sort=4
          default: 1
          oneOf:
            - type: number
              minimum: 0
              maximum: 500
            - $ref: '#/components/schemas/ServiceAutoscaling'
        auto_shutdown:
          $ref: '#/components/schemas/Autoshutdown'
        allow_interception:
          description: >-
            +label=Allow intercepts

            +usage=Whether to allow intercepts to be applied for this service.

            This would inject an additional sidecar in each pod of the service.
            Not recommended on production
          type: boolean
          default: false
        rollout_strategy:
          description: >-
            +label=Rollout strategy

            +usage=Strategy to dictate how a rollout should happen when a new
            release for this service is made
            [Docs](https://docs.truefoundry.com/docs/rollout-strategy)

            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/Rolling'
            - $ref: '#/components/schemas/Canary'
            - $ref: '#/components/schemas/BlueGreen'
          discriminator:
            propertyName: type
      allOf:
        - $ref: '#/components/schemas/BaseService'
      required:
        - type
        - replicas
      title: Service
      x-fern-type-name: Service
    AsyncService:
      description: +docs=Describes the configuration for the async-service
      type: object
      properties:
        type:
          description: +value=async-service
          type: string
          enum:
            - async-service
        replicas:
          description: >-
            +label=Replicas

            +usage=Deploy multiple instances of your pods to distribute incoming
            traffic across them, ensuring effective load balancing.

            +icon=fa-clone
          default: 1
          oneOf:
            - type: number
              minimum: 0
              maximum: 500
            - $ref: '#/components/schemas/AsyncServiceAutoscaling'
        rollout_strategy:
          $ref: '#/components/schemas/Rolling'
        worker_config:
          $ref: '#/components/schemas/WorkerConfig'
        sidecar:
          $ref: '#/components/schemas/AsyncProcessorSidecar'
      allOf:
        - $ref: '#/components/schemas/BaseService'
      required:
        - type
        - replicas
        - worker_config
      title: AsyncService
      x-fern-type-name: AsyncService
    Job:
      description: +docs=Describes the configuration for the job
      type: object
      required:
        - type
        - name
        - image
        - trigger
      properties:
        type:
          description: +value=job
          type: string
          enum:
            - job
        name:
          description: >-
            +usage=Name of the job

            +sort=1

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z](?:[a-z0-9]|-(?!-)){1,30}[a-z0-9]$
        image:
          description: >-
            +docs=Specify whether you want to deploy a Docker image or build and
            deploy from source code

            +label=Deploy a Docker image or build and deploy from source code

            +icon=fa-solid fa-cloud-arrow-up:#21B6A8

            +sort=200

            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/Build'
            - $ref: '#/components/schemas/Image'
          discriminator:
            propertyName: type
        trigger:
          description: |-
            +docs=Specify the trigger
            +sort=300
            +discriminator=type
          type: object
          default:
            type: manual
          oneOf:
            - $ref: '#/components/schemas/Manual'
            - $ref: '#/components/schemas/Schedule'
          discriminator:
            propertyName: type
        trigger_on_deploy:
          description: |-
            +docs=Trigger on deploy
            +sort=350
            +usage=Trigger the job after deploy immediately
          type: boolean
        params:
          description: |-
            +label=Params for input
            +uiType=StructsTable
            +usage=Configure params and pass it to create different job runs
            +sort=400
          type: array
          items:
            $ref: '#/components/schemas/Param'
        env:
          description: >-
            +label=Environment Variables

            +usage=Configure environment variables to be injected in the service
            either as plain text or secrets.
            [Docs](https://docs.truefoundry.com/docs/env-variables)

            +icon=fa-globe

            +sort=500
          type: object
          default: null
          additionalProperties:
            type: string
          nullable: true
        resources:
          $ref: '#/components/schemas/Resources'
        alerts:
          description: >-
            +label=Alerts

            +usage=Configure alerts to be sent when the job
            starts/fails/completes

            +icon=fa-bell

            +sort=650
          type: array
          items:
            $ref: '#/components/schemas/JobAlert'
        retries:
          description: >-
            +label=Retries

            +usage=Specify the maximum number of attempts to retry a job before
            it is marked as failed.

            +icon=fa-repeat

            +sort=700
          type: integer
          minimum: 0
          maximum: 10
          default: 0
        timeout:
          description: |-
            +label=Timeout
            +usage=Job timeout in seconds.
            +icon=fa-clock
            +sort=800
          type: integer
          minimum: 0
          exclusiveMinimum: true
          maximum: 432000
        concurrency_limit:
          description: |-
            +label=Concurrency Limit
            +usage=Number of runs that can run concurrently
            +icon=fa-copy
            +sort=900
          type: integer
          minimum: 0
          exclusiveMinimum: true
        service_account:
          description: +sort=1000
          type: string
        mounts:
          description: >-
            +usage=Configure data to be mounted to job pod(s) as a string,
            secret or volume.
            [Docs](https://docs.truefoundry.com/docs/mounting-volumes-job)

            +todo=Add discriminator for union types in array
          type: array
          items:
            type: object
            oneOf:
              - $ref: '#/components/schemas/SecretMount'
              - $ref: '#/components/schemas/StringDataMount'
              - $ref: '#/components/schemas/VolumeMount'
        labels:
          description: +label=Labels
          type: object
          additionalProperties:
            type: string
        kustomize:
          $ref: '#/components/schemas/Kustomize'
        workspace_fqn:
          description: |-
            +label=Workspace FQN
            +docs=Fully qualified name of the workspace
            +uiType=Hidden
          type: string
      title: Job
      x-fern-type-name: Job
    Notebook:
      description: +docs=Describes the configuration for the service
      type: object
      properties:
        type:
          description: +value=notebook
          type: string
          enum:
            - notebook
        image:
          $ref: '#/components/schemas/WorkbenchImage'
        cull_timeout:
          description: >-
            +label=Stop after (minutes of inactivity)

            +usage=Stop the notebook instance after this much time in minutes of
            inactivity.

            The notebook instance will be stopped even if the notebook is open
            in your browser, but nothing is running on the notebook.

            +sort=5

            +uiProps={"descriptionInline":true}
          type: integer
          minimum: 5
          default: 30
      allOf:
        - $ref: '#/components/schemas/BaseWorkbenchInput'
      required:
        - type
        - image
      title: Notebook
      x-fern-type-name: Notebook
    Codeserver:
      description: +docs=Describes the configuration for the code server
      type: object
      properties:
        type:
          description: +value=codeserver
          type: string
          enum:
            - codeserver
        image:
          $ref: '#/components/schemas/WorkbenchImage'
      allOf:
        - $ref: '#/components/schemas/BaseWorkbenchInput'
      required:
        - type
        - image
      title: Codeserver
      x-fern-type-name: Codeserver
    SSHServer:
      description: +docs=Describes the configuration for the ssh server
      type: object
      properties:
        type:
          description: +value=ssh-server
          type: string
          enum:
            - ssh-server
        image:
          $ref: '#/components/schemas/WorkbenchImage'
        ssh_public_key:
          description: >-
            +label: SSH Public Key

            +usage=Add Your SSH Public Key, this will be used to authenticate
            you to the SSH Server.  \

            You can find it using `cat ~/.ssh/id_rsa.pub` in Mac/Linux or `type
            $home\.ssh\id_rsa.pub` in Windows Powershell.  \

            You can also generate a new SSH key pair using `ssh-keygen -t rsa`
            in your local terminal. (same for both Mac/Linux and Windows
            Powershell)

            +uiType=TextArea

            +sort=4
          type: string
        cull_timeout:
          description: >-
            +label=Stop after (minutes of inactivity)

            +usage=Stop the SSH Server instance after this much time in minutes
            of inactivity. The instance is considered active if there is at
            least one active SSH connection (a client connected to the SSH
            server), or if a background job is running using tmux or screen, or
            if the pod has restarted.

            +sort=5

            +uiProps={"descriptionInline":true, "warningMessage":"Please note
            that stop after inactivity is only available for images with
            tag(including custom images) >= v0.3.10"}
          type: integer
          minimum: 5
      allOf:
        - $ref: '#/components/schemas/BaseWorkbenchInput'
      required:
        - type
        - image
        - ssh_public_key
      title: SSHServer
      x-fern-type-name: SSHServer
    RStudio:
      description: +docs=Describes the configuration for the Rstudio server
      type: object
      properties:
        type:
          description: +value=rstudio
          type: string
          enum:
            - rstudio
        image:
          $ref: '#/components/schemas/WorkbenchImage'
      allOf:
        - $ref: '#/components/schemas/BaseWorkbenchInput'
      required:
        - type
        - image
      title: RStudio
      x-fern-type-name: RStudio
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
    Volume:
      type: object
      required:
        - type
        - name
        - config
      properties:
        type:
          description: +value=volume
          type: string
          enum:
            - volume
        name:
          description: >-
            +sort=1

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +usage=Name of the Volume. This will be set as the volume name.
          type: string
          pattern: ^[a-z](?:[a-z0-9]|-(?!-)){1,30}[a-z0-9]$
        config:
          description: >-
            +sort=2

            +label=Volume Config

            +message=Volume Configuration, can be either Dynamically provisioned
            or statically provisioned.

            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/DynamicVolumeConfig'
            - $ref: '#/components/schemas/StaticVolumeConfig'
          discriminator:
            propertyName: type
        volume_browser:
          $ref: '#/components/schemas/VolumeBrowser'
        workspace_fqn:
          description: |-
            +label=Workspace FQN
            +docs=Fully qualified name of the workspace
            +uiType=Hidden
          type: string
      title: Volume
      x-fern-type-name: Volume
    ApplicationSet:
      description: +docs=Describes the configuration for the application set
      type: object
      required:
        - type
        - name
      properties:
        type:
          description: +value=application-set
          type: string
          enum:
            - application-set
        name:
          description: |-
            +label=Name
            +usage=Name of the application set.
            +icon=fa-font
          type: string
        components:
          description: |-
            +label=Components
            +usage=Array of components with their specifications.
            +icon=fa-puzzle-piece
            +uiType=AppComponents
            +todo=Add discriminator for union types in array
          type: array
          items:
            type: object
            oneOf:
              - $ref: '#/components/schemas/Service'
              - $ref: '#/components/schemas/AsyncService'
              - $ref: '#/components/schemas/Job'
              - $ref: '#/components/schemas/Helm'
        template:
          description: |-
            +label=Template
            +usage=Template to be used for the application set.
            +icon=fa-file
            +uiType=Hidden
          type: string
        values:
          description: >-
            +label=Values

            +usage=Values to be used to render components for the application
            set.

            +icon=fa-file

            +uiType=YamlInput

            +uiProps={"allowAllValues":true}
          type: object
        workspace_fqn:
          description: |-
            +label=Workspace FQN
            +docs=Fully qualified name of the workspace
            +uiType=Hidden
          type: string
        convert_template_manifest:
          description: >-
            +label=Convert Template Manifest

            +docs=Flag to indicate if the template manifest should be converted
            to TrueFoundry manifest

            +uiType=Hidden
          type: boolean
      title: ApplicationSet
      x-fern-type-name: ApplicationSet
    Workflow:
      description: +docs=Describes the configuration for the worflow
      type: object
      required:
        - type
        - name
        - source
        - workflow_file_path
      properties:
        type:
          description: +value=workflow
          type: string
          enum:
            - workflow
        name:
          description: >-
            +usage=Name of the workflow

            +sort=1

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z](?:[a-z0-9]|-(?!-)){1,30}[a-z0-9]$
        source:
          description: |-
            +docs=Source Code for the workflow, either local or remote
            +label=Source Code for your workflow
            +icon=fa-solid fa-cloud-arrow-up:#21B6A8
            +sort=200
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/LocalSource'
            - $ref: '#/components/schemas/RemoteSource'
          discriminator:
            propertyName: type
        workflow_file_path:
          description: |-
            +label=Workflow File Path
            +docs=Path to the workflow file relative to the project root path
            +sort=550
          type: string
        flyte_entities:
          description: |-
            +label=Flyte Entities
            +todo=Add discriminator for union types in array
          type: array
          items:
            type: object
            oneOf:
              - $ref: '#/components/schemas/FlyteTask'
              - $ref: '#/components/schemas/FlyteWorkflow'
              - $ref: '#/components/schemas/FlyteLaunchPlan'
        alerts:
          description: +label=Alerts
          type: array
          items:
            $ref: '#/components/schemas/WorkflowAlert'
      title: Workflow
      x-fern-type-name: Workflow
    SparkJob:
      type: object
      required:
        - type
        - name
        - image
        - entrypoint
        - driver_config
        - executor_config
      properties:
        type:
          description: |-
            +value=spark-job
            +sort=1
          type: string
          enum:
            - spark-job
        name:
          description: >-
            +label=Name

            +usage=Name of the job

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +sort=2
          type: string
          pattern: ^[a-z](?:[a-z0-9]|-(?!-)){1,30}[a-z0-9]$
        image:
          description: >-
            +label=Deploy a Docker image & Specify Spark Version

            +usage=The image to use for driver and executors. Must have spark
            installed. Spark version must match the version in the image.

            +sort=500

            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/SparkImage'
            - $ref: '#/components/schemas/SparkImageBuild'
          discriminator:
            propertyName: type
        entrypoint:
          description: |-
            +label=Entrypoint
            +sort=1000
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/SparkJobPythonEntrypoint'
            - $ref: '#/components/schemas/SparkJobScalaEntrypoint'
            - $ref: '#/components/schemas/SparkJobJavaEntrypoint'
            - $ref: '#/components/schemas/SparkJobPythonNotebookEntrypoint'
            - $ref: '#/components/schemas/SparkJobScalaNotebookEntrypoint'
          discriminator:
            propertyName: type
        driver_config:
          $ref: '#/components/schemas/SparkDriverConfig'
        executor_config:
          $ref: '#/components/schemas/SparkExecutorConfig'
        env:
          description: >-
            +label=Environment Variables

            +usage=Configure environment variables to be injected in the service
            either as plain text.
            [Docs](https://docs.truefoundry.com/docs/env-variables)

            +icon=fa-globe

            +sort=21000
          type: object
          default: null
          nullable: true
        spark_conf:
          description: >-
            +label=Spark Config Properties

            +usage=Extra configuration properties to be passed to the spark job.
            [Docs](https://spark.apache.org/docs/latest/configuration.html)

            +icon=fa-gear:#68BBE3

            +sort=21500
          type: object
          default: null
          additionalProperties:
            type: string
            minLength: 1
          nullable: true
        mounts:
          description: >-
            +label=Mounts

            +usage=Configure volumes to be mounted to driver and executors.
            [Docs](https://docs.truefoundry.com/docs/mounting-volumes-job)

            +sort=22000

            +uiType=Mounts
          type: array
          items:
            $ref: '#/components/schemas/VolumeMount'
        retries:
          description: >-
            +label=Retries

            +usage=Specify the maximum number of attempts to retry a job before
            it is marked as failed.

            +icon=fa-repeat

            +sort=23000
          type: integer
          minimum: 0
          maximum: 10
          default: 0
        service_account:
          description: |-
            +label=Service Account
            +sort=24000
          type: string
        workspace_fqn:
          description: |-
            +label=Workspace FQN
            +docs=Fully qualified name of the workspace
            +uiType=Hidden
          type: string
      title: SparkJob
      x-fern-type-name: SparkJob
    Intercept:
      type: object
      required:
        - type
        - name
        - rules
      properties:
        type:
          description: +value=intercept
          type: string
          enum:
            - intercept
        name:
          description: >-
            +sort=1

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +usage=Name of the service this intercept is for
          type: string
          pattern: ^[a-z](?:[a-z0-9]|-(?!-)){1,30}[a-z0-9]$
        rules:
          type: array
          items:
            type: object
            required:
              - port
              - action
            properties:
              port:
                description: |-
                  +docs=Port of service to be intercepted
                  +label=Service port
                type: integer
                minimum: 0
                maximum: 65535
              match:
                $ref: '#/components/schemas/HeaderMatch'
              action:
                description: |-
                  +docs=Action to take on succesful match
                  +label=Action
                  +discriminator=type
                type: object
                oneOf:
                  - $ref: '#/components/schemas/ForwardAction'
                  - $ref: '#/components/schemas/MirrorAction'
                discriminator:
                  propertyName: type
      title: Intercept
      x-fern-type-name: Intercept
    Application:
      type: object
      properties:
        id:
          type: string
        fqn:
          type: string
        name:
          type: string
        type:
          allOf:
            - $ref: '#/components/schemas/ApplicationType'
        createdBySubject:
          $ref: '#/components/schemas/Subject'
        tenantName:
          type: string
        metadata:
          $ref: '#/components/schemas/ApplicationMetadata'
        lifecycleStage:
          allOf:
            - $ref: '#/components/schemas/ApplicationLifecycleStage'
        workspaceId:
          type: string
        lastVersion:
          type: integer
        activeVersion:
          type: integer
        createdAt:
          format: date-time
          type: string
        updatedAt:
          format: date-time
          type: string
        recommendations:
          description: Recommendations for this application
          type: array
          items:
            $ref: '#/components/schemas/Recommendation'
        alerts:
          description: Alerts for this application
          type: array
          items:
            $ref: '#/components/schemas/Alert'
        alertsSummary:
          type: object
          description: Summary of alerts for this application
          additionalProperties: true
        applicationDebugInfos:
          description: Debug infos for this application
          type: array
          items:
            $ref: '#/components/schemas/ApplicationDebugInfo'
        potentialProblems:
          description: Potential problems with the application
          type: array
          items:
            $ref: '#/components/schemas/ApplicationProblem'
        autopilot:
          type: object
          additionalProperties: true
        workspaceFqn:
          type: string
        createdBy:
          type: string
        deployment:
          $ref: '#/components/schemas/Deployment'
        activeDeploymentId:
          type: string
        lastDeploymentId:
          type: string
      required:
        - createdBySubject
        - lifecycleStage
        - autopilot
      title: Application
      x-fern-type-name: Application
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
    BuildInfo:
      type: object
      properties:
        name:
          type: string
        status:
          type: number
          enum:
            - 20
            - 30
            - 40
          description: |2-

                  Status of the build:
                    STARTED = 20,
                    SUCCEEDED = 30,
                    FAILED = 40
                
          example: 20
      required:
        - name
        - status
      title: BuildInfo
      x-fern-type-name: BuildInfo
    DeploymentStatus:
      type: object
      properties:
        id:
          type: string
        deploymentId:
          type: string
        status:
          allOf:
            - $ref: '#/components/schemas/DeploymentStatusValue'
        state:
          type: object
        transition:
          allOf:
            - $ref: '#/components/schemas/DeploymentTransition'
        message:
          type: string
        retryCount:
          type: number
        createdAt:
          format: date-time
          type: string
        updatedAt:
          format: date-time
          type: string
      title: DeploymentStatus
      x-fern-type-name: DeploymentStatus
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
    ServiceAutoscaling:
      type: object
      properties:
        metrics:
          description: |-
            +label=Autoscaling metrics
            +usage=Metrics to use for the autoscaler
            +sort=4
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/CPUUtilizationMetric'
            - $ref: '#/components/schemas/RPSMetric'
            - $ref: '#/components/schemas/CronMetric'
          discriminator:
            propertyName: type
      allOf:
        - $ref: '#/components/schemas/BaseAutoscaling'
      required:
        - metrics
      title: ServiceAutoscaling
      x-fern-type-name: ServiceAutoscaling
    Autoshutdown:
      type: object
      required:
        - wait_time
      properties:
        wait_time:
          description: >-
            +label=Wait Time

            +usage=The period to wait after the last received request before
            scaling the replicas to 0. This value should be high enough to allow
            for the replicas of the service to come up to avoid premature
            scaling down.
          type: integer
          minimum: 0
          maximum: 604800
          default: 900
      title: Autoshutdown
      x-fern-type-name: Autoshutdown
    Rolling:
      description: >-
        +docs=This strategy updates the pods in a rolling fashion such that a
        subset of the

        total pods are replaced with new version at one time.

        A commonly used strategy can be to have maxUnavailablePercentage close
        to 0 so that there

        is no downtime and keep the maxSurgePercentage to around 25%. If you are
        anyways running

        a large number of pods, the service can often tolerate a few pods going
        down - so you

        max maxUnavailablePercentage = 10 and maxSurgePercentage=0. You can read
        about it more

        [here](https://spot.io/resources/kubernetes-autoscaling/5-kubernetes-deployment-strategies-roll-out-like-the-pros/)

        +label=Rolling update strategy
      type: object
      required:
        - type
        - max_unavailable_percentage
        - max_surge_percentage
      properties:
        type:
          description: +value=rolling_update
          type: string
          enum:
            - rolling_update
        max_unavailable_percentage:
          description: >-
            +label=Max unavailable(%)

            +usage=Percentage of total replicas that can be brought down at one
            time.

            For a value of 25 when replicas are set to 12 this would mean
            minimum (25% of 12) = 3 pods might be unavailable during the
            deployment.

            Setting this to a higher value can help in speeding up the
            deployment process.
          type: integer
          minimum: 0
          maximum: 100
          default: 25
        max_surge_percentage:
          description: >-
            +label=Max Surge(%)

            +usage=Percentage of total replicas of updated image that can be
            brought up over the total replicas count.

            For a value of 25 when replicas are set to 12 this would mean
            (12+(25% of 12) = 15) pods might be running at one time.

            Setting this to a higher value can help in speeding up the
            deployment process.
          type: integer
          minimum: 0
          maximum: 100
          default: 25
      title: Rolling
      x-fern-type-name: Rolling
    Canary:
      description: >-
        +docs=This strategy brings up the new release without bringing the older
        release down. Traffic is shifted from the older release to the newer
        release in a staged manner.

        This can help with verifying the health of the new release without
        shifting complete traffic.

        +label=Canary strategy
      type: object
      required:
        - type
        - steps
      properties:
        type:
          description: +value=canary
          type: string
          enum:
            - canary
        steps:
          description: >-
            +docs=These steps would be executed in order to enable shifting of
            traffic slowly from stable to canary version

            +label=Steps
          type: array
          items:
            $ref: '#/components/schemas/CanaryStep'
      title: Canary
      x-fern-type-name: Canary
    BlueGreen:
      description: >-
        +docs=This strategy brings up the new release completely before
        switching the complete load to the new release.

        This minimizes the time that two versions are serving traffic at the
        same time.

        +label=Blue Green strategy
      type: object
      required:
        - type
      properties:
        type:
          description: +value=blue_green
          type: string
          enum:
            - blue_green
        enable_auto_promotion:
          description: >-
            +docs=Promote the new release to handle the complete traffic. A
            manual promotion would be needed if this is disabled

            +label=Auto-promotion
          type: boolean
          default: false
        auto_promotion_seconds:
          description: >-
            +docs=Promote the new release to handle the complete traffic after
            waiting for these many seconds

            +label=Auto-promotion seconds
          type: integer
          minimum: 0
          default: 30
      title: BlueGreen
      x-fern-type-name: BlueGreen
    BaseService:
      type: object
      required:
        - name
        - image
        - ports
      properties:
        name:
          description: >-
            +usage=Name of the service. This uniquely identifies this service in
            the workspace.

            > Name can only contain alphanumeric characters and '-' and can be
            atmost 25 characters long

            +sort=1

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z](?:[a-z0-9]|-(?!-)){1,30}[a-z0-9]$
        image:
          description: >-
            +docs=Specify whether you want to deploy a Docker image or build and
            deploy from source code

            +label=Deploy a Docker image or build and deploy from source code

            +icon=fa-solid fa-cloud-arrow-up:#21B6A8

            +sort=2

            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/Build'
            - $ref: '#/components/schemas/Image'
          discriminator:
            propertyName: type
        artifacts_download:
          $ref: '#/components/schemas/ArtifactsDownload'
        resources:
          $ref: '#/components/schemas/Resources'
        env:
          description: >-
            +label=Environment Variables

            +usage=Configure environment variables to be injected in the service
            either as plain text or secrets.
            [Docs](https://docs.truefoundry.com/docs/env-variables)

            +icon=fa-globe

            +sort=6
          type: object
          default: null
          additionalProperties:
            type: string
          nullable: true
        ports:
          description: >-
            +docs=Specify the ports you want the service to be exposed to

            +label=Configure ports and endpoints to route customer traffic

            +usage=Expose the deployment to make it accessible over the internet
            or keep it private. Implement authentication to restrict access.
            [Docs](https://docs.truefoundry.com/docs/define-ports-and-domains)

            +icon=fa-plug

            +sort=4
          type: array
          items:
            $ref: '#/components/schemas/Port'
        service_account:
          type: string
        mounts:
          description: >-
            +usage=Configure data to be mounted to service pod(s) as a string,
            secret or volume.
            [Docs](https://docs.truefoundry.com/docs/mounting-volumes-service)

            +sort=10011

            +todo=Add discriminator for union types in array
          type: array
          items:
            type: object
            oneOf:
              - $ref: '#/components/schemas/SecretMount'
              - $ref: '#/components/schemas/StringDataMount'
              - $ref: '#/components/schemas/VolumeMount'
        labels:
          description: +label=Labels
          type: object
          additionalProperties:
            type: string
        kustomize:
          $ref: '#/components/schemas/Kustomize'
        liveness_probe:
          $ref: '#/components/schemas/HealthProbe'
        readiness_probe:
          $ref: '#/components/schemas/HealthProbe'
        startup_probe:
          $ref: '#/components/schemas/HealthProbe'
        workspace_fqn:
          description: |-
            +label=Workspace FQN
            +docs=Fully qualified name of the workspace
            +uiType=Hidden
          type: string
      title: BaseService
      x-fern-type-name: BaseService
    AsyncServiceAutoscaling:
      type: object
      properties:
        metrics:
          description: |-
            +label=Autoscaling metrics
            +usage=Metrics to use for the autoscaler
            +sort=4
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/SQSQueueMetricConfig'
            - $ref: '#/components/schemas/NATSMetricConfig'
            - $ref: '#/components/schemas/KafkaMetricConfig'
            - $ref: '#/components/schemas/CronMetric'
            - $ref: '#/components/schemas/AMQPMetricConfig'
          discriminator:
            propertyName: type
      allOf:
        - $ref: '#/components/schemas/BaseAutoscaling'
      required:
        - metrics
      title: AsyncServiceAutoscaling
      x-fern-type-name: AsyncServiceAutoscaling
    WorkerConfig:
      type: object
      required:
        - input_config
        - num_concurrent_workers
      properties:
        input_config:
          description: |-
            +label=Input Config
            +usage=Input Config
            +sort=1
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/SQSInputConfig'
            - $ref: '#/components/schemas/NATSInputConfig'
            - $ref: '#/components/schemas/KafkaInputConfig'
            - $ref: '#/components/schemas/AMQPInputConfig'
          discriminator:
            propertyName: type
        output_config:
          description: |-
            +label=Output Config
            +usage=Output Config
            +sort=2
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/SQSOutputConfig'
            - $ref: '#/components/schemas/NATSOutputConfig'
            - $ref: '#/components/schemas/CoreNATSOutputConfig'
            - $ref: '#/components/schemas/KafkaOutputConfig'
            - $ref: '#/components/schemas/AMQPOutputConfig'
          discriminator:
            propertyName: type
        num_concurrent_workers:
          description: |-
            +label=Number of Concurrent Workers
            +usage=Number of concurrent workers to spawn for the processor
            +sort=3
          type: integer
          minimum: 1
          maximum: 10
          default: 1
      title: WorkerConfig
      x-fern-type-name: WorkerConfig
    AsyncProcessorSidecar:
      type: object
      required:
        - destination_url
      properties:
        destination_url:
          description: |-
            +label=Destination URL
            +usage=URL for the processor to invoke
          type: string
        request_timeout:
          description: |-
            +label=Request Timeout Seconds
            +usage=Timeout for the invoke request in seconds
          type: integer
          minimum: 1
          default: 10
        sidecar_image:
          description: >-
            +label=Sidecar Image

            +usage=Image for the processor sidecar (This field will be
            deprecated in the future)
          type: string
      title: AsyncProcessorSidecar
      x-fern-type-name: AsyncProcessorSidecar
    Build:
      description: |-
        +docs=Describes how we build our code into a Docker image.
        +label=Source Code (Build and deploy source code)
        +icon=fa-code
      type: object
      required:
        - type
        - build_source
        - build_spec
      properties:
        type:
          description: +value=build
          type: string
          enum:
            - build
        docker_registry:
          description: >-
            +docs=FQN of the container registry. You can the FQN of your desired
            container registry (or add one)

            in the  Integrations
            page[Integrations](https://app.truefoundry.tech/integrations?tab=docker-registry)
            page

            +label=Docker Registry

            +usage=FQN of the container registry. If you can't find your
            registry here,

            add it through the [Integrations](/integrations?tab=docker-registry)
            page
          type: string
          pattern: ^\S+$
        build_source:
          description: |-
            +docs=Source code location.
            +label=Fetch source code to build and deploy
            +icon=fa-code
            +discriminator=type
            +sort=1
          type: object
          oneOf:
            - $ref: '#/components/schemas/RemoteSource'
            - $ref: '#/components/schemas/GitSource'
            - $ref: '#/components/schemas/LocalSource'
          discriminator:
            propertyName: type
        build_spec:
          description: >-
            +docs=Instructions to build a container image out of the build
            source

            +label=Build using DockerFile or using Buildpack

            +icon=fa-wrench

            +discriminator=type

            +sort=2
          type: object
          oneOf:
            - $ref: '#/components/schemas/DockerFileBuild'
            - $ref: '#/components/schemas/PythonBuild'
          discriminator:
            propertyName: type
      title: Build
      x-fern-type-name: Build
    Image:
      description: >-
        +docs=Describes that we are using a pre-built image stored in a Docker
        Image registry

        +label=Docker Image (Deploy an existing image)

        +icon=fa-brands fa-docker:#0db7ed
      type: object
      required:
        - type
        - image_uri
      properties:
        type:
          description: +value=image
          type: string
          enum:
            - image
        image_uri:
          description: >-
            +label=Image URI

            +usage=The image URI. Specify the name of the image and the tag.

            If the image is in Dockerhub, you can skip registry-url (for e.g.
            `tensorflow/tensorflow`).

            You can use an image from a private registry using Advanced fields

            +placeholder=registry-url/account/image:version (e.g.
            docker.io/tensorflow/tensorflow)
          type: string
          pattern: ^\S+$
        docker_registry:
          description: >-
            +docs=FQN of the container registry. You can the FQN of your desired
            container registry (or add one)

            in the  Integrations
            page[Integrations](https://app.truefoundry.tech/integrations?tab=docker-registry)
            page

            +label=Docker Registry

            +usage=FQN of the container registry. If you can't find your
            registry here,

            add it through the [Integrations](/integrations?tab=docker-registry)
            page
          type: string
          pattern: ^\S+$
        command:
          description: >-
            +label=Command Override

            +usage=Override the command to run when container starts.

            When deploying a Job, the command can be templatized by defining
            `params` and referencing them in command

            E.g. `python main.py --learning_rate {{learning_rate}}`
          oneOf:
            - type: string
            - type: array
              items:
                type: string
      title: Image
      x-fern-type-name: Image
    Manual:
      description: >-
        +docs=Describes that we are going to manually trigger our job.

        +label=Manual

        +usage=Trigger the job manually.
        [Docs](https://docs.truefoundry.com/docs/deploy-a-cron-job)
      type: object
      required:
        - type
      properties:
        type:
          description: +value=manual
          type: string
          enum:
            - manual
      title: Manual
      x-fern-type-name: Manual
    Schedule:
      description: >-
        +docs=Describes that we are going to schedule our job to run at a
        schedule, making our job a cron job.

        +label=Schedule

        +usage=Run the job on a schedule.
        [Docs](https://docs.truefoundry.com/docs/deploy-a-cron-job)
      type: object
      required:
        - type
        - schedule
        - concurrency_policy
      properties:
        type:
          description: +value=scheduled
          type: string
          enum:
            - scheduled
        schedule:
          description: >-
            +docs=Specify the schedule for this job to be run periodically in
            cron format. [Learn
            more](https://docs.truefoundry.com/docs/deploy-a-cron-job)

            +usage=Specify the schedule for this job to be run periodically in
            cron format.

            ```

            * * * * *

            | | | | |

            | | | | |___ day of week (0-6) (Sunday is 0)

            | | | |_____ month (1-12)

            | | |_______ day of month (1-31)

            | |_________ hour (0-23)

            |___________ minute (0-59)

            ```
          type: string
        concurrency_policy:
          description: >-
            +usage=Choose whether to allow this job to run while another
            instance of the job is running, or to replace the currently running
            instance. Allow

            will enable multiple instances of this job to run. Forbid will keep
            the current instance of the job running and stop a new instance from
            being run.

            Replace will terminate any currently running instance of the job and
            start a new one.
          type: string
          enum:
            - Forbid
            - Allow
            - Replace
          default: Forbid
        timezone:
          description: >-
            +usage=Timezone against which the cron schedule will be calculated,
            e.g. "Asia/Tokyo". Default is machine's local time.

            https://docs.truefoundry.com/docs/list-of-supported-timezones
          type: string
      title: Schedule
      x-fern-type-name: Schedule
    Param:
      type: object
      required:
        - name
      properties:
        name:
          description: >-
            +usage=Name of the param

            +message=name can contain lower case alphabets, digits, underscore
            (_) and hypen (-). It can be 32 characters long, should start with
            an alphabet, and should end with either an alphabet or digit.
          type: string
          pattern: ^[a-z][a-z0-9\-_]{0,30}[a-z0-9]$
        description:
          description: |-
            +usage=Description of param
            +message=description cannot be longer than 127 characters
          type: string
          pattern: ^.{1,127}$
        default:
          description: |-
            +usage=Default value or placeholder
            +message=default value cannot be longer than 127 characters
            +label=Default value
          type: string
          pattern: ^.{0,127}$
        param_type:
          type: string
          enum:
            - string
            - ml_repo
          default: string
      title: Param
      x-fern-type-name: Param
    Resources:
      description: >-
        +docs=Describes the resource constraints for the application so that it
        can be deployed accordingly on the cluster

        To learn more you can go
        [here](https://www.truefoundry.com/docs/resources-cpu-memory-storage)

        +icon=fa-microchip

        +label=Resources

        +usage=Configure resource allocations, specify node constraints and
        capacity types to improve performance and reduce expenses.
        [Docs](https://www.truefoundry.com/docs/resources-cpu-memory-storage)
      type: object
      required:
        - cpu_request
        - cpu_limit
        - memory_request
        - memory_limit
        - ephemeral_storage_request
        - ephemeral_storage_limit
      properties:
        cpu_request:
          description: >-
            +label=CPU Request

            +sort=1

            +usage=Requested CPU which determines the minimum cost incurred. The
            CPU usage can exceed the requested

            amount, but not the value specified in the limit. 1 CPU means 1 CPU
            core. Fractional CPU can be requested

            like `0.5` or `0.05`
          type: number
          minimum: 0.001
          maximum: 256
          default: 0.2
        cpu_limit:
          description: >-
            +label=CPU Limit

            +usage=CPU limit beyond which the usage cannot be exceeded. 1 CPU
            means 1 CPU core. Fractional CPU can be requested

            like `0.5`. CPU limit should be >= cpu request.

            +sort=2
          type: number
          minimum: 0.001
          maximum: 256
          default: 0.5
        memory_request:
          description: >-
            +label=Memory Request

            +usage=Requested memory which determines the minimum cost incurred.
            The unit of memory is in megabytes(MB).

            So 1 means 1 MB and 2000 means 2GB.

            +sort=3
          type: integer
          minimum: 1
          maximum: 4000000
          default: 200
        memory_limit:
          description: >-
            +label=Memory Limit

            +usage=Memory limit after which the application will be killed with
            an OOM error. The unit of memory is

            in megabytes(MB). So 1 means 1 MB and 2000 means 2GB. MemoryLimit
            should be greater than memory request.

            +sort=4
          type: integer
          minimum: 1
          maximum: 4000000
          default: 500
        ephemeral_storage_request:
          description: >-
            +label=Storage Request

            +usage=Requested disk storage. The unit of memory is in
            megabytes(MB).

            This is ephemeral storage and will be wiped out on pod restarts or
            eviction

            +sort=5
          type: integer
          minimum: 1
          maximum: 2000000
          default: 1000
        ephemeral_storage_limit:
          description: >-
            +label=Storage Limit

            +usage=Disk storage limit. The unit of memory is in megabytes(MB).
            Exceeding this limit will result in eviction.

            It should be greater than the request. This is ephemeral storage and
            will be wiped out on pod restarts or eviction

            +sort=6
          type: integer
          minimum: 1
          maximum: 2000000
          default: 2000
        shared_memory_size:
          description: >-
            +label=Shared Memory Size (MB)

            +usage=Define the shared memory requirements for your workload.
            Machine learning libraries like Pytorch can use Shared Memory

            for inter-process communication. If you use this, we will mount a
            `tmpfs` backed volume at the `/dev/shm` directory.

            Any usage will also count against the workload's memory limit
            (`resources.memory_limit`) along with your workload's memory usage.

            If the overall usage goes above `resources.memory_limit` the user
            process may get killed.

            Shared Memory Size cannot be more than the defined Memory Limit for
            the workload.
          type: integer
          minimum: 64
          maximum: 4000000
        node:
          description: >-
            +label=Node

            +usage=This field determines how the underlying node resource is to
            be utilized

            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/NodeSelector'
            - $ref: '#/components/schemas/NodepoolSelector'
          discriminator:
            propertyName: type
        devices:
          description: >-
            +label=Devices

            +usage=Define custom device or accelerator requirements for your
            workload. We currently support NVIDIA GPUs, AWS Inferentia
            Accelerators, Single Host TPU Slices.

            +todo=Add discriminator for union types in array
          type: array
          items:
            type: object
            oneOf:
              - $ref: '#/components/schemas/NvidiaGPU'
              - $ref: '#/components/schemas/AWSInferentia'
              - $ref: '#/components/schemas/NvidiaMIGGPU'
              - $ref: '#/components/schemas/NvidiaTimeslicingGPU'
              - $ref: '#/components/schemas/GcpTPU'
      title: Resources
      x-fern-type-name: Resources
    JobAlert:
      description: |-
        +docs=Describes the configuration for the job alerts
        +label=Alert
      type: object
      properties:
        notification_channel:
          description: |-
            +label=Notification Channel
            +usage=Specify the notification channel to send alerts to
            +uiType=Hidden
            +sort=660
          type: string
          minLength: 1
        to_emails:
          description: >-
            +label=To Emails

            +usage=List of recipients' email addresses if the notification
            channel is Email.

            +docs=Specify the emails to send alerts to

            +uiType=Hidden

            +sort=665
          type: array
          items:
            type: string
            minLength: 1
        notification_target:
          $ref: '#/components/schemas/NotificationTarget'
        on_start:
          description: |-
            +label=On Start
            +usage=Send an alert when the job starts
            +sort=670
          type: boolean
          default: false
        on_completion:
          type: boolean
          default: false
        on_failure:
          description: |-
            +label=On Failure
            +usage=Send an alert when the job fails
            +sort=690
          type: boolean
          default: true
      title: JobAlert
      x-fern-type-name: JobAlert
    SecretMount:
      type: object
      required:
        - type
        - mount_path
        - secret_fqn
      properties:
        type:
          description: +value=secret
          type: string
          enum:
            - secret
        mount_path:
          description: |-
            +label=File path
            +usage=Absolute file path where the file will be created.
            +message=Please enter a valid file path
          type: string
          pattern: ^\/(?:[^/\n]+\/*)*[^/\n]+(\.[^/\n]+)?$
        secret_fqn:
          description: |-
            +label=Secret
            +usage=The TrueFoundry secret whose value will be the file content.
          type: string
          pattern: ^tfy-secret:\/\/.+:.+:.+$
      title: SecretMount
      x-fern-type-name: SecretMount
    StringDataMount:
      type: object
      required:
        - type
        - mount_path
        - data
      properties:
        type:
          description: +value=string
          type: string
          enum:
            - string
        mount_path:
          description: |-
            +label=File Path
            +usage=Absolute file path where the file will be created.
            +message=Please enter a valid file path
          type: string
          pattern: ^\/(?:[^/\n]+\/*)*[^/\n]+(\.[^/\n]+)?$
        data:
          description: |-
            +label=Data
            +usage=The file content.
          type: string
      title: StringDataMount
      x-fern-type-name: StringDataMount
    VolumeMount:
      type: object
      required:
        - type
        - mount_path
        - volume_fqn
      properties:
        type:
          description: +value=volume
          type: string
          enum:
            - volume
        mount_path:
          description: |-
            +label=Volume mount path
            +usage=Absolute file path where the volume will be mounted.
            +message=Please enter a valid mount path
          type: string
          pattern: ^\/(?:[^/\n]+\/*)*[^/\n]+(\.[^/\n]+)?$
        sub_path:
          description: >-
            +label=Sub Path

            +usage=Sub path within the volume to mount. Defaults to root of the
            volume.
          type: string
          pattern: ^(?:[^/\n]+/*)*[^/\n]+(\.[^/\n]+)?$
        volume_fqn:
          description: |-
            +label=Volume
            +usage=The TrueFoundry volume that needs to be mounted.
          type: string
          pattern: ^tfy-volume:\/\/.+:.+:.+$
      title: VolumeMount
      x-fern-type-name: VolumeMount
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
    WorkbenchImage:
      description: +usage=Workbench Image with persistent environment (Python 3.11.6)
      type: object
      required:
        - image_uri
      properties:
        image_uri:
          description: >-
            +label=Image URI

            +usage=The image URI. Specify the name of the image and the tag.

            If the image is in Dockerhub, you can skip registry-url (for e.g.
            `tensorflow/tensorflow`).

            You can use an image from a private registry using Advanced fields

            +placeholder=registry-url/account/image:version
          type: string
        build_script:
          description: >-
            +label=Build Script

            +usage=The build script to run when building the image.

            This will be executed as the last step in the docker build process
            as the root user (RUN DEBIAN_FRONTEND=noninteractive bash -ex
            build_script.sh)

            +placeholder=Enter the build script

            +uiType=CodeEditor

            +uiProps={"language":"shell"}
          type: string
          minLength: 1
          maxLength: 5120
        docker_registry:
          description: >-
            +docs=FQN of the container registry. You can the FQN of your desired
            container registry (or add one)

            in the  Integrations
            page[Integrations](https://app.truefoundry.tech/integrations?tab=docker-registry)
            page

            +label=Docker Registry

            +uiType=IntegrationSelect

            +uiProps={"integrationType":"docker-registry"}

            +usage=FQN of the container registry. If you can't find your
            registry here,

            add it through the [Integrations](/integrations?tab=docker-registry)
            page
          type: string
      title: WorkbenchImage
      x-fern-type-name: WorkbenchImage
    BaseWorkbenchInput:
      description: +docs=Describes the configuration for the service
      type: object
      required:
        - name
        - home_directory_size
      properties:
        name:
          description: >-
            +usage=Name of the workbench. This uniquely identifies this
            workbench in the workspace.

            > Name can only contain alphanumeric characters and '-' and can be
            atmost 25 characters long

            +sort=1

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z](?:[a-z0-9]|-(?!-)){1,30}[a-z0-9]$
        home_directory_size:
          description: >-
            +label=Home Directory Size in GB (Persistent)

            +usage=Size of the home directory for the workbench (Persistent
            Storage)

            +sort=6
          type: integer
          minimum: 5
          maximum: 64000
          default: 20
        resources:
          $ref: '#/components/schemas/Resources'
        env:
          description: >-
            +label=Environment Variables

            +usage=Configure environment variables to be injected in the service
            either as plain text or secrets.
            [Docs](https://docs.truefoundry.com/docs/environment-variables-and-secrets-jobs)

            +sort=10110
          type: object
          default: null
          additionalProperties:
            type: string
          nullable: true
        mounts:
          description: >-
            +usage=Configure data to be mounted to workbench pod(s) as a string,
            secret or volume.
            [Docs](https://docs.truefoundry.com/docs/mounting-volumes-job)

            +sort=10111

            +todo=Add discriminator for union types in array
          type: array
          items:
            type: object
            oneOf:
              - $ref: '#/components/schemas/SecretMount'
              - $ref: '#/components/schemas/StringDataMount'
              - $ref: '#/components/schemas/VolumeMount'
        service_account:
          description: +sort=10113
          type: string
        kustomize:
          $ref: '#/components/schemas/Kustomize'
        workspace_fqn:
          description: |-
            +label=Workspace FQN
            +docs=Fully qualified name of the workspace
            +uiType=Hidden
          type: string
      title: BaseWorkbenchInput
      x-fern-type-name: BaseWorkbenchInput
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
    DynamicVolumeConfig:
      description: +label=Create new volume
      type: object
      required:
        - type
        - storage_class
        - size
      properties:
        type:
          description: |-
            +label=Volume Type
            +value=dynamic
            +usage=Volume Type for the volume.
          type: string
          enum:
            - dynamic
        storage_class:
          description: |-
            +label=Storage Class Name
            +usage=Name of the storage class to be used for the volume.
          type: string
        size:
          description: |-
            +label=Size(Gi)
            +unit=Gi
            +usage=Size of volume in Gi
          type: integer
          minimum: 1
          maximum: 64000
      title: DynamicVolumeConfig
      x-fern-type-name: DynamicVolumeConfig
    StaticVolumeConfig:
      description: +label=Use existing volume
      type: object
      required:
        - type
        - persistent_volume_name
      properties:
        type:
          description: |-
            +label=Volume Type
            +value=static
            +usage=Volume Type for the volume.
          type: string
          enum:
            - static
        persistent_volume_name:
          description: |-
            +label=Persistent Volume
            +usage=Persistent Volume Name of the volume to be used.
          type: string
      title: StaticVolumeConfig
      x-fern-type-name: StaticVolumeConfig
    VolumeBrowser:
      description: +label=Volume Browser
      type: object
      required:
        - endpoint
      properties:
        username:
          description: >-
            +message=3 to 10 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number.

            +usage=Username for logging in the volume browser.

            +sort=1
          type: string
          pattern: ^[a-z][a-z0-9]{1,8}[a-z0-9]$
        password_secret_fqn:
          description: >-
            +label=Password Secret FQN

            +usage=TFY Secret containing the password for logging in the volume
            browser.

            +sort=2
          type: string
          pattern: ^tfy-secret:\/\/.+:.+:.+$
        endpoint:
          $ref: '#/components/schemas/Endpoint'
        service_account:
          description: |-
            +label=Service Account Name
            +usage=Kubernetes Service account name for the volume browser.
            +sort=4
          type: string
      title: VolumeBrowser
      x-fern-type-name: VolumeBrowser
    LocalSource:
      description: >-
        +docs=Describes that we are using code stored in a local developement
        environment to build our image

        +label=Local

        +icon=fa-folder:black

        +sort=100
      type: object
      required:
        - type
        - project_root_path
        - local_build
      properties:
        type:
          description: +value=local
          type: string
          enum:
            - local
        project_root_path:
          description: +usage=Local project root path.
          type: string
          default: ./
        local_build:
          description: run docker build locally
          type: boolean
          default: true
      title: LocalSource
      x-fern-type-name: LocalSource
    RemoteSource:
      description: >-
        +docs=Describes that we are using code stored in a remote repository to
        build our image

        +label=S3

        +icon=fa-brands fa-aws:black

        +sort=200
      type: object
      required:
        - type
        - remote_uri
      properties:
        type:
          description: +value=remote
          type: string
          enum:
            - remote
        remote_uri:
          description: |-
            +docs=Remote repository URI
            +label=Remote URI
          type: string
      title: RemoteSource
      x-fern-type-name: RemoteSource
    FlyteTask:
      type: object
      required:
        - template
      properties:
        template:
          $ref: '#/components/schemas/FlyteTaskTemplate'
        description: {}
      title: FlyteTask
      x-fern-type-name: FlyteTask
    FlyteWorkflow:
      type: object
      required:
        - template
      properties:
        template:
          $ref: '#/components/schemas/FlyteWorkflowTemplate'
        description: {}
      title: FlyteWorkflow
      x-fern-type-name: FlyteWorkflow
    FlyteLaunchPlan:
      type: object
      required:
        - id
        - spec
        - closure
      properties:
        id:
          $ref: '#/components/schemas/FlyteLaunchPlanID'
        spec:
          $ref: '#/components/schemas/FlyteLaunchPlanSpec'
        closure: {}
      title: FlyteLaunchPlan
      x-fern-type-name: FlyteLaunchPlan
    WorkflowAlert:
      description: |-
        +docs=Describes the configuration for the workflow alerts
        +label=Alert
      type: object
      properties:
        notification_target:
          $ref: '#/components/schemas/NotificationTarget'
        on_completion:
          description: |-
            +label=On Completion
            +usage=Send an alert when the job completes
            +sort=680
          type: boolean
          default: false
        on_failure:
          description: |-
            +label=On Failure
            +usage=Send an alert when the job fails
            +sort=690
          type: boolean
          default: true
      title: WorkflowAlert
      x-fern-type-name: WorkflowAlert
    SparkImage:
      description: >-
        +docs=Describes that we are using a pre-built image stored in a Docker
        Image registry

        +label=Deploy an existing image

        +icon=fa-brands fa-docker:#0db7ed
      type: object
      required:
        - type
        - spark_version
        - image_uri
      properties:
        type:
          description: +value=spark-image
          type: string
          enum:
            - spark-image
        spark_version:
          description: >-
            --- Spark Specific Field ---

            +label=Spark Version

            +usage=Spark version should match the spark version installed in the
            image.

            +sort=1000
          type: string
          default: 3.5.2
        image_uri:
          description: >-
            +label=Image URI

            +usage=The image URI. Specify the name of the image and the tag.

            If the image is in Dockerhub, you can skip registry-url (for e.g.
            `tensorflow/tensorflow`).

            You can use an image from a private registry using Advanced fields

            +placeholder=registry-url/account/image:version (e.g.
            docker.io/tensorflow/tensorflow)

            +sort=1001
          type: string
          pattern: ^\S+$
        docker_registry:
          description: >-
            +docs=FQN of the container registry. You can the FQN of your desired
            container registry (or add one)

            in the  Integrations
            page[Integrations](https://app.truefoundry.tech/integrations?tab=docker-registry)
            page

            +label=Docker Registry

            +usage=FQN of the container registry. If you can't find your
            registry here,

            add it through the [Integrations](/integrations?tab=docker-registry)
            page

            +sort=1002
          type: string
          pattern: ^\S+$
      title: SparkImage
      x-fern-type-name: SparkImage
    SparkImageBuild:
      description: |-
        TODO (gw): Fix sorting for this such that it looks similar to SparkImage
        +docs=Describes that we are building a new image based on the spec
        +label=Build a new image
        +icon=fa-brands fa-docker:#0db7ed
      type: object
      required:
        - type
        - build_source
        - build_spec
      properties:
        type:
          description: +value=spark-image-build
          type: string
          enum:
            - spark-image-build
        docker_registry:
          description: >-
            +docs=FQN of the container registry. You can the FQN of your desired
            container registry (or add one)

            in the  Integrations
            page[Integrations](https://app.truefoundry.tech/integrations?tab=docker-registry)
            page

            +label=Docker Registry

            +usage=FQN of the container registry. If you can't find your
            registry here,

            add it through the [Integrations](/integrations?tab=docker-registry)
            page

            +sort=1002
          type: string
          pattern: ^\S+$
        build_source:
          description: >-
            TODO (gw): The following is a hack till the uiType GitSelect is
            fixed fron frontend

            +label=Fetch source code

            +sort=1003

            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/GitSource'
            - $ref: '#/components/schemas/RemoteSource'
          discriminator:
            propertyName: type
        build_spec:
          $ref: '#/components/schemas/SparkBuild'
      title: SparkImageBuild
      x-fern-type-name: SparkImageBuild
    SparkJobPythonEntrypoint:
      description: +label=python
      type: object
      required:
        - type
        - main_application_file
      properties:
        type:
          description: +value=python
          type: string
          enum:
            - python
        main_application_file:
          description: >-
            +label=Main Application File

            +usage=The main application file to be executed by the spark job.
            Relative path in case of git repository.

            +message=Filename should have .py extension

            +sort=5

            +placeholder=For example: local:///path/to/file.py,
            s3:///bucket/path/to/file.py, etc.
          type: string
        arguments:
          description: |-
            +label=Arguments
            +usage=Arguments to be passed to the main application file.
            +sort=6
          type: string
      title: SparkJobPythonEntrypoint
      x-fern-type-name: SparkJobPythonEntrypoint
    SparkJobScalaEntrypoint:
      description: +label=scala
      type: object
      required:
        - type
        - main_application_file
        - main_class
      properties:
        type:
          description: +value=scala
          type: string
          enum:
            - scala
        main_application_file:
          description: >-
            +label=Main Application File

            +usage=The main application file to be executed by the spark job.

            +message=Filename should have .jar extension

            +sort=5

            +placeholder=For example: local:///path/to/file.jar,
            s3:///bucket/path/to/file.jar, etc.
          type: string
        main_class:
          description: |-
            +label=Main Class
            +usage=The main class to be executed by the spark job.
            +sort=6
            +required=true
            +message=The main class must be a valid Java class name.
          type: string
        arguments:
          description: |-
            +label=Arguments
            +usage=Arguments to be passed to the main application file.
            +sort=7
          type: string
      title: SparkJobScalaEntrypoint
      x-fern-type-name: SparkJobScalaEntrypoint
    SparkJobJavaEntrypoint:
      description: +label=java
      type: object
      required:
        - type
        - main_application_file
        - main_class
      properties:
        type:
          description: +value=java
          type: string
          enum:
            - java
        main_application_file:
          description: >-
            +label=Main Application File

            +usage=The main application file to be executed by the spark job.

            +message=Filename should have .jar extension

            +sort=5

            +placeholder=For example: local:///path/to/file.jar,
            s3:///bucket/path/to/file.jar, etc.
          type: string
        main_class:
          description: |-
            +label=Main Class
            +usage=The main class to be executed by the spark job.
            +sort=6
            +required=true
            +message=The main class must be a valid Java class name.
          type: string
        arguments:
          description: |-
            +label=Arguments
            +usage=Arguments to be passed to the main application file.
            +sort=7
          type: string
      title: SparkJobJavaEntrypoint
      x-fern-type-name: SparkJobJavaEntrypoint
    SparkJobPythonNotebookEntrypoint:
      description: +label=python notebook
      type: object
      required:
        - type
        - main_application_file
      properties:
        type:
          description: +value=python-notebook
          type: string
          enum:
            - python-notebook
        main_application_file:
          description: >-
            +label=Main Application File

            +usage=The main application file to be executed by the spark job.
            Relative path in case of git repository.

            +message=Filename should have .ipynb extension

            +sort=5

            +placeholder=For example: local:///path/to/file.ipynb,
            s3:///bucket/path/to/file.ipynb, etc.
          type: string
      title: SparkJobPythonNotebookEntrypoint
      x-fern-type-name: SparkJobPythonNotebookEntrypoint
    SparkJobScalaNotebookEntrypoint:
      description: +label=scala notebook
      type: object
      required:
        - type
        - main_application_file
      properties:
        type:
          description: +value=scala-notebook
          type: string
          enum:
            - scala-notebook
        main_application_file:
          description: >-
            +label=Main Application File

            +usage=The main application file to be executed by the spark job.
            Relative path in case of git repository.

            +message=Filename should have .ipynb extension

            +sort=5

            +placeholder=For example: local:///path/to/file.ipynb,
            s3:///bucket/path/to/file.ipynb, etc.
          type: string
      title: SparkJobScalaNotebookEntrypoint
      x-fern-type-name: SparkJobScalaNotebookEntrypoint
    SparkDriverConfig:
      description: +label=Driver Config
      type: object
      properties:
        resources:
          $ref: '#/components/schemas/Resources'
      title: SparkDriverConfig
      x-fern-type-name: SparkDriverConfig
    SparkExecutorConfig:
      description: +label=Executor Config
      type: object
      required:
        - instances
      properties:
        instances:
          description: |-
            +label=Executor Instances
            +discriminator=type
          type: object
          default:
            type: fixed
            count: 1
          oneOf:
            - $ref: '#/components/schemas/SparkExecutorFixedInstances'
            - $ref: '#/components/schemas/SparkExecutorDynamicScaling'
          discriminator:
            propertyName: type
        resources:
          $ref: '#/components/schemas/Resources'
      title: SparkExecutorConfig
      x-fern-type-name: SparkExecutorConfig
    HeaderMatch:
      type: object
      required:
        - type
        - name
        - exact_match
      properties:
        type:
          description: +value=header
          type: string
          enum:
            - header
        name:
          description: |-
            +docs=Header name to match on
            +label=Name
          type: string
        exact_match:
          description: |-
            +docs=Header value to match on
            +label=Value
          type: string
          default: ''
      title: HeaderMatch
      x-fern-type-name: HeaderMatch
    ForwardAction:
      type: object
      required:
        - type
        - service_uri
        - port
      properties:
        type:
          description: +value=forward
          type: string
          enum:
            - forward
        service_uri:
          description: >-
            +docs=Service uri to redirect to. This is the kubernetes fully
            qualified domain name that the target service can be reached on

            Eg - For a service called `sample-service` in workspace `ws` this
            would be `sample-service.ws.svc.cluster.local

            +label=Service URI
          type: string
        port:
          description: |-
            +docs=Port to redirect the service traffic to
            +label=Target port
          type: integer
          minimum: 0
          maximum: 65535
      title: ForwardAction
      x-fern-type-name: ForwardAction
    MirrorAction:
      type: object
      required:
        - type
        - service_uri
        - port
        - percentage
      properties:
        type:
          description: +value=mirror
          type: string
          enum:
            - mirror
        service_uri:
          description: >-
            +docs=Service uri to mirror to. This is the kubernetes fully
            qualified domain name that the target service can be reached on

            Eg - For a service called `sample-service` in workspace `ws` this
            would be `sample-service.ws.svc.cluster.local

            +label=Service URI
          type: string
        port:
          description: |-
            +docs=Port to redirect the service traffic to
            +label=Target port
          type: integer
          minimum: 0
          maximum: 65535
        percentage:
          description: |-
            +docs=Percentage of requests to mirror
            +label=Percentage
          type: integer
          minimum: 0
          maximum: 100
      title: MirrorAction
      x-fern-type-name: MirrorAction
    ApplicationType:
      type: string
      enum:
        - async-service
        - service
        - job
        - spark-job
        - helm
        - notebook
        - codeserver
        - rstudio
        - ssh-server
        - volume
        - application
        - application-set
        - intercept
        - workflow
      title: ApplicationType
      x-fern-type-name: ApplicationType
    ApplicationMetadata:
      type: object
      properties:
        paused:
          type: boolean
      title: ApplicationMetadata
      x-fern-type-name: ApplicationMetadata
    ApplicationLifecycleStage:
      type: string
      enum:
        - active
        - deleting
        - deletion_failed
      title: ApplicationLifecycleStage
      x-fern-type-name: ApplicationLifecycleStage
    Alert:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        timestamps:
          type: array
          items:
            format: date-time
            type: string
        startTime:
          format: date-time
          type: string
        resolvedTime:
          format: date-time
          type: string
        applicationId:
          type: string
        clusterId:
          type: string
        tenantName:
          type: string
        fingerprint:
          type: string
        updatedAt:
          type: string
        createdAt:
          type: string
        applicationDebugInfoId:
          type: string
      required:
        - name
        - timestamps
        - startTime
        - clusterId
        - fingerprint
      title: Alert
      x-fern-type-name: Alert
    ApplicationDebugInfo:
      type: object
      properties:
        id:
          type: string
        applicationId:
          type: string
        application:
          $ref: '#/components/schemas/Application'
        debugInfo:
          type: object
          additionalProperties: true
        createdAt:
          format: date-time
          type: string
        updatedAt:
          format: date-time
          type: string
      required:
        - applicationId
        - debugInfo
      title: ApplicationDebugInfo
      x-fern-type-name: ApplicationDebugInfo
    ApplicationProblem:
      type: object
      properties:
        name:
          type: string
        description:
          type: string
      required:
        - name
        - description
      title: ApplicationProblem
      x-fern-type-name: ApplicationProblem
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
    DeploymentStatusValue:
      type: string
      enum:
        - INITIALIZED
        - BUILD_SUCCESS
        - BUILD_FAILED
        - DEPLOY_FAILED
        - DEPLOY_FAILED_WITH_RETRY
        - DEPLOY_SUCCESS
        - ROLLOUT_STARTED
        - SET_TRAFFIC
        - PAUSED
        - FAILED
        - CANCELLED
        - REDEPLOY_STARTED
        - BUILDING
      title: DeploymentStatusValue
      x-fern-type-name: DeploymentStatusValue
    DeploymentTransition:
      type: string
      enum:
        - BUILDING
        - DEPLOYING
        - REUSING_EXISTING_BUILD
        - COMPONENTS_DEPLOYING
        - WAITING
        - ''
      title: DeploymentTransition
      x-fern-type-name: DeploymentTransition
    CPUUtilizationMetric:
      type: object
      required:
        - type
        - value
      properties:
        type:
          description: +value=cpu_utilization
          type: string
          enum:
            - cpu_utilization
        value:
          description: >-
            +label=CPU utilization %

            +usage=Percentage of cpu request averaged over all replicas which
            the autoscaler should try to maintain
          type: integer
          minimum: 1
          maximum: 100
      title: CPUUtilizationMetric
      x-fern-type-name: CPUUtilizationMetric
    RPSMetric:
      type: object
      required:
        - type
        - value
      properties:
        type:
          description: +value=rps
          type: string
          enum:
            - rps
        value:
          description: >-
            +label=Requests per second

            +usage=Average request per second averaged over all replicas that
            autoscaler should try to maintain
          type: number
          minimum: 0
          exclusiveMinimum: true
      title: RPSMetric
      x-fern-type-name: RPSMetric
    CronMetric:
      type: object
      required:
        - type
        - start
        - end
        - timezone
      properties:
        type:
          description: +value=cron
          type: string
          enum:
            - cron
        desired_replicas:
          description: >-
            +label=Desired Replicas

            +usage=Desired number of replicas during the given interval. Default
            value is max_replicas.
          type: integer
          minimum: 1
        start:
          description: |-
            +label=Start Schedule
            +docs=Cron expression indicating the start of the cron schedule.
            +usage=Cron expression indicating the start of the cron schedule.
            ```
            * * * * *
            | | | | |
            | | | | |___ day of week (0-6) (Sunday is 0)
            | | | |_____ month (1-12)
            | | |_______ day of month (1-31)
            | |_________ hour (0-23)
            |___________ minute (0-59)
            ```
          type: string
        end:
          description: |-
            +label=End Schedule
            +docs=Cron expression indicating the end of the cron schedule.
            +usage=Cron expression indicating the end of the cron schedule.
            ```
            * * * * *
            | | | | |
            | | | | |___ day of week (0-6) (Sunday is 0)
            | | | |_____ month (1-12)
            | | |_______ day of month (1-31)
            | |_________ hour (0-23)
            |___________ minute (0-59)
            ```
          type: string
        timezone:
          description: >-
            +usage=Timezone against which the cron schedule will be calculated,
            e.g. "Asia/Tokyo". Default is machine's local time.

            https://docs.truefoundry.com/docs/list-of-supported-timezones
          type: string
          default: UTC
      title: CronMetric
      x-fern-type-name: CronMetric
    BaseAutoscaling:
      type: object
      required:
        - min_replicas
        - max_replicas
      properties:
        min_replicas:
          description: |-
            +label=Minimum replicas
            +usage=Minimum number of replicas to keep available
            +sort=1
          type: integer
          minimum: 0
          default: 1
        max_replicas:
          description: |-
            +label=Maximum replicas
            +usage=Maximum number of replicas allowed for the component.
            +sort=2
          type: integer
          minimum: 1
          maximum: 500
        polling_interval:
          description: |-
            +label=Polling Interval
            +usage=This is the interval to check each trigger on.
          type: integer
          minimum: 0
          default: 30
      title: BaseAutoscaling
      x-fern-type-name: BaseAutoscaling
    CanaryStep:
      type: object
      required:
        - weight_percentage
      properties:
        weight_percentage:
          description: >-
            +docs=Percentage of total traffic to be shifted to the canary
            release.

            The rest will continue to go to the existing deployment

            +label=Canary weight percentage

            +unit=%

            +placeholder=Weight
          type: integer
          minimum: 0
          maximum: 100
        pause_duration:
          description: >-
            +docs=Duration for which to pause the release. The release process
            will wait for these seconds before proceeding to the next step.

            If this is not set, the step will pause indefinitely on this step

            +label=Pause duration

            +unit=seconds

            +placeholder=Duration
          type: integer
          minimum: 0
      title: CanaryStep
      x-fern-type-name: CanaryStep
    ArtifactsDownload:
      description: >-
        +docs=Describes the configuration for the artifacts cache

        +label=Artifacts Download

        +usage=Download and cache models in a volume to enhance loading speeds
        and reduce costs by avoiding repeated downloads.
        [Docs](https://docs.truefoundry.com/docs/download-and-cache-models)
      type: object
      required:
        - artifacts
      properties:
        cache_volume:
          $ref: '#/components/schemas/ArtifactsCacheVolume'
        artifacts:
          description: |-
            +label=Artifacts
            +usage=List of artifacts to be cached
            +todo=Add discriminator for union types in array
          type: array
          items:
            type: object
            oneOf:
              - $ref: '#/components/schemas/TrueFoundryArtifactSource'
              - $ref: '#/components/schemas/HuggingfaceArtifactSource'
      title: ArtifactsDownload
      x-fern-type-name: ArtifactsDownload
    Port:
      description: +docs=Describes the ports the service should be exposed to.
      type: object
      required:
        - port
        - protocol
        - expose
      properties:
        port:
          description: +usage=Port number to expose.
          type: integer
          minimum: 1
          maximum: 65535
          default: 80
        protocol:
          description: +usage=Protocol for the port.
          type: string
          enum:
            - TCP
            - UDP
          default: TCP
        expose:
          description: +usage=Expose the port
          type: boolean
          default: true
        app_protocol:
          description: >-
            +label=Application Protocol

            +usage=Application Protocol for the port.

            Select the application protocol used by your service. For most use
            cases, this should be `http`(HTTP/1.1).

            If you are running a gRPC server, select the `grpc` option.

            This is only applicable if `expose=true`.
          type: string
          enum:
            - http
            - grpc
            - tcp
          default: http
        host:
          description: >-
            +usage=Host e.g. ai.example.com, app.truefoundry.com

            +message=Upto 253 characters, each part of host should be at most 63
            characters long, can contain alphabets, digits and hypen, must begin
            and end with an alphanumeric characters. Parts must be separated by
            periods (.)
          type: string
          pattern: >-
            ^((([a-zA-Z0-9\-]{1,63}\.)([a-zA-Z0-9\-]{1,63}\.)*([A-Za-z]{1,63}))|(((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)))$
        path:
          description: >-
            +usage=Path e.g. /v1/api/ml/, /v2/docs/

            +message=Should begin and end with a forward slash (/). Each part
            can can contain alphabets, digits and hypen, must begin and end with
            an alphanumeric characters. Parts should be separated by forward
            slashes (/)
          type: string
          pattern: ^(/([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-_\.]*[a-zA-Z0-9]))*/$
        rewrite_path_to:
          description: >-
            +label=Rewrite Path to

            +usage=Rewrite the path prefix to a different path.

            If `path` is `/v1/api` and `rewrite_path_to` is `/api`. The URI in
            the HTTP request `http://0.0.0.0:8080/v1/api/houses` will be
            rewritten to `http://0.0.0.0:8080/api/houses` before the request is
            forwarded your service.

            Defaults to `/`.

            This is only applicable if `path` is given.

            +message=Should begin and end with a forward slash (/). Each part
            can can contain alphabets, digits and hypen, must begin and end with
            an alphanumeric characters. Parts should be separated by forward
            slashes (/)
          type: string
          pattern: ^(/([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-_\.]*[a-zA-Z0-9]))*/$
        auth:
          description: |-
            +usage=Authentication method for inbound traffic
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/BasicAuthCreds'
            - $ref: '#/components/schemas/JwtAuthConfig'
            - $ref: '#/components/schemas/TrueFoundryInteractiveLogin'
          discriminator:
            propertyName: type
      title: Port
      x-fern-type-name: Port
    HealthProbe:
      description: >-
        +docs=Describes the configuration for the Health Probe's

        To learn more you can go
        [here](https://docs.truefoundry.com/docs/liveness-readiness-probe)

        +icon=fa-heart

        +uiType=HealthProbe
      type: object
      required:
        - config
      properties:
        config:
          $ref: '#/components/schemas/HttpProbe'
        initial_delay_seconds:
          description: >-
            +usage=Time to wait after container has started before checking the
            endpoint
          type: integer
          minimum: 0
          maximum: 36000
          default: 0
        period_seconds:
          description: +usage=How often to check the endpoint
          type: integer
          minimum: 1
          maximum: 36000
          default: 10
        timeout_seconds:
          description: >-
            +usage=Time to wait for a response from the endpoint before
            considering it down
          type: integer
          minimum: 1
          maximum: 36000
          default: 1
        success_threshold:
          description: >-
            +usage=Number of successful responses from the endpoint before
            container is considered healthy
          type: integer
          minimum: 1
          maximum: 5000
          default: 1
        failure_threshold:
          description: >-
            +usage=Number of consecutive failures before the container is
            considered down
          type: integer
          minimum: 1
          maximum: 5000
          default: 3
      title: HealthProbe
      x-fern-type-name: HealthProbe
    SQSQueueMetricConfig:
      type: object
      required:
        - type
        - queue_length
      properties:
        type:
          description: +value=sqs
          type: string
          enum:
            - sqs
        queue_length:
          description: >-
            +label=Queue length

            +usage=Upper limit of the number of backlog messages the auto-scaler
            will try to maintain per replica. If you set this number to 10 and
            have 30 messages in the queue and one replica, the auto-scaler will
            scale the number of replicas to 3.
          type: integer
          minimum: 1
        auth:
          description: |-
            +label=Authentication
            +usage=Receive permissions on the SQS Queue
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/AwsSqsAccessKeyBasedAuth'
            - $ref: '#/components/schemas/AwsSqsAssumedRoleBasedAuth'
          discriminator:
            propertyName: type
      title: SQSQueueMetricConfig
      x-fern-type-name: SQSQueueMetricConfig
    NATSMetricConfig:
      type: object
      required:
        - type
        - lag_threshold
      properties:
        type:
          description: +value=nats
          type: string
          enum:
            - nats
        lag_threshold:
          description: >-
            +label=Lag Threshold

            +usage=Upper limit of the number of backlog messages the auto-scaler
            will try to maintain per replica. If you set this number to 10 and
            have 30 messages in the stream and one replica, the auto-scaler will
            scale the number of replicas to 3.
          type: integer
          minimum: 1
      title: NATSMetricConfig
      x-fern-type-name: NATSMetricConfig
    KafkaMetricConfig:
      type: object
      required:
        - type
        - lag_threshold
      properties:
        type:
          description: +value=kafka
          type: string
          enum:
            - kafka
        lag_threshold:
          description: >-
            +label=Lag Threshold

            +usage=Upper limit of the number of backlog messages the auto-scaler
            will try to maintain per replica. If you set this number to 10 and
            have 30 messages in the stream and one replica, the auto-scaler will
            scale the number of replicas to 3.
          type: integer
          minimum: 1
      title: KafkaMetricConfig
      x-fern-type-name: KafkaMetricConfig
    AMQPMetricConfig:
      type: object
      required:
        - type
        - queue_length
      properties:
        type:
          description: +value=amqp
          type: string
          enum:
            - amqp
        queue_length:
          description: >-
            +label=Queue Length

            +usage=Upper limit of the number of backlog messages the auto-scaler
            will try to maintain per replica. If you set this number to 10 and
            have 30 messages in the stream and one replica, the auto-scaler will
            scale the number of replicas to 3.
          type: integer
          minimum: 1
      title: AMQPMetricConfig
      x-fern-type-name: AMQPMetricConfig
    SQSInputConfig:
      description: |-
        +docs=Describes the configuration for the input SQS worker
        +label=SQS
      type: object
      required:
        - type
        - queue_url
        - region_name
        - visibility_timeout
        - wait_time_seconds
      properties:
        type:
          description: +value=sqs
          type: string
          enum:
            - sqs
        queue_url:
          description: |-
            +label=Queue URL
            +usage=AWS SQS Queue URL of Subscriber
            +sort=1
          type: string
        region_name:
          description: |-
            +label=Region Name
            +usage=AWS Region Name
            +sort=2
          type: string
        visibility_timeout:
          description: >-
            +label=Visibility Timeout (seconds)

            +usage=A period during which Amazon SQS prevents all consumers from
            receiving and processing the message. If one message takes 5 seconds
            to process, you can set this number to 7 or any number higher than
            5. This will ensure that while the message is being processed, it
            will not be available to other replicas. For more information, see
            [here](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-visibility-timeout.html)

            +sort=3
          type: integer
          minimum: 1
          maximum: 43200
        wait_time_seconds:
          description: >-
            +label=Wait Time Seconds

            +usage=Wait timeout for long polling. For more information, see
            [here](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-short-and-long-polling.html)
          type: integer
          minimum: 1
          maximum: 20
          default: 19
        auth:
          $ref: '#/components/schemas/AWSAccessKeyAuth'
      title: SQSInputConfig
      x-fern-type-name: SQSInputConfig
    NATSInputConfig:
      description: |-
        +docs=Describes the configuration for the input NATS worker
        +label=NATS
      type: object
      required:
        - type
        - nats_url
        - stream_name
        - root_subject
        - consumer_name
        - wait_time_seconds
      properties:
        type:
          description: +value=nats
          type: string
          enum:
            - nats
        nats_url:
          description: |-
            +label=NATS URL
            +usage=Input NATS URL
            +sort=1
          type: string
        stream_name:
          description: |-
            +label=Stream Name
            +usage=Name of the NATS stream
            +sort=2
          type: string
        root_subject:
          description: >-
            +label=Root Subject

            +usage=Root subject of input NATS

            +message=Input NATS root subject should only contain alphanumeric
            letters, dashes(-), and periods(.)

            +sort=3
          type: string
          pattern: ^[a-zA-Z0-9][a-zA-Z0-9\-.]+[a-zA-Z0-9]$
        consumer_name:
          description: >-
            +label=Consumer Name

            +usage=Consumer name of input NATS

            +message=Consumer name should only contain alphanumeric letters,
            dashes(-), and underscores(_)

            +sort=4
          type: string
          pattern: ^[a-zA-Z0-9][a-zA-Z0-9\-_]+[a-zA-Z0-9]$
        wait_time_seconds:
          description: |-
            +label=Wait Time Seconds
            +usage=Wait timeout for long polling.
            +sort=5
          type: integer
          minimum: 1
          maximum: 20
          default: 19
        nats_metrics_url:
          description: >-
            +label=NATS metrics URL

            +usage=URL for the NATS metrics endpoint. It is compulsory if you
            want to use NATS autoscaling.

            +message=NATS Metrics URL should be a valid HTTP/HTTPS URL
          type: string
          pattern: ^(http(s?)://).*$
        auth:
          $ref: '#/components/schemas/NATSUserPasswordAuth'
      title: NATSInputConfig
      x-fern-type-name: NATSInputConfig
    KafkaInputConfig:
      description: |-
        +docs=Describes the configuration for the input Kafka worker
        +label=Kafka
      type: object
      required:
        - type
        - bootstrap_servers
        - topic_name
        - consumer_group
        - tls
      properties:
        type:
          description: +value=kafka
          type: string
          enum:
            - kafka
        bootstrap_servers:
          description: >-
            +label=Bootstrap servers

            +usage='Kafka Bootstrap servers - Comma separated list of Kafka
            brokers "hostname:port" to connect to for bootstrap'

            +sort=1
          type: string
        topic_name:
          description: |-
            +label=Topic Name
            +usage=Kafka topic to subscribe to
            +sort=2
          type: string
        consumer_group:
          description: >-
            +label=Consumer Group Name

            +usage=The name of the consumer group to join for dynamic partition
            assignment

            +sort=3
          type: string
        tls:
          description: |-
            +label=TLS
            +usage=TLS configuration for SASL authentication
          type: boolean
          default: true
        wait_time_seconds:
          description: |-
            +label=Wait Time Seconds
            +usage=Wait timeout for long polling.
          type: integer
          minimum: 1
          maximum: 300
          default: 10
        auth:
          $ref: '#/components/schemas/KafkaSASLAuth'
      title: KafkaInputConfig
      x-fern-type-name: KafkaInputConfig
    AMQPInputConfig:
      description: |-
        +docs=Describes the configuration for the input AMQP worker
        +label=AMQP
      type: object
      required:
        - type
        - url
        - queue_name
        - wait_time_seconds
      properties:
        type:
          description: +value=amqp
          type: string
          enum:
            - amqp
        url:
          description: |-
            +label=Queue URL
            +usage=AMQP Queue URL of Subscriber
            +sort=1
          type: string
          pattern: >-
            ^(amqp|amqps?)://(?:[^:@]+(?::[^:@]+)?@)?([^/?]+)(?:/([^?]+))?/?([^?]+)?(?:\?(.*))?$
        queue_name:
          description: |-
            +label=Queue Name
            +usage=AMQP Queue Name
            +sort=2
          type: string
        wait_time_seconds:
          description: |-
            +label=Wait Time Seconds
            +usage=Wait timeout for long polling.
          type: integer
          minimum: 1
          default: 5
      title: AMQPInputConfig
      x-fern-type-name: AMQPInputConfig
    SQSOutputConfig:
      description: |-
        +docs=Describes the configuration for the output SQS worker
        +label=SQS
      type: object
      required:
        - type
        - queue_url
        - region_name
      properties:
        type:
          description: +value=sqs
          type: string
          enum:
            - sqs
        queue_url:
          description: |-
            +label=Queue URL
            +usage=AWS SQS Queue URL of Publisher
            +sort=1
          type: string
        region_name:
          description: |-
            +label=Region Name
            +usage=AWS Region Name
            +sort=2
          type: string
        auth:
          $ref: '#/components/schemas/AWSAccessKeyAuth'
      title: SQSOutputConfig
      x-fern-type-name: SQSOutputConfig
    NATSOutputConfig:
      description: |-
        +docs=Describes the configuration for the output NATS worker
        +label=NATS
      type: object
      required:
        - type
        - nats_url
        - root_subject
      properties:
        type:
          description: +value=nats
          type: string
          enum:
            - nats
        nats_url:
          description: |-
            +label=NATS URL
            +usage=Output NATS URL
            +sort=1
          type: string
        root_subject:
          description: >-
            +label=Root Subject

            +usage=Root subject of output NATS

            +message=Output NATS root subject should only contain alphanumeric
            letters, dashes(-), and periods(.)

            +sort=2
          type: string
          pattern: ^[a-zA-Z0-9][a-zA-Z0-9\-.]+[a-zA-Z0-9]$
        auth:
          $ref: '#/components/schemas/NATSUserPasswordAuth'
      title: NATSOutputConfig
      x-fern-type-name: NATSOutputConfig
    CoreNATSOutputConfig:
      description: |-
        +docs=Describes the configuration for the output Core NATS worker
        +label=Core NATS
      type: object
      required:
        - type
        - nats_url
        - root_subject
      properties:
        type:
          description: +value=core-nats
          type: string
          enum:
            - core-nats
        nats_url:
          description: |-
            +label=NATS URL
            +usage=Output NATS URL
            +sort=1
          type: string
        root_subject:
          description: >-
            +label=Root Subject

            +usage=Root subject of output NATS

            +message=Output NATS root subject should only contain alphanumeric
            letters, dashes(-), and periods(.)

            +sort=2
          type: string
          pattern: ^[a-zA-Z0-9][a-zA-Z0-9\-.]+[a-zA-Z0-9]$
        auth:
          $ref: '#/components/schemas/NATSUserPasswordAuth'
      title: CoreNATSOutputConfig
      x-fern-type-name: CoreNATSOutputConfig
    KafkaOutputConfig:
      description: |-
        +docs=Describes the configuration for the output Kafka worker
        +label=Kafka
      type: object
      required:
        - type
        - bootstrap_servers
        - topic_name
        - tls
      properties:
        type:
          description: +value=kafka
          type: string
          enum:
            - kafka
        bootstrap_servers:
          description: >-
            +label=Bootstrap servers

            +usage='Kafka Bootstrap servers - Comma separated list of Kafka
            brokers "hostname:port" to connect to for bootstrap'

            +sort=1
          type: string
        topic_name:
          description: |-
            +label=Topic Name
            +usage=Kafka topic to publish to
            +sort=2
          type: string
        tls:
          description: |-
            +label=TLS
            +usage=TLS configuration for SASL authentication
          type: boolean
          default: true
        auth:
          $ref: '#/components/schemas/KafkaSASLAuth'
      title: KafkaOutputConfig
      x-fern-type-name: KafkaOutputConfig
    AMQPOutputConfig:
      description: |-
        +docs=Describes the configuration for the output AMQP worker
        +label=AMQP
      type: object
      required:
        - type
        - url
        - routing_key
      properties:
        type:
          description: +value=amqp
          type: string
          enum:
            - amqp
        url:
          description: |-
            +label=Queue URL
            +usage=AMQP Queue URL of Publisher
            +sort=1
          type: string
          pattern: >-
            ^(amqp|amqps?)://(?:[^:@]+(?::[^:@]+)?@)?([^/?]+)(?:/([^?]+))?/?([^?]+)?(?:\?(.*))?$
        routing_key:
          description: |-
            +label=Routing Key
            +usage=AMQP Routing Key to publish to.
            +sort=2
          type: string
        exchange_name:
          description: |-
            +label=Exchange Name
            +usage=AMQP Exchange Name
            +sort=3
          type: string
      title: AMQPOutputConfig
      x-fern-type-name: AMQPOutputConfig
    GitSource:
      description: >-
        +docs=Describes that we are using code stored in a git repository to
        build our image

        +label=Git Source

        +icon=fa-solid fa-code-branch:black

        +sort=300
      type: object
      required:
        - type
        - repo_url
        - ref
      properties:
        type:
          description: +value=git
          type: string
          enum:
            - git
        repo_url:
          description: >-
            +label=Repo URL

            +usage=The repository URL.

            +sort=1

            +message=Needs to be a valid Github, Bitbucket, Azure Repos or
            Gitlab link
          type: string
          pattern: ^(https?://)\S+$
        ref:
          description: |-
            +label=Commit SHA
            +usage=The commit SHA.
            +sort=2
          type: string
          pattern: ^\S+$
        branch_name:
          description: |-
            +label=Branch Name
            +usage=Selecting branch will select latest commit SHA of the branch.
            +sort=3
          type: string
          pattern: ^\S+$
      title: GitSource
      x-fern-type-name: GitSource
    DockerFileBuild:
      description: |-
        +docs=Describes that we are using a dockerfile to build our image
        +label=Docker File (I already have Docker File)
        +icon=fa-brands fa-docker:#0db7ed
      type: object
      required:
        - type
        - dockerfile_path
        - build_context_path
      properties:
        type:
          description: +value=dockerfile
          type: string
          enum:
            - dockerfile
        dockerfile_path:
          description: >-
            +label=Path to Dockerfile

            +usage=The file path of the Dockerfile relative to project root
            path.
          type: string
          default: ./Dockerfile
        build_context_path:
          description: >-
            +label=Path to build context

            +usage=Build context path for the Dockerfile relative to project
            root path.
          type: string
          default: ./
        command:
          description: >-
            +label=Command Override

            +usage=Override the command to run when the container starts

            When deploying a Job, the command can be templatized by defining
            `params` and referencing them in command

            E.g. `python main.py --learning_rate {{learning_rate}}`
          oneOf:
            - type: string
            - type: array
              items:
                type: string
        build_args:
          description: +label=Build arguments to pass to docker build
          type: object
          additionalProperties:
            type: string
        build_secrets:
          description: >-
            +label=Build secrets to pass to docker build

            +usage=Build secrets you can use in run statements inside dockerfile
            like RUN --mount=type=secret,id=pip_index_url ... you can learn more
            about build secrets here
            https://www.truefoundry.com/docs/docker-build-secrets

            +uiProps={"showTableView": true, "widthRatios": {"id":0.4286}}
          type: array
          items:
            $ref: '#/components/schemas/BuildSecret'
      title: DockerFileBuild
      x-fern-type-name: DockerFileBuild
    PythonBuild:
      description: >-
        +docs=Describes that we are using python to build a container image with
        a specific python version and pip packages installed.

        +label=Python Code (I don't have Dockerfile)

        +icon=fa-brands fa-python:#306998
      type: object
      required:
        - type
        - build_context_path
        - command
      properties:
        type:
          description: +value=tfy-python-buildpack
          type: string
          enum:
            - tfy-python-buildpack
        python_version:
          description: >-
            +label=Python version

            +usage=Python version to run your application. Should be one of the
            tags listed on [Official Python Docker
            Page](https://hub.docker.com/_/python)

            +message=Please enter a valid Python version tag

            +sort=10002
          type: string
          pattern: ^\d+(\.\d+){1,2}([\-\.a-z0-9]+)?$
        build_context_path:
          description: |-
            +label=Path to build context
            +usage=Build path relative to project root path.
            +sort=10000
          type: string
          default: ./
        requirements_path:
          description: |-
            `Path to build context`
            +label=Path to requirements
            +usage=Path to `requirements.txt` relative to
            `Path to build context`
            +uiType=Hidden
            +sort=10004
          type: string
        pip_packages:
          description: |-
            +label=Pip packages to install
            +usage=Define pip package requirements.
            In Python/YAML E.g. ["fastapi>=0.90,<1.0", "uvicorn"]
            +placeholder=Enter a pip package name E.g. fastapi>=0.90,<1.0
            +uiType=Hidden
          type: array
          items:
            type: string
        python_dependencies:
          description: |-
            +label=Python dependencies
            +usage=Python dependencies to install
            +sort=10004
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/Pip'
            - $ref: '#/components/schemas/UV'
            - $ref: '#/components/schemas/Poetry'
          discriminator:
            propertyName: type
        apt_packages:
          description: |-
            +label=List of Debian packages to install.
            +usage=Debian packages to install via `apt get`.
            In Python/YAML E.g. ["git", "ffmpeg", "htop"]
            +placeholder=Enter a debian package name E.g. ffmpeg
          type: array
          items:
            type: string
        command:
          description: >-
            Command will be set as the Entrypoint of the generated

            image.

            +label=Command

            +usage=Command to run when the container starts.

            Command will be set as the Entrypoint of the generated image.

            When deploying a Job, the command can be templatized by defining
            `params` and referencing them in command

            E.g. `python main.py --learning_rate {{learning_rate}}`

            +sort=10001
          oneOf:
            - type: string
            - type: array
              items:
                type: string
        cuda_version:
          description: >-
            +label=CUDA Version

            +usage=Version of CUDA Toolkit and CUDNN to install in the image

            These combinations are based off of publically available docker
            images on docker hub

            You can also specify a valid tag of the form
            {cuda_version_number}-cudnn{cudnn_version_number}-{runtime|devel}-ubuntu{ubuntu_version}

            Refer https://hub.docker.com/r/nvidia/cuda/tags for valid set of
            values

            Note: We use deadsnakes ubuntu ppa to add Python that currently
            supports only Ubuntu 18.04, 20.04 and 22.04

            +sort=10003
          type: string
          pattern: >-
            ^((\d+\.\d+(\.\d+)?-cudnn(\d+)?-(runtime|devel)-ubuntu\d+\.\d+)|11\.0-cudnn8|11\.1-cudnn8|11\.2-cudnn8|11\.3-cudnn8|11\.4-cudnn8|11\.5-cudnn8|11\.6-cudnn8|11\.7-cudnn8|11\.8-cudnn8|12\.0-cudnn8|12\.1-cudnn8|12\.2-cudnn8|12\.3-cudnn9|12\.4-cudnn9|12\.5-cudnn9|12\.6-cudnn9|12\.8-cudnn9|12\.9-cudnn9)$
      title: PythonBuild
      x-fern-type-name: PythonBuild
    NodeSelector:
      description: >-
        +label=Node selector

        +usage=Constraints to select a Node - Specific GPU / Instance Families,
        On-Demand/Spot.
      type: object
      required:
        - type
      properties:
        type:
          description: +value=node_selector
          type: string
          enum:
            - node_selector
        instance_families:
          description: >-
            +label=Instance family

            +usage=Instance family of the underlying machine to use. Multiple
            instance families can be supplied.

            The workload is guaranteed to be scheduled on one of them.
          type: array
          items:
            type: string
        capacity_type:
          description: >-
            +label=Capacity Type

            +usage=Configure what type of nodes to run the app. By default no
            placement logic is applied.

            "spot_fallback_on_demand" will try to place the application on spot
            nodes but will fallback to on-demand when spot nodes are not
            available.

            "spot" will strictly place the application on spot nodes.

            "on_demand" will strictly place the application on on-demand nodes.
          type: string
          enum:
            - spot_fallback_on_demand
            - spot
            - on_demand
      title: NodeSelector
      x-fern-type-name: NodeSelector
    NodepoolSelector:
      description: |-
        +label=Nodepool selector
        +usage=Specify one or more nodepools to run your application on.
      type: object
      required:
        - type
      properties:
        type:
          description: +value=nodepool_selector
          type: string
          enum:
            - nodepool_selector
        nodepools:
          description: >-
            +label=Nodepools

            +usage=Nodepools where you want to run your workload. Multiple
            nodepools can be selected.
             The workload is guaranteed to be scheduled on one of the nodepool
          type: array
          items:
            type: string
      title: NodepoolSelector
      x-fern-type-name: NodepoolSelector
    NvidiaGPU:
      type: object
      required:
        - type
        - count
      properties:
        type:
          description: +value=nvidia_gpu
          type: string
          enum:
            - nvidia_gpu
        name:
          description: >-
            +label=GPU Name

            +usage=Name of the Nvidia GPU. One of [P4, P100, V100, T4, A10G,
            A100_40GB, A100_80GB]

            This field is required for Node Selector and can be ignored in
            Nodepool Selector.

            One instance of the card contains the following amount of memory -

            P4: 8 GB, P100: 16 GB, V100: 16 GB, T4: 16 GB, A10G: 24 GB,
            A100_40GB: 40GB, A100_80GB: 80 GB
          type: string
        count:
          description: >-
            +label=GPU Count

            +usage=Count of GPUs to provide to the application

            Note the exact count and max count available for a given GPU type
            depends on cloud provider and cluster type.
          type: integer
          minimum: 1
          maximum: 16
      title: NvidiaGPU
      x-fern-type-name: NvidiaGPU
    AWSInferentia:
      type: object
      required:
        - type
        - count
      properties:
        type:
          description: +value=aws_inferentia
          type: string
          enum:
            - aws_inferentia
        name:
          description: >-
            +label=Inferentia accelerator name

            +usage=Name of the AWS Inferentia Accccelerator. One of [INF1,
            INF2].

            This field is required for Node Selector and can be ignored in
            Nodepool Selector.
          type: string
        count:
          description: >-
            +label=Count

            +usage=Count of Inferentia accelerator chips to provide to the
            application
          type: integer
          minimum: 1
          maximum: 16
      title: AWSInferentia
      x-fern-type-name: AWSInferentia
    NvidiaMIGGPU:
      type: object
      required:
        - type
        - profile
      properties:
        type:
          description: +value=nvidia_mig_gpu
          type: string
          enum:
            - nvidia_mig_gpu
        name:
          description: >-
            +label=GPU Name

            +usage=Name of the Nvidia GPU. One of [P4, P100, V100, T4, A10G,
            A100_40GB, A100_80GB]

            This field is required for Node Selector and can be ignored in
            Nodepool Selector.

            One instance of the card contains the following amount of memory -

            P4: 8 GB, P100: 16 GB, V100: 16 GB, T4: 16 GB, A10G: 24 GB,
            A100_40GB: 40GB, A100_80GB: 80 GB
          type: string
        profile:
          description: >-
            +label=MIG Profile

            +usage=Name of the MIG profile to use. One of the following based on
            gpu type

            Please refer to
            https://docs.nvidia.com/datacenter/tesla/mig-user-guide/#supported-mig-profiles
            for more details

            A30 - [1g.6gb, 2g.12gb, 4g.24gb]

            A100 40 GB - [1g.5gb, 1g.10gb, 2g.10gb, 3g.20gb, 4g.20gb, 7g.40gb]

            A100 80 GB / H100 80 GB - [1g.10gb, 1g.20gb, 2g.20gb, 3g.40gb,
            4g.40gb, 7g.80gb]

            H100 94 GB - [1g.12gb, 1g.24gb, 2g.24gb, 3g.47gb, 4g.47gb, 7g.94gb]

            H100 96 GB - [1g.12gb, 1g.24gb, 2g.24gb, 3g.48gb, 4g.48gb, 7g.96gb]

            H200 141 GB - [1g.18gb, 1g.35gb, 2g.35gb, 3g.71gb, 4g.71gb]

            B200 180 GB - [1g.23gb, 1g.45gb, 2g.45gb, 3g.90gb, 4g.90gb,
            7g.180gb]
          type: string
          enum:
            - 1g.6gb
            - 2g.12gb
            - 1g.5gb
            - 1g.10gb
            - 2g.10gb
            - 3g.20gb
            - 4g.20gb
            - 1g.20gb
            - 2g.20gb
            - 3g.40gb
            - 4g.40gb
            - 1g.12gb
            - 1g.24gb
            - 2g.24gb
            - 3g.47gb
            - 4g.47gb
            - 3g.48gb
            - 4g.48gb
            - 1g.18gb
            - 1g.35gb
            - 2g.35gb
            - 3g.71gb
            - 4g.71gb
            - 1g.23gb
            - 1g.45gb
            - 2g.45gb
            - 3g.90gb
            - 4g.90gb
            - 4g.24gb
            - 7g.40gb
            - 7g.80gb
            - 7g.94gb
            - 7g.96gb
            - 7g.180gb
      title: NvidiaMIGGPU
      x-fern-type-name: NvidiaMIGGPU
    NvidiaTimeslicingGPU:
      type: object
      required:
        - type
        - gpu_memory
      properties:
        type:
          description: +value=nvidia_timeslicing_gpu
          type: string
          enum:
            - nvidia_timeslicing_gpu
        name:
          description: >-
            +label=GPU Name

            +usage=Name of the Nvidia GPU. One of [P4, P100, V100, T4, A10G,
            A100_40GB, A100_80GB]

            This field is required for Node Selector and can be ignored in
            Nodepool Selector.

            One instance of the card contains the following amount of memory -

            P4: 8 GB, P100: 16 GB, V100: 16 GB, T4: 16 GB, A10G: 24 GB,
            A100_40GB: 40GB, A100_80GB: 80 GB
          type: string
        gpu_memory:
          description: >-
            +label=GPU Memory (MB)

            +usage=Amount of GPU memory (in MB) to allocate. Please note, this
            limit is not being enforced today but will be in future.
            Applications are expected to operate in co-opertative mode
          type: integer
          minimum: 1
          maximum: 400000
      title: NvidiaTimeslicingGPU
      x-fern-type-name: NvidiaTimeslicingGPU
    GcpTPU:
      type: object
      required:
        - type
        - name
        - topology
      properties:
        type:
          description: +value=gcp_tpu
          type: string
          enum:
            - gcp_tpu
        name:
          description: |-
            +label=TPU Type name
            +usage=Name of the TPU Type. One of
              - `tpu-v4-podslice` (TPU v4, ct4p)
              - `tpu-v5-lite-device` (TPU v5e, ct5l)
              - `tpu-v5-lite-podslice`  (TPU v5e, ct5lp)
              - `tpu-v5p-slice` (TPU v5p, ct5p)
          type: string
          pattern: ^tpu-[a-z\d\-]+$
        topology:
          description: >-
            +label=Slice Topology

            +usage=Topology of the TPU slices. Currently only single-host
            topology is supported.
             Please refer to [TPUs on GKE docs](https://cloud.google.com/kubernetes-engine/docs/concepts/tpus#plan-tpu-configuration)
             Allowed Values:
               - `2x2x1` for `tpu-v4-podslice`
               - One of `1x1`, `2x2`, `2x4` for `tpu-v5-lite-device` and `tpu-v5-lite-podslice`
               - `2x2x1` for `tpu-v5p-slice`
          type: string
          pattern: ^\d+x\d+(x\d+)?$
      title: GcpTPU
      x-fern-type-name: GcpTPU
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
    Endpoint:
      type: object
      required:
        - host
      properties:
        host:
          description: >-
            +usage=Host e.g. ai.example.com, app.truefoundry.com

            +message=Upto 253 characters, each part of host should be at most 63
            characters long, can contain alphabets, digits and hypen, must begin
            and end with an alphanumeric characters. Parts must be separated by
            periods (.)
          type: string
          pattern: >-
            ^((([a-zA-Z0-9\-]{1,63}\.)([a-zA-Z0-9\-]{1,63}\.)*([A-Za-z]{1,63}))|(((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)))$
        path:
          description: >-
            +usage=Path e.g. /v1/api/ml/, /v2/docs/

            +message=Should begin and end with a forward slash (/). Each part
            can can contain alphabets, digits and hypen, must begin and end with
            an alphanumeric characters. Parts should be separated by forward
            slashes (/)
          type: string
          pattern: ^(/([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-_\.]*[a-zA-Z0-9]))*/$
      title: Endpoint
      x-fern-type-name: Endpoint
    FlyteTaskTemplate:
      type: object
      required:
        - id
        - custom
      properties:
        id:
          $ref: '#/components/schemas/FlyteTaskID'
        custom:
          $ref: '#/components/schemas/FlyteTaskCustom'
      title: FlyteTaskTemplate
      x-fern-type-name: FlyteTaskTemplate
    FlyteWorkflowTemplate:
      type: object
      required:
        - id
      properties:
        id:
          $ref: '#/components/schemas/FlyteWorkflowID'
      title: FlyteWorkflowTemplate
      x-fern-type-name: FlyteWorkflowTemplate
    FlyteLaunchPlanID:
      type: object
      required:
        - resourceType
        - name
      properties:
        resourceType:
          type: string
          enum:
            - LAUNCH_PLAN
        name:
          type: string
      title: FlyteLaunchPlanID
      x-fern-type-name: FlyteLaunchPlanID
    FlyteLaunchPlanSpec:
      type: object
      required:
        - workflowId
      properties:
        workflowId:
          $ref: '#/components/schemas/FlyteWorkflowID'
      title: FlyteLaunchPlanSpec
      x-fern-type-name: FlyteLaunchPlanSpec
    SparkBuild:
      description: >-
        +docs=Describes that we are using python to build a container image with
        a specific python version and pip packages installed.

        +label=Build specification

        +icon=fa-brands fa-python:#306998
      type: object
      required:
        - type
        - spark_version
        - build_context_path
      properties:
        type:
          description: +value=tfy-spark-buildpack
          type: string
          enum:
            - tfy-spark-buildpack
        spark_version:
          description: >-
            --- Spark Specific Field ---

            +label=Spark Version

            +usage=Spark version should match the spark version installed in the
            image.

            +sort=1001
          type: string
          default: 3.5.2
        build_context_path:
          description: |-
            +label=Path to build context
            +usage=Build path relative to project root path.
          type: string
          default: ./
        requirements_path:
          description: |-
            `Path to build context`
            +label=Path to requirements
            +usage=Path to `requirements.txt` relative to
            `Path to build context`
          type: string
      title: SparkBuild
      x-fern-type-name: SparkBuild
    SparkExecutorFixedInstances:
      description: +label=Fixed Instances
      type: object
      required:
        - type
        - count
      properties:
        type:
          description: +value=fixed
          type: string
          enum:
            - fixed
        count:
          description: |-
            +label=Instances Count
            +usage=Number of instances to start
          type: integer
          minimum: 0
          maximum: 500
          default: 1
      title: SparkExecutorFixedInstances
      x-fern-type-name: SparkExecutorFixedInstances
    SparkExecutorDynamicScaling:
      description: +label=Dynamic Scaling
      type: object
      required:
        - type
        - min
        - max
      properties:
        type:
          description: +value=dynamic
          type: string
          enum:
            - dynamic
        min:
          description: |-
            +label=Min Instances
            +usage=Minimum number of instances to start / scale down to
            +sort=100
          type: integer
          minimum: 0
          maximum: 500
          default: 1
        max:
          description: |-
            +label=Max Instances
            +usage=Maximum number of instances to scale up to
            +sort=200
          type: integer
          minimum: 0
          maximum: 500
          default: 1
      title: SparkExecutorDynamicScaling
      x-fern-type-name: SparkExecutorDynamicScaling
    ArtifactsCacheVolume:
      description: |-
        +docs=Describes the volume that will be used to cache the models
        +label=Artifacts Cache Volume
      type: object
      required:
        - storage_class
        - cache_size
      properties:
        storage_class:
          description: |-
            +label=Storage Class
            +usage=Storage class of the Volume where artifacts will be cached
          type: string
        cache_size:
          description: >-
            +label=Cache Size (GB)

            +usage=Size of the Volume (in GB) where artifacts will be cached.
            Should be greater than twice the size of artifacts getting cached
          type: integer
          minimum: 1
          maximum: 1000
          default: 200
      title: ArtifactsCacheVolume
      x-fern-type-name: ArtifactsCacheVolume
    TrueFoundryArtifactSource:
      description: |-
        +docs=Input for Artifact from TrueFoundry Artifact Registry
        +label=TrueFoundry Artifact Source
      type: object
      required:
        - type
        - artifact_version_fqn
        - download_path_env_variable
      properties:
        type:
          description: +value=truefoundry-artifact
          type: string
          enum:
            - truefoundry-artifact
        artifact_version_fqn:
          description: >-
            +label=Artifact or Model Version FQN

            +usage=Artifact or Model Version FQN of the artifact to be
            downloaded
          type: string
        download_path_env_variable:
          description: >-
            +label=Download Path Environment Variable

            +usage=Environment variable which will contain the download path of
            the artifact
          type: string
      title: TrueFoundryArtifactSource
      x-fern-type-name: TrueFoundryArtifactSource
    HuggingfaceArtifactSource:
      description: |-
        +docs=Input for Artifact from Huggingface Model Hub
        +label=Huggingface Model Source
      type: object
      required:
        - type
        - model_id
        - revision
        - ignore_patterns
        - download_path_env_variable
      properties:
        type:
          description: +value=huggingface-hub
          type: string
          enum:
            - huggingface-hub
        model_id:
          description: |-
            +label=Model ID
            +usage=Model ID of the artifact to be downloaded
          type: string
        revision:
          description: |-
            +label=Revision
            +usage=Revision of the artifact to be downloaded
          type: string
        ignore_patterns:
          description: |-
            +label=Ignore Patterns
            +usage=List of patterns to ignore while downloading the artifact
          type: array
          items:
            type: string
          default:
            - '*.h5'
            - '*.ot'
            - '*.tflite'
            - '*.msgpack'
        download_path_env_variable:
          description: >-
            +label=Download Path Environment Variable

            +usage=Environment variable which will contain the download path of
            the artifact
          type: string
      title: HuggingfaceArtifactSource
      x-fern-type-name: HuggingfaceArtifactSource
    BasicAuthCreds:
      description: +label=Username and password
      type: object
      required:
        - type
        - username
        - password
      properties:
        type:
          description: +value=basic_auth
          type: string
          enum:
            - basic_auth
        username:
          description: |-
            +label=Username for service auth
            +message=Username for the user to authenticate with
            +sort=1
          type: string
        password:
          description: |-
            +label=Password for service auth
            +message=Password for the user to authenticate with
            +sort=2
          type: string
      title: BasicAuthCreds
      x-fern-type-name: BasicAuthCreds
    JwtAuthConfig:
      description: |-
        +label=JWT Authentication
        +usage=Configure JWT-based authentication using JWKS
      type: object
      required:
        - type
        - integration_fqn
      properties:
        type:
          description: +value=jwt_auth
          type: string
          enum:
            - jwt_auth
        integration_fqn:
          description: >-
            +docs=FQN of the JWT Auth integration. You can use the FQN of your
            desired jwt auth integration (or add one)

            in the  Integrations page[Integrations](/integrations) page

            +label=JWT Auth Integration

            +sort=200

            +usage=FQN of the JWT Auth integration. If you can't find your
            integration here,

            add it through the [Integrations](/integrations) page

            +uiType=IntegrationSelect

            +uiProps={"integrationType":"jwt-auth"}
          type: string
        enable_login:
          description: |-
            +label=Enable Login
            +usage=Enable login for the service
            +sort=300
          type: boolean
        claims:
          description: |-
            +label=Claims
            +usage=List of key-value pairs of claims to verify in the JWT token
            +sort=400
          type: array
          items:
            type: object
            required:
              - key
              - values
            properties:
              key:
                type: string
              values:
                type: array
                items:
                  type: string
        bypass_auth_paths:
          description: >-
            +label=Paths that will bypass auth

            +usage=List of paths that will bypass auth.

            needs to start with a forward slash(/) and should not contain
            wildcards(*)

            +sort=500
          type: array
          items:
            type: string
            pattern: ^/[^*]*
      title: JwtAuthConfig
      x-fern-type-name: JwtAuthConfig
    TrueFoundryInteractiveLogin:
      description: +label=Login with truefoundry
      type: object
      required:
        - type
      properties:
        type:
          description: +value=truefoundry_oauth
          type: string
          enum:
            - truefoundry_oauth
        bypass_auth_paths:
          description: >-
            +label=Paths that will bypass auth

            +usage=List of paths that will bypass auth.

            needs to start with a forward slash(/) and should not contain
            wildcards(*)
          type: array
          items:
            type: string
            pattern: ^/[^*]*
      title: TrueFoundryInteractiveLogin
      x-fern-type-name: TrueFoundryInteractiveLogin
    HttpProbe:
      description: >-
        +docs=Describes the Instructions for assessing container health by
        executing an HTTP GET request.

        To learn more you can go
        [here](https://docs.truefoundry.com/docs/liveness-readiness-probe)

        +label=Instructions for assessing container health by executing an HTTP
        GET request.
      type: object
      required:
        - type
        - path
        - port
      properties:
        type:
          description: |-
            +sort=1
            +label=Request Type
            +value=http
          type: string
          enum:
            - http
        path:
          description: |-
            +usage=Path to the health check endpoint
            +sort=2
          type: string
        port:
          description: |-
            +usage=Listening port for the health check endpoint
            +sort=3
          type: integer
          minimum: 0
          maximum: 65535
        host:
          description: |-
            +sort=4
            +usage=Host name to connect to, defaults to the pod IP
          type: string
        scheme:
          description: |-
            +sort=5
            +usage=Scheme to use for connecting to the host
          type: string
          default: HTTP
      title: HttpProbe
      x-fern-type-name: HttpProbe
    AwsSqsAccessKeyBasedAuth:
      description: +label=AWS Access Key Based Auth
      type: object
      required:
        - type
        - access_key_id
        - secret_access_key
      properties:
        type:
          description: +value=access-key-based
          type: string
          enum:
            - access-key-based
        access_key_id:
          description: |-
            +label=AWS access key ID
            +usage=AWS access key ID that you want to use for authentication.
            +sort=100
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
        secret_access_key:
          description: >-
            +label=AWS access key secret

            +usage=AWS access key secret that you want to use for
            authentication.

            +sort=200

            +uiType=SecretSelect

            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: AwsSqsAccessKeyBasedAuth
      x-fern-type-name: AwsSqsAccessKeyBasedAuth
    AwsSqsAssumedRoleBasedAuth:
      description: +label=AWS Assumed Role Based Auth
      type: object
      required:
        - type
        - assumed_role_arn
      properties:
        type:
          description: +value=assumed-role-based
          type: string
          enum:
            - assumed-role-based
        assumed_role_arn:
          description: |-
            +label=AWS Assumed Role ARN
            +usage=ARN of the role you want to assume.
            +sort=100
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^arn:(aws|aws-us-gov|aws-cn):iam::[0-9]{12}:role(/[\w+=,.@-]+)+$
      title: AwsSqsAssumedRoleBasedAuth
      x-fern-type-name: AwsSqsAssumedRoleBasedAuth
    AWSAccessKeyAuth:
      type: object
      required:
        - aws_access_key_id
        - aws_secret_access_key
      properties:
        aws_access_key_id:
          description: |-
            +label=AWS Access Key ID
            +usage=AWS Access Key ID
            +sort=1
          type: string
        aws_secret_access_key:
          description: |-
            +label=AWS Secret Access Key
            +usage=AWS Secret Access Key for the user to authenticate with
            +sort=2
          type: string
        aws_session_token:
          description: >-
            +label=AWS Session Token

            +usage=AWS Session Token, only required when using temporary
            credentials

            +sort=3
          type: string
      title: AWSAccessKeyAuth
      x-fern-type-name: AWSAccessKeyAuth
    NATSUserPasswordAuth:
      description: |-
        +docs=NATS User Password Authentication
        +label=NATS User Password Authentication
      type: object
      required:
        - account_name
        - user
        - password
      properties:
        account_name:
          description: |-
            +label=Account Name
            +usage=Name of the NATS account
            +sort=1
          type: string
          default: $G
        user:
          description: |-
            +label=User
            +usage=User for NATS authentication
            +sort=2
          type: string
        password:
          description: |-
            +label=Password
            +usage=Password for NATS authentication
            +sort=3
          type: string
      title: NATSUserPasswordAuth
      x-fern-type-name: NATSUserPasswordAuth
    KafkaSASLAuth:
      type: object
      required:
        - username
        - password
      properties:
        username:
          description: |-
            +label=Username
            +usage=Username for SASL authentication
            +sort=1
          type: string
        password:
          description: |-
            +label=Password
            +usage=Password for SASL authentication
            +sort=2
          type: string
      title: KafkaSASLAuth
      x-fern-type-name: KafkaSASLAuth
    BuildSecret:
      type: object
      required:
        - id
        - value
      properties:
        id:
          description: |-
            +label=ID
            +usage=The ID of the build secret.
            +sort=1
          type: string
          pattern: ^\S+$
        value:
          description: |-
            +label=Value
            +usage=The value of the build secret.
            +uiType=SecretSelect
            +uiProps={"typeOptionsList":["password","secret"]}
            +sort=2
          type: string
        env:
          description: >-
            +label=Environment Variable Name

            +usage=The name of the environment variable to inject the secret
            into.

            +uiType=Hidden

            +sort=3
          type: string
          pattern: ^[_a-zA-Z0-9][_a-zA-Z0-9]{1,31}$
      title: BuildSecret
      x-fern-type-name: BuildSecret
    Pip:
      description: |-
        +usage=Use pip to install requirements files and packages
        +label=Pip
      type: object
      required:
        - type
      properties:
        type:
          description: +value=pip
          type: string
          enum:
            - pip
        requirements_path:
          description: >-
            +label=Path to requirements

            +usage=Path to `requirements.txt` relative to `Path to build
            context`
          type: string
        pip_packages:
          description: |-
            +label=Pip packages to install
            +usage=Define pip package requirements.
            In Python/YAML E.g. ["fastapi>=0.90,<1.0", "uvicorn"]
            +placeholder=Enter a pip package name E.g. fastapi>=0.90,<1.0
          type: array
          items:
            type: string
      title: Pip
      x-fern-type-name: Pip
    UV:
      description: |-
        +usage=Use `uv` to setup env.
        Your build context root must contain `pyproject.toml` and `uv.lock`
        +label=uv
      type: object
      required:
        - type
      properties:
        type:
          description: +value=uv
          type: string
          enum:
            - uv
        uv_version:
          description: |-
            +label=UV version
            +usage=UV version to use
          type: string
          default: latest
        sync_options:
          description: |-
            +label=Sync options
            +usage=Sync options to pass to uv command
          type: string
      title: UV
      x-fern-type-name: UV
    Poetry:
      description: |-
        +usage=Use `poetry` to setup env.
        Your build context root must contain `pyproject.toml` and `poetry.lock`
      type: object
      required:
        - type
      properties:
        type:
          description: +value=poetry
          type: string
          enum:
            - poetry
        poetry_version:
          description: |-
            +label=Poetry version
            +usage=Poetry version to use
          type: string
          default: latest
          oneOf:
            - enum:
                - latest
            - pattern: ^\d+(\.\d+){1,2}([\-\.a-z0-9]+)?$
        install_options:
          description: |-
            +label=Install options
            +usage=install options to pass to poetry command
          type: string
      title: Poetry
      x-fern-type-name: Poetry
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
    FlyteTaskID:
      type: object
      required:
        - resourceType
        - name
      properties:
        resourceType:
          type: string
          enum:
            - TASK
        name:
          type: string
      title: FlyteTaskID
      x-fern-type-name: FlyteTaskID
    FlyteTaskCustom:
      type: object
      required:
        - truefoundry
      properties:
        truefoundry:
          description: +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/PythonTaskConfig'
            - $ref: '#/components/schemas/ContainerTaskConfig'
            - $ref: '#/components/schemas/PySparkTaskConfig'
            - $ref: '#/components/schemas/DatabricksJobTaskConfig'
          discriminator:
            propertyName: type
      title: FlyteTaskCustom
      x-fern-type-name: FlyteTaskCustom
    FlyteWorkflowID:
      type: object
      required:
        - resourceType
        - name
      properties:
        resourceType:
          type: string
          enum:
            - WORKFLOW
        name:
          type: string
      title: FlyteWorkflowID
      x-fern-type-name: FlyteWorkflowID
    PythonTaskConfig:
      description: +docs=Describes the configuration for the python function task
      type: object
      required:
        - type
        - image
      properties:
        type:
          description: +value=python-task-config
          type: string
          enum:
            - python-task-config
        image:
          description: |-
            +label=Image Spec
            +docs=Specification for the image to be used for the task
            +sort=100
            +usage=Specify the image spec for the task
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/TaskPythonBuild'
            - $ref: '#/components/schemas/TaskDockerFileBuild'
          discriminator:
            propertyName: type
        env:
          description: >-
            +label=Environment Variables

            +usage=Configure environment variables to be injected in the task
            either as plain text or secrets.
            [Docs](https://docs.truefoundry.com/docs/env-variables)

            +icon=fa-globe

            +sort=200
          type: object
          default: null
          additionalProperties:
            type: string
          nullable: true
        resources:
          $ref: '#/components/schemas/Resources'
        mounts:
          description: |-
            +usage=Configure data to be mounted to Workflow pod(s) as a volume.
            +sort=400
            +todo=Add discriminator for union types in array
          type: array
          items:
            type: object
            oneOf:
              - $ref: '#/components/schemas/SecretMount'
              - $ref: '#/components/schemas/StringDataMount'
              - $ref: '#/components/schemas/VolumeMount'
        service_account:
          description: |-
            +label=Service Account
            +sort=500
          type: string
      title: PythonTaskConfig
      x-fern-type-name: PythonTaskConfig
    ContainerTaskConfig:
      type: object
      required:
        - type
        - image
      properties:
        type:
          description: +value=container-task-config
          type: string
          enum:
            - container-task-config
        image:
          description: >-
            +docs=Specify whether you want to deploy a Docker image or build and
            deploy from source code

            +label=Deploy a Docker image or build and deploy from source code

            +icon=fa-solid fa-cloud-arrow-up:#21B6A8

            +sort=200

            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/Build'
            - $ref: '#/components/schemas/Image'
          discriminator:
            propertyName: type
        env:
          description: >-
            +label=Environment Variables

            +usage=Configure environment variables to be injected in the task
            either as plain text or secrets.
            [Docs](https://docs.truefoundry.com/docs/env-variables)

            +icon=fa-globe

            +sort=200
          type: object
          default: null
          additionalProperties:
            type: string
          nullable: true
        resources:
          $ref: '#/components/schemas/Resources'
        mounts:
          description: |-
            +usage=Configure data to be mounted to Workflow pod(s) as a volume.
            +sort=400
            +todo=Add discriminator for union types in array
          type: array
          items:
            type: object
            oneOf:
              - $ref: '#/components/schemas/SecretMount'
              - $ref: '#/components/schemas/StringDataMount'
              - $ref: '#/components/schemas/VolumeMount'
        service_account:
          description: |-
            +label=Service Account
            +sort=500
          type: string
      title: ContainerTaskConfig
      x-fern-type-name: ContainerTaskConfig
    PySparkTaskConfig:
      type: object
      required:
        - type
        - image
        - driver_config
        - executor_config
      properties:
        type:
          description: +value=pyspark-task-config
          type: string
          enum:
            - pyspark-task-config
        image:
          $ref: '#/components/schemas/TaskPySparkBuild'
        driver_config:
          $ref: '#/components/schemas/SparkDriverConfig'
        executor_config:
          $ref: '#/components/schemas/SparkExecutorConfig'
        spark_conf:
          description: >-
            +label=Spark Config Properties

            +usage=Extra configuration properties to be passed to the spark job.
            [Docs](https://spark.apache.org/docs/latest/configuration.html)

            +icon=fa-gear:#68BBE3

            +sort=3000
          type: object
          default: null
          additionalProperties:
            type: string
            minLength: 1
          nullable: true
        env:
          description: >-
            +label=Environment Variables

            +usage=Configure environment variables to be injected in the task
            either as plain text or secrets.
            [Docs](https://docs.truefoundry.com/docs/env-variables)

            +icon=fa-globe

            +sort=4000
          type: object
          default: null
          additionalProperties:
            type: string
          nullable: true
        mounts:
          description: |-
            +usage=Configure data to be mounted to Workflow pod(s) as a volume.
            +sort=5000
          type: array
          items:
            $ref: '#/components/schemas/VolumeMount'
        service_account:
          description: |-
            +label=Service Account
            +sort=6000
          type: string
      title: PySparkTaskConfig
      x-fern-type-name: PySparkTaskConfig
    DatabricksJobTaskConfig:
      description: >-
        +docs=Task that triggers a Databricks job via API and polls until
        completion. Uses image and optional resources; no env or mounts.
        Execution is handled by the CLI (trigger + poll).

        +label=Databricks Job Task

        +icon=fa-solid fa-database:#FF3621
      type: object
      required:
        - type
        - image
        - workspace_host
        - job_id
      properties:
        type:
          description: +value=databricks-job-task-config
          type: string
          enum:
            - databricks-job-task-config
        image:
          description: |-
            +label=Image Spec
            +docs=Specification for the image to be used for the task
            +sort=100
            +usage=Specify the image spec for the task
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/TaskPythonBuild'
            - $ref: '#/components/schemas/TaskDockerFileBuild'
          discriminator:
            propertyName: type
        workspace_host:
          description: >-
            +label=Workspace Host

            +usage=Databricks workspace URL, e.g.
            https://<workspace>.cloud.databricks.com

            +sort=150
          type: string
        job_id:
          description: >-
            +label=Job ID

            +usage=The Databricks job ID to run (from Databricks workspace
            Jobs).

            +sort=200
          type: string
        service_account:
          description: |-
            +label=Service Account
            +sort=500
          type: string
        job_parameters:
          description: >-
            +label=Job Parameters

            +usage=Optional parameters to pass to the job run (e.g. notebook
            params, jar params).

            +sort=300
          type: object
          default: null
          additionalProperties:
            type: string
          nullable: true
        timeout_seconds:
          description: >-
            +label=Timeout (seconds)

            +usage=Maximum seconds to wait for the job run to complete. Used by
            CLI when polling.

            +sort=500
          type: number
          default: null
          nullable: true
        skip_wait_for_completion:
          description: >-
            +label=Skip Wait for Completion

            +usage=If false, the task waits for the Databricks job run to
            complete (trigger and poll). If true, only triggers the job and
            returns. Default false.

            +sort=505
          type: boolean
          default: false
        env:
          description: >-
            +label=Environment Variables

            +usage=Configure environment variables to be injected in the task
            either as plain text or secrets.
            [Docs](https://docs.truefoundry.com/docs/env-variables)

            +icon=fa-globe

            +sort=200
          type: object
          default: null
          additionalProperties:
            type: string
          nullable: true
        resources:
          $ref: '#/components/schemas/Resources'
      title: DatabricksJobTaskConfig
      x-fern-type-name: DatabricksJobTaskConfig
    TaskPythonBuild:
      description: |-
        +docs=Describes the configuration for the python build for a task
        +label=Python Buid Spec
        +icon=fa-brands fa-python:#306998
      type: object
      required:
        - type
      properties:
        type:
          description: |-
            +uiType=Hidden
            +value=task-python-build
          type: string
          enum:
            - task-python-build
        docker_registry:
          description: >-
            +docs=FQN of the container registry. You can the FQN of your desired
            container registry (or add one)

            in the  Integrations
            page[Integrations](https://app.truefoundry.tech/integrations?tab=docker-registry)
            page

            +label=Docker Registry

            +usage=FQN of the container registry. If you can't find your
            registry here,

            add it through the [Integrations](/integrations?tab=docker-registry)
            page
          type: string
        python_version:
          description: >-
            +label=Python version

            +usage=Python version to run your application. Should be one of the
            tags listed on [Official Python Docker
            Page](https://hub.docker.com/_/python)

            +message=Please enter a valid Python version tag
          type: string
          pattern: ^\d+(\.\d+){1,2}([\-\.a-z0-9]+)?$
        requirements_path:
          description: |-
            `Path to build context`
            +label=Path to requirements
            +usage=Path to `requirements.txt` relative to
            `Path to build context`
          type: string
        pip_packages:
          description: |-
            +label=Pip packages to install
            +usage=Define pip package requirements.
            In Python/YAML E.g. ["fastapi>=0.90,<1.0", "uvicorn"]
            +placeholder=Enter a pip package name E.g. fastapi>=0.90,<1.0
          type: array
          items:
            type: string
        apt_packages:
          description: |-
            +label=List of Debian packages to install.
            +usage=Debian packages to install via `apt get`.
            In Python/YAML E.g. ["git", "ffmpeg", "htop"]
            +placeholder=Enter a debian package name E.g. ffmpeg
          type: array
          items:
            type: string
        cuda_version:
          description: >-
            +label=CUDA Version

            +usage=Version of CUDA Toolkit and CUDNN to install in the image

            These combinations are based off of publically available docker
            images on docker hub

            You can also specify a valid tag of the form
            {cuda_version_number}-cudnn{cudnn_version_number}-{runtime|devel}-ubuntu{ubuntu_version}

            Refer https://hub.docker.com/r/nvidia/cuda/tags for valid set of
            values

            Note: We use deadsnakes ubuntu ppa to add Python that currently
            supports only Ubuntu 18.04, 20.04 and 22.04
          type: string
          pattern: >-
            ^((\d+\.\d+(\.\d+)?-cudnn(\d+)?-(runtime|devel)-ubuntu\d+\.\d+)|11\.0-cudnn8|11\.1-cudnn8|11\.2-cudnn8|11\.3-cudnn8|11\.4-cudnn8|11\.5-cudnn8|11\.6-cudnn8|11\.7-cudnn8|11\.8-cudnn8|12\.0-cudnn8|12\.1-cudnn8|12\.2-cudnn8|12\.3-cudnn9|12\.4-cudnn9|12\.5-cudnn9|12\.6-cudnn9|12\.8-cudnn9|12\.9-cudnn9)$
      title: TaskPythonBuild
      x-fern-type-name: TaskPythonBuild
    TaskDockerFileBuild:
      description: |-
        +docs=Describes the configuration for the docker build for a task
        +label=Docker File
        +icon=fa-brands fa-docker:#0db7ed
      type: object
      required:
        - type
        - dockerfile_path
      properties:
        type:
          description: |-
            +uiType=Hidden
            +value=dockerfile
          type: string
          enum:
            - task-dockerfile-build
        docker_registry:
          description: >-
            +docs=FQN of the container registry. You can the FQN of your desired
            container registry (or add one)

            in the  Integrations
            page[Integrations](https://app.truefoundry.tech/integrations?tab=docker-registry)
            page

            +label=Docker Registry

            +usage=FQN of the container registry. If you can't find your
            registry here,

            add it through the [Integrations](/integrations?tab=docker-registry)
            page
          type: string
        dockerfile_path:
          description: >-
            +label=Path to Dockerfile

            +usage=The file path of the Dockerfile relative to project root
            path.
          type: string
          default: ./Dockerfile
        build_args:
          description: +label=Build arguments to pass to docker build
          type: object
          additionalProperties:
            type: string
      title: TaskDockerFileBuild
      x-fern-type-name: TaskDockerFileBuild
    TaskPySparkBuild:
      description: |-
        +docs=Describes the configuration for the PySpark build for a task
        +label=PySpark Build Spec
        +icon=fa-brands fa-python:#306998
      type: object
      required:
        - type
        - spark_version
      properties:
        type:
          description: |-
            +uiType=Hidden
            +value=task-pyspark-build
          type: string
          enum:
            - task-pyspark-build
        spark_version:
          description: >-
            --- Spark Specific Field ---

            +label=Spark Version

            +usage=Spark version should match the spark version installed in the
            image.

            +sort=1001
          type: string
          default: 3.5.2
        container_image:
          description: >-
            +label=Container Image

            +usage=Custom container image URI. If provided, this image is used
            instead of the

            default Spark base image (public.ecr.aws/bitnami/spark). The image
            must be

            Debian-based and have Python and Spark pre-installed.

            +sort=1002
          type: string
          pattern: ^\S+$
        docker_registry:
          description: >-
            +docs=FQN of the container registry. You can the FQN of your desired
            container registry (or add one)

            in the  Integrations
            page[Integrations](https://app.truefoundry.tech/integrations?tab=docker-registry)
            page

            +label=Docker Registry

            +usage=FQN of the container registry. If you can't find your
            registry here,

            add it through the [Integrations](/integrations?tab=docker-registry)
            page

            +uiType=IntegrationSelect

            +uiProps={"descriptionKey":"registry_url","integrationType":"docker-registry"}
          type: string
        requirements_path:
          description: |-
            `Path to build context`
            +label=Path to requirements
            +usage=Path to `requirements.txt` relative to
            `Path to build context`
          type: string
        pip_packages:
          description: |-
            +label=Pip packages to install
            +usage=Define pip package requirements.
            In Python/YAML E.g. ["fastapi>=0.90,<1.0", "uvicorn"]
            +placeholder=Enter a pip package name E.g. fastapi>=0.90,<1.0
          type: array
          items:
            type: string
        apt_packages:
          description: |-
            +label=List of Debian packages to install.
            +usage=Debian packages to install via `apt get`.
            In Python/YAML E.g. ["git", "ffmpeg", "htop"]
            +placeholder=Enter a debian package name E.g. ffmpeg
          type: array
          items:
            type: string
      title: TaskPySparkBuild
      x-fern-type-name: TaskPySparkBuild
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````