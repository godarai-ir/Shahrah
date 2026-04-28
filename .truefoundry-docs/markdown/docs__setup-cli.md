> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Setup For CLI

> Install and configure the TrueFoundry CLI for managing deployments, experiments, and platform resources.

## TrueFoundry CLI Library

TrueFoundry CLI library to help you interact with the platform programmatically by

* Interacting with the deployments side of TrueFoundry, enabling you to manage workspaces, deployments, applications, and view logs.
* Providing experiment tracking capabilities, allowing you to track ML experiments and interact with ML repositories within TrueFoundry.

## Install Truefoundry CLI

<Info>
  You need Python ≥ 3.9, ≤ 3.14 to use the Truefoundry CLI.
</Info>

```bash lines theme={"dark"}
pip install -U "truefoundry"
```

<Tip>
  Optional features can be installed with extras.

  * `workflow` extra to work with [Workflows](/docs/workflow-concepts).
  * `spark` extra to work with [Spark Tasks](/docs/spark-task) in workflows.

  <CodeGroup>
    ```bash Workflows Without Spark Tasks lines theme={"dark"}
    pip install -U "truefoundry[workflow]"
    ```

    ```bash Workflows With Spark Tasks lines theme={"dark"}
    pip install -U "truefoundry[workflow,spark]"
    ```
  </CodeGroup>

  **NOTE:** `workflow` and `spark` extras are only supported on Python ≥ 3.9, ≤ 3.12
</Tip>

## Authenticate CLI

To be able to use the TrueFoundry CLI, you first need to authenticate. You can login either via browser or using API Key.

<Tabs>
  <Tab title="Login via Browser using Device Code">
    To log in to **TrueFoundry**, run the following command:

    ```bash lines theme={"dark"}
    tfy login --host <your-truefoundry-host> # e.g. host: https://myorg.truefoundry.cloud
    ```

    <Tip>
      The value for `--host` is the first part of the url when you open the platform. This domain might be subdomain of your organization. The below example shows `https://myorg.truefoundry.cloud`

      <img src="https://mintcdn.com/truefoundry/DdP_2rhue4AQQlob/images/28d53c11-768377df9c43a945e1fcf682bb0c146021cbe5f776fec459e1b4247116e2da65-image.png?fit=max&auto=format&n=DdP_2rhue4AQQlob&q=85&s=75815a04e18ad0546dcf4f8be76fe3d4" alt="" width="2058" height="380" data-path="images/28d53c11-768377df9c43a945e1fcf682bb0c146021cbe5f776fec459e1b4247116e2da65-image.png" />
    </Tip>

    This will prompt you to open a URL and show you a `Submit` button.

    <img src="https://mintcdn.com/truefoundry/5CkapnZ7CyjQJ4bx/images/docs/login-code-screen.png?fit=max&auto=format&n=5CkapnZ7CyjQJ4bx&q=85&s=bebafa5f9bb16d2ae8d189a99bf5c50d" alt="" width="3752" height="1488" data-path="images/docs/login-code-screen.png" />

    Once you click on Submit, you will be signed in to the CLI.

    <img src="https://mintcdn.com/truefoundry/5CkapnZ7CyjQJ4bx/images/docs/login-cli-flow.png?fit=max&auto=format&n=5CkapnZ7CyjQJ4bx&q=85&s=c067d7307c822fcbf0e3ade9a2efc5e4" alt="" width="2338" height="254" data-path="images/docs/login-cli-flow.png" />
  </Tab>

  <Tab title="Login using API Key">
    For scenarios where code interacts with TrueFoundry APIs or where opening a browser is not feasible, you can log in using the TrueFoundry API Key. The steps for this are as follows:

    <Steps>
      <Step title="Generate a TrueFoundry API Key">
        Generate a Truefoundry API Key from Access -> Personal Access Tokens.

        <iframe provider="app.supademo.com" href="https://app.supademo.com/embed/cmb6f77un3o41ppkpytsnu7ls?embed_v=2" height="450px" width="100%" src="https://app.supademo.com/embed/cmb6f77un3o41ppkpytsnu7ls?embed_v=2" style={{ border:"none",display:"flex",margin:"auto" }} />

        <Warning>
          Your API Key is a sensitive piece of information - so **Never** share it with unauthorized individuals. You will not be able to see the value of the API key on the Truefoundry platform again - so please save it in a secure location.
        </Warning>

        <Info>
          For production environments, you can use [Virtual Account API Keys](/docs/generating-truefoundry-api-keys#virtual-accounts) instead of Personal Access Tokens to authenticate. Virtual Account API Keys are not tied to a user and hence don't expire when the user leaves the organization.
        </Info>
      </Step>

      <Step title="Use the API key to login">
        You can use the API key to login in two ways:

        #### Option 1: Setting environment variables

        Set the `TFY_API_KEY` and `TFY_HOST` environment variables.

        * Set `TFY_HOST` to your TrueFoundry host URL. e.g. `https://your-domain.truefoundry.com`
        * Set `TFY_API_KEY` to your API key.

        ```bash lines theme={"dark"}
        export TFY_HOST=<your-truefoundry-host>
        export TFY_API_KEY=<your-api-key>
        ```

        You can now use the Truefoundry SDK to interact with the platform.

        #### Option 2: Using the `login` command on CLI

        When using the CLI, you can provide the API key directly within the command:

        ```bash lines theme={"dark"}
        tfy login --host <your-truefoundry-host> --api-key <your-api-key>
        ```
      </Step>
    </Steps>
  </Tab>
</Tabs>
