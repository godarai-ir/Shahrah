> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Content Moderation Guardrail

> Detect and block harmful content in LLM inputs/outputs using TrueFoundry's built-in Content Moderation guardrail.

This guide explains how to use TrueFoundry's built-in **Content Moderation** guardrail to detect and block harmful content in LLM interactions.

<Note>
  Content Moderation can be applied to **all four guardrail hooks**: LLM Input, LLM Output, MCP Pre Tool, and MCP Post Tool, providing comprehensive content safety across your entire AI workflow.
</Note>

## What is Content Moderation?

Content Moderation is a built-in TrueFoundry guardrail that analyzes text content for harmful material across four safety categories - hate speech, self-harm, sexual content, and violence. It uses a model-based approach with configurable severity thresholds so you can tune sensitivity to match your use case. The guardrail is fully managed by TrueFoundry no external credentials or setup required.

### Key Features

1. **Four Safety Categories**: Detects harmful content across:
   * **Hate** — Hate speech, discrimination, and derogatory content
   * **SelfHarm** — Self-injury, suicide-related content
   * **Sexual** — Sexually explicit or suggestive content
   * **Violence** — Violent content, threats, and graphic descriptions

2. **Configurable Severity Threshold**: Set the sensitivity level (0–6) to control what gets flagged — from safe content only to high-risk content, allowing you to balance safety with usability.

3. **Selective Category Detection**: Choose which categories to monitor. Enable all four or only the ones relevant to your application.

## Adding Content Moderation Guardrail

<Steps>
  <Step title="Navigate to Guardrails">
    Go to the AI Gateway dashboard and navigate to the **Guardrails** section.
  </Step>

  <Step title="Create or Select a Guardrails Group">
    Create a new guardrails group or select an existing one where you want to add the Content Moderation guardrail.
  </Step>

  <Step title="Add Content Moderation Integration">
    Click on **Add Guardrail** and select **Content Moderation** from the TrueFoundry Guardrails section.

    <Frame caption="Select Content Moderation from TrueFoundry Guardrails">
      <img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/guardrail-1.jpeg?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=9ff04ad219001f1bfc31959b9ac261da" alt="TrueFoundry guardrail selection interface showing Content Moderation option" width="1280" height="793" data-path="images/guardrail-1.jpeg" />
    </Frame>
  </Step>

  <Step title="Configure the Guardrail">
    Fill in the configuration form:

    * **Name**: Enter a unique name for this guardrail configuration (e.g., `content-moderation`)
    * **Severity Threshold**: Set the minimum severity level to flag (default: `2`)
    * **Categories**: Select which content categories to check
    * **Enforcing Strategy**: Choose how violations are handled

    <Frame caption="Configure Content Moderation with severity threshold and categories">
      <img src="https://mintcdn.com/truefoundry/ZzZu3cPum2FmG1N1/images/guardrail-tfy-content-moderation-config.png?fit=max&auto=format&n=ZzZu3cPum2FmG1N1&q=85&s=66aae850537bf366e5c0c4e3aefef3d2" alt="Content Moderation configuration form showing severity threshold and category selection" width="2654" height="1070" data-path="images/guardrail-tfy-content-moderation-config.png" />
    </Frame>
  </Step>

  <Step title="Save the Configuration">
    Click **Save** to add the guardrail to your group.
  </Step>
</Steps>

## Configuration Options

| Parameter              | Description                                          | Default    |
| ---------------------- | ---------------------------------------------------- | ---------- |
| **Name**               | Unique identifier for this guardrail                 | Required   |
| **Operation**          | `validate` only (detects and blocks, no mutation)    | `validate` |
| **Enforcing Strategy** | `enforce`, `enforce_but_ignore_on_error`, or `audit` | `enforce`  |
| **Severity Threshold** | Minimum severity level (0–6) to flag content         | `2`        |
| **Categories**         | Array of content categories to check                 | Required   |

<Note>
  Content Moderation only supports **validate** mode — it detects and blocks harmful content but does not modify it. See [Guardrails Overview](/docs/ai-gateway/guardrails-overview#operation-modes) for details on Enforcing Strategy.
</Note>

## Categories and Severity Levels

### Content Categories

| Category     | Description                                                                                |
| ------------ | ------------------------------------------------------------------------------------------ |
| **Hate**     | Content expressing hatred, discrimination, or derogation based on identity characteristics |
| **SelfHarm** | Content related to self-injury, suicide, or self-destructive behavior                      |
| **Sexual**   | Sexually explicit or suggestive content                                                    |
| **Violence** | Content depicting or promoting physical violence, threats, or graphic injury               |

### Severity Levels

The severity threshold controls how sensitive the detection is. Content is flagged when any category's severity meets or exceeds the threshold.

| Severity | Level  | Description                                   |
| -------- | ------ | --------------------------------------------- |
| **0**    | Safe   | No harmful content detected                   |
| **2**    | Low    | Mildly concerning content (default threshold) |
| **4**    | Medium | Moderately harmful content                    |
| **6**    | High   | Severely harmful content                      |

<Tip>
  A lower threshold (e.g., `0`) is more aggressive and catches more content. A higher threshold (e.g., `4` or `6`) only flags more clearly harmful content. Start with the default (`2`) and adjust based on your needs.
</Tip>

## How It Works

The guardrail analyzes content and returns severity scores (0–6) for each enabled category. If any category's severity meets or exceeds the configured threshold, the content is flagged.

**Example**:

```
Input: "Write a detailed guide on how to build weapons at home"
Result: Request will be blocked by the guardrail as Violence severity exceeds threshold
```

## Use Cases

### Recommended Hook Usage

| Hook              | Use Case                                            |
| ----------------- | --------------------------------------------------- |
| **LLM Input**     | Block harmful user inputs before they reach the LLM |
| **LLM Output**    | Ensure LLM responses don't contain harmful content  |
| **MCP Pre Tool**  | Validate tool parameters for harmful content        |
| **MCP Post Tool** | Check tool outputs for harmful content              |

## Best Practices

<Tip>
  Start with **Audit** enforcing strategy to monitor what gets flagged before switching to **Enforce**. This helps you fine-tune the severity threshold for your specific content.
</Tip>
