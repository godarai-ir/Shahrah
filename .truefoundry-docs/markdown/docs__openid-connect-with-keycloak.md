> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# OIDC with Keycloak

> Register a Keycloak client and integrate it with TrueFoundry for OpenID Connect based authentication.

## Register a New Keycloak Client

Navigate to **Clients -> Create Client** to create a new Keycloak Client. Set the following values:

* Client Type: OpenID Connect in General Settings tab
* Client ID: truefoundry or any other name you prefer in General Settings tab

<img src="https://mintcdn.com/truefoundry/iMid4yIOHvzf4Z4V/images/sso-keycloak-client-reg-1.png?fit=max&auto=format&n=iMid4yIOHvzf4Z4V&q=85&s=a88388e83699e3ae0a49397f204d58d2" width="2446" height="1510" data-path="images/sso-keycloak-client-reg-1.png" />

* Set Client Authentication to On in Capability Config tab
* Select Standard flow, Direct access grants and OAuth 2.0 Device Authorization Grant under Authentication flow in Capability Config tab.

<img src="https://mintcdn.com/truefoundry/2MMcllD7kMlpnaWX/images/sso-keycloak-client-reg-2.png?fit=max&auto=format&n=2MMcllD7kMlpnaWX&q=85&s=62962930868aa4bcba5f66a497c65900" width="2436" height="1502" data-path="images/sso-keycloak-client-reg-2.png" />

* Set Valid Redirect URIs to `https://login.truefoundry.com/oauth2/callback` in Login Settings tab. Root URL can be set as your control plane URL.

<img src="https://mintcdn.com/truefoundry/2MMcllD7kMlpnaWX/images/sso-keycloak-client-reg-3.png?fit=max&auto=format&n=2MMcllD7kMlpnaWX&q=85&s=e74a6388bb6127c664c1f27a8b2e3f69" width="2436" height="1502" data-path="images/sso-keycloak-client-reg-3.png" />

<Note>
  Please note that we do not support RSA-OAEP as encryption algorithm. Please ensure that Keys with RSA-OAEP algorithm are not used. Support algorithms are ES384, HS384, ES256, HS256, HS512, PS384, RS384, PS256, RS256, PS512, RS512, none, ES512.

  <Accordion title="You can disable algorithm by disabling the Key provider under Realm Settings -> Keys.">
    <img src="https://mintcdn.com/truefoundry/2MMcllD7kMlpnaWX/images/sso-keycloak-client-reg-4.png?fit=max&auto=format&n=2MMcllD7kMlpnaWX&q=85&s=443b44539bf6e6f14d46a3473fb6a435" width="2446" height="1514" data-path="images/sso-keycloak-client-reg-4.png" />
  </Accordion>
</Note>

## Integrate with TrueFoundry

To integrate Keycloak with TrueFoundry, Add the following configuration to [TrueFoundry SSO settings](/docs/sso#integrate-sso-with-truefoundry)

* **SSO Provider**: `Custom`
* Select `OIDC` as Authentication configuration
* Set `Client ID` to the client ID of your Keycloak application
* Set `Client Secret` to the client secret of your Keycloak application. Can be found in `Credentials` tab of your Keycloak client.
* Set `Issuer` to the issuer of your Keycloak application. Example: `https://<keycloak-host>/realms/<realm-name>`

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/2MMcllD7kMlpnaWX/images/sso-tfy-form.png?fit=max&auto=format&n=2MMcllD7kMlpnaWX&q=85&s=2d0dc47c9999ac2441cc5d2dd8a3ee0a" width="2058" height="1470" data-path="images/sso-tfy-form.png" />
</Frame>

Click on Save button to save the changes.

***
