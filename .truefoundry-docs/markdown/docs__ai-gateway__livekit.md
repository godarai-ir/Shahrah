> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# LiveKit

> Learn how to use livekit with TrueFoundry AI Gateway, including setup steps, use cases, and production-ready examples.

This guide provides instructions for integrating [LiveKit](https://livekit.io/) with the TrueFoundry AI Gateway for real-time AI voice and video applications.

## What is LiveKit?

LiveKit is an open-source platform for building real-time audio and video applications with AI agents. With TrueFoundry AI Gateway integration, you can route your LiveKit AI agent requests through the Gateway for enhanced security, cost tracking, and access controls.

### Key Features of LiveKit

1. **Real-Time Multimodal AI Agents**: LiveKit enables building [conversational AI experiences](https://livekit.io/) with sub-100ms latency for voice and video interactions. Stream audio and video between AI models and users in real-time, creating natural, responsive AI agent conversations with minimal delay.

2. **Scalable WebRTC Infrastructure**: It can handle everything from small applications to [enterprise-scale deployments](https://docs.livekit.io/) with thousands of concurrent participants. The platform provides robust networking with speaker detection, simulcast, and end-to-end encryption.

3. **Comprehensive SDK Ecosystem**: Offers modern, full-featured client SDKs across all major platforms with [consistent APIs and extensive documentation](https://docs.livekit.io/). The platform includes first-class support for integrating AI models, STT, TTS, and real-time APIs, making it easy to build sophisticated AI-powered applications.

## Prerequisites

Before integrating LiveKit with TrueFoundry, ensure you have:

1. **TrueFoundry Account**: Create a [TrueFoundry account](https://www.truefoundry.com/register) and follow the instructions in our [Gateway Quick Start Guide](https://docs.truefoundry.com/gateway/quick-start)
2. **LiveKit Account**: Set up a LiveKit Cloud account or deploy LiveKit server
3. **Python Environment**: Python 3.8+ with pip package manager

## Setup Process

### Step 1: Install Dependencies

```bash lines theme={"dark"}
pip install \
    "livekit-agents[deepgram,openai,cartesia,silero]~=1.0" \
    "python-dotenv"
```

Add to your `.env` file:

```bash lines theme={"dark"}
# TrueFoundry Configuration
OPENAI_API_KEY=<Your TrueFoundry API Key>
TRUEFOUNDRY_BASE_URL=<Your TrueFoundry Gateway URL>
TRUEFOUNDRY_MODEL_ID=<Your TrueFoundry Model ID>

# Third-party Services
DEEPGRAM_API_KEY=<Your Deepgram API Key>
CARTESIA_API_KEY=<Your Cartesia API Key>

# LiveKit Configuration
LIVEKIT_API_KEY=<Your LiveKit API Key>
LIVEKIT_API_SECRET=<Your LiveKit API Secret>
LIVEKIT_URL=<Your LiveKit Server URL>
```

<Frame>
  <img src="https://mintcdn.com/truefoundry/n3EuZuJ0K8wBFp1G/images/new-code-snippet.png?fit=max&auto=format&n=n3EuZuJ0K8wBFp1G&q=85&s=3634c2dc8c3565fd77ab896d3fd07ed9" width="2940" height="1664" data-path="images/new-code-snippet.png" />
</Frame>

You will get your API key, base URL, and model name directly from the unified code snippet above in the TrueFoundry playground. Copy these values into your `.env` file as shown.

### Step 2: Basic LiveKit Agent with TrueFoundry

Create a simple LiveKit agent that uses TrueFoundry's AI Gateway:

```python filename="livekit_agent.py" lines theme={"dark"}
import os
from dotenv import load_dotenv
from livekit import agents
from livekit.agents import AgentSession, Agent, RoomInputOptions
from livekit.plugins import (
    openai,
    cartesia,
    deepgram,
    silero,
)

load_dotenv()


class TrueFoundryAssistant(Agent):
    def __init__(self) -> None:
        super().__init__(
            instructions="You are a helpful AI assistant powered by TrueFoundry. "
                        "You can help with questions, provide information, and have natural conversations. "
                        "Keep your responses conversational and helpful."
        )


async def entrypoint(ctx: agents.JobContext):
    # Create session with TrueFoundry Gateway integration
    session = AgentSession(
        stt=deepgram.STT(
            model="nova-2", 
            language="en"
        ),
        llm=openai.LLM(
            model=os.getenv("TRUEFOUNDRY_MODEL_ID", "openai-main/gpt-4o"),
            api_key=os.getenv("OPENAI_API_KEY"),  # Your TrueFoundry API Key
            base_url=os.getenv("TRUEFOUNDRY_BASE_URL"),  # Your TrueFoundry Gateway URL
            temperature=0.7,
        ),
        tts=cartesia.TTS(
            model="sonic-english",
            voice="79a125e8-cd45-4c13-8a67-188112f4dd22"  # Pleasant voice
        ),
        vad=silero.VAD.load(
            min_speaking_duration=0.3,
            min_silence_duration=0.8,
        ),
    )

    await ctx.connect()

    # Start the session with the assistant
    await session.start(
        room=ctx.room,
        agent=TrueFoundryAssistant(),
    )

    # Send initial greeting
    await session.generate_reply(
        instructions="Welcome the user warmly and let them know you're powered by TrueFoundry. "
                   "Ask how you can help them today."
    )


if __name__ == "__main__":
    agents.cli.run_app(agents.WorkerOptions(entrypoint_fnc=entrypoint))
```

### Step 3: Run Your Agent

```bash lines theme={"dark"}
python livekit_agent.py dev
```

That's it! Your LiveKit agent is now integrated with TrueFoundry Gateway and ready for real-time AI conversations.
