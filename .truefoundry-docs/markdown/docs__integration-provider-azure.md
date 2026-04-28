> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Azure

> Connect Azure services to TrueFoundry with secure configurations.

TrueFoundry supports integrating with multiple Azure services like ABS, ACR, Azure Vault, AKS, Azure OpenAI Models, Azure Repos etc.

To integrate any of the above services, you simply need to add your Azure account as a provider account and add integrations for the same as shown below:

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/clya859c90ddhthf78u580x6e" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/clya859c90ddhthf78u580x6e" style={{ border:"none",display:"flex",margin:"auto" }} />

**Share access with users, teams or everyone in your TrueFoundry account**

As shown in the previous slides, you can share access of each integration with users, teams or everyone in your TrueFoundry account. This would allow them to view and use the integration. Only tenant-admins can edit the integrations.

### Azure Blob Storage(ABS) Integration

Follow the steps below to connect your Azure blob storage to TrueFoundry:

1. Create a [Azure Storage account](https://learn.microsoft.com/en-us/azure/storage/common/storage-account-create?tabs=azure-portal) in your resource group
   1. **Instance details** - You must `Geo-redundant storage` to make sure your data is available through other regions in case of region unavailability.
   2. **Security** - Make sure
      1. DISABLE `Allow enabling anonymous access on individual containers`
      2. ENABLE `Enable storage account key access`
   3. **Network access** - ENABLE `Allow public access from all networks`
   4. **Recovery** - You can keep it to default for 7 days.
2. Create an [Azure container](https://learn.microsoft.com/en-us/azure/storage/blobs/blob-containers-portal) inside the above storage account.
3. Search for `CORS` from the left panel and for `Blob service` (optional for `File service` `Queue service` and `Table Service`, only apply the change if you are using them) select the below options
   1. Allowed Origins - `*` or your control plane URL
   2. Allowed Methods - `GET, POST, PUT`
   3. Allowed Headers - `*`
   4. Exposed Headers - `Etag`
   5. MaxAgeSeconds - `3600`
4. Collect the following information
   1. [Standard endpoint](https://learn.microsoft.com/en-us/azure/storage/common/storage-account-overview#standard-endpoints) - Endpoint of the blob storage Once the container is created we need to get the standard endpoint of the blob storage along with the container which will look something like this. Replace this with your storage account name and the container name.

      ```javascript lines theme={"dark"}
      https://*mystorageaccount*.blob.core.windows.net/*mycontainer*/
      ```
   2. Connection string - From the Azure portal in your storage account, head over to the `Security + Networking` section under `Access keys` which will contain the `Connection String` .
5. Head over to the Integrations tab from the sidebar and follow the steps shown in the previous demo to complete the integration.

### Azure Container Registry(ACR) Integration

To add Azure Container registry, follow the steps below:

1. Create an [Azure container registry](https://learn.microsoft.com/en-us/azure/container-registry/container-registry-intro) in your Azure resource group.
   1. **Connectivity access** - Public from (all networks)
2. Collect the following information
   1. Your container registry name will be in the format `<name>.azurecr.io`
   2. The username and password for the container registry can be copied from Access keys under Settings in the left panel of the Azure portal in your container registry resource.\
      Enable the Admin User to copy the username and the password
3. Head over to the Integrations tab from the sidebar and follow the steps shown in the previous demo to complete the integration.

### Azure AKS integration

To add support of Azure AKS integration, follow the below steps

1. Create an [Azure Ad App registration](https://learn.microsoft.com/en-us/entra/identity-platform/quickstart-register-app?tabs=certificate#register-an-application)
2. Create a client secret for the Azure AD application with custom expiration set to a long time. You can do this from the left panel -> Certificates and Secrets.
3. Head over to the AKS cluster and add `Reader`, `Monitoring Reader` and `Azure Kubernetes Service Contributor Role` role to the above Azure AD application. This can be done from Azure portal -> kubernetes services -> AKS cluster -> Access Control.
4. Head over to the Integrations tab from the sidebar and follow the steps shown in the previous demo to complete the integration.

### Azure Key Vault Integration

To add Azure Key Vault as a secret store, follow the below steps:

**Prerequisites**

1. Create an [Azure Key Vault](https://learn.microsoft.com/en-us/azure/key-vault/general/quick-create-portal) in your resource group if you don't have one already.
   1. Navigate to the [Azure Portal](https://portal.azure.com) and search for "Key Vaults".
   2. Click **+ Create** to create a new Key Vault.
   3. Select your subscription and resource group.
   4. Enter a unique name for your Key Vault
   5. Select the region and pricing tier.
   6. Click **Review + create** and then **Create**.
2. Create an [Azure Ad App registration](https://learn.microsoft.com/en-us/entra/identity-platform/quickstart-register-app?tabs=certificate#register-an-application) if not done in the previous step.
3. Assign the `Key Vault Secrets Officer` role to the above Azure AD application. This can be done from Azure portal -> Key Vault -> Access Control.
4. Head over to the Integrations tab from the sidebar and follow the steps shown in the previous demo to complete the integration.
