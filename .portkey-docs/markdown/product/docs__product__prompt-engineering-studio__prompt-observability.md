> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Prompt Observability

Portkey's Prompt Observability provides comprehensive insights into how your prompts are performing in production. This feature allows you to track usage, monitor performance metrics, and analyze trends to continuously improve your prompts based on real-world usage.

<Info>
  This feature is available on all Portkey [plans](https://portkey.ai/pricing).
</Info>

## Overview

Prompt Observability gives you visibility into your prompt usage and performance through analytics dashboards, detailed logs, and template history. By monitoring these metrics, you can identify which prompts are performing well and which need optimization, helping you make data-driven decisions about your AI applications.

## Accessing Prompt Observability

You can access observability data in several ways:

1. **From the Prompt Template page**: View history and performance metrics for a specific prompt
2. **From the Analytics dashboard**: Filter analytics by prompt ID
3. **From the Logs section**: Filter logs by prompt ID to see detailed usage information

## Prompt Analytics

The Analytics dashboard provides high-level metrics for your prompts, showing important information like costs, token usage, latency, request volume, and user engagement. You can easily filter your prompts using `prompt-id` in the analytics dashboard.

<Frame>
  <img src="https://mintcdn.com/portkey-docs/Buc1Vm2P31GSPm3S/images/product/dashboard.png?fit=max&auto=format&n=Buc1Vm2P31GSPm3S&q=85&s=6c880d5ae17f22884436ad3c7b3347d9" width="600" height="348" data-path="images/product/dashboard.png" />
</Frame>

The dashboard enables you to understand trends in your prompt usage over time and identify potential opportunities for optimization. For more details on using the analytics dashboard and available filters, refer to Portkey's [Analytics documentation](/product/observability/analytics).

## Prompt Logs

The Logs section on Portkey's dashboard provides detailed information about each individual prompt call, giving you visibility into exactly how your prompts are being used in real-time. You can easily filter your prompts using `prompt-id` in the logs view.

<Frame>
  <img src="https://mintcdn.com/portkey-docs/_Cb_bj7tVjxcfwsu/images/product/product-2.avif?fit=max&auto=format&n=_Cb_bj7tVjxcfwsu&q=85&s=831f1cb98919f6a6d9e2a467db031653" width="800" height="492" data-path="images/product/product-2.avif" />
</Frame>

Each log entry shows the timestamp, model used, request path, user, tokens consumed, cost, and status. This granular data helps you understand exactly how your prompts are performing in production and identify any issues that need attention.

For information on filtering and searching logs, refer to Portkey's [Logs documentation](/product/observability/logs).

<Note>
  **Render Calls**: Note that `prompts.render` API calls are not logged in the observability features. Only `prompts.completions` calls are tracked.
</Note>

## Prompt Template History

Each prompt template includes a "Recent" tab that shows the history of calls made using that specific template:

<Frame>
  <img src="https://mintcdn.com/portkey-docs/_Cb_bj7tVjxcfwsu/images/product/product-2.avif?fit=max&auto=format&n=_Cb_bj7tVjxcfwsu&q=85&s=831f1cb98919f6a6d9e2a467db031653" width="800" height="492" data-path="images/product/product-2.avif" />
</Frame>

This chronological view makes it easy to see how your template is being used and how it's performing over time. You can quickly access detailed information about each call directly from this history view.

The template history is particularly useful when you're iterating on a prompt design, as it allows you to see the immediate impact of your changes. Combined with [Prompt Versioning](/product/prompt-engineering-studio/prompt-versioning), this gives you a complete view of your prompt's evolution and performance.

## Next Steps

Now that you understand how to monitor your prompts, explore these related features:

* [Prompt Versioning](/product/prompt-engineering-studio/prompt-versioning) - Track changes to your prompts over time
* [Prompt API](/product/prompt-engineering-studio/prompt-api) - Integrate optimized prompts into your applications
* [Prompt Playground](/product/prompt-engineering-studio/prompt-playground) - Test and refine your prompts based on observability insights
* [Prompt Partials](/product/prompt-engineering-studio/prompt-partial) - Create reusable components for your prompts
* [Tool Library](/product/prompt-engineering-studio/tool-library) - Enhance your prompts with specialized tools
