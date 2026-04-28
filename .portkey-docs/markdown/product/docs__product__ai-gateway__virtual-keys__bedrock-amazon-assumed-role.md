> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Connect Bedrock with Amazon Assumed Role

> How to create a new integration for Bedrock using Amazon Assumed Role Authentication

<Info>
  Available on all plans.
</Info>

<Card title="Set Up Bedrock Authentication for Enterprise" href="https://github.com/Portkey-AI/helm/blob/main/charts/portkey-gateway/docs/Bedrock.md">
  On the Enterprise plan and need to connect to Bedrock using an AWS assumed role? Check out the documentation here.
</Card>

## Select AWS Assumed Role Authentication

Create a new integration on Portkey, select **Bedrock** as the provider and **AWS Assumed Role** as the authentication method.

<Frame>
  <img src="https://mintcdn.com/portkey-docs/VWP2Y8zxPP5N4jE6/images/product/ai-gateway/Bedrock-Assumed-Role.png?fit=max&auto=format&n=VWP2Y8zxPP5N4jE6&q=85&s=ff849d28d7d11002c6566b8b0b53edc0" width="518" height="436" data-path="images/product/ai-gateway/Bedrock-Assumed-Role.png" />
</Frame>

## Create an AWS Role for Portkey to Assume

This role you create will be used by Portkey to execute InvokeModel commands on Bedrock models in your AWS account. The setup process will establish a minimal-permission ("least privilege") role and set it up to allow Portkey to assume this role.

### Create a permission policy in your AWS account using the following JSON

```json theme={"system"}
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "BedrockConsole",
      "Effect": "Allow",
      "Action": [
        "bedrock:InvokeModel",
        "bedrock:InvokeModelWithResponseStream"
        ],
      "Resource": "*"
    }
  ]
}
```

<Frame>
  <img src="https://mintcdn.com/portkey-docs/VWP2Y8zxPP5N4jE6/images/product/ai-gateway/create-policy.png?fit=max&auto=format&n=VWP2Y8zxPP5N4jE6&q=85&s=186a0227920b403714707763a43a30ca" width="1562" height="985" data-path="images/product/ai-gateway/create-policy.png" />
</Frame>

### Create a new IAM role

Choose *AWS account* as the trusted entity type. If you set an external ID be sure to copy it, we will need it later.

<Frame>
  <img src="https://mintcdn.com/portkey-docs/Buc1Vm2P31GSPm3S/images/product/ai-gateway/create-role.png?fit=max&auto=format&n=Buc1Vm2P31GSPm3S&q=85&s=434208a7ba3ff1b8ecb56712fa710f5e" width="1556" height="955" data-path="images/product/ai-gateway/create-role.png" />
</Frame>

### Add the above policy to the role

Search for the policy you created above and add it to the role.

<Frame>
  <img src="https://mintcdn.com/portkey-docs/VWP2Y8zxPP5N4jE6/images/product/ai-gateway/add-policy.png?fit=max&auto=format&n=VWP2Y8zxPP5N4jE6&q=85&s=39bf86a1ad2117ddafabe778b3c59008" width="1557" height="910" data-path="images/product/ai-gateway/add-policy.png" />
</Frame>

### Configure Trust Relationship for the role

Once the role is created, open the role and navigate to the *Trust relationships* tab and click *Edit trust policy*.
This is where you will add the Portkey AWS account as a trusted entity.

```sh Portkey Account ARN theme={"system"}
arn:aws:iam::299329113195:role/portkey-app
```

<Note>
  The above ARN only works for our [hosted app](https://app.portkey.ai/).<br />

  To enable **Assumed Role for AWS in your Portkey Enterprise deployment**, you can refer to [this guide](https://github.com/Portkey-AI/helm/blob/main/charts/portkey-gateway/docs/Bedrock.md). If you face any issue, please reach out to us at [support@portkey.ai](mailto:support@portkey.ai).
</Note>

Paste the following JSON into the trust policy editor and click *Update Trust Policy*.

```json theme={"system"}
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::299329113195:role/portkey-app"
      },
      "Action": "sts:AssumeRole",
      "Condition": {}
}
]
}
```

If you set an external ID, add it to the condition as shown below.

```json theme={"system"}
  {
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Principal": {
          "AWS": "<Portkey Account ARN>"
        },
        "Action": "sts:AssumeRole",
        "Condition": {
          "StringEquals": {
            "sts:ExternalId": "<your external ID>"
          }
        }
      }
    ]
}
```

## Configure the integration with the role ARN

Once the role is created, copy the role ARN and paste it into the Bedrock integrations modal in Portkey along with the external ID if you set one and the AWS region you are using.

<Frame>
  <img src="https://mintcdn.com/portkey-docs/VWP2Y8zxPP5N4jE6/images/product/ai-gateway/add-role-arn.png?fit=max&auto=format&n=VWP2Y8zxPP5N4jE6&q=85&s=52d14a3d58333ca95478e6fd75b7b59a" width="482" height="259" data-path="images/product/ai-gateway/add-role-arn.png" />
</Frame>

You're all set! You can now use the providers inherited from your integration to invoke Bedrock models.
