> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# User, Team, Account and Tenant Management

> A high level introduction to Truefoundry's account management hierarchy and key concepts.

Truefoundry provides a hierarchical structure for managing users, teams, resources, and access control. This structure is designed to offer flexibility as well as security for companies of various sizes.

<img src="https://mintcdn.com/truefoundry/Gz34IEy99_S9XhHC/images/docs/platform/user-team-account-management.png?fit=max&auto=format&n=Gz34IEy99_S9XhHC&q=85&s=4667a014ca9a60ed6eeaf373b67e789b" width="3417" height="2372" data-path="images/docs/platform/user-team-account-management.png" />

The key entities in the hierarchy above are:

1. **Control Plane**: This is one deployment of the control-plane software. In the SAAS hosted version of Truefoundry, this is already managed by us. In case you are self-hosting the control-plane, you will get your own instance and can create the tenants under that.

2. **Tenant**: This is a top level entity in the control-plane. Different tenants are completely isolated accounts, each with their own users, teams, roles, SSO settings, etc.

<Warning>
  There is no interaction between the data in different tenants.

  **Usage in hosted version**: This is used to separate the the different customers on our hosted control-plane.

  **Usage in self-hosted version**: This tenant wide isolation is less useful and should only be used if there are entirely different units within the organization - each with their own platform and IT teams.
</Warning>

All the resources can be created directly under the tenant or under an account within a tenant.

3. **Tenant-Level Resources**: These are the resources that can be created directly under the tenant. These resources can be used and managed by teams/users/virtual accounts at the tenant level. In most cases, just creating resources at the tenant level and alloting them to different teams/users is enough and we don't need to create accounts within a tenant.

4. **Account**: A tenant-level admin can create multiple accounts and assign users to it with a certain set of permissions. Each account has settings that can be configured by the account creator. Each account will have an admin which can grant permission to other users within that account. The permission set will be a subset of the permissions granted to the account admin.

A user belonging to an account will have access to the following resources:

* Resources within the same account which this user has access to.
* Resources within the tenant that this user has access to.
* This user cannot be granted access to any resources that belong to another account of which this user is not a member of.

The resources within the tenant and account are as follows:

<img src="https://mintcdn.com/truefoundry/Gz34IEy99_S9XhHC/images/docs/platform/hierarchy-with-resources.png?fit=max&auto=format&n=Gz34IEy99_S9XhHC&q=85&s=79515201c0f13d201e0df1fb7f120c43" width="5552" height="3520" data-path="images/docs/platform/hierarchy-with-resources.png" />

This hierarchy enables the following functionality:

1. **Tenant-level resources can be shared with certain accounts**: This allows platform teams to create some central resources that can be used by different teams within the organization. A few examples of such scenarios are:

   * AI Engineering: Tenant-admin (Platform) team creates a Kubernetes cluster and allots different workspaces in that cluster to different accounts. For example account1/team1 has admin access to tenant-cluster/workspace1, account1/team2 has admin access to tenant-cluster/workspace2, etc.
   * AI Gateway: Tenant-admin (Platform) team creates a model provider account and allots different models to different accounts. For example account1/team1 has access to model1, account1/team2 has access to model2, etc.

2. **Every account can create and manage their own resources**: This allows flexibility to different teams to create and manage their own resources within their account. A good example of this is each team managing their own prompts.

3. **Billing Isolation**: The usage from each account will go into the billing for that account. This allows different teams to have their own budgets and usage limits.

There are three primary user entities in Truefoundry:

1. **Users**: These are the individuals who have access to the platform. They have a unique email address, and can be deactivated if the employee leaves the organization.

<Note>
  Users can belong to the overall tenant(root account) or just be scoped to an account.
</Note>

2. **Teams**: These are groups of users which can be added together to a resource. For example, you can add a team as a viewer to a cluster, admin to a workspace, etc. Using teams, you can just grant access of resources to the team and then just manage the members within the team.

<Note>
  Teams can belong to the overall tenant(root account) or just be scoped to an account.
</Note>

3. **Virtual Accounts**: Virtual accounts are not mapped to any user. They can be created by tenant-admins or account-admins to provide access to resources for applications / code. For e.g., if your code wants to access a Truefoundry API or model in the AI Gateway layer, it will need a valid token to access the API. In this case, we should not provide a user token, since it will be deactivated if the user leaves the organization.

<Note>
  Virtual accounts can belong to the overall tenant(root account) or just be scoped to an account.
</Note>

## FAQ

<AccordionGroup>
  <Accordion title="Can I create multiple tenants if I am self-hosting the control-plane?">
    We do have the support of creating multiple tenants if you are self-hosting the control-plane. However, this is not enabled by default and you will need to talk to Truefoundry team to enable this. This also falls in a different pricing tier and should only be needed in very large organizations that have entirely different platform/IT teams.
  </Accordion>

  <Accordion title="Can I just create resources and allot them to different teams/users at the tenant level? Why do I need accounts?">
    Yes, in most cases, just creating resources at the tenant level and alloting them to different teams/users is enough and we don't need to create accounts within a tenant. This is the simplest and most common use case. This can easily model 10-100 teams within a company if they are relatively in a flat hierarchy.
    In this case, if you give permissions to some of these teams to also create resources, its possible to give access of resources created by one of these teams to some other team. While this is flexible, it might sometimes not be desirable where in every team needs to work in an isolated way. Creating accounts within a tenant ensures that cross-sharing is not possible from the platform itself. If you just operate in teams within the tenant, you have to manually ensure the isolation.
  </Accordion>

  <Accordion title="Can I create multiple tenants in the hosted version of control-plane?">
    You can by just signing up multiple times on the hosted control-plane. Each of these will be a separate tenant and you will have to login to them separately. There is no global view of both the tenants you will get in one place.
  </Accordion>
</AccordionGroup>
