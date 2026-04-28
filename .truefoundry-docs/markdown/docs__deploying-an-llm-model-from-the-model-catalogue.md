> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Deploying LLMs

> Deploy LLMs from the model catalogue. Simple steps to serve models in production.

TrueFoundry simplifies the process of deploying LLM by automaticaly figuring out the most optimal way of deploying any LLM and configuring the correct set of GPUs. We also enable model caching by default and make it really easy to configure autoscaling based on your needs.

## Key Advantages of deploying LLMs via TrueFoundry

1. **Flexibility to choose any model servers**: TrueFoundry support **vLLM, SGLang and TRT-LLM** as model-servers for deploying your LLMs. This allows you to have complete flexibility in choosing the most optimal model-server for your model - hence guaranteeing the faster inference for your LLMs.
2. **Model caching**: LLMs are quite big in size ranging from 5 GB to 150GB. Downloading such models on every restart incurs huge amount of networking cost and leads to lower startup times. TrueFoundry handles the part of downloading the model, caching it and mounting it to all pods - hence providing really fast startup and autoscaling time and lowering network costs.
3. **Image Streaming**: TrueFoundry implemented image streaming and caching which leads to 3X faster download times for vLLM and SGLang images.
4. **Sticky Routing**: In case of LLMs, its advantageous to route requests with the same prefix to the same GPU machine to leverage KV cache optimization. When a model processes a sequence, it stores key-value pairs in memory. If subsequent requests share the same prefix (like a conversation context), routing them to the same instance allows the model to reuse the cached computations, significantly reducing latency and improving throughput. Trueofundry supports sticky routing to make LLM inferences faster.
5. **Inbuilt Observability**: TrueFoundry automatically exposes all the GPU metrics like GPU utilization, temperature and GPU memory for every LLM deployment.
6. **One click addition to LLM Gateway**: The LLMs deployed via TrueFoundry can be added to the Gateway - providing a unified API and governance for all self-hosted models.
7. **Fast and Optimal Autoscaling**: For LLMs, the best metric for autoscaling is requests per second - which is by default supported in TrueFoundry. It also makes autscaling much faster using Model Caching and Image streaming.
8. **Scale to 0**: GPUs can be quite expensive - and hence scale to 0 is a much needed functionality (specially in dev environments) to be able to lower costs when LLMs are not being used.

To deploy a LLM, you can either choose one of the models in the LLM catalogue or **paste the HuggingFace LLM url**. TrueFoundry will try its best to figure out the details from the models page and come up with the best deployment pattern for it. This should work for most of the LLM models - but you might have to tweak in some cases in case we don't find all the information on the HuggingFace page.

## Deploying LLM Model

<Info>
  Before you begin, ensure you have the following:

  * **Workspace**:\
    To deploy your LLM, you'll need a workspace. If you don't have one, you can create it using this guide: [Create a Workspace](/docs/key-concepts#creating-a-workspace)
</Info>

Let's deploy a Llama2-7B LLM through the Model Catalouge

<iframe href="https://app.supademo.com/embed/1RNetCVUSjUp07_3uNvYk" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/1RNetCVUSjUp07_3uNvYk" style={{ border: "none", display: "flex", margin: "auto" }} />

<Info>
  ### Accessing Gated/Private Models

  To access private models enter a [Secret](/docs/manage-secrets) containing your [HuggingFace Token](https://huggingface.co/docs/hub/en/security-tokens) from the HuggingFace account that has access to the model

  <Frame caption="">
    <img src="https://mintcdn.com/truefoundry/5ZMNUF311tgGO_yT/images/hf-gated-models-secret-picker.png?fit=max&auto=format&n=5ZMNUF311tgGO_yT&q=85&s=8410ecbf6c39760e128c446b76a4e8c3" width="1892" height="1292" data-path="images/hf-gated-models-secret-picker.png" />
  </Frame>
</Info>

## Sending requests to your deployed LLM

You can send requests to each LLM through either the **"Completions"** endpoint or the **"Chat Completions"**.

**Note:** Chat Completions Endpoint is available for models which have prompt templates for chat defined.

You can send requests to your LLM using both the normal and streaming methods.

You can get the code to send your request in any language using the OpenAPI tab. Follow the instructions below.

<iframe href="https://app.supademo.com/embed/vBFCewDc0nKvnTd3qsx0W" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/vBFCewDc0nKvnTd3qsx0W" style={{ border: "none", display: "flex", margin: "auto" }} />

<CodeGroup>
  ```json Request Body (Chat) lines theme={"dark"}
  {
    "model": "<YOUR_MODEL_NAME_HERE>",
    "messages": [
      {
        "role": "system",
        "content": "You are a helpful assistant."
      },
      {
        "role": "user",
        "content": "Hello!"
      }
    ]
  }
  ```

  ```json Request Body (Completions) lines theme={"dark"}
  {
    "model": "<YOUR_MODEL_NAME_HERE>",
    "prompt": "This is a test prompt"
  }
  ```
</CodeGroup>

You can find `MODEL_NAME` with a GET request on `<YOUR_ENDPOINT_HERE>/v1/models`Here is a sample response

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/DdP_2rhue4AQQlob/images/2e1d51da-64018df-image.png?fit=max&auto=format&n=DdP_2rhue4AQQlob&q=85&s=dae814a7471233787a7dcedeaaa55057" width="1598" height="482" data-path="images/2e1d51da-64018df-image.png" />
</Frame>

For example, here is a Python code snippet to send a request to your LLM

<CodeGroup>
  ```python Python (Chat Endpoint) lines theme={"dark"}
  # pip install requests

  import json

  import requests

  URL = "<YOUR_ENDPOINT_HERE>/v1/chat/completions"
  headers = {"Content-Type": "application/json", "Accept": "text/event-stream"}
  payload = {
    "model": "<YOUR_MODEL_NAME_HERE>",
    "messages": [
      {
        "role": "system",
        "content": "You are a helpful assistant."
      },
      {
        "role": "user",
        "content": "Hello!"
      }
    ]
  }

  response = requests.post(URL, json=payload, headers=headers)

  print(response.json())
  ```

  ```python Python (Completions Endpoint) lines theme={"dark"}
  # pip install requests

  import json

  import requests

  URL = "<YOUR_ENDPOINT_HERE>/v1/completions"
  headers = {"Content-Type": "application/json", "Accept": "text/event-stream"}
  payload = {
    "model": "llama-2-7b-chat-hf",
    "prompt": "This is a test prompt"
  }

  response = requests.post(URL, json=payload, headers=headers)

  print(response.json())
  ```
</CodeGroup>

#### Making requests via streaming client

Streaming requests allow you to receive the generated text from the LLM as it is being produced, without having to wait for the entire response to be generated. This method is useful for applications that require real-time processing of the generated text.

<CodeGroup>
  ```python Chat Endpoint lines theme={"dark"}

  from openai import OpenAI
  import json

  client = OpenAI(
    base_url="<YOUR_ENDPOINT_HERE>/v1",
    api_key="TEST",
  )
  stream = client.chat.completions.create(
      messages = [
         {"role": "user", "content": "Enter your prompt here"},
      ],
      model= "<YOUR_MODEL_NAME_HERE>",
      stream=True,
      max_tokens=500,
  )
  for chunk in stream:
      if chunk.choices[0].delta.content is not None:
          print(chunk.choices[0].delta.content, end="")
  ```
</CodeGroup>

## Additional Configuration

Optionally, in the advanced options, you might want to

* [Add Authentication to Endpoints](/docs/define-ports-and-domains#adding-basic-authentication-to-service)

  <Info>
    ### Basic Auth with OpenAI SDK

    If you add Basic Authentication with username and password, you can pass them to OpenAI SDK using `default_headers` argument

    <CodeGroup>
      ```python python lines theme={"dark"}
      import base64
      from openai import OpenAI

      username = "..."
      password = "..."
      credentials = f"AME:{password}"
      encoded_credentials = base64.b64encode(credentials.encode()).decode()

      client = OpenAI(
        base_url="<YOUR_ENDPOINT_HERE>/v1",
        api_key="TEST",
        default_headers={"Authorization": f"Basic {encoded_credentials}"},
      )
      ```
    </CodeGroup>
  </Info>

* Configure [Autoscaling](/docs/autoscaling-overview) and [Rollout Strategy](/docs/rollout-strategy)

## Making LLM in TrueFoundry Model Registry Deployable

If you are logging a LLM in TrueFoundry Model Registry, we would need some metadata to make it deployable - Specifically `pipeline_tag`, `library_name` and `base_model` / `huggingface_model_url`

You can add / update these using the `truefoundry` Python SDK. Make sure to complete the [Setup for CLI](/docs/setup-cli) steps

<CodeGroup>
  ```html Shell lines theme={"dark"}
  pip install -U truefoundry
  tfy login --host <Your TrueFoundry Platform Url>
  ```
</CodeGroup>

Here is an example:

Say we finetuned a model based on `Meta-Llama-3-8B-Instruct` and are logging it in TrueFoundry Model Registry, we can add metadata like so:

1. **Updating an already logged model**

<Frame caption="Get the model version FQN">
  <img src="https://mintcdn.com/truefoundry/JFTbQOWMkMfvFjDC/images/b0587c3b-b7c1847-image.png?fit=max&auto=format&n=JFTbQOWMkMfvFjDC&q=85&s=ed46dff497a2f65f4e0b9a82a3a6382e" width="3002" height="790" data-path="images/b0587c3b-b7c1847-image.png" />
</Frame>

<CodeGroup>
  ```python Python lines theme={"dark"}
  from truefoundry.ml import get_client, TransformersFramework

  logged_model_fqn = "<Copy the FQN from UI and Paste>"

  client = get_client()
  model_version = client.get_model_version_by_fqn(fqn=logged_model_fqn)
  model_version.framework = TransformersFramework(
    library_name="transformers",
    pipeline_tag="text-generation",
  )
  model_version.update()
  ```
</CodeGroup>

2. **Or, While logging a new model**

<CodeGroup>
  ```python Python lines theme={"dark"}
  from truefoundry.ml import get_client, TransformersFramework

  client = get_client()
  model_version = client.log_model(
    ml_repo="my-ml-repo",
    name="my-finetuned-llama-3-8b",
    model_file_or_folder="path/to/local/model", # Model location on disk
    framework=TransformersFramework(
        library_name="transformers",
        pipeline_tag="text-generation",
    ),
  )  
  print(model_version.fqn)
  ```
</CodeGroup>

With this done, you should see a deploy button on the UI
