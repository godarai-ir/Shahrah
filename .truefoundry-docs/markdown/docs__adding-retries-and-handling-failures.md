> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Adding Retries And Handling Failures

> Learn how to configure retries and handle failures in production workflows.

TrueFoundry workflows provide robust mechanisms for handling task failures and retrying failed tasks. Here's how you can implement these features:

### Task Retries

You can configure automatic retries for individual tasks using the `retries` parameter in the `@task` decorator:

<CodeGroup>
  ```python Python lines theme={"dark"}
  @task(task_config=task_config, retries=3)
  def my_task():
     ...
  ```
</CodeGroup>

This configuration will attempt to execute the task up to 3 additional times if it fails.

Note: These retries are specifically user retries (if the code fails due to a code error).

If there are infrastructure issues, like spot-interruptions and errors like OOM killed, they are considered as infra failure and can be configured using a parameter called `max-node-retries-system-failures` which is a cluster level setting. The default value of this field is `3`. Please contact your system admin to change this value.

### Workflow Failure Handling

To handle failures at the workflow level, you can define a failure handler task and specify it using the `on_failure` parameter in the `@workflow` decorator:

```python lines theme={"dark"}
from truefoundry.workflow import task, workflow

@task(task_config=task_config)
def handle_failure():
   print("Handling Failure/Sending Notification")
	 ...

@workflow(on_failure=handle_failure)
def data_pipeline():
   ...
```

If your workflow fails, this will run the "handle\_failure" task towards the end. This can be used to clean up some resources or database entries or files and also send alert notification.

***
