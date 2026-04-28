> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Cerebras

> Add and configure Cerebras models through TrueFoundry's AI Gateway

### Adding Models

This section explains the steps to add Cerebras models and configure the required access controls.

<Steps>
  <Step title="Navigate to Cerebras Models in AI Gateway">
    From the TrueFoundry dashboard, navigate to `AI Gateway` > `Models` and select `Cerebras`.

    <Frame caption="Navigate to Cerebras Models">
      <img src="https://mintcdn.com/truefoundry/j2SsFkRhddQ9EdtM/images/cerebras-page.png?fit=max&auto=format&n=j2SsFkRhddQ9EdtM&q=85&s=0338eb18ac3f6858d6e4d61e2f0a6858" alt="Navigating to Cerebras Provider Account in AI Gateway" width="1600" height="824" data-path="images/cerebras-page.png" />
    </Frame>
  </Step>

  <Step title="Add Cerebras Account Details">
    Click `Add Cerebras Account`. Give a unique name to your Cerebras account and complete the form with your Cerebras authentication details (API Key).
    Add collaborators to your account. You can read more about access control [here](/docs/ai-gateway/gateway-access-control).

    <Frame caption="Cerebras Account Form">
      <img src="https://mintcdn.com/truefoundry/j2SsFkRhddQ9EdtM/images/cerebras-integrations.png?fit=max&auto=format&n=j2SsFkRhddQ9EdtM&q=85&s=b70bf2b7768e0514666fe001ab41e969" alt="Cerebras account configuration form with fields for API key and collaborators" width="1294" height="957" data-path="images/cerebras-integrations.png" />
    </Frame>
  </Step>

  <Step title="Add Models by Model ID">
    Click on `+ Add Model` to open the form for adding a new model. For Cerebras, you don't select from a list. Instead, you need to get the `Model ID` from their [docs](https://inference-docs.cerebras.ai/models/overview) page.

    <Frame caption="Find Model ID in Cerebras Docs">
      <img src="https://mintcdn.com/truefoundry/j2SsFkRhddQ9EdtM/images/cerebras-docs.png?fit=max&auto=format&n=j2SsFkRhddQ9EdtM&q=85&s=12f124d7dd87eee1532e30e02c97e30a" alt="Cerebras Docs showing available models with ids" width="1839" height="957" data-path="images/cerebras-docs.png" />
    </Frame>

    Use this `Model ID` when adding the model in TrueFoundry.

    <Frame caption="Adding a Cerebras model to TrueFoundry">
      <img src="https://mintcdn.com/truefoundry/hHqnVVkoQBuawwO_/images/docs/cerebras-add-model.png?fit=max&auto=format&n=hHqnVVkoQBuawwO_&q=85&s=14a2aabcd19a10e32dbbc87fdd29787c" alt="Model addition form for Cerebras with fields for Name and Model ID configuration" width="1036" height="957" data-path="images/docs/cerebras-add-model.png" />
    </Frame>
  </Step>
</Steps>

### Inference

After adding the models, you can perform inference using an OpenAI-compatible API via the Playground or by integrating with your own application.

<Frame caption="Infer Model in Playground or Get Code Snippet to integrate in your application">
  <img src="https://mintcdn.com/truefoundry/hHqnVVkoQBuawwO_/images/docs/cerebras-inference.png?fit=max&auto=format&n=hHqnVVkoQBuawwO_&q=85&s=c7054b85fab1985a599bf0f0bb896b0e" alt="Code Snippet and Try in Playgroud Buttons for each model" width="1599" height="827" data-path="images/docs/cerebras-inference.png" />
</Frame>
