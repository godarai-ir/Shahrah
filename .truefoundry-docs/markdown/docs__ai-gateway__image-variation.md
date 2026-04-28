> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Image Variation API (/images/variations)

> Generate creative variations of existing images using TrueFoundry AI Gateway

## Provider capabilities

The table below summarizes gateway support for this endpoint by provider.

<Info>
  Legend:

  * **✅** Supported by Provider and Truefoundry
  * <Icon icon="circle-xmark" iconType="regular" color="red" /> Provided by provider, but not by Truefoundry
  * <Icon icon="circle-minus" iconType="regular" /> Provider does not support this feature
</Info>

| Provider     | Variation                                                   |
| ------------ | ----------------------------------------------------------- |
| OpenAI       | ✅                                                           |
| Azure OpenAI | <Icon icon="circle-minus" iconType="regular" />             |
| Bedrock      | ✅                                                           |
| Vertex       | <Icon icon="circle-minus" iconType="regular" />             |
| Anthropic    | <Icon icon="circle-minus" iconType="regular" />             |
| Cohere       | <Icon icon="circle-minus" iconType="regular" />             |
| Gemini       | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| Groq         | <Icon icon="circle-minus" iconType="regular" />             |
| Together-AI  | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| xAI          | <Icon icon="circle-minus" iconType="regular" />             |
| DeepInfra    | <Icon icon="circle-xmark" iconType="regular" color="red" /> |

For every gateway endpoint and provider, see [Supported APIs](/docs/ai-gateway/intro-to-llm-gateway#supported-apis).

TrueFoundry AI Gateway supports image variation capabilities that allow you to generate new versions of existing images while preserving the core visual elements and style. This feature is perfect for creating artistic variations, exploring different compositions, or generating multiple options from a single source image.

## Supported Providers

* **OpenAI**: Supports `dall-e-2` model
* **AWS Bedrock**: Supports `amazon-nova-canvas` model

## Requirements

| Provider    | Model              | Format     | Size Limit | Image Count   | Image Requirements |
| ----------- | ------------------ | ---------- | ---------- | ------------- | ------------------ |
| OpenAI      | dall-e-2           | Square PNG | \< 4MB     | 1 image only  | Must be square     |
| AWS Bedrock | amazon-nova-canvas | PNG, JPEG  | -          | Upto 5 images | -                  |

## Example Usage

The meaning of the parameters are present in the [OpenAI API documentation](https://platform.openai.com/docs/api-reference/images/create-variation). The AWS Bedrock specfic parameters are explained [here](https://docs.aws.amazon.com/nova/latest/userguide/image-gen-req-resp-structure.html#:~:text=The%20following%20imageVariationParams%20fields%20are%20used%20in%20this%20request%3A)

<CodeGroup>
  ```python OpenAI lines theme={"dark"}
  from openai import OpenAI

  BASE_URL = "{GATEWAY_BASE_URL}"
  API_KEY = "your-truefoundry-api-key"

  # Configure OpenAI client with TrueFoundry settings
  client = OpenAI(
      api_key=API_KEY,
      base_url=BASE_URL,
  )

  response = client.images.create_variation(
      model="openai-main/dall-e-2",
      image=open("source_image.png", "rb"),  # Source image for variation
      n=3,  # Number of variations to generate
      size="1024x1024"
  )

  print(response)
  ```

  ```python AWS Bedrock lines theme={"dark"}
  from openai import OpenAI

  BASE_URL = "{GATEWAY_BASE_URL}"
  API_KEY = "your-truefoundry-api-key"

  # Configure OpenAI client with TrueFoundry settings
  client = OpenAI(
      api_key=API_KEY,
      base_url=BASE_URL,
  )

  response = client.images.create_variation(
      model="openai-main/dall-e-2",
      image=[ open("source_image.png", "rb"), open("source_image.png", "rb") ]  # Source image for variation
      n=3,  # Number of variations to generate
      size="1024x1024",
      extra_body={
          "similarityStrength": 0.5,
          "negativeText": "cats",
          "cfgScale": 7.0,
          "seed": 42,
          "size": "1024x1024",
      }
  )

  print(response)
  ```
</CodeGroup>

## Response Format

The API returns an `ImagesResponse` object containing:

```json lines theme={"dark"}
{
  "created": 1589478378,
  "data": [
    {
      "url": "https://..."
    },
    {
      "url": "https://..."
    },
    {
      "url": "https://..."
    }
  ]
}

```
