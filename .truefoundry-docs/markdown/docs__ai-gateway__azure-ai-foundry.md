> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Azure Foundry

> Add and configure Azure AI Foundry models in TrueFoundry's AI Gateway

### Adding Models

This section explains the steps to add Azure AI Foundry models and configure the required access controls.

<Steps>
  <Step title="Navigate to Azure AI Foundry Models in AI Gateway">
    From the TrueFoundry dashboard, navigate to `AI Gateway` > `Models` and select `Azure AI Foundry`.

    <Frame caption="Navigate to Azure AI Foundry Models">
      <img src="https://mintcdn.com/truefoundry/FeKcq2n1MMm83Par/images/Screenshot2025-07-22at4.19.00PM-min.png?fit=max&auto=format&n=FeKcq2n1MMm83Par&q=85&s=bff6f39dfe08538bf2534e4bf79b3672" alt="Navigating to Azure AI Foundry Provider Account in AI Gateway" width="3592" height="2008" data-path="images/Screenshot2025-07-22at4.19.00PM-min.png" />
    </Frame>
  </Step>

  <Step title="Add Azure AI Foundry Account Details">
    Click `Add Azure AI Foundry Account`. Give a unique name to your Azure AI Foundry account.
    Add collaborators to your account. You can read more about access control [here](/docs/ai-gateway/gateway-access-control).

    <Info>
      Authentication (API Key or Certificate) is configured per-model in the next step, not at the account level.
    </Info>

    <img src="https://mintcdn.com/truefoundry/FeKcq2n1MMm83Par/images/Screenshot2025-07-22at7.31.37PM-min.png?fit=max&auto=format&n=FeKcq2n1MMm83Par&q=85&s=e0cf8381b8ba56143841bb3cc989a1d0" alt="Azure AI Foundry account configuration form with collaborator fields" width="2520" height="2004" data-path="images/Screenshot2025-07-22at7.31.37PM-min.png" />
  </Step>

  <Step title="Add Models from Azure AI Foundry">
    Click on `+ Add Model` to open the model form. For Azure AI Foundry, you add models based on your **deployments** in Azure. First, ensure you have deployed a model in your Azure AI Foundry project. You can follow Microsoft's instructions [here](https://learn.microsoft.com/en-us/azure/ai-foundry/how-to/model-catalog-overview).

    Once deployed, navigate to your deployment in the Azure AI Foundry portal to find the **Target URI** (endpoint URL), **Deployment Name**, and **API Key**.

    <Frame caption="Find Endpoint URL and Model Name in Azure AI Foundry">
      <img src="https://mintcdn.com/truefoundry/nRWhyNXPvFcPqtDW/images/azure-foundry-deployment-annotated.png?fit=max&auto=format&n=nRWhyNXPvFcPqtDW&q=85&s=89459ec540c4e0ea59272358860b2e4b" alt="Azure portal showing deployed model Target URI, deployment name, and API key" width="2948" height="1518" data-path="images/azure-foundry-deployment-annotated.png" />
    </Frame>

    Fill in the model form with the following details:

    | Field                     | What to Enter                                                                      | Where to Find in Azure                          |
    | ------------------------- | ---------------------------------------------------------------------------------- | ----------------------------------------------- |
    | **Display Name**          | name to identify this model in TrueFoundry                                         | Your choice (e.g., `gpt-4o-mini`)               |
    | **Azure Deployment Name** | The deployment name from Azure AI Foundry (not the base model name)                | Deployments → click deployment → **Name** field |
    | **Azure Endpoint**        | The full endpoint URL excluding the API path and query parameters (see below)      | Deployments → click deployment → **Target URI** |
    | **Authentication**        | API Key or [Certificate-based auth](/docs/ai-gateway/azure-entra-certificate-auth) | Deployments → click deployment → **Key**        |
    | **Model Types**           | Select the capabilities of this model (Chat, Embedding, etc.)                      | Based on the model you deployed                 |

    #### Setting the Azure Endpoint

    Copy the **Target URI** from your deployment in the Azure AI Foundry portal, but **remove the API path and query parameters**, gateway appends those automatically.

    | Model type                                                      | Azure gives you                                                                    | Enter in TrueFoundry                                 |
    | --------------------------------------------------------------- | ---------------------------------------------------------------------------------- | ---------------------------------------------------- |
    | **Most models** (OpenAI, Mistral, DeepSeek, Meta, Cohere, etc.) | `https://<resource>.services.ai.azure.com/models/chat/completions?api-version=...` | `https://<resource>.services.ai.azure.com/models`    |
    | **Anthropic (Claude)**                                          | `https://<resource>.services.ai.azure.com/anthropic/v1/messages?api-version=...`   | `https://<resource>.services.ai.azure.com/anthropic` |
    | **Realtime models**                                             | `https://<resource>.services.ai.azure.com/openai/realtime?api-version=...`         | `https://<resource>.services.ai.azure.com`           |

    <Frame caption="Adding an Azure AI Foundry model to TrueFoundry">
      <img src="https://mintcdn.com/truefoundry/nRWhyNXPvFcPqtDW/images/azure-ai-foundry-model-addition.png?fit=max&auto=format&n=nRWhyNXPvFcPqtDW&q=85&s=3e0840d2bc8fba1c10220de25c12ac6c" alt="Model addition form for Azure AI Foundry with fields for model name, endpoint URL, and authentication" width="1564" height="1518" data-path="images/azure-ai-foundry-model-addition.png" />
    </Frame>

    <Info>
      Azure AI Foundry integration supports various AI models including OpenAI, Meta Llama, Mistral, DeepSeek, Cohere, and Anthropic Claude models deployed in your Azure account.
    </Info>
  </Step>
</Steps>

### Inference

After adding the models, you can perform inference using an OpenAI-compatible API via the Playground or by integrating with your own application.

<Frame caption="Test in Playground">
  <img src="https://mintcdn.com/truefoundry/FeKcq2n1MMm83Par/images/Screenshot2025-07-22at4.13.25PM-min.png?fit=max&auto=format&n=FeKcq2n1MMm83Par&q=85&s=38c13c20bcea01e1dc836c0d10bd4251" alt="Code Snippet and Try in Playgroud Buttons for each model" width="3600" height="2012" data-path="images/Screenshot2025-07-22at4.13.25PM-min.png" />
</Frame>

### Mistral OCR - Document Processing

Extract text from documents while maintaining structure and formatting (headers, paragraphs, lists, tables) using the Mistral OCR model via Azure AI Foundry. Returns markdown format and supports multiple formats including PDFs, images (png, jpeg, avif), and office documents (pptx, docx).

<Warning>
  This endpoint cannot be used via the Mistral SDK. The Mistral SDK automatically appends `/v1` to the base URL, which causes a URL mismatch (e.g. the request is sent to `<base_url>/v1/ocr` instead of `<base_url>/ocr`). Use direct HTTP requests (e.g. `requests` in Python) as shown below. See the [open GitHub issue](https://github.com/mistralai/client-python/issues/292) for details.
</Warning>

```python lines theme={"dark"}
import base64
import requests
import json

def encode_pdf(pdf_path):
    with open(pdf_path, "rb") as pdf_file:
        return base64.b64encode(pdf_file.read()).decode("utf-8")

pdf_path = "path-to-your-pdf"
base64_pdf = encode_pdf(pdf_path)

url = "{GATEWAY_BASE_URL}/proxy/ocr"

headers = {
    "Content-Type": "application/json",
    "Authorization": "Bearer your-truefoundry-api-key"
}

# Use your TrueFoundry Azure Foundry model name (e.g. azure-foundry/mistral-ocr)
payload = {
    "model": "azure-foundry/mistral-ocr",
    "document": {
        "type": "document_url",
        "document_url": f"data:application/pdf;base64,{base64_pdf}"
    },
    "include_image_base64": True
}

response = requests.post(url, headers=headers, json=payload)

if response.status_code == 200:
    # Save output to file
    with open("ocr_output.json", "w") as f:
        json.dump(response.json(), f, indent=2)
    print("OCR output saved to ocr_output.json")
else:
    print(f"Error: {response.status_code}")
    print(response.text)
```
