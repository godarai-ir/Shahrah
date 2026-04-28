> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Cost Monitoring

> Step-by-step guide for cost monitoring, explaining configuration, best practices, and real-world usage on TrueFoundry.

TrueFoundry integrates with Kubecost to provide cost visibility for your deployments. Kubecost is one of the applications that can be installed on the cluster. Once you install the application, you will start getting cost updates on the TrueFoundry portal. However, by default, Kubecost is not able to pull the spot instance prices - for this you need to setup Spot Datafeed collection on your AWS account. This guide provides a quick way to setup spot data-feed collection.

### AWS Spot Integration

1. Run the following bash script. Please replace `profile`, `region` and `cluster_name` before running

<CodeGroup>
  ```powershell shell lines theme={"dark"}
  # Enter your profile, region, and cluster name
  export profile="default"
  export region=""
  export cluster_name=""

  # (Do not edit) Declares some variables
  export bucket_name=$cluster_name-kubecost
  export account_id=$(aws sts get-caller-identity --query "Account" --output text --profile $profile)
  export oidc_provider=$(aws eks describe-cluster --name $cluster_name --region $region --query "cluster.identity.oidc.issuer" --output text --profile $profile | sed -e "s/^https:\/\///")
  export role_arn_name=$cluster_name-kubecost-role-arn
  export namespace=kubecost
  export service_account=kubecost-cost-analyzer


  # Create s3 bucket to store spot data feed
  aws s3api create-bucket \
      --bucket $bucket_name \
      --region $region \
      --object-ownership ObjectWriter \
      --profile $profile

  # Subscribe to spot data feed
  aws ec2 create-spot-datafeed-subscription \
      --bucket $bucket_name \
      --region $region \
      --profile $profile


  # Create IAM policy to access the s3 bucket
  ## Create policy.json
  cat >policy.json <<EOF
  {
    "Version": "2012-10-17",
    "Statement": [
      {
          "Sid": "SpotDataFeed",
          "Effect": "Allow",
          "Action": [
            "s3:ListAllMyBuckets",
            "s3:ListBucket",
            "s3:List*",
            "s3:Get*"
          ],
          "Resource": "arn:aws:s3:::$bucket_name*"
      }
    ]
  } 
  EOF

  ## Create policy
  aws iam create-policy \
      --policy-name SpotDataFeed \
      --policy-document file://policy.json \
      --profile $profile

  # Create IAM role to assume the role as service account
  ## Create trust-relationship.json
  cat >trust-relationship.json <<EOF
  {
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Principal": {
          "Federated": "arn:aws:iam::$account_id:oidc-provider/$oidc_provider"
        },
        "Action": "sts:AssumeRoleWithWebIdentity",
        "Condition": {
          "StringEquals": {
            "$oidc_provider:aud": "sts.amazonaws.com",
            "$oidc_provider:sub": "system:serviceaccount:$namespace:$service_account"
          }
        }
      }
    ]
  }
  EOF

  ## Create role
  aws iam create-role \
      --role-name $role_arn_name \
      --assume-role-policy-document file://trust-relationship.json \
      --profile $profile

  ## Attach policy to role
  aws iam attach-role-policy \
      --role-name $role_arn_name \
      --policy-arn=arn:aws:iam::$account_id:policy/SpotDataFeed \
      --profile $profile

  # Print the role ARN
  echo "Role ARN: arn:aws:iam::$account_id:role/$role_arn_name"
  ```
</CodeGroup>

2. Nagivate to `Workspaces` in TrueFoundry UI and edit `kubecost` workspace in your cluster.
3. Switch to`Show Advanced` options and add Service Account with name: `kubecost-cost-analyzer` and role arn value as `arn:aws:iam::<account_id>:role/<role-arn-name>`
4. Navigate to `Deployments > Helm` and edit `kubecost`. Then, add following section to values:

<CodeGroup>
  ```shell values lines theme={"dark"}
  kubecostProductConfigs:
    projectID: "<account_id>"
    awsSpotDataBucket: <bucket-name>
    awsSpotDataPrefix: ""
    awsSpotDataRegion: <region>
  ```
</CodeGroup>

### Azure Spot Integration

Coming soon...

### GCP Spot Integration

Coming soon...

***
