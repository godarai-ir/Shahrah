> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Quick Start Guide: Setup & Integration

> Quickly set up the TrueFoundry AI Gateway, connect model providers, and access LLMs with step-by-step instructions.

<Steps>
  <Step title="Sign Up">
    Go to [TrueFoundry](https://truefoundry.com) and click **Signup**.

    <Frame caption="Welcome Page">
      <img src="https://mintcdn.com/truefoundry/DKL5mQTvtv1ulV_-/images/2026-03-17_10.34.13.png?fit=max&auto=format&n=DKL5mQTvtv1ulV_-&q=85&s=a1da189be28ec4db2aa96f942aa00766" alt="Screenshot of the TrueFoundry Welcome Page with signup option" width="3024" height="1724" data-path="images/2026-03-17_10.34.13.png" />
    </Frame>

    Create an account and verify your email. You will be redirected to the landing page.
  </Step>

  <Step title="Add Your Models">
    Select the provider you want to add models from and add models after providing API Key. You can add multiple API keys from the same provider by creating separate provider accounts.

    <Frame caption="AI Gateway Page">
      <img src="https://mintcdn.com/truefoundry/iMid4yIOHvzf4Z4V/images/quick-start-4.jpg?fit=max&auto=format&n=iMid4yIOHvzf4Z4V&q=85&s=36b91f858a443ec4e2935efab71a4871" alt="AI Gateway Models Page showing provider selection" width="2520" height="1636" data-path="images/quick-start-4.jpg" />
    </Frame>

    ### Add Provider Account and Models

    A Provider Account represents one account of a model provider (e.g. OpenAI, Anthropic, AWS Bedrock).

    <Tip>
      You can add multiple accounts per provider, each with their own API keys. Each account can have multiple models.
    </Tip>

    <div style={{ overflow:"hidden", height:"450px" }}>
      <iframe provider="app.supademo.com" href="https://app.supademo.com/embed/cmmu4mpbc1sjaf3u9b48pg8ah?embed_v=2" typeofembed="iframe" height="500px" width="100%" src="https://app.supademo.com/embed/cmmu4mpbc1sjaf3u9b48pg8ah?embed_v=2" style={{ border:"none", display:"block", marginTop:"-25px" }} />
    </div>

    Once submitted, your provider accounts and models will appear under the **Models** tab.

    <Note>
      To add models later, go to **AI Gateway > Models**, select a provider account on the left, and click **Add Model** in the top right. From the models listing page, click **Try in Playground** on any model to start experimenting.

      <img src="https://mintcdn.com/truefoundry/l9NusOvGV2-nsX7f/images/model-addition.png?fit=max&auto=format&n=l9NusOvGV2-nsX7f&q=85&s=c0fa03193210a4e08c848237615de6a9" alt="Add Models to Provider Account interface in TrueFoundry" width="2678" height="654" data-path="images/model-addition.png" />
    </Note>
  </Step>

  <Step title="Try Models in the Playground">
    The LLM Playground lets you test models from all connected providers directly in the UI.

    <Frame caption="AI Gateway Playground Page">
      <img src="https://mintcdn.com/truefoundry/Pv3zXte3KtZeCZSK/images/pg-new.png?fit=max&auto=format&n=Pv3zXte3KtZeCZSK&q=85&s=da82dd093f6051992861615576901a66" alt="AI Gateway Playground Page for testing LLM models" width="2686" height="1720" data-path="images/pg-new.png" />
    </Frame>
  </Step>

  <Step title="Integrate with Code">
    To call Gateway models from your code, you need **Gateway Base URL**, **API Key**, and **Model ID**.

    All three are available in the **Code Snippets** tab of the Playground.

    <Frame caption="AI Gateway Code Snippets Page">
      <img src="https://mintcdn.com/truefoundry/Pv3zXte3KtZeCZSK/images/playground-new.png?fit=max&auto=format&n=Pv3zXte3KtZeCZSK&q=85&s=86c9cb57fd8294257617ee71f05e56e0" alt="Gateway Code Snippets Page with code generation options" width="2686" height="1720" data-path="images/playground-new.png" />
    </Frame>

    ### Gateway Base URL

    Your Gateway Base URL (`GATEWAY_BASE_URL`) is the endpoint for the AI Gateway service.

    * **TrueFoundry SaaS** users: `https://gateway.truefoundry.ai`
    * **Self-hosted** deployments: found in the Code Snippet section of the Playground.

    ### API Key

    Authenticate with the Gateway using one of:

    * **[Personal Access Token (PAT)](/docs/ai-gateway/authentication)** — recommended for developers during development
    * **[Virtual Account Token (VAT)](/docs/ai-gateway/authentication#virtual-account-token-vat)** — recommended for applications.

    Create a token from the **Access** section in the TrueFoundry platform.

    <Frame>
      <img src="https://mintcdn.com/truefoundry/zp-556n-et40aC3a/images/code-snippet-new-gateway.png?fit=max&auto=format&n=zp-556n-et40aC3a&q=85&s=a7bc77977c1ea074ad00f65100b7ab34" alt="TrueFoundry playground showing the Gateway Base URL and API key" width="2116" height="1716" data-path="images/code-snippet-new-gateway.png" />
    </Frame>

    ### Example Usage

    ```python theme={"dark"}
    from openai import OpenAI

    client = OpenAI(
        api_key="your_truefoundry_api_key",
        base_url="{GATEWAY_BASE_URL}"
    )
    ```

    Or set them as environment variables:

    ```bash theme={"dark"}
    export OPENAI_BASE_URL="{GATEWAY_BASE_URL}"
    export OPENAI_API_KEY="your_truefoundry_api_key"
    ```

    The Playground also provides ready-to-use snippets for `OpenAI SDK`, `Langchain`, `Python (Streaming/Non-Streaming)`, `Node.js`, `LangGraph`, `LlamaIndex`, `Google ADK`, `cURL`, and more.
  </Step>
</Steps>
