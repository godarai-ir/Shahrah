> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# AWS SQS

> Step-by-step guide for aws sqs, explaining configuration, best practices, and real-world usage on TrueFoundry.

## Provisioning the queue

Creating an AWS Simple Queue Service (SQS) queue is a straightforward process that can be accomplished using the AWS Management Console or AWS Command Line Interface (CLI). Here's a step-by-step guide for creating an AWS SQS queue through the AWS Management Console:

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/-mGxlBr0P7TKaq5UYONTQ" typeofembed="iframe" height="494px" width="100%" src="https://app.supademo.com/embed/-mGxlBr0P7TKaq5UYONTQ" style={{ border: "none", display: "flex", margin: "auto" }} />

<Warning>
  Note:

  The visibility timeout has a significant impact on the behavior of asynchronous services.

  It determines how long a message remains hidden from consumers after it's fetched from the queue, playing a vital role in preventing multiple consumers from processing the same message concurrently.

  For example, if your worker process typically takes around 5 seconds to complete a task, it's advisable to set the Visibility Timeout to at least twice that duration, which in this case would be 10 seconds.

  If the Visibility Timeout is set too low, there's a risk of multiple consumers attempting to process the same message simultaneously, potentially leading to conflicts and errors in your system. It's essential to strike the right balance to ensure efficient and orderly message processing.
</Warning>

Once you click on **Create queue**, you'll receive a confirmation message indicating the successful creation of the queue.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/FrY4JbiyZud2He3p/images/195801da-23e8a38-Screenshot_2023-10-20_at_17.54.10.png?fit=max&auto=format&n=FrY4JbiyZud2He3p&q=85&s=ba977afb60e420ba4cd06bd30d774849" width="3840" height="1890" data-path="images/195801da-23e8a38-Screenshot_2023-10-20_at_17.54.10.png" />
</Frame>

## Configuring TrueFoundry Async Service with AWS SQS

You will have to specify these configurations for AWS SQS Input Worker:

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/cMwl3Xaf29OHlMI05MT6A" typeofembed="iframe" height="480px" width="100%" src="https://app.supademo.com/embed/cMwl3Xaf29OHlMI05MT6A" style={{ border: "none", display: "flex", margin: "auto" }} />

## Configuring Autoscaling for AWS SQS Queue

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/ffc53dff-9823d30-Screenshot_2023-12-05_at_10.43.06_AM.png?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=af52b084441e18dcb78424ba906f7ff4" width="2940" height="527" data-path="images/ffc53dff-9823d30-Screenshot_2023-12-05_at_10.43.06_AM.png" />
</Frame>

**AWS SQS Average Backlog** is defined as the AWS SQS pending queue length averaged over all replicas that the autoscale is designed to maintain.

The **pending queue length** refers to the number of messages that are currently in the queue but have not yet been processed. These are messages waiting to be consumed and processed by the relevant workers or services.

The **average backlog** is the average or mean value of the pending queue length across multiple replicas. In a distributed and auto-scaling system, there can be multiple instances or replicas of your service, each with its queue. The average backlog provides a way to measure the workload across all replicas.

This **Average Backlog** is a valuable metric for determining how to scale your application efficiently.

<Info>
  Note:

  This metric is only available in case you are using AWS SQS for your input queue OR authentication is configured under autoscaling.
</Info>

## Authentication for AWS SQS

### AWS IAM User

You can configure AWS IAM User with permissions to read and write to SQS Queue and generate Access key pair. You can configure AWS Access Key ID and AWS Access Secret Key in input config of SQS.

<img src="https://mintcdn.com/truefoundry/Cwv1mi1V5pPIB0ae/images/async-service-sqs-auth-access-keys.png?fit=max&auto=format&n=Cwv1mi1V5pPIB0ae&q=85&s=62a70117ac9d09a4babdd4aafa98861a" width="1726" height="1364" data-path="images/async-service-sqs-auth-access-keys.png" />

### AWS IAM Role with Web Identity

You can configure AWS IAM Role along with Web Identity.

<Accordion title="Configuration for IAM Role to be attached to Async service">
  **Policy:**

  ```
  {
      "Version": "2012-10-17",
      "Statement": [
          {
              "Action": [
                  "sqs:*"
              ],
              "Effect": "Allow",
              "Resource": "arn:aws:sqs:us-east-1:1234567890:my-sqs" # AWS ARN for your SQS
          }
      ]
  }
  ```

  **Trust Relationship:**

  ```
  {
      "Version": "2012-10-17",
      "Statement": [
          {
              "Sid": "",
              "Effect": "Allow",
              "Principal": {
                  "AWS": "arn:aws:iam::1234567890:role/keda-operator-platform-role" # IAM Role ARN attached to keda operator pod.
              },
              "Action": "sts:AssumeRole"
          },
          {
              "Effect": "Allow",
              "Principal": {
                  "Federated": "arn:aws:iam::1234567890:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/ABCDEFG1234567890ABCDEFG" # AWS OIDC ARN for EKS
              },
              "Action": "sts:AssumeRoleWithWebIdentity",
              "Condition": {
                  "StringEquals": {
                      "oidc.eks.us-east-1.amazonaws.com/id/ABCDEFG1234567890ABCDEFG:aud": "sts.amazonaws.com", # AWS OIDC Claim
                      "oidc.eks.us-east-1.amazonaws.com/id/ABCDEFG1234567890ABCDEFG:sub": "system:serviceaccount:my-workspace:my-async-service" # AWS OIDC claim, Namespace and serviceaccount for async service
                  }
              }
          }
      ]
  }
  ```
</Accordion>

Add the Role ARN for the above created role as IRSA annotations in serviceaccount used in your async service application. Also, when enabling autoscaling, add the Role ARN in authentication.

<img src="https://mintcdn.com/truefoundry/Cwv1mi1V5pPIB0ae/images/async-service-sqs-auth-iam-based.png?fit=max&auto=format&n=Cwv1mi1V5pPIB0ae&q=85&s=acdaa788dd8afe87faa74e5385e9d790" width="1804" height="984" data-path="images/async-service-sqs-auth-iam-based.png" />

<Note>
  Please ensure platform IAM role is attached to Keda operator pods to be able to pull metrics when autoscaling is enabled.

  <Accordion title="Configuration for IAM Role to be attached to Keda operator">
    **Policy:**

    ```
    {
        "Statement": [
            {
                "Action": "sts:AssumeRole",
                "Effect": "Allow",
                "Resource": "*"
            }
        ],
        "Version": "2012-10-17"
    }
    ```

    **Trust relationship:**

    ```{ theme={"dark"}
        "Version": "2012-10-17",
        "Statement": [
            {
                "Effect": "Allow",
                "Principal": {
                    "Federated": "arn:aws:iam::1234567890:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/ABCDEFG1234567890ABCDEFG"
                },
                "Action": "sts:AssumeRoleWithWebIdentity",
                "Condition": {
                    "StringEquals": {
                        "oidc.eks.us-east-1.amazonaws.com/id/ABCDEFG1234567890ABCDEFG:aud": "sts.amazonaws.com",
                        "oidc.eks.us-east-1.amazonaws.com/id/ABCDEFG1234567890ABCDEFG:sub": "system:serviceaccount:keda:keda-operator" # AWS OIDC claim, Namespace and serviceaccount for keda
                    }
                }
            }
        ]
    }
    ```
  </Accordion>
</Note>

## Parameters for SQS Average Backlog

* **Queue lag threshold**: This is the maximum number of messages each replica should handle. If there are more messages than the threshold, the auto-scaler adds replicas to share the workload.

## Configuring AWS SQS Average Backlog

### Through the User Interface (UI)

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/clpry1j8m0f23peq5ey81cjur" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/clpry1j8m0f23peq5ey81cjur" style={{ border: "none", display: "flex", margin: "auto" }} />

### Via the Python SDK

In your Service deployment code **`deploy.py`**, include the following:

<CodeGroup>
  ```python Python highlight={12-20} lines theme={"dark"}
  from truefoundry.deploy import AsynceService, Build, DockerFileBuild, Port, AsyncServiceAutoscaling,SQSQueueMetricConfig

  service = AsyncService(
      name="my-async-service",
      image=Build(build_spec=DockerFileBuild()),
      ports=[
        Port(
          host="your_host",
          port=8501
        )
      ]
      replicas=AsyncServiceAutoscaling(
          min_replicas=1,
          max_replicas=3,
          metrics=SQSQueueMetricConfig(
              queue_length=30
          ),
          cooldown_period=300,
          polling_interval=30
      )
  )
  service.deploy(workspace_fqn="YOUR_WORKSPACE_FQN")
  ```
</CodeGroup>
