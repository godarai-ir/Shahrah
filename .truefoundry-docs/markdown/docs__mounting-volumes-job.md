> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Mounting Volumes

> Mount persistent volumes to jobs for sharing data across training runs without re-downloading datasets.

When we run model training jobs, the job will need to download the data from somewhere in order to be able to train the model. While doing this once is fine, but in case if we are planning to run multiple training jobs with different hyperparameters or other settings on the same set of data, it will be wasteful from a time and cost perspective for every job to download the same set of data again and again.

This is where volumes can help. We can download the data once to a volume and then mount it to multiple jobs. They can then pickup the data from the volume as if it was downloaded on disk at the location where the volume is mounted.

## Mounting Volumes

To use a persistent volume, we will first need to create one and then attach it to our deployments. You can learn how to create volumes using the [Creating a Volume](/docs/creating-a-volume) guide

## Attaching Volumes to a Deployment

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/wibf0r91duo7vzNt9E10i " typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/wibf0r91duo7vzNt9E10i " style={{ border: "none", display: "flex", margin: "auto" }} />

## Using mounted volumes in your deployment

Once you've attached a volume to your deployment, you can use it like any other directory within your job's workflow. For instance, if you've downloaded a dataset, named `dataset.csv`, to a volume mounted at `/data`, you can directly access it from within the Job as shown in the code below:

<CodeGroup>
  ```python Python lines theme={"dark"}
  import os
  import pandas as pd
  ...

  # Load the dataset from the mounted volume
  data = pd.read_csv("/data/dataset.csv")

  # Train the model
  ...
  ```
</CodeGroup>

***
