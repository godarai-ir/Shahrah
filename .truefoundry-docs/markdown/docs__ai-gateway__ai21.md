> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# AI21

> Add and configure AI21 language models in TrueFoundry's AI Gateway

### Adding Models

This section explains the steps to add AI21 models and configure the required access controls.

<Steps>
  <Step title="Navigate to AI21 Models in AI Gateway">
    From the TrueFoundry dashboard, navigate to `AI Gateway` > `Models` and select `AI21`.

    <Frame caption="Navigate to AI21 Models">
      <img src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/Screenshot2025-07-23at2.37.03PM-min.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=d6e685f1c5bf3659c2c2f737334af99f" alt="Navigating to AI21 Provider Account in AI Gateway" width="3590" height="2008" data-path="images/docs/Screenshot2025-07-23at2.37.03PM-min.png" />
    </Frame>
  </Step>

  <Step title="Add AI21 Account Details">
    Click `Add AI21 Account`. Give a unique name to your AI21 account and complete the form with your AI21 authentication details (API Key). Add collaborators to your account, this will give access to the account to other users/teams. Learn more about access control [here](/docs/ai-gateway/gateway-access-control).

    <Frame caption="AI21 Model Account Form">
      <img src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/Screenshot2025-07-23at2.59.31PM-min.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=08bf3ca07123e8a68253299c5907154d" alt="AI21 account configuration form with fields for API key and collaborators" width="2524" height="2006" data-path="images/docs/Screenshot2025-07-23at2.59.31PM-min.png" />
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

<img src="https://mintcdn.com/truefoundry/jw406UAsc7ErYUq8/images/Screenshot2025-07-23at3.30.48PM-min.png?fit=max&auto=format&n=jw406UAsc7ErYUq8&q=85&s=1013bb3c27cf61ea46480ee3d91237ad" alt="Code Snippet and Try in Playgroud Buttons for each model" width="3599" height="2008" data-path="images/Screenshot2025-07-23at3.30.48PM-min.png" />
