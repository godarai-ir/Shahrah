> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Customizing CI/CD Templates

> Learn how to customize CI/CD templates on TrueFoundry for faster, consistent deployments.

The CI/CD templates can be configured on a control plane level (available for customers using TrueFoundry in separate hosted control plane only). This customization can be used to add custom image scanning workflows, security scans etc to the CI/CD pipelines or add whatever custom logic user wants to add.

### How do CI/CD templates get rendered?

TrueFoundry renders different CI/CD templates as shown in this [document](/docs/setting-up-cicd-for-your-service) automatically based on the deployment.\
All the templates can be found [here](https://github.com/truefoundry/infra-charts/tree/main/charts/tfy-configs/files/cicd-templates).

There is a file called `cicd-providers.yaml` which lists the providers need to be enabled like github, bitbucket and gitlab. All other files contain templates to configure CI/CD based on different cases like:

1. Whether image is built on TrueFoundry or built elsewhere.
2. Whether the spec is stored as a yaml in your repository (complete GitOps) or not.

These are different templates for gitlab, bitbucket and github.

### How to customize these templates for your Control Plane?

The CI/CD templates are supplied to TrueFoundry's control plane via a config map. These config map values can be overridden by updating the values of `truefoundry` helm chart

You will need to update the values of tfy-configs in the `truefoundry` helm chart.

You need to define the content of `cicd-providers.yaml`with your provider specific details. Here is the [default file.](https://github.com/truefoundry/infra-charts/blob/main/charts/tfy-configs/files/cicd-templates/cicd-providers.yaml)\
Apart from this you need to define all the the other templates of a particular provider (for e.g. github) by taking reference from the [files in this folder.](https://github.com/truefoundry/infra-charts/tree/main/charts/tfy-configs/files/cicd-templates) ( You should ideally change only the `steps` field in each yaml file (or maybe the description). No other field should be changed.)

<CodeGroup>
  ````yaml values.yaml (truefoundry helm chart) lines theme={"dark"}
  ...
  tfy-configs:
    ...
    configs:
      cicdTemplates:
        cicd-providers.yaml: |
          - id: github
            name: GitHub
            icon: github
            enabled: true
          - id: gitlab
            name: GitLab
            icon: gitlab
            enabled: false
          - id: bitbucket
            name: Bitbucket
            icon: bitbucket
            enabled: false
        github-actions-git-source.yaml: |
          ...
        github-actions-local-source.yaml: |
          ...
        github-actions-self-build-image.yaml: |
          ...
        github-actions-git-source-patch-application.yaml: |
          ...
        github-actions-self-build-image-patch-application.yaml: |
          name: Deploy docker image on TrueFoundry with out spec in Git repository
          cicd_provider_id: github
          enabled: true
          description: "The application spec will be stored and maintained from TrueFoundry UI. The docker image is built in your CI pipeline and then the image uri is patched in the spec and deployed to truefoundry."
          deployment_mode: patch-application
          build_source: local
          recommended_environment: dev
          image_builder: self
          steps:
            - label: Generate API Key
              icon: null
              usage: Generate an API Key to authenticate and deploy applications
              type: generate-api-key
            - label: Add API Key to Github Secrets
              icon: null
              usage: null
              type: markdown-content
              args:
                content: |
                  In your GitHub Repository, navigate to **Settings > Secrets and Variables > Actions**.
                  Add a new secret called `TFY_API_KEY` and set the generated api key as value
            - label: Create GitHub Action
              icon: null
              usage: |
                Add the below workflow as `tfy-deploy.yaml` in your github workflow directory (`.github/workflows/`).
                Following GitHub Action will be triggered on each push to `main` branch
              type: markdown-content
              args:
                content: |
                  > **Note:** Please read through the `env` section and Image Build Section and update them for your registry and repo.
                
                
                  ```yaml lines
                  name: Deploy to TrueFoundry

                  on:
                    push:
                      branches:
                        - 'main'

                  permissions:
                    id-token: write
                    contents: read

                  env:
                    TFY_HOST: {{ TRUEFOUNDRY_TFY_HOST }}
                    TFY_API_KEY: $\{{ secrets.TFY_API_KEY }}
                    APPLICATION_FQN: {{ TRUEFOUNDRY_APPLICATION_FQN }}
                    
                    # Update these with your Docker Registry and Repository
                    DOCKER_REGISTRY: docker.io
                    DOCKER_REPO_NAME: $\{{ github.event.repository.name }}
                    
                    DOCKER_IMAGE_REPO: $\{{ env.DOCKER_REGISTRY }}/$\{{ env.DOCKER_REPO_NAME }}
                    DOCKER_IMAGE_TAG: $\{{ github.sha }}
                    DOCKER_IMAGE_URI: "$\{{ env.DOCKER_IMAGE_REPO }}:$\{{ env.DOCKER_IMAGE_TAG }}"

                  jobs:
                    build_deploy:
                      name: Build Image
                      runs-on: ubuntu-latest
                      timeout-minutes: 30
                      steps:
                        - name: Checkout code
                          uses: actions/checkout@v3

                        - name: Set up Docker Buildx
                          uses: docker/setup-buildx-action@v3

                        ### Image Build Section ###

                        # Build your image, push it
                        # Here is a sample, you can replace this with your registry specific steps.
                        # The registry here should be also be linked in Integrations on TrueFoundry

                        # Please see https://github.com/docker/login-action?tab=readme-ov-file#usage for examples
                        name: Login to Docker Hub
                        uses: docker/login-action@v3
                        with:
                          registry: $\{{ env.DOCKER_REGISTRY }}
                          username: $\{{ secrets.DOCKER_REGISTRY_USERNAME }}
                          password: $\{{ secrets.DOCKER_REGISTRY_PASSWORD }}
                          
                        - name: Build and push image
                          uses: docker/build-push-action@v5
                          with:
                            platforms: linux/amd64
                            context: .
                            push: true
                            tags: $\{{ env.DOCKER_IMAGE_URI }}
                            cache-from: type=registry,ref=$\{{ env.DOCKER_IMAGE_REPO }}:buildcache
                            cache-to: mode=max,image-manifest=true,type=registry,ref=$\{{ env.DOCKER_IMAGE_REPO }}:buildcache

                        ############################

                        - name: Set up Python
                          uses: actions/setup-python@v4
                          with:
                            python-version: 3.11

                        - name: Install dependencies
                          run: |
                            pip install "truefoundry<1.0.0"

                        - name: Deploy to workspace
                          run: |
                            tfy patch-application --application-fqn $\{{ env.APPLICATION_FQN }} --patch='{"image": {"image_uri": "$\{{ env.DOCKER_IMAGE_URI }}"}}'
  ````

  ```
  ```
</CodeGroup>

**Note**

* There must be a file named: `cicd-providers.yaml` in the CI/CD templates. You can use [this](https://github.com/truefoundry/infra-charts/blob/main/charts/tfy-configs/files/cicd-templates/cicd-providers.yaml) as reference to create a CI/CD providers file:
* All the keys must be file names of format `*.yaml`
* You must define all the templates you need in the `truefoundry` values.

***
