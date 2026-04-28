> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Passing Files/Artifacts Between Tasks

> Use FlyteFile and FlyteDirectory to pass files and directories between tasks in TrueFoundry workflows.

In Workflow, FlyteFile and FlyteDirectory are special types representing files and directory that can be passed between tasks in a workflow. It helps manage and track files and directories, ensuring that they can be accessed and moved across different execution environments.

## Passing files across tasks

<CodeGroup>
  ```python Python lines theme={"dark"}
  from truefoundry.workflow import (
      PythonTaskConfig,
      TaskPythonBuild,
      task,
      workflow,
      FlyteFile
  )
  from truefoundry.deploy import Resources

  task_config = PythonTaskConfig(
      image=TaskPythonBuild(
          python_version="3.9",
          pip_packages=["truefoundry[workflow]==0.4.7rc0"],
      ),
      resources=Resources(cpu_request=0.45),
      service_account="default",
  )

  @task(task_config=task_config)
  def create_file() -> FlyteFile:
      file_path = "/tmp/sample.txt"
      with open(file_path, "w") as f:
          f.write("Hello World!")
      return FlyteFile(file_path)


  @task(task_config=task_config)
  def read_and_print_file(flyte_file: FlyteFile):
      with open(flyte_file, "r") as f:
          content = f.read()
          print("File content:", content)
          

  @workflow
  def simple_file_workflow():
      file = create_file()
      read_and_print_file(flyte_file=file)

  if __name__ == "__main__": 
      simple_file_workflow()
  ```
</CodeGroup>

* In the above example we are creating a `sample.txt` file in a `create_file` task and then returning that file as FlyteFile and then in `read_and_print_file` task we are taking the flyteFile as input and then printing the content of the file.

## Passing directory across tasks

<CodeGroup>
  ```python Python lines theme={"dark"}
  from truefoundry.workflow import (
      PythonTaskConfig,
      TaskPythonBuild,
      task,
      workflow,
      FlyteDirectory
  )
  from truefoundry.deploy import Resources

  task_config = PythonTaskConfig(
      image=TaskPythonBuild(
          python_version="3.9",
          pip_packages=["truefoundry[workflow]==0.4.7rc0"],
      ),
      resources=Resources(cpu_request=0.45),
      service_account="default",
  )

  @task(task_config=task_config)
  def create_directory() -> FlyteDirectory:
      import os

      dir_path = "/tmp/sample_directory"
      os.makedirs(dir_path, exist_ok=True)
      
      # Create multiple text files in the directory
      for i in range(3):
          with open(os.path.join(dir_path, f"file_{i}.txt"), "w") as f:
              f.write(f"Content of file {i}\n")

      return FlyteDirectory(dir_path)


  @task(task_config=task_config)
  def read_and_print_directory(directory: FlyteDirectory) -> str:
      import os

      # List all files in the directory
      for filename in os.listdir(directory):
          file_path = os.path.join(directory, filename)
          with open(file_path, "r") as f:
              content = f.read()
              print(f"Contents of {filename}:")
              print(content)
      return "Done reading and printing the directory."

  @workflow
  def simple_directory_workflow():
      directory = create_directory()
      read_and_print_directory(directory=directory)
  ```
</CodeGroup>

* In the above example, we are creating multiple files in `/tmp/sample_directory` directory in `create_directory` task and then we are passing this directory in `read_and_print_directory` task to print the files in that directory and the content of these files.

***
