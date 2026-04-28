> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Deploy Helm Charts

> Complete guide to deploying Helm charts in TrueFoundry with support for multiple repository types, Kustomize integration, and advanced configurations

TrueFoundry provides a way to deploy Helm charts directly through the platform, making it easy to deploy any Helm chart without needing to use kubectl or other command-line tools. It also helps keep track of the helm charts that are deployed along with their versions and change history.

Truefoundry also provides the following additional features for helm charts deployment:

1. **Support for multiple repository types**: Truefoundry allows deploying a Helm chart from public/private helm repository, OCI registry, and or your own Git repository.
2. **Support for Kustomize patches and additional manifests**: Truefoundry allows you to add Kustomize patches and additional manifests to the Helm chart deployment. This is specially useful in case the helm chart doesn't allow you to customize the values you need.
3. **Support for secrets management**: While installing helm charts, we often need to create kubernetes secrets manually and then refer to them in the helm chart. Truefoundry makes this process secure by allowing you to create the secrets in Truefoundry(on your secret manager) and then add a kubernetes secret manifest via Kustomize to the helm chart deployment. You can read more on this in the steps below.
4. **Validation for cluster-scoped objects**: Helm-charts downloaded from the internet can have malicious code in them, which can impact the security of your cluster. Truefoundry validates the helm chart by checking for any cluster-scoped objects in the helm chart.

<Note>
  Cluster-scoped objects (like ClusterRole, ClusterRoleBinding, or non-namespace
  scoped resources) cannot be applied when deploying Helm charts unless you have
  cluster admin privileges. This restriction prevents workspace users from
  creating cluster-level resources that could impact other workloads.
</Note>

5. **Pause Helm Chart**: Truefoundry allows you to pause the helm chart deployment. This is useful in case you want to stop the deployment and resume it later to save cost. Pausing the helm chart will scale all the pods in the helm chart to 0.

## Step-by-Step Deployment Guide

<Steps>
  <Step title="Navigate to Deployments">
    Log in to your TrueFoundry dashboard and click on **Deployments** in the
    left sidebar, then click **New** to create a new deployment.

    <img src="https://mintcdn.com/truefoundry/WsXGoMBEEqYmepmC/images/new_deployment_button.png?fit=max&auto=format&n=WsXGoMBEEqYmepmC&q=85&s=6ca438319dda9b0df29e2af73fdd105a" alt="TrueFoundry Deployments page showing the New button" width="2696" height="232" data-path="images/new_deployment_button.png" />
  </Step>

  <Step title="Select Helm">
    In the application type page, click on **show advanced** and select **Helm**.

    <img src="https://mintcdn.com/truefoundry/2McLD1HKh8epqCAW/images/helm_deployment_option.png?fit=max&auto=format&n=2McLD1HKh8epqCAW&q=85&s=37208e979da68e774c555e9f4ba5e395" alt="Deploy new Helm modal showing Helm option selected" width="2664" height="1468" data-path="images/helm_deployment_option.png" />
  </Step>

  <Step title="Choose Your Chart Source">
    Now you need to tell TrueFoundry where your Helm chart is located. You have three options:

    <Frame caption="Helm Chart Deployment Form">
      <img src="https://mintcdn.com/truefoundry/RotK-PRWkDa8_6SL/images/helm_repo_types.png?fit=max&auto=format&n=RotK-PRWkDa8_6SL&q=85&s=adfab23cdd1bda4967a84c6a47e87046" alt="TrueFoundry Helm chart deployment form showing the main configuration interface" width="1774" height="606" data-path="images/helm_repo_types.png" />
    </Frame>

    <br />

    <AccordionGroup>
      <Accordion title="Option 1: Public Helm Repository" icon="folder">
        You can use this option to deploy charts from public helm repositories like Bitnami, Helm Charts, etc.

        <Frame caption="HelmRepo Configuration">
          <img src="https://mintcdn.com/truefoundry/RotK-PRWkDa8_6SL/images/helm_repo_page.png?fit=max&auto=format&n=RotK-PRWkDa8_6SL&q=85&s=05b60964729140599032138c6b402519" alt="HelmRepo configuration showing repository URL, chart name, and version selection" width="1802" height="1276" data-path="images/helm_repo_page.png" />
        </Frame>

        **What you need to fill:**

        * `Helm repository URL`: The URL of the chart repository (e.g., `https://charts.bitnami.com/bitnami`)
        * `Chart name`: The name of the chart (e.g., `redis`, `postgresql`)
        * `Version`: The specific version you want to deploy

        **Example:** Deploy Redis from Bitnami

        ```yaml lines theme={"dark"}
        source:
          type: helm-repo
          repo_url: https://charts.bitnami.com/bitnami
          chart: redis
          version: 22.0.7
        ```
      </Accordion>

      <Accordion title="Option 2: Container Registry (OCI)" icon="server">
        You can use this option to deploy charts from container registries like Docker Hub, Google Container Registry, etc.

        <Info>
          To deploy charts from private container registries, you have to add the container registry as an integration first. To do this, go to **Integrations** → **Add Integration Provider** and select your registry provider and add an integration for your registry. You can refer to the [Integrations Overview](/docs/integrations-overview) guide for more details on how to add integrations.
        </Info>

        <Frame caption="OCIRepo Configuration">
          <img src="https://mintcdn.com/truefoundry/zwfmtH9v9mtcuJm4/images/oci_helm_repo.png?fit=max&auto=format&n=zwfmtH9v9mtcuJm4&q=85&s=2c837800355f9822e1550436732c7d53" alt="OCIRepo configuration showing OCI chart URL and version fields" width="1812" height="1364" data-path="images/oci_helm_repo.png" />
        </Frame>

        **What you need to fill:**

        * `OCI chart URL`: The OCI URL of your chart (e.g., `oci://registry-1.docker.io/bitnamicharts/redis`)
        * `Version`: The specific version of the chart
        * `Container Registry`: This is only needed if you are deploying helm chart from your private container registry. Toggle the **Show Advanced Fields** and select the integration from the dropdown which contains your helm charts.

        **Example:** Deploy Redis from Docker Hub

        ```yaml lines theme={"dark"}
        source:
          type: oci-repo
          oci_chart_url: oci://registry-1.docker.io/bitnamicharts/redis
          version: 22.0.7
        ```
      </Accordion>

      <Accordion title="Option 3: Git Repository" icon="file">
        You can use this option to deploy charts from your own Git repositories.

        <Info>
          To deploy charts from private Git repositories, you have to create a repository secret first, you can refer to the [Private Repository Configuration](/docs/deploy-helm-charts#github-private-repository) section on how to do that.
        </Info>

        <Frame caption="GitHelmRepo Configuration">
          <img src="https://mintcdn.com/truefoundry/tTsoiVRuN3baNHEd/images/git_helm_repo.png?fit=max&auto=format&n=tTsoiVRuN3baNHEd&q=85&s=8217887360c861f65b8ecb92dc374675" alt="GitHelmRepo configuration showing Git repository URL, revision, and path fields" width="1802" height="1376" data-path="images/git_helm_repo.png" />
        </Frame>

        **What you need to fill:**

        * `Git repository URL`: The URL of your Git repository
        * `Revision`: Branch, tag, or commit SHA to use (e.g., `main`, `v1.0.0`)
        * `Path`: Path to the chart within the repository (e.g., `charts/my-app`)

        **Example:** Deploy from your own Git repo

        ```yaml lines theme={"dark"}
        source:
          type: git-helm-repo
          git_repo_url: https://github.com/your-org/helm-charts.git
          revision: main
          path: charts/redis
        ```
      </Accordion>
    </AccordionGroup>
  </Step>

  <Step title="Configure Your Application">
    This is where you can update the values file, add kustomize patches and additional manifests.

    <AccordionGroup>
      <Accordion title="Basic Values Configuration" icon="file">
        You can override the default values of the helm chart by updating the values in values block. This is useful if you want to change the default values of the helm chart.

        <Frame caption="Values Configuration">
          <img src="https://mintcdn.com/truefoundry/OP229afdEVxDmaze/images/values_block.png?fit=max&auto=format&n=OP229afdEVxDmaze&q=85&s=83c95dac498c26aa707de3a34a16c2fb" alt="Values configuration section showing options for values files and block files" width="1782" height="1370" data-path="images/values_block.png" />
        </Frame>

        **Example:** Configure Redis with a password and persistent storage.

        ```yaml lines theme={"dark"}
        values:
          auth:
            enabled: true
            password: "my-secure-password"
          master:
            persistence:
              enabled: true
              size: 8Gi
          service:
            type: LoadBalancer
            ports:
              redis: 6379
        ```
      </Accordion>

      <Accordion title="Using Secrets (Recommended)" icon="link">
        Instead of putting passwords directly in your configuration, you can use TrueFoundry's secure secret management.

        **How to use secrets:**

        1. First, create a secret in TrueFoundry, you can refer to the [Secret Management](/docs/manage-secrets#creating-and-managing-secrets) page for more details on how to create a secret.
        2. Copy the secret's FQN (it looks like `tfy-secret://truefoundry:secret-name:key`)
        3. Use it in your secret manifest instead of plain text password.

        <Info>
          Truefoundry secrets are supported only in manifest block that too only for kubernetes secret manifest which are using `stringData` field to store the secret.
        </Info>

        <Frame caption="Secrets Configuration">
          <img src="https://mintcdn.com/truefoundry/b0zck08nRIZEm4rs/images/secret_in_manifest.png?fit=max&auto=format&n=b0zck08nRIZEm4rs&q=85&s=2362196b494a021d42f84ef13c4eeb7d" alt="Secrets configuration section showing options for secrets" width="1810" height="1494" data-path="images/secret_in_manifest.png" />
        </Frame>

        **Example:** Using a secret for Redis password, as you can see in the above image, the password is referenced using the secret FQN.

        ```yaml lines theme={"dark"}
        # Secret to create a redis password secret
        apiVersion: v1
        kind: Secret
        metadata:
          name: redis-secret
        type: Opaque
        stringData:
          redis-password: tfy-secret://truefoundry:redis-secrets:password
        ```

        <Info>
          This is much more secure than putting passwords directly in your configuration files!
        </Info>
      </Accordion>

      <Accordion title="Advanced Customization (Optional)" icon="route">
        **When to use this:** If you need to modify the Kubernetes resources that Helm creates, or add additional resources.

        **Two main options:**

        1. **Kustomize Patches**: Modify existing resources (e.g., add annotations, change resource limits), you can refer to the [Kustomize](/docs/kustomize) page for more details on how to use kustomize patches.

        <Frame caption="Kustomize Configuration">
          <img src="https://mintcdn.com/truefoundry/RotK-PRWkDa8_6SL/images/kustomize_for_helm_page.png?fit=max&auto=format&n=RotK-PRWkDa8_6SL&q=85&s=8402cb8f37f3bd49ee0a045f76830e59" alt="Kustomize configuration showing patch and additional manifests options" width="1800" height="1364" data-path="images/kustomize_for_helm_page.png" />
        </Frame>

        2. **Additional Manifests**: Add new Kubernetes resources (e.g., To expose your app with a VirtualService, you can add a VirtualService manifest here.)

        <Frame caption="Additional Manifests Configuration">
          <img src="https://mintcdn.com/truefoundry/g9bibSt_o82HsdND/images/additional_manifest_example.png?fit=max&auto=format&n=g9bibSt_o82HsdND&q=85&s=ebb6910aade8dcddcb1f976d5eadb8ae" alt="Additional manifests configuration showing options for additional manifests" width="1810" height="1364" data-path="images/additional_manifest_example.png" />
        </Frame>

        **Example:** Expose Redis with a custom endpoint

        ```yaml lines theme={"dark"}
        # Additional manifest to expose Redis endpoint
        apiVersion: networking.istio.io/v1beta1
        kind: VirtualService
        metadata:
          name: redis-virtualservice
          namespace: your-workspace
        spec:
          hosts:
            - <your-custom-domain>
          gateways:
            - <istio-gateway-name>/<istio-gateway-namespace>
          tcp:
            - match:
                - port: 6379
              route:
                - destination:
                    host: redis-redis-master.<your-workspace-name>.svc.cluster.local
                    port:
                      number: 6379
        ```
      </Accordion>
    </AccordionGroup>
  </Step>

  <Step title="Deploy and Monitor">
    Click **Submit** to deploy your chart. TrueFoundry will:

    1. Download your chart
    2. Apply your configuration
    3. Create the necessary Kubernetes resources
    4. Show you the deployment status

    <Frame caption="Deployment Status">
      <img src="https://mintcdn.com/truefoundry/yI0uagtNkPQEVgMd/images/deploy_helm_chart_form.png?fit=max&auto=format&n=yI0uagtNkPQEVgMd&q=85&s=0cd2e3551d78d88941931e197de28c45" alt="Deployment status showing the deployment status" width="1816" height="1494" data-path="images/deploy_helm_chart_form.png" />
    </Frame>

    You can monitor the deployment progress and view logs by clicking on your deployment in the deployments list.
  </Step>
</Steps>

## Private Repository Configuration

TrueFoundry allows you to deploy Helm charts from private repositories by configuring repository credentials using the Kubernetes manifest deployment feature or by adding repository integrations.

<AccordionGroup>
  <Accordion title="GitHub Private Repository" icon="file">
    Configure access to private GitHub repositories containing Helm charts.

    <Steps>
      <Step title="Add GitHub as Integration">
        First, add your private GitHub repository as an integration in TrueFoundry, you can refer to the [Github Integration](/docs/github-integration-set-up) guide for more details on how to add an github repository as an integration.
      </Step>

      <Step title="Create GitHub Personal Access Token">
        First, create a personal access token in GitHub:

        1. Go to [https://github.com/settings/tokens](https://github.com/settings/tokens) (click on "Generate new token (classic)")
        2. Select scopes: `repo` (for private repositories)
        3. Copy the generated token

        <Warning>
          Store your token securely. You won't be able to see it again after
          creation.
        </Warning>
      </Step>

      <Step title="Deploy Repository Secret using Kubernetes Manifest">
        Use TrueFoundry's Kubernetes manifest deployment to create the repository
        secret. Follow the [Deploy Kubernetes
        Manifests](/docs/deploy-kubernetes-manifests) guide and deploy this manifest:

        ```yaml lines theme={"dark"}
        apiVersion: v1
        kind: Secret
        type: Opaque
        metadata:
          name: repo-1752233801
          labels:
            argocd.argoproj.io/secret-type: repository
          namespace: argocd
        stringData:
          url: https://github.com/your-org/helm-charts.git
          type: git
          project: tfy-apps
          password: github_pat_11BAP2UBQ0tDSEVgbtphEM_18bH8fYWa6qhGCIWZsZIeIMm5NAEKYUIOfdfdfdfdfHER7RDDfbTAa
          username: x-access-token
        ```

        **Before deploying, replace:**

        * `https://github.com/your-org/helm-charts.git` with your actual repository URL
        * `github_pat_11BAP2UBQ0tDSEVgbtphEM_18bH8fYWa6qhGCIWZsZIeIMm5NAEKYUIOfdfdfdfdfHER7RDDfbTAa` with your actual GitHub token
        * `repo-1752233801` with a unique name for your secret

        <Frame caption="Deployed Secret">
          <img src="https://mintcdn.com/truefoundry/g9bibSt_o82HsdND/images/argo_repo_secret_deployment.png?fit=max&auto=format&n=g9bibSt_o82HsdND&q=85&s=29d5ce69e7bbd570f42c933ae6992b67" alt="Deployed secret showing the secret created in the argocd namespace" width="1804" height="1492" data-path="images/argo_repo_secret_deployment.png" />
        </Frame>
      </Step>

      <Step title="Deploy Helm Chart from Private Repository">
        Now you can deploy Helm charts from your private repository:

        1. Go to **Deployments** → **New** → **Helm**
        2. Select **Git Repository** as your chart source
        3. Enter your private repository URL
        4. Specify the chart path and revision
        5. Configure your values and deploy
      </Step>
    </Steps>
  </Accordion>
</AccordionGroup>

## Complete Example: Deploying Redis

Let's walk through a real example of deploying Redis (a popular caching database) with proper configuration and security:

<AccordionGroup>
  <Accordion title="Step 1: Choose Redis Chart" icon="folder">
    First, we'll tell TrueFoundry to use the Redis chart from Bitnami (a popular chart repository):

    <Frame caption="Redis Chart Configuration">
      <img src="https://mintcdn.com/truefoundry/KZ9SZ6S9x8hzKvZy/images/redis_info.png?fit=max&auto=format&n=KZ9SZ6S9x8hzKvZy&q=85&s=7d3dee171a73d1e3d8647748cec1d6f8" alt="Redis chart configuration showing the chart name and version" width="1780" height="872" data-path="images/redis_info.png" />
    </Frame>

    This tells TrueFoundry: "Use the Redis chart from Bitnami's repository, version 22.0.5"
  </Accordion>

  <Accordion title="Step 2: Set Up Secure Password" icon="link">
    Instead of putting a password directly in our configuration, we'll use TrueFoundry's secure secret management:

    1. **Create a secret in TrueFoundry:**

       * Go to **Secrets** in the left sidebar
       * Create a new secret called `redis-secrets`
       * Add a key called `password` with your desired password

           <Frame caption="Creating Redis Password Secret">
             <img src="https://mintcdn.com/truefoundry/b0zck08nRIZEm4rs/images/secret_in_manifest.png?fit=max&auto=format&n=b0zck08nRIZEm4rs&q=85&s=2362196b494a021d42f84ef13c4eeb7d" width="1810" height="1494" data-path="images/secret_in_manifest.png" />
           </Frame>

    2. **Copy the secret reference:** It will look like `tfy-secret://truefoundry:redis-secrets:password`
  </Accordion>

  <Accordion title="Step 3: Add additional manifests to create a redis password secret" icon="file">
    Now we'll configure how Redis should run - with a password, persistent storage, and external access:

    ```yaml lines theme={"dark"}
    # Additional manifest to create a redis password secret
    apiVersion: v1
      kind: Secret
      metadata:
        name: redis-secret
      type: Opaque
      data:
        redis-password: <base64-encoded-password>
    ```

    **Before deploying, replace:**

    * `redis-secret` with a unique name for your secret
    * `redis-password` with your desired password

    <Frame caption="Creating Redis Password Secret">
      <img src="https://mintcdn.com/truefoundry/b0zck08nRIZEm4rs/images/secret_in_manifest.png?fit=max&auto=format&n=b0zck08nRIZEm4rs&q=85&s=2362196b494a021d42f84ef13c4eeb7d" width="1810" height="1494" data-path="images/secret_in_manifest.png" />
    </Frame>
  </Accordion>

  <Accordion title="Step 4: Configure Redis Settings" icon="file">
    Now we'll configure how Redis should run - with a password, persistent storage, and external access:

    ```yaml lines theme={"dark"}
    # Redis configuration with security and persistence
    values:
      auth:
        enabled: true
        existingSecret: redis-secret
        existingSecretPasswordKey: redis-password

      master:
        service:
          type: LoadBalancer  # This makes Redis accessible from outside
          ports:
            redis: 6379
        persistence:
          enabled: true       # This saves data even if Redis restarts
          size: 8Gi

      replica:
        replicaCount: 2       # Run 2 backup Redis instances
        service:
          type: ClusterIP     # Internal access only
        persistence:
          enabled: true
          size: 4Gi

      metrics:
        enabled: true         # Enable monitoring
        serviceMonitor:
          enabled: true
    ```

    **What this does:**

    * Sets up a secure password using our secret
    * Makes Redis accessible from outside (LoadBalancer)
    * Saves data permanently (persistence)
    * Runs backup instances for reliability
    * Enables monitoring
  </Accordion>

  <Accordion title="Step 5: Custom Endpoint (Optional)" icon="globe">
    If you want to access Redis through a custom URL (like `redis.your-app.com`), you can add this advanced configuration:

    ```yaml lines theme={"dark"}
    # Custom endpoint configuration
    apiVersion: networking.istio.io/v1beta1
    kind: VirtualService
    metadata:
      name: redis-virtualservice
      namespace: your-workspace
    spec:
      hosts:
        - redis.your-workspace.ml.demo.truefoundry.cloud
      gateways:
        - istio-system/tfy-wildcard-alb
      tcp:
        - match:
            - port: 6379
          route:
            - destination:
                host: redis-redis-master.your-workspace.svc.cluster.local
                port:
                  number: 6379
    ```

    **When to use this:** Only if you need a custom domain name for your Redis instance.
  </Accordion>
</AccordionGroup>

<Frame caption="Final Redis Deployment Result">
  <img src="https://mintcdn.com/truefoundry/7DvUItQJFvQFavq_/images/redis_deployment.png?fit=max&auto=format&n=7DvUItQJFvQFavq_&q=85&s=ac50e2b9b12c018c91a42a73d5fed3fd" alt="Final Redis deployment showing the complete deployment with all configurations, Kustomize patches, and VirtualService applied" width="2662" height="1342" data-path="images/redis_deployment.png" />
</Frame>
