> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Introduction

> Overview of ML Repo for managing experiments and artifacts.

ML Repositories (referred to Repositories) in the portal are like respositories in Git, except they are for data. They store ML related data like runs, models, artifacts, etc.

Like in Git repositories, we can have Member, Editor and Admin roles, similarily ML Repositories also have these roles to define who can read and write data to the repository.

<img src="https://mintcdn.com/truefoundry/5CkapnZ7CyjQJ4bx/images/docs/ml-repositories-intro.png?fit=max&auto=format&n=5CkapnZ7CyjQJ4bx&q=85&s=28ea88f7ee2d1f979e05460f3816f6f9" width="3840" height="1936" data-path="images/docs/ml-repositories-intro.png" />

<Note>
  **Each repository is backed by a Blob Storage (S3, GCS, Azure Blob, MinIO, etc) to store the data in the repo**. Hence, to create a ML Repo, we need to connect one or more Blob Storages to Truefoundry.

  Two repositories can be linked to different paths within the same blob storage - however, one repository cannot be linked to multiple blob storages.
</Note>

The key data entities that a ML repository can store are:

* **Models** : Model comprises of model file and some metadata. Each Model can have multiple versions.
* **Artifacts** : Artifact is a collection of files. Each Artifact can have multiple versions.
* **Prompts** : Prompts are used to store the inputs to LLMs to use in applications / agents.
* **Runs** : A run represents a single experiment which in the context of Machine Learning is one specific model (say Logistic Regression), with a fixed set of hyper-parameters. Metrics, and parameters (details below) are all logged under a specific run. A run can also have links to models, artifacts, prompts that were either created in that run or used in that run.
  * **Parameters**: Parameters or HyperParameters that define your experiment and Machine Learning model. For example, `learning_rate, cache_size`.
  * **Metric**: Metrics are values that help you to evaluate and compare different runs. For example, `accuracy, f1 score`.
  * **Tags**: Tags are labels for a run. A tag is represented by a string tag name and value. For example, `env`.

<img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/e55ff63c-80d57f4-image_9.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=dc8c81f127679dace9ab0ed1410810a8" width="2043" height="714" data-path="images/e55ff63c-80d57f4-image_9.png" />
