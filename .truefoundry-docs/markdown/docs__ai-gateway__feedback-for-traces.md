> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Log Feedback on requests

> Learn how to create, modify, and retrieve feedback for trace spans.

## Providing Feedback for Traces

You can create, modify, or delete feedback for trace spans. Feedback includes a rating from 1 (lowest) to 5 (highest) to evaluate the quality of a trace span.

### Creating New Feedback

You can create new feedback for traces using the API. Specify the target trace span using `feedbackTargetId` (returned in the headers of every AI Gateway response as `x-tfy-feedback-target-id`).

The feedback target ID is returned in the headers of every AI Gateway response as `x-tfy-feedback-target-id`. Currently, the feedback target ID only represents the root span in the trace.

```python lines theme={"dark"}
import requests

# Make API request
response = requests.post(
    "https://{control_plane_url}/api/svc/v1/gateway-feedback",
    headers={
        "Authorization": "Bearer YOUR_API_TOKEN",
        "Content-Type": "application/json"
    },
    json={
        "target": {
          "feedbackTargetId": "eyJzcGFuSWQiOiJhMmQ2Yjg4MmFhMTViZTU1IiwidHJhY2VJZCI6IjAxOWFiMmE5NmQ1MDcwN2ZhMzA5MGMxYjcyMDFkNWJkIn0="
        },
        "rating": 4,
        "comment": "Optional comment", # Optional
        "metadata": {"key": "value"} # Optional
    }
)

response.raise_for_status()
data = response.json()
```

The response contains the feedback ID, which can be used to modify or delete the feedback.

```json lines theme={"dark"}
{
  "data": {
    "id": "019ab2a98fc6e80483eae496e2dd611d"
  }
}
```

### Modifying Existing Feedback

Existing feedback can be modified using the following API request:

The feedback target ID is returned in the headers of every AI Gateway response as `x-tfy-feedback-target-id`. Currently, the feedback target ID only represents the root span in the trace.

```python lines theme={"dark"}
import requests

# Make API request
response = requests.put(
    "https://{control_plane_url}/api/svc/v1/gateway-feedback/019ab2a98fc6e80483eae496e2dd611d",
    headers={
        "Authorization": "Bearer YOUR_API_TOKEN",
        "Content-Type": "application/json"
    },
    json={
        "target": {
          "feedbackTargetId": "eyJzcGFuSWQiOiJhMmQ2Yjg4MmFhMTViZTU1IiwidHJhY2VJZCI6IjAxOWFiMmE5NmQ1MDcwN2ZhMzA5MGMxYjcyMDFkNWJkIn0="
        },
        "rating": 4,
        "comment": "Optional comment", # Optional
        "metadata": {"key": "value"} # Optional
    }
)

response.raise_for_status()
data = response.json()
```

### Deleting Existing Feedback

Existing feedback can be deleted using the following API request:

The feedback target ID is returned in the headers of every AI Gateway response as `x-tfy-feedback-target-id`. Currently, the feedback target ID only represents the root span in the trace.

```python lines theme={"dark"}
import requests

# Make API request
response = requests.delete(
    "https://{control_plane_url}/api/svc/v1/gateway-feedback/019ab2a98fc6e80483eae496e2dd611d",
    headers={
        "Authorization": "Bearer YOUR_API_TOKEN",
        "Content-Type": "application/json"
    },
    json={
        "target": {
          "feedbackTargetId": "eyJzcGFuSWQiOiJhMmQ2Yjg4MmFhMTViZTU1IiwidHJhY2VJZCI6IjAxOWFiMmE5NmQ1MDcwN2ZhMzA5MGMxYjcyMDFkNWJkIn0="
        }
    }
)

response.raise_for_status()
data = response.json()
```

## Reading Feedback

Feedback for any given trace span is displayed alongside the span in the trace view.

<Frame caption="AI Gateway - Feedback for trace span">
  <img src="https://mintcdn.com/truefoundry/05v1PNKB_X4eGMQ9/images/feedback-for-trace-with-star.png?fit=max&auto=format&n=05v1PNKB_X4eGMQ9&q=85&s=e7251528a2173a47520a65f203a02547" alt="Feedback for trace" width="1562" height="232" data-path="images/feedback-for-trace-with-star.png" />
</Frame>

You can also view the complete list of feedback for a trace span in the `Raw Data` section under the `tfyGatewayFeedbacks` field.
