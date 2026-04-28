> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Benchmarking LLMs

> Measure token generation throughput, Time to First Token (TTFT), Inter Token Latency of LLMs via the chat completions API

## Deploying the LLM Benchmarking Tool

The LLM Benchmarking Tool allows you to measure key performance metrics of language models, including token generation throughput, Time to First Token (TTFT), and Inter Token Latency.

### Deployment via Application Catalog

The simplest way to deploy the LLM Benchmarking Tool is through the Application Catalog:

1. Navigate to the Application Catalog and select "Benchmark LLM performance"

<Frame caption="Application Catalog">
  <img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/e1ec180a-e8e605fa3e3fb3bd39d04329833c702b6427babd560aab4864b517390c4b804c-image.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=c69bbfcf631f6327e79800d823b5895a" width="2140" height="850" data-path="images/e1ec180a-e8e605fa3e3fb3bd39d04329833c702b6427babd560aab4864b517390c4b804c-image.png" />
</Frame>

2. Fill in the required deployment information, including Name and Host endpoint for your benchmarking tool.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/f7623298-3ad67c4d4002e04b2af4ea1e308fb631175202c3557872769b6adb0b28add3ee-image.png?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=87ec82447509a40e17244bcf135998c1" width="2144" height="572" data-path="images/f7623298-3ad67c4d4002e04b2af4ea1e308fb631175202c3557872769b6adb0b28add3ee-image.png" />
</Frame>

## Model Configuration

### Basic Configuration Parameters

#### Load Test Parameters

* **Number of users (peak concurrency)**: Maximum number of concurrent users for the load test
* **Ramp up (users started/second)**: Rate at which new users are added to the test
* **Host**: Base URL for LLM API Server with OpenAI compatible endpoints (v1/chat/completions)

#### Model Settings

* **Tokenizer (HuggingFace tokenizer to use to count tokens)**: HuggingFace tokenizer identifier used to count tokens in prompts and responses
* **Model (Name of the model in chat/completions payload)**: Model identifier used in the API request payload

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/jw406UAsc7ErYUq8/images/a15b7934-fe05f205d45529cb1365aac3d1e1d39ce2306aeeaf2c8bc0b77268cbc949d646-image.png?fit=max&auto=format&n=jw406UAsc7ErYUq8&q=85&s=b0b4495dd75d312688ba88728a3af51d" width="1786" height="1164" data-path="images/a15b7934-fe05f205d45529cb1365aac3d1e1d39ce2306aeeaf2c8bc0b77268cbc949d646-image.png" />
</Frame>

#### Finding Parameters for TrueFoundry Deployed Models

When using a model deployed on TrueFoundry, you can find the required parameters in the model's deployment spec:

1. **Model Name**: Look under the `env` section for `MODEL_NAME`

   <CodeGroup>
     ```bash YAML lines theme={"dark"}
     env:
       MODEL_NAME: nousresearch-meta-llama-3-1-8b-instruct
     ```
   </CodeGroup>

2. **Host**: Find the host URL under the `ports` section

   <CodeGroup>
     ```yaml YAML lines theme={"dark"}
     ports:
       - host: example-model.truefoundry.tech
     ```
   </CodeGroup>

3. **Tokenizer**: Look for the `model_id` under `artifacts_download.artifacts`

   <CodeGroup>
     ```python YAML lines theme={"dark"}
     artifacts_download:
       artifacts:
         - type: huggingface-hub
           model_id: NousResearch/Meta-Llama-3.1-8B-Instruct
     ```
   </CodeGroup>

#### Finding Parameters for External Models

When using external models (like GPT-4), you'll need to configure the following parameters:

1. **Model Name**: Use the model identifier from your provider (e.g., `gpt-4`, `gpt-4o`)
2. **Tokenizer**: Find an equivalent tokenizer on HuggingFace (e.g., `Quivr/gpt-4o`)
3. **Host**: Your external provider's API endpoint
4. **API Key**: Your provider's API key (e.g., OpenAI API key)

#### Finding Parameters for TrueFoundry AI Gateway Models

When using models through TrueFoundry's AI Gateway:

1. Navigate to the AI Gateway in your TrueFoundry workspace

2. Select the model you want to benchmark

3. Click on the `</> Code` button to view the API integration code

4. From the code example, you can find:

   * **Host**: The base URL in the request (e.g., `https://truefoundry.tech/api/llm/chat/completions`)
   * **Model Name**: The model identifier in the request payload (e.g., `"model": "openai-main/gpt-4o"`)
   * **OpenAI API Key**: Generate one using the "Generate API Key" button
   * **Tokenizer**: Find an equivalent tokenizer on HuggingFace (e.g., `Quivr/gpt-4o`)

### Prompt Configuration

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/e32ebeb1-9cf6ee10d625f6d137f48807ca45d6476a243367b70db0bddacafb558ac5ebb9-image.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=7d71575784ca38dc02f5b4080b9a81e7" width="1706" height="880" data-path="images/e32ebeb1-9cf6ee10d625f6d137f48807ca45d6476a243367b70db0bddacafb558ac5ebb9-image.png" />
</Frame>

* **Max Output Tokens**: Maximum number of tokens allowed in the model's response
* **Use Random Prompts**: Whether to use randomly generated prompts for testing
* **Use Single Prompt**: Whether to use a single prompt for all test requests
* **Ignore EOS**: Whether to ignore end-of-sequence tokens during token counting

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/FrY4JbiyZud2He3p/images/16f514d3-f0dbe9f76e2d391aeed4fe890f3e691257a05530393738fb2da83c03a032a45e-image.png?fit=max&auto=format&n=FrY4JbiyZud2He3p&q=85&s=dda4a87a35bd5ce4ea5ba4a61f3eedf7" width="1698" height="1226" data-path="images/16f514d3-f0dbe9f76e2d391aeed4fe890f3e691257a05530393738fb2da83c03a032a45e-image.png" />
</Frame>

* **Prompt Min Tokens**: Minimum number of tokens in the input prompt (not used with random or single prompts)
* **Prompt Max Tokens**: Maximum number of tokens in the input prompt (not used with random or single prompts)

## Viewing Benchmark Results

After running the benchmark, you'll see comprehensive performance metrics displayed in charts:

* Requests per Second
* Active Users
* Tokens per Second
* Response Time Seconds
* Response Time First Token (ms)
* Inter Token Latency (ms)

<Frame caption="Benchmarking Results">
  <img src="https://mintcdn.com/truefoundry/4MAaF__cLD4iud16/images/5b6591fd-e7f9b5b307fef272ff4170873fe5a4fda3c7e79d5ae51a6c47603d6fb310312b-image.png?fit=max&auto=format&n=4MAaF__cLD4iud16&q=85&s=c0a527149fbc77f9cac0947dea8fbb39" width="3004" height="1782" data-path="images/5b6591fd-e7f9b5b307fef272ff4170873fe5a4fda3c7e79d5ae51a6c47603d6fb310312b-image.png" />
</Frame>

***
