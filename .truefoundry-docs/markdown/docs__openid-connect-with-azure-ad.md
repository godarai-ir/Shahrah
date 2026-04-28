> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# OIDC with Azure AD

> Configure OpenID Connect with Azure AD or Microsoft Entra ID for single sign-on access to TrueFoundry.

## Configure OpenID Connect with Azure Active Directory/Microsoft Entra ID

Once you have completed this configuration you may enable an OpenID Connect “Login with Azure AD” button for TrueFoundry dashboard. See [Microsoft Entra ID - Register An App Quickstart Guide](https://learn.microsoft.com/en-us/entra/identity-platform/quickstart-register-app) as an additional reference.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/FrY4JbiyZud2He3p/images/232a0c10-58fc69bdaeabbd88997d4c2bdb61f94cf11d99b1c2dd4c8f15ce32f69dd985e8-image.png?fit=max&auto=format&n=FrY4JbiyZud2He3p&q=85&s=0c2307596d003a24b63def900f65e97c" width="2998" height="1328" data-path="images/232a0c10-58fc69bdaeabbd88997d4c2bdb61f94cf11d99b1c2dd4c8f15ce32f69dd985e8-image.png" />
</Frame>

## Register a New Azure Active Directory Application

You will first need to login to the [Azure Portal](https://azure.microsoft.com/en-us/features/azure-portal/).

Once logged in, navigate to **Azure Active Directory -> App Registrations -> New Registration** to create a new Azure Active Directory Application.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/f5d01358-fd1144a448b0e2a4f687bd452c8595bf8fe4f9109d09dbbe8dd22bfc0139f03a-azure.png?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=2d1cc8cbc115abd9226c828c0cc9085f" width="1600" height="1620" data-path="images/f5d01358-fd1144a448b0e2a4f687bd452c8595bf8fe4f9109d09dbbe8dd22bfc0139f03a-azure.png" />
</Frame>

Here we have configured our application `Redirect URI`. Use redirect URL value as`https://login.truefoundry.com/oauth2/callback`.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/4MAaF__cLD4iud16/images/721c415a-f46419deacef695b73654f1ca037bbf992e28570796ec8490e979dc24b1b930e-azure-setup.png?fit=max&auto=format&n=4MAaF__cLD4iud16&q=85&s=17f46863b6694ce315d7b58ed687125f" width="1600" height="1620" data-path="images/721c415a-f46419deacef695b73654f1ca037bbf992e28570796ec8490e979dc24b1b930e-azure-setup.png" />
</Frame>

Once the application has been created, note the `Application (client) ID` and the `Directory (tenant) ID`. These will be used respectively as the **Client Id** value and to construct the Issuer value in your TrueFoundry OpenID Connect Identity Provider configuration.

### Device Code Flow

To enable the Device Code login flow, follow the steps below:

* Navigate to `Authentication` tab under `Manage` section in your app registration page
* Scroll down to `Advanced settings` and enable `Allow public client flows`

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/JFTbQOWMkMfvFjDC/images/aff9c9e9-9979d7367ecaba7354e95f6f1488b32d30634d81ccbbb3ef600173be188ca390-image_22.png?fit=max&auto=format&n=JFTbQOWMkMfvFjDC&q=85&s=e8e7844e6fd84524e69835d47d28b1b3" width="2148" height="1620" data-path="images/aff9c9e9-9979d7367ecaba7354e95f6f1488b32d30634d81ccbbb3ef600173be188ca390-image_22.png" />
</Frame>

## Create a New Azure Active Directory Application Secret

Navigate to **Azure Active Directory -> App Registrations ->\[Your Application] -> Certificates & secrets -> New client secret** to create a new Azure Active Directory Application Client Secret.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/f3523cc2-80b02d0720645e514a257a8c6179db4a8886ce9f109c68e0f07ce562dc97bbd9-azure-id.png?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=9548a3642c5fd77c2c1c006c5a908376" width="1600" height="1620" data-path="images/f3523cc2-80b02d0720645e514a257a8c6179db4a8886ce9f109c68e0f07ce562dc97bbd9-azure-id.png" />
</Frame>

Note the `VALUE` of the created client secret. This will be used as the Client secret value in your TrueFoundry OpenID Connect Identity Provider configuration.

## Integrate with TrueFoundry

To integrate Azure AD with TrueFoundry, Add the following configuration to [TrueFoundry SSO settings](/docs/sso#integrate-sso-with-truefoundry)

* **Issuer URL**: Example: `https://login.microsoftonline.com/{tenantId}/v2.0`
* **Tenant ID**: Tenant or Directory ID of your Azure application.
* **Client ID**: Application ID of your Azure application
* **Client Secret**: Secret value of client secret created in the above step

***
