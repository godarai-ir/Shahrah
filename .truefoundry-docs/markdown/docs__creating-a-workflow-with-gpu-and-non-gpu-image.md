> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Creating A Workflow With Different Container Images

> Step-by-step guide for creating a workflow with gpu and non gpu image, explaining configuration, best practices, and real-world usage on…

This guide helps you to create a workflow with tasks that have different task configs and different container applications. This can have variety of applications like keeping different configuration (resources and container image) for steps requiring GPUs and steps not requiring GPUs

In this example, We will write a simple workflow to get the latest price of Bitcoin, Ethereum, and Litecoin, the main aim of this example is to understand the different types of configs. we will create two configs, one for GPU and one for CPU like this.

## Prerequsite

Before you proceed with the guide, make sure you have the following:

* **TrueFoundry CLI**: Set up and configure the TrueFoundry CLI tool on your local machine by following the [Setup for CLI](/docs/setup-cli) guide.
* **Workspace**: To deploy your workflow, you'll need a workspace. If you don't have one, you can create it using this guide: [Creating a Workspace](/docs/key-concepts#creating-a-workspace)or seek assistance from your cluster administrator.

## Creating the workflow

Create a workflow\.py where we will write the code for our workflow and place in the project root folder. (With other dependent files and requirements.txt)

**`workflow.py`**

<CodeGroup>
  ```python workflow.py lines theme={"dark"}
  from truefoundry.workflow import (
      task,
      workflow,
      PythonTaskConfig,
      TaskPythonBuild,
  )
  from truefoundry.deploy import Resources, NvidiaGPU
  import requests
  import json

  cpu_config = PythonTaskConfig(
      image=TaskPythonBuild(
          python_version="3.9",
          pip_packages=["truefoundry[workflow]==0.4.8"],
          # requirements_path="requirements.txt"
      ),
      resources=Resources(cpu_request=0.5)
  )

  gpu_config = PythonTaskConfig(
      image=TaskPythonBuild(
          python_version="3.9",
          pip_packages=["truefoundry[workflow]==0.4.8", "pynvml==11.5.0"],
          # requirements_path="requirements.txt",
          cuda_version="11.5-cudnn8",
      ),
      env={
          "NVIDIA_DRIVER_CAPABILITIES": "compute,utility",
          "NVIDIA_VISIBLE_DEVICES": "all",
      },
      resources=Resources(cpu_request=1.5, devices=[NvidiaGPU(name="T4", count=1)])
  )


  # Python Task: Fetch real-time prices for multiple cryptocurrencies from the CoinGecko API
  @task(task_config=cpu_config)
  def fetch_crypto_data() -> str:
      response = requests.get(
          "https://api.coingecko.com/api/v3/simple/price?ids=bitcoin,ethereum,litecoin&vs_currencies=usd"
      )
      return response.text


  # Python Task: Process the data to extract prices
  @task(task_config=gpu_config)
  def extract_prices(data: str) -> dict:
      from pynvml import nvmlDeviceGetCount, nvmlInit

      nvmlInit()
      assert nvmlDeviceGetCount() > 0

      json_data = json.loads(data)
      prices = {
          "bitcoin": json_data["bitcoin"]["usd"],
          "ethereum": json_data["ethereum"]["usd"],
          "litecoin": json_data["litecoin"]["usd"],
      }
      return prices


  # Workflow: Combine all tasks
  @workflow
  def crypto_workflow() -> dict:
      data = fetch_crypto_data()
      prices = extract_prices(data=data)
      return prices
  ```
</CodeGroup>

* As you can see for the `fetch_crypto_data` function we have defined the resource which does not use GPU whereas task `extract_prices` uses GPU and hence we have defined the env variable and we have also used the pynvml package to check whether GPU is present or not by asserting the condition that `assert nvmlDeviceGetCount() > 0`.

Now run the below command in the terminal to deploy your workflow, replace `<workfspace-fqn>` with the [workspace fqn](/docs/key-concepts#getting-workspace-fqn) which you can find on the UI.

<CodeGroup>
  ```shell Shell lines theme={"dark"}
  tfy deploy workflow \
    --name multi-image-workflow \
    --file workflow.py \
    --workspace_fqn "Paste your workspace FQN here"
  ```
</CodeGroup>

***
