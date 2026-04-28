> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Download Models/Artifacts

> Download and cache models efficiently to reduce cold starts and latency.

If you are deploying models as APIs, you will need the model file to be present along with the service so that you can load and serve it. There are a few ways people usually end up doing this:

**Bake the model file into the docker image**

This works, however, it has the following disadvantages:

1. The docker image size can become huge if the model file is big.
2. We will need to rebuild the docker image if the model file changes.
3. Every time the pod needs to come up on a new machine, it will end up downloading the model file again.

**Download the model file from S3 at the start of the server**

1. In this case, every time a pod comes up, we need to download the model first which leads to high startup time.
2. It also means more networking costs every time the models are downloaded.

**Create a volume, download the model file on the volume, and mount the volume to your service**

This approach has a lot of benefits:

1. Smaller docker image size leads to faster startup
2. Faster loading of models from volume leads the faster startup times
3. Cost savings because of not having to download the model repeatedly.

However, this can be a bit cumbersome to set up since the steps need to be done manually. Also, one more thing to be careful of here is the upgrade process from one model version to another. The key issues here are:

1. Upgrading requires manual intervention of having to download the new version of the model on the volume.
2. If we swap the older model on the volume with the newer model, and a pod restarts just after that, it will lead to one pod using the newer version of the model, even though we haven't started the deployment process.

To overcome the above issues, you can use the `Artifacts Download` feature in TrueFoundry which automatically provisions a volume, downloads the model onto the volume, mounts the volume and also handles upgrades safely. The way this works during the first deployment is:

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/s4Aj2_qGCrSP-zc8/images/7c849d9b-97b2833-image.png?fit=max&auto=format&n=s4Aj2_qGCrSP-zc8&q=85&s=2404547969b2deddfd641f0814fcd994" width="4466" height="817" data-path="images/7c849d9b-97b2833-image.png" />
</Frame>

As you can see in the diagram above, the user will need to define the link to the model to be downloaded and an environment variable in which TrueFoundry will provide the path where the model is present. The reason this is done this way is to provide a smooth and failsafe upgrade process. When you are deploying again, the following flow will be followed:

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/FrY4JbiyZud2He3p/images/18ea4912-e22e467-image.png?fit=max&auto=format&n=FrY4JbiyZud2He3p&q=85&s=16a7f991bc6bd0e5b4abed13802c4650" width="5266" height="1778" data-path="images/18ea4912-e22e467-image.png" />
</Frame>

In this scenario, the model is not downloaded if there is no change in the artifact version. However, if there is a change, TrueFoundry downloads the new model at a different path and then restarts all the servers with the new path in the environment variable. It also cleans up the older model once the rollout is fully complete. This makes it easy to switch back to the older version during a rollout if anything goes wrong.

### Using Artifact Download Feature to cache models

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/6X0FXduT3F11v3jh-Yz61" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/6X0FXduT3F11v3jh-Yz61" style={{ border: "none", display: "flex", margin: "auto" }} />

### Using the Cached Model

In your Service code, to use the cached model, you can utilize the `Download Path Environment Variable` you specified while configuring the `Download and Cache Artifact`.

You can do one of these two things, Get it directly in your code using `os.environ.get()` or Pass it as a command-line argument if your service needs it that way.

1. #### Read the download path from environment variable

For example, if you specify `MODEL_ID_1` as the`Download Path Environment Variable`, to download `model.joblib` file, you can access the cached model like this:

<CodeGroup>
  ```python Python lines theme={"dark"}
  import os
  import joblib

  # Get the model_path using the environment variable
  model_path_1 = os.environ.get("MODEL_ID_1")

  # Load the model using the specified path
  model_1 = joblib.load(os.path.join(model_path_1, "model.joblib"))
  ```
</CodeGroup>

2. #### Pass the download path as a command-line argument

For example, in case you are deploying an LLM on top of the vLLM model server using the `vllm/vllm-openai:v0.2.6` docker image. The following takes the model and tokenizer as a command-line argument.

You can pass the `Download Path Environment Variable` in the arguments.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/4MAaF__cLD4iud16/images/685789cb-f95033f-Screenshot_2024-01-26_at_12.08.47_AM.png?fit=max&auto=format&n=4MAaF__cLD4iud16&q=85&s=95a3bb14a5ec1ff9736a7547e65b98e7" width="2940" height="798" data-path="images/685789cb-f95033f-Screenshot_2024-01-26_at_12.08.47_AM.png" />
</Frame>
