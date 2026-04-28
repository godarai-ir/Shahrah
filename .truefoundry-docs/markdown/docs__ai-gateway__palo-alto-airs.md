> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Palo Alto Prisma AIRS Integration

> Learn how to use palo alto airs with TrueFoundry AI Gateway, including setup steps, use cases, and production-ready examples.

This guide explains how to integrate Palo Alto Prisma AIRS with TrueFoundry to enhance the safety and compliance of your LLM applications.

## What is Palo Alto Prisma AIRS?

Palo Alto Prisma AIRS (AI Runtime Security) is an enterprise-grade AI security platform that provides comprehensive protection for AI applications and LLM deployments.

### Key Features of Palo Alto Prisma AIRS

1. **Real-Time AI Threat Detection**: Prisma AIRS provides continuous monitoring and analysis of AI interactions, detecting malicious prompts, injection attacks, and suspicious behavior patterns in real-time. The platform uses advanced [AI runtime security](https://www.paloaltonetworks.com/network-security/ai-runtime-security) capabilities to identify and block threats before they can compromise your AI systems.

2. **Comprehensive Policy Enforcement**: Implement and enforce granular security policies across your entire AI infrastructure, including content filtering, access controls, and usage restrictions. The platform ensures compliance with industry standards and organizational policies while maintaining the performance and functionality of your AI applications.

3. **Enterprise-Grade Compliance and Auditing**: Built-in compliance frameworks and detailed audit trails help organizations meet regulatory requirements and maintain governance over AI deployments. Prisma AIRS provides comprehensive reporting and analytics to track AI usage, security incidents, and policy violations across your [Palo Alto Networks security ecosystem](https://www.paloaltonetworks.com/).

### Adding Palo Alto Prisma AIRS Integration

To add Palo Alto Prisma AIRS to your TrueFoundry setup, follow these steps:

**Fill in the Guardrails Group Form**

* **Name**: Enter a name for your guardrails group.
* **Collaborators**: Add collaborators who will have access to this group.
* **Palo Alto Prisma AIRS Config**:
  * **Name**: Enter a name for the Palo Alto Prisma AIRS configuration.
  * **Profile Name**: The unique name of your Palo Alto Prisma AIRS profile. You can create a new profile in the Palo Alto Prisma AIRS portal and get the name from there.
* **Palo Alto Prisma AIRS Authentication Data**:
  * **API Key**: The API key for the Palo Alto Prisma AIRS.\
    This key is required to authenticate requests to Palo Alto Prisma AIRS. You can obtain it from the Palo Alto Prisma AIRS portal by navigating to your Palo Alto Prisma AIRS resource, then selecting "Keys and Endpoint" in the resource menu. Ensure you keep this key secure, as it grants access to your Palo Alto Prisma AIRS resource.

<Frame caption="Fill in the Palo Alto Prisma AIRS Form">
  <img src="https://mintcdn.com/truefoundry/iMid4yIOHvzf4Z4V/images/palo-alto-airs.jpeg?fit=max&auto=format&n=iMid4yIOHvzf4Z4V&q=85&s=3425705c5f2bdb4b3a6df5b6d1d19310" alt="TrueFoundry interface for configuring Palo Alto Prisma AIRS with fields for name, profile name, and API key" width="1280" height="504" data-path="images/palo-alto-airs.jpeg" />
</Frame>

#### Response Structure

The Palo Alto Prisma AIRS API returns a response with the following structure:

<AccordionGroup>
  <Accordion title="Example Response: Injection Detected">
    This is an example response from Palo Alto Prisma AIRS, where injection is detected. This will be blocked and a 400 error will be returned.

    ```json lines theme={"dark"}
      {
        "report_id": "R82f1e879-0000-49af-9345-da907431c08f",
        "scan_id": "82f1e879-0000-49af-9345-da907431c08f",
        "tr_id": 1234,
        "profile_id": "12345678-0000-1234-1234-123456789012",
        "profile_name": "ai-dummy-profile",
        "category": "malicious",
        "action": "block",
        "prompt_detected": {
          "url_cats": true,
          "dlp": true,
          "injection": true,
          "toxic_content": true,
          "malicious_code": true,
          "agent": true,
          "topic_violation": true
        },
        "response_detected": {
          "url_cats": true,
          "dlp": true,
          "db_security": true,
          "toxic_content": true,
          "malicious_code": true,
          "agent": true,
          "ungrounded": true,
          "topic_violation": true
        },
        "prompt_detection_details": {
          "topic_guardrails_details": {
            "allowed_topics": [
              "string"
            ],
            "blocked_topics": [
              "string"
            ]
          }
        },
        "response_detection_details": {
          "topic_guardrails_details": {
            "allowed_topics": [
              "string"
            ],
            "blocked_topics": [
              "string"
            ]
          }
        },
        "created_at": "2024-07-29T15:51:28.071Z",
        "completed_at": "2024-07-29T15:51:28.071Z"
      }
    ```
  </Accordion>
</AccordionGroup>

#### Validation Logic

TrueFoundry uses the Palo Alto Prisma AIRS response to determine content safety and compliance:

* If the `response.action` is `block`, the request will be blocked and a 400 error is returned.
* If the `response.action` is `allow`, the request will be allowed to proceed.
