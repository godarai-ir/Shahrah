> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# SambaNova

> Add and configure SambaNova models through TrueFoundry's AI Gateway

### Adding Models

This section explains the steps to add SambaNova models and configure the required access controls.

<Steps>
  <Step title="Navigate to SambaNova Models in AI Gateway">
    From the TrueFoundry dashboard, navigate to `AI Gateway` > `Models` and select `SambaNova`.

    <Frame caption="Navigate to SambaNova Models">
      <img src="https://mintcdn.com/truefoundry/00jXIHQoSJESU4_H/images/sambanova-page.png?fit=max&auto=format&n=00jXIHQoSJESU4_H&q=85&s=ff6640f6523b661c753f80300cce6214" alt="Navigating to SambaNova Provider Account in AI Gateway" width="1600" height="828" data-path="images/sambanova-page.png" />
    </Frame>
  </Step>

  <Step title="Add SambaNova Account Details">
    Click `Add SambaNova Account`. Give a unique name to your SambaNova account and complete the form with your SambaNova authentication details (API Key).
    Add collaborators to your account. You can read more about access control [here](/docs/ai-gateway/gateway-access-control).

    <Frame caption="SambaNova Account Form">
      <img src="https://mintcdn.com/truefoundry/00jXIHQoSJESU4_H/images/sambanova-integrations.png?fit=max&auto=format&n=00jXIHQoSJESU4_H&q=85&s=0b0c307bdb0524b6de18c42349182aa6" alt="SambaNova account configuration form with fields for API key and collaborators" width="1300" height="962" data-path="images/sambanova-integrations.png" />
    </Frame>
  </Step>

  <Step title="Add Models by Model ID">
    Click on `+ Add Model` to open the form for adding a new model. For SambaNova, you don't select from a list. Instead, you need to get the `Model ID` from their [docs](https://docs.sambanova.ai/docs/en/models/sambacloud-models) page.

    <Frame caption="Find Model ID in SambaNova Docs">
      <img src="https://mintcdn.com/truefoundry/00jXIHQoSJESU4_H/images/sambanova-docs.png?fit=max&auto=format&n=00jXIHQoSJESU4_H&q=85&s=1deda9968d4c43f70095bdf577d321be" alt="SambaNova Docs showing available models with ids" width="1600" height="821" data-path="images/sambanova-docs.png" />
    </Frame>

    Use this `Model ID` when adding the model in TrueFoundry.

    <Frame caption="Adding a SambaNova model to TrueFoundry">
      <img src="https://mintcdn.com/truefoundry/RPGBvdj58v511tQq/images/docs/sambanova-add-model.png?fit=max&auto=format&n=RPGBvdj58v511tQq&q=85&s=df83d4e5f80d626a080cad0571843830" alt="Model addition form for SambaNova with fields for Name and Model ID configuration" width="1037" height="961" data-path="images/docs/sambanova-add-model.png" />
    </Frame>
  </Step>
</Steps>

### Inference

After adding the models, you can perform inference using an OpenAI-compatible API via the Playground or by integrating with your own application.

<Frame caption="Infer Model in Playground or Get Code Snippet to integrate in your application">
  <img src="https://mintcdn.com/truefoundry/RPGBvdj58v511tQq/images/docs/sambanova-inference.png?fit=max&auto=format&n=RPGBvdj58v511tQq&q=85&s=4efe240bc1f3158f0886ca3f75f7bf00" alt="Code Snippet and Try in Playgroud Buttons for each model" width="1600" height="828" data-path="images/docs/sambanova-inference.png" />
</Frame>
