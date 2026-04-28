> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Github

> Create a GitHub App and configure the integration with TrueFoundry for repository access and CI/CD.

## Creating the Github App

To enable GitHub integrations, you need to create a Github app. Follow these steps to create the Github app and integrate it with truefoundry:

**You need to have the admin access to create a Github app**

1. In the upper-right corner of any page on GitHub, click your profile photo.
2. Navigate to your account settings.
   * For an app owned by a personal account, click **Settings**.
   * For an app owned by an organization:
     1. Click **Your organizations**.
     2. To the right of the organization, click **Settings**.
3. In the left sidebar, click **Developer settings**.
   <Frame caption="">
     <img src="https://mintcdn.com/truefoundry/ZH5BEiH2UvJ7x_p-/images/Screenshot-2025-05-13-at-2.17.00-PM.png?fit=max&auto=format&n=ZH5BEiH2UvJ7x_p-&q=85&s=82e755ad3826cb765b475d494b00d849" width="2610" height="816" data-path="images/Screenshot-2025-05-13-at-2.17.00-PM.png" />
   </Frame>
4. In the left sidebar, click **GitHub Apps**
   <Frame caption="">
     <img src="https://mintcdn.com/truefoundry/ZH5BEiH2UvJ7x_p-/images/Screenshot-2025-05-13-at-2.42.29-PM.png?fit=max&auto=format&n=ZH5BEiH2UvJ7x_p-&q=85&s=86dc45e0e14b3bd2583130f424aeffea" width="2380" height="844" data-path="images/Screenshot-2025-05-13-at-2.42.29-PM.png" />
   </Frame>
5. Now click on **New GitHub App**
6. Fill in the details as follows:
   * App name: TrueFoundry
   * Homepage URL: [https://control-plane-url](https://control-plane-url) For example [https://control-plane.truefoundry.com](https://control-plane.truefoundry.com)
     <Frame caption="">
       <img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/github-integrations-form-1.png?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=8fb6f85ee81c57cbcf8bd081f88e7f99" width="2126" height="1112" data-path="images/github-integrations-form-1.png" />
     </Frame>
   * Callback URL: [https://control-plane-url/api/svc/v1/vcs/github/callback](https://control-plane-url/api/svc/v1/vcs/github/callback)
     <Frame caption="">
       <img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/github-integration-form-2.png?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=fff12122b92ae2109e1bdd925505d762" width="2280" height="832" data-path="images/github-integration-form-2.png" />
     </Frame>
   * Setup URL: [https://control-plane-url/api/svc/v1/vcs/github/callback](https://control-plane-url/api/svc/v1/vcs/github/callback)
     <Frame caption="">
       <img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/github-integration-form-3.png?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=ea343b006bb2b47f5e8f4f19ba243b50" width="2070" height="504" data-path="images/github-integration-form-3.png" />
     </Frame>
   * Webhook URL: [https://control-plane-url/api/svc/v1/vcs/github/webhook-callback](https://control-plane-url/api/svc/v1/vcs/github/webhook-callback)
     <Frame caption="">
       <img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/github-integration-form-4.png?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=2364afe1acd0b92682fe966ba7c51232" width="2306" height="902" data-path="images/github-integration-form-4.png" />
     </Frame>
     * Select the `Enable SSL Verification` checkbox
   * Now in the permissions section, Give the read only access to contents in the Repository permissions
     <Frame caption="">
       <img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/github-integration-form-5.png?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=a24d46cac885473ef45b0ebfa073e126" width="2128" height="252" data-path="images/github-integration-form-5.png" />
     </Frame>
   * Now in the `Where can this GitHub App be installed?` section, Select `Any Account`
     <Frame caption="">
       <img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/github-integration-form-6.png?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=7889cc0ca6457005bb3c16fb898e7ecf" width="2328" height="468" data-path="images/github-integration-form-6.png" />
     </Frame>
7. Now generate a new client secret and copy it
   <Frame caption="">
     <img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/github-integration-form-7.png?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=9085c46667acf3c76362a4861da66764" width="1914" height="346" data-path="images/github-integration-form-7.png" />
   </Frame>
8. Now generate a new private key(section present at the bottom) and copy it
   <Frame caption="">
     <img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/github-integration-form-8.png?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=f6410f9cb8ad33b24a660b29906ff55b" width="1932" height="838" data-path="images/github-integration-form-8.png" />
   </Frame>
9. Now provide AppId, AppName, AppURL, Private Key(base64 encode) as env to servicefoundry-server

```
GITHUB_INSTALLATION_URL=https://github.com/apps/<app-name>/installations/new
GITHUB_PRIVATE_KEY=<private-key-in-base64-format>
GITHUB_APP_ID="<app-id>" # Please note to add quotes to this value
GITHUB_PAT=<github-personal-access-token>
```
