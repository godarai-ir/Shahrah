> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Logging Data in Job Runs

> Track job runs and view logs for debugging and monitoring.

## Prerequisites

To log data from your job runs, you need access to an ML Repository (MLRepo). ML Repositories store models, data, artifacts, and prompts and are backed by blob storage like S3, GCS, or Azure Blob Storage.

<Accordion title="Setting up ML Repository Access">
  #### Step 1: Create or Access an ML Repository

  If you don't have an ML Repository yet, you'll need to create one:

  1. **Prerequisite - Blob Storage Integration**: Before creating a Repository, connect one or more Blob Storages to TrueFoundry:

     * [AWS S3](/docs/integration-provider-aws)
     * [Google Cloud Storage](/docs/integration-provider-gcp)
     * [Azure Blob Storage](/docs/integration-provider-azure)
     * Any S3 API Compatible Storage

  2. **Create Repository**: Go to Platform → Repositories tab and create a new ML Repository
</Accordion>

## Grant ML Repository Access to Workspace

To enable your job to log data, you need to grant access to the ML Repository for your workspace:

1. Go to Platform → Workspaces tab
2. Edit your workspace
3. In the "ML Repositories" section, grant access to your ML Repository
4. Choose appropriate permissions (Viewer or Editor)

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/cm3mwh0ej0mooahr7tqkjhe8c?embed_v=2" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/cm3mwh0ej0mooahr7tqkjhe8c?embed_v=2" style={{ border: "none", display: "flex", margin: "auto" }} />

## Creating Run and Logging Data

A run is used to represent a single ML experiment. You can create a run at the beginning of your script or notebook, log parameters, metrics, artifacts, models, tags and finally end the run. This provides an easy to keep track of all data related to ML experiments.

A quick code snippet to create a run and end it:

```python lines theme={"dark"}
from truefoundry.ml import get_client

client = get_client()
run = client.create_run(ml_repo="iris-demo", run_name="svm-model")
# Your code here.
run.end()
```

You can organize multiple runs under a single ml\_repo. For example, the run `svm-model` will be created under the ml\_repo `iris-demo`.

Once you've created runs and logged data, you can view them in the TrueFoundry dashboard. Navigate to your job in the Platform → Applications tab, click on the job name, and go to the "Job Runs" tab to see all executions with their status, metrics, and parameters.

<Frame caption="Job Runs Dashboard - Example showing multiple runs with different statuses">
  <img src="https://mintcdn.com/truefoundry/KZ9SZ6S9x8hzKvZy/images/run_on_job_page.png?fit=max&auto=format&n=KZ9SZ6S9x8hzKvZy&q=85&s=a0430a67fbc0c9c12ccd88ab4a97245a" alt="Job runs dashboard showing multiple runs with different statuses including finished, terminated, and failed runs" width="2652" height="1216" data-path="images/run_on_job_page.png" />
</Frame>

## Logging Different Types of Data

<AccordionGroup>
  <Accordion title="Creating a run and ending it">
    ```python Python lines theme={"dark"}
    from truefoundry.ml import get_client

    client = get_client()
    run = client.create_run(ml_repo="iris-demo", run_name="svm-model")
    # Your code here.
    run.end()
    ```
  </Accordion>

  <Accordion title="Adding tags to a run">
    ```python highlight={5} lines theme={"dark"}
    from truefoundry.ml import get_client

    client = get_client()
    run = client.create_run(ml_repo="iris-demo", run_name="svm-model")
    run.set_tags({"env": "development", "task": "classification"})
    # Your code here.
    run.end()
    ```

    You can view the tags from the dashboard and also create new tags.

    <img src="https://mintcdn.com/truefoundry/4MAaF__cLD4iud16/images/6e1f1d92-4251517-Adding_Tags.png?fit=max&auto=format&n=4MAaF__cLD4iud16&q=85&s=b0f1e0f0bd65b3d323670e079572c2aa" width="2890" height="1220" data-path="images/6e1f1d92-4251517-Adding_Tags.png" />
  </Accordion>

  <Accordion title="Logging parameters">
    Parameters are used to store the configuration of a run. This can be either the inputs to your script or the hyperparameters of your model during training like `learning_rate`, `cache_size`.
    The parameter values are stringified before storing.

    You can log parameters using the `log_params` as shown below:

    ```python highlight={6} lines theme={"dark"}
    from truefoundry.ml import get_client

    client = get_client()
    run = client.create_run(ml_repo="iris-demo", run_name="svm-model")

    run.log_params({"cache_size": 200.0, "kernel": "linear"})

    run.end()
    ```

    <Info>
      Parameters are immutable and you cannot change the value of param once logged.
      If you need to change the value of param, it basically means that you are
      changing your input configuration and it's best to create a new run for that.
    </Info>

    #### Viewing logged parameter in the dashboard

    <img src="https://mintcdn.com/truefoundry/s4Aj2_qGCrSP-zc8/images/7f64f35d-be7bcf3-View_logged.png?fit=max&auto=format&n=s4Aj2_qGCrSP-zc8&q=85&s=06dd74445948dfa766a35ce254252385" width="3022" height="1140" data-path="images/7f64f35d-be7bcf3-View_logged.png" />

    #### Filtering runs based on parameter value

    To filters runs, click on top right corner of the screen to apply the required filter.

    <img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/e68b6fe3-9c3da3d-param.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=f484fea6fdb0120cb65817b5762129c1" width="3022" height="782" data-path="images/e68b6fe3-9c3da3d-param.png" />

    #### Capturing command-line arguments in the run

    We can capture command-line arguments directly from the `argparse.Namespace` object.

    <CodeGroup>
      ```python Python lines theme={"dark"}
      import argparse
      from truefoundry.ml import get_client

      parser = argparse.ArgumentParser()
      parser.add_argument("--batch_size", type=int, required=True)
      args = parser.parse_args()

      client = get_client()
      run = client.create_run(ml_repo="iris-demo")

      run.log_params(args)

      run.end()
      ```
    </CodeGroup>
  </Accordion>

  <Accordion title="Logging metrics">
    Metrics are values that help you to evaluate and compare different runs - for e.g. `accuracy`, `f1 score`. You can log any output of your script as a metric.

    You can capture metrics using the [`log_metrics`](/docs/truefoundry-ml-reference/runs#function-log-metrics) method.

    ```python highlight={6} lines theme={"dark"}
    from truefoundry.ml import get_client

    client = get_client()
    run = client.create_run(ml_repo="iris-demo", run_name="svm-model")
    run.log_params({"cache_size": 200.0, "kernel": "linear"})
    run.log_metrics(metric_dict={"accuracy": 0.7, "loss": 0.6})

    run.end()
    ```

    These metrics can be seen in Truefoundry dashboard. Filters can be used on metrics values to filter out runs as shown in the figure.

    <Frame caption="Metrics Overview">
      <img src="https://mintcdn.com/truefoundry/DdP_2rhue4AQQlob/images/2dac81ef-7983f19-filter2.png?fit=max&auto=format&n=DdP_2rhue4AQQlob&q=85&s=26c4bdde91c9887db74d7c2f8ff654a8" width="3022" height="1730" data-path="images/2dac81ef-7983f19-filter2.png" />
    </Frame>

    <Frame caption="Filter runs on the basis of metrics">
      <img src="https://mintcdn.com/truefoundry/s4Aj2_qGCrSP-zc8/images/86e6cdb0-82b97e0-filter1.png?fit=max&auto=format&n=s4Aj2_qGCrSP-zc8&q=85&s=0a8e6515c441ee9373b598b00d190a61" width="3022" height="928" data-path="images/86e6cdb0-82b97e0-filter1.png" />
    </Frame>

    ### Step-wise metric logging in the run

    You can capture step-wise metrics too using the `step` argument.

    ```python lines theme={"dark"}
    for global_step in range(1000):
        run.log_metrics(metric_dict={"accuracy": 0.7, "loss": 0.6}, step=global_step)
    ```

    The stepwise-metrics can be visualized as graphs in the dashboard.

    <Frame caption="Step-wise metrics">
      <img src="https://mintcdn.com/truefoundry/PSBc0bX31_cIC7pm/images/d7d1514e-7a516f2-filter3.png?fit=max&auto=format&n=PSBc0bX31_cIC7pm&q=85&s=957e2b9f32497f21786b69714386c1ef" width="3022" height="1610" data-path="images/d7d1514e-7a516f2-filter3.png" />
    </Frame>

    #### Should I use epoch or global step as a value for the `step` argument in the run?

    If available you should use the global step as a value for the `step` argument. To capture epoch-level metric aggregates, you can use the following pattern.

    <CodeGroup>
      ```python Python lines theme={"dark"}
      run.log_metrics(
        (metric_dict = { "epoch/train_accuracy": 0.7, epoch: epoch }),
        (step = global_step)
      );
      ```
    </CodeGroup>
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

  <Accordion title="Log Images">
    You can also log images in different steps in a run. Images can be associated with a step number, in case you are running multiple epochs in training and want to log the images at different steps.

    `PIL` package is needed to log images. To install the PIL package, run

    ```
    pip install pillow
    ```

    Here is the sample code to log images from different sources:

    ```python lines theme={"dark"}
    from truefoundry.ml import get_client, Image
    import numpy as np
    import PIL.Image

    client = get_client()
    run = client.create_run(
        ml_repo="my-classification-project",
    )

    imarray = np.random.randint(low=0, high=256, size=(100, 100, 3))
    im = PIL.Image.fromarray(imarray.astype("uint8")).convert("RGB")
    im.save("result_image.jpeg")

    images_to_log = {
        "logged-image-array": Image(data_or_path=imarray),
        "logged-pil-image": Image(data_or_path=im),
        "logged-image-from-path": Image(data_or_path="result_image.jpeg"),
    }

    run.log_images(images_to_log, step=1)
    run.end()
    ```

    Images are represented and logged using this class in TrueFoundry.

    You can initialize `truefoundry.ml.Image` by either by using a local path or you can use a numpy array / PIL.Image object.

    You can also log caption and the actual and predicted values for an image as shown in the examples below.

    #### Logging images with caption and a class label

    ```python wrap lines theme={"dark"}
    from keras.datasets import mnist
    from truefoundry.ml import get_client, Image
    import time
    import numpy as np

    data = mnist.load_data()
    (X_train, y_train), (X_test, y_test) = data

    client = get_client()
    run = client.create_run("mnist-sample")

    actuals = list(y_test)
    predictions = list(np.random.randint(9, size=10))

    img_dict = {}
    for i in range(10):
        img_dict[str(i)] = Image(
            data_or_path=X_train[i],
            caption="mnist sample",
            class_groups={
                "actuals": str(actuals[i]),
                "predictions": str(predictions[i])
                },
        )

    run.log_images(img_dict)
    ```

    The logged images can be visualized in the TrueFoundry dashboard.

    <img src="https://mintcdn.com/truefoundry/OHzlp6GY5G-JfKle/images/c91261ee-a562885-MNIST_sample_images.png?fit=max&auto=format&n=OHzlp6GY5G-JfKle&q=85&s=ac61379b873cf9859c9e79c3f2cacd29" width="3004" height="1606" data-path="images/c91261ee-a562885-MNIST_sample_images.png" />

    You can also log images with multi-label classification problems.

    ```python lines theme={"dark"}
    images_to_log = {
        "logged-image-array": truefoundry.ml.Image(
            data_or_path=imarray,
            caption="testing image logging",
            class_groups={"actuals": ["dog", "human"], "predictions": ["cat", "human"]},
        ),
    }

    run.log_images(images_to_log, step=1)
    ```
  </Accordion>

  <Accordion title="Log Plots">
    You can also log plots in a run and visualize them in the TrueFoundry Dashboard. You can associate a plot with a step number, in case you are running multiple epochs in training and want to log the plots at different steps.

    You can log custom matplotlib, plotly plots as shown in examples below:

    <Tabs>
      <Tab title="Matplotlib Plot">
        ```python lines theme={"dark"}
        from truefoundry.ml import get_client
        from sklearn.metrics import ConfusionMatrixDisplay
        import matplotlib.pyplot as plt

        client = get_client()
        run = client.create_run(
            ml_repo="my-classification-project",
        )

        ConfusionMatrixDisplay.from_predictions(["spam", "ham"], ["ham", "ham"])

        run.log_plots({"confusion_matrix": plt}, step=1)
        run.end()
        ```
      </Tab>

      <Tab title="Seaborn Plot">
        ```python lines theme={"dark"}
        from truefoundry.ml import get_client
        from matplotlib import pyplot as plt
        import seaborn as sns

        # create a run in mlfoundry
        client = get_client()
        run = client.create_run(
            ml_repo="my-classification-project",
        )

        sns.set_theme(style="ticks", palette="pastel")

        # Load the example tips dataset
        tips = sns.load_dataset("tips")

        # Draw a nested boxplot to show bills by day and time
        sns.boxplot(x="day", y="total_bill", hue="smoker", palette=["m", "g"], data=tips)
        sns.despine(offset=10, trim=True)

        run.log_plots({"seaborn": plt})

        run.end()
        ```
      </Tab>

      <Tab title="Plotly Plot">
        ```python lines theme={"dark"}
        from truefoundry.ml import get_client
        import plotly.express as px

        client = get_client()
        run = client.create_run(ml_repo="my-classification-project")

        df = px.data.tips()
        fig = px.histogram(
            df,
            x="total_bill",
            y="tip",
            color="sex",
            marginal="rug",
            hover_data=df.columns,
        )

        plots_to_log = {
            "distribution-plot": fig,
        }

        run.log_plots(plots_to_log, step=1)
        run.end()
        ```
      </Tab>
    </Tabs>

    You can visualize the logged plots in the TrueFoundry Dashboard.

    <Columns cols={2}>
      <Card>
        <img src="https://mintcdn.com/truefoundry/PSBc0bX31_cIC7pm/images/da4209a5-b1b24bb-Plotly_Image.png?fit=max&auto=format&n=PSBc0bX31_cIC7pm&q=85&s=0d3161cd456d1ee880c926cb90bf2d7e" width="3004" height="1500" data-path="images/da4209a5-b1b24bb-Plotly_Image.png" />
      </Card>

      <Card>
        <img src="https://mintcdn.com/truefoundry/FrY4JbiyZud2He3p/images/25c6b61c-e5f529e-Visualizing_the_logged_plots.png?fit=max&auto=format&n=FrY4JbiyZud2He3p&q=85&s=5de12f60994a22d6a792637211b451ec" width="512" height="467" data-path="images/25c6b61c-e5f529e-Visualizing_the_logged_plots.png" />
      </Card>
    </Columns>
  </Accordion>
</AccordionGroup>

## Complete Examples

Here are comprehensive examples that demonstrate how to deploy a job and log data during machine learning training:

<AccordionGroup>
  <Accordion title="MNIST Training Script with Logging">
    This example shows the complete training script that logs parameters, metrics, plots, and models:

    ```python lines theme={"dark"}
    import argparse
    import matplotlib.pyplot as plt
    import tensorflow as tf
    from tensorflow.keras.datasets import mnist
    from truefoundry.ml import get_client

    # Parse command line arguments
    parser = argparse.ArgumentParser()
    parser.add_argument("--num_epochs", type=int, default=4)
    parser.add_argument("--learning_rate", type=float, default=0.01)
    parser.add_argument(
        "--ml_repo",
        type=str,
        required=True,
        help="The name of the ML Repo to track metrics and models"
    )
    args = parser.parse_args()

    # Load and preprocess the MNIST dataset
    (x_train, y_train), (x_test, y_test) = mnist.load_data()
    x_train = x_train / 255.0
    x_test = x_test / 255.0

    print(f"The number of train images: {len(x_train)}")
    print(f"The number of test images: {len(x_test)}")

    # Initialize TrueFoundry client and create a run
    client = get_client()
    run = client.create_run(ml_repo=args.ml_repo, run_name="mnist-training")

    try:
        # Log sample images as plots
        plt.figure(figsize=(10, 5))
        for i in range(10):
            plt.subplot(2, 5, i + 1)
            plt.imshow(x_train[i], cmap="gray")
            plt.title(f"Label: {y_train[i]}")
            plt.axis("off")
        plt.tight_layout()
        run.log_plots({"sample_images": plt})

        # Define and compile the model
        model = tf.keras.Sequential([
            tf.keras.layers.Flatten(input_shape=(28, 28)),
            tf.keras.layers.Dense(128, activation="relu"),
            tf.keras.layers.Dense(10, activation="softmax"),
        ])

        optimizer = tf.keras.optimizers.Adam(learning_rate=args.learning_rate)
        model.compile(
            optimizer=optimizer,
            loss="sparse_categorical_crossentropy",
            metrics=["accuracy"]
        )

        # Log hyperparameters
        run.log_params({
            "optimizer": "adam",
            "learning_rate": args.learning_rate,
            "num_epochs": args.num_epochs,
            "loss": "sparse_categorical_crossentropy",
            "metrics": ["accuracy"],
            "model_architecture": "Sequential with Dense layers"
        })

        # Train the model
        history = model.fit(
            x_train, y_train,
            epochs=args.num_epochs,
            validation_data=(x_test, y_test)
        )

        # Evaluate the model
        loss, accuracy = model.evaluate(x_test, y_test)
        print(f"Test loss: {loss}")
        print(f"Test accuracy: {accuracy}")

        # Log metrics for each epoch
        history_dict = history.history
        for epoch in range(args.num_epochs):
            run.log_metrics({
                "train_accuracy": history_dict['accuracy'][epoch],
                "val_accuracy": history_dict['val_accuracy'][epoch],
                "train_loss": history_dict['loss'][epoch],
                "val_loss": history_dict['val_loss'][epoch]
            }, step=epoch + 1)

        # Log final metrics
        run.log_metrics({
            "final_test_accuracy": accuracy,
            "final_test_loss": loss
        })

        # Save and log the trained model
        model.save("mnist_model.h5")
        run.log_model(
            name="mnist-handwritten-digits",
            model_file_or_folder="mnist_model.h5",
            framework="tensorflow",
            description="Neural network model to recognize handwritten digits from MNIST dataset",
            metadata={
                "test_accuracy": accuracy,
                "test_loss": loss,
                "num_epochs": args.num_epochs,
                "learning_rate": args.learning_rate
            }
        )

        # Add tags for organization
        run.set_tags({
            "dataset": "mnist",
            "model_type": "neural_network",
            "framework": "tensorflow",
            "task": "classification"
        })

        print("Training completed successfully!")

    except Exception as e:
        print(f"Training failed: {e}")
        raise
    finally:
        # End the run
        run.end()
    ```

    ### Key Features Demonstrated:

    * **Parameter Logging**: Hyperparameters like learning rate, epochs, and model configuration
    * **Metrics Logging**: Training and validation accuracy/loss for each epoch
    * **Plot Logging**: Sample images from the dataset
    * **Model Logging**: Saved model with metadata and framework information
    * **Tag Organization**: Categorizing the run for easy filtering
    * **Error Handling**: Proper exception handling to ensure runs are marked correctly
  </Accordion>

  <Accordion title="Deploy MNIST Training Job">
    This example shows how to deploy a parameterized job that can be run multiple times with different configurations:

    ```python lines theme={"dark"}
    import logging
    import argparse

    from truefoundry.deploy import Build, Job, LocalSource, Param, PythonBuild, Resources

    logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(name)s] %(levelname)-8s %(message)s")

    # Parse command line arguments
    parser = argparse.ArgumentParser()
    parser.add_argument("--workspace_fqn", type=str, required=True, help="FQN of the workspace to deploy to")
    args = parser.parse_args()

    # Define the job specifications
    job = Job(
        name="mnist-train-job",
        image=Build(
            build_source=LocalSource(local_build=False),
            build_spec=PythonBuild(
                python_version="3.11",
                command="python train.py --num_epochs {{num_epochs}} --ml_repo {{ml_repo}}",
                requirements_path="requirements.txt",
            ),
        ),
        params=[
            Param(name="num_epochs", default="4"),
            Param(name="ml_repo", param_type="ml_repo"),
        ],
        resources=Resources(
            cpu_request=0.5,
            cpu_limit=0.5,
            memory_request=1500,
            memory_limit=2000
        ),
    )

    # Deploy the job
    deployment = job.deploy(workspace_fqn=args.workspace_fqn, wait=False)
    print(f"Job deployed successfully! FQN: {deployment.fqn}")
    ```

    ### Key Features:

    * **Parameterized Job**: Uses `Param` objects to make the job configurable
    * **Python Build**: Automatically builds the container from source code
    * **Resource Configuration**: Specifies CPU and memory requirements
    * **ML Repository Integration**: Links to an ML repository for logging
  </Accordion>
</AccordionGroup>

## Accessing Detailed Run Information

In the Job Runs table, you'll notice that the **"RUN DETAILS"** column contains clickable links. When you click on any run details link, you'll be taken to a comprehensive view of that specific run, which includes:

* **Overview Tab**: Key metrics and hyperparameters used in the run
* **Results Tab**: Detailed metrics and performance data
* **Models Tab**: All logged models with their metadata
* **Artifacts Tab**: Files and artifacts associated with the run

<Info>
  **Pro Tip**: Use the run details view to analyze your experiments, compare
  different hyperparameter configurations, and track the progress of your
  machine learning projects over time.
</Info>
