> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Getting Started

> Get started with ML Repo by creating a repository, setting up the CLI, and logging your first experiment.

### Create a ML Repo

<Info>
  ML Repositories are backed by a blob storage. So, you need to have atleast one blob storage integration to create an ML Repo. You can follow the guides to integrate [AWS S3](/docs/integration-provider-aws), [Google Cloud Storage](/docs/integration-provider-gcp), [Azure Blob Storage](/docs/integration-provider-azure) or any S3 compatible storage to TrueFoundry.
</Info>

You can then create an ML Repo from the ML Repo's tab in the Platform.

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/cm3o2yns311aoahr701yoz457?embed_v=2" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/cm3o2yns311aoahr701yoz457?embed_v=2" style={{ border: "none", display: "flex", margin: "auto" }} />

### Setup the TrueFoundry CLI

To get started, we need to have the truefoundry library installed. You can install it following the instructions in the [CLI Setup](/docs/setup-cli) docs.

### Add Log Lines to your code

You can use the code below to create a run, log metrics and parameters and then finally end the run.

```python Python lines theme={"dark"}
from truefoundry.ml import get_client

client = get_client()
# Create a new run
run = client.create_run(ml_repo="iris-demo", run_name="svm-model")
# Log parameters
run.log_params({"learning_rate": 0.001})
# Log metrics
run.log_metrics({"accuracy": 0.7, "loss": 0.6})
# End the run
run.end()
```

<AccordionGroup>
  <Accordion title="Create and end a run">
    ```python Python lines theme={"dark"}
    from truefoundry.ml import get_client

    client = get_client()
    run = client.create_run(ml_repo="iris-demo", run_name="svm-model")
    # Your code here.
    run.end()
    ```
  </Accordion>

  <Accordion title="Add tags to a run">
    ```python highlight={5} lines theme={"dark"}
    from truefoundry.ml import get_client

    client = get_client()
    run = client.create_run(ml_repo="iris-demo", run_name="svm-model")
    run.set_tags({"env": "development", "task": "classification"})
    # Your code here.
    run.end()
    ```
  </Accordion>

  <Accordion title="Log parameters">
    ```python highlight={6} lines theme={"dark"}
    from truefoundry.ml import get_client

    client = get_client()
    run = client.create_run(ml_repo="iris-demo", run_name="svm-model")

    run.log_params({"cache_size": 200.0, "kernel": "linear"})

    run.end()
    ```
  </Accordion>

  <Accordion title="Log metrics">
    ```python highlight={6} lines theme={"dark"}
    from truefoundry.ml import get_client

    client = get_client()
    run = client.create_run(ml_repo="iris-demo", run_name="svm-model")
    run.log_params({"cache_size": 200.0, "kernel": "linear"})
    run.log_metrics(metric_dict={"accuracy": 0.7, "loss": 0.6})

    run.end()
    ```
  </Accordion>

  <Accordion title="Log Artifacts">
    ```python highlight={18-31} lines theme={"dark"}
    import os
    from truefoundry.ml import get_client
    from truefoundry.ml import ArtifactPath

    client = get_client()
    run = client.create_run(ml_repo="iris-demo", run_name="svm-model")
    run.log_params({"cache_size": 200.0, "kernel": "linear"})
    run.log_metrics(metric_dict={"accuracy": 0.7, "loss": 0.6})

    # Just creating sample files to log as artifacts
    # os.makedirs("my-folder", exist_ok=True)
    # with open("my-folder/file-inside-folder.txt", "w") as f:
    #     f.write("Hello!")

    # with open("just-a-file.txt", "w") as f:
    #     f.write("Hello from file!")

    artifact_version = run.log_artifact(
        name="my-artifact",
        artifact_paths=[
            # Add files and folders here, `ArtifactPath` takes source and destination
            # source can be single file path or folder path
            # destination can be file path or folder path
            # Note: When source is a folder path, destination is always interpreted as folder path
            ArtifactPath(src="just-a-file.txt"),
            ArtifactPath(src="my-folder/", dest="cool-dir"),
            ArtifactPath(src="just-a-file.txt", dest="cool-dir/copied-file.txt")
        ],
        description="This is a sample artifact",
        metadata={"created_by": "my-username"}
    )
    print(artifact_version.fqn)
    run.end()
    ```
  </Accordion>

  <Accordion title="Log Models">
    ```python highlight={8-12} lines theme={"dark"}
    from truefoundry.ml import get_client

    client = get_client()
    run = client.create_run(ml_repo="iris-demo", run_name="svm-model")
    run.log_params({"cache_size": 200.0, "kernel": "linear"})
    run.log_metrics(metric_dict={"accuracy": 0.7, "loss": 0.6})

    model_version = run.log_model(
        name="name-for-the-model",
        model_file_or_folder="path/to/model/file/or/folder/on/disk",
        framework=<None or Framework> # Check 
    )
    run.end()
    ```
  </Accordion>
</AccordionGroup>

This run will now appear on the TrueFoundry dashboard under ML Repos Tab.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/JFTbQOWMkMfvFjDC/images/b05b872e-9cd8988-Screenshot_134.png?fit=max&auto=format&n=JFTbQOWMkMfvFjDC&q=85&s=f10362bf32df8e44669b04a9361211fc" width="1920" height="873" data-path="images/b05b872e-9cd8988-Screenshot_134.png" />
</Frame>

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/JFTbQOWMkMfvFjDC/images/b044a431-befcc7a-Screenshot_135.png?fit=max&auto=format&n=JFTbQOWMkMfvFjDC&q=85&s=00906677575cb9fcefffd4541f698141" width="1920" height="898" data-path="images/b044a431-befcc7a-Screenshot_135.png" />
</Frame>

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/JFTbQOWMkMfvFjDC/images/b6fd15e6-930ce79-Screenshot_136.png?fit=max&auto=format&n=JFTbQOWMkMfvFjDC&q=85&s=df1356ec5891e303fc05b28440c17956" width="1920" height="894" data-path="images/b6fd15e6-930ce79-Screenshot_136.png" />
</Frame>

Congratulations! You have successfully created a new ML repository. It is now ready for use, and you can start populating it with your data, code, models, and other related resources.
