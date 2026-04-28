> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Aporia

[Aporia](https://www.aporia.com/) provides state-of-the-art Guardrails for any AI workload. With Aporia, you can setup powerful, multimodal AI Guardrails and just add your Project ID to Portkey to enable them for your Portkey requests.

Aporia supports Guardrails for `Prompt Injections`, `Prompt Leakage`, `SQL Enforcement`, `Data Leakage`, `PII Leakage`, `Profanity Detection`, and many more!

Browse Aporia's docs for more info on each of the Guardrail:

<Card title="Quickstart - AporiaAporia" href="https://gr-docs.aporia.com/get-started/quickstart" icon={<img src="/images/guardrails/apple-touch-icon.png"  />}>
  Learn more about Patronus AI and their offerings.
</Card>

## Using Aporia with Portkey

### 1. Add Aporia API Key to Portkey

* Navigate to the `Integrations` page under `Settings`
* Click on the edit button for the Aporia integration and add your API key

<Frame>
  <img src="https://mintcdn.com/portkey-docs/QFjngWBmb6CT7QXC/images/guardrails/ap-1.png?fit=max&auto=format&n=QFjngWBmb6CT7QXC&q=85&s=7d6d1bd0e12b9fbfbb6920f84fa6cbf0" width="1812" height="422" data-path="images/guardrails/ap-1.png" />
</Frame>

### 2. Add Aporia's Guardrail Check

* Now, navigate to the `Guardrails` page
* Search for `Validate - Project` Guardrail Check and click on `Add`
* Input your corresponding Aporia Project ID where you are defining the policies
* Save the check, set any actions you want on the check, and create the Guardrail!

| Check Name          | Description                                                                         | Parameters         | Supported Hooks                      |
| :------------------ | :---------------------------------------------------------------------------------- | :----------------- | :----------------------------------- |
| Validate - Projects | Runs a project containing policies set in Aporia and returns a PASS or FAIL verdict | Project ID: string | beforeRequestHooks afterRequestHooks |

<Frame>
  <img src="https://mintcdn.com/portkey-docs/IbI4RvWwDz6X1dr5/images/guardrails/ap-2.png?fit=max&auto=format&n=IbI4RvWwDz6X1dr5&q=85&s=94f5dceffe21ceec631c5df8795839f3" width="1212" height="558" data-path="images/guardrails/ap-2.png" />
</Frame>

Your Aporia Guardrail is now ready to be added to any Portkey request you'd like!

### 3. Add Guardrail ID to a Config and Make Your Request

* When you save a Guardrail, you'll get an associated Guardrail ID - add this ID to the `before_request_hooks` or `after_request_hooks` params in your Portkey Config
* Save this Config and pass it along with any Portkey request you're making!

Your requests are now guarded by your Aporia policies and you can see the Verdict and any action you take directly on Portkey logs! More detailed logs for your requests will also be available on your Aporia dashboard.

***

## Get Support

If you face any issues with the Aporia integration, just ping the @Aporia team on the [community forum](https://discord.gg/portkey-llms-in-prod-1143393887742861333).
