> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Running Workflow Locally

> Run TrueFoundry workflows locally for rapid testing and debugging before deploying to the platform.

You can run workflow locally by simply running your worklow python file, you can do:

<CodeGroup>
  ```shell Shell lines theme={"dark"}
  python3 workflow.py
  ```
</CodeGroup>

### Important points to remember:

* Only your code logic is tested while running the workflow but the environment in which worklow runs can be different from your local environment and there can be difference in execution speed or time taken for workflow to complete.
* If you are running workflow locally let's say in Windows and then when the workflow runs in Linux environment, then there are chances of code not working properly if any of the library you are using is only compatible with windows or behaves differently in linux.

***
