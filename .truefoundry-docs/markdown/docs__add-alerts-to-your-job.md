> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Add Alerts To Your Job

> Step-by-step guide to adding alerts for jobs to monitor failures and performance.

## Pre-Requisites

Setup a `Notification Channel Integration` on the Integrations page. Please follow this [document](/docs/add-alerts-to-your-job) to add an integration.

Currently two types of Notification Channels are supported:

1. Slack (Webhook URL based): For this, please create a [webhook](https://api.slack.com/messaging/webhooks) for your slack and then add it as an integration in **Slack Provider Account**
2. Slack (Bot Token based): For this, please create a [bot token](https://api.slack.com/quickstart) for your slack and then add it as an integration in **Slack Provider Account** (Slack Bot Integration). *It requires chat:write and chat:write:public scope.* You can add slack channels to send to respective slack channel.
3. Email (SMTP credentials Integration): This can be found in **Custom Provider Account.** You can configure the SMTP Credentials of the mail server, from\_email and to\_emails and use it to send notifications.

## How to Configure

1. In the Job form, please click on "Show Advanced Fields" and find the "Alerts" section

   <Frame caption="">
     <img src="https://mintcdn.com/truefoundry/OHzlp6GY5G-JfKle/images/c3671944-d6665850748e75924d3433f02ddaf2779721c06229c5824cf0eeda15cae36dbc-Screenshot_2025-04-14_at_4.01.16_PM.png?fit=max&auto=format&n=OHzlp6GY5G-JfKle&q=85&s=d8f53db49e71ad5a1fccae2e0968a276" width="3014" height="1698" data-path="images/c3671944-d6665850748e75924d3433f02ddaf2779721c06229c5824cf0eeda15cae36dbc-Screenshot_2025-04-14_at_4.01.16_PM.png" />
   </Frame>

2. Select your notification target and the triggers (on\_start, on\_completion, on\_failure).

3. There are three types of targets, one is email and the others are slack webhook and slack bot, you need to add an email/slack webhook/slack bot integration respectively and select it in notification channel field to use that integration for sending notifications.

4. Click on Submit.

Once the Alerts are created you can see the alerts in your Slack/Email. Here is a sample Alert:

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/FrY4JbiyZud2He3p/images/01ebfd96-864fdecc7c286610b7deb610a6a4afaccfcda690236ac61f6e606b46492030a1-image.png?fit=max&auto=format&n=FrY4JbiyZud2He3p&q=85&s=af96a4812ebc1f6418dcb4f4d415eaa9" width="1412" height="380" data-path="images/01ebfd96-864fdecc7c286610b7deb610a6a4afaccfcda690236ac61f6e606b46492030a1-image.png" />
</Frame>

***
