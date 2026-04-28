> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Secret Management API

> Create and manage secret groups using TrueFoundry REST API with examples for create, update, and search.

The TrueFoundry API allows you to create and manage secret groups, which are collections of key-value pairs stored securely. Secret groups can be created, searched, and updated through REST API endpoints.

<Note>
  Replace `YOUR_CONTROL_PLANE_URL` with your actual TrueFoundry control plane
  URL and use the `TFY_API_KEY` environment variable for authentication throughout this guide.
</Note>

## Prerequisites

Before using the Secret Management API, ensure you have:

1. **TrueFoundry API Server URL**: `<control-plane-url>/api/svc`
2. **API Key**: Set the `TFY_API_KEY` environment variable for authentication

## API Endpoints

### 1. Create or Update Secret Group

Creates a new secret group or updates an existing one.

**Endpoint:** `PUT /v1/secret-groups` [Refer](https://truefoundry.com/docs/api-reference/secret-groups/create-or-update-secret-group)

**Headers:**

```
Authorization: Bearer <TFY_API_KEY>
Content-Type: application/json
Accept: application/json
```

**Request Body:**

```json theme={"dark"}
{
  "manifest": {
    "name": "secret-group-name",
    "type": "secret-group",
    "integration_fqn": "internal:aws:aws-1:secret-store:internal-secret-store",
    "collaborators": [
      {
        "role_id": "secret-group-admin",
        "subject": "user:email@example.com"
      },
      {
        "role_id": "secret-group-editor",
        "subject": "team:team-name"
      }
    ]
  }
}
```

**Manifest Fields:**

* `name` (string, required): Name of the secret group. If the name is 5 characters or less, it's recommended to append `-tenant` suffix.
* `type` (string, required): Always set to `"secret-group"`
* `integration_fqn` (string, required): Integration fully qualified name. Default: `"internal:aws:aws-1:secret-store:internal-secret-store"`
* `collaborators` (array, required): List of collaborators with their roles
  * `role_id` (string): Role identifier (`"secret-group-admin"` or `"secret-group-editor"`)
  * `subject` (string): Subject in format `"user:email@example.com"` or `"team:team-name"`

**Response:**

```json theme={"dark"}
{
  "data": {
    "id": "secret-group-id",
    "name": "secret-group-name",
    ...
  }
}
```

**Example:**

```bash theme={"dark"}
curl -X PUT "${TRUEFOUNDRY_API_SERVER_URL}/v1/secret-groups" \
  -H "Authorization: Bearer ${TFY_API_KEY}" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{
    "manifest": {
      "name": "my-tenant",
      "type": "secret-group",
      "integration_fqn": "internal:aws:aws-1:secret-store:internal-secret-store",
      "collaborators": [
        {
          "role_id": "secret-group-admin",
          "subject": "user:admin@example.com"
        },
        {
          "role_id": "secret-group-editor",
          "subject": "team:tenant-secret-access"
        }
      ],
      "ownedBy": {
        "account": "root-account"
      }
    }
  }'
```

### 2. Add Secrets to Secret Group

Adds secrets (key-value pairs) to an existing secret group.

**Endpoint:** `PUT /v1/secret-groups/{secret_group_id}` [Refer](https://truefoundry.com/docs/api-reference/secret-groups/update-secret-group)

**Path Parameters:**

* `secret_group_id` (string, required): The ID of the secret group returned from the create operation

**Headers:**

```
Authorization: Bearer <TFY_API_KEY>
Content-Type: application/json
```

**Request Body:**

```json theme={"dark"}
{
  "secrets": [
    {
      "key": "SECRET_KEY_1",
      "value": "secret-value-1"
    },
    {
      "key": "SECRET_KEY_2",
      "value": "secret-value-2"
    }
  ]
}
```

**Secrets Array:**

* Each object in the `secrets` array contains:
  * `key` (string, required): The secret key/name
  * `value` (string, optional): The secret value

<Note>
  Even if updating the value of one secret in the secret group, you need to pass the keys for all other secrets. For example, to update only `SECRET_KEY_1` to `secret-value-1-updated`, the payload would be:
</Note>

```json theme={"dark"}
{
  "secrets": [
    {
      "key": "SECRET_KEY_1",
      "value": "secret-value-1-updated"
    },
    {
      "key": "SECRET_KEY_2"
    }
  ]
}
```

If a key is missing in the payload, that secret will be deleted from the secret group.

**Example:**

```bash theme={"dark"}
curl -X PUT "${TRUEFOUNDRY_API_SERVER_URL}/v1/secret-groups/{secret_group_id}" \
  -H "Authorization: Bearer ${TFY_API_KEY}" \
  -H "Content-Type: application/json" \
  -d '{
    "secrets": [
      {
        "key": "DATABASE_PASSWORD",
        "value": "my-secure-password"
      },
      {
        "key": "API_KEY",
        "value": "sk-1234567890"
      }
    ]
  }'
```

### 3. Search Secret Groups

Searches for existing secret groups by fqn.

**Endpoint:** `GET /v1/secret-groups?fqn={fqn}` [Refer](https://truefoundry.com/docs/api-reference/secret-groups/list-secret-groups)

**Query Parameters:**

* `fqn` (string, required): Use FQN to search for a specific secret group. Format `<tenant-name>:<secret-group-name>`

**Headers:**

```
Authorization: Bearer <TFY_API_KEY>
Accept: application/json
```

**Response:**

```json theme={"dark"}
{
  "data": [
    {
      "id": "secret-group-id",
      "name": "secret-group-name",
      "associatedSecrets": [
        {
          "name": "SECRET_KEY_1"
        },
        {
          "name": "SECRET_KEY_2"
        }
      ],
      ...
    }
  ]
}
```

**Example:**

```bash theme={"dark"}
curl -X GET "${TRUEFOUNDRY_API_SERVER_URL}/v1/secret-groups?fqn=my-tenant:my-secret-group" \
  -H "Authorization: Bearer ${TFY_API_KEY}" \
  -H "Accept: application/json"
```
