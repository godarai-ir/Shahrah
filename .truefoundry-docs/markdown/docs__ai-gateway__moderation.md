> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Moderation API (/moderations)

> Learn how to use the OpenAI Moderations API through TrueFoundry's AI Gateway to identify potentially harmful content in text and images.

## Provider capabilities

The table below summarizes gateway support for this endpoint by provider.

<Info>
  Legend:

  * **✅** Supported by Provider and Truefoundry
  * <Icon icon="circle-xmark" iconType="regular" color="red" /> Provided by provider, but not by Truefoundry
  * <Icon icon="circle-minus" iconType="regular" /> Provider does not support this feature
</Info>

| Provider     | Moderation                                                  |
| ------------ | ----------------------------------------------------------- |
| OpenAI       | ✅                                                           |
| Azure OpenAI | <Icon icon="circle-minus" iconType="regular" />             |
| Anthropic    | <Icon icon="circle-minus" iconType="regular" />             |
| Bedrock      | <Icon icon="circle-minus" iconType="regular" />             |
| Vertex       | <Icon icon="circle-minus" iconType="regular" />             |
| Cohere       | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| Gemini       | <Icon icon="circle-minus" iconType="regular" />             |
| Groq         | <Icon icon="circle-minus" iconType="regular" />             |
| Cerebras     | <Icon icon="circle-minus" iconType="regular" />             |
| Together-AI  | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| xAI          | <Icon icon="circle-minus" iconType="regular" />             |
| DeepInfra    | <Icon icon="circle-xmark" iconType="regular" color="red" /> |

For every gateway endpoint and provider, see [Supported APIs](/docs/ai-gateway/intro-to-llm-gateway#supported-apis).

### Supported Providers

* OpenAI

Use the [moderations](https://platform.openai.com/docs/api-reference/moderations) endpoint to identify content that may violate usage policies, such as text or images involving violence, hate speech, harassment, self-harm, or sexual content.

You can use two models for this endpoint:

* `omni-moderation-latest`: This model and all snapshots support more categorization options and multi-modal inputs.
* `text-moderation-latest` **(Legacy)**: Older model that supports only text inputs and fewer input categorizations. The newer omni-moderation models will be the best choice for new applications.

You can add these models in your OpenAI provider account.

Code Snippets

```python lines theme={"dark"}
from openai import OpenAI

BASE_URL = "{GATEWAY_BASE_URL}"
API_KEY = "your-truefoundry-api-key"

# Configure OpenAI client with TrueFoundry settings
client = OpenAI(
    api_key=API_KEY,
    base_url=BASE_URL,
)

response = client.moderations.create(
    model="openai-main/omni-moderation-latest",
    input=[
        {"type": "text", "text": "Text to check for moderation"},
        {
            "type": "image_url",
            "image_url": {
                "url": "https://example.com/image.png",
            }
        },
    ],
)

print(response)
```

Here's a full example output, where the input is an image from a single frame of a war movie. The model correctly predicts indicators of violence in the image, with a `violence` category score of greater than 0.8.

```
{
  "id": "modr-970d409ef3bef3b70c73d8232df86e7d",
  "model": "omni-moderation-latest",
  "results": [
    {
      "flagged": true,
      "categories": {
        "sexual": false,
        "sexual/minors": false,
        "harassment": false,
        "harassment/threatening": false,
        "hate": false,
        "hate/threatening": false,
        "illicit": false,
        "illicit/violent": false,
        "self-harm": false,
        "self-harm/intent": false,
        "self-harm/instructions": false,
        "violence": true,
        "violence/graphic": false
      },
      "category_scores": {
        "sexual": 2.34135824776394e-7,
        "sexual/minors": 1.6346470245419304e-7,
        "harassment": 0.0011643905680426018,
        "harassment/threatening": 0.0022121340080906377,
        "hate": 3.1999824407395835e-7,
        "hate/threatening": 2.4923252458203563e-7,
        "illicit": 0.0005227032493135171,
        "illicit/violent": 3.682979260160596e-7,
        "self-harm": 0.0011175734280627694,
        "self-harm/intent": 0.0006264858507989037,
        "self-harm/instructions": 7.368592981140821e-8,
        "violence": 0.8599265510337075,
        "violence/graphic": 0.37701736389561064
      },
      "category_applied_input_types": {
        "sexual": [
          "image"
        ],
        "sexual/minors": [],
        "harassment": [],
        "harassment/threatening": [],
        "hate": [],
        "hate/threatening": [],
        "illicit": [],
        "illicit/violent": [],
        "self-harm": [
          "image"
        ],
        "self-harm/intent": [
          "image"
        ],
        "self-harm/instructions": [
          "image"
        ],
        "violence": [
          "image"
        ],
        "violence/graphic": [
          "image"
        ]
      }
    }
  ]
}
```

#### Key Fields

* `flagged`: `true` if any content is considered harmful.
* `categories`: Dictionary of category flags, `true` means the input falls under that harmful category.
* `category_scores`: Probability scores (0 to 1) for each category.
* `category_applied_input_types` : Indicates whether the category was triggered by the text or image.
