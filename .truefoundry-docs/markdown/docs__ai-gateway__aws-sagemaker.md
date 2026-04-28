> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# AWS Sagemaker

> Configure and use AWS Sagemaker models through TrueFoundry's AI Gateway

### Adding Models

This section explains the steps to add AWS Sagemaker models and configure the required access controls.

<Steps>
  <Step title="Navigate to AWS Sagemaker Models in AI Gateway">
    From the TrueFoundry dashboard, navigate to `AI Gateway` > `Models` and select `AWS Sagemaker`.

    <Frame caption="Navigate to AWS Sagemaker Models">
      <img src="https://mintcdn.com/truefoundry/eUE5tK_OsaTr86Ss/images/create_sagemaker_model.png?fit=max&auto=format&n=eUE5tK_OsaTr86Ss&q=85&s=9ca348851f20d6d76f938748dad0d01e" alt="Navigating to AWS Sagemaker Provider Account in AI Gateway" width="3014" height="1492" data-path="images/create_sagemaker_model.png" />
    </Frame>
  </Step>

  <Step title="Add AWS Sagemaker Account Name and Collaborators">
    Give a unique name for the Sagemaker account which will be used to refer later in the models. The models in the account will be referred to as `@providername/@modelname`. Add collaborators to your account. You can decide which users/teams have access to the models in the account (**User Role**) and who can add/edit/remove models in this account (**Manager Role**). You can read more about access control [here](/docs/ai-gateway/gateway-access-control).

    <Frame caption="AWS Sagemaker Model Account Form">
      <img src="https://mintcdn.com/truefoundry/KEC3h_6YJZSnQGY8/images/add_sagemaker_provider_account.png?fit=max&auto=format&n=KEC3h_6YJZSnQGY8&q=85&s=07e5ca37ab3eb6e656c25a3c83f7e552" alt="AWS Sagemaker account configuration form with fields for API key and collaborators" width="3016" height="1500" data-path="images/add_sagemaker_provider_account.png" />
    </Frame>
  </Step>

  <Step title="Add Region and Authentication">
    Select the default AWS region for the models in this account. \*\*The account-level region serves as the default for all models unless explicitly overridden at the model level. \*\*Provide the authentication details on how the gateway can access the Sagemaker models. Truefoundry supports both AWS Access Key/Secret Key and Assume Role based authentication. You can read below on how to generate the access/secret keys or roles.

    <Accordion title="Get AWS Authentication Details">
      **Using AWS Access Key and Secret**

      1. Create an IAM user (or choose an existing IAM user) following [these steps](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_create.html).
      2. Add required permission for this user. The following policy grants permission to invoke all model in your available regions (To check the list of available regions for different models, refer to [AWS Sagemaker](https://docs.aws.amazon.com/sagemaker/latest/dg/api-permissions-reference.html?)).
      3.       <CodeGroup>
                 ```json JSON lines theme={"dark"}
                 {
                   "Version": "2012-10-17",
                   "Statement": [
                     {
                       "Effect": "Allow",
                       "Sid": "InvokeAllModels",
                       "Action": [
                         "sagemaker:InvokeEndpoint",
                       ],
                       "Resource": [
                         "arn:aws:sagemaker:<region>:<account-id>:endpoint/<endpoint-name>"
                 	    ]
                     }
                   ]
                 }
                 ```
               </CodeGroup>
      4. Create an access key for this user [as per this doc](https://docs.aws.amazon.com/IAM/latest/UserGuide/access-keys-admin-managed.html#admin-create-access-key).
      5. Use this access key and secret while adding the provider account to authenticate requests to the Sagemaker model.

      **Using Assumed Role**

      1. You can also directly specify a role that can be assumed by the service account attached to the pods running AI Gateway.
      2. Read more about how assumed roles work [here](https://docs.aws.amazon.com/STS/latest/APIReference/API_AssumeRole.html).
    </Accordion>
  </Step>

  <Step title="Add Models">
    * To add the model, first enter the model name to save and access it in truefoundry. Then, enter the model ID and the region where the model is deployed.
    * The model ID is the name of the endpoint in sagemaker. You can go to the endpoint page in sagemaker and copy the endpoint name.

    <Frame caption="Navigate to Endpoints in Sagemaker and copy the endpoint name which is your model id">
      <img src="https://mintcdn.com/truefoundry/iHnxgGrqAWdesLuk/images/get_sagemaker_model_endpoint.png?fit=max&auto=format&n=iHnxgGrqAWdesLuk&q=85&s=7e4aa925b6356d3f1f9912527ad6ac26" alt="Sagemaker model configuration form in TrueFoundry with fields for model name, model ID and region" width="3022" height="1308" data-path="images/get_sagemaker_model_endpoint.png" />
    </Frame>

    <Frame caption="Add Sagemaker Model in TrueFoundry">
      <img src="https://mintcdn.com/truefoundry/KEC3h_6YJZSnQGY8/images/add_sagemaker_model.png?fit=max&auto=format&n=KEC3h_6YJZSnQGY8&q=85&s=7da258157c11817225e53d95590fcb38" alt="Sagemaker model configuration form in TrueFoundry with fields for model name, model ID and region" width="3020" height="1496" data-path="images/add_sagemaker_model.png" />
    </Frame>
  </Step>
</Steps>

### Inference

After adding the models, you can perform inference using an OpenAI-compatible API via the Playground or integrate with your own application.

* To try in playground, click on the `Try in Playground` button.

  <Info>
    Only chat completion is supported in Playground, to use embedding, you can
    use the code snippet to integrate in your application.
  </Info>

  <Frame caption="Infer Model in Playground or Get Code Snippet to integrate in your application">
    <img
      src="https://mintcdn.com/truefoundry/RQu8Oq_dVhJCS0cG/images/sagemaker_try_in_playground.png?fit=max&auto=format&n=RQu8Oq_dVhJCS0cG&q=85&s=c5beba6eeaf059559d2c4b6977b16a2e"
      alt="Code Snippet and Try in Playgroud Buttons for each
  model"
      width="3006"
      height="1526"
      data-path="images/sagemaker_try_in_playground.png"
    />
  </Frame>

* To get the code snippet to inference the model in your application using OpenAI-compatible API, click on the `Get Code Snippet` button.

  <Frame caption="Get Code Snippet to integrate in your application">
    <img
      src="https://mintcdn.com/truefoundry/iHnxgGrqAWdesLuk/images/get_code_snippet_playground.png?fit=max&auto=format&n=iHnxgGrqAWdesLuk&q=85&s=def6f16e88195f42ac3cf9a189d7a45c"
      alt="Code Snippet to integrate in your
  application"
      width="2984"
      height="1022"
      data-path="images/get_code_snippet_playground.png"
    />
  </Frame>

## Access other models with gateway using proxy

If you want to access other models type like reranker or traditional machine learning models hosted on sagemaker, you can use the proxy endpoint to access them by using truefoundry gateway.

To use the proxy endpoint, you can use the following endpoint to proxy request via truefoundry gateway:

```
{GATEWAY_BASE_URL}/v1/endpoints/{modelId}/invocations
```

Here, `{modelId}` is the model id of the model you want to access, you can get it from the endpoints page in sagemaker.

Also you have to give the provider name in the header in the following format:

```
X-TFY-PROVIDER-NAME: sagemaker
```

Also create a sagmaker provider integration in truefoundry gateway and select the type of model as embeddings or chat.

you can see the example below on how to proxy the request to model in python:

```python lines theme={"dark"}
import requests
import json

URL = "{GATEWAY_BASE_URL}/v1/endpoints/{modelId}/invocations"
API_KEY = "your-truefoundry-api-key"

headers = {
    "Content-Type": "application/json",
    "Authorization": f"Bearer {API_KEY}",
    "X-TFY-LOGGING-CONFIG": '{"enabled": true}'
    "X-TFY-PROVIDER-NAME": "sagemaker" # sagemaker-provider-name
}

payload = {
    "model": "sagemaker/modelId", # Your TrueFoundry Model name
    ... # other input request parameters which are supported by the model
}

response = requests.post(URL, headers=headers, json=payload)
print(response.json())
```

### FAQ:

<Accordion title="How to override the default cost of models?">
  In case you have custom pricing for your models, you can override the default cost by clicking on Edit Model button and then choosing the `Private Cost Metric` option.

  <Columns>
    <Frame caption="Edit Model">
      <img src="https://mintcdn.com/truefoundry/5CkapnZ7CyjQJ4bx/images/docs/edit-gateway-model.png?fit=max&auto=format&n=5CkapnZ7CyjQJ4bx&q=85&s=19ad20b9cb350dcb5ac5b9af5e6c94f6" alt="Edit model button and interface for AWS Sagemaker model" width="3840" height="1936" data-path="images/docs/edit-gateway-model.png" />
    </Frame>

    <Frame caption="Set custom cost metric">
      <img src="https://mintcdn.com/truefoundry/5CkapnZ7CyjQJ4bx/images/docs/edit-private-cost-gateway-model.png?fit=max&auto=format&n=5CkapnZ7CyjQJ4bx&q=85&s=8159f0e17e523b68634ec00f920e74ed" alt="Custom cost metric configuration form with input fields for pricing" width="3840" height="1936" data-path="images/docs/edit-private-cost-gateway-model.png" />
    </Frame>
  </Columns>
</Accordion>
