> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Define Ports And Domains

> Define ports and domains for your services. Configure routing, exposure, and access correctly.

When you create a Service, you need to specify the ports that the Service will listen on. A service can also have mulitple ports.
Truefoundry allows you to define the ports as well as the authentication strategies at the port level.

<Frame>
  <img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/docs/ports-section-in-service.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=8aec2a548f26cfe4899b6fbdd4c47280" width="3834" height="1936" data-path="images/docs/ports-section-in-service.png" />
</Frame>

The key values in the Ports section are described as follows:

### Port Number

This is the port number that the service will listen on. This is usually 8000 for FastAPI, 8501 for Streamlit and 7860 for Gradio.
If you are writing a Dockerfile, this will usually be the port that you write after the `EXPOSE` instruction.

### Port Protocol

You can choose the protocol to be `HTTP` in most cases where you are using REST APIs. If hosting a GRPC service, please choose `GRPC`.

### Expose the port

This is governed by the `Expose` checkbox. Exposing a service will enable it to be called from anywhere using the endpoint - other users or services living outside your cluster will also be able to access it. If you do not expose a port, it will only be accessible from within the cluster using the URL: **`servicename-workspacename.svc.cluster.local:port`**

<Tip>
  If you expose a port, it will still be accessible from within the cluster using the URL: **`servicename-workspacename.svc.cluster.local:port`**
</Tip>

### Endpoint

If you expose a port, we will need to map a URL to the endpoint using which we can access the service. You should see one/multiple domains in the dropdown - depending on what has been configured by the Infra team during the cluster setup. The domains for a cluster are configured for each cluster in the Cluster Settings page.

<Columns cols={2}>
  <Card title="Set domains for cluster">
    <img src="https://mintcdn.com/truefoundry/5CkapnZ7CyjQJ4bx/images/docs/edit-domains-cluster.png?fit=max&auto=format&n=5CkapnZ7CyjQJ4bx&q=85&s=063d9f0ff1bb03f30a1e0c9b6a5aab1b" width="3836" height="1934" data-path="images/docs/edit-domains-cluster.png" />
  </Card>

  <Card title="View domains for cluster">
    <img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/docs/view-domains-cluster.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=031a266bec9f1712ccf0da22c949599b" width="3834" height="1926" data-path="images/docs/view-domains-cluster.png" />
  </Card>
</Columns>

<Note>
  The domains are mapped to the LoadBalancer IP of the cluster. Once the traffic reaches the loadbalancer, the routing to all the services inside the cluster is done using Istio Ingress Gateway.
</Note>

Domains can be of two types:

* **Wildcard Domain(`*.your-organization.com`)**: If we map a wildcard domain to the loadbalancer, every service deployed in the cluster can get its own subdomain like service1.your-organization.com, service2.your-organization.com etc. <Tip>We recommend using wildcard domains since many frontend applications do not work properly without a subdomain.</Tip>
* **Non-wildcard Domain: (`your-organization.com`)**: If we map a non-wildcard domain to the loadbalancer, every service deployed in the cluster will share the same domain. They will be routed to via different paths - E.g. your-organization.com/service1 and your-organization.com/service2.

The endpoint is of the following form:

```
https://[<subdomain>].your-organization.com/[<path-to-service>]
```

* `subdomain` can be set only if the domain in the cluster is a wildcard domain.
* The UI fills up default values for the subdomain and path to service. However, you can change them as per your requirements.

<Warning>
  Two services cannot have the same endpoint. If you try to deploy a second service with the same endpoint, Truefoundry will show you an error and ask you to change the endpoint.
</Warning>
