> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Configuring Resources

> Step-by-step guide for configuring resources, explaining configuration, best practices, and real-world usage on TrueFoundry.

When we define the task config, we will need to define the resources that the task will need for successful execution. This will ensure that Kubernetes can allot the correct set of resources and the application continues to run smoothly.

## Setting up the resources in Task Config

You can specify the resources for each task in task config. Each task config takes resource as an argument where you can define the resource limit based on the workload of that task.

Here is a sample resources object.

<CodeGroup>
  ```python Setting Resources {12-23} lines theme={"dark"}
  from truefoundry.deploy import Resources, NodeSelector, NvidiaGPU
  from truefoundry.workflow import task, workflow, PythonTaskConfig, TaskPythonBuild

  sample_task_config = PythonTaskConfig(
      # `TaskPythonBuild` helps specify the details of your Python Code.
      # These details will be used to templatize a DockerFile to build your Docker Image
      image=TaskPythonBuild(
          python_version="3.9",
          # Pip packages to install.`truefoundry[workflow]` is a mandatory dependency
          pip_packages=["truefoundry[workflow]"],
      ),
      resources=Resources( 
           cpu_request=2,
           cpu_limit=4,
           memory_request=12000,
           memory_limit=14000,
   .       # Supported GPUs are -> "T4", "A100_40GB", "A100_80GB", "V100"
           devices=[NvidiaGPU(name="T4", count=1)]],
           node=NodeSelector(
           	 # possible values "spot", "on_demand", "spot_fallback_on_demand"
             capacity_type="spot_fallback_on_demand"
           )
       ),
      service_account="<service-account>",
  )

  @task(task_config=sample_task_config)
  def my_task():
    ...
  ```
</CodeGroup>

Here is a brief explanation of what the different fields refer to:

| Resource              | Unit               | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| --------------------- | ------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **CPU**               | CPU cores          | Processing power required by the application. Defined in terms of CPU cores. 1 CPU unit is equivalent to 1 virtual core. It's possible to ask for fractional CPU resources like 0.1 or even 0.01. You will need to specify CPU requests and limits. The number you define in the CPU request will always be reserved for the application. Your application can occasionaly take more CPU than what is requestes till the CPU limit, but not beyond that. For e.g, if CPU request is 0.5 and limit is 1, it means that the application has 0.5 CPU reserved for itself. CPU usage can go upto 1 if there is CPU available on the machine - else it will be throttled.                                                                                                                                                                                                                                                                                                                                                                                                                             |
| **Memory**            | Megabytes (MB)     | Defined as an integer and the unit is Megabytes. So a value of 1 means 1 MB of memory and 1000 means 1GB of memory. You need to specify memory requests and limits. The number you define in the memory request will always be reserved for the application. Your application can occasionaly take more memory than what is requested till the memory limit. If the application takes up more memory than the limit, then the application will be killed and restarted. For e.g, if memory request is 500 MB and limit is 1000 MB, it means that you application will always have 500MB of RAM. You can have spikes of memory usage till 1 GB, beyond which the application will be killed and restarted.                                                                                                                                                                                                                                                                                                                                                                                        |
| **Ephemeral Storage** | Megabytes (MB)     | Temporary disk space to keep code, artifacts, etc which is lost when the pod is terminated. Defined as an integer and the unit is Megabytes. A value of 1 means 1 MB of disk space and 1000 means 1GB of disk space. You need to specify ephemeral Storage requests and limits. If you specify 1 GB as request and 5 GB as limit, you will have guaranteed access to 1GB of disk space. You can go upto 5GB in case there is disk left on the machine, but we shouldn't rely on this. If the application tries to take up more than 5GB, the application will be killed.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| **GPU**               | GPU-Type and Count | Graphics processing power required by the application. You need firstly specify which GPU you want to provision for your Application (GPUs can be of the following types: K80, P4, P100, V100, T4, A10G, A100\_40GB, A100\_80GB, etc.). You can find more details about these GPUs [here](https://cloud.google.com/compute/docs/gpus). In the case of **AWS or GCP**, you can do this by selecting the **GPU-Type** (read more [here](/docs/resources-cpu-memory-storage#selecting-a-specific-instance-type). In case you are using **Azure or some other cloud** you will just have to specify the **Nodepool** that contains the GPU (read more [here](/docs/resources-cpu-memory-storage#selecting-a-kubernetes-nodepool). Secondly, you need to specify the **GPU-Count**. Please note that if you ask for GPU count as 2 and type as A100, you will get a machine with atleast 2 A100 GPU cards. Its possible in some cloud providers that one machine has 4 A100 GPU cards. In this case, your application will use 2 of the 4 GPU cards and another application can use the rest 2 cards. |
| **Shared Memory**     | Megabytes (MB)     | Shared memory size is needed for data sharing between processes. This is useful in certain scenarios, for example, Torch Dataloaders rely on shared memory to efficiently share data between multiple workers during training. Defined as an integer and the unit is Megabytes. A value of 1 means 1 MB of shared memory and 1000 means 1GB of shared memory. In case your use-case requires Shared memory, and the usage exceeds the Shared memory size, your applications replica will get killed                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |

## Using spot instances to run tasks

TrueFoundry supports different capacity types for node selection, allowing you to optimize costs while maintaining reliability. You can configure this through the `NodeSelector` in your resource configuration.

### Available Capacity Types

1. **spot** - Uses spot instances exclusively

   1. Most cost-effective option (up to 70% cheaper than on-demand)
   2. Suitable for fault-tolerant workloads
   3. Risk of interruption if spot capacity is unavailable

2. **on\_demand** - Uses on-demand instances exclusively

   * Highest availability and reliability
   * No interruption risk
   * Standard pricing

3. **spot\_fallback\_on\_demand** - Uses spot instances with automatic fallback to on\_demand instances

   * Attempts to use spot instances first
   * Automatically falls back to on-demand if spot is unavailable
   * Balances cost savings with reliability

***
