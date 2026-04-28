> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Generic

> This page provides an overview of the architecture, requirements and steps to install the TrueFoundry compute plane cluster in your generic cluster

## Requirements:

The requirements to setup compute plane in existing cluster is as follows:

* <Icon icon="square-check" iconType="regular" /> K8s version 1.28 or higher with ability to run 250 nodes and 4096 pods with GPU and CPU depending on your use case.
* <Icon icon="square-check" iconType="regular" /> Egress access to container registries - `public.ecr.aws`, `quay.io`, `ghcr.io`, `tfy.jfrog.io`, `docker.io/natsio`, `nvcr.io`, `registry.k8s.io` so that we can download the docker images for argocd, nats, gpu operator, argo rollouts, argo workflows, istio, keda, etc.
* <Icon icon="square-check" iconType="regular" /> We need a domain to map to the service endpoints and certificate to encrypt the traffic. A wildcard domain like \*.services.example.com is preferred. TrueFoundry can do path based routing like `services.example.com/tfy/*`, however, many frontend applications do not support this. For certificate, check [this](#setting-up-tls-in-generic-cluster) document for more details.
* <Icon icon="square-check" iconType="regular" /> Support for block storage for disk volumes and NFS storage for artifact storage.
* <Icon icon="square-check" iconType="regular" /> A bucket and container registry will be required for the cluster.
* <Icon icon="square-check" iconType="regular" /> An IP pool is required for the load balancer to get an IP address. Check [metallb](https://metallb.io/) for more details.

## Setting up compute plane

TrueFoundry compute plane infrastructure, for generic cluster, is provisioned using helm. You can download the helm commands and execute them on your cluster.

<Steps>
  <Step title="Choose an existing cluster">
    Go to the platform section in the left panel and click on `Clusters`. You can click on `Attach Existing Cluster` . Read the requirements and if everything is satisfied, click on `Continue`.

    <img src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/create-compute-plane-screenshot-1.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=b3febf85743f0b5d32adb737e23eadb6" width="3840" height="1938" data-path="images/docs/create-compute-plane-screenshot-1.png" />
  </Step>

  <Step title="Fill up the form to generate the helm commands">
    A form will be presented with the details for the new cluster to be created. Fill in with your cluster details. Click `Submit` when done

    <Tabs>
      <Tab title="Attach Existing Cluster">
        The key fields to fill up here are:

        * `Cluster Name` - Your cluster name.
        * `Cluster Addons` - TrueFoundry needs to install addons like ArgoCD, ArgoWorkflows, Keda, Istio, etc. Please disable the addons that are already installed on your cluster so that truefoundry installation does not overrride the existing configuration and affect your existing workloads.
      </Tab>
    </Tabs>
  </Step>

  <Step title="Copy the helm command and execute it on your cluster">
    You will be presented with `helm` commands to install the TrueFoundry compute plane.

    <img src="https://mintcdn.com/truefoundry/5CkapnZ7CyjQJ4bx/images/docs/how-to-deploy-your-own-cloud/generic-compute-plane-helm-script.png?fit=max&auto=format&n=5CkapnZ7CyjQJ4bx&q=85&s=207f1416f7a7f7ff26ec99a5cd612df7" width="1538" height="1642" data-path="images/docs/how-to-deploy-your-own-cloud/generic-compute-plane-helm-script.png" />
  </Step>

  <Step title="Verify the cluster is showing as connected in the platform">
    Once the helm commands are executed, the cluster will be shown as connected in the platform.
  </Step>

  <Step title="Create DNS Record">
    We can get the load-balancer's IP address by going to the platform section in the bottom left panel under the Clusters section. Under the preferred cluster, you'll see the load balancer IP address under the `Base Domain URL` section.

    <img src="https://mintcdn.com/truefoundry/5CkapnZ7CyjQJ4bx/images/docs/how-to-deploy-your-own-cloud/generic-compute-plane-load-balancer-address.png?fit=max&auto=format&n=5CkapnZ7CyjQJ4bx&q=85&s=19736feeda516cf400b61e8811bf16b5" width="2786" height="372" data-path="images/docs/how-to-deploy-your-own-cloud/generic-compute-plane-load-balancer-address.png" />

    Create a DNS record in your DNS provider with the following details

    | Record Type | Record Name        | Record value              |
    | ----------- | ------------------ | ------------------------- |
    | CNAME       | \*.tfy.example.com | LOADBALANCER\_IP\_ADDRESS |
  </Step>

  <Step title="Setup routing and TLS for deploying workloads to your cluster">
    Follow the instructions [here](#setting-up-tls-in-generic-cluster) to setup DNS and TLS for deploying workloads to your cluster.
  </Step>

  <Step title="Configure Node Pools">
    On managed Kubernetes (EKS, GKE, AKS), TrueFoundry automatically discovers node pools. For generic or on-prem clusters, you need to label your nodes so TrueFoundry can identify and group them into node pools.

    **1. Add a node pool label to every node**

    Pick a label key (e.g. `truefoundry.com/nodepool`) and apply it to each node with the pool name as the value:

    ```bash theme={"dark"}
    kubectl label nodes <NODE_NAME> truefoundry.com/nodepool=<POOL_NAME>
    ```

    Then tell TrueFoundry which label key to look for:

    1. Go to **Clusters** in the platform, click **Edit** on your cluster.
    2. Toggle on **Advanced Fields**.
    3. Under **Node Label Keys**, set the **Nodepool Selector Label** to your chosen key (e.g. `truefoundry.com/nodepool`).

    **2. Label GPU nodes with the GPU type**

    If you have GPU node pools, each GPU node must also carry a GPU type label so TrueFoundry can match workloads to the right hardware:

    ```bash theme={"dark"}
    kubectl label nodes <NODE_NAME> truefoundry.com/gpu_type=<GPU_TYPE>
    ```

    Supported `GPU_TYPE` values: `A10G`, `A10_12GB`, `A10_24GB`, `A10_4GB`, `A10_8GB`, `A100_40GB`, `A100_80GB`, `H100_80GB`, `H100_94GB`, `H200`, `L4`, `L40S`, `P100`, `P4`, `T4`, `V100`.
  </Step>

  <Step title="Start deploying workloads to your cluster">
    You can start by going [here](https://docs.truefoundry.com/docs/deploy-first-service#deploy-from-github)
  </Step>
</Steps>

***

## Setting up TLS in Generic Cluster

There are two primary ways to add TLS to the load balancer in a generic cluster.

<AccordionGroup>
  <Accordion title="Can I use cert-manager to set up TLS in a generic cluster?">
    Yes. We recommend using [cert-manager with Let's Encrypt](https://cert-manager.io/docs/getting-started/) if you have a DNS provider.
  </Accordion>

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
</AccordionGroup>
