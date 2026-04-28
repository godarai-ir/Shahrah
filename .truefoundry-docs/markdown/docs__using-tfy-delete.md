> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Using tfy delete

> Delete TrueFoundry resources using tfy delete and a manifest file

## What is `tfy delete`?

`tfy delete` is a CLI command that lets you delete existing resources, such as applications and workspaces, by pointing to the same manifest file you used to create them with [`tfy apply`](/docs/using-tfy-apply).

Instead of navigating the UI or remembering resource identifiers, you can simply pass the manifest and `tfy delete` will identify and remove the corresponding resource from the platform.

> **Note:** If you haven't set up the CLI yet, see the [CLI setup guide](/docs/setup-cli).

## Basic Usage

To delete a resource described in a manifest file:

```bash lines theme={"dark"}
tfy delete -f <your-manifest-file>.yaml
```

You can also pass multiple files at once:

```bash lines theme={"dark"}
tfy delete -f manifest1.yaml -f manifest2.yaml
```

## Deleting an Application

If you previously deployed an application using a manifest like the one below:

```yaml lines theme={"dark"}
name: my-fastapi-app
type: application
image:
  type: image
  image_uri: my-registry/my-fastapi-app:latest
  command: uvicorn app:app --host 0.0.0.0 --port 8000
ports:
  - port: 8000
    protocol: TCP
replicas: 1
```

You can delete it by running:

```bash lines theme={"dark"}
tfy delete -f application-manifest.yaml
```

## Deleting a Workspace

Similarly, if you have a workspace manifest:

```yaml lines theme={"dark"}
name: my-dev-workspace
type: workspace
cluster: my-cluster
```

Delete it with:

```bash lines theme={"dark"}
tfy delete -f workspace-manifest.yaml
```

## Deleting Other Resources

`tfy delete` works with any resource type that supports manifest-based management. For example, a provider account manifest used with `tfy apply`:

```yaml lines theme={"dark"}
name: openai-account-cli
type: provider-account/openai
auth_data:
  api_key: tfy-secret://truefoundry:openai-group:API_KEY
  type: api-key
```

Can be deleted with:

```bash lines theme={"dark"}
tfy delete -f provider-account-manifest.yaml
```
