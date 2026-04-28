> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Deploy Compute Plane

> Connect Kubernetes Cluster in your cloud account to TrueFoundry Control Plane

TrueFoundry provides an easy way to attach a new compute plane to the control plane. It provides the OpenTofu/Terraform code using which you can bring up all the necessary components of the compute plane in your cloud account. The key components of the compute plane are:

1. **Kubernetes Cluster** - This is the primary compute on which the applications deployed by the control plane run.
   This can be EKS, GKE, AKS, Openshift, Oracle Kubernetes engine or any other standard Kubernetes cluster. The control plane should have read access to the cluster configuration.
   The following addons need to be present on the compute-plane cluster depending on your usecase. **You can bring your own addons and have full flexibility on the configuration of the addons.**

<AccordionGroup>
  <Accordion title="ArgoCD (Essential)" default>
    TrueFoundry relies on [ArgoCD](https://argo-cd.readthedocs.io/en/stable/) to deploy applications to the compute-plane cluster. The infra applications are deployed in the *default* project in argocd
    while the user deployed applications are deployed in *tfy-apps* project.

    If you are **using your own ArgoCD**, please make sure of the following requirements:

    1. Ensure argocd has access to create argo applications in all namespace. For this following things must be set

    ```yaml lines theme={"dark"}
    server.extraArgs[0]="--insecure"
    server.extraArgs[1]="--application-namespaces=*"
    controller.extraArgs[0]="--application-namespaces=*"
    ```

    2. Create a *tfy-apps* project with the following spec.

    ```yaml lines theme={"dark"}
    apiVersion: argoproj.io/v1alpha1
    kind: AppProject
    metadata:
      name: tfy-apps
      namespace: argocd
    spec:
      clusterResourceWhitelist:
      - group: '*'
        kind: '*'
      destinations:
      - namespace: '*'
        server: '*'
      sourceNamespaces:
      - '*'
      sourceRepos:
      - '*'
    ```

    3. Add the following configuration to the configmap of argocd to force argocd to use `labels` instead of `annotations` for the applications.

    ```yaml lines theme={"dark"}
    configs:
      cm:
        application.instanceLabelKey: argocd.argoproj.io/instance
        application.resourceTrackingMethod: label
        controller.resource.health.persist: 'true'
    ```

    You can find the ArgoCD configuration file that Truefoundry installs by default [here](https://github.com/truefoundry/infra-charts/blob/main/charts/tfy-k8s-generic-inframold/templates/argocd.yaml).
  </Accordion>

  <Accordion title="Istio (Essential)">
    [Istio](https://istio.io/) is a really powerful service mesh and ingress controller. TrueFoundry uses Istio as the primary ingress controller in the compute-plane cluster. We don't inject the sidecar by
    default - its only injected in cases where needed for the following usecases:

    1. [Request Count Based autoscaling](/docs/autoscaling-overview#enable-autoscaling)
    2. Oauth based authentication and authorization for Jupyter Notebooks.
    3. [Intercepts](/docs/intercepts) feature to redirect / mirror traffic to other applications.
    4. [Authentication](/docs/endpoint-authentication) for services deployed on the cluster.

    If you are already using Istio in your cluster, Truefoundry should be able to work with it without any additional configuration. Truefoundry agent automatically discovers the istio gateways and exposed the domains to the control plane.

    <Note>
      Please ensure that if you have multiple Istio gateways, they do not have the same domains configured. If that is the case, then we will [need to specify which gateway to use](/docs/tfy-agent#define-which-istio-gateway-to-use-for-the-truefoundry-components) for the Truefoundry components as a variable in the tfy-agent helm chart.
    </Note>

    There are three istio components that TrueFoundry installs:

    1. **istio-base** - These are the bunch of CRDs that are required for Istio to work. You can find the argocd configuration [here](https://github.com/truefoundry/infra-charts/blob/main/charts/tfy-k8s-generic-inframold/templates/istio/istio-base.yaml).
    2. **istio-discovery** - This is pilot service that is responsible for the discovery of the services in the cluster. You can find the argocd configuration [here](https://github.com/truefoundry/infra-charts/blob/main/charts/tfy-k8s-generic-inframold/templates/istio/istio-discovery.yaml).
    3. **tfy-istio-ingress** - This is the ingress gateway that is responsible for the ingress of the services to the cluster. You can find the argocd configuration [here](https://github.com/truefoundry/infra-charts/blob/main/charts/tfy-k8s-generic-inframold/templates/istio/tfy-istio-ingress.yaml).
  </Accordion>

  <Accordion title="ArgoRollouts (Essential)">
    [Argo Rollouts](https://argoproj.github.io/argo-rollouts/) is used to power the [canary and blue-green rollout strategies](/docs/rollout-strategy) in TrueFoundry.

    If you are already using Argo Rollouts in your cluster, Truefoundry should be able to work with it without any additional configuration.

    You can find the argocd configuration [here](https://github.com/truefoundry/infra-charts/blob/main/charts/tfy-k8s-generic-inframold/templates/argo-rollouts.yaml).
  </Accordion>

  <Accordion title="Prometheus (Essential)">
    [Prometheus](https://prometheus.io/) is used to power the [metrics](/docs/monitor-your-service) feature on the platform. It also powers the [autoscaling](/docs/autoscaling-overview), [autoshutdown](/docs/scale-service-to-0) and autopilot features of the platform. TrueFoundry uses the opensource [kube-prometheus-stack](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack) for running prometheus in the cluster.

    If you are already using kube-prometheus-stack in your cluster, TrueFoundry should be able to work with it with the following configuration changes:

    ```yaml lines theme={"dark"}
      kube-state-metrics:
        metricsLabelsAllowlist:
        - pods=[truefoundry.com/application,truefoundry.com/component-type,truefoundry.com/component,truefoundry.com/application-id]
        - nodes=[karpenter.sh/capacity-type,eks.amazonaws.com/capacityType,kubernetes.azure.com/scalesetpriority,kubernetes.azure.com/accelerator,cloud.google.com/gke-provisioning,node.kubernetes.io/instance-type]
    ```

    and

    ```yaml lines theme={"dark"}
      alertmanager:
        alertmanagerSpec:
          alertmanagerConfigMatcherStrategy:
            type: None
    ```

    You can find the argocd configuration [here](https://github.com/truefoundry/infra-charts/blob/main/charts/tfy-k8s-generic-inframold/templates/prometheus.yaml).
  </Accordion>

  <Accordion title="ArgoWorkflows (Optional)">
    TrueFoundry uses [Argo Workflows](https://argo-workflows.readthedocs.io/en/latest/) to power the [Jobs feature](/docs/introduction-to-a-job) on the platform.

    If you are already using Argo Workflows in your cluster, Truefoundry should be able to work with the following configuration:

    ```yaml lines theme={"dark"}
      controller:
        workflowDefaults:
          spec:
            activeDeadlineSeconds: 432000
            ttlStrategy:
              secondsAfterCompletion: 3600
        namespaceParallelism: 1000
        parallelism: 1000
    ```

    You can find the argocd configuration [here](https://github.com/truefoundry/infra-charts/blob/main/charts/tfy-k8s-generic-inframold/templates/argo-workflows.yaml).
  </Accordion>

  <Accordion title="Keda (Optional)">
    [Keda](https://keda.sh/) is used to power the [autoscaling](/docs/autoscaling-overview#autoscaling) feature on the platform. TrueFoundry uses the opensource [keda](https://github.com/kedacore/keda) for event driver autoscaling in the cluster.

    If you are already using Keda in your cluster, TrueFoundry should be able to work without any additional configuration.

    You can find the argocd configuration [here](https://github.com/truefoundry/infra-charts/blob/main/charts/tfy-k8s-generic-inframold/templates/keda.yaml).
  </Accordion>

  <Accordion title="Victoria logs (Optional)">
    [Victoria logs](https://victoriametrics.com/products/victorialogs/) and [Vector](https://vector.dev/) are used to power the [logs](/docs/monitor-your-service) feature on the platform. This is optional and you can choose to provide your own logging solution.

    If you are already using Victoria logs in your cluster, Truefoundry should be able to work without any additional configuration. If you are already using vector to ingest logs, Truefoundry should be able to work with the following configuration:

    ```yaml lines theme={"dark"}
      customConfig:
        sinks:
          vlogs:
            type: elasticsearch
            query:
              _time_field: timestamp
              _stream_fields: namespace,pod,container,stream,truefoundry_com_application,truefoundry_com_deployment_version,truefoundry_com_component_type,truefoundry_com_retry_number,job_name,sparkoperator_k8s_io_app_name,truefoundry_com_buildName
            inputs:
              - parser
        transforms:
          parser:
            type: remap
            inputs:
              - k8s
            source: >
              if .message == "" {
                .message = " "
              }

              # Extract basic pod information

              .service = .kubernetes.container_name

              .container = .kubernetes.container_name

              .app = .kubernetes.container_name

              .pod = .kubernetes.pod_name

              .node = .kubernetes.pod_node_name

              .namespace = .kubernetes.pod_namespace

              .job_name = .kubernetes.job_name

              # Extract ALL pod labels dynamically using for_each

              pod_labels = object(.kubernetes.pod_labels) ?? {}

              # Iterate through all pod labels and add them with a prefix

              for_each(pod_labels) -> |key, value| {
                label_key = replace(replace(replace(key, ".", "_"), "/", "_"), "-", "_")
                . = set!(., [label_key], string(value) ?? "")
              }

              # Clean up kubernetes metadata

              del(.kubernetes)

              del(.file)
        sources:
          k8s:
            type: kubernetes_logs
            glob_minimum_cooldown_ms: 1000
    ```

    You can find the argocd configuration [here](https://github.com/truefoundry/infra-charts/blob/main/charts/tfy-k8s-generic-inframold/templates/tfy-logs.yaml).
  </Accordion>

  <Accordion title="GPU Operator (Optional)">
    [GPU Operator](https://artifacthub.io/packages/helm/truefoundry/tfy-gpu-operator) is used to deploy workloads on the GPU nodes. It's a TrueFoundry provided helm chart that's based on [Nvidia's GPU operator](https://github.com/NVIDIA/gpu-operator).

    If you are already using nvidia's GPU Operator in your cluster, TrueFoundry should be able to work without any additional configuration.

    You can find the argocd configuration for the following cloud providers

    1. [AWS](https://github.com/truefoundry/infra-charts/blob/main/charts/tfy-k8s-aws-eks-inframold/templates/tfy-gpu-operator.yaml)
    2. [GCP](https://github.com/truefoundry/infra-charts/blob/main/charts/tfy-k8s-gcp-gke-standard-inframold/templates/tfy-gpu-operator.yaml)
    3. [Azure](https://github.com/truefoundry/infra-charts/blob/main/charts/tfy-k8s-azure-aks-inframold/templates/tfy-gpu-operator.yaml)
    4. [Generic](https://github.com/truefoundry/infra-charts/blob/main/charts/tfy-k8s-generic-inframold/templates/tfy-gpu-operator.yaml).
  </Accordion>

  <Accordion title="Grafana (Optional)">
    [Grafana](https://grafana.com/) is a monitoring tool that can be installed to view the metrics, logs and create dashboards on the cluster. TrueFoundry doesn't direcly use grafana to power the monitoring dashboard on the platform but it is available to view additional cluster level metrics as a separate addon.

    If you are using Grafana in your cluster, you can use it for monitoring the cluster. But if you want to use the TrueFoundry provided Grafana, you can install the TrueFoundry [grafana helm chart](https://artifacthub.io/packages/helm/truefoundry/tfy-grafana) that comes with a lot of inbuilt dashboards for cluster monitoring.

    You can find the argocd configuration [here](https://github.com/truefoundry/infra-charts/blob/main/charts/tfy-k8s-generic-inframold/templates/grafana.yaml).
  </Accordion>

  <Accordion title="[AWS Only] Karpenter (Essential)">
    [Karpenter](https://karpenter.sh/) is required for supporting dynamic node provisioning on AWS EKS.

    If you are already using Karpenter in your cluster, Truefoundry should be able to work with the following additional configuration:

    1. Install [eks-node-monitoring-agent](https://github.com/aws/eks-node-monitoring-agent) helm chart.
    2. Configure Karpenter to use the eks-node-monitoring-agent.

    ```yaml lines theme={"dark"}
    settings:
      featureGates:
        nodeRepair: true
    ```

    You can find the karpenter argocd configuration [here](https://github.com/truefoundry/infra-charts/blob/main/charts/tfy-k8s-aws-eks-inframold/templates/tfy-aws/karpenter.yaml).

    We also install tfy-karpenter-config which is another helm chart that installs the [nodepools](https://karpenter.sh/docs/concepts/nodepools/) and [nodeclasses](https://karpenter.sh/docs/concepts/nodeclasses/). If you are already using Karpenter in your cluster, TrueFoundry requires following nodepool types to be present:

    | Nodepool Type    | Configuration                                                                                                                                        | Purpose                                                                                  |
    | :--------------- | :--------------------------------------------------------------------------------------------------------------------------------------------------- | :--------------------------------------------------------------------------------------- |
    | Critical         | amd64 linux on-demand nodepool with taint `class.truefoundry.com/component=critical:NoSchedule` and label `class.truefoundry.com/component=critical` | For running TrueFoundry critical workloads like prometheus, victoria-logs and tfy-agent. |
    | GPU nodepool     | amd64 linux on-demand/spot (both) with taint `nvidia.com/gpu=true:NoSchedule` and label `nvidia.com/gpu.deploy.operands=true`                        | For running user deployed GPU applications.                                              |
    | Default nodepool | amd64 linux on-demand/spot (both) without any taints                                                                                                 | For running user deployed CPU applications.                                              |

    You can find the tfy-karpenter-config argocd configuration [here](https://github.com/truefoundry/infra-charts/blob/main/charts/tfy-k8s-aws-eks-inframold/templates/tfy-aws/karpenter-config.yaml).
  </Accordion>

  <Accordion title="[AWS Only] Metrics-Server (Essential)">
    [Metrics-Server](https://github.com/kubernetes-sigs/metrics-server) is required on AWS EKS cluster for autoscaling.

    If you are already using Metrics-Server in your cluster, Truefoundry should be able to work without any additional configuration.

    You can find the argocd configuration [here](https://github.com/truefoundry/infra-charts/blob/main/charts/tfy-k8s-aws-eks-inframold/templates/metrics-server.yaml).
  </Accordion>

  <Accordion title="[AWS Only] AWS EBS CSI Driver (Essential)">
    [AWS EBS CSI Driver](https://docs.aws.amazon.com/eks/latest/userguide/ebs-csi.html) is required for supporting EBS volumes on EKS cluster.

    If you are already using AWS EBS CSI Driver in your cluster, Truefoundry should be able to work without any additional configuration. We do expect a default storage class to be present in the cluster preferrably gp3 backed by encrypted volumes.

    You can find the argocd configuration [here](https://github.com/truefoundry/infra-charts/blob/main/charts/tfy-k8s-aws-eks-inframold/templates/tfy-aws/aws-ebs-csi-driver.yaml).
  </Accordion>

  <Accordion title="[AWS Only] AWS EFS CSI Driver (Optional)">
    [AWS EFS CSI Driver](https://docs.aws.amazon.com/eks/latest/userguide/efs-csi.html) is required for supporting EFS volumes for EKS cluster.

    If you are already using AWS EFS CSI Driver in your cluster, Truefoundry should be able to work without any additional configuration. We do expect a storage class to be present in the cluster which can be used for mounting EFS volumes.

    You can find the argocd configuration [here](https://github.com/truefoundry/infra-charts/blob/main/charts/tfy-k8s-aws-eks-inframold/templates/tfy-aws/aws-efs-csi-driver.yaml).
  </Accordion>

  <Accordion title="[AWS Only] AWS Load Balancer Controller (Essential)">
    [AWS Load Balancer Controller](https://docs.aws.amazon.com/eks/latest/userguide/aws-load-balancer-controller.html) is required for supporting load balancer on EKS.

    If you are already using AWS Load Balancer Controller in your cluster, Truefoundry should be able to work without any additional configuration.

    You can find the argocd configuration [here](https://github.com/truefoundry/infra-charts/blob/main/charts/tfy-k8s-aws-eks-inframold/templates/tfy-aws/aws-load-balancer-controller.yaml).
  </Accordion>

  <Accordion title="[AWS Only] TFY Inferentia Operator (Optional)">
    [TFY Inferentia Operator](https://artifacthub.io/packages/helm/truefoundry/tfy-inferentia-operator) is required for supporting Inferentia machines on EKS.

    If you are already using Inferentia Operator in your cluster, TrueFoundry should be able to work without any additional configuration.

    You can find the argocd configuration [here](https://github.com/truefoundry/infra-charts/blob/main/charts/tfy-k8s-aws-eks-inframold/templates/tfy-aws/tfy-inferentia-operator.yaml).
  </Accordion>

  <Accordion title="Cert-Manager (Optional)">
    [Cert-Manager](https://cert-manager.io/) is required for provisioning certificates for exposing services. In AWS you can use the [AWS Certificate Manager](https://aws.amazon.com/certificate-manager/) to provision the certificates. For more details on how to setup the certificates, please refer to the [TrueFoundry documentation](/docs/add-certificate-for-tls).
  </Accordion>
</AccordionGroup>

To simplify installation, TrueFoundry packages these addons together in cloud-specific inframold charts, so you do not need to install them one by one:

* [AWS EKS Inframold (`tfy-k8s-aws-eks-inframold`)](https://artifacthub.io/packages/helm/truefoundry/tfy-k8s-aws-eks-inframold)
* [Azure AKS Inframold (`tfy-k8s-azure-aks-inframold`)](https://artifacthub.io/packages/helm/truefoundry/tfy-k8s-azure-aks-inframold)
* [GCP GKE Inframold (`tfy-k8s-gcp-gke-standard-inframold`)](https://artifacthub.io/packages/helm/truefoundry/tfy-k8s-gcp-gke-standard-inframold)
* [Generic Kubernetes Inframold (`tfy-k8s-generic-inframold`)](https://artifacthub.io/packages/helm/truefoundry/tfy-k8s-generic-inframold)

<Info>
  We have different inframold helm charts per cloud provider since there are differences in the addons required in the compute-plane depending on the cloud provider. For e.g, we have Karpenter, Inferentia Operator which are required in AWS only. Also the tolerations and affinity settings are different for each cloud provider.
</Info>

2. **Docker Registry** - This is the registry where the docker images built by the control plane are pushed. This can be ECR, GCR, ACR, Quay, Dockerhub, JFrog, Harbour, or your own docker registry.\
   Quay or any other standard docker registry. The control plane should have access to push to this registry.
3. **Blob Storage** (Optional) - This will be used to store the ML models and artifacts. This is optional and only needed if you intend to use the model registry
   feature. The control plane should have access to read and write to this blob storage. Possible integrations are AWS S3, Azure Blob Storage, GCP Storage, Minio or any other S3 compatible storage.
4. **Secret Store** (Optional) - This will be used to store the secrets for the applications. This is optional and only needed if you intend to use
   the secret store feature. The control plane should have access to read and write to this secret store. Possible integrations are AWS ParameterStore, AWS Secrets Manager, Azure Key Vault, GCP Secret Manager.
5. **DNS and Certificates** - We need to setup a domain name and point it to the load balancer in the compute-plane cluster. This will allow us to provide domain names to the workloads deployed on the cluster. The domain name can be a single domain or a wild card domain. For e.g. if you point a domain like `*.example.com` (wildcard domain) to the load balancer, the services will be exposed like `service1.example.com`, `service2.example.com` etc. However, if you point a domain like `tfy.example.com` (non-wildcard domain) to the load balancer, the services will be exposed like `tfy.example.com/service1`, `tfy.example.com/service2` etc. **Many frontend applications do not work with path based routing (the latter case) and hence we recommend using wildcard domains.** We also need to provision certificates to terminate the TLS traffic on the load balancer. For this we can use the [AWS Certificate Manager](https://aws.amazon.com/certificate-manager/) or cert-manager with GCP Cloud DNS or Azure DNS. You can also bring your pre-created certificates.

## FAQ

<AccordionGroup>
  <Accordion title="How to install in an air-gapped / restricted network environment?">
    In an air-gapped environment, the compute plane cluster cannot pull images or helm charts from the internet. You must make all artifacts available in a private registry and deploy a policy engine to rewrite image/helm references at runtime.

    ### Upload Artifacts to Your Private Registry

    First, upload all images shipped as part of your inframold installation to your private registry. You can either:

    * configure a mirror/pull-through cache for `tfy.jfrog.io`, or
    * download and upload artifacts using the script below.

    All required images are listed in an `artifacts-manifest.json` in the [infra-charts](https://github.com/truefoundry/infra-charts) repository for the inframold chart you are installing.

    Use the manifest URL that matches your inframold chart:

    | Inframold Chart                      | `artifacts-manifest.json` URL                                                                                                                                                                                                                      |
    | ------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
    | `tfy-k8s-aws-eks-inframold`          | [https://github.com/truefoundry/infra-charts/blob/main/charts/tfy-k8s-aws-eks-inframold/artifacts-manifest.json](https://github.com/truefoundry/infra-charts/blob/main/charts/tfy-k8s-aws-eks-inframold/artifacts-manifest.json)                   |
    | `tfy-k8s-azure-aks-inframold`        | [https://github.com/truefoundry/infra-charts/blob/main/charts/tfy-k8s-azure-aks-inframold/artifacts-manifest.json](https://github.com/truefoundry/infra-charts/blob/main/charts/tfy-k8s-azure-aks-inframold/artifacts-manifest.json)               |
    | `tfy-k8s-gcp-gke-standard-inframold` | [https://github.com/truefoundry/infra-charts/blob/main/charts/tfy-k8s-gcp-gke-standard-inframold/artifacts-manifest.json](https://github.com/truefoundry/infra-charts/blob/main/charts/tfy-k8s-gcp-gke-standard-inframold/artifacts-manifest.json) |
    | `tfy-k8s-generic-inframold`          | [https://github.com/truefoundry/infra-charts/blob/main/charts/tfy-k8s-generic-inframold/artifacts-manifest.json](https://github.com/truefoundry/infra-charts/blob/main/charts/tfy-k8s-generic-inframold/artifacts-manifest.json)                   |

    <Steps>
      <Step title="Download the artifacts manifest">
        ```shell Shell theme={"dark"}
        wget <link_to_artifacts_manifest_for_your_inframold_chart> -O artifacts-manifest.json
        ```
      </Step>

      <Step title="Authenticate with your target registry">
        ```shell Shell theme={"dark"}
        docker login --username <username> --password <password> <target_registry>
        ```
      </Step>

      <Step title="Set up the upload script">
        Clone the repo and install dependencies:

        ```shell Shell theme={"dark"}
        git clone https://github.com/truefoundry/infra-charts.git
        python3 -m venv venv && source venv/bin/activate
        pip install -r infra-charts/scripts/upload-artifacts/requirements.txt
        ```
      </Step>

      <Step title="Run the upload script for images and helm charts">
        Run once for images, once for helm charts:

        ```shell Shell theme={"dark"}
        # Push container images
        python infra-charts/scripts/upload-artifacts/upload_artifacts.py \
          image artifacts-manifest.json <registry_url>

        # Push helm charts
        python infra-charts/scripts/upload-artifacts/upload_artifacts.py \
          helm artifacts-manifest.json <registry_url>
        ```

        The script accepts three arguments:

        * `artifact_type` — either `image` or `helm`
        * `file_path` — path to the `artifacts-manifest.json` file
        * `destination_registry` — your private registry URL
      </Step>
    </Steps>

    <Tip>
      If you set up a mirror or pull-through cache of TrueFoundry's registry (`tfy.jfrog.io`) in your own private registry, you can skip the artifact upload step and go directly to the Kyverno setup below.
    </Tip>

    ### Configure Kyverno Policy Engine

    [Kyverno](https://kyverno.io/) is a CNCF policy engine that intercepts admission requests and rewrites image/helm repository references to point to your private registry. This is required because TrueFoundry does not currently provide a direct way to override all image URIs in the inframold chart values.

    <Steps>
      <Step title="Install Kyverno">
        Create `kyverno-values.yaml`:

        ```yaml YAML theme={"dark"}
        admissionController:
          rbac:
            clusterRole:
              extraResources:
                - apiGroups:
                    - 'argoproj.io'
                  resources:
                    - applications
                  verbs:
                    - get
                    - list
                    - watch
                    - patch
                    - update

        backgroundController:
          rbac:
            clusterRole:
              extraResources:
                - apiGroups:
                    - 'argoproj.io'
                  resources:
                    - applications
                  verbs:
                    - get
                    - list
                    - watch

        reportsController:
          rbac:
            clusterRole:
              extraResources:
                - apiGroups:
                    - 'argoproj.io'
                  resources:
                    - applications
                  verbs:
                    - get
                    - list
                    - watch
        ```

        Install the chart:

        ```shell Shell theme={"dark"}
        helm repo add kyverno https://kyverno.github.io/kyverno/
        helm repo update kyverno
        helm install kyverno kyverno/kyverno \
          --namespace kyverno \
          --create-namespace \
          -f kyverno-values.yaml
        ```
      </Step>

      <Step title="Install ArgoCD CRDs">
        These CRDs must be installed before the `tfy-kyverno-config` chart because the Kyverno policies reference ArgoCD `Application` resources for Helm repository mutation.

        ```shell Shell theme={"dark"}
        kubectl apply -f https://raw.githubusercontent.com/argoproj/argo-cd/master/manifests/crds/application-crd.yaml
        ```
      </Step>

      <Step title="Install the tfy-kyverno-config chart">
        Create `kyverno-config-values.yaml`:

        ```yaml YAML theme={"dark"}
        replaceImageRegistry:
          enabled: true
          excludeNamespaces:
            - kube-system
            - kube-public
            - kube-node-lease
          includeNamespaces: []
          registryReplacementMap:
            "*": <registry_url>

        replaceArgoHelmRepo:
          enabled: true
          excludeNamespaces: []
          includeNamespaces: []
          registryReplacementMap:
            "*": <registry_url>
        ```

        Install the chart:

        ```shell Shell theme={"dark"}
        helm install tfy-kyverno-config <registry_url>/infra-charts/tfy-kyverno-config \
          -n kyverno -f kyverno-config-values.yaml
        ```
      </Step>
    </Steps>

    Then proceed with the standard compute plane installation. Image references will be transparently redirected to your private registry by the policy engine.
  </Accordion>
</AccordionGroup>
