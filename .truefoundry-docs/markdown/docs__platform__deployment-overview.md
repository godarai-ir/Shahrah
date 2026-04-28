> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Installation and Deployment Overview

> Description of the different deployment options for Truefoundry.

Truefoundry provides different deployment options depending on your cloud provider, existing infrastructure and the components that you want to deploy. Before understanding the different deployment options, its important to read about the overall architecture here: [control-plane architecture](/docs/platform/control-plane-architecture), [compute-plane architecture](/docs/platform/compute-plane-architecture), [gateway-plane architecture](/docs/platform/gateway-plane-architecture).

## Deployment Options

| Deployment Scenario                                                                 | Installation                                                                                    | Remarks                                                                                          | Cost of Hosting                                                                                         | Pricing Tier |
| ----------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------- | ------------ |
| **1. AI Gateway SAAS only**<br />Gateway                                            | No Installation Needed                                                                          | Use managed AI Gateway for LLM requests, no compute plane or control plane hosting needed.       | 0                                                                                                       | Starter      |
| **2. Gateway Plane only**<br />Gateway                                              | [Deploy Gateway Plane](/docs/platform/deploy-gateway-plane)                                     | Host the Gateway plane for LLM requests and use control-plane hosted by Truefoundry.             | \~\$600/month                                                                                           | Enterprise   |
| **3. Control Plane + Gateway Plane**<br />AI Gateway                                | [Deploy Control Plane and Gateway Plane](/docs/platform/deploy-control-plane-and-gateway-plane) | Use only AI Gateway, full control, suitable if you don't need compute plane/app deployment.      | \~\$800-1000/month                                                                                      | Enterprise   |
| **4. Compute Plane + AI Gateway SAAS**<br />AI Deployment, AI Gateway               | [Deploy Compute Plane](/docs/infrastructure/deploy-compute-plane)                               | Deploy models/services (Compute Plane) on your own infrastructure while using managed AI Gateway | Fixed cost of \~\$200/month. Cost scales as compute plane scales.                                       | Pro          |
| **5. Compute Plane only**<br />AI Deployment                                        | [Deploy Compute Plane](/docs/infrastructure/deploy-compute-plane)                               | Host the compute plane for model/services deployment, control-plane hosted by Truefoundry.       | Fixed cost of \~\$200/month. Cost scales as compute plane scales.                                       | Pro          |
| **6. Compute Plane + Control Plane**<br />AI Deployment                             | [Deploy Control and Compute Plane](/docs/deploy-control-and-compute-plane)                      | Complete self-hosted for AI Deployment, full control over infra, no Gateway product.             | Fixed cost \~\$600/month. Cost scales as compute plane scales.                                          | Enterprise   |
| **7. Control Plane + Gateway Plane + Compute Plane**<br />AI Deployment, AI Gateway | [Deploy Control and Compute Plane](/docs/deploy-control-and-compute-plane)                      | Full flexibility, full control; self-host the entire platform.                                   | Fixed cost \~\$1200/month. Cost scales as the compute plane scales and requests through gateway scales. | Enterprise   |

The key deployment options described above are:

<AccordionGroup>
  <Accordion title="Option 1: AI Gateway SAAS only">
    This is a fully managed solution on Truefoundry's secure cloud infrastructure with enterprise-grade features. You don't need to deploy any infrastructure on your end. This includes usage of only the AI Gateway modle and not AI Deployment.

    <Frame caption="Truefoundry Managed SAAS">
      <img src="https://mintcdn.com/truefoundry/41rnemFLxWdEyTu9/images/docs/platform/ai-gateway-saas-only.png?fit=max&auto=format&n=41rnemFLxWdEyTu9&q=85&s=f701ce2d4ae21718df4a400a7f8f5511" alt="Truefoundry Managed SAAS Gateway Only" width="4225" height="2536" data-path="images/docs/platform/ai-gateway-saas-only.png" />
    </Frame>

    <Tip>
      This is ideal for smaller, mid-size or entprises that want to use Truefoundry AI gateway without the operational overhead of self-hosting.
    </Tip>

    The key features and advantages of this deployment option are:

    1. **Globally distributed gateway to minimize latency**: Truefoundry gateway is deployed in multiple regions of the world across multiple zones and multiple cloud providers to provide low latency and high availability. Learn more about our [globally distributed infrastructure](/docs/ai-gateway/globally-distributed-saas).
    2. **Zero Overhead of maintenance**: There is no overhead of maintaining infrastructure and you can get access to the latest features and improvements.
    3. Data is encrypted at rest and in transit.
    4. Truefoundry Infrastructure is SOC2, ISO27001, GDPR, and HIPAA compliant

    **Installation:** No Installation Needed.
  </Accordion>

  <Accordion title="Option2: Gateway Plane only">
    This is a deployment of the gateway plane on your own infrastructure. The control-plane is hosted by Truefoundry and it doesn't include the AI Deployment module. The gateway exports the request-response data to an ingestion server running on the Truefoundry control plane. From the ingestion server, the data can be stored in a blob storage bucket either on the Truefoundry side or on your own infrastructure — depending on your data residency and compliance requirements. The control-plane stores the metrics and has access to the bucket containing the request-response data.

    <Frame caption="Gateway Plane and Data Storage on your own infrastructure">
      <img src="https://mintcdn.com/truefoundry/YDcIjALfJg9bdCyU/images/docs/platform/gateway-plane-only.png?fit=max&auto=format&n=YDcIjALfJg9bdCyU&q=85&s=9c3cdb1d0f1028130ec306442053a597" alt="Gateway Plane and Data Storage on your own infrastructure" width="1024" height="688" data-path="images/docs/platform/gateway-plane-only.png" />
    </Frame>

    The key features about this mode of deployment are:

    1. **LLM Traffic stays within your own premises**: All LLM traffic stays within your own infrastructure and Truefoundry doesn't come into the live path of a request to LLM.

    2. **Flexible data storage**: The request-response data flows from the gateway to the ingestion server on the Truefoundry control plane, and can then be stored in a bucket on the Truefoundry side or in a bucket on your own infrastructure. If stored on your end, you retain full ownership of the data. The data is stored in parquet format — so you can use it for analytics, debugging, and evaluation via Spark, DuckDB, Athena, or any tool of your choice.

    3. **Management of Gateway**: The gateway needs to be managed on your end.

    4. **Truefoundry control plane has access to the bucket containing the data**: Whether the data is stored on the Truefoundry side or on your infrastructure, the control plane needs access to the bucket to let you browse request logs on the Truefoundry dashboard. The control plane also compacts the logs and creates indexes to make querying faster.

    <Info>
      **If the bucket is on your infrastructure, you must grant the Truefoundry control plane access to it. Without this access, you will not be able to browse logs or use the dashboard features.**
    </Info>

    <Note>
      When the data is stored on your infrastructure and you browse request logs in the Truefoundry dashboard, the data is fetched from your blob storage — so you might incur egress charges from your cloud provider. The data might be cached temporarily in the control plane for faster queries. Egress charges also apply when the control plane compacts logs and creates indexes.
    </Note>

    **Installation:** [Deploy Gateway Plane](/docs/platform/deploy-gateway-plane)
  </Accordion>

  <Accordion title="Option3: Control Plane + Gateway Plane">
    Host both the gateway plane and the control plane on your own infrastructure.

    In this case, everything except the authentication server and analytics server, everything is hosted on your own infrastructure.

    <Frame caption="Control Plane, Gateway Plane and Data Storage on your own infrastructure">
      <img src="https://mintcdn.com/truefoundry/5CkapnZ7CyjQJ4bx/images/docs/gateway-control-plane-customer-end.png?fit=max&auto=format&n=5CkapnZ7CyjQJ4bx&q=85&s=28ed7825a58c58b903c58428110365ac" alt="Control Plane, Gateway Plane and Data Storage on your own infrastructure" width="3634" height="2803" data-path="images/docs/gateway-control-plane-customer-end.png" />
    </Frame>

    <Note>
      The only data sent to authentication/licensing server are the emails of the employees using the platform and the count of the requests flowing through the gateway. To understand how SSO works with our central authentication server, [refer to this page](/docs/sso). This helps us keep track of the licenses and billing.
    </Note>

    **Installation:** [Deploy Control Plane and Gateway Plane](/docs/platform/deploy-control-plane-and-gateway-plane)
  </Accordion>

  <Accordion title="Option 4: Compute Plane + AI Gateway SAAS">
    Deploy models/services (Compute Plane) on your own infrastructure while using managed AI Gateway.

    <Frame caption="Compute Plane + AI Gateway SAAS">
      <img src="https://mintcdn.com/truefoundry/41rnemFLxWdEyTu9/images/docs/platform/compute-plane-ai-gateway-saas.png?fit=max&auto=format&n=41rnemFLxWdEyTu9&q=85&s=2a38c001d73119222a7b121ff76da4d6" alt="Compute Plane + AI Gateway SAAS" width="3641" height="2170" data-path="images/docs/platform/compute-plane-ai-gateway-saas.png" />
    </Frame>

    **Installation:** [Deploy Compute Plane](/docs/infrastructure/deploy-compute-plane).
  </Accordion>

  <Accordion title="Option 5: Compute Plane only">
    Host the compute plane for model/services deployment, control-plane hosted by Truefoundry. This doesn't include the AI Gateway product.

    <Frame caption="Compute Plane only">
      <img src="https://mintcdn.com/truefoundry/SlXC71NQlRpPV_MA/images/docs/platform/compute-plane-only.png?fit=max&auto=format&n=SlXC71NQlRpPV_MA&q=85&s=f71e3de15cad62c1ad27c5d49c8f76f7" alt="Compute Plane only" width="3141" height="2052" data-path="images/docs/platform/compute-plane-only.png" />
    </Frame>

    **Installation:** [Deploy Compute Plane](/docs/infrastructure/deploy-compute-plane) (AWS, GCP, Azure, or generic).
  </Accordion>

  <Accordion title="Option 6: Compute Plane + Control Plane">
    Complete self-hosted for AI Deployment, full control over infra, no Gateway product.

    <Frame caption="Compute Plane + Control Plane">
      <img src="https://mintcdn.com/truefoundry/41rnemFLxWdEyTu9/images/docs/platform/compute-plane-control-plane.png?fit=max&auto=format&n=41rnemFLxWdEyTu9&q=85&s=f433a7c2ac7790965c1bfbcaf95183c0" alt="Compute Plane + Control Plane" width="3141" height="2236" data-path="images/docs/platform/compute-plane-control-plane.png" />
    </Frame>

    **Installation:** [Deploy Control and Compute Plane](/docs/deploy-control-and-compute-plane) (AWS, GCP, Azure, or generic control plane guides).
  </Accordion>

  <Accordion title="Option 7: Control Plane + Gateway Plane + Compute Plane">
    Self-host the control-plane, gateway plane and compute plane all in your own infrastructure.

    <Frame caption="Control Plane + Gateway Plane + Compute Plane">
      <img src="https://mintcdn.com/truefoundry/SlXC71NQlRpPV_MA/images/docs/platform/control-plane-compute-plane-gateway-plane.png?fit=max&auto=format&n=SlXC71NQlRpPV_MA&q=85&s=1a679168594e439a2c93d628b588bcb9" alt="Control Plane + Gateway Plane + Compute Plane" width="3199" height="2350" data-path="images/docs/platform/control-plane-compute-plane-gateway-plane.png" />
    </Frame>

    **Installation:** [Deploy Control and Compute Plane](/docs/deploy-control-and-compute-plane).
  </Accordion>
</AccordionGroup>

## Understanding the Installation Process

Truefoundry software ships as a combination of OpenTofu/Terraform code and helm charts. Your exact deployment process will depend on your current state of infrastructure and the modules you want to deploy. Here's a brief overview of how the compute-plane, gateway-plane and control-plane are deployed.

### Compute Plane

This comprises of a Kubernetes cluster and few add-ons on the Kubernetes cluster as described [here](/docs/platform/compute-plane-architecture). If don't have a Kubernetes cluster, Truefoundry can also provide the OpenTofu/Terraform code to provision the cluster on AWS, GCP or Azure. The OpenTofu/Terraform code brings up the cluster, installs the add-ons on the cluster, and creates roles to grant the control-plane permission to DockerRegistry, SecretsManager, BlobStorage and the Kubernetes cluster itself. Each of the addons is installed as an argo-application on the Kubernetes cluster. You can find the list of addons in the [compute plane overview](/docs/infrastructure/deploy-compute-plane) page.

Once the addons are installed on the cluster, the management of the addons (upgrade, value changes) are done from the Truefoundry dashboard itself and terrraform is not required. OpenTofu/Terraform is only used to upgrade the cluster or make other changes in the infrastructure.

<img src="https://mintcdn.com/truefoundry/yBeFg05DaZI7O318/images/docs/platform/addons-management-controlplane.png?fit=max&auto=format&n=yBeFg05DaZI7O318&q=85&s=c0f29940e40d3a56f025a38391a17227" alt="Compute Plane Addons Management" width="3840" height="1868" data-path="images/docs/platform/addons-management-controlplane.png" />

### Gateway Plane

The gateway plane ships as a single helm-chart that can be deployed on any Kubernetes cluster. The gateway has no external dependencies and only **needs to be able to connect to the control-plane via a secure WebSocket connection**. The gateway plane is also stateless and has no database or storage attached to it.

The Truefoundry gateway helm chart is available in this Github repository: [tfy-llm-gateway](https://github.com/truefoundry/infra-charts/tree/main/charts/tfy-llm-gateway)

### Control Plane

The control-plane ships as a single helm-chart that can be deployed on any Kubernetes cluster. It also requires a PostgreSQL database to store the data and connection to a blob storage to store the data. You can either bring your own Kubernetes cluster, Postgres and blobstorage or Truefoundry can help provision the same using our OpenTofu/Terraform code.

The control-plane helm chart is available in this Github repository: [truefoundry](https://github.com/truefoundry/infra-charts/tree/main/charts/truefoundry)

The control-plane helm chart includes the gateway helm chart as a dependency to make it easier to install both the control-plane and the gateway in one go.

## Overview of Helm Charts in Truefoundry

The key helm-charts in Truefoundry and their composition and usage is as follows:

| Helm Chart                             | Component                                         | Description                                                                                                                                                                                                        |
| -------------------------------------- | ------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **truefoundry**                        | Control-Plane + Gateway (optionally)              | The control-plane helm chart. You only need this chart if you are self-hosting the control-plane.                                                                                                                  |
| **tfy-llm-gateway**                    | Gateway Only                                      | The gateway helm chart - you only need to install this if you are self-hosting the gateway.                                                                                                                        |
| **tfy-k8s-aws-eks-inframold**          | AWS Compute-plane + Control-Plane(optionally)     | The AWS Inframold helm chart that contains all the addons in compute-plane and also the truefoundry control-plane. You can disable the control-plane installation if you are only installing the compute-plane     |
| **tfy-k8s-gcp-gke-standard-inframold** | GCP Compute-plane + Control-Plane(optionally)     | The GCP Inframold helm chart that contains all the addons in compute-plane and also the truefoundry control-plane. You can disable the control-plane installation if you are only installing the compute-plane     |
| **tfy-k8s-azure-aks-inframold**        | Azure Compute-plane + Control-Plane(optionally)   | The Azure Inframold helm chart that contains all the addons in compute-plane and also the truefoundry control-plane. You can disable the control-plane installation if you are only installing the compute-plane   |
| **tfy-k8s-generic-inframold**          | On-prem Compute-plane + Control-Plane(optionally) | The Generic Inframold helm chart that contains all the addons in compute-plane and also the truefoundry control-plane. You can disable the control-plane installation if you are only installing the compute-plane |
