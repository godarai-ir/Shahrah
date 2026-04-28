> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# ToolJet

> Add AI capabilities to ToolJet apps with Portkey

ToolJet is a low-code platform for building apps. Integrate Portkey to add AI-powered chat, completions, and embeddings to your ToolJet apps.

## Prerequisites

* **Portkey API Key** from [app.portkey.ai/api-keys](https://app.portkey.ai/api-keys)
* **Provider slug** from [Model Catalog](https://app.portkey.ai/model-catalog)
* **ToolJet Account** with Marketplace Plugin access

<Card title="Watch Demo" icon="play" href="https://www.youtube.com/watch?v=VGb7HGRkam0">
  Quick walkthrough of ToolJet's UI components
</Card>

## Setup

<Steps>
  <Step title="Install Portkey Plugin">
    Go to **ToolJet Dashboard → Plugins → Marketplace** → Search **Portkey** → **Install**.
  </Step>

  <Step title="Add as Data Source">
    1. Go to **Data Sources → Plugins → Add Portkey**
    2. Enter:
       * **Authorization**: Your Portkey API Key
       * **Default Provider**: Your provider slug (e.g., `@openai-prod`)
    3. **Test** the connection

    <Frame>
      <img src="https://mintcdn.com/portkey-docs/_Cb_bj7tVjxcfwsu/images/supported-llm/2.Portkey%20datasource.png?fit=max&auto=format&n=_Cb_bj7tVjxcfwsu&q=85&s=d07c5537f38428c1498d8d7daca4af8e" width="1448" height="1488" data-path="images/supported-llm/2.Portkey datasource.png" />
    </Frame>
  </Step>

  <Step title="Use in Your App">
    1. Go to **Queries → Add Datasource → Select Portkey**
    2. Choose an operation (Chat, Completion, Embedding, Prompt)
    3. **Run** to verify

    <Frame>
      <img src="https://mintcdn.com/portkey-docs/_Cb_bj7tVjxcfwsu/images/supported-llm/3.Use%20in%20ToolJet.png?fit=max&auto=format&n=_Cb_bj7tVjxcfwsu&q=85&s=be8e00ff727d40159e9c62b665329a8f" width="2934" height="1674" data-path="images/supported-llm/3.Use in ToolJet.png" />
    </Frame>
  </Step>

  <Step title="Connect to UI">
    1. Add **Text Input** and **Button** widgets
    2. Configure Button's **onClick** to execute the Portkey query
    3. Display results in a **Text Box**
  </Step>
</Steps>

## Supported Operations

<AccordionGroup>
  <Accordion title="Chat">
    Generate chat completions from messages.

    **Parameters:** Messages, Model, Max Tokens, Temperature, Stop Sequences, Metadata

    <Frame>
      <img src="https://mintcdn.com/portkey-docs/5HryU7OAng7DdvDd/images/supported-llm/5.chat.png?fit=max&auto=format&n=5HryU7OAng7DdvDd&q=85&s=ace66568d71409a64be44ea4d34d4176" width="1634" height="1248" data-path="images/supported-llm/5.chat.png" />
    </Frame>
  </Accordion>

  <Accordion title="Completion">
    Generate text completions from a prompt.

    **Parameters:** Prompt, Model, Max Tokens, Temperature, Stop Sequences, Metadata

    <Frame>
      <img src="https://mintcdn.com/portkey-docs/_Cb_bj7tVjxcfwsu/images/supported-llm/4.completion.png?fit=max&auto=format&n=_Cb_bj7tVjxcfwsu&q=85&s=b94f0ba2d4b53e83f2fd194ea44f181c" width="1646" height="1252" data-path="images/supported-llm/4.completion.png" />
    </Frame>
  </Accordion>

  <Accordion title="Prompt Completion">
    Use pre-defined prompts from Portkey.

    **Parameters:** Prompt ID, Variables, Parameters, Metadata

    <Frame>
      <img src="https://mintcdn.com/portkey-docs/5HryU7OAng7DdvDd/images/supported-llm/6.prompt-completion.png?fit=max&auto=format&n=5HryU7OAng7DdvDd&q=85&s=0fb2208da628ec4653806c9c33575a8f" width="1660" height="1230" data-path="images/supported-llm/6.prompt-completion.png" />
    </Frame>
  </Accordion>

  <Accordion title="Create Embedding">
    Generate embeddings for text.

    **Parameters:** Input, Model, Metadata

    <Frame>
      <img src="https://mintcdn.com/portkey-docs/5HryU7OAng7DdvDd/images/supported-llm/7.embedding.png?fit=max&auto=format&n=5HryU7OAng7DdvDd&q=85&s=700b88984982ba33e77e2c7aba42df77" width="1654" height="838" data-path="images/supported-llm/7.embedding.png" />
    </Frame>
  </Accordion>
</AccordionGroup>

<Info>
  For all operations, you can optionally specify a **Config** or override the default **Provider**.
</Info>

## Troubleshooting

| Issue                | Solution                                   |
| -------------------- | ------------------------------------------ |
| Authentication Error | Check API key and provider slug            |
| Slow Response        | Adjust temperature or max\_tokens          |
| CORS Issues          | Ensure API settings allow ToolJet's domain |

## Next Steps

<CardGroup cols={2}>
  <Card title="Model Catalog" icon="list" href="/product/model-catalog">
    Set up providers
  </Card>

  <Card title="Configs" icon="gear" href="/product/ai-gateway/configs">
    Add routing and caching
  </Card>
</CardGroup>
