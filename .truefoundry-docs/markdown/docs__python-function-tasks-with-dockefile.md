> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Using Dockerfile For Python Function Task

> Use a custom Dockerfile with Python function tasks in TrueFoundry workflows to install extra dependencies.

In TrueFoundry workflow you can pass the dockerfile path in the task config, this is useful when you want to install a binary and use it in your python function task like using a jq command, etc.

## TaskDockerFileBuild

* To use dockerfile for python task, we need to import `TaskDockerFileBuild` and use it in `PythonTaskConfig`.

<CodeGroup>
  ```python Python lines theme={"dark"}
  from truefoundry.workflow import task, workflow, PythonTaskConfig, TaskDockerFileBuild
  from truefoundry.deploy import Resources

  task_config = PythonTaskConfig(
      image=TaskDockerFileBuild(
          dockerfile_path="Dockerfile",
      ),
      resources=Resources(cpu_request=0.5)
  )
  ```
</CodeGroup>

## Example

In this guide, we will see how to write a Python function task with a dockerfile. In this example, we will take a string as an input and then will calculate the SHA56 hash from the input. We will be passing the

## Prerequisites

Before you proceed with the guide, make sure you have the following:

* **TrueFoundry CLI**: Set up and configure the TrueFoundry CLI tool on your local machine by following the [Setup for CLI](/docs/setup-cli) guide.
* **Workspace**: To deploy your workflow, you'll need a workspace. If you don't have one, you can create it using this guide: [Creating a Workspace](/docs/key-concepts#creating-a-workspace) or seek assistance from your cluster administrator.

## Creating the workflow

Create a workflow\.py where we will write the code for our workflow and place it in the project root directory with your Dockerfile.

Your directory structure will then appear as follows:

```plain Directory Structure theme={"dark"}
.
├── workflow.py
└── Dockerfile
```

`Dockerfile`

<CodeGroup>
  ```python Dockerfile lines theme={"dark"}
  # Use an official Python runtime as a parent image
  FROM python:3.10-slim

  # Install jq binary
  RUN apt-get update && apt-get install -y jq

  # Set the working directory
  WORKDIR /app

  RUN pip install "truefoundry[workflow]==0.4.8"

  # Copy the current directory contents into the container
  COPY . /app

  # Set the default command to run Python
  CMD ["python"]
  ```
</CodeGroup>

`workflow.py`

<CodeGroup>
  ```python workflow.py lines theme={"dark"}
  from truefoundry.workflow import task, workflow, PythonTaskConfig, TaskDockerFileBuild
  from truefoundry.deploy import Resources

  task_config = PythonTaskConfig(
      image=TaskDockerFileBuild(
          dockerfile_path="Dockerfile",
      ),
      resources=Resources(cpu_request=0.45)
  )


  @task(task_config=task_config)
  def run_jq(input_json: str) -> str:
      import subprocess

      # Run the jq binary using Python's subprocess module
      process = subprocess.run(
          ["jq", "."],  # jq command to pretty-print JSON
          input=input_json,
          text=True,
          capture_output=True,
      )
      return process.stdout


  @workflow
  def my_workflow(input_json: str) -> str:
      return run_jq(input_json=input_json)
  ```
</CodeGroup>

* As you can see, we have given the dockerfile\_path argument in PythonTaskConfig where the path to the docker file is used as its value.

Now run the below command in the terminal to deploy your workflow, replace `<workfspace-fqn>` with the [workspace fqn](/docs/key-concepts#getting-workspace-fqn) which you can find on the UI.

<CodeGroup>
  ```shell Shell lines theme={"dark"}
  tfy deploy workflow \
    --name dockerfile-build-wf \
    --file workflow.py \
    --workspace_fqn "Paste your workspace FQN here"
  ```
</CodeGroup>

***
