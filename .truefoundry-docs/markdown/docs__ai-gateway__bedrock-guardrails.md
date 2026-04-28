> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# AWS Bedrock Guardrail Integration

> Learn how to use Amazon Bedrock guardrails via the TrueFoundry AI Gateway.

This guide explains how to integrate AWS Bedrock Guardrails with TrueFoundry to enhance the safety and compliance of your LLM applications.

## What is AWS Bedrock Guardrails?

AWS Bedrock Guardrails is Amazon's managed service that helps implement safeguards for generative AI applications.

### Key Features of AWS Bedrock Guardrails

1. **Content Filters**: Detect and filter harmful text or image content in user inputs and model responses. You can adjust filter strength for predefined categories such as hate speech, insults, sexual content, violence, misconduct, and prompt attacks. [AWS Bedrock Guardrails](https://aws.amazon.com/bedrock/guardrails/) provides configurable safeguards that help you build responsible AI applications at scale.

2. **Sensitive Information Filters**: Identify and redact personally identifiable information (PII) in user inputs and model outputs. This feature supports predefined PII types and allows for custom regex patterns to detect sensitive data, helping protect user privacy and maintain compliance with data protection regulations.

3. **Contextual Grounding Checks**: Detect and filter hallucinations in model responses by verifying their grounding in source information and relevance to user queries. This helps ensure the factual accuracy and trustworthiness of AI-generated content through [AWS's advanced validation mechanisms](https://docs.aws.amazon.com/bedrock/latest/userguide/guardrails.html).

#### How to create a guardrail in AWS Bedrock

To create guardrails in AWS Bedrock, follow these steps:

1. **Sign in to the AWS Management Console**
   * Go to the [AWS Bedrock Console](https://console.aws.amazon.com/bedrock/).

2. **Navigate to Guardrails**
   * In the Bedrock console sidebar, select **Guardrails**.

3. **Create a New Guardrail**
   * Click on **Create guardrail**.
   * Enter a **name** and an optional **description** for your guardrail.

4. **Define Guardrail Policies**
   * Configure the policies you want to enforce. This may include:
     * **Content filters**: Block or allow specific types of content (e.g., hate speech, violence, PII).
     * **Add denied topics**: Add topics that you want to deny.
     * **Add word filters**: Add words that you want to filter.
     * **Add sensitive information filters**: Add sensitive information that you want to filter.
     * **Add contextual grounding checks**: Add contextual grounding checks that you want to enforce.
       <Frame caption="Fill in the Guardrail Form">
         <img src="https://mintcdn.com/truefoundry/OHzlp6GY5G-JfKle/images/bedrock-guardrail-1.jpeg?fit=max&auto=format&n=OHzlp6GY5G-JfKle&q=85&s=1de1cf6e07b3044981d719539ffcf468" alt="AWS Bedrock console showing guardrail creation form with content filter options and policy configuration" width="1280" height="839" data-path="images/bedrock-guardrail-1.jpeg" />
       </Frame>

5. **Review and Save**
   * Review your guardrail configuration.
   * Click **Create** to save the guardrail.

6. **Note the Guardrail Identifier and Version**
   * After creation, note the **Guardrail Identifier** and **Version**. You will need these values to integrate with TrueFoundry.

7. **(Optional) Test Your Guardrail**
   * Use the Bedrock console to test your guardrail with sample inputs and verify that it enforces your policies as expected.

For more details, refer to the [AWS Bedrock Guardrails documentation](https://docs.aws.amazon.com/bedrock/latest/userguide/guardrails.html).

Ensure these prerequisites are met to smoothly integrate AWS Bedrock Guardrails into your TrueFoundry environment.

### Adding AWS Bedrock Guardrail Integration

To add AWS Bedrock Guardrails to your TrueFoundry setup, follow these steps:

**Fill in the Guardrails Group Form**

* **Name**: Enter a name for your guardrails group.
* **Collaborators**: Add collaborators who will have access to this group.
* **AWS Bedrock Guardrail Config**:
  * **Name**: Enter a name for the AWS Bedrock Guardrail configuration.
  * **Operation**: The operation type to use for the Guardrail.
    * **Validate**: Guardrails with this operation are used to validate requests. These guardrails are run in parallel.
    * **Mutate**: Guardrails with this operation can both validate and mutate requests. Mutate guardrails are run sequentially.
  * **Guardrail Identifier**: Provide the identifier for the Bedrock guardrail.
  * **Guardrail Version**: Specify the version of the guardrail.
* **AWS Authentication Data**:\
  You can authenticate with AWS in one of two ways:
  1. **Access Key Based Authentication**
     * **Access Key ID** and **Secret Access Key**: Enter the AWS credentials that have permission to invoke Bedrock Guardrails.
     * <Note>
         For security and best practices, use an IAM user or role with the minimum required permissions to access Bedrock Guardrails. Avoid using root credentials.
       </Note>
  2. **Assumed Role Based Authentication**
     * **Role ARN**: Provide the Amazon Resource Name (ARN) of the IAM role to assume for accessing Bedrock Guardrails.
     * **External ID (if required)**: Enter the external ID if your organization requires it for cross-account access.
     * **Region**: Specify the AWS region where your Bedrock Guardrail is deployed.
     * <Note>
         The system will use AWS STS to assume the provided role. Ensure the role has the necessary permissions for Bedrock Guardrails and that trust relationships are configured correctly.
       </Note>
  <Note>
    You need `bedrock:ApplyGuardrail` permission: This permission allows your IAM entity to call the ApplyGuardrail API, which is the core function for applying guardrails to input or output data.
  </Note>

<Frame caption="Fill in the AWS Bedrock Guardrail Form">
  <img src="https://mintcdn.com/truefoundry/OHzlp6GY5G-JfKle/images/bedrock-guardrail.jpeg?fit=max&auto=format&n=OHzlp6GY5G-JfKle&q=85&s=49805f6d37a54623beaff9897635167b" width="1280" height="805" data-path="images/bedrock-guardrail.jpeg" />
</Frame>

### How AWS Bedrock Guardrails Validation Works

When you integrate AWS Bedrock Guardrails with TrueFoundry, the system sends the last message to AWS Bedrock's ApplyGuardrail API and receives a response that indicates whether the content violates any configured policies.

#### Response Structure

The AWS Bedrock ApplyGuardrail API returns a response with the following structure:

<AccordionGroup>
  <Accordion title="Example Response: Validate Operation (PII Detection)">
    This is an example response from AWS Bedrock for a <strong>Validate</strong> operation, where PII (Personally Identifiable Information) is detected and blocked.

    ```json lines theme={"dark"}
    {
      "action": "GUARDRAIL_INTERVENED",
      "actionReason": "Guardrail blocked.",
      "assessments": [
          {
              "invocationMetrics": {
                  "guardrailCoverage": {
                      "textCharacters": {
                          "guarded": 39,
                          "total": 39
                      }
                  },
                  "guardrailProcessingLatency": 264,
                  "usage": {
                      "contentPolicyImageUnits": 0,
                      "contentPolicyUnits": 0,
                      "contextualGroundingPolicyUnits": 0,
                      "sensitiveInformationPolicyFreeUnits": 0,
                      "sensitiveInformationPolicyUnits": 1,
                      "topicPolicyUnits": 0,
                      "wordPolicyUnits": 0
                  }
              },
              "sensitiveInformationPolicy": {
                  "piiEntities": [
                      {
                          "action": "BLOCKED",
                          "detected": true,
                          "match": "Hrithik",
                          "type": "NAME"
                      }
                  ]
              }
          }
      ],
      "blockedResponse": "Sorry, the model cannot answer this question.",
      "guardrailCoverage": {
          "textCharacters": {
              "guarded": 39,
              "total": 39
          }
      },
      "output": [
          {
              "text": "Sorry, the model cannot answer this question."
          }
      ],
      "outputs": [
          {
              "text": "Sorry, the model cannot answer this question."
          }
      ],
      "usage": {
          "contentPolicyImageUnits": 0,
          "contentPolicyUnits": 0,
          "contextualGroundingPolicyUnits": 0,
          "sensitiveInformationPolicyFreeUnits": 0,
          "sensitiveInformationPolicyUnits": 1,
          "topicPolicyUnits": 0,
          "wordPolicyUnits": 0
      }
    }
    ```
  </Accordion>

  <Accordion title="Example Response: Mutate Operation (PII Redaction)">
    This is an example response from AWS Bedrock for a <strong>Mutate</strong> operation, where PII is detected and anonymized (redacted) in the output.

    ```json lines theme={"dark"}
    {
        "action": "GUARDRAIL_INTERVENED",
        "actionReason": "Guardrail masked.",
        "assessments": [
            {
                "invocationMetrics": {
                    "guardrailCoverage": {
                        "textCharacters": {
                            "guarded": 39,
                            "total": 39
                        }
                    },
                    "guardrailProcessingLatency": 160,
                    "usage": {
                        "contentPolicyImageUnits": 0,
                        "contentPolicyUnits": 0,
                        "contextualGroundingPolicyUnits": 0,
                        "sensitiveInformationPolicyFreeUnits": 0,
                        "sensitiveInformationPolicyUnits": 1,
                        "topicPolicyUnits": 0,
                        "wordPolicyUnits": 0
                    }
                },
                "sensitiveInformationPolicy": {
                    "piiEntities": [
                        {
                            "action": "ANONYMIZED",
                            "detected": true,
                            "match": "Hrithik",
                            "type": "NAME"
                        }
                    ]
                }
            }
        ],
        "guardrailCoverage": {
            "textCharacters": {
                "guarded": 39,
                "total": 39
            }
        },
        "output": [
            {
                "text": "Hi I am {NAME}, Can you spell my name?"
            }
        ],
        "outputs": [
            {
                "text": "Hi I am {NAME}, Can you spell my name?"
            }
        ],
        "usage": {
            "contentPolicyImageUnits": 0,
            "contentPolicyUnits": 0,
            "contextualGroundingPolicyUnits": 0,
            "sensitiveInformationPolicyFreeUnits": 0,
            "sensitiveInformationPolicyUnits": 1,
            "topicPolicyUnits": 0,
            "wordPolicyUnits": 0
        }
    }
    ```
  </Accordion>
</AccordionGroup>

#### Validation Logic

TrueFoundry relies on the Bedrock Guardrails response to determine content safety:

* If Bedrock marks any filter as blocked, the content will also be blocked in TrueFoundry.
* In **validate** mode, requests are blocked if Bedrock detects any violations.
* In **mutate** mode, TrueFoundry uses the modified output provided by Bedrock, which may have sensitive or violating content removed or anonymized.

This ensures your content is handled according to Bedrock's guardrail decisions and your configured policies.
