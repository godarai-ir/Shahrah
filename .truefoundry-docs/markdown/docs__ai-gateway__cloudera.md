> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Cloudera

> Add and configure Cloudera models in TrueFoundry's AI Gateway

### Adding Models

This section explains the steps to add Cloudera models and configure the required access controls.

<Steps>
  <Step title="Navigate to Cloudera Models in AI Gateway">
    From the TrueFoundry dashboard, navigate to `AI Gateway` > `Models` and select `Cloudera`.

    <Frame caption="Navigate to Cloudera Models">
      <img src="https://mintcdn.com/truefoundry/gmANRVM1Uk3etyPP/images/Screenshot2025-11-06at5.48.41PM.png?fit=max&auto=format&n=gmANRVM1Uk3etyPP&q=85&s=4d6c9d60449ffe02de2bf44c8cf23bb0" alt="Navigating to Cloudera Provider Account in AI Gateway" width="3600" height="2016" data-path="images/Screenshot2025-11-06at5.48.41PM.png" />
    </Frame>
  </Step>

  <Step title="Add Cloudera Account and Authentication">
    Give a unique name for the Cloudera account. This will be used to refer to the models later. Provide the authentication details for the gateway to access your Cloudera CDP environment. TrueFoundry supports both `CDP Workload Authentication` (recommended) and `Auth Token based` authentication.

    <Accordion title="Get Cloudera Authentication Details">
      **Using CDP Workload Authentication (Recommended):**

      CDP Workload Authentication is the recommended approach for production environments as it provides better security and access control. This method generates secure workload tokens using your Access key and private key.

      Choose **CDP Workload Auth Token**.

      * Enter your `Access Key ID`, `Private Key`, and `Workload Name`.
      * Enter your `CDP Base URL`. This is the base URL for CDP API services used for authentication token generation. You can find your CDP Base URL in your Cloudera CDP environment settings or documentation.

      <Note>
        Store your `Private Key` in a Secret Group and reference it using the secret's Fully Qualified Name (FQN) in the Private Key field. This ensures your private key is securely managed and not exposed in plain text.
      </Note>

      <Note>
        The `CDP Base URL` is different from the Model Base URL. The CDP Base URL is used for authentication token generation, while the Model Base URL (configured per model) is used for inference requests.
      </Note>

      <Frame caption="CDP Workload Authentication">
        <img src="https://mintcdn.com/truefoundry/acKb5p55patKgIwU/images/Screenshot2025-11-25at1.56.30PM.png?fit=max&auto=format&n=acKb5p55patKgIwU&q=85&s=24de2b45a7cebc5b13180f8461b1ef82" alt="Cloudera Account Configuration Form with Access Key ID, Private Key, Workload Name, and CDP Base URL Fields" width="2526" height="2012" data-path="images/Screenshot2025-11-25at1.56.30PM.png" />
      </Frame>

      **Using Auth Token:**

      Auth Token is suitable for development and testing environments.

      * Choose **Token Auth**.
      * Enter your pre-generated `Auth Token`. You can generate this token using the CDP CLI or from your user profile settings.

      <Frame caption="Auth Token Authentication">
        <img src="https://mintcdn.com/truefoundry/acKb5p55patKgIwU/images/Screenshot2025-11-25at1.57.12PM.png?fit=max&auto=format&n=acKb5p55patKgIwU&q=85&s=a909c5c63032d19c9666701b2859247c" alt="Cloudera Account Configuration Form with Auth Token Field" width="2526" height="2008" data-path="images/Screenshot2025-11-25at1.57.12PM.png" />
      </Frame>
    </Accordion>
  </Step>

  <Step title="Add Models">
    Click `+ Add Model` to add a new model configuration. Provide the following:

    * **Model ID**: The unique identifier of the model in the Cloudera Model Registry. This is used in the request body and can differ from the endpoint name in the URL. Find this in the Cloudera CDP Model Registry or Model Endpoint Details UI.
    * **Base URL**: The model endpoint URL up to `/v1`. Copy the full endpoint URL from your Cloudera AI Inference service and remove the last two path components (e.g., remove `/chat/completions`). The Base URL should end with `/v1`.

    <Frame caption="Add Cloudera Model in TrueFoundry">
      <img src="https://mintcdn.com/truefoundry/acKb5p55patKgIwU/images/Screenshot2025-11-25at1.58.00PM.png?fit=max&auto=format&n=acKb5p55patKgIwU&q=85&s=8494f19803432e07a07be3ccea1cb38d" alt="Cloudera model configuration form in TrueFoundry with display name, model ID, and base URL fields" width="2518" height="2008" data-path="images/Screenshot2025-11-25at1.58.00PM.png" />
    </Frame>
  </Step>
</Steps>

### Inference

After adding the models, you can perform inference using an OpenAI-compatible API via the Playground or by integrating it with your own application.

<Frame caption="Infer Model in Playground or Get Code Snippet">
  <img src="https://mintcdn.com/truefoundry/gmANRVM1Uk3etyPP/images/Screenshot2025-11-06at5.50.17PM.png?fit=max&auto=format&n=gmANRVM1Uk3etyPP&q=85&s=a2cd24cc466e50de8f2e525fac555031" alt="Code Snippet and Try in Playgroud Buttons for each model" width="3584" height="2008" data-path="images/Screenshot2025-11-06at5.50.17PM.png" />
</Frame>
