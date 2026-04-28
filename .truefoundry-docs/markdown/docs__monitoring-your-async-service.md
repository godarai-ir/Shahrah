> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Monitor Your Async Service

> Monitor async service health with consumer lag, processing rate, and latency metrics plus real-time logs.

Async service involves your main HTTP service code which you are writing and a sidecar component which picks up items from the queue and calls yoru HTTP service with the payload. The sidecar code has been written by TrueFoundry and emits useful metrics which can help you understand the overall state of the system.

## Metrics

To see the metrics for a particular Async Service, you need to select that Async Service from the deployments section, then click on the **`Metrics`** button in the top right corner of the dashboard.

<iframe href="https://app.supademo.com/embed/ffiG8xBWzNUBPH9qsv8cT" typeofembed="iframe" height="480px" width="100%" provider="app.supademo.com" src="https://app.supademo.com/embed/ffiG8xBWzNUBPH9qsv8cT" style={{ border: "none", display: "flex", margin: "auto" }} />

The key metrics to monitor and pay attention to are:

* **Consumer Lag**: If you had to choose one metric to look at to understand the state of the service, Consumer Lag would be that metric. Consumer Lag denotes how many items are present in the queue that have not yet been processed by your service. Ideally the consumer lag should be less than the number of messages that one pod can process at any given point of time. If the consumer lag rises, it means that your service is processing messages at a slower rate than the rate at which items are enqueue into the queue. If its rising, you should increase the number of replicas of your service so that your processing rate can be faster.

* **Messages Processing Rate**: This number shows how many messages per second if your service processing. This is the combined processing rate of all the pods of your service. This number should be equal to the rate at which messages are enqueud in the queue, else the consumer lag will start growing.

* **Processing Time (ms)**: Processing time is the time taken by your service to process the message. This will be the same time which your service would have taken if it would have been directly invoked via HTTP. This is important to understand the latency that the clients will incur for each request. Please note that the total latency that the client will see is more than the processing time.

<Info>
  Total Latency = Time to put message in queue + time spent in queue + time to move message from queue to HTTP service by sidecar + processing time + response publish time
</Info>

* **Response Publish Time(ms)**: This is the time taken by sidecar to publish the response received from the HTTP service to the output queue. This number should usually be very small unless the sidecar is having issues talking to the output queue.

* **Message Processing Failure % (%):**\
  Tracks the percentage of messages for which your service returns non-2XX HTTP codes. This helps keep track of the cases when your service is either erroring out or you have intentionally provided error return codes.

* **Consumer Pull Failure Rate:** This denotes errors from sidecar while pulling messages from queue and should almost never happen. The most probably reason for this number being greater than 0 will be issue connecting to the input queue which might be a network or authentication error. If you are ever encountering this, please see the logs of the service from async\_processor and that should indicate the error.

## Logs

### Service Level Logs

You can see your own HTTP service logs by clicking on the logs button.

<iframe href="https://app.supademo.com/embed/7fzAP0zhFBr_ePtCmVzUO" typeofembed="iframe" height="480px" width="100%" provider="app.supademo.com" src="https://app.supademo.com/embed/7fzAP0zhFBr_ePtCmVzUO" style={{ border: "none", display: "flex", margin: "auto" }} />

You will see your own service logs inteleaved with the sidecar logs. All the log lines starting with `async_processor:INFO` belong to the sidecar. In steady state, you will hardly see any logs from the sidecar and most logs will belong to your service. The sidecar outputs the logs when its facing any sort of errors which might help you debug your service.

### Pod Level Logs

You can also filter out the sidecar logs in the pod logs by choosing the sidecar container.

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/kNlBko2mosGTLWyHHzzCd" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/kNlBko2mosGTLWyHHzzCd" style={{ border: "none", display: "flex", margin: "auto" }} />
