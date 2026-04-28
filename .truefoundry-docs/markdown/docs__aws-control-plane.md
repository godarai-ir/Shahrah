> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# AWS

> This page provides an architecture overview, requirements and steps to setup a TrueFoundry control plane cluster in AWS.

The architecture of a TrueFoundry compute plane is as follows:

<Note>
  <p>
    This section is coming soon.
  </p>
</Note>

<Accordion title="Access Policies Overview">
  | Policy                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      | Description                                                                                                                                                                                                                                                                      |
  | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
  | [ELBControllerPolicy](https://github.com/terraform-aws-modules/terraform-aws-iam/blob/f37809108f86d8fbdf17f735df734bf4abe69315/modules/iam-role-for-service-accounts-eks/policies.tf#L752)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  | Role assumed by load balancer controller to provision ELB when a service of type LoadBalancer is created                                                                                                                                                                         |
  | [KarpenterPolicy](https://github.com/terraform-aws-modules/terraform-aws-iam/blob/f37809108f86d8fbdf17f735df734bf4abe69315/modules/iam-role-for-service-accounts-eks/policies.tf#L619) and [SQSPolicy](https://github.com/truefoundry/terraform-aws-truefoundry-karpenter/blob/76f04a4c4abf1dc22e97b80994015c8e2e77d06f/sqs.tf#L8)                                                                                                                                                                                                                                                                                                                                                                                                                                                                          | Role assumed by Karpenter to dynamically provision nodes and handle spot node termination                                                                                                                                                                                        |
  | [EFSPolicy](https://github.com/truefoundry/terraform-aws-truefoundry-efs/blob/574d6680e1fc0ff4cff8c0f4df507c6f13659f10/efs.tf#L8)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | Role assumed by EFS CSI to provision and attach EFS volumes                                                                                                                                                                                                                      |
  | [EBSPolicy](https://docs.aws.amazon.com/aws-managed-policy/latest/reference/AmazonEBSCSIDriverPolicy.html)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  | Role assumed by EBS CSI to provision and attach EBS volumes                                                                                                                                                                                                                      |
  | RolePolicy with policies for:- [ECR](https://github.com/truefoundry/terraform-aws-truefoundry-platform-features/blob/98bb8c09e5760dd5c2d557e27b7a94e2056da266/iam.tf#L57-L98), [S3](https://github.com/truefoundry/terraform-aws-truefoundry-platform-features/blob/98bb8c09e5760dd5c2d557e27b7a94e2056da266/iam.tf#L5-L18), [SSM](https://github.com/truefoundry/terraform-aws-truefoundry-platform-features/blob/98bb8c09e5760dd5c2d557e27b7a94e2056da266/iam.tf#L20-L36), [EKS](https://github.com/truefoundry/terraform-aws-truefoundry-platform-features/blob/98bb8c09e5760dd5c2d557e27b7a94e2056da266/iam.tf#L100-L144)<br />Use the [trust relationship](https://github.com/truefoundry/terraform-aws-truefoundry-platform-features/blob/98bb8c09e5760dd5c2d557e27b7a94e2056da266/iam.tf#L192-L229). | Role assumed by TrueFoundry to allow access to ECR, S3, and SSM services. If you are using TrueFoundry's control plane the role will be assumed by `arn:aws:iam::416964291864:role/tfy-ctl-euwe1-production-truefoundry-deps` otherwise it will be your control plane's IAM role |
  | ClusterRole with policies:<br />- [AmazonEKSClusterPolicy](https://docs.aws.amazon.com/aws-managed-policy/latest/reference/AmazonEKSClusterPolicy.html)<br />- [AmazonEKSVPCResourceControllerPolicy](https://docs.aws.amazon.com/aws-managed-policy/latest/reference/AmazonEKSVPCResourceController.html)<br />- EncryptionPolicy                                                                                                                                                                                                                                                                                                                                                                                                                                                                          | Role that provides Kubernetes permissions to manage the cluster lifecycle, networking, and encryption                                                                                                                                                                            |
  | NodeRole with policies: [AmazonEC2ContainerRegistryReadOnlyPolicy](https://docs.aws.amazon.com/aws-managed-policy/latest/reference/AmazonEC2ContainerRegistryReadOnly.html), [AmazonEKS\_CNI\_Policy](https://docs.aws.amazon.com/aws-managed-policy/latest/reference/AmazonEKS_CNI_Policy.html), [AmazonEKSWorkerNodePolicy](https://docs.aws.amazon.com/aws-managed-policy/latest/reference/AmazonEKSWorkerNodePolicy.html),   [AmazonSSMManagedInstanceCorePolicy](https://docs.aws.amazon.com/aws-managed-policy/latest/reference/AmazonSSMManagedInstanceCore.html)                                                                                                                                                                                                                                    | Role assumed by EKS nodes to work with AWS resources for ECR access, IP assignment, and cluster registration                                                                                                                                                                     |

  EncryptionPolicy to create and manage key for encryption:

  ```json lines theme={"dark"}
  {  
      "Statement": [  
          {  
              "Action": [  
                  "kms:Encrypt",  
                  "kms:Decrypt",  
                  "kms:ListGrants",  
                  "kms:DescribeKey"  
              ],  
              "Effect": "Allow",  
              "Resource": "arn:aws:kms:<region>:<aws_account_id>:key/<key_id>"  
          }  
      ],  
      "Version": "2012-10-17"  
  }
  ```
</Accordion>

Setting up TrueFoundry control plane on your own cloud involves creating the infrastructure to support the platform and then installing the platform itself.

## Setting up Infrastructure

## Requirements:

The requirements to setup control plane in each of the scenarios is as follows:

* <Icon icon="square-check" iconType="regular" /> Billing and [STS](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_temp_enable-regions.html#sts-regions-activate-deactivate) must be enabled for the AWS account.
* <Icon icon="square-check" iconType="regular" /> Please make sure you have enough quotas for GPU/Inferentia instances on the account depending on your usecase. You can check and increase quotas at [AWS EC2 service quotas](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-resource-limits.html)
* <Icon icon="square-check" iconType="regular" /> Please make sure you have created a certifcate for your domain in AWS Certificate Manager (ACM) and have the ARN of the certificate ready. This is required to setup TLS for the load balancer.
* <Icon icon="square-check" iconType="regular" /> Postgres database with the following requirements:
  * Version: >= 13
  * Instance Types: `db.t3.medium` or `db.t4g.medium`
  * Storage: 20GB of type `gp3` with autoscale enabled to 30GB
  * Encryption: Enabled
  * **For PostgreSQL 17+:** Set `force_ssl` parameter to `0` (off) in parameter group if you need to allow non-SSL connections (default is `1`)
  * **Security Group:** Ensure RDS security group allows inbound traffic from EKS node security groups
* <Icon icon="square-check" iconType="regular" /> S3 bucket to store the intermediate code while building the docker image.
* <Icon icon="square-check" iconType="regular" /> Egress Access to TrueFoundry:
  * [https://auth.truefoundry.com](https://auth.truefoundry.com) - Central Authentication Server for licensing and authentication
  * [https://login.truefoundry.com](https://login.truefoundry.com) - Login UI for the central authentication server
  * [https://catalogue.truefoundry.com](https://catalogue.truefoundry.com) - Central Repository for fetching catalogues for latest model, their public cost, mcp servers, etc.
  * [https://analytics.truefoundry.com](https://analytics.truefoundry.com) - Analytics Server for sending usage analytics to TrueFoundry.
* <Icon icon="square-check" iconType="regular" /> DNS: Domain for control plane and service endpoints. One endpoint to point to the control plane service (e.g., platform.example.com) and the other to point to the compute plane service (e.g., tfy.example.com/service1). **The control-plane URL must be reachable from the compute-plane**. The developers will need to access the TrueFoundry UI at the provided domain.
* <Icon icon="square-check" iconType="regular" /> We will need a certificate ARN (for the domain provided above) to attach to the loadbalancer so as to terminate TLS traffic at the load balancer. This will allow the services we deploy on the cluster to be accessed via HTTPS. We recommend using [AWS Certificate Manager](/docs/add-certificate-for-tls#aws) to add TLS to the load balancer. You can read the instructions in Step 2 below on how to create the certificate in AWS Certificate Manager.
* <Icon icon="square-check" iconType="regular" /> You need to have enough permissions on the AWS account to create the resources needed for the compute plane. Check [this](#permissions-required-to-create-the-infrastructure) for more details. We usually recommend admin permission on the AWS account, but if you need the exact set of fine-grained permissions, you can check the list of permissions below:

<CodeGroup>
  ```json json lines expandable theme={"dark"}
  {
    "Version": "2012-10-17",
    "Statement": [
      {
        "Sid": "ClusterScopedCloudFormationIamProfilesAndRds",
        "Effect": "Allow",
        "Action": [
          "cloudformation:*",
          "iam:AddRoleToInstanceProfile",
          "iam:CreateInstanceProfile",
          "iam:DeleteInstanceProfile",
          "iam:GetInstanceProfile",
          "iam:RemoveRoleFromInstanceProfile",
          "iam:TagInstanceProfile",
          "rds:AddTagsToResource",
          "rds:CreateDBInstance",
          "rds:CreateTenantDatabase",
          "rds:DeleteDBInstance",
          "rds:DeleteTenantDatabase",
          "rds:DescribeDBInstances",
          "rds:RemoveTagsFromResource"
        ],
        "Resource": [
          "arn:aws:iam::${AWS_ACCOUNT_ID}:instance-profile/*",
          "arn:aws:cloudformation:*:${AWS_ACCOUNT_ID}:stackset/${CLUSTER_NAME}*:*",
          "arn:aws:cloudformation:*:${AWS_ACCOUNT_ID}:stack/${CLUSTER_NAME}*/*",
          "arn:aws:rds:*:${AWS_ACCOUNT_ID}:db:${CLUSTER_NAME}*"
        ]
      },
      {
        "Sid": "OidcProvidersAndRdsSubnetGroups",
        "Effect": "Allow",
        "Action": [
          "iam:CreateOpenIDConnectProvider",
          "iam:DeleteOpenIDConnectProvider",
          "iam:GetOpenIDConnectProvider",
          "iam:TagOpenIDConnectProvider",
          "rds:AddTagsToResource",
          "rds:CreateDBInstance",
          "rds:CreateDBSubnetGroup",
          "rds:DeleteDBInstance",
          "rds:DeleteDBSubnetGroup",
          "rds:DescribeDBSubnetGroups",
          "rds:ListTagsForResource",
          "rds:RemoveTagsFromResource"
        ],
        "Resource": [
          "arn:aws:rds:*:${AWS_ACCOUNT_ID}:subgrp:${CLUSTER_NAME}*",
          "arn:aws:iam::${AWS_ACCOUNT_ID}:oidc-provider/*"
        ]
      },
      {
        "Sid": "IamPoliciesAndRdsDbParameterManagement",
        "Effect": "Allow",
        "Action": [
          "iam:CreatePolicy",
          "iam:DeletePolicy",
          "iam:GetPolicy",
          "iam:GetPolicyVersion",
          "iam:ListPolicyVersions",
          "iam:TagPolicy",
          "rds:AddTagsToResource",
          "rds:CopyDBParameterGroup",
          "rds:CreateDBCluster",
          "rds:CreateDBClusterEndpoint",
          "rds:CreateDBClusterParameterGroup",
          "rds:CreateDBClusterSnapshot",
          "rds:CreateDBInstance",
          "rds:CreateDBInstanceReadReplica",
          "rds:CreateDBParameterGroup",
          "rds:CreateDBProxy",
          "rds:CreateDBProxyEndpoint",
          "rds:CreateDBSecurityGroup",
          "rds:CreateDBShardGroup",
          "rds:CreateDBSnapshot",
          "rds:CreateDBSubnetGroup",
          "rds:CreateEventSubscription",
          "rds:DeleteDBParameterGroup",
          "rds:DescribeDBInstances",
          "rds:DescribeDBParameterGroups",
          "rds:DescribeDBParameters",
          "rds:ListTagsForResource",
          "rds:ModifyDBParameterGroup",
          "rds:RemoveTagsFromResource",
          "rds:ResetDBParameterGroup"
        ],
        "Resource": [
          "arn:aws:iam::${AWS_ACCOUNT_ID}:policy/tfy-*",
          "arn:aws:iam::${AWS_ACCOUNT_ID}:policy/${CLUSTER_NAME}-*",
          "arn:aws:iam::${AWS_ACCOUNT_ID}:policy/AmazonEKS_Karpenter_Controller_Policy*",
          "arn:aws:iam::${AWS_ACCOUNT_ID}:policy/AmazonEKS_CNI_Policy*",
          "arn:aws:iam::${AWS_ACCOUNT_ID}:policy/AmazonEKS_AWS_Load_Balancer_Controller*",
          "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryFullAccess",
          "arn:aws:rds:*:${AWS_ACCOUNT_ID}:db:*",
          "arn:aws:rds:*:${AWS_ACCOUNT_ID}:pg:*"
        ]
      },
      {
        "Sid": "GlobalInfraAndCloudFormationSupportActions",
        "Effect": "Allow",
        "Action": [
          "cloudformation:*",
          "ec2:*",
          "eks:*",
          "elasticfilesystem:*",
          "iam:GetRole",
          "iam:ListPolicies",
          "kms:*",
          "rds:DescribeEngineDefaultClusterParameters",
          "rds:DescribeEngineDefaultParameters",
          "route53:AssociateVPCWithHostedZone",
          "s3:ListAllMyBuckets",
          "s3:ListBucket",
          "sts:GetCallerIdentity"
        ],
        "Resource": "*"
      },
      {
        "Sid": "ClusterScopedIamRoles",
        "Effect": "Allow",
        "Action": "iam:*",
        "Resource": "arn:aws:iam::${AWS_ACCOUNT_ID}:role/${CLUSTER_NAME}*"
      },
      {
        "Sid": "ClusterAndTruefoundryS3Access",
        "Effect": "Allow",
        "Action": "s3:*",
        "Resource": [
          "arn:aws:s3:::${CLUSTER_NAME}*",
          "arn:aws:s3:::${CLUSTER_NAME}*/*",
          "arn:aws:s3:::truefoundry*",
          "arn:aws:s3:::truefoundry*/*"
        ]
      },
      {
        "Sid": "ClusterEventsSqsAndLogs",
        "Effect": "Allow",
        "Action": [
          "events:*",
          "logs:*",
          "sqs:*"
        ],
        "Resource": [
          "arn:aws:events:*:${AWS_ACCOUNT_ID}:rule/${CLUSTER_NAME}*",
          "arn:aws:sqs:*:${AWS_ACCOUNT_ID}:${CLUSTER_NAME}*",
          "arn:aws:logs:*:${AWS_ACCOUNT_ID}:log-group:*",
          "arn:aws:logs:*:${AWS_ACCOUNT_ID}:log-group:*:log-stream:*"
        ]
      }
    ]
  }
  ```
</CodeGroup>

Regarding the VPC and EKS cluster, you can decide between the following scenarios:

<Tabs>
  <Tab title="New VPC and New EKS Cluster">
    1. <Icon icon="square-check" iconType="regular" /> The new VPC should will have a CIDR range of /20 or larger, at least 2 availability zones and private subnets with CIDR `/24` or larger. This is to ensure capacity for \~250 instances and 4096 pods.
    2. <Icon icon="square-check" iconType="regular" /> If you want to use a smaller network range for your EKS cluster, TrueFoundry supports [EKS custom networking](https://docs.aws.amazon.com/eks/latest/userguide/cni-custom-network.html) as well.
    3. <Icon icon="square-check" iconType="regular" /> A NAT gateway will be provisioned to provide internet access to the private subnets.
    4. <Icon icon="square-check" iconType="regular" /> We should have egress access to `public.ecr.aws`, `quay.io`, `ghcr.io`, `tfy.jfrog.io`, `docker.io/natsio`, `nvcr.io`, `registry.k8s.io` so that we can download the docker images for argocd, nats, gpu operator, argo rollouts, argo workflows, istio, keda, etc.
  </Tab>

  <Tab title="Existing VPC and New EKS Cluster">
    1. <Icon icon="square-check" iconType="regular" /> The existing VPC should have min 2 private subnets in different AZs with CIDR /24. This ensures capacity for \~250 instances and 4096 pods. The VPC should have NAT gateway for private subnets. If you want to use a smaller network range for your EKS cluster, TrueFoundry supports [EKS custom networking](https://docs.aws.amazon.com/eks/latest/userguide/cni-custom-network.html) as well.
    2. <Icon icon="square-check" iconType="regular" /> If you want to have a load balancer in the public subnet there should be atleast one public subnet in the existing VPC with min CIDR range of /28.
    3. <Icon icon="square-check" iconType="regular" /> The VPC should have [Auto-assign IP address](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-launch-instance-wizard.html#liw-network-settings:~:text=Auto%2Dassign%20Public,IPv4%20addresses.) enabled. It should also have [DNS support](https://docs.aws.amazon.com/glue/latest/dg/set-up-vpc-dns.html) and [DNS hostnames](https://docs.aws.amazon.com/glue/latest/dg/set-up-vpc-dns.html) enabled.

    Your subnets must have the following tags for the TrueFoundry OpenTofu/Terraform code to work with them.

    | Resource Type           | Required Tags                                                                                                                     | Description                                                                                 |
    | ----------------------- | --------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------- |
    | Private Subnets         | - `kubernetes.io/cluster/${clusterName}`: `"shared"`<br />- `subnet`: `"private"`<br />- `kubernetes.io/role/internal-elb`: `"1"` | Tags required for EKS to properly manage internal load balancers and subnet identification  |
    | Public Subnets          | - `kubernetes.io/cluster/${clusterName}`: `"shared"`<br />- `subnet`: `"public"`<br />- `kubernetes.io/role/elb`: `"1"`           | Tags required for EKS to properly manage external load balancers and subnet identification  |
    | EKS Node Security Group | - `karpenter.sh/discovery`: `"${clusterName}"`                                                                                    | This tag is required for Karpenter to discover and manage node provisioning for the cluster |
  </Tab>

  <Tab title="Existing EKS Cluster">
    1. <Icon icon="square-check" iconType="regular" /> EKS Version should be 1.30 or higher with [IRSA](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html) enabled.
    2. <Icon icon="square-check" iconType="regular" /> EBS CSI Driver should be installed [Installation Guide](https://github.com/kubernetes-sigs/aws-ebs-csi-driver/blob/master/docs/install.md) - Required for persistent volume support for Notebooks, SSH. If not installed TrueFoundry will install it for you.
    3. <Icon icon="square-check" iconType="regular" /> EFS CSI Driver should be installed [Installation Guide](https://github.com/kubernetes-sigs/aws-efs-csi-driver/blob/master/docs/README.md) - Required for model and data caching. If not installed TrueFoundry will install it for you.
    4. <Icon icon="square-check" iconType="regular" /> AWS Load Balancer Controller (>=v2.12.0) should be installed [Installation Guide](https://kubernetes-sigs.github.io/aws-load-balancer-controller/latest/deploy/installation/) - Required for Ingress and Service type LoadBalancer support. If not installed TrueFoundry will install it for you.
    5. <Icon icon="square-check" iconType="regular" /> Although this is not compulsory, we highly recommend Karpenter to be installed on the cluster. It makes a lot of functionalities in TrueFoundry easier, faster and cost-effective. If not installed TrueFoundry will install it for you.
  </Tab>
</Tabs>

## Setting up control plane

TrueFoundry control plane infrastructure is provisioned using OpenTofu/Terraform. You can download the OpenTofu/Terraform code for your exact account by filling up your account details and downloading a script that can be executed on your local machine. To perform the below steps, you need to [register](https://www.truefoundry.com/register) an account on TrueFoundry and login to the platform.

<Steps>
  <Step title="Enable Deployment Feature in the Platform (Optional)">
    To enable the deployment feature which allows you to deploy services through the platform, you need to enable it;

    * In the left hand navigation, go to `Settings` then `Platform Feature Visibility` under `Preferences`
    * Click on `Edit` button. Then enable the toggle for `Enable Deployment`

    <img src="https://mintcdn.com/truefoundry/bWzUilIOzt9sRNdU/images/docs/platform/enable-deployment.png?fit=max&auto=format&n=bWzUilIOzt9sRNdU&q=85&s=4932c230f6d6a6b969ed3d83c942be2b" width="1510" height="408" data-path="images/docs/platform/enable-deployment.png" />

    * Click on `Save` button.

    This will enable the deployment feature in the platform and allow you to create either a control plane and compute plane.

    <img src="https://mintcdn.com/truefoundry/bWzUilIOzt9sRNdU/images/docs/platform/deployment-platform.png?fit=max&auto=format&n=bWzUilIOzt9sRNdU&q=85&s=71e7b321682305cce46f6105c61a6eab" width="1511" height="647" data-path="images/docs/platform/deployment-platform.png" />
  </Step>

  <Step title="Choose to create a new cluster or attach an existing cluster">
    Go to the platform section in the left panel and click on `Clusters`. Add the following value at the end of your URL `&controlPlaneSetupEnabled=true`. This will enable the control plane installation for you. You can click on `Create New Cluster` or `Attach Existing Cluster` depending on your use case. Read the requirements and if everything is satisfied, click on `Continue`.

    <img src="https://mintcdn.com/truefoundry/IHwtCTyIC4s5BfZO/images/docs/platform/create-control-plane.png?fit=max&auto=format&n=IHwtCTyIC4s5BfZO&q=85&s=7c6845f32a7f7a6573a4b9615a5cbaf3" width="2638" height="1570" data-path="images/docs/platform/create-control-plane.png" />
  </Step>

  <Step title="Get Domain and Certificate ARN">
    We will need two domains and certificate ARNs to point to the load balancer that we will be creating in the next step. Let's say you have a domain like `*.services.example.com` - we will be creating a DNS record with this later in Step 6. We recommend using [AWS Certificate Manager (ACM)](https://docs.aws.amazon.com/acm/latest/userguide/acm-overview.html) to create the certificate since it's easier to manage and renew the certificates automatically. To generate a certificate ARN, please follow the steps below. If you are not using AWS Certificate Manager, you can skip this step and continue to the next step.

    <Accordion title="Create the certificate in AWS Certificate Manager">
      1. Navigate to AWS Certificate Manager in the AWS console
      2. [Request a public certificate](https://docs.aws.amazon.com/acm/latest/userguide/gs-acm-request-public.html)
      3. Specify your domain (e.g., `*.services.example.com`)
      4. Choose DNS validation (recommended)
      5. Add the CNAME records provided by ACM to your DNS provider. Follow the [official AWS guide for DNS validation](https://docs.aws.amazon.com/acm/latest/userguide/dns-validation.html). For detailed steps on adding CNAME records, see [AWS documentation on DNS validation](https://docs.aws.amazon.com/acm/latest/userguide/gs-acm-validate-dns.html)
      6. Wait for the certificate to change to "Active" status (this may take 30 minutes or longer)
      7. Copy the certificate ARN for the next step (format will be like: `arn:aws:acm:region:account:certificate/certificate-id`)
    </Accordion>
  </Step>

  <Step title="Fill up the form to generate the opentofu/terraform code">
    A form will be presented with the details for the new cluster to be created. Fill in with your cluster details. Click `Submit` when done

    <Tabs>
      <Tab title="Create New Cluster">
        The key fields to fill up here are:

        * `Cluster Name` - A name for your cluster.
        * `Region` - The region where you want to create the cluster.
        * `Network Configuration` - Choose between `New VPC` or `Existing VPC` depending on your use case.
        * `Authentication` - This is how you are authenticated to AWS on your local machine. It's used to configure OpenTofu/Terraform to authenticate with AWS.
        * `S3 Bucket for OpenTofu/Terraform State` - OpenTofu/Terraform state will be stored in this bucket. It can be a preexisting bucket or a new bucket name. The new bucket will automatically be created by our script.
        * `Control Plane Configuration` - Control plane URL and the database details. You can chose between `PostgreSQL on kubernetes` or `Managed PostgreSQL (RDS)` or `Existing PostgreSQL configuration` depending on your use case.
        * `Load Balancer Configuration` - This is to configure the load balancer for your cluster. You can choose between `Public` or `Private` Load Balancer, it defaults to `Public`. You can also add certificate ARNs and domain names for the load balancer but these are optional.
      </Tab>

      <Tab title="Attach Existing Cluster">
        The key fields to fill up here are:

        * `Region` - The region where your cluster is already created.
        * `Cluster Configuration` - Provide the details of the existing cluster like the name of the cluster, URL of the OIDC provider, and the other required ARNs on the form.
        * `Cluster Addons` - TrueFoundry needs to install addons like ArgoCD, ArgoWorkflows, Keda, Istio, etc. Please disable the addons that are already installed on your cluster so that truefoundry installation does not overrride the existing configuration and affect your existing workloads.
        * `Network Configuration` - Provide the details of the existing VPC and subnets where the cluster is already created.
        * `Authentication` - This is how you are authenticated to AWS on your local machine. It's used to configure OpenTofu/Terraform to authenticate with AWS.
        * `S3 Bucket for OpenTofu/Terraform State` - OpenTofu/Terraform state will be stored in this bucket. It can be a preexisting bucket or a new bucket name. The new bucket will automatically be created by our script.
        * `Control Plane Configuration` - Control plane URL and the database details. You can chose between `PostgreSQL on kubernetes` or `Managed PostgreSQL (RDS)` or `Existing PostgreSQL configuration` depending on your use case.
        * `Load Balancer Configuration` - This is to configure the load balancer for your cluster. You can choose between `Public` or `Private` Load Balancer, it defaults to `Public`. You can also add certificate ARNs and domain names for the load balancer but these are optional.
      </Tab>
    </Tabs>

    Enter the domain and the certificate ARN that we got in previous step in the form as shown below.

    <img src="https://mintcdn.com/truefoundry/TOuU47bDo5UpP1Lw/images/docs/platform/domain-load-balancer-aws-cluster.png?fit=max&auto=format&n=TOuU47bDo5UpP1Lw&q=85&s=e92347fe1883ba78460cbde414b5ccea" width="3834" height="1864" data-path="images/docs/platform/domain-load-balancer-aws-cluster.png" />
  </Step>

  <Step title="Copy the curl command and execute it on your local machine">
    You will be presented with a `curl` command to download and execute the script. The script will take care of installing the pre-requisites, downloading OpenTofu/Terraform code and running it on your local machine to create the cluster. This will take around 40-50 minutes to complete.

    <img src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/curl-screenshot.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=320ca44c465ebc4d46aeb15528b5f61f" width="2110" height="772" data-path="images/docs/curl-screenshot.png" />
  </Step>

  <Step title="Create DNS Record">
    Once the script is executed, create the DNS record for the control plane url. To get the load balancer IP address, you can check the kubernetes service of type `LoadBalancer` in the `istio-system` namespace. You can run the following command to get the IP address.

    ```bash lines theme={"dark"}
    kubectl get svc -n istio-system tfy-istio-ingress -ojsonpath='{.status.loadBalancer.ingress[0].hostname}'
    ```

    This will give you the login screen to the control plane through which you can login via the same credentials used to register the tenant.
    Create a DNS record in your route 53 or your DNS provider with the following details

    | Record Type | Record Name            | Record Value              |
    | ----------- | ---------------------- | ------------------------- |
    | CNAME       | CONTROL\_PLANE\_DOMAIN | LOADBALANCER\_IP\_ADDRESS |
  </Step>

  <Step title="Attach the compute plane to the control plane">
    We will need to attach the same cluster as compute plane so that we can manage it form the platform. For this, you need to go to the platform section in the left panel and click on `Clusters`. Click on `Attach Existing Cluster` and fill in the details of the control plane cluster. The key fields to fill up here are:

    * `Cluster Name` - The name of the cluster.
    * `Cluster Addons` - Unselect all the addons as we have installed them while bringing up the control plane.
    * `Network Configuration` - Networking configuration of the control plane cluster.
    * `Authentication` - This is how you are authenticated to AWS on your local machine. It's used to configure OpenTofu/Terraform to authenticate with AWS.
    * `S3 Bucket for OpenTofu/Terraform State` - OpenTofu/Terraform state will be stored in this bucket. It can be a preexisting bucket or a new bucket name. You can use the same bucket that we used for the control plane and change the bucket key to be used for OpenTofu/Terraform state file.
    * `Platform Features` - This is to decide which features like BlobStorage, ClusterIntegration, ParameterStore, DockerRegistry and SecretsManager will be enabled for your cluster. To read more on how these integrations are used in the platform, please refer to the [platform features](/docs/infrastructure/deploy-compute-plane) page.

    <img src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/create-compute-plane-screenshot-1.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=b3febf85743f0b5d32adb737e23eadb6" width="3840" height="1938" data-path="images/docs/create-compute-plane-screenshot-1.png" />
  </Step>

  <Step title="Copy the curl command and execute it on your local machine">
    You will be presented with a `curl` command to download and execute the script. The script will take care of installing the pre-requisites, downloading OpenTofu/Terraform code and running it on your local machine to create the cluster. This will take around 40-50 minutes to complete.

    <img src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/curl-screenshot.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=320ca44c465ebc4d46aeb15528b5f61f" width="2110" height="772" data-path="images/docs/curl-screenshot.png" />
  </Step>

  <Step title="Verify the cluster is showing as connected in the platform">
    Once the script is executed, the cluster will be shown as connected in the platform.
  </Step>

  <Step title="Enable the deployment of workloads on the cluster from the platform">
    To enable the deployment feature which allows you to deploy services through the platform, you need to enable it;

    * In the left hand navigation, go to `Settings` then `Platform Feature Visibility` under `Preferences`
    * Click on `Edit` button. Then enable the toggle for `Enable Deployment`
    * Click on `Save` button.
  </Step>

  <Step title="Start deploying workloads to your cluster">
    You can start by going [here](https://docs.truefoundry.com/docs/deploy-first-service#deploy-from-github)
  </Step>
</Steps>

## FAQ

<AccordionGroup>
  <Accordion title="Can I use cert-manager to add TLS to the load balancer and not use AWS Certificate Manager?">
    Yes, you can use cert-manager to add TLS to the load balancer and not use AWS Certificate Manager. You can follow the instructions [here](https://cert-manager.io/docs/getting-started/) to install cert-manager and add TLS to the load balancer.
  </Accordion>

  <Accordion title="Can I use my own certificate and key files to add TLS to the load balancer?">
    Yes, please consult this [guide](/docs/add-certificate-for-tls#cert-files) to add your own certificate and key files to the load balancer.
  </Accordion>

  <Accordion title="How to enable SSL for PostgreSQL connections?">
    The TrueFoundry control plane supports SSL connections to PostgreSQL. You can configure SSL by setting the `DB_SSL_MODE` environment variable in your `truefoundry-values.yaml`.

    Supported `DB_SSL_MODE` values:

    | Mode          | Encryption | Certificate Validation | Use Case                                                        |
    | ------------- | ---------- | ---------------------- | --------------------------------------------------------------- |
    | `disable`     | No         | No                     | Local development or trusted networks                           |
    | `no-verify`   | Yes        | No                     | Managed databases with self-signed or unverified certs          |
    | `require`     | Yes        | Yes (system CA store)  | When you have a valid CA certificate and want full verification |
    | `verify-ca`   | Yes        | Yes (custom CA)        | Same as `require` but explicitly checks CA                      |
    | `verify-full` | Yes        | Yes (CA + hostname)    | Strictest mode, validates CA and hostname                       |

    SSL certificate environment variables:

    | Variable           | Purpose                                        | Required                                           |
    | ------------------ | ---------------------------------------------- | -------------------------------------------------- |
    | `DB_SSL_CA_PATH`   | Path to the server CA certificate file         | For `require`, `verify-ca`, or `verify-full` modes |
    | `DB_SSL_CERT_PATH` | Path to the client certificate file (for mTLS) | Not needed for AWS RDS                             |
    | `DB_SSL_KEY_PATH`  | Path to the client private key file (for mTLS) | Not needed for AWS RDS                             |

    **Scenario 1: Encrypted connection without certificate validation (`no-verify`)**

    This is the simplest option for AWS RDS. It encrypts the connection but skips server certificate validation.

    ```yaml truefoundry-values.yaml wrap lines theme={"dark"}
    servicefoundryServer:
      env:
        DB_SSL_MODE: "no-verify"
    mlfoundryServer:
      env:
        DB_SSL_MODE: "no-verify"
    ```

    **Scenario 2: Encrypted connection with certificate validation (`require`)**

    This mode encrypts the connection and validates the server certificate. You must provide the AWS RDS CA bundle so the application can verify the RDS server certificate.

    Download the RDS CA bundle and create a Kubernetes Secret:

    <Note>
      You can download the RDS CA bundle from the [AWS documentation](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/UsingWithRDS.SSL.html).
    </Note>

    ```bash wrap lines theme={"dark"}
    # Create the Kubernetes Secret
    kubectl create secret generic db-ca-certificate \
      --from-file=rds-combined-ca-bundle.pem=rds-combined-ca-bundle.pem \
      -n truefoundry
    ```

    Then configure `truefoundry-values.yaml` to mount the CA bundle and set `DB_SSL_CA_PATH`:

    ```yaml truefoundry-values.yaml wrap lines theme={"dark"}
    servicefoundryServer:
      env:
        DB_SSL_MODE: "require"
        DB_SSL_CA_PATH: "/etc/ssl/custom/rds-combined-ca-bundle.pem"
      extraVolumes:
        - name: db-ca-cert
          secret:
            secretName: db-ca-certificate
      extraVolumeMounts:
        - name: db-ca-cert
          mountPath: /etc/ssl/custom
          readOnly: true
    mlfoundryServer:
      env:
        DB_SSL_MODE: "require"
        DB_SSL_CA_PATH: "/etc/ssl/custom/rds-combined-ca-bundle.pem"
      extraVolumes:
        - name: db-ca-cert
          secret:
            secretName: db-ca-certificate
      extraVolumeMounts:
        - name: db-ca-cert
          mountPath: /etc/ssl/custom
          readOnly: true
    ```

    Upgrade the Helm release to apply the changes:

    ```bash wrap lines theme={"dark"}
    helm upgrade --install truefoundry oci://tfy.jfrog.io/tfy-helm/truefoundry -n truefoundry --create-namespace -f truefoundry-values.yaml
    ```
  </Accordion>

  <Accordion title="How to enable and access control plane monitoring (Grafana)?">
    TrueFoundry ships with a built-in monitoring stack that includes Grafana dashboards for the control plane. To enable it, add the following to your `truefoundry-values.yaml`:

    ```yaml truefoundry-values.yaml theme={"dark"}
    truefoundryMonitoring:
      enabled: true
      grafana:
        grafana.ini:
          auth.jwt:
            jwk_set_url: >-
              https://<your-truefoundry-control-plane-url>/api/svc/v1/keys/<tenant-name>/jwks
    ```

    Then upgrade the Helm release to apply the changes:

    ```bash theme={"dark"}
    helm upgrade --install truefoundry oci://tfy.jfrog.io/tfy-helm/truefoundry \
      -n truefoundry --create-namespace \
      -f truefoundry-values.yaml
    ```

    Once enabled, platform **admins** can access the Grafana dashboard at:

    ```
    https://<your-truefoundry-control-plane-url>/admin/grafana/
    ```

    <Note>
      * Replace `<your-truefoundry-control-plane-url>` with your actual control plane domain (e.g., `app.example.com`) and `<tenant-name>` with your TrueFoundry tenant name provided during onboarding.
      * Only users with the **admin** role can access this endpoint.
      * Make sure to include the trailing `/` at the end of the URL.
      * If you already have Prometheus or VictoriaLogs in your cluster, you can point the monitoring stack to them using `externalServices` instead of installing new instances.
    </Note>

    For the full configuration reference, see the [Control Plane Monitoring](/docs/platform/controlplane-monitoring) guide.
  </Accordion>
</AccordionGroup>
