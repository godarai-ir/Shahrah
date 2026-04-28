> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Truefoundry Architecture

> Truefoundry provides a split-plane architecture to decouple compute, gateway and control-plane and can be deployed on your own infrastructure.

Truefoundry comprises of a split-plane architecture that comprises of a control-plane as the orchestration layer and the compute, gateway and data planes as the execution and data layer.

<Frame caption="Truefoundry Platform Architecture">
  <img src="https://mintcdn.com/truefoundry/SlXC71NQlRpPV_MA/images/docs/platform/truefoundry-platform-architecture.png?fit=max&auto=format&n=SlXC71NQlRpPV_MA&q=85&s=ec5d62a9c636cf0939a7fa02247510a0" alt="Truefoundry Platform Architecture" width="3851" height="2716" data-path="images/docs/platform/truefoundry-platform-architecture.png" />
</Frame>

The key components in the architecture are:

1. **Global Authentication/Licensing Server**: This is a global licensing/authentication server. It is used to authenticate every user who logs into the control-plane. This is useful for Truefoundry to track the number of users in the platform and control licensing at a central layer. **The global authentication server cannot be shipped to customer's infrastructure.** The global authentication server is hosted at [https://auth.truefoundry.com](https://auth.truefoundry.com) and [https://login.truefoundry.com](https://login.truefoundry.com).

<Note>
  The only information being passed from the control-plane to the global authentication server is the emails of the users logging into the Truefoundry platform. To understand how SSO works with our central authentication server, [refer to this page](/docs/sso).
</Note>

2. **Analytics Server**: This is a global analytics server that collects metrics related to the usage of the platform. The key metrics being collected are the number of clusters connected to the control-plane, the addons installed in each of the clusters along with their versions, the version of the control-plane and the number of requests flowing through the gateway. The analytics server is hosted at [https://analytics.truefoundry.com](https://analytics.truefoundry.com).

3. **Control Plane**: The control plane is where the users/developers log in to manage and monitor compute, models, pipelines, services. It stores all the configuration data and interacts with the compute, gateway and data planes as well as external services like Docker Registry, Secret Manager, etc. To read more on what the control plane comprises of, check [Control Plane Architecture](/docs/platform/control-plane-architecture).

<Note>
  Truefoundry provides a hosted and managed control-plane to which you can attach your own compute and data plane. The control-plane can also be hosted within your own VPC as part of our enterprise plan.
</Note>

4. **Compute Plane**: The compute plane is the actual Kubernetes cluster where the models, services, jobs and pipelines deployed by the users run as part of the AI engineering module. **You don't need the compute plane if you are only using the AI gateway features.** To read more on what the compute plane comprises of, refer to [Compute Plane Architecture](/docs/platform/compute-plane-architecture).

<Note>
  Truefoundry doesn't provide compute plane and you need to bring your own cloud account or on-prem Kubernetes cluster.
</Note>

5. **Gateway Plane**: This is the LLM/MCP proxy that sits between the users and the LLM/MCP servers. It is responsible for providing unified API access to LLMs and MCP servers with enterprise-grade security and observability. To read more on what the gateway plane comprises of, refer to [Gateway Plane Architecture](/docs/platform/gateway-plane-architecture).

<Note>
  Truefoundry provides a managed and globally distributed gateway plane. It can also be deployed on your own infrastructure.
</Note>

6. **Data Plane**: The data plane stores the models and artifacts for the AI engineering module. You can bring your own blob storage to store the models/artifacts or use the Truefoundry managed one.
