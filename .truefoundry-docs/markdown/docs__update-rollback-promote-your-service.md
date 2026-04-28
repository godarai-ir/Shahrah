> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Update, Rollback, Promote

> Update service versions, rollback to previous deployments, and promote services between environments safely.

During the lifecycle of an application, we will usually encounter the scenarios of upgrading the version of the application, rolling back the active version or promoting the application from dev to staging and then to production. TrueFoundry faciliates these flows and categorizes them into 3 workflows: Updating, Rolling back and Promoting.

* **Updating**: Modifying the existing code or configuration of a service to introduce new features, fix bugs, or enhance performance.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/s4Aj2_qGCrSP-zc8/images/8d492bdf-87e108a-Untitled-2023-10-03-0731-15.png?fit=max&auto=format&n=s4Aj2_qGCrSP-zc8&q=85&s=01a9c6d1bd48cffe3b0fabb9dfbaa8c8" width="1233" height="474" data-path="images/8d492bdf-87e108a-Untitled-2023-10-03-0731-15.png" />
</Frame>

* **Rolling Back**: Reverting a service to a previous, known-good state in case of unexpected issues or performance degradation.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/s4Aj2_qGCrSP-zc8/images/7b52f181-2409fa7-Untitled-2023-10-03-0731-16.png?fit=max&auto=format&n=s4Aj2_qGCrSP-zc8&q=85&s=4e524446ea49b8dd2a4ae1bcded5d36c" width="2652" height="669" data-path="images/7b52f181-2409fa7-Untitled-2023-10-03-0731-16.png" />
</Frame>

* **Promoting**: Moving a service from a lower environment (e.g., development or staging) to a higher environment (e.g., production) for being consumed by end users.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/s4Aj2_qGCrSP-zc8/images/7e18c922-528ae27-Untitled-2023-10-03-0731-19.png?fit=max&auto=format&n=s4Aj2_qGCrSP-zc8&q=85&s=4a756acba2157ca45f054560d0e042ad" width="2139" height="1066" data-path="images/7e18c922-528ae27-Untitled-2023-10-03-0731-19.png" />
</Frame>

## Updating Services

To update a service in TrueFoundry, simply deploy the service with the desired changes, maintaining the same service name and workspace. TrueFoundry will automatically detect the update and create a new version.

Alternatively, you can update your Machine Learning Service using the **`Edit`** button to modify your Service's configuration, following the instructions below. This will also create and deploy a new version with your changes.

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/uag-o124hmbQiKYqisyTL" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/uag-o124hmbQiKYqisyTL" style={{ border: "none", display: "flex", margin: "auto" }} />

## Rolling Back Services

TrueFoundry retains a comprehensive history of your service's deployed versions. You can easily redeploy a previous stable state, following the steps outlined in the demo below:

This action will create a new version with the configuration of the redeployed version, effectively rolling back the service to its previous state. For e.g. if the current version is v10, and you rollback to v5, a new version v11 will be created with the same spec as v5.

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/1q4n5IQ9MPNpYNgIKqssG" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/1q4n5IQ9MPNpYNgIKqssG" style={{ border: "none", display: "flex", margin: "auto" }} />

## Promoting Services

Usually, there will be multiple environments for an application for different stages of testing. While most companies maintain 3 environments: dev, staging, and production, your company might have fewer or more environments depending on the requirements. Independent of the number of environments, the problem statement of promoting an application from one environment to the other remains the same. Let's say we are promoting a service from staging to production. A few things to keep in mind are:

1. The docker image URI of current production should be overwritten by the image URI currently running in staging.
2. Settings like autoscaling and replicas in production shouldn't be overwritten with what is there in staging.
3. Settings like readiness probe, and liveness probe should be changed with what is running actively in staging.
4. We need to make sure environment variables are consistent among staging and production. If a new environment variable is added to staging, and we forget to add it to production while promoting, it can be disastrous.

TrueFoundry promote flow takes these factors into account and shows the correct diff to the user while promoting to avoid manual errors.

Let's say you've developed an initial version of a ML service for an ongoing project. This service has been deployed to your team's development workspace. After making some refinements, your development service is ready to move to production. The production environment can be another workspace on the same cluster or on a different cluster. To promote this to a production environment, make sure you have the production workspace ready and then follow the steps below:

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/2rfq33G4ipWYNebMvAcr_" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/2rfq33G4ipWYNebMvAcr_" style={{ border: "none", display: "flex", margin: "auto" }} />

<Info>
  ### Please remember to make modifications in the prod service after promoting

  1. **Environment Variables**: You might need to modify the environment variables to access different DBs, etc.
  2. **Resource Scaling**: To handle production traffic, you might need to change the number of replicas or configure autoscaling.
</Info>

The journey doesn't end once the service is deployed in production. You will constantly make changes to the service in development environment and then promote again when you want to upgrade production again.

When you promote again, TrueFoundry will analyze the configuration differences between your enhanced development service and the existing production service. It will highlight the things that are going to be changed in the current production environment like docker image uri, environment variables, liveness/readiness probes and ignore the parts like autoscaling that are going to be retained in the production service. This helps avoid manual errors during the promote flow.

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/ad1x0M1SL0HfB_Syx3kG- " typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/ad1x0M1SL0HfB_Syx3kG- " style={{ border: "none", display: "flex", margin: "auto" }} />

By leveraging Promote Flow, you can confidently introduce enhancements to your production service without compromising its stability and security.

***
