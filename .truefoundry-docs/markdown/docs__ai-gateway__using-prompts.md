> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Using Prompts

> Learn how to use using prompts with TrueFoundry AI Gateway, including setup steps, use cases, and production-ready examples.

Users can use prompts in two ways:

1. **Prompt Run API** - Server-side rendering where the Gateway handles prompt rendering
2. **SDK** - Client-side rendering where you fetch and render the prompt locally

In both the cases, you need to get the Prompt Version FQN to use the prompt.

<Accordion title="How to get the Prompt Version FQN" id="get-prompt-version-fqn">
  You can get the FQN of the prompt version from the Prompt Registry as shown in the image below.

  <Frame>
    <img src="https://mintcdn.com/truefoundry/YDqmXEab_DxBZe5e/images/prompt-copy-fqn.png?fit=max&auto=format&n=YDqmXEab_DxBZe5e&q=85&s=4742a66140310a50023ab100c888b32a" alt="Getting FQN from prompt registry" width="3024" height="1134" data-path="images/prompt-copy-fqn.png" />
  </Frame>
</Accordion>

## Prompt Run API (server-side rendering)

Use the AI Gateway by passing the prompt version FQN in the request body. The Gateway will internally render the prompt and execute it.

**Parameters:**

* `prompt_version_fqn` - The fully qualified name of your prompt version
* `prompt_variables`\[Optional] - Variables to substitute in your prompt template

<Note>
  **Important considerations:**

  * If the prompt version doesn't have a model configured, pass the model in the request body using the `model` parameter
  * If you specify a model in the request body and your prompt version already has a model configured, the request body model takes precedence
  * Any messages passed in the request body will be appended to the messages defined in the prompt version
</Note>

You can get the exact code snippet you need for running the prompt from the playground or the Code Snippet button in the Playground or Prompt Registry.

<Accordion title="How to get code snippets for prompts" id="gateway-snippets">
  ### From Playground

  1. Navigate to the Playground in your TrueFoundry dashboard
  2. Select and load the prompt as described in the document here [Playground](/docs/ai-gateway/prompt-playground#loading-existing-prompts)
  3. Click on the "Code Snippet" button to generate the exact code you need
  4. Copy the generated code snippet for your application

  <Frame>
    <img src="https://mintcdn.com/truefoundry/YDqmXEab_DxBZe5e/images/prompt-playground-code.png?fit=max&auto=format&n=YDqmXEab_DxBZe5e&q=85&s=71ae2d556e349b5e0796a1df9ba1e19e" alt="Getting code snippet from playground" width="3024" height="1710" data-path="images/prompt-playground-code.png" />
  </Frame>

  ### From Prompt Registry

  1. Go to the Prompt Registry in your TrueFoundry dashboard
  2. Find and select the prompt version you want to use
  3. Click on the "Code Snippet" button in the prompt details
  4. Copy the generated code snippet with the correct prompt version FQN

  <Frame>
    <img src="https://mintcdn.com/truefoundry/YDqmXEab_DxBZe5e/images/prompt-use-gateway-code.png?fit=max&auto=format&n=YDqmXEab_DxBZe5e&q=85&s=07a99c6bf79f37b9b828847cc5caadc3" alt="Getting code snippet from prompt registry" width="3024" height="1704" data-path="images/prompt-use-gateway-code.png" />
  </Frame>
</Accordion>

<CodeGroup>
  ```python TrueFoundry AI Gateway (Recommended) theme={"dark"}
  from openai import OpenAI

  client = OpenAI(
      api_key="your-tfy-api-key",
      base_url="{GATEWAY_BASE_URL}"
  )

  stream = client.chat.completions.create(
      messages=[],
      model="",
      stream=True,
      extra_headers={
          "X-TFY-METADATA": '{"your_custom_key":"your_custom_value"}',
          "X-TFY-LOGGING-CONFIG": '{"enabled": true}',
      },
      extra_body={
          "prompt_version_fqn": "chat_prompt:truefoundry/default/my-second-prompt:1",
          "prompt_variables": {
              "name": "John Doe",
              "age": "30"
          }
      },
  )

  for chunk in stream:
      if (
          chunk.choices
          and len(chunk.choices) > 0
          and chunk.choices[0].delta.content is not None
      ):
          print(chunk.choices[0].delta.content, end="", flush=True)
  ```
</CodeGroup>

## SDK (client-side rendering)

You can fetch the template and render it client-side and then make the request to the model with whatever OpenAI client you are using.

<Accordion title="How to get code snippets for prompts" id="sdk-snippets">
  1. Go to the Prompt Registry in your TrueFoundry dashboard
  2. Find and select the prompt version you want to use
  3. Click on the "Use Via SDK" button to view the code snippet

  <Frame>
    <img src="https://mintcdn.com/truefoundry/YDqmXEab_DxBZe5e/images/prompt-use-sdk-code.png?fit=max&auto=format&n=YDqmXEab_DxBZe5e&q=85&s=2ee3372c2582049d33df845718438627" alt="Getting code snippet from prompt registry" width="3024" height="1712" data-path="images/prompt-use-sdk-code.png" />
  </Frame>
</Accordion>

<CodeGroup>
  ```python theme={"dark"}
  from truefoundry import client, render_prompt
  from openai import OpenAI

  # Fetch the prompt template
  prompt_template = client.prompt_versions.get_by_fqn(
      fqn="chat_prompt:truefoundry/default/my-second-prompt:1",
  ).data.manifest

  # Render the prompt with input variables
  variables = {}  # Add your variables here
  prompt = render_prompt(prompt_template, variables=variables)

  # Send request to OpenAI directly
  response = OpenAI().chat.completions.create(
      messages=prompt["messages"],
      model=prompt.get("model") or "gpt-4o-mini",
      **(prompt.get("parameters") or {}),
  )

  print(response.choices[0].message.content)
  ```
</CodeGroup>
