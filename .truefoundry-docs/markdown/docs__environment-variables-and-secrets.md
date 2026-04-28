> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Environment Variables And Secrets

> Manage environment variables and secrets securely across services.

Environment variables are key-value pairs stored outside of the code itself, accessible to the application during runtime. This approach effectively separates configuration from the code, offering several benefits:

* **Secure Storage of Sensitive Data**: Sensitive information, such as database credentials or API keys, is not embedded directly into the code, reducing the risk of exposure.
* **Ease of Configuration Management**: Configuration can be easily managed and adapted based on the environment (development, staging, production, etc.), without modifying the code.

### Example: Hardcoding Information vs. Using Environment Variables

For instance, consider accessing database from a Service in a traditional approach. You might hardcode the connection details directly into the code:

<CodeGroup>
  ```python Python lines theme={"dark"}
  import psycopg2

  # Connect to the database
  conn = psycopg2.connect(
    database="your-database-name",
    user="your-username",
    password="your-password"
  )

  # Retrieve the secret value from the database
  cursor = conn.cursor()
  cursor.execute("SELECT some_value FROM table")
  value = cursor.fetchone()[0]
  ```
</CodeGroup>

However, if we hardcode the values, it will be difficult to manage across different environments since the database credentials will be different across dev, staging and prod. To solve this issue, you can utilize environment variables to store and access these values securely.

Configure environment variables like `YOUR-DATABASE-URL`, `YOUR-USERNAME`, and `YOUR-PASSWORD` during deployment configuration, and retrieve them within your code using `os.environ(ENV_VAR_NAME)`

<CodeGroup>
  ```python Python lines theme={"dark"}
  import os

  # Retrieve the secret value from the environment variable
  database_url = os.environ['YOUR-DATABASE-URL']
  database_url = os.environ['YOUR-USERNAME']
  database_url = os.environ['YOUR-PASSWORD']

  # Use the secret value to connect to the database
  conn = psycopg2.connect(database_url)

  # Retrieve the value from the database
  cursor = conn.cursor()
  cursor.execute("SELECT some_value FROM table")
  value = cursor.fetchone()[0]
  ```
</CodeGroup>

By separating configuration data from the code, you can seamlessly switch between different environments (production, development, etc.) without altering the code itself. Simply modify the environment variables during deployment configuration, allowing you to maintain a consistent codebase across different environments.

## Secrets: Enhanced Protection for Highly Sensitive Data

For sensitive keys like database passwords, api-keys, its not advisable to provide them directly as environment variables in the deployment configuration since everyone can see the deployment spec and get access to these sensitive passwords. Hence, we usually store such sensitive values in secret managers (AWS SSM, AWS Secret Manager, GCP Secret Manager, Azure Vault or Hashicorp Vault) and then only provide the key to the secret in the environment configuration. Translating the key to the actual value is done at a different step.

TrueFoundry provides an easy way to put the sensitive values in Secret Managers and then just provide the FQN of the secret (of the form `tfy-secret://user:my-secret-group:my-secret`) in the deployment spec. You can read about how to create the secrets [here](/docs/manage-secrets) . TrueFoundry will automatically fetch the value and inject it into the environment at runtime.

## Setting environment variables and secrets

You can directly provide the value of the environment variable in the `Value` field if its not sensitive. If the value is sensitive, you can first create the secret and then provide the FQN of the secret in the `Value` field.

<Note>
  To search for secrets, start typing `tfy` in the Value field and a dropdown will appear with all the secrets that you have access to.
</Note>

<iframe href="https://app.supademo.com/embed/a2Wk1NyiiAbYE8B0wjP08" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/a2Wk1NyiiAbYE8B0wjP08" style={{ border: "none", display: "flex", margin: "auto" }} />

<Tip>If you want to paste a lot of environment variables in one shot, you can do so by switching to Raw Mode and pasting the content of .env file. </Tip>

<img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/env-raw-values.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=aa83e3af7b859b7eba983e27db9821a1" width="3020" height="1712" data-path="images/env-raw-values.png" />
