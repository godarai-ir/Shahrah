> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# OpenRouter

> Add and configure OpenRouter models in TrueFoundry's AI Gateway

### Adding Models

This section explains the steps to add OpenRouter models and configure the required access controls.

<Steps>
  <Step title="Navigate to OpenRouter Models in AI Gateway">
    From the TrueFoundry dashboard, navigate to `AI Gateway` > `Models` and select `OpenRouter`.

    <Frame caption="Navigate to OpenRouter Models">
      <img src="https://mintcdn.com/truefoundry/n3EuZuJ0K8wBFp1G/images/openrouter-0.png?fit=max&auto=format&n=n3EuZuJ0K8wBFp1G&q=85&s=658858c38388e1ed6e19e4452fb6e74f" alt="Navigating to OpenRouter Provider Account in AI Gateway" width="3840" height="1988" data-path="images/openrouter-0.png" />
    </Frame>
  </Step>

  <Step title="Add OpenRouter Account Details">
    Click `Add OpenRouter Account`. Give a unique name to your OpenRouter account and complete the form with your OpenRouter authentication details (API Key). Add collaborators to your account, this will give access to the account to other users/teams. Learn more about access control [here](/docs/ai-gateway/gateway-access-control).

    <Frame caption="OpenRouter API Key Configuration">
      <img src="https://mintcdn.com/truefoundry/n3EuZuJ0K8wBFp1G/images/openrouter-apikey.png?fit=max&auto=format&n=n3EuZuJ0K8wBFp1G&q=85&s=9f4f1a6f408da3999fff31b3c0cd1a91" alt="OpenRouter account configuration form with fields for API key and collaborators" width="3840" height="1984" data-path="images/openrouter-apikey.png" />
    </Frame>
  </Step>

  <Step title="Add Models">
    To add a model, you need to get the model ID from OpenRouter:

    1. Go to [OpenRouter's models page](https://openrouter.ai/models)
    2. Find the model you want to use
    3. Copy the model ID (e.g., `anthropic/claude-3-haiku`, `openai/gpt-4o-mini`)
    4. Paste the model ID in the "Add Model" form in TrueFoundry AI Gateway

    <Frame caption="OpenRouter Models Page">
      <img src="https://mintcdn.com/truefoundry/n3EuZuJ0K8wBFp1G/images/openrouter-model-page.png?fit=max&auto=format&n=n3EuZuJ0K8wBFp1G&q=85&s=868acbc5aeaec1ba275e9be4ab729ac0" alt="OpenRouter models page showing available models and their IDs" width="3840" height="1984" data-path="images/openrouter-model-page.png" />
    </Frame>

    <Frame caption="Adding Model in TrueFoundry">
      <img src="https://mintcdn.com/truefoundry/n3EuZuJ0K8wBFp1G/images/openrouter-add-model.png?fit=max&auto=format&n=n3EuZuJ0K8wBFp1G&q=85&s=4b1ccca2512215910cc192ea4fb0d12a" alt="Adding OpenRouter model in TrueFoundry AI Gateway" width="3840" height="1984" data-path="images/openrouter-add-model.png" />
    </Frame>

    <Note>
      You can add any model available on OpenRouter by copying its model ID from the OpenRouter website. TrueFoundry AI Gateway supports all text and multimodal models available through OpenRouter.
    </Note>

    <Info>
      The complete list of models supported by OpenRouter can be found [here](https://openrouter.ai/models). Make sure to use the exact model ID as shown on OpenRouter.
    </Info>
  </Step>
</Steps>

### Inference

After adding the models, you can perform inference using an OpenAI-compatible API via the Playground or by integrating with your own application.

<img src="https://mintcdn.com/truefoundry/n3EuZuJ0K8wBFp1G/images/openrouter-inference.png?fit=max&auto=format&n=n3EuZuJ0K8wBFp1G&q=85&s=174de27b9e7e7447cb7fbca8838f7bf2" alt="Code Snippet and Try in Playground Buttons for each model" width="3840" height="1988" data-path="images/openrouter-inference.png" />
