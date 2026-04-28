> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Using Raw Container Task

> Step-by-step guide for container task, explaining configuration, best practices, and real-world usage on TrueFoundry.

In TrueFoundry workflow, a container task allows you to execute a command or script inside a Docker container without needing to write a custom Python function for that specific task. It provides a way to use pre-built Docker images and run commands directly within those containers.

## ContainerTaskConfig

* Just like python task config, in container task, we have `ContainerTaskConfig` which take, image spec as image, env, resource and service\_account as input.

<CodeGroup>
  ```python Python lines theme={"dark"}
  from truefoundry.workflow import task, workflow, ContainerTaskConfig, ContainerTask
  from truefoundry.deploy import Image

  container_task = ContainerTask(
      name="container_task",
      task_config=ContainerTaskConfig(
          image=Image(
              image_uri="...",
              command=[...],
          )
      ),
  )
  ```
</CodeGroup>

## Example

Now lets take a look at the example where we will just use the alpine base image and then print the number from 1 to 10 and will print their sum.

## Prerequisites

Before you proceed with the guide, make sure you have the following:

* **TrueFoundry CLI**: Set up and configure the TrueFoundry CLI tool on your local machine by following the [Setup for CLI](/docs/setup-cli) guide.
* **Workspace**: To deploy your workflow, you'll need a workspace. If you don't have one, you can create it using this guide: [Creating a Workspace](/docs/key-concepts#creating-a-workspace)or seek assistance from your cluster administrator.

## Creating the workflow

Create a workflow\.py where we will write the code for our workflow and place in the project root folder. (With other dependent files and requirements.txt)

`workflow.py`

<CodeGroup>
  ```python workflow.py lines theme={"dark"}
  from truefoundry.workflow import task, workflow, ContainerTaskConfig, ContainerTask
  from truefoundry.deploy import Image

  container_task = ContainerTask(
      name="container_task",
      task_config=ContainerTaskConfig(
          image=Image(
              image_uri="alpine:3.12",
              command=[
                  "/bin/sh",
                  "-c",
                  """
                  # Generate a sequence of numbers from 1 to 10
                  seq 1 10 | tee /var/data/numbers.txt

                  # Calculate the sum of these numbers
                  sum=$(seq 1 10 | awk '{s+=$1} END {print s}')

                  # Save the sum to a file
                  echo "Sum: $sum"
                  """,
              ],
          )
      ),
  )


  @workflow
  def sum_of_numbers() -> str:
      container_task()
      return "The sum of numbers from 1 to 10 has been calculated."
  ```
</CodeGroup>

* So here we have defined container\_task where we are using alpine image and the command which is basically a bash script to print the number from 1 to 10 and their sum.
* The container task can be called in the same way as you call the python task.

Now run the below command in the terminal to deploy your workflow, replace `<workfspace-fqn>` with the [workspace fqn](/docs/key-concepts#getting-workspace-fqn) which you can find on the UI.

<CodeGroup>
  ```shell Shell lines theme={"dark"}
  tfy deploy workflow \
    --name raw-container-wf \
    --file workflow.py \
    --workspace_fqn "Paste your workspace FQN here"
  ```
</CodeGroup>

***
