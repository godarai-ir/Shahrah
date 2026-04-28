> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Using TPUs

> Deploy apps on Single Host TPU v4, v5e, v5p slices

<Warning>
  ### Compatibility Notes

  * Currently we only support Single Host TPU Topologies. If you have a use case that can benefit from Multi Host Topologies please reach out to us
  * Currently we only support provisioning/scheduling TPU workloads using Node Auto Provisioning (NAP). We'll add support for nodepools soon.
</Warning>

## Prerequistes

### Kubernetes Version

While different TPU types require different mininum GKE version, considering all TPU types, Node Auto Provisioning Support and ability to run without privileded mode, we recommend using **1.28.7-gke.1020000 or later** and **1.29.2-gke.1035000 or later**

Please refer to following links for up to date requirements

* [https://cloud.google.com/kubernetes-engine/docs/concepts/tpus#availability](https://cloud.google.com/kubernetes-engine/docs/concepts/tpus#availability)
* [https://cloud.google.com/kubernetes-engine/docs/concepts/node-auto-provisioning#tpus](https://cloud.google.com/kubernetes-engine/docs/concepts/node-auto-provisioning#tpus)
* [https://cloud.google.com/kubernetes-engine/docs/how-to/node-auto-provisioning#configuring\_tpus](https://cloud.google.com/kubernetes-engine/docs/how-to/node-auto-provisioning#configuring_tpus)
* [https://cloud.google.com/kubernetes-engine/docs/concepts/tpus#plan-tpu-configuration](https://cloud.google.com/kubernetes-engine/docs/concepts/tpus#plan-tpu-configuration)

### Regional Availability

TPUs are available in limited zones. Following table lists the availability for Single Host TPUs.

Please refer to following links for up to date availability:

* [https://cloud.google.com/kubernetes-engine/docs/concepts/tpus#availability](https://cloud.google.com/kubernetes-engine/docs/concepts/tpus#availability)

| TPU Type     | us-west1-c | us-west4-a | us-central1-a | us-central2-b | us-east1-c | us-east1-d | us-east5-a | us-east5-c | europe-west4-b |
| ------------ | ---------- | ---------- | ------------- | ------------- | ---------- | ---------- | ---------- | ---------- | -------------- |
| v4           |            |            |               | ✅             |            |            |            |            |                |
| v5e Device   |            |            | ✅             |               |            |            |            |            | ✅              |
| v5e PodSlice | ✅          | ✅          |               |               | ✅          |            |            |            |                |
| v5p          |            |            |               |               |            | ✅          | ✅          | ✅          |                |

### Quota

Please make sure you have enough quota available in the respective region. You can apply for Quotas on your GCP Console: [https://console.cloud.google.com/iam-admin/quotas](https://console.cloud.google.com/iam-admin/quotas)

### Node Auto Provisioning

Once you have Quota, add the TPU accelerators to NAP's limits. Please follow the guide at [https://cloud.google.com/kubernetes-engine/docs/how-to/node-auto-provisioning#config\_file](https://cloud.google.com/kubernetes-engine/docs/how-to/node-auto-provisioning#config_file) to update these limits.

<Info>
  ### CPU and Memory Limits

  TPU VMs often have large amounts of CPU and Memory. Please make sure to also adjust CPU and Memory resource limits in NAP config. Inadequate limits on CPU and Memory might block TPU nodes from scaling
</Info>

Here is what a sample config looks like for adding TPU accelerators

<CodeGroup>
  ```yaml yaml {9-20} lines theme={"dark"}
  resourceLimits:
    - resourceType: 'cpu'
      minimum: 0
      maximum: 1000
    - resourceType: 'memory'
      minimum: 0
      maximum: 10000
    # ...
    - resourceType: 'tpu-v4-podslice'
      minimum: 0
      maximum: 32
    - resourceType: 'tpu-v5-lite-podslice'
      minimum: 0
      maximum: 32
    - resourceType: 'tpu-v5-lite-device'
      minimum: 0
      maximum: 32
    - resourceType: 'tpu-v5p-slice'
      minimum: 0
      maximum: 32
  autoprovisioningLocations:
    # Change these zones according to your cluster
    - us-central2-b
  management:
    autoRepair: true
    autoUpgrade: true
  shieldedInstanceConfig:
    enableSecureBoot: true
    enableIntegrityMonitoring: true
  diskSizeGb: 100
  diskType: pd-balanced
  ```
</CodeGroup>

## Adding TPU Devices

### Using Spec or Python SDK

<Info>
  ### Tip

  You can use the UI to generate equivalent YAML or Python Spec

  <Frame caption="">
    <img src="https://mintcdn.com/truefoundry/DdP_2rhue4AQQlob/images/462a465d-1a47c15-image.png?fit=max&auto=format&n=DdP_2rhue4AQQlob&q=85&s=56243819c84244a050845c5f8fb3c99f" width="1810" height="974" data-path="images/462a465d-1a47c15-image.png" />
  </Frame>
</Info>

You can add TPU by simply adding it under the `resources.devices` section like folllows:

<CodeGroup>
  ```python YAML {8-10} lines theme={"dark"}
  name: my-service
  type: service
  ...
  resources:
    node:
      type: node_selector
    devices:
      - name: tpu-v5-lite-device
        type: gcp_tpu
        topology: 1x1
    cpu_request: 21
    cpu_limit: 23
    memory_request: 39100
    memory_limit: 46000
    ephemeral_storage_request: 20000
    ephemeral_storage_limit: 100000
  ...
  ```
</CodeGroup>

<CodeGroup>
  ```python Python {2,10-12} lines theme={"dark"}
  from truefoundry.deploy import Service, Resources, NodeSelector
  from truefoundry.deploy import GcpTPU, TPUType
  ...


  service = Service(
      name=" my-service",
      ...
      resources=Resources(
          devices=[
              GcpTPU(name=TPUType.V5_LITE_DEVICE, topology="1x1")
          ],
          cpu_request=21,
          cpu_limit=23,
          memory_request=39100,
          memory_limit=46000,
          ephemeral_storage_request=20000,
          ephemeral_storage_limit=100000,
          node=NodeSelector()
      )
  )
  ```
</CodeGroup>

Supported values for name and topologies:

| TPU Type         | Name                   | Supported Topologies | Equivalent Devices Count |
| ---------------- | ---------------------- | -------------------- | ------------------------ |
| v4 PodSlice      | `tpu-v4-podslice`      | `2x2x1`              | 4                        |
| v5 Lite Device   | `tpu-v5-lite-device`   | `1x1`, `2x2`, `2x4`  | 1, 4, 8                  |
| v5 Lite PodSlice | `tpu-v5-lite-podslice` | `1x1`, `2x2`, `2x4`  | 1, 4, 8                  |
| v5p Slice        | `tpu-v5p-slice`        | `2x2x1`              | 4                        |

### Using UI

You can add TPUs to your Service / Job / Notebook / SSH Server by simply selecting a TPU Type and a topology from the **Resources** section

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/FrY4JbiyZud2He3p/images/26783d6a-57e8f82-image.png?fit=max&auto=format&n=FrY4JbiyZud2He3p&q=85&s=5a31b9afd792f883ae40a6caccb42bc5" width="1794" height="792" data-path="images/26783d6a-57e8f82-image.png" />
</Frame>

## Examples

<Info>
  ### Before you begin

  We recommend reading through following pages first to understand basics of deployment with TrueFoundry

  * [TrueFoundry Key Concepts](/docs/key-concepts)
  * [Deploy Service using Python SDK](/docs/deploy-service-programatically)
  * [Deploy and Run Job using Python SDK](/docs/deploy-job-using-python-sdk)
  * [Secrets](/docs/manage-secrets)
</Info>

### Enumerate TPU chips

This simple example Job installs Jax and enumerates the assigned TPU devices

<CodeGroup>
  ```bash truefoundry.yaml lines theme={"dark"}
  name: enumerate-tpu-devices
  type: job
  image:
    type: image
    image_uri: python:3.11
    command: bash /run.sh
  mounts:
    - type: string
      mount_path: /run.sh
      data: |
        #!/bin/bash
        pip install --quiet --upgrade 'jax[tpu]>0.3.0' -f https://storage.googleapis.com/jax-releases/libtpu_releases.html
        python -c 'import jax; print(jax.devices())'
  retries: 0
  trigger:
    type: manual
  resources:
    node:
      type: node_selector
      capacity_type: spot_fallback_on_demand
    devices:
      - name: tpu-v5-lite-device
        type: gcp_tpu
        topology: 1x1
    cpu_request: 21
    cpu_limit: 23
    memory_request: 39100
    memory_limit: 46000
    ephemeral_storage_request: 20000
    ephemeral_storage_limit: 100000
  trigger_on_deploy: true
  ```
</CodeGroup>

On submitting, it would bring up a TPU node and produce the following output

```python lines theme={"dark"}
[TpuDevice(id=0, process_index=0, coords=(0,0,0), core_on_chip=0)]
```

### Running Gemma 2B with Jax

This [Sample Notebook](https://github.com/sanchit-gandhi/notebooks/blob/a72529a4344ad409dc3ab7634f5e9100de242839/jax_gemma.ipynb) demonstrates running Gemma 2B using Jax in an interactive environment. The notebook can be uploaded directly to Notebook instance running with TPUs on TrueFoundry

### Deploying Gemma 2B LLM with Jestream & MaxText on TPU v5e

This section is adapted from official GKE Guide: [https://cloud.google.com/kubernetes-engine/docs/tutorials/serve-gemma-tpu-jetstream](https://cloud.google.com/kubernetes-engine/docs/tutorials/serve-gemma-tpu-jetstream)

**Step 1: Get Access to the model on Kaggle**

* Request Access to the Model on Kaggle: [https://cloud.google.com/kubernetes-engine/docs/tutorials/serve-gemma-tpu-jetstream#sign-consent](https://cloud.google.com/kubernetes-engine/docs/tutorials/serve-gemma-tpu-jetstream#sign-consent)
* Create Kaggle Access Token: [https://cloud.google.com/kubernetes-engine/docs/tutorials/serve-gemma-tpu-jetstream#generate-token](https://cloud.google.com/kubernetes-engine/docs/tutorials/serve-gemma-tpu-jetstream#generate-token) . You will download a `kaggle.json` which we will use in the next step

**Step 2: Add`kaggle.json` to Secrets**

Create a [Secret](/docs/manage-secrets) add the contents of `kaggle.json` like follows.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/JFTbQOWMkMfvFjDC/images/b2d80198-ddfce6e-image.png?fit=max&auto=format&n=JFTbQOWMkMfvFjDC&q=85&s=45ca71ef9aa90237a28da75800fe24ec" width="1630" height="1584" data-path="images/b2d80198-ddfce6e-image.png" />
</Frame>

Once you have saved, keep the secret fqn handy, we'll use it later

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/4MAaF__cLD4iud16/images/661c704e-cf6ec4a-image.png?fit=max&auto=format&n=4MAaF__cLD4iud16&q=85&s=516088f9e96054e8a827d6331943b035" width="2568" height="478" data-path="images/661c704e-cf6ec4a-image.png" />
</Frame>

**Step 3: Create Workspace, GCS Bucket, GCP Service Account and Link Them**

* Create a [Workspace](/docs/key-concepts#creating-a-workspace) to deploy the model to
* Create a GCS Bucket from GCP Console
* Run the following script to create GCP Service Account with correct permissions. Make sure to edit your project id, bucket name, regions, workspace name

<CodeGroup>
  ```bash Shell lines theme={"dark"}
  #!/bin/bash

  gcloud config set project your-project
  PROJECT_ID=$(gcloud config get project)
  BUCKET_NAME=tpu-models-bucket
  REGION=us-central1
  LOCATION=us-central1-a
  WORKSPACE_NAME=tpu-models-workspace
  SA_NAME=tpu-models-sa


  gcloud iam service-accounts create "${SA_NAME}"

  gcloud storage buckets add-iam-policy-binding \
      "gs://${BUCKET_NAME}" \
      --member "serviceAccount:${SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"  \
      --role roles/storage.objectUser \
      --project "${PROJECT_ID}"
  gcloud storage buckets add-iam-policy-binding \
    "gs://${BUCKET_NAME}" \
    --member "serviceAccount:${SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com" \
    --role roles/storage.insightsCollectorService \
    --project "${PROJECT_ID}"

  gcloud iam service-accounts add-iam-policy-binding \
      "${SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com" \
      --role roles/iam.workloadIdentityUser \
      --member "serviceAccount:${PROJECT_ID}.svc.id.goog[${WORKSPACE_NAME}/${SA_NAME}]"
  ```
</CodeGroup>

* Finally edit the workspace you created and add the service account principal

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/DdP_2rhue4AQQlob/images/42c8bb69-b8462a9-image.png?fit=max&auto=format&n=DdP_2rhue4AQQlob&q=85&s=26f908f4f1bd1ef04b32e926b9f5f798" width="1654" height="1602" data-path="images/42c8bb69-b8462a9-image.png" />
</Frame>

**Step 4: Deploy and Run the Job to convert the model to MaxText Engine format**

This Job pulls in the model from Kaggle using the `kaggle.json` secret we created earlier and puts them in your GCS bucket using the serviceaccount we created earlier

<CodeGroup>
  ```python Python lines theme={"dark"}
  import logging
  from truefoundry.deploy import Job, Image, Resources, GcpTPU, TPUType, SecretMount

  logging.basicConfig(level=logging.INFO)

  job = Job(
      name="maxengine-ckpt-conv",
      # Mention your bucket in command below
      image=Image(
          image_uri="us-docker.pkg.dev/cloud-tpu-images/inference/inference-checkpoint:v0.2.0",
          command="/usr/bin/checkpoint_converter.sh -b=tpu-models-bucket -m=google/gemma/maxtext/2b-it/3",
      ),
      # Service account to use to write to the bucket
      service_account="tpu-models-sa",
      # Secret to use for fetching the model checkpoint from Kaggle
      mounts=[
          SecretMount(
              mount_path="/kaggle/kaggle.json",
              secret_fqn="tfy-secret://your-org:kaggle-secrets:kaggle-json"
          )
      ],
      resources=Resources(
          cpu_request=20,
          cpu_limit=20,
          memory_request=150000,
          memory_limit=190000,
          ephemeral_storage_request=30000,
          ephemeral_storage_limit=40000,
          devices=[
              GcpTPU(
                  name=TPUType.V5_LITE_DEVICE,
                  topology="2x2"
              )
          ]
      ),
      trigger_on_deploy=True,
  )

  # Add the fqn of the Workspace we created earlier
  job.deploy(workspace_fqn="<your-workspace-fqn>", wait=False)
  ```
</CodeGroup>

Once this runs successfully, you will get the base and converted checkpoints in your bucket

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/DdP_2rhue4AQQlob/images/2de9378f-5a6aacb-image.png?fit=max&auto=format&n=DdP_2rhue4AQQlob&q=85&s=9c556be9921029fb0df245439afa899b" width="2584" height="790" data-path="images/2de9378f-5a6aacb-image.png" />
</Frame>

**Step 5: Deploy the model via MaxText Engine**

<CodeGroup>
  ```python Python lines theme={"dark"}
  import logging
  from servicefoundry import Service, Image, Port, Resources, GcpTPU, TPUType

  logging.basicConfig(level=logging.INFO)

  service = Service(
      name="maxengine-server",
      # Model name and bucket are mentioned in the command
      image=Image(
          image_uri="us-docker.pkg.dev/cloud-tpu-images/inference/maxengine-server:v0.2.0",
          command="/usr/bin/maxengine_server_entrypoint.sh model_name=gemma-2b tokenizer_path=assets/tokenizer.gemma per_device_batch_size=4 max_prefill_predict_length=1024 max_target_length=2048 async_checkpointing=false ici_fsdp_parallelism=1 ici_autoregressive_parallelism=-1 ici_tensor_parallelism=1 scan_layers=false weight_dtype=bfloat16 load_parameters_path=gs://tpu-models-bucket/final/unscanned/gemma_2b-it/0/checkpoints/0/items",
      ),
      ports=[
          Port(port=9000, expose=False),
      ],
      # Service account to use to read from the bucket
      service_account="tpu-models-sa",
      resources=Resources(
          cpu_request=20,
          cpu_limit=24,
          memory_request=40000,
          memory_limit=48000,
          ephemeral_storage_request=30000,
          ephemeral_storage_limit=40000,
          devices=[
              GcpTPU(
                  name=TPUType.V5_LITE_DEVICE,
                  topology="1x1"
              )
          ]
      ),
  )

  # Add the fqn of the Workspace we created earlier
  service.deploy(workspace_fqn="<your-workspace-fqn>", wait=False)
  ```
</CodeGroup>

This runs the maxengine server in gRPC mode on port 9000

**Step 5: Deploy a HTTP Head Server**

Finally we can deploy a http head server that is just a gRPC client wrapped in FastAPI app

<CodeGroup>
  ```python Python lines theme={"dark"}
  import logging
  from servicefoundry import Service, Image, LocalSource, DockerFileBuild, Port, Resources

  logging.basicConfig(level=logging.INFO)

  service = Service(
      name="jetstream-http",
      image=Image(image_uri="truefoundrycloud/jetstream-http-patched:v0.2.0"),
      ports=[
          Port(
              port=8000,
              host="<Your Host>"
          ),
      ],
      resources=Resources(
          cpu_request=0.5,
          cpu_limit=1,
          memory_request=500,
          memory_limit=1000,
          ephemeral_storage_request=1000,
          ephemeral_storage_limit=4000,
      ),
  )

  # Add the fqn of the Workspace we created earlier
  service.deploy(workspace_fqn="<your-workspace-fqn>", wait=False)
  ```
</CodeGroup>

Once deployed, you can make request with the following body on the `/generate` endpoint

```json lines theme={"dark"}
{
  "server": "maxengine-server.tpu-models-workspace.svc.cluster.local",
  "port": "9000",
  "prompt": "This is a sample prompt ",
  "priority": 0,
  "max_tokens": 500
}
```

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/e3121a39-6a18b0f-image.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=6793cf65d5e75ad0dbf3b89dc39b0bc8" width="3014" height="1484" data-path="images/e3121a39-6a18b0f-image.png" />
</Frame>

***
