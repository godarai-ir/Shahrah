> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Manage Teams

> Detailed guide to how to add, manage and delete teams in Truefoundry

Teams allow Admins to group multiple users together and ease out the role assignment in resources. Teams can be managed manually or in an automated way by setting up groups in Single Sign-On (SSO) with your Identity Provider (IdP).

### Create Team and add members

Admins can create manual teams and manage the members. Admins can also add `managers` to team with access to further manage the team members themself. To create a team in TrueFoundry, follow these steps:

<iframe href="https://app.supademo.com/embed/Sz73nKcCDHYkQIBfifhg5" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/Sz73nKcCDHYkQIBfifhg5" style={{ border: "none", display: "flex", margin: "auto" }} />

### Grant Access of a Resource to Team

Once a team has been created, you can grant it access to resources. To do so, follow these steps:

<iframe href="https://app.supademo.com/embed/YJv0iwe-AWSiiAUY8hss8" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/YJv0iwe-AWSiiAUY8hss8" style={{ border: "none", display: "flex", margin: "auto" }} />

## Manage Teams via Gitops

Admins can manage teams programmatically by using the `tfy apply` command on the team YAML file. You can get the instructions by clicking on the `Apply Using YAML` in the Edit form.

<img src="https://mintcdn.com/truefoundry/2fPZGuuvhgm5F-Rk/images/docs/platform/team-apply-using-yaml-button.png?fit=max&auto=format&n=2fPZGuuvhgm5F-Rk&q=85&s=cd6d36485de0350d445eab5364a07916" alt="Apply Using YAML button" width="3024" height="1642" data-path="images/docs/platform/team-apply-using-yaml-button.png" />

<img src="https://mintcdn.com/truefoundry/2fPZGuuvhgm5F-Rk/images/docs/platform/team-apply-using-yaml-instructions.png?fit=max&auto=format&n=2fPZGuuvhgm5F-Rk&q=85&s=40701f224e811ea992e7d78493faa1fd" alt="Apply Using YAML instructions" width="3024" height="1642" data-path="images/docs/platform/team-apply-using-yaml-instructions.png" />

## Manage Teams via API

You can find all the team related APIs in the [TrueFoundry SDK documentation](/docs/truefoundry_sdk/teams).
