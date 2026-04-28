> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# ElevenLabs

> Add and configure ElevenLabs Models in TrueFoundry's AI Gateway

### Adding Models

Add ElevenLabs models to the AI Gateway and set up access in three steps.

<Steps>
  <Step title="Navigate to ElevenLabs Models in AI Gateway">
    In the TrueFoundry dashboard, go to **AI Gateway** → **Models** and select **ElevenLabs**.

    <Frame caption="Navigate to ElevenLabs Models">
      <img className="hidden dark:block" src="https://mintcdn.com/truefoundry/5j9IxXGQZJVFsqLO/images/Screenshot2026-02-23at7.49.29PM-1.png?fit=max&auto=format&n=5j9IxXGQZJVFsqLO&q=85&s=4b0b4464e77ef19a6b15091bf063bf48" alt="Navigating to ElevenLabs Provider Account in AI Gateway" width="3600" height="2006" data-path="images/Screenshot2026-02-23at7.49.29PM-1.png" />

      <img className="dark:hidden" src="https://mintcdn.com/truefoundry/5j9IxXGQZJVFsqLO/images/Screenshot2026-02-23at7.49.29PM.png?fit=max&auto=format&n=5j9IxXGQZJVFsqLO&q=85&s=9b2097aed224fd79dc2c36788c3c752b" alt="Navigating to ElevenLabs Provider Account in AI Gateway" width="3600" height="2006" data-path="images/Screenshot2026-02-23at7.49.29PM.png" />
    </Frame>
  </Step>

  <Step title="Add ElevenLabs Account Details">
    Click **Add ElevenLabs Account**. Enter a unique account name and your ElevenLabs **API Key**. Optionally add collaborators so other users or teams can use this account. See [Gateway access control](/docs/ai-gateway/gateway-access-control) for details.

    <Frame caption="ElevenLabs Model Account Form">
      <img className="hidden dark:block" src="https://mintcdn.com/truefoundry/5j9IxXGQZJVFsqLO/images/Screenshot2026-02-23at7.50.57PM.png?fit=max&auto=format&n=5j9IxXGQZJVFsqLO&q=85&s=6d20c77cfa9d9f544ddec4ee881b67a2" alt="ElevenLabs account configuration form with fields for API key and collaborators" width="2528" height="2002" data-path="images/Screenshot2026-02-23at7.50.57PM.png" />

      <img className="dark:hidden" src="https://mintcdn.com/truefoundry/5j9IxXGQZJVFsqLO/images/Screenshot2026-02-23at7.50.57PM-1.png?fit=max&auto=format&n=5j9IxXGQZJVFsqLO&q=85&s=9a0a4b8eda5377b2272398134ad58cde" alt="ElevenLabs account configuration form with fields for API key and collaborators" width="2528" height="2002" data-path="images/Screenshot2026-02-23at7.50.57PM-1.png" />
    </Frame>
  </Step>

  <Step title="Add Models">
    Click **+ Add Model** and fill in the **Display name**, **Model ID**, and **Model type**.

    <Frame>
      <img src="https://mintcdn.com/truefoundry/5j9IxXGQZJVFsqLO/images/Screenshot2026-02-23at7.54.49PM.png?fit=max&auto=format&n=5j9IxXGQZJVFsqLO&q=85&s=7ed41e843c136a881ec02ec286c8918e" alt="Screenshot2026 02 23at7 54 49PM" width="2516" height="2000" data-path="images/Screenshot2026-02-23at7.54.49PM.png" />
    </Frame>

    <Warning>
      For ElevenLabs, the **Model ID** and **Display name** must be identical.
    </Warning>
  </Step>
</Steps>

### Inference

Once models are added, use the **ElevenLabs Native SDK** to call TTS and STT endpoints.

<img src="https://mintcdn.com/truefoundry/5j9IxXGQZJVFsqLO/images/Screenshot2026-02-23at7.56.04PM-2.png?fit=max&auto=format&n=5j9IxXGQZJVFsqLO&q=85&s=462c1db1f276b95830134825a92bf9a8" alt="Code Snippet and Try in Playground Buttons for each Deepinfra model" className="hidden dark:block" width="3590" height="2014" data-path="images/Screenshot2026-02-23at7.56.04PM-2.png" />

<img src="https://mintcdn.com/truefoundry/5j9IxXGQZJVFsqLO/images/Screenshot2026-02-23at7.56.04PM-1.png?fit=max&auto=format&n=5j9IxXGQZJVFsqLO&q=85&s=4a62d6d085b3d715e7ceced1aa445cfa" alt="Code Snippet and Try in Playground Buttons for each Deepinfra model" className="dark:hidden" width="3590" height="2014" data-path="images/Screenshot2026-02-23at7.56.04PM-1.png" />
