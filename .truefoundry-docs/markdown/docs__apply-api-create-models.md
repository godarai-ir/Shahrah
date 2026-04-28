> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Add and List Models

> Learn how to add and list models (OpenAI, Anthropic, AWS Bedrock, etc.) using TrueFoundry APIs with comprehensive examples

TrueFoundry allows you to connect and use models from providers like OpenAI, Anthropic, AWS Bedrock, Google Vertex AI, and many other providers from a single place. Once added, these models can be used in TrueFoundry AI Gateway and your applications.

The models are saved as provider integrations in TrueFoundry. The group of models are saved as a provider account. You can add multiple models to a single provider account; the models in the same provider account are called integrations.

<Note>
  Replace `YOUR_CONTROL_PLANE_URL` with your actual TrueFoundry control plane
  URL and `YOUR_API_KEY` with your API key throughout this guide.
</Note>

## Add Models

### API Overview

Here is the curl command example to add a model as a provider account:

```bash lines theme={"dark"}
curl -X PUT "YOUR_CONTROL_PLANE_URL/api/svc/v1/apply" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "manifest": {
      "type": "provider-account/openai",
      "name": "my-openai-account",
      "auth_data": {
        "type": "openai-api-key",
        "api_key": "sk-..."
      },
      "integrations": [
        {
          "type": "openai-model",
          "name": "gpt-4",
          "model_name": "gpt-4",
          "model_type": "chat"
        }
      ]
    },
    "dryRun": false
  }'
```

### Request Schema

The Apply API accepts a `TrueFoundryApplyRequest` object with the following structure:

<ParamField body="manifest" type="ModelProviderAccount | ModelManifest | ArtifactManifest | Service | ..." required>
  The manifest object defining the resource to be created or updated. For model
  provider accounts, this should be a `ModelProviderAccount` object.
</ParamField>

<ParamField body="dryRun" type="boolean" default="false">
  When set to `true`, performs a dry run without actually creating or updating
  the resource. Useful for validation.
</ParamField>

#### ModelProviderAccount Schema

For creating model provider accounts, the manifest should be a `ModelProviderAccount` object. This is a union type that can be one of any of the following provider account types:

**Supported Provider Account Types:**

* [🔗 AwsBedrockProviderAccount](/docs/truefoundry_sdk/types#awsbedrockprovideraccount) - AWS Bedrock models
* [🔗 GoogleVertexProviderAccount](/docs/truefoundry_sdk/types#googlevertexprovideraccount) - Google Vertex AI models
* [🔗 GoogleGeminiProviderAccount](/docs/truefoundry_sdk/types#googlegeminiprovideraccount) - Google Gemini models
* [🔗 AzureOpenAIProviderAccount](/docs/truefoundry_sdk/types#azureopenaiprovideraccount) - Azure OpenAI models
* [🔗 AnthropicProviderAccount](/docs/truefoundry_sdk/types#anthropicprovideraccount) - Anthropic Claude models
* [🔗 OpenaiProviderAccount](/docs/truefoundry_sdk/types#openaiprovideraccount) - OpenAI models
* [🔗 CohereProviderAccount](/docs/truefoundry_sdk/types#cohereprovideraccount) - Cohere models
* [🔗 GroqProviderAccount](/docs/truefoundry_sdk/types#groqprovideraccount) - Groq models
* [🔗 MistralAIProviderAccount](/docs/truefoundry_sdk/types#mistralaiprovideraccount) - Mistral AI models
* [🔗 TogetherAIProviderAccount](/docs/truefoundry_sdk/types#togetheraiprovideraccount) - Together AI models
* [🔗 OllamaProviderAccount](/docs/truefoundry_sdk/types#ollamaprovideraccount) - Ollama models
* [🔗 SelfHostedModelProviderAccount](/docs/truefoundry_sdk/types#selfhostedmodelprovideraccount) - Self-hosted models
* For more details, see the [ModelProviderAccount](/docs/truefoundry_sdk/types#modelprovideraccount) type.

**Common Fields for All Provider Accounts:**

<ParamField body="type" type="string" required>
  The provider account type (e.g., "provider-account/openai",
  "provider-account/anthropic", etc.)
</ParamField>

<ParamField body="name" type="string" required pattern="^[a-z][a-z0-9\\-]{1,30}[a-z0-9]$">
  Name of the provider account. Must be 3-32 characters, lowercase alphanumeric
  with hyphens, cannot start with a number.
</ParamField>

<ParamField body="auth_data" type="object" required>
  Authentication data specific to the provider (API keys, credentials, etc.)
</ParamField>

<ParamField body="integrations" type="array" required>
  List of model integrations available through this provider account
</ParamField>

<ParamField body="collaborators" type="array">
  Optional list of users who have access to this provider account
</ParamField>

### Response Schema

```json lines theme={"dark"}
{
  "id": "provider-account-123",
  "name": "my-openai-account",
  "fqn": "tenant-name/openai/my-openai-account",
  "provider": "openai",
  "manifest": {
    "type": "provider-account/openai",
    "name": "my-openai-account",
    "auth_data": {
      "type": "openai-api-key",
      "api_key": "sk-..."
    }
  },
  "integrations": [
    {
      "id": "integration-456",
      "name": "gpt-4",
      "fqn": "tenant-name/openai/my-openai-account/model/gpt-4",
      "type": "model",
      "manifest": {
        "type": "openai-model",
        "name": "gpt-4",
        "model_name": "gpt-4",
        "model_type": "chat"
      },
      "providerAccountFqn": "tenant-name/openai/my-openai-account",
      "createdBySubject": {
        "subjectSlug": "user@example.com"
      },
      "createdAt": "2024-01-15T10:30:00Z",
      "updatedAt": "2024-01-15T10:30:00Z"
    }
  ],
  "createdBySubject": {
    "subjectSlug": "user@example.com"
  },
  "createdAt": "2024-01-15T10:30:00Z",
  "updatedAt": "2024-01-15T10:30:00Z",
  "action": "CREATE"
}
```

### Response Fields

<ResponseField name="id" type="string">
  Unique identifier for the provider account.
</ResponseField>

<ResponseField name="name" type="string">
  Name of the provider account.
</ResponseField>

<ResponseField name="fqn" type="string">
  Fully qualified name of the provider account (tenant/provider/name).
</ResponseField>

<ResponseField name="provider" type="string">
  The provider type (e.g., "openai", "anthropic", "aws-bedrock").
</ResponseField>

<ResponseField name="manifest" type="object">
  The provider account manifest that was applied, excluding integrations.
</ResponseField>

<ResponseField name="integrations" type="array">
  List of provider integrations (models) associated with this account.
</ResponseField>

<ResponseField name="createdBySubject" type="object">
  Information about the user who created the provider account.
</ResponseField>

<ResponseField name="createdAt" type="string">
  Timestamp when the provider account was created.
</ResponseField>

<ResponseField name="updatedAt" type="string">
  Timestamp when the provider account was last updated.
</ResponseField>

<ResponseField name="action" type="string">
  The action performed: "CREATE" for new resources, "UPDATE" for existing ones.
</ResponseField>

### Examples

<Expandable title="OpenAI Provider Account">
  For connecting to OpenAI's API:

  ```json theme={"dark"}
  {
    "type": "provider-account/openai",
    "name": "my-openai-account",
    "auth_data": {
      "type": "openai-api-key",
      "api_key": "sk-your-openai-api-key"
    },
    "integrations": [
      {
        "type": "openai-model",
        "name": "gpt-4",
        "model_name": "gpt-4",
        "model_type": "chat"
      }
    ]
  }
  ```
</Expandable>

<Expandable title="Anthropic Provider Account">
  For connecting to Anthropic's Claude API:

  ```json theme={"dark"}
  {
    "type": "provider-account/anthropic",
    "name": "my-anthropic-account",
    "auth_data": {
      "type": "anthropic-api-key",
      "api_key": "sk-ant-your-anthropic-api-key"
    },
    "integrations": [
      {
        "type": "anthropic-model",
        "name": "claude-3-sonnet",
        "model_name": "claude-3-sonnet-20240229",
        "model_type": "chat"
      }
    ]
  }
  ```
</Expandable>

<Expandable title="AWS Bedrock Provider Account">
  For connecting to AWS Bedrock:

  ```json theme={"dark"}
  {
    "type": "provider-account/aws-bedrock",
    "name": "my-aws-bedrock-account",
    "region": "us-east-1",
    "auth_data": {
      "type": "aws-access-key",
      "access_key_id": "your-access-key",
      "secret_access_key": "your-secret-key"
    },
    "integrations": [
      {
        "type": "bedrock-model",
        "name": "claude-3-sonnet",
        "model_name": "anthropic.claude-3-sonnet-20240229-v1:0",
        "model_type": "chat"
      }
    ]
  }
  ```
</Expandable>

### Error Handling

The Apply API may return various error responses. Common scenarios include:

* **400 Bad Request**: Invalid manifest structure or missing required fields
* **401 Unauthorized**: Invalid or missing API key
* **403 Forbidden**: Insufficient permissions for the operation
* **404 Not Found**: Provider account not found
* **409 Conflict**: Provider account with the same name already exists
* **422 Unprocessable Entity**: Invalid provider credentials or configuration

## List Models

You can list and retrieve model integrations using the List Provider Integrations API. This allows you to query existing model integrations with filtering and pagination support.

### API Overview

Here is the curl command example to list model integrations:

```bash lines theme={"dark"}
curl -X GET "YOUR_CONTROL_PLANE_URL/api/svc/v1/provider-integrations?type=model" \
  -H "accept: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

### Query Parameters

<ParamField query="type" type="string" default="model">
  Filter integrations by type. For models, use `model`.
</ParamField>

<ParamField query="fqn" type="string">
  Filter by fully qualified name (FQN) of a specific integration. Cannot be used with `id`.
</ParamField>

<ParamField query="id" type="string">
  Filter by integration ID. Cannot be used with `fqn`.
</ParamField>

<ParamField query="offset" type="number" default="0">
  Pagination offset - number of items to skip.
</ParamField>

<ParamField query="limit" type="number">
  Maximum number of items to return. If not specified, returns all items.
</ParamField>

### Response Schema

```json lines theme={"dark"}
{
  "data": [
    {
      "id": "cmfpjat8x04yg01oe1jng7l5l",
      "name": "gpt-4",
      "fqn": "truefoundry:openai:my-openai-account:model:gpt-4",
      "type": "model",
      "manifest": {
        "name": "gpt-4",
        "type": "integration/model/openai",
        "model_id": "gpt-4",
        "model_types": ["chat"],
        "cost": {
          "metric": "public_cost"
        }
      },
      "metadata": {},
      "providerAccountFqn": "truefoundry:openai:my-openai-account",
      "providerAccount": {
        "id": "cmc93zjju01h001pfdr3l437d",
        "name": "my-openai-account",
        "fqn": "truefoundry:openai:my-openai-account",
        "provider": "openai",
        "manifest": {
          "name": "my-openai-account",
          "type": "provider-account/openai",
          "auth_data": {
            "type": "api-key",
            "api_key": "dummy-key"
          },
          "collaborators": [
            {
              "role_id": "provider-account-manager",
              "subject": "user:user@example.com"
            },
            {
              "role_id": "provider-account-access",
              "subject": "team:everyone"
            }
          ]
        },
        "createdBySubject": {
          "subjectId": "cm7t30kps08zz01s78jsgdtru",
          "subjectSlug": "user@example.com",
          "subjectType": "user",
          "subjectPatName": "user-pat-1",
          "subjectDisplayName": "John Doe"
        },
        "createdAt": "2024-01-15T10:30:00.000Z",
        "updatedAt": "2024-01-15T10:30:00.000Z",
        "createdBy": "user@example.com"
      },
      "createdBySubject": {
        "subjectId": "cm7t30kps08zz01s78jsgdtru",
        "subjectSlug": "user@example.com",
        "subjectType": "user",
        "subjectPatName": "user-pat-1",
        "subjectDisplayName": "John Doe"
      },
      "createdAt": "2024-01-15T10:30:00.000Z",
      "updatedAt": "2024-01-15T10:30:00.000Z",
      "createdBy": "user@example.com"
    }
  ],
  "pagination": {
    "total": 1,
    "offset": 0,
    "limit": 10
  }
}
```

### Response Fields

<ResponseField name="data" type="array">
  Array of model integration objects.

  <Expandable title="Integration Object Properties">
    <ResponseField name="id" type="string">
      Unique identifier for the model integration.
    </ResponseField>

    <ResponseField name="name" type="string">
      Name of the model integration.
    </ResponseField>

    <ResponseField name="fqn" type="string">
      Fully qualified name of the integration in the format: `tenant:provider:account:model:name` (colon-separated).
    </ResponseField>

    <ResponseField name="type" type="string">
      Type of the integration (e.g., "model").
    </ResponseField>

    <ResponseField name="manifest" type="object">
      The model integration manifest (auth\_data is excluded for security).
    </ResponseField>

    <ResponseField name="metadata" type="object">
      Additional metadata associated with the integration.
    </ResponseField>

    <ResponseField name="providerAccountFqn" type="string">
      FQN of the provider account this integration belongs to.
    </ResponseField>

    <ResponseField name="providerAccount" type="object">
      Full provider account object that contains this integration.

      <Expandable title="Provider Account Properties">
        <ResponseField name="id" type="string">
          Unique identifier for the provider account.
        </ResponseField>

        <ResponseField name="name" type="string">
          Name of the provider account.
        </ResponseField>

        <ResponseField name="fqn" type="string">
          Fully qualified name of the provider account.
        </ResponseField>

        <ResponseField name="provider" type="string">
          Provider type (e.g., "openai", "anthropic", "aws-bedrock").
        </ResponseField>

        <ResponseField name="manifest" type="object">
          Provider account manifest including auth\_data and collaborators.
        </ResponseField>

        <ResponseField name="createdBySubject" type="object">
          Information about the user who created the provider account.
        </ResponseField>

        <ResponseField name="createdAt" type="string">
          Timestamp when the provider account was created.
        </ResponseField>

        <ResponseField name="updatedAt" type="string">
          Timestamp when the provider account was last updated.
        </ResponseField>

        <ResponseField name="createdBy" type="string">
          ID of the user who created the provider account.
        </ResponseField>
      </Expandable>
    </ResponseField>

    <ResponseField name="createdBySubject" type="object">
      Information about the user who created the integration.

      <Expandable title="Subject Properties">
        <ResponseField name="subjectId" type="string">
          Unique identifier for the subject.
        </ResponseField>

        <ResponseField name="subjectType" type="string">
          Type of subject (e.g., "user").
        </ResponseField>

        <ResponseField name="subjectSlug" type="string">
          Slug identifier for the subject (e.g., email address).
        </ResponseField>

        <ResponseField name="subjectPatName" type="string | null">
          Personal Access Token name if the action was performed using a PAT, or null.
        </ResponseField>

        <ResponseField name="subjectDisplayName" type="string | null">
          Display name of the subject, or null if not available.
        </ResponseField>
      </Expandable>
    </ResponseField>

    <ResponseField name="createdAt" type="string">
      Timestamp when the integration was created.
    </ResponseField>

    <ResponseField name="updatedAt" type="string">
      Timestamp when the integration was last updated.
    </ResponseField>

    <ResponseField name="createdBy" type="string">
      ID of the user who created the integration.
    </ResponseField>
  </Expandable>
</ResponseField>

<ResponseField name="pagination" type="object">
  Pagination metadata.

  <Expandable title="Pagination Properties">
    <ResponseField name="total" type="number">
      Total number of integrations matching the query.
    </ResponseField>

    <ResponseField name="offset" type="number">
      Current pagination offset.
    </ResponseField>

    <ResponseField name="limit" type="number">
      Maximum number of items returned per page.
    </ResponseField>
  </Expandable>
</ResponseField>

### Examples

<Expandable title="List All Models">
  To list all model integrations in your tenant:

  ```bash theme={"dark"}
  curl -X GET "YOUR_CONTROL_PLANE_URL/api/svc/v1/provider-integrations?type=model" \
    -H "accept: application/json" \
    -H "Authorization: Bearer YOUR_API_KEY"
  ```
</Expandable>

<Expandable title="List Models with Pagination">
  To retrieve models with pagination (10 items starting from offset 20):

  ```bash theme={"dark"}
  curl -X GET "YOUR_CONTROL_PLANE_URL/api/svc/v1/provider-integrations?type=model&limit=10&offset=20" \
    -H "accept: application/json" \
    -H "Authorization: Bearer YOUR_API_KEY"
  ```
</Expandable>

<Expandable title="Get Specific Model by FQN">
  To retrieve a specific model by its fully qualified name:

  ```bash theme={"dark"}
  curl -X GET "YOUR_CONTROL_PLANE_URL/api/svc/v1/provider-integrations?type=model&fqn=truefoundry:openai:my-openai-account:model:gpt-4" \
    -H "accept: application/json" \
    -H "Authorization: Bearer YOUR_API_KEY"
  ```
</Expandable>

<Expandable title="Get Specific Model by ID">
  To retrieve a specific model by its integration ID:

  ```bash theme={"dark"}
  curl -X GET "YOUR_CONTROL_PLANE_URL/api/svc/v1/provider-integrations?type=model&id=integration-456" \
    -H "accept: application/json" \
    -H "Authorization: Bearer YOUR_API_KEY"
  ```
</Expandable>

### Error Handling

The List API may return various error responses. Common scenarios include:

* **400 Bad Request**: Invalid query parameters (e.g., both `fqn` and `id` provided)
* **401 Unauthorized**: Invalid or missing API key
* **404 Not Found**: Integration with specified `id` or `fqn` not found
