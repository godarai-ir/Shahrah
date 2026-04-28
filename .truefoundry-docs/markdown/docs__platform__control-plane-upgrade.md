> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Control Plane Upgrade

> Upgrade the control plane safely with minimal downtime.

<Note>
  **This section is applicable only if you are self-hosting the control-plane**. In case Truefoundry is hosting the control-plane, all upgrades are managed by Truefoundry automatically.
</Note>

Truefoundry control-plane is usually released twice a week to our hosted control-plane. Releases for self-hosted control-plane are usually a week behind our managed saas control-plane. All the releases can be found in our [Changelog page](https://docs.truefoundry.com/changelog).

Each version is tagged with a version number and date. It also contains the instructions to upgrade the control-plane.

<Frame caption="Release Changelog">
  <img src="https://mintcdn.com/truefoundry/9DTHMpRlYCET7n5O/images/docs/platform/release-changelog.png?fit=max&auto=format&n=9DTHMpRlYCET7n5O&q=85&s=3665c020811e7a58e8523e03585faa49" width="3840" height="2024" data-path="images/docs/platform/release-changelog.png" />
</Frame>

In case there is any breaking chnage or the values.yaml of the control-plane helm chart needs to change, it will be clearly mentioned in the changelog.

<Info>
  Truefoundry team can also help you in the upgrade process. We can come on a screen share and be available during the upgrade. **We usually recommend upgrading the control-plane atleast once a month to keep it up to date.**
</Info>

### Compute Plane Dependencies

The control-plane is compatible with older versions of the compute-plane unless mentioned in the changelog. We do however have a recommended version of the different components of the compute-plane that you should try to maintain since that is the most tested configuration that we maintain. You can track the dependencies here: [https://releases.truefoundry.com/control-plane-versions](https://releases.truefoundry.com/control-plane-versions)

This will highlight the versions of ArgoCD, GPU Operator, and other addons that the control-plane version is tested with.

<img src="https://mintcdn.com/truefoundry/9DTHMpRlYCET7n5O/images/docs/platform/releases-control-plane-dependencies.png?fit=max&auto=format&n=9DTHMpRlYCET7n5O&q=85&s=41fef8cf3442bfca5679d84ab4cdad3c" width="3840" height="1868" data-path="images/docs/platform/releases-control-plane-dependencies.png" />

### Gateway Plane Dependencies

The control-plane chart has a dependency on the gateway-plane chart. Every control-plane version is compatible with the gateway plane that is part of the control-plane chart. You can find the gateway-plane version in the [control-plane Chart.yaml file](https://github.com/truefoundry/infra-charts/blob/main/charts/truefoundry/Chart.yaml#L32).

<Note>
  Control-plane newer versions are compatible with older versions of the gateway-plane unless mentioned in the changelog. However, newer features in the control-plane might not be available in the older gateway-plane versions.
</Note>
