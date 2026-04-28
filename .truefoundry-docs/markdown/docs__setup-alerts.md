> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Setup Alerts

> Configure Prometheus AlertManager alerts with Email, Slack, or PagerDuty notification channels for monitoring.

Alerts are important to monitor the health of your applications and get notified if something goes wrong. Truefoundry makes it easy to setup alerts for all your applications including Service, Async Service, Job, Helm Deployment, Volume, Notebook and SSH server. You can set up notification channels in Email, Slack or Pagerduty to be notified of the alerts.

Truefoundry uses [Prometheus AlertManager](https://prometheus.io/docs/alerting/latest/alertmanager/) to power the alerts.

### Key Components of an Alert

An alert primarily comprises of two components:

1. **Alert Rule**: This is a rule in terms of a PromQL query that is evaluated periodically to check if its true. If its true for a `configured duration`, the alert is triggered.

<Tip>
  Truefoundry provides the PromQL expressions of the most commonly used alerts which should suffice for most usecases. So you don't need to necessarily learn PromQL to setup alerts.
</Tip>

2. **Severity**: This is the severity of the alert which is used to categorize if the alert needs immediate attention or not. It can be either `warning` or `critical`. This can be used in PagerDuty to route the alert to the proper channel.
3. **Notification Channel**: These are the channels where the alerts will be sent once they are triggered.

## Setting up alerts in Truefoundry Services

To setup alerts, you need to follow the steps below:

### 1. Setup Notification Channels

<Note>
  You need to create an integration with either Slack, Email or PagerDuty before setting up a notification channel. If this is not already done, please refer to Adding [slack](/docs/slack-bot-integration), [email](/docs/email-integration), and [pagerduty](/docs/pagerduty-integration) integrations documentation.
</Note>

Before setting up the alerts, you have to configure the notification channels and add one or more notification channels to send alerts to.

<Frame>
  <img src="https://mintcdn.com/truefoundry/OHzlp6GY5G-JfKle/images/configure-notification-target.png?fit=max&auto=format&n=OHzlp6GY5G-JfKle&q=85&s=3d5e2a23dd90cb86411ef8d960fa8a06" alt="Configure Notification Channels" width="2826" height="1052" data-path="images/configure-notification-target.png" />
</Frame>

You can choose any among Email, SlackBot or Pagerduty as the notification channel to send alerts.

<Frame>
  <img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/email-channel.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=74109a872d681c8fffd21593ebe88e07" alt="Email Channel" width="1570" height="1038" data-path="images/email-channel.png" />
</Frame>

<Note> You can add multiple notification channels to send alerts to. </Note>

### 2. Create Alert

You can choose among the already available alerts or create your own custom alert. In most of the cases, the existing alerts should suffice for your usecase. Here are a few of the alerts already available:

<AccordionGroup>
  <Accordion title="HighCPUThrottleRate" default>
    This alert is triggered when the container CPU throttling rate exceeds 25% in the last 5 minutes, indicating CPU resource exhaustion. This means it might be affecting the latency or throughput of the service. **The solution is usually to increase the CPU request and limit of the service.**
  </Accordion>

  <Accordion title="PodIsCrashLooping">
    This alert is triggered when a pod restarts more than 5 times within a 1 hour period, indicating a crash loop. **You should check the logs and events to understand the root cause of the crash.**
  </Accordion>

  <Accordion title="PodIsNotHealthy">
    This alert is triggered when a pod is not in a running state for more than 15 minutes. This can happen if the pod is in pending, unknown, or failed status. **This can happen if we cannot provision an instance or there is an application error. You can check the logs and events to understand the root cause of the pod not being healthy.**
  </Accordion>

  <Accordion title="RequestSuccessRateIsLow">
    This alert is triggered when the requests are failing and we are getting 5XX responses. This usually happens because of a bug in the application code or some pods not running successfully.

    <Warning>
      This alert only takes into account the requests going through the load-balancer. This means that if you are calling the service directly using `<your-service-name>.svc.cluster.local`, it will not be taken into account. It should be called using the domain url which should provide in the Ports section after clicking on `Expose` button.
    </Warning>
  </Accordion>
</AccordionGroup>

<Frame>
  <img src="https://mintcdn.com/truefoundry/OHzlp6GY5G-JfKle/images/create-alert-rule.png?fit=max&auto=format&n=OHzlp6GY5G-JfKle&q=85&s=edf8b9aa71b18bb27a28a1caed35647b" alt="Create Alert Rule" width="2850" height="1086" data-path="images/create-alert-rule.png" />
</Frame>

If you need something apart from the above, you can create your own alert using the **New Alert Rule** form in the UI. To create and verify the PromQL query, we recommend using Grafana UI to test the query. The key fields to fill up in the form are:

* `Name`: A descriptive name for your alert.

* `Description`: (Optional) Briefly describe what this alert monitors.

* `Prometheus Expression`: Enter the Prometheus query that defines the alert condition. For example:

  ```prometheus lines theme={"dark"}
  sum(rate(http_requests_total{status!="2xx"}[5m])) by (service) > 5
  ```

  This triggers if there are more than 5 non-2xx HTTP responses in 5 minutes for any service.

* `Trigger After (seconds)`: How long the condition must be true before triggering the alert.

* `Severity`: Choose between `Warning` and `Critical`.

* `Notification Enabled`: Enable or disable notifications for this rule.

<Frame>
  <img src="https://mintcdn.com/truefoundry/OHzlp6GY5G-JfKle/images/create-alert-rule-form.png?fit=max&auto=format&n=OHzlp6GY5G-JfKle&q=85&s=7a54642935fb8e42c3f6dd15f2336902" alt="Create Alert Rule Form" width="1638" height="1486" data-path="images/create-alert-rule-form.png" />
</Frame>

### Applying AlertRules via YAML in GitOps

You can apply AlertRules via YAML in GitOps. You can copy the YAML from the Code icon in the Alerts page.

<Columns cols={2}>
  <Card>
    <img src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/alerts-code-button.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=68c15c4835f171cc69e1290b83f783b9" width="3840" height="1936" data-path="images/docs/alerts-code-button.png" />
  </Card>

  <Card>
    <img src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/alerts-yaml-ui.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=4dba9ed3eaa321ee538d38244e0bec33" width="3840" height="1938" data-path="images/docs/alerts-yaml-ui.png" />
  </Card>
</Columns>
