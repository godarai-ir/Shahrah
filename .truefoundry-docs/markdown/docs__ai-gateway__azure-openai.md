> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Azure OpenAI

> Add and configure Azure OpenAI models through TrueFoundry's AI Gateway

### Adding Models

This section explains the steps to add Azure OpenAI models and configure the required access controls.

<Steps>
  <Step title="Navigate to Azure OpenAI Models in AI Gateway">
    From the TrueFoundry dashboard, navigate to `AI Gateway` > `Models` and select `Azure OpenAI`.

    <Frame caption="Navigate to Azure OpenAI Models">
      <img src="https://mintcdn.com/truefoundry/FeKcq2n1MMm83Par/images/Screenshot2025-07-22at2.47.19PM-min.png?fit=max&auto=format&n=FeKcq2n1MMm83Par&q=85&s=470445afd901609429b4b9a751547290" alt="Navigating to Azure OpenAI Provider Account in AI Gateway" width="3600" height="1996" data-path="images/Screenshot2025-07-22at2.47.19PM-min.png" />
    </Frame>
  </Step>

  <Step title="Add Azure OpenAI Account Details">
    Click `Add Azure OpenAI Account`. Give a unique name to your Azure OpenAI account and complete the form with your Azure authentication details (API Key). For enhanced security, you can also use [certificate-based authentication](/docs/ai-gateway/azure-entra-certificate-auth).
    Add collaborators to your account. You can read more about access control [here](/docs/ai-gateway/gateway-access-control).

    <img src="https://mintcdn.com/truefoundry/FeKcq2n1MMm83Par/images/Screenshot2025-07-22at7.31.51PM-min.png?fit=max&auto=format&n=FeKcq2n1MMm83Par&q=85&s=a48040a58ed3f1860f576bc583c11c85" alt="Azure OpenAI account configuration form with fields for API key and collaborators" width="2522" height="2006" data-path="images/Screenshot2025-07-22at7.31.51PM-min.png" />
  </Step>

  <Step title="Add Models by Deployment Name">
    Click on `+ Add Model` to open the form for adding a new model. For Azure OpenAI, you don't select from a list. Instead, you add models based on your **deployments** in Azure. The `Model ID` in TrueFoundry corresponds to your **Deployment Name** in Azure. First, ensure you have deployed a model in your Azure OpenAI service. You can follow Microsoft's instructions [here](https://learn.microsoft.com/en-us/azure/ai-services/openai/how-to/create-resource?pivots=web-portal). Once deployed, you can find the deployment name in the `Deployments` tab of your Azure AI Studio.

    <Frame caption="Find Model Deployment Name in Azure">
      <img src="https://mintcdn.com/truefoundry/s4Aj2_qGCrSP-zc8/images/7bf8e768-956b58837d4ea6b4a76dbb0da5003d5cecaa64d2988482dca907e5faecb949b2-Screenshot_2025-02-03_at_10.20.09_AM.png?fit=max&auto=format&n=s4Aj2_qGCrSP-zc8&q=85&s=6ac280a61732b0c8b0fd9bcd3212dfa5" alt="Azure AI Studio showing deployed models with deployment names highlighted" width="3020" height="1618" data-path="images/7bf8e768-956b58837d4ea6b4a76dbb0da5003d5cecaa64d2988482dca907e5faecb949b2-Screenshot_2025-02-03_at_10.20.09_AM.png" />
    </Frame>

    Use this deployment name as the `Model ID` when adding the model in TrueFoundry.

    <Frame caption="Adding an Azure OpenAI model to TrueFoundry">
      <img src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/Screenshot2025-07-22at2.59.20PM1-min.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=bcd968d1126b1564ea20072858e89793" alt="Model addition form for Azure OpenAI with fields for model ID and endpoint configuration" width="2022" height="2008" data-path="images/docs/Screenshot2025-07-22at2.59.20PM1-min.png" />
    </Frame>

    <Info>
      Azure OpenAI integration supports any OpenAI model that you have deployed in your Azure account.
    </Info>
  </Step>
</Steps>

### Inference

After adding the models, you can perform inference using an OpenAI-compatible API via the Playground or by integrating with your own application.

<Frame caption="Infer Model in Playground or Get Code Snippet to integrate in your application">
  <img src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/Screenshot2025-07-22at2.59.40PM-min.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=e99175f5bae9352d4682fd5e340e9359" alt="Code Snippet and Try in Playgroud Buttons for each model" width="3600" height="2006" data-path="images/docs/Screenshot2025-07-22at2.59.40PM-min.png" />
</Frame>
