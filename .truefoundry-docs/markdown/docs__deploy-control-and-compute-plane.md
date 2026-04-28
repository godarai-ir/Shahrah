> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Deploy Control Plane and Compute Plane

> Deploy TrueFoundry Control Plane and Compute Plane in your own infrastructure

Truefoundry Control-Plane ships as a helm chart ([https://github.com/truefoundry/infra-charts/tree/main/charts/truefoundry](https://github.com/truefoundry/infra-charts/tree/main/charts/truefoundry)) that can be deployed on any Kubernetes cluster. The installation process is different based on if you are spinning up a Kubernetes cluster from scratch or using an existing cluster to install the control-plane.

<Tip>
  Please read the [control-plane architecture](/docs/platform/control-plane-architecture) to understand the key components of the control plane.
</Tip>

### Control-plane Dependencies

The TrueFoundry control-plane requires a Kubernetes cluster, a Postgres database, and Blob Storage. Below is a table outlining popular options for each, along with typical cost ranges. Actual cost may vary depending on region, resource choice, scale, and provider pricing. All costs are rough monthly estimates.

| Component                                                              | Supported Options                                                                                                                                     | Typical Cost Range (USD/month)                                                    | Remarks                                                                                                                                                                             |
| ---------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Kubernetes Cluster**<br /> K8s version >= 1.28                       | AWS EKS<br /> GCP GKE<br /> Azure AKS<br /> Self-Managed K8s                                                                                          | \$400-1200 (depending on if we want AI deployment or AI Gateway or both features) | Kubernetes cluster is required to host the [control-plane helm chart](https://github.com/truefoundry/infra-charts/tree/main/charts/truefoundry) which brings up all the components. |
| **Postgres Database**<br /> Postgres version >= 13, Storage size: 30GB | -AWS RDS/Postgres (db.t4g.medium) <br /> - GCP CloudSQL (db-custom-1-3840)<br /> -Azure Postgres (GP\_Standard\_D4ds\_v5)<br /> -Self-hosted Postgres | \$200-400 (Depending on the cloud provider and the instance type)                 | The database stores all the data including deployment configuration, model registry metadata, user information, gateway configuration, etc.                                         |
| **Blob Storage**                                                       | AWS S3<br /> GCP Cloud Storage<br /> Azure Blob Storage<br /> Any S3-compatible storage like Minio                                                    | Depends on storage (Provision for 200GB min)                                      | Stores ML models and artifacts, Gateway request and response logs                                                                                                                   |
| **Domain Name**                                                        | Any provider (Route53, Cloudflare, etc.)                                                                                                              | Typically free with domain or varies per provider                                 | We need one domain to be pointed to the control-plane dashboard so that the UI can be accessed over the browser.                                                                    |

### Control-Plane Compute Requirements

The control-plane has different components which can be disabled or enabled based on your needs. The control-plane helm chart also has the Gateway plane bundled with it in case you want to deploy it. Here's the compute requirements based on the features you want to enable in the control-plane.

<Tabs>
  <Tab title="Small (Dev)">
    The small tier is recommended for POC/dev purposes. Here all the components are deployed on Kubernetes and in non HA mode (single replica). This is suitable if you are just testing out the different features of Truefoundry.

    <Warning>
      This setup brings up 1 replica of the services and is not highly-available. It can enable you to test the features but we do not recommend this for production mode.
    </Warning>

    |                 Configuration                |   CPU / Memory   | Block Storage PV |                           Min Nodes                          |     Cost    |
    | :------------------------------------------: | :--------------: | :--------------: | :----------------------------------------------------------: | :---------: |
    |        **AI Deployment + AI Gateway**        |   2 vCPU / 8GB   |    60GB <br />   | 2 <br /> <sub>*Pods should be spread over min 2 nodes*</sub> | \~ \$120 pm |
    |            **AI Deployment Only**            |   1 vCPU / 4GB   |    50GB <br />   | 2 <br /> <sub>*Pods should be spread over min 2 nodes*</sub> |  \~ \$60 pm |
    |              **AI Gateway Only**             |   2 vCPU / 8GB   |    60GB <br />   | 2 <br /> <sub>*Pods should be spread over min 2 nodes*</sub> | \~ \$120 pm |
    |        **AI Gateway + Gateway Plane**        |   3 vCPU / 10GB  |    60GB <br />   | 2 <br /> <sub>*Pods should be spread over min 2 nodes*</sub> | \~ \$120 pm |
    | **Postgres <br /> (Deployed on Kubernetes)** | 0.5 vCPU / 0.5GB |    5GB <br />    |                                                              |             |
    |      Blob Storage <br /> (S3 Compatible)     |                  |       20GB       |                                                              |             |
  </Tab>

  <Tab title="Medium (Prod)">
    The medium tier is configured for production and will suffice teams of 10-500 members.

    <Note>
      **This configuration can handle upto 10,000 deployments daily and around 50 Kubernetes compute-plane clusters.
      The AI Gateway is configured for a minimum 3 replicas which can handle around 500 requests/second to LLMs.**

      It's configured to be horizontally scalable and autoscale when the load increases. The Block Storage and S3 are used to store LLM request logs. The size is dependent on the size and number of requests and should be set as per the expected usage.
    </Note>

    |              Configuration             |  CPU / Memory  | Block Storage PV |                           Min Nodes                          |     Cost     |
    | :------------------------------------: | :------------: | :--------------: | :----------------------------------------------------------: | :----------: |
    |     **AI Deployment + AI Gateway**     | 16 vCPU / 56GB |   450GB <br />   | 3 <br /> <sub>*Pods should be spread over min 3 nodes*</sub> | \~ \$1000 pm |
    |         **AI Deployment Only**         |  8 vCPU / 24GB |   250GB <br />   | 3 <br /> <sub>*Pods should be spread over min 3 nodes*</sub> |  \~ \$400 pm |
    |           **AI Gateway Only**          | 16 vCPU / 48GB |   250GB <br />   | 3 <br /> <sub>*Pods should be spread over min 3 nodes*</sub> |  \~ \$800 pm |
    |     **AI Gateway + Gateway Plane**     | 20 vCPU / 60GB |   250GB <br />   | 3 <br /> <sub>*Pods should be spread over min 3 nodes*</sub> | \~ \$1200 pm |
    | **Postgres <br /> (Managed Database)** |  2 vCPU / 4GB  |    30GB <br />   |                                                              |              |
    |   Blob Storage <br /> (S3 Compatible)  |                |       500GB      |                                                              |              |
  </Tab>

  <Tab title="Large (Prod)">
    The large tier is configured for production and will suffice organizations of 500-50000 members.

    <Note>
      **This configuration can handle upto 1 million deployments daily and over 500 Kubernetes compute-plane clusters. The AI Gateway is configured for a minimum 10 replicas which can handle around 2000 requests/second to LLMs.**

      It's configured to be horizontally scalable and autoscale when the load increases. The Block Storage and S3 are used to store LLM request logs. The size is dependent on the size and number of requests and should be set as per the expected usage.
    </Note>

    |              Configuration             |  CPU / Memory  | Block Storage PV |                            Min Nodes                           |     Cost     |
    | :------------------------------------: | :------------: | :--------------: | :------------------------------------------------------------: | :----------: |
    |     **AI Deployment + AI Gateway**     | 32 vCPU / 72GB |   750GB <br />   |  10 <br /> <sub>*Pods should be spread over min 5 nodes*</sub> | \~ \$1600 pm |
    |         **AI Deployment Only**         | 16 vCPU / 32GB |   300GB <br />   |  5 <br /> <sub>*Pods should be spread over min 5 nodes*</sub>  | \~ \$1000 pm |
    |           **AI Gateway Only**          | 32 vCPU / 64GB |   400GB <br />   | 10 <br /> <sub>*Pods should be spread over min 10 nodes*</sub> | \~ \$1400 pm |
    |     **AI Gateway + Gateway Plane**     | 40 vCPU / 80GB |   400GB <br />   | 10 <br /> <sub>*Pods should be spread over min 10 nodes*</sub> | \~ \$1800 pm |
    | **Postgres <br /> (Managed Database)** |  8 vCPU / 16GB |   100GB <br />   |                                                                |              |
    |   Blob Storage <br /> (S3 Compatible)  |                |      1000GB      |                                                                |              |
  </Tab>
</Tabs>

### Deploying Control-Plane

There are different scenarios to install the control-plane depending on if you are spinning up a new Kubernetes cluster or using an existing cluster.

**1. Spin up a new Kubernetes cluster and deploy control-plane**

Truefoundry provides OpenTofu/Terraform code to spin up a new Kubernetes cluster and deploy the control-plane on it. You can customize the network configuration and Kubernetes configuration as per your needs. Please follow the guides below:

* [AWS Control Plane](/docs/aws-control-plane)
* [GCP Control Plane](/docs/gcp-control-plane)
* [Azure Control Plane](/docs/azure-control-plane)

**2. Deploy control-plane on an existing Kubernetes cluster**

Install the control-plane helm chart using helm install. Follow this [guide](/docs/installing-control-plane-using-helm-chart) to install the control-plane helm chart.

<Tip>
  If your environment is air-gapped or has restricted network access, refer to the air-gapped FAQ in the [control plane](/docs/platform/deploy-control-plane-and-gateway-plane#how-to-install-in-an-air-gapped--restricted-network-environment) or [compute plane](/docs/infrastructure/deploy-compute-plane#how-to-install-in-an-air-gapped--restricted-network-environment) installation guides for additional setup steps.
</Tip>
