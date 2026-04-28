> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Creating A Cron Workflow

> Schedule recurring workflows with cron on TrueFoundry. Learn setup, timing options, retries, and best practices.

A cron workflow is a type of workflow that is scheduled to run at specific intervals, similar to how cron jobs work in Unix-like systems. It allows you to automate the execution of workflows based on a predefined schedule, which can be expressed using cron syntax. Cron jobs are always scheduled in UTC timezone.

The cron job can be scheduled be passing the `execution_configs` in the workflow decorator

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

As you can see we have defined the execution config where the schedule is set to run every 10 minutes. The expression is define in cron expression format.

***
