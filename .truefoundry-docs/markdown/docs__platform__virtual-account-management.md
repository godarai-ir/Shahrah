> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Manage Virtual Accounts

> Detailed guide to how to add, manage and delete Virtual Accounts in Truefoundry

Virtual accounts are non-user accounts that can be assumed by applications or services to access resources in Truefoundry. They can be created by admins and each virtual account will have a token using which the application can access the resources.

<Tip>
  It is usually recommended to create one virtual account per application and scope it to the minimum set of permissions required by the application.
</Tip>

## Create Virtual Accounts

You can create and assign permissions to a virtual account as shown below:

<iframe typeofembed="iframe" href="https://app.supademo.com/embed/cm7m23pd504np1x0i0xny4ido?embed_v=2" height="450px" width="100%" src="https://app.supademo.com/embed/cm7m23pd504np1x0i0xny4ido?embed_v=2" style={{ border: "none", display: "flex", margin: "auto" }} />

Once you create a virtual account, you can get the token for it by clicking on the `Get Token` button.

<img src="https://mintcdn.com/truefoundry/hN5wOa3gNueWZ6WB/images/docs/platform/get-token-virtual-account.png?fit=max&auto=format&n=hN5wOa3gNueWZ6WB&q=85&s=48a4a265d23788c0696a26feae45cc0f" alt="Get Token button" width="3024" height="1644" data-path="images/docs/platform/get-token-virtual-account.png" />

Using a virtual account, you can autorotate the tokens, set notifications on rotation and also sync the token tox a secretmanager of your choice.

### Configure auto-rotation of virtual account tokens

You can configure auto-rotate to automatically rotate the token at an interval. In auto-rotation, new token will be generated and can be retrieved using UI or API while the older token will be active for a grace period that is configurable.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/9lYm7n0BF3i5yINM/images/platform-virtual-account-mng-auto-rotate.png?fit=max&auto=format&n=9lYm7n0BF3i5yINM&q=85&s=b7648acab340030272880611a6aa7d80" width="1744" height="1148" data-path="images/platform-virtual-account-mng-auto-rotate.png" />
</Frame>

### Configure notification on token rotation

You can get notified via email or Slack when a token is rotated. To configure notification, you first need to add an integration for email or Slack. You can find the instructions [here](/docs/email-integration) and [here](/docs/slack-bot-integration).

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/9lYm7n0BF3i5yINM/images/platform-virtual-account-mng-notifications.png?fit=max&auto=format&n=9lYm7n0BF3i5yINM&q=85&s=9d54c3c834cb84712a2decf5db11312a" width="1752" height="1124" data-path="images/platform-virtual-account-mng-notifications.png" />
</Frame>

### Configure secret store sync for virtual account tokens

You can configure Truefoundry to automatically sync the virtual account token to a secret store of your choice. To use this feature, you first need to integrate a secret store with Truefoundry. Truefoundry supports integrations with AWS Parameter Store, AWS Secrets Manager, Google Secret Manager, HashiCorp Vault, Azure Vault, etc. You can find the instructions [here](/docs/integrations-secret-store).

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/9lYm7n0BF3i5yINM/images/platform-virtual-account-mng-secret-store-sync.png?fit=max&auto=format&n=9lYm7n0BF3i5yINM&q=85&s=e7d96669245ac1aabdfcb125f829ce18" width="888" height="384" data-path="images/platform-virtual-account-mng-secret-store-sync.png" />
</Frame>

If you configure the secrets to sync to a secret store, the virtual account token will be stored in the secret store in a secret path of your choice. The token will automatically be synced to the secret store when the token is rotated.
