> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Finetuning LLMs

> Finetune Llama, Mistral, Mixtral and more on one or more GPUs

LLMs are pre-trained on massive datasets of text and code. This makes them versatile for various tasks, but they may not perform optimally on your specific domain or data.

Finetuning allows you to train these models on your data, enhancing their performance and tailoring them to your unique requirements.

Fine-tuning with TrueFoundry allows you to bring your data, and fine-tune popular Open Source LLM's such as Llama 2, Mistral, Zephyr, Mixtral, and more. This is made easy, as we provide pre-configured options for resources and use the optimal training techniques available. You can choose to perform fine-tuning either using **Jobs** or **Notebooks**. You can further, easily track the progress of finetuning through ML-Repositories.

<Info>
  ### Supported Architectures and Sizes

  We support model sizes of up to 70B for the following model architectures

  * [llama](https://huggingface.co/models?other=llama)
  * [mistral](https://huggingface.co/models?other=mistral)
  * [qwen](https://huggingface.co/models?other=qwen)
  * [qwen2](https://huggingface.co/models?other=qwen2)
  * [gemma](https://huggingface.co/models?other=gemma)
  * [gemma2](https://huggingface.co/models?other=gemma2)
  * [phi](https://huggingface.co/models?other=phi)
  * [phi3](https://huggingface.co/models?other=phi3)
  * [phi4](https://huggingface.co/models?other=phi4)

  Following architectures are supported on best effort basis

  * [mixtral](https://huggingface.co/models?other=mixtral)
  * [gemma3](https://huggingface.co/models?other=gemma3)
  * [falcon](https://huggingface.co/models?other=falcon)
  * [deci](https://huggingface.co/models?other=deci)
  * [mpt](https://huggingface.co/models?other=mpt)
  * [gpt\_bigcode](https://huggingface.co/models?other=gpt_bigcode)
  * [gpt\_neox](https://huggingface.co/models?other=gpt_neox)
</Info>

## QLoRA

For fine-tuning, TrueFoundry embraces the QLoRA technique, a cutting-edge technique that revolutionizes fine-tuning by balancing power and efficiency. This technique uses clever tricks to stay compact, so you can fine-tune on smaller hardware (even just one GPU), saving time, money, and resources, all while maintaining top performance.

## Pre-requisites

Before you begin, ensure you have the following:

* **Workspace**:\
  To deploy your LLM, you'll need a workspace. If you don't have one, you can create it using this guide: [Create a Workspace](/docs/key-concepts#creating-a-workspace) or seek assistance from your cluster administrator.

## Setting up the Training Data

We support two different data formats:

### Chat

Data needs to be in `jsonl` format with each line containing a whole conversation in OpenAI Chat format

Each line contains a key called `messages`. Each `messages` key contains a list of messages, where each message is a dictionary with `role` and `content` keys. The `role` key can be either `user`, `assistant` or `system` and the `content` key contains the message content.

Example:

<CodeGroup>
  ```javascript Chat Type Dataset lines theme={"dark"}
  {"messages": [{"role": "system", "content": "Marv is a factual chatbot that is also sarcastic."}, {"role": "user", "content": "What's the capital of France?"}, {"role": "assistant", "content": "Paris"}, {"role": "user", "content": "Can you be more sarcastic?"}, {"role": "assistant", "content": "Paris, as if everyone doesn't know that already."}]}
  {"messages": [{"role": "system", "content": "Marv is a factual chatbot that is also sarcastic."}, {"role": "user", "content": "Who wrote 'Romeo and Juliet'?"}, {"role": "assistant", "content": "William Shakespeare"}, {"role": "user", "content": "Can you be more sarcastic?"}, {"role": "assistant", "content": "Oh, just some guy named William Shakespeare. Ever heard of him?"}]}
  {"messages": [{"role": "system", "content": "Marv is a factual chatbot that is also sarcastic."}, {"role": "user", "content": "How far is the Moon from Earth?"}, {"role": "assistant", "content": "384,400 kilometers"}, {"role": "user", "content": "Can you be more sarcastic?"}, {"role": "assistant", "content": "Around 384,400 kilometers. Give or take a few, like that really matters."}]}
  ...
  ```
</CodeGroup>

### Completion

Data needs to be in `jsonl` format with each line containing a json encoded string containing two keys prompt and completion.

Example:

<CodeGroup>
  ```javascript Completion Type Dataset lines theme={"dark"}
  {"prompt": "What is 2 + 2?", "completion": "The answer to 2 + 2 is 4"}
  {"prompt": "Flip a coin", "completion": "I flipped a coin and the result is heads!"}
  {"prompt": "<prompt text>", "completion": "<ideal generated text>"}
  ...
  ```
</CodeGroup>

You can further split your data into training data and evaluation data.

Once your data is prepared, you need to store the data somewhere. You can choose where to store your data:

* **TrueFoundry Artifact**: Upload it as a TrueFoundry artifact for easy access.
* **Cloud Storage**: Upload it to a cloud storage service.
* **Local Machine**: Save it directly on your computer.

#### Upload to a TrueFoundry Artifact

If you prefer to upload your training data directly to TrueFoundry as an artifact, follow the [Add Artifacts via UI](/docs/log-artifacts#ui), and Upload your `.jsonl` training data file.

#### Upload to a cloud storage

You can upload your data to a S3 Bucket using the following command:

<CodeGroup>
  ```bash Shell lines theme={"dark"}
  aws s3 cp "path-to-training-data-file-locally" s3://bucket-name/dir-to-store-file
  ```
</CodeGroup>

Once done you can generate a pre-signed URL of the S3 Object using the following command:

<CodeGroup>
  ```shell Shell lines theme={"dark"}
  aws s3 presign s3://bucket-name/path-to-training-data-file-in-s3
  ```
</CodeGroup>

<Frame caption="Output of uploading file to AWS S3 and getting the pre-signed URL">
  <img src="https://mintcdn.com/truefoundry/FrY4JbiyZud2He3p/images/1acce976-50d99fe-upload-output.png?fit=max&auto=format&n=FrY4JbiyZud2He3p&q=85&s=00660bdbba3bcefab158bc80e3cae583" width="1395" height="162" data-path="images/1acce976-50d99fe-upload-output.png" />
</Frame>

Now you can use this pre-signed URL in the fine-tuning job / notebook.

Similarly, you can also upload to [AZURE BLOB](https://learn.microsoft.com/en-us/python/api/azure-storage-blob/azure.storage.blob?view=azure-python#azure-storage-blob-generate-blob-sas) and [GCP GCS](https://cloud.google.com/storage/docs/samples/storage-generate-signed-url-v4#storage_generate_signed_url_v4-python).

## Fine-Tuning a LLM

Now that your data is prepared, you can start the fine-tuning.

Once your data is ready, you can now start fine-tuning your LLM. Here you have two options, deploying a fine-tuning notebook for experimentation or launching a dedicated fine-tuning job.

1. **Notebooks: Experimentation Playground**

Notebooks offer an ideal setup for explorative and iterative fine-tuning. You can experiment on a small subset of data, trying different hyperparameters to figure out the ideal configuration for the best performance. Thanks to the interactive setup, you can analyze the intermediate results to gain deeper insights into the LLM's behavior and response to different training parameters.

Therefore, notebooks are strongly recommended for early-stage exploration and hyperparameter tuning.

2. **Jobs: Reliable and Scalable**

Once you've identified the optimal hyperparameters and configuration through experimentation, transitioning to a deployment job helps you fine-tune on whole dataset and facilitates rapid and reliable training. It ensures consistent and reproducible training runs, as well as built-in retry mechanisms automatically handle any hiccups, ensuring seamless training without manual intervention

Consequently, deployment jobs are the preferred choice for large-scale LLM finetuning, particularly when the optimal configuration has been established through prior experimentation.

## Hyperparameters

Fine-tuning an LLM requires adjusting key parameters to optimize its performance on your specific task. Here are some crucial hyperparameters to consider:

* **Epochs**: This determines the number of times the model iterates through the entire training dataset.\
  Too many epochs can lead to overfitting, and too few might leave the model undertrained. You should start with a moderate number and increase until the validation performance starts dropping.
* **Learning Rate**: This defines how quickly the model updates its weights based on errors.\
  Too high can cause instability and poor performance, and too low can lead to slow learning. Start small and gradually increase if the finetuning is slow.
* **Batch Size**: This controls how many data points the model processes before adjusting its internal parameters. Choose a size based on memory constraints and desired training speed. Too high can strain resources, and too low might lead to unstable updates.
* **Lora Alpha and R**: These control the adaptive scaling of weights in the Lora architecture, improving efficiency for large models. These are useful parameters for generalization. High values might lead to instability, low values might limit potential performance.
* **Max Length** : This defines the maximum sequence length the model can process at once.\
  Choose based on your task's typical input and output lengths. Too short can truncate context, and too long can strain resources and memory.

The optimal values for these hyperparameters depend on your specific LLM, task, and dataset. Be prepared to experiment and iteratively refine your settings for optimal performance.

### Fine-Tuning using a Notebook

<iframe href="https://app.supademo.com/embed/clrrvgftj0swypezg06vnokad" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/clrrvgftj0swypezg06vnokad" style={{ border: "none", display: "flex", margin: "auto" }} />

### Fine-Tuning using a Job

Before you start, you will first need to create an ML Repo (this will be used to store your training metrics and artifacts, such as your checkpoints and models) and give your workspace access to the ML Repo. You can read more about ML Repo's [here](/docs/key-concepts#mlrepo)

<iframe href="https://app.supademo.com/embed/clrrzr3z9124gpezg0c8sjp0c" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/clrrzr3z9124gpezg0c8sjp0c" style={{ border: "none", display: "flex", margin: "auto" }} />

Now that your ML Repo is set up, you can create the fine-tuning job.

<iframe href="https://app.supademo.com/embed/d2cr17Xo7Nhpxc-uWHDDU" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/d2cr17Xo7Nhpxc-uWHDDU" style={{ border: "none", display: "flex", margin: "auto" }} />

## Deploying the Fine-Tuned Model

Once your Fine-tuning is complete, the next step is to deploy the fine-tuned LLM.

<iframe href="https://app.supademo.com/embed/clrrynypm0zpypezg40y9wbtu" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/clrrynypm0zpypezg40y9wbtu" style={{ border: "none", display: "flex", margin: "auto" }} />

You can learn more about how to send requests to your Deploy LLM using the following [guide](/docs/deploying-an-llm-model-from-the-model-catalogue#sending-request-to-your-llm)

## Advanced: Merging LoRa Adapters and uploading merged model

Currently the finetuning Job/Notebook only converts the best checkpoint to a merged model to save GPU compute time. But sometimes we might want to pick an intermediate checkpoint, merge it and re-upload it as a model.

<Warning>
  ### Upcoming Improvments

  We are working on building this feature as part of the platform. Till then the code in this section can be run locally or in a Notebook on the platform
</Warning>

First, let's make sure we have following requirements installed

```javascript lines theme={"dark"}
--extra-index-url https://download.pytorch.org/whl/cu121
torch==2.3.0+cu121
tokenizers==0.19.1
transformers==4.42.3
accelerate==0.31.0
peft==0.11.1
truefoundry>=0.5.5,<1.0.0
```

Make sure you are logged in to TrueFoundry

```html lines theme={"dark"}
tfy login --host <Your TrueFoundry Platform URL>
```

Next, save this script

<CodeGroup>
  ```python merge_and_upload.py lines theme={"dark"}
  import os
  import math
  import os
  import re
  import shutil
  from typing import Any, Dict, Optional

  import numpy as np
  from huggingface_hub import scan_cache_dir
  import argparse

  import json
  from truefoundry.ml import get_client, TransformersFramework
  from transformers import AutoModelForCausalLM, AutoTokenizer
  from peft import PeftModel


  def get_or_create_run(
      ml_repo: str, run_name: str, auto_end: bool = False, create_ml_repo: bool = False
  ):
      client = get_client()
      if create_ml_repo:
          client.create_ml_repo(ml_repo=ml_repo)
      try:
          run = client.get_run_by_name(ml_repo=ml_repo, run_name=run_name)
      except Exception as e:
          if "RESOURCE_DOES_NOT_EXIST" not in str(e):
              raise
          run = client.create_run(ml_repo=ml_repo, run_name=run_name, auto_end=auto_end)
      return run


  def log_model_to_truefoundry(
      run,
      model_name: str,
      model_dir: str,
      hf_hub_model_id: str,
      metadata: Optional[Dict[str, Any]] = None,
      step: int = 0,
  ):
      metadata = metadata or {}
      print("Uploading Model...")
      hf_cache_info = scan_cache_dir()
      files_to_save = []
      for repo in hf_cache_info.repos:
          if repo.repo_id == hf_hub_model_id:
              for revision in repo.revisions:
                  for file in revision.files:
                      if file.file_path.name.endswith(".py"):
                          files_to_save.append(file.file_path)
                  break

      # copy the files to output_dir of pipeline
      for file_path in files_to_save:
          match = re.match(r".*snapshots\/[^\/]+\/(.*)", str(file_path))
          if match:
              relative_path = match.group(1)
              destination_path = os.path.join(model_dir, relative_path)
              os.makedirs(os.path.dirname(destination_path), exist_ok=True)
              shutil.copy(str(file_path), destination_path)
          else:
              print("Python file in hf model cache in unknown path:", file_path)

      metadata.update(
          {
            "pipeline_tag": "text-generation",
            "library_name": "transformers",
            "base_model": hf_hub_model_id,
            "huggingface_model_url": f"https://huggingface.co/{hf_hub_model_id}"
          }
      )
      metadata = {
          k: v
          for k, v in metadata.items()
          if isinstance(v, (int, float, np.integer, np.floating)) and math.isfinite(v)
      }
      run.log_model(
          name=model_name,
          model_file_or_folder=model_dir,
          framework=TransformersFramework(pipeline_tag="text-generation", library_name="transformers", base_model=hf_hub_model_id),
          metadata=metadata,
          step=step,
      )
      print(f"You can view the model at {run.dashboard_link}?tab=models")


  def merge_and_upload(
      hf_hub_model_id: str,
      ml_repo: str,
      run_name: str,
      artifact_version_fqn: str,
      saved_model_name: str,
      dtype: str = "bfloat16",
      device_map: str = "auto",
  ):
      import torch
      client = get_client()
      if device_map.startswith("{"):
          device_map = json.loads(device_map)
      artifact_version = client.get_artifact_version_by_fqn(artifact_version_fqn)
      lora_model_path = artifact_version.download()
      tokenizer = AutoTokenizer.from_pretrained(hf_hub_model_id)
      model = AutoModelForCausalLM.from_pretrained(hf_hub_model_id, device_map=device_map, torch_dtype=getattr(torch, dtype))
      model = PeftModel.from_pretrained(model, lora_model_path)
      model = model.merge_and_unload(progressbar=True)

      merged_model_dir = os.path.abspath("./merged")
      os.makedirs(merged_model_dir, exist_ok=True)
      tokenizer.save_pretrained(merged_model_dir)
      model.save_pretrained(merged_model_dir)
      run = get_or_create_run(
          ml_repo=ml_repo,
          run_name=run_name,
          auto_end=False,
          create_ml_repo=False,
      )
      log_model_to_truefoundry(
          run=run,
          model_name=saved_model_name,
          model_dir=merged_model_dir,
          hf_hub_model_id=hf_hub_model_id,
          metadata={
              "checkpoint": artifact_version_fqn,
          },
          step=artifact_version.step,
      )


  def main():
      parser = argparse.ArgumentParser()
      parser.add_argument("--hf_hub_model_id", type=str, required=True)
      parser.add_argument("--ml_repo", type=str, required=True)
      parser.add_argument("--run_name", type=str, required=True)
      parser.add_argument("--artifact_version_fqn", type=str, required=True)
      parser.add_argument("--saved_model_name", type=str, required=True)
      parser.add_argument("--dtype", type=str, default="bfloat16", choices=["bfloat16", "float16", "float32"])
      parser.add_argument("--device_map", type=str, default="auto")
      args = parser.parse_args()
      merge_and_upload(
          hf_hub_model_id=args.hf_hub_model_id,
          ml_repo=args.ml_repo,
          run_name=args.run_name,
          artifact_version_fqn=args.artifact_version_fqn,
          saved_model_name=args.saved_model_name,
          dtype=args.dtype,
          device_map=args.device_map,
      )


  if __name__ == "__main__":
      main()
  ```
</CodeGroup>

Finally, run this script.

Run `--help` to see help

<CodeGroup>
  ```bash python merge_and_upload.py --help lines theme={"dark"}
  usage: merge_and_upload.py [-h] --hf_hub_model_id HF_HUB_MODEL_ID --ml_repo ML_REPO --run_name RUN_NAME --artifact_version_fqn ARTIFACT_VERSION_FQN
                           --saved_model_name SAVED_MODEL_NAME [--dtype {bfloat16,float16,float32}] [--device_map DEVICE_MAP]

  options:
    -h, --help            show this help message and exit
    --hf_hub_model_id HF_HUB_MODEL_ID
                          HuggingFace Hub model id to merge the LoRa adapter with. E.g. `stas/tiny-random-llama-2`
    --ml_repo ML_REPO     ML repo to log the merged model to
    --run_name RUN_NAME   Name of the run to log the merged model to. If the run does not exist, it will be created.
    --artifact_version_fqn ARTIFACT_VERSION_FQN
                          Artifact version FQN of the LoRa adapter to merge with the HF model
    --saved_model_name SAVED_MODEL_NAME
                          Name of the model to log to MLFoundry
    --dtype {bfloat16,float16,float32}
                          Data type to load the base model
    --device_map DEVICE_MAP
                          device_map to use when loading the model. auto, cpu, or a dictionary of device_map
  ```
</CodeGroup>

E.g. usage which merges checkpoint `artifact:truefoundry/llm-experiments/ckpt-finetune-2024-03-14T05-00-55:7` with its base model `stas/tiny-random-llama-2` and uploads it as `finetuned-tiny-random-llama-checkpoint-7` to `run finetune-2024-03-14T05-00-55` in ML Repo `llm-experiments`

<Frame caption="Grab the checkpoint's artifact version FQN from the Artifacts Tab of your Job Run Output">
  <img src="https://mintcdn.com/truefoundry/PSBc0bX31_cIC7pm/images/d76805ca-278f62f-image.png?fit=max&auto=format&n=PSBc0bX31_cIC7pm&q=85&s=b527c69258b3ee7f7cbfd59178a6ab72" width="2890" height="1586" data-path="images/d76805ca-278f62f-image.png" />
</Frame>

```bash lines theme={"dark"}
python merge_and_upload.py \
  --hf_hub_model_id stas/tiny-random-llama-2 \
  --ml_repo llm-experiments \
  --run_name finetune-2024-03-14T05-00-55 \
  --artifact_version_fqn artifact:truefoundry/llm-experiments/ckpt-finetune-2024-03-14T05-00-55:7 \
  --saved_model_name finetuned-tiny-random-llama-checkpoint-7 \
  --dtype bfloat16 \
  --device_map auto
```
