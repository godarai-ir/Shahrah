> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# TrueFoundry ML Client

> TrueFoundry ML client reference for tracking experiments, logging artifacts, and managing models programmatically.

## Installation

```bash lines theme={"dark"}
pip install truefoundry
```

## Quick Example

```python lines theme={"dark"}
from truefoundry.ml import get_client

# Initialize the client
client = get_client()

# Create a run
run = client.create_run(
    ml_repo="my-classification-project",
    run_name="svm-with-rbf-kernel"
)

# Log metrics and parameters
run.log_metrics({"accuracy": 0.95, "loss": 0.1})
run.log_params({"learning_rate": 0.01, "epochs": 100})

# End the run
run.end()
```

## Available Modules

<div className="space-y-0">
  <a href="/docs/truefoundry-ml-reference/truefoundry-ml" className="group flex items-center justify-between py-1 px-2 bg-white border-b border-gray-200 hover:bg-gray-50 transition-colors">
    <h3 className="font-medium text-gray-900 group-hover:text-blue-700 transition-colors">TrueFoundry ML Client</h3>

    <div className="flex items-center space-x-2">
      <span className="text-blue-500 group-hover:text-blue-700 transition-colors">→</span>
    </div>
  </a>

  <a href="/docs/truefoundry-ml-reference/runs" className="group flex items-center justify-between py-1 px-2 bg-white border-b border-gray-200 hover:bg-gray-50 transition-colors">
    <h3 className="font-medium text-gray-900 group-hover:text-blue-700 transition-colors">Run</h3>

    <div className="flex items-center space-x-2">
      <span className="text-blue-500 group-hover:text-blue-700 transition-colors">→</span>
    </div>
  </a>
</div>
