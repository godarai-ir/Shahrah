> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Handling Failure using Failure Task/Node

> Handle workflow failures gracefully using retries, alerts, and recovery steps.

The failure task feature enables you to designate a specific task to execute in the event of a failure within your workflow.

This is useful when you want to handle a failure in a specific way, such as retrying the task or sending an email notification or cleaning up the resources after a failure.

<Warning>
  Important Note: Failure task or function accepts an optional `err` parameter
  which is an instance of `FlyteError`. To use the FlyteError, you will need the
  truefoundry version to be greater than or equal to 0.9.1 and the name of the
  parameter should be `err` only.
</Warning>

## How to use Failure Task/Node

* let's say we have a workflow with a task that fails

```python lines theme={"dark"}
import typing
from truefoundry.workflow import (
    PythonTaskConfig,
    TaskPythonBuild,
    task,
    workflow,
    FlyteError,
)

task_config = PythonTaskConfig(
    image=TaskPythonBuild(
        pip_packages=["truefoundry[workflow]==0.9.1"],
    ),
)


@task(task_config=task_config)
def normal_task(param_1: str):
    print(f"param_1: {param_1}")
    raise Exception("This task will fail")

@task(task_config=task_config)
def failure_task(param_1: str, param_2: str, err: typing.Optional[FlyteError] = None):
    print(f"param_1: {param_1}")
    print(f"param_2: {param_2}")
    print(f"error: {err.message}")
    print("This task will be executed in the event of a failure")

@workflow(on_failure=failure_task)
def my_workflow(param_1: str, param_2: str):
    normal_task(param_1)

```

* In the above example, we have defined two task functions `normal_task` and `failure_task`. One is a normal task that will fail and the other is a failure task that will be executed in the event of a failure.

<Warning>
  The failure task function input parameters should be completely same as the
  workflow function parameters else it will raise an error.
</Warning>

* Failure task function also accepts an optional `err` parameter which is an instance of `FlyteError` which you can import from `truefoundry.workflow`
  and this will contain the error message or stack trace of the failure.

* So this workflow will execute the `normal_task` and if it fails, the `failure_task` will be executed

In this way you can handle the failure of a task in a workflow.
