> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Cohere

> Add and configure Cohere models in TrueFoundry's AI Gateway

### Adding Models

This section explains the steps to add Cohere models and configure the required access controls.

<Steps>
  <Step title="Navigate to Cohere Models in AI Gateway">
    From the TrueFoundry dashboard, navigate to `AI Gateway` > `Models` and select `Cohere`.

    <Frame caption="Navigate to Cohere Models">
      <img src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/Screenshot2025-07-23at2.35.39PM-min.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=19f2cee0c3c9f212c4f9bacb5f902117" alt="Navigating to Cohere Provider Account in AI Gateway" width="3596" height="2004" data-path="images/docs/Screenshot2025-07-23at2.35.39PM-min.png" />
    </Frame>
  </Step>

  <Step title="Add Cohere Account Details">
    Click `Add Cohere Account`. Give a unique name to your Cohere account and complete the form with your Cohere authentication details (API Key). Add collaborators to your account, this will give access to the account to other users/teams. Learn more about access control [here](/docs/ai-gateway/gateway-access-control).

    <Frame caption="Cohere Model Account Form">
      <img src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/Screenshot2025-07-23at2.59.11PM-min.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=c51e4ccea4b99ba69d1b268cadd32153" alt="Cohere account configuration form with fields for API key and collaborators" width="2536" height="2016" data-path="images/docs/Screenshot2025-07-23at2.59.11PM-min.png" />
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

<img src="https://mintcdn.com/truefoundry/jw406UAsc7ErYUq8/images/Screenshot2025-07-23at3.29.22PM-min.png?fit=max&auto=format&n=jw406UAsc7ErYUq8&q=85&s=4522d54c985c09c391881e6a9ab24c71" alt="Code Snippet and Try in Playgroud Buttons for each model" width="3598" height="2014" data-path="images/Screenshot2025-07-23at3.29.22PM-min.png" />
