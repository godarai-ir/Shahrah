> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Infra Set Up For Workflows

> Infrastructure setup guide for running workflows reliably at scale.

Setting up workflows in a Workload Cluster (already connected to truefoundry) requires the following configuration to be done:

### Requirements:

1. Cloud Storage Bucket (S3/GCS/AzureBlob)
2. Bucket Access - IAM permissions to access the bucket

<Note>
  If you have created the compute plane cluster using TrueFoundry, a bucket and the required IAM roles are created and attached as an integration provider in the platform.

  Ideally the Blob Storage and the cluster should be in the same region, if created manually.
</Note>

### Step 1: Create an integration provider for the cloud storage bucket

1. Follow this [document](/docs/integrations-overview) to create an integration provider for your cloud provider. You need to create an IAM role with the right trust relationship for the workflow propeller to access the storage bucket
2. In the "Clusters" page, edit your cluster from the three dots and update the "Workflow Storage Integration" to use your cloud storage bucket.

### Step 2: Install Workflow Propeller in the cluster

To install the workflow propeller, you need to follow the steps below depending on the cloud provider.

1. From the cluster page, click on the "Add-Ons" tab and scroll to the "Tfy Workflow Propeller" section.
2. Click on "Install" and proceed with workspace creation.
3. The values for the helm chart are given below depending on the cloud provider.
4. There are few common values that are common to all the cloud providers. These can be found from the three dots against your cluster and then clicking on "Show Cluster Token".
   * Tenant name - This is the tenant name of your TrueFoundry account.
   * Control Plane URL - This is the URL of your TrueFoundry control plane.
   * TFY agent token - This is the token of the cluster you connected.

<Tabs>
  <Tab title="AWS">
    For AWS we need the following values

    1. AWS S3 bucket name - This is the name of the S3 bucket used in the integration provider.
    2. AWS region - This is the region of the S3 bucket.
    3. AWS IAM role ARN - AWS IAM role ARN used in the integration provider.

    Final values file should like this. Ensure to replace the placeholders with the actual values.

    <CodeGroup>
      ```yaml YAML lines theme={"dark"}
      global:
        tenantName: <TENANT_NAME>
        controlPlaneUrl: <CONTROL_PLANE_URL_IN_HTTPS_FORMAT>
      flyte-core:
        storage:
          type: s3
          bucketName: <AWS_S3_BUCKET_NAME>
          connection:
            region: <AWS_REGION>
            auth-type: iam
          enable-multicontainer: true
        configmap:
          core:
            propeller:
              metadata-prefix: s3://<AWS_S3_BUCKET_NAME>/tfy-workflow-propeller/metatdata
              rawoutput-prefix: s3://<AWS_S3_BUCKET_NAME>/tfy-workflow-propeller/raw_data
          admin:
            admin:
              Command:
                - echo
                - <TFY_AGENT_TOKEN>
              endpoint: <CONTROL_PLANE_URL>:443
        flyteadmin:
          serviceAccount:
            alwaysCreate: true
        flytepropeller:
          serviceAccount:
            annotations:
              eks.amazonaws.com/role-arn: <AWS_IAM_ROLE_ARN>
      tfySignedURLServer:
        env:
          AWS_REGION: <AWS_REGION>
          S3_BUCKET_NAME: s3://<AWS_S3_BUCKET_NAME>
          DEFAULT_CLOUD_PROVIDER: aws
        enabled: true
      ```
    </CodeGroup>
  </Tab>

  <Tab title="GCP">
    For GCP we need the following values

    1. GCP project ID - This is the project ID of the GCP project used in the integration provider.
    2. GCP region - This is the region of the GCS bucket.
    3. GCP Service Account Email - GCP Service Account Email used in the integration provider.

    Final values file should like this. Ensure to replace the placeholders with the actual values.

    <CodeGroup>
      ```yaml YAML lines theme={"dark"}
      global:
        tenantName: <TENANT_NAME>
        controlPlaneUrl: <CONTROL_PLANE_URL_IN_HTTPS_FORMAT>
      flyte-core:
        storage:
          gcs:
            projectId: <GCP_PROJECT_ID>
          type: gcs
          bucketName: <GCP_GCS_BUCKET_NAME>
        configmap:
          core:
            propeller:
              metadata-prefix: gs://<GCP_GCS_BUCKET_NAME>/tfy-workflow-propeller/metadata
              rawoutput-prefix: gs://<GCP_GCS_BUCKET_NAME>/tfy-workflow-propeller/raw_data
          admin:
            admin:
              Command:
                - echo
                - <TFY_AGENT_TOKEN>
              endpoint: <CONTROL_PLANE_URL>:443
        flytepropeller:
          serviceAccount:
            annotations:
              iam.gke.io/gcp-service-account: <GCP_SERVICE_ACCOUNT_EMAIL>
      tfySignedURLServer:
        env:
          GS_BUCKET_NAME: <GCP_GCS_BUCKET_NAME>
          DEFAULT_CLOUD_PROVIDER: gcp
        enabled: true
      ```
    </CodeGroup>
  </Tab>

  <Tab title="Azure">
    For Azure we need the following values

    1. Azure Blob URI - This is the URI of the Azure Blob Storage used in the integration provider.
    2. Azure Storage Account Name - This is the name of the Azure Storage Account used in the integration provider. You can find this in the azure blob URI.
    3. Azure Storage Account Key - Azure Storage Account Key used in the integration provider. You can find this in the azure connection string.
    4. Azure Storage Container Name - Azure Storage Container Name used in the integration provider. You can find this in the azure blob URI.

    Final values file should like this. Ensure to replace the placeholders with the actual values.

    <CodeGroup>
      ```yaml YAML lines theme={"dark"}
      global:
        tenantName: <TENANT_NAME>
        controlPlaneUrl: <CONTROL_PLANE_URL_IN_HTTPS_FORMAT>
      flyte-core:
        storage:
          type: custom
          custom:
            stow:
              kind: azure
              config:
                key: "<AZURE_STORAGE_ACCOUNT_KEY>"
                account: "<AZURE_STORAGE_ACCOUNT_NAME>"
            type: stow
            container: "<AZURE_STORAGE_CONTAINER_NAME>"
            connection: {}
            enable-multicontainer: true
        configmap:
          k8s:
            plugins:
              k8s:
                default-env-vars:
                  - AZURE_STORAGE_ACCOUNT_NAME: <AZURE_STORAGE_ACCOUNT_NAME>
                  - AZURE_STORAGE_ACCOUNT_KEY: <AZURE_STORAGE_ACCOUNT_KEY>
          core:
            propeller:
              metadata-prefix: abfs://<AZURE_STORAGE_CONTAINER_NAME>/tfy-workflow-propeller/metadata
              rawoutput-prefix: abfs://<AZURE_STORAGE_CONTAINER_NAME>/tfy-workflow-propeller/raw_data
          admin:
            admin:
              Command:
                - echo
                - <TFY_AGENT_TOKEN>
              endpoint: <CONTROL_PLANE_URL>:443
      tfySignedURLServer:
        enabled: false
      ```
    </CodeGroup>
  </Tab>

  <Tab title="Generic/On-prem">
    For Generic/on-prem we need the following values

    1. S3-compatible bucket name - This is the name of the S3 bucket used in the integration provider.
    2. Region - This is the region of the S3-compatible bucket.
    3. Access Key ID - Access Key ID used in the integration provider.
    4. Secret Access Key - Secret Access Key used in the integration provider.

    Final values file should like this. Ensure to replace the placeholders with the actual values.

    <CodeGroup>
      ```yaml YAML lines theme={"dark"}
      global:
        tenantName: <TENANT_NAME>
        controlPlaneUrl: <CONTROL_PLANE_URL_IN_HTTPS_FORMAT>
      flyte-core:
        storage:
          type: s3
          bucketName: <S3_BUCKET_NAME>
          connection:
            region: <REGION>
            auth-type: iam
          enable-multicontainer: true
        configmap:
          core:
            propeller:
              metadata-prefix: s3://<S3_BUCKET_NAME>/tfy-workflow-propeller/metatdata
              rawoutput-prefix: s3://<S3_BUCKET_NAME>/tfy-workflow-propeller/raw_data
          admin:
            admin:
              Command:
                - echo
                - <TFY_AGENT_TOKEN>
              endpoint: <CONTROL_PLANE_URL>:443
        flyteadmin:
          serviceAccount:
            alwaysCreate: true
        flytepropeller:
          podEnv:
            - name: AWS_ACCESS_KEY_ID
              value: <ACCESS_KEY_ID>
            - name: AWS_SECRET_ACCESS_KEY
              value: <SECRET_ACCESS_KEY>
            - name: AWS_REGION
              value: us-east-1
      tfySignedURLServer:
        env:
          AWS_REGION: <REGION>
          S3_BUCKET_NAME: s3://<S3_BUCKET_NAME>
          DEFAULT_CLOUD_PROVIDER: aws
          AWS_ACCESS_KEY_ID: <ACCESS_KEY_ID>
          AWS_SECRET_ACCESS_KEY: <SECRET_ACCESS_KEY>
        enabled: true
      ```
    </CodeGroup>
  </Tab>
</Tabs>

***
