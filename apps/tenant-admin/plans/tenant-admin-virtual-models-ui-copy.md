# Tenant Admin Virtual Models UI Copy

## Navigation
- Virtual Models

## List
- Title: Virtual Models
- Description: Create stable model aliases with priority routing, retries, and fallback.
- Empty title: No virtual models yet
- Empty description: Create a virtual model to route traffic across provider targets.

## Form
- Create title: Create virtual model
- Edit title: Edit virtual model
- Alias label: Alias
- Strategy label: Routing strategy
- Strategy value: Priority (MVP)
- Availability label: Availability
- Availability help: Disabled virtual models cannot be used in gateway requests.

## Target Editor
- Section title: Target chain
- Section description: Traffic starts at target 1 and falls back in order.
- Provider label: Provider
- Model label: Provider model
- Retry attempts label: Retry attempts
- Retry delay label: Retry delay (ms)
- Timeout label: Timeout (sec)
- Retry status codes label: Retry status codes
- Fallback status codes label: Fallback status codes

## Preview
- Title: Effective routing preview
- Description: Requests use this order until a target succeeds.

## Validation Messages
- Alias is required.
- Alias must contain only letters, numbers, dash, or underscore.
- At least one target is required.
- Each target must include a provider.
- Each target must include a model.
- Target order values must be unique.
- Retry attempts must be between 0 and 5.
- Retry delay must be between 0 and 10000 ms.

## Confirmations
- Unsaved changes: You have unsaved changes. Continue and discard them?
- Disable: Disable this virtual model? Requests using this alias will fail.
