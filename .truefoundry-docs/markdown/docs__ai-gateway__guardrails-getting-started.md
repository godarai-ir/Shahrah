> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Getting Started

> Learn how to set up and implement guardrails to ensure safe and compliant LLM interactions and MCP tool invocations

To get started with integrating and using Guardrails, follow the steps below:

<Steps titleSize="h3">
  <Step title="Create a Guardrails Group">
    A guardrails group serves as a container for multiple guardrail integrations. You can assign who can add/edit/remove guardrails in a group and who can access the guardrails in a group. There are two type of roles for collaborators:

    * **Manager:** Can create, edit, delete and use the guardrail integrations in the group.
    * **User:** Can use the guardrail integrations in the group.

    To create a guardrails group, navigate to the AI Gateway -> Guardrail. Give your group a name. In our case, we've named it `openai-guardrail`.

    <Frame caption="Create a Guardrails Group">
      <img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/guardrail-getting-started-1.jpeg?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=1ca56c1ed0e29fba3b3c188120f5f03c" width="1280" height="696" data-path="images/guardrail-getting-started-1.jpeg" />
    </Frame>

    <Frame caption="Give your group a name">
      <img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/guardrail-getting-started-2.jpeg?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=e9b96647fc54b0e2a3e9f579bf7bce12" width="1280" height="694" data-path="images/guardrail-getting-started-2.jpeg" />
    </Frame>

    <Tip>
      **A Guardrail Group is the minimum level of access control for guardrails.**

      A common usage pattern is to create an org-wide global guardrails group which contains the organization wide guardrails. The Platform team is the manager and the entire org are users of the guardrail.

      For team/product specific guardrails, its better to create individual guardrail groups per product / team.
    </Tip>
  </Step>

  <Step title="Add Guardrail Integration in Guardrails Group">
    You can add integrations to your guardrails for content moderation and safety. Let's create an OpenAI guardrail integration, which we will later use as an output guardrail in our configuration example. This will allow us to automatically moderate and filter responses from the LLM using OpenAI's moderation capabilities.

    1. Select the OpenAI Moderation from the list of integrations.
           <Frame caption="Select the OpenAI Moderation Integration">
             <img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/guardrail-getting-started-3.jpeg?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=84a81a6966acce25d2812070443bb6c7" alt="Guardrail integration selection interface showing OpenAI Moderation option" width="1280" height="694" data-path="images/guardrail-getting-started-3.jpeg" />
           </Frame>

    2. Fill in the OpenAI Moderation Guardrail Config form. For this tutorial, we've named it `tfy-openai-moderation`.
           <Frame caption="Fill in the OpenAI Moderation Form">
             <img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/guardrail-getting-started-4.jpeg?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=4f529a1b9c312f6c6657fcea4b022fb5" alt="OpenAI Moderation configuration form with fields for name, API key, and operation type" width="1280" height="1003" data-path="images/guardrail-getting-started-4.jpeg" />
           </Frame>

    3. Save the guardrail group.

    You are now ready to use this OpenAI guardrail integration in your guardrail configuration to moderate LLM inputs. This means the guardrail will be applied to the input prompts before they are sent to the LLM, helping to ensure that any unsafe or non-compliant content is detected and handled prior to model processing.

    <Note>
      To customize moderation sensitivity for specific categories such as `harassment`, `sexual`, or `hate`, enable the **Category Thresholds** option. You can then adjust the threshold values for each category according to your requirements.

      <img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/guardrail-getting-started-5.jpeg?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=c52f6af69dc046dac59737d1dc51f097" width="1280" height="997" data-path="images/guardrail-getting-started-5.jpeg" />
    </Note>
  </Step>

  <Step title="Test Guardrails in Playground">
    The Playground allows you to test guardrails on all four hooks before deploying to production.

    ### Testing LLM Input/Output Guardrails

    1. **Navigate to the Playground**\
       Navigate to the AI Gateway → Playground Tab.

    2. **Select Guardrails**\
       On the left side, you'll see options for **LLM Input Guardrails** and **LLM Output Guardrails**.

       * Click on either option depending on which hook you want to test.
       * Add the guardrail you want to apply by selecting from the list.

       As we have already created an OpenAI guardrail named `tfy-openai-moderation` in the previous steps, select this guardrail under **LLM Input Guardrails**.

           <Frame caption="Select Guardrails">
             <img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/guardrail-playground.jpeg?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=48b973f0a228b737e9f0d7999843b580" alt="AI Gateway playground interface with guardrail selection panel on the left side" width="1280" height="697" data-path="images/guardrail-playground.jpeg" />
           </Frame>

    3. **Test the Guardrail**

       * Enter a prompt that would typically be flagged as unsafe or offensive.
       * Send the request to the model.

       You should see that the prompt is blocked by the guardrail, demonstrating that your configuration is working as expected.

    ### Testing MCP Tool Guardrails

    Once you have configured guardrail rules in **AI Gateway → Controls → Guardrails**, you can test MCP Pre Tool and Post Tool guardrails from the Playground:

    1. **Configure Guardrail Rules First**\
       Navigate to **AI Gateway → Controls → Guardrails** and create rules that target your MCP servers/tools with `mcp_tool_pre_invoke_guardrails` or `mcp_tool_post_invoke_guardrails`.

    2. **Use the Playground with MCP Tools**\
       In the Playground, when you invoke MCP tools (via agents or tool calls), the configured guardrails will automatically execute on the Pre Tool and Post Tool hooks.

    3. **View Results in Traces**\
       After execution, navigate to **Monitor → Request Traces** to see detailed information about which guardrails were triggered on each MCP tool invocation, including:
       * Pre Tool guardrail validation results
       * Post Tool guardrail findings and mutations
       * Pass/fail status for each guardrail

    <Tip>
      Testing in the Playground is an easy way to validate that your guardrails are correctly configured before deploying to production. For MCP guardrails, use the Request Traces view to verify guardrails are triggering as expected on tool invocations.
    </Tip>
  </Step>

  <Step title="Trigger Guardrails in LLM requests from code">
    You can pass guardrails as headers at a per-request level by providing the `X-TFY-GUARDRAILS` header. You can copy the relevant code snippet from the Playground section.

    1. **Select the Guardrail to Apply**\
       On the left side, choose the **LLM Input Guardrails** and **LLM Output Guardrails** sections, and select the guardrails you want to apply. For this tutorial, we've selected the `tfy-openai-moderation` guardrail for both input and output.
           <Frame caption="Select the Guardrail to Apply">
             <img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/guardrail-selector-playground.jpeg?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=80f0da9ff25e285e4d29a652236ca458" alt="Guardrail selector in playground showing selected OpenAI moderation guardrail" width="1280" height="713" data-path="images/guardrail-selector-playground.jpeg" />
           </Frame>

    2. **Get the Code Snippet**\
       After configuring your request and selecting the desired guardrail(s), click the **Code** button at the top right of the Playground.\
       In the code snippet section, you will see ready-to-use examples for different SDKs and `curl`.\
       **Note:** The generated code will automatically include the necessary `X-TFY-GUARDRAILS` header with your selected guardrails.
           <Frame caption="Code Snippet Section">
             <img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/guardrail-code-section.jpeg?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=244546d2473402d25e39b512c7d14bcf" alt="Code snippet section showing language options for implementing guardrails" width="1280" height="689" data-path="images/guardrail-code-section.jpeg" />
           </Frame>

    3. **Copy and Run the Code**\
       Copy the generated `curl` command (or code for your preferred SDK) and run it.\
       The guardrails you selected in the Playground will be applied automatically, as reflected in the request headers.
           <Frame caption="Copy the code snippet">
             <img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/guardrail-code-curl.jpeg?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=9713efa3e3b5436a2b979e944f665179" alt="Curl command example with X-TFY-GUARDRAILS header for applying guardrails" width="1280" height="446" data-path="images/guardrail-code-curl.jpeg" />
           </Frame>

    ### X-TFY-GUARDRAILS Header Format

    The header accepts JSON with the following fields:

    ```json theme={"dark"}
    {
      "llm_input_guardrails": ["group/guardrail-name"],
      "llm_output_guardrails": ["group/guardrail-name"],
      "mcp_tool_pre_invoke_guardrails": ["group/guardrail-name"],
      "mcp_tool_post_invoke_guardrails": ["group/guardrail-name"]
    }
    ```

    <Tip>
      You can use this approach with any HTTP client or SDK by adding the `X-TFY-GUARDRAILS` header to your request. This allows you to dynamically apply guardrails per request without changing your global configuration.
    </Tip>

    <Note>
      For backward compatibility, `input_guardrails` maps to `llm_input_guardrails` and `output_guardrails` maps to `llm_output_guardrails`.
    </Note>
  </Step>

  <Step title="Configure application of guardrails at gateway layer">
    1. **Navigate to Guardrail Rules**\
       Go to **AI Gateway → Controls → Guardrails** to access the Guardrail Rules page.

    2. **Create a Guardrail Configuration**\
       Click on **Create/Edit Guardrail Config**. Fill in the required details for your guardrail configuration:
       * Define `when` conditions to target specific users, models, metadata, or MCP servers
       * Configure guardrails for any of the four hooks:
         * `llm_input_guardrails` - Before LLM request
         * `llm_output_guardrails` - After LLM response
         * `mcp_tool_pre_invoke_guardrails` - Before MCP tool invocation
         * `mcp_tool_post_invoke_guardrails` - After MCP tool returns

    3. **Save the Configuration**\
       After filling out the form, click **Save** to apply your guardrail configuration.

    <Frame caption="Configure application of guardrails at gateway layer">
      <img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/guardrail-via-config.jpeg?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=738b7d6703ddf4c4e8ef75e2512a8936" alt="Gateway-level guardrail configuration interface with YAML editor for rules" width="1280" height="689" data-path="images/guardrail-via-config.jpeg" />
    </Frame>

    For more details on how to configure guardrails, see the [Guardrails Configuration Guide](/docs/ai-gateway/guardrails-configuration).

    <Tip>
      Configuring guardrails at the gateway layer is recommended for organization-wide enforcement. This centralizes guardrail management and auditing, eliminating the need to set headers on every request. MCP tool guardrails are especially important for securing agentic workflows.
    </Tip>
  </Step>

  <Step title="Monitor Guardrail Execution">
    After configuring guardrails, you can verify they are being applied correctly by monitoring execution in real-time through the AI Gateway dashboard:

    1. Navigate to **AI Gateway → Monitor → Request Traces**
    2. View detailed traces showing which guardrails were triggered on each hook
    3. See findings, mutations, and execution timing for each guardrail
    4. Filter by guardrail status to quickly find blocked or flagged requests

    <Frame caption="Guardrail latency visible in Request Traces">
      <img src="https://mintcdn.com/truefoundry/jcJfesAzCw8ThHaY/images/guardrail-request-traces.png?fit=max&auto=format&n=jcJfesAzCw8ThHaY&q=85&s=0c36d4b75b93f70049344ff8278dae13" alt="Request Traces view showing guardrail execution latency, result type, scope, and applied entity for each span" width="1024" height="560" data-path="images/guardrail-request-traces.png" />
    </Frame>

    <Tip>
      Use the Request Traces view to debug guardrail behavior, identify false positives during audit mode rollout, and verify your configuration is working as expected.
    </Tip>
  </Step>

  <Step title="Detect Guardrail Violations Programmatically">
    When a guardrail blocks a request, the Gateway returns a `400` status code with `error.type` set to `guardrail_checks_failed`. You can use this to handle violations in your application.

    ```python theme={"dark"}
    from openai import OpenAI

    client = OpenAI(
        api_key="your_truefoundry_api_key",
        base_url="{GATEWAY_BASE_URL}"
    )

    try:
        response = client.chat.completions.create(
            model="openai-main/gpt-4o-mini",
            messages=[
                {"role": "user", "content": "How do I write a function to calculate factorial?"}
            ]
        )
        print("Response:", response.choices[0].message.content)
    except Exception as e:
        if hasattr(e, 'response') and e.response is not None:
            error_data = e.response.json()
            if error_data.get('error', {}).get('type') == 'guardrail_checks_failed':
                print("Guardrail violation detected!")
                checks = error_data.get('guardrail_checks', {})
                for hook in ['llm_input_guardrails', 'llm_output_guardrails',
                             'mcp_tool_pre_invoke_guardrails', 'mcp_tool_post_invoke_guardrails']:
                    if checks.get(hook):
                        print(f"Violation on: {hook}")
    ```

    <Note>
      The `guardrail_checks` object in the error response only contains the hooks that were evaluated for the request.
    </Note>
  </Step>
</Steps>
