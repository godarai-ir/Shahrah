> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Using Fractional GPUs

> Share a single GPU across multiple workloads using Nvidia MIG and TimeSlicing for cost-efficient deployments.

Fractional GPUs enable us to allocate multiple workloads to a single GPU which can be useful in the following scenarios:

1. The workloads take around 2-3 GB of VRAM - so you can allocate multiple replicas of this workload on a single GPU which has around 16GB of VRAM or more.
2. Each workload has sparse traffic and its not able to max out the GPU usage.

There are two ways to use fractional GPUs:

1. **TimeSlicing**: In this approach, you can slice a GPU into a few fixed number of fractional parts and then choose a fraction of GPU for this workload. For e.g, we can decide to divide a GPU into 10 slices and then request 3 slices for one workload, 5 slices for workload2 and 2 slices for the workload3. This means that worload1 will use 0.3 GPU (compute + 30% of VRAM), workload2 will use 0.5 GPU and workload3 will use 0.2 GPU. However, timeslicing is only used for scheduling the workloads on the same GPU - it doesn't mean any actual isolation on the machine. For example, if the GPU machine has 16GB VRAM, its on the user to actually make sure that workload1 takes less than 4.8GB VRAM, workload2 takes less than 8 GB of VRAM and workload3 takes less than 3.2 GB of VRAM. If one workload starts taking more memory suddenly, it can lead to crashing of the other processes. The compute is also shared but one workload can go upto using the complete GPU if the other workloads are idle - its basically context-switching among the three workloads. You can read about this more [here.](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/gpu-sharing.html#understanding-time-slicing-gpus)
2. **MIG (Multi-Instance GPUs)**: This is a feature provided by Nvidia only on the A100, H100 and newer generation GPUs - it doesn't work on other GPUs. We can divide the GPUs into a fixed number of configurable parts as mentioned in the table below. The workloads can choose one of the slices and they will get compute and memory isolation. The instances are not exactly the complete fractions of the GPU, but more discrete units as mentioned in the table below. For e.g. let's say we divide one A100 GPU of 40GB into 7 parts - then we can place 7 workloads each using around 1/7 GPU and 5GB VRAM. Please note that we cannot simply provide 2 slices to one workload in this case and expect it to get 2/7 GPU and 10 GB VRAM. Each workload can only get one slice in this case. Here are the different **MIG profiles**:

A brief comparison chart between TimeSlicing and MIG is as follows:

| Feature                 | TimeSlicing                                                                                                                            | MIG (Multi-Instance GPU)                                                                                                         |
| ----------------------- | -------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| **GPU Support**         | Works on most GPUs.                                                                                                                    | Only supported on NVIDIA A100 and H100 GPUs.                                                                                     |
| **Isolation**           | **No real isolation.**  User is responsible for memory management. Potential for crashes if one workload exceeds its allocated memory. | **Strong Isolation.**  Compute and memory are isolated between instances.  Guaranteed resource allocation.                       |
| **Resource Allocation** | Divides GPU into fractional parts (e.g., 0.3, 0.5, 0.2). Workloads can use these fractional parts.                                     | Divides GPU into pre-defined, discrete instance types (as per NVIDIA's configurations). Workloads are assigned entire instances. |
| **VRAM Management**     | User-managed.  VRAM allocation is not enforced by the hardware.                                                                        | Hardware-enforced. Each instance has dedicated VRAM.                                                                             |
| **Compute Sharing**     | Compute is shared via context-switching. Workloads can potentially use the entire GPU when others are idle.                            | Compute is partitioned and isolated to each instance.  No sharing of compute resources beyond the instance's allocation.         |
| **Flexibility**         | More flexible in terms of resource allocation fractions (e.g., can request 0.3, 0.5, etc.).                                            | Limited to NVIDIA's pre-defined instance types.  Less flexible in terms of fine-grained resource requests.                       |

## Deploying on Fractional GPUs via MIG

<Steps>
  <Step title="Create a Nodepool with MIG enabled">
    We will need to create a separate nodepool for MIG enabled GPUs. Every GPU has different MIG profiles as mentioned in this page: [https://docs.nvidia.com/datacenter/tesla/mig-user-guide/index.html](https://docs.nvidia.com/datacenter/tesla/mig-user-guide/index.html). For e.g. here are the MIG profiles for A100 GPU:

    | GPU         | GPU Compute Fraction / Instance | Number of instances per GPU | GPU Memory / Instance | Configuration Name | GPU Instance Profile (for Azure) |
    | ----------- | ------------------------------- | --------------------------- | --------------------- | ------------------ | -------------------------------- |
    | A100 (40GB) | 1/7                             | 7                           | 5GB                   | 1g.5gb             | MIG1g                            |
    | A100 (40GB) | 2/7                             | 3                           | 10GB                  | 2g.10gb            | MIG2g                            |
    | A100 (40GB) | 3/7                             | 2                           | 20GB                  | 3g.20gb            | MIG3g                            |
    | A100 (80GB) | 1/7                             | 7                           | 10GB                  | 1g.10gb            | MIG1g                            |
    | A100 (80GB) | 2/7                             | 3                           | 20GB                  | 2g.20gb            | MIG2g                            |
    | A100 (80GB) | 3/7                             | 2                           | 40GB                  | 3g.40gb            | MIG3g                            |

    While creating the nodepool, we will need to select the MIG profile. Here are the steps to do it in different cloud providers:

    <Tabs>
      <Tab title="Azure">
        Create a Nodepool with MIG enabled using the argument `--gpu-instance-profile` of Azure CLI.

        ```bash Shell lines theme={"dark"}
        az aks nodepool add \
        --cluster-name <your cluster name> \
        --resource-group <your resource group> \
        --no-wait \
        --enable-cluster-autoscaler \
        --eviction-policy Delete \
        --node-count 0 \
        --max-count 20  \
        --min-count 1 \
        --node-osdisk-size 200 \
        --scale-down-mode Delete \
        --os-type Linux \
        --node-taints "nvidia.com/gpu=Present:NoSchedule" \
        --name a100mig7 \
        --node-vm-size Standard_NC24ads_A100_v4 \
        --priority Spot \
        --os-sku Ubuntu \
        --gpu-instance-profile MIG1g
        ```
      </Tab>

      <Tab title="GCP">
        Create a nodepool and pass the mig\_profile in `accelerator` by passing `gpu_partition_size=1g.5gb`\[OR one of the allowed values for MIG profile values mentioned above]

        ```bash Shell lines theme={"dark"}
        gcloud container node-pools create a100-40-mig-1g5gb \                                                                             INT ✘  
          --project=<enter your project name> \
          --region=<enter your region> \
          --cluster=<enter your cluster name here> \
          --machine-type=a2-highgpu-1g \
          --accelerator type=nvidia-tesla-a100,count=1,gpu-partition-size=1g.5gb \
          --enable-autoscaling \
          --total-min-nodes 0 \
          --total-max-nodes 4 \
          --min-provision-nodes 0 \
          --num-nodes 0
        ```
      </Tab>

      <Tab title="AWS">
        It is not trivial to currently support MIG GPUs on AWS in a managed way, although if you want to try the feature out -> Please refer to these [docs](https://aws.amazon.com/blogs/containers/maximizing-gpu-utilization-with-nvidias-multi-instance-gpu-mig-on-amazon-eks-running-more-pods-per-gpu-for-enhanced-performance/)
      </Tab>
    </Tabs>
  </Step>

  <Step title="Deploy your workload on the MIG nodepool">
    Once you have created the nodepool, you will be able to see the MIG nodepool in the available nodepools in `Resources` section in deployment.

    <Warning>
      It might take 10 mins for the newly created nodepool to be visible on the Truefoundry UI. You can force sync the nodepools by going to Platform -> Clusters -> Sync Cluster.
    </Warning>

    To deploy a workload that utilizes fractional GPU, start deploying your service/job on truefoundry and in the "Resources" section, select nodepool selector.
    You can now see the Fractional GPUs on the UI which you can select (as shown below)

    <Frame caption="Using MIG GPU">
      <img src="https://mintcdn.com/truefoundry/DdP_2rhue4AQQlob/images/3b48b7e9-97d76f5-Screenshot_2024-03-26_at_3.53.42_PM.png?fit=max&auto=format&n=DdP_2rhue4AQQlob&q=85&s=69e3f1cb03fbad37943551c9a7b2ce8f" width="2108" height="1012" data-path="images/3b48b7e9-97d76f5-Screenshot_2024-03-26_at_3.53.42_PM.png" />
    </Frame>
  </Step>
</Steps>

## Deploying on Fractional GPUs via Timeslicing

<Steps>
  <Step title="Create a Nodepool with Timeslicing enabled">
    <Tabs>
      <Tab title="Azure">
        Create a Nodepool with `device-plugin.config` pointing to the correct time-slicing config with Azure CLI.

        ```bash Shell lines theme={"dark"}
        az aks nodepool add \
        --cluster-name <your cluster name> \
        --resource-group <your resource group> \
        --no-wait \
        --enable-cluster-autoscaler \
        --eviction-policy Delete \
        --node-count 0 \
        --max-count 20  \
        --min-count 0 \
        --node-osdisk-size 200 \
        --scale-down-mode Delete \
        --os-type Linux \
        --node-taints "nvidia.com/gpu=Present:NoSchedule" \
        --name a100mig7 \
        --node-vm-size Standard_NC24ads_A100_v4 \
        --priority Spot \
        --os-sku Ubuntu \
        --labels nvidia.com/device-plugin.config=time-sliced-10
        ```
      </Tab>

      <Tab title="GCP">
        Here's the command to create a nodepool with timeslicing enabled on GCP.

        ```bash lines theme={"dark"}
        gcloud container node-pools create a100-40-frac-10 \                                                                                   ✔  
                --project=tfy-devtest \
                --region=us-central1 \
                --cluster=tfy-gtl-b-us-central-1 \
                --machine-type=a2-highgpu-1g \
                --accelerator type=nvidia-tesla-a100,count=1,gpu-sharing-strategy=time-sharing,max-shared-clients-per-gpu=10 \
                --enable-autoscaling \
                --total-min-nodes 0 \
                --total-max-nodes 4 \
                --min-provision-nodes 0 \
                --num-nodes 0
        ```
      </Tab>

      <Tab title="AWS">
        Create nodegroup on AWS EKS with the following label:

        ```yaml lines theme={"dark"}
        labels:
          "nvidia.com/device-plugin.config": "time-sliced-10"
        ```
      </Tab>
    </Tabs>
  </Step>

  <Step title="Deploy your workload on the Timeslicing nodepool">
    Once you have created the nodepool, you will be able to see the timesliced nodepool in the available nodepools in `Resources` section in deployment.

    <Warning>
      It might take 10 mins for the newly created nodepool to be visible on the Truefoundry UI. You can force sync the nodepools by going to Platform -> Clusters -> Sync Cluster.
    </Warning>

    To deploy a workload that utilizes fractional GPU, start deploying your service/job on truefoundry and in the "Resources" section, select nodepool selector.
    You can now see the timesliced GPUs on the UI which you can select (as shown below)

    <Frame caption="Using Timeslicing GPU">
      <img src="https://mintcdn.com/truefoundry/DdP_2rhue4AQQlob/images/2be0b32a-ea9657c-Screenshot_2024-03-26_at_3.53.51_PM.png?fit=max&auto=format&n=DdP_2rhue4AQQlob&q=85&s=7e9a00c0edb56282c4a8fe8a72f1da21" width="2116" height="1106" data-path="images/2be0b32a-ea9657c-Screenshot_2024-03-26_at_3.53.51_PM.png" />
    </Frame>
  </Step>
</Steps>
