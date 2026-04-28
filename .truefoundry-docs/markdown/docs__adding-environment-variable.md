> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Adding Environment Variable

> How to add and manage environment variables for jobs and services on TrueFoundry.

Environment can be added in both Python task and Container task, so when you define the python task config or container task config, we can give the env as follows:

* **Python task config**: you can pass the environment variables in the config by using env argument which takes dict as input, where key and value both are strings.

<CodeGroup>
  ```python Python lines theme={"dark"}
  from truefoundry.workflow import task, workflow, PythonTaskConfig, TaskPythonBuild
  from truefoundry.deploy import Resources

  task_config = PythonTaskConfig(
      image=TaskPythonConfig(...),
      env={
          "env_variable_name": "value_of_env_variable",
      },
    	...
  )

  @task(task_config=task_config)
  ...
  ```
</CodeGroup>

* **Container task config**: The container task config also takes env as input argument, similar to python task config , where you can pass the env variable in key value pair as dict.

<CodeGroup>
  ```python Python lines theme={"dark"}
  from truefoundry.workflow import task, workflow, ContainerTaskConfig, ContainerTask
  from truefoundry.deploy import Image

  container_task = ContainerTask(
      name="container_task",
      task_config=ContainerTaskConfig(
          image=Image(...),
          env={
              "env_variable_name": "value_of_env_variable",
          },
      ),
  )

  ...
  ```
</CodeGroup>

### Using secrets as an environment variable

you can also pass the secret as an env variable in a workflow. To use a secret, instead of passing the value of env, you can pass the secret FQN of the secret in place of value.

<CodeGroup>
  ```python python lines theme={"dark"}
  from truefoundry.workflow import task, workflow, PythonTaskConfig, TaskDockerFileBuild
  from truefoundry.deploy import Resources

  task_config = PythonTaskConfig(
      image=TaskDockerFileBuild(...),
      env={
          "env_variable_name": "value_of_env_variable",
          # Env with secret fqn
          "MY_SECRET": "tfy-secret://user:my-secret-group:my-secret",
      },
    	...
  )

  @task(task_config=task_config)
  ...
  ```
</CodeGroup>

***
