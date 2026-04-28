> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# AWS

> This page provides an architecture overview, requirements and steps to setup a TrueFoundry compute plane cluster in AWS

The architecture of a TrueFoundry compute plane is as follows:

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/s4Aj2_qGCrSP-zc8/images/87feb78c-de29b95-AWS_5.png?fit=max&auto=format&n=s4Aj2_qGCrSP-zc8&q=85&s=ebf9cf0ff15f75ecf4d115020e71a3d5" width="1779" height="1182" data-path="images/87feb78c-de29b95-AWS_5.png" />
</Frame>

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

## Requirements:

The requirements to setup compute plane in each of the scenarios is as follows:

* <Icon icon="square-check" iconType="regular" /> Billing and [STS](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_temp_enable-regions.html#sts-regions-activate-deactivate) must be enabled for the AWS account.
* <Icon icon="square-check" iconType="regular" /> Please make sure you have enough quotas for GPU/Inferentia instances on the account depending on your usecase. You can check and increase quotas at [AWS EC2 service quotas](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-resource-limits.html)
* <Icon icon="square-check" iconType="regular" /> Please make sure you have created a certifcate for your domain in AWS Certificate Manager (ACM) and have the ARN of the certificate ready. This is required to setup TLS for the load balancer.
* <Icon icon="square-check" iconType="regular" /> You need to have enough permissions on the AWS account to create the resources needed for the compute plane. Check [this](#permissions-required-to-create-the-infrastructure) for more details. We usually recommend admin permission on the AWS account, but if you need the exact set of fine-grained permissions, you can check the list of permissions below:

```json json lines expandable theme={"dark"}
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "cloudformation:DescribeStacks",
                "cloudformation:ListStacks",
                "eks:*",
                "ec2:*",
                "iam:GetRole",
                "iam:ListPolicies",
                "elasticfilesystem:*",
                "kms:*",
                "route53:AssociateVPCWithHostedZone",
                "s3:ListAllMyBuckets",
                "sts:GetCallerIdentity"
            ],
            "Resource": "*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "events:*"
            ],
            "Resource": "arn:aws:events:$REGION:$ACCOUNT_ID:rule/$CLUSTER_NAME*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "iam:AddRoleToInstanceProfile",
                "iam:CreateInstanceProfile",
                "iam:DeleteInstanceProfile",
                "iam:GetInstanceProfile",
                "iam:TagInstanceProfile",
                "iam:UntagInstanceProfile",
                "iam:RemoveRoleFromInstanceProfile"
            ],
            "Resource": "arn:aws:iam::$ACCOUNT_ID:instance-profile/$CLUSTER_NAME*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "cloudformation:CreateStack",
                "cloudformation:DeleteStack",
                "cloudformation:GetTemplate"
            ],
            "Resource": "arn:aws:cloudformation:$REGION:$ACCOUNT_ID:stack/$CLUSTER_NAME*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "iam:CreateOpenIDConnectProvider",
                "iam:DeleteOpenIDConnectProvider",
                "iam:GetOpenIDConnectProvider",
                "iam:TagOpenIDConnectProvider"
            ],
            "Resource": "arn:aws:iam::$ACCOUNT_ID:oidc-provider/*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "iam:CreatePolicy",
                "iam:DeletePolicy",
                "iam:GetPolicy",
                "iam:TagPolicy",
                "iam:UntagPolicy",
                "iam:GetPolicyVersion",
                "iam:ListPolicyVersions"
            ],
            "Resource": [
                "arn:aws:iam::$ACCOUNT_ID:policy/tfy-*",
                "arn:aws:iam::$ACCOUNT_ID:policy/truefoundry-*",
                "arn:aws:iam::$ACCOUNT_ID:policy/AmazonEKS_Karpenter_Controller_Policy*",
                "arn:aws:iam::$ACCOUNT_ID:policy/AmazonEKS_CNI_Policy*",
                "arn:aws:iam::$ACCOUNT_ID:policy/AmazonEKS_AWS_Load_Balancer_Controller*",
                "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryFullAccess"
            ]
        },
        {
            "Effect": "Allow",
            "Action": [
                "iam:AttachRolePolicy",
                "iam:CreateRole",
                "iam:CreateServiceLinkedRole",
                "iam:DeleteRole",
                "iam:DetachRolePolicy",
                "iam:ListAttachedRolePolicies",
                "iam:ListInstanceProfilesForRole",
                "iam:ListRolePolicies",
                "iam:TagRole",
                "iam:PassRole"
            ],
            "Resource": "arn:aws:iam::$ACCOUNT_ID:role/$CLUSTER_NAME*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "logs:*"
            ],
            "Resource": [
                "arn:aws:logs:$REGION:$ACCOUNT_ID:log-group:/aws/eks/$CLUSTER_NAME*",
                "arn:aws:logs:$REGION:$ACCOUNT_ID:log-group::log-stream:"
            ]
        },
        {
            "Effect": "Allow",
            "Action": [
                "s3:*"
            ],
            "Resource": [
                "arn:aws:s3:::$CLUSTER_NAME*",
                "arn:aws:s3:::$CLUSTER_NAME*/*"
            ]
        },
        {
            "Effect": "Allow",
            "Action": "sqs:*",
            "Resource": "arn:aws:sqs:$REGION:$ACCOUNT_ID:$CLUSTER_NAME*"
        },
        {
            "Effect": "Allow",
            "Action": "ssm:GetParameters",
            "Resource": "arn:aws:ssm:$REGION:$ACCOUNT_ID:parameter/tfy*"
        }
    ]
}
```

Regarding the VPC and EKS cluster, you can decide between the following scenarios:

<Tabs>
  <Tab title="New VPC and New EKS Cluster">
    1. <Icon icon="square-check" iconType="regular" /> The new VPC should will have a CIDR range of /20 or larger, at least 2 availability zones and private subnets with CIDR `/24` or larger. This is to ensure capacity for \~250 instances and 4096 pods.
    2. <Icon icon="square-check" iconType="regular" /> If you are using [custom networking](https://docs.aws.amazon.com/eks/latest/userguide/cni-custom-network.html),
       you need to have CGNAT IP address in each AZ. CGNAT space and route tables will be setup in the VPC.
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
    2. <Icon icon="square-check" iconType="regular" /> EBS CSI Driver should be installed [Installation Guide](https://github.com/kubernetes-sigs/aws-ebs-csi-driver/blob/master/docs/install.md) - Required for persistent volume support for Notebooks, SSH.
    3. <Icon icon="square-check" iconType="regular" /> EFS CSI Driver should be installed [Installation Guide](https://github.com/kubernetes-sigs/aws-efs-csi-driver/blob/master/docs/README.md) - Required for model and data caching.
    4. <Icon icon="square-check" iconType="regular" /> AWS Load Balancer Controller (>=v2.12.0) should be installed [Installation Guide](https://kubernetes-sigs.github.io/aws-load-balancer-controller/latest/deploy/installation/) - Required for Ingress and Service type LoadBalancer support. The appropriate IAM roles for service account (IRSA) should be created.
    5. <Icon icon="square-check" iconType="regular" /> Although this is not compulsory, we highly recommend Karpenter to be installed on the cluster. It makes a lot of functionalities in TrueFoundry easier, faster and cost-effective.
  </Tab>
</Tabs>

## Setting up compute plane

TrueFoundry compute plane infrastructure is provisioned using OpenTofu/Terraform. You can download the OpenTofu/Terraform code for your exact account by filling up your account details and downloading a script that can be executed on your local machine.

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
    Go to the platform section in the left panel and click on `Clusters`. You can click on `Create New Cluster` or `Attach Existing Cluster` depending on your use case. Read the requirements and if everything is satisfied, click on `Continue`.

    <img src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/create-compute-plane-screenshot-1.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=b3febf85743f0b5d32adb737e23eadb6" width="3840" height="1938" data-path="images/docs/create-compute-plane-screenshot-1.png" />
  </Step>

  <Step title="Get Domain and Certificate ARN">
    We will need a domain and certificate ARN to point to the load balancer that we will be creating in the next step. Let's say you have a domain like `*.services.example.com` - we will be creating a DNS record with this later in Step 6. We recommend using [AWS Certificate Manager (ACM)](https://docs.aws.amazon.com/acm/latest/userguide/acm-overview.html) to create the certificate since it's easier to manage and renew the certificates automatically. To generate a certificate ARN, please follow the steps below. If you are not using AWS Certificate Manager, you can skip this step and continue to the next step.

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

  <Step title="Fill up the form to generate the OpenTofu/Terraform code">
    A form will be presented with the details for the new cluster to be created. Fill in with your cluster details. Click `Submit` when done

    <Tabs>
      <Tab title="Create New Cluster">
        The key fields to fill up here are:

        * `Cluster Name` - A name for your cluster.
        * `Region` - The region where you want to create the cluster.
        * `Network Configuration` - Choose between `New VPC` or `Existing VPC` depending on your use case.
        * `Authentication` - This is how you are authenticated to AWS on your local machine. It's used to configure OpenTofu/Terraform to authenticate with AWS.
        * `S3 Bucket for OpenTofu/Terraform State` - OpenTofu/Terraform state will be stored in this bucket. It can be a preexisting bucket or a new bucket name. The new bucket will automatically be created by our script.
        * `Load Balancer Configuration` - This is to configure the load balancer for your cluster. You can choose between `Public` or `Private` Load Balancer, it defaults to `Public`. You can also add certificate ARNs and domain names for the load balancer but these are optional.
        * `Platform Features` - This is to decide which features like BlobStorage, ClusterIntegration, ParameterStore, DockerRegistry and SecretsManager will be enabled for your cluster. To read more on how these integrations are used in the platform, please refer to the [platform features](/docs/infrastructure/deploy-compute-plane) page.
      </Tab>

      <Tab title="Attach Existing Cluster">
        The key fields to fill up here are:

        * `Region` - The region where your cluster is already created.
        * `Cluster Configuration` - Provide the details of the existing cluster like the name of the cluster, URL of the OIDC provider, and the other required ARNs on the form.
        * `Cluster Addons` - TrueFoundry needs to install addons like ArgoCD, ArgoWorkflows, Keda, Istio, etc. Please disable the addons that are already installed on your cluster so that truefoundry installation does not overrride the existing configuration and affect your existing workloads.
        * `Network Configuration` - Provide the details of the existing VPC and subnets where the cluster is already created.
        * `Authentication` - This is how you are authenticated to AWS on your local machine. It's used to configure OpenTofu/Terraform to authenticate with AWS.
        * `S3 Bucket for OpenTofu/Terraform State` - OpenTofu/Terraform state will be stored in this bucket. It can be a preexisting bucket or a new bucket name. The new bucket will automatically be created by our script.
        * `Load Balancer Configuration` - This is to configure the load balancer for your cluster. You can choose between `Public` or `Private` Load Balancer, it defaults to `Public`. You can also add certificate ARNs and domain names for the load balancer but these are optional.
        * `Platform Features` - This is to decide which features like BlobStorage, ClusterIntegration, ParameterStore, DockerRegistry and SecretsManager will be enabled for your cluster. To read more on how these integrations are used in the platform, please refer to the [platform features](/docs/infrastructure/deploy-compute-plane) page.
      </Tab>
    </Tabs>

    Enter the domain and the certificate ARN that we got in previous step in the form as shown below.

    <img src="https://mintcdn.com/truefoundry/TOuU47bDo5UpP1Lw/images/docs/platform/domain-load-balancer-aws-cluster.png?fit=max&auto=format&n=TOuU47bDo5UpP1Lw&q=85&s=e92347fe1883ba78460cbde414b5ccea" width="3834" height="1864" data-path="images/docs/platform/domain-load-balancer-aws-cluster.png" />
  </Step>

  <Step title="Copy the curl command and execute it on your local machine">
    You will be presented with a `curl` command to download and execute the script. The script will take care of installing the pre-requisites, downloading OpenTofu/Terraform code and running it on your local machine to create the cluster. This will take around 40-50 minutes to complete.

    <img src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/curl-screenshot.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=320ca44c465ebc4d46aeb15528b5f61f" width="2110" height="772" data-path="images/docs/curl-screenshot.png" />
  </Step>

  <Step title="Verify the cluster is showing as connected in the platform">
    Once the script is executed, the cluster will be shown as connected in the platform.
  </Step>

  <Step title="Create DNS Record">
    We can get the load-balancer's IP address by going to the platform section in the bottom left panel under the Clusters section. Under the preferred cluster, you'll see the load balancer IP address under the `Base Domain URL` section.

    <Frame>
      <img src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/cluster-page-base-domain-url.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=effcaf4ad8dedd6a1da28bd694ded547" width="4416" height="1262" data-path="images/docs/cluster-page-base-domain-url.png" />
    </Frame>

    Create a DNS record in your route 53 or your DNS provider with the following details

    | Record Type | Record Name        | Record value              |
    | ----------- | ------------------ | ------------------------- |
    | CNAME       | \*.tfy.example.com | LOADBALANCER\_IP\_ADDRESS |
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
    If you have your own certificate files (for example, from another certificate provider or self-signed), you can use them directly with TrueFoundry.

    1. Create a Kubernetes secret with your certificate and key, or create a self-signed certificate:

           <CodeGroup>
             ```shell Shell lines theme={"dark"}
             # Generate a self-signed certificate
             openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
               -keyout tls.key -out tls.crt \
               -subj "/CN=*.example.com" \
               -addext "subjectAltName = DNS:example.com,DNS:*.example.com"
             ```
           </CodeGroup>

           <CodeGroup>
             ```shell Shell lines theme={"dark"}
             # Create secret from local certificate files
             kubectl create secret tls example-com-tls \
               --cert=path/to/cert/file \
               --key=path/to/key/file \
               -n istio-system
             ```
           </CodeGroup>

    2. Once the secret is created, head over to the cluster page and navigate to the `tfy-istio-ingress` add-on. Add the secret name in the `tfyGateway.spec.servers[1].tls.credentialName` section and ensure that `tfyGateway.spec.servers[1].port.protocol` is set to `HTTPS`. Here we are using `example-com-tls` as the secret name, which contains the certificate and key.

           <CodeGroup>
             ```yaml YAML lines theme={"dark"}
                 servers:
                   - <REDACTED>
                   - hosts:
                       - "*"
                     port:
                       name: https-tfy-wildcard
                       number: 443
                       protocol: HTTPS
                     tls:
                       mode: SIMPLE
                       credentialName: example-com-tls
             ```
           </CodeGroup>

    <Warning>
      Self-signed certificates will cause browser warnings. They should only be used for testing or internal systems. To connect to services with self-signed certificates, you have to pass the CA certificate to verify the SSL certificate.
    </Warning>
  </Accordion>

  <Accordion title="How to review the OpenTofu/Terraform code that is generated by the platform?">
    In Step 4 in the guide above, when you run the curl command, the OpenTofu/Terraform code will be downloaded to your local machine. The script will ask you before executing the OpenTofu/Terraform code at which point you can stop the execution and review the OpenTofu/Terraform code generated by the platform.
  </Accordion>

  <Accordion title="How do I set up custom networking in EKS?">
    By default, Amazon VPC CNI assigns Pods an IP address from the primary subnet. If the primary subnet CIDR is too small, the CNI may not be able to acquire enough secondary IP addresses for your Pods. Custom networking solves this by using a secondary CIDR for Pod IPs. For more details, see the [Amazon EKS Custom Networking documentation](https://docs.aws.amazon.com/eks/latest/best-practices/custom-networking.html).

    **Steps:**

    1. Attach a secondary CIDR to the VPC (e.g. `100.64.0.0/16` per [RFC 6598](https://datatracker.ietf.org/doc/html/rfc6598)) and create new subnets with that CIDR in the same AZs as your primary subnets.
    2. Ensure the secondary subnets are added to the route tables of your primary subnets (this typically happens automatically).
    3. Configure the AWS VPC CNI to use the secondary subnets.

       If you used TrueFoundry's OpenTofu/Terraform code, add this to the [EKS module](https://github.com/truefoundry/terraform-aws-truefoundry-cluster):

    ```hcl theme={"dark"}
    cluster_addons_vpc_cni_additional_configurations = {
        env = {
          AWS_VPC_K8S_CNI_CUSTOM_NETWORK_CFG = "true"
          ENI_CONFIG_LABEL_DEF               = "topology.kubernetes.io/zone"
        }
    }
    ```

    Otherwise, set the env variables directly:

    ```bash theme={"dark"}
    kubectl set env daemonset aws-node -n kube-system "AWS_VPC_K8S_CNI_CUSTOM_NETWORK_CFG=true"
    kubectl set env daemonset aws-node -n kube-system "ENI_CONFIG_LABEL_DEF=topology.kubernetes.io/zone"
    ```

    4. Create an `ENIConfig` resource for each AZ where your EKS nodes run:

    ```yaml theme={"dark"}
    apiVersion: crd.k8s.amazonaws.com/v1alpha1
    kind: ENIConfig
    metadata:
      name: us-east-1a   # AZ of the EKS nodes
    spec:
      securityGroups:
        - sg-0dff111a1d11c1c11   # security group of the EKS nodes
      subnet: subnet-011b111c1f11fdf11   # subnet with the secondary CIDR
    ```

    5. Restart the nodes one by one to apply the changes. Pods will be rescheduled onto nodes with secondary IP addresses.
  </Accordion>
</AccordionGroup>
