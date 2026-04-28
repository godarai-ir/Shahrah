> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Customize Build Workflow

> Inject custom logic into TrueFoundry's image build pipeline for services, jobs, and workflows. Run scripts before, during, or after the build and push steps.

When TrueFoundry builds a Docker image for a **service**, **job**, **workflow**, or other workload, it runs a fixed sequence of steps: download code, log in to the registry, build the image, and push it.

You can **customize this build workflow** by overriding the main build script and adding your own logic-for example, security scans, tests, or custom tooling-**before**, **during**, or **after** those steps. Your custom script runs in order with the platform’s built-in steps.

<Info>
  This customization applies to the **platform-level** build workflow (Helm values). It affects how images are built whenever the platform builds an image for any service, job, or workflow.
</Info>

## When would you use this?

Common use cases:

* **Security or quality checks**: Run a scanner (e.g. SonarQube, Trivy) on the source before building.
* **Pre-build steps**: Install tools, run linters, or run tests before the image is built and pushed.
* **Post-build steps**: Notify external systems, tag images, or run checks after the image is pushed.

You do this by writing a **custom script**, putting it in a ConfigMap so the build pod can run it, and then wiring it into the platform’s build script in Helm. Follow the steps below.

## What the build script looks like

The platform runs a **main build script** for every image build: a sequence of steps. You override it in Helm and **insert** calls to your custom script where you want (before, during, or after the built-in steps).

The default sequence is:

1. **`download-code.sh`** - fetch source code
2. **`registry-login.sh`** - log in to the container registry
3. **`wait-for-builder.sh`** - wait for the builder
4. **`build-and-push.sh`** - build the image and push it
5. **`update-build.sh`** - report build status

Your custom script is **mounted at `/custom-scripts/`** in the build pod. In the main script you add a line like `/custom-scripts/<your-script-filename>` where you want it to run-e.g. after step 1 (pre-build scan) or after step 4 (post-build notify).

The structure looks like this:

```yaml theme={"dark"}
script: |
  download-code.sh

  # Your custom script runs here (e.g. scan, test, lint)
  /custom-scripts/<your-script-filename>

  registry-login.sh
  wait-for-builder.sh
  build-and-push.sh

  # Or run it here (e.g. notify, tag)
  # /custom-scripts/<your-script-filename>

  update-build.sh '{"status":"SUCCEEDED"}'
```

<Warning>
  **Do not remove** any of the existing steps (`download-code.sh`, `registry-login.sh`, `wait-for-builder.sh`, `build-and-push.sh`, `update-build.sh`) from the script. Removing them can break the build workflow.
</Warning>

Once you know where your script fits in this sequence, you can write the script (Step 1), put it in a ConfigMap (Step 2), and then add the full configuration in Helm (Step 3).

## Add custom logic to the build pipeline

<Steps>
  <Step title="Write your custom script">
    A **script** here is a bash file that runs inside the build pod when the platform builds an image.

    You write it once; in the next steps you’ll put it in a ConfigMap, mount it at **`/custom-scripts/`**, and call it from the main build script (e.g. **`/custom-scripts/scan-before-build.sh`**) at the point you choose-before, during, or after the built-in steps. You’re defining the **logic** that runs at that point (e.g. a scan, test, or notification).

    Create a bash script with that logic. Remember the filename-you’ll use it for the ConfigMap (Step 2) and for the invocation line in the main build script (Step 3).

    Example: scanning source code with SonarQube before the image is built and pushed:

    <CodeGroup>
      ```bash scan-before-build.sh theme={"dark"}
      #!/bin/bash

      # Example: scan source code using SonarQube
      # Ref: https://docs.sonarsource.com/sonarqube-server/latest/analyzing-source-code/scanners/sonarscanner/#running-from-zip-file
      sonar-scanner $SOURCE_CODE_DOWNLOAD_PATH -Dsonar.token=$SONARQUBE_AUTH_TOKEN
      ```
    </CodeGroup>

    <Note>
      **Use the absolute path** when invoking your script from the build workflow (e.g. `/custom-scripts/scan-before-build.sh`). The default working directory in the build pod is `/scripts`.
    </Note>

    <Info>
      **Environment variables:** If your script needs extra env vars, add them in Step 3 under `tfyBuild.truefoundryWorkflows.extraEnvs`. Do **not** overwrite these reserved variables: `SOURCE_CODE_DOWNLOAD_PATH`, `DOCKER_REGISTRY_URL`, `DOCKER_REGISTRY_USERNAME`, `DOCKER_REGISTRY_PASSWORD`, `DOCKER_REPO`, `DOCKER_TAG`, `CALLBACK_URL`. You can read and use them in your script if needed.
    </Info>
  </Step>

  <Step title="Create a ConfigMap from your script">
    The build workflow runs in a pod that must have access to your script. A **ConfigMap** stores the script so the pod can mount it as a file.

    Create the ConfigMap so the script is available when the build runs:

    <CodeGroup>
      ```bash Terminal theme={"dark"}
      kubectl create configmap <configmap-name> --from-file=<script-file-name> -n truefoundry
      ```
    </CodeGroup>

    Example: if your script is `scan-before-build.sh` and you want the ConfigMap named `my-build-scripts`:

    ```bash theme={"dark"}
    kubectl create configmap my-build-scripts --from-file=scan-before-build.sh -n truefoundry
    ```
  </Step>

  <Step title="Add your script to the TrueFoundry build workflow">
    You need to:

    1. **Mount the ConfigMap** so the build pod can read your script from `/custom-scripts/`.
    2. **Set the main build script** in Helm so it calls your script at the right place (see [What the build script looks like](#what-the-build-script-looks-like) for the sequence).

    In your Helm values, set `tfyBuild.truefoundryWorkflows` as below. Replace `<config-map-name>` with the ConfigMap from Step 2, and `<your-script-filename>` with the script file from Step 1 (e.g. `scan-before-build.sh`).

    <Warning>
      Keep all existing steps in the script. Do not remove `download-code.sh`, `registry-login.sh`, `wait-for-builder.sh`, `build-and-push.sh`, or `update-build.sh`.
    </Warning>

    <Note>
      Set **`defaultMode: 511`** on the ConfigMap volume so the mounted script file is executable.
    </Note>

    <CodeGroup>
      ```yaml values.yaml theme={"dark"}
      tfyBuild:
        truefoundryWorkflows:
          extraVolumes:
            - name: custom-script
              configMap:
                name: <config-map-name>
                defaultMode: 511
          extraVolumeMounts:
            - name: custom-script
              mountPath: /custom-scripts
          sfyBuilder:
            script: |
              download-code.sh

              # Run your script from Step 1 (mounted at /custom-scripts/)
              /custom-scripts/<your-script-filename>

              registry-login.sh
              wait-for-builder.sh
              build-and-push.sh

              # Optional: run another script from Step 1 after build and push (e.g. notify, tag)
              # /custom-scripts/<your-script-filename>

              update-build.sh '{"status":"SUCCEEDED"}'
      ```
    </CodeGroup>

    * **`extraVolumes`** / **`extraVolumeMounts`**: Mount your ConfigMap at `/custom-scripts` so the pod can execute your script.
    * **`defaultMode: 511`**: Makes the file executable (required for the script to run).
    * **`sfyBuilder.script`**: This is the **main build script** that runs for every image build. The logic you add here (including calls to `/custom-scripts/...`) runs in order with the built-in steps.

    <Info>
      To pass **environment variables** to your custom script, add them under `tfyBuild.truefoundryWorkflows.extraEnvs`. Do not overwrite the reserved variables listed in Step 1.
    </Info>

    After you apply these values and the platform runs a build, it will use this script and your custom logic will run at the positions you defined.
  </Step>
</Steps>

***
