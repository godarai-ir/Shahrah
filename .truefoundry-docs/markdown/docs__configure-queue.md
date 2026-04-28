> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Configure Queue

> Step-by-step guide for configure queue, explaining configuration, best practices, and real-world usage on TrueFoundry.

The Queue configuration is the most important configuration in the Async Service. You will find these settings under the Worker Config section in the form.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/PSBc0bX31_cIC7pm/images/d809d833-548cbb9-Screenshot_2024-04-12_at_12.58.11_AM.png?fit=max&auto=format&n=PSBc0bX31_cIC7pm&q=85&s=c0c44e49c395db05cddd24174995b31f" width="2866" height="1710" data-path="images/d809d833-548cbb9-Screenshot_2024-04-12_at_12.58.11_AM.png" />
</Frame>

There are two sections: **Input Config** and **Output Config**

### Input Config

This is the input queue configuration and has to be provided. This configuration comprises of primarily two parts:

1. Url of the Queueing system.
2. Authentication config to talk to the queueing system.

The exact settings differs based on the queue. You can find more details about it in the Queue Integrations for AWS SQS, Kafka, NATS and Google AQMP.

### OutputConfig

The output queue configuration is optional and only needed if you want to sidecar to write the output back to a queue. In many cases, your HTTP service might be writing the output directly to S3 or updating some database - in which case you can leave the output config empty. If you want to write the output back to a queue so that it can be consumed by some other system, then we will need to provision a output queue.

TrueFoundry supports the following queue integrations for Output Queue: SQS, Core Nats, Nats, Kafka and AMQP. If you want to support for some other queue, feel free to reach out to us or contribute the queue adapter at [https://github.com/truefoundry/async\_service](https://github.com/truefoundry/async_service). The settings for the Queue are very similar to the input configs and can be found in the docs of the individual Queue integrations.
