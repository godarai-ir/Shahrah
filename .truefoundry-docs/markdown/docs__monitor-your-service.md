> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# View Logs, Metrics And Events

> Monitor your deployed services with pod-level logs, resource metrics, events, and Grafana dashboard integration.

TrueFoundry provides you with Logs, Metrics and Events to monitor your deployments to identify and debug issues.

## Browse Logs using UI

Logs are records of events that occur in your deployment, such as requests to your services, errors that occur, and system messages. TrueFoundry provides you with logs at 2 levels:

1. **Deployment Level Logs**: This will show the combined logs from all the pods in the deployment.
2. **Pod Level Logs:** This allows you to view logs for an individual pod in the deployment.

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/hhyLU7IWe5MI49Bri1mON" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/hhyLU7IWe5MI49Bri1mON" style={{ border: "none", display: "flex", margin: "auto" }} />

## Metrics Dashboard

Metrics dashboard provides a visual representation of the metrics collected from your application like CPU Usage, GPU Usage, Network Usage etc. This allows you quick and easy way to identify fluctuations and potential performance bottlenecks. TrueFoundry provides you with metrics dashboards at two levels:

1. **Deployment Level Metrics**: This will show the combined metrics from all the pods in the deployment.
2. **Pod Level Metrics:** This allows you to view metrics for an individual pod in the deployment.

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/ilhLQKst11EIv3HKuBf1Y" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/ilhLQKst11EIv3HKuBf1Y" style={{ border: "none", display: "flex", margin: "auto" }} />

## View Events List

Events are the occurrences that happen in your TrueFoundry deployment. They can be triggered by a variety of things, a few examples of which are:

* Deploying a new service version
* Starting or stopping a pod
* A pod crashing

Events can be used to track the progress of your deployment and to troubleshoot any problems that may occur. These are the standard Kubernetes events - you probably don't need to look at them often unless you are debugging your pod not starting.

TrueFoundry provides you with events at 2 levels:

1. **Deployment Level Events**: This will show the combined events from all the pods in the deployment.
2. **Pod Level Events:** This allows you to view events for an individual pod in the deployment.

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/_vAkIV-0QQd6s58h-lvx0" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/_vAkIV-0QQd6s58h-lvx0" style={{ border: "none", display: "flex", margin: "auto" }} />

## Set up Monitoring through Grafana

[Grafana](https://grafana.com) is a popular open-source dashboarding tool that can plot the different metrics in Prometheus. Since TrueFoundry services metrics are present in Prometheus, they can be viewed in Grafana also. Exporting the metrics dashboard to Grafana will enable the following scenarios:

* **Customizable dashboards**: Grafana allows you to create custom dashboards that display the metrics that are most important to you.
* **Alerting**: You can set up alerts in Grafana to notify you when certain metrics reach specific thresholds.
* **Integration with other tools**: Grafana integrates with a variety of other tools, such as Prometheus and Elasticsearch, so you can consolidate your monitoring data into a single platform.

To get the Grafana dashboard, you can use `Export Metrics Dashboard to Grafana` feature as described below.

### Prerequisites

Before you can start exporting your Service's Metrics Dashboard to Grafana, you need to complete the following prerequisites:

* **Install Grafana on your Cluster**: Follow the instructions provided below to install Grafana on your cluster.

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/mAx2RjHhs6NL4ZSW4y9uH" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/mAx2RjHhs6NL4ZSW4y9uH" style={{ border: "none", display: "flex", margin: "auto" }} />

* **Obtain Grafana Credentials**: Once Grafana is installed, follow the instructions below and obtain the necessary credentials. These credentials will be used later in the process of accessing your Grafana Dashboard

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/aoXKtutVsKgQsNtJO5vWb" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/aoXKtutVsKgQsNtJO5vWb" style={{ border: "none", display: "flex", margin: "auto" }} />

### Exporting Metrics Dashboard

To export the Metrics Dashboard to Grafana you can follow the instruction given below:

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/khIV7lKPDcPMPjXk6N8_N" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/khIV7lKPDcPMPjXk6N8_N" style={{ border: "none", display: "flex", margin: "auto" }} />

## Integrate with your own monitoring and logging solution

You can integrate any of your existing monitoring and logging solutions like Datadog, Newrelic, Prometheus, Cloudwatch, Splunk, etc for the applications deployed on TrueFoundry. Since TrueFoundry applications run on a Kubernetes cluster deployed on your account, you can install the respective agents of the monitoring solutions on the cluster and get all the logs and metrics show up on your monitoring dashboard.

***
