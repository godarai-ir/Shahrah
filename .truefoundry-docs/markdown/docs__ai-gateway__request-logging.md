> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# View Traces

> Control which requests are traced using headers or environment variables.

TrueFoundry AI Gateway allows you to control which requests are logged.

## Controlling Request Logging

### Using HTTP Headers

To enable logging for a specific request, include the `X-TFY-LOGGING-CONFIG` header:

```json lines theme={"dark"}
{
  "enabled": true
}
```

To disable logging for a specific request, set:

```json lines theme={"dark"}
{
  "enabled": false
}
```

### Using Global Settings

You can also control logging behavior globally using AI Gateway global settings:

<Note>
  In the AI Gateway UI, click `Settings` in the left sidebar to open global settings.
</Note>

| Mode                | Description                                                                                                                                                                           |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `HEADER_CONTROLLED` | Logging depends on the `enabled` value in the `X-TFY-LOGGING-CONFIG` header. If the header is absent or set to `true`, logging will occur. If set to `false`, no logging will happen. |
| `ALWAYS`            | All requests are logged regardless of the `enabled` value.                                                                                                                            |
| `NEVER`             | No requests are logged regardless of the `enabled` value.                                                                                                                             |

## Viewing Request Logs

You can view all logged requests in the TrueFoundry UI. Go to `AI Gateway` > `Monitor` > `Requests`

<Frame caption="Request logs in the Monitor section">
  <img src="https://mintcdn.com/truefoundry/QPyNP8qMPAoy2O9d/images/gateway-request-logs.png?fit=max&auto=format&n=QPyNP8qMPAoy2O9d&q=85&s=54fd4487da2d27caa008c782abdcec48" width="3014" height="1582" data-path="images/gateway-request-logs.png" />
</Frame>

## Code Example

```python lines theme={"dark"}
from openai import OpenAI

BASE_URL = "{GATEWAY_BASE_URL}"
API_KEY = "your-truefoundry-api-key"

client = OpenAI(
    api_key=API_KEY,
    base_url=BASE_URL,
    default_headers={
        "X-TFY-LOGGING-CONFIG": '{"enabled": true}'
    }
)
```

## Data Access Rules for Request Logs and Metrics

Please refer to [Data Access Rules for Traces and Metrics](/docs/ai-gateway/data-access) for more information.
