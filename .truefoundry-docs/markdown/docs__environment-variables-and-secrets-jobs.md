> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Environment Variables And Secrets

> Configure environment variables and secrets specifically for jobs.

Environment variables are key-value pairs stored outside of the code itself, accessible to the application during runtime. This approach effectively separates configuration from the code, offering several benefits:

* **Secure Storage of Sensitive Data**: Sensitive information, such as database credentials or API keys, is not embedded directly into the code, reducing the risk of exposure.
* **Ease of Configuration Management**: Configuration can be easily managed and adapted based on the environment (development, staging, production, etc.), without modifying the code.

### Example: Hardcoding Information vs. Using Environment Variables

Consider a scenario where you're training a machine learning model and need to store the resulting model in Amazon S3. In a traditional approach, you might hardcode the S3 access credentials directly into the code:

<CodeGroup>
  ```python Python lines theme={"dark"}
  import boto3
  import os

  # Hardcoded S3 credentials
  aws_access_key_id = "your-aws-access-key-id"
  aws_secret_access_key = "your-aws-secret-access-key"
  bucket_name = "your-bucket-name"

  # Train the machine learning model
  # ...

  # Create an S3 client using the hardcoded credentials
  s3_client = boto3.client('s3', aws_access_key_id, aws_secret_access_key)

  # Upload the trained model to S3
  s3_client.upload_file('trained_model.pkl', bucket_name, 'trained_model.pkl')
  ```
</CodeGroup>

However, if we hardcode the values, it will be difficult to manage across different environments since the S3 access credentials and bucket-name will be different across dev, staging and prod. To solve this issue, you can utilize environment variables to store and access these values securely.

Configure environment variables like `YOUR-AWS-ACCESS-KEY-ID`, `YOUR-AWS-SECRET-ACCESS-KEY`, and `YOUR-BUCKET-NAME` during deployment configuration, and retrieve them within your code using `os.environ(ENV_VAR_NAME)`

<CodeGroup>
  ```python Python lines theme={"dark"}
  import boto3
  import os

  # Hardcoded S3 credentials
  aws_access_key_id = os.environ["YOUR-AWS-ACCESS-KEY-ID"]
  aws_secret_access_key = os.environ["YOUR-AWS-SECRET-ACCESS-KEY"]
  bucket_name = os.environ["YOUR-BUCKET-NAME"]

  # Train the machine learning model
  # ...

  # Create an S3 client using the hardcoded credentials
  s3_client = boto3.client('s3', aws_access_key_id, aws_secret_access_key)

  # Upload the trained model to S3
  s3_client.upload_file('trained_model.pkl', bucket_name, 'trained_model.pkl')
  ```
</CodeGroup>

By separating configuration data from the code, you can seamlessly switch between different environments (production, development, etc.) without altering the code itself. Simply modify the environment variables during deployment configuration, allowing you to maintain a consistent codebase across different environments.

## Secrets: Enhanced Protection for Highly Sensitive Data

For sensitive keys like database passwords, api-keys, its not advisable to provide them directly as environment variables in the deployment configuration since everyone can see the deployment spec and get access to these sensitive passwords. Hence, we usually store such sensitive values in secret managers (AWS SSM, AWS Secret Manager, GCP Secret Manager, Azure Vault or Hashicorp Vault) and then only provide the key to the secret in the environment configuration. Translating the key to the actual value is done at a different step.

TrueFoundry provides an easy way to put the sensitive values in Secret Managers and then just provide the FQN of the secret (of the form `tfy-secret://user:my-secret-group:my-secret`) in the deployment spec. You can read about how to create the secrets [here](/docs/manage-secrets). TrueFoundry will automatically fetch the value and inject it into the environment at runtime.

## How to inject environment variables and secrets in TrueFoundry

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/9D3Z3_pyMKNOqYd-ghPPi" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/9D3Z3_pyMKNOqYd-ghPPi" style={{ border: "none", display: "flex", margin: "auto" }} />

***
