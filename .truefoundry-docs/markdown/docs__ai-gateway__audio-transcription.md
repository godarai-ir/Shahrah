> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Speech to Text API

> Convert audio files to text using speech-to-text models through TrueFoundry's AI Gateway

## Provider capabilities

The table below summarizes gateway support for this endpoint by provider.

<Info>
  Legend:

  * **✅** Supported by Provider and Truefoundry
  * <Icon icon="circle-xmark" iconType="regular" color="red" /> Provided by provider, but not by Truefoundry
  * <Icon icon="circle-minus" iconType="regular" /> Provider does not support this feature
</Info>

| Provider         | Transcription                                               |
| ---------------- | ----------------------------------------------------------- |
| OpenAI           | ✅                                                           |
| Azure OpenAI     | ✅                                                           |
| Azure AI Foundry | ✅                                                           |
| Anthropic        | <Icon icon="circle-minus" iconType="regular" />             |
| Bedrock          | <Icon icon="circle-minus" iconType="regular" />             |
| Vertex           | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| Cohere           | <Icon icon="circle-minus" iconType="regular" />             |
| Gemini           | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| Groq             | ✅                                                           |
| Together-AI      | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| xAI              | <Icon icon="circle-minus" iconType="regular" />             |
| DeepInfra        | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| DeepGram         | ✅                                                           |
| Cartesia         | ✅                                                           |
| ElevenLabs       | ✅                                                           |

For every gateway endpoint and provider, see [Supported APIs](/docs/ai-gateway/intro-to-llm-gateway#supported-apis).

**Speech to text (STT)** turns speech (audio files) into text. The gateway supports two ways to call it:

| Approach                        | Use for                                      | Base path                                      |
| ------------------------------- | -------------------------------------------- | ---------------------------------------------- |
| **OpenAI-compatible API**       | OpenAI, Azure OpenAI, Azure AI Foundry, Groq | `{GATEWAY_BASE_URL}`                           |
| **Provider proxy** (native SDK) | Deepgram, Cartesia, ElevenLabs, Vertex       | `{GATEWAY_BASE_URL}/stt/{providerAccountName}` |

## Add models to the gateway

Before you can call the Speech to Text API, add your STT models to TrueFoundry through a provider account. When adding a model, select **Audio Transcription** as the model type.

| Provider         | Setup guide                                           |
| ---------------- | ----------------------------------------------------- |
| OpenAI           | [OpenAI](/docs/ai-gateway/openai)                     |
| Azure OpenAI     | [Azure OpenAI](/docs/ai-gateway/azure-openai)         |
| Azure AI Foundry | [Azure AI Foundry](/docs/ai-gateway/azure-ai-foundry) |
| Groq             | [Groq](/docs/ai-gateway/groq)                         |
| Deepgram         | [Deepgram](/docs/ai-gateway/deepgram)                 |
| Cartesia         | [Cartesia](/docs/ai-gateway/cartesia)                 |
| ElevenLabs       | [ElevenLabs](/docs/ai-gateway/elevenlabs)             |
| Vertex           | [Google Vertex](/docs/ai-gateway/google-vertex)       |

## Code snippets

**Before you start:** Replace `{GATEWAY_BASE_URL}` with your AI Gateway Base URL ([how to find it](/docs/ai-gateway/quick-start#gateway-base-url)) and `your-tfy-api-key` with your TrueFoundry API key. For the provider proxy, replace `{providerAccountName}` with the display name of your provider account on TrueFoundry.

<Note>
  **Model names:** For audio (STT/TTS), the model ID in code must match the **display name** of the model on your TrueFoundry provider account.
</Note>

<Note>
  **Which SDK to use:** For OpenAI, Azure OpenAI, Azure AI Foundry, and Groq, use the OpenAI SDK (same API). For Deepgram, Cartesia, ElevenLabs, and Vertex, use each provider’s native SDK with the gateway URL above.
</Note>

<CodeGroup>
  ```python OpenAI lines theme={"dark"}
  from openai import OpenAI

  BASE_URL = "{GATEWAY_BASE_URL}"
  API_KEY = "your-tfy-api-key"

  client = OpenAI(
      api_key=API_KEY,
      base_url=BASE_URL,
  )

  with open("/path/to/audio.mp3", "rb") as audio_file:
      response = client.audio.transcriptions.create(
          model="openai-main/whisper-1",  # truefoundry model name
          file=audio_file,
      )

  print(response)
  ```

  ```python Azure OpenAI lines theme={"dark"}
  from openai import OpenAI

  BASE_URL = "{GATEWAY_BASE_URL}"
  API_KEY = "your-tfy-api-key"

  client = OpenAI(
      api_key=API_KEY,
      base_url=BASE_URL,
  )

  with open("/path/to/audio.mp3", "rb") as audio_file:
      response = client.audio.transcriptions.create(
          model="azure-openai-main/your-whisper-deployment", # truefoundry model name
          file=audio_file,
      )

  print(response)
  ```

  ```python Azure AI Foundry lines theme={"dark"}
  from openai import OpenAI

  BASE_URL = "{GATEWAY_BASE_URL}"
  API_KEY = "your-tfy-api-key"

  client = OpenAI(
      api_key=API_KEY,
      base_url=BASE_URL,
  )

  with open("/path/to/audio.mp3", "rb") as audio_file:
      response = client.audio.transcriptions.create(
          model="azure-foundry-main/your-whisper-deployment", # truefoundry model name
          file=audio_file,
      )

  print(response)
  ```

  ```python Groq lines theme={"dark"}
  from openai import OpenAI

  BASE_URL = "{GATEWAY_BASE_URL}"
  API_KEY = "your-tfy-api-key"

  client = OpenAI(
      api_key=API_KEY,
      base_url=BASE_URL,
  )

  with open("/path/to/audio.mp3", "rb") as audio_file:
      response = client.audio.transcriptions.create(
          model="groq-main/whisper-large-v3", # truefoundry model name
          file=audio_file,
      )

  print(response)
  ```

  ```python Deepgram lines theme={"dark"}
  from deepgram import DeepgramClient
  from deepgram.environment import DeepgramClientEnvironment

  TFY_API_KEY = "your-tfy-api-key"
  AUDIO_FILE = "/path/to/audio.mp3"

  env = DeepgramClientEnvironment(
      base="{GATEWAY_BASE_URL}/stt/{deepgramProviderAccountName}",
      production="",
      agent="",
  )
  client = DeepgramClient(
  	api_key="dummy",
  	environment=env,
  	headers={"x-tfy-api-key": TFY_API_KEY},
  )

  # Transcribe a remote URL
  response = client.listen.v1.media.transcribe_url(
      url="https://dpgr.am/spacewalk.wav",
      model="nova-3",
      smart_format=True
  )
  print(response)

  # Or transcribe a local file
  with open(AUDIO_FILE, "rb") as audio_file:
      response = client.listen.v1.media.transcribe_file(
          request=audio_file.read(),
          model="nova-3", # actual model id
          smart_format=True
      )
  print(response)
  ```

  ```python Cartesia lines theme={"dark"}
  import asyncio
  import httpx
  from cartesia import AsyncCartesia

  TFY_API_KEY = "your-tfy-api-key"
  BASE_URL = "{GATEWAY_BASE_URL}/stt/{cartesiaProviderAccountName}"
  INPUT_FILE = "/path/to/audio.wav"

  client = AsyncCartesia(
      api_key="dummy",
      base_url=BASE_URL,
      http_client=httpx.AsyncClient(
          headers={
              "x-tfy-api-key": TFY_API_KEY,
          },
      ),
  )


  async def main():
      with open(INPUT_FILE, "rb") as f:
          response = await client.stt.transcribe(
              file=f,
              model="ink-whisper", # actual model id
              language="en",
              timestamp_granularities=["word"],
          )
      print(response)


  asyncio.run(main())
  ```

  ```python ElevenLabs lines theme={"dark"}
  import httpx
  from elevenlabs import ElevenLabs

  BASE_URL = "{GATEWAY_BASE_URL}/stt/{elevenLabsProviderAccountName}"
  TFY_API_KEY = "your-tfy-api-key"
  AUDIO_FILE_PATH = "/path/to/audio.mp3"

  client = ElevenLabs(
      api_key="dummy",
      base_url=BASE_URL,
      httpx_client=httpx.Client(
          headers={
              "x-tfy-api-key": TFY_API_KEY,
          },
      ),
  )

  with open(AUDIO_FILE_PATH, "rb") as audio_file:
      response = client.speech_to_text.convert(
          file=audio_file,
          model_id="scribe_v2", # actual model id
          tag_audio_events=True,
          timestamps_granularity="word",
      )

  print(response)
  ```

  ```python Google Vertex lines theme={"dark"}
  # Google Cloud Speech-to-Text V2: https://cloud.google.com/speech-to-text/v2/docs
  # pip install google-cloud-speech
  from google.cloud.speech_v2 import SpeechClient
  from google.cloud.speech_v2.types import cloud_speech
  from google.api_core.client_options import ClientOptions
  from google.auth.credentials import AnonymousCredentials

  TFY_API_KEY = "your-tfy-api-key"
  BASE_URL = "{GATEWAY_BASE_URL}/stt/{vertexProviderAccountName}"

  client = SpeechClient(
      credentials=AnonymousCredentials(),
      transport="rest",
      client_options=ClientOptions(api_endpoint=BASE_URL),
  )

  # Inject TFY auth header into the REST transport session
  client._transport._session.headers["x-tfy-api-key"] = TFY_API_KEY

  with open("/path/to/audio.mp3", "rb") as f:
      audio_content = f.read()

  config = cloud_speech.RecognitionConfig(
      auto_decoding_config=cloud_speech.AutoDetectDecodingConfig(),
      language_codes=["en-US"],
      model="long",  # actual model id (long, short, telephony, etc.)
      features=cloud_speech.RecognitionFeatures(
          enable_automatic_punctuation=True,
      ),
  )

  response = client.recognize(
      recognizer="projects/{GCP_PROJECT_ID}/locations/global/recognizers/_",
      config=config,
      content=audio_content,
  )

  for result in response.results:
      alt = result.alternatives[0]
      print(f"Transcript: {alt.transcript}")
      print(f"Confidence: {alt.confidence}")
  ```
</CodeGroup>

## Response

The shape of the response depends on the provider. Use `print(response)` to inspect it, or refer to each provider’s SDK docs for the exact structure.
