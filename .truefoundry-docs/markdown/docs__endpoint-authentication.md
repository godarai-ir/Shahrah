> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Add authentication to your Endpoints

> Configure endpoint authentication. Secure your services with tokens and access rules.

Truefoundry makes it easy to add authentication to your endpoints. You can authenticate your
endpoint in any of the following ways:

1. **Username and Password Authentication**: In this case, you can provide a username:password combination as a security measure to guard the point. This is very easy to configure but requires passing the credentials around.

2. **JWT Authentication**: In this case, you can provide a JWT token as a security measure to guard the point. In this case, the request to the endpoint will need to have a valid JWT token issued by your identity providers (IDP) like Amazon Cognito, Google, Okta, or Microsoft Entra ID. The gateway will validate the JWT using the public key from the IDP's JWKS endpoint.

3. **Login With Truefoundry**: In this case, you can provide a Truefoundry login to guard the point. This is very easy to configure by just clicking on the checkbox to enable. In this case, the user will need to be logged in to Truefoundry to be able to access the endpoint.

<img src="https://mintcdn.com/truefoundry/5CkapnZ7CyjQJ4bx/images/docs/endpoint-authentication.png?fit=max&auto=format&n=5CkapnZ7CyjQJ4bx&q=85&s=1e4668865ee0485f54e5e7786533093b" width="3024" height="1708" data-path="images/docs/endpoint-authentication.png" />

## Username and Password Authentication

To configure this, please choose the `Username and Password` option after clicking on `Enable Authentication` and provide a value for username and password (Let's say you set the username to *`hello`* and the password to *`pass`*). Once you update the deployment, you can access the endpoint with the credentials in the following ways:

<Tabs>
  <Tab title="Access via Browser">
    When attempting to access the endpoint from a web browser, you will be prompted to enter your credentials.

    <Frame caption="">
      <img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/faee8696-465deac-browser-access.png?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=ebf152465dc861ee9b745d197c8a0395" width="2686" height="690" data-path="images/faee8696-465deac-browser-access.png" />
    </Frame>
  </Tab>

  <Tab title="Access via Python code">
    When using Python, you can pass the credentials in the following manner:

    ```python lines theme={"dark"}
    import requests
    endpoint = 'https://fastapi-my-workspace-8080.my-org.com/'
    response = requests.get(endpoint, auth=('hello', 'pass'))
    ```
  </Tab>

  <Tab title="Access via cURL">
    When using cURL, you'll need to include the credentials as part of the command:

    ```bash cURL lines theme={"dark"}
    curl -v -X GET --user hello:pass https://fastapi-my-workspace-8080.my-org.com/
    ```
  </Tab>
</Tabs>

## JWT Authentication

To setup JWT authentication, we first need to setup a **Custom JWT Auth** integration that tells Truefoundry how to validate the JWT token and also do login via Oauth in case you want Oauth2 based login for your service.

### Create a Custom JWT Auth Integration

1. Go to the **Platform** page in the left sidebar and click on Integration Providers. Click on **Add Integration Provider** button and choose the **Custom** option.

<img src="https://mintcdn.com/truefoundry/5CkapnZ7CyjQJ4bx/images/docs/custom-jwt-auth-integration.png?fit=max&auto=format&n=5CkapnZ7CyjQJ4bx&q=85&s=e198463ce1acb7c82fded623dfbce006" width="3836" height="1938" data-path="images/docs/custom-jwt-auth-integration.png" />

2. Provide the following details about the JWT Integration:

* **Name**: A name to refer to this integration later.
* **Issuer**: The base URL of the authentication provider
* **JWKS URL**: The JSON Web Key Set URI for JWT verification.
* **Oauth2 Client Configuration (Optional)**: Enable Login Provider and provide the values for `Client ID`, `Client Secret`, `Authorization URL`, `Token URL`, and any `Scopes` (e.g. `openid`, `profile`, `email`, etc.)

To get the above values, you will need to create / use an existing application with your identity provider such as Amazon Cognito, Google, Okta, or Azure AD. The steps for each of these providers are outlines below:

<Accordion title="Amazon Cognito">
  <Steps>
    <Step title="Create a Cognito User Pool">
      Follow the [AWS Cognito guide](https://docs.aws.amazon.com/cognito/latest/developerguide/getting-started-user-pools-application.html) to create a Cognito application or use an existing one.
    </Step>

    <Step title="Create an Application">
      * Choose Traditional Web Application as the Application Type.
      * Skip adding the return URL for now - it can be added later after deploying the TrueFoundry service.
      * Save ClientID and ClientSecret
    </Step>

    <Step title="Copy Details from OpenID Configuration URL">
      For e.g. the URL will look something like [https://cognito-idp.us-east-1.amazonaws.com/us-east-1\_GOoTGBS6e/.well-known/openid-configuration](https://cognito-idp.us-east-1.amazonaws.com/us-east-1_GOoTGBS6e/.well-known/openid-configuration)

      On opening the URL, you will see a JSON response and we need to extract the following details: `issuer`, `jwks_uri`, `authorization_endpoint`, `token_endpoint`, `scopes_supported`
    </Step>

    <Step title="Integrate with TrueFoundry">
      * Use the values from the previous step in the integration
      * For Client Secret, you can either directly add the value in the form or first create a [Truefoundry Secret](/docs/manage-secrets) and then use the FQN of the secret.
      * Include appropriate scopes as required for your integration (e.g., `openid`, `email`, etc.).
      * Retain JWT Source as Access Token (no changes needed).

      <Frame caption="Sample Custom JWT Auth integration for Amazon Cognito">
        <img src="https://mintcdn.com/truefoundry/s4Aj2_qGCrSP-zc8/images/7b073d6c-c0b6f400a4551110ea793b56288634cac0485f3b76e8833d5173c70db4ad6092-Screenshot_2025-03-10_at_9.54.28_AM.png?fit=max&auto=format&n=s4Aj2_qGCrSP-zc8&q=85&s=e7f132d2835a24e90471108ec7eb9a94" width="2076" height="1506" data-path="images/7b073d6c-c0b6f400a4551110ea793b56288634cac0485f3b76e8833d5173c70db4ad6092-Screenshot_2025-03-10_at_9.54.28_AM.png" />
      </Frame>
    </Step>
  </Steps>
</Accordion>

<Accordion title="Google OAuth2">
  <Steps>
    <Step title="Create a Google Cloud Project">
      Follow the [Google Cloud Console guide](https://console.cloud.google.com/) to create a Google Cloud Project (if not already available)
    </Step>

    <Step title="Enable the OAuth 2.0 API">
      * Go to APIs & Services > Credentials.
      * Click **Create Credentials** and select **OAuth Client ID**.
    </Step>

    <Step title="Create an OAuth 2.0 Application">
      * Choose **"Web Application"** as the Application Type.
      * Skip adding the Authorized Redirect URI for now - it can be added later after deploying the TrueFoundry service.
      * Save the Client ID and Client Secret
    </Step>

    <Step title="Copy Details from OpenID Configuration URL">
      * Open the URL [https://accounts.google.com/.well-known/openid-configuration](https://accounts.google.com/.well-known/openid-configuration)
      * Copy the values for `issuer`, `jwks_uri`, `authorization_endpoint`, `token_endpoint`, `scopes_supported`
    </Step>

    <Step title="Integrate with TrueFoundry">
      * Use the values from the previous step in the integration
      * For Client Secret, you can either directly add the value in the form or first create a [Truefoundry Secret](/docs/manage-secrets) and then use the FQN of the secret.
      * Include the mandatory `openid` scope, along with other required scopes (e.g., `email` etc.).
      * Set JWT Source to `ID Token` instead of Access Token.

      <Frame caption="Custom JWT Auth integration for Google OAuth2">
        <img src="https://mintcdn.com/truefoundry/FrY4JbiyZud2He3p/images/0f74c880-6f95a5881fd9d038fed95474f31b89a460eb5066dff6641e9fae28e483757037-Screenshot_2025-03-17_at_10.53.58_AM.png?fit=max&auto=format&n=FrY4JbiyZud2He3p&q=85&s=22cae1d455ddf596d06ffe6b8d0bb2b9" width="1988" height="1582" data-path="images/0f74c880-6f95a5881fd9d038fed95474f31b89a460eb5066dff6641e9fae28e483757037-Screenshot_2025-03-17_at_10.53.58_AM.png" />
      </Frame>
    </Step>
  </Steps>

  #### Steps

  1. **Create a Google Cloud Project** (if not already available) at [Google Cloud Console](https://console.cloud.google.com/).

  2. **Enable the OAuth 2.0 API:**

     * Go to **APIs & Services > Credentials**.
     * Click **"Create Credentials"** and select **OAuth Client ID**.

  3. **Create an OAuth 2.0 Application:**

     * Choose **"Web Application"** as the **Application Type**.
     * **Skip adding the Authorized Redirect URI** for now; it can be added later after deploying the TrueFoundry service.

  4. **Save the following details:**

     * **Client ID**
     * **Client Secret**

  5. **Open the OpenID Configuration:**

     * Example:

       ```javascript lines theme={"dark"}
       https://accounts.google.com/.well-known/openid-configuration
       ```

     * This contains required fields like:

       * **Issuer**
       * **JWKS URI**
       * **Authorization URL**
       * **Token URL**
       * **Supported Scopes**

  #### Integration with TrueFoundry

  * Use the values from the OpenID configuration in the integration.

  * For **Client Secret**, you can:

    * Create a **TrueFoundry Secret**, or
    * Directly add the value.

  * **Include the mandatory`openid` scope**, along with other required scopes (e.g., `email` etc.).

  * **Set JWT Source to`ID Token`** instead of Access Token.

  **Example configuration**

  <Frame caption="Custom JWT Auth integration for Google OAuth2">
    <img src="https://mintcdn.com/truefoundry/FrY4JbiyZud2He3p/images/0f74c880-6f95a5881fd9d038fed95474f31b89a460eb5066dff6641e9fae28e483757037-Screenshot_2025-03-17_at_10.53.58_AM.png?fit=max&auto=format&n=FrY4JbiyZud2He3p&q=85&s=22cae1d455ddf596d06ffe6b8d0bb2b9" width="1988" height="1582" data-path="images/0f74c880-6f95a5881fd9d038fed95474f31b89a460eb5066dff6641e9fae28e483757037-Screenshot_2025-03-17_at_10.53.58_AM.png" />
  </Frame>
</Accordion>

<Accordion title="Okta">
  <Steps>
    <Step title="Create an Okta Developer Account">
      **Create an Okta Developer Account** (if not already available) at [https://developer.okta.com/signup/](https://developer.okta.com/signup/).
    </Step>

    <Step title="Create an Application">
      * Choose Web Application as the Application Type.
      * Skip adding the Sign-in redirect URI for now - it can be added later after deploying the TrueFoundry service.
      * Save ClientID, Client Secret and Okta Domain (e.g., `https://dev-123456.okta.com`)
    </Step>

    <Step title="Copy Details from OpenID Configuration URL">
      * Open the URL [https://dev-123456.okta.com/.well-known/openid-configuration](https://dev-123456.okta.com/.well-known/openid-configuration)
      * Copy the values for `issuer`, `jwks_uri`, `authorization_endpoint`, `token_endpoint`, `scopes_supported`
    </Step>

    <Step title="Integrate with TrueFoundry">
      * Use the values from the previous step in the integration
      * For Client Secret, you can either directly add the value in the form or first create a [Truefoundry Secret](/docs/manage-secrets) and then use the FQN of the secret.
      * Include appropriate scopes as required for your integration (e.g., `offline_access`, `openid`, `email`, etc.).
      * Retain JWT Source as Access Token (no changes needed).

      <Frame caption="Custom JWT Auth integration for Okta">
        <img src="https://mintcdn.com/truefoundry/s4Aj2_qGCrSP-zc8/images/958a7aca-6efe55259b54fa2de5c073753107808d37ae60ff25d4ae860df4a1fda6a33ad1-Screenshot_2025-03-18_at_10.24.34_AM.png?fit=max&auto=format&n=s4Aj2_qGCrSP-zc8&q=85&s=070441d9f763849259d6a10b2a668853" width="1886" height="1374" data-path="images/958a7aca-6efe55259b54fa2de5c073753107808d37ae60ff25d4ae860df4a1fda6a33ad1-Screenshot_2025-03-18_at_10.24.34_AM.png" />
      </Frame>
    </Step>
  </Steps>
</Accordion>

<Accordion title="Microsoft Entra ID (Azure AD)">
  <Steps>
    <Step title="Create an Azure AD Application">
      * Follow the [Microsoft Entra ID guide](https://learn.microsoft.com/en-us/azure/active-directory/develop/quickstart-register-app) to create an application or use an existing one.
      * Save the Client ID, Client Secret and Tenant ID.
      * Skip adding the redirect URI initially. This can be added later once the TrueFoundry service deployment is created.
    </Step>

    <Step title="Copy Details from OpenID Configuration URL">
      * Open the URL [https://login.microsoftonline.com/{tenant_id}/.well-known/openid-configuration](https://login.microsoftonline.com/\{tenant_id}/.well-known/openid-configuration)
      * Copy the values for `issuer`, `jwks_uri`, `authorization_endpoint`, `token_endpoint`, `scopes_supported`
    </Step>

    <Step title="Integrate with TrueFoundry">
      * Use the values from the previous step in the integration
      * For Client Secret, you can either directly add the value in the form or first create a [Truefoundry Secret](/docs/manage-secrets) and then use the FQN of the secret.
      * Include the mandatory`openid` scope, along with other required scopes (e.g., `email`, `profile`, etc.).
      * Set JWT Source to`ID Token`

      <Frame caption="Custom JWT Auth integration for Azure AD/Microsoft Entra ID">
        <img src="https://mintcdn.com/truefoundry/jw406UAsc7ErYUq8/images/a45cc158-41bac5460fea6e32af287b5e547ee15b5c88cccbbf8c9cb8ef1e26c9cb4e0824-Screenshot_2025-03-17_at_10.48.15_AM.png?fit=max&auto=format&n=jw406UAsc7ErYUq8&q=85&s=df59c8e278de6d987b558f4a270d3a79" width="1988" height="1304" data-path="images/a45cc158-41bac5460fea6e32af287b5e547ee15b5c88cccbbf8c9cb8ef1e26c9cb4e0824-Screenshot_2025-03-17_at_10.48.15_AM.png" />
      </Frame>
    </Step>
  </Steps>
</Accordion>

### Enabling JWT Authentication in Your Service

Once you have set up the Custom JWT Auth integration, you can select `Enable Authentication` in the Ports section and choose `JWT Integration` from the options below.

<Frame caption="Sample JWT Auth Config while deploying a service">
  <img src="https://mintcdn.com/truefoundry/5CkapnZ7CyjQJ4bx/images/docs/jwt-auth-ports-service.png?fit=max&auto=format&n=5CkapnZ7CyjQJ4bx&q=85&s=8b88afff114486aca49c5175dbdb37c7" width="3834" height="1932" data-path="images/docs/jwt-auth-ports-service.png" />
</Frame>

<Tip>
  You can use the **Enable Login** option to enable OAuth2-based login for your service. This will redirect users to the OAuth2 provider's login page and, after authentication, returned to your application. This is useful for web applications and UI applications like Streamlit, Gradio. For backend APIs that only need to validate JWTs in requests, this option is not required.
</Tip>

A few key points to note:

* You can enforce additional security by verifying claims in the token by providing the key and the accepted values for that key. For e.g. if we want to restrict access to users with emails from `truefoundry.com`, we can provide key as `email` and accepted values as `*@truefoundry.com` in the Claims section.

* If **OAuth2-based login** is enabled, you must configure the **Redirect URL** in your OAuth2 provider settings.
  * After deploying the service, the Redirect URL will be displayed on the deployment page.
  * Copy this URL and add it to your OAuth2 provider's **redirect URIs** section.

<Frame caption="Redirect URL for OAuth2 Login">
  <img src="https://mintcdn.com/truefoundry/s4Aj2_qGCrSP-zc8/images/73fb1fee-2c858a61683f0c5a7954d727935866eb9d565b1228005d85923fb3e8b6300f09-Screenshot_2025-03-11_at_10.30.10_AM.png?fit=max&auto=format&n=s4Aj2_qGCrSP-zc8&q=85&s=3ecde1ae03ddce1fae7f6902025cc056" width="2008" height="692" data-path="images/73fb1fee-2c858a61683f0c5a7954d727935866eb9d565b1228005d85923fb3e8b6300f09-Screenshot_2025-03-11_at_10.30.10_AM.png" />
</Frame>

### Sending request to your Authenticated Endpoint

While sending request to an endpoint with JWT authentication enabled, you need to provide the token in the `Authorization`header. This header carries the token, allowing your service to verify and authenticate the request based on the credentials and claims present in the JWT.

Here is an example of how to send a request using curl:

```powershell cURL lines theme={"dark"}
curl -X GET "https://your-service-endpoint.truefoundry.cloud/api/data" \
     -H "Authorization: Bearer YOUR_JWT_HERE"
```

<Note> Unauthenticated requests would return `403`. </Note>

## Login with Truefoundry

If you want a quick solution to enable authentication for your service - and the users are all Truefoundry users - you can enable login with Truefoundry. In this case, the users will be prompted with a Login to Truefoundry screen and after authentication, they will be redirected to your application. To enable this, you need to tick `Enable Authentication` and select `Login with Truefoundry` option.

<Frame caption="Login With Truefoundry">
  <img src="https://mintcdn.com/truefoundry/5CkapnZ7CyjQJ4bx/images/docs/login-truefoundry-ports-service.png?fit=max&auto=format&n=5CkapnZ7CyjQJ4bx&q=85&s=c9fb8667844131dec2a1be00e24c9845" width="3024" height="1720" data-path="images/docs/login-truefoundry-ports-service.png" />
</Frame>
