> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# SSO Overview

> Enable single sign-on for TrueFoundry with identity providers like GSuite, Azure AD, Okta, or Keycloak.

TrueFoundry allows your team to enable SSO with your Identity Provider (IdP) by leveraging OpenID Connect (OIDC) or Security Assertion Markup Language (SAML). TrueFoundry SSO integration provides a seamless way to sign in with your own IdPs and also eliminates the need for employees to maintain/enter credentials.

TrueFoundry can integrate with the following IdP providers: **GSuite**, **AzureAD**, **Okta**, **Keycloak**

<Tip>
  If you don't see the name of your IdP provider above, there is a high chance
  your IdP is also supported as long as it support OpenID Connect (OIDC) or SAML
  protocol.
</Tip>

## How does TrueFoundry SSO work?

Truefoundry maintains a centralized Authentication Server that handles the authentication flow to the platform (as described [here](/docs/platform/architecture)). The SSO integration of your IDP is done by configuring the Authentication Server to use your IdP as the Identity Provider. **This helps us control the licensing and track the number of users accessing the platform.**

When a user logs into the control plane, the control-plane redirects to login.truefoundry.com, which is the Authentication Server. The Authentication Server then redirects to the IdP's login page, which then redirects back to the Authentication Server with an authorization code. The Authentication Server then uses this authorization code to fetch user information from the IdP and create a new user in the platform if they don't exist. The user is then redirected to the control plane with the authentication tokens. The flow of the request is described as follows:

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/iMid4yIOHvzf4Z4V/images/sso-flow-dia-1.png?fit=max&auto=format&n=iMid4yIOHvzf4Z4V&q=85&s=191ff71b269a641e2c283c3941356baf" width="2082" height="990" data-path="images/sso-flow-dia-1.png" />
</Frame>

<Steps>
  <Step title="Request Login Page">
    Browser initiates the OAuth flow by requesting the login page from TrueFoundry Control Plane, with the redirect URL being the TrueFoundry Control Plane URL.

    <Accordion title="Example Request/Response">
      ```bash Request lines theme={"dark"}
      GET /api/svc/v1/oauth2/authorize?tenantName={tenant_name}
          &controlPlaneURL=https%3A%2F%2Fyour.example.com
          &redirectURL=https%3A%2F%2Fyour.example.com%2Fauth%2Fcallback
      HTTP/1.1 
      Host: your.example.com
      ```

      ```bash Response lines theme={"dark"}
      HTTP/1.1 200 OK
      Content-Type: text/html; charset=utf-8
      Content-Length: 100
      Connection: keep-alive
      Date: Fri, 27 Jun 2025 07:00:00 GMT
        {
          "authorizationEndpoint":"https://login.truefoundry.com/oauth2/authorize?client_id={application_id}&response_type=code&scope=openid+email+profile+offline_access&redirect_uri=https%3A%2F%2Fyour.example.com%2Fauth%2Fcallback"
        }
      ```
    </Accordion>
  </Step>

  <Step title="Send Login Page">
    Browser redirects to the Login Page URL powered by <Tooltip tip="https://login.truefoundry.com">TrueFoundry Auth Server</Tooltip>. The login page provides an interactive form for the user to choose the type of authentication, such as password-based or SSO.

    <Frame caption="">
      <img src="https://mintcdn.com/truefoundry/iMid4yIOHvzf4Z4V/images/sso-flow-dia-2.png?fit=max&auto=format&n=iMid4yIOHvzf4Z4V&q=85&s=3c60e69582778070c50e349ec5d8acf8" width="3018" height="1770" data-path="images/sso-flow-dia-2.png" />
    </Frame>
  </Step>

  <Step title="Redirect to SSO Identity Provider page">
    On selecting SSO, the browser is redirected to the SSO Identity Provider’s login page with the redirect URL as the <Tooltip tip="https://login.truefoundry.com">TrueFoundry Auth Server</Tooltip>

    <Accordion title="Example Request/Response">
      ```bash Request  lines theme={"dark"}
      POST /oauth2/<client_id>/v1/authorize 
      HTTP/1.1 
      Host: example.sso.com 
      Content-Type: application/x-www-form-urlencoded 
      Origin: https://login.truefoundry.com 
      Body: login_hint=
            &scope=openid+groups+email
            &response_type=code
            &code_challenge_method=S256
            &redirect_uri=https%3A%2F%2Flogin.truefoundry.com%2Foauth2%2Fcallback
            &state={encoded_state}
            &client_id={client_id}
            &code_challenge={code_challenge}
      ```

      ```bash Response lines theme={"dark"}
      HTTP/1.1 302 Found
      Cache-Control: no-cache, no-store
      Pragma: no-cache
      Server: nginx
      Date: Fri, 27 Jun 2025 07:00:00 GMT
      Location: https://login.truefoundry.com/oauth2/callback?code={authorization_code}&state={encoded_state}
      ```
    </Accordion>

    <Check>
      All requests to the Identity Provider are signed using a RS256 key pair. In case of SAML, destination assertion is enabled ensuring secure transfer of user authentication information between the SAML identity provider (IdP) and TrueFoundry Auth Server (SP).
    </Check>

    <Frame caption="">
      <img src="https://mintcdn.com/truefoundry/iMid4yIOHvzf4Z4V/images/sso-flow-dia-8.png?fit=max&auto=format&n=iMid4yIOHvzf4Z4V&q=85&s=4c640aaba27ea66ff5de8312f0cf4631" width="1260" height="532" data-path="images/sso-flow-dia-8.png" />
    </Frame>
  </Step>

  <Step title="Redirect to the TrueFoundry Auth Server with code">
    On successful authentication by SSO Identity Provider, the browser redirects to the configured redirect URL, i.e. <Tooltip tip="https://login.truefoundry.com">TrueFoundry Auth Server</Tooltip>, with an authorisation code. The Auth server then uses this authorisation code to fetch user information using a POST HTTP request to the SSO Identity Provider.

    <Accordion title="Example Request/Response">
      ```bash Request lines theme={"dark"}
      GET /oauth2/callback?code={authorization_code}&state={encoded_state} 
      HTTP/1.1 
      Host: login.truefoundry.com
      ```

      ```bash Response lines theme={"dark"}
      HTTP/1.1 302 Found
      Cache-Control: no-cache, no-store
      Pragma: no-cache
      Server: nginx
      Date: Fri, 27 Jun 2025 07:00:00 GMT
      Location: https://your.example.com/auth/callback?code={authorization_code}&userState=Authenticated
      ```
    </Accordion>

    <Check>
      To protect data integrity and privacy, all data in transit to and from TrueFoundry Control Plane or Auth Server is encrypted using TLS 1.2 or higher.
    </Check>
  </Step>

  <Step title="Validate & Create User if it doesn't exist">
    <Tooltip tip="https://login.truefoundry.com">TrueFoundry Auth Server</Tooltip> validates the response from the SSO Identity provider and maps the user based on the email claim with the user's email in the database, and creates a new entry if not present already.

    <Accordion title="Example Request/Response">
      This step is performed by the <Tooltip tip="https://login.truefoundry.com">TrueFoundry Auth Server</Tooltip> and is not visible to the user.

      Get Token from SSO Identity Provider:

      ```bash Request lines theme={"dark"}
      POST /oauth2/{client_id}/v1/token
      HTTP/1.1
      Host: example.sso.com
      Content-Type: application/x-www-form-urlencoded
      Origin: https://login.truefoundry.com
      Body: code={authorization_code}
            &grant_type=authorization_code
            &redirect_uri=https://login.truefoundry.com/oauth2/callback
      ```

      ```bash Response lines theme={"dark"}
      HTTP/1.1 200 OK
      Content-Type: application/json
      Date: Fri, 27 Jun 2025 07:00:00 GMT
        {
          "token_type" : "Bearer",
          "expires_in" : 3600,
          "access_token" : "{access_token}",
          "scope" : "groups openid email",
          "id_token" : "{id_token}"
        }
      ```

      Get User information from SSO Identity Provider:

      ```bash Request lines theme={"dark"}
      GET /oauth2/{client_id}/v1/userinfo
      HTTP/1.1
      Host: example.sso.com
      Authorization: Bearer {access_token}
      ```

      ```bash Response lines theme={"dark"}
      HTTP/1.1 200 OK
      Content-Type: application/json
      Date: Fri, 27 Jun 2025 07:00:00 GMT
        {
          "sub" : "{sso_user_id}",
          "email" : "example@example.com",
          "email_verified" : true,
          "groups" : ["{group_1}", "{group_2}"]
        }
      ```

      Link user via email `example@example.com` to user in TrueFoundry Auth Server and redirects to TrueFoundry Control Plane.
    </Accordion>

    <Note>
      Any user deactivated from the platform is rejected at this step.
    </Note>
  </Step>

  <Step title="Redirect to TrueFoundry Control Plane with the OAuth authorisation code">
    On user validation, the <Tooltip tip="https://login.truefoundry.com">TrueFoundry Auth Server</Tooltip> redirects the control to TrueFoundry Control Plane via browser with an authorisation code

    <Accordion title="Example Request/Response">
      ```bash Request lines theme={"dark"}
      GET /auth/callback?code={authorization_code}&userState=Authenticated
      HTTP/1.1 
      Host: your.example.com
      ```

      ```bash Response lines theme={"dark"}
      HTTP/1.1 200 OK
      Date: Fri, 27 Jun 2025 07:00:00 GMT
      ```
    </Accordion>
  </Step>

  <Step title="Request tokens using OAuth authorisation code">
    TrueFoundry Control Plane uses the code to request tokens using a POST HTTP request to the <Tooltip tip="https://login.truefoundry.com">TrueFoundry Auth Server</Tooltip>.

    <Accordion title="Example Request/Response">
      ```bash Request lines theme={"dark"}
      POST /api/svc/v1/oauth2/token
      HTTP/1.1 
      Host: your.example.com
      Content-Type: application/json
      Origin: https://your.example.com
      Body: {
              "tenantName":"{tenant_name}",
              "code":"{authorization_code}",
              "grantType":"authorization_code",
              "redirectURI":"https://your.example.com/auth/callback"
            }
      ```

      ```bash Response lines theme={"dark"}
      HTTP/1.1 201 Created
      Date: Fri, 27 Jun 2025 07:00:00 GMT
      Set-Cookie: accessToken=<access_token>; HttpOnly; Path=/; Max-Age=86400; Expires=Sat, 28 Jun 2025 07:00:00 GMT
      Set-Cookie: refreshToken=<refresh_token>; HttpOnly; Path=/; Max-Age=604800; Expires=Thu, 03 Jul 2025 07:00:00 GMT
      ```
    </Accordion>
  </Step>

  <Step title="Return authentication tokens">
    On successful code validation, <Tooltip tip="https://login.truefoundry.com">TrueFoundry Auth Server</Tooltip> responds with authentication tokens that include an access token and a refresh token, signed by <Tooltip tip="https://login.truefoundry.com">TrueFoundry Auth Server</Tooltip>.

    <Info>
      By default, the access token is **valid** for **1 day** and the refresh token is valid for **7 days**. You can change the token expiry by contacting support.
    </Info>
  </Step>

  <Step title="Set tokens as an HttpOnly Cookie">
    TrueFoundry Control Plane sets these authentication tokens as HTTP-only cookies in the Browser. All further requests to TrueFoundry Control Plane contain the same cookies and are used for authentication and authorisation at the API server layer

    <Frame caption="">
      <img src="https://mintcdn.com/truefoundry/iMid4yIOHvzf4Z4V/images/sso-flow-dia-9.png?fit=max&auto=format&n=iMid4yIOHvzf4Z4V&q=85&s=aad6cdb091d628642959cf32da02c2c0" width="1614" height="360" data-path="images/sso-flow-dia-9.png" />
    </Frame>
  </Step>
</Steps>

## Configure SSO in TrueFoundry

Go to **Platform -> Settings -> SSO**, toggle **Enabled**, pick your **SSO Provider** (Okta, Google, Azure AD, or Custom), then choose **OIDC** or **SAML v2** under Authentication Configuration.

<img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/docs/set-sso-settings.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=a796756c337f61bc4517d87391e64e25" alt="SSO Settings" width="3838" height="1866" data-path="images/docs/set-sso-settings.png" />

You can enable the SSO settings and fill up the form with the fields as described below:

<img src="https://mintcdn.com/truefoundry/2MMcllD7kMlpnaWX/images/sso-tfy-form.png?fit=max&auto=format&n=2MMcllD7kMlpnaWX&q=85&s=2d0dc47c9999ac2441cc5d2dd8a3ee0a" alt="SSO configuration in TrueFoundry" width="2058" height="1470" data-path="images/sso-tfy-form.png" />

You can choose any of the SSO providers: **Google**, **AzureAD**, **Okta**, or custom (**KeyCloak** can be integrated using Custom). In case you are using any of the above, you can read our specialized documentation for each of them:

1. [OIDC with AzureAD](/docs/openid-connect-with-azure-ad)
2. [OIDC with Okta](/docs/openid-connect-with-okta)
3. [OIDC with Keycloak](/docs/openid-connect-with-keycloak)
4. [SAML v2 with AzureAD](/docs/saml-v2-with-azure-ad)

In case you don't find the guide above, you can follow the steps below for OIDC and SAML v2:

<Tabs>
  <Tab title="OpenID Connect (OIDC)">
    <Steps>
      <Step title="Create a client application in your IdP">
        Set the redirect/callback URL to
        `https://login.truefoundry.com/oauth2/callback`
      </Step>

      <Step title="Fill up the TrueFoundry SSO settings form">
        Fill up the TrueFoundry SSO settings form with the fields as described
        below: - `Client ID` - `Client Secret` - `Discover endpoints` — If you
        enable this as True, we will automatically discover the endpoints from the
        issuer URL. If you mark this as disabled, you will need to manually set the
        `Authorization Endpoint`, `Token Endpoint` and `UserInfo Endpoint` - `Issuer
                    URL` - your IdP’s issuer - used with discovery to auto‑populate endpoints. -
        `Scopes` - Space separated list of scopes. Defaults to `openid email`
      </Step>
    </Steps>
  </Tab>

  <Tab title="SAML v2">
    <Steps>
      <Step title="Create a service provider in your IdP">
        When adding **ACS / Single sign‑on URL** or **Audience / SP Entity ID** you can use placeholder values for now and update it once you save the configuration in TrueFoundry.
      </Step>

      <Tip>
        **Simplify claim configuration with attribute mapping(Optional)**

        SAML identity providers often use complex, vendor-specific attribute names for user information. For example, attributes like `http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress` or `http://schemas.xmlsoap.org/ws/2005/05/identity/claims/nameidentifier` require you to configure custom claim names in TrueFoundry's advanced settings.

        **Recommended approach:** Configure attribute mapping in your IdP to use standard, simple names that match TrueFoundry's defaults:

        | Attribute Name | Map to                                 | Purpose                           |
        | -------------- | -------------------------------------- | --------------------------------- |
        | `sub`          | Your IdP's user ID field               | User's unique identifier          |
        | `email`        | Your IdP's email field                 | User's email address              |
        | `groups`       | Your IdP's groups field (if supported) | User's group memberships for RBAC |

        This allows TrueFoundry's default settings to work automatically without needing to specify custom claim names in the advanced settings.

        <Frame caption="Example: Configuring attribute mappings in Okta">
          <img src="https://mintcdn.com/truefoundry/uaD0tzrO0Vxuaycb/images/docs/updated-sso-claim.png?fit=max&auto=format&n=uaD0tzrO0Vxuaycb&q=85&s=e06478e7a368daaad3f5dd9360dbfe80" alt="SAML attribute mapping configuration" width="1460" height="834" data-path="images/docs/updated-sso-claim.png" />
        </Frame>
      </Tip>

      <Step title="Fill up the TrueFoundry SSO settings form">
        Fill up the TrueFoundry SSO settings form with the fields as described below:

        * `Identity Provider Endpoint (IdP SSO URL)` — your IdP's SAML login URL.
        * `X.509 Certificate (PEM)` — the IdP signing certificate used to verify SAML responses.
      </Step>

      <Step title="Update the IdP configuration">
        After you save your SAML configuration:

        * Replace the placeholder `ACS / Single sign‑on URL` with the actual `Callback URL` shown in TrueFoundry.
        * Replace the placeholder `Audience / SP Entity ID` with the actual `Issuer` shown in TrueFoundry portal.

        <img src="https://mintcdn.com/truefoundry/2MMcllD7kMlpnaWX/images/sso-saml-metadata.png?fit=max&auto=format&n=2MMcllD7kMlpnaWX&q=85&s=ea9616502066bd8f6f902f1d85eaba10" alt="SSO SAML metadata in TrueFoundry" width="1406" height="576" data-path="images/sso-saml-metadata.png" />
      </Step>
    </Steps>
  </Tab>
</Tabs>

### Advanced Customizations (apply to both OIDC and SAML)

You can customize the login button that will be shown to the user while logging into Truefoundry and token settings.

* `Button Text` / `Button Image URL` - customize the login button.
* `Email Claim` - claim/attribute carrying the user's email (default: `email`).
* `Unique ID Claim` - claim/attribute carrying a unique user ID (default: `sub`).

### Enable SCIM Provisioning

TrueFoundry supports [SCIM (System for Cross-domain Identity Management)](https://scim.cloud/) to automatically create and update users and groups from your identity provider. SCIM provisioning is available for SAML v2 SSO configurations.

<Info>
  SCIM enables automatic user lifecycle management — when you add or remove
  users in your identity provider, TrueFoundry will automatically sync those
  changes.
</Info>

**Steps to configure SCIM:**

<Note>
  #### SCIM: Group Name Convention

  Only alphanumeric characters and "-" are allowed. Group names can be up to 36 characters in length.

  <Accordion title="Example Group Names">
    | Names                                             | Status                            |
    | ------------------------------------------------- | --------------------------------- |
    | `engineering-team`                                | ✅ Valid                           |
    | `data-science-ml`                                 | ✅ Valid                           |
    | `product123`                                      | ✅ Valid                           |
    | `team-alpha-beta-gamma-delta-epsilon`             | ✅ Valid (36 characters)           |
    | `admin_users`                                     | ❌ Invalid (contains underscore)   |
    | `sales team`                                      | ❌ Invalid (contains space)        |
    | `finance@ops`                                     | ❌ Invalid (contains @ symbol)     |
    | `very-long-group-name-exceeding-limit-characters` | ❌ Invalid (exceeds 36 characters) |
  </Accordion>
</Note>

<Steps>
  <Step title="Enable SCIM when creating SAML SSO">
    When configuring a new SAML v2 SSO provider, check the **Enable SCIM** checkbox at the bottom of the authentication configuration form.

    <img src="https://mintcdn.com/truefoundry/nPY6hmSlBVBPJNoL/images/enable_scim.png?fit=max&auto=format&n=nPY6hmSlBVBPJNoL&q=85&s=3e3e73b466f7bc1f1f69cfc2e128d3f8" alt="Enable SCIM checkbox in SAML SSO configuration" width="1526" height="812" data-path="images/enable_scim.png" />
  </Step>

  <Step title="Save the SSO configuration">
    Complete the SAML configuration with your Identity Provider Endpoint and X.509
    Certificate, then save the configuration.
  </Step>

  <Step title="Get the SCIM URL">
    After saving, expand your SSO configuration to view the SCIM URL along with other metadata. Copy this URL — you'll need it for your identity provider's SCIM settings.

    <img src="https://mintcdn.com/truefoundry/f3KTstSSq3H4Ys2R/images/scim_url.png?fit=max&auto=format&n=f3KTstSSq3H4Ys2R&q=85&s=856a4bdbb60d9ebe72a42569a3a81a58" alt="SCIM URL displayed in SSO configuration" width="1408" height="622" data-path="images/scim_url.png" />
  </Step>

  <Step title="Get the SCIM Token">
    Click the **key icon** (Get SCIM Token) next to your SSO configuration to generate and copy the SCIM authentication token.

    <img src="https://mintcdn.com/truefoundry/83e0kg2LyHcfZ91Y/images/get_scim_token.png?fit=max&auto=format&n=83e0kg2LyHcfZ91Y&q=85&s=7fa1462dd1356c95b6f651b2e641b62d" alt="Get SCIM Token option" width="1388" height="588" data-path="images/get_scim_token.png" />

    <Warning>
      Store this token securely. You'll need it to authenticate SCIM requests from your identity provider.
    </Warning>
  </Step>

  <Step title="Configure your Identity Provider">
    In your identity provider's SCIM provisioning settings:

    * Set the **SCIM Base URL** to the SCIM URL from TrueFoundry
    * Set the **Authentication** method to Bearer Token
    * Use the SCIM Token as the Bearer token value
  </Step>
</Steps>
