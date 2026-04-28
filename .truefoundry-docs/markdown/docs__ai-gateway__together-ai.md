> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Together AI

> Learn how to add and configure Together AI models in TrueFoundry AI Gateway for seamless integration and inference.

### Adding Models

This section explains the steps to add Together AI models and configure the required access controls.

<Steps>
  <Step title="Navigate to Together AI Models in AI Gateway">
    From the TrueFoundry dashboard, navigate to `AI Gateway` > `Models` and select `Together AI`.

    <Frame caption="Navigate to Together AI Models">
      <img src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/Screenshot2025-07-23at2.50.29PM-min.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=ad165f2aee7e7b07de1bb4ae653df54b" alt="Navigating to Together AI Provider Account in AI Gateway" width="3594" height="2004" data-path="images/docs/Screenshot2025-07-23at2.50.29PM-min.png" />
    </Frame>
  </Step>

  <Step title="Add Together AI Account Details">
    Click `Add Together AI Account`. Give a unique name to your Together AI account and complete the form with your Together AI authentication details (API Key). Add collaborators to your account, this will give access to the account to other users/teams. Learn more about access control [here](/docs/ai-gateway/gateway-access-control).

    <Frame caption="Together AI Model Account Form">
      <img src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/Screenshot2025-07-23at3.02.18PM-min.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=3812c8181d8b14ed6b968e147529b540" alt="Together AI account configuration form with fields for API key and collaborators" width="2510" height="2014" data-path="images/docs/Screenshot2025-07-23at3.02.18PM-min.png" />
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

<img src="https://mintcdn.com/truefoundry/jw406UAsc7ErYUq8/images/Screenshot2025-07-23at3.38.53PM-min.png?fit=max&auto=format&n=jw406UAsc7ErYUq8&q=85&s=62b2976854929421e14a7a8564b2ea76" alt="Code Snippet and Try in Playgroud Buttons for each model" width="3599" height="2002" data-path="images/Screenshot2025-07-23at3.38.53PM-min.png" />
