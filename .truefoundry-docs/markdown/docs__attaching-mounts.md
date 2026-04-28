> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Attaching Mounts

> Step-by-step guide for attaching mounts, explaining configuration, best practices, and real-world usage on TrueFoundry.

In some Workflows or tasks, you might need a file to be present at a certain path for the workflow to work. Or you have data in directories that need to be mounted for the task to complete. TrueFoundry allows you to mount files in workflow in the following ways:

## Volume Mount

If your task needs access to multiple files of data or multiple directories, you can use a volume to store the data and then mount the volume at the desired path. You can learn more about Volumes [here](/docs/introduction-to-volume).

To use a persistent volume, we will first need to create one and then attach it to your deployment. You can learn how to create volumes using the [Creating a Volume](/docs/creating-a-volume)guide

you can mount the volume for a task in the following way.

<CodeGroup>
  ```python Python lines theme={"dark"}
  from truefoundry.deploy import VolumeMount, Resources
  from truefoundry.workflow import task, workflow, PythonTaskConfig, TaskDockerFileBuild

  task_config = PythonTaskConfig(
      image=TaskDockerFileBuild(
  		...
      mounts=[
        VolumeMount(
          mount_path="/model", #or your desired path
          volume_fqn="your-volume-fqn"
        )
      ]
  )
  ```
</CodeGroup>

## Secret Mount

This is similar to string mount, except, in this case, you will directly provide a TrueFoundry Secret FQN. You can read more about Secrets and how to create them [here](/docs/manage-secrets).

The content in the secret will be dumped into a file and mounted in the provided location. A good use-case of this is for mounting Google credentials file which you might need to access Google services in a task.

<CodeGroup>
  ```python Python lines theme={"dark"}
  from truefoundry.deploy import SecretMount, Resources
  from truefoundry.workflow import task, workflow, PythonTaskConfig, TaskDockerFileBuild

  task_config = PythonTaskConfig(
      image=TaskDockerFileBuild(
  		...
      mounts=[
        SecretMount(
        	mount_path="/etc/google-credentials.json",
          secret_fqn="tfy-secret://user:my-secret-group:my-secret"
        )
      ]
  )
  ```
</CodeGroup>

## String Mount

This can be useful if you need a small configuration file to be present at a certain file path. To configure this, you need to provide the path where it needs to be mounted and the string data that should be in that file. A example of this can be if you want to apply yaml specs for a service through task, then you can use string mount to define the specs.

<CodeGroup>
  ```python Python lines theme={"dark"}
  from truefoundry.deploy import StringDataMount, Resources
  from truefoundry.workflow import task, workflow, PythonTaskConfig, TaskDockerFileBuild

  task_config = PythonTaskConfig(
      image=TaskDockerFileBuild(
  		...
      mounts=[
  				StringDataMount(
        		mount_path="/etc/nginx/conf.d/my_config.conf",
          	data="server { listen 80; }"
        )
      ]
  )
  ```
</CodeGroup>

***
