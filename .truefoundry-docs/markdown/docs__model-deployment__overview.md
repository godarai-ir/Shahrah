> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Overview

> Overview of model deployment options and patterns.

There are different ways to deploy the models as an API - depending on the framework / type of model.

<Frame caption="Model Deployment Options">
  <img src="https://mintcdn.com/truefoundry/5CkapnZ7CyjQJ4bx/images/docs/model-deployment-options.png?fit=max&auto=format&n=5CkapnZ7CyjQJ4bx&q=85&s=4078981b3d13d10dbe4d6624aaf74dd8" width="3128" height="2032" data-path="images/docs/model-deployment-options.png" />
</Frame>

TrueFoundry doesn't provide any client side framework for model deployment. We believe there are a lot of great open-source frameworks to build inference services and we don't want to build yet another one.
This also helps avoid any vendor lock-in with TrueFoundry since you don't need to change your code to deploy models on TrueFoundry or migrate to another platform.

TrueFoundry supports deploying models from different frameworks. You can deploy models from HuggingFace, or any custom models that you have logged in the TrueFoundry model registry or your own custom inference code in
any of the frameworks. All model API deployments in TrueFoundry are abstractions on top of the [Service Deployment](/docs/introduction-to-a-service.mdx) feature.
That's why its highly recommended to get familiar with service deployment first.

## Bring your own inference code and model in any framework

TrueFoundry can deploy model inference code in any framework that you are using. Here are a few examples of deploying models in most commonly used frameworks and model servers:

<CardGroup cols={3}>
  <Card title="HuggingFace" href="/docs/model-deployment/deploy-model-from-huggingface" arrow>
    Deploy Transformers / Diffusers models with vLLM, SGLang, Nvidia Triton, etc.
  </Card>

  <Card title="Scikit Learn & XGBoost" href="/docs/model-deployment/deploy-sklearn-xgboost" arrow>
    Deploy Scikit Learn and XGBoost models with FastAPI or Nvidia PyTriton.
  </Card>

  <Card title="FastAPI" href="/docs/model-deployment/deploy-model-fastapi" arrow>
    Most flexible option that can wrap any inference code.
  </Card>

  <Card title="LitServe" href="/docs/model-deployment/deploy-litserve" arrow>
    Wrap any model with LitServe with optional features like dynamic batching and advanced features.
  </Card>

  <Card title="AWS Multi Model Server" href="/docs/model-deployment/deploy-model-mms" arrow>
    Deploy models with AWS Multi Model Server.
  </Card>

  <Card title="TorchServe" href="/docs/model-deployment/deploy-torchserve" arrow>
    Deploy Pytorch models with TorchServe.
  </Card>

  <Card title="TensorFlow Serve" href="/docs/model-deployment/deploy-tfserve" arrow>
    Deploy TensorFlow models with TensorFlow Serve.
  </Card>

  <Card title="Mlflow Serve" href="/docs/model-deployment/deploy-model-mlflow" arrow>
    Deploy Mlflow models with Mlflow Serve.
  </Card>
</CardGroup>
