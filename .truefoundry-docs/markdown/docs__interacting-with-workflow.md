> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Interacting With Workflow

> Learn how to trigger, monitor, and debug workflows.

## Triggering the workflow

* Workflows can be triggered from UI, CLI and also via python SDK

## Triggering from UI

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/cm0ch5b231nnz128v6m3zkqsc?embed_v=2" typeofembed="iframe" height="500px" width="100%" src="https://app.supademo.com/embed/cm0ch5b231nnz128v6m3zkqsc?embed_v=2" style={{ border: "none", display: "flex", margin: "auto" }} />

## Triggering via cli

* Make sure the truefoundry\[Workflow] is installed and is set up or you can refer to [this](/docs/setup-cli) guide to set up the cli.

* Then copy the application fqn from the UI of the workflow

* <Frame caption="">
    <img src="https://mintcdn.com/truefoundry/FrY4JbiyZud2He3p/images/0a4f3d3c-94139916b440e1dad24ac4b7be5c9e62e106f2067a2d330d5776c52abcba0c0b-image.png?fit=max&auto=format&n=FrY4JbiyZud2He3p&q=85&s=996fe98927c118b3514256252ca8b435" width="1807" height="220" data-path="images/0a4f3d3c-94139916b440e1dad24ac4b7be5c9e62e106f2067a2d330d5776c52abcba0c0b-image.png" />
  </Frame>

  Then deploy using the following command

<CodeGroup>
  ```html Shell lines theme={"dark"}
  tfy trigger workflow <application-fqn>
  ```
</CodeGroup>

* if workflow accepts the input parameters you can pass them as arguments

<CodeGroup>
  ```html Shell lines theme={"dark"}
  tfy trigger workflow <application-fqn> --input1 <input1> --input2 <input2>
  ```
</CodeGroup>

## Triggering via Python SDK

<CodeGroup>
  ```python trigger.py lines theme={"dark"}
  from truefoundry.deploy import trigger_workflow

  trigger_workflow(application_fqn="<application-fqn>", inputs={"input-key": "input-value"})
  ```
</CodeGroup>

* Save the above file and run the following command

<CodeGroup>
  ```shell Shell lines theme={"dark"}
  python trigger.py
  ```
</CodeGroup>

## Checking workflow graph

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/cm0dm6f411y2u128vkgbctl9b?embed_v=2" typeofembed="iframe" height="458px" width="100%" src="https://app.supademo.com/embed/cm0dm6f411y2u128vkgbctl9b?embed_v=2" style={{ border: "none", display: "flex", margin: "auto" }} />

* You can check the workflow graph by clicking on any of the workflow on workfow page and then clicking on graph page as shown in above tutorial
* The workflow execution graph shows the static graph of the workflow where each rectangle represents the tasks and the edges represent the execution order of each tasks, where a edge from one node to another node represent the connection from one task to another task.

## Checking live execution graph of a workflow

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/cm0dmo3rn1yfo128vjvu9r8eb?embed_v=2" typeofembed="iframe" height="434px" width="100%" src="https://app.supademo.com/embed/cm0dmo3rn1yfo128vjvu9r8eb?embed_v=2" style={{ border: "none", display: "flex", margin: "auto" }} />

* Graph of the workflow can be viewed on the execution page of that particular workflow.
* The rectangles represent the task and the edges represent the connection between nodes, so the edge from one task to another tells about the next task or step in the workflow after the current task is executed successfully.
* The yellow color around the task means that the task is running and if it turns red that means it is in a failed state and workflow has failed and if it is green in color then it is executed successfully.
* The dotted rectangle represents that the task is a conditional task and the highlighted border suggests which tasks were executed based on the condition defined in the workflow.
* If there are multiple nested conditional tasks then on clicking the task that was executed, then clicking on the conditional task will open another nested condition and now there will be two dotted rectangles around tasks as seen in above tutorial and this will keep on increasing until the task are executed successfully and there are no nested conditions left.

## Checking Execution Logs, Metrics and Events

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/cm0dmwx981ygi128vz7zb70nv?embed_v=2" typeofembed="iframe" height="428px" width="100%" src="https://app.supademo.com/embed/cm0dmwx981ygi128vz7zb70nv?embed_v=2" style={{ border: "none", display: "flex", margin: "auto" }} />

* You can check the logs, metrics, and events for each task individually by clicking on that particular task or node.
* For conditional task, you can check the metrics and logs of the task that were executed in the workflow.
* For map task, when you click on map task, a new page opens where you can see all the subtasks and each sub-task has its own logs metrics and events.

## Terminating the workflow

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/cm0dngbza1ym9128v036zxjxk?embed_v=2" typeofembed="iframe" height="431px" width="100%" src="https://app.supademo.com/embed/cm0dngbza1ym9128v036zxjxk?embed_v=2" style={{ border: "none", display: "flex", margin: "auto" }} />

***
