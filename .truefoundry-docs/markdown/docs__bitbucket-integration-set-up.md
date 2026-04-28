> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Bitbucket

> Create a Bitbucket OAuth Consumer app and configure the integration with TrueFoundry for repository access.

## Creating the Bitbucket App

To enable Bitbucket integration in your app, you need to create a Bitbucket OAuth Consumer app using your BitBucket account. This Bitbucket app can then be authenticated to access repositories of users' account.

Follow these steps to create the Bitbucket OAuth Consumer app and integrate it with truefoundry:

* Sign in to [bitbucker.org](https://bitbucket.org/). Navigate to Workspace Settings(preferably for your organization).

* Navigate to `OAuth Consumers` under `Apps and features`.

  <Frame caption="">
    <img src="https://mintcdn.com/truefoundry/4MAaF__cLD4iud16/images/56347ae8-82b368d54aeb969f578d004f1b17f0cc49b34ef4126d6cbf7c205a5ed7778f63-image_32.png?fit=max&auto=format&n=4MAaF__cLD4iud16&q=85&s=e63ac341d535c3023dec352265b52d6e" width="3016" height="1716" data-path="images/56347ae8-82b368d54aeb969f578d004f1b17f0cc49b34ef4126d6cbf7c205a5ed7778f63-image_32.png" />
  </Frame>

* Click the button to `Add consumer`

* Fill in the following details:

  * Name: This can be anything. Consider something like \<Organization>'s Bitbucket or \<Your Name>'s Bitbucket or something else descriptive.

  * Redirect URI: `https://app.example-org.truefoundry.com/api/svc/v1/vcs/bitbucket/callback` (replace `app.example-org.truefoundry.com` with your control plane URL)

  * Select the following permissions:

    * Account: Email, Read
    * Projects: Read
    * Repositories: Read, Write(to allow adding CI/CD PR)
    * Pull requests: Read, Write(to allow adding CI/CD PR)

    <Frame caption="">
      <img src="https://mintcdn.com/truefoundry/OHzlp6GY5G-JfKle/images/c1a50fbc-b031c8d81c2204e52083915f90aa92c5b9efa75ba2e4f4804a923ddf16a838af-image.png?fit=max&auto=format&n=OHzlp6GY5G-JfKle&q=85&s=692bf8af4810ac1dbccacd4428b239d8" width="1254" height="818" data-path="images/c1a50fbc-b031c8d81c2204e52083915f90aa92c5b9efa75ba2e4f4804a923ddf16a838af-image.png" />
    </Frame>

* Apply. Copy the `Key` and `Secret` from created app.

* Set environment variables in servicefoundryServer :\
  BITBUCKET\_CLIENT\_ID=\$`{Key}` BITBUCKET\_CLIENT\_SECRET=\$`{Secret}`

* If using CI/CD for control plane deployment follow below:

  * Add these keys and values in the `truefoundry-creds` k8s secret present in `truefoundry` namespace.\
    BITBUCKET\_CLIENT\_ID=\$`{Key}` BITBUCKET\_CLIENT\_SECRET=\$`{Secret}`
  * In the truefoundry application values file, use the value as `${k8s-secret/truefoundry-creds/BITBUCKET_CLIENT_ID}` and `${k8s-secret/truefoundry-creds/BITBUCKET_CLIENT_SECRET}`

### Link Bitbucket

Bitbucket is linked via OAuth. This will allow use of any repository you have access to. We can easily link our Bitbucket account in few steps as mentioned below:

* Click on the `Link Bitbucket` button present in the top right corner of the Bitbucket tab for linking your Bitbucket account as shown below:

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/4MAaF__cLD4iud16/images/5fbb4e00-2c8634c-image.png?fit=max&auto=format&n=4MAaF__cLD4iud16&q=85&s=2b6ac3981f5c2ec39c74275f1ddf59e4" width="1920" height="908" data-path="images/5fbb4e00-2c8634c-image.png" />
</Frame>

* Your bitbucket account which is signed in to your system will get linked. You will now be able to see your account on the dashboard as shown below:

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/e6ea7604-df6704b-image.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=d467cbcc48493d2cb7acf90d8af529b8" width="1920" height="952" data-path="images/e6ea7604-df6704b-image.png" />
</Frame>

* You can also see the **connected repositories** of your account by clicking on `⚙️` button on the right side of that particular account.
* **Note** : In this connected repositories section you will only see the repositories that you have admin access. Only these repositories can be used to make deployment on the platform.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/OHzlp6GY5G-JfKle/images/c2363bf2-65774ac-image.png?fit=max&auto=format&n=OHzlp6GY5G-JfKle&q=85&s=337652789e726d6b1dce533900edb120" width="1920" height="944" data-path="images/c2363bf2-65774ac-image.png" />
</Frame>

* You can also **unlink your account** by `unlink` button. It can be seen in the above figure.
