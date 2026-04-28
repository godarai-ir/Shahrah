> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Introduction To Workflow

> Powered by Flyte (An OpenSource Workflow Orchestrator)

A TrueFoundry workflow is a structured sequence of tasks that enables efficient management and execution of complex computational processes, especially in data processing and machine learning. It offers a scalable, maintainable, and reusable approach to orchestrating and automating tasks.

**Key features of TrueFoundry workflows:**

* **Task Orchestration:** Seamlessly manage and execute tasks in a defined order using Directed Acyclic Graphs **(DAGs)**.
* **End-to-End Automation:** Streamline the entire lifecycle of data and machine learning workflows, from data ingestion to model deployment.
* **Scalability and Reusability:** Effortlessly scale workflows to handle large datasets or complex tasks.

### Architecture

TrueFoundry Workflows are based on an OpenSource Workflow Orchestrator [Flyte](https://docs.flyte.org/en/latest/)

* We ship Flyte's Control Plane Components with TrueFoundry's Control Plane (No additional Setup Needed).
* Each Kubernetes cluster where you need to run workflows needs the following setup to be done. (Installation of Flyte's Data Plane Components). Follow this [document](/docs/infra-set-up-for-workflows) for the same.

### Key Consideration while building a workflow.

* Install the truefoundry workflow and set up the CLI, refer to [this](/docs/setup-cli) guide for setting up the CLI.
* You need to define the workflow in a python file like [this](/docs/creating-your-first-workflow#workflowpy). Functions representing tasks and workflows need to decorated with `@workflow` and `@task`decorators.
* Define task and task config to configure the usage of task in workflow, check the [different types of task](/docs/task-config)config that are available for usage in workflow.
* To know about how to interact with the workflow, refer to [Interacting with workflow](/docs/interacting-with-workflow).
* To run a raw container as a task, refer to the [Using raw container task guide](/docs/container-task).
* To run workflow at a fix interval of time, create a cron workflow, to learn more about cron workflow, refer [this](/docs/cron-workflow).
* To run PySpark jobs as a task in your workflow, use a Spark task with `PySparkTaskConfig`; refer to the [Spark task guide](/docs/spark-task) for setup and examples.
* To trigger a Databricks job from workflow, use a Databricks task with `DatabricksJobTaskConfig`; refer to the [Databricks task guide](/docs/databricks-task) for setup and examples.
* To speed up the execution of heavy task, you can use map task to create sub task and run simultaneously, refer to [this](/docs/creating-a-map-task) guide to see how you can define map task.
* To have a conditional execution of task or to select task based on conditions, use [conditional task.](/docs/introduction-to-workflow)

***
