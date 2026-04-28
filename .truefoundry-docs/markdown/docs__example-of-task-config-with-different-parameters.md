> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Example Of Task Config With Different Parameters

> See task config examples with different parameters and execution options.

A task config takes various parameters as input, in this example, we will see how to use or input all these parameters in detail.

## Example

<CodeGroup>
  ```python Python lines theme={"dark"}
  from truefoundry.workflow import (
      PythonTaskConfig,
      TaskPythonBuild,
      task,
      workflow,
  )
  from truefoundry.deploy import Resources, SecretMount, VolumeMount, StringDataMount, NvidiaGPU

  task_config = PythonTaskConfig(
      image=TaskPythonBuild(
          python_version="3.9",
          pip_packages=["truefoundry[workflow]==0.4.7rc0"],
          #To install pip packages from a requirements file uncomment this and provide the path to the requirements file
          # requirements_path="requirements.txt", 
          apt_packages=[
              "git",
              "ffmpeg",
          ]
      ),
      resources=Resources(
          cpu_request=0.45,
          cpu_limit=0.5,
          memory_request=100,
          memory_limit=150,
          ephemeral_storage_request=100,
          ephemeral_storage_limit=150,
          devices=[NvidiaGPU(name="T4", count=1)],
          capacity_type="spot_fallback_on_demand"
      ),
      service_account="default",
      mounts=[
          # To use volume mount or secret mount uncomment the following lines and provide the correct values
          # VolumeMount(mount_path="/tmp/data", volume_fqn="tfy-volume://tfy-usea1-devtest:nikp-wf-test:wf-test"),
          # SecretMount(mount_path="/tmp/secret", secret_fqn="tfy-secret://truefoundry:test-secret-1730850194619:abctestkey3"),
          StringDataMount(mount_path="/tmp/stringdata", data="Enter your data here"),
      ]
  )

  # Task 1: Extract data (simulates loading data from an external source)
  @task(task_config=task_config, retries=3)
  def print_string_mount_data() -> str:
      with open("/tmp/stringdata", "r") as f:
          data = f.read()
          print(f"Data: {data}")
      return data


  @workflow
  def simple_workflow() -> str:
      data = print_string_mount_data()
      return data

  if __name__ == "__main__":
      print(f"data_pipeline(): {simple_workflow()}")
  ```
</CodeGroup>

* As you can see in the above example, you can pass the path of the requirements file to the `requirements_path` parameter in image section of task config to install libraries mentioned in requirements files. Make sure that this file path is relative to the root directory.
* You can also define the apt packages you want to install by defining them `apt_packages` parameter.
* To mount the secrets or volume or string data, you can define them in the mounts parameter as seen in example.
* you can also set the retries on failure for a particular task by passing the `retries` argument in the `@task` decorator as seen in above example.

***
