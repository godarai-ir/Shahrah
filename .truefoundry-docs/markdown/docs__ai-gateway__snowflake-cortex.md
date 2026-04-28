> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Snowflake Cortex

> Add and configure Cortex models through TrueFoundry's AI Gateway

### Adding Models

This section explains the steps to add Cortex models and configure the required access controls.

<Steps>
  <Step title="Navigate to Cortex Models in AI Gateway">
    From the TrueFoundry dashboard, navigate to `AI Gateway` > `Models` and select `Snowflake Cortex`.

    <Frame caption="Navigate to Snowflake Cortex Models">
      <img src="https://mintcdn.com/truefoundry/qdHpJvNuVln8agiy/images/Snowflakecortex-page.png?fit=max&auto=format&n=qdHpJvNuVln8agiy&q=85&s=2bc4880ddb183efdb8c63da4f29d2470" alt="Navigating to Snowflake Cortex Provider Account in AI Gateway" width="1600" height="832" data-path="images/Snowflakecortex-page.png" />
    </Frame>
  </Step>

  <Step title="Add Snowflake Cortex Account Details">
    Click `Add Snowflake Cortex Account`. Give a unique name to your Snowflake cortex account and complete the form with your Snowflake cortex authentication details (PAT Token)
    and account identifier for snowflake. For more details on account identifier check [here](https://docs.snowflake.com/en/user-guide/admin-account-identifier).
    Then add collaborators to your account. You can read more about access control [here](/docs/ai-gateway/gateway-access-control).

    <Frame caption="Snowflake Cortex Account Form">
      <img src="https://mintcdn.com/truefoundry/qdHpJvNuVln8agiy/images/Snowflakecortex-integrations.png?fit=max&auto=format&n=qdHpJvNuVln8agiy&q=85&s=f7fa5bf90567f59dcef33c8597f9f88e" alt="Snowflake cortex account configuration form with fields for API key and collaborators" width="1294" height="957" data-path="images/Snowflakecortex-integrations.png" />
    </Frame>
  </Step>

  <Step title="Add Models by Model ID">
    Click on `+ Add Model` to open the form for adding a new model. For Snowflake cortex, you don't select from a list. Instead, you need to get the `Model ID` from their [docs](https://docs.snowflake.com/en/user-guide/snowflake-cortex/aisql) page.

    <Frame caption="Find Model ID in Snowflake Cortex Docs">
      <img src="https://mintcdn.com/truefoundry/qdHpJvNuVln8agiy/images/Snowflakecortex-docs.png?fit=max&auto=format&n=qdHpJvNuVln8agiy&q=85&s=afca4a4498434dc0a4044ffed7d74f71" alt="Snowflake Cortex docs showing available models with ids" width="1858" height="969" data-path="images/Snowflakecortex-docs.png" />
    </Frame>

    Use this `Model ID` when adding the model in TrueFoundry.

    <Frame caption="Adding a Snowflake cortex model to TrueFoundry">
      <img src="https://mintcdn.com/truefoundry/aRB3-wXHAGvGOYRm/images/docs/Snowflakecortex-add-model.png?fit=max&auto=format&n=aRB3-wXHAGvGOYRm&q=85&s=e3461b8848d93d81307524aa52e3070b" alt="Model addition form for Snowflake cortex with fields for Name and Model ID configuration" width="1041" height="962" data-path="images/docs/Snowflakecortex-add-model.png" />
    </Frame>
  </Step>
</Steps>

### Inference

After adding the models, you can perform inference using an OpenAI-compatible API via the Playground or by integrating with your own application.

<Frame caption="Infer Model in Playground or Get Code Snippet to integrate in your application">
  <img src="https://mintcdn.com/truefoundry/aRB3-wXHAGvGOYRm/images/docs/Snowflakecortex-inference.png?fit=max&auto=format&n=aRB3-wXHAGvGOYRm&q=85&s=b0e5b2e23d22261a307fe0fd7f15122f" alt="Code Snippet and Try in Playgroud Buttons for each model" width="1600" height="824" data-path="images/docs/Snowflakecortex-inference.png" />
</Frame>
