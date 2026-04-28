> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Configure Ports

> Step-by-step guide for configure ports, explaining configuration, best practices, and real-world usage on TrueFoundry.

If you are using Async Service via the library route, you need not have any ports exposed in your service.

However, if you are using HTTP service along with a sidecar, you need to provide a port on your HTTP service on which the sidecar will send the request. You need not expose this port externally since the sidecar calls the HTTP service calls it via the localhost url.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/ed5ac9b0-584504e-Screenshot_2024-04-13_at_4.01.00_PM.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=3e51d2f792795e27685fe034dbfc3ae0" width="2306" height="1160" data-path="images/ed5ac9b0-584504e-Screenshot_2024-04-13_at_4.01.00_PM.png" />
</Frame>
