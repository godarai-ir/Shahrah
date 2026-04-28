> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Playground

> Create, test, save and compare prompts in the Playground.

Truefoundry's LLM Playground is your interactive environment for developing, testing, and refining AI prompts and building agents. It provides a comprehensive interface to experiment with different models, configurations, and parameters before deploying your prompts in production.

## Getting Started with the Playground

Follow these steps to create and test your prompts:

1. **Choose a model** - Select from available AI models
2. **Add System message** - Define the AI's role and behavior
3. **Add User message** - Provide the input or query
4. **Configure parameters** - Adjust creativity, response length, guardrails, and structured output
5. **Run to test** - Execute your prompt and review the results
6. **Save the prompt** - Store as a reusable template for future use

<Frame>
  <img src="https://mintcdn.com/truefoundry/vx_CIFagyjPNEVU1/images/Screenshot2025-09-24at6.44.05PM.png?fit=max&auto=format&n=vx_CIFagyjPNEVU1&q=85&s=867cbc86f6fe84b18fbfe98368e94025" alt="TrueFoundry AI Gateway Playground interface showing model selection, system message input, user message input, and parameter configuration options" width="3600" height="2008" data-path="images/Screenshot2025-09-24at6.44.05PM.png" />
</Frame>

*Main Playground interface with model selection, message inputs, and parameter controls.*

## Managing Prompts

### Saving Your Prompt in the Playground

Once you're satisfied with your prompt's performance, save it for future use. Click the `Save` button and provide:

* **ML Repo** - Select the repository (see [ML Repos](/docs/introduction-to-ml-repo))
* **Name** - Give your prompt a descriptive name
* **Commit message** - Add a brief description of changes

<Frame>
  <img src="https://mintcdn.com/truefoundry/XUZt-P9vPWFRpzC2/images/Screenshot2025-09-24at4.39.45PM.png?fit=max&auto=format&n=XUZt-P9vPWFRpzC2&q=85&s=87e657698ccda43483687d9964ae4fc5" alt="TrueFoundry AI Gateway Playground with highlighted Save button in the top right corner after configuring prompt parameters" width="3600" height="2010" data-path="images/Screenshot2025-09-24at4.39.45PM.png" />
</Frame>

*Save button highlighted in the top right corner.*

<Frame>
  <img src="https://mintcdn.com/truefoundry/XUZt-P9vPWFRpzC2/images/Screenshot2025-09-24at4.40.38PM.png?fit=max&auto=format&n=XUZt-P9vPWFRpzC2&q=85&s=313b07b091036dbe7093f2d6ccd278c8" alt="Modal dialog for saving prompt template showing ML Repo selection, prompt name field, and commit message field" width="3598" height="2006" data-path="images/Screenshot2025-09-24at4.40.38PM.png" />
</Frame>

*Save dialog with ML repo selection, name field, and commit message.*

### Loading Existing Prompts

The Playground allows you to work with previously saved prompts, enabling you to:

* **Build on previous work** - Start from existing templates
* **Collaborate with teammates** - Share and iterate on prompts
* **Continue where you left off** - Resume prompt development

<Frame>
  <img src="https://mintcdn.com/truefoundry/XUZt-P9vPWFRpzC2/images/Screenshot2025-09-24at4.43.53PM.png?fit=max&auto=format&n=XUZt-P9vPWFRpzC2&q=85&s=2af2ce52cd6fdc3bffcf669ef2818d24" alt="TrueFoundry AI Gateway Playground Load Prompt interface showing dropdown to select existing prompt templates" width="3600" height="2008" data-path="images/Screenshot2025-09-24at4.43.53PM.png" />
</Frame>

*Load Prompt dropdown showing available saved templates.*

When you load a prompt in the Playground, all configurations are restored exactly as saved — including model parameters, guardrails, structured output, variables, and routing settings.

<Frame>
  <img src="https://mintcdn.com/truefoundry/XUZt-P9vPWFRpzC2/images/Screenshot2025-09-24at4.43.13PM.png?fit=max&auto=format&n=XUZt-P9vPWFRpzC2&q=85&s=344e2da5ad7cfe33a10f1d83c9ba8ff9" alt="TrueFoundry AI Gateway Playground displaying a loaded prompt with all previous configurations restored including system message, parameters, and settings" width="3582" height="2008" data-path="images/Screenshot2025-09-24at4.43.13PM.png" />
</Frame>

*Loaded prompt with all previous configurations restored.*
Once the prompt is loaded, you can edit, experiment with the prompt and save it again to create a new version.

## Advanced Playground Features

<AccordionGroup>
  <Accordion title="Input Variables - Dynamic Content" id="input-variables">
    Create reusable prompts in the Playground with variables like `{{customer_name}}` or `{{product_type}}`.
    You can then run these prompts by passing values. Please refer to [Using Prompts](/docs/ai-gateway/using-prompts) for more details on running prompts with variables.

    <Frame>
      <img src="https://mintcdn.com/truefoundry/YDqmXEab_DxBZe5e/images/prompt-variable.png?fit=max&auto=format&n=YDqmXEab_DxBZe5e&q=85&s=e32926d9efae31e8695fe0fe51037e54" alt="TrueFoundry AI Gateway Playground input variables configuration showing variable definition interface with placeholder examples like customer_name and product_type" width="3024" height="1704" data-path="images/prompt-variable.png" />
    </Frame>

    *Input variables configuration with dynamic placeholders.*
  </Accordion>

  <Accordion title="Guardrails - Configure Input and Output Guardrails" id="guardrails">
    Configure guardrails in the Playground to ensure model responses stay appropriate and on-topic.

    Learn more: [Guardrails and Security Overview](https://docs.truefoundry.com/gateway/guardrails-overview)

    <Frame>
      <img src="https://mintcdn.com/truefoundry/XUZt-P9vPWFRpzC2/images/Screenshot2025-09-24at5.02.42PM.png?fit=max&auto=format&n=XUZt-P9vPWFRpzC2&q=85&s=2645747827c081c779591b327fd0f937" alt="TrueFoundry AI Gateway Playground showing guardrails configuration panel with available guardrail options to select" width="3600" height="2000" data-path="images/Screenshot2025-09-24at5.02.42PM.png" />
    </Frame>

    *Guardrails configuration panel with available guardrail options.*

    <Frame>
      <img src="https://mintcdn.com/truefoundry/vx_CIFagyjPNEVU1/images/Screenshot2025-09-24at7.12.18PM.png?fit=max&auto=format&n=vx_CIFagyjPNEVU1&q=85&s=f3a6596d7a8c454814408775ac219051" alt="TrueFoundry AI Gateway Playground guardrails configuration interface showing selected guardrail policies applied to the prompt" width="3600" height="2008" data-path="images/Screenshot2025-09-24at7.12.18PM.png" />
    </Frame>

    *Selected guardrail policies applied to the prompt.*
  </Accordion>

  <Accordion title="Virtual models — load balancing and fallback" id="routing">
    Use the Playground with **virtual models** to exercise load balancing, fallback, and retries across gateway targets.

    Learn more: [Virtual Models](https://docs.truefoundry.com/docs/ai-gateway/virtual-model)

    <Frame>
      <img src="https://mintcdn.com/truefoundry/XUZt-P9vPWFRpzC2/images/Screenshot2025-09-24at5.58.41PM.png?fit=max&auto=format&n=XUZt-P9vPWFRpzC2&q=85&s=f8406c40d2dc685ade2701b6e63be44b" alt="TrueFoundry AI Gateway Playground routing configuration panel showing model selection and routing policy options" width="3580" height="2002" data-path="images/Screenshot2025-09-24at5.58.41PM.png" />
    </Frame>

    *Routing configuration from Playground*

    <Frame>
      <img src="https://mintcdn.com/truefoundry/XUZt-P9vPWFRpzC2/images/Screenshot2025-09-24at5.59.56PM.png?fit=max&auto=format&n=XUZt-P9vPWFRpzC2&q=85&s=cb23aab22674aaede4ee3d64a656c9f5" alt="TrueFoundry AI Gateway Playground routing configuration interface showing configured routing policies with fallback and load balancing settings" width="3600" height="2002" data-path="images/Screenshot2025-09-24at5.59.56PM.png" />
    </Frame>

    *Configured routing policies with fallback and load balancing settings.*
  </Accordion>

  <Accordion title="Structured Output - Configure Response Format for LLM Responses" id="structured-output">
    Configure the Playground to get responses in specific formats such as JSON.

    <Frame>
      <img src="https://mintcdn.com/truefoundry/XUZt-P9vPWFRpzC2/images/Screenshot2025-09-24at5.09.03PM.png?fit=max&auto=format&n=XUZt-P9vPWFRpzC2&q=85&s=d9ee05f7a9fdb59353d4e842f65b7a38" alt="TrueFoundry AI Gateway Playground structured output configuration showing schema definition interface with agent assistance" width="3600" height="2010" data-path="images/Screenshot2025-09-24at5.09.03PM.png" />
    </Frame>

    *Structured output configuration with schema definition interface.*

    <Frame>
      <img src="https://mintcdn.com/truefoundry/XUZt-P9vPWFRpzC2/images/Screenshot2025-09-24at5.15.22PM.png?fit=max&auto=format&n=XUZt-P9vPWFRpzC2&q=85&s=e8c0fedd95c3b5a8165820d3c44cc080" alt="TrueFoundry AI Gateway Playground showing generated JSON schema for structured output with defined properties and data types" width="3596" height="2000" data-path="images/Screenshot2025-09-24at5.15.22PM.png" />
    </Frame>

    *Generated JSON schema with defined properties and data types.*
  </Accordion>

  <Accordion title="MCP Servers - Add Tools and MCP Servers" id="mcp-servers">
    Use the Playground to extend AI capabilities with external tools and services.

    <Frame>
      <img src="https://mintcdn.com/truefoundry/XUZt-P9vPWFRpzC2/images/Screenshot2025-09-24at5.18.10PM.png?fit=max&auto=format&n=XUZt-P9vPWFRpzC2&q=85&s=bd3df0384f07b4332b141f3136bad692" alt="TrueFoundry AI Gateway Playground MCP Servers configuration panel showing available MCP server options and connection settings" width="3596" height="2006" data-path="images/Screenshot2025-09-24at5.18.10PM.png" />
    </Frame>

    *MCP Servers configuration panel with available tools.*

    <Frame>
      <img src="https://mintcdn.com/truefoundry/XUZt-P9vPWFRpzC2/images/Screenshot2025-09-24at5.53.39PM.png?fit=max&auto=format&n=XUZt-P9vPWFRpzC2&q=85&s=00679c250f1dad6357b81c3eca584a94" alt="TrueFoundry AI Gateway Playground showing configured MCP Servers with connected external tools and services for enhanced AI capabilities" width="3600" height="2006" data-path="images/Screenshot2025-09-24at5.53.39PM.png" />
    </Frame>

    *Configured MCP Servers with connected tools.*
  </Accordion>
</AccordionGroup>

## Comparison View (Coming Soon)

Compare multiple prompt versions and their configurations in the playground side by side.
