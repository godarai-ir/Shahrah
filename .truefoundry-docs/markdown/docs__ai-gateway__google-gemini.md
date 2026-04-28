> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Google Gemini

> Add and configure Google Gemini models in TrueFoundry's AI Gateway

### Adding Models

This section explains the steps to add Google Gemini models and configure the required access controls.

<Steps>
  <Step title="Navigate to Google Gemini Models in AI Gateway">
    From the TrueFoundry dashboard, navigate to `AI Gateway` > `Models` and select `Google Gemini`.

    <Frame caption="Navigate to Google Gemini Model">
      <img src="https://mintcdn.com/truefoundry/FqWRGsBw-rkZbVrY/images/Screenshot2025-09-25at7.17.33PM.png?fit=max&auto=format&n=FqWRGsBw-rkZbVrY&q=85&s=9d90ddae30fbd8e751f260b15d0b1274" alt="Navigating to Google Gemini Provider Account in AI Gateway" width="3600" height="2010" data-path="images/Screenshot2025-09-25at7.17.33PM.png" />
    </Frame>
  </Step>

  <Step title="Add Google Gemini Account Details">
    Click `Add Google Gemini Account`. Give a unique name to your Gemini account and complete the form with your Gemini authentication details (API Key). Add collaborators to your account, this will give access to the account to other users/teams. Learn more about access control [here](/docs/ai-gateway/gateway-access-control).

    <Frame caption="Google Gemini Model Account Form">
      <img src="https://mintcdn.com/truefoundry/FqWRGsBw-rkZbVrY/images/Screenshot2025-09-25at7.19.23PM.png?fit=max&auto=format&n=FqWRGsBw-rkZbVrY&q=85&s=9241a76658a6e9aec7abe38b1d9f1b1e" alt="Google Gemini account configuration form with fields for API key and collaborators" width="2522" height="2004" data-path="images/Screenshot2025-09-25at7.19.23PM.png" />
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

<img src="https://mintcdn.com/truefoundry/FqWRGsBw-rkZbVrY/images/Screenshot2025-09-25at7.20.22PM.png?fit=max&auto=format&n=FqWRGsBw-rkZbVrY&q=85&s=b8356ce8317fbe21158a5ccb0e450a05" alt="Code Snippet and Try in Playgroud Buttons for each model" width="3600" height="2022" data-path="images/Screenshot2025-09-25at7.20.22PM.png" />
