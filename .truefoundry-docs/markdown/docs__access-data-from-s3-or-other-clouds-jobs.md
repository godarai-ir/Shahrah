> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Access Data From S3 Or Other Clouds

> Learn how to access data from S3 and other cloud storage in batch jobs on TrueFoundry.

In some instances, your Service may need to access data stored in S3 or other cloud storage platforms. To facilitate this access, you can employ one of two approaches:

#### Credential-Based Access through environment variables

This approach involves defining specific environment variables that contain the necessary credentials for accessing the cloud storage platform. For instance, to access S3, you would set environment variables for the AWS access key ID and secret access key, the environment variables being: `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`

<iframe href="https://app.supademo.com/embed/2JFenD-_FNcIg6hzTI1Zm?embed_v=2" typeofembed="iframe" height="475px" width="100%" provider="app.supademo.com" src="https://app.supademo.com/embed/2JFenD-_FNcIg6hzTI1Zm?embed_v=2" style={{ border: "none", display: "flex", margin: "auto" }} />

#### IAM Role-Based Access through Service Account

The second approach is to provide your Job with a Role with the necessary permission through Service Accounts.

Service Accounts provide a streamlined approach to managing access without the need for tokens or complex authentication methods. This approach involves creating a Principal IAM Role within your cloud platform, granting it the necessary permissions for your project's requirements. Here are detailed guides for creating Principal IAM Roles in your respective cloud platforms and integrating them as Service Accounts within the workspace:

* AWS: [Authenticate to AWS services using IAM service account](/docs/use-aws-services-using-iam-serviceaccount)
* GCP: [Authenticate to GCP using IAM serviceaccount](/docs/authenticate-to-gcp-using-iam-serviceaccount)

Once you've configured the Service Account within the workspace, you can simply toggle the **`Show Advanced Fields`** option at the bottom of the form. This will reveal an expanded set of options, from which you can select the desired Service Account using the provided dropdown menu.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/DdP_2rhue4AQQlob/images/387e68cd-b561def-Screenshot_2023-11-28_at_12.36.50_AM.png?fit=max&auto=format&n=DdP_2rhue4AQQlob&q=85&s=5e70e3e2084e10a4c542b9fdf42645d9" width="2350" height="284" data-path="images/387e68cd-b561def-Screenshot_2023-11-28_at_12.36.50_AM.png" />
</Frame>

***
