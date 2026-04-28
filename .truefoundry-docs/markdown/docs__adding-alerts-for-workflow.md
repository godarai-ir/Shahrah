> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Adding Alerts For Workflow

> Learn how to add alerts for workflows to track execution and failures.

## Pre-Requisites

Setup a `Notification Channel Integration`on the Integrations page. Please follow this [document](/docs/adding-alerts-for-workflow) to add an integration.

Currently two types of Notification Channels are supported:

1. Slack (Webhook URL based): For this, please create a [webhook](https://api.slack.com/messaging/webhooks) for your slack and then add it as an integration in **Slack Provider Account** (Slack Webhook Integration)
2. Slack (Bot Token based): For this, please create a [bot token](https://api.slack.com/quickstart) for your slack and then add it as an integration in **Slack Provider Account** (Slack Bot Integration). *It requires chat:write and chat:write:public scope.* You can add slack channels to send to respective slack channel.
3. Email (SMTP credentials Integration): This can be found in **Custom Provider Account.** You can configure the SMTP Credentials of the mail server, from\_email and to\_emails and use it to send notifications.

## How to Configure

### Define alerts in workflow Decorator

* You can define or configure the alerts in the workflow decorator in your workflow file.

<CodeGroup>
  ```python Python lines theme={"dark"}
  from truefoundry.workflow import ...
  from truefoundry.deploy import Resources, WorkflowAlert, Email, SlackWebhook

  ...
  ...
  @workflow(
      alerts=[
          WorkflowAlert(
              notification_target=Email(
                  notification_channel="<Paste your email notification integration fqn here>",
                  to_emails=[
                      "demo@truefoundry.com"
                  ],
              ),
              on_completion=True,
              on_failure=True,
          ),
          WorkflowAlert(
              notification_target=SlackWebhook(
                  notification_channel="<Paste your slack webhook notification integration fqn here>",
              ),
              on_completion=True,
              on_failure=True,
          ),
          WorkflowAlert(
              notification_target=SlackBot(
                  notification_channel="<Paste your slack bot notification integration fqn here>",
                  channels:[
                      "#workflow-notifications"
                  ]
              ),
              on_completion=True,
              on_failure=True,
          ),
      ]
  )
  def my_workflow():
  	...
  ```
</CodeGroup>

<Warning>
  ### ⚠️Make sure that the truefoundry package version is set to latest version in task config.

  For example in pip\_packages filed in python task config, make sure that trufoundry\[workflow] version is 0.6.3 or newer and not older one, or make sure that version which you are mentioning supports alerts.
</Warning>

* Right now the alert notification is supported only for workflow completion(successful completio and workflow failure(failure includes failed executions and aborted as well as timed out workflow executions).
* There are three types of targets, one is email and the others are slack webhook and slack bot, you need to add an email/slack webhook/slack bot integration respectively and select it in notification channel field to use that integration for sending notifications.
* Now when you will deploy the workflow, the alerts will be set.

### Define alerts in deploy.py file

* Another way to define alerts is to define workflow in deploy.py file and set alerts in it.

<CodeGroup>
  ```python deploy.py lines theme={"dark"}
  from truefoundry.deploy import Workflow, WorkflowAlert, SlackWebhook, Email


  wf = Workflow(
      name="test-email-alerts-2",
      workflow_file_path="nested.py",
      alerts=[
          WorkflowAlert(
              notification_target=SlackWebhook(
                  notification_channel="<Paste your slack webhook notification integration fqn here>",
              ),
              on_completion=True,
              on_failure=True,
          ),
          WorkflowAlert(
              notification_target=Email(
                  notification_channel="<Paste your email notification integration fqn here>",
                  to_emails=["demo@truefoundry.com"],
              ),
              on_completion=True,
              on_failure=True,
          ),
          WorkflowAlert(
              notification_target=SlackBot(
                  notification_channel="<Paste your slack bot notification integration fqn here>",
                  channels=["#workflow-notifications"],
              ),
              on_completion=True,
              on_failure=True,
          ),
      ],
  )

  wf.deploy("<workspace-fqn>", wait=False)
  ```
</CodeGroup>

* After defining this, you can just run `python deploy.py` and your workflow will be deployed and alerts will be set.

***
