> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Migrate Sagemaker Pytorch Endpoint

> Migrate SageMaker PyTorch endpoints to TrueFoundry.

This guide describes how to deploy a Pytorch models deployed via Sagemaker endpoint in TrueFoundry.
For this we will need to adapt the existing inference script in Sagemaker to the TrueFoundry platform.

## Existing Code

A Sagemaker deployment typically contains code in the form of the following file tree -

```
code/
├── inference.py
└── requirements.txt
```

* **inference.py** - This is the inference handler that implements the Sagemaker functions like model\_fn, input\_fn, predict\_fn, output\_fn etc
* **requirements.txt** - This contains any additional Python packages needed by the inference handler

Apart from these, there are also:

* **Model artifacts** - Generated model files (e.g. `model.pth`). These may reside on your S3 buckets.
* **Sagemaker deployment code** (e.g. `sagemaker_deploy.py`) - Code to call Sagemaker to deploy the model as an endpoint

A sample example is shown below:

<CodeGroup>
  ```python inference.py lines theme={"dark"}
  # defining model and loading weights to it.
  import json
  import torch

  class Net(torch.nn.Module):
    ...

   
  def model_fn(model_dir):
      model = Net()
      with open(os.path.join(model_dir, "model.pth"), "rb") as f:
          model.load_state_dict(torch.load(f))
      model.to(device).eval()
      return model


  # data preprocessing
  def input_fn(request_body, request_content_type):
      assert request_content_type == "application/json"
      data = json.loads(request_body)["inputs"]
      data = torch.tensor(data, dtype=torch.float32, device=device)
      return data


  # inference
  def predict_fn(input_object, model):
      with torch.no_grad():
          prediction = model(input_object)
      return prediction


  # postprocess
  def output_fn(predictions, content_type):
      assert content_type == "application/json"
      res = predictions.cpu().numpy().tolist()
      return json.dumps(res)
  ```

  ```python sagemaker_deploy.py lines theme={"dark"}
  from sagemaker.pytorch import PyTorchModel
  from sagemaker.serializers import JSONSerializer
  from sagemaker.deserializers import JSONDeserializer


  pytorch_model = PyTorchModel(
      model_data = s3_model_uri,
      role=role,                  
      entry_point=<"inference.py">,
      framework_version = "1.12.1",
      py_version='py38',
      name = <"model_name">,
  )

  predictor = pytorch_model.deploy(
      serializer = JSONSerializer(),
      deserializer = JSONDeserializer(),
      endpoint_name = <"endpoint_name">,
      instance_type = "ml.g4dn.xlarge",
      initial_instance_count = 2,
  )
  ```
</CodeGroup>

## Deploying the model on TrueFoundry

Broadly speaking these are the things we shall do -

* Enclose the inference handler within a Docker container containing `torchserve` to support pytorch-based models
* Upload the model artifact as a TrueFoundry Artifact to make it accessible from the running container
* Launch a TrueFoundry deployment utilizing the above two pieces

<Steps>
  <Step title="Upload the Pytorch model artifacts to the TrueFoundry Model Registry">
    The existing model will look something like:

    ```
    model/
    └── model.pth
    ```

    Upload the model to the TrueFoundry Model registry either via code or UI.

    <Tabs>
      <Tab title="Upload via Code">
        ```python upload_model.py lines theme={"dark"}
        import mlfoundry

        client = mlfoundry.get_client()

        model_version = client.log_model(
            ml_repo=<"ml_repo_name">,
            name=<"my-pytorch-model">,
            model_file_or_folder=<"model/">,  # Path to directory containing the model
            framework=mlfoundry.ModelFramework.PYTORCH,
            description="TorchServe + Sagemaker compatible model artifact",
        )
        print("Model Version FQN:", model_version.fqn)
        ```
      </Tab>

      <Tab title="Upload Via UI">
        To upload the model via UI, navigate to the model registry and click on the `Upload Model` button. You can read the [guide](/docs/model-registry#upload-model-from-ui) to learn more.
      </Tab>
    </Tabs>
  </Step>

  <Step title="Create a Python script to launch the torchserve process at startup">
    ```python main.py lines theme={"dark"}
    import os
    # We are reusing sagemaker's open source toolkit to manage torchserve
    from sagemaker_pytorch_serving_container import serving
    import shutil

    def main():
        home_dir = os.getenv("HOME")
        # MODEL_DIR is populated by truefoundry artifacts. Model will be downloaded here
        artifact_model_dir = os.getenv("MODEL_DIR")
        if not artifact_model_dir:
            raise ValueError("`MODEL_DIR` must be set in environment and point to a model directory")
        # Base dir where models are supposed to be present
        base_dir = os.getenv("SAGEMAKER_BASE_DIR", os.path.join(home_dir, "model-store"))
        model_dir = os.path.join(base_dir, "model")
        # Copying over the model artifacts to the model dir
        shutil.copytree(src=artifact_model_dir, dst=model_dir)
        # Copying over the code to model dir
        shutil.copytree(src=os.path.join(home_dir, "code"), dst=os.path.join(model_dir, "code"))
        # Launching the torchserve process
        serving.main()


    if __name__ == '__main__':
        main()
    ```
  </Step>

  <Step>
    3. Next, we'll write a `Dockerfile` that can create the TrueFoundry application

    <CodeGroup>
      ```shell Dockerfile lines theme={"dark"}
      # We pick a gpu enabled torchserve image as base. Others can be found here - https://hub.docker.com/r/pytorch/torchserve/tags
      FROM --platform=linux/amd64 pytorch/torchserve:0.9.0-gpu
      ENV SAGEMAKER_BASE_DIR=/home/model-server/model-store
      RUN pip install -U pip setuptools wheel && \
          pip install --no-cache-dir \
              sagemaker-pytorch-inference==2.0.22 \
              scikit-learn \
              pandas \
              scipy==1.10.1
      USER root
      RUN touch /etc/sagemaker-ts.properties && \
          chown model-server:model-server /etc/sagemaker-ts.properties
      USER model-server
      WORKDIR /home/model-server/
      # Assuming code is the directory where inference code and requirements.txt is present
      COPY code ./code
      COPY main.py ./
      EXPOSE 8080
      ```
    </CodeGroup>
  </Step>

  <Step>
    4. Now let's go ahead and write a `deploy.py` script that can be used with TrueFoundry to get a service deployed. Here you'll need to change the following

       * Service Name - Name for the service we'll deploy
       * Entrypoint Script Name (value for `SAGEMAKER_PROGRAM`) - The code file name containing `model_fn`, `input_fn`, `predict_fn` and `output_fn`
       * Model Version FQN - The FQN obtained from `upload_model.py`

    <CodeGroup>
      ```python deploy.py lines theme={"dark"}
      import argparse
      import logging
      from typing import Optional

      from servicefoundry import (
          Service, Build, LocalSource,
          DockerFileBuild, Port, ArtifactsDownload,
          TrueFoundryArtifactSource,
          ArtifactsCacheVolume, Resources, GPUType, NvidiaGPU,
      )

      logging.basicConfig(level=logging.INFO, format=logging.BASIC_FORMAT)


      def main():
          service = Service(
              name=<"service_name">,
              image=Build(
                  build_source=LocalSource(local_build=False),
                  build_spec=DockerFileBuild(
                      dockerfile_path="./Dockerfile",
                      command="python main.py"
                  )
              ),
              ports=[Port(port=8080, host=<"host.app.example.com">, path=<"/">)],
              env={
                  # This should be the `entry_point` argument, the code file containing model_fn, predict_fn, etc
                  "SAGEMAKER_PROGRAM": <"inference.py">
              },
              artifacts_download=ArtifactsDownload(
                  artifacts=[              	
                      TrueFoundryArtifactSource(
                          # This should be the model version fqn obtained by running `upload_model.py`
                          artifact_version_fqn=<"model_version_fqn">
                          download_path_env_variable="MODEL_DIR",
                      )
                  ],
              ),
              resources=Resources(
                  cpu_request=1,
                  cpu_limit=4,
                  memory_request=8000,
                  memory_limit=16000,
                  ephemeral_storage_request=10000,
                  ephemeral_storage_limit=16000,
                  devices=[
                      NvidiaGPU(name=GPUType.T4, count=1)
                  ]
              ),
              liveness_probe=HealthProbe(
                  config=HttpProbe(path="/ping", port=8080),
                  initial_delay_seconds=30,
                  period_seconds=10,
                  timeout_seconds=1,
                  success_threshold=1,
                  failure_threshold=5,
              ),
              readiness_probe=HealthProbe(
                  config=HttpProbe(path="/ping", port=8080),
                  initial_delay_seconds=30,
                  period_seconds=10,
                  timeout_seconds=1,
                  success_threshold=1,
                  failure_threshold=5,
              ),
          )
          service.deploy(workspace_fqn=<"workspace_fqn">, wait=False)


      if __name__ == '__main__':
          main()
      ```
    </CodeGroup>
  </Step>

  <Step>
    Deploy using `truefoundry`

    <CodeGroup>
      ```shell Shell lines theme={"dark"}
      $ python deploy.py
      ```
    </CodeGroup>
  </Step>

  <Step>
    Once the deployment has gone through, it can be tested using this script -

    ```python test_endpoint.py lines theme={"dark"}
    if __name__ == "__main__":
        # The URL of the endpoint you're sending the request to
        url = 'https://<host.app.example.com>/predictions/model'

        # Your JSON payload
        data = {"inputs": <serialized_input>}

        # Send the POST request with the JSON payload
        response = requests.post(url, json=data)

        # Check the response
        if response.status_code == 200:
            print("Request successful.")
            # Process response data if needed
            print(response.json())
        else:
            print("Request failed.", response.status_code)
            print(response.text)
    ```
  </Step>
</Steps>
