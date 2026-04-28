> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Text to Speech API

> Convert text to speech through TrueFoundry's AI Gateway

## Provider capabilities

The table below summarizes gateway support for this endpoint by provider.

<Info>
  Legend:

  * **✅** Supported by Provider and Truefoundry
  * <Icon icon="circle-xmark" iconType="regular" color="red" /> Provided by provider, but not by Truefoundry
  * <Icon icon="circle-minus" iconType="regular" /> Provider does not support this feature
</Info>

| Provider         | Text To Speech                                              |
| ---------------- | ----------------------------------------------------------- |
| OpenAI           | ✅                                                           |
| Azure OpenAI     | ✅                                                           |
| Azure AI Foundry | ✅                                                           |
| Anthropic        | <Icon icon="circle-minus" iconType="regular" />             |
| Bedrock          | <Icon icon="circle-minus" iconType="regular" />             |
| Vertex           | ✅                                                           |
| Cohere           | <Icon icon="circle-minus" iconType="regular" />             |
| Gemini           | ✅                                                           |
| Groq             | ✅                                                           |
| Together-AI      | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| xAI              | <Icon icon="circle-minus" iconType="regular" />             |
| DeepInfra        | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| DeepGram         | ✅                                                           |
| Cartesia         | ✅                                                           |
| ElevenLabs       | ✅                                                           |

For every gateway endpoint and provider, see [Supported APIs](/docs/ai-gateway/intro-to-llm-gateway#supported-apis).

**Text to speech (TTS)** turns written text into audio. The gateway supports two ways to call it:

| Approach                  | Use for                                        | Base path                                      |
| ------------------------- | ---------------------------------------------- | ---------------------------------------------- |
| **OpenAI-compatible API** | OpenAI, Azure OpenAI, Azure AI Foundry, Groq   | `{GATEWAY_BASE_URL}`                           |
| **Native SDK**            | Deepgram, Cartesia, ElevenLabs, Gemini, Vertex | `{GATEWAY_BASE_URL}/tts/{providerAccountName}` |

## Add models to the gateway

Before you can call the Text to Speech API, add your TTS models to TrueFoundry through a provider account. When adding a model, select **Text to Speech** as the model type.

| Provider         | Setup guide                                           |
| ---------------- | ----------------------------------------------------- |
| OpenAI           | [OpenAI](/docs/ai-gateway/openai)                     |
| Azure OpenAI     | [Azure OpenAI](/docs/ai-gateway/azure-openai)         |
| Azure AI Foundry | [Azure AI Foundry](/docs/ai-gateway/azure-ai-foundry) |
| Groq             | [Groq](/docs/ai-gateway/groq)                         |
| Deepgram         | [Deepgram](/docs/ai-gateway/deepgram)                 |
| Cartesia         | [Cartesia](/docs/ai-gateway/cartesia)                 |
| ElevenLabs       | [ElevenLabs](/docs/ai-gateway/elevenlabs)             |
| Gemini           | [Google Gemini](/docs/ai-gateway/google-gemini)       |
| Vertex           | [Google Vertex](/docs/ai-gateway/google-vertex)       |

## Code snippets

**Before you start:** Replace `{GATEWAY_BASE_URL}` with your gateway base URL and `your-tfy-api-key` with your TrueFoundry API key. For the native SDK, replace `{providerAccountName}` with the display name of your provider account on TrueFoundry.

<Note>
  **Model names:** For audio (STT/TTS), the model ID in code must match the **display name** of the model on your TrueFoundry provider account.
</Note>

<Note>
  **Which SDK to use:** For OpenAI, Azure OpenAI, Azure AI Foundry, and Groq, use the OpenAI SDK (same API). For Deepgram, Cartesia, ElevenLabs, Gemini, and Vertex, use each provider’s native SDK with the gateway URL above.
</Note>

<CodeGroup>
  ```python OpenAI lines theme={"dark"}
  from pathlib import Path
  from openai import OpenAI

  BASE_URL = "{GATEWAY_BASE_URL}"
  API_KEY = "your-tfy-api-key"

  client = OpenAI(api_key=API_KEY, base_url=BASE_URL)

  output_path = Path(__file__).parent / "output.mp3"
  with client.audio.speech.with_streaming_response.create(
      model="openai-main/gpt-4o-mini-tts",  # truefoundry model name
      voice="alloy",
      input="Hello, how are you?",
  ) as response:
      response.stream_to_file(output_path)

  print(output_path)
  ```

  ```python Azure OpenAI lines theme={"dark"}
  from pathlib import Path
  from openai import OpenAI

  BASE_URL = "{GATEWAY_BASE_URL}"
  API_KEY = "your-tfy-api-key"

  client = OpenAI(api_key=API_KEY, base_url=BASE_URL)

  output_path = Path(__file__).parent / "output.mp3"
  with client.audio.speech.with_streaming_response.create(
      model="azure-openai-main/your-tts-deployment",  # truefoundry model name
      voice="alloy",
      input="Hello, how are you?",
  ) as response:
      response.stream_to_file(output_path)

  print(output_path)
  ```

  ```python Azure AI Foundry lines theme={"dark"}
  from pathlib import Path
  from openai import OpenAI

  BASE_URL = "{GATEWAY_BASE_URL}"
  API_KEY = "your-tfy-api-key"

  client = OpenAI(api_key=API_KEY, base_url=BASE_URL)

  output_path = Path(__file__).parent / "output.mp3"
  with client.audio.speech.with_streaming_response.create(
      model="azure-foundry-main/your-tts-deployment",  # truefoundry model name
      voice="alloy",
      input="Hello, how are you?",
  ) as response:
      response.stream_to_file(output_path)

  print(output_path)
  ```

  ```python Groq lines theme={"dark"}
  from pathlib import Path
  from openai import OpenAI

  BASE_URL = "{GATEWAY_BASE_URL}"
  API_KEY = "your-tfy-api-key"

  client = OpenAI(api_key=API_KEY, base_url=BASE_URL)

  output_path = Path(__file__).parent / "output.mp3"
  with client.audio.speech.with_streaming_response.create(
      model="groq-main/your-tts-model",  # truefoundry model name
      voice="alloy",
      input="Hello, how are you?",
  ) as response:
      response.stream_to_file(output_path)

  print(output_path)
  ```

  ```python Deepgram lines theme={"dark"}
  from deepgram import DeepgramClient
  from deepgram.environment import DeepgramClientEnvironment

  TFY_API_KEY = "your-tfy-api-key"
  OUTPUT_FILE = "/path/to/output.mp3"

  env = DeepgramClientEnvironment(
      base="{GATEWAY_BASE_URL}/tts/{deepgramProviderAccountName}",
      production="",
      agent="",
  )
  client = DeepgramClient(
  	api_key="dummy",
  	environment=env,
  	headers={"x-tfy-api-key": TFY_API_KEY},
  )

  response = client.speak.v1.audio.generate(
      text="Hello! This is a test of Deepgram text to speech via the gateway.",
      model="aura-2-thalia-en",  # actual model id
  )

  with open(OUTPUT_FILE, "wb") as f:
      for chunk in response:
          f.write(chunk)

  print(OUTPUT_FILE)
  ```

  ```python Cartesia lines theme={"dark"}
  import asyncio
  import httpx
  from cartesia import AsyncCartesia

  TFY_API_KEY = "your-tfy-api-key"
  BASE_URL = "{GATEWAY_BASE_URL}/tts/{cartesiaProviderAccountName}"
  OUTPUT_FILE = "/path/to/output.wav"

  client = AsyncCartesia(
      api_key="dummy",
      base_url=BASE_URL,
      http_client=httpx.AsyncClient(
          headers={"x-tfy-api-key": TFY_API_KEY},
      ),
  )


  async def main():
      response = await client.tts.generate(
          model_id="sonic-3",  # actual model id
          transcript="Welcome to Cartesia Sonic!",
          voice={"mode": "id", "id": "a0e99841-438c-4a64-b679-ae501e7d6091"},
          language="en",
          output_format={
              "container": "wav",
              "sample_rate": 44100,
              "encoding": "pcm_s16le",
          },
      )
      with open(OUTPUT_FILE, "wb") as f:
          f.write(await response.read())
      print(OUTPUT_FILE)


  asyncio.run(main())
  ```

  ```python ElevenLabs lines theme={"dark"}
  import httpx
  from elevenlabs import ElevenLabs

  BASE_URL = "{GATEWAY_BASE_URL}/tts/{elevenLabsProviderAccountName}"
  TFY_API_KEY = "your-tfy-api-key"
  OUTPUT_FILE = "/path/to/output.mp3"

  client = ElevenLabs(
      api_key="dummy",
      base_url=BASE_URL,
      httpx_client=httpx.Client(
          headers={"x-tfy-api-key": TFY_API_KEY},
      ),
  )

  response = client.text_to_speech.convert(
      voice_id="JBFqnCBsd6RMkjVDRZzb",
      output_format="mp3_44100_128",
      text="The first move is what sets everything in motion.",
      model_id="eleven_v3",  # actual model id
  )

  with open(OUTPUT_FILE, "wb") as f:
      for chunk in response:
          f.write(chunk)

  print(OUTPUT_FILE)
  ```

  ```python Gemini lines theme={"dark"}
  import wave
  from google import genai
  from google.genai import types

  def save_wav(filename, pcm, channels=1, rate=24000, sample_width=2):
      with wave.open(filename, "wb") as wf:
          wf.setnchannels(channels)
          wf.setsampwidth(sample_width)
          wf.setframerate(rate)
          wf.writeframes(pcm)

  TFY_API_KEY = "your-tfy-api-key"
  BASE_URL = "{GATEWAY_BASE_URL}/tts/{geminiProviderAccountName}"

  client = genai.Client(
      api_key="dummy",
      http_options=types.HttpOptions(
          base_url=BASE_URL,
          headers={"x-tfy-api-key": TFY_API_KEY},
      ),
  )

  response = client.models.generate_content(
      model="gemini-2.5-flash-preview-tts",  # actual model id
      contents="Say cheerfully: Have a wonderful day!",
      config=types.GenerateContentConfig(
          response_modalities=["AUDIO"],
          speech_config=types.SpeechConfig(
              voice_config=types.VoiceConfig(
                  prebuilt_voice_config=types.PrebuiltVoiceConfig(voice_name="Kore"),
              )
          ),
      ),
  )

  audio_data = response.candidates[0].content.parts[0].inline_data.data
  save_wav("output.wav", audio_data)
  print("output.wav")
  ```

  ```python Vertex lines theme={"dark"}
  from google.cloud import texttospeech
  from google.api_core.client_options import ClientOptions
  from google.auth.credentials import AnonymousCredentials

  TFY_API_KEY = "your-tfy-api-key"
  BASE_URL = "{GATEWAY_BASE_URL}/tts/{vertexProviderAccountName}"
  MODEL_NAME = "vertex-main/chirp"  # truefoundry model name 

  client_options = ClientOptions(api_endpoint=BASE_URL)
  client = texttospeech.TextToSpeechClient(
      transport="rest",
      client_options=client_options,
      credentials=AnonymousCredentials(),
  )
  client._transport._session.headers.update({
      "x-tfy-api-key": TFY_API_KEY,
      "x-tfy-model-name": MODEL_NAME,
  })

  response = client.synthesize_speech(
      input=texttospeech.SynthesisInput(text="Hello, how are you doing?"),
      voice=texttospeech.VoiceSelectionParams(
          language_code="en-US",
          name="en-US-Wavenet-D",
      ),
      audio_config=texttospeech.AudioConfig(
          audio_encoding=texttospeech.AudioEncoding.MP3,
      ),
  )

  with open("output.mp3", "wb") as f:
      f.write(response.audio_content)

  print("output.mp3")
  ```
</CodeGroup>

## Response

TTS responses are audio bytes. Each snippet above writes the audio to a file and prints the output path.
