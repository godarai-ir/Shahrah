> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Rollout Strategy

> Configure how to rollout a new version of your application

Rolling out a new version of your application shouldn't incur downtime for your application - specially for production applications. In some cases, you might also want to test the new version of your application on a small segment of traffic before rolling out to everyone to catch unexpected issues. This is where the rollout strategy comes in.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/PSBc0bX31_cIC7pm/images/d8708bed-28c175c-image.png?fit=max&auto=format&n=PSBc0bX31_cIC7pm&q=85&s=adc3cba9d5e40a396c57fcf63f9b01c4" width="704" height="404" data-path="images/d8708bed-28c175c-image.png" />
</Frame>

Let's imagine you have 5 replicas running for v1 of your service. We want to roll out our new shiny version v2. There are a few ways we can roll out v2:

1. **Stop all 5 replicas of v1 and then bring up 5 replicas of v2.** - This will cause a downtime of the application while the 5 replicas of v2 are starting up and hence is not a great idea.
2. **Start 1 replica of v2 - when 1 replica of v2 is up, stop 1 replica of v1**. Continue the process again for the remaining 4 replicas till all 5 replicas are swapped - This makes sure there are always 5 replicas of v2 that are running and there is no disruption in quality of the service. However, there is a small duration when 6 replicas will be live since we only stop v1 replica after v2 replica is live. This strategy is called ***Rolling Update Strategy.***
3. **Stop 1 replica of v1, then start 1 replica of v2**. Continue the process till all 5 replicas of v1 are swapped. If we do this, there is a small period of time when 4 replicas are serving the traffic and hence this can cause a slight disruption in traffic, specially if all 5 replicas are fully required to serve the traffic. This strategy is called ***Rolling Update Strategy.***
4. **Bring up 1 replica of v2 - once it is up, divert 20% traffic to v2. Then stop one replica of v1, bring up another replica of v2 and divert additional 20% traffic** - hence making 40% traffic served by v2 and 60% traffic served by v1. You can continue this process till 100% traffic is served by v2 and 0% is served by v1 - This is what is called ***Canary deployment.*** . This might seem similar to the Rolling Update strategy defined above - however, in case of Canary, you can shift traffic smaller than the traffic handled by one pod, pause in between the rollout and control the traffic shift increments in a more finegrained way.
5. **Bring up 5 replicas completely of v2, shift 100% traffic to v2 and then stop the 5 replicas of v1.** - This is what is called ***BlueGreen Deployment.***

In the approaches #2 and #3, we might want to swap more then 1 replica at a time to increase the speed of rollout. Since otherwise, if there are a 1000 replicas, rolling out one replica at a time can cause the deployment to take multiple hours.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/s4Aj2_qGCrSP-zc8/images/74724764-578efdd-Screenshot_2023-11-14_at_11.14.19_PM.png?fit=max&auto=format&n=s4Aj2_qGCrSP-zc8&q=85&s=0fbc941667a94473e81957a971c66c57" width="2090" height="370" data-path="images/74724764-578efdd-Screenshot_2023-11-14_at_11.14.19_PM.png" />
</Frame>

Using TrueFoundry, you can configure the rollout strategy according to your needs. The different rollout strategies and their configuration are described below.

## Rolling Update Strategy

Rolling Update is a deployment strategy that seamlessly replaces the old version of your application with the new one, ensuring a gradual transition without disruptions. In this case, we want to be able to configure how many replicas of the old version get replaced with the new version and whether we create new replicas of the new version before shutting off the replicas of the older version. This can be configured using the following parameters:

* **Max Surge(%)**: This defines the max number of replicas of the older and newer version of the application that can be live at any point during the rollout. This is expressed as a percentage of the current number of the replicas. If the current number of replicas is 10 and we set max surge to 20%, this means that 12 replicas can at max be live any point. What this implies is the system will bring up 2 replicas of v2 before shutting down 2 v1 replicas.
* **Max Unavailable(%)**: This defines how many minimum replicas (old + new version combined) should be available during the rollout. It is expressed as a percentage of the current number of the replicas.. For example, if replica count is set to 12 for the deployment and max unavailable is set to 25%, then minimum of 9 replicas need to be live at any point during the rollout. What this means is the system can bring down 3 replicas of v1 before starting the 3 replicas of v2.

Let's assume we are upgrading the deployment from v1 to v2 assuming both Max surge and max unavailable are set to 25% and initial replicas are 12.

| Step                                                                                             | v1 Replicas | v2 Replicas | Total Replicas |
| ------------------------------------------------------------------------------------------------ | ----------- | ----------- | -------------- |
| Before rollout                                                                                   | 12          | 0           | 12             |
| It will bring down 3 v1 replicas because minimum available replica(max unavailable) should be 9. | 9           | 0           | 9              |
| It will bring up 6 v2 replicas because maximum available replicas(max surge) can be 15           | 9           | 6           | 15             |
| Once all v2 replicas are available then it will again bring down 6 v1 replicas                   | 3           | 6           | 9              |
| It will bring up 3 v2 replicas because the total should be 12 at rollout completion              | 0           | 12          | 12             |
| Rollout Complete                                                                                 | 0           | 12          | 12             |

<Tip>
  A good default startegy is to keep max\_unavailable at 10% and max\_surge at 10%. If you are working with GPU machines and don't want additional machines to come up during rollout, you can set max\_surge to 0% and max\_unavailable to 20%.
</Tip>

### Configuring Rolling Update for an Application

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/QL01FrhiG6rgt2Jm1zTwM" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/QL01FrhiG6rgt2Jm1zTwM" style={{ border: "none", display: "flex", margin: "auto" }} />

One limitation that we have in rolling update is that we cannot control how much traffic goes to v1 and v2. This is where Canary rollout comes into play.

## Canary

Canary deployment is a cautious approach to rolling out updates where a small subset of users gets to try the new version first before it's made available to everyone. It is ideal for minimizing risks associated with major updates to test the new deployment before a full rollout.

Some key terms to get comfortable with before we can understand this strategy.

1. **Canary Steps:** You can have multiple stages/steps in your canary rollout. You get absolute control control at each step regarding how much traffic should be redirected to the new version and should it automatically move to the next step after a certain time interval or a manual trigger is needed to move it to the next step.
2. **Set traffic weight to:** This is the percentage of traffic that should be redirected to the new version at each stage.
3. **Pause for:** This helps to define if the rollout should move to next step after a certain time or a manual trigger is needed.

Let's assume we are upgrading the deployment from v1 to v2, assuming we want to:

1. Deploy total 10% traffic to new version and wait for 60 seconds.
2. Deploy total 30% traffic to new version and wait until we manually trigger to next step.
3. Deploy total 80% traffic to new version and wait for 30 seconds.

| Steps                       | v1 Traffic(%) | v2 Traffic(%) | Pause For(seconds) |
| --------------------------- | ------------- | ------------- | ------------------ |
| Before rollout              | 100%          |               | NA                 |
| 1 - Automatic               | 90%           | 10%           | 60                 |
| 2 - Automatic               | 70%           | 30%           | Pause Indefintely  |
| 3- Manual trigger next step | 20%           | 80%           | 30                 |
| Rollout complete            | 0%            | 100%          | NA                 |

### Configuring Canary Rollout for an Application

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/v7Be0iLjE654nt7i_qc1_" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/v7Be0iLjE654nt7i_qc1_" style={{ border: "none", display: "flex", margin: "auto" }} />

## Blue-Green:

Blue-Green is a strategy that ensures a seamless transition between different versions by maintaining two identical environments: one for the current version (blue) and one for the new version (green). It minimises downtime and provides an instant fallback mechanism in case issues arise in the new version.

Let's assume we are upgrading the deployment from v1(blue) to v2(green):

1. We will create v2 as a full clone of v1. If there were 12 replicas in v1 then we create exact 12 replicas in v2.
2. Once all the pods are ready in v2, we move full traffic to v2.
3. This moving traffic from v1 to v2 can be automatic or it can be manual depending on what we configure.

For configuring Blue Green, you can adjust the following settings:

* **Auto-promotion:** Toggle this setting to automatically promote the new version to active once it passes all tests and checks.
* **Auto-promotion seconds:** Specify the duration, in seconds, to wait before automatically promoting the new version to active.

<Warning>
  One thing to be cautious of while using blue-green deployment is that you need to have the resources to bring up double the number of replicas. If you had 12 replicas in v1 then 12 v2 replicas are brought up - so we need to be able to bring up 24 replicas at once. While cloud can scale easily for CPU machines, this assumption might not hold true for GPU machines and on-prem infrastructure.
</Warning>

### Configuring Blue Green Rollout for an application

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/XvZhbnkM2eN6Pf1kL3-ML" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/XvZhbnkM2eN6Pf1kL3-ML" style={{ border: "none", display: "flex", margin: "auto" }} />
