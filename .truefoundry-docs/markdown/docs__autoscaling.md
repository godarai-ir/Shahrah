> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Autoscaling

> Step-by-step guide for autoscaling, explaining configuration, best practices, and real-world usage on TrueFoundry.

When traffic or resource usage isn't constant, we want the number of replicas to dynamically adjust based on the Queue Backlog. We need to define the minimum and maximum number of replicas in this case and the autoscaling strategy (based on the Queue Backlog) will decide what should be the number of replicas between the min and max replicas.

## Autoscaling Configuration

Autoscaling configuration involves setting minimum and maximum replica counts as well as defining metrics that trigger autoscaling actions. Here are the available settings for autoscaling:

* **Minimum Replicas**: The minimum number of replicas to keep available.
* **Maximum Replicas**: The maximum number of replicas to keep available.
* **Cooldown Period** (Advanced Settings): The period to wait after the last trigger is reported active before scaling the resource back to 0.
* **Polling Interval** (Advanced Settings): This is the interval to check each trigger on

## Configuring Autoscaling via UI

To configure autoscaling parameters for your service via the UI, follow these steps:

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/UTorLi_7TqoJSmjeCwQNl" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/UTorLi_7TqoJSmjeCwQNl" style={{ border: "none", display: "flex", margin: "auto" }} />

## Async Service Autoscaling Metrics

For your async service, autoscaling metrics play a crucial role in dynamically adjusting resource allocation to meet changing demands while maintaining optimal performance.

* **AWS SQS Average Backlog**: AWS SQS pending queue length averaged over all replicas which autoscaler will try to maintain.\
  This option is only available when the Input Worker is based on AWS SQS.\
  Learn more about [AWS SQS Backlog Metric](/docs/aws-sqs#configuring-autoscaling-for-aws-sqs-queue)
