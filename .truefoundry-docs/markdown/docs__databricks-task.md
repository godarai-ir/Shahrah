> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Creating A Databricks Task

> Learn how to trigger Databricks jobs from a TrueFoundry workflow using DatabricksJobTaskConfig, configure workspace and job settings, and set up authentication.

A Databricks task triggers an existing Databricks job from your Flyte workflow. You use the same `@task` decorator as other workflow tasks, but pass **`task_config=DatabricksJobTaskConfig(...)`** so the task runs by calling the Databricks Jobs API. The task triggers the job with `run_now()`, and by default waits for the run to complete (or times out and cancels the run).

### Example

<CodeGroup>
  ```python Python theme={"dark"}
  from truefoundry.workflow import (
      DatabricksJobTaskConfig,
      TaskPythonBuild,
      task,
      workflow,
  )

  @task(
      task_config=DatabricksJobTaskConfig(
          image=TaskPythonBuild(
              pip_packages=["truefoundry[workflow]"],
          ),
          workspace_host="https://<your-workspace>.cloud.databricks.com",
          service_account="flyte-databricks-sa",
          job_id="123",
          timeout_seconds=2000,
      )
  )
  def run_databricks_job():
      print("Databricks job complete")

  @workflow()
  def my_workflow():
      run_databricks_job()
  ```
</CodeGroup>

Trigger the workflow as usual; the task will trigger the specified Databricks job and, by default, wait for it to complete (or until `timeout_seconds`).

### How it runs

* The task calls Databricks `jobs.run_now()` with an **idempotency token** derived from the Flyte execution ID so the same logical run is not submitted twice.
* If **`skip_wait_for_completion`** is `False`: the task polls until the run terminates or is skipped, or until **`timeout_seconds`** elapses. On timeout, the run is canceled and a `RuntimeError` is raised. On success, the task completes after the Databricks run finishes.
* If **`skip_wait_for_completion`** is `True`: the task returns immediately after triggering the job; it does not wait for completion.

### DatabricksJobTaskConfig

<ParamField body="image" type="TaskPythonBuild | TaskDockerFileBuild" required>
  Image spec for the task (e.g. `TaskPythonBuild` with `pip_packages`). The task process runs in this image; the actual job execution happens in Databricks.
</ParamField>

<ParamField body="workspace_host" type="str" required>
  Databricks workspace URL (e.g. `https://<workspace>.cloud.databricks.com`).
</ParamField>

<ParamField body="job_id" type="str" required>
  The Databricks job ID to run (from your Databricks workspace Jobs).
</ParamField>

<ParamField body="service_account" type="str" required={false}>
  Kubernetes service account name for the task pod. Used for OIDC token exchange when `DATABRICKS_PERSONAL_ACCESS_TOKEN` is not set (workload identity federation).
</ParamField>

<ParamField body="job_parameters" type="dict[str, str]" required={false}>
  Optional parameters to pass to the job run (e.g. notebook params, jar params).
</ParamField>

<ParamField body="timeout_seconds" type="float" required={false}>
  Maximum seconds to wait for the job run to complete. If not set, defaults to 20 minutes. When the timeout is reached, the run is canceled and a `RuntimeError` is raised.
</ParamField>

<ParamField body="skip_wait_for_completion" type="bool" required={false}>
  If `False` (default), the task triggers the job and waits for it to complete (or timeout). If `True`, the task only triggers the job and returns immediately without waiting.
</ParamField>

<ParamField body="env" type="dict[str, str]" required={false}>
  Environment variables for the task (plain or secret refs). Use this to inject credentials such as **`DATABRICKS_PERSONAL_ACCESS_TOKEN`** so the task can authenticate to Databricks. See [Environment variables](https://docs.truefoundry.com/docs/env-variables).
</ParamField>

<ParamField body="resources" type="Resources" required={false}>
  Optional CPU/memory resources for the task pod.
</ParamField>

### Authentication

The task supports two ways to authenticate to Databricks:

1. **Personal Access Token (PAT)** — If **`DATABRICKS_PERSONAL_ACCESS_TOKEN`** is set in the task’s environment, the task uses it to authenticate. You can provide the PAT via the task config **`env`** (e.g. a secret reference) or via Flyte Propeller’s Kubernetes default env vars so it is available to the task pod.
2. **OAuth token federation** — If `DATABRICKS_PERSONAL_ACCESS_TOKEN` is not set, the task uses OAuth token federation (workload identity). This requires **`DATABRICKS_SERVICE_PRINCIPAL_CLIENT_ID`** in the task’s environment and a **service account** on the task pod (set **`service_account`** in `DatabricksJobTaskConfig`). The Kubernetes service account token is exchanged for a Databricks access token via the workspace OIDC endpoint; the Databricks workspace must have OIDC federation configured.

| Variable                                     | Required      | Description                                                                                                                                         |
| -------------------------------------------- | ------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| **`DATABRICKS_PERSONAL_ACCESS_TOKEN`**       | No\*          | Databricks Personal Access Token. If set, the task uses PAT auth. Can be provided via task **`env`** (e.g. secret ref) or cluster default env vars. |
| **`DATABRICKS_SERVICE_PRINCIPAL_CLIENT_ID`** | For OIDC only | Client ID of the Databricks service principal. Required when using OAuth token federation (i.e. when PAT is not set).                               |

\* Either PAT or OIDC (client ID + service account) must be configured.

For more on Databricks authentication:

* **[Authorize service principal access with OAuth (OAuth M2M)](https://docs.databricks.com/aws/en/dev-tools/auth/oauth-m2m)** — Creating an OAuth secret and using client ID + secret for workspace-level operations.
* **[Authenticate using OAuth token federation](https://docs.databricks.com/aws/en/dev-tools/auth/oauth-federation)** — Using workload identity tokens (e.g. Kubernetes OIDC) instead of Databricks secrets.

### Checklist

* Use **`@task(task_config=DatabricksJobTaskConfig(...))`** with required **`image`**, **`workspace_host`**, and **`job_id`**.
* For auth: set **`DATABRICKS_PERSONAL_ACCESS_TOKEN`** (e.g. in task **`env`** as a secret ref) for PAT auth, or use OAuth token federation with **`DATABRICKS_SERVICE_PRINCIPAL_CLIENT_ID`** and **`service_account`**.
* Optionally set **`env`**, **`job_parameters`**, **`timeout_seconds`**, **`skip_wait_for_completion`**, and **`resources`** as needed.
* Register the task in a **`@workflow`** and call it like any other task (e.g. `run_databricks_job()`).
