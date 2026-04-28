> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Liveness/Readiness Probe

> Configure liveness and readiness probes for services.

TrueFoundry helps you deploy on Kubernetes and one of the key advantages here is that Kubernetes will automatically restart your process if something goes wrong. It can also stop routing requests to a replica of a service which is not yet ready or healthy to serve the traffic. The way this works is by telling Kubernetes to keep checking at some regular intervals some API endpoints on your service to detect if the service is ready and healthy. There are two types of such health checks:

* **Liveness probe**: Checks whether the service's replica is currently healthy. If the service is not healthy, the replica will be terminated and another one will be started.
* **Readiness probe**: Checks whether the service's replica is ready to receive traffic. Until the readiness probe succeeds, no incoming traffic will be routed to this replica.

The main difference between liveness and readiness probes is that liveness probes check if the service is running, while readiness probes check if the service is ready to serve requests. Readiness probes are also used to prevent traffic from being routed to replicas that are still starting up or that are in the process of being restarted.

For example, a machine learning service may take some time to load its model before it is ready to serve requests. During this time, the service would still be considered to be live (i.e., the liveness probe would return a success response), but it would not be ready to serve requests (i.e., the readiness probe would not return a success response). This would prevent traffic from being routed to that replica until the model has been loaded and initialized.

## How to add Liveness and Readiness Probes to your deployed services

To add liveness and readiness probes to your deployed service, you will need to:

* Create two endpoints in your service, one for the liveness probe and one for the readiness probe. The endpoints should return a successful response if the service is healthy and a failed response if the service is unhealthy.
* Configure your deployment to use the health check endpoints. You can do this via the user interface (UI) or via the Python SDK.

### Example FastAPI health check endpoints

The following example shows two simple FastAPI health check endpoints `/livez` and `/readyz` that will be used by the Health Checks:

<CodeGroup>
  ```python app.py lines theme={"dark"}
  from contextlib import asynccontextmanager

  from fastapi import FastAPI

  app = FastAPI(root_path=os.getenv("TFY_SERVICE_ROOT_PATH"))

  # Flag to indicate whether the model is loaded and ready
  model_loaded = False

  @app.get("/livez")
  def liveness():
    # Check if the service is running
    return True


  @app.get("/readyz")
  def readyness():
    # Check if the service is ready to receive traffic
    # For example, check if the model is loaded and if the initialization task is complete.
    # Once loaded, set the model_loaded flag to True
    return model_loaded


  @app.get("/")
  async def root():
    return {"message": "Hello World"}
  ```
</CodeGroup>

### Configuring Health Check for your Service

Here are the parameters for liveness and readiness probes:

| Parameter               | Description                                                                                                       |
| ----------------------- | ----------------------------------------------------------------------------------------------------------------- |
| path                    | The path to the health check endpoint.                                                                            |
| port                    | The port on which the health check endpoint is listening.                                                         |
| initial\_delay\_seconds | Number of seconds to wait after the container has started before checking the endpoint.                           |
| period\_seconds         | How often to check the endpoint.                                                                                  |
| timeout\_seconds        | Number of seconds to wait for a response from the endpoint before considering it to be down.                      |
| success\_threshold      | Number of times in a row the endpoint must respond successfully before the container is considered to be healthy. |
| failure\_threshold      | Number of times in a row the endpoint can fail to respond before the container is considered to be down.          |

For example, let's say we have the following configuration:

| Parameter               | Value   |
| ----------------------- | ------- |
| path                    | /health |
| port                    | 8080    |
| initial\_delay\_seconds | 10      |
| period\_seconds         | 5       |
| timeout\_seconds        | 2       |
| success\_threshold      | 3       |
| failure\_threshold      | 2       |

This configuration will cause the probe to check the /health endpoint on port 8080 every 5 seconds. The probe will wait for 2 seconds for a response before failing the probe. The probe will require 3 consecutive successful probes before marking the container as healthy, and it will allow 2 consecutive failed probes before marking the container as unhealthy.

#### Setting Liveness / Readiness Probes via the UI

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/TPrNSTgeJZYvKRrgDGJIa" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/TPrNSTgeJZYvKRrgDGJIa" style={{ border: "none", display: "flex", margin: "auto" }} />

## Viewing health check logs

You can go to the **Pods** tab on the dashboard, and then click on the logs.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/s4Aj2_qGCrSP-zc8/images/99379a4c-665836d-Screenshot_2022-12-26_at_4.22.36_AM.png?fit=max&auto=format&n=s4Aj2_qGCrSP-zc8&q=85&s=7734d21c6d19a89a631365b62e704660" width="2880" height="750" data-path="images/99379a4c-665836d-Screenshot_2022-12-26_at_4.22.36_AM.png" />
</Frame>

You will be able to see the logs of the replica, including the logs of the health check probes.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/DdP_2rhue4AQQlob/images/430c487d-8a28990-Screenshot_2022-12-26_at_4.23.58_AM.png?fit=max&auto=format&n=DdP_2rhue4AQQlob&q=85&s=a8a20fcdb4a7cd865fa7828a307d7534" width="2880" height="682" data-path="images/430c487d-8a28990-Screenshot_2022-12-26_at_4.23.58_AM.png" />
</Frame>

The `GET /livez` and `GET /readyz` giving `200 OK` messages indicate that the health checks are passing and that the service is healthy.

***
