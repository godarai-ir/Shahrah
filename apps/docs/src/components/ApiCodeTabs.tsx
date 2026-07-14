import CodeBlock from '@theme/CodeBlock';
import TabItem from '@theme/TabItem';
import Tabs from '@theme/Tabs';

type JsonValue =
  | string
  | number
  | boolean
  | null
  | JsonValue[]
  | { [key: string]: JsonValue };

type SdkOperation =
  | 'chat.completions.create'
  | 'responses.create'
  | 'embeddings.create'
  | 'moderations.create'
  | 'images.generate'
  | 'images.edit'
  | 'images.createVariation'
  | 'audio.speech.create'
  | 'audio.transcriptions.create'
  | 'audio.translations.create';

type CustomSnippets = Partial<
  Record<'python' | 'nodejs' | 'rest' | 'openaiPython' | 'openaiNodejs', string>
>;

type ApiCodeTabsProps = {
  apiKey?: string;
  baseUrl?: string;
  body?: JsonValue;
  endpoint: string;
  groupId?: string;
  headers?: Record<string, string>;
  method?: 'POST' | 'GET' | 'DELETE' | 'PATCH' | 'PUT';
  sdkOperation?: SdkOperation;
  snippets?: CustomSnippets;
};

const DEFAULT_BASE_URL = 'https://YOUR_WORKSPACE.godarai.ir';
const DEFAULT_API_KEY = 'YOUR_GODARAI_TOKEN';

function toJson(value: JsonValue | undefined) {
  return JSON.stringify(value ?? {}, null, 2);
}

function toPythonLiteral(value: JsonValue | undefined) {
  return toJson(value)
    .replace(/\btrue\b/g, 'True')
    .replace(/\bfalse\b/g, 'False')
    .replace(/\bnull\b/g, 'None');
}

function indent(value: string, spaces: number) {
  const prefix = ' '.repeat(spaces);
  return value
    .split('\n')
    .map((line) => `${prefix}${line}`)
    .join('\n');
}

function pythonHeaders(apiKey: string, headers?: Record<string, string>) {
  const values = {
    Authorization: `Bearer ${apiKey}`,
    'Content-Type': 'application/json',
    ...headers,
  };

  return Object.entries(values)
    .map(([key, value]) => `        ${JSON.stringify(key)}: ${JSON.stringify(value)},`)
    .join('\n');
}

function nodeHeaders(apiKey: string, headers?: Record<string, string>) {
  const values = {
    Authorization: `Bearer ${apiKey}`,
    'Content-Type': 'application/json',
    ...headers,
  };

  return Object.entries(values)
    .map(([key, value]) => `    ${JSON.stringify(key)}: ${JSON.stringify(value)},`)
    .join('\n');
}

function pythonExtraHeaders(headers?: Record<string, string>) {
  if (!headers || Object.keys(headers).length === 0) {
    return '';
  }

  return `    extra_headers={
${Object.entries(headers)
  .map(([key, value]) => `        ${JSON.stringify(key)}: ${JSON.stringify(value)},`)
  .join('\n')}
    }`;
}

function nodeRequestOptions(headers?: Record<string, string>) {
  if (!headers || Object.keys(headers).length === 0) {
    return '';
  }

  return `, {
  headers: {
${Object.entries(headers)
  .map(([key, value]) => `    ${JSON.stringify(key)}: ${JSON.stringify(value)},`)
  .join('\n')}
  },
}`;
}

function restHeaders(apiKey: string, headers?: Record<string, string>) {
  const values = {
    Authorization: `Bearer ${apiKey}`,
    'Content-Type': 'application/json',
    ...headers,
  };

  return Object.entries(values)
    .map(([key, value]) => `  --header ${JSON.stringify(`${key}: ${value}`)} \\`)
    .join('\n');
}

function toPythonKwargs(value: JsonValue | undefined) {
  const inner = toPythonLiteral(value).slice(1, -1).trim();
  if (!inner) {
    return '';
  }

  return inner
    .split('\n')
    .map((line) => line.replace(/^  "([A-Za-z_][A-Za-z0-9_]*)": /, '$1='))
    .join('\n');
}

function httpPythonSnippet(props: ApiCodeTabsProps) {
  const body = toPythonLiteral(props.body);
  const baseUrl = props.baseUrl ?? DEFAULT_BASE_URL;
  const apiKey = props.apiKey ?? DEFAULT_API_KEY;
  return `import requests

response = requests.${props.method?.toLowerCase() ?? 'post'}(
    "${baseUrl}${props.endpoint}",
    headers={
${pythonHeaders(apiKey, props.headers)}
    },
    json=${body},
    timeout=60,
)

response.raise_for_status()
print(response.json())`;
}

function nodeFetchSnippet(props: ApiCodeTabsProps) {
  const body = toJson(props.body);
  const baseUrl = props.baseUrl ?? DEFAULT_BASE_URL;
  const apiKey = props.apiKey ?? DEFAULT_API_KEY;
  return `const response = await fetch("${baseUrl}${props.endpoint}", {
  method: "${props.method ?? 'POST'}",
  headers: {
${nodeHeaders(apiKey, props.headers)}
  },
  body: JSON.stringify(${body}),
});

if (!response.ok) {
  throw new Error(\`Request failed with status \${response.status}\`);
}

const data = await response.json();
console.log(data);`;
}

function restSnippet(props: ApiCodeTabsProps) {
  const baseUrl = props.baseUrl ?? DEFAULT_BASE_URL;
  const apiKey = props.apiKey ?? DEFAULT_API_KEY;
  return `curl --request ${props.method ?? 'POST'} "${baseUrl}${props.endpoint}" \\
${restHeaders(apiKey, props.headers)}
  --data '${toJson(props.body)}'`;
}

function pythonSdkSnippet(
  operation: SdkOperation,
  body: JsonValue | undefined,
  apiKey: string,
  baseUrl: string,
  headers?: Record<string, string>,
) {
  return `from openai import OpenAI

client = OpenAI(
    api_key="${apiKey}",
    base_url="${baseUrl}/v1",
)

response = client.${operation}(
${indent(toPythonKwargs(body), 4)}${headers && Object.keys(headers).length > 0 ? ',' : ''}
${pythonExtraHeaders(headers)}
)

print(response)`;
}

function nodeSdkSnippet(
  operation: SdkOperation,
  body: JsonValue | undefined,
  apiKey: string,
  baseUrl: string,
  headers?: Record<string, string>,
) {
  return `import OpenAI from "openai";

const client = new OpenAI({
  apiKey: "${apiKey}",
  baseURL: "${baseUrl}/v1",
});

const response = await client.${operation}(${toJson(body)}${nodeRequestOptions(headers)});

console.log(response);`;
}

export default function ApiCodeTabs({
  apiKey = DEFAULT_API_KEY,
  baseUrl = DEFAULT_BASE_URL,
  body,
  endpoint,
  groupId = 'api-code-language',
  headers,
  method = 'POST',
  sdkOperation,
  snippets = {},
}: ApiCodeTabsProps) {
  const props = { apiKey, baseUrl, body, endpoint, groupId, headers, method, sdkOperation, snippets };
  const openaiPython =
    snippets.openaiPython ??
    (sdkOperation ? pythonSdkSnippet(sdkOperation, body, apiKey, baseUrl, headers) : undefined);
  const openaiNodejs =
    snippets.openaiNodejs ??
    (sdkOperation ? nodeSdkSnippet(sdkOperation, body, apiKey, baseUrl, headers) : undefined);

  return (
    <Tabs className="api-code-tabs" groupId={groupId} queryString="language">
      <TabItem value="python" label="Python">
        <CodeBlock language="python">
          {snippets.python ?? httpPythonSnippet(props)}
        </CodeBlock>
      </TabItem>
      <TabItem value="nodejs" label="NodeJS">
        <CodeBlock language="js">
          {snippets.nodejs ?? nodeFetchSnippet(props)}
        </CodeBlock>
      </TabItem>
      <TabItem value="rest" label="REST API">
        <CodeBlock language="bash">{snippets.rest ?? restSnippet(props)}</CodeBlock>
      </TabItem>
      {openaiPython ? (
        <TabItem value="openai-python" label="OpenAI Python SDK">
          <CodeBlock language="python">{openaiPython}</CodeBlock>
        </TabItem>
      ) : null}
      {openaiNodejs ? (
        <TabItem value="openai-nodejs" label="OpenAI NodeJS SDK">
          <CodeBlock language="js">{openaiNodejs}</CodeBlock>
        </TabItem>
      ) : null}
    </Tabs>
  );
}
