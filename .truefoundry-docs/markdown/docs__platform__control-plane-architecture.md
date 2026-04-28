> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Truefoundry Control Plane

> Architecture of the TrueFoundry Control Plane which is the brain of the Truefoundry platform.

The control-plane is the brain of the Truefoundry platform. It is responsible for managing the overall configuration and orchestration of the platform components.

<Note>
  We provide a hosted control-plane that you can use or also self-host the control-plane as part of our enterprise plan
</Note>

<Frame caption="Truefoundry Control-Plane">
  <img src="https://mintcdn.com/truefoundry/Gz34IEy99_S9XhHC/images/docs/platform/control-plane-internal.png?fit=max&auto=format&n=Gz34IEy99_S9XhHC&q=85&s=e47a2ec28d0c2b1a69205ed9722e4dc8" alt="Truefoundry Control-Plane" width="3416" height="2118" data-path="images/docs/platform/control-plane-internal.png" />
</Frame>

The control-plane comprises of the following key components. Different components are required for different use-cases. If you are not using those features, they can be disabled and will not be installed.

| Component                  | AI Deployment                                               | AI Gateway                                                  | Description                                                                                                                                                                                                       |
| -------------------------- | ----------------------------------------------------------- | ----------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **UI**                     | <Icon icon="square-check" iconType="regular" />             | <Icon icon="square-check" iconType="regular" />             | This is the UI component to view deployments, and other resources.                                                                                                                                                |
| **Backend Microservice**   | <Icon icon="square-check" iconType="regular" />             | <Icon icon="square-check" iconType="regular" />             | Truefoundry comprises of multiple backend services that handle various aspects like authorization, deployment control flow, CRUD APIs, interaction with database and external services, etc.                      |
| **PostgreSQL**             | <Icon icon="square-check" iconType="regular" />             | <Icon icon="square-check" iconType="regular" />             | Database to store user information, deployment information, etc, gateway configuration, etc. This can be deployed on Kubernetes in dev-mode, however, we recommend using managed database like RDS in production. |
| **Blob Storage**           | <Icon icon="square-check" iconType="regular" />             | <Icon icon="square-check" iconType="regular" />             | Control-plane needs access to a blob storage to store code uploads, gateway logs and traces.                                                                                                                      |
| **Queue**                  | <Icon icon="square-check" iconType="regular" />             | <Icon icon="square-check" iconType="regular" />             | Truefoundry uses NATS as the queue and cache layer to communicate all data from the control plane to the compute and gateway planes.                                                                              |
| **Otel Collector**         | <Icon icon="square-xmark" iconType="regular" color="red" /> | <Icon icon="square-check" iconType="regular" />             | Otel collector to collect tracing data from AI Gateway                                                                                                                                                            |
| **Ingestor**               | <Icon icon="square-xmark" iconType="regular" color="red" /> | <Icon icon="square-check" iconType="regular" />             | Ingestor receives the logs from the OTEL collector and dumps them to S3 or any other blob storage.                                                                                                                |
| **Controller**             | <Icon icon="square-check" iconType="regular" />             | <Icon icon="square-xmark" iconType="regular" color="red" /> | The controller is responsible for handling all connections from the tfy-agents in different compute-plane clusters.                                                                                               |
| **Image Builder**          | <Icon icon="square-check" iconType="regular" />             | <Icon icon="square-xmark" iconType="regular" color="red" /> | ImageBuilder is used in AI deployment to build docker images on Kubernetes cluster using buildkit.                                                                                                                |
| **Workflows Microservice** | <Icon icon="square-check" iconType="regular" />             | <Icon icon="square-xmark" iconType="regular" color="red" /> | Workflows microservice is used in AI deployment to manage Truefoundry workflows. This is needed only if you are using [workflows feature](/docs/introduction-to-workflow).                                        |
| **Spark History Server**   | <Icon icon="square-check" iconType="regular" />             | <Icon icon="square-xmark" iconType="regular" color="red" /> | Spark history server is used in AI deployment to view the logs and traces of the spark jobs deployed on Truefoundry . This is needed only if you are using deploying Spark jobs on Truefoundry.                   |

Control-plane also integrates with a number of external services to make it easier to manage your deployments and configuration. These integrations are optional and you can only enable the ones you need. The key types of integrations and their usecases are as follows:

| **Integration Type**                                                         |                  AI Deployment                  |                          AI Gateway                         | Description                                                                                                                      |
| ---------------------------------------------------------------------------- | :---------------------------------------------: | :---------------------------------------------------------: | -------------------------------------------------------------------------------------------------------------------------------- |
| **Docker Registry**                                                          | <Icon icon="square-check" iconType="regular" /> | <Icon icon="square-xmark" iconType="regular" color="red" /> | Needed to pull images for deployed services. Required for building and deploying AI services.                                    |
| **Blob Storage**                                                             | <Icon icon="square-check" iconType="regular" /> |       <Icon icon="square-check" iconType="regular" />       | Used to store code uploads, model artifacts, logs, and gateway traces.                                                           |
| **External Secrets Manager (e.g., AWS Secrets Manager, GCP Secret Manager)** | <Icon icon="square-check" iconType="regular" /> |       <Icon icon="square-check" iconType="regular" />       | Used to securely inject sensitive credentials into services.                                                                     |
| **Authentication System (OIDC, SSO)**                                        | <Icon icon="square-check" iconType="regular" /> |       <Icon icon="square-check" iconType="regular" />       | Provides user authentication and single sign-on capabilities for the platform.                                                   |
| **Notifications (Slack, Email, Webhooks)**                                   | <Icon icon="square-check" iconType="regular" /> |       <Icon icon="square-check" iconType="regular" />       | Allows configuring notifications for deployment status, failure alerts, and other platform events.                               |
| **Git Repository Integration**                                               | <Icon icon="square-check" iconType="regular" /> | <Icon icon="square-xmark" iconType="regular" color="red" /> | Enables seamless integration with Git repositories for CI/CD workflows, code synchronization, and automated deployment triggers. |

### Supported Integrations

<Tabs>
  <Tab title="Docker Registry">
    <CardGroup cols={3}>
      <Card>
        <div className="flex items-center gap-2 w-full">
          <Icon icon="aws" />

          <span className="text-[15px] font-bold sm:text-base">AWS Container Registry</span>
        </div>
      </Card>

      <Card>
        <div className="flex items-center gap-2 w-full">
          <Icon icon="google" />

          <span className="text-[15px] font-bold sm:text-base">GCP Artifact Registry</span>
        </div>
      </Card>

      <Card>
        <div className="flex items-center gap-2 w-full">
          <Icon icon="microsoft" />

          <span className="text-[15px] font-bold sm:text-base">Azure Container Registry</span>
        </div>
      </Card>

      <Card>
        <div className="flex items-center gap-2 w-full">
          <Icon icon="docker" />

          <span className="text-[15px] font-bold sm:text-base">Docker Hub</span>
        </div>
      </Card>

      <Card>
        <div className="flex items-center gap-2 w-full">
          <Icon icon="quay" />

          <span className="text-[15px] font-bold sm:text-base">Quay</span>
        </div>
      </Card>

      <Card>
        <div className="flex items-center gap-2 w-full">
          <Icon icon="jfrog" />

          <span className="text-[15px] font-bold sm:text-base">JFrog Artifactory</span>
        </div>
      </Card>

      <Card>
        <div className="flex items-center gap-2 w-full">
          <Icon icon="server" />

          <span className="text-[15px] font-bold sm:text-base">Self-hosted registries</span>
        </div>
      </Card>
    </CardGroup>
  </Tab>

  <Tab title="Blob Storage">
    <CardGroup cols={3}>
      <Card>
        <div className="flex items-center gap-2 w-full">
          <Icon icon="aws" />

          <span className="text-[15px] font-bold sm:text-base">AWS S3</span>
        </div>
      </Card>

      <Card>
        <div className="flex items-center gap-2 w-full">
          <Icon icon="google" />

          <span className="text-[15px] font-bold sm:text-base">GCP Storage</span>
        </div>
      </Card>

      <Card>
        <div className="flex items-center gap-2 w-full">
          <Icon icon="microsoft" />

          <span className="text-[15px] font-bold sm:text-base">Azure Storage</span>
        </div>
      </Card>

      <Card>
        <div className="flex items-center gap-2 w-full">
          <Icon icon="minio" />

          <span className="text-[15px] font-bold sm:text-base">MinIO</span>
        </div>
      </Card>

      <Card>
        <div className="flex items-center gap-2 w-full">
          <Icon icon="purestorage" />

          <span className="text-[15px] font-bold sm:text-base">PureStorage</span>
        </div>
      </Card>

      <Card>
        <div className="flex items-center gap-2 w-full">
          <Icon icon="netapp" />

          <span className="text-[15px] font-bold sm:text-base">NetApp</span>
        </div>
      </Card>
    </CardGroup>
  </Tab>

  <Tab title="Secret Manager">
    <CardGroup cols={3}>
      <Card>
        <div className="flex items-center gap-2 w-full">
          <Icon icon="aws" />

          <span className="text-[15px] font-bold sm:text-base">AWS Parameter Store</span>
        </div>
      </Card>

      <Card>
        <div className="flex items-center gap-2 w-full">
          <Icon icon="aws" />

          <span className="text-[15px] font-bold sm:text-base">AWS Secrets Manager</span>
        </div>
      </Card>

      <Card>
        <div className="flex items-center gap-2 w-full">
          <Icon icon="google" />

          <span className="text-[15px] font-bold sm:text-base">GCP Secret Manager</span>
        </div>
      </Card>

      <Card>
        <div className="flex items-center gap-2 w-full">
          <Icon icon="microsoft" />

          <span className="text-[15px] font-bold sm:text-base">Azure Key Vault</span>
        </div>
      </Card>
    </CardGroup>
  </Tab>

  <Tab title="Notifications">
    <CardGroup cols={3}>
      <Card>
        <div className="flex items-center gap-2 w-full">
          <Icon icon="envelope" />

          <span className="text-[15px] font-bold sm:text-base">Email</span>
        </div>
      </Card>

      <Card>
        <div className="flex items-center gap-2 w-full">
          <Icon icon="slack" />

          <span className="text-[15px] font-bold sm:text-base">Slack Webhook/Bot</span>
        </div>
      </Card>

      <Card>
        <div className="flex items-center gap-2 w-full">
          <Icon icon="pagerduty" />

          <span className="text-[15px] font-bold sm:text-base">PagerDuty</span>
        </div>
      </Card>
    </CardGroup>
  </Tab>

  <Tab title="Git Repositories">
    <CardGroup cols={3}>
      <Card>
        <div className="flex items-center gap-2 w-full">
          <Icon icon="github" />

          <span className="text-[15px] font-bold sm:text-base">GitHub</span>
        </div>
      </Card>

      <Card>
        <div className="flex items-center gap-2 w-full">
          <Icon icon="gitlab" />

          <span className="text-[15px] font-bold sm:text-base">GitLab</span>
        </div>
      </Card>

      <Card>
        <div className="flex items-center gap-2 w-full">
          <Icon icon="bitbucket" />

          <span className="text-[15px] font-bold sm:text-base">Bitbucket</span>
        </div>
      </Card>

      <Card>
        <div className="flex items-center gap-2 w-full">
          <Icon icon="microsoft" />

          <span className="text-[15px] font-bold sm:text-base">Azure Repos</span>
        </div>
      </Card>
    </CardGroup>
  </Tab>
</Tabs>
