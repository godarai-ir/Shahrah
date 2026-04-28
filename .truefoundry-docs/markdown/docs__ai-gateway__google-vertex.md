> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Google Vertex

> Add and configure Google Vertex AI models including Gemini, Anthropic, and Mistral in TrueFoundry's AI Gateway

### Adding Models

This section explains the steps to add Google Vertex AI models and configure the required access controls.

<Steps>
  <Step title="Navigate to Google Vertex Models in AI Gateway">
    From the TrueFoundry dashboard, navigate to `AI Gateway` > `Models` and select `Google Vertex`.

    <Frame caption="Navigate to Google Vertex Models">
      <img src="https://mintcdn.com/truefoundry/s4Aj2_qGCrSP-zc8/images/Screenshot2025-06-24at3.41.39PM-min.png?fit=max&auto=format&n=s4Aj2_qGCrSP-zc8&q=85&s=872a46ac63dd635f002c30938938dbe1" alt="Navigating to Google Vertex Provider Account in AI Gateway" width="3600" height="2006" data-path="images/Screenshot2025-06-24at3.41.39PM-min.png" />
    </Frame>
  </Step>

  <Step title="Add Google Vertex Account and Authentication">
    Give a unique name to your Google Vertex account. This will be used to refer to the models later. Add collaborators to your account, this will give access to the account to other users/teams. Learn more about access control [**here**](https://docs.truefoundry.com/gateway/gateway-access-control).

    <Accordion title="Get Google Vertex Authentication Details">
      **Required IAM Role**

      The Google Cloud identity used by the gateway (a service account, whether referenced by a key, by GKE Workload Identity, or by Workload Identity Federation) must have the `Vertex AI User` role (`roles/aiplatform.user`), which includes the `aiplatform.endpoints.predict` permission required by the gateway.

      The gateway supports three authentication methods. Pick the one that matches your deployment.

      **1. Using Service Account JSON Key**

      This method works for all deployment types (GKE, EKS, AKS, on-premises, or the SaaS Gateway).

      * Generate a Service Account JSON key by following the official Google Cloud documentation [here](https://cloud.google.com/iam/docs/creating-managing-service-account-keys).
      * The service account must have the `Vertex AI User` role.
      * When adding the provider account in TrueFoundry, select **Service account key file** as the authentication type and paste the JSON key into the **Service account key JSON** field (or store it as a secret and reference it).

      **2. Using Workload Identity Federation (Keyless, Cross-Cloud)**

      Workload Identity Federation (WIF) lets the gateway authenticate to Google Cloud **without service account keys**, even when running **outside of GKE** — for example, on Amazon EKS, Azure AKS, or on-premises Kubernetes clusters. It works by exchanging a short-lived Kubernetes service account token for a Google Cloud access token through Google's Security Token Service.

      <Info>
        Workload Identity Federation is the recommended approach for production deployments running outside of GKE. It eliminates long-lived service account keys while supporting any Kubernetes environment, and it also works on the **SaaS** version of the AI Gateway.
      </Info>

      **Prerequisites**

      1. A Google Cloud project with Vertex AI enabled.
      2. A **Workload Identity Pool** and **Provider** configured in Google Cloud IAM. Follow the official guide: [Configure Workload Identity Federation with Kubernetes](https://cloud.google.com/iam/docs/workload-identity-federation-with-kubernetes).
      3. A Google Cloud IAM service account that the federated identity can impersonate, with the `Vertex AI User` role (`roles/aiplatform.user`) granted on the project.
      4. The Kubernetes service account used by the gateway must have permission to issue `TokenRequest` resources for itself. The TrueFoundry-provided Helm chart configures this RBAC automatically.

      **Generate the credential configuration JSON**

      Use the `gcloud` CLI to generate the credential configuration file:

      ```bash theme={"dark"}
      gcloud iam workload-identity-pools create-cred-config \
        projects/<PROJECT_NUMBER>/locations/global/workloadIdentityPools/<POOL_ID>/providers/<PROVIDER_ID> \
        --service-account=<GSA_EMAIL> \
        --credential-source-type=programmatic \
        --output-file=credential-config.json
      ```

      This produces a JSON file with `"type": "external_account"` describing the identity pool, audience, and STS token-exchange endpoints. It is **not** a private key.

      **Configure in TrueFoundry**

      When adding or editing the Vertex AI provider account:

      1. Select **Workload Identity Federation file** as the authentication type.
      2. Paste the contents of the generated `credential-config.json` into the **Key file content** field, or store it as a secret and reference it.

      <Warning>
        Resumable file uploads (used for some batch and fine-tuning workflows that upload files to Google Cloud Storage via signed URLs) are not yet supported with Workload Identity Federation. If you rely on those flows, use a Service Account JSON key instead.
      </Warning>

      **3. Using GCP Workload Identity on GKE (Self-Hosted Gateway only)**

      When running the gateway inside Google Kubernetes Engine (GKE), you can rely on GKE's built-in Workload Identity, which lets a Kubernetes service account (KSA) act as a Google Cloud IAM service account (GSA) automatically through the GKE metadata server.

      <Info>
        GKE Workload Identity is GKE-specific. Pods using the configured KSA authenticate as the associated GSA when accessing Google Cloud APIs, with no extra configuration on the gateway side.
      </Info>

      To set up GKE Workload Identity, follow the official Google Cloud documentation: [Configure Workload Identity on GKE](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity).

      When adding the Vertex AI provider account in TrueFoundry, **leave the authentication section empty** — the gateway will automatically pick up GKE Workload Identity credentials via Application Default Credentials (ADC).

      <Note>
        GCP Workload Identity (GKE ADC) does **not** work on the SaaS version of the Gateway, and it only works when the gateway runs inside a GKE cluster. For all other environments, use Workload Identity Federation or a Service Account JSON key.
      </Note>
    </Accordion>

    <Frame caption="Add Vertex Provider Account">
      <img src="https://mintcdn.com/truefoundry/FeKcq2n1MMm83Par/images/Screenshot2025-06-24at4.06.27PM-min.png?fit=max&auto=format&n=FeKcq2n1MMm83Par&q=85&s=2c98df837648bda353e02b656d5691dc" alt="Google Vertex account configuration form with fields for name, project ID, service account JSON, and region" width="2642" height="2002" data-path="images/Screenshot2025-06-24at4.06.27PM-min.png" />
    </Frame>
  </Step>

  <Step title="Configure Project ID and Region">
    Provide your Google Cloud **Project ID** and a default **Region** for all models under this account. You can override the region for individual models later.

    **Project ID**

    * You can find your **Project ID** in the top-right corner of your Google Cloud Console.

    <Frame caption="Finding your Project ID in Google Cloud Console">
      <img src="https://mintcdn.com/truefoundry/OHzlp6GY5G-JfKle/images/c6c18af2-b3f40abdc4b83512a66eecc762ce07a9d8af5ff88efb7570b55684fe30ee7437-Screenshot_2025-01-30_at_11.24.27_AM.png?fit=max&auto=format&n=OHzlp6GY5G-JfKle&q=85&s=3cf75338f529454814d69ea804fe66dd" alt="Google Cloud Console header showing project ID location in the dropdown menu" width="3016" height="1424" data-path="images/c6c18af2-b3f40abdc4b83512a66eecc762ce07a9d8af5ff88efb7570b55684fe30ee7437-Screenshot_2025-01-30_at_11.24.27_AM.png" />
    </Frame>

    **Region**

    * Specify a default region for all models under this account. You can override this region for individual models later.
  </Step>

  <Step title="Add Models">
    You can either select available models from the list or add them manually by clicking `+ Add Model`. When adding a model manually, the `Model ID` format depends on the provider.

    <Accordion title="Adding Google (Gemini) Models">
      Select a Gemini model from the list or add it manually.

      * **Model ID Format**: `google/<vertex-model-id>`
      * **Example**: `google/gemini-1.5-pro`

      You can find the Model ID in the Google Cloud Console.

      <Frame caption="Find Gemini Model ID in Google Console">
        <img src="https://mintcdn.com/truefoundry/s4Aj2_qGCrSP-zc8/images/Screenshot2025-06-24at3.35.57PM-min.png?fit=max&auto=format&n=s4Aj2_qGCrSP-zc8&q=85&s=8602f748c8e104db216de6a7ac500a29" alt="Google Cloud Console showing Gemini model details with model ID highlighted" width="3076" height="1788" data-path="images/Screenshot2025-06-24at3.35.57PM-min.png" />
      </Frame>
    </Accordion>

    <Accordion title="Adding Anthropic Models">
      Select a Claude model from the list or add it manually.

      * **Model ID Format**: `anthropic/<vertex-model-id>`
      * **Example**: `anthropic/claude-3-5-sonnet-v2@20241022`

      <Frame caption="Find Anthropic Model ID in Google Console">
        <img src="https://mintcdn.com/truefoundry/FeKcq2n1MMm83Par/images/Screenshot2025-06-24at4.24.26PM.png?fit=max&auto=format&n=FeKcq2n1MMm83Par&q=85&s=5e84496e9a4e01a8f4f4df9233f9b2e7" alt="Google Cloud Console showing Anthropic Claude model details with model ID highlighted" width="2994" height="1906" data-path="images/Screenshot2025-06-24at4.24.26PM.png" />
      </Frame>
    </Accordion>

    <Accordion title="Adding Mistral AI Models">
      Select a Mistral model from the list or add it manually.

      * **Model ID Format**: `mistralai/<vertex-model-id>`
      * **Example**: `mistralai/mistral-large-2411@001`

      <Frame caption="Find Mistral Model ID in Google Console">
        <img src="https://mintcdn.com/truefoundry/OHzlp6GY5G-JfKle/images/c6fb5521-ce958f64be249cf6a9a1cfc3a10ba22412b8a1b17ffdc1d46edfafc1a692b4a6-Screenshot_2025-02-03_at_9.58.14_AM.png?fit=max&auto=format&n=OHzlp6GY5G-JfKle&q=85&s=3508a99f8b4602f221d65d3b688dfbbc" alt="Google Cloud Console showing Mistral AI model details with model ID highlighted" width="2992" height="1534" data-path="images/c6fb5521-ce958f64be249cf6a9a1cfc3a10ba22412b8a1b17ffdc1d46edfafc1a692b4a6-Screenshot_2025-02-03_at_9.58.14_AM.png" />
      </Frame>
    </Accordion>

    <Note>
      When adding any model manually, you can specify a **Region** to override the default one set at the account level.
    </Note>
  </Step>
</Steps>

### Inference

After adding the models, you can perform inference using an OpenAI-compatible API via the Playground or by integrating it with your own application.

<Frame caption="Infer Model in Playground or Get Code Snippet">
  <img src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/Screenshot2025-07-22at4.52.04PM-min.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=c1917cc4fe04f43ce77abc7cff651839" alt="Code Snippet and Try in Playgroud Buttons for each Google Vertex model" width="3600" height="2002" data-path="images/docs/Screenshot2025-07-22at4.52.04PM-min.png" />
</Frame>

### FAQs

<Accordion title="Do I need to add multiple provider accounts for different regions?">
  No. You can set a default region at the account level and override it for each individual model if needed. This allows you to use models from different regions with a single provider account.
</Accordion>

<Accordion title="Which authentication method should I choose?">
  * **Service Account JSON Key** — Works everywhere (any cloud, on-prem, SaaS Gateway). Simplest to set up, but requires you to manage and rotate a long-lived secret.
  * **Workload Identity Federation** — Recommended for production. Keyless, works on any Kubernetes cluster (EKS, AKS, GKE, on-prem) and on the SaaS Gateway. Requires a one-time setup of a Workload Identity Pool in Google Cloud.
  * **GCP Workload Identity (GKE)** — Only available when the self-hosted gateway runs inside a GKE cluster. Keyless and zero-config on the gateway side, but does not work on the SaaS Gateway or outside of GKE.

  |                                         | Service Account Key       | Workload Identity Federation    | GCP Workload Identity (GKE) |
  | --------------------------------------- | ------------------------- | ------------------------------- | --------------------------- |
  | Works on GKE                            | Yes                       | Yes                             | Yes                         |
  | Works on EKS / AKS / on-prem            | Yes                       | Yes                             | No                          |
  | Works on SaaS Gateway                   | Yes                       | Yes                             | No                          |
  | Key management required                 | Yes                       | No                              | No                          |
  | Requires credential JSON in TrueFoundry | Yes (service account key) | Yes (`external_account` config) | No (leave empty)            |
</Accordion>

<Accordion title="What is the difference between GCP Workload Identity and Workload Identity Federation?">
  Both are keyless authentication mechanisms, but they target different environments.

  **GCP Workload Identity** is a GKE-only feature. The GKE metadata server automatically maps a Kubernetes service account to a Google Cloud IAM service account. The gateway picks this up through Application Default Credentials (ADC) when no auth data is configured. It does not work on the SaaS Gateway or outside of GKE.

  **Workload Identity Federation** is a broader Google Cloud feature that works across any Kubernetes cluster (EKS, AKS, on-prem, and GKE) and on the SaaS Gateway. It requires you to provide an `external_account` credential configuration JSON (generated via `gcloud iam workload-identity-pools create-cred-config`). The gateway exchanges a short-lived Kubernetes service account token for a Google Cloud access token through Google's Security Token Service.
</Accordion>

<Accordion title="When should I use Gemini vs Vertex AI? What's the difference?">
  **Gemini** is generally recommended for individual developers and prototyping use cases, while **Vertex AI** is recommended for production and enterprise use cases.

  Vertex AI offers everything available in the Gemini API and more, including:

  * More secure auth using service accounts instead of API keys
  * A Model Garden that includes multiple third-party models
  * Access to provisioned throughput

  You can read more about this here:

  * [Migrate to Cloud](https://ai.google.dev/gemini-api/docs/migrate-to-cloud)
  * [Migrate from Google AI](https://docs.cloud.google.com/vertex-ai/generative-ai/docs/migrate/migrate-google-ai#google-ai)
</Accordion>
