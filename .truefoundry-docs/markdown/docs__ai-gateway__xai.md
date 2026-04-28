> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# xAI

> Add and configure xAI (Grok) models in TrueFoundry's AI Gateway

### Adding Models

This section explains the steps to add xAI models and configure the required access controls.

<Steps>
  <Step title="Navigate to xAI Models in AI Gateway">
    From the TrueFoundry dashboard, navigate to `AI Gateway` > `Models` and select `xAI`.

    <Frame caption="Navigate to xAI Models">
      <img src="https://mintcdn.com/truefoundry/MMr_lFGYMwcxFTFb/images/docs/xai-provider-selection.png?fit=max&auto=format&n=MMr_lFGYMwcxFTFb&q=85&s=44ef7449da46b13a4be2e3ab7b0b0dd4" alt="xAI provider selection in TrueFoundry AI Gateway models page" width="2880" height="1628" data-path="images/docs/xai-provider-selection.png" />
    </Frame>
  </Step>

  <Step title="Add xAI Account Details">
    Click `Add xAI Account`. Give a unique name to your xAI account and complete the form with your xAI authentication details (API Key). Add collaborators to your account, this will give access to the account to other users/teams. Learn more about access control [here](/docs/ai-gateway/gateway-access-control).

    <Frame caption="xAI Model Account Form">
      <img src="https://mintcdn.com/truefoundry/MMr_lFGYMwcxFTFb/images/docs/xai-account-form.png?fit=max&auto=format&n=MMr_lFGYMwcxFTFb&q=85&s=16f0d68a3fcd48c40168ed3a0ceee40e" alt="xAI account configuration form with fields for API key and collaborators" width="2014" height="1620" data-path="images/docs/xai-account-form.png" />
    </Frame>
  </Step>

  <Step title="Add Models">
    If you see the model you want to add in the list of checkboxes, we support public model cost for these models.

    <Note>
      (Optional) If the model you are looking for is not present in the options, you can add it using `+ Add Model` at the end of list (scroll down to see the option) by filling the form.
    </Note>

    <Info>
      TrueFoundry AI Gateway supports all chat completion models in xAI. The complete list of models supported by xAI can be found [here](https://docs.x.ai/docs/models).
    </Info>
  </Step>
</Steps>

### Inference

After adding the models, you can perform inference using an OpenAI-compatible API via the Playground or by integrating with your own application.

<img src="https://mintcdn.com/truefoundry/MMr_lFGYMwcxFTFb/images/docs/xai-model-playground.png?fit=max&auto=format&n=MMr_lFGYMwcxFTFb&q=85&s=990aec45890eb289fa065cbb30573d67" alt="Code Snippet and Try in Playground Buttons for each xAI model" width="2880" height="1632" data-path="images/docs/xai-model-playground.png" />
