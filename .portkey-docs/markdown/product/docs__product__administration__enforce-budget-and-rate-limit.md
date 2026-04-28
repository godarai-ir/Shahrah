> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Enforce Budget Limits and Rate Limits for Your API Keys

> Configure budget and rate limits on API keys to effectively manage AI spending and usage across your organization

<Info>
  Available on **Enterprise** plan and select **Pro** customers.
</Info>

<Tip>
  Looking to set limits for an entire workspace instead of individual API keys? See [Enforce Workspace Budget and Rate Limits](/product/administration/enforce-workspace-budget-limts-and-rate-limits).
</Tip>

## Overview

For enterprises deploying AI at scale, maintaining financial oversight and operational control is crucial. Portkey's governance features for API keys provide finance teams, IT departments, and executives with the transparency and guardrails needed to confidently scale AI adoption across the organization.

By implementing budget and rate limits on API keys at both organization and workspace levels, you can:

* Prevent unexpected cost overruns through automated spending caps
* Maintain performance and availability through usage rate controls
* Receive timely notifications when thresholds are approached
* Enforce consistent governance policies across teams and departments

These capabilities ensure your organization can innovate with AI while maintaining predictable costs and usage patterns.

## Budget Limits

Budget limits allow you to set maximum LLM spending or token usage thresholds on your API keys, automatically preventing further usage when limits are reached.

When creating or editing an API key, you can establish spending parameters that align with your financial planning:

### Setting Up Budget Limits

When creating a new API key or editing an existing one:

1. Toggle on **Add Budget Limit**
2. Choose between two limit types:
   * **Cost**: Set a maximum spend in USD (minimum \$1)
   * **Tokens**: Set a maximum token usage

<Frame caption="Budget limit configuration">
  <img src="https://mintcdn.com/portkey-docs/Buc1Vm2P31GSPm3S/images/product/budget-limits-options.png?fit=max&auto=format&n=Buc1Vm2P31GSPm3S&q=85&s=67237265eb18ead901be8717e2d61f8d" width="873" height="452" data-path="images/product/budget-limits-options.png" />
</Frame>

### Alert Thresholds

You can configure alert thresholds to receive notifications before reaching your full budget:

1. Enter a value in the **Alert Threshold** field
2. When usage reaches this threshold, notifications will be sent to configured recipients
3. The API key continues to function until the full budget limit is reached

### Periodic Reset Options

Budget limits can be set to automatically reset at regular intervals:

<Frame caption="Periodic reset options">
  <img src="https://mintcdn.com/portkey-docs/Buc1Vm2P31GSPm3S/images/product/periodic-reset.png?fit=max&auto=format&n=Buc1Vm2P31GSPm3S&q=85&s=ed4ba4826e5b77301d05979a861fda34" width="870" height="268" data-path="images/product/periodic-reset.png" />
</Frame>

* **No Periodic Reset**: The budget limit applies until exhausted
* **Reset Weekly**: Budget limits reset every Sunday at 12 AM UTC
* **Reset Monthly**: Budget limits reset on the 1st of each month at 12 AM UTC

## Rate Limits

Rate limits control how frequently an API key can be used, helping you maintain application performance and prevent unexpected usage spikes.

### Setting Up Rate Limits

When creating a new API key or editing an existing one:

1. Toggle on **Add Rate Limit**
2. Choose your limit type:
   * **Requests**: Limit based on number of API calls
   * **Tokens**: Limit based on token consumption
3. Specify the limit value and time interval

<Frame caption="Rate limit configuration">
  <img src="https://mintcdn.com/portkey-docs/_Cb_bj7tVjxcfwsu/images/product/rate-limit.png?fit=max&auto=format&n=_Cb_bj7tVjxcfwsu&q=85&s=0e0ce0c4811a056ce4bc79fac708ea9e" width="873" height="323" data-path="images/product/rate-limit.png" />
</Frame>

### Time Intervals

Rate limits can be applied using three different time intervals:

* **Per Minute**: For granular control of high-frequency applications
* **Per Hour**: For balanced control of moderate usage
* **Per Day**: For broader usage management

When a rate limit is reached, subsequent requests are rejected until the time interval resets.

## Email Notifications

Email notifications keep relevant stakeholders informed about API key usage and when limits are approached or reached.

### Configuring Notifications

To set up email notifications for an API key with budget limits:

1. Toggle on **Email Notifications** when creating/editing an API key
2. Add recipient email addresses:
   * Type an email address and click **New** or press Enter
   * Add multiple recipients as needed

<Frame caption="Email notification setup">
  <img src="https://mintcdn.com/portkey-docs/Buc1Vm2P31GSPm3S/images/product/email.png?fit=max&auto=format&n=Buc1Vm2P31GSPm3S&q=85&s=1ba2522939e6bfeef6cd4e6df352edf2" width="1077" height="176" data-path="images/product/email.png" />
</Frame>

### Default Recipients

When limits are reached or thresholds are crossed, Portkey automatically sends notifications to:

* Organization administrators
* Organization owners
* The API key creator/owner

You can add additional recipients such as finance team members, department heads, or project managers who need visibility into AI usage.

## Availability

These features are available to Portkey Enterprise customers and select Pro users. To enable these features for your account, please contact [support@portkey.ai](mailto:support@portkey.ai) or join the [Portkey Discord](https://portkey.ai/community) community.

To learn more about the Portkey Enterprise plan, [schedule a consultation](https://portkey.sh/demo-16).
