> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Globally Distributed SAAS Gateway

> Learn about Truefoundry's globally distributed AI Gateway infrastructure deployed across multiple regions and cloud providers

Truefoundry's globally distributed gateway is deployed across multiple regions and cloud providers to ensure high availability, low latency, and resilience against regional or cloud provider disruptions. The global gateway URL is `https://gateway.truefoundry.ai`.

<Info>
  **`gateway.truefoundry.ai` is the unified endpoint for both the AI Model Gateway and the MCP Gateway.**

  Whether you are routing LLM inference requests (OpenAI-compatible API, etc.) or connecting to MCP (Model Context Protocol) servers, all traffic goes through the same globally distributed infrastructure. This means MCP Gateway deployments benefit from the same multi-region, multi-cloud availability described on this page.
</Info>

<Frame caption="Truefoundry SAAS Global Gateway Architecture — Control Plane in Ireland with Gateway Planes distributed globally">
  <img src="https://mintcdn.com/truefoundry/vouMdj56xtoFDdZJ/images/docs/gateway-multi-region.png?fit=max&auto=format&n=vouMdj56xtoFDdZJ&q=85&s=ddbf007d6cfa6bfc368dfb1a6dd58fb8" alt="Architecture diagram showing the Control Plane in Ireland (Europe) managing configuration, with Gateway Planes handling LLM and MCP traffic across 14 global regions connected via NATS" width="1376" height="768" data-path="images/docs/gateway-multi-region.png" />
</Frame>

## Features

1. **Globally Distributed**: Deployed across more than 12 regions around the globe and across 3 multiple cloud providers for maximum availability while minimizing latency.
2. **Automated Failover**: All traffic is routed to the nearest gateway for minimum latency. In case of regional downtime, traffic is automatically routed to closest healthy regions ensuring uninterrupted service.
3. **Multi-Cloud Deployment**: Distributed across multiple cloud providers to be tolerant to cloud provider-specific disruptions.
4. **Data Encryption**: Data is encrypted at rest and in transit.
5. **Compliance**: Truefoundry Infrastructure is SOC2, ISO27001, GDPR, and HIPAA compliant

## Architecture

The SAAS global deployment follows the same [Gateway Plane Architecture](/docs/platform/gateway-plane-architecture) used across all Truefoundry deployments. It consists of two key components:

* **Control Plane** — Manages all gateway configuration including models, users, teams, virtual accounts, rate-limiting, and routing configs. The SAAS control plane is hosted in **Ireland (Europe)**.
* **Gateway Planes** — Stateless, horizontally scalable gateway instances that handle all production traffic (LLM requests, MCP requests, etc.). These are deployed across the regions listed in the [Regional Deployments](#regional-deployments) section below.

The gateway planes subscribe to the control plane for configuration updates via NATS and perform all authentication, authorization, rate-limiting, and load-balancing checks in-memory with no external calls in the request path. For a detailed breakdown of the request flow, performance benchmarks, and FAQs, see the [Gateway Plane Architecture](/docs/platform/gateway-plane-architecture) page.

<Note>
  The specific regions and locations where gateway planes are deployed are subject to change based on Truefoundry's internal infrastructure needs. Regions may be added, removed, or relocated without prior notice.
</Note>

## Global Deployment

For most use cases, we recommend using the global endpoint which automatically routes to the nearest healthy gateway:

| Deployment           | Global Endpoint                  |
| -------------------- | -------------------------------- |
| Global (Auto-routed) | `https://gateway.truefoundry.ai` |

<Tip>
  Using the global URL (`https://gateway.truefoundry.ai`) is recommended as it automatically routes your requests to the nearest healthy gateway instance based on your geographic location to give you the minimum possible latency.
</Tip>

## Regional Deployments

Each gateway region has its own URL and associated metadata. Every request routed through the SaaS gateway is automatically enriched with the `tfy_gateway_region` and `tfy_gateway_zone` metadata keys that identify which gateway region and zone handled the request. The **Region** and **Zone** columns in the table below show the values these keys will contain.

<Warning>
  **Please do not use the regional endpoints in production since they are subject to change at any point without notice.**

  We recommend using the global endpoint (`https://gateway.truefoundry.ai`) for production use, or the region-specific endpoint described in the [Multi-regional Deployments](#multi-regional-deployments) section.
</Warning>

| Physical Location                   | Cloud Provider | Region | Zone  | Regional Endpoint                    |
| ----------------------------------- | -------------- | ------ | ----- | ------------------------------------ |
| North Virginia, United States (ORF) | AWS            | `US`   | `ORF` | `https://orf.gateway.truefoundry.ai` |
| San Francisco, United States (SFO)  | Azure          | `US`   | `SFO` | `https://sfo.gateway.truefoundry.ai` |
| Dallas, Texas, United States (DFW)  | GCP            | `US`   | `DFW` | `https://dfw.gateway.truefoundry.ai` |
| Toronto, Canada (YYZ)               | GCP            | `CA`   | `YYZ` | `https://yyz.gateway.truefoundry.ai` |
| Sao Paulo, Brazil (GRU)             | GCP            | `SA`   | `GRU` | `https://gru.gateway.truefoundry.ai` |
| London, United Kingdom (LHR)        | AWS            | `EU`   | `LHR` | `https://lhr.gateway.truefoundry.ai` |
| Madrid, Spain (MAD)                 | GCP            | `EU`   | `MAD` | `https://mad.gateway.truefoundry.ai` |
| Gavle, Sweden (GVX)                 | Azure          | `EU`   | `GVX` | `https://gvx.gateway.truefoundry.ai` |
| Cape Town, South Africa (CPT)       | AWS            | `AF`   | `CPT` | `https://cpt.gateway.truefoundry.ai` |
| Doha, Qatar (DIA)                   | GCP            | `US`   | `DIA` | `https://dia.gateway.truefoundry.ai` |
| Mumbai, India (BOM)                 | AWS            | `IN`   | `BOM` | `https://bom.gateway.truefoundry.ai` |
| Singapore, Singapore (SIN)          | AWS            | `AP`   | `SIN` | `https://sin.gateway.truefoundry.ai` |
| Melbourne, Australia (MEL)          | AWS            | `AU`   | `MEL` | `https://mel.gateway.truefoundry.ai` |
| Sydney, Australia (SYD)             | AWS            | `AU`   | `SYD` | `https://syd.gateway.truefoundry.ai` |

<Tip>
  The `tfy_gateway_region` and `tfy_gateway_zone` keys are set automatically — you do not need to send them. They are part of the resolved metadata on every request and can be used for:

  * **[Virtual model routing](/docs/ai-gateway/virtual-model#metadata-based-target-filtering)** — Route to region-specific model deployments using `metadata_match`
  * **[Data routing](/docs/ai-gateway/data-routing)** — Send logs to region-specific destinations based on gateway location
  * **[Data access filters](/docs/ai-gateway/data-access)** — Scope log visibility by gateway region
  * **[Request filtering](/docs/ai-gateway/log-custom-metadata)** — Filter request logs by region in the dashboard

  For example, you can route US gateway traffic to Azure US-deployed models and EU gateway traffic to Azure EU-deployed models. See [region-based routing with virtual models](/docs/ai-gateway/virtual-model#metadata-based-target-filtering) for a worked example.
</Tip>

## Multi-regional Deployments

Multi-regional endpoints automatically route your requests to the closest healthy gateway within a specific geographic region. If all regional locations are unavailable, traffic is routed to the designated fallback regions.

| Region        | Multi-regional Endpoint             | Primary Locations                                       | Fallback Locations    |
| ------------- | ----------------------------------- | ------------------------------------------------------- | --------------------- |
| United States | `https://us.gateway.truefoundry.ai` | North Virginia (ORF), San Francisco (SFO), Dallas (DFW) | Toronto, Canada (YYZ) |
| Europe        | `https://eu.gateway.truefoundry.ai` | London (LHR), Madrid (MAD), Gavle (GVX)                 | Doha, Qatar (DIA)     |
| Australia     | `https://au.gateway.truefoundry.ai` | Sydney (SYD), Melbourne (MEL)                           | Singapore (SIN)       |

## Gateway Status Monitoring

To track the status of each gateway deployment and receive real-time updates on service availability, visit our status page:

**Gateway Status Page**: [status.truefoundry.com](https://status.truefoundry.com)

You can expand the **AI Gateway** section to see per-region uptime:

<Frame caption="Per-region gateway uptime on the status page">
  <img src="https://mintcdn.com/truefoundry/vouMdj56xtoFDdZJ/images/docs/ai-gateway/gateway-status-regions.png?fit=max&auto=format&n=vouMdj56xtoFDdZJ&q=85&s=1f8ae1820ff3c5b84c363085f2e38d1c" alt="AI Gateway expanded view showing per-region uptime for each gateway deployment" width="1024" height="560" data-path="images/docs/ai-gateway/gateway-status-regions.png" />
</Frame>

### Subscribe to Status Updates

Stay informed about gateway availability by subscribing to status notifications:

1. Visit the [Gateway Status Page](https://status.truefoundry.com)
2. Click the **Get Updates** button in the top right
3. Choose your preferred notification method:
   * Email notifications
   * RSS Feed
   * On a custom webhook

<Frame caption="Click the Get Updates button on the status page to subscribe to notifications">
  <img src="https://mintcdn.com/truefoundry/vouMdj56xtoFDdZJ/images/docs/ai-gateway/gateway-status-page.png?fit=max&auto=format&n=vouMdj56xtoFDdZJ&q=85&s=c3c34669bc02b31b7bc0ce06d1d5a188" alt="Truefoundry status page with the Get Updates button highlighted in the top right" width="1024" height="518" data-path="images/docs/ai-gateway/gateway-status-page.png" />
</Frame>

## Connecting Your Private Models or MCP Servers to the Gateway

When you use the SAAS gateway to proxy requests to models or MCP servers hosted within your own infrastructure (e.g. inside a VPC), the gateway needs a network path to reach those endpoints. Depending on your security requirements and infrastructure setup, there are several approaches to establish this connectivity.

<AccordionGroup>
  <Accordion title="Option 1: Reverse Tunnel Agent">
    Instead of exposing your model or MCP server publicly, you can run `cloudflared` inside your infrastructure and establish an outbound tunnel to a Truefoundry-managed public domain. This avoids opening inbound ports while still allowing the gateway to reach your private services.

    <Note>
      This is only available as part of our higher tier enterprise plans.
    </Note>

    **Setup flow**

    1. Raise a request with the Truefoundry team with the following details:
       * `private_service_url`: The URL of the private service that hosts your MCP server or model. You can also setup a load balancer in front of multiple MCP servers or models to route traffic as needed. Eg - For an mcp server running within the same kubernetes cluster as cloudflared, it would be something like `http://mcp-server.default.svc.cluster.local:8080`.

    2. Truefoundry shares the following with you:
       * `public_domain` - The public domain that you can use to access your private service. For example `<tenant-name>.private-endpoints.truefoundry.ai`.
       * `auth_token` - The token that you can use to run `cloudflared` in your environment.

    3. Run `cloudflared` in your environment:

           <Tabs>
             <Tab title="VM">
               ```bash theme={"dark"}
               cloudflared tunnel run --token <auth_token>
               ```
             </Tab>

             <Tab title="Kubernetes (Helm)">
               Create a `values.yaml` file:

               ```yaml theme={"dark"}
               tunnel:
                 token: <auth_token>
               ```

               Install `tfy-cloudflared`:

               ```bash theme={"dark"}
               helm repo add truefoundry https://truefoundry.github.io/infra-charts/
               helm repo update
               helm install tfy-cloudflared truefoundry/tfy-cloudflared -f values.yaml -n cloudflared --create-namespace
               ```

               Verify the installation that the pods are running:

               ```bash theme={"dark"}
               kubectl get pods -n cloudflared
               ```
             </Tab>
           </Tabs>

    4. You can now create the remote MCP server using `https://<tenant-name>.private-endpoints.truefoundry.ai/mcp`. For detailed MCP server setup instructions, see [Getting started with MCP servers](/docs/ai-gateway/mcp/mcp-server-getting-started).
  </Accordion>

  <Accordion title="Option 2: Private Link">
    For organizations that require fully private network connectivity with no public internet exposure, you can establish a Private Link connection between the Truefoundry gateway and your VPC. This creates a private, direct network path that never traverses the public internet.

    <Note>
      Private Link connections are region-specific. You will need to set up a Private Link in each cloud region where you need private connectivity to the gateway. Refer to the [Regional Deployments](#regional-deployments) table above to identify the regions and cloud providers available, then contact the [Truefoundry team](https://truefoundry.com/contact) to get the specific service endpoint names for your selected regions.

      **This is only available as part of our higher tier enterprise plans.**
    </Note>

    <AccordionGroup>
      <Accordion title="AWS PrivateLink">
        **AWS PrivateLink** allows you to privately connect your VPC to the Truefoundry gateway without exposing traffic to the public internet.

        **Prerequisites:**

        * An AWS account with a VPC in a [supported gateway region](#regional-deployments)
        * Appropriate IAM permissions to create VPC endpoints

        **Steps:**

        1. Contact the [Truefoundry team](https://truefoundry.com/contact) to get the **VPC Endpoint Service Name** for your desired region.
        2. In the AWS Console, navigate to **VPC → Endpoints → Create Endpoint**.
        3. Select **Find service by name** and enter the service name provided by Truefoundry.
        4. Select the VPC and subnets where your model or MCP server is running.
        5. Attach a security group that allows traffic on the required ports.
        6. Create the endpoint and wait for the connection to be accepted by Truefoundry.
        7. Once active, use the VPC endpoint DNS name as the target for your model or MCP server configuration in the gateway.

        ```bash theme={"dark"}
        # Verify connectivity from within your VPC
        curl -X GET "https://<vpc-endpoint-dns>/health"
        ```
      </Accordion>

      <Accordion title="GCP Private Service Connect">
        **GCP Private Service Connect** enables private connectivity from your VPC to the Truefoundry gateway.

        **Prerequisites:**

        * A GCP project with a VPC in a [supported gateway region](#regional-deployments)
        * Appropriate IAM permissions to create Private Service Connect endpoints

        **Steps:**

        1. Contact the [Truefoundry team](https://truefoundry.com/contact) to get the **Service Attachment URI** for your desired region.
        2. In the GCP Console, navigate to **Network Services → Private Service Connect**.
        3. Click **Create Endpoint** and select **Published service**.
        4. Enter the service attachment URI provided by Truefoundry.
        5. Select the VPC network and subnetwork where your model or MCP server is running.
        6. Assign an internal IP address for the endpoint.
        7. Create the endpoint and wait for the connection to be accepted by Truefoundry.
        8. Use the assigned internal IP address or configure a DNS entry to route traffic through the Private Service Connect endpoint.

        ```bash theme={"dark"}
        # Verify connectivity from within your VPC
        curl -X GET "http://<internal-ip>/health"
        ```
      </Accordion>

      <Accordion title="Azure Private Link">
        **Azure Private Link** provides private connectivity from your VNet to the Truefoundry gateway.

        **Prerequisites:**

        * An Azure subscription with a VNet in a [supported gateway region](#regional-deployments)
        * Appropriate permissions to create private endpoints

        **Steps:**

        1. Contact the [Truefoundry team](https://truefoundry.com/contact) to get the **Private Link Service Resource ID** for your desired region.
        2. In the Azure Portal, navigate to **Private Link Center → Private Endpoints → Create**.
        3. Select your subscription, resource group, and region.
        4. Under **Resource**, select **Connect to an Azure resource by resource ID or alias** and enter the resource ID provided by Truefoundry.
        5. Select the VNet and subnet where your model or MCP server is running.
        6. Configure DNS integration to automatically create a private DNS zone (recommended).
        7. Create the private endpoint and wait for the connection to be approved by Truefoundry.
        8. Use the private endpoint's IP address or DNS name to route traffic.

        ```bash theme={"dark"}
        # Verify connectivity from within your VNet
        curl -X GET "http://<private-endpoint-ip>/health"
        ```
      </Accordion>
    </AccordionGroup>

    **Tradeoffs:**

    * Most secure option — traffic never leaves the cloud provider's private network.
    * Requires setup in each region where you need connectivity.
    * Involves coordination with the Truefoundry team for service endpoint provisioning.
    * May incur additional cloud provider charges for Private Link / Private Service Connect endpoints and data transfer.
  </Accordion>

  <Accordion title="Option 3: Self-Host the Gateway">
    If none of the above options meet your requirements, you can self-host the Truefoundry gateway plane within your own infrastructure. In this setup, the gateway runs inside your VPC and has direct network access to your models and MCP servers — eliminating any need for public exposure or cross-network connectivity. [Read more about self-hosting the gateway plane](/docs/platform/deployment-overview#option2-gateway-plane-only).
  </Accordion>
</AccordionGroup>

## FAQ

<AccordionGroup>
  <Accordion title="What is the round trip latency to the SAAS gateway?">
    Your client is automatically routed to the closest gateway region, so the round trip time (RTT) from your application to the gateway typically ranges from **20–50ms**.

    If you are seeing higher latencies, please let us know and we will be happy to add another region closer to your use case.

    If you are **self-hosting the gateway** within your own infrastructure, the RTT from your application to the gateway will be on the order of **\~1ms** when both are running in the same cluster.
  </Accordion>
</AccordionGroup>
