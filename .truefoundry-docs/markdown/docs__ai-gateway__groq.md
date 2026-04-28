> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Groq

> Add and configure Groq models in TrueFoundry's AI Gateway for high-performance inference

### Adding Models

This section explains the steps to add Groq models and configure the required access controls.

<Steps>
  <Step title="Navigate to Groq Models in AI Gateway">
    From the TrueFoundry dashboard, navigate to `AI Gateway` > `Models` and select `Groq`.

    <Frame caption="Navigate to Groq Models">
      <img src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/Screenshot2025-07-23at2.38.52PM-min.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=41113947bd29e5e0a2eb61a6364094c8" alt="Groq provider selection in TrueFoundry AI Gateway models page" width="3596" height="2000" data-path="images/docs/Screenshot2025-07-23at2.38.52PM-min.png" />
    </Frame>
  </Step>

  <Step title="Add Groq Account Details">
    Click `Add Groq Account`. Give a unique name to your Groq account and complete the form with your Groq authentication details (API Key). Add collaborators to your account, this will give access to the account to other users/teams. Learn more about access control [here](/docs/ai-gateway/gateway-access-control).

    <Frame caption="Groq Model Account Form">
      <img src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/Screenshot2025-07-23at3.16.03PM-min.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=211f44dfc33fd9cfafcaed4ce9c3ef54" alt="Groq account configuration form with fields for API key and collaborators" width="2508" height="2002" data-path="images/docs/Screenshot2025-07-23at3.16.03PM-min.png" />
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

<img src="https://mintcdn.com/truefoundry/jw406UAsc7ErYUq8/images/Screenshot2025-07-23at3.33.18PM-min.png?fit=max&auto=format&n=jw406UAsc7ErYUq8&q=85&s=3ca25515f4905e59047fec3dab4d750e" alt="Code Snippet and Try in Playgroud Buttons for each Groq model" width="3590" height="2014" data-path="images/Screenshot2025-07-23at3.33.18PM-min.png" />
