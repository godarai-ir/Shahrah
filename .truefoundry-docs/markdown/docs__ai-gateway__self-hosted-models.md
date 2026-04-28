> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Self-Hosted Models

> Learn how to add and configure self-hosted or external AI models to TrueFoundry AI Gateway for centralized management and inference.

If you run LLMs on your own infrastructure — on-premises GPUs, private cloud instances, or any model server not deployed through TrueFoundry — you can connect them to the AI Gateway by providing the endpoint URL and authentication details.

Once registered, these models appear in the gateway model catalog alongside cloud providers and benefit from all gateway features: routing, guardrails, rate limiting, cost tracking, and observability.

### Prerequisites

Your model server must expose an HTTP endpoint. The gateway works best with **OpenAI-compatible APIs**, which is the default output format for popular inference frameworks:

* [vLLM](https://docs.vllm.ai) — OpenAI-compatible by default
* [Ollama](https://ollama.com) — OpenAI-compatible API
* [SGLang](https://sgl-project.github.io) — OpenAI-compatible by default
* [Text Generation Inference (TGI)](https://huggingface.co/docs/text-generation-inference) — supports OpenAI-compatible mode

### Adding Self-Hosted Models to the Gateway

Follow these steps to add a self-hosted or external model (deployed on TrueFoundry or elsewhere) to the AI Gateway.

<Steps>
  <Step title="Navigate to Self Hosted Models in AI Gateway">
    From the TrueFoundry dashboard, navigate to `AI Gateway` > `Models` and select `Self Hosted Models`.

    <Frame caption="Add model to AI Gateway">
      <img src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/Screenshot2025-07-23at10.33.47AM-min.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=477591a20530d12c470d88f37b42c9af" alt="Navigating to Self Hosted Provider Account in AI Gateway" width="3600" height="2006" data-path="images/docs/Screenshot2025-07-23at10.33.47AM-min.png" />
    </Frame>
  </Step>

  <Step title="Add Collaborators">
    Add collaborators to your account, this will give access to the account to other users/teams. Learn more about access control [here](/docs/ai-gateway/gateway-access-control).

    <Frame caption="Add collaborators for the model">
      <img src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/Screenshot2025-07-23at10.48.42AM1-min.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=6ce3abd930763cb2d6aaf8a07ec8b139" alt="Self Hosted account configuration form with fields for API key and collaborators" width="2516" height="2002" data-path="images/docs/Screenshot2025-07-23at10.48.42AM1-min.png" />
    </Frame>
  </Step>

  <Step title="Configure Model Details">
    Give a name to your hosted model and add `ModelId`, add the `URL` to the hosted model. Choose the correct Model Server for your model and optionally add Auth Data which will be used for Authentication of request to the model.

    <img src="https://mintcdn.com/truefoundry/jw406UAsc7ErYUq8/images/Screenshot2025-07-23at10.36.07AM-min.png?fit=max&auto=format&n=jw406UAsc7ErYUq8&q=85&s=db5f79fe759f5120c1c0253a27b7dc66" alt="Configuration form showing fields for model ID, URL, and authentication" width="2028" height="2014" data-path="images/Screenshot2025-07-23at10.36.07AM-min.png" />
  </Step>
</Steps>

### Inference

After adding the models, you can perform inference using an OpenAI-compatible API via the Playground or by integrating with your own application.

<Frame caption="Inference with Self-Hosted Model">
  <img
    src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/Screenshot2025-07-23at10.36.45AM-min.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=c4c5e575b9aa572aa33c90e932e035be"
    alt="Code Snippet and Try in Playground Buttons for each
model"
    width="3600"
    height="2006"
    data-path="images/docs/Screenshot2025-07-23at10.36.45AM-min.png"
  />
</Frame>
