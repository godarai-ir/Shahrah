> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Set Up CI/CD

> Configure CI/CD pipelines for your services using GitHub Actions, GitLab CI/CD, Bitbucket, or Jenkins.

Setting up CI/CD enables your code to be deployed automatically whenever you push changes to your main branch. TrueFoundry integrates seamlessly with popular CI/CD systems like GitHub Actions, GitLab CI/CD, Bitbucket Pipelines, and Jenkins.

<Note>
  TrueFoundry doesn't provide CI capabilities - it integrates with external
  CI/CD systems. The build process always happens on TrueFoundry to save cost
  and time.
</Note>

## How it works

TrueFoundry's CI/CD integration is designed for simplicity and efficiency:

* **Automatic builds**: All Docker images are built on TrueFoundry's infrastructure, which is 3-15X faster than traditional CI pipelines
* **Cost optimization**: No build costs in your CI pipeline - everything happens on TrueFoundry
* **Simple decision**: You only need to decide whether to keep the TrueFoundry YAML spec in your Git repository
* **Automatic deployment**: Every commit to your main branch triggers an automatic deployment

<Tip>
  TrueFoundry can build images 3-15X faster than your CI pipeline. You can read
  [this
  blog](https://www.truefoundry.com/blog/enabling-3-15x-faster-docker-image-builds-with-truefoundry-on-kubernetes)
  to learn more about the performance benefits.
</Tip>

## Key decision: YAML spec location

The only decision you need to make is whether to keep the TrueFoundry YAML spec in your Git repository:

**Keep YAML spec in Git** (Recommended)

* Full configuration versioning and tracking
* Complete control over service configuration

**Don't keep YAML spec in Git**

* Simpler setup and maintenance
* Configuration changes via TrueFoundry UI

<Warning>
  If you keep the YAML spec in Git and deploy via CI/CD, avoid changing
  configuration via the TrueFoundry UI. Any UI changes will be overwritten by
  the Git configuration on the next deployment.
</Warning>

{" "}

## Setting up CI/CD

TrueFoundry provides a simple setup process that generates the appropriate CI/CD configuration based on your preferences.

<Steps>
  <Step title="Click on the three dots on the top right corner of the application page and select Setup CI/CD">
    <img src="https://mintcdn.com/truefoundry/FrY4JbiyZud2He3p/images/060e6452-2761c700216d55134101b245a63e507f003d66803014256cd5b5939719ad56de-image.png?fit=max&auto=format&n=FrY4JbiyZud2He3p&q=85&s=dddb35d46a16f40cc173eea1c9c3f8b2" alt="Setup CI/CD button with provider selection options" width="1439" height="360" data-path="images/060e6452-2761c700216d55134101b245a63e507f003d66803014256cd5b5939719ad56de-image.png" />
  </Step>

  <Step title="Choose your CI/CD provider and decide whether to keep the YAML spec in your Git repository">
    <img src="https://mintcdn.com/truefoundry/kIGYcCynaeFa2Exr/images/ci_cd_option_selection_form.png?fit=max&auto=format&n=kIGYcCynaeFa2Exr&q=85&s=2e4c7f524820073e53cbbc8d30e361e7" alt="YAML spec location decision interface" width="1642" height="720" data-path="images/ci_cd_option_selection_form.png" />

    This is the only decision you need to make. TrueFoundry will automatically generate
    the appropriate CI/CD configuration based on your choice.
  </Step>

  <Step title="Follow the generated setup steps based on your repository platform">
    TrueFoundry will provide you with:

    * API key generation instructions, click on the button to generate the API key and copy it.
          <Frame caption="">
            <img src="https://mintcdn.com/truefoundry/kIGYcCynaeFa2Exr/images/ci_cd_api_key_gen_step.png?fit=max&auto=format&n=kIGYcCynaeFa2Exr&q=85&s=7b1a9288268a29823b793d5185c0a0a2" width="1568" height="352" data-path="images/ci_cd_api_key_gen_step.png" />
          </Frame>
    * GitHub secrets configuration, add the API key to the GitHub secrets for the repository.
          <Frame caption="">
            <img src="https://mintcdn.com/truefoundry/mFhPWcFYt_ds3OC2/images/add_api_key_to_github_ci_cd.png?fit=max&auto=format&n=mFhPWcFYt_ds3OC2&q=85&s=f498b3d104e8d39bc597fa35e7762ba4" width="1526" height="296" data-path="images/add_api_key_to_github_ci_cd.png" />
          </Frame>
    * Download the TrueFoundry specification file(if you have selected yes to keep the YAML spec in your Git repository)
          <Frame caption="">
            <img src="https://mintcdn.com/truefoundry/KCz-JU1lrb2tDqiE/images/download_tfy_spec_for_ci_cd.png?fit=max&auto=format&n=KCz-JU1lrb2tDqiE&q=85&s=346836f32bf3a048331442bc8ef5bae4" width="1548" height="400" data-path="images/download_tfy_spec_for_ci_cd.png" />
          </Frame>
    * Copythe CI/CD workflow or action file and store it in the repository as written in the step based on your platform.
          <Frame caption="">
            <img src="https://mintcdn.com/truefoundry/kIGYcCynaeFa2Exr/images/ci_cd_file_download.png?fit=max&auto=format&n=kIGYcCynaeFa2Exr&q=85&s=d75347c160030ab018c175fb87b3ac96" width="1566" height="1248" data-path="images/ci_cd_file_download.png" />
          </Frame>
  </Step>
</Steps>

## Required Secrets

Make sure to configure these secrets in your CI/CD system:

| Secret Name                | Description                            | Required For            |
| -------------------------- | -------------------------------------- | ----------------------- |
| `TFY_API_KEY`              | TrueFoundry API key for authentication | All deployments         |
| `DOCKER_REGISTRY_USERNAME` | Docker registry username               | Image-based deployments |
| `DOCKER_REGISTRY_PASSWORD` | Docker registry password/token         | Image-based deployments |
