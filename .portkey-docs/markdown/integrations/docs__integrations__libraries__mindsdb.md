> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# MindsDB

> Integrate MindsDB with Portkey for production-grade AI pipelines

MindsDB connects data sources and LLMs for AI automation. Portkey adds:

* **1600+ LLMs** — Access any provider through MindsDB
* **Observability** — Track costs, latency, and usage
* **Reliability** — Caching, routing, guardrails via [Configs](/product/ai-gateway/configs)

## Prerequisites

1. Install MindsDB via [Docker](https://docs.mindsdb.com/setup/self-hosted/docker) or [Docker Desktop](https://docs.mindsdb.com/setup/self-hosted/docker-desktop)
2. Install Portkey dependencies per [MindsDB instructions](https://docs.mindsdb.com/setup/self-hosted/docker#install-dependencies)
3. Get your [Portkey API key](https://app.portkey.ai/api-keys)

## Setup

<Steps>
  <Step title="Create AI Engine">
    <CodeGroup>
      ```sql Using Config ID theme={"system"}
      CREATE ML_ENGINE portkey_engine
      FROM portkey
      USING
          portkey_api_key = 'PORTKEY_API_KEY',
          config = 'pc-your-config-id';
      ```

      ```sql Using Provider Slug theme={"system"}
      CREATE ML_ENGINE portkey_engine
      FROM portkey
      USING
          portkey_api_key = 'PORTKEY_API_KEY',
          provider = '@openai-prod';
      ```
    </CodeGroup>

    <Note>
      Use [Configs](/product/ai-gateway/configs) to add guardrails, caching, conditional routing, and more.
    </Note>
  </Step>

  <Step title="Create Model">
    ```sql theme={"system"}
    CREATE MODEL portkey_model
    PREDICT answer
    USING
        engine = 'portkey_engine',
        temperature = 0.2;
    ```
  </Step>

  <Step title="Query">
    ```sql theme={"system"}
    SELECT question, answer
    FROM portkey_model
    WHERE question = 'Where is Stockholm located?';
    ```

    Output:

    ```
    +-----------------------------+--------------------------------------------------------------------------------+
    | question                    | answer                                                                         |
    +-----------------------------+--------------------------------------------------------------------------------+
    | Where is Stockholm located? | Stockholm is the capital of Sweden, located on the south-central east coast... |
    +-----------------------------+--------------------------------------------------------------------------------+
    ```
  </Step>
</Steps>

## Next Steps

<CardGroup cols={2}>
  <Card title="MindsDB Use Cases" icon="book" href="https://docs.mindsdb.com/use-cases/overview">
    Explore MindsDB examples
  </Card>

  <Card title="Portkey Configs" icon="gear" href="/product/ai-gateway/configs">
    Add routing, caching, guardrails
  </Card>
</CardGroup>
