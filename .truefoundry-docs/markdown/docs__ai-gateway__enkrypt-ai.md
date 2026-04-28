> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Enkrypt AI Guardrail Integration

> Learn how to use enkrypt ai with TrueFoundry AI Gateway, including setup steps, use cases, and production-ready examples.

This guide explains how to integrate [Enkrypt AI Guardrail](https://www.enkryptai.com/) with TrueFoundry to enhance the safety and compliance of your LLM applications.

## What is Enkrypt AI?

Enkrypt AI is an AI safety and security platform that provides comprehensive guardrails for Large Language Model (LLM) applications.

### Key Features of Enkrypt AI

1. **Multi-Layered Content Security**: Enkrypt AI provides comprehensive content analysis including toxicity detection, PII identification, prompt injection prevention, and bias detection. The platform uses advanced ML models to analyze content across multiple dimensions, ensuring [robust AI safety](https://www.enkryptai.com/) and preventing harmful or inappropriate outputs.

2. **Flexible Validation and Mutation Operations**: Support for both validation-only and mutation operations allows for different security approaches. Validation operations run in parallel for performance, while mutation operations can sequentially process and redact sensitive content like PII, providing flexible content handling based on your security requirements.

3. **Real-Time Threat Protection**: Advanced detection capabilities for prompt injection attacks, sponge attacks, and other adversarial inputs that could compromise AI system integrity. Enkrypt AI's [security guardrails](https://www.enkryptai.com/) continuously monitor and protect against evolving AI-specific threats while maintaining system performance and user experience.

### Adding Enkrypt AI Guardrail Integration

To add Enkrypt AI Guardrail to your TrueFoundry setup, follow these steps:

**Fill in the Guardrails Group Form**

* **Name**: Enter a name for your guardrails group.
* **Collaborators**: Add collaborators who will have access to this group.
* **Enkrypt AI Guardrail Config**:
  * **Name**: Enter a name for the Enkrypt AI Guardrail configuration.
  * **Guardrail Name**: The unique name of your Enkrypt AI guardrail. You can create a new guardrail in the Enkrypt AI portal and get the name from there.
  * **Operation**: The operation type to use for the Guardrail.
    * **Validate**: Guardrails with this operation are used to validate requests. These guardrails are run in parallel.
    * **Mutate**: Guardrails with this operation can both validate and mutate requests. Mutate guardrails are run sequentially.
* **Enkrypt AI Authentication Data**:
  * **API Key**: The API key for the Enkrypt AI.\
    This key is required to authenticate requests to Enkrypt AI. You can obtain it from the Enkrypt AI portal by navigating to your Enkrypt AI resource, then selecting "Keys and Endpoint" in the resource menu. Ensure you keep this key secure, as it grants access to your Enkrypt AI resource.

<Frame caption="Fill in the Enkrypt AI Guardrail Form">
  <img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/enkrypt-ai-guardrail.jpeg?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=1b601f55c571193e23a9103e96d4e337" alt="TrueFoundry interface for configuring Enkrypt AI Guardrail with fields for name, operation type, guardrail name, and API key" width="1280" height="603" data-path="images/enkrypt-ai-guardrail.jpeg" />
</Frame>

#### Response Structure

The Enkrypt AI Guardrail API returns a response with the following structure:

<AccordionGroup>
  <Accordion title="Example Response: PII Detection">
    This is an example response from Enkrypt AI Guardrail, where PII is detected and anonymized (redacted) in the output. In this example, we have configured topics as "Phone". The response includes a summary of detected categories, detailed detection results, and the redacted text.

    ```json lines theme={"dark"}
      {
        "summary": {
          "keyword_detected": 0,
          "nsfw": 0,
          "toxicity": [],
          "injection_attack": 0,
          "on_topic": 1,
          "pii": 1,
          "sponge_attack": 0,
          "bias": 0
        },
        "details": {
          "keyword_detector": {
            "detected_keywords": [],
            "detected_counts": {},
            "redacted_text": "Hi I am Hrihtik, I like iPhone"
          },
          "nsfw": {
            "sfw": 0.5073823928833008,
            "nsfw": 0.43124663829803467
          },
          "toxicity": {
            "toxicity": 0.002005422255024314,
            "severe_toxicity": 9.976350702345371e-05,
            "obscene": 0.00023746110673528165,
            "threat": 0.00011596245894907042,
            "insult": 0.0002040357794612646,
            "identity_hate": 0.00016701706044841558
          },
          "injection_attack": {
            "safe": "0.812537",
            "attack": "0.187463",
            "most_unsafe_content": "Hi I am Hrihtik"
          },
          "topic_detector": {
            "Phone": 0.6327297687530518
          },
          "pii": {
            "entities": {
              "pii": {},
              "secrets": {},
              "ip_address": {},
              "url": {},
              "email_address": {},
              "person": {
                "<PERSON_0>": "Hrihtik"
              },
              "phone_number": {},
              "location": {},
              "credit_card": {},
              "social_security_number": {}
            },
            "text": "Hi I am <PERSON_0>",
            "key": "6c2fb109f5044318af6904d9a2bac5f1"
          },
          "sponge_attack": {
            "sponge_attack": 0
          },
          "bias": {
            "bias_detected": false,
            "debiased_text": "Hi, I am Hrihtik."
          }
        }
      }
    ```
  </Accordion>
</AccordionGroup>

#### Validation Logic

TrueFoundry uses the Enkrypt AI Guardrail response to determine content safety and compliance:

* If the operation is set to `Validate` and a policy violation is detected, the request will be blocked and a 400 error is returned.
* If the operation is set to `Mutate` and PII or a configured keyword is detected (and no other violation is detected), the detected information will be redacted and the redacted message will be returned in the response.
* For other types of violations under `Mutate`, the request will be blocked and a 400 error is returned.
* If no violations are detected, the original content is passed through unchanged.
