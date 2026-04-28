> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Image Edit API (/images/edits)

> Edit and transform images using text prompts via TrueFoundry AI Gateway

## Provider capabilities

The table below summarizes gateway support for this endpoint by provider.

<Info>
  Legend:

  * **✅** Supported by Provider and Truefoundry
  * <Icon icon="circle-xmark" iconType="regular" color="red" /> Provided by provider, but not by Truefoundry
  * <Icon icon="circle-minus" iconType="regular" /> Provider does not support this feature
</Info>

| Provider     | Edit                                                        |
| ------------ | ----------------------------------------------------------- |
| OpenAI       | ✅                                                           |
| Azure OpenAI | ✅                                                           |
| Bedrock      | ✅                                                           |
| Vertex       | ✅                                                           |
| Anthropic    | <Icon icon="circle-minus" iconType="regular" />             |
| Cohere       | <Icon icon="circle-minus" iconType="regular" />             |
| Gemini       | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| Groq         | <Icon icon="circle-minus" iconType="regular" />             |
| Together-AI  | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| xAI          | <Icon icon="circle-minus" iconType="regular" />             |
| DeepInfra    | <Icon icon="circle-xmark" iconType="regular" color="red" /> |

For every gateway endpoint and provider, see [Supported APIs](/docs/ai-gateway/intro-to-llm-gateway#supported-apis).

The Image Edit API lets you modify images using text instructions. You can edit specific parts of an image, add new elements, or extend the image beyond its original boundaries. Just provide your source image and describe what changes you want to make.

## Supported Providers

* **OpenAI**: Supports `dall-e-2` and `gpt-image-1` models
* **Vertex AI**: Supports `imagen-3.0-generate-001` model
* **AWS Bedrock**: Supports `amazon-nova-canvas` model
* **Azure OpenAI**: Supports `gpt-image-1` model

## Requirements

| Provider     | Model                     | Format         | Size Limit | Image Count     |
| ------------ | ------------------------- | -------------- | ---------- | --------------- |
| OpenAI       | gpt-image-1               | PNG, WebP, JPG | \< 50MB    | Up to 16 images |
| OpenAI       | dall-e-2                  | Square PNG     | \< 4MB     | 1 image only    |
| Vertex AI    | imagen-3.0-capability-001 | PNG, JPG       | \< 20MB    | 1 image only    |
| AWS Bedrock  | amazon-nova-canvas        | PNG, JPG       | \< 5MB     | 1 image only    |
| Azure OpenAI | gpt-image-1               | PNG, WebP, JPG | \< 50MB    | Up to 16 images |

## Example Usage

<AccordionGroup>
  <Accordion title="OpenAI (gpt-image-1, dall-e-2)">
    OpenAI supports both `gpt-image-1` (up to 16 images, PNG/WebP/JPG, \<50MB) and `dall-e-2` (1 square PNG image, \<4MB). The meaning of the parameters are present in the [OpenAI API documentation](https://platform.openai.com/docs/api-reference/images/create-edit).

    <CodeGroup>
      ```python gpt-image-1 lines theme={"dark"}
      from openai import OpenAI

      BASE_URL = "{GATEWAY_BASE_URL}"
      API_KEY = "your-truefoundry-api-key"

      # Configure OpenAI client with TrueFoundry settings
      client = OpenAI(
          api_key=API_KEY,
          base_url=BASE_URL,
      )

      response = client.images.edit(
          model="openai-main/gpt-image-1",
          image=[
              open("image1.png", "rb"),  # First source image
              open("image2.png", "rb")   # Second source image (up to 16 images supported)
          ],
          prompt="Replace the background with a beach scene and add palm trees on both sides",
          mask=open("mask.png", "rb")    # Optional mask to specify edit areas
      )

      print(response.data[0].url)
      ```

      ```python dall-e-2 lines theme={"dark"}
      from openai import OpenAI

      BASE_URL = "{GATEWAY_BASE_URL}"
      API_KEY = "your-truefoundry-api-key"

      # Configure OpenAI client with TrueFoundry settings
      client = OpenAI(
          api_key=API_KEY,
          base_url=BASE_URL,
      )

      response = client.images.edit(
          model="openai-main/dall-e-2",
          image=open("image.png", "rb"),  # Single square PNG image (under 4MB)
          prompt="Add a sunset in the background",
          mask=open("mask.png", "rb")     # Optional mask
      )

      print(response.data[0].url)
      ```
    </CodeGroup>

    <Note>
      `gpt-image-1` supports up to 16 images in PNG, WebP, or JPG format with a 50MB size limit per image. `dall-e-2` requires a single square PNG image under 4MB.
    </Note>
  </Accordion>

  <Accordion title="Vertex AI (imagen-3.0-capability-001)">
    Vertex AI's Imagen 3.0 model supports editing a single image in PNG or JPG format (max 20MB). The meaning of the parameters are present in the [Google Vertex documentation](https://cloud.google.com/vertex-ai/generative-ai/docs/image/edit-insert-objects).

    <CodeGroup>
      ```python Vertex lines theme={"dark"}
      from openai import OpenAI

      BASE_URL = "{GATEWAY_BASE_URL}"
      API_KEY = "your-truefoundry-api-key"

      # Configure OpenAI client with TrueFoundry settings
      client = OpenAI(
          api_key=API_KEY,
          base_url=BASE_URL,
      )

      response = client.images.edit(
          model="image-edit/imagen-3.0-capability-001",
          image=open("image.png", "rb"), # REQUIRED, source image
          prompt="Replace the background with a beach scene and add palm trees on both sides", # Optional
          mask=open("mask.png", "rb"), # REQUIRED
          n = 2,
          extra_body={
              "maskMode": "MASK_MODE_BACKGROUND",
              "maskClasses": [162, 170], # ONLY if maskMode="MASK_MODE_SEMANTIC"
              "dilation": 0.01,
              "baseSteps": 35,
              "editMode": "EDIT_MODE_INPAINT_REMOVAL"
          }

      print(response.data[0].b64_json)
      ```
    </CodeGroup>

    <Note>
      Vertex AI supports a single image in PNG or JPG format with a maximum size of 20MB.
    </Note>
  </Accordion>

  <Accordion title="AWS Bedrock (amazon-nova-canvas)">
    AWS Bedrock's Nova Canvas model supports editing a single image in PNG or JPG format (max 5MB).

    <Accordion title="Amazon Nova Canvas">
      The meaning of the parameters are present in the [AWS Bedrock documentation](https://docs.aws.amazon.com/nova/latest/userguide/image-gen-req-resp-structure.html).

      <CodeGroup>
        ```python Amazon Nova Canvas lines theme={"dark"}
        from openai import OpenAI

        BASE_URL = "{GATEWAY_BASE_URL}"
        API_KEY = "your-truefoundry-api-key"

        # Configure OpenAI client with TrueFoundry settings
        client = OpenAI(
            api_key=API_KEY,
            base_url=BASE_URL,
        )

        response = client.images.edit(
            model="tfy-ai-bedrock/amazon-nova-canvas"
            image=open("image.png", "rb"), # REQUIRED, source image
            prompt="Replace the background with a beach scene and add palm trees on both sides",
            # Mask image OR mask prompt is required, not both
            mask=open("mask.png", "rb")    # Mask image to specify edit areas
            
            extra_body={
        	    "taskType": "OUTPAINTING"          # Optional: defaults to INPAINTING
        	    "maskPrompt": "Box in the center." # Required if mask image is not specified, should NOT be included if mask image is specified
        	    "negativeText": "dogs, cats"       # Optional: A text prompt to define what not to include in the image.
        	    
        	    # If taskType=OUTPAINTING
        	    "outPaintingMode": "PRECISE"       # Optional: DEFAULT | PRECISE
        	    
        	    # For VIRTUAL_TRY_ON, need to use curl request only
        	  }
        )

        print(response.data[0].b64_json)
        ```

        ```python Virtual try-on lines  theme={"dark"}
        import requests
        import json
        import base64

        # --- Configuration ---
        # !! Replace with your actual token and file paths !!
        API_URL = "{GATEWAY_BASE_URL}/images/edits"
        BEARER_TOKEN = "YOUR_BEARER_TOKEN"  # Replace with your actual token
        IMAGE_PATH = "image.jpg"            # Path to your main image
        REFERENCE_IMAGE_PATH = "reference.jpg"  # Path to your reference image
        MASK_IMAGE_PATH = "mask.jpg"        # Path for the mask (since we use maskType=IMAGE)

        # ---------------------

        # Define the non-file form fields
        form_data = {
            "model": "tfy-ai-bedrock/amazon-nova-canvas",
            "maskType": "IMAGE",           # Example: Using IMAGE maskType
            
            # --- General Optional Fields ---
            "n": 1,
            "size": "1024x1024",
            "cfgScale": 9.0,
            "seed": 42,
            "quality": "hd",
            "preserveBodyPose": "ON",
            "preserveHands": "ON",
            "preserveFace": "ON",
            "mergeStyle": "BALANCED",
            "returnMask": "true"
            
            # --- Fields for maskType=GARMENT ---
            # "garmentClass": "UPPER_BODY",
            # "maskShape": "CONTOUR",
            
            # --- Fields for maskType=PROMPT ---
            # "maskPrompt": "box in the center",
            # "maskShape": "CONTOUR",
        }

        # Define the headers
        headers = {
            "Authorization": f"Bearer {BEARER_TOKEN}"
            # 'Content-Type: multipart/form-data' is set automatically by requests
        }

        try:
            # Open the files in binary read mode
            with open(IMAGE_PATH, "rb") as image_file, \
                 open(REFERENCE_IMAGE_PATH, "rb") as ref_image_file, \
                 open(MASK_IMAGE_PATH, "rb") as mask_file:

                # Define the files to be uploaded
                # The format is ('filename', file_object, 'content_type')
                files_to_upload = {
                    "image": (IMAGE_PATH, image_file, "image/jpeg"),
                    "referenceImage": (REFERENCE_IMAGE_PATH, ref_image_file, "image/jpeg"),
                    "mask": (MASK_IMAGE_PATH, mask_file, "image/jpeg") # Required for maskType=IMAGE
                }

                print(f"Sending request to {API_URL}...")
                
                # Make the POST request
                response = requests.post(
                    API_URL,
                    headers=headers,
                    data=form_data,    # Non-file form fields
                    files=files_to_upload  # File form fields
                )

                # Check for HTTP errors
                response.raise_for_status()

                print("Request successful.")
                response_json = response.json()

                # Process the response (similar to your original script)
                if "data" in response_json and len(response_json["data"]) > 0:
                    first_result = response_json["data"][0]

                    if "url" in first_result:
                        print(f"Image URL: {first_result['url']}")

                    if "b64_json" in first_result:
                        print("Found base64 image. Saving to edited_image.png...")
                        image_bytes = base64.b64decode(first_result["b64_json"])
                        
                        with open("edited_image.png", "wb") as f:
                            f.write(image_bytes)
                        print("Image saved as edited_image.png")
                    
                    # You might also want to check for the returned mask if returnMask=true
                    if "mask" in first_result and "b64_json" in first_result["mask"]:
                         print("Found base64 mask. Saving to returned_mask.png...")
                         mask_bytes = base64.b64decode(first_result["mask"]["b64_json"])
                         with open("returned_mask.png", "wb") as f:
                             f.write(mask_bytes)
                         print("Mask saved as returned_mask.png")

                else:
                    print("Response received, but no data found or data is empty.")
                    print("Response JSON:", response_json)

        except requests.exceptions.HTTPError as e:
            print(f"HTTP error: {e}")
            print("Response body:", e.response.text)
        except requests.exceptions.RequestException as e:
            print(f"Request error: {e}")
        except FileNotFoundError as e:
            print(f"File error: {e}. Make sure files exist at the specified paths.")
        except (KeyError, json.JSONDecodeError) as e:
            print(f"Parsing error: {e}")
        except Exception as e:
            print(f"An unexpected error occurred: {e}")
        ```
      </CodeGroup>
    </Accordion>

    <Accordion title="Stability AI">
      The meaning of the parameters are present in the [AWS Bedrock documentation](https://docs.aws.amazon.com/bedrock/latest/userguide/stable-image-services.html).

      <CodeGroup>
        ```python Inpaint lines theme={"dark"}
        from openai import OpenAI

        BASE_URL = "{GATEWAY_BASE_URL}"
        API_KEY = "your-truefoundry-api-key"

        # Configure OpenAI client with TrueFoundry settings
        client = OpenAI(
            api_key=API_KEY,
            base_url=BASE_URL,
        )

        response = client.images.edit(
            model="tfy-ai-bedrock/stable-image-inpaint1",
            image=open("image.png", "rb"), # REQUIRED, source image
            prompt="Replace the background with a beach scene and add palm trees on both sides", # REQUIRED
        		output_format="jpeg", # Optional
            mask=open(filePath, "rb"), # Optional
            extra_body={
               "style_preset": "anime", # Optional
               "negative_prompt": "dogs", # Optional
               "seed": 42, # Optional
               "grow_mask": 7 # Optional
           }

        print(response.data[0].b64_json)
        ```

        ```python Search And Recolor lines theme={"dark"}
        from openai import OpenAI

        BASE_URL = "{GATEWAY_BASE_URL}"
        API_KEY = "your-truefoundry-api-key"

        # Configure OpenAI client with TrueFoundry settings
        client = OpenAI(
            api_key=API_KEY,
            base_url=BASE_URL,
        )

        response = client.images.edit(
            model="tfy-ai-bedrock/stable-image-search-recolor-v1",
            image=open("image.png", "rb"), # REQUIRED, source image
            prompt="Replace the background with a beach scene and add palm trees on both sides", # REQUIRED
        		output_format="jpeg", # Optional
            mask=open(filePath, "rb"), # Optional
            extra_body={
               "select_prompt": "Select the jacket.", # REQUIRED
               "style_preset": "anime", # Optional
               "negative_prompt": "dogs", # Optional
               "seed": 42, # Optional
               "grow_mask": 7 # Optional
           }

        print(response.data[0].b64_json)
        ```

        ```python Search and Replace lines theme={"dark"}
        from openai import OpenAI

        BASE_URL = "{GATEWAY_BASE_URL}"
        API_KEY = "your-truefoundry-api-key"

        # Configure OpenAI client with TrueFoundry settings
        client = OpenAI(
            api_key=API_KEY,
            base_url=BASE_URL,
        )

        response = client.images.edit(
            model="tfy-ai-bedrock/stable-image-search-replace-v1",
            image=open("image.png", "rb"), # REQUIRED, source image
            prompt="Replace the background with a beach scene and add palm trees on both sides", # REQUIRED
        		output_format="jpeg", # Optional
            mask=open(filePath, "rb"), # Optional
            extra_body={
               "search_prompt": "Select the jacket.", # REQUIRED
               "style_preset": "anime", # Optional
               "negative_prompt": "dogs", # Optional
               "seed": 42, # Optional
               "grow_mask": 7 # Optional
           }

        print(response.data[0].b64_json)
        ```

        ```python Erase lines theme={"dark"}
        from openai import OpenAI

        BASE_URL = "{GATEWAY_BASE_URL}"
        API_KEY = "your-truefoundry-api-key"

        # Configure OpenAI client with TrueFoundry settings
        client = OpenAI(
            api_key=API_KEY,
            base_url=BASE_URL,
        )

        response = client.images.edit(
            model="tfy-ai-bedrock/stable-image-erase-object-v1",
            image=open("image.png", "rb"), # REQUIRED, source image
            prompt="Replace the background with a beach scene and add palm trees on both sides", # REQUIRED
        		output_format="jpeg", # Optional
            mask=open(filePath, "rb"), # Optional
            extra_body={
               "seed": 42, # Optional
               "grow_mask": 7 # Optional
           }

        print(response.data[0].b64_json)
        ```

        ```python Remove Backgroud lines theme={"dark"}
        from openai import OpenAI

        BASE_URL = "{GATEWAY_BASE_URL}"
        API_KEY = "your-truefoundry-api-key"

        # Configure OpenAI client with TrueFoundry settings
        client = OpenAI(
            api_key=API_KEY,
            base_url=BASE_URL,
        )

        response = client.images.edit(
            model="tfy-ai-bedrock/stable-image-remove-background-v1",
            image=open("image.png", "rb"), # REQUIRED, source image
        	output_format="jpeg", # Optional
        		
        print(response.data[0].b64_json)
        ```
      </CodeGroup>
    </Accordion>

    <Note>
      AWS Bedrock supports a single image in PNG or JPG format with a maximum size of 5MB.
    </Note>
  </Accordion>

  <Accordion title="Azure OpenAI (gpt-image-1)">
    Azure OpenAI's `gpt-image-1` model supports up to 16 images in PNG, WebP, or JPG format (max 50MB each). The meaning of the parameters are present in the [Azure OpenAI documentation](https://learn.microsoft.com/en-us/azure/ai-foundry/openai/how-to/dall-e?tabs=gpt-image-1#specify-image-edit-api-options).

    <CodeGroup>
      ```python Python lines theme={"dark"}
      from openai import OpenAI

      BASE_URL = "{GATEWAY_BASE_URL}"
      API_KEY = "your-truefoundry-api-key"

      # Configure OpenAI client with TrueFoundry settings
      client = OpenAI(
          api_key=API_KEY,
          base_url=BASE_URL,
      )

      response = client.images.edit(
          model="azure-main/gpt-image-1",
          image=[
              open("image1.png", "rb"),  # First source image
              open("image2.png", "rb")   # Second source image (up to 16 images supported)
          ],
          prompt="Replace the background with a beach scene and add palm trees on both sides",
          mask=open("mask.png", "rb")    # Optional mask to specify edit areas
      )

      print(response.data[0].url)
      ```
    </CodeGroup>

    <Note>
      Azure OpenAI supports up to 16 images in PNG, WebP, or JPG format with a 50MB size limit per image.
    </Note>
  </Accordion>
</AccordionGroup>

## Response Format

The API returns an `ImagesResponse` object containing:

```python lines theme={"dark"}
ImagesResponse(
    created=1755685741,
    data=[
        Image(
            url='https://oaidalleapiprodscus.blob.core.windows.net/private/org-ojH41IdW0UR2VlysxKUx8AjA/user-9QSCTtrOEHbbiQRFfFbwT8fx/img-PlwCalRpn4j5jQxG1wKvQYGc.png?st=2025-08-20T09%3A29%3A01Z&se=2025-08-20T11%3A29%3A01Z&sp=r&sv=2024-08-04&sr=b&rscd=inline&rsct=image/png&skoid=32836cae-d25f-4fe9-827b-1c8c59c442cc&sktid=a48cca56-e6da-484e-a814-9c849652bcb3&skt=2025-08-20T10%3A01%3A35Z&ske=2025-08-21T10%3A01%3A35Z&sks=b&skv=2024-08-04&sig=g21zsMrRuM8aRjO5lLyVVwxZD7K4Ng1OoI7QZ5e8Y4Q%3D',
            b64_json=None,
            revised_prompt=None
        )
    ],
    background=None,
    output_format=None,
    quality=None,
    size=None,
    usage=None
)
```
