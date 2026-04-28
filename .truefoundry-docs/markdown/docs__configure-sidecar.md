> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Configure Sidecar

> Step-by-step guide for configure sidecar, explaining configuration, best practices, and real-world usage on TrueFoundry.

To run the async service in sidecar mode as mentioned in [Introduction](/docs/introduction-to-async-service#use-sidecar), we will need to enable the sidecar in the deployment form. This is available in the advanced settings in the Deployment form.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/s4Aj2_qGCrSP-zc8/images/7d1e6ced-1cf4a60-Screenshot_2024-04-14_at_12.36.33_PM.png?fit=max&auto=format&n=s4Aj2_qGCrSP-zc8&q=85&s=724115856bc9e17700e0ca1b10c12ede" width="2326" height="1316" data-path="images/7d1e6ced-1cf4a60-Screenshot_2024-04-14_at_12.36.33_PM.png" />
</Frame>

We will have to provide the url in your service to which the sidecar can send the HTTP request. For e.g. if you have a path called `/predict` in your service and you are running it on port 8000, the destination url will be `http://0.0.0.0:8000/predict`

<Info>
  ### Sidecar destination url should be `http://0.0.0.0:<port>/<path>`
</Info>

The port number is the same as what you provided in the Ports section in the deployment form. The reason you have to provide it here again since you might have multiple ports and you want the sidecar to call one of the ports specified.
