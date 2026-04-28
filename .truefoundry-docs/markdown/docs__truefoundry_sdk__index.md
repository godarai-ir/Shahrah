> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# TrueFoundry Python SDK

> Complete Python SDK for TrueFoundry - Access and manage resources on TrueFoundry

## Installation

<Info>
  You need Python ≥ 3.9, ≤ 3.14 to use the Truefoundry Python SDK.
</Info>

```bash lines theme={"dark"}
pip install -U "truefoundry"
```

Please see the [Setup for CLI](/docs/setup-cli) for more details.

<Tip>
  Looking for TypeScript SDK? Check out the [TrueFoundry TypeScript SDK](https://www.npmjs.com/package/truefoundry-sdk)
</Tip>

## Quick Example

<CodeGroup>
  ```python Initialize With API Key lines theme={"dark"}
  from truefoundry import TrueFoundry

  # Initialize the client
  client = TrueFoundry(
      base_url="https://your-control-plane-url.truefoundry.com",
      api_key="your_api_key"
  )

  # List your applications
  applications = client.applications.list()
  for app in applications:
      print(f"Application: {app.name}")
  ```

  ```python Using Existing Login Session lines theme={"dark"}
  from truefoundry import client

  # List your applications
  applications = client.applications.list()
  for app in applications:
      print(f"Application: {app.name}")
  ```
</CodeGroup>

## Available Clients

<div className="space-y-0">
  <a href="/docs/truefoundry_sdk/main_client" className="group flex items-center justify-between py-1 px-2 bg-white border-b border-gray-200 hover:bg-gray-50 transition-colors">
    <h3 className="font-medium text-gray-900 group-hover:text-blue-700 transition-colors">TrueFoundry Client</h3>

    <div className="flex items-center space-x-2">
      <span className="text-xs text-gray-500">Main client</span>
      <span className="text-blue-500 group-hover:text-blue-700 transition-colors">→</span>
    </div>
  </a>

  <a href="/docs/truefoundry_sdk/agent_skill_versions" className="group flex items-center justify-between py-1 px-2 bg-white border-b border-gray-200 hover:bg-gray-50 transition-colors">
    <h3 className="font-medium text-gray-900 group-hover:text-blue-700 transition-colors">Agent Skill Versions</h3>

    <div className="flex items-center space-x-2">
      <span className="text-xs text-gray-500">3 methods</span>
      <span className="text-blue-500 group-hover:text-blue-700 transition-colors">→</span>
    </div>
  </a>

  <a href="/docs/truefoundry_sdk/agent_skills" className="group flex items-center justify-between py-1 px-2 bg-white border-b border-gray-200 hover:bg-gray-50 transition-colors">
    <h3 className="font-medium text-gray-900 group-hover:text-blue-700 transition-colors">Agent Skills</h3>

    <div className="flex items-center space-x-2">
      <span className="text-xs text-gray-500">4 methods</span>
      <span className="text-blue-500 group-hover:text-blue-700 transition-colors">→</span>
    </div>
  </a>

  <a href="/docs/truefoundry_sdk/alerts" className="group flex items-center justify-between py-1 px-2 bg-white border-b border-gray-200 hover:bg-gray-50 transition-colors">
    <h3 className="font-medium text-gray-900 group-hover:text-blue-700 transition-colors">Alerts</h3>

    <div className="flex items-center space-x-2">
      <span className="text-xs text-gray-500">1 methods</span>
      <span className="text-blue-500 group-hover:text-blue-700 transition-colors">→</span>
    </div>
  </a>

  <a href="/docs/truefoundry_sdk/application_versions" className="group flex items-center justify-between py-1 px-2 bg-white border-b border-gray-200 hover:bg-gray-50 transition-colors">
    <h3 className="font-medium text-gray-900 group-hover:text-blue-700 transition-colors">Application Versions</h3>

    <div className="flex items-center space-x-2">
      <span className="text-xs text-gray-500">2 methods</span>
      <span className="text-blue-500 group-hover:text-blue-700 transition-colors">→</span>
    </div>
  </a>

  <a href="/docs/truefoundry_sdk/applications" className="group flex items-center justify-between py-1 px-2 bg-white border-b border-gray-200 hover:bg-gray-50 transition-colors">
    <h3 className="font-medium text-gray-900 group-hover:text-blue-700 transition-colors">Applications</h3>

    <div className="flex items-center space-x-2">
      <span className="text-xs text-gray-500">9 methods</span>
      <span className="text-blue-500 group-hover:text-blue-700 transition-colors">→</span>
    </div>
  </a>

  <a href="/docs/truefoundry_sdk/artifact_versions" className="group flex items-center justify-between py-1 px-2 bg-white border-b border-gray-200 hover:bg-gray-50 transition-colors">
    <h3 className="font-medium text-gray-900 group-hover:text-blue-700 transition-colors">Artifact Versions</h3>

    <div className="flex items-center space-x-2">
      <span className="text-xs text-gray-500">10 methods</span>
      <span className="text-blue-500 group-hover:text-blue-700 transition-colors">→</span>
    </div>
  </a>

  <a href="/docs/truefoundry_sdk/artifacts" className="group flex items-center justify-between py-1 px-2 bg-white border-b border-gray-200 hover:bg-gray-50 transition-colors">
    <h3 className="font-medium text-gray-900 group-hover:text-blue-700 transition-colors">Artifacts</h3>

    <div className="flex items-center space-x-2">
      <span className="text-xs text-gray-500">5 methods</span>
      <span className="text-blue-500 group-hover:text-blue-700 transition-colors">→</span>
    </div>
  </a>

  <a href="/docs/truefoundry_sdk/clusters" className="group flex items-center justify-between py-1 px-2 bg-white border-b border-gray-200 hover:bg-gray-50 transition-colors">
    <h3 className="font-medium text-gray-900 group-hover:text-blue-700 transition-colors">Clusters</h3>

    <div className="flex items-center space-x-2">
      <span className="text-xs text-gray-500">6 methods</span>
      <span className="text-blue-500 group-hover:text-blue-700 transition-colors">→</span>
    </div>
  </a>

  <a href="/docs/truefoundry_sdk/data_directories" className="group flex items-center justify-between py-1 px-2 bg-white border-b border-gray-200 hover:bg-gray-50 transition-colors">
    <h3 className="font-medium text-gray-900 group-hover:text-blue-700 transition-colors">Data Directories</h3>

    <div className="flex items-center space-x-2">
      <span className="text-xs text-gray-500">9 methods</span>
      <span className="text-blue-500 group-hover:text-blue-700 transition-colors">→</span>
    </div>
  </a>

  <a href="/docs/truefoundry_sdk/environments" className="group flex items-center justify-between py-1 px-2 bg-white border-b border-gray-200 hover:bg-gray-50 transition-colors">
    <h3 className="font-medium text-gray-900 group-hover:text-blue-700 transition-colors">Environments</h3>

    <div className="flex items-center space-x-2">
      <span className="text-xs text-gray-500">4 methods</span>
      <span className="text-blue-500 group-hover:text-blue-700 transition-colors">→</span>
    </div>
  </a>

  <a href="/docs/truefoundry_sdk/events" className="group flex items-center justify-between py-1 px-2 bg-white border-b border-gray-200 hover:bg-gray-50 transition-colors">
    <h3 className="font-medium text-gray-900 group-hover:text-blue-700 transition-colors">Events</h3>

    <div className="flex items-center space-x-2">
      <span className="text-xs text-gray-500">1 methods</span>
      <span className="text-blue-500 group-hover:text-blue-700 transition-colors">→</span>
    </div>
  </a>

  <a href="/docs/truefoundry_sdk/jobs" className="group flex items-center justify-between py-1 px-2 bg-white border-b border-gray-200 hover:bg-gray-50 transition-colors">
    <h3 className="font-medium text-gray-900 group-hover:text-blue-700 transition-colors">Jobs</h3>

    <div className="flex items-center space-x-2">
      <span className="text-xs text-gray-500">5 methods</span>
      <span className="text-blue-500 group-hover:text-blue-700 transition-colors">→</span>
    </div>
  </a>

  <a href="/docs/truefoundry_sdk/logs" className="group flex items-center justify-between py-1 px-2 bg-white border-b border-gray-200 hover:bg-gray-50 transition-colors">
    <h3 className="font-medium text-gray-900 group-hover:text-blue-700 transition-colors">Logs</h3>

    <div className="flex items-center space-x-2">
      <span className="text-xs text-gray-500">1 methods</span>
      <span className="text-blue-500 group-hover:text-blue-700 transition-colors">→</span>
    </div>
  </a>

  <a href="/docs/truefoundry_sdk/ml_repos" className="group flex items-center justify-between py-1 px-2 bg-white border-b border-gray-200 hover:bg-gray-50 transition-colors">
    <h3 className="font-medium text-gray-900 group-hover:text-blue-700 transition-colors">Ml Repos</h3>

    <div className="flex items-center space-x-2">
      <span className="text-xs text-gray-500">4 methods</span>
      <span className="text-blue-500 group-hover:text-blue-700 transition-colors">→</span>
    </div>
  </a>

  <a href="/docs/truefoundry_sdk/model_versions" className="group flex items-center justify-between py-1 px-2 bg-white border-b border-gray-200 hover:bg-gray-50 transition-colors">
    <h3 className="font-medium text-gray-900 group-hover:text-blue-700 transition-colors">Model Versions</h3>

    <div className="flex items-center space-x-2">
      <span className="text-xs text-gray-500">5 methods</span>
      <span className="text-blue-500 group-hover:text-blue-700 transition-colors">→</span>
    </div>
  </a>

  <a href="/docs/truefoundry_sdk/models" className="group flex items-center justify-between py-1 px-2 bg-white border-b border-gray-200 hover:bg-gray-50 transition-colors">
    <h3 className="font-medium text-gray-900 group-hover:text-blue-700 transition-colors">Models</h3>

    <div className="flex items-center space-x-2">
      <span className="text-xs text-gray-500">5 methods</span>
      <span className="text-blue-500 group-hover:text-blue-700 transition-colors">→</span>
    </div>
  </a>

  <a href="/docs/truefoundry_sdk/personal_access_tokens" className="group flex items-center justify-between py-1 px-2 bg-white border-b border-gray-200 hover:bg-gray-50 transition-colors">
    <h3 className="font-medium text-gray-900 group-hover:text-blue-700 transition-colors">Personal Access Tokens</h3>

    <div className="flex items-center space-x-2">
      <span className="text-xs text-gray-500">5 methods</span>
      <span className="text-blue-500 group-hover:text-blue-700 transition-colors">→</span>
    </div>
  </a>

  <a href="/docs/truefoundry_sdk/prompt_versions" className="group flex items-center justify-between py-1 px-2 bg-white border-b border-gray-200 hover:bg-gray-50 transition-colors">
    <h3 className="font-medium text-gray-900 group-hover:text-blue-700 transition-colors">Prompt Versions</h3>

    <div className="flex items-center space-x-2">
      <span className="text-xs text-gray-500">5 methods</span>
      <span className="text-blue-500 group-hover:text-blue-700 transition-colors">→</span>
    </div>
  </a>

  <a href="/docs/truefoundry_sdk/prompts" className="group flex items-center justify-between py-1 px-2 bg-white border-b border-gray-200 hover:bg-gray-50 transition-colors">
    <h3 className="font-medium text-gray-900 group-hover:text-blue-700 transition-colors">Prompts</h3>

    <div className="flex items-center space-x-2">
      <span className="text-xs text-gray-500">5 methods</span>
      <span className="text-blue-500 group-hover:text-blue-700 transition-colors">→</span>
    </div>
  </a>

  <a href="/docs/truefoundry_sdk/secret_groups" className="group flex items-center justify-between py-1 px-2 bg-white border-b border-gray-200 hover:bg-gray-50 transition-colors">
    <h3 className="font-medium text-gray-900 group-hover:text-blue-700 transition-colors">Secret Groups</h3>

    <div className="flex items-center space-x-2">
      <span className="text-xs text-gray-500">7 methods</span>
      <span className="text-blue-500 group-hover:text-blue-700 transition-colors">→</span>
    </div>
  </a>

  <a href="/docs/truefoundry_sdk/secrets" className="group flex items-center justify-between py-1 px-2 bg-white border-b border-gray-200 hover:bg-gray-50 transition-colors">
    <h3 className="font-medium text-gray-900 group-hover:text-blue-700 transition-colors">Secrets</h3>

    <div className="flex items-center space-x-2">
      <span className="text-xs text-gray-500">3 methods</span>
      <span className="text-blue-500 group-hover:text-blue-700 transition-colors">→</span>
    </div>
  </a>

  <a href="/docs/truefoundry_sdk/teams" className="group flex items-center justify-between py-1 px-2 bg-white border-b border-gray-200 hover:bg-gray-50 transition-colors">
    <h3 className="font-medium text-gray-900 group-hover:text-blue-700 transition-colors">Teams</h3>

    <div className="flex items-center space-x-2">
      <span className="text-xs text-gray-500">5 methods</span>
      <span className="text-blue-500 group-hover:text-blue-700 transition-colors">→</span>
    </div>
  </a>

  <a href="/docs/truefoundry_sdk/traces" className="group flex items-center justify-between py-1 px-2 bg-white border-b border-gray-200 hover:bg-gray-50 transition-colors">
    <h3 className="font-medium text-gray-900 group-hover:text-blue-700 transition-colors">Traces</h3>

    <div className="flex items-center space-x-2">
      <span className="text-xs text-gray-500">1 methods</span>
      <span className="text-blue-500 group-hover:text-blue-700 transition-colors">→</span>
    </div>
  </a>

  <a href="/docs/truefoundry_sdk/users" className="group flex items-center justify-between py-1 px-2 bg-white border-b border-gray-200 hover:bg-gray-50 transition-colors">
    <h3 className="font-medium text-gray-900 group-hover:text-blue-700 transition-colors">Users</h3>

    <div className="flex items-center space-x-2">
      <span className="text-xs text-gray-500">12 methods</span>
      <span className="text-blue-500 group-hover:text-blue-700 transition-colors">→</span>
    </div>
  </a>

  <a href="/docs/truefoundry_sdk/virtual_accounts" className="group flex items-center justify-between py-1 px-2 bg-white border-b border-gray-200 hover:bg-gray-50 transition-colors">
    <h3 className="font-medium text-gray-900 group-hover:text-blue-700 transition-colors">Virtual Accounts</h3>

    <div className="flex items-center space-x-2">
      <span className="text-xs text-gray-500">8 methods</span>
      <span className="text-blue-500 group-hover:text-blue-700 transition-colors">→</span>
    </div>
  </a>

  <a href="/docs/truefoundry_sdk/workspaces" className="group flex items-center justify-between py-1 px-2 bg-white border-b border-gray-200 hover:bg-gray-50 transition-colors">
    <h3 className="font-medium text-gray-900 group-hover:text-blue-700 transition-colors">Workspaces</h3>

    <div className="flex items-center space-x-2">
      <span className="text-xs text-gray-500">6 methods</span>
      <span className="text-blue-500 group-hover:text-blue-700 transition-colors">→</span>
    </div>
  </a>
</div>

## Additional Resources

<div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-8">
  <a href="/docs/truefoundry_sdk/types" className="block p-4 bg-gray-50 border border-gray-200 rounded-lg hover:bg-gray-100 transition-colors">
    <h3 className="font-semibold text-gray-900 mb-2">📋 Data Types</h3>
    <p className="text-sm text-gray-600">Explore all input and output types used in the SDK</p>
  </a>

  <a href="/docs/truefoundry_sdk/enums" className="block p-4 bg-gray-50 border border-gray-200 rounded-lg hover:bg-gray-100 transition-colors">
    <h3 className="font-semibold text-gray-900 mb-2">🔢 Enums</h3>
    <p className="text-sm text-gray-600">View all available enum values and constants</p>
  </a>
</div>

## Getting Help

* **Documentation**: Browse through the client documentation above
* **GitHub**: [TrueFoundry Python SDK](https://github.com/truefoundry/truefoundry-python-sdk)
* **Support**: Reach out to our support team for assistance
