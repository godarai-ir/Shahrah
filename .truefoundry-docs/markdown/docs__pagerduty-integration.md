> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# PagerDuty

> Set up PagerDuty as a notification channel for TrueFoundry alerts and incident management.

To receive alerts and notifications in TrueFoundry, you need to set up notification integrations for PagerDuty. This guide walks you through adding PagerDuty notification integration, describing each field in the forms, and providing helpful links for setup.

## PagerDuty Notification Integration

To add a PagerDuty notification integration:

1. **Go to the Integrations page** in the TrueFoundry UI.
2. **Click on 'Add Integration Provider'.**
3. **Select 'PagerDuty' as the provider type.**
4. **Choose 'PagerDuty Integration'** in the integrations section.
5. **Fill out the form fields as shown below:**

<Frame>
  <img src="https://mintcdn.com/truefoundry/iMid4yIOHvzf4Z4V/images/pagerduty-integration.png?fit=max&auto=format&n=iMid4yIOHvzf4Z4V&q=85&s=670f206fe87c1a3247a8c5fa3e90d9e9" alt="PagerDuty Integration Form" width="1678" height="1400" data-path="images/pagerduty-integration.png" />
</Frame>

### Field Descriptions

* **Name**: A unique name for this integration provider (e.g., "dev-pagerduty").
* **Display Name**: The display name for this PagerDuty integration (e.g., "pagerduty-prometheus-channel").
* **Integration Key**: The integration key from your PagerDuty service (required for sending alerts).
* **Integration URL**: The PagerDuty Events API URL (example: `https://events.pagerduty.com/generic/2010-04-15/create_event.json`).
* **Access Control (Optional)**: List of users or groups allowed to use this integration. By default, set to `everyone`.

> **How to get PagerDuty Integration Key:**
>
> 1. Go to your PagerDuty Services page.
> 2. Select the service you want to integrate with Prometheus.
> 3. Under "Integrations", click "Add Integration" and select "Prometheus" from the list and fill other details.
>
> <Warning>
>   Make sure you select prometheus from the list of integrations, else the
>   alerts will not be sent to PagerDuty.
> </Warning>
>
> 4. Copy the generated Integration Key and Integration URL and use it in the form above.
> 5. For more details, see the [PagerDuty Prometheus Integration Guide](https://www.pagerduty.com/docs/guides/prometheus-integration-guide/).

## Access Control

For all integrations, you can restrict access to specific users or groups by editing the **Access Control** field. By default, it is set to `everyone` so all users can use the integration.

## Final Steps

* After filling all required fields, click **Add Integration Provider** to save your integration.
* You can now use these integrations as notification channels when setting up alerts for your applications.

***

With these integrations in place, you can receive real-time alerts via PagerDuty for all your TrueFoundry applications!
