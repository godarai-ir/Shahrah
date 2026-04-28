> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Set Resources

> Configure CPU, memory, ephemeral storage, GPU, and shared memory resources for your TrueFoundry deployments.

When we deploy a service or job or any other application, we will need to define the resources that the application needs. This will ensure that Kubernetes can allot the correct set of resources and the application continues to run smoothly.

There are a few key resources you will need to configure:

| Resource              | Unit               | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| --------------------- | ------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **CPU**               | CPU cores          | This defines the processing power required by the application in terms of CPU cores. 1 CPU unit is equivalent to 1 virtual core. It's possible to ask for fractional CPU resources like 0.1 or even 0.01 (minimum is 0.01). While defining CPU, you will need to specify requests and limits. The number you define in the CPU request will always be reserved for the application. Your application can occasionaly take more CPU than what is requestes till the CPU limit, but not beyond that. **Any value above the request is opportunistic and not guaranteed to be available**. <br />For e.g, if CPU request is 0.5 and limit is 1, it means that the application has 0.5 CPU reserved for itself. CPU usage can go upto 1 if there is spare CPU available on the machine - else it will be throttled. |
| **Memory**            | Megabytes (MB)     | Defined as an integer and the unit is Megabytes. So a value of 1 means 1 MB of memory and 1000 means 1GB of memory. You need to specify memory requests and limits. The number you define in the memory request will always be reserved for the application. Your application can occasionaly take more memory than what is requested till the memory limit. **Any value above the request is opportunistic and not guaranteed to be available**. If the application takes up more memory than the limit, then the application will be killed and restarted. <br /> For e.g, if memory request is 500 MB and limit is 1000 MB, it means that you application will always have 500MB of RAM. You can have spikes of memory usage till 1 GB, beyond which the application will be killed and restarted.           |
| **Ephemeral Storage** | Megabytes (MB)     | Temporary disk space to keep code, artifacts, etc which is lost when the pod is terminated. Defined as an integer and the unit is Megabytes. A value of 1 means 1 MB of disk space and 1000 means 1GB of disk space. You need to specify ephemeral Storage requests and limits. If you specify 1 GB as request and 5 GB as limit, you will have guaranteed access to 1GB of disk space. You can go upto 5GB in case there is disk left on the machine, but we shouldn't rely on this. If the application tries to take up more than 5GB, the application will be killed.                                                                                                                                                                                                                                        |
| **GPU**               | GPU-Type and Count | This allows you to specify which GPU you want to provision for your application (GPUs can be of the following types: **K80, P4, P100, V100, T4, A10G, A100\_40GB, A100\_80GB**, etc.). Secondly, you need to specify the GPU-Count. Please note that if you ask for GPU count as 2 and type as A100, you will get a machine with atleast 2 A100 GPU cards. Its possible in some cloud providers that one machine has 4 A100 GPU cards. In this case, your application will use 2 of the 4 GPU cards and another application can use the rest 2 cards.                                                                                                                                                                                                                                                           |
| **Shared Memory**     | Megabytes (MB)     | Shared memory size is needed for data sharing between processes. This is useful in certain scenarios, for example, Torch Dataloaders rely on shared memory to efficiently share data between multiple workers during training. Defined as an integer and the unit is Megabytes. A value of 1 means 1 MB of shared memory and 1000 means 1GB of shared memory. In case your use-case requires Shared memory, and the usage exceeds the Shared memory size, your applications' replica will get killed.                                                                                                                                                                                                                                                                                                           |

## Setting Resources for your application

You can select the resources from the `Resources` section in the deployment form.

<img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/docs/set-resources-service.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=8ce26e8c4997052edf49c673554d1e41" width="3836" height="1936" data-path="images/docs/set-resources-service.png" />

The key things to fill up here are:

1. Choose CPU or GPU along with its type.
2. Fill up the CPU, Memory and Ephemeral Storage request and limit fields.

<Tip>
  Requests should be set based on the minimum resources your application needs to function normally under load. **Requests are guaranteed to be allocated to your application**.

  Limits should be set to prevent the application from using more resources than is acceptable. Setting the limits higher than requests allows the application to burst if it requires more resources for a short duration. **Limits define the maximum amount of resources a container is allowed to consume if there are extra resources available**.

  If you are unsure about the limit, its usually a good idea to set it to twice the request. So if you are CPU request is 0.5, you can set the limit to 1. If you are requesting 1 GB RAM, you can set the limit to 2 GB.
</Tip>

3. Select the capacity type - we have the following options here:

* `On-demand` - This will deploy the application on on-demand instances.
* `Spot` - This will deploy the application on spot instances.
* `Prefer Spot (with fallback on On-demand)` - This will try to deploy the application on spot instances. If the spot instances are not available, it will bring up the application on on-demand instances. If a spot instance is preempted, we will try to replace it with another spot instance, failing which, it will be shifted to an on-demand instance.

<Accordion title="What are On-demand and Spot Instances?">
  Spot and on-demand instances are two types of instances offered by cloud providers like AWS, GCP, and Azure.

  On-demand instances are the standard, pay-as-you-go compute resources offered by cloud providers. You pay for the compute time you use, with no long-term commitments.

  Spot instances are spare compute capacity offered by cloud providers at a significantly discounted price (often up to 90% off the on-demand price). However, the catch is that the cloud provider can reclaim these instances with minimal notice (typically a 2-minute warning) if demand for that instance type increases. **The spot prices also fluctuate based on demand and supply**.

  Here's an example of a snapshot of the spot vs on-demand prices of a few instance types in AWS, GCP and Azure:

  <Tabs>
    <Tab title="AWS">
      | Instance Type | On-Demand Price (approx.) | Spot Price (approx.) | Discount (%) |
      | ------------- | ------------------------- | -------------------- | ------------ |
      | t3.medium     | \$0.0416/hr               | \~$0.01-$0.02/hr     | \~50-75% off |
      | m6i.large     | \$0.077/hr                | \~$0.02-$0.03/hr     | \~60-75% off |
      | c6i.large     | \$0.085/hr                | \~$0.02-$0.03/hr     | \~65-75% off |
      | r6i.large     | \$0.113/hr                | \~$0.03-$0.04/hr     | \~65-70% off |
      | i4.large      | \$0.15/hr                 | \~$0.04-$0.05/hr     | \~65-70% off |
    </Tab>

    <Tab title="GCP">
      | Instance Type                 | On-Demand Price (relative)                  | Preemptible Price |
      | ----------------------------- | ------------------------------------------- | ----------------- |
      | E2 (e2-medium, e2-standard-4) | Base price                                  | \~70-80% discount |
      | N1 (n1-standard-4)            | Slightly higher than E2                     | \~70-80% discount |
      | N2 (n2-standard-4)            | Slightly higher than N1, better performance | \~70-80% discount |
      | C2 (c2-standard-4)            | Higher CPU performance price                | \~70-80% discount |
      | M1 (m1-ultramem-40)           | Higher price for memory                     | \~70-80% discount |
    </Tab>

    <Tab title="Azure">
      | Instance Type | On-Demand Price (approx.) | Spot Price (approx.) | Discount (%) |
      | ------------- | ------------------------- | -------------------- | ------------ |
      | D2s\_v2       | \$0.146/hr                | \~\$0.0248/hr        | \~83% off    |
      | NC4as T4 v3   | \$0.52/hr                 | \~\$0.14/hr          | \~70% off    |
      | D2ads v6      | \$0.1140/hr               | \~\$.0160/hr         | \~86% off    |
    </Tab>
  </Tabs>
</Accordion>

<Accordion title="Which capacity type option should I choose?" defaultOpen={true}>
  <Tip>
    We usually recommend using the `Prefer Spot (with fallback on On-demand)` or `Any`(in Azure) option for all development workloads.

    For production services, if you are running more than 5 replicas of a service, we recommend using the `Prefer Spot (with fallback on On-demand)` option. Else, we recommend using the `On-demand` option.
    For jobs, we recommend using the `On-demand` option if your jobs are short (less than 3 hours). For long-running jobs, you can use the `Prefer Spot (with fallback on On-demand)` option only if **you implement checkpointing and can resume from the last checkpoint where the earlier job was killed** because of spot instance preemption.

    **You almost never need to use the `Spot` option**.
  </Tip>

  Spot instances are a good option for nodes when:

  * You are running multiple replicas of a service - hence the chances of all the spot machines going down at the same time is extremely low.
  * You are running a long running training job and saving checkpoints which allows you to resume the training from the last checkpoint in case the machine is taken away.
  * Your service can handle going down and restarting on a different node.
</Accordion>

<Accordion title="How does TrueFoundry handle spot interruptions?">
  Truefoundry's handling of the spot instances depends on the cloud provider. Here's what happens in each of the cloud providers:

  <Tabs>
    <Tab title="AWS">
      Truefoundry uses [Karpenter](https://karpenter.sh/) to orchestrate spot and on-demand instance provisioning. Karpenter automatically monitors the availability of spot instances and can bring up another instance in case a spot instance is taken away. We configure Karpenter to prefer a balance between cost and availability - so that we don't just prioritize the cheapest option always which might have a high rate of preemption. AWS also sends a signal 2 minutes before a spot instance is taken away. Truefoundry configures Karpenter to start bringing up a new instance the moment AWS sends this signal to reduce the node provisioning time to a few seconds.

      **Because of this, the downtime of a pod to switch to another node is actually just a few seconds (depending on docker image download time)** and hence spot instance based scheduling works quite well in AWS. We have also seen a average uptime of around 12 hours for spot instances - although this varies based on the region and the instance type.
    </Tab>

    <Tab title="GCP">
      In GKE clusters, Truefoundry uses Node AutoProvisioning (NAP) to automatically provision spot and on-demand nodes and orchestrate the fallback between these two capacity types.
    </Tab>

    <Tab title="Azure">
      In Azure, we create nodepools of both spot and on-demand nodes. There is a fallback from Spot to On\_demand if spot nodes are not available, but there is no way to configure preference to spot instance as of now. When we choose the `Any` capacity type, it can either be a spot or on-demand instance. Azure usually has the best availability of spot instances - with uptimes ranging upto 30 days for spot instances.
    </Tab>
  </Tabs>
</Accordion>

## Advanced Options

### Selecting a specific instance type

<Warning> We usually do not recommend selecting a specific instance type, unless you have a specific reason to do so - like improved performance on a specific instance type.</Warning>

You can select an instance type by clicking on the `Advanced Options` button in Resources section and then selecting `Instance Family` option.

<img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/docs/select-instance-type-resources.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=aa04b8a059a1c890dfd6f772edd4b170" width="3840" height="1934" data-path="images/docs/select-instance-type-resources.png" />

### Selecting a Kubernetes nodepool

If you are using the nodepools in Kubernetes and want to specifically select a certain nodepool to deploy your application, you can do so by clicking on the `Nodepool Selector` option and then selecting the nodepool. We automatically parse the possible options from the available nodepools to show you the filtered set of nodepools - for e.g. if you select Spot, it will filter the set of nodepools that have spot instances.

<img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/docs/nodepool-selector.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=43130f84acb753c5109929adafb1b669" width="3838" height="1938" data-path="images/docs/nodepool-selector.png" />
