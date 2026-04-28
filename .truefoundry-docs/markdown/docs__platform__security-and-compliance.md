> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Security and Compliance

> Learn about Truefoundry's enterprise-grade security features, compliance certifications, and data protection measures

Truefoundry is built with security and compliance as foundational principles. Our platform provides enterprise-grade security features, comprehensive compliance certifications, and robust data protection to ensure your AI workloads and data remain secure.

For detailed security information, compliance reports, and trust documentation, visit [trust.truefoundry.com](https://trust.truefoundry.com).

## Our Security Philosophy

Our security strategy is anchored in two core principles: **defense in depth** and **secure by design**. As an AI platform that orchestrates model deployments, handles MCP communications, and manages sensitive data flows across your infrastructure, security is not an afterthought—it's embedded into every layer of our architecture.

We believe modern software security requires proactive measures rather than reactive responses. Our approach combines industry-standard security practices with secure defaults and automated tooling to create a robust security posture that scales with your needs.

<CardGroup cols={2}>
  <Card title="Defense in Depth" icon="shield-halved">
    Multiple layers of security controls ensure that no single point of failure can compromise the system
  </Card>

  <Card title="Secure by Design" icon="lock">
    Security considerations are integrated from the earliest design stages, not bolted on later
  </Card>

  <Card title="Automated Security" icon="robot">
    Continuous automated scanning with 30+ security tools on every code change
  </Card>

  <Card title="Human Expertise" icon="user-shield">
    Expert security reviews and threat modeling complement automated tooling
  </Card>
</CardGroup>

### Beyond Automated Tools: Human-Centered Security

While our automated security tooling provides comprehensive coverage, we recognize that true security requires human expertise and collaborative oversight. Our security posture extends beyond automated scanning:

<AccordionGroup>
  <Accordion title="Expert Security Code Reviews">
    Our security engineers conduct thorough code reviews focusing on logic flaws, business logic vulnerabilities, and complex attack vectors that automated tools might miss. Every significant change undergoes security-focused peer review.
  </Accordion>

  <Accordion title="Threat Modeling & Risk Assessment">
    Regular security assessments evaluate our attack surface, identify potential threat vectors, and validate our defense mechanisms against real-world attack scenarios. We continuously update our threat models as the landscape evolves.
  </Accordion>

  <Accordion title="Security Architecture Review">
    All major design decisions undergo security architecture review to ensure security considerations are embedded from the earliest stages. This prevents costly security retrofits and ensures consistent protection.
  </Accordion>

  <Accordion title="Penetration Testing">
    Regular penetration testing by internal and external security experts validates our defenses and identifies potential weaknesses before they can be exploited.
  </Accordion>

  <Accordion title="Community-Driven Security">
    We actively engage with the security research community, maintain responsible disclosure processes, and leverage collective intelligence to identify and address emerging threats promptly.
  </Accordion>

  <Accordion title="Security Champion Program">
    Developers across the organization receive security training and act as security advocates within their teams, creating a culture of security awareness that permeates every aspect of development.
  </Accordion>
</AccordionGroup>

### Comprehensive Security Pipeline

Our security pipeline operates at multiple levels, ensuring that vulnerabilities are caught as early as possible in the development lifecycle:

<Steps>
  <Step title="Pre-commit Security Gates">
    Before any code reaches our repository, it must pass through rigorous pre-commit hooks including:

    * **Static analysis** with Bandit for common security issues
    * **Semantic pattern matching** with Semgrep for complex vulnerability patterns
    * **Secrets detection** with Dodgy and Gitleaks for hardcoded credentials
    * **Type checking** and code quality enforcement
  </Step>

  <Step title="Continuous Integration Security">
    Our CI/CD pipelines implement automated security scanning on every pull request and commit:

    * **30+ security scans** trigger automatically on every PR
    * **CodeQL and Semgrep** for deep semantic code analysis
    * **Gitleaks** for comprehensive secret detection across git history
    * **pip-audit and npm audit** for dependency vulnerability scanning
    * **Container security assessment** for all containerized components
  </Step>

  <Step title="Supply Chain Security">
    We maintain strict oversight of our software supply chain:

    * Automated dependency vulnerability scanning
    * Software Bill of Materials (SBOM) generation for complete transparency
    * License compliance checking to ensure all components meet security standards
    * Continuous monitoring of upstream dependencies for new vulnerabilities
  </Step>

  <Step title="Container Security Hardening">
    Our containerized deployments follow security best practices:

    * Multi-stage builds to minimize attack surface
    * Minimal base images with the latest security updates
    * Non-root user execution and read-only filesystems
    * Comprehensive scanning with Grype
  </Step>

  <Step title="Runtime Security Monitoring">
    Beyond build-time security, we implement runtime monitoring to detect and respond to potential threats in production environments.
  </Step>
</Steps>

<Info>
  This multi-layered approach ensures that security vulnerabilities are identified and addressed at the earliest possible stage, reducing risk and maintaining the integrity of our platform.
</Info>

## Compliance Certifications

Truefoundry maintains multiple compliance certifications to meet enterprise security and regulatory requirements:

* **SOC 2 Type II**: Certified for security, availability, processing integrity, confidentiality, and privacy
* **GDPR**: Compliant with the General Data Protection Regulation for data privacy
* **HIPAA**: Compliant with Health Insurance Portability and Accountability Act for healthcare data

<Frame caption="Our Compliance Certifications">
  <Columns cols={3}>
    <Card>
      <img src="https://mintcdn.com/truefoundry/VhkkzLbS0sN_6xZz/images/icons/SOC.png?fit=max&auto=format&n=VhkkzLbS0sN_6xZz&q=85&s=01bddac60c60b709fc49096ed82c33fe" alt="SOC 2 Type II Logo" width="679" height="679" data-path="images/icons/SOC.png" />
    </Card>

    <Card>
      <img src="https://mintcdn.com/truefoundry/VhkkzLbS0sN_6xZz/images/icons/gdpr.png?fit=max&auto=format&n=VhkkzLbS0sN_6xZz&q=85&s=18b9e6154fa6892052755a4649298261" alt="GDPR Logo" width="679" height="679" data-path="images/icons/gdpr.png" />
    </Card>

    <Card>
      <img src="https://mintcdn.com/truefoundry/VhkkzLbS0sN_6xZz/images/icons/hipaa.png?fit=max&auto=format&n=VhkkzLbS0sN_6xZz&q=85&s=d53d0d7925d1a759c429e61d100fc1cb" alt="HIPAA Logo" width="679" height="679" data-path="images/icons/hipaa.png" />
    </Card>
  </Columns>
</Frame>

<Info>
  Compliance certifications apply to Truefoundry's managed infrastructure. For
  self-hosted deployments, compliance depends on your infrastructure and
  security controls.
</Info>

## Data Security

<Tabs>
  <Tab title="Self-Hosted Deployment">
    ### Data Residency and Sovereignty

    If you are self-hosting the control-plane, Truefoundry's architecture ensures that your data remains under your control:

    * **Data Stays in Your Environment**: When deployed on your infrastructure, all data, models, and artifacts remain within your cloud account or on-premises environment
    * **No Data Egress**: Deploying on your own infrastructure eliminates data egress costs and ensures data never leaves your environment

    Truefoundry control-plane communicates with our central authentication/licensing server and exchanges the following information:

    <AccordionGroup>
      <Accordion title="Emails of the users logging into the Truefoundry platform">
        Truefoundry control-plane uses our remote authentication server to
        authenticate the users logging into the Truefoundry platform. If you have
        added SSO, this will be a chained authentication flow where in the user is
        validated by Truefoundry and your IdP both. To read more on this, please refer
        to our [SSO documentation](/docs/sso).
      </Accordion>

      <Accordion title="Usage analytics">
        The control-plane communicates with our central analytics server ([https://analytics.truefoundry.com](https://analytics.truefoundry.com)) and exchanges the following information:

        * The number of clusters connected to the control-plane
        * The addons installed in each of the clusters along with their versions
        * The version of the control-plane
        * The usage metrics from the AI Gateway

        This is an example of a sample payload that is sent to the analytics server. This request is sent once every hour from the control-plane to the analytics server.

        ```json expandable theme={"dark"}
        {
          "clusters": [
            {
              "clusterName": "test-cluster1",
              "kubernetesVersion": "v1.33.4-eks-e386d34",
              "provider": "aws",
              "tenantName": "tenant1",
              "controlPlaneVersion": "0.90.0",
              "controlPlaneURL": "https://app.truefoundry.com",
              "truefoundryControlPlane": true,
              "controlPlaneProvider": "aws",
              "isConnected": true,
              "addons": [
                {
                  "name": "truefoundry",
                  "version": "0.90.0",
                  "healthStatus": "Healthy",
                  "syncStatus": "OutOfSync",
                  "sourceNamespace": "argocd",
                  "destinationNamespace": "truefoundry",
                  "addonName": "truefoundry"
                },
                {
                  "name": "tfy-gpu-operator",
                  "version": "0.4.0",
                  "healthStatus": "Healthy",
                  "syncStatus": "OutOfSync",
                  "sourceNamespace": "argocd",
                  "destinationNamespace": "tfy-gpu-operator",
                  "addonName": "tfy-gpu-operator"
                },
                {
                  "name": "prometheus",
                  "version": "69.6.0",
                  "healthStatus": "Healthy",
                  "syncStatus": "Synced",
                  "sourceNamespace": "argocd",
                  "destinationNamespace": "prometheus",
                  "addonName": "prometheus"
                },
                {
                  "name": "keda",
                  "version": "2.17.1",
                  "healthStatus": "Healthy",
                  "syncStatus": "Synced",
                  "sourceNamespace": "argocd",
                  "destinationNamespace": "keda",
                  "addonName": "keda"
                },
                {
                  "name": "argocd",
                  "version": "9.4.14",
                  "healthStatus": "Healthy",
                  "syncStatus": "Synced",
                  "sourceNamespace": "argocd",
                  "destinationNamespace": "argocd",
                  "addonName": "argocd"
                },
                {
                  "name": "argo-rollout",
                  "version": "2.39.5",
                  "healthStatus": "Healthy",
                  "syncStatus": "Synced",
                  "sourceNamespace": "argocd",
                  "destinationNamespace": "argo-rollouts",
                  "addonName": "argo-rollouts"
                },
                {
                  "name": "aws-efs-csi-driver",
                  "version": "3.1.8",
                  "healthStatus": "Healthy",
                  "syncStatus": "Synced",
                  "sourceNamespace": "aws-efs-csi-driver",
                  "destinationNamespace": "aws-efs-csi-driver",
                  "addonName": "aws-efs-csi-driver"
                },
                {
                  "name": "tfy-logs",
                  "version": "0.1.10",
                  "healthStatus": "Healthy",
                  "syncStatus": "OutOfSync",
                  "sourceNamespace": "argocd",
                  "destinationNamespace": "tfy-logs",
                  "addonName": "tfy-logs"
                },
                {
                  "name": "aws-ebs-csi-driver",
                  "version": "2.41.0",
                  "healthStatus": "Healthy",
                  "syncStatus": "Synced",
                  "sourceNamespace": "aws-ebs-csi-driver",
                  "destinationNamespace": "aws-ebs-csi-driver",
                  "addonName": "aws-ebs-csi-driver"
                },
                {
                  "name": "metrics-server",
                  "version": "3.12.1",
                  "healthStatus": "Healthy",
                  "syncStatus": "Synced",
                  "sourceNamespace": "kube-system",
                  "destinationNamespace": "kube-system",
                  "addonName": "metrics-server"
                },
                {
                  "name": "karpenter",
                  "version": "0.5.6",
                  "healthStatus": "Healthy",
                  "syncStatus": "Synced",
                  "sourceNamespace": "kube-system",
                  "destinationNamespace": "kube-system",
                  "addonName": "tfy-karpenter"
                },
                {
                  "name": "tfy-prometheus-config",
                  "version": "0.2.14",
                  "healthStatus": "Healthy",
                  "syncStatus": "OutOfSync",
                  "sourceNamespace": "prometheus",
                  "destinationNamespace": "prometheus",
                  "addonName": "tfy-prometheus-config"
                },
                {
                  "name": "argo-workflows",
                  "version": "0.45.12",
                  "healthStatus": "Healthy",
                  "syncStatus": "Synced",
                  "sourceNamespace": "argo-workflows",
                  "destinationNamespace": "argo-workflows",
                  "addonName": "argo-workflows"
                },
                {
                  "name": "tfy-istio-ingress",
                  "version": "0.3.2",
                  "healthStatus": "Healthy",
                  "syncStatus": "OutOfSync",
                  "sourceNamespace": "istio-system",
                  "destinationNamespace": "istio-system",
                  "addonName": "tfy-istio-ingress"
                },
                {
                  "name": "istio-base",
                  "version": "1.25.4",
                  "healthStatus": "Healthy",
                  "syncStatus": "OutOfSync",
                  "sourceNamespace": "istio-system",
                  "destinationNamespace": "istio-system",
                  "addonName": "istio-base"
                },
                {
                  "name": "istio-discovery",
                  "version": "1.25.4",
                  "healthStatus": "Healthy",
                  "syncStatus": "OutOfSync",
                  "sourceNamespace": "istio-system",
                  "destinationNamespace": "istio-system",
                  "addonName": "istio-discovery"
                },
                {
                  "name": "elasti",
                  "version": "0.1.16",
                  "healthStatus": "Healthy",
                  "syncStatus": "Synced",
                  "sourceNamespace": "elasti",
                  "destinationNamespace": "elasti",
                  "addonName": "elasti"
                }
              ],
              "autopilotStatistics": {
                "clusterName": "tfy-prod-euwe1",
                "tenantName": "internal",
                "autopilot": { "addons": false, "service": false },
                "p50CpuUsage": 4.49,
                "p50CpuRequested": 65.822,
                "p50CpuAllocatable": 211.02,
                "p50MemoryUsage": 133.447,
                "p50MemoryRequested": 255.933,
                "p50MemoryAllocatable": 649.102,
                "p50GpuCountUsage": 0,
                "p50GpuCountRequested": 0,
                "p50GpuCountProvisioned": 0,
                "p50GpuMemoryUsage": 0,
                "p50GpuMemoryProvisioned": 0
              }
            }
          ]
        }
        ```

        <h4 id="gateway-analytics">Gateway Analytics</h4>

        The AI Gateway also sends usage and configuration analytics to the same analytics server. This data helps us understand gateway utilization patterns and improve the service.

        The gateway analytics payload includes:

        * **Usage metrics**: Token consumption, request counts, error rates, provider and model usage breakdown, guardrails execution statistics, routing decisions, and MCP tool calls
        * **Configuration metrics**: Counts of configured provider accounts, models, guardrails, MCP servers, rate/budget limit rules, and prompts

        ```json expandable theme={"dark"}
        {
          "tenantName": "tenant1",
          "data": {
            "1738800000": {
              "usage": {
                "inputTokens": 125000,
                "outputTokens": 48000,
                "totalTokens": 173000,
                "totalRequestCount": 1250,
                "totalErrorCounts": 12,
                "providerWiseModelCounts": {
                  "openai": 800,
                  "anthropic": 350,
                  "azure": 100
                },
                "modelWiseCounts": {
                  "openai": {
                    "gpt-4o": 500,
                    "gpt-4o-mini": 300
                  },
                  "anthropic": {
                    "claude-3-5-sonnet": 250,
                    "claude-3-haiku": 100
                  },
                  "azure": {
                    "gpt-4": 100
                  }
                },
                "guardrailsRuntime": {
                  "totalEligibleRequests": 1250,
                  "guardrailPass": 1200,
                  "guardrailFail": 35,
                  "guardrailMutate": 10,
                  "guardrailRunError": 5
                },
                "routingData": {
                  "totalRouting": 1250,
                  "tfyRateLimited": 8,
                  "tfyBudgetLimited": 3
                },
                "totalMCPToolCalls": 75
              },
              "config": {
                "models": {
                  "totalProviderAccounts": 5,
                  "totalModels": 12,
                  "totalVirtualModels": 3
                },
                "guardrails": {
                  "totalGuardrailGroups": 4,
                  "totalGuardrailIntegrations": 6,
                  "guardrailCountsByIntegrationType": {
                    "azure-content-safety": 2,
                    "openai-moderation": 2,
                    "tfy-content-moderation": 2
                  }
                },
                "mcpServers": {
                  "totalMcpGroups": 2,
                  "totalMcpServers": 8
                },
                "rules": {
                  "totalRateLimitRules": 15,
                  "totalBudgetLimitRules": 8,
                  "totalGuardrailConfigRules": 6,
                  "otelExportConfigured": true
                },
                "prompts": {
                  "totalPrompts": 24,
                  "totalPromptVersions": 67
                }
              }
            }
          }
        }
        ```
      </Accordion>

      <Accordion title="Public Pricing of instance types and models">
        The control-plane communicates with our central catalogue server
        ([https://catalogue.truefoundry.com](https://catalogue.truefoundry.com)) to get the public pricing of the instance
        types of AWS/GCP/Azure that are shown on the Truefoundry dashboard. We also
        fetch the public pricing of the LLM models that are added to the AI Gateway.
      </Accordion>

      <Accordion title="Crash and Error Logs">
        Truefoundry uses Sentry to collect crash and error logs from the control-plane
        and send it to Sentry's servers. This helps us identify and fix issues
        quickly.

        **Sensitive data and PII**: Sensitive data (such as user prompts and other
        personally identifiable information) is not sent to Sentry. [Sentry does not
        send or log PII by
        default](https://docs.sentry.io/platforms/python/data-management/sensitive-data/#personally-identifiable-information-pii)
      </Accordion>
    </AccordionGroup>

    ### Encryption

    * **Encryption at Rest**: Encrypted at rest using AES-256 encryption
    * **Encryption in Transit**: All network communications use TLS 1.2 or higher to encrypt data in transit
    * **Secret Management**: Integration with cloud-native secret managers (AWS Secrets Manager, Azure Key Vault, GCP Secret Manager) for secure credential storage
  </Tab>

  <Tab title="Managed Deployment">
    ### Data Residency and Sovereignty

    Truefoundry's infrastructure is globally distributed and can cater to your data residency requirements.

    * **Bring Your Own Storage**: Use your own blob storage (S3, Azure Blob, GCS) to maintain full control over data storage and access
    * **Decide your region of gateway deployment and logs storage**: In our SAAS hosted gateway, you can choose the region of the gateway deployment and the logs storage.

    ### Encryption

    All data is protected with industry-standard encryption:

    * **Encryption at Rest**: Data stored in Truefoundry-managed infrastructure is encrypted at rest using AES-256 encryption
    * **Encryption in Transit**: All network communications use TLS 1.2 or higher to encrypt data in transit
    * **Secret Management**: Integration with cloud-native secret managers (AWS Secrets Manager, Azure Key Vault, GCP Secret Manager) for secure credential storage

    ### Data Isolation

    * **Multi-Tenancy**: Truefoundry provides secure multi-tenant isolation at the workspace and resource level
    * **Network Segmentation**: Control plane, compute plane, and gateway plane operate independently with proper network segmentation
    * **Resource-Level Access Control**: Granular permissions ensure users can only access resources they're authorized to use
  </Tab>
</Tabs>

## Access Control and Authentication

<AccordionGroup>
  <Accordion title="Authentication Methods">
    Truefoundry supports multiple authentication mechanisms:

    * **Single Sign-On (SSO)**: Integrate with your identity provider (IdP) including SAML 2.0 and OIDC-compatible providers like Google, Microsoft Entra ID, Keycloak, Okta, etc. Learn more about our [SSO documentation](/docs/sso).
    * **JWT Token Authentication**: API access using JWT tokens from your identity provider
    * **API Keys**: Secure API key-based authentication for programmatic access
    * **Token Rotation**: Truefoundry supports automated token rotation for the API keys and tokens used to authenticate with the Truefoundry platform. Learn more about our [Token Rotation documentation](/docs/platform/virtual-account-management#configure-auto-rotation-of-virtual-account-tokens).
  </Accordion>

  <Accordion title="Authorization and Access Control">
    Comprehensive role-based access control (RBAC) ensures users have appropriate permissions:

    * **Tenant-Level Permissions**: Control access at the organization level with Tenant Admin and Tenant Member roles
    * **Resource-Level Permissions**: Granular permissions for workspaces, clusters, ML repositories, secret groups, and provider accounts
    * **Team-Based Access**: Organize users into teams with shared permissions and access controls
    * **Fine-Grained Permissions**: Viewer, Editor, and Admin roles for different resources

    For detailed information on access control, see [Access Control](/docs/collaboration-and-access-control).
  </Accordion>
</AccordionGroup>

## Infrastructure Security

<AccordionGroup>
  <Accordion title="Architecture Security">
    Truefoundry's split-plane architecture provides security through separation, ensuring sensitive data and workloads remain under your control:

    **Data Never Leaves Your Environment**

    * The compute plane runs entirely within your Kubernetes cluster (EKS, GKE, AKS, OpenShift, or on-premises). Truefoundry does not provide or access compute resources directly—all workloads execute in your controlled environment
    * Your models, datasets, and artifacts stay within your cloud account or on-premises infrastructure, eliminating data egress risks
    * The computeplane load balancer can be put behing a private endpoint limiting all external access to the compute plane.
    * The compute plane doesn't require any ingress from the control-plane. Its the compute plane that intiates the connection to the control-plane. The `tfy-agent` connects to the control plane via outbound-only secure WebSocket connections, meaning no inbound ports need to be opened in your firewall
    * ControlPlane can be selfhosted within your VPC in our enterprise plan.
  </Accordion>

  <Accordion title="Network Security">
    * **Private Network Deployment**: Deploy control plane within your VPC for complete network isolation
    * **TLS/SSL Encryption**: All API communications encrypted with TLS 1.2+
    * **Firewall Integration**: Compatible with your existing firewall and network security policies
    * **VPC Isolation**: Support for deployment within private VPCs and on-premises networks
  </Accordion>

  <Accordion title="Container Security">
    * **Image Scanning**: Integration with container registries and security scanning tools
    * **Least Privilege**: Containers run with minimal required permissions
    * **Security Contexts**: Kubernetes security contexts for pod-level security controls
    * **Secret Injection**: Secure environment variable and secret management without exposing credentials

    <Note>
      Truefoundry internally follows a policy to scan all images and artifacts and not release software with any critical known vulnerabilities that have a fix. We have an SLA to fix all critical vulnerabilities within 10 days if a fix is available.
    </Note>
  </Accordion>
</AccordionGroup>

## Audit and Monitoring

<AccordionGroup>
  <Accordion title="Audit Logging">
    Comprehensive audit logging tracks all platform activities:

    * **Complete Activity History**: All user actions, resource changes, and API calls are logged
    * **Detailed Audit Trails**: Track who performed what action, when, and on which resource
    * **Export Capabilities**: Export audit logs to your logging platform (Splunk, Datadog, etc.)
    * **API Access**: Programmatic access to audit logs via Truefoundry API

    To read more about audit logging, please refer to our [Audit Logging](/docs/platform/audit-logging) documentation.
  </Accordion>

  <Accordion title="Security Monitoring">
    <Tabs>
      <Tab title="Self-Hosted Deployment">
        Since the infrastructure is self-hosted, all your existing monitoring,
        alerting and security solutions can be used to monitor the Truefoundry
        infrastructure.
      </Tab>

      <Tab title="Managed Deployment">
        We have a dedicated team that monitors the Truefoundry infrastructure 24/7
        and is available to be paged for any incident in your control-plane. We used
        centralized monitoring and alerting solutions to monitor the Truefoundry
        infrastructure.
      </Tab>
    </Tabs>
  </Accordion>
</AccordionGroup>

## AI Gateway Security

### Gateway Security Features

Truefoundry's AI Gateway provides enterprise-grade security for LLM access:

* **Authentication**: Multiple authentication methods including OIDC, JWT, and API keys
* **Authorization**: Role-based access control for models and endpoints
* **Guardrails Integration**: Support for content safety, PII detection, and prompt injection prevention
* **Request/Response Logging**: Secure logging of API requests and responses for compliance and debugging

### Content Safety

Integration with enterprise guardrail providers:

* **Content Filtering**: Azure AI Content Safety, OpenAI Moderation, and custom guardrails
* **PII Detection and Redaction**: Automatic detection and redaction of personally identifiable information
* **Prompt Injection Prevention**: Protection against prompt injection attacks
* **Toxicity Detection**: Real-time detection of harmful or inappropriate content

For more information on guardrails, see [AI Gateway Guardrails](/docs/ai-gateway/ecosystem).

## Incident Response and Business Continuity

<AccordionGroup>
  <Accordion title="High Availability">
    * **Globally Distributed**: Gateway deployed across 12+ regions and multiple cloud providers including AWS, GCP and Azure.
    * **Automated Failover**: Automatic routing to healthy regions in case of regional downtime
    * **Multi-Cloud Multi-region deployment of control-plane**: Control-Plane is deployed across three regions and three cloud providers in active passive mode and has disaster recovery enabled.
    * **Redundancy**: High availability configurations for critical components
  </Accordion>

  <Accordion title="Disaster Recovery">
    Truefoundry is deployed in multiple-regions and multiple-cloud providers. The control-plane is deployed in active passive mode and has disaster recovery enabled.
  </Accordion>

  <Accordion title="Status Monitoring">
    Status of all truefoundry hosted components can be tracked at [status.truefoundry.com](https://status.truefoundry.com).

    <img src="https://mintcdn.com/truefoundry/L6_ximWisGPdTTa9/images/docs/platform/status-monitoring-page.png?fit=max&auto=format&n=L6_ximWisGPdTTa9&q=85&s=cf31d0289cbf93794720d71878017ed6" alt="Status Page" width="3840" height="1866" data-path="images/docs/platform/status-monitoring-page.png" />

    You can also subscribe to the status updates by clicking on the "Get Updates" button in the top right corner of the status page.
  </Accordion>

  <Accordion title="Backup and Recovery">
    * **Configuration Backups**: Regular backups of platform configuration and metadata
    * **Disaster Recovery**: Comprehensive disaster recovery procedures
    * **Data Backup**: Integration with your backup solutions for data plane storage
    * **Recovery Procedures**: Documented procedures for rapid recovery from incidents
  </Accordion>
</AccordionGroup>

## Vulnerability Management, Logging and Monitoring

<Tabs>
  <Tab title="Self-Hosted Deployment">
    * A regular vulnerability scanning program in place to ensure that security issues are identified and remediated promptly, as outlined in the Technical Vulnerability Management section of the Operations Security Policy.
    * There is auditing and logging enabled across the systems and network devices to maintain visibility into system activity. AWS CloudTrail is integrated with CloudWatch as a centralized logging and monitoring solution. This setup correlates events and generates alerts for our team whenever a potential security incident is detected.
    * Additionally, 24/7 monitoring of security alerts is maintained through this system to ensure continuous oversight and rapid response.
  </Tab>

  <Tab title="Managed Deployment">
    In case of Managed Deployment, the platform will be deployed on the customer's cloud account. The customer's infra/platform team will be responsible for setting up security logging and monitoring
  </Tab>
</Tabs>

## Security Reporting

<AccordionGroup>
  <Accordion title="Vulnerability Reporting">
    If you discover a security vulnerability, please report it responsibly:

    * **Security Contact**: Report vulnerabilities to [security@truefoundry.com](mailto:security@truefoundry.com)
    * **Responsible Disclosure**: We follow responsible disclosure practices
    * **Response Time**: We aim to respond to security reports within 48 hours
  </Accordion>

  <Accordion title="Security Updates">
    * **Security Advisories**: Published on [trust.truefoundry.com](https://trust.truefoundry.com)
    * **Release Notes**: Security updates included in platform release notes
    * **Notifications**: Security-critical updates communicated to customers
  </Accordion>
</AccordionGroup>

## Data Privacy

<AccordionGroup>
  <Accordion title="Privacy Controls">
    * **Data Minimization**: Only collect and process data necessary for platform operation
    * **User Data Control**: Users can manage their own data and access permissions
    * **Data Retention**: Configurable data retention policies
    * **Right to Deletion**: Support for data deletion requests in compliance with privacy regulations
  </Accordion>

  <Accordion title="GDPR Compliance">
    Truefoundry is GDPR compliant and provides:

    * **Data Subject Rights**: Support for data access, rectification, and deletion requests
    * **Privacy by Design**: Privacy considerations built into platform architecture
    * **Data Processing Agreements**: Standard data processing agreements available
    * **Privacy Policy**: Comprehensive privacy policy available on our website
  </Accordion>
</AccordionGroup>

## Third-Party Security

<AccordionGroup>
  <Accordion title="Vendor Security">
    * **Security Assessments**: Regular security assessments of third-party vendors and integrations
    * **Secure Integrations**: All integrations use secure authentication and encryption
    * **Vendor Monitoring**: Continuous monitoring of vendor security practices
  </Accordion>

  <Accordion title="Integration Security">
    * **OAuth 2.0**: Secure OAuth-based integrations with cloud providers
    * **API Security**: All API integrations use secure authentication methods
    * **Credential Management**: Secure storage and rotation of third-party credentials
  </Accordion>
</AccordionGroup>

### Security Best Practices - Recommendations for Customers

1. **Enable SSO**: Use single sign-on with your identity provider for centralized authentication
2. **Implement Least Privilege**: Grant users only the minimum permissions required for their role
3. **Regular Access Reviews**: Periodically review and audit user access and permissions
4. **Enable Audit Logging**: Monitor audit logs regularly for suspicious activities
5. **Use Secret Management**: Store sensitive credentials in secret managers, not in code
6. **Network Segmentation**: Deploy Truefoundry in isolated network segments when possible
7. **Regular Updates**: Keep Truefoundry components updated to the latest versions
8. **Enable Guardrails**: Use content safety guardrails for AI Gateway deployments

<Note>
  Security is a shared responsibility. While Truefoundry provides secure
  infrastructure and platform features, customers are responsible for securing
  their applications, data, and access controls within their environment.
</Note>

## Sub-Processors

<Tabs>
  <Tab title="Self-Hosted Deployment">
    Here is the list of sub-processors:

    1. **FusionAuth**: Used for authentication
    2. **AWS**: Used for hosting, storage and security monitoring
    3. **Sentry**: Used for error and crash logs
    4. **Posthog**: Used for user event and analytics
  </Tab>

  <Tab title="Managed Deployment">
    Here is the list of sub-processors:

    1. **FusionAuth**: Used for authentication
    2. **Sentry**: Used for error and crash logs
    3. **Posthog**: Used for user event and analytics
  </Tab>
</Tabs>

We have signed a Data Processing Agreement (DPA) with all sub-processors.
Regarding data storage, data is stored in Europe for AWS, FusionAuth, and Posthog. Sentry data is stored in the US.
