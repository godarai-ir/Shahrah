> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# SAML V2 With Azure AD

> Configure SAML v2 authentication with Azure AD to enable enterprise single sign-on for TrueFoundry.

## Configure SAML v2 for Azure AD/Microsoft Entra ID

This documentation will guide you in configuring SAML v2 IdP for Azure AD/Microsoft Entra ID. In this case, TrueFoundry will act as Service Provider (SP) to Azure AD (IdP). Functionally, the result will allow you to display a “Login with Azure AD” button on your TrueFoundry login page and connect via SAML to Azure AD users/applications.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/DdP_2rhue4AQQlob/images/413a9df8-f33d20d6c06da13f7a2d824b545714ea2f3ed4275763c806df274f4b7d32cf30-image.png?fit=max&auto=format&n=DdP_2rhue4AQQlob&q=85&s=b4cff858a5b64bcb36fdd1a5857f9e86" width="2998" height="1328" data-path="images/413a9df8-f33d20d6c06da13f7a2d824b545714ea2f3ed4275763c806df274f4b7d32cf30-image.png" />
</Frame>

## Create an Application in Azure

If you have already configured an Azure AD Enterprise application, skip this section. If you have not, please follow the brief steps outlined below:

1. From the [Azure account portal](https://portal.azure.com/#home) navigate to `Enterprise Applications`.
2. At the top of the screen click on `New application`.
3. Click on `Create your own application`.
4. Name the application
5. Select the third option - `Integrate any other application you don't find in the gallery (Non-gallery)`.
6. Click `Create`

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/DdP_2rhue4AQQlob/images/2c75cce4-3a2c02f08494bbcf16bfbcb922d1f359542d17170e6fa7034ee8fc604bcabe16-tf-create-app.png?fit=max&auto=format&n=DdP_2rhue4AQQlob&q=85&s=abf438684fa435d7e86b78d77b5b53bb" width="1600" height="738" data-path="images/2c75cce4-3a2c02f08494bbcf16bfbcb922d1f359542d17170e6fa7034ee8fc604bcabe16-tf-create-app.png" />
</Frame>

## Configure Your Azure Application

From your application home screen, click on Single sign-on. Select the SAML option.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/s4Aj2_qGCrSP-zc8/images/889cf40f-32983e8d642d7a12242a1f479d34b41e23b62d13ca3a7c0c3eac489dd8559a39-tf-saml.png?fit=max&auto=format&n=s4Aj2_qGCrSP-zc8&q=85&s=6d3a49ac004afb31a922d0f90f818950" width="1600" height="1120" data-path="images/889cf40f-32983e8d642d7a12242a1f479d34b41e23b62d13ca3a7c0c3eac489dd8559a39-tf-saml.png" />
</Frame>

## Integrate with TrueFoundry

To integrate Azure AD with TrueFoundry, Add the following configuration to [TrueFoundry SSO settings](/docs/sso#integrate-sso-with-truefoundry)

* **Login URL**: This value can be obtained from your Azure AD Application as demonstrated below. You will want to copy the Login URL value from Azure AD into this field.

* <Frame caption="">
    <img src="https://mintcdn.com/truefoundry/4MAaF__cLD4iud16/images/6555fa9e-f1cb5d6b42e2984de1746be8c4b5b306053136a8aebd77b4dbd67dba73bf3e54-image.png?fit=max&auto=format&n=4MAaF__cLD4iud16&q=85&s=d67fad85e8b2994ad558676aadf5bce1" width="1600" height="260" data-path="images/6555fa9e-f1cb5d6b42e2984de1746be8c4b5b306053136a8aebd77b4dbd67dba73bf3e54-image.png" />
  </Frame>

  **Verification key**: From the overview of your SAML application in Azure AD, under step three, you should find a Certificate (Base64) to download.

  <Frame caption="">
    <img src="https://mintcdn.com/truefoundry/OHzlp6GY5G-JfKle/images/ba491df1-41cd3621730f73a0ff070207fad73c2956a4ff6d0a9c6e1a53e9548c1fb26be7-image.png?fit=max&auto=format&n=OHzlp6GY5G-JfKle&q=85&s=5b330807bc170db7d1b41f8b304f5ef0" width="1600" height="343" data-path="images/ba491df1-41cd3621730f73a0ff070207fad73c2956a4ff6d0a9c6e1a53e9548c1fb26be7-image.png" />
  </Frame>

***
