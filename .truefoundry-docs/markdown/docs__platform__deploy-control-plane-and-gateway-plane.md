> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Deploy AI Gateway

> Learn how to deploy TrueFoundry's AI Gateway on your own infrastructure with detailed compute requirements and installation instructions.

## Key Components

To install the complete control plane on your own infrastructure, you need to install the following components:

* Truefoundry Control Plane + Gateway (Shipped as a single helm chart called `truefoundry`)
* PostgreSQL Database (Managed or Self-Hosted with PostgreSQL >= 13)
* Blob Storage (S3, GCS, Azure Container or any other S3 compatible storage)

## Compute Requirements

Truefoundry ships as a helm chart ([https://github.com/truefoundry/infra-charts/tree/main/charts/truefoundry](https://github.com/truefoundry/infra-charts/tree/main/charts/truefoundry)) that has configurable options to either deploy both Deployment and AI Gateway feature or just choose the one of them according to your needs. The compute requirements change based on the set of features and the scale of the number of users and requests.

Here are a few scenarios that you can choose from based on your needs.

<Tabs>
  <Tab title="Small (Dev)">
    The small tier is recommended for development purposes. Here all the components are deployed on Kubernetes and in non HA mode (single replica). This is suitable if you are just testing out the different features of Truefoundry.

    <Warning>
      This setup brings up 1 replica of the services and is not highly-available. It
      can enable you to test the features but we do not recommend this for
      production mode.
    </Warning>

    <Tabs>
      <Tab title="AWS">
        |                         Component                         |    CPU   | Memory |                                  Storage                                  |                           Min Nodes                          |                        Remarks                        |
        | :-------------------------------------------------------: | :------: | :----: | :-----------------------------------------------------------------------: | :----------------------------------------------------------: | :---------------------------------------------------: |
        | Helm-Chart <br /> *(AI Gateway Control Plane components)* |  6 vCPU  |  12GB  | 60GB <br /> <sub>*Persistent Volumes (Block Storage) On Kubernetes*</sub> | 2 <br /> <sub>*Pods should be spread over min 2 nodes*</sub> |   <sub>Cost: \~ \$220 pm(EC2 and EC2 others) </sub>   |
        |      Helm-Chart <br /> *(AI Gateway component only)*      |  1 vCPU  |  512Mi |                                     -                                     |  1 <br /> <sub>*Pods should be spread over min 1 node*</sub> |    <sub>Cost: \~ \$35 pm(EC2 and EC2 others) </sub>   |
        |          Postgres <br /> (Deployed on Kubernetes)         | 0.5 vCPU |  0.5GB |  5GB <br /> <sub>*Persistent Volumes (Block Storage) On Kubernetes*</sub> |                                                              | <sub>Cost: \~ \$15 pm (RDS compute and storage)</sub> |
        |            Blob Storage <br /> (S3 Compatible)            |          |        |                                    20GB                                   |                                                              |        <sub>Cost: \~ \$3 pm (S3 storage)</sub>        |
      </Tab>

      <Tab title="GCP">
        |                         Component                         |    CPU   | Memory |                                  Storage                                  |                           Min Nodes                          |            Remarks            |
        | :-------------------------------------------------------: | :------: | :----: | :-----------------------------------------------------------------------: | :----------------------------------------------------------: | :---------------------------: |
        | Helm-Chart <br /> *(AI Gateway Control Plane components)* |  6 vCPU  |  12GB  | 60GB <br /> <sub>*Persistent Volumes (Block Storage) On Kubernetes*</sub> | 2 <br /> <sub>*Pods should be spread over min 2 nodes*</sub> | <sub>Cost: \~ Coming up</sub> |
        |      Helm-Chart <br /> *(AI Gateway component only)*      |  1 vCPU  |  512Mi |                                     -                                     |  1 <br /> <sub>*Pods should be spread over min 1 node*</sub> | <sub>Cost: \~ Coming up</sub> |
        |          Postgres <br /> (Deployed on Kubernetes)         | 0.5 vCPU |  0.5GB |  5GB <br /> <sub>*Persistent Volumes (Block Storage) On Kubernetes*</sub> |                                                              | <sub>Cost: \~ Coming up</sub> |
        |            Blob Storage <br /> (S3 Compatible)            |          |        |                                    20GB                                   |                                                              | <sub>Cost: \~ Coming up</sub> |
      </Tab>

      <Tab title="Azure">
        |                         Component                         |    CPU   | Memory |                                  Storage                                  |                           Min Nodes                          |            Remarks            |
        | :-------------------------------------------------------: | :------: | :----: | :-----------------------------------------------------------------------: | :----------------------------------------------------------: | :---------------------------: |
        | Helm-Chart <br /> *(AI Gateway Control Plane components)* |  6 vCPU  |  12GB  | 60GB <br /> <sub>*Persistent Volumes (Block Storage) On Kubernetes*</sub> | 2 <br /> <sub>*Pods should be spread over min 2 nodes*</sub> | <sub>Cost: \~ Coming up</sub> |
        |      Helm-Chart <br /> *(AI Gateway component only)*      |  1 vCPU  |  512Mi |                                     -                                     |  1 <br /> <sub>*Pods should be spread over min 1 node*</sub> | <sub>Cost: \~ Coming up</sub> |
        |          Postgres <br /> (Deployed on Kubernetes)         | 0.5 vCPU |  0.5GB |  5GB <br /> <sub>*Persistent Volumes (Block Storage) On Kubernetes*</sub> |                                                              | <sub>Cost: \~ Coming up</sub> |
        |            Blob Storage <br /> (S3 Compatible)            |          |        |                                    20GB                                   |                                                              | <sub>Cost: \~ Coming up</sub> |
      </Tab>
    </Tabs>
  </Tab>

  <Tab title="Medium (Prod)">
    The medium tier is configured for production and will suffice teams of 10-500 members.

    <Note>
      **The AI Gateway is configured for a minimum 3 replicas (1 vCPU 1GB each) which can handle around 500 requests/second to LLMs.**

      It's configured to be horizontally scalable and autoscale when the load increases. The Block Storage and S3 are used to store LLM request logs. The size is dependent on the size and number of requests and should be set as per the expected usage.
    </Note>

    <Tabs>
      <Tab title="AWS">
        |                         Component                         |   CPU   | Memory | Storage |                           Min Nodes                          |                        Remarks                        |
        | :-------------------------------------------------------: | :-----: | :----: | :-----: | :----------------------------------------------------------: | :---------------------------------------------------: |
        | Helm-Chart <br /> *(AI Gateway Control Plane components)* | 14 vCPU |  32GB  |  250GB  | 3 <br /> <sub>*Pods should be spread over min 3 nodes*</sub> |   <sub>Cost: \~ \$600 pm (EC2 and EC2 others)</sub>   |
        |      Helm-Chart <br /> *(AI Gateway component only)*      |  3 vCPU |   3GB  |    -    | 3 <br /> <sub>*Pods should be spread over min 3 nodes*</sub> |   <sub>Cost: \~ \$105 pm (EC2 and EC2 others)</sub>   |
        |             Postgres <br /> (Managed Database)            |  2 vCPU |   4GB  |   30GB  |                                                              | <sub>Cost: \~ \$60 pm (RDS compute and storage)</sub> |
        |            Blob Storage <br /> (S3 Compatible)            |         |        |  500GB  |                                                              |        <sub>Cost: \~ \$15 pm (S3 storage)</sub>       |
      </Tab>

      <Tab title="GCP">
        |                         Component                         |   CPU   | Memory | Storage |                           Min Nodes                          |            Remarks            |
        | :-------------------------------------------------------: | :-----: | :----: | :-----: | :----------------------------------------------------------: | :---------------------------: |
        | Helm-Chart <br /> *(AI Gateway Control Plane components)* | 14 vCPU |  32GB  |  250GB  | 3 <br /> <sub>*Pods should be spread over min 3 nodes*</sub> | <sub>Cost: \~ Coming up</sub> |
        |      Helm-Chart <br /> *(AI Gateway component only)*      |  3 vCPU |   3GB  |    -    | 3 <br /> <sub>*Pods should be spread over min 3 nodes*</sub> | <sub>Cost: \~ Coming up</sub> |
        |             Postgres <br /> (Managed Database)            |  2 vCPU |   4GB  |   30GB  |                                                              | <sub>Cost: \~ Coming up</sub> |
        |            Blob Storage <br /> (S3 Compatible)            |         |        |  500GB  |                                                              | <sub>Cost: \~ Coming up</sub> |
      </Tab>

      <Tab title="Azure">
        |                         Component                         |   CPU   | Memory | Storage |                           Min Nodes                          |            Remarks            |
        | :-------------------------------------------------------: | :-----: | :----: | :-----: | :----------------------------------------------------------: | :---------------------------: |
        | Helm-Chart <br /> *(AI Gateway Control Plane components)* | 14 vCPU |  32GB  |  250GB  | 3 <br /> <sub>*Pods should be spread over min 3 nodes*</sub> | <sub>Cost: \~ Coming up</sub> |
        |      Helm-Chart <br /> *(AI Gateway component only)*      |  3 vCPU |   3GB  |    -    | 3 <br /> <sub>*Pods should be spread over min 3 nodes*</sub> | <sub>Cost: \~ Coming up</sub> |
        |             Postgres <br /> (Managed Database)            |  2 vCPU |   4GB  |   30GB  |                                                              | <sub>Cost: \~ Coming up</sub> |
        |            Blob Storage <br /> (S3 Compatible)            |         |        |  500GB  |                                                              | <sub>Cost: \~ Coming up</sub> |
      </Tab>
    </Tabs>
  </Tab>

  <Tab title="Large (Prod)">
    The large tier is configured for production and will suffice organizations of 500-50000 members.

    <Note>
      **The AI Gateway is configured for a minimum 10 replicas (1 vCPU 1GB each) which can handle around 2000 requests/second to LLMs.**

      It's configured to be horizontally scalable and autoscale when the load increases. The Block Storage and S3 are used to store LLM request logs. The size is dependent on the size and number of requests and should be set as per the expected usage.
    </Note>

    <Tabs>
      <Tab title="AWS">
        |                         Component                         |   CPU   | Memory | Storage |                            Min Nodes                           |                        Remarks                        |
        | :-------------------------------------------------------: | :-----: | :----: | :-----: | :------------------------------------------------------------: | :---------------------------------------------------: |
        | Helm-Chart <br /> *(AI Gateway Control Plane components)* | 32 vCPU |  64GB  |  400GB  | 10 <br /> <sub>*Pods should be spread over min 10 nodes*</sub> |   <sub>Cost: \~ \$1400 pm (EC2 and EC2 others)</sub>  |
        |      Helm-Chart <br /> *(AI Gateway component only)*      | 10 vCPU |  10GB  |    -    | 10 <br /> <sub>*Pods should be spread over min 10 nodes*</sub> |   <sub>Cost: \~ \$350 pm (EC2 and EC2 others)</sub>   |
        |             Postgres <br /> (Managed Database)            |  2 vCPU |   4GB  |   30GB  |                                                                | <sub>Cost: \~ \$60 pm (RDS compute and storage)</sub> |
        |            Blob Storage <br /> (S3 Compatible)            |         |        |  1000GB |                                                                |        <sub>Cost: \~ \$30 pm (S3 storage)</sub>       |
      </Tab>

      <Tab title="GCP">
        |                         Component                         |   CPU   | Memory | Storage |                            Min Nodes                           |            Remarks            |
        | :-------------------------------------------------------: | :-----: | :----: | :-----: | :------------------------------------------------------------: | :---------------------------: |
        | Helm-Chart <br /> *(AI Gateway Control Plane components)* | 32 vCPU |  64GB  |  400GB  | 10 <br /> <sub>*Pods should be spread over min 10 nodes*</sub> | <sub>Cost: \~ Coming up</sub> |
        |      Helm-Chart <br /> *(AI Gateway component only)*      | 10 vCPU |  10GB  |    -    | 10 <br /> <sub>*Pods should be spread over min 10 nodes*</sub> | <sub>Cost: \~ Coming up</sub> |
        |             Postgres <br /> (Managed Database)            |  2 vCPU |   4GB  |   30GB  |                                                                | <sub>Cost: \~ Coming up</sub> |
        |            Blob Storage <br /> (S3 Compatible)            |         |        |  1000GB |                                                                | <sub>Cost: \~ Coming up</sub> |
      </Tab>

      <Tab title="Azure">
        |                         Component                         |   CPU   | Memory | Storage |                            Min Nodes                           |            Remarks            |
        | :-------------------------------------------------------: | :-----: | :----: | :-----: | :------------------------------------------------------------: | :---------------------------: |
        | Helm-Chart <br /> *(AI Gateway Control Plane components)* | 32 vCPU |  64GB  |  400GB  | 10 <br /> <sub>*Pods should be spread over min 10 nodes*</sub> | <sub>Cost: \~ Coming up</sub> |
        |      Helm-Chart <br /> *(AI Gateway component only)*      | 10 vCPU |  10GB  |    -    | 10 <br /> <sub>*Pods should be spread over min 10 nodes*</sub> | <sub>Cost: \~ Coming up</sub> |
        |             Postgres <br /> (Managed Database)            |  2 vCPU |   4GB  |   30GB  |                                                                | <sub>Cost: \~ Coming up</sub> |
        |            Blob Storage <br /> (S3 Compatible)            |         |        |  1000GB |                                                                | <sub>Cost: \~ Coming up</sub> |
      </Tab>
    </Tabs>
  </Tab>
</Tabs>

## Prerequisites for Installation

1. **Kubernetes Cluster**: K8s cluster 1.27+.
2. Support for dynamic provisioning of storage for PVC (for e.g AWS EBS, Azure Disk etc.) and support for ingress controller (for e.g. Nginx Ingress Controller) or istio service mesh for exposing the control plane dashboard and AI Gateway at an endpoint.
3. Domain to map the ingress of the Control Plane dashboard and AI Gateway along with certificate for the domain.
   <Info>
     This Domain will be referred as Control Plane URL in our documentation.
   </Info>
4. Egress Access from TrueFoundry:
   * [https://auth.truefoundry.com](https://auth.truefoundry.com) - Central Authentication Server for licensing and authentication
   * [https://login.truefoundry.com](https://login.truefoundry.com) - Login UI for the central authentication server
   * [https://catalogue.truefoundry.com](https://catalogue.truefoundry.com) - Central Repository for fetching catalogues for latest model, their public cost, mcp servers, etc.
   * [https://analytics.truefoundry.com](https://analytics.truefoundry.com) - Analytics Server for sending usage analytics to TrueFoundry.
5. **Tenant Name, Licence key, and image pull secret** - This will be given by TrueFoundry team. Make sure your organization is registered(([https://truefoundry.com/register](https://truefoundry.com/register))) on TrueFoundry.
   <Note>
     One Tenant Name and Licence key must only be used to setup one Control
     Plane. Later, switching to new tenant name and licence key would lead to
     complete data lose of existing control plane.
   </Note>
6. PostgreSQL database. We usually recommend managed PostgreSQL database (For e.g. AWS RDS, or Google Cloud SQL, or Azure Database for PostgreSQL) for production environments.
   * PostgreSQL version >= 13
   * IOPS: Default (suitable for dev/testing).
   * For PostgreSQL 17+: Disable SSL, for AWS: by setting `force_ssl` parameter to 0 in the parameter group, for Azure: by setting `require_secure_transport` parameter to `false` in the parameter group
   * For instance requirements, refer to the [Compute Requirements](#compute-requirements) section.
     <Note>
       In case, you do not have a managed database just for testing purposes,
       set `devMode` to true in the values file to spin up a local PostgreSQL
       database.
     </Note>
7. Blob Storage to store the AI Gateway request logs (either S3, GCS, Azure Blob Storage, or any other S3 compatible storage). You can find the instructions in the guide below.

## Installation Instructions

<Tabs>
  <Tab title="AWS">
    <Steps>
      <Step title="Create S3 Bucket">
        Create a [S3 Bucket](https://docs.aws.amazon.com/AmazonS3/latest/userguide/creating-bucket.html) with following config:

        * Make sure the bucket has lifecycle configuration to [abort multipart upload](https://docs.aws.amazon.com/AmazonS3/latest/userguide/mpu-abort-incomplete-mpu-lifecycle-config.html) set for 7 days.
        * Make sure [CORS](https://docs.aws.amazon.com/AmazonS3/latest/userguide/enabling-cors-examples.html) is applied on the bucket with the below configuration:

        ```json lines theme={"dark"}
        [
          {
            "AllowedHeaders": ["*"],
            "AllowedMethods": ["GET", "POST", "PUT"],
            "AllowedOrigins": ["*"],
            "ExposeHeaders": ["ETag"],
            "MaxAgeSeconds": 3000
          }
        ]
        ```
      </Step>

      <Step title="Setup Control Plane Platform IAM Role">
        <Accordion title="Creating AWS IAM Role for Control Plane">
          Control Plane IAM Role needs to have permission to access the S3 bucket created in the previous step.

          * Create a new IAM role for Control Plane with a suitable name like `tfy-control-plane-platform-deps`
          * Add the following trust policy to the Control Plane IAM Role:

          ```json lines theme={"dark"}
          {
            "Version": "2012-10-17",
            "Statement": [
                {
                    "Effect": "Allow",
                    "Principal": {
                        "Federated": "arn:aws:iam::<ACCOUNT_ID>:oidc-provider/oidc.eks.<AWS_REGION>.amazonaws.com/id/<OIDC_ID>"
                    },
                    "Action": "sts:AssumeRoleWithWebIdentity",
                    "Condition": {
                        "StringEquals": {
                            "oidc.eks.<AWS_REGION>.amazonaws.com/id/<OIDC_ID>:sub": [
                                "system:serviceaccount:truefoundry:truefoundry",
                            ]
                        }
                    }
                }
            ]
          }
          ```

          <Note>
            In place of `<ACCOUNT_ID>`, `<AWS_REGION>`, and `<OIDC_ID>` you can also give the values from your EKS cluster.
            You can find the OIDC\_ID from the EKS cluster. Also, here we are assuming that the service account is `truefoundry` and the namespace is `truefoundry`, you can change it as per your needs.
          </Note>

          Create a IAM Policy to allow access to the S3 Bucket with following config:

          ```json lines theme={"dark"}
          {
            "Statement": [
              {
                "Sid": "S3",
                "Effect": "Allow",
                "Action": ["s3:*"],
                "Resource": [
                  "arn:aws:s3:::<YOUR_S3_BUCKET_NAME>",
                  "arn:aws:s3:::<YOUR_S3_BUCKET_NAME>/*"
                ]
              }
            ],
            "Version": "2012-10-17"
          }
          ```

          Attach the IAM Policy to the Control Plane Platform IAM Role. You can also attach the IAM policy to access AWS bedrock models from the link [here](/docs/ai-gateway/aws-bedrock#get-aws-authentication-details).

          <Note>
            If you are integrating with AWS bedrock models from a different AWS account, you can check the [FAQ section](/docs/ai-gateway/install-only-llm-gateway#how-to-integrate-with-aws-bedrock-models-from-a-different-aws-account).
          </Note>
        </Accordion>
      </Step>

      <Step title="Create Postgres RDS Database">
        Create a Postgres RDS instance of size `db.t3.medium` with storage size of 30GB.

        <Warning>
          **Important Configuration Notes:** - **For PostgreSQL 17:** Disable SSL by
          setting `force_ssl` parameter to `0` in the parameter group - **Security
          Group:** Ensure your RDS security group has inbound rules allowing traffic
          from EKS node security groups
        </Warning>

        <Note>
          In case you want to setup PostgreSQL on Kubernetes and not use RDS for testing purposes, skip this step and set `devMode` to true in the values file below
        </Note>
      </Step>

      <Step title="Create Kubernetes Secrets">
        We will create two secrets in this step:

        1. Store the License Key and DB Credentials
        2. Store the Image Pull Secret

        <Accordion title="Create Kubernetes Secret for License Key and DB Credentials">
          We need to create a [Kubernetes secret](https://github.com/truefoundry/infra-charts/blob/main/charts/truefoundry/README.md#using-k8s-secret-for-required-fields) containing the licence key and db credentials.

          <Note>
            If you are using PostgreSQL on Kubernetes in the dev mode, the values will be as follows:

            DB\_HOST: \<HELM\_RELEASE\_NAME>-postgresql.\<NAMESPACE>.svc.cluster.local // eg. truefoundry-postgresql.truefoundry.svc.cluster.local

            DB\_NAME: truefoundry

            DB\_USERNAME: postgres # In order to use custom username, please update the same at `postgresql.auth.username`

            DB\_PASSWORD: randompassword # You can change this to any value here.
          </Note>

          ```yaml truefoundry-creds.yaml lines theme={"dark"}
          apiVersion: v1
          kind: Secret
          metadata:
            name: truefoundry-creds
          type: Opaque
          stringData:
            TFY_API_KEY: <TFY_API_KEY> # Provided by TrueFoundry team
            DB_HOST: <DB_HOST>
            DB_NAME: <DB_NAME>
            DB_USERNAME: <DB_USERNAME>
            DB_PASSWORD: <DB_PASSWORD>
          ```

          Apply the secret to the Kubernetes cluster (Assuming you are installing the control plane in the `truefoundry` namespace)

          ```bash lines theme={"dark"}
          kubectl apply -f truefoundry-creds.yaml -n truefoundry
          ```
        </Accordion>

        <Accordion title="Create Kubernetes Secret for Image Pull Secret">
          We need to create a [Image Pull Secret](https://github.com/truefoundry/infra-charts/blob/main/charts/truefoundry/README.md#using-k8s-secret-for-required-fields) to enable pulling the truefoundry images from the private registry.

          ```yaml truefoundry-image-pull-secret.yaml lines theme={"dark"}
          apiVersion: v1
          kind: Secret
          metadata:
            name: truefoundry-image-pull-secret
          type: kubernetes.io/dockerconfigjson
          data:
            .dockerconfigjson: <IMAGE_PULL_SECRET> # Provided by TrueFoundry team
          ```

          Apply the secret to the Kubernetes cluster (Assuming you are installing the control plane in the `truefoundry` namespace)

          ```bash lines theme={"dark"}
          kubectl apply -f truefoundry-image-pull-secret.yaml -n truefoundry
          ```
        </Accordion>
      </Step>

      <Step title="Create HelmChart Values file">
        Create a values file as given below and replace the following values:

        * **Control Plane URL**: URL that you will map to the control plane dashboard (e.g., `https://truefoundry.example.com`)
        * **Tenant Name**: Tenant name provided by TrueFoundry team
        * **AWS S3 Bucket Name**: Name of the S3 bucket you created in the previous step (e.g., `my-truefoundry-bucket`)
        * **AWS Region**: Region of the S3 bucket you created in the previous step (e.g., `us-west-2`)
        * **Control Plane IAM Role ARN**: ARN of the IAM role you created in the previous step (e.g., `arn:aws:iam::123456789012:role/tfy-control-plane-platform-deps`)

        ```yaml truefoundry-values.yaml wrap lines theme={"dark"}
        global:
          # Domain to map the platform to
          controlPlaneURL: https://example.com

          # Ask TrueFoundry team to provide these
          tenantName: <TENANT_NAME>

          # Choose the resource tier as per your needs
          resourceTier: medium # or small or large

          # This is the reference to the secrets we created in the previous step
          existingTruefoundryCredsSecret: "truefoundry-creds"
          imagePullSecrets:
            - name: "truefoundry-image-pull-secret"
          ## Add if you have restricted public registry access
          # image:
          #   pullSecretNames:
          #   - "truefoundry-image-pull-secret"

          config:
            defaultCloudProvider: "aws"
            storageConfiguration:
              awsS3BucketName: "<AWS_S3_BUCKET_NAME>"
              awsRegion: "<AWS_REGION>"

          serviceAccount:
            annotations:
              eks.amazonaws.com/role-arn: <CONTROL_PLANE_IAM_ROLE_ARN>

          ingress:
            hosts:
              - example.com
            enabled: true
            annotations: {}
            ingressClassName: nginx # Replace with your ingress class name

        # In case, you want to spin up PostgreSQL on kubernetes, enable this
        # Please add creds and host details in the secret `truefoundry-creds` in the previous step
        devMode:
          enabled: false
        tags:
          llmGateway: true
          llmGatewayRequestLogging: true

        # Disable few dependencies for only LLM Gateway setup
        tfyBuild:
          enabled: false
        sfyManifestService:
          enabled: false
        tfyController:
          enabled: false
        tfyConfigs:
          enabled: false
        ```
      </Step>

      <Step title="Install Helm chart">
        ```bash wrap lines theme={"dark"}
        helm upgrade --install truefoundry oci://tfy.jfrog.io/tfy-helm/truefoundry -n truefoundry --create-namespace -f truefoundry-values.yaml
        ```
      </Step>
    </Steps>
  </Tab>

  <Tab title="GCP">
    <Steps>
      <Step title="Create GCP Storage Bucket">
        We will need a GCS bucket to store the LLM request logs.

        <Accordion title="Create GCS bucket">
          To setup GCP Storage for control plane, follow the steps below:

          1. Create a [GCP bucket](https://cloud.google.com/storage/docs/creating-buckets).
             * Make sure to add the lifecycle configurations on the bucket to delete multipart upload after 7 days.
             * For this go to GCP bucket -> Lifecycle -> Add a rule
             * Select `Delete multi-part upload` for 7 days

          2. We also need to add the CORS policy to the GCP bucket. Right now adding the CORS policy to the GCP bucket is not possible through the console so for this, we will use *gsutil*

             1. Create a file called `cors.json` using the below command

                       <CodeGroup>
                         ```powershell Shell lines theme={"dark"}
                         cat > cors.json <<EOF
                         [
                             {
                               "origin": ["*"],
                               "method": ["GET", "POST", "PUT"],
                               "maxAgeSeconds": 3600
                             }
                         ]
                         EOF
                         ```
                       </CodeGroup>

          3. Attach the above CORS policy to the service account by running the following command using [gsutils](https://cloud.google.com/storage/docs/gsutil_install)

                       <CodeGroup>
                         ```shell Shell lines theme={"dark"}
                         gsutil cors set cors.json  gs://BUCKET_NAME
                         ```
                       </CodeGroup>

          4. Create a [custom IAM role](https://cloud.google.com/iam/docs/creating-custom-roles#creating) with the following permissions and add to the `serviceaccount` created above:

                       <CodeGroup>
                         ```javascript JSON lines theme={"dark"}
                         [
                           "storage.objects.create",
                           "storage.objects.delete",
                           "storage.objects.get",
                           "storage.objects.list",
                           "storage.objects.update",
                           "storage.buckets.create",
                           "storage.buckets.get",
                           "storage.buckets.list",
                           "storage.buckets.create",
                           "storage.buckets.update",
                           "storage.multipartUploads.create",
                           "storage.multipartUploads.list",
                           "storage.multipartUploads.listParts",
                           "storage.multipartUploads.abort",
                           "resourcemanager.projects.get",
                         ];
                         ```
                       </CodeGroup>

          5. Add the following [IAM condition](https://cloud.google.com/iam/docs/conditions-overview) - `resource.name.startsWith('projects/\_/buckets/<bucket name>}')` to the above IAM role.
        </Accordion>
      </Step>

      <Step title="Create Google Cloud SQL for PostgreSQL database">
        Create a Google Cloud SQL for PostgreSQL database of size `db-custom-1-3840` with storage size of 30GB.

        <Warning>
          **Important Configuration Notes:** - **For PostgreSQL 17+:** Disable SSL
          Security > Allow unencrypted network traffic - **Firewall Rules:** Ensure your
          database firewall rules allow inbound traffic from GKE node pools
        </Warning>

        <Note>
          In case you want to setup PostgreSQL on Kubernetes in the dev mode, skip this step and set `devMode` to true in the values file in the steps below.
        </Note>
      </Step>

      <Step title="Create Kubernetes Secrets">
        We will create two secrets in this step:

        1. Store the License Key and DB Credentials
        2. Store the Image Pull Secret

        <Accordion title="Create Kubernetes Secret for License Key and DB Credentials">
          We need to create a [Kubernetes secret](https://github.com/truefoundry/infra-charts/blob/main/charts/truefoundry/README.md#using-k8s-secret-for-required-fields) containing the licence key and db credentials.

          <Note>
            If you are using PostgreSQL on Kubernetes in the dev mode, the values will be as follows:

            DB\_HOST: \<HELM\_RELEASE\_NAME>-postgresql.\<NAMESPACE>.svc.cluster.local // eg. truefoundry-postgresql.truefoundry.svc.cluster.local

            DB\_NAME: truefoundry

            DB\_USERNAME: postgres # In order to use custom username, please update the same at `postgresql.auth.username`

            DB\_PASSWORD: randompassword # You can change this to any value here.
          </Note>

          ```yaml truefoundry-creds.yaml lines theme={"dark"}
          apiVersion: v1
          kind: Secret
          metadata:
            name: truefoundry-creds
          type: Opaque
          stringData:
            TFY_API_KEY: <TFY_API_KEY> # Provided by TrueFoundry team
            DB_HOST: <DB_HOST>
            DB_NAME: <DB_NAME>
            DB_USERNAME: <DB_USERNAME>
            DB_PASSWORD: <DB_PASSWORD>
          ```

          Apply the secret to the Kubernetes cluster (Assuming you are installing the control plane in the `truefoundry` namespace)

          ```bash lines theme={"dark"}
          kubectl apply -f truefoundry-creds.yaml -n truefoundry
          ```
        </Accordion>

        <Accordion title="Create Kubernetes Secret for Image Pull Secret">
          We need to create a [Image Pull Secret](https://github.com/truefoundry/infra-charts/blob/main/charts/truefoundry/README.md#using-k8s-secret-for-required-fields) to enable pulling the truefoundry images from the private registry.

          ```yaml truefoundry-image-pull-secret.yaml lines theme={"dark"}
          apiVersion: v1
          kind: Secret
          metadata:
            name: truefoundry-image-pull-secret
          type: kubernetes.io/dockerconfigjson
          data:
            .dockerconfigjson: <IMAGE_PULL_SECRET> # Provided by TrueFoundry team
          ```

          Apply the secret to the Kubernetes cluster (Assuming you are installing the control plane in the `truefoundry` namespace)

          ```bash lines theme={"dark"}
          kubectl apply -f truefoundry-image-pull-secret.yaml -n truefoundry
          ```
        </Accordion>
      </Step>

      <Step title="Create Helm chart Values file">
        Create a values file as given below and replace the following values:

        * Control Plane URL: URL that you will map to the control plane dashboard.
        * Tenant Name: Tenant name provided by TrueFoundry team.
        * GCP Project ID: GCP project ID provided by TrueFoundry team.
        * GCP Storage Bucket Name: Name of the GCS bucket you created in the previous step.
        * GCP Service Account Name: Name of the GCP service account you created in the previous step.

        ```yaml truefoundry-values.yaml wrap expandable lines theme={"dark"}
        global:
          # Domain to map the control plane dashboard
          controlPlaneURL: https://example.com

          # Ask TrueFoundry team to provide these
          tenantName: <TENANT_NAME>

          # Choose the resource tier as per your needs
          resourceTier: medium # or small or large

          # This is the reference to the secrets we created in the previous step
          existingTruefoundryCredsSecret: "truefoundry-creds"
          imagePullSecrets:
            - name: "truefoundry-image-pull-secret"
          ## Add if you have restricted public registry access
          # image:
          #   pullSecretNames:
          #   - "truefoundry-image-pull-secret"

          config:
            defaultCloudProvider: "gcp"
            storageConfiguration:
              googleCloudProjectId: "<GCP_PROJECT_ID>"
              googleCloudStorageBucketName: "<GCP_STORAGE_BUCKET_NAME>"

          serviceAccount:
            annotations:
              iam.gke.io/gcp-service-account: <CONTROL_PLANE_GCP_SERVICE_ACCOUNT_NAME>@<PROJECT_ID>.iam.gserviceaccount.com

          ingress:
            hosts:
              - example.com
            enabled: true
            annotations: {}
            ingressClassName: nginx # Replace with your ingress class name

        # In case, you want to spin up PostgreSQL on kubernetes, enable this
        devMode:
          enabled: false
        tags:
          llmGateway: true
          llmGatewayRequestLogging: true

        # Disable few dependencies for only LLM Gateway setup
        tfyBuild:
          enabled: false
        sfyManifestService:
          enabled: false
        tfyController:
          enabled: false
        tfyConfigs:
          enabled: false
        ```
      </Step>

      <Step title="Install Helm chart">
        ```bash wrap lines theme={"dark"}
        helm upgrade --install truefoundry oci://tfy.jfrog.io/tfy-helm/truefoundry -n truefoundry --create-namespace -f truefoundry-values.yaml
        ```
      </Step>
    </Steps>
  </Tab>

  <Tab title="Azure">
    <Steps>
      <Step title="Create Azure Blob Storage">
        <Accordion title="Create Azure Blob Storage">
          To setup Azure Blob Storage for control plane, follow the steps below:

          1. Create a [Azure Storage account](https://learn.microsoft.com/en-us/azure/storage/common/storage-account-create?tabs=azure-portal) in your resource group
             1. **Instance details** - You must `Geo-redundant storage` to make sure your data is available through other regions in case of region unavailability.

             2. **Security** - Make sure
                1. DISABLE `Allow enabling anonymous access on individual containers`
                2. ENABLE `Enable storage account key access`

             3. **Network access** - ENABLE `Allow public access from all networks`

             4. **Recovery** - You can keep it to default for 7 days.

          2. Create an [Azure container](https://learn.microsoft.com/en-us/azure/storage/blobs/blob-containers-portal) inside the above storage account.

          3. Search for `CORS` from the left panel and for `Blob service` (optional for `File service` `Queue service` and `Table Service`, only apply the change if you are using them) select the below options
             1. Allowed Origins - `*` or your control plane URL
             2. Allowed Methods - `GET, POST, PUT`
             3. Allowed Headers - `*`
             4. Exposed Headers - `Etag`
             5. MaxAgeSeconds - `3600`

          4. Collect the following information
             1. [Standard endpoint](https://learn.microsoft.com/en-us/azure/storage/common/storage-account-overview#standard-endpoints) - Endpoint of the blob storage Once the container is created we need to get the standard endpoint of the blob storage along with the container which will look something like this. Replace this with your storage account name and the container name.
                ```javascript lines theme={"dark"}
                https://*mystorageaccount*.blob.core.windows.net/*mycontainer*/
                ```
             2. Connection string - From the Azure portal in your storage account, head over to the `Security + Networking` section under `Access keys` which will contain the `Connection String` .
        </Accordion>
      </Step>

      <Step title="Create Azure Database for PostgreSQL database">
        Create a PostgreSQL database of size `GP_Standard_D4ds_v5` with storage size of 30GB.

        <Warning>
          **Important Configuration Notes:** - **For PostgreSQL 17+:** Disable SSL,
          allowing both encrypted and unencrypted client communications, you can change
          the server parameter `require_secure_transport` to OFF - **Security Group:**
          Ensure your database firewall rules allow inbound traffic from AKS node pools
        </Warning>

        <Note>
          In case you want to setup PostgreSQL on Kubernetes in the dev mode, skip this step and set `devMode` to true in the values file in the steps below.
        </Note>
      </Step>

      <Step title="Create Kubernetes Secrets">
        We will create two secrets in this step:

        1. Store the License Key and DB Credentials
        2. Store the Image Pull Secret

        <Accordion title="Create Kubernetes Secret for License Key and DB Credentials">
          We need to create a [Kubernetes secret](https://github.com/truefoundry/infra-charts/blob/main/charts/truefoundry/README.md#using-k8s-secret-for-required-fields) containing the licence key and db credentials.

          <Note>
            If you are using PostgreSQL on Kubernetes in the dev mode, the values will be as follows:

            DB\_HOST: \<HELM\_RELEASE\_NAME>-postgresql.\<NAMESPACE>.svc.cluster.local // eg. truefoundry-postgresql.truefoundry.svc.cluster.local

            DB\_NAME: truefoundry

            DB\_USERNAME: postgres # In order to use custom username, please update the same at `postgresql.auth.username`

            DB\_PASSWORD: randompassword # You can change this to any value here.
          </Note>

          ```yaml truefoundry-creds.yaml lines theme={"dark"}
          apiVersion: v1
          kind: Secret
          metadata:
            name: truefoundry-creds
          type: Opaque
          stringData:
            TFY_API_KEY: <TFY_API_KEY> # Provided by TrueFoundry team
            DB_HOST: <DB_HOST>
            DB_NAME: <DB_NAME>
            DB_USERNAME: <DB_USERNAME>
            DB_PASSWORD: <DB_PASSWORD>
          ```

          Apply the secret to the Kubernetes cluster (Assuming you are installing the control plane in the `truefoundry` namespace)

          ```bash lines theme={"dark"}
          kubectl apply -f truefoundry-creds.yaml -n truefoundry
          ```
        </Accordion>

        <Accordion title="Create Kubernetes Secret for Image Pull Secret">
          We need to create a [Image Pull Secret](https://github.com/truefoundry/infra-charts/blob/main/charts/truefoundry/README.md#using-k8s-secret-for-required-fields) to enable pulling the truefoundry images from the private registry.

          ```yaml truefoundry-image-pull-secret.yaml lines theme={"dark"}
          apiVersion: v1
          kind: Secret
          metadata:
            name: truefoundry-image-pull-secret
          type: kubernetes.io/dockerconfigjson
          data:
            .dockerconfigjson: <IMAGE_PULL_SECRET> # Provided by TrueFoundry team
          ```

          Apply the secret to the Kubernetes cluster (Assuming you are installing the control plane in the `truefoundry` namespace)

          ```bash lines theme={"dark"}
          kubectl apply -f truefoundry-image-pull-secret.yaml -n truefoundry
          ```
        </Accordion>
      </Step>

      <Step title="Create HelmChart Values file">
        Create a values file as given below and replace the following values:

        * `Control Plane URL`: URL that you will map to the control plane dashboard.
        * `Tenant Name`: Tenant name provided by TrueFoundry team.
        * `Azure Blob URI`: Azure Blob URI provided by TrueFoundry team.
        * `Azure Blob Connection String`: Azure Blob Connection String provided by TrueFoundry team.

        ```yaml truefoundry-values.yaml wrap expandable lines theme={"dark"}
        global:
          # Domain to map the platform to
          controlPlaneURL: https://example.com

          # Ask TrueFoundry team to provide these
          tenantName: <TENANT_NAME>

          # This is the reference to the secrets we created in the previous step
          existingTruefoundryCredsSecret: "truefoundry-creds"
          imagePullSecrets:
            - name: "truefoundry-image-pull-secret"
          ## Add if you have restricted public registry access
          # image:
          #   pullSecretNames:
          #   - "truefoundry-image-pull-secret"

          # Choose the resource tier as per your needs
          resourceTier: medium # or small or large

          config:
            defaultCloudProvider: "azure"
            storageConfiguration:
              azureBlobUri: "<AZURE_BLOB_URI>"
              # Add the connection string as part of truefoundry-creds secret(refer - https://github.com/truefoundry/infra-charts/blob/main/charts/truefoundry/README.md#using-k8s-secret-for-additional-fields):
              azureBlobConnectionString: ${k8s-secret/truefoundry-creds/AZURE_BLOB_CONNECTION_STRING}
          ingress:
            hosts:
              - example.com
            enabled: true
            annotations: {}
            ingressClassName: nginx # Replace with your ingress class name
        # In case, you want to spin up PostgreSQL on kubernetes, enable this
        # Please add creds and host details in the secret `truefoundry-creds`
        devMode:
          enabled: false
        tags:
          llmGateway: true
          llmGatewayRequestLogging: true

        tfyBuild:
          enabled: false
        sfyManifestService:
          enabled: false
        tfyController:
          enabled: false
        tfyConfigs:
          enabled: false
        ```
      </Step>

      <Step title="Install Helm chart">
        ```bash wrap lines theme={"dark"}
        helm upgrade --install truefoundry oci://tfy.jfrog.io/tfy-helm/truefoundry -n truefoundry --create-namespace -f truefoundry-values.yaml
        ```
      </Step>
    </Steps>
  </Tab>

  <Tab title="Openshift">
    In OpenShift clusters, the values file will be slightly different. The rest of the steps remains the exact same as what you would do for AWS/GCP/Azure.

    The values file will be as follows:

    ```yaml truefoundry-values.yaml wrap lines theme={"dark"}
    global:
      # Domain to map the control plane dashboard
      controlPlaneURL: https://example.com

      # Ask TrueFoundry team to provide these
      tenantName: <TENANT_NAME>

      # This is the reference to the secrets we created in the previous step
      existingTruefoundryCredsSecret: "truefoundry-creds"
      imagePullSecrets:
        - name: "truefoundry-image-pull-secret"
      ## Add if you have restricted public registry access
      # image:
      #   pullSecretNames:
      #   - "truefoundry-image-pull-secret"

      # Choose the resource tier as per your needs
      resourceTier: medium # or small or large

      config:
        defaultCloudProvider: "aws"
        storageConfiguration:
          awsS3BucketName: "<AWS_S3_BUCKET_NAME>"
          awsRegion: "<AWS_REGION>"

      serviceAccount:
        annotations:
          eks.amazonaws.com/role-arn: <CONTROL_PLANE_IAM_ROLE_ARN>
      ingress:
        hosts:
          - example.com
        enabled: true
        annotations: {}
        ingressClassName: nginx # Replace with your ingress class name
        # For haproxy openshift, please add the following annotation to the ingress
        haproxy.router.openshift.io/rewrite-target: /
    # In case, you want to spin up PostgreSQL on kubernetes, enable this
    # Please add creds and host details in the secret `truefoundry-creds`
    devMode:
      enabled: false
    tags:
      llmGateway: true
      llmGatewayRequestLogging: true

    # Disable few dependencies for only LLM Gateway setup
    tfyBuild:
      enabled: false
    sfyManifestService:
      enabled: false
    tfyController:
      enabled: false
    tfyConfigs:
      enabled: false
    ```
  </Tab>

  <Tab title="On-Prem">
    <Steps>
      <Step title="Get S3 compatible storage">
        You can use Minio or any other S3 compatible storage.

        <Accordion title="Using Minio (Self-Hosted S3-Compatible Storage)">
          If you're using Minio or another S3-compatible storage solution, you'll need to provide the following configuration:

          1. **Bucket Name**: Name of the S3-compatible bucket
          2. **Region**: Region identifier (can be any string for Minio)
          3. **Access Key ID**: Access key for authentication
          4. **Secret Access Key**: Secret key for authentication
          5. **Endpoint URL**: Full URL to your S3-compatible storage endpoint

          These values will be used in the configuration steps below.
        </Accordion>
      </Step>

      <Step title="Create Postgres database">
        If you have a managed postgres database, we highly recommend you to use it. If you don't have one, you can create a PostgreSQL database of size equivalent to `db.t3.medium` with storage size of 30GB.

        <Warning>
          **Important Configuration Notes:**

          * **For PostgreSQL 17+:** Disable SSL by setting `force_ssl` parameter to `0` in the parameter group
          * **Security Group:** Ensure your database security group/firewall rules allow traffic from Kubernetes node groups
        </Warning>

        <Accordion title="Using Managed PostgreSQL">
          For production environments, we strongly recommend using a managed PostgreSQL service:

          1. Create a PostgreSQL instance with the specifications from the [Compute Requirements](#compute-requirements) section
          2. Create a database named `truefoundry` (or your preferred name)
          3. Create a user with full permissions on this database
          4. Note down the following details:
             * **DB\_HOST**: Database hostname or IP address
             * **DB\_PORT**: Database port (usually 5432)
             * **DB\_NAME**: Database name
             * **DB\_USERNAME**: Database username
             * **DB\_PASSWORD**: Database password
        </Accordion>

        <Accordion title="Using PostgreSQL on Kubernetes (Dev/Testing Only)">
          For development or testing purposes only, you can deploy PostgreSQL on Kubernetes:

          1. Set `devMode.enabled: true` in the values file (shown in a later step)
          2. The PostgreSQL will be automatically deployed with the control plane

          This setup is NOT highly-available and should NOT be used for production environments.
        </Accordion>
      </Step>

      <Step title="Create Kubernetes Secrets">
        We will create two secrets in this step:

        1. Store the License Key and DB Credentials
        2. Store the Image Pull Secret

        <Accordion title="Create Kubernetes Secret for License Key and DB Credentials">
          We need to create a [Kubernetes secret](https://github.com/truefoundry/infra-charts/blob/main/charts/truefoundry/README.md#using-k8s-secret-for-required-fields) containing the licence key and db credentials.

          <Note>
            If you are using PostgreSQL on Kubernetes in the dev mode, the values will be as follows:

            DB\_HOST: \<HELM\_RELEASE\_NAME>-postgresql.\<NAMESPACE>.svc.cluster.local // eg. truefoundry-postgresql.truefoundry.svc.cluster.local

            DB\_NAME: truefoundry

            DB\_USERNAME: postgres # In order to use custom username, please update the same at `postgresql.auth.username`

            DB\_PASSWORD: randompassword # You can change this to any value here.
          </Note>

          ```yaml truefoundry-creds.yaml lines theme={"dark"}
          apiVersion: v1
          kind: Secret
          metadata:
            name: truefoundry-creds
          type: Opaque
          stringData:
            TFY_API_KEY: <TFY_API_KEY> # Provided by TrueFoundry team
            DB_HOST: <DB_HOST>
            DB_NAME: <DB_NAME>
            DB_USERNAME: <DB_USERNAME>
            DB_PASSWORD: <DB_PASSWORD>
          ```

          Apply the secret to the Kubernetes cluster (Assuming you are installing the control plane in the `truefoundry` namespace)

          ```bash lines theme={"dark"}
          kubectl apply -f truefoundry-creds.yaml -n truefoundry
          ```
        </Accordion>

        <Accordion title="Create Kubernetes Secret for Image Pull Secret">
          We need to create a [Image Pull Secret](https://github.com/truefoundry/infra-charts/blob/main/charts/truefoundry/README.md#using-k8s-secret-for-required-fields) to enable pulling the truefoundry images from the private registry.

          ```yaml truefoundry-image-pull-secret.yaml lines theme={"dark"}
          apiVersion: v1
          kind: Secret
          metadata:
            name: truefoundry-image-pull-secret
          type: kubernetes.io/dockerconfigjson
          data:
            .dockerconfigjson: <IMAGE_PULL_SECRET> # Provided by TrueFoundry team
          ```

          Apply the secret to the Kubernetes cluster (Assuming you are installing the control plane in the `truefoundry` namespace)

          ```bash lines theme={"dark"}
          kubectl apply -f truefoundry-image-pull-secret.yaml -n truefoundry
          ```
        </Accordion>
      </Step>

      <Step title="Create HelmChart Values file">
        Create a values file as given below and replace the following values:

        * `Control Plane URL`: URL that you will map to the control plane dashboard.
        * `Tenant Name`: Tenant name provided by TrueFoundry team.
        * `AWS S3 Bucket Name`: AWS S3 Bucket Name
        * `AWS Region`: AWS Region
        * `AWS Access Key ID`: AWS Access Key ID
        * `AWS Secret Access Key`: AWS Secret Access Key
        * `AWS Endpoint URL`: URL for your S3 compatible Blob storage

        ```yaml truefoundry-values.yaml wrap expandable lines theme={"dark"}
        global:
          # Domain to map the control plane dashboard
          controlPlaneURL: https://example.com

          # Ask TrueFoundry team to provide these
          tenantName: <TENANT_NAME>

          # This is the reference to the secrets we created in the previous step
          existingTruefoundryCredsSecret: "truefoundry-creds"
          imagePullSecrets:
            - name: "truefoundry-image-pull-secret"
          ## Add if you have restricted public registry access
          # image:
          #   pullSecretNames:
          #   - "truefoundry-image-pull-secret"

          # Choose the resource tier as per your needs
          resourceTier: medium # or small or large

          config:
            defaultCloudProvider: "aws"
            storageConfiguration:
              # In case of S3 compatible storage, provide the details here
              awsS3BucketName: "<AWS_S3_BUCKET_NAME_EQUIVALENT>"
              awsRegion: "<AWS_REGION_EQUIVALENT>"
          ingress:
            hosts:
              - example.com
            enabled: true
            annotations: {}
            ingressClassName: nginx # Replace with your ingress class name
        # In case, you want to spin up PostgreSQL on kubernetes, enable this
        # Please add creds and host details in the secret `truefoundry-creds`
        devMode:
          enabled: false
        tags:
          llmGateway: true
          llmGatewayRequestLogging: true

        servicefoundryServer:
          env:
            AWS_ACCESS_KEY_ID: <AWS_ACCESS_KEY_ID_EQUIVALENT>
            AWS_SECRET_ACCESS_KEY: <AWS_SECRET_ACCESS_KEY_EQUIVALENT>
            AWS_ENDPOINT_URL: <AWS_ENDPOINT_URL_EQUIVALENT>
            # AWS_ALLOW_HTTP: <AWS_ALLOW_HTTP_EQUIVALENT> (optional)

        deltaFusionIngestor:
          env:
            AWS_ACCESS_KEY_ID: <AWS_ACCESS_KEY_ID_EQUIVALENT>
            AWS_SECRET_ACCESS_KEY: <AWS_SECRET_ACCESS_KEY_EQUIVALENT>
            AWS_ENDPOINT_URL: <AWS_ENDPOINT_URL_EQUIVALENT>
            # AWS_ALLOW_HTTP: <AWS_ALLOW_HTTP_EQUIVALENT> (optional)

        deltaFusionQueryServer:
          env:
            AWS_ACCESS_KEY_ID: <AWS_ACCESS_KEY_ID_EQUIVALENT>
            AWS_SECRET_ACCESS_KEY: <AWS_SECRET_ACCESS_KEY_EQUIVALENT>
            AWS_ENDPOINT_URL: <AWS_ENDPOINT_URL_EQUIVALENT>
            # AWS_ALLOW_HTTP: <AWS_ALLOW_HTTP_EQUIVALENT> (optional)

        # Disable few dependencies for only LLM Gateway setup
        tfyBuild:
          enabled: false
        sfyManifestService:
          enabled: false
        tfyController:
          enabled: false
        tfyConfigs:
          enabled: false
        ```
      </Step>

      <Step title="Install Helm chart">
        ```bash wrap lines theme={"dark"}
        helm upgrade --install truefoundry oci://tfy.jfrog.io/tfy-helm/truefoundry -n truefoundry --create-namespace -f truefoundry-values.yaml
        ```
      </Step>
    </Steps>
  </Tab>
</Tabs>

## FAQ

<AccordionGroup>
  <Accordion title="How to add multiple gateway planes to the control plane?">
    You can add multiple gateway planes to the control plane by following the steps below:

    <Steps>
      <Step title="Create Kubernetes Secret for License Key and DB Credentials">
        We will create two secrets in this step:

        1. Store the License Key
        2. Store the Image Pull Secret

        <AccordionGroup>
          <Accordion title="Create Kubernetes Secret for License Key">
            We need to create a [Kubernetes secret](https://github.com/truefoundry/infra-charts/blob/main/charts/truefoundry/README.md#using-k8s-secret-for-required-fields) containing the licence key.

            <Info>
              Same license key will be used for all the gateway planes as used for the
              control plane
            </Info>

            ```yaml truefoundry-creds.yaml lines theme={"dark"}
            apiVersion: v1
            kind: Secret
            metadata:
              name: truefoundry-creds
            type: Opaque
            stringData:
              TFY_API_KEY: <TFY_API_KEY>
            ```

            Apply the secret to the Kubernetes cluster (Assuming you are installing the control plane in the `truefoundry` namespace)

            ```bash lines theme={"dark"}
            kubectl apply -f truefoundry-creds.yaml -n truefoundry
            ```
          </Accordion>

          <Accordion title="Create Kubernetes Secret for Image Pull Secret">
            We need to create a [Image Pull Secret](https://github.com/truefoundry/infra-charts/blob/main/charts/truefoundry/README.md#using-k8s-secret-for-required-fields) to enable pulling the truefoundry images from the private registry.

            <Info>
              Same image pull secret will be used for all the gateway planes as used for the
              control plane. Use your credentials if you are pulling TrueFoundry images from
              your registry.
            </Info>

            ```yaml truefoundry-image-pull-secret.yaml lines theme={"dark"}
            apiVersion: v1
            kind: Secret
            metadata:
              name: truefoundry-image-pull-secret
            type: kubernetes.io/dockerconfigjson
            data:
              .dockerconfigjson: <IMAGE_PULL_SECRET> # Provided by TrueFoundry team
            ```

            Apply the secret to the Kubernetes cluster (Assuming you are installing the control plane in the `truefoundry` namespace)

            ```bash lines theme={"dark"}
            kubectl apply -f truefoundry-image-pull-secret.yaml -n truefoundry
            ```
          </Accordion>
        </AccordionGroup>
      </Step>

      <Step title="Create Helm chart Values file for gateway plane">
        Create a values file as given below and replace the following values:

        * `CONTROL_PLANE_URL`: URL that you will map to the control plane dashboard.
        * `TENANT_NAME`: Tenant name provided by TrueFoundry team.
        * `GATEWAY_ENDPOINT_HOST`: The domain where you will expose the gateway endpoint (e.g., `gateway.example.com`)

        ```yaml truefoundry-gateway-values.yaml wrap expandable lines theme={"dark"}
        global:
          # This is the reference to the secrets we created in the previous step
          imagePullSecrets:
            - name: "truefoundry-image-pull-secret"

          # Choose the resource tier as per your needs
          resourceTier: medium # or small or large
          controlPlaneURL: <CONTROL_PLANE_URL> # eg. https://example-company.truefoundry.cloud
          tenantName: <TENANT_NAME>

        ingress:
          enabled: true
          annotations: {}
          ingressClassName: nginx
          tls: []
          hosts:
            - <GATEWAY_ENDPOINT_HOST>

        # Optional: Istio configuration (if using Istio instead of standard ingress)
        # istio:
        #   virtualservice:
        #     hosts:
        #       - <GATEWAY_ENDPOINT_HOST>
        #     enabled: true
        #     retries:
        #       enabled: true
        #       retryOn: gateway-error
        #     gateways:
        #       - istio-system/tfy-wildcard
        #     annotations: {}
        ```
      </Step>

      <Step title="Install Helm chart for gateway plane">
        ```bash wrap lines theme={"dark"}
        helm upgrade --install tfy-llm-gateway oci://tfy.jfrog.io/tfy-helm/tfy-llm-gateway -n truefoundry --create-namespace -f truefoundry-gateway-values.yaml
        ```
      </Step>
    </Steps>
  </Accordion>

  <Accordion title="Can I use my Artifactory as a mirror to pull images?">
    Yes. You can configure your Artifactory to mirror our registry.

    <Note>
      Credentials for accessing the TrueFoundry private registry are required and
      will be provided during onboarding.
    </Note>

    **1. Registry Configuration**

    * **URL**: `https://tfy.jfrog.io/`

    **2. Update Helm values**

    ```yaml wrap lines theme={"dark"}
    global:
      image:
        registry: <YOUR_REGISTRY> # Replace with your registry
    postgresql:
      image:
        registry: <YOUR_REGISTRY> # Replace with your registry, use this if `devMode` is enabled
    ```
  </Accordion>

  <Accordion title="Can I copy images to my own private registry?">
    Yes. We provide a [script](https://github.com/truefoundry/infra-charts/blob/main/scripts/clone_images_to_your_registry.sh) that uses the `truefoundry` Helm Chart to identify and copy required images to your private registry.

    <Note>
      Credentials for accessing the TrueFoundry private registry are required and
      will be provided during onboarding.
    </Note>

    <Tabs>
      <Tab title="Generic Registry">
        **1. Install required dependencies**

        * [Skopeo](https://github.com/containers/skopeo/blob/main/install.md)
          * Used to perform the image copy operation.
        * [Helm](https://helm.sh/docs/intro/install/)
          * Used to get the list of images from the TrueFoundry Helm Chart.

        **2. Add TrueFoundry Helm Chart repository**

        ```bash wrap lines theme={"dark"}
        helm repo add truefoundry https://truefoundry.github.io/infra-charts
        helm repo update
        ```

        **3. Authenticate to the TrueFoundry source registry**

        ```bash wrap lines theme={"dark"}
        skopeo login -u <USERNAME> -p <PASSWORD> https://tfy.jfrog.io/
        ```

        <Note>
          Replace `<USERNAME>` with the TrueFoundry registry username.\
          Replace `<PASSWORD>` with the TrueFoundry registry password.
        </Note>

        **4. Authenticate to your destination registry**

        ```bash wrap lines theme={"dark"}
        skopeo login -u <USERNAME> -p <PASSWORD> <YOUR_REGISTRY>
        ```

        <Note>
          Replace `<USERNAME>` with your registry username.\
          Replace `<PASSWORD>` with your registry password.\
          Replace `<YOUR_REGISTRY>` with the URL of your registry.

          Skopeo will use authentication details for a registry that was previously authenticated with `docker login`.

          Alternatively, you can use the `--dest-user` and `--dest-password` flags to provide the username and password for the destination registry.
        </Note>

        **5. Run Clone Image Script**

        ```bash wrap lines theme={"dark"}
        export TRUEFOUNDRY_HELM_CHART_VERSION=<TRUEFOUNDRY_HELM_CHART_VERSION>
        export TRUEFOUNDRY_HELM_VALUES_FILE=<TRUEFOUNDRY_HELM_VALUES_FILE>
        export DEST_REGISTRY=<YOUR_DESTINATION_REGISTRY>

        # Dry-run example
        curl -s https://raw.githubusercontent.com/truefoundry/infra-charts/main/scripts/clone_images_to_your_registry.sh | bash -s -- --helm-chart truefoundry --helm-version $TRUEFOUNDRY_HELM_CHART_VERSION --helm-values $TRUEFOUNDRY_HELM_VALUES_FILE --dest-registry $DEST_REGISTRY --dry-run

        # Live example
        curl -s https://raw.githubusercontent.com/truefoundry/infra-charts/main/scripts/clone_images_to_your_registry.sh | bash -s -- --helm-chart truefoundry --helm-version $TRUEFOUNDRY_HELM_CHART_VERSION --helm-values $TRUEFOUNDRY_HELM_VALUES_FILE --dest-registry $DEST_REGISTRY
        ```

        <Note>
          Replace `<TRUEFOUNDRY_HELM_CHART_VERSION>` with the version of the Truefoundry
          helm chart you want to use. You can find the latest version in the
          [changelog](https://docs.truefoundry.com/changelog).

          Replace `<TRUEFOUNDRY_HELM_VALUES_FILE>` with the path to the values file you created in the [Installation Instructions](#installation-instructions).

          Replace `<DEST_REGISTRY>` with the URL of your registry.
        </Note>

        **6. Update the Helm values file to use your registry**

        ```yaml wrap lines theme={"dark"}
        global:
          image:
            registry: <YOUR_REGISTRY> # Replace with your registry
        postgresql:
          image:
            registry: <YOUR_REGISTRY> # Replace with your registry, use this if `devMode` is enabled
        ```
      </Tab>

      <Tab title="AWS ECR Registry">
        **1. Install required dependencies**

        * [Skopeo](https://github.com/containers/skopeo/blob/main/install.md)
          * Used to perform the image copy operation.
        * [Helm](https://helm.sh/docs/intro/install/)
          * Used to get the list of images from the TrueFoundry Helm Chart.
        * [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html)
          * Used to perform AWS ECR actions to validate and create repositories.

        **2. Add TrueFoundry Helm Chart repository**

        ```bash wrap lines theme={"dark"}
        helm repo add truefoundry https://truefoundry.github.io/infra-charts
        helm repo update
        ```

        **3. Authenticate to the TrueFoundry source registry**

        ```bash wrap lines theme={"dark"}
        skopeo login -u <USERNAME> -p <PASSWORD> https://tfy.jfrog.io/
        ```

        <Note>
          Replace `<USERNAME>` with the TrueFoundry registry username.\
          Replace `<PASSWORD>` with the TrueFoundry registry password.
        </Note>

        **4. Authenticate to your destination registry**

        ```bash wrap lines theme={"dark"}
        # Set your AWS profile
        export AWS_PROFILE=<AWS_PROFILE>

        # Authenticate to ECR using the profile
        aws ecr get-login-password --region us-west-2 | skopeo login --username AWS --password-stdin <YOUR_ECR_REGISTRY>
        ```

        <Note>
          Replace `<AWS_PROFILE>` with your AWS profile name.\
          Replace `<YOUR_ECR_REGISTRY>` with the URL of your ECR registry (ex. `123456789012.dkr.ecr.us-east-2.amazonaws.com`).

          Skopeo will use authentication details for a registry that was previously authenticated with `docker login`.
        </Note>

        **5. Run Clone Image Script**

        * This script creates required ECR repositories and copies images.
        * Optionally append a path to your registry URL to namespace repositories (e.g., `123456789012.dkr.ecr.us-east-2.amazonaws.com/truefoundry`).

        ```bash wrap lines theme={"dark"}
        export TRUEFOUNDRY_HELM_CHART_VERSION=<TRUEFOUNDRY_HELM_CHART_VERSION>
        export TRUEFOUNDRY_HELM_VALUES_FILE=<TRUEFOUNDRY_HELM_VALUES_FILE>
        export DEST_REGISTRY=<YOUR_DESTINATION_REGISTRY>

        # Dry-run example
        curl -s https://raw.githubusercontent.com/truefoundry/infra-charts/main/scripts/clone_images_to_your_registry.sh | bash -s -- --helm-chart truefoundry --helm-version $TRUEFOUNDRY_HELM_CHART_VERSION --helm-values $TRUEFOUNDRY_HELM_VALUES_FILE --dest-registry $DEST_REGISTRY --dry-run

        # Live example
        curl -s https://raw.githubusercontent.com/truefoundry/infra-charts/main/scripts/clone_images_to_your_registry.sh | bash -s -- --helm-chart truefoundry --helm-version $TRUEFOUNDRY_HELM_CHART_VERSION --helm-values $TRUEFOUNDRY_HELM_VALUES_FILE --dest-registry $DEST_REGISTRY
        ```

        <Note>
          Replace `<TRUEFOUNDRY_HELM_CHART_VERSION>` with the TrueFoundry Helm chart version. Find the latest version in the [changelog](https://docs.truefoundry.com/changelog).

          Replace `<TRUEFOUNDRY_HELM_VALUES_FILE>` with the path to your values file from [Installation Instructions](#installation-instructions).

          Replace `<YOUR_DESTINATION_ECR_REGISTRY>` with your ECR registry URL (e.g., `123456789012.dkr.ecr.us-east-2.amazonaws.com/truefoundy`).
        </Note>

        **6. Update the Helm values file to use your registry**

        ```yaml wrap lines theme={"dark"}
        global:
          image:
            registry: <YOUR_REGISTRY> # Replace with your registry
        postgresql:
          image:
            registry: <YOUR_REGISTRY> # Replace with your registry, use this if `devMode` is enabled
        ```
      </Tab>
    </Tabs>
  </Accordion>

  <Accordion title="How to install in an air-gapped / restricted network environment?">
    An air-gapped environment is isolated from the internet. Since the control plane and gateway plane ship as a single helm chart (`truefoundry`), you only need to make the container images available in your private registry and update the helm values to point to it.

    1. **Copy images** to your private registry — set up a [registry mirror](#can-i-use-my-artifactory-as-a-mirror-to-pull-images) or [copy images directly](#can-i-copy-images-to-my-own-private-registry) using the steps described in the FAQs above
    2. **Update helm values** to point to your private registry (see the helm value overrides in the same FAQs above)
    3. **Continue with the standard installation** on this page
  </Accordion>

  <Accordion title="How to integrate with AWS bedrock models from a different AWS account?">
    You can integrate with AWS bedrock models from a different AWS account by following the steps below:

    1. Add the following IAM policy to the control plane IAM role so that it can assume the IAM role of the AWS account that has the bedrock models:

    ```json lines theme={"dark"}
    {
      "Statement": [
        {
          "Action": "sts:AssumeRole",
          "Effect": "Allow",
          "Resource": "*"
        }
      ],
      "Version": "2012-10-17"
    }
    ```

    2. In the IAM role in the destination AWS account (which has bedrock access), add the following trust policy to allow the control plane IAM role to assume it:

    ```json lines theme={"dark"}
    {
      "Statement": [
        {
          "Effect": "Allow",
          "Principal": {
            "AWS": "<CONTROL_PLANE_IAM_ROLE_ARN>"
          },

          "Action": "sts:AssumeRole"
        }
      ],
      "Version": "2012-10-17"
    }
    ```

    3. Now you can use the IAM role of the destination AWS account while [integrating AWS bedrock models](/docs/ai-gateway/aws-bedrock) in the TrueFoundry AI gateway.
  </Accordion>

  <Accordion title="Do we need any NFS volumes in Kubernetes for the AI Gateway or Control Plane?">
    No, we only need block storage for installing and running Truefoundry. This should be supported via the CSI driver and only ReadWriteOnce access is required.
  </Accordion>

  <Accordion title="What is the structure of access logs">
    We log access information in standard output with the following format:

    1. logfmt
    2. json

    These can be switched with the help of an environment variable to the AI Gateway installation. (Default: logfmt)

    ### Log format

    Standard log format structure:

    ```wrap lines theme={"dark"}
    time="%START_TIME%" level=%LEVEL% ip=%IP_ADDRESS% tenant=%TENANT_NAME% user=%SUBJECT_TYPE%:%SUBJECT_SLUG% model=%MODEL_ID% method=%METHOD% path=%PATH% status=%STATUS_CODE% time_taken=%DURATION%ms trace_id=%TRACE_ID%
    ```

    | Log operator  | Details                                                                                                             |
    | ------------- | ------------------------------------------------------------------------------------------------------------------- |
    | START\_TIME   | ISO timestamp for request start. eg. 2025-08-12 13:34:50                                                            |
    | LEVEL         | info\|warn\|error                                                                                                   |
    | IP\_ADDRESS   | IP address of the caller. eg. ::ffff:10.99.55.142                                                                   |
    | TENANT\_NAME  | Name of the tenant. eg. truefoundry                                                                                 |
    | SUBJECT\_TYPE | user\|virtualaccount                                                                                                |
    | SUBJECT\_SLUG | Email or virtual account name. eg. [tfy-user@truefoundry.com](mailto:tfy-user@truefoundry.com)\|demo-virtualaccount |
    | MODEL\_ID     | Model ID. eg. openai-default/gpt-5                                                                                  |
    | METHOD        | GET\|POST\|PUT                                                                                                      |
    | PATH          | Path of the request. eg. /api/inference/openai/chat/completions                                                     |
    | STATUS\_CODE  | 200\|400\|401\|403\|429\|500                                                                                        |
    | DURATION      | Duration of the request. eg. 12                                                                                     |
    | TRACE\_ID     | Trace ID of the request                                                                                             |

    **Examples:**

    ```wrap lines theme={"dark"}
    time="2025-08-12 13:34:50" level=info ip=::ffff:10.99.55.142 tenant=truefoundry user=virtualaccount:demo-virtualaccount model=openai-default/gpt-5 method=POST path=/api/inference/openai/chat/completions status=200 time_taken=53ms trace_id=587b2a946c13f62f9160674a8c983ce3
    ```
  </Accordion>

  <Accordion title="How to use SSO directly without using TrueFoundry Auth Server?">
    By default, the control plane uses the TrueFoundry Auth Server for user authentication. However, you can configure it to use your own external identity provider instead. We support both OIDC and SAML-compliant identity providers. [Read more](https://www.truefoundry.com/docs/deploy-control-plane-with-external-oauth)
  </Accordion>

  <Accordion title="Requests to the gateway are timing out after a certain duration">
    If your LLM requests are timing out after a certain duration, the first thing to check is the **traces** in the TrueFoundry dashboard. Look at the request duration — if you see requests consistently timing out at exactly 60 seconds, the issue is almost certainly the **load balancer**, not the TrueFoundry AI Gateway. **The TrueFoundry gateway does not impose any request timeout.**

    <img src="https://mintcdn.com/truefoundry/yR_clVDeJDlQkXKY/images/docs/ai-gateway/alb-timeout-traces.png?fit=max&auto=format&n=yR_clVDeJDlQkXKY&q=85&s=56c0579b9e5bf6f68663f242242499ff" alt="Traces showing requests timing out at 60 seconds" width="3470" height="1090" data-path="images/docs/ai-gateway/alb-timeout-traces.png" />

    This commonly happens when an **Application Load Balancer (ALB)** is placed in front of the gateway to expose it. The default **Connection idle timeout** on AWS ALBs is 60 seconds, which is too short for long-running LLM inference requests (especially streaming responses or large prompts).

    **Solution:** Increase the idle timeout on your AWS ALB to a higher value (e.g., 300 seconds or more).

    You can find this setting in the AWS Console under **EC2 → Load Balancers → Select your ALB → Attributes tab → Connection idle timeout**.

    <img src="https://mintcdn.com/truefoundry/yR_clVDeJDlQkXKY/images/docs/ai-gateway/aws-alb-idle-timeout-setting.png?fit=max&auto=format&n=yR_clVDeJDlQkXKY&q=85&s=c1b0431e90ea8b9cb765a7f9cb21e18d" alt="AWS ALB Connection idle timeout setting" width="1024" height="447" data-path="images/docs/ai-gateway/aws-alb-idle-timeout-setting.png" />

    You can also update it via the AWS CLI:

    ```bash theme={"dark"}
    aws elbv2 modify-load-balancer-attributes \
      --load-balancer-arn <YOUR_ALB_ARN> \
      --attributes Key=idle_timeout.timeout_seconds,Value=300
    ```

    <Info>
      If you are using an ingress controller (e.g., NGINX Ingress) in addition to the ALB, also verify that the ingress controller's proxy timeout settings are configured appropriately.
    </Info>
  </Accordion>

  <Accordion title="Can I get TrueFoundry metrics in Victoria Metrics instead of Prometheus?">
    Yes. TrueFoundry supports exporting metrics to [Victoria Metrics](https://victoriametrics.com/) as an alternative to Prometheus. To enable this, add the following to your `truefoundry-values.yaml` file and upgrade the Helm release:

    <Note>
      This only installs the `VMServiceScrape` and related custom resources for
      scraping TrueFoundry metrics. It does **not** deploy Victoria Metrics itself —
      you are responsible for installing and managing your own Victoria Metrics
      instance.
    </Note>

    ```yaml truefoundry-values.yaml wrap lines theme={"dark"}
    victoriaMetricsMonitoring:
      enabled: true
    ```

    Then upgrade the Helm release to apply the changes:

    ```bash wrap lines theme={"dark"}
    helm upgrade --install truefoundry oci://tfy.jfrog.io/tfy-helm/truefoundry -n truefoundry --create-namespace -f truefoundry-values.yaml
    ```
  </Accordion>

  <Accordion title="How to enable SSL for PostgreSQL connections?">
    The TrueFoundry control plane supports SSL connections to PostgreSQL. You can configure SSL by setting the `DB_SSL_MODE` environment variable in your `truefoundry-values.yaml`.

    Supported `DB_SSL_MODE` values:

    | Mode          | Encryption | Certificate Validation | Use Case                                                        |
    | ------------- | ---------- | ---------------------- | --------------------------------------------------------------- |
    | `disable`     | No         | No                     | Local development or trusted networks                           |
    | `no-verify`   | Yes        | No                     | Managed databases with self-signed or unverified certs          |
    | `require`     | Yes        | Yes (system CA store)  | When you have a valid CA certificate and want full verification |
    | `verify-ca`   | Yes        | Yes (custom CA)        | Same as `require` but explicitly checks CA                      |
    | `verify-full` | Yes        | Yes (CA + hostname)    | Strictest mode, validates CA and hostname                       |

    SSL certificate environment variables:

    | Variable           | Purpose                                        | Required                                                     |
    | ------------------ | ---------------------------------------------- | ------------------------------------------------------------ |
    | `DB_SSL_CA_PATH`   | Path to the server CA certificate file         | For `require`, `verify-ca`, or `verify-full` modes           |
    | `DB_SSL_CERT_PATH` | Path to the client certificate file (for mTLS) | Only for mTLS (GCP Cloud SQL, Azure Database for PostgreSQL) |
    | `DB_SSL_KEY_PATH`  | Path to the client private key file (for mTLS) | Only for mTLS (GCP Cloud SQL, Azure Database for PostgreSQL) |

    <Note>
      The certificate requirements vary by cloud provider. AWS RDS only needs the server CA bundle (`DB_SSL_CA_PATH`), while GCP Cloud SQL and Azure Database for PostgreSQL may require all three certificate paths when client certificate authentication (mTLS) is enabled. Refer to the cloud-specific control plane documentation for detailed examples.
    </Note>

    **Scenario 1: Encrypted connection without certificate validation (`no-verify`)**

    This is the simplest option for managed databases. It encrypts the connection but skips server certificate validation.

    ```yaml truefoundry-values.yaml wrap lines theme={"dark"}
    servicefoundryServer:
      env:
        DB_SSL_MODE: "no-verify"
    mlfoundryServer:
      env:
        DB_SSL_MODE: "no-verify"
    ```

    **Scenario 2: Encrypted connection with certificate validation (`require`)**

    This mode encrypts the connection and validates the server certificate. You must provide the appropriate certificate files for your database provider. The example below shows the full configuration with all three certificate paths (for GCP/Azure mTLS). For AWS RDS, only `DB_SSL_CA_PATH` is needed.

    Create a Kubernetes Secret containing your certificate files:

    ```bash wrap lines theme={"dark"}
    # AWS RDS (CA bundle only)
    kubectl create secret generic db-ssl-certs \
      --from-file=ca-certificate.crt=/path/to/your/ca-certificate.crt \
      -n truefoundry

    # GCP Cloud SQL / Azure (full mTLS)
    kubectl create secret generic db-ssl-certs \
      --from-file=ca-certificate.crt=/path/to/server-ca.pem \
      --from-file=client-cert.pem=/path/to/client-cert.pem \
      --from-file=client-key.pem=/path/to/client-key.pem \
      -n truefoundry
    ```

    Then configure `truefoundry-values.yaml` to mount the certificates and set the SSL paths:

    ```yaml truefoundry-values.yaml wrap lines theme={"dark"}
    servicefoundryServer:
      env:
        DB_SSL_MODE: "require"
        DB_SSL_CA_PATH: "/etc/ssl/custom/ca-certificate.crt"
        # Only needed for mTLS (GCP Cloud SQL, Azure Database for PostgreSQL)
        DB_SSL_CERT_PATH: "/etc/ssl/custom/client-cert.pem"
        DB_SSL_KEY_PATH: "/etc/ssl/custom/client-key.pem"
      extraVolumes:
        - name: db-ssl-certs
          secret:
            secretName: db-ssl-certs
      extraVolumeMounts:
        - name: db-ssl-certs
          mountPath: /etc/ssl/custom
          readOnly: true
    mlfoundryServer:
      env:
        DB_SSL_MODE: "require"
        DB_SSL_CA_PATH: "/etc/ssl/custom/ca-certificate.crt"
        # Only needed for mTLS (GCP Cloud SQL, Azure Database for PostgreSQL)
        DB_SSL_CERT_PATH: "/etc/ssl/custom/client-cert.pem"
        DB_SSL_KEY_PATH: "/etc/ssl/custom/client-key.pem"
      extraVolumes:
        - name: db-ssl-certs
          secret:
            secretName: db-ssl-certs
      extraVolumeMounts:
        - name: db-ssl-certs
          mountPath: /etc/ssl/custom
          readOnly: true
    ```

    Upgrade the Helm release to apply the changes:

    ```bash wrap lines theme={"dark"}
    helm upgrade --install truefoundry oci://tfy.jfrog.io/tfy-helm/truefoundry -n truefoundry --create-namespace -f truefoundry-values.yaml
    ```
  </Accordion>

  <Accordion title="How to configure custom CA certificates?">
    If your TrueFoundry deployment needs to trust custom Certificate Authorities (e.g., for internal services, private registries, or corporate proxies), you can configure custom CA certificates in the Helm chart.

    There are two methods to provide custom CA certificates:

    ### Method 1: Pass customCA as a multiline string

    You can directly provide the CA certificate content as a multiline string in your `values.yaml`:

    ```yaml truefoundry-values.yaml wrap lines theme={"dark"}
    global:
      customCA:
        enabled: true
        certificate: |
          -----BEGIN CERTIFICATE-----
          MIIDXTCCAkWgAwIBAgIJAKZ7VqHEqvmKMA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV
          BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
          ... (rest of your certificate) ...
          -----END CERTIFICATE-----
    ```

    This method is suitable when you have one or a few CA certificates to add.

    ### Method 2: Use an existing ConfigMap containing CA certificate(s)

    If you already have your custom CA certificates in a Kubernetes ConfigMap, you can reference it directly. An initContainer will merge the custom CA with the system CAs.

    <Steps>
      <Step title="Create a ConfigMap with your custom CA certificate(s)">
        Create a Kubernetes ConfigMap containing your custom CA certificate(s):

        ```bash wrap lines theme={"dark"}
        kubectl create configmap custom-ca-certificates \
          --from-file=ca-certificates.crt=custom-ca.crt \
          -n truefoundry
        ```

        Alternatively, if you want to create it from a YAML file:

        ```yaml custom-ca-configmap.yaml wrap lines theme={"dark"}
        apiVersion: v1
        kind: ConfigMap
        metadata:
          name: custom-ca-certificates
          namespace: truefoundry
        data:
          ca-certificates.crt: |
            -----BEGIN CERTIFICATE-----
            ... (your custom CA certificate content) ...
            -----END CERTIFICATE-----
        ```

        Apply the ConfigMap:

        ```bash wrap lines theme={"dark"}
        kubectl apply -f custom-ca-configmap.yaml
        ```
      </Step>

      <Step title="Reference the ConfigMap in your Helm values">
        Update your `truefoundry-values.yaml` to reference the ConfigMap:

        ```yaml truefoundry-values.yaml wrap lines theme={"dark"}
        global:
          customCA:
            enabled: true
            existingConfigMap:
              name: custom-ca-certificates
        ```
      </Step>

      <Step title="Upgrade the Helm installation">
        Apply the changes by upgrading your Helm release:

        ```bash wrap lines theme={"dark"}
        helm upgrade --install truefoundry oci://tfy.jfrog.io/tfy-helm/truefoundry \
          -n truefoundry --create-namespace -f truefoundry-values.yaml
        ```
      </Step>
    </Steps>

    ### Method 2b: Use an existing ConfigMap with `overrideCAList`

    If you want the ConfigMap to **replace** the system CA bundle entirely instead of merging, set `overrideCAList` to `true`. In this mode, the ConfigMap is mounted directly at `/etc/ssl/certs/` (no initContainer is used), so the ConfigMap must contain the **full** CA bundle (system + custom CAs).

    <Steps>
      <Step title="Prepare your CA certificate file">
        Add your custom CA certificate(s) to your system's CA bundle. On a Linux system with the certificate file saved as `custom-ca.crt`:

        ```bash wrap lines theme={"dark"}
        # Copy the certificate to the CA directory
        sudo cp custom-ca.crt /usr/local/share/ca-certificates/

        # Update the CA certificates bundle
        sudo update-ca-certificates
        ```

        This will generate or update `/etc/ssl/certs/ca-certificates.crt` with your custom CA included (system CAs + your custom CA).
      </Step>

      <Step title="Create a ConfigMap from the complete ca-certificates.crt file">
        Create a Kubernetes ConfigMap containing the complete CA bundle:

        ```bash wrap lines theme={"dark"}
        kubectl create configmap custom-ca-certificates \
          --from-file=ca-certificates.crt=/etc/ssl/certs/ca-certificates.crt \
          -n truefoundry
        ```

        Alternatively, if you want to create it from a YAML file:

        ```yaml custom-ca-configmap.yaml wrap lines theme={"dark"}
        apiVersion: v1
        kind: ConfigMap
        metadata:
          name: custom-ca-certificates
          namespace: truefoundry
        data:
          ca-certificates.crt: |
            -----BEGIN CERTIFICATE-----
            ... (your complete ca-certificates.crt content including system + custom CAs) ...
            -----END CERTIFICATE-----
        ```

        Apply the ConfigMap:

        ```bash wrap lines theme={"dark"}
        kubectl apply -f custom-ca-configmap.yaml
        ```
      </Step>

      <Step title="Reference the ConfigMap in your Helm values with overrideCAList">
        Update your `truefoundry-values.yaml` to reference the ConfigMap with `overrideCAList` enabled:

        ```yaml truefoundry-values.yaml wrap lines theme={"dark"}
        global:
          customCA:
            enabled: true
            existingConfigMap:
              name: custom-ca-certificates
              overrideCAList: true
        ```
      </Step>

      <Step title="Upgrade the Helm installation">
        Apply the changes by upgrading your Helm release:

        ```bash wrap lines theme={"dark"}
        helm upgrade --install truefoundry oci://tfy.jfrog.io/tfy-helm/truefoundry \
          -n truefoundry --create-namespace -f truefoundry-values.yaml
        ```
      </Step>
    </Steps>

    <Warning>
      When `overrideCAList` is set to `true`, the ConfigMap is mounted directly replacing the system CA bundle. Your ConfigMap must contain the complete CA bundle (system CAs + your custom CAs). If you only include your custom CAs, all standard public CA trust will be lost and outbound HTTPS connections to public services will fail.
    </Warning>

    <Info>
      The custom CA certificates will be mounted into all TrueFoundry pods and added to the system's trust store. This ensures that all outgoing HTTPS connections from TrueFoundry services will trust your custom CAs.
    </Info>

    <Note>
      After adding custom CA certificates, verify that your TrueFoundry pods have restarted and are running correctly. You may need to restart existing pods for the changes to take effect.
    </Note>
  </Accordion>

  <Accordion title="How to enable and access control plane monitoring (Grafana)?">
    TrueFoundry ships with a built-in monitoring stack that includes Grafana dashboards for the control plane. To enable it, add the following to your `truefoundry-values.yaml`:

    ```yaml truefoundry-values.yaml theme={"dark"}
    truefoundryMonitoring:
      enabled: true
      grafana:
        grafana.ini:
          auth.jwt:
            jwk_set_url: >-
              https://<your-truefoundry-control-plane-url>/api/svc/v1/keys/<tenant-name>/jwks
    ```

    Then upgrade the Helm release to apply the changes:

    ```bash theme={"dark"}
    helm upgrade --install truefoundry oci://tfy.jfrog.io/tfy-helm/truefoundry \
      -n truefoundry --create-namespace \
      -f truefoundry-values.yaml
    ```

    Once enabled, platform **admins** can access the Grafana dashboard at:

    ```
    https://<your-truefoundry-control-plane-url>/admin/grafana/
    ```

    <Note>
      * Replace `<your-truefoundry-control-plane-url>` with your actual control plane domain (e.g., `app.example.com`) and `<tenant-name>` with your TrueFoundry tenant name provided during onboarding.
      * Only users with the **admin** role can access this endpoint.
      * Make sure to include the trailing `/` at the end of the URL.
      * If you already have Prometheus or VictoriaLogs in your cluster, you can point the monitoring stack to them using `externalServices` instead of installing new instances.
    </Note>

    For the full configuration reference, see the [Control Plane Monitoring](/docs/platform/controlplane-monitoring) guide.
  </Accordion>

  <Accordion title="How do you add default metadata to all requests passing via the gateway?">
    You can attach default metadata to every request that passes through the AI Gateway by setting the `DEFAULT_GATEWAY_METADATA` environment variable on the gateway. The value should be a JSON string of key-value pairs.

    Add the following to your gateway configuration in values file of the gateway plane:

    ```yaml theme={"dark"}
    tfy-llm-gateway:
      env:
        DEFAULT_GATEWAY_METADATA: '{"org":"internal"}'
    ```

    <Note>
      The metadata key-value pairs will be automatically included in every request routed through the gateway. You can use this to tag requests with organizational identifiers, environment labels, or any other metadata your downstream systems need.
    </Note>
  </Accordion>

  <Accordion title="How to use HTTPRoute to route traffic using Kubernetes Gateway API?">
    The TrueFoundry Helm charts support the [Kubernetes Gateway API](https://gateway-api.sigs.k8s.io/) as an alternative to standard `Ingress` resources. Use `HTTPRoute` when your cluster uses a Gateway API-compatible controller (e.g. Envoy Gateway, Istio, NGINX Gateway Fabric, GKE Gateway).

    **Control plane (truefoundry chart)**

    Add the following to your `truefoundry-values.yaml`, setting `parentRefs` to point to your existing `Gateway`:

    ```yaml truefoundry-values.yaml theme={"dark"}
    global:
      httpRoute:
        enabled: true
        parentRefs:
          - name: my-gateway        # Name of your Gateway resource
            namespace: gateway-system  # Namespace where the Gateway is deployed
            sectionName: https      # Listener section on the Gateway (e.g. http or https)
        hostnames:
          - "app.example.com"       # Hostname that this HTTPRoute should match
    ```

    Then apply:

    ```bash theme={"dark"}
    helm upgrade --install truefoundry oci://tfy.jfrog.io/tfy-helm/truefoundry \
      -n truefoundry --create-namespace \
      -f truefoundry-values.yaml
    ```

    <Note>
      * Only one routing method should be enabled at a time. Disable `global.ingress.enabled` and `global.virtualservice.enabled` when using `httpRoute`.
      * The `sectionName` must match a named listener on your `Gateway` resource. Omit it if your Gateway has a single unnamed listener.
      * TLS termination is handled by the parent `Gateway` — no TLS configuration is needed on the `HTTPRoute` itself.
    </Note>
  </Accordion>
</AccordionGroup>
