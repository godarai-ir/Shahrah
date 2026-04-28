> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Create Runs and Log Data

> Step-by-step guide for create run, explaining configuration, best practices, and real-world usage on TrueFoundry.

A run is used to represent a single invocation of a job, a script or a ML experiment. You can create a run at the beginning of your script or notebook, log parameters, metrics, artifacts, models, tags and finally end the run. This provides an easy to keep track of all data related to job runs or ML experiments.

A quick code snippet to create a run and end it:

```python lines theme={"dark"}
from truefoundry.ml import get_client

client = get_client()
run = client.create_run(ml_repo="iris-demo", run_name="svm-model")
# Your code here.
run.end()
```

You can organize multiple runs under a single ml\_repo. For example, the run `svm-model` will be created under the ml\_repo `iris-demo`.

You can view these runs in the TrueFoundry dashboard.

<Frame caption="TrueFoundry Dashboard">
  <img src="https://mintcdn.com/truefoundry/s4Aj2_qGCrSP-zc8/images/9e148d23-7fa8afc-mlfoundry_dashboard.png?fit=max&auto=format&n=s4Aj2_qGCrSP-zc8&q=85&s=c4d9d3c9021f55aa43942f45bb2dd31d" width="3022" height="1474" data-path="images/9e148d23-7fa8afc-mlfoundry_dashboard.png" />
</Frame>

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

    You can view the tags from the dashboard and also create new tags.

    <img src="https://mintcdn.com/truefoundry/4MAaF__cLD4iud16/images/6e1f1d92-4251517-Adding_Tags.png?fit=max&auto=format&n=4MAaF__cLD4iud16&q=85&s=b0f1e0f0bd65b3d323670e079572c2aa" width="2890" height="1220" data-path="images/6e1f1d92-4251517-Adding_Tags.png" />
  </Accordion>

  <Accordion title="Log parameters">
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
      Parameters are immutable and you cannot change the value of param once logged. If you need to change the value of param, it basically means that you are changing your input configuration and it's best to create a new run for that.
    </Info>

    #### Viewing logged parameter in dashboard

    <img src="https://mintcdn.com/truefoundry/s4Aj2_qGCrSP-zc8/images/7f64f35d-be7bcf3-View_logged.png?fit=max&auto=format&n=s4Aj2_qGCrSP-zc8&q=85&s=06dd74445948dfa766a35ce254252385" width="3022" height="1140" data-path="images/7f64f35d-be7bcf3-View_logged.png" />

    #### Filtering runs bases on parameter value

    To filters runs, click on top right corner of the screen to apply the required filter.

    <img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/e68b6fe3-9c3da3d-param.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=f484fea6fdb0120cb65817b5762129c1" width="3022" height="782" data-path="images/e68b6fe3-9c3da3d-param.png" />

    #### Capturing command-line arguments

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

  <Accordion title="Log metrics">
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

    ### Step-wise metric logging

    You can capture step-wise metrics too using the `step` argument.

    ```python lines theme={"dark"}
    for global_step in range(1000):
        run.log_metrics(metric_dict={"accuracy": 0.7, "loss": 0.6}, step=global_step)
    ```

    The stepwise-metrics can be visualized as graphs in the dashboard.

    <Frame caption="Step-wise metrics">
      <img src="https://mintcdn.com/truefoundry/PSBc0bX31_cIC7pm/images/d7d1514e-7a516f2-filter3.png?fit=max&auto=format&n=PSBc0bX31_cIC7pm&q=85&s=957e2b9f32497f21786b69714386c1ef" width="3022" height="1610" data-path="images/d7d1514e-7a516f2-filter3.png" />
    </Frame>

    #### Should I use epoch or global step as a value for the `step` argument?

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

## Accessing Runs in TrueFoundry

To interact with runs in TrueFoundry, you can use the provided methods in the TrueFoundryClient class. Here are the different possibilities to access runs:

<AccordionGroup>
  <Accordion title="Get a Run by ID">
    To retrieve an existing run by its ID, use the[`get_run_by_id`](/docs/mlfoundry#function-get-run-by-id) method:

    ```python lines theme={"dark"}
    client = TrueFoundryClient();
    run = client.get_run_by_id("run_id_here");
    ```
  </Accordion>

  <Accordion title="Get a Run by Fully Qualified Name (FQN)">
    If you have the fully qualified name (FQN) of a run, which follows the pattern tenant\_name/ml\_repo/run\_name, you can use the `get_run_by_fqn` method:

    ```python Python lines theme={"dark"}
    client = TrueFoundryClient();
    run = client.get_run_by_fqn("tenant_name/ml_repo/run_name");
    ```
  </Accordion>

  <Accordion title="Get All Runs for a Project">
    To retrieve all the runs' names and IDs for a project, use the [`get_all_runs`](/docs/mlfoundry#function-get-all-runs) method:

    ```python Python lines theme={"dark"}
    client = TrueFoundryClient();
    runs_df = client.get_all_runs((ml_repo = "project_name_here"));
    ```
  </Accordion>

  <Accordion title="Search Runs">
    You can search for runs that match specific criteria using the [`search_runs`](/docs/mlfoundry#function-search-runs) method:

    ```python Python lines theme={"dark"}
    client = TrueFoundryClient()
    runs = client.search_runs(
        ml_repo="project_name_here",
        filter_string="metrics.accuracy > 0.75",
        order_by=["metric.accuracy DESC"],
    )
    for run in runs:
        print(run)
    ```
  </Accordion>

  <Accordion title="Get Tags for a Run">
    You can use the `get_tags` method. It returns a dictionary.

    ```python lines theme={"dark"}
    from truefoundry.ml import get_client

    client = get_client()
    run = client.get_run("run-id-of-the-run")

    print(run.get_tags())
    ```
  </Accordion>

  <Accordion title="Get Parameters for a Run">
    You can use the [`get_params`](/docs/truefoundry-ml-reference/runs#function-get-params) method. It returns a dictionary

    ```python lines theme={"dark"}
    from truefoundry.ml import get_client

    client = get_client()
    run = client.get_run("run-id-of-the-run")

    print(run.get_params())
    ```
  </Accordion>

  <Accordion title="Get Metrics for a Run">
    You can use the `get_metrics`method. It returns a dictionary.

    ```python lines theme={"dark"}
    from truefoundry.ml import get_client

    client = get_client()
    run = client.get_run("run-id-of-the-run")

    metrics = run.get_metrics()

    for metric_name, metric_history in metrics.items():
        print(f"logged metrics for metric {metric_name}:")
        for metric in metric_history:
            print(f"value: {metric.value}")
            print(f"step: {metric.step}")
            print(f"timestamp_ms: {metric.timestamp}")
            print("--")

    run.end()
    ```
  </Accordion>
</AccordionGroup>

## FAQs

<AccordionGroup>
  <Accordion title="Can anyone create a run under my ml_repo?">
    You will need to have minimum of `Project Editor` role to create a run under a ml\_repo. `Project Viewer` role does not have permission to create a run.
  </Accordion>

  <Accordion title="Can I use runs as a context manager?">
    Yes, we can use runs as a context manager. A run will be automatically ended after the execution exits the `with` block.

    <CodeGroup>
      ```python Python lines theme={"dark"}
      client.create_ml_repo("iris-demo")

      run = client.create_run(ml_repo="iris-demo", run_name="svm-model")
      with run:
          # Your code here.
          ...

      # No need to call run.end()
      ```
    </CodeGroup>
  </Accordion>

  <Accordion title="Are run names unique?">
    Yes. run names under a ml\_repo are unique. If a run name already exists, we add a suffix to make it unique.\
    If you do not pass a run name while creating a run, we generate a random name.

    <CodeGroup>
      ```python Python lines theme={"dark"}
      from truefoundry.ml import get_client

      client = get_client()
      run = client.create_run(ml_repo="iris-demo")

      print(run.run_name)
      run.end()
      ```
    </CodeGroup>
  </Accordion>

  <Accordion title="How runs are identified?">
    Runs are identified by by their `id`.

    <CodeGroup>
      ```python Python lines theme={"dark"}
      from truefoundry.ml import get_client

      client = get_client()
      run = client.create_run(ml_repo="iris-demo")

      print(run.run_id)
      run.end()
      ```
    </CodeGroup>
  </Accordion>
</AccordionGroup>
