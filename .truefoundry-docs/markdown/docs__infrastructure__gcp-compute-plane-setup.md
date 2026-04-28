> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# GCP

> This page provides an overview of the architecture, requirements and steps to install the TrueFoundry compute plane cluster in GCP

The architecture of a TrueFoundry compute plane is as follows:

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/DdP_2rhue4AQQlob/images/4c4c369b-9f78484-GCP_1.png?fit=max&auto=format&n=DdP_2rhue4AQQlob&q=85&s=f1dce08f7d2e9e779f5aad5c9a855146" width="1779" height="1182" data-path="images/4c4c369b-9f78484-GCP_1.png" />
</Frame>

<Accordion title="Access Policies Overview">
  | Policy                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | Description                                                                                                                                                                                                                                                                                                                                                                                              |
  | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
  | RolePolicy with policies for - [Artifact registry](https://github.com/truefoundry/terraform-google-truefoundry-platform-features/blob/c1d198b003e6673e65a64c8e262bf8c07fb1f6ea/iam.tf#L117-L144), [Secrets manager](https://github.com/truefoundry/terraform-google-truefoundry-platform-features/blob/c1d198b003e6673e65a64c8e262bf8c07fb1f6ea/iam.tf#L16-L28), [Blob storage](https://github.com/truefoundry/terraform-google-truefoundry-platform-features/blob/c1d198b003e6673e65a64c8e262bf8c07fb1f6ea/iam.tf#L52-L68), [Cluster viewer](https://github.com/truefoundry/terraform-google-truefoundry-platform-features/blob/c1d198b003e6673e65a64c8e262bf8c07fb1f6ea/iam.tf#L92-L99), [IAM serviceaccount token creator](https://github.com/truefoundry/terraform-google-truefoundry-platform-features/blob/c1d198b003e6673e65a64c8e262bf8c07fb1f6ea/iam.tf#L160), [Logging viewer](https://github.com/truefoundry/terraform-google-truefoundry-platform-features/blob/c1d198b003e6673e65a64c8e262bf8c07fb1f6ea/iam.tf#L168) | Role `<cluster_name>-platform-user` with permissions for:<br />- Creating and managing blob storage buckets<br />- Managing secrets in secret manager<br />- Pulling and pushing images to artifact registry<br />- Enabling cloud integration for GCP (node level details)<br />- Viewing cluster autoscaler logs<br />- Creating Service Account keys (Service Account key creation should be allowed) |
</Accordion>

## Requirements:

The common requirements to setup compute plane in each of the scenarios is as follows:

* <Icon icon="square-check" iconType="regular" /> Billing must be enabled for the GCP account.
* <Icon icon="square-check" iconType="regular" /> Following APIs must be enabled in the project -
  <Accordion title="Required APIs" default>
    * Compute Engine API - This API must be enabled for Virtual Machines
    * Kubernetes Engine API - This API must be enabled for Kubernetes clusters
    * Storage Engine API - This API must be enabled for GCP Blob storage - Buckets
    * Artifact Registry API - This API must be enabled for docker registry and image builds
    * Secrets Manager API - This API must be enabled to support Secret management
  </Accordion>
* <Icon icon="square-check" iconType="regular" /> Egress access to container registries - `public.ecr.aws`, `quay.io`, `ghcr.io`, `tfy.jfrog.io`, `docker.io/natsio`, `nvcr.io`, `registry.k8s.io` so that we can download the docker images for argocd, nats, gpu operator, argo rollouts, argo workflows, istio, keda, etc.
* <Icon icon="square-check" iconType="regular" /> We need a domain to map to the service endpoints and certificate to encrypt the traffic. A wildcard domain like \*.services.example.com is preferred. TrueFoundry can do path based routing like `services.example.com/tfy/*`, however, many frontend applications do not support this. For certificate, check [this](#setting-up-tls-in-gcp) document for more details.
* <Icon icon="square-check" iconType="regular" /> Enough quotas for CPU/GPU instances must be present depending on your usecase. You can check and increase quotas at [GCP compute quotas](https://cloud.google.com/compute/quotas)
* <Icon icon="square-check" iconType="regular" /> Service account key creation should be allowed for the service account used by the platform.

<Tabs>
  <Tab title="New VPC and New GKE Cluster">
    1. <Icon icon="square-check" iconType="regular" /> The new VPC subnet should have a CIDR range of /24 or larger. Secondary ranges for pods (min /20) and services (min /24) are required. Secondary range can be from a non-routable range.This is to ensure capacity for \~250 instances and 4096 pods.
    2. <Icon icon="square-check" iconType="regular" /> User/serviceaccount to provision the infrastructure.
  </Tab>

  <Tab title="Existing VPC and New GKE Cluster">
    1. <Icon icon="square-check" iconType="regular" /> The existing VPC subnet should have a CIDR range of /24 or larger. Secondary ranges for pods (min /20) and services (min /24) are required. Secondary range can be from a non-routable range. This is to ensure capacity for \~250 instances and 4096 pods. Secondary ranges for pods **should be named** as `pods` and secondary ranges for services **should be named** as `services`.
    2. If your want to use a shared VPC (subnet) for the GKE cluster, ensure that you have followed this [GCP document](https://cloud.google.com/kubernetes-engine/docs/how-to/cluster-shared-vpc) and have given all the [right permissions](https://cloud.google.com/kubernetes-engine/docs/how-to/cluster-shared-vpc#summary_of_roles_granted_on_subnets) to the internal project serivceaccount
    3. <Icon icon="square-check" iconType="regular" /> The VPC should have Cloud router and cloud NAT for private subnets. Port 80 and 443 should be open for the load balancer. Allow all traffic between the subnets. Port 443, 6443, 8443, 9443 and 15017 should be allowed from the GKE control plane.
    4. <Icon icon="square-check" iconType="regular" /> User/serviceaccount to provision the infrastructure.
  </Tab>

  <Tab title="Existing GKE Cluster">
    1. <Icon icon="square-check" iconType="regular" /> GKE Version should be 1.30 or later.
    2. <Icon icon="square-check" iconType="regular" /> [NAP](https://cloud.google.com/kubernetes-engine/docs/how-to/node-auto-provisioning) should be enabled for the cluster. Ensure min and max for GPUS are also set.

    ```bash bash lines theme={"dark"}
      gcloud container clusters update CLUSTER_NAME \
          --enable-autoprovisioning \
          --autoprovisioning-resource-limits=nvidia-tesla-t4=0:256 \
          --location=REGION
    ```

    3. <Icon icon="square-check" iconType="regular" /> Workload identity should be enabled for the cluster with workload pool `project_id.svc.id.goog`.

    ```bash bash lines theme={"dark"}
    # Enable Workload Identity feature
    gcloud container clusters update CLUSTER_NAME \
        --workload-pool=PROJECT_ID.svc.id.goog \
        --region=REGION

    # Enable Workload Identity on the node pool
    gcloud container node-pools update NODE_POOL_NAME \
        --cluster=CLUSTER_NAME \
        --workload-identity-config=workload-pool=PROJECT_ID.svc.id.goog \
        --region=REGION
    ```
  </Tab>
</Tabs>

## Setting up compute plane

TrueFoundry compute plane infrastructure is provisioned using OpenTofu/Terraform. You can download the OpenTofu/Terraform code for your exact account by filling up your account details and downloading a script that can be executed on your local machine.

<Steps>
  <Step title="Enable Deployment Feature in the Platform (Optional)">
    To enable the deployment feature which allows you to deploy services through the platform, you need to enable it;

    * In the left hand navigation, go to `Settings` then `Platform Feature Visibility` under `Preferences`
    * Click on `Edit` button. Then enable the toggle for `Enable Deployment`

    <img src="https://mintcdn.com/truefoundry/bWzUilIOzt9sRNdU/images/docs/platform/enable-deployment.png?fit=max&auto=format&n=bWzUilIOzt9sRNdU&q=85&s=4932c230f6d6a6b969ed3d83c942be2b" width="1510" height="408" data-path="images/docs/platform/enable-deployment.png" />

    * Click on `Save` button.

    This will enable the deployment feature in the platform and allow you to create either a control plane and compute plane.

    <img src="https://mintcdn.com/truefoundry/bWzUilIOzt9sRNdU/images/docs/platform/deployment-platform.png?fit=max&auto=format&n=bWzUilIOzt9sRNdU&q=85&s=71e7b321682305cce46f6105c61a6eab" width="1511" height="647" data-path="images/docs/platform/deployment-platform.png" />
  </Step>

  <Step title="Choose to create a new cluster or attach an existing cluster">
    Go to the platform section in the left panel and click on `Clusters`. You can click on `Create New Cluster` or `Attach Existing Cluster` depending on your use case. Read the requirements and if everything is satisfied, click on `Continue`.

    <img src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/create-compute-plane-screenshot-1.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=b3febf85743f0b5d32adb737e23eadb6" width="3840" height="1938" data-path="images/docs/create-compute-plane-screenshot-1.png" />
  </Step>

  <Step title="Fill up the form to generate the OpenTofu/Terraform code">
    A form will be presented with the details for the new cluster to be created. Fill in with your cluster details. Click `Submit` when done

    <Tabs>
      <Tab title="Create New Cluster">
        The key fields to fill up here are:

        * `Region` - The region and availability zones where you want to create the cluster.
        * `Project ID` - The project ID where you want to create the cluster.
        * `Cluster Name` - A name for your cluster.
        * `Cluster Version` and `Master node IPv4 block` - The version of the cluster and the IPv4 block for the master nodes.
        * `Network Configuration` - Choose between `New network` or `Existing network` depending on your use case.
        * `DNS Configuration` - Configure the DNS zone and domains that will point to the cluster’s load balancer. This also provisions a TLS certificate for those domains. Select New DNS Zone or Existing DNS Zone if you want TrueFoundry to provision DNS in GCP. If you use an external DNS provider (e.g., Route53, Cloudflare), you can skip this section.
                  <img src="https://mintcdn.com/truefoundry/BTVIAjzc1bwh10GK/images/docs/platform/gcp-dns-configuration.png?fit=max&auto=format&n=BTVIAjzc1bwh10GK&q=85&s=6b3765100a3cfdf482b3fb70f10ca611" width="1359" height="548" data-path="images/docs/platform/gcp-dns-configuration.png" />
        * `GCS Bucket for OpenTofu/Terraform State` - OpenTofu/Terraform state will be stored in this bucket. It can be a preexisting bucket or a new bucket name. The new bucket will automatically be created by our script.
        * `Platform Features` - This is to decide which features like BlobStorage, ClusterIntegration, Container Registry and Secrets Manager will be enabled for your cluster. To read more on how these integrations are used in the platform, please refer to the [platform features](/docs/infrastructure/deploy-compute-plane) page.
      </Tab>

      <Tab title="Attach Existing Cluster">
        The key fields to fill up here are:

        * `Region` - The region and availability zones where you want to create the cluster.
        * `Project ID` - The project ID where you want to create the cluster.
        * `Cluster Name` - A name for your cluster.
        * `Cluster Addons` - TrueFoundry needs to install addons like ArgoCD, ArgoWorkflows, Keda, Istio, etc. Please disable the addons that are already installed on your cluster so that truefoundry installation does not overrride the existing configuration and affect your existing workloads.
        * `DNS Configuration` - Configure the DNS zone and domains that will point to the cluster’s load balancer. This also provisions a TLS certificate for those domains. Select New DNS Zone or Existing DNS Zone if you want TrueFoundry to provision DNS in GCP. If you use an external DNS provider (e.g., Route53, Cloudflare), you can skip this section.
                  <img src="https://mintcdn.com/truefoundry/BTVIAjzc1bwh10GK/images/docs/platform/gcp-dns-configuration.png?fit=max&auto=format&n=BTVIAjzc1bwh10GK&q=85&s=6b3765100a3cfdf482b3fb70f10ca611" width="1359" height="548" data-path="images/docs/platform/gcp-dns-configuration.png" />
        * `GCS Bucket for OpenTofu/Terraform State` - OpenTofu/Terraform state will be stored in this bucket. It can be a preexisting bucket or a new bucket name. The new bucket will automatically be created by our script.
        * `Platform Features` - This is to decide which features like BlobStorage, ClusterIntegration, ParameterStore, DockerRegistry and SecretsManager will be enabled for your cluster. To read more on how these integrations are used in the platform, please refer to the [platform features](/docs/infrastructure/deploy-compute-plane) page.
      </Tab>
    </Tabs>
  </Step>

  <Step title="Copy the curl command and execute it on your local machine">
    You will be presented with a `curl` command to download and execute the script. The script will take care of installing the pre-requisites, downloading OpenTofu/Terraform code and running it on your local machine to create the cluster. This will take around 40-50 minutes to complete.

    <img src="https://mintcdn.com/truefoundry/5CkapnZ7CyjQJ4bx/images/docs/how-to-deploy-your-own-cloud/gcp-compute-plane-bootstrap-script.png?fit=max&auto=format&n=5CkapnZ7CyjQJ4bx&q=85&s=f8445f54142e15cd56d89bc0356e7b9c" width="1546" height="618" data-path="images/docs/how-to-deploy-your-own-cloud/gcp-compute-plane-bootstrap-script.png" />
  </Step>

  <Step title="Verify the cluster is showing as connected in the platform">
    Once the script is executed, the cluster will be shown as connected in the platform.
  </Step>

  <Step title="Create DNS Record">
    We can get the load-balancer's IP address by going to the platform section in the bottom left panel under the Clusters section. Under the preferred cluster, you'll see the load balancer IP address under the `Base Domain URL` section.

    <img src="https://mintcdn.com/truefoundry/5CkapnZ7CyjQJ4bx/images/docs/how-to-deploy-your-own-cloud/gcp-compute-plane-load-balancer-address.png?fit=max&auto=format&n=5CkapnZ7CyjQJ4bx&q=85&s=bc7b874ffb4f7bb3ae2d56ab205e98a5" width="2804" height="524" data-path="images/docs/how-to-deploy-your-own-cloud/gcp-compute-plane-load-balancer-address.png" />

    Create a DNS record in google cloud dns or your DNS provider with the following details

    | Record Type | Record Name        | Record value              |
    | ----------- | ------------------ | ------------------------- |
    | A.          | \*.tfy.example.com | LOADBALANCER\_IP\_ADDRESS |
  </Step>

  <Step title="Start deploying workloads to your cluster">
    You can start by going [here](/docs/deploy-first-service#deploy-from-github)
  </Step>
</Steps>

## FAQ

<Accordion title="Can I use my own certificate and key files to add TLS to the load balancer?">
  If you have your own certificate files (for example, from another certificate provider or self-signed), you can use them directly with TrueFoundry.

  1. Create a Kubernetes secret with your certificate and key, or create a self-signed certificate:

       <CodeGroup>
         ```shell Shell lines theme={"dark"}
         # Generate a self-signed certificate
         openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
           -keyout tls.key -out tls.crt \
           -subj "/CN=*.example.com" \
           -addext "subjectAltName = DNS:example.com,DNS:*.example.com"
         ```
       </CodeGroup>

       <CodeGroup>
         ```shell Shell lines theme={"dark"}
         # Create secret from local certificate files
         kubectl create secret tls example-com-tls \
           --cert=path/to/cert/file \
           --key=path/to/key/file \
           -n istio-system
         ```
       </CodeGroup>

  2. Once the secret is created, head over to the cluster page and navigate to the `tfy-istio-ingress` add-on. Add the secret name in the `tfyGateway.spec.servers[1].tls.credentialName` section and ensure that `tfyGateway.spec.servers[1].port.protocol` is set to `HTTPS`. Here we are using `example-com-tls` as the secret name, which contains the certificate and key.

       <CodeGroup>
         ```yaml YAML lines theme={"dark"}
             servers:
               - <REDACTED>
               - hosts:
                   - "*"
                 port:
                   name: https-tfy-wildcard
                   number: 443
                   protocol: HTTPS
                 tls:
                   mode: SIMPLE
                   credentialName: example-com-tls
         ```
       </CodeGroup>

  <Warning>
    Self-signed certificates will cause browser warnings. They should only be used for testing or internal systems. To connect to services with self-signed certificates, you have to pass the CA certificate to verify the SSL certificate.
  </Warning>
</Accordion>
