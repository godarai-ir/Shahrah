> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Model Registry

> Store, version, and manage your ML models in the TrueFoundry model registry with stage transitions and metadata.

TrueFoundry provides a comprehensive model registry to store and manage models. Models are stored in [repositories](/docs/introduction-to-ml-repo.mdx) which are backed by S3/GCS/Azure Container/Minio storage on your cloud account. The key functionalities provided by the model registry are:

### Upload Model From UI

You can either upload a model file from disk or import a model saved in your S3 bucket.

<iframe href="https://app.supademo.com/embed/cmb9jz99m13fzsn1razlqttox?embed_v=2" typeofembed="iframe" height="475px" width="100%" provider="app.supademo.com" src="https://app.supademo.com/embed/cmb9jz99m13fzsn1razlqttox?embed_v=2" />

### Log Model In Code

<Note>
  Install the truefoundry library following the instructions [here](/docs/setup-cli.mdx)
</Note>

You can log models of all frameworks in the registry. Example of code to log SkLearn and Transformers are provided below:

<Tabs>
  <Tab title="Custom Framework">
    <CodeGroup>
      ```python Python lines theme={"dark"}
      from truefoundry.ml import get_client

      client = get_client()
      model_version = client.log_model(
          ml_repo="name-of-the-ml-repo",
          name="name-for-the-model",
          model_file_or_folder="path/to/model/file/or/folder/on/disk",
          framework=None
      )
      ```
    </CodeGroup>

    <Tip>
      Any file or folder can be saved as model by passing it in `model_file_or_folder` and `framework` can be set to `None`
    </Tip>
  </Tab>

  <Tab title="SkLearn">
    ```python lines theme={"dark"}
    from truefoundry.ml import get_client, SklearnFramework, infer_signature

    import joblib
    import numpy as np
    from sklearn.pipeline import make_pipeline
    from sklearn.preprocessing import StandardScaler
    from sklearn.svm import SVC

    X = np.array([[-1, -1], [-2, -1], [1, 1], [2, 1]])
    y = np.array([1, 1, 2, 2])
    clf = make_pipeline(StandardScaler(), SVC(gamma='auto'))
    clf.fit(X, y)
    joblib.dump(clf, "sklearn-pipeline.joblib")

    client = get_client()
    client.create_ml_repo(  # This is only required once
        name="my-classification-project",
        # This controls which bucket is used.
        # You can get this from Integrations > Blob Storage.
        storage_integration_fqn='<storage_integration_fqn>'
    )
    model_version = client.log_model(
        ml_repo="my-classification-project",
        name="my-sklearn-model",
        description="A simple sklearn pipeline",
        model_file_or_folder="sklearn-pipeline.joblib",
        framework=SklearnFramework(),
        metadata={"accuracy": 0.99, "f1": 0.80},
    )
    print(model_version.fqn)
    ```
  </Tab>

  <Tab title="Transformers">
    ```python Huggingface Transformers lines theme={"dark"}
    import os
    import shutil

    from truefoundry.ml import get_client, TransformersFramework
    from huggingface_hub import snapshot_download

    MODEL_ID = "HuggingFaceTB/SmolLM2-135M"   # Model Id from huggingface hub
    PIPELINE_TAG = "text-generation"          # Task the model is trained for
    LOCAL_PATH = "./transformers-model"       # Path for local download

    snapshot_download(
        MODEL_ID,
        revision=None,
        cache_dir=None,
        local_dir=LOCAL_PATH,
        ignore_patterns=["*.h5", "*.ot"],
        local_dir_use_symlinks=False,
    )

    if os.path.exists(os.path.join(LOCAL_PATH, '.cache')):
        shutil.rmtree(os.path.join(LOCAL_PATH, '.cache'))


    ML_REPO = "my-llm-project"         # ML Repo to upload to
    MODEL_NAME = "smollm2-135m"        # Model Name to upload as

    client = get_client()
    model_version = client.log_model(
        ml_repo=ML_REPO,
        name=MODEL_NAME,
        model_file_or_folder=LOCAL_PATH,
        framework=TransformersFramework(
            pipeline_tag=PIPELINE_TAG
        ),
    )
    print(model_version.fqn)
    ```
  </Tab>

  <Tab title="Pytorch">
    <CodeGroup>
      ```python Python lines theme={"dark"}
      from truefoundry.ml import get_client

      client = get_client()
      model_version = client.log_model(
          ml_repo="name-of-the-ml-repo",
          name="name-for-the-model",
          model_file_or_folder="path/to/model/file/or/folder/on/disk",
          framework=PyTorchFramework()
      )
      ```
    </CodeGroup>
  </Tab>

  <Tab title="TensorFlow">
    <CodeGroup>
      ```python Python lines theme={"dark"}
      from truefoundry.ml import get_client

      client = get_client()
      model_version = client.log_model(
          ml_repo="name-of-the-ml-repo",
          name="name-for-the-model",
          model_file_or_folder="path/to/model/file/or/folder/on/disk",
          framework=TensorFlowFramework()
      )
      ```
    </CodeGroup>
  </Tab>
</Tabs>

For other frameworks, you can use the following classes in `truefoundry.ml`

`FastAIFramework`, `GluonFramework`, `H2OFramework`, `KerasFramework`, `LightGBMFramework`, `ONNXFramework`, `PaddleFramework`, `SklearnFramework`, `SpaCyFramework`, `StatsModelsFramework`, `TransformersFramework`, `XGBoostFramework`

<Note>
  Any subsequent calls to `log_model` with the same `name` would create a new version of this model - `v2`, `v3` and so on.
</Note>

### View and manage versions

The logged model can be found in the Models tab. It can also be accessed from inside the MLRepo.

<Frame caption>
  <img src="https://mintcdn.com/truefoundry/OHzlp6GY5G-JfKle/images/cfca5ece-e43453146ee3833c9ea8bedd721ac9b037a4b1143730d540744aab3d62ff75fe-image.png?fit=max&auto=format&n=OHzlp6GY5G-JfKle&q=85&s=ef550b5dff3216fa339aec2101332fe5" alt="" width="3598" height="826" data-path="images/cfca5ece-e43453146ee3833c9ea8bedd721ac9b037a4b1143730d540744aab3d62ff75fe-image.png" />
</Frame>

You can view the details of each model version from there on.

<Frame caption>
  <img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/e37fca1a-2322ed688a2684a7d557648d2091433190a60c53104740d70b3321492b8a37d1-image_1.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=ac9f0216d567466b1ad75033cdceca2d" alt="" width="3444" height="888" data-path="images/e37fca1a-2322ed688a2684a7d557648d2091433190a60c53104740d70b3321492b8a37d1-image_1.png" />
</Frame>

<Note>
  Once a model version is created, the model version files are immutable. Only fields like description, framework, metadata can be updated using CLI or UI.
</Note>

### Using the Model in your code

For every model version, you can find the code to download it in your inference service.

<Frame caption="Code Snippet to download model">
  <img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/docs/use-model-in-code-snippet.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=f02947e081cc310dc3ce6dde10c9c6f3" alt="" width="3024" height="1664" data-path="images/docs/use-model-in-code-snippet.png" />
</Frame>
