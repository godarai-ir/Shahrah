> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Cartesia

> Add and configure Cartesia Models in TrueFoundry's AI Gateway

### Adding Models

Add Cartesia models to the AI Gateway and set up access in three steps.

<Steps>
  <Step title="Navigate to Cartesia Models in AI Gateway">
    In the TrueFoundry dashboard, go to **AI Gateway** → **Models** and select **Cartesia**.

    <Frame caption="Navigate to Cartesia Models">
      <img className="hidden dark:block" src="https://mintcdn.com/truefoundry/5j9IxXGQZJVFsqLO/images/Screenshot2026-02-23at8.02.07PM-2.png?fit=max&auto=format&n=5j9IxXGQZJVFsqLO&q=85&s=aa9ffeb76e25e56d4e3e2a70a0e14c26" alt="Navigating to Cartesia Provider Account in AI Gateway" width="3600" height="2006" data-path="images/Screenshot2026-02-23at8.02.07PM-2.png" />

      <img className="dark:hidden" src="https://mintcdn.com/truefoundry/5j9IxXGQZJVFsqLO/images/Screenshot2026-02-23at8.02.07PM-3.png?fit=max&auto=format&n=5j9IxXGQZJVFsqLO&q=85&s=2014a5fbab97d6a57c6112da257ee911" alt="Navigating to Cartesia Provider Account in AI Gateway" width="3600" height="2006" data-path="images/Screenshot2026-02-23at8.02.07PM-3.png" />
    </Frame>
  </Step>

  <Step title="Add Cartesia Account Details">
    Click **Add Cartesia Account**. Enter a unique account name and your Cartesia **API Key**. Optionally add collaborators so other users or teams can use this account. See [Gateway access control](/docs/ai-gateway/gateway-access-control) for details.

    <Frame caption="Cartesia Model Account Form">
      <img className="hidden dark:block" src="https://mintcdn.com/truefoundry/5j9IxXGQZJVFsqLO/images/Screenshot2026-02-23at8.03.55PM-1.png?fit=max&auto=format&n=5j9IxXGQZJVFsqLO&q=85&s=a07da662d0f60d93b9de75c8bc1f6ecc" alt="Cartesia account configuration form with fields for API key and collaborators" width="2514" height="2003" data-path="images/Screenshot2026-02-23at8.03.55PM-1.png" />

      <img className="dark:hidden" src="https://mintcdn.com/truefoundry/5j9IxXGQZJVFsqLO/images/Screenshot2026-02-23at7.50.57PM-1.png?fit=max&auto=format&n=5j9IxXGQZJVFsqLO&q=85&s=9a0a4b8eda5377b2272398134ad58cde" alt="Cartesia account configuration form with fields for API key and collaborators" width="2528" height="2002" data-path="images/Screenshot2026-02-23at7.50.57PM-1.png" />
    </Frame>
  </Step>

  <Step title="Add Models">
    Click **+ Add Model** and fill in the **Display name**, **Model ID**, and **Model type**.

    <Frame>
      <img src="https://mintcdn.com/truefoundry/5j9IxXGQZJVFsqLO/images/Screenshot2026-02-23at7.54.49PM.png?fit=max&auto=format&n=5j9IxXGQZJVFsqLO&q=85&s=7ed41e843c136a881ec02ec286c8918e" alt="Screenshot2026 02 23at7 54 49PM" width="2516" height="2000" data-path="images/Screenshot2026-02-23at7.54.49PM.png" />
    </Frame>

    <Warning>
      For Cartesia, the **Model ID** and **Display name** must be identical.
    </Warning>
  </Step>
</Steps>

### Inference

Once models are added, use the **Cartesia Native SDK** to call TTS and STT endpoints.

<img src="https://mintcdn.com/truefoundry/5j9IxXGQZJVFsqLO/images/Screenshot2026-02-23at8.05.09PM.png?fit=max&auto=format&n=5j9IxXGQZJVFsqLO&q=85&s=041b1fcc9752dd372dadc7a527181595" alt="Code Snippet and Try in Playground Buttons for each Deepinfra model" className="hidden dark:block hidden dark:block" width="3586" height="2010" data-path="images/Screenshot2026-02-23at8.05.09PM.png" />

<img src="https://mintcdn.com/truefoundry/5j9IxXGQZJVFsqLO/images/Screenshot2026-02-23at8.05.09PM-1.png?fit=max&auto=format&n=5j9IxXGQZJVFsqLO&q=85&s=a695085b03ad4866a3d61f4b4d2cf2d9" alt="Code Snippet and Try in Playground Buttons for each Deepinfra model" className="hidden dark:block dark:hidden" width="3586" height="2010" data-path="images/Screenshot2026-02-23at8.05.09PM-1.png" />

<img src="https://mintcdn.com/truefoundry/5j9IxXGQZJVFsqLO/images/Screenshot2026-02-23at7.56.04PM-1.png?fit=max&auto=format&n=5j9IxXGQZJVFsqLO&q=85&s=4a62d6d085b3d715e7ceced1aa445cfa" alt="Code Snippet and Try in Playground Buttons for each Deepinfra model" className="dark:hidden" width="3590" height="2014" data-path="images/Screenshot2026-02-23at7.56.04PM-1.png" />
