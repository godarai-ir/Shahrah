> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Databricks

> Add and configure Databricks models in TrueFoundry's AI Gateway

### Adding Models

This section explains the steps to add Databricks models and configure the required access controls.

<Steps>
  <Step title="Navigate to Databricks Models in AI Gateway">
    From the TrueFoundry dashboard, navigate to `AI Gateway` > `Models` and select `Databricks`.

    <Frame caption="Navigate to Databricks Models">
      <img src="https://mintcdn.com/truefoundry/FeKcq2n1MMm83Par/images/Screenshot2025-07-02at8.51.28PM-min.png?fit=max&auto=format&n=FeKcq2n1MMm83Par&q=85&s=72eb4b2d57fc9998575d1a9728f8ec01" alt="Navigating to Databricks Provider Account in AI Gateway" width="3598" height="2002" data-path="images/Screenshot2025-07-02at8.51.28PM-min.png" />
    </Frame>
  </Step>

  <Step title="Add Databricks Account and Authentication">
    Give a unique name for the Databricks account. This will be used to refer to the models later. Provide the authentication details for the gateway to access your Databricks models. Truefoundry supports both Service Principal and Personal Access Token (PAT) based authentication.

    <Accordion title="Get Databricks Authentication Details">
      **Using Service Principal (Recommended):**

      Service Principal authentication is the recommended approach for production environments as it provides better security and access control.

      * Choose **Service Principal Auth**.
      * Enter your Databricks Service Principal `Client ID` and `OAuth Secret`.

      <Frame caption="Service Principal Authentication">
        <img src="https://mintcdn.com/truefoundry/FeKcq2n1MMm83Par/images/Screenshot2025-07-02at8.55.51PM-min.png?fit=max&auto=format&n=FeKcq2n1MMm83Par&q=85&s=dfd052802cc4ce784e68bed78f14ad57" alt="Databricks Account Configuration Form with Service Principal Client ID and OAuth Secret Fields" width="2644" height="2008" data-path="images/Screenshot2025-07-02at8.55.51PM-min.png" />
      </Frame>

      **Using Personal Access Token (PAT):**

      Personal Access Tokens are suitable for development and testing environments.

      * Choose **Databricks API Key Based Auth**.
      * Enter your `PAT`.

      <Frame caption="Personal Access Token (PAT) Authentication">
        <img src="https://mintcdn.com/truefoundry/FeKcq2n1MMm83Par/images/Screenshot2025-07-02at8.56.44PM-min.png?fit=max&auto=format&n=FeKcq2n1MMm83Par&q=85&s=06ed619205411374790a3bcfcbb48230" alt="Databricks Account Configuration Form with PAT Field" width="2678" height="2006" data-path="images/Screenshot2025-07-02at8.56.44PM-min.png" />
      </Frame>
    </Accordion>

    Finally, enter your Databricks workspace URL (e.g., `https://<workspace_id>.databricks.com`).
  </Step>

  <Step title="Add Models">
    Click `+ Add Model` to add a new model configuration. The `Model ID` in TrueFoundry **must exactly match** the **serving endpoint name** in your Databricks workspace.

    <Accordion title="How to Set Up Databricks Serving Endpoints">
      1. **Access Databricks Serving**: In your Databricks workspace, navigate to **Serving** in the left sidebar and click **Create serving endpoint**.

               <Frame>
                 <img src="https://mintcdn.com/truefoundry/FeKcq2n1MMm83Par/images/Screenshot2025-07-02at9.19.16PM-min.png?fit=max&auto=format&n=FeKcq2n1MMm83Par&q=85&s=3b7bc2883f2e66ef3b6a581c83ec8f53" alt="Databricks workspace interface showing the Serving option in the left sidebar menu" width="3600" height="2006" data-path="images/Screenshot2025-07-02at9.19.16PM-min.png" />
               </Frame>
      2. **Configure Endpoint**:

         * **Endpoint name**: Choose a descriptive name. This name will be your **Model ID** in TrueFoundry.
         * **Served Entity**: Choose from **Foundation Models** or your custom models.

               <Frame>
                 <img src="https://mintcdn.com/truefoundry/FeKcq2n1MMm83Par/images/Screenshot2025-07-02at9.20.24PM-min.png?fit=max&auto=format&n=FeKcq2n1MMm83Par&q=85&s=3097c0c985f9dfc6a381d34228d3a67a" alt="Databricks serving endpoint configuration form with fields for endpoint name and model selection" width="3592" height="2004" data-path="images/Screenshot2025-07-02at9.20.24PM-min.png" />
               </Frame>
      3. **Deploy and Verify**: Click **Create** and wait for the deployment to become **Ready**.
    </Accordion>

    <Frame caption="Add Databricks Model in TrueFoundry">
      <img src="https://mintcdn.com/truefoundry/FeKcq2n1MMm83Par/images/Screenshot2025-07-02at9.26.43PM-min.png?fit=max&auto=format&n=FeKcq2n1MMm83Par&q=85&s=e2466dfa164f107a6929bebd9df06ba6" alt="Databricks model configuration form in TrueFoundry with display name and model ID fields" width="2712" height="2010" data-path="images/Screenshot2025-07-02at9.26.43PM-min.png" />
    </Frame>
  </Step>
</Steps>

### Inference

After adding the models, you can perform inference using an OpenAI-compatible API via the Playground or by integrating it with your own application.

<Frame caption="Infer Model in Playground or Get Code Snippet">
  <img src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/Screenshot2025-07-22at4.41.30PM-min.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=c54cdd7479f216f0775cbc32cdfab18b" alt="Code Snippet and Try in Playgroud Buttons for each model" width="3594" height="1998" data-path="images/docs/Screenshot2025-07-22at4.41.30PM-min.png" />
</Frame>
