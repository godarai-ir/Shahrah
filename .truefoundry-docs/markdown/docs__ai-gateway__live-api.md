> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Live / Realtime API (/live)

> Bidirectional realtime streaming (Live API) over WebSocket through TrueFoundry's AI Gateway

## Provider capabilities

The table below summarizes gateway support for this endpoint by provider.

<Info>
  Legend:

  * **✅** Supported by Provider and Truefoundry
  * <Icon icon="circle-xmark" iconType="regular" color="red" /> Provided by provider, but not by Truefoundry
  * <Icon icon="circle-minus" iconType="regular" /> Provider does not support this feature
</Info>

| Provider         | Live / Realtime API |
| ---------------- | ------------------- |
| Gemini           | ✅                   |
| Vertex           | ✅                   |
| OpenAI           | ✅                   |
| Azure AI Foundry | ✅                   |

For every gateway endpoint and provider, see [Supported APIs](/docs/ai-gateway/intro-to-llm-gateway#supported-apis).

The **Realtime API** enables low-latency, bidirectional streaming over a persistent WebSocket connection. You can send and receive text and audio in real time, enabling use cases like voice assistants and interactive agents. The gateway proxies the WebSocket connection to the provider using each provider's native SDK.

| Approach                        | Supported providers                                                      | Base path                                                               |
| ------------------------------- | ------------------------------------------------------------------------ | ----------------------------------------------------------------------- |
| **Provider proxy** (native SDK) | Google Gemini, Google Vertex AI, OpenAI, Azure AI Foundry / Azure OpenAI | `{GATEWAY_BASE_URL}/live/{providerAccountName}` (use `wss://` protocol) |

**Before you start:** Replace `{GATEWAY_BASE_URL}` with your AI Gateway Base URL ([how to find it](/docs/ai-gateway/quick-start#gateway-base-url)) and `your-tfy-api-key` with your TrueFoundry API key. Replace `{providerAccountName}` with the display name of your provider account on TrueFoundry. For WebSocket connections, use the `wss://` protocol with the gateway host.

<Note>
  **Model names:** The model ID in code must match the **display name** of the model on your TrueFoundry provider account.
</Note>

<Note>
  **Which SDK to use:** Use the `google-genai` Python SDK for Google Gemini and Google Vertex AI, the `openai` Python SDK for OpenAI and Azure AI Foundry, and the `azure-ai-voicelive` Python SDK for Azure AI Foundry (alternative) — all pointed at the gateway WebSocket URL above.
</Note>

## Add models to the gateway

Before you can use the Realtime API, add your realtime models to TrueFoundry through a provider account. When adding a model, select **Realtime** as the model type.

| Provider         | Setup guide                                           |
| ---------------- | ----------------------------------------------------- |
| Google Gemini    | [Google Gemini](/docs/ai-gateway/google-gemini)       |
| Google Vertex AI | [Google Vertex](/docs/ai-gateway/google-vertex)       |
| OpenAI           | [OpenAI](/docs/ai-gateway/openai)                     |
| Azure AI Foundry | [Azure AI Foundry](/docs/ai-gateway/azure-ai-foundry) |
| Azure OpenAI     | [Azure OpenAI](/docs/ai-gateway/azure-openai)         |

## Code snippet

After adding the models, you can get a ready-to-use code snippet from the TrueFoundry platform or use the examples below.

The example below demonstrates a realtime audio session, streaming microphone input to the model and playing back audio responses through the speaker. You can adapt the code to use other modalities as needed.

<AccordionGroup>
  <Accordion title="Google Gemini">
    ```python lines theme={"dark"}
    """
    Gemini Live API - Realtime Audio Streaming
    pip install google-genai pyaudio
    """
    import asyncio
    import pyaudio
    from google import genai
    from google.genai import types

    FORMAT = pyaudio.paInt16
    CHANNELS = 1
    SEND_SAMPLE_RATE = 16000
    RECEIVE_SAMPLE_RATE = 24000
    CHUNK_SIZE = 1024

    API_KEY = "your-tfy-api-key"
    MODEL = "gemini-live-2.5-flash"  # actual model id
    BASE_URL = "{GATEWAY_BASE_URL}/live/{geminiProviderAccountName}"

    client = genai.Client(
        http_options={
            "base_url": BASE_URL,
            "headers": {
                "Authorization": f"Bearer {API_KEY}",
            }
        },
        api_key=API_KEY,
    )

    CONFIG = types.LiveConnectConfig(
        response_modalities=["AUDIO"],
        speech_config=types.SpeechConfig(
            voice_config=types.VoiceConfig(
                prebuilt_voice_config=types.PrebuiltVoiceConfig(voice_name="Zephyr")
            )
        ),
        # Enable transcription to get text versions of user and model speech.
        # Remove these lines if transcription is not needed.
        input_audio_transcription=types.AudioTranscriptionConfig(),
        output_audio_transcription=types.AudioTranscriptionConfig(),
    )

    pya = pyaudio.PyAudio()

    async def main():
        try:
            async with client.aio.live.connect(model=MODEL, config=CONFIG) as session:
                print("Connected!")

                # Record audio from microphone and send to session
                mic_info = pya.get_default_input_device_info()
                mic_stream = pya.open(
                    format=FORMAT, channels=CHANNELS, rate=SEND_SAMPLE_RATE,
                    input=True, input_device_index=mic_info["index"],
                    frames_per_buffer=CHUNK_SIZE,
                )

                # Speaker output for receiving audio
                speaker_stream = pya.open(
                    format=FORMAT, channels=CHANNELS, rate=RECEIVE_SAMPLE_RATE,
                    output=True,
                )

                audio_in_queue = asyncio.Queue()
                current_speaker = None  # Track who is currently speaking

                async def send_audio():
                    while True:
                        data = await asyncio.to_thread(
                            mic_stream.read, CHUNK_SIZE, exception_on_overflow=False
                        )
                        await session.send_realtime_input(audio={"data": data, "mime_type": "audio/pcm"})

                async def receive_audio():
                    nonlocal current_speaker
                    while True:
                        turn = session.receive()
                        was_interrupted = False
                        async for response in turn:
                            if response.server_content and response.server_content.model_turn:
                                for part in response.server_content.model_turn.parts:
                                    if part.inline_data:
                                        audio_in_queue.put_nowait(part.inline_data.data)
                                    if part.text and not part.thought:  # skip model thinking
                                        print(part.text, end="", flush=True)

                            # Print transcriptions if enabled above
                            if hasattr(response, "server_content") and response.server_content:
                                sc = response.server_content
                                if hasattr(sc, "input_transcription") and sc.input_transcription and sc.input_transcription.text:
                                    if current_speaker != "user":
                                        if current_speaker is not None:
                                            print()  # end previous line
                                        print("[You]: ", end="", flush=True)
                                        current_speaker = "user"
                                    print(sc.input_transcription.text, end="", flush=True)
                                if hasattr(sc, "output_transcription") and sc.output_transcription and sc.output_transcription.text:
                                    if current_speaker != "model":
                                        if current_speaker is not None:
                                            print()  # end previous line
                                        print("[Model]: ", end="", flush=True)
                                        current_speaker = "model"
                                    print(sc.output_transcription.text, end="", flush=True)
                                if hasattr(sc, "interrupted") and sc.interrupted:
                                    was_interrupted = True

                        # Only clear the audio queue on interruption.
                        # On normal turn completion, let play_audio finish playing
                        # all enqueued chunks to avoid losing audio.
                        if was_interrupted:
                            while not audio_in_queue.empty():
                                audio_in_queue.get_nowait()

                async def play_audio():
                    while True:
                        data = await audio_in_queue.get()
                        await asyncio.to_thread(speaker_stream.write, data)

                async with asyncio.TaskGroup() as tg:
                    tg.create_task(send_audio())
                    tg.create_task(receive_audio())
                    tg.create_task(play_audio())

        except Exception as e:
            print(f"Error: {e}")
        finally:
            pya.terminate()

    asyncio.run(main())
    ```
  </Accordion>

  <Accordion title="Google Vertex AI">
    ```python lines theme={"dark"}
    """
    Gemini Live API (Vertex AI) - Realtime Audio Streaming
    pip install google-genai pyaudio google-auth
    """
    import asyncio
    import pyaudio
    import google.auth.credentials
    from google import genai
    from google.genai import types

    FORMAT = pyaudio.paInt16
    CHANNELS = 1
    SEND_SAMPLE_RATE = 16000
    RECEIVE_SAMPLE_RATE = 24000
    CHUNK_SIZE = 1024

    API_KEY = "your-tfy-api-key"
    MODEL = "gemini-live-2.5-flash"  # actual model id
    BASE_URL = "{GATEWAY_BASE_URL}/live/{vertexProviderAccountName}"


    class _GatewayCredentials(google.auth.credentials.Credentials):
        """Bypasses local ADC; the gateway handles Vertex AI authentication."""

        def __init__(self, token):
            super().__init__()
            self.token = token

        def refresh(self, request):
            pass

        @property
        def valid(self):
            return True


    client = genai.Client(
        http_options={
            "base_url": BASE_URL,
            "headers": {"Authorization": f"Bearer {API_KEY}"},
        },
        vertexai=True,
        project="your-gcp-project",
        location="us-central1",
        credentials=_GatewayCredentials(API_KEY),
    )

    CONFIG = types.LiveConnectConfig(
        response_modalities=["AUDIO"],
        speech_config=types.SpeechConfig(
            voice_config=types.VoiceConfig(
                prebuilt_voice_config=types.PrebuiltVoiceConfig(voice_name="Zephyr")
            )
        ),
        # Enable transcription to get text versions of user and model speech.
        # Remove these lines if transcription is not needed.
        input_audio_transcription=types.AudioTranscriptionConfig(),
        output_audio_transcription=types.AudioTranscriptionConfig(),
    )

    pya = pyaudio.PyAudio()

    async def main():
        try:
            async with client.aio.live.connect(model=MODEL, config=CONFIG) as session:
                print("Connected!")

                # Record audio from microphone and send to session
                mic_info = pya.get_default_input_device_info()
                mic_stream = pya.open(
                    format=FORMAT, channels=CHANNELS, rate=SEND_SAMPLE_RATE,
                    input=True, input_device_index=mic_info["index"],
                    frames_per_buffer=CHUNK_SIZE,
                )

                # Speaker output for receiving audio
                speaker_stream = pya.open(
                    format=FORMAT, channels=CHANNELS, rate=RECEIVE_SAMPLE_RATE,
                    output=True,
                )

                audio_in_queue = asyncio.Queue()
                current_speaker = None  # Track who is currently speaking

                async def send_audio():
                    while True:
                        data = await asyncio.to_thread(
                            mic_stream.read, CHUNK_SIZE, exception_on_overflow=False
                        )
                        await session.send_realtime_input(audio={"data": data, "mime_type": "audio/pcm"})

                async def receive_audio():
                    nonlocal current_speaker
                    while True:
                        turn = session.receive()
                        was_interrupted = False
                        async for response in turn:
                            if response.server_content and response.server_content.model_turn:
                                for part in response.server_content.model_turn.parts:
                                    if part.inline_data:
                                        audio_in_queue.put_nowait(part.inline_data.data)
                                    if part.text and not part.thought:  # skip model thinking
                                        print(part.text, end="", flush=True)

                            # Print transcriptions if enabled above
                            if hasattr(response, "server_content") and response.server_content:
                                sc = response.server_content
                                if hasattr(sc, "input_transcription") and sc.input_transcription and sc.input_transcription.text:
                                    if current_speaker != "user":
                                        if current_speaker is not None:
                                            print()  # end previous line
                                        print("[You]: ", end="", flush=True)
                                        current_speaker = "user"
                                    print(sc.input_transcription.text, end="", flush=True)
                                if hasattr(sc, "output_transcription") and sc.output_transcription and sc.output_transcription.text:
                                    if current_speaker != "model":
                                        if current_speaker is not None:
                                            print()  # end previous line
                                        print("[Model]: ", end="", flush=True)
                                        current_speaker = "model"
                                    print(sc.output_transcription.text, end="", flush=True)
                                if hasattr(sc, "interrupted") and sc.interrupted:
                                    was_interrupted = True

                        # Only clear the audio queue on interruption.
                        # On normal turn completion, let play_audio finish playing
                        # all enqueued chunks to avoid losing audio.
                        if was_interrupted:
                            while not audio_in_queue.empty():
                                audio_in_queue.get_nowait()

                async def play_audio():
                    while True:
                        data = await audio_in_queue.get()
                        await asyncio.to_thread(speaker_stream.write, data)

                async with asyncio.TaskGroup() as tg:
                    tg.create_task(send_audio())
                    tg.create_task(receive_audio())
                    tg.create_task(play_audio())

        except Exception as e:
            print(f"Error: {e}")
        finally:
            pya.terminate()

    asyncio.run(main())
    ```
  </Accordion>

  <Accordion title="OpenAI">
    ```python lines theme={"dark"}
    """
    OpenAI Realtime API - Audio Streaming
    Ref: https://github.com/openai/openai-python/blob/main/examples/realtime/audio_util.py

    Requires Python 3.11+
    pip install "openai[realtime]" numpy sounddevice
    """
    import base64
    import asyncio
    import threading

    import numpy as np
    import sounddevice as sd

    from openai import AsyncOpenAI
    from openai.resources.realtime.realtime import AsyncRealtimeConnection

    SAMPLE_RATE = 24000
    CHANNELS = 1
    CHUNK_LENGTH_S = 0.05

    API_KEY = "your-tfy-api-key"
    MODEL = "gpt-4o-realtime-preview"  # actual model id

    client = AsyncOpenAI(
        api_key=API_KEY,
        websocket_base_url="wss://{GATEWAY_HOST}/live/{openaiProviderAccountName}",
    )


    class AudioPlayerAsync:
        def __init__(self):
            self.queue = []
            self.lock = threading.Lock()
            self.stream = sd.OutputStream(
                callback=self._callback, samplerate=SAMPLE_RATE,
                channels=CHANNELS, dtype=np.int16,
                blocksize=int(CHUNK_LENGTH_S * SAMPLE_RATE),
            )
            self.playing = False

        def _callback(self, outdata, frames, time, status):
            with self.lock:
                data = np.empty(0, dtype=np.int16)
                while len(data) < frames and self.queue:
                    item = self.queue.pop(0)
                    needed = frames - len(data)
                    data = np.concatenate((data, item[:needed]))
                    if len(item) > needed:
                        self.queue.insert(0, item[needed:])
                if len(data) < frames:
                    data = np.concatenate((data, np.zeros(frames - len(data), dtype=np.int16)))
            outdata[:] = data.reshape(-1, 1)

        def add_data(self, data: bytes):
            with self.lock:
                self.queue.append(np.frombuffer(data, dtype=np.int16))
                if not self.playing:
                    self.playing = True
                    self.stream.start()

        def stop(self):
            self.playing = False
            self.stream.stop()
            with self.lock:
                self.queue = []

        def terminate(self):
            self.stream.close()


    async def send_mic_audio(connection: AsyncRealtimeConnection):
        read_size = int(SAMPLE_RATE * 0.02)
        stream = sd.InputStream(channels=CHANNELS, samplerate=SAMPLE_RATE, dtype="int16")
        stream.start()
        try:
            while True:
                if stream.read_available < read_size:
                    await asyncio.sleep(0)
                    continue
                data, _ = stream.read(read_size)
                await connection.input_audio_buffer.append(
                    audio=base64.b64encode(data).decode("utf-8"),
                )
                await asyncio.sleep(0)
        except KeyboardInterrupt:
            pass
        finally:
            stream.stop()
            stream.close()


    async def main():
        player = AudioPlayerAsync()
        try:
            async with client.realtime.connect(model=MODEL) as connection:
                print("Connected!")

                await connection.session.update(session={
                    "type": "realtime",
                    "output_modalities": ["audio"],
                    "audio": {
                        "input": {
                            "turn_detection": {"type": "server_vad"},
                            # Enable input audio transcription (user speech to text).
                            # Remove this if input transcription is not needed.
                            "transcription": {"model": "gpt-4o-transcribe", "language": "en"},
                        },
                        "output": {
                            "voice": "alloy"
                        }
                    }
                })

                async def receive_events():
                    async for event in connection:
                        if event.type == "response.output_audio.delta":
                            player.add_data(base64.b64decode(event.delta))
                        # Output transcript (model speech to text), enabled by default
                        elif event.type == "response.output_audio_transcript.delta":
                            print(event.delta, end="", flush=True)
                        elif event.type == "response.output_audio_transcript.done":
                            print()
                        # Input transcript (user speech to text), requires transcription config above
                        elif event.type == "conversation.item.input_audio_transcription.completed":
                            print(f"\n[You]: {event.transcript}")
                        elif event.type == "input_audio_buffer.speech_started":
                            player.stop()
                        elif event.type == "error":
                            print(f"\n[ERROR] {event}")

                print("Start speaking! (Ctrl+C to stop)\n")
                async with asyncio.TaskGroup() as tg:
                    tg.create_task(send_mic_audio(connection))
                    tg.create_task(receive_events())

        except Exception as e:
            print(f"Error: {e}")
        finally:
            player.terminate()

    asyncio.run(main())
    ```
  </Accordion>

  <Accordion title="Azure AI Foundry / Azure OpenAI">
    ```python lines theme={"dark"}
    """
    OpenAI Realtime API via Azure AI Foundry / Azure OpenAI - Audio Streaming
    Ref: https://github.com/openai/openai-python/blob/main/examples/realtime/audio_util.py

    Requires Python 3.11+
    pip install "openai[realtime]" numpy sounddevice
    """
    import base64
    import asyncio
    import threading

    import numpy as np
    import sounddevice as sd

    from openai import AsyncOpenAI
    from openai.resources.realtime.realtime import AsyncRealtimeConnection

    SAMPLE_RATE = 24000
    CHANNELS = 1
    CHUNK_LENGTH_S = 0.05

    API_KEY = "your-tfy-api-key"
    MODEL = "gpt-4o-realtime-preview"  # actual model id

    client = AsyncOpenAI(
        api_key=API_KEY,
        websocket_base_url="wss://{GATEWAY_HOST}/live/{azureFoundryProviderAccountName}",
    )


    class AudioPlayerAsync:
        def __init__(self):
            self.queue = []
            self.lock = threading.Lock()
            self.stream = sd.OutputStream(
                callback=self._callback, samplerate=SAMPLE_RATE,
                channels=CHANNELS, dtype=np.int16,
                blocksize=int(CHUNK_LENGTH_S * SAMPLE_RATE),
            )
            self.playing = False

        def _callback(self, outdata, frames, time, status):
            with self.lock:
                data = np.empty(0, dtype=np.int16)
                while len(data) < frames and self.queue:
                    item = self.queue.pop(0)
                    needed = frames - len(data)
                    data = np.concatenate((data, item[:needed]))
                    if len(item) > needed:
                        self.queue.insert(0, item[needed:])
                if len(data) < frames:
                    data = np.concatenate((data, np.zeros(frames - len(data), dtype=np.int16)))
            outdata[:] = data.reshape(-1, 1)

        def add_data(self, data: bytes):
            with self.lock:
                self.queue.append(np.frombuffer(data, dtype=np.int16))
                if not self.playing:
                    self.playing = True
                    self.stream.start()

        def stop(self):
            self.playing = False
            self.stream.stop()
            with self.lock:
                self.queue = []

        def terminate(self):
            self.stream.close()


    async def send_mic_audio(connection: AsyncRealtimeConnection):
        read_size = int(SAMPLE_RATE * 0.02)
        stream = sd.InputStream(channels=CHANNELS, samplerate=SAMPLE_RATE, dtype="int16")
        stream.start()
        try:
            while True:
                if stream.read_available < read_size:
                    await asyncio.sleep(0)
                    continue
                data, _ = stream.read(read_size)
                await connection.input_audio_buffer.append(
                    audio=base64.b64encode(data).decode("utf-8"),
                )
                await asyncio.sleep(0)
        except KeyboardInterrupt:
            pass
        finally:
            stream.stop()
            stream.close()


    async def main():
        player = AudioPlayerAsync()
        try:
            async with client.realtime.connect(model=MODEL) as connection:
                print("Connected!")

                await connection.session.update(session={
                    "type": "realtime",
                    "output_modalities": ["audio"],
                    "audio": {
                        "input": {
                            "turn_detection": {"type": "server_vad"},
                            # Enable input audio transcription (user speech to text).
                            # Remove this if input transcription is not needed.
                            "transcription": {"model": "gpt-4o-transcribe", "language": "en"},
                        },
                        "output": {
                            "voice": "alloy"
                        }
                    }
                })

                async def receive_events():
                    async for event in connection:
                        if event.type == "response.output_audio.delta":
                            player.add_data(base64.b64decode(event.delta))
                        # Output transcript (model speech to text), enabled by default
                        elif event.type == "response.output_audio_transcript.delta":
                            print(event.delta, end="", flush=True)
                        elif event.type == "response.output_audio_transcript.done":
                            print()
                        # Input transcript (user speech to text), requires transcription config above
                        elif event.type == "conversation.item.input_audio_transcription.completed":
                            print(f"\n[You]: {event.transcript}")
                        elif event.type == "input_audio_buffer.speech_started":
                            player.stop()
                        elif event.type == "error":
                            print(f"\n[ERROR] {event}")

                print("Start speaking! (Ctrl+C to stop)\n")
                async with asyncio.TaskGroup() as tg:
                    tg.create_task(send_mic_audio(connection))
                    tg.create_task(receive_events())

        except Exception as e:
            print(f"Error: {e}")
        finally:
            player.terminate()

    asyncio.run(main())
    ```
  </Accordion>

  <Accordion title="Azure AI Foundry / Azure OpenAI (VoiceLive SDK)">
    ```python lines theme={"dark"}
    # pip install "azure-ai-voicelive[aiohttp]"

    import asyncio
    from azure.core.credentials import AccessToken
    from azure.ai.voicelive.aio import connect
    from azure.ai.voicelive.models import (
        RequestSession, Modality, InputAudioFormat, OutputAudioFormat,
        ServerVad, ServerEventType,
    )

    API_KEY = "your-tfy-api-key"
    MODEL = "gpt-4o-realtime-preview"  # actual model id
    ENDPOINT = "wss://{GATEWAY_HOST}/live/{azureFoundryProviderAccountName}"


    class BearerTokenCredential:
        """Sends token as Authorization: Bearer header instead of api-key header."""
        def __init__(self, token: str):
            self._token = token

        async def get_token(self, *scopes, **kwargs):
            return AccessToken(self._token, 0)

        async def close(self):
            pass

        async def __aenter__(self):
            return self

        async def __aexit__(self, *args):
            pass


    async def main():
        async with connect(
            endpoint=ENDPOINT,
            credential=BearerTokenCredential(API_KEY),
            model=MODEL,
        ) as conn:
            session = RequestSession(
                modalities=[Modality.TEXT, Modality.AUDIO],
                instructions="You are a helpful assistant.",
                input_audio_format=InputAudioFormat.PCM16,
                output_audio_format=OutputAudioFormat.PCM16,
                turn_detection=ServerVad(
                    threshold=0.5,
                    prefix_padding_ms=300,
                    silence_duration_ms=500,
                ),
            )
            await conn.session.update(session=session)

            async for evt in conn:
                print(f"Event: {evt.type}")
                if evt.type == ServerEventType.RESPONSE_DONE:
                    break

    asyncio.run(main())
    ```
  </Accordion>
</AccordionGroup>

## Tool calling

You can define tools (functions) that the model can invoke during a live session. The model will return a tool call when it decides to use a function, and you send the result back to continue the conversation. See the [Gemini Live API tools documentation](https://ai.google.dev/gemini-api/docs/live-api/tools) and [OpenAI Realtime API function calling documentation](https://platform.openai.com/docs/guides/realtime#function-calls) for more details.

<Note>
  You can adapt this pattern to call [TrueFoundry MCP tools](/ai-gateway/mcp/mcp-overview) as well — replace the dummy function with an MCP tool invocation via the [MCP Gateway SDK](/ai-gateway/mcp-gateway-sdk-usage).
</Note>

<AccordionGroup>
  <Accordion title="Google Gemini">
    ```python theme={"dark"}
    """
    Gemini Live API - Tool Calling with Audio
    Ref: https://ai.google.dev/gemini-api/docs/live-api/tools

    pip install google-genai pyaudio
    """
    import asyncio
    import pyaudio
    from google import genai
    from google.genai import types

    FORMAT = pyaudio.paInt16
    CHANNELS = 1
    SEND_SAMPLE_RATE = 16000
    RECEIVE_SAMPLE_RATE = 24000
    CHUNK_SIZE = 1024

    API_KEY = "your-tfy-api-key"
    MODEL = "gemini-live-2.5-flash"  # actual model id
    BASE_URL = "{GATEWAY_BASE_URL}/live/{geminiProviderAccountName}"

    client = genai.Client(
        http_options={
            "base_url": BASE_URL,
            "headers": {
                "Authorization": f"Bearer {API_KEY}",
            }
        },
        api_key=API_KEY,
    )


    # --- Dummy tool implementation (replace with your own logic) ---
    def get_weather(location: str) -> dict:
        """Returns mock weather data for the given location."""
        return {
            "location": location,
            "temperature": "15°C",
            "condition": "Foggy",
            "humidity": "85%",
        }


    TOOL_HANDLERS = {
        "get_weather": lambda args: get_weather(args["location"]),
    }

    # Define function declarations
    get_weather_declaration = {
        "name": "get_weather",
        "description": "Gets the current weather for a given location.",
        "parameters": {
            "type": "object",
            "properties": {
                "location": {
                    "type": "string",
                    "description": "The city or place to get weather for",
                }
            },
            "required": ["location"],
        },
    }

    tools = [{"function_declarations": [get_weather_declaration]}]

    CONFIG = types.LiveConnectConfig(
        response_modalities=["AUDIO"],
        speech_config=types.SpeechConfig(
            voice_config=types.VoiceConfig(
                prebuilt_voice_config=types.PrebuiltVoiceConfig(voice_name="Zephyr")
            )
        ),
        # Enable transcription to get text versions of user and model speech.
        # Remove these lines if transcription is not needed.
        input_audio_transcription=types.AudioTranscriptionConfig(),
        output_audio_transcription=types.AudioTranscriptionConfig(),
        tools=tools,
    )

    pya = pyaudio.PyAudio()

    async def main():
        try:
            async with client.aio.live.connect(model=MODEL, config=CONFIG) as session:
                # Try saying: "What's the weather in San Francisco?"
                print("Connected! Try saying: 'What's the weather in San Francisco?' to trigger tool calling.")

                # Record audio from microphone and send to session
                mic_info = pya.get_default_input_device_info()
                mic_stream = pya.open(
                    format=FORMAT, channels=CHANNELS, rate=SEND_SAMPLE_RATE,
                    input=True, input_device_index=mic_info["index"],
                    frames_per_buffer=CHUNK_SIZE,
                )

                # Speaker output for receiving audio
                speaker_stream = pya.open(
                    format=FORMAT, channels=CHANNELS, rate=RECEIVE_SAMPLE_RATE,
                    output=True,
                )

                audio_in_queue = asyncio.Queue()
                current_speaker = None  # Track who is currently speaking

                async def send_audio():
                    while True:
                        data = await asyncio.to_thread(
                            mic_stream.read, CHUNK_SIZE, exception_on_overflow=False
                        )
                        await session.send_realtime_input(audio={"data": data, "mime_type": "audio/pcm"})

                async def receive_audio():
                    nonlocal current_speaker
                    while True:
                        turn = session.receive()
                        was_interrupted = False
                        async for response in turn:
                            # Handle tool calls
                            if response.tool_call:
                                function_responses = []
                                for fc in response.tool_call.function_calls:
                                    print(f"\n[Tool Call] {fc.name}({fc.args})")

                                    handler = TOOL_HANDLERS.get(fc.name)
                                    if handler:
                                        result = handler(fc.args)
                                    else:
                                        result = {"error": f"Unknown tool: {fc.name}"}
                                    print(f"[Tool Result] {result}")

                                    function_responses.append(
                                        types.FunctionResponse(
                                            id=fc.id,
                                            name=fc.name,
                                            response={"result": result},
                                        )
                                    )

                                await session.send_tool_response(
                                    function_responses=function_responses
                                )
                                continue

                            if response.server_content and response.server_content.model_turn:
                                for part in response.server_content.model_turn.parts:
                                    if part.inline_data:
                                        audio_in_queue.put_nowait(part.inline_data.data)
                                    if part.text and not part.thought:
                                        print(part.text, end="", flush=True)

                            # Print transcriptions if enabled above
                            if hasattr(response, "server_content") and response.server_content:
                                sc = response.server_content
                                if hasattr(sc, "input_transcription") and sc.input_transcription and sc.input_transcription.text:
                                    if current_speaker != "user":
                                        if current_speaker is not None:
                                            print()  # end previous line
                                        print("[You]: ", end="", flush=True)
                                        current_speaker = "user"
                                    print(sc.input_transcription.text, end="", flush=True)
                                if hasattr(sc, "output_transcription") and sc.output_transcription and sc.output_transcription.text:
                                    if current_speaker != "model":
                                        if current_speaker is not None:
                                            print()  # end previous line
                                        print("[Model]: ", end="", flush=True)
                                        current_speaker = "model"
                                    print(sc.output_transcription.text, end="", flush=True)
                                if hasattr(sc, "interrupted") and sc.interrupted:
                                    was_interrupted = True

                        # Only clear the audio queue on interruption.
                        # On normal turn completion, let play_audio finish playing
                        # all enqueued chunks to avoid losing audio.
                        if was_interrupted:
                            while not audio_in_queue.empty():
                                audio_in_queue.get_nowait()

                async def play_audio():
                    while True:
                        data = await audio_in_queue.get()
                        await asyncio.to_thread(speaker_stream.write, data)

                print("Start speaking! (Ctrl+C to stop)\n")
                async with asyncio.TaskGroup() as tg:
                    tg.create_task(send_audio())
                    tg.create_task(receive_audio())
                    tg.create_task(play_audio())

        except Exception as e:
            print(f"Error: {e}")
        finally:
            pya.terminate()

    asyncio.run(main())
    ```
  </Accordion>

  <Accordion title="Google Vertex AI">
    ```python theme={"dark"}
    """
    Gemini Live API (Vertex AI) - Tool Calling with Audio
    Ref: https://ai.google.dev/gemini-api/docs/live-api/tools

    pip install google-genai pyaudio google-auth
    """
    import asyncio
    import pyaudio
    import google.auth.credentials
    from google import genai
    from google.genai import types

    FORMAT = pyaudio.paInt16
    CHANNELS = 1
    SEND_SAMPLE_RATE = 16000
    RECEIVE_SAMPLE_RATE = 24000
    CHUNK_SIZE = 1024

    API_KEY = "your-tfy-api-key"
    MODEL = "gemini-live-2.5-flash"  # actual model id
    BASE_URL = "{GATEWAY_BASE_URL}/live/{vertexProviderAccountName}"


    class _GatewayCredentials(google.auth.credentials.Credentials):
        """Bypasses local ADC; the gateway handles Vertex AI authentication."""
        def __init__(self, token):
            super().__init__()
            self.token = token
        def refresh(self, request):
            pass
        @property
        def valid(self):
            return True


    client = genai.Client(
        http_options={
            "base_url": BASE_URL,
            "headers": {"Authorization": f"Bearer {API_KEY}"},
        },
        vertexai=True,
        project="your-gcp-project",
        location="us-central1",
        credentials=_GatewayCredentials(API_KEY),
    )


    # --- Dummy tool implementation (replace with your own logic) ---
    def get_weather(location: str) -> dict:
        """Returns mock weather data for the given location."""
        return {
            "location": location,
            "temperature": "15°C",
            "condition": "Foggy",
            "humidity": "85%",
        }


    TOOL_HANDLERS = {
        "get_weather": lambda args: get_weather(args["location"]),
    }

    # Define function declarations
    get_weather_declaration = {
        "name": "get_weather",
        "description": "Gets the current weather for a given location.",
        "parameters": {
            "type": "object",
            "properties": {
                "location": {
                    "type": "string",
                    "description": "The city or place to get weather for",
                }
            },
            "required": ["location"],
        },
    }

    tools = [{"function_declarations": [get_weather_declaration]}]

    CONFIG = types.LiveConnectConfig(
        response_modalities=["AUDIO"],
        speech_config=types.SpeechConfig(
            voice_config=types.VoiceConfig(
                prebuilt_voice_config=types.PrebuiltVoiceConfig(voice_name="Zephyr")
            )
        ),
        # Enable transcription to get text versions of user and model speech.
        # Remove these lines if transcription is not needed.
        input_audio_transcription=types.AudioTranscriptionConfig(),
        output_audio_transcription=types.AudioTranscriptionConfig(),
        tools=tools,
    )

    pya = pyaudio.PyAudio()

    async def main():
        try:
            async with client.aio.live.connect(model=MODEL, config=CONFIG) as session:
                # Try saying: "What's the weather in San Francisco?"
                print("Connected! Try saying: 'What's the weather in San Francisco?' to trigger tool calling.")

                # Record audio from microphone and send to session
                mic_info = pya.get_default_input_device_info()
                mic_stream = pya.open(
                    format=FORMAT, channels=CHANNELS, rate=SEND_SAMPLE_RATE,
                    input=True, input_device_index=mic_info["index"],
                    frames_per_buffer=CHUNK_SIZE,
                )

                # Speaker output for receiving audio
                speaker_stream = pya.open(
                    format=FORMAT, channels=CHANNELS, rate=RECEIVE_SAMPLE_RATE,
                    output=True,
                )

                audio_in_queue = asyncio.Queue()
                current_speaker = None  # Track who is currently speaking

                async def send_audio():
                    while True:
                        data = await asyncio.to_thread(
                            mic_stream.read, CHUNK_SIZE, exception_on_overflow=False
                        )
                        await session.send_realtime_input(audio={"data": data, "mime_type": "audio/pcm"})

                async def receive_audio():
                    nonlocal current_speaker
                    while True:
                        turn = session.receive()
                        was_interrupted = False
                        async for response in turn:
                            # Handle tool calls
                            if response.tool_call:
                                function_responses = []
                                for fc in response.tool_call.function_calls:
                                    print(f"\n[Tool Call] {fc.name}({fc.args})")

                                    handler = TOOL_HANDLERS.get(fc.name)
                                    if handler:
                                        result = handler(fc.args)
                                    else:
                                        result = {"error": f"Unknown tool: {fc.name}"}
                                    print(f"[Tool Result] {result}")

                                    function_responses.append(
                                        types.FunctionResponse(
                                            id=fc.id,
                                            name=fc.name,
                                            response={"result": result},
                                        )
                                    )

                                await session.send_tool_response(
                                    function_responses=function_responses
                                )
                                continue

                            if response.server_content and response.server_content.model_turn:
                                for part in response.server_content.model_turn.parts:
                                    if part.inline_data:
                                        audio_in_queue.put_nowait(part.inline_data.data)
                                    if part.text and not part.thought:
                                        print(part.text, end="", flush=True)

                            # Print transcriptions if enabled above
                            if hasattr(response, "server_content") and response.server_content:
                                sc = response.server_content
                                if hasattr(sc, "input_transcription") and sc.input_transcription and sc.input_transcription.text:
                                    if current_speaker != "user":
                                        if current_speaker is not None:
                                            print()  # end previous line
                                        print("[You]: ", end="", flush=True)
                                        current_speaker = "user"
                                    print(sc.input_transcription.text, end="", flush=True)
                                if hasattr(sc, "output_transcription") and sc.output_transcription and sc.output_transcription.text:
                                    if current_speaker != "model":
                                        if current_speaker is not None:
                                            print()  # end previous line
                                        print("[Model]: ", end="", flush=True)
                                        current_speaker = "model"
                                    print(sc.output_transcription.text, end="", flush=True)
                                if hasattr(sc, "interrupted") and sc.interrupted:
                                    was_interrupted = True

                        # Only clear the audio queue on interruption.
                        # On normal turn completion, let play_audio finish playing
                        # all enqueued chunks to avoid losing audio.
                        if was_interrupted:
                            while not audio_in_queue.empty():
                                audio_in_queue.get_nowait()

                async def play_audio():
                    while True:
                        data = await audio_in_queue.get()
                        await asyncio.to_thread(speaker_stream.write, data)

                print("Start speaking! (Ctrl+C to stop)\n")
                async with asyncio.TaskGroup() as tg:
                    tg.create_task(send_audio())
                    tg.create_task(receive_audio())
                    tg.create_task(play_audio())

        except Exception as e:
            print(f"Error: {e}")
        finally:
            pya.terminate()

    asyncio.run(main())
    ```
  </Accordion>

  <Accordion title="OpenAI">
    ```python theme={"dark"}
    """
    OpenAI Realtime API - Tool Calling with Audio
    Ref: https://platform.openai.com/docs/guides/realtime#function-calls

    Requires Python 3.11+
    pip install "openai[realtime]" numpy sounddevice
    """
    import json
    import base64
    import asyncio
    import threading

    import numpy as np
    import sounddevice as sd

    from openai import AsyncOpenAI
    from openai.resources.realtime.realtime import AsyncRealtimeConnection

    SAMPLE_RATE = 24000
    CHANNELS = 1
    CHUNK_LENGTH_S = 0.05

    API_KEY = "your-tfy-api-key"
    MODEL = "gpt-4o-realtime-preview"  # actual model id

    client = AsyncOpenAI(
        api_key=API_KEY,
        websocket_base_url="wss://{GATEWAY_HOST}/live/{openaiProviderAccountName}",
    )


    # --- Dummy tool implementation (replace with your own logic) ---
    def get_weather(location: str) -> dict:
        """Returns mock weather data for the given location."""
        return {
            "location": location,
            "temperature": "15°C",
            "condition": "Foggy",
            "humidity": "85%",
        }


    TOOL_HANDLERS = {
        "get_weather": lambda args: get_weather(args["location"]),
    }


    class AudioPlayerAsync:
        def __init__(self):
            self.queue = []
            self.lock = threading.Lock()
            self.stream = sd.OutputStream(
                callback=self._callback, samplerate=SAMPLE_RATE,
                channels=CHANNELS, dtype=np.int16,
                blocksize=int(CHUNK_LENGTH_S * SAMPLE_RATE),
            )
            self.playing = False

        def _callback(self, outdata, frames, time, status):
            with self.lock:
                data = np.empty(0, dtype=np.int16)
                while len(data) < frames and self.queue:
                    item = self.queue.pop(0)
                    needed = frames - len(data)
                    data = np.concatenate((data, item[:needed]))
                    if len(item) > needed:
                        self.queue.insert(0, item[needed:])
                if len(data) < frames:
                    data = np.concatenate((data, np.zeros(frames - len(data), dtype=np.int16)))
            outdata[:] = data.reshape(-1, 1)

        def add_data(self, data: bytes):
            with self.lock:
                self.queue.append(np.frombuffer(data, dtype=np.int16))
                if not self.playing:
                    self.playing = True
                    self.stream.start()

        def stop(self):
            self.playing = False
            self.stream.stop()
            with self.lock:
                self.queue = []

        def terminate(self):
            self.stream.close()


    async def send_mic_audio(connection: AsyncRealtimeConnection):
        read_size = int(SAMPLE_RATE * 0.02)
        stream = sd.InputStream(channels=CHANNELS, samplerate=SAMPLE_RATE, dtype="int16")
        stream.start()
        try:
            while True:
                if stream.read_available < read_size:
                    await asyncio.sleep(0)
                    continue
                data, _ = stream.read(read_size)
                await connection.input_audio_buffer.append(
                    audio=base64.b64encode(data).decode("utf-8"),
                )
                await asyncio.sleep(0)
        except KeyboardInterrupt:
            pass
        finally:
            stream.stop()
            stream.close()


    async def main():
        player = AudioPlayerAsync()
        try:
            async with client.realtime.connect(model=MODEL) as connection:
                # Try saying: "What's the weather in San Francisco?"
                print("Connected! Try saying: 'What's the weather in San Francisco?' to trigger tool calling.")

                await connection.session.update(session={
                    "type": "realtime",
                    "output_modalities": ["audio"],
                    "tools": [
                        {
                            "type": "function",
                            "name": "get_weather",
                            "description": "Gets the current weather for a given location.",
                            "parameters": {
                                "type": "object",
                                "properties": {
                                    "location": {
                                        "type": "string",
                                        "description": "The city or place to get weather for",
                                    }
                                },
                                "required": ["location"],
                            },
                        }
                    ],
                    "audio": {
                        "input": {
                            "turn_detection": {"type": "server_vad"},
                            # Enable input audio transcription (user speech to text).
                            # Remove this if input transcription is not needed.
                            "transcription": {"model": "gpt-4o-transcribe"},
                        },
                        "output": {
                            "voice": "alloy"
                        }
                    }
                })

                pending_tool_calls: dict[str, dict] = {}

                async def receive_events():
                    async for event in connection:
                        if event.type == "response.output_audio.delta":
                            player.add_data(base64.b64decode(event.delta))
                        # Output transcript (model speech to text), enabled by default
                        elif event.type == "response.output_audio_transcript.delta":
                            print(event.delta, end="", flush=True)
                        elif event.type == "response.output_audio_transcript.done":
                            print()
                        # Input transcript (user speech to text), requires transcription config above
                        elif event.type == "conversation.item.input_audio_transcription.completed":
                            print(f"\n[You]: {event.transcript}")
                        elif event.type == "input_audio_buffer.speech_started":
                            player.stop()
                        # Tool call handling
                        elif event.type == "response.output_item.added":
                            item = event.item
                            if item.type == "function_call":
                                pending_tool_calls[item.call_id] = {"name": item.name, "arguments": ""}
                        elif event.type == "response.function_call_arguments.delta":
                            if event.call_id in pending_tool_calls:
                                pending_tool_calls[event.call_id]["arguments"] += event.delta
                        elif event.type == "response.function_call_arguments.done":
                            call_id = event.call_id
                            if call_id in pending_tool_calls:
                                tool = pending_tool_calls.pop(call_id)
                                args = json.loads(tool["arguments"])
                                print(f"\n[Tool Call] {tool['name']}({args})")

                                handler = TOOL_HANDLERS.get(tool["name"])
                                if handler:
                                    result = handler(args)
                                else:
                                    result = {"error": f"Unknown tool: {tool['name']}"}
                                print(f"[Tool Result] {result}")

                                await connection.conversation.item.create(item={
                                    "type": "function_call_output",
                                    "call_id": call_id,
                                    "output": json.dumps(result),
                                })
                                await connection.response.create()
                        elif event.type == "error":
                            print(f"\n[ERROR] {event}")

                print("Start speaking! (Ctrl+C to stop)\n")
                async with asyncio.TaskGroup() as tg:
                    tg.create_task(send_mic_audio(connection))
                    tg.create_task(receive_events())

        except Exception as e:
            print(f"Error: {e}")
        finally:
            player.terminate()

    asyncio.run(main())
    ```
  </Accordion>

  <Accordion title="Azure AI Foundry">
    ```python theme={"dark"}
    """
    OpenAI Realtime API via Azure AI Foundry - Tool Calling with Audio
    Ref: https://platform.openai.com/docs/guides/realtime#function-calls

    Requires Python 3.11+
    pip install "openai[realtime]" numpy sounddevice
    """
    import json
    import base64
    import asyncio
    import threading

    import numpy as np
    import sounddevice as sd

    from openai import AsyncOpenAI
    from openai.resources.realtime.realtime import AsyncRealtimeConnection

    SAMPLE_RATE = 24000
    CHANNELS = 1
    CHUNK_LENGTH_S = 0.05

    API_KEY = "your-tfy-api-key"
    MODEL = "gpt-4o-realtime-preview"  # actual model id

    client = AsyncOpenAI(
        api_key=API_KEY,
        websocket_base_url="wss://{GATEWAY_HOST}/live/{azureFoundryProviderAccountName}",
    )


    # --- Dummy tool implementation (replace with your own logic) ---
    def get_weather(location: str) -> dict:
        """Returns mock weather data for the given location."""
        return {
            "location": location,
            "temperature": "15°C",
            "condition": "Foggy",
            "humidity": "85%",
        }


    TOOL_HANDLERS = {
        "get_weather": lambda args: get_weather(args["location"]),
    }


    class AudioPlayerAsync:
        def __init__(self):
            self.queue = []
            self.lock = threading.Lock()
            self.stream = sd.OutputStream(
                callback=self._callback, samplerate=SAMPLE_RATE,
                channels=CHANNELS, dtype=np.int16,
                blocksize=int(CHUNK_LENGTH_S * SAMPLE_RATE),
            )
            self.playing = False

        def _callback(self, outdata, frames, time, status):
            with self.lock:
                data = np.empty(0, dtype=np.int16)
                while len(data) < frames and self.queue:
                    item = self.queue.pop(0)
                    needed = frames - len(data)
                    data = np.concatenate((data, item[:needed]))
                    if len(item) > needed:
                        self.queue.insert(0, item[needed:])
                if len(data) < frames:
                    data = np.concatenate((data, np.zeros(frames - len(data), dtype=np.int16)))
            outdata[:] = data.reshape(-1, 1)

        def add_data(self, data: bytes):
            with self.lock:
                self.queue.append(np.frombuffer(data, dtype=np.int16))
                if not self.playing:
                    self.playing = True
                    self.stream.start()

        def stop(self):
            self.playing = False
            self.stream.stop()
            with self.lock:
                self.queue = []

        def terminate(self):
            self.stream.close()


    async def send_mic_audio(connection: AsyncRealtimeConnection):
        read_size = int(SAMPLE_RATE * 0.02)
        stream = sd.InputStream(channels=CHANNELS, samplerate=SAMPLE_RATE, dtype="int16")
        stream.start()
        try:
            while True:
                if stream.read_available < read_size:
                    await asyncio.sleep(0)
                    continue
                data, _ = stream.read(read_size)
                await connection.input_audio_buffer.append(
                    audio=base64.b64encode(data).decode("utf-8"),
                )
                await asyncio.sleep(0)
        except KeyboardInterrupt:
            pass
        finally:
            stream.stop()
            stream.close()


    async def main():
        player = AudioPlayerAsync()
        try:
            async with client.realtime.connect(model=MODEL) as connection:
                # Try saying: "What's the weather in San Francisco?"
                print("Connected! Try saying: 'What's the weather in San Francisco?' to trigger tool calling.")

                await connection.session.update(session={
                    "type": "realtime",
                    "output_modalities": ["audio"],
                    "tools": [
                        {
                            "type": "function",
                            "name": "get_weather",
                            "description": "Gets the current weather for a given location.",
                            "parameters": {
                                "type": "object",
                                "properties": {
                                    "location": {
                                        "type": "string",
                                        "description": "The city or place to get weather for",
                                    }
                                },
                                "required": ["location"],
                            },
                        }
                    ],
                    "audio": {
                        "input": {
                            "turn_detection": {"type": "server_vad"},
                            # Enable input audio transcription (user speech to text).
                            # Remove this if input transcription is not needed.
                            "transcription": {"model": "gpt-4o-transcribe"},
                        },
                        "output": {
                            "voice": "alloy"
                        }
                    }
                })

                pending_tool_calls: dict[str, dict] = {}

                async def receive_events():
                    async for event in connection:
                        if event.type == "response.output_audio.delta":
                            player.add_data(base64.b64decode(event.delta))
                        # Output transcript (model speech to text), enabled by default
                        elif event.type == "response.output_audio_transcript.delta":
                            print(event.delta, end="", flush=True)
                        elif event.type == "response.output_audio_transcript.done":
                            print()
                        # Input transcript (user speech to text), requires transcription config above
                        elif event.type == "conversation.item.input_audio_transcription.completed":
                            print(f"\n[You]: {event.transcript}")
                        elif event.type == "input_audio_buffer.speech_started":
                            player.stop()
                        # Tool call handling
                        elif event.type == "response.output_item.added":
                            item = event.item
                            if item.type == "function_call":
                                pending_tool_calls[item.call_id] = {"name": item.name, "arguments": ""}
                        elif event.type == "response.function_call_arguments.delta":
                            if event.call_id in pending_tool_calls:
                                pending_tool_calls[event.call_id]["arguments"] += event.delta
                        elif event.type == "response.function_call_arguments.done":
                            call_id = event.call_id
                            if call_id in pending_tool_calls:
                                tool = pending_tool_calls.pop(call_id)
                                args = json.loads(tool["arguments"])
                                print(f"\n[Tool Call] {tool['name']}({args})")

                                handler = TOOL_HANDLERS.get(tool["name"])
                                if handler:
                                    result = handler(args)
                                else:
                                    result = {"error": f"Unknown tool: {tool['name']}"}
                                print(f"[Tool Result] {result}")

                                await connection.conversation.item.create(item={
                                    "type": "function_call_output",
                                    "call_id": call_id,
                                    "output": json.dumps(result),
                                })
                                await connection.response.create()
                        elif event.type == "error":
                            print(f"\n[ERROR] {event}")

                print("Start speaking! (Ctrl+C to stop)\n")
                async with asyncio.TaskGroup() as tg:
                    tg.create_task(send_mic_audio(connection))
                    tg.create_task(receive_events())

        except Exception as e:
            print(f"Error: {e}")
        finally:
            player.terminate()

    asyncio.run(main())
    ```
  </Accordion>
</AccordionGroup>

## References

* [Google Gemini Live API documentation](https://ai.google.dev/gemini-api/docs/live)
* [Google Vertex AI Live API documentation](https://cloud.google.com/vertex-ai/generative-ai/docs/live-api)
* [OpenAI Realtime API documentation](https://platform.openai.com/docs/guides/realtime)
* [Azure Voice Live API documentation](https://learn.microsoft.com/en-us/azure/ai-services/speech-service/voice-live)
* [Azure AI VoiceLive Python SDK](https://learn.microsoft.com/en-us/python/api/overview/azure/ai-voicelive-readme)
