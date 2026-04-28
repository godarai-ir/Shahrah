> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# OIDC with Okta

> Configure OpenID Connect with Okta to enable single sign-on for TrueFoundry dashboard access.

## Configure OpenID Connect with Okta

Once you have completed this configuration, you may enable an OpenID Connect "Login with Okta" button for TrueFoundry dashboard.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/4MAaF__cLD4iud16/images/603d5fe9-b3e0689d15b7e1537e5bf6bda9767232b61a5681d9305bb8bc116970d18e290c-image.png?fit=max&auto=format&n=4MAaF__cLD4iud16&q=85&s=96a008e9cd27df7ddd14e68ed51a6001" width="2998" height="1328" data-path="images/603d5fe9-b3e0689d15b7e1537e5bf6bda9767232b61a5681d9305bb8bc116970d18e290c-image.png" />
</Frame>

## Prerequisites

* A TrueFoundry instance running on a publicly accessible URL.
* An [Okta Workforce Identity Cloud Account](https://www.okta.com/) connected to a business email address. The [Okta Verify app](https://help.okta.com/en-us/Content/Topics/Mobile/okta-verify-overview.htm) if your Okta account is not configured to send SMS messages for 2FA.

## Okta Configuration

Log in to Okta and navigate to the **Admin** panel.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/4MAaF__cLD4iud16/images/606dff84-2f8e74705c33b0bed7f776e27690639fdfcba91d7fe43a45534ae1546a02a7ad-okta-admin.png?fit=max&auto=format&n=4MAaF__cLD4iud16&q=85&s=999db0b2ff0ce4c4222ecc54dfc8fd89" width="1600" height="1627" data-path="images/606dff84-2f8e74705c33b0bed7f776e27690639fdfcba91d7fe43a45534ae1546a02a7ad-okta-admin.png" />
</Frame>

Navigate to Applications -> Applications and click the `Create App Integration` button.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/DdP_2rhue4AQQlob/images/4d1d1636-ac1dc892837ae782031a832b2f041c560bdd6cc61d2f2d13550048813bf0b563-okta-dash.png?fit=max&auto=format&n=DdP_2rhue4AQQlob&q=85&s=d1c5bcf8e3bd4a87d76e5d5bbb11fb02" width="1600" height="1029" data-path="images/4d1d1636-ac1dc892837ae782031a832b2f041c560bdd6cc61d2f2d13550048813bf0b563-okta-dash.png" />
</Frame>

Select OIDC - OpenID Connect and Web Application then click `Next`.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/FrY4JbiyZud2He3p/images/2477b5ba-c4fead2210f5d5937551119c91262c12e5fdc0ad3e4617555f0df8ba30c39f86-okta-next.png?fit=max&auto=format&n=FrY4JbiyZud2He3p&q=85&s=5388ba37dcc30a59b9732724d74c6b1f" width="1600" height="1627" data-path="images/2477b5ba-c4fead2210f5d5937551119c91262c12e5fdc0ad3e4617555f0df8ba30c39f86-okta-next.png" />
</Frame>

Enter an App integration name and ensure that Authorization Code is checked. Then, supply a Sign-in redirect URI as `https://login.truefoundry.com/oauth2/callback`.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/DdP_2rhue4AQQlob/images/30675878-227cd70c8718fd1ee7bee6a26edb25acf10974a263c19022aa387267bc9c0895-okta-url.png?fit=max&auto=format&n=DdP_2rhue4AQQlob&q=85&s=f39e3e205f1a7484da6580458512c95d" width="1600" height="1337" data-path="images/30675878-227cd70c8718fd1ee7bee6a26edb25acf10974a263c19022aa387267bc9c0895-okta-url.png" />
</Frame>

Under the Assignments section, select Skip group assignment for now . Hit `Save`.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/4MAaF__cLD4iud16/images/582f5f2a-42e1bb323715e8a8b3daf63c4144ea28aadfdaf8e65318e71753c12b4fe8c005-okta-save.png?fit=max&auto=format&n=4MAaF__cLD4iud16&q=85&s=cd0b3694561108fdb16fa02310243827" width="1600" height="1627" data-path="images/582f5f2a-42e1bb323715e8a8b3daf63c4144ea28aadfdaf8e65318e71753c12b4fe8c005-okta-save.png" />
</Frame>

Copy the **Client ID** and **Secret** into a text file for later use.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/FrY4JbiyZud2He3p/images/19a7700e-278a942f72cda2588c9b764fe0e2ab7dc051eed4f30f845d056e2c0453275229-okta-use.png?fit=max&auto=format&n=FrY4JbiyZud2He3p&q=85&s=3af8770e78766cbbf11e93389c4fa0ce" width="1600" height="1627" data-path="images/19a7700e-278a942f72cda2588c9b764fe0e2ab7dc051eed4f30f845d056e2c0453275229-okta-use.png" />
</Frame>

Under **Sign On**, navigate to the section **OpenID Connect ID Token** and change the Issuer to use the `Okta URL`. Hit `Save` and copy this URL into a text file

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/s4Aj2_qGCrSP-zc8/images/821e989e-866f062e54aae359f5e8c102c31a04b83a04e59bb303858ca808b528870f32a0-okta-file.png?fit=max&auto=format&n=s4Aj2_qGCrSP-zc8&q=85&s=44454d8d609aacf5dae22cdc7fa3bfae" width="1600" height="1627" data-path="images/821e989e-866f062e54aae359f5e8c102c31a04b83a04e59bb303858ca808b528870f32a0-okta-file.png" />
</Frame>

Navigate to **Directory -> People** and click on the username of the user you’d like to authenticate. Then click `Assign Applications` and hit the Assign button next to the one you created. Then click `Save and Go Back` and `Done`.

## Integrate with TrueFoundry

To integrate Okta with TrueFoundry, Add the following configuration to [TrueFoundry SSO settings](/docs/sso#integrate-sso-with-truefoundry)

* **Issuer URL**: Issuer of your Okta application. Example: `https://{your-okta-domain}.okta.com`
* **Client ID**: Application ID of your Okta application
* **Client Secret**: The secret value of the client secret created in the above step

***
