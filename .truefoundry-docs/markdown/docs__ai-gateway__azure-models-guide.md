> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Choosing an Azure Resource Type

> Understand Azure OpenAI vs Azure AI Foundry and choose the right integration for TrueFoundry's AI Gateway

TrueFoundry supports both **Azure OpenAI** and **Azure AI Foundry** resource types for integrating Azure-hosted models with the AI Gateway. This guide explains the differences and helps you pick the right one.

## Azure OpenAI

A specialized Azure resource type that provides access to **OpenAI models only** (GPT-5.2, GPT-4.1, o3, o4-mini, etc.) through Azure's infrastructure. It is a managed service focused on providing a secure, enterprise-grade wrapper around OpenAI's APIs.

* Access limited to OpenAI models (GPT-5 series, GPT-4.1 series, o-series reasoning models, embeddings, image generation, etc.)
* Uses Azure deployment names to identify models
* Supports pay-per-token and provisioned throughput pricing
* Best suited when you only need OpenAI models

## Azure AI Foundry

A broader Azure resource type that includes **everything Azure OpenAI offers**, plus access to models from other providers like Anthropic (Claude), Mistral, Meta (Llama), Cohere, and more. It is the default resource type for new projects in Microsoft's Foundry portal.

* Access to all models including OpenAI, Anthropic, Mistral, Meta, Cohere, etc.
* Additional platform capabilities like agent orchestration, evaluations, and multi-model routing
* Recommended by Microsoft for new deployments

<Info>
  Azure AI Foundry is a superset of Azure OpenAI. You're not choosing one *instead of* the other. Foundry includes all Azure OpenAI capabilities plus additional models and platform features.
</Info>

## Comparison

|                                            | Azure AI Foundry | Azure OpenAI |
| ------------------------------------------ | ---------------- | ------------ |
| OpenAI models (GPT-5.2, GPT-4.1, o3, etc.) | ✓                | ✓            |
| Anthropic models (Claude)                  | ✓                | ✗            |
| Mistral, Meta, Cohere, and others          | ✓                | ✗            |
| Azure OpenAI API compatibility             | ✓                | ✓            |
| Agent orchestration and evaluations        | ✓                | ✗            |
| Recommended for new setups                 | ✓                |              |

For a detailed comparison from Microsoft, see [Choose an Azure resource type for AI Foundry](https://learn.microsoft.com/en-us/azure/ai-foundry/concepts/resource-types).

## Endpoint Formats

Azure AI Foundry uses a unified endpoint format (referred to as the Microsoft Foundry format) for most models:

```
https://<resource-name>.services.ai.azure.com/models/chat/completions?api-version=<version>
```

Anthropic models on Azure AI Foundry use a separate endpoint path:

```
https://<resource-name>.services.ai.azure.com/anthropic/v1/messages
```

<Note>
  Azure AI Foundry also supports an Azure OpenAI-compatible endpoint format (`<resource-name>.cognitiveservices.azure.com/openai/...`), but TrueFoundry uses the Foundry format shown above. You may see the OpenAI-compatible URL in the Azure portal for OpenAI model deployments. This is expected, both formats work, but you should use the Foundry format endpoint when configuring models in TrueFoundry.
</Note>

When configuring models in TrueFoundry, you provide the endpoint URL from your Azure AI Foundry deployment.

## What About Azure Hub?

Azure Hub is a management layer on top of Azure AI Foundry or Azure OpenAI that provides additional capabilities for organizing projects, governance, and team collaboration. When creating a Hub, you select an underlying AI Foundry or Azure OpenAI resource.

Since mid 2025, Microsoft has been moving most Hub capabilities directly into the **Foundry** resource type, bringing agents, models, and tools together under a single resource. For TrueFoundry integration, the relevant resource is the underlying **Azure AI Foundry** (or Azure OpenAI) resource itself. The Hub layer doesn't affect how you configure models in the AI Gateway.

## Which Should I Use?

**Starting fresh**: Use **Azure AI Foundry**. It gives you the broadest model access with full backward compatibility with Azure OpenAI APIs. Set up your resource in the [Azure AI Foundry portal](https://ai.azure.com).

**Already using Azure OpenAI**: Your existing setup continues to work with TrueFoundry. If you want access to non-OpenAI models or newer Foundry capabilities, Microsoft provides an upgrade path that preserves your existing API endpoints, deployments, and security configuration. See [Upgrade Azure OpenAI to Foundry](https://learn.microsoft.com/en-us/azure/ai-foundry/how-to/upgrade-azure-openai).

**Only need OpenAI models**: If you only need OpenAI models and already have an Azure OpenAI resource, there's no need to migrate. Both resource types are fully supported in TrueFoundry.

## Further Reading

* [Azure OpenAI or Azure AI Foundry](https://learn.microsoft.com/en-us/answers/questions/5572779/azure-openai-or-azure-ai-foundry) - Detailed community discussion on when to use each
* [Choose an Azure resource type](https://learn.microsoft.com/en-us/azure/ai-foundry/concepts/resource-types) - Official comparison of all Azure AI resource types
* [Upgrade Azure OpenAI to Foundry](https://learn.microsoft.com/en-us/azure/ai-foundry/how-to/upgrade-azure-openai) - Migration guide preserving existing endpoints and config

## Next Steps

* [Set up Azure AI Foundry](/docs/ai-gateway/azure-ai-foundry) **(Recommended)**
* [Set up Azure OpenAI](/docs/ai-gateway/azure-openai)
* [Certificate-based authentication](/docs/ai-gateway/azure-entra-certificate-auth)
