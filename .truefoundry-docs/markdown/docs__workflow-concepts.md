> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Workflow Concepts

> Understand core workflow concepts including workflows, tasks, decorators, and task execution flow in TrueFoundry.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/OHzlp6GY5G-JfKle/images/bf19dce7-1ce5177d229cc45b21a43e361f9e2a075dbb4c65aa2ec62442ec5a699645a71d-image.png?fit=max&auto=format&n=OHzlp6GY5G-JfKle&q=85&s=0969443e9d8951688306cd7b57211395" width="1013" height="475" data-path="images/bf19dce7-1ce5177d229cc45b21a43e361f9e2a075dbb4c65aa2ec62442ec5a699645a71d-image.png" />
</Frame>

## Workflow

**Definition**: A workflow is a composition of tasks, arranged in a specific order, that defines the execution logic. Workflows can include conditional logic, loops, and parallel execution.

* Right now only static workflows are supported.
* There can be only one Workflow in an application.
* Workflow can be defined by wrapping a function with the @workflow decorator.

<CodeGroup>
  ```python Python lines theme={"dark"}
  from truefoundry.workflow import task, workflow


  @task(...)
  def say_hello() -> str:
      return "Hello, World!"


  @workflow
  def hello_world_wf() -> str:
      res = say_hello()
      return res
  ```
</CodeGroup>

## Task

**Definition**: Task is the smallest unit of execution. It represent a single operation or function, such as a data transformation, model training step, or data retrieval, etc.

* Types of Tasks:

  * **Python Tasks:** :Regular Python functions decorated with @task and are called in the workflow function.
  * **Container Tasks**: A Container Task is a task that runs a Docker container as part of the workflow execution. This allows you to encapsulate any piece of logic or computation within a Docker container and execute it within the workflow.

Each task takes a `task_config` as a parameter to learn more about task config click [here](/docs/task-config).

<CodeGroup>
  ```python Python lines theme={"dark"}
  from truefoundry.workflow import task, workflow, PythonTaskConfig, TaskDockerFileBuild,  ContainerTaskConfig, ContainerTask
  from truefoundry.deploy import Resources, Image

  #Python task config
  task_config = PythonTaskConfig(
      ...
  )
  @task(task_config=task_config)
  def python_task() -> str:
      ...
    
  #Container task with task config
  container_task = ContainerTask(
      name="container_task",
      task_config=ContainerTaskConfig(...),
  )
  ```
</CodeGroup>

## Map Task

A map task is a specialized task that enables the parallel execution of a single task across multiple inputs, this task is used to speed up the execution and it is used when the input is too large or the function is a heavy task and requires a huge amount of resources and time.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/DdP_2rhue4AQQlob/images/423e4e2f-14367b896e57ca3c2addf615d778b1145a9a2ac1a4fdfe267e9902947416dc8e-image.png?fit=max&auto=format&n=DdP_2rhue4AQQlob&q=85&s=c354113d45e3f316b0b4142f582966a7" width="480" height="643" data-path="images/423e4e2f-14367b896e57ca3c2addf615d778b1145a9a2ac1a4fdfe267e9902947416dc8e-image.png" />
</Frame>

To learn more about Map task and its usage, refer to [Creating a Map Task](/docs/creating-a-map-task) guide.

## Conditional Task

A conditional task allows you to decide the execution path of the workflow based on the condition defined. It his used when you have different execution paths that are dependent on the input or output of the previous task.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/jw406UAsc7ErYUq8/images/aa966969-b453d17c08115f3e1e020fd936884a3ae94b4e84ab108e30a55e86a7b6a24227-image.png?fit=max&auto=format&n=jw406UAsc7ErYUq8&q=85&s=fba478dd80c31dece1ce47bf603ed8d2" width="574" height="693" data-path="images/aa966969-b453d17c08115f3e1e020fd936884a3ae94b4e84ab108e30a55e86a7b6a24227-image.png" />
</Frame>

To learn more about Conditional task and its usage, refer to [Creating a Conditional Task](/docs/conditional-task) guide.

## Cron Workflow

* A cron workflow is a type of workflow that is scheduled to run at specific intervals, similar to how cron jobs work in Unix-like systems. It allows you to automate the execution of workflows based on a predefined schedule, which can be expressed using cron syntax. Cron jobs are always scheduled in UTC timezone.
* Each workflow takes an optional execution config as parameter in which you can define the schedule or interval at which you want to run the workflow.
* The cron Workflow can be scheduled be passing the `execution_configs` in the workflow decorator

<CodeGroup>
  ```python Python lines theme={"dark"}
  from truefoundry.workflow import workflow, ExecutionConfig

  @workflow(
      execution_configs=[
          ExecutionConfig(
              schedule="*/10 * * * *",
          )
      ]
  )
  def your_workflow():
      ...
  ```
</CodeGroup>

***
