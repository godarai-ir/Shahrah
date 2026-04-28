> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Gitlab

> Create a GitLab application and configure the integration with TrueFoundry for repository access and CI/CD.

## Creating the Gitlab App

To enable Gitlab integrations, you need to create a Gitlab app. Follow these steps to create the Gitlab app and integrate it with truefoundry:

* Sign in to GitLab.com. On the left sidebar, select your avatar and select Edit profile.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/s4Aj2_qGCrSP-zc8/images/7e2ec34e-5ce0fa2-first.png?fit=max&auto=format&n=s4Aj2_qGCrSP-zc8&q=85&s=496a9613a891dcda45fbe948adc402c5" width="1607" height="761" data-path="images/7e2ec34e-5ce0fa2-first.png" />
</Frame>

* On the left sidebar, select **Applications**. Click on the **Add new application** button right side.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/s4Aj2_qGCrSP-zc8/images/77a6157f-71048c1-second.png?fit=max&auto=format&n=s4Aj2_qGCrSP-zc8&q=85&s=6d4228455b71e96157160aa3e3832e8b" width="1900" height="568" data-path="images/77a6157f-71048c1-second.png" />
</Frame>

* Provide the required details for **Add new application**.

  * Name: This can be anything. Consider something like `<Organization>`'s GitLab or `<Your Name>`'s GitLab or something else descriptive.
  * Redirect URI: `https://app.example-org.truefoundry.com/api/svc/v1/vcs/gitlab/callback`
  * Enable the **Confidential** check box.
  * Select the following scopes: read\_api, read\_user, read\_repository, write\_repository.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/OHzlp6GY5G-JfKle/images/c89809e5-4b2eef2-third.png?fit=max&auto=format&n=OHzlp6GY5G-JfKle&q=85&s=feab5ac0f3b2e1334189f30605fe4646" width="971" height="877" data-path="images/c89809e5-4b2eef2-third.png" />
</Frame>

* Select **Save application**.
* You should now see an **Application ID** and **Secret**.

<Frame caption="List of connected repositories">
  <img src="https://mintcdn.com/truefoundry/FrY4JbiyZud2He3p/images/16852009-bbb451d-fourth.png?fit=max&auto=format&n=FrY4JbiyZud2He3p&q=85&s=bd94a00f766b50c44b946dd01cf8dbf4" width="1896" height="873" data-path="images/16852009-bbb451d-fourth.png" />
</Frame>

List of connected repositories

* Set environment variables in servicefoundry server :
  ```
  GITLAB_APP_ID: ''
  GITLAB_APP_SECRET: ''
  GITLAB_SCOPE: 'read_api read_user read_repository write_repository'
  ```
