> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Deploying NVIDIA NIM Models

> Deploy optimized TensorRT-LLM Engines using NIM Containers

<Info>
  Supported Model Types

  Currently we list NIM models of following types

  * Large Language Models (LLMs)
  * Vision Language Models (VLMs)
  * Embedding Models
  * Reranking Models
</Info>

## Adding `nvcr.io` Docker Registry

1. Generate an API Key from [https://org.ngc.nvidia.com/setup/api-keys](https://org.ngc.nvidia.com/setup/api-keys)

   Make sure to give it access to `NGC Catalog`

   <img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/docs/nim-ngc-api-key.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=e54f9c1f9a441521d30c688aa917ee74" alt="image.png" width="3010" height="1554" data-path="images/docs/nim-ngc-api-key.png" />
2. Add a Custom Docker Registry to the Platform

   * Registry URL: `nvcr.io`
   * Username: `$oauthtoken`
   * Password: The API Key from the previous step

   <img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/docs/nvcr-docker-registry.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=e3cbbe09592ca8f1a544cc0b1419fa98" alt="image.png" width="3004" height="1598" data-path="images/docs/nvcr-docker-registry.png" />

## Adding NGC API Key to Secrets

1. Add the same API Key as a [Secret](/docs/manage-secrets) on the Platform. We are calling the secret `NGC_API_KEY`

   <img src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/add-ngc-api-key-as-secret.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=1952ebd8b42a0ece7fb81b12fbe1e886" alt="Add Ngc Api Key As Secret Pn" width="3006" height="980" data-path="images/docs/add-ngc-api-key-as-secret.png" />

## Deploying a NIM Model

1. From `New Deployment` page, select `NVIDIA NIM`.

   * Select the workspace you want to deploy to
   * Select the NVCR Model Registry Integration we created in the previous step
   * Select the NGC API Key Secret we created in the previous step
   * Select the model you want to deploy

   <img src="https://mintcdn.com/truefoundry/5CkapnZ7CyjQJ4bx/images/docs/new-deployment-nvidia-nim.png?fit=max&auto=format&n=5CkapnZ7CyjQJ4bx&q=85&s=6c3178dc063d1f4e9847411fab510106" alt="image.png" width="1878" height="1596" data-path="images/docs/new-deployment-nvidia-nim.png" />
2. Click Next. You will be presented with optimized profiles (for `latency` or `throughput` ) for differerent precision and GPU options for which TRT-LLM engines are prebuilt and available. You can select any of the profile and Continue to Deployment.

   <img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/docs/nim-deployment-options.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=128b4ef6962eda63e39f26d86486d90e" alt="Nim Deployment Options Pn" width="1890" height="1604" data-path="images/docs/nim-deployment-options.png" />

## (Optional) Caching NIM Model to External Volume

<Tip>
  Recommended for Large Models and Production Environments
</Tip>

To avoid re-downloading Model on every restart, you can [Create a Volume](/docs/creating-a-volume) and [Mount the Volume ](/docs/mounting-volumes-service)at `/opt/nim/.cache`

<img src="https://mintcdn.com/truefoundry/4MAaF__cLD4iud16/images/4fcd568c-5d90c572d92dd8c09ec0590fba11f5058f486f2d663594b95086fec59f074340-image.png?fit=max&auto=format&n=4MAaF__cLD4iud16&q=85&s=46b0f10b7657b450d7ed3a4a507a8623" alt="4fcd568c 5d90c572d92dd8c09ec0590fba11f5058f486f2d663594b95086fec59f074340 Image Pn" width="1872" height="1586" data-path="images/4fcd568c-5d90c572d92dd8c09ec0590fba11f5058f486f2d663594b95086fec59f074340-image.png" />

## Running Inferences

You can now run inferences via the OpenAPI tab. You can also [Add the Model to LLM Gateway](/docs/ai-gateway/self-hosted-models) using the button on the top

<img src="https://mintcdn.com/truefoundry/s4Aj2_qGCrSP-zc8/images/84390370-9d4356719c8ce8f41dc04108e037fab0f980d78a06cbbb8f5d85f55bac0d5887-image.png?fit=max&auto=format&n=s4Aj2_qGCrSP-zc8&q=85&s=3bcb22f7b2cf25a2a5f9c9ce783d5dcf" alt="84390370 9d4356719c8ce8f41dc04108e037fab0f980d78a06cbbb8f5d85f55bac0d5887 Image Pn" width="3000" height="1578" data-path="images/84390370-9d4356719c8ce8f41dc04108e037fab0f980d78a06cbbb8f5d85f55bac0d5887-image.png" />
