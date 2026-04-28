> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Introduction to Async Service

> Introduction to async services and event-driven execution.

Async Service in TrueFoundry allows us to put input requests in a queue, process them asynchronously and then put the outputs in another queue. You can use an async service if one or more of the following scenarios applies to your use case:

1. **Large payload size:** If the input to your service / model is really large, it might be useful to put the payload in S3 or some blob storage and then trigger the async service to process the payload. The async service can then download the S3 object and do the processing.
2. **Large processing time:** If the time takes to respond to a request is high(from seconds to minutes), it almost becomes essential to use an async service since HTTP requests in a normal service can start timing out.
3. **Scale to 0:** You can scale an async service to 0 if there are no items in the queue.
4. **High reliability in case of traffic surges:** If you are using a simple HTTP service and traffic suddenly rises, your service will start throwing 5XX errors in case autoscaling is slow. Async service offers a higher reliability since the messages will be stored in the queue and processed when the service scales up.

To build an async service, we need to provision an input queue where we will be writing the messages. TrueFoundry supports the following queues for input queues. Each of the queue points to a page on how you can provision the queue.

1. **AWS SQS**
2. **Nats queue**
3. **Kafka**
4. **Google AMQP**

You can consume messages from the queue either via our async library or via a sidecar.

## Use async library

Since an async service consumes messages from a queue, we will need the logic to pull and push items to the queue. While you can write this logic in your own code, it can be cumbersome and you will have to change your code if you ever decide to switch the queue system underneath. To overcome this, TrueFoundry provides a [simple open-source Python library](https://github.com/truefoundry/async_service) that integrates with different queue frameworks and you just need to implement the processing handler. You can follow the README in the Github repository to integrate it into your codebase.

### Architecture:

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/FrY4JbiyZud2He3p/images/13753e37-11698bc-image.png?fit=max&auto=format&n=FrY4JbiyZud2He3p&q=85&s=44a7e2e065559201c34122c47d6d84c7" width="2143" height="784" data-path="images/13753e37-11698bc-image.png" />
</Frame>

<Info>
  ### We recommend using the Python Library only if your processing code is in Python and the time to process is greater than 1 minute.

  For all other use cases, the approach provided below using a sidecar is recommended.
</Info>

## Use sidecar

In this case, the python library mentioned above is packaged into another docker image and runs alongside your main code in a separate container as a sidecar. You will need to write your processing code as HTTP service.

### Architecture

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/4MAaF__cLD4iud16/images/61f96f62-02fe08f-async-architecture.png?fit=max&auto=format&n=4MAaF__cLD4iud16&q=85&s=1018f68d3a1d66c0f20c1fd46837f5fb" width="2655" height="1312" data-path="images/61f96f62-02fe08f-async-architecture.png" />
</Frame>

The tfy-async-sidecar component consumes items from the queue and calls your HTTP service with the payload as a POST request. The tfy-async sidecar code is open source and available [here](https://github.com/truefoundry/async_service). It has adapters to consume the messages from the different types of queue as mentioned above. If you need support for additional queues please feel free to reach out, or contribute the queue adapter on this [repository](https://github.com/truefoundry/async_service).

Once the sidecar pulls the message from the queue, it needs to deliver it to the HTTP service which is written by the user. There are no constraints on the HTTP service and you can write it using your own preferred framework like FastAPI, Flask, ExpressJS, etc. The service needs to expose an endpoint where it will accept the messages in the queue as the body in the POST request. You can provide that endpoint as an input in the async service deployment spec.

After the HTTP service receives the message, it processes it and returns the response to the sidecar. If we configure an output queue, the sidecar can then go and write the data back to the output queue. You can also handle the writing in your service code and not add the output queue configuration if you want.

### Acking Logic with the Input queue

The sidecar acks the message to the input queue once it has received the response from the HTTP service and written the response to the output queue. If there is a failure at any of the intermediate steps, the queue will redeliver the message to one of the replicas again for processing. This ensures that there is a high level of reliability for the input messages to be processed.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/_wMKf6ALFfIRnUSQ/images/d0562395-9c5ce9d-Untitled-2023-10-03-0731.excalidraw-2.svg?fit=max&auto=format&n=_wMKf6ALFfIRnUSQ&q=85&s=be967c7928dde303afa1786c52f3507d" width="2338" height="2596" data-path="images/d0562395-9c5ce9d-Untitled-2023-10-03-0731.excalidraw-2.svg" />
</Frame>

## Getting Started

If you're new to the TrueFoundry Async Service, follow the comprehensive guide given below on how to deploy your service as an asynchronous service. Please make sure you have [deployed a simple service](/docs/deploy-first-service#deploy-from-github) on TrueFoundry before embarking on this tutorial. You will need to have the following components ready before deploying an async service.

1. HTTP service
2. Provisioned Queue

You can follow [Deploy your first Async Service](/docs/deploy-your-first-asyncservice) guide to understand the basics of deploying Async Service.

You can configure the settings further using the guides below:

1. Dockerize the code to be deployed.
2. Define the resources requirements for your service - [Define Resources](/docs/resources-cpu-memory-storage) - While the documentation is specific for services, it's the same as what is required for Async services.
3. Define Queue Configuration
4. \[Optional] Defining environment variables and secrets to be injected into the code - [Environment Variables and Secrets](/docs/environment-variables-and-secrets-jobs).
5. [Update, Rollback, Promote your Async Service](/docs/update-rollback-promote-your-service): While the documentation is specifically for Services, the Update, Rollback, and Promote process follows a similar flow for Jobs.
6. [Setting up CI/CD for your Async Service](/docs/setting-up-cicd-for-your-service): While the documentation is specifically for Services, the CI/CD setup process follows a similar flow for Jobs.
