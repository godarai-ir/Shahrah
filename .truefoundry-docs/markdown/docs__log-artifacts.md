> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Artifacts and Artifact Versions

> Uploading files and directories as Artifacts. Downloading contents of Artifact to disk.

An Artifact on TrueFoundry is a bunch of files and folders stored on remote storage (cloud buckets like S3, GCS, Azure Blob) along with optional metadata fields. Additionally, artifacts are versioned automatically i.e. Each artifact once created has an associated version number and is immutable (contents cannot be changed). These artifacts can be retrieved later in any context and used in downstream applications.

## How artifacts are organized

In an organization, you can create multiple [ML Repo](/docs/ml-repo-quickstart#create-a-ml-repo). You can log one or more artifacts in each of these ML Repo. Under an ML Repo, an artifact is primarily identified by its name. Every time an artifact is created with the same under the same ML Repo, a new version is automatically created for that artifact.

The following diagram explains the hierarchy:

<Frame caption="How Artifact Versions are organized">
  <img src="https://mintcdn.com/truefoundry/DdP_2rhue4AQQlob/images/3cc715f7-ce31dc7-image.png?fit=max&auto=format&n=DdP_2rhue4AQQlob&q=85&s=93e7f08ec2112bcf1ddb23361279a29b" width="991" height="748" data-path="images/3cc715f7-ce31dc7-image.png" />
</Frame>

This hierarchy is also reflected in the Fully Qualified Name of any Artifact Version

<Frame caption="Artifact Version FQN parts">
  <img src="https://mintcdn.com/truefoundry/OHzlp6GY5G-JfKle/images/c4cc9643-5eef292-image.png?fit=max&auto=format&n=OHzlp6GY5G-JfKle&q=85&s=eb8b91cf1db38785b326597d39f28bcd" width="1310" height="214" data-path="images/c4cc9643-5eef292-image.png" />
</Frame>

<Info>
  ### Artifact vs Artifact Version

  Keep in mind that when you log an artifact with some name, you get back an artifact version under the same name. Therefore, when we refer to an artifact, we are almost always referring to a specific version of that artifact instead of the artifact itself (which represents all versions under it).
</Info>

### Relationship between Artifacts and Runs

In the above diagram, we can also see some dotted lines between artifacts and runs under the ML Repo. This is because a logged artifact version can have an associated source run denoting where it came from. Such a relationship is established when an artifact version is created using a [TrueFoundry Run](/docs/create-run) instead of direct logging.

Logging an artifact via a run can be useful, say, when we are performing an ML experiment and we have hyperparameters, metrics, images, and plots along with some files and folders. We can create a run to log hyperparameters, metrics, images, and plots and use the same run to log files and folders as an artifact. This way we can trace back useful metadata from the created artifact version.

On the other hand, you might have some dataset that might be used across multiple runs in the same ML Repo. In such a case, it might not make sense to associate logging of the dataset itself with some Run.

## Uploading an Artifact

You can log an artifact via two methods

1. Python SDK
2. User Interface

### Python SDK

To log an artifact you need

* `ml_repo` - Name of the ML Repo
* `name` - Name for the artifact
* `artifact_paths` - A list of tuples of `(<local file or directory path>, <remote file or directory path>)` indicating which file or folder to copy where relative to the root of the artifact.

Please see [`log_artifact`](/docs/mlfoundry#function-log-artifact) for complete reference.

<CodeGroup>
  ```python Direct Via Client lines theme={"dark"}
  import os
  from truefoundry.ml import get_client, ArtifactPath

  client = get_client()
  os.makedirs("my-folder", exist_ok=True)
  with open("my-folder/file-inside-folder.txt", "w") as f:
      f.write("Hello!")

  with open("just-a-file.txt", "w") as f:
      f.write("Hello from file!")

  artifact_version = client.log_artifact(
      ml_repo="my-ml-repo-1",
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
  ```

  ```python Via TrueFoundry Run lines theme={"dark"}
  import os
  from truefoundry.ml import get_client
  from truefoundry.ml import ArtifactPath

  client = get_client()
  client.create_ml_repo(ml_repo="my-ml-repo-1")

  os.makedirs("my-folder", exist_ok=True)
  with open("my-folder/file-inside-folder.txt", "w") as f:
      f.write("Hello!")

  with open("just-a-file.txt", "w") as f:
      f.write("Hello from file!")

  run = client.create_run(ml_repo="my-ml-repo-1", run_name="my-run-name")

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
</CodeGroup>

This is what it looks like on the dashboard. You can find the Artifact Version under ML Repos > Your ML Repo Name > Artifacts > Your Artifact Name

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/e1150c2e-17922e9-image.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=4739d8cb82ad8e06ba857773ca0e723d" width="2666" height="1276" data-path="images/e1150c2e-17922e9-image.png" />
</Frame>

### UI

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/qLfbCYWFClPrK0PEVtbiZ" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/qLfbCYWFClPrK0PEVtbiZ" style={{ border: "none", display: "flex", margin: "auto" }} />

1. On the platform's dashboard, click on the "ML Repos" tab in the left panel.
2. Select an ML Repo from the list or Create a new ML Repo and then select it.
3. Go to the "Other Artifacts" tab at the top.
4. Create a new Artifact by clicking on the "Create Artifact" Button.
5. Configure the artifact by providing the following details
   1. Artifact Name: Give a descriptive name to your Artifact for easy identification.
6. After filling in the details, click on the "Create" button to create the new Artifact.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/OHzlp6GY5G-JfKle/images/c195f7c8-6ad3ddc-Screenshot_2023-07-28_at_3.15.10_PM.png?fit=max&auto=format&n=OHzlp6GY5G-JfKle&q=85&s=0a79138cb3aa3566b1e8190889efdf5f" width="2880" height="1620" data-path="images/c195f7c8-6ad3ddc-Screenshot_2023-07-28_at_3.15.10_PM.png" />
</Frame>

7. The new Artifact should now be listed among the other Artifacts in your ML Repo. Click on the Artifact Name

8. The Artifact should be empty. On the Artifact page, locate and click the "New Artifact Version" button.

9. In the modal that appears,

   1. Upload the necessary files or models that you want to include as part of this Artifact version.
   2. Optionally, you can also add any relevant Artifact metadata to provide additional information about this version.

10. After filling in the details, click on the "Create" button

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/DdP_2rhue4AQQlob/images/4b753f93-7d9867d-Screenshot_2023-07-28_at_4.46.04_PM.png?fit=max&auto=format&n=DdP_2rhue4AQQlob&q=85&s=de790b90ef4689035aa8c653b43ce63c" width="2880" height="1624" data-path="images/4b753f93-7d9867d-Screenshot_2023-07-28_at_4.46.04_PM.png" />
</Frame>

Congratulations! You have successfully uploaded the new Artifact version, which includes the files and resources you added.

## Get the Artifact Version And Download Contents

You can use the [`get_artifact_version_by_fqn`](/docs/mlfoundry#function-get-artifact-version-by-fqn) method. It takes in `fqn` as an argument. FQN can be obtained from the UI dashboard or saved from the `fqn` property on [`ArtifactVersion`](/docs/artifact#kbdclasskbd-artifactversion) returned by [`log_artifact`](/docs/mlfoundry#function-log-artifact)

<Info>
  ### Access Permissions

  The user or service account accessing this artifact version should have at least `Viewer` permission on the ML Repo of the artifact version.
</Info>

<CodeGroup>
  ```python Python lines theme={"dark"}
  from truefoundry.ml import get_client

  client = get_client()
  artifact_version = client.get_artifact_version_by_fqn(fqn="<your artifact fqn>") # E.g. "artifact:truefoundry/user/my-classification-project/sklearn-artifact:1"

  # download the artifact contents to disk
  download_info = artifact_version.download(path="path/to/your/location")
  print(download_info)
  ```
</CodeGroup>

***
