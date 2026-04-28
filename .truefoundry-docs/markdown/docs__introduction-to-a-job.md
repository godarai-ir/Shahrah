> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Introduction

> Understand what a job is and how it runs on TrueFoundry.

TrueFoundry jobs enable you to run task-oriented workloads which are meant to run for a certain duration to complete a task, then terminate and release the resources.

Here are some scenarios where Jobs are particularly well-suited:

* **Model Training**: Train machine learning models on large datasets, where the resource gets freed up once the training is complete.
* **Maintenance and Cleanup**: Schedule routine maintenance tasks, such as data backups, model retraining, report generation etc.
* **Batch Inference**: Perform large-scale batch inference tasks, such as processing large volumes of data using trained models, leveraging Job's ability to handle parallel workloads efficiently.
* **Spark Jobs**: Run Apache Spark jobs on your kubernetes cluster, with dynamic scaling.

TrueFoundry makes it easy to configure various aspects of your job deployment.

<Columns cols={3}>
  <Card title="Dockerize Code" href="/docs/dockerize-code">
    Deploy from Github, local machine or a prebuilt image.
  </Card>

  <Card title="Customize Resources" href="/docs/resources-cpu-memory-storage">
    Set CPU, GPU, memory resources and spot/on-demand instances.
  </Card>

  <Card title="Environment Variables And Secrets" href="/docs/environment-variables-and-secrets-jobs">
    Set environment variables and secrets for your job.
  </Card>

  <Card title="Schedule Job" href="/docs/deploy-job-from-a-public-github-repository#schedule-trigger">
    Schedule Job to run at a specific time.
  </Card>

  <Card title="Trigger your job" href="/docs/deploy-job-from-a-public-github-repository#manual-trigger">
    Trigger your job manually
  </Card>

  <Card title="Parameterize Job" href="/docs/deploy-a-job-with-additional-parameters">
    Parameterize your job to enable ease of changing argument values.
  </Card>

  <Card title="Retries and Timeout" href="/docs/retries-and-timeout">
    Set retries and timeout for your job in case the job gets stuck or fails for some reason.
  </Card>

  <Card title="Concurrency" href="/docs/concurrency-limits">
    Set concurrency limit for your job to specify how many instances of a Job can run at once.
  </Card>

  <Card title="Access Cloud Services" href="/docs/access-data-from-s3-or-other-clouds-services">
    Access S3 / GCS /Azure Container / other cloud managed services.
  </Card>

  <Card title="Mounting Volumes" href="/docs/mounting-volumes-job">
    Mount volumes to cache data
  </Card>

  <Card title="Deploy Programatically" href="/docs/deploy-job-programatically">
    Deploy using Python and CLI
  </Card>

  <Card title="Setup CI/CD" href="/docs/setting-up-cicd-for-your-service">
    Setup with your favorite CI/CD tool.
  </Card>

  <Card title="View Metrics" href="/docs/monitor-your-service">
    View the most important metrics for your job run.
  </Card>

  <Card title="View Logs" href="/docs/monitor-your-service">
    View the logs on a job run or per pod.
  </Card>

  <Card title="Set Up Alerts" href="/docs/set-up-alerts">
    Setup alerts and nottifications for your job
  </Card>

  <Card title="Clone, Update, and Rollback" href="/docs/update-rollback-promote-your-service">
    Clone, update version, rollback to previous version and promote to production
  </Card>
</Columns>

***
