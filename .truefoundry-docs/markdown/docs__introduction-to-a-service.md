> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Introduction

> Learn the basics of services, endpoints, and deployments.

A TrueFoundry Service represents a continuously running application that typically provides a set of APIs for interaction. Services can be dynamically scaled based on incoming traffic or resource demands.

Services are perfect for scenarios where real-time responses are essential, such as:

* Hosting Real-time Model Inference (e.g., Flask, FastAPI)
* Fueling Dynamic Website Backends
* Creating Model Demos (e.g., Streamlit, Gradio)

TrueFoundry makes it easy to configure various aspects of your service deployment.

<Columns cols={3}>
  <Card title="Dockerize Code" href="/docs/dockerize-code">
    Deploy from Github, local machine or a prebuilt image.
  </Card>

  <Card title="Customize Resources" href="/docs/resources-cpu-memory-storage">
    Set CPU, GPU, memory resources and spot/on-demand instances.
  </Card>

  <Card title="Environment Variables And Secrets" href="/docs/environment-variables-and-secrets">
    Set environment variables and secrets for your service.
  </Card>

  <Card title="Pause/Resume" href="/docs/pause-resume-service">
    Pause/Resume to save cost
  </Card>

  <Card title="Domain Mapping to Endpoints" href="/docs/define-ports-and-domains">
    Expose ports and map DNS to endpoints.
  </Card>

  <Card title="Endpoint Authentication" href="/docs/endpoint-authentication">
    Add authentication to your endpoints.
  </Card>

  <Card title="Autoscaling" href="/docs/autoscaling-overview">
    Autoscale on CPU, traffic or time based policies.
  </Card>

  <Card title="Rollout Strategy" href="/docs/rollout-strategy">
    Configure Canary, Blue/Green rollout strategies.
  </Card>

  <Card title="Access Cloud Services" href="/docs/access-data-from-s3-or-other-clouds-services">
    Access S3 / GCS /Azure Container / other cloud managed services.
  </Card>

  <Card title="Health Probes" href="/docs/liveness-readiness-probe">
    Setup probes to check the health of your service.
  </Card>

  <Card title="Mounting Volumes" href="/docs/mounting-volumes-service">
    Mount volumes to cache data
  </Card>

  <Card title="Deploy Programatically" href="/docs/deploy-service-programatically">
    Deploy using Python and CLI
  </Card>

  <Card title="Setup CI/CD" href="/docs/setting-up-cicd-for-your-service">
    Setup with your favorite CI/CD tool.
  </Card>

  <Card title="Clone, Update, and Rollback" href="/docs/update-rollback-promote-your-service">
    Clone, update version, rollback to previous version and promote to production
  </Card>

  <Card title="View Metrics" href="/docs/monitor-your-service">
    View the most important metrics for your service.
  </Card>

  <Card title="View Logs" href="/docs/monitor-your-service">
    View the logs on a service or per pod.
  </Card>

  <Card title="Set Up Alerts" href="/docs/set-up-alerts">
    Setup alerts and nottifications for your service
  </Card>
</Columns>
