> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Mounting Volumes/Files

> Mount volumes, secrets, or configuration strings to your services for persistent data and runtime settings.

In some applications, you might need a file to be present at a certain path for the application to work. Or you have data in directories that need to be mounted for the application to work. You can mount data to your service either by mounting a volume, secret or a string.

<Tabs>
  <Tab title="Volume">
    If your application needs access to multiple files of data or multiple directories, you can use a volume to store the data and then mount the volume at the desired path. You can learn more about Volumes [here](/docs/introduction-to-volume). **To use a persistent volume, we will first need to create one and then attach it to your deployment**. You can learn how to create volumes using the [Creating a Volume](/docs/creating-a-volume) guide.

    <Note>
      You can only mount volumes in the same workspace as your service.
    </Note>

    <iframe provider="app.supademo.com" href="https://app.supademo.com/embed/r-RLmi4qJTsy0vSvh30Ld" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/r-RLmi4qJTsy0vSvh30Ld" style={{ border: "none", display: "flex", margin: "auto" }} />
  </Tab>

  <Tab title="String">
    This can be useful if you need a small configuration file to be present at a certain file path. To configure this, you need to provide the path where it needs to be mounted and the string data that should be in that file. **A good example can be a Nginx configuration file that you can mount along with the Nginx docker image.**

    <iframe provider="app.supademo.com" href="https://app.supademo.com/embed/qrpHmqJS1qtdJSu4exeGm" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/qrpHmqJS1qtdJSu4exeGm" style={{ border: "none", display: "flex", margin: "auto" }} />
  </Tab>

  <Tab title="Secret">
    This is similar to string mount, except, in this case, you will directly provide a TrueFoundry Secret FQN. You can read more about Secrets and how to create them [here](/docs/manage-secrets).

    The content in the secret will be dumped into a file and mounted in the provided location. **A good usecase of this is for mounting Google credentials file which you might need to access Google services.**

    <iframe provider="app.supademo.com" href="https://app.supademo.com/embed/0--BuOA6US4apaovMtx3l " typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/0--BuOA6US4apaovMtx3l " style={{ border: "none", display: "flex", margin: "auto" }} />
  </Tab>
</Tabs>

## Using mounted files in your deployment

Once you have attached a file to your deployment, you can use it in your deployment like any other file. For example, if you mounted the file to `/etc/config.json`, you can access the file in the `/etc/config.json` path

<CodeGroup>
  ```python Python lines theme={"dark"}
  import json

  with open("/etc/config.json", "r") as f:
    config_data = json.load(f)

  # Access specific values from the config data dictionary
  api_key = config_data["api_key"]
  database_url = config_data["database_url"]

  # Use these values in your application logic
  print(f"API key: {api_key}")
  print(f"Database URL: {database_url}")
  ```
</CodeGroup>

***
