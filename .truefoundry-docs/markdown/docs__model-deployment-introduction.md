> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Introduction

> Introduction to the model deployment lifecycle from training and logging to serving models in production.

TrueFoundry serves the entire lifecycle of a datascientist starting from running a Jupyter Notebook, training the models, logging the models to a model registry and then deploying the model in batch or realtime mode.

### The ML Lifecycle

<Steps>
  <Step title="Start a Jupyter Notebook or Connect Local VSCode/Cursor to remote compute">
    To start a Jupyter Notebook, follow the docs here: [Launch Jupyter Notebook](/docs/launch-notebooks).

    To start a remote SSH server and develop on VSCode by connecting to remote compute, [launch a SSH server](/docs/launch-an-ssh-server)
  </Step>

  <Step title="Run Your Training Job">
    Notebooks are great for experimentation and in some cases, can suffice for running small training jobs - however, in most cases, you will want to deploy the training job
    to a remote server where you can run it manually or on a schedule. Use the [Jobs](/docs/introduction-to-a-job) feature in TrueFoundry to create training job.
  </Step>

  <Step title="Log Your Trained Model and Metrics">
    You can log your models in the TrueFoundry model registry using the Python SDK or UI. You will need to first create a [repository](/docs/introduction-to-ml-repo) to store the models in. You can log your models
    using this [guide](/docs/log-models).
  </Step>

  <Step title="Deploy your model">
    You can deploy the models either in realtime or batch mode. To deploy it in batch mode, use the [Jobs](/docs/introduction-to-a-job) feature.
    To deploy the models in realtime mode, use the [Services](/docs/introduction-to-a-service.mdx) feature in TrueFoundry.
  </Step>
</Steps>
