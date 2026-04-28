> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# SQL Sanitizer Guardrail

> Detect and sanitize risky SQL patterns in LLM outputs using TrueFoundry's built-in SQL Sanitizer guardrail.

This guide explains how to use TrueFoundry's built-in **SQL Sanitizer** guardrail to detect and handle potentially dangerous SQL patterns in LLM interactions.

## What is SQL Sanitizer?

SQL Sanitizer is a built-in TrueFoundry guardrail that identifies and handles risky SQL patterns in text content. It can detect destructive SQL statements, unsafe query patterns, and potential SQL injection vectors. The guardrail runs directly within the AI Gateway without requiring external API calls.

### Key Features

1. **Destructive Statement Detection**: Identifies dangerous SQL operations including:
   * `DROP` statements that can delete tables or databases
   * `TRUNCATE` statements that remove all data from tables
   * `ALTER` statements that modify table structures
   * `GRANT` and `REVOKE` statements that change permissions

2. **Unsafe Pattern Detection**: Identifies risky query patterns:
   * `DELETE` statements without `WHERE` clauses
   * `UPDATE` statements without `WHERE` clauses
   * String interpolation patterns that may indicate SQL injection vulnerabilities

3. **Flexible Operation Modes**:
   * **Validate**: Detect risky patterns and block requests/responses
   * **Mutate**: Detect patterns, optionally strip comments, and continue with sanitized output

## Adding SQL Sanitizer Guardrail

To add SQL Sanitizer to your TrueFoundry setup, follow these steps:

<Steps>
  <Step title="Navigate to Guardrails">
    Go to the AI Gateway dashboard and navigate to the **Guardrails** section.
  </Step>

  <Step title="Create or Select a Guardrails Group">
    Create a new guardrails group or select an existing one where you want to add the SQL Sanitizer guardrail.
  </Step>

  <Step title="Add SQL Sanitizer Integration">
    Click on **Add Guardrail** and select **SQL Sanitizer** from the TrueFoundry Guardrails section.

    <Frame caption="Select SQL Sanitizer from TrueFoundry Guardrails">
      <img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/guardrail-1.jpeg?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=9ff04ad219001f1bfc31959b9ac261da" alt="TrueFoundry guardrail selection interface showing SQL Sanitizer option" width="1280" height="793" data-path="images/guardrail-1.jpeg" />
    </Frame>
  </Step>

  <Step title="Configure the Guardrail">
    Fill in the configuration form with your desired settings (see Configuration Options below).
  </Step>

  <Step title="Save the Configuration">
    Click **Save** to add the guardrail to your group.
  </Step>
</Steps>

## Configuration Options

| Parameter                                  | Default    | Description                                              |
| ------------------------------------------ | ---------- | -------------------------------------------------------- |
| **Operation**                              | `validate` | `validate` (block) or `mutate` (sanitize + continue)     |
| **Priority**                               | `1`        | Execution order for mutate guardrails (lower runs first) |
| **Enforcing Strategy**                     | `enforce`  | `enforce`, `enforce_but_ignore_on_error`, or `audit`     |
| **Block DROP/TRUNCATE/ALTER/GRANT/REVOKE** | `true`     | Block destructive statements                             |
| **Strip SQL Comments**                     | `true`     | Remove `--` and `/* */` comments                         |
| **Block DELETE/UPDATE without WHERE**      | `true`     | Block bulk operations                                    |
| **Require Parameterization**               | `false`    | Flag string interpolation patterns                       |

<Note>
  See [Guardrails Overview](/docs/ai-gateway/guardrails-overview#operation-modes) for Operation Modes and Enforcing Strategy details.
</Note>

### Validate vs Mutate

| Behavior                        | Validate      | Mutate           |
| ------------------------------- | ------------- | ---------------- |
| Blocked statements (DROP, etc.) | Block request | Log but continue |
| DELETE/UPDATE without WHERE     | Block request | Log but continue |
| SQL comments                    | Preserved     | **Stripped**     |
| Interpolation patterns          | Block request | Log but continue |

## Detected SQL Patterns

### Destructive Statements

| Statement        | Risk Level | Description                          |
| ---------------- | ---------- | ------------------------------------ |
| `DROP TABLE`     | Critical   | Permanently deletes tables           |
| `DROP DATABASE`  | Critical   | Permanently deletes entire databases |
| `TRUNCATE TABLE` | Critical   | Removes all rows from a table        |
| `ALTER TABLE`    | High       | Modifies table structure             |
| `GRANT`          | High       | Adds permissions                     |
| `REVOKE`         | High       | Removes permissions                  |

### Unsafe Query Patterns

| Pattern                        | Risk Level | Description               |
| ------------------------------ | ---------- | ------------------------- |
| `DELETE FROM table` (no WHERE) | High       | Deletes all rows in table |
| `UPDATE table SET` (no WHERE)  | High       | Updates all rows in table |

### SQL Injection Detection (Require Parameterization)

When **Require Parameterization** is enabled, the guardrail detects potential SQL injection vectors by looking for string interpolation patterns:

| Pattern Type                      | Detection      | Example                         |
| --------------------------------- | -------------- | ------------------------------- |
| **String Concatenation**          | `+` operator   | `"SELECT * FROM " + table_name` |
| **Python % Formatting**           | `%.` pattern   | `"SELECT * FROM %s" % table`    |
| **Template Literals / F-strings** | `{...}` braces | `f"SELECT * FROM {table}"`      |

<Warning>
  Interpolation detection is heuristic-based and may have false positives (e.g., JSON data with braces). Enable this option only when you specifically need to enforce parameterized queries.
</Warning>

### SQL Comment Patterns

The guardrail detects and can strip SQL comments:

| Comment Type      | Pattern     | Example                                |
| ----------------- | ----------- | -------------------------------------- |
| **Line Comment**  | `-- ...`    | `SELECT * FROM users; -- admin bypass` |
| **Block Comment** | `/* ... */` | `SELECT * /* hidden */ FROM users`     |

<Tip>
  Comments can be used to hide malicious SQL. Enable **Strip SQL Comments** to remove them, especially when processing user-provided or LLM-generated SQL.
</Tip>

## How It Works

1. Recursively scans all string content (including nested objects/arrays)
2. Strips comments (if enabled) before checking patterns
3. Checks for blocked statements and missing WHERE clauses
4. Returns verdict and optionally sanitized content (up to 10 issues per request)

```json theme={"dark"}
{
  "error": null,
  "verdict": false,
  "data": {
    "issues": [
      { "type": "blocked_statement", "message": "Blocked SQL statement detected: DROP", "field": "messages[0].content" },
      { "type": "delete_without_where", "message": "DELETE without WHERE clause detected", "field": "messages[0].content" }
    ],
    "sanitized": false,
    "explanation": "Detected 2 SQL issue(s): blocked_statement, delete_without_where"
  },
  "transformedData": {
    "request": { "json": null },
    "response": { "json": null }
  },
  "transformed": false
}
```

### Validate Mode

When configured in **validate** mode, the guardrail blocks requests containing risky patterns.

<AccordionGroup>
  <Accordion title="Blocked: DROP statement">
    ```sql theme={"dark"}
    DROP TABLE users;
    ```

    **Result**: Blocked - `blocked_statement: Blocked SQL statement detected: DROP`
  </Accordion>

  <Accordion title="Blocked: DELETE without WHERE">
    ```sql theme={"dark"}
    DELETE FROM orders;
    ```

    **Result**: Blocked - `delete_without_where: DELETE without WHERE clause detected`
  </Accordion>

  <Accordion title="Allowed: Safe SELECT query">
    ```sql theme={"dark"}
    SELECT * FROM users WHERE id = 1;
    ```

    **Result**: Allowed - No SQL issues detected
  </Accordion>

  <Accordion title="Allowed: DELETE with WHERE">
    ```sql theme={"dark"}
    DELETE FROM orders WHERE created_at < '2024-01-01';
    ```

    **Result**: Allowed - DELETE has WHERE clause
  </Accordion>
</AccordionGroup>

### Mutate Mode

When configured in **mutate** mode, the guardrail:

* **Strips SQL comments** from the output (line comments `--` and block comments `/* */`)
* **Logs all detected issues** but allows the request to continue
* **Returns sanitized content** with comments removed

<AccordionGroup>
  <Accordion title="Sanitized: Line comment removed">
    **Input**:

    ```sql theme={"dark"}
    SELECT * FROM users; -- bypass authentication
    ```

    **Output**:

    ```sql theme={"dark"}
    SELECT * FROM users;
    ```

    **Issues logged**: Comment stripped (no blocking issue)
  </Accordion>

  <Accordion title="Sanitized: Block comment removed">
    **Input**:

    ```sql theme={"dark"}
    SELECT * /* admin */ FROM users WHERE role = 'user'
    ```

    **Output**:

    ```sql theme={"dark"}
    SELECT *  FROM users WHERE role = 'user'
    ```
  </Accordion>

  <Accordion title="Flagged but allowed: DROP in mutate mode">
    **Input**:

    ```sql theme={"dark"}
    DROP TABLE temp_data;
    ```

    **Output**: Same as input (not modified)\
    **Issues logged**: `blocked_statement: Blocked SQL statement detected: DROP`\
    **Verdict**: Allowed (mutate mode doesn't block)
  </Accordion>
</AccordionGroup>

<Tip>
  **Use mutate mode carefully**: In mutate mode, even dangerous statements like DROP are logged but not blocked. Use this mode only when you have additional safeguards in place or need to sanitize comments while monitoring for issues.
</Tip>

## Use Cases

### Text-to-SQL Applications

When building applications that convert natural language to SQL, apply SQL Sanitizer on **MCP Pre Tool** to validate queries before they're executed by the database tool:

```yaml theme={"dark"}
name: guardrails-control
type: gateway-guardrails-config
rules:
  - id: text-to-sql-safety
    when:
      target:
        operator: or
        conditions:
          mcpServers:
            values:
              - database-tools
            condition: in
          mcpTools:
            values:
              - execute_query
            condition: in
      subjects:
        operator: and
        conditions:
          in:
            - team:data-engineering
          not_in:
            - user:db-admin@example.com
    llm_input_guardrails: []
    llm_output_guardrails: []
    mcp_tool_pre_invoke_guardrails:
      - my-guardrail-group/sql-sanitizer
    mcp_tool_post_invoke_guardrails: []
```

This ensures dangerous SQL (DROP, TRUNCATE, DELETE without WHERE) is blocked before reaching the database.

### Input Validation

Apply SQL Sanitizer to user inputs to detect potential SQL injection attempts:

```bash theme={"dark"}
curl -X POST "{GATEWAY_BASE_URL}/chat/completions" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -H 'X-TFY-GUARDRAILS: {"llm_input_guardrails":["my-guardrail-group/sql-sanitizer"],"llm_output_guardrails":[]}' \
  -d '{
    "model": "openai-main/gpt-4o-mini",
    "messages": [
      {"role": "user", "content": "Search for user: Robert'); DROP TABLE users;--"}
    ]
  }'
```

## Best Practices

<Note>
  SQL Sanitizer provides defense-in-depth but doesn't replace database permissions. Always configure database users with minimum required privileges.
</Note>

<Warning>
  **Pattern-based detection**: Case-insensitive regex matching. May have false positives in prose text. Complex obfuscation may bypass detection.
</Warning>

## Recommended Hooks

| Hook              | Use Case                                            |
| ----------------- | --------------------------------------------------- |
| **LLM Output**    | Validate SQL in LLM responses (text-to-SQL apps)    |
| **MCP Pre Tool**  | Validate SQL **before** execution by database tools |
| **MCP Post Tool** | Sanitize SQL in tool outputs                        |

```yaml theme={"dark"}
name: guardrails-control
type: gateway-guardrails-config
rules:
  - id: database-tool-safety
    when:
      target:
        operator: or
        conditions:
          mcpServers:
            values:
              - database-tools
            condition: in
      subjects:
        operator: and
        conditions:
          in:
            - team:engineering
          not_in:
            - user:db-admin@example.com
    llm_input_guardrails: []
    llm_output_guardrails: []
    mcp_tool_pre_invoke_guardrails:
      - my-guardrail-group/sql-sanitizer
    mcp_tool_post_invoke_guardrails: []
```

<Warning>
  Always apply SQL Sanitizer to **MCP Pre Tool** for database tools to prevent dangerous SQL execution.
</Warning>

## Customizing Rules

**Read-only apps**: Enable all blocking options.

**Write-enabled apps**: Keep DROP/TRUNCATE/ALTER blocking, optionally disable DELETE/UPDATE without WHERE checks.
