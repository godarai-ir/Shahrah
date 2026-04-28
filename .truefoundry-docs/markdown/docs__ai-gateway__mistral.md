> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Mistral AI

> Add and configure Mistral AI models in TrueFoundry's AI Gateway

### Adding Models

This section explains the steps to add Mistral AI models and configure the required access controls.

<Steps>
  <Step title="Navigate to Mistral AI Models in AI Gateway">
    From the TrueFoundry dashboard, navigate to `AI Gateway` > `Models` and select `Mistral AI`.

    <Frame caption="Navigate to Mistral Models">
      <img src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/Screenshot2025-07-23at2.39.52PM-min.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=27dbf8bd6476b97a91438d2da456422d" alt="Navigating to Mistral Provider Account in AI Gateway" width="3596" height="2010" data-path="images/docs/Screenshot2025-07-23at2.39.52PM-min.png" />
    </Frame>
  </Step>

  <Step title="Add Mistral Account Details">
    Click `Add Mistral Account`. Give a unique name to your Mistral AI account and complete the form with your Mistral authentication details (API Key). Add collaborators to your account, this will give access to the account to other users/teams. Learn more about access control [here](/docs/ai-gateway/gateway-access-control).

    <Frame caption="Mistral AI Model Account Form">
      <img src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/Screenshot2025-07-23at3.01.29PM-min.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=c67625ac84817a8661789ec38c8773df" alt="Mistral account configuration form with fields for API key and collaborators" width="2522" height="2018" data-path="images/docs/Screenshot2025-07-23at3.01.29PM-min.png" />
    </Frame>
  </Step>

  <Step title="Add Models">
    Select the model from the list. If you see the model you want to add in the list of checkboxes, we support public model cost for these models.

    <Note>
      (Optional) If the model you are looking for is not present in the options, you can add it using `+ Add Model` at the end of list (scroll down to see the option) by filling the form.
    </Note>
  </Step>
</Steps>

### Inference

After adding the models, you can perform inference using an OpenAI-compatible API via the Playground or by integrating with your own application.

<img src="https://mintcdn.com/truefoundry/jw406UAsc7ErYUq8/images/Screenshot2025-07-23at3.34.21PM-min.png?fit=max&auto=format&n=jw406UAsc7ErYUq8&q=85&s=2c63f33bb39c0a6652ecc7ef6ede8b32" alt="Code Snippet and Try in Playgroud Buttons for each model" width="3600" height="2004" data-path="images/Screenshot2025-07-23at3.34.21PM-min.png" />
