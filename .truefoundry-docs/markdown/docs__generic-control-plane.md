> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Deploy Generic Control Plane

> Learn how to deploy TrueFoundry's Control Plane on your generic Kubernetes cluster with detailed compute requirements and installation instructions.

## Key Components

To install the control plane on your generic Kubernetes infrastructure, you need to install the following components:

* TrueFoundry Control Plane (Shipped as a helm chart called `truefoundry`)
* PostgreSQL Database (Managed or Self-Hosted with PostgreSQL >= 13)
* Blob Storage (S3, Minio, or any other S3 compatible storage)

## Compute Requirements

TrueFoundry ships as a helm chart ([https://github.com/truefoundry/infra-charts/tree/main/charts/truefoundry](https://github.com/truefoundry/infra-charts/tree/main/charts/truefoundry)) that has configurable options to deploy the control plane according to your needs. The compute requirements change based on the scale of the number of users and requests.

Here are a few scenarios that you can choose from based on your needs.

<Tabs>
  <Tab title="Small (Dev)">
    The small tier is recommended for development purposes. Here all the components are deployed on Kubernetes and in non HA mode (single replica). This is suitable if you are just testing out the different features of TrueFoundry.

    <Warning>
      This setup brings up 1 replica of the services and is not highly-available. It
      can enable you to test the features but we do not recommend this for
      production mode.
    </Warning>

    |                    Component                   |    CPU   | Memory |                                  Storage                                  |                           Min Nodes                          |                                             Remarks                                            |
    | :--------------------------------------------: | :------: | :----: | :-----------------------------------------------------------------------: | :----------------------------------------------------------: | :--------------------------------------------------------------------------------------------: |
    | Helm-Chart <br /> *(Control Plane components)* |  4 vCPU  |   8GB  | 60GB <br /> <sub>*Persistent Volumes (Block Storage) On Kubernetes*</sub> | 2 <br /> <sub>*Pods should be spread over min 2 nodes*</sub> |                                                                                                |
    |    Postgres <br /> (Deployed on Kubernetes)    | 0.5 vCPU |  0.5GB |  5GB <br /> <sub>*Persistent Volumes (Block Storage) On Kubernetes*</sub> |                                                              | <sub>PostgreSQL version >= 13</sub> <br /> <sub>IOPS: Default (suitable for dev/testing)</sub> |
    |       Blob Storage <br /> (S3 Compatible)      |          |        |                                    20GB                                   |                                                              |                                                                                                |
  </Tab>

  <Tab title="Medium (Prod)">
    The medium tier is configured for production and will suffice teams of 10-500 members.

    |                    Component                   |   CPU   | Memory | Storage |                           Min Nodes                          |                                             Remarks                                             |
    | :--------------------------------------------: | :-----: | :----: | :-----: | :----------------------------------------------------------: | :---------------------------------------------------------------------------------------------: |
    | Helm-Chart <br /> *(Control Plane components)* | 16 vCPU |  48GB  |  250GB  | 3 <br /> <sub>*Pods should be spread over min 3 nodes*</sub> |                                                                                                 |
    |       Postgres <br /> (Managed Database)       |  2 vCPU |   4GB  |   30GB  |                                                              | <sub>PostgreSQL version >= 13</sub> <br /> <sub>IOPS: >=3000 (recommended for production)</sub> |
    |       Blob Storage <br /> (S3 Compatible)      |         |        |  500GB  |                                                              |                                                                                                 |
  </Tab>

  <Tab title="Large (Prod)">
    The large tier is configured for production and will suffice organizations of 500-50000 members.

    |                    Component                   |   CPU   | Memory | Storage |                            Min Nodes                           |                                               Remarks                                              |
    | :--------------------------------------------: | :-----: | :----: | :-----: | :------------------------------------------------------------: | :------------------------------------------------------------------------------------------------: |
    | Helm-Chart <br /> *(Control Plane components)* | 32 vCPU |  64GB  |  400GB  | 10 <br /> <sub>*Pods should be spread over min 10 nodes*</sub> |                                                                                                    |
    |       Postgres <br /> (Managed Database)       |  2 vCPU |   4GB  |   30GB  |                                                                | <sub>PostgreSQL version >= 13</sub> <br /> <sub>IOPS: >=12000 (high-performance requirement)</sub> |
    |       Blob Storage <br /> (S3 Compatible)      |         |        |  1000GB |                                                                |                                                                                                    |
  </Tab>
</Tabs>

## Prerequisites for Installation

1. **Kubernetes Cluster**: K8s cluster 1.27+

2. **Support for dynamic provisioning** of storage for PVC (such as local storage provisioners, NFS, [Longhorn](https://longhorn.io/), [OpenEBS](https://openebs.io/), or other on-premises storage solutions) and support for exposing the control plane dashboard via an ingress controller:
   * **Ingress controller** (e.g., Nginx Ingress Controller)
   * **Istio service mesh**

3. **Domain to map the ingress** of the Control Plane dashboard along with certificate for the domain.

   <Info>
     This Domain will be referred as Control Plane URL in our documentation.
   </Info>

4. **Egress Access from TrueFoundry:**
   * [https://auth.truefoundry.com](https://auth.truefoundry.com) - Central Authentication Server for licensing and authentication
   * [https://login.truefoundry.com](https://login.truefoundry.com) - login UI for the central authentication server
   * [https://catalogue.truefoundry.com](https://catalogue.truefoundry.com) - central repository for fetching catalogues for latest model, their public cost, mcp servers, etc.
   * [https://analytics.truefoundry.com](https://analytics.truefoundry.com) - analytics server for sending usage analytics to TrueFoundry.

5. **Tenant Name, Licence key, and image pull secret** - This will be given by TrueFoundry team. Make sure your organization is registered(([https://truefoundry.com/register](https://truefoundry.com/register))) on TrueFoundry.

   <Note>
     One Tenant Name and Licence key must only be used to setup one Control
     Plane. Later, switching to new tenant name and licence key would lead to
     complete data lose of existing control plane.
   </Note>

6. **PostgreSQL database** (version >= 13). We recommend managed PostgreSQL database for production environments. For instance requirements, refer to the [Compute Requirements](#compute-requirements) section.

   <Note>
     In case you do not have a managed database for testing purposes, set
     `devMode` to true in the values file to spin up a local PostgreSQL
     database.
   </Note>

7. **Blob Storage** to store artifacts (S3, Minio, or any other S3 compatible storage).

## Installation Instructions

<Steps>
  <Step title="Setup Blob Storage">
    You need to configure S3-compatible blob storage (such as Minio, AWS S3, or any other S3-compatible storage) to store artifacts, logs, and AI Gateway request logs.

    <Accordion title="Using Minio (Self-Hosted S3-Compatible Storage)">
      If you're using Minio or another S3-compatible storage solution, you'll need to provide the following configuration:

      1. **Bucket Name**: Name of the S3-compatible bucket
      2. **Region**: Region identifier (can be any string for Minio)
      3. **Access Key ID**: Access key for authentication
      4. **Secret Access Key**: Secret key for authentication
      5. **Endpoint URL**: Full URL to your S3-compatible storage endpoint

      These values will be used in the configuration steps below.
    </Accordion>

    <Info>
      Make sure your blob storage is accessible from your Kubernetes cluster and
      note down all credentials and endpoints for the next steps.
    </Info>
  </Step>

  <Step title="Setup PostgreSQL Database">
    Ensure you have a PostgreSQL database (version >= 13) set up and accessible from your Kubernetes cluster.

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

    <Warning>
      **PostgreSQL SSL Configuration:** For PostgreSQL 17+, SSL must be disabled in
      the database configuration. This can typically be done by setting SSL-related
      parameters to false or 0 in your database configuration.
    </Warning>
  </Step>

  <Step title="Add Helm Repository">
    Add the ArgoCD helm repository:

    ```bash lines theme={"dark"}
    helm repo add argo https://argoproj.github.io/argo-helm
    helm repo update argo
    ```
  </Step>

  <Step title="Install ArgoCD">
    Install ArgoCD if not already installed:

    ```bash lines theme={"dark"}
    helm install argocd argo/argo-cd \
    --namespace argocd \
    --create-namespace \
    --version 9.4.14 \
    --set server.extraArgs[0]="--insecure" \
    --set server.extraArgs[1]="--application-namespaces=*" \
    --set controller.extraArgs[0]="--application-namespaces=*" \
    --set applicationSet.enabled=false \
    --set notifications.enabled=false \
    --set dex.enabled=false
    ```
  </Step>

  <Step title="Create Kubernetes Secrets">
    We will create two secrets in this step:

    1. Store the License Key and DB Credentials
    2. Store the Image Pull Secret

    <Accordion title="Create Kubernetes Secret for License Key and DB Credentials">
      We need to create a [Kubernetes secret](https://github.com/truefoundry/infra-charts/blob/main/charts/truefoundry/README.md#using-k8s-secret-for-required-fields) containing the licence key and database credentials that you configured in the previous step.

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

  <Step title="Create HelmChart Values File">
    Create a values file as given below and replace the following values:

    * **Control Plane URL**: URL that you will map to the control plane dashboard (e.g., `https://truefoundry.example.com`)
    * **Tenant Name**: Tenant name provided by TrueFoundry team
    * **Blob Storage Configuration**: Use the bucket name, region, access keys, and endpoint URL from the "Setup Blob Storage" step above. Depending on the storage you can chose your `defaultCloudProvider` as `aws`, `gcp`, `azure` etc
    * You can disable any addons that you don't need by setting the `enabled` flag to `false`

    ```yaml truefoundry-values.yaml lines theme={"dark"}
    tenantName: ""
    controlPlaneURL: ""
    clusterName: ""

    tolerations: []
    affinity: {}

    argocd:
      enabled: true
    argoWorkflows:
      enabled: true
    argoRollouts:
      enabled: true
    certManager:
      enabled: true
    metricsServer:
      enabled: false # enable if you don't have metrics server already enabled
    gpu:
      enabled: true

    truefoundry:
      enabled: true
      devMode:
        enabled: false # enable for dev mode
      virtualservice:
        enabled: false
        hosts: []
        gateways: []
        annotations: {}
      existingTruefoundryCredsSecret: ""
      imagePullSecrets: []

      defaultCloudProvider: ""
      storageConfiguration: {}

      tolerations: []
      affinity: {}

    tfyLogs:
      enabled: true
    istio:
      enabled: true
      tfyGateway:
        httpsRedirect: false
        tls:
          enabled: false # make this true for encrypted traffic. Ensure you have TLS certs as k8s secrets in istio-system namespace
          mode: SIMPLE
          credentialName: "" # secret name containing TLS certificates
        domains: ["*"]

    keda:
      enabled: true
    kubecost:
      enabled: true

    prometheus:
      enabled: true
      alertmanager:
        config:
          receivers:
            - name: tfy-agent
              webhook_configs:
                - url: http://tfy-agent.tfy-agent.svc.cluster.local:3000/alerts
                  send_resolved: true
      config:
        enabled: true

    tfyAgent:
      enabled: false
    elasti:
      enabled: true
    helm:
      resourcePolicy: keep
    ```
  </Step>

  <Step title="Install TrueFoundry Control Plane">
    Install the TrueFoundry control plane using the OCI helm chart:

    ```bash lines theme={"dark"}
    helm upgrade --install tfy-k8s-generic-inframold oci://tfy.jfrog.io/tfy-helm/tfy-k8s-generic-inframold \
    -n argocd --create-namespace -f truefoundry-values.yaml
    ```
  </Step>

  <Step title="Configure Load Balancer and DNS">
    1. Get the IP address of the load balancer:

    ```bash lines theme={"dark"}
    kubectl get svc -n istio-system
    ```

    2. Point your control plane domain to this IP address in your DNS provider.

    3. Configure [TLS certificates](/docs/add-certificate-for-tls) for the load balancer to terminate TLS traffic.
  </Step>

  <Step title="Verify Installation">
    1. Wait for all pods to be running:

    ```bash lines theme={"dark"}
    kubectl get pods -n argocd
    kubectl get pods -n truefoundry
    ```

    2. Access the control plane at your configured URL and login with the same credentials used to register the tenant.
  </Step>

  <Step title="Start deploying workloads to your cluster">
    You can start by going [here](https://docs.truefoundry.com/docs/deploy-first-service#deploy-from-github)
  </Step>
</Steps>

## Setting Up Compute Plane

After installing the control plane, you need to set up a compute plane where your applications will run.

<Note>
  **Compute Plane Installation**

  For complete instructions on setting up the compute plane on your generic cluster, please refer to the [Generic Compute Plane Setup](/docs/infrastructure/generic-compute-plane-setup) guide.
</Note>

<Warning>
  **Important: Setting Up Control Plane as a Compute Plane**

  When installing the control plane, you need to **set up the control plane cluster as a compute plane** managed by TrueFoundry itself. This allows you to manage control plane upgrades and configuration through the TrueFoundry UI.

  To enable this, you **MUST disable all add-ons** to avoid conflicts with existing cluster components. The add-ons that should be disabled include:

  * **ArgoCD** - Already installed separately in Step 2
  * **Argo Workflows** - Conflicts with control plane installation
  * **Argo Rollouts** - Conflicts with control plane installation
  * **Istio** - Conflicts if already present
  * **Cert Manager** - Conflicts if already present
  * **Prometheus** - Conflicts if already present
  * **KEDA** - Conflicts if already present

  These add-ons will be automatically disabled when you use the TrueFoundry chart for control plane installation (they are not included in the values file above).
</Warning>

# FAQ

<AccordionGroup>
  <Accordion title="Can I use my Artifactory as a mirror to pull images?">
    Yes. You can configure your Artifactory to mirror our registry.

    <Note>
      Credentials for accessing the TrueFoundry private registry are required and
      will be provided during onboarding.
    </Note>

    As applications created in the control plane are deployed in multiple namespace, you need to create the image pull secret in all the namespaces.

    Update Helm values as follows:

    ```yaml wrap lines theme={"dark"}
    registry: <YOUR_REGISTRY> # Replace with your registry

    imagePullSecrets:
      - name: <IMAGE_PULL_SECRET_NAME> # Replace with your image pull secret name
    ```

    Following script you can use to create the image pull secret in all the namespaces

    ```bash wrap lines theme={"dark"}
    NAMESPACES=(argocd prometheus truefoundry keda istio-system tfy-agent cert-manager argo-workflows argo-rollouts tfy-logs kubecost elasti tfy-gpu-operator)

    for ns in "${NAMESPACES[@]}"; do kubectl create namespace "$ns"; kubectl create secret docker-registry <IMAGE_PULL_SECRET_NAME> --namespace="$ns" --docker-server=<YOUR_REGISTRY> --docker-username=<USERNAME> --docker-password=<PASSWORD> --docker-email=<EMAIL>; done
    ```

    <Note>
      * Replace `<IMAGE_PULL_SECRET_NAME>` with your image pull secret name.

      * Replace `<YOUR_REGISTRY>` with your registry URL. Ensure you are also adding your namespace as TrueFoundry will preserve the namespace structure of the images.

      * Replace `<USERNAME>` with your registry username.

      * Replace `<PASSWORD>` with your registry password.

      * Replace `<EMAIL>` with your registry email.
    </Note>
  </Accordion>

  <Accordion title="Can I copy images to my own private registry?">
    Yes, you can copy the images to your own private registry. To get the entire list of images and instructions to push the images to your registry, you can refer to the [air-gapped installation FAQ](/docs/platform/deploy-control-plane-and-gateway-plane#how-to-install-in-an-air-gapped--restricted-network-environment) in the control plane installation guide.

    The script maintains the namespacing structure of the images in the destination registry. For example, if image is `quay.io/argoproj/argo-rollouts:v1.8.2`, it will be pushed to `<YOUR_REGISTRY>/<NAMESPACE>/argoproj/argo-rollouts:v1.8.2` where `<NAMESPACE>` is the namespace of the image in which you want to push the images.

    Once the images are pushed to your registry, you can update the Helm values as follows:

    ```yaml wrap lines theme={"dark"}
    registry: <YOUR_REGISTRY>

    imagePullSecrets:
      - name: <IMAGE_PULL_SECRET_NAME>
    ```

    Following script you can use to create the image pull secret in all the namespaces

    ```bash wrap lines theme={"dark"}
    NAMESPACES=(argocd prometheus truefoundry keda istio-system tfy-agent cert-manager argo-workflows argo-rollouts tfy-logs kubecost elasti tfy-gpu-operator)

    for ns in "${NAMESPACES[@]}"; do kubectl create namespace "$ns"; kubectl create secret docker-registry <IMAGE_PULL_SECRET_NAME> --namespace="$ns" --docker-server=<YOUR_REGISTRY> --docker-username=<USERNAME> --docker-password=<PASSWORD> --docker-email=<EMAIL>; done
    ```

    <Note>
      * Replace `<IMAGE_PULL_SECRET_NAME>` with your image pull secret name.

      * Replace `<YOUR_REGISTRY>` with your registry URL. Ensure you are also adding your namespace as TrueFoundry will preserve the namespace structure of the images.

      * Replace `<USERNAME>` with your registry username.

      * Replace `<PASSWORD>` with your registry password.

      * Replace `<EMAIL>` with your registry email.
    </Note>
  </Accordion>

  <Accordion title="How to use SSO directly without using TrueFoundry Auth Server?">
    By default, the control plane uses the TrueFoundry Auth Server for user authentication. However, you can configure it to use your own external identity provider instead. We support both OIDC and SAML-compliant identity providers. [Read more](https://www.truefoundry.com/docs/deploy-control-plane-with-external-oauth)
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
    # Server CA only (e.g. AWS RDS)
    kubectl create secret generic db-ssl-certs \
      --from-file=ca-certificate.crt=/path/to/your/ca-certificate.crt \
      -n truefoundry

    # Full mTLS (e.g. GCP Cloud SQL, Azure Database for PostgreSQL)
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
</AccordionGroup>
