> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Email Integration

> Set up email integrations for alerts and notifications in TrueFoundry.

To receive alerts and notifications in TrueFoundry, you need to set up notification integrations for Email. This guide walks you through adding Email notification integration, describing each field in the forms, and providing helpful links for setup.

## Email Notification Integration

To add an Email notification integration:

1. **Go to the Integrations Tab** in the TrueFoundry UI.
2. **Click on 'Add Integration Provider'.**
3. **Select 'Custom' as the provider type.**
4. **Choose 'Email Notification Channel'** in the integrations section.
5. **Fill out the form fields as shown below:**

<Frame>
  <img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/email-integration.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=5a3d89a416678c5c3b5d44e2f25c6ca1" alt="Email Integration Form" width="1682" height="1500" data-path="images/email-integration.png" />
</Frame>

### Field Descriptions

* **Name**: A unique name for this integration provider (e.g., "Production SMTP").
* **Custom Basic Auth (Optional)**: Enable if your SMTP server requires custom authentication.
* **Display Name**: The display name for this email notification channel (e.g., "Alerts Email Channel").
* **From Email**: The email address that will appear as the sender of alert emails.
* **SMTP Credentials**:
  * **Host**: The SMTP server hostname (e.g., `smtp.gmail.com`).
  * **Username**: The username for SMTP authentication (often your email address).
  * **Password**: The password or app-specific password for SMTP authentication.
  * **Port**: The port for SMTP (commonly `587` for TLS, `465` for SSL, or `25` for non-encrypted).
  * **TLS**: Enable if your SMTP server requires TLS encryption (recommended for most providers).
* **Access Control (Optional)**: List of users or groups allowed to use this integration. By default, set to `everyone`.

> **Note:**
>
> * For more information about SMTP and how to get your SMTP credentials, see [What is SMTP? (Mailgun Docs)](https://www.mailgun.com/blog/deliverability/what-is-smtp/).
> * If using Gmail, you may need to generate an [App Password](https://support.google.com/accounts/answer/185833?hl=en) and enable "Less secure app access" or use OAuth2.

### How to get SMTP credentials?

#### AWS SES

* For AWS SES, you can use the IAM user credentials to get the SMTP credentials. Follow the [AWS SES documentation](https://docs.aws.amazon.com/ses/latest/dg/smtp-credentials.html#smtp-credentials-console) to get the SMTP credentials.
* SMTP Host can be found in the AWS SES console under the "SMTP Settings" section.
* Add your domain as verified Identity in the AWS SES console. Follow the [AWS SES documentation](https://docs.aws.amazon.com/ses/latest/dg/receiving-email-verification.html) to verify your domain.
