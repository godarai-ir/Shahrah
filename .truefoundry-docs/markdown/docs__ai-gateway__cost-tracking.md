> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Cost Tracking

> Set up and manage cost tracking for AI model usage with public and private pricing options.

TrueFoundry AI Gateway makes it easy to track and manage costs for all your AI model usage. This guide walks you through setting up cost tracking, viewing detailed breakdowns, and analyzing your spending.

## Setting Up Cost Tracking

To set up cost tracking for your AI models:

1. From the TrueFoundry dashboard, navigate to `AI Gateway` > `Models`
2. Select any Provider Account
3. Click on the `Edit` button for an existing model or `+ Add Model` to create a new one
4. In the model configuration screen, you'll find cost options

<Frame caption="Navigating to Model Cost Configuration Interface">
  <img src="https://mintcdn.com/truefoundry/jw406UAsc7ErYUq8/images/Screenshot2025-08-08at10.27.56AM-min.png?fit=max&auto=format&n=jw406UAsc7ErYUq8&q=85&s=96508f8223e0465fd8a7b6cb95821085" alt="Navigating to Model Cost Configuration Interface" width="3600" height="2008" data-path="images/Screenshot2025-08-08at10.27.56AM-min.png" />
</Frame>

TrueFoundry offers two simple ways to track costs:

* **Public Cost** — Uses provider-published rates from our open-source pricing catalog. Cost per token is auto-populated and updated; ideal for most popular models.
* **Private Cost** — Lets you set custom pricing for models without public rates, custom contracts, or fine-tuned models.

<Tabs>
  <Tab title="Public Cost – Automatic Pricing">
    <Frame caption="Setting up public cost">
      <img src="https://mintcdn.com/truefoundry/iMid4yIOHvzf4Z4V/images/public-cost.jpeg?fit=max&auto=format&n=iMid4yIOHvzf4Z4V&q=85&s=bcf3022dd1a536d5480a6ad9da06a137" alt="Interface showing how to enable public cost tracking for a model" width="1149" height="1280" data-path="images/public-cost.jpeg" />
    </Frame>

    Public cost uses pre-configured pricing based on the provider's published rates:

    * Automatically populated cost per token
    * Continuously updated using provider-published rates
    * Available for most popular models

    <Accordion title="How we get public pricing of models">
      TrueFoundry maintains an open-source pricing catalog in the [truefoundry/models GitHub repository](https://github.com/truefoundry/models). This repository acts as the pricing database used by AI Gateway for public cost tracking.

      The same pricing data is also viewable on the public [TrueFoundry Models dashboard](https://www.truefoundry.com/models), so you can inspect model pricing outside of your workspace as well.

      When AI Gateway calculates cost using public pricing, it considers:

      * **Region-wise pricing**: If a model has different rates by deployment region, the matching regional rate is used. For example, AWS Bedrock’s [Nova Lite](https://github.com/truefoundry/models/blob/main/providers/aws-bedrock/us.amazon.nova-lite-v1:0.yaml) uses different input/output costs per token in `us-west-2`, `us-west-1`, `eu-central-1`, and `eu-west-1`.
      * **Tiered pricing**: If a provider defines usage tiers (e.g. different rates at different volume thresholds), the applicable tier is selected. For example, [Gemini 2.5 Pro](https://github.com/truefoundry/models/blob/main/providers/google-gemini/gemini-2.5-pro.yaml) has base rates and higher rates from 200K tokens; [Gemini 1.5 Flash](https://github.com/truefoundry/models/blob/main/providers/google-gemini/gemini-1.5-flash.yaml) has tiers starting at 128K tokens.

      This helps ensure tracked costs align more closely with how providers bill in real-world scenarios.
    </Accordion>
  </Tab>

  <Tab title="Private Cost – Custom Pricing">
    <Frame caption="Setting up private cost">
      <img src="https://mintcdn.com/truefoundry/iMid4yIOHvzf4Z4V/images/private-cost.jpeg?fit=max&auto=format&n=iMid4yIOHvzf4Z4V&q=85&s=da195b7eff984a98bee55a060146b07d" alt="Interface showing how to configure custom pricing for a model" width="1157" height="1280" data-path="images/private-cost.jpeg" />
    </Frame>

    Private cost lets you set custom pricing for:

    * Models without public pricing
    * Custom pricing contracts
    * Fine-tuned models
  </Tab>
</Tabs>

## Viewing Your Costs

Once set up, you can easily view and analyze your costs in the Metrics section. Go to `AI Gateway` > `Metrics`.

<Frame caption="Metrics page overview">
  <img src="https://mintcdn.com/truefoundry/jw406UAsc7ErYUq8/images/Screenshot2025-08-05at12.17.53PM-min.png?fit=max&auto=format&n=jw406UAsc7ErYUq8&q=85&s=301c58dd282bbad7faa5d7d127d63571" alt="Overview of the metrics page showing total costs and usage trends" width="3600" height="2010" data-path="images/Screenshot2025-08-05at12.17.53PM-min.png" />
</Frame>

The Metrics page shows your total costs, usage trends, and provides interactive filters to analyze your spending patterns.

## Cost Breakdowns

View your costs from different perspectives with a single click:

<Columns cols={3}>
  <Card title="Cost by user">
    <img src="https://mintcdn.com/truefoundry/jw406UAsc7ErYUq8/images/Screenshot2025-08-05at12.16.23PM-min.png?fit=max&auto=format&n=jw406UAsc7ErYUq8&q=85&s=0d07362b6101e3aa074c3ee0dd1c9039" alt="Table showing cost breakdown by individual users" width="1660" height="758" data-path="images/Screenshot2025-08-05at12.16.23PM-min.png" />
  </Card>

  <Card title="Cost by model">
    <img src="https://mintcdn.com/truefoundry/jw406UAsc7ErYUq8/images/Screenshot2025-08-05at12.16.10PM-min.png?fit=max&auto=format&n=jw406UAsc7ErYUq8&q=85&s=5f79785c172932f14e5a73a61ffab2e4" alt="Table showing cost breakdown by individual models" width="1658" height="768" data-path="images/Screenshot2025-08-05at12.16.10PM-min.png" />
  </Card>

  <Card title="Cost by team">
    <img src="https://mintcdn.com/truefoundry/jw406UAsc7ErYUq8/images/Screenshot2025-08-05at12.16.57PM-min.png?fit=max&auto=format&n=jw406UAsc7ErYUq8&q=85&s=5f760e7f4edd2cbebe7897d5750c7886" alt="Table showing cost breakdown by team" width="1664" height="766" data-path="images/Screenshot2025-08-05at12.16.57PM-min.png" />
  </Card>
</Columns>

Each view helps you understand different aspects of your AI usage:

* **User view**: Identify high-usage individuals
* **Model view**: See which models cost the most
* **Team view**: Track department or project spending

## Exporting Cost Data

<Frame caption="Downloading cost data">
  <img src="https://mintcdn.com/truefoundry/jw406UAsc7ErYUq8/images/Screenshot2025-08-05at12.21.20PM-min.png?fit=max&auto=format&n=jw406UAsc7ErYUq8&q=85&s=14822ded00d5f275d3eece81d7674441" alt="Interface showing how to download raw cost and usage data" width="3600" height="2338" data-path="images/Screenshot2025-08-05at12.21.20PM-min.png" />
</Frame>

Need to analyze your data further? Easily export it:

1. Go to the `Metrics` section
2. Click on the `3 dots` button and then click on `Download Raw Data`
3. Choose the fields you want to `groupBy` the data

### Custom Grouping Options

You can customize how data is grouped in your exports. Simply select the fields you want to group by, such as username, model\_name, or teams, to get exactly the data organization you need for your analysis.

<Frame caption="Custom grouping options">
  <img src="https://mintcdn.com/truefoundry/jw406UAsc7ErYUq8/images/Screenshot2025-08-05at12.22.54PM-min.png?fit=max&auto=format&n=jw406UAsc7ErYUq8&q=85&s=ce627d8c9d76b6e1190887f4484aa972" alt="Screenshot showing groupBy options for username, model_name, and teams" width="3058" height="2078" data-path="images/Screenshot2025-08-05at12.22.54PM-min.png" />
</Frame>
