> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Azure Repos

> Follow this step to enable Azure Repos integrations

## Creating the Azure Repos App

This guide walks you through the process of enabling Azure Repo integration with TrueFoundry by creating an app registration in Microsoft Entra ID (formerly Azure AD) and providing the necessary client environment variables to TrueFoundry. This allows the TrueFoundry Control Plane to authenticate and interact with Azure Repos on behalf of the registered app.

## Prerequisites

Ensure you have:

* Access to the Azure portal with the necessary permissions to create an app registration.
* Administrator or sufficient privileges in your Azure DevOps organization.
* The ability to configure environment variables in TrueFoundry.
* The control plane URL of your TrueFoundry deployment.

## Step 1: Create an App Registration in Microsoft Entra ID

1. **Log in to the Azure Portal**:
   * Go to [Azure Portal](https://portal.azure.com) and sign in with your Microsoft account.
2. **Navigate to Microsoft Entra ID (Azure Active Directory)**:
   * From the left-hand navigation pane, select **Microsoft Entra ID** (formerly Azure Active Directory).
3. **Register a New Application**:
   * Under **Manage**, select **App registrations** and click on **New registration**.
   * **Name**: Provide a name for the application (e.g., `TrueFoundry Integration`).
   * **Supported account types**: Choose **Accounts in this organizational directory only** (Single tenant) if your Azure DevOps organisation is linked to the current tenant. Otherwise, choose multi-tenant option.
   * **Redirect URI**: Under **Redirect URI (optional)**, select **Web** from the dropdown and enter the following URL:

     ```
     <control-plane-url>/api/svc/v1/vcs/azure/callback
     ```

     Replace `<control-plane-url>` with the control plane URL of your TrueFoundry deployment. It will be something like `https://org.truefoundry.cloud` and is typically used to access the TrueFoundry Dashboard.
   * Click **Register** to create the app.
4. **Configure API Permissions**:

   * After registration, you'll be redirected to the app’s overview page.
   * Under **Manage**, select **API permissions**.
   * Click on **Add a permission**.

     **a. Azure DevOps**:

     * Select **APIs my organization uses**.
     * Search for **Azure DevOps** and select it.
     * Choose **Delegated permissions**.
     * Select the permission `vso.code_manage` to allow management of code repositories.
     * Click **Add permissions**.

     **b. Microsoft Graph**:

     * Click **Add a permission** again.
     * Select **Microsoft Graph**.
     * Choose **Delegated permissions**.
     * Select the permissions `offline_access` and `User.Read`.
     * Click **Add permissions**.
   * Click on **Grant admin consent** for your organization to ensure the permissions are active and granted.

   <Frame caption="">
     <img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/eec36a26-2a837db37d5ca9cc858182dafca1488b9dfd54ea202790ec358ea85f88ee6a78-permissions.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=e46a5906c061fccbaee61beb038aff2e" alt="" width="2212" height="758" data-path="images/eec36a26-2a837db37d5ca9cc858182dafca1488b9dfd54ea202790ec358ea85f88ee6a78-permissions.png" />
   </Frame>
5. **Generate a Client Secret**:
   * Under **Manage**, select **Certificates & secrets**.
   * Under **Client secrets**, click **New client secret**.
   * Provide a description (e.g., `TrueFoundry Secret`) and choose an expiry duration.
   * Click **Add**.
   * Copy the **Value** of the client secret immediately. You won't be able to retrieve it later.
6. **Gather Required Information**:
   * Navigate back to the **Overview** section of the app registration.
   * Note down the **Application (client) ID**.

## Step 2: Provide Environment Variables to TrueFoundry

In TrueFoundry, you need to configure the following environment variables using the details from the Microsoft Entra ID app registration. Edit the Helm deployment of `truefoundry` and add the following variables in the `env` section under `servicefoundry-server`

1. **AZURE\_CLIENT\_ID**: The Application (client) ID of your app registration.
2. **AZURE\_CLIENT\_SECRET**: The Client Secret generated in the previous step.

<CodeGroup>
  ```yaml YAML lines theme={"dark"}
  servicefoundry-server:
    # other fields
    env:
      # other envs
      AZURE_CLIENT_ID: <client id of azure app registration>
      AZURE_CLIENT_SECRET: <client secret>
  ```
</CodeGroup>

Now you can deploy applications from your Azure Repos to TrueFoundry by linking repositories. You can link repositories by going to `Integrations` > `Git` > `Azure Repos` on the TrueFoundry Dashboard.

***
