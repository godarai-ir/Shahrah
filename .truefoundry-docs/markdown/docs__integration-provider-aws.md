> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# AWS

> Integrate AWS services with TrueFoundry. Configure access and resources.

TrueFoundry supports integrating with multiple AWS services like S3, ECR, SSM, EKS etc.

To integrate any of the above services, you simply need to add your AWS account as a provider account and add integrations in the `Integration Providers` section in the `Platform` page.

### Share access with users, teams or everyone in your TrueFoundry account

As shown in the previous slides, you can share access of each integration with users, teams or everyone in your TrueFoundry account. This would allow them to view and use the integration. Only tenant-admins can edit the integrations.

## Generate Access Key or Assumed Role

The document below will guide you on how to create an IAM role with assume role and add the required permissions to that role. If you have used the TrueFoundry OpenTofu/Terraform code, an IAM role will already be created and attached here. You can also create a new one to allow for specific permissions if needed.

You can create an IAM user with the required permissions and generate an access key and secret key to add integration. For AWS, TrueFoundry supports both IAM user and IAM role.

### Create an IAM role with assume role

The role should have following trust policy added.

<CodeGroup>
  ```json json lines theme={"dark"}
  {
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Principal": {
          "AWS": "arn:aws:iam::416964291864:role/tfy-ctl-euwe1-production-truefoundry-deps"
        },
        "Action": "sts:AssumeRole",
        "Condition": {}
      }
    ]
  }
  ```
</CodeGroup>

If you plan to use IAM role for workflows, make sure to add the following trust relationship to it so that workflow propeller can assume this role to access the storage bucket.

<CodeGroup>
  ```json JSON lines theme={"dark"}
  {
      "Effect": "Allow",
      "Principal": {
          "Federated": "arn:aws:iam::<ACCOUNT_ID>:oidc-provider/oidc.eks.<AWS_REGION>.amazonaws.com/id/<OIDC_ID>"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
          "StringEquals": {
              "oidc.eks.<AWS_REGION>.amazonaws.com/id/<OIDC_ID>:sub": "system:serviceaccount:tfy-workflow-propeller:flytepropeller",
              "oidc.eks.<AWS_REGION>.amazonaws.com/id/<OIDC_ID>:aud": "sts.amazonaws.com"
          }
      }
  }
  ```
</CodeGroup>

To create this IAM role, you must first save this trust policy in your local system as a JSON file. You can then use the following commands to create the role.

<CodeGroup>
  ```shell Shell lines theme={"dark"}
  aws iam create-role --role-name <ROLE_NAME> --assume-role-policy-document file://<TRUST_POLICY_FILE.json>

  ```
</CodeGroup>

## Following steps are required for each integration

You can also attach the following policies to the IAM role/user for each integration. Ensure that you replace the `AWS_REGION` and `ACCOUNT_ID` with your own.

<AccordionGroup>
  <Accordion title="S3 integration">
    **\[Pre-requisite] Create a [S3 Bucket](https://docs.aws.amazon.com/AmazonS3/latest/userguide/creating-bucket.html) with following config**

    1. Make sure the bucket has lifecycle configuration to [abort multipart upload](https://docs.aws.amazon.com/AmazonS3/latest/userguide/mpu-abort-incomplete-mpu-lifecycle-config.html) set for 7 days.
    2. Make sure [CORS](https://docs.aws.amazon.com/AmazonS3/latest/userguide/enabling-cors-examples.html) is applied on the bucket with the below configuration:

    <CodeGroup>
      ```json JSON lines theme={"dark"}
      [
        {
          "AllowedHeaders": ["*"],
          "AllowedMethods": ["GET", "POST", "PUT"],
          "AllowedOrigins": ["*"],
          "ExposeHeaders": ["ETag"],
          "MaxAgeSeconds": 3000
        }
      ]
      ```
    </CodeGroup>

    **Required Policies**

    <CodeGroup>
      ```json JSON lines theme={"dark"}
      {
        "Sid": "S3",
        "Effect": "Allow",
        "Action": ["s3:*"],
        "Resource": [
          "arn:aws:s3:::<YOUR_S3_BUCKET_NAME>",
          "arn:aws:s3:::<YOUR_S3_BUCKET_NAME>/*"
        ]
      }
      ```
    </CodeGroup>
  </Accordion>

  <Accordion title="ECR integration">
    Attach the following policy to the IAM role/user for integrating ECR.

    <CodeGroup>
      ```json JSON lines theme={"dark"}
      [
        {
          "Sid": "ECR",
          "Effect": "Allow",
          "Action": [
            "ecr:GetRegistryPolicy",
            "ecr:DescribeImageScanFindings",
            "ecr:GetLifecyclePolicyPreview",
            "ecr:CreateRepository",
            "ecr:GetDownloadUrlForLayer",
            "ecr:DescribeImageReplicationStatus",
            "ecr:ListTagsForResource",
            "ecr:BatchGetRepositoryScanningConfiguration",
            "ecr:GetRegistryScanningConfiguration",
            "ecr:PutImage",
            "ecr:BatchGetImage",
            "ecr:DescribeRepositories",
            "ecr:BatchCheckLayerAvailability",
            "ecr:GetRepositoryPolicy",
            "ecr:GetLifecyclePolicy",
            "ecr:ListImages",
            "ecr:InitiateLayerUpload",
            "ecr:CompleteLayerUpload",
            "ecr:DescribeImages",
            "ecr:DeleteRepository",
            "ecr:UploadLayerPart"
          ],
          "Resource": ["arn:aws:ecr:AWS_REGION:ACCOUNT_ID:repository/tfy-*"]
        },
        {
          "Sid": "ECR",
          "Effect": "Allow",
          "Action": [
            "ecr:DescribeRegistry",
            "ecr:GetAuthorizationToken",
            "sts:GetServiceBearerToken"
          ],
          "Resource": ["*"]
        }
      ]
      ```
    </CodeGroup>
  </Accordion>

  <Accordion title="SSM integration">
    Attach the following policy to the IAM role/user for integrating parameter store.

    <CodeGroup>
      ```json JSON lines theme={"dark"}
      {
        "Sid": "SSM",
        "Effect": "Allow",
        "Action": [
          "ssm:GetParameter",
          "ssm:GetParameters",
          "ssm:PutParameter",
          "ssm:DeleteParameter",
          "ssm:DeleteParameters",
          "ssm:GetParameterHistory"
        ],
        "Resource": ["arn:aws:ssm:AWS_REGION:ACCOUNT_ID:parameter/tfy-secret/*"]
      }
      ```
    </CodeGroup>
  </Accordion>

  <Accordion title="AWS Secrets Manager integration (Optional)">
    Attach the following policy to the IAM role/user for integrating AWS Secrets Manager.

    <CodeGroup>
      ```json JSON lines theme={"dark"}
      {
        "Sid": "SecretsManager",
        "Effect": "Allow",
        "Action": [
          "secretsmanager:GetSecretValue",
          "secretsmanager:DescribeSecret",
          "secretsmanager:CreateSecret",
          "secretsmanager:DeleteSecret",
          "secretsmanager:UpdateSecret",
          "secretsmanager:ListSecrets",
          "secretsmanager:PutSecretValue"
        ],
        "Resource": ["arn:aws:secretsmanager:AWS_REGION:ACCOUNT_ID:secret:/tfy-secret/*"]
      }
      ```
    </CodeGroup>
  </Accordion>

  <Accordion title="EKS integration">
    Attach the following policy to the IAM role/user for integrating EKS.

    <CodeGroup>
      ```json JSON lines theme={"dark"}
      {
          "Statement": [
              {
                  "Action": [
                      "eks:ListUpdates",
                      "eks:ListTagsForResource",
                      "eks:ListPodIdentityAssociations",
                      "eks:ListNodegroups",
                      "eks:ListInsights",
                      "eks:ListFargateProfiles",
                      "eks:ListAddons",
                      "eks:DescribeUpdate",
                      "eks:DescribePodIdentityAssociation",
                      "eks:DescribeNodegroup",
                      "eks:DescribeInsight",
                      "eks:DescribeFargateProfile",
                      "eks:DescribeCluster",
                      "eks:DescribeAddon",
                      "eks:AccessKubernetesApi",
                      "eks:UpdateNodegroupConfig"
                  ],
                  "Effect": "Allow",
                  "Resource": [
                      "arn:aws:eks:AWS_REGION:AWS_ACCOUNT_ID:podidentityassociation/CLUSTER_NAME/*",
                      "arn:aws:eks:AWS_REGION:AWS_ACCOUNT_ID:nodegroup/CLUSTER_NAME/*/*",
                      "arn:aws:eks:AWS_REGION:AWS_ACCOUNT_ID:identityproviderconfig/CLUSTER_NAME/*/*/*",
                      "arn:aws:eks:AWS_REGION:AWS_ACCOUNT_ID:fargateprofile/CLUSTER_NAME/*/*",
                      "arn:aws:eks:AWS_REGION:AWS_ACCOUNT_ID:cluster/CLUSTER_NAME",
                      "arn:aws:eks:AWS_REGION:AWS_ACCOUNT_ID:addon/CLUSTER_NAME/*/*",
                      "arn:aws:eks:AWS_REGION:AWS_ACCOUNT_ID:dashboard/*",
                      "arn:aws:eks:AWS_REGION:AWS_ACCOUNT_ID:eks-anywhere-subscription/*",
                      "arn:aws:eks:AWS_REGION:AWS_ACCOUNT_ID:capability/CLUSTER_NAME/*/*/*"
                  ]
              },
              {
                  "Action": [
                      "eks:ListClusters",
                      "eks:DescribeAddonVersions",
                      "eks:DescribeAddonConfiguration",
                      "ec2:DescribeRegions",
                      "eks:CreateAccessEntry",
                      "eks:DeleteAccessEntry",
                      "eks:AssociateAccessPolicy",
                      "eks:DisassociateAccessPolicy"
                  ],
                  "Effect": "Allow",
                  "Resource": "*"
              }
          ],
          "Version": "2012-10-17"
      }
      ```
    </CodeGroup>
  </Accordion>

  <Accordion title="Bedrock integration (Optional)">
    The following policy grants permission to invoke any foundation model available on Bedrock in your available regions (To check the list of available regions for different models, refer to [AWS Bedrock](https://docs.aws.amazon.com/bedrock/latest/userguide/inference-profiles-support.html)). You can configure `Resource` list to control which models can be accessed through the AI Gateway.

    <CodeGroup>
      ```json JSON lines theme={"dark"}
      {
        "Version": "2012-10-17",
        "Statement": [
          {
            "Effect": "Allow",
            "Sid": "InvokeAllModels",
            "Action": [
              "bedrock:InvokeModel",
              "bedrock:InvokeModelWithResponseStream"
            ],
            "Resource": ["arn:aws:bedrock:*::foundation-model/*"]
          }
        ]
      }
      ```
    </CodeGroup>
  </Accordion>
</AccordionGroup>

***
