> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Deploy your first Async Service

> Deploy your first async service. Learn architecture, setup, and scaling basics.

To deploy an async service, we will need to write an HTTP service using any framework like Fastapi, Flask, etc. In this example, we will be using [FastAPI](https://fastapi.tiangolo.com/). We will need to expose a POST endpoint. We will also need to provision a queue to which we will be pushing the messages.

## Pre-requisites

Before you proceed with the guide, please ensure that you have setup a Workspace. To deploy your service, you'll need a workspace. If you don't have one, you can create it using this guide: [Creating a Workspace](/docs/key-concepts#creating-a-workspace)

### HTTP Service

We've already created a FastAPI service for the Iris classification problem, and you can find the code in our [GitHub Repository](https://github.com/truefoundry/getting-started-examples/tree/main/deploy-model-with-fastapi).

Please visit the repository to familiarize yourself with the code you'll be deploying.

#### Project Structure

The project files are organized as follows:

<CodeGroup>
  ```plain Directory Structure theme={"dark"}
  .
  ├── server.py - Contains FastAPI code for inference.
  ├── iris_classifier.joblib - The model file.
  └── requirements.txt - Lists dependencies.
  ```
</CodeGroup>

All these files are located in the same directory.

### Queue

We will be using SQS for this first example. You can use any of the other queue integrations if you want. To provision the AWS SQS Queue, you can refer to the [AWS SQS](/docs/aws-sqs) guide.

## Initiating Deployment via UI

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/cVte4HCNmevUB8mVwYoV_" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/cVte4HCNmevUB8mVwYoV_" style={{ border: "none", display: "flex", margin: "auto" }} />

Use these configs for the deployment form:

* **Repo URL**: [https://github.com/truefoundry/getting-started-examples](https://github.com/truefoundry/getting-started-examples)
* **Path to build context**: ./deploy-model-with-fastapi/
* **Command**: `uvicorn server:app --host 0.0.0.0 --port 8000`
* **Path to requirements**: requirements.txt
* **Port**: 8000
* **Destination URL** (Inside Sidecar): [http://0.0.0.0:8000/predict](http://0.0.0.0:8000/predict) (You can find this option in the `Advanced` section of the deployment form under `Sidecar` section)

<Info>
  ### What we did above:

  In the example above, We firstly specified the details regarding our HTTP Service.

  Given we only had Python code and a requirements.txt. We didn't have a prewritten docker file - so we chose the **Python Code** option - to let TrueFoundry templatize a Dockerfile from the details provided about the application and build the Docker image for us.

  We give these details in the Build context field, where we specify the directory in the GitHub repository where our service code resides (`./deploy-model-with-fastapi/`). We also specify the command that we need to use to run our service (`uvicorn server:app --host 0.0.0.0 --port 8000`).

  Next, we setup the AWS SQS Queue following the [Creating and Utilizing an AWS SQS Queue](/docs/aws-sqs) documentation.

  We specify the port that we want our service to listen on (`8000`).

  Finally, we configure the **Sidecar**. Sidecar is a component that will consume message from a queue and forwards them as POST requests to your HTTP service.
</Info>

## View your deployed service

Once you click **`Submit`**, your deployment will be successful in a few seconds, and your Async Service will be displayed as active (green), indicating that it's up and running.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/FrY4JbiyZud2He3p/images/28688c1d-2fbcff0-Screenshot_2023-12-05_at_10.29.05_AM.png?fit=max&auto=format&n=FrY4JbiyZud2He3p&q=85&s=c2f7b5684971c7cca996b6fcecf3ea1b" width="2940" height="1055" data-path="images/28688c1d-2fbcff0-Screenshot_2023-12-05_at_10.29.05_AM.png" />
</Frame>

Congratulations! You've successfully deployed your first Async service.

## Interacting with the Application

### Async Service Specific Dashboard

<iframe href="https://app.supademo.com/embed/HEwdbPDBkEChLiOYpn-TH" typeofembed="iframe" height="475px" width="100%" provider="app.supademo.com" src="https://app.supademo.com/embed/HEwdbPDBkEChLiOYpn-TH" style={{ border: "none", display: "flex", margin: "auto" }} />

### Sending Requests to your Async Service

The way to trigger the async service will be to push messages to the queue so that the sidecar and pull messages from the queue and call the HTTP service that we have written above. We can verify if things are working by seeing the logs of the service.

To dump messages into the queue, we will have to use the queue library to push messages to the queue. A sample code to push messages to the SQS queue is as follows:

```python lines theme={"dark"}
import json
import uuid
import boto3

def send_request(input_sqs_url: str):
    sqs = boto3.client("sqs")
    request_id = str(uuid.uuid4())

    payload = {
        "sepal_length": 7.0,
        "sepal_width": 3.2,
        "petal_length": 4.7,
        "petal_width": 1.4,
    }

    sqs.send_message(
        QueueUrl=input_sqs_url,
        MessageBody=json.dumps({"request_id": request_id, "body": payload})
    )

if __name__ == "__main__":
    send_request(input_sqs_url="YOUR_INPUT_SQS_URL")
```

Run the above Python script using `python send_async_request.py`
