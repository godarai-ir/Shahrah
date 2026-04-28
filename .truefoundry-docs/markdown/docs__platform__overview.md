> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Truefoundry Overview

> Truefoundry enables developer productivity and governance by making AI deployments easier and providing an AI gateway to access models, MCP servers and agents.

Truefoundry is a a cloud-agnostic platform for building, deploying and monitoring AI applications while enabling complete governance and security within an organization. It provides the following two key modules: **AI Engineering** and **AI Gateway**.

<Tip>
  All the components in Truefoundry are modular and you can decide to use only the components you need. For e.g., if you don't need the AI engineering module, you can just use the AI gateway module.
</Tip>

## AI Engineering

AI engineering module primarily enables datascientists to deploy their models, agents and workflows on your own infrastructure while providing a single place to manage all the AI assets. It abstracts out the underlying infrastructure to enable rapid experimentation and deployment, while making sure it adheres to the guardrails and principles set by the organization.

<Note>
  Truefoundry doesn't provide compute. You bring your own cloud account or on-prem hardware. Truefoundry will connect with it and enable you to deploy your models, agents and workflows.
  All the models and artifacts are also stored on your own storage.
</Note>

<CardGroup cols={3}>
  <Card title="Jupyter Notebooks / Remote SSH" href="/docs/launch-notebooks">
    Start Jupyter Notebooks or connect your existing IDE to remote compute on any cloud/on-prem hardware including GPUs.
  </Card>

  <Card title="Train Models / Batch Inference" href="/docs/introduction-to-a-job">
    Use the Jobs feature in Truefoundry to run training or batch inference jobs either manually or on a schedule.
  </Card>

  <Card title="Model Registry" href="/docs/model-registry">
    Store and Version your models and artifacts.
  </Card>

  <Card title="Model Inference" href="/docs/introduction-to-workflow">
    Deploy your models in any framework (Transformers, PyTorch, TensorFlow, SkLearn, XGBoost, etc) as realtime APIs.
  </Card>

  <Card title="Workflows" href="/docs/introduction-to-workflow">
    Deploy and monitor complex ML pipelines.
  </Card>

  <Card title="Service Deployment" href="/docs/introduction-to-a-service">
    Deploy any service (REST, gRPC, etc) or your Streamlit, Gradio, Flask, FastAPI applications.
  </Card>

  <Card title="LLM Deployment" href="/docs/deploying-an-llm-model-from-the-model-catalogue">
    Deploy LLMs from HuggingFace/own model registry using vLLM / Sglang /TRTLLM with low latency, high througput and faster autoscaling.
  </Card>

  <Card title="LLM Finetuning" href="/docs/finetuning-a-model-from-the-model-catalogue">
    Finetune LLM models on your own data from the model catalogue or any HuggingFace LLM model.
  </Card>

  <Card title="Async Inference" href="/docs/introduction-to-async-service">
    Deploy async inference services backed by queue of your choice to handle inference with higher latency.
  </Card>
</CardGroup>

## AI Gateway

The AI gateway provides a single interface to access all the LLMs and AI models, MCP servers and agents within an organization. It comes with access control, key management, governance and monitoring inbuilt which enables developers to focus on building great applications without worrying about the underlying models, keys, observability and platform teams to impose rate and budget limits, access control, audit and security guardrails.

<CardGroup cols={3}>
  <Card title="LLM Gateway" href="/docs/ai-gateway/intro-to-llm-gateway">
    Call 1000+ LLM models using a single API.
  </Card>

  <Card title="MCP Registry" href="/docs/ai-gateway/mcp/mcp-server-getting-started">
    Deploy MCP servers from the model catalogue.
  </Card>

  <Card title="MCP Gateway" href="/docs/ai-gateway/mcp/mcp-gateway-auth-security">
    Deploy MCP servers from the model catalogue.
  </Card>

  <Card title="Prompt Management" href="/docs/ai-gateway/prompt-management">
    Create, store and version prompts and use them via the Gateway.
  </Card>

  <Card title="Tracing and Observability" href="/docs/ai-gateway/analytics">
    Trace and monitor all requests across LLMs and MCP servers going through the Gateway.
  </Card>

  <Card title="Agent Gateway">
    Register all agents in the Gateway and call them via a single endpoint. Coming Soon.
  </Card>
</CardGroup>
