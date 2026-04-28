> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Prompt Versioning

> Understand version history and the Prompt Registry; load specific versions into the Playground.

## Understanding Prompt Versions

TrueFoundry automatically tracks every change you make to a prompt template, enabling experimentation, history, and collaboration. Each time you modify a prompt, a new version is created with a unique version number, allowing you to:

* Track changes and version history
* Use any version and reverting applications to a previous version of prompt
* Collaborate with team members on prompt development

## Prompt Registry

The Prompt Registry is your central hub for managing all saved prompts. Access all your prompts through the Prompt Registry if you have read or write permissions.

<Frame>
  <img src="https://mintcdn.com/truefoundry/YDqmXEab_DxBZe5e/images/prompt-registry-listing.png?fit=max&auto=format&n=YDqmXEab_DxBZe5e&q=85&s=c2dee6f9d881b9a25df7f3d2d632fd23" alt="Prompt Registry showing all saved prompts" width="3016" height="1146" data-path="images/prompt-registry-listing.png" />
</Frame>

The registry displays:

* **Name**: The prompt identifier
* **Latest Version**: Current version number
* **Model Name**: Associated model (if configured)
* **Description**: Brief description of the prompt's purpose
* **Repository**: The workspace/repository location
* **Created By**: Author information and creation time

## Viewing Prompt Versions

Click any prompt in the registry to see detailed version history and configuration.

<Frame>
  <img src="https://mintcdn.com/truefoundry/YDqmXEab_DxBZe5e/images/prompt-registry-show-versions-try-playground.png?fit=max&auto=format&n=YDqmXEab_DxBZe5e&q=85&s=41f59126988ea964c2731c8cb8205e6f" alt="Prompt details with Try in Playground button" width="3024" height="1708" data-path="images/prompt-registry-show-versions-try-playground.png" />
</Frame>

You can also load any version directly into the Playground for testing and experimentation.

Click the **"Try in Playground"** button to:

* Load the selected prompt version into the Playground
* Test the prompt with different inputs
* Make modifications and create new versions
* Generate code snippets for implementation

### Compare Versions (Diff Mode)

Use the **Version Difference** tab to see exactly what changed between versions in standard github like diff view.

<Frame>
  <img src="https://mintcdn.com/truefoundry/YDqmXEab_DxBZe5e/images/prompt-diff-mode.png?fit=max&auto=format&n=YDqmXEab_DxBZe5e&q=85&s=f756246bfbb2e20e67b3b2e8b5065a9c" alt="Version difference showing changes between prompt versions" width="3024" height="1708" data-path="images/prompt-diff-mode.png" />
</Frame>
