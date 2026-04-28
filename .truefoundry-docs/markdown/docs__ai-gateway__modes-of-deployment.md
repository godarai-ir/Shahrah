> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Deployment Options

> Explore the different Gateway deployment options ranging from self-hosting to fully managed

Truefoundry provides different deployment options for AI gateway depending on if you want to use SAAS hosted gateway or self-host the gateway. Before understanding the different deployment options, its important to read about the [gateway-plane architecture](/docs/platform/gateway-plane-architecture).

## Deployment Options

| Deployment Scenario                                                 | Modules    | Pricing Tier | Remarks                                                                                     | Cost of Hosting      |
| ------------------------------------------------------------------- | ---------- | ------------ | ------------------------------------------------------------------------------------------- | -------------------- |
| **1. AI Gateway SAAS only**                                         | Gateway    | Starter      | Use managed AI Gateway for LLM requests, no compute plane or control plane hosting needed.  | 0                    |
| **2. Saas AI Gateway with data storage on your own infrastructure** | Gateway    | Starter      | Use managed AI Gateway for LLM requests, no compute plane or control plane hosting needed.  | Storage cost in S3   |
| **3. Gateway Plane only**                                           | Gateway    | Enterprise   | Host the Gateway plane for LLM requests and use control-plane hosted by Truefoundry.        | \~\$600/month        |
| **4. Control Plane + Gateway Plane**                                | AI Gateway | Enterprise   | Use only AI Gateway, full control, suitable if you don't need compute plane/app deployment. | \~ \$800-1000/month. |

The key deployment options described above are:

### Option 1: AI Gateway SAAS only

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

### Option 2: Saas AI Gateway with data storage on your own infrastructure

In this case, the gateway and control-plane are managed by Truefoundry on its own infrastructure, but the LLM data is stored on your own infrastructure.

<Frame caption="Truefoundry Managed Gateway with data storage on your own infrastructure">
  <img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/docs/truefoundry-gateway-with-customer-storage.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=d426fad2f72c221f2eb268565ac849b6" alt="Truefoundry Managed SAAS" width="3774" height="2384" data-path="images/docs/truefoundry-gateway-with-customer-storage.png" />
</Frame>

Tha advantages of this deployment option are as follows:

1. **No Infrastructure Management**: In this case, the gateway is globally distributed, available and fully managed by Truefoundry. Learn more about our [globally distributed infrastructure](/docs/ai-gateway/globally-distributed-saas). You just need to provide the blob storage(AWS S3 Bucket, or GCS Bucket or Azure Blob Storage or any other S3 compatible storage) and the control-plane will use it to store the LLM data.

2. **You retain data ownership**: You retain full ownership of the request-response data since the its stored on a bucket on your end. The data are stored in parquet format - so you can use them for analytics, debugging and evaluation via Spark, DuckDB, Athena or any tool of your choice.

<Note>
  The actual request response data in this case flows through Truefoundry control-plane and then gets stored in a blob storage on your infrastructure. The control-plane might cache some of the data for faster queries, but doesn't do any long term retention of the data.

  When you are browsing the request logs in the Truefoundry dashboard, the data will be fetched from your blob storage - so you might incur egress charges from your cloud provider.
</Note>

### Option3: Gateway Plane on your own infrastructure

This is a deployment of the gateway plane on your own infrastructure. The control-plane is hosted by Truefoundry. The gateway exports the request-response data to the ingestion server which then stores the data in your own blob storage. The control-plane stores the metrics and has access to the bucket containing the request-response data.

<Frame caption="Gateway Plane and Data Storage on your own infrastructure">
  <img src="https://mintcdn.com/truefoundry/YDcIjALfJg9bdCyU/images/docs/platform/gateway-plane-only.png?fit=max&auto=format&n=YDcIjALfJg9bdCyU&q=85&s=9c3cdb1d0f1028130ec306442053a597" alt="Gateway Plane and Data Storage on your own infrastructure" width="1024" height="688" data-path="images/docs/platform/gateway-plane-only.png" />
</Frame>

The key features about this mode of deployment are:

1. **LLM Traffic stays within your own premises**: All LLM traffic stays within your own infrastructure and Truefoundry doesn't come into the live path of a request to LLM.

2. **You retain full control over your data**: You retain full ownership of the request-response data since the its stored on a bucket on your end. The data are stored in parquet format - so you can use them for analytics, debugging and evaluation via Spark, DuckDB, Athena or any tool of your choice.

3. **Management of Gateway and Ingestion Service**: The gateway and ingestion service availability needs to be managed on your end.

4. **Truefoundry control plane has access to the bucket containing the data**: This access helps you browse the request logs on the truefoundry dashboard. The control-plane also need to access the logs to be able to compact it and create indexes on it to make it faster to query.

<Info>
  **You will not be be able to use this feature if you don't give access to Truefoundry control plane access to your bucket.**
</Info>

<Note>
  When you are browsing the request logs in the Truefoundry dashboard, the data will be fetched from your blob storage - so you might incur egress charges from your cloud provider. The data might be cached temporarily in the control-plane for faster queries. There are also egress charges when the control-plane compacts the logs and creates indexes on it.
</Note>

### Option4: Control Plane + Gateway Plane

Host both the gateway plane and the control plane on your own infrastructure.

In this case, everything except the authentication server and analytics server, everything is hosted on your own infrastructure.

<Frame caption="Control Plane, Gateway Plane and Data Storage on your own infrastructure">
  <img src="https://mintcdn.com/truefoundry/5CkapnZ7CyjQJ4bx/images/docs/gateway-control-plane-customer-end.png?fit=max&auto=format&n=5CkapnZ7CyjQJ4bx&q=85&s=28ed7825a58c58b903c58428110365ac" alt="Control Plane, Gateway Plane and Data Storage on your own infrastructure" width="3634" height="2803" data-path="images/docs/gateway-control-plane-customer-end.png" />
</Frame>

<Note>
  The only data sent to authentication/licensing server are the emails of the employees using the platform and the count of the requests flowing through the gateway. To understand how SSO works with our central authentication server, [refer to this page](/docs/sso). This helps us keep track of the licenses and billing.
</Note>
