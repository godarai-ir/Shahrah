> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Manage User Roles & Permissions

> Configure custom roles with scoped permissions for AI Gateway, MCP servers, agents, and platform resources.

TrueFoundry offers custom roles to be assigned to users with scoped permissions in the platform. These roles allow fine-grained access control over various resources and actions.

Key aspects of custom roles:

1. Each role can have multiple permissions across different resource types
2. Permissions define what actions can be performed on specific resources
3. Roles can be assigned to users to grant them the associated permissions

### Available Permissions

#### AI Gateway

<AccordionGroup title="AI Gateway">
  <Accordion title="Provider Account (Model/Virtual Model/Guardrail Group)">
    | Permission              | Description                                                 |
    | :---------------------- | :---------------------------------------------------------- |
    | Create Provider Account | Allows creation of new provider accounts                    |
    | Delete Provider Account | Allows deleting existing provider accounts                  |
    | Manage Provider Account | Allows updating provider account configuration              |
    | Read Provider Account   | Allows viewing provider account details                     |
    | Use Integrations        | Allows using integrations associated with provider accounts |
  </Accordion>

  <Accordion title="MCP/Virtual MCP Server">
    | Permission        | Description                              |
    | :---------------- | :--------------------------------------- |
    | Create MCP Server | Allows creation of new MCP servers       |
    | Delete MCP Server | Allows deleting existing MCP servers     |
    | Manage MCP Server | Allows updating MCP server configuration |
    | Read MCP Server   | Allows viewing MCP server details        |
    | Use MCP Server    | Allows using MCP server for operations   |
  </Accordion>

  <Accordion title="Agent">
    | Permission    | Description                     |
    | :------------ | :------------------------------ |
    | Create Agent  | Allows creation of new Agents   |
    | Delete Agents | Allows deleting existing Agents |
    | Manage Agents | Allows updating Agents          |
    | Read Agents   | Allows viewing Agents details   |
  </Accordion>

  <Accordion title="Gateway Controls">
    | Permission              | Description                                              |
    | :---------------------- | :------------------------------------------------------- |
    | List Gateway Controls   | Allows listing and viewing gateway controls              |
    | Manage Gateway Controls | Allows creating, updating, and deleting gateway controls |
  </Accordion>
</AccordionGroup>

#### AI Engineering

<AccordionGroup title="AI Engineering">
  <Accordion title="Cluster">
    | Permission      | Description                                      |
    | :-------------- | :----------------------------------------------- |
    | Create Cluster  | Allows creation of new clusters                  |
    | Delete Cluster  | Allows deleting existing clusters                |
    | Manage Clusters | Allows updating and configuring cluster settings |
    | Read Cluster    | Allows viewing cluster details and configuration |
  </Accordion>

  <Accordion title="Workspace">
    | Permission       | Description                               |
    | :--------------- | :---------------------------------------- |
    | Create Workspace | Allows creation of new workspaces         |
    | Delete Workspace | Allows deleting existing workspaces       |
    | Manage Workspace | Allows updating workspace configuration   |
    | Read Workspace   | Allows viewing workspace details          |
    | List Workspaces  | Allows listing and viewing all workspaces |
  </Accordion>

  <Accordion title="Application">
    | Permission          | Description                                          |
    | :------------------ | :--------------------------------------------------- |
    | List Applications   | Allows listing and viewing applications              |
    | Manage Applications | Allows creating, updating, and deleting applications |
  </Accordion>

  <Accordion title="Environment">
    | Permission          | Description                                          |
    | :------------------ | :--------------------------------------------------- |
    | List Environments   | Allows listing and viewing environments              |
    | Manage Environments | Allows creating, updating, and deleting environments |
  </Accordion>

  <Accordion title="Policy">
    | Permission      | Description                                      |
    | :-------------- | :----------------------------------------------- |
    | List Policies   | Allows listing and viewing policies              |
    | Manage Policies | Allows creating, updating, and deleting policies |
  </Accordion>
</AccordionGroup>

#### Common

<AccordionGroup title="Common">
  <Accordion title="Role">
    | Permission   | Description                                   |
    | :----------- | :-------------------------------------------- |
    | List Roles   | Allows listing and viewing roles              |
    | Manage Roles | Allows creating, updating, and deleting roles |
  </Accordion>

  <Accordion title="User">
    | Permission   | Description                                   |
    | :----------- | :-------------------------------------------- |
    | List Users   | Allows listing and viewing users              |
    | Manage Users | Allows creating, updating, and deleting users |
  </Accordion>

  <Accordion title="Team">
    | Permission  | Description                                    |
    | :---------- | :--------------------------------------------- |
    | Create Team | Allows creation of new teams                   |
    | Delete Team | Allows deleting existing teams                 |
    | Manage Team | Allows updating team configuration and members |
    | Read Team   | Allows viewing team details and members        |
  </Accordion>

  <Accordion title="Virtual Account">
    | Permission             | Description                                              |
    | :--------------------- | :------------------------------------------------------- |
    | Create Virtual Account | Allows creation of new virtual accounts                  |
    | Read Virtual Account   | Allows viewing virtual account details                   |
    | Manage Virtual Account | Allows updating virtual account configuration and tokens |
    | Delete Virtual Account | Allows deleting existing virtual accounts                |
  </Accordion>

  <Accordion title="External Identity">
    | Permission                 | Description                                                 |
    | :------------------------- | :---------------------------------------------------------- |
    | List External Identities   | Allows listing and viewing external identities              |
    | Manage External Identities | Allows creating, updating, and deleting external identities |
  </Accordion>

  <Accordion title="Integrations">
    | Permission              | Description                                                 |
    | :---------------------- | :---------------------------------------------------------- |
    | Create Provider Account | Allows creation of new provider accounts                    |
    | Delete Provider Account | Allows deleting existing provider accounts                  |
    | Manage Provider Account | Allows updating provider account configuration              |
    | Read Provider Account   | Allows viewing provider account details                     |
    | Use Integrations        | Allows using integrations associated with provider accounts |
  </Accordion>

  <Accordion title="Repository">
    | Permission        | Description                              |
    | :---------------- | :--------------------------------------- |
    | Create Repository | Allows creation of new repositories      |
    | Delete Data       | Allows deleting data within repositories |
    | Delete Repository | Allows deleting existing repositories    |
    | Manage Repository | Allows updating repository configuration |
    | Read Data         | Allows reading data from repositories    |
    | Read Repository   | Allows viewing repository details        |
    | Write Data        | Allows writing data to repositories      |
  </Accordion>

  <Accordion title="Secret Group">
    | Permission          | Description                                |
    | :------------------ | :----------------------------------------- |
    | Create Secret Group | Allows creation of new secret groups       |
    | Delete Secret Group | Allows deleting existing secret groups     |
    | Manage Secret Group | Allows updating secret group configuration |
    | Read Data           | Allows reading secret values               |
    | Read Secret Group   | Allows viewing secret group details        |
    | Write Data          | Allows writing or updating secret values   |
  </Accordion>

  <Accordion title="Global Settings">
    | Permission      | Description                                  |
    | :-------------- | :------------------------------------------- |
    | List Settings   | Allows listing and viewing platform settings |
    | Manage Settings | Allows updating platform settings            |
  </Accordion>
</AccordionGroup>

## Managing Default Roles

TrueFoundry allows customizing default roles for giving flexibility to Admins to decide. For example, we can update the Default team manager role to allow them manage their virtual accounts on their own.

<Frame>
  <img src="https://mintcdn.com/truefoundry/XnXVxVsk1dnqxsr8/images/access-roles-default-roles-tab.png?fit=max&auto=format&n=XnXVxVsk1dnqxsr8&q=85&s=713124c42645f8bfde691c6f077d58ab" width="3024" height="1128" data-path="images/access-roles-default-roles-tab.png" />
</Frame>

<Warning>
  Updating permissions for `Admin` role is not allowed and by default contains
  all the permissions.
</Warning>

## Managing Custom Roles

Manage custom roles with fine-grained permissions on different resources at tenant level. These custom roles can then be assigned to Users.

<Steps>
  <Step title="Create new Custom Role">
    Navigate to `Access > Roles > Custom Roles`. Click on `Create Role`.

    <Frame>
      <img src="https://mintcdn.com/truefoundry/iWPat58JYfgUGHrb/images/access-roles-custom-role-create-new.png?fit=max&auto=format&n=iWPat58JYfgUGHrb&q=85&s=7d8a92baa5b936887c5b58bcbebfdb71" width="1766" height="750" data-path="images/access-roles-custom-role-create-new.png" />
    </Frame>
  </Step>

  <Step title="Fill the form">
    <ParamField body="Name" type="string" required>
      A unique name to identify this role. Should follow the format as Alphanumeric
      characters and hyphens.
    </ParamField>

    <ParamField body="Display Name" type="string" required>
      A user-friendly readable display name for this Role.
    </ParamField>

    <ParamField body="Description" type="string" required>
      A breif description about this role and its permissions.
    </ParamField>

    <ParamField body="Permissions" type="array" required>
      Permissions for the role. Refer to [Available Permissions](https://www.truefoundry.com/docs/platform/manage-user-roles-and-permissions#available-permissions) to understand these Permissions.

      <Note>
        * Permissions are applied to all the resources in the tenant.

        * Provider Account should be used to give permissions for Models, Guardrails, and Integrations.
      </Note>
    </ParamField>

    <Frame>
      <img src="https://mintcdn.com/truefoundry/pVeIvHPc1_yR1riQ/images/access-roles-custom-role-read-only-form.png?fit=max&auto=format&n=pVeIvHPc1_yR1riQ&q=85&s=b3c527a95f7b86e2712f759dc21039cc" width="3024" height="1714" data-path="images/access-roles-custom-role-read-only-form.png" />
    </Frame>

    <Accordion title="Example for read-only Platform User">
      <Frame>
        <img src="https://mintcdn.com/truefoundry/XnXVxVsk1dnqxsr8/images/access-roles-custom-role-read-only-user.png?fit=max&auto=format&n=XnXVxVsk1dnqxsr8&q=85&s=2ac9283bbcd4680ce2cbce10791ee8e0" width="1176" height="1078" data-path="images/access-roles-custom-role-read-only-user.png" />
      </Frame>
    </Accordion>
  </Step>

  <Step title="Assign Roles to user">
    Navigate to `Access > Users`. Select the user and click on `Edit User`.

    <Frame>
      <img src="https://mintcdn.com/truefoundry/pVeIvHPc1_yR1riQ/images/access-roles-custom-role-read-only-user-edit.png?fit=max&auto=format&n=pVeIvHPc1_yR1riQ&q=85&s=90b87bb0bb4704541af78d1a78d0fa7b" width="1768" height="794" data-path="images/access-roles-custom-role-read-only-user-edit.png" />
    </Frame>

    Select the role and `Save Changes`.

    <Frame>
      <img src="https://mintcdn.com/truefoundry/pVeIvHPc1_yR1riQ/images/access-roles-custom-role-read-only-user-assign.png?fit=max&auto=format&n=pVeIvHPc1_yR1riQ&q=85&s=a37d526108dfab3805365a6bca814e1a" width="3024" height="1360" data-path="images/access-roles-custom-role-read-only-user-assign.png" />
    </Frame>

    <Note>User can be assigned to only one role at a time.</Note>
  </Step>
</Steps>

## FAQ

<AccordionGroup>
  <Accordion title="How to allow Teams create and manage their Virtual Accounts without making them Admin?">
    You can achieve this by updating the default role for `Team Manager` to allow them to create or manage their `Virtual Account`.

    <Steps>
      <Step title="Edit default Team Manager Role">
        Navigate to `Access > Roles > Default Roles` and click on `Edit` button for the default `Team Manager` role.

        <Frame>
          <img src="https://mintcdn.com/truefoundry/iWPat58JYfgUGHrb/images/access-roles-default-role-team-manager-edit.png?fit=max&auto=format&n=iWPat58JYfgUGHrb&q=85&s=c6111f1a3f86b0f627dcf6b9c58515b3" width="1760" height="746" data-path="images/access-roles-default-role-team-manager-edit.png" />
        </Frame>
      </Step>

      <Step title="Update Team Manager role">
        Update the permissions based on your requirements.

        * **Read Virtual Account**: If you
          want to allow Team Managers view Virtual Accounts owned by their teams.
        * **Create Virtual Account**: If you
          want to allow Team Managers create new Virtual Accounts owned by their teams.
        * **Manage Virtual Account**: If you
          want to allow Team Managers update Virtual Accounts owned by their teams. This also includes permission to get the token.
        * **Delete Virtual Account**: If you
          want to allow Team Managers delete Virtual Accounts owned by their teams.

        <Frame>
          <img src="https://mintcdn.com/truefoundry/iWPat58JYfgUGHrb/images/access-roles-default-role-team-manager-edit-form.png?fit=max&auto=format&n=iWPat58JYfgUGHrb&q=85&s=f30744dd6964c503ec85bd6a25069a94" width="3022" height="1356" data-path="images/access-roles-default-role-team-manager-edit-form.png" />
        </Frame>

        <Warning>
          These permissions apply to all the Team Managers in the Account irrespective
          of their Teams.
        </Warning>
      </Step>

      <Step title="Manage Virtual Accounts">
        Team Managers should now be able to view, create, manage or delete Virtual Account owned by their Team based on the permissions given.

        <Frame>
          <img src="https://mintcdn.com/truefoundry/iWPat58JYfgUGHrb/images/access-va-list-team-owned-by.png?fit=max&auto=format&n=iWPat58JYfgUGHrb&q=85&s=847c85dbbba86b4599c0208de17c08ac" width="3024" height="1366" data-path="images/access-va-list-team-owned-by.png" />
        </Frame>
      </Step>
    </Steps>
  </Accordion>

  <Accordion title="How to allow Users create and manage their Virtual MCP Servers without making them Admin?">
    You can achieve this by assigning a custom role to those `Users` to allow them to create or manage their `Virtual MCP Servers`.

    <Steps>
      <Step title="Create a new Custom Role">
        Navigate to `Access > Roles > Custom Roles` and click on `Create Role` button.

        <Frame>
          <img src="https://mintcdn.com/truefoundry/iWPat58JYfgUGHrb/images/access-roles-custom-role-create-new.png?fit=max&auto=format&n=iWPat58JYfgUGHrb&q=85&s=7d8a92baa5b936887c5b58bcbebfdb71" width="1766" height="750" data-path="images/access-roles-custom-role-create-new.png" />
        </Frame>
      </Step>

      <Step title="Fill the form">
        Fill the form with required details and add `Create MCP Server` permission to allow just create a new MCP/Virtual MCP Server.

        <Frame>
          <img src="https://mintcdn.com/truefoundry/iWPat58JYfgUGHrb/images/access-roles-custom-role-mcp-server-create-form.png?fit=max&auto=format&n=iWPat58JYfgUGHrb&q=85&s=352c4c21588387f55782ada9c23c6009" width="3022" height="1716" data-path="images/access-roles-custom-role-mcp-server-create-form.png" />
        </Frame>

        <Info>
          This only allows the Users to create a new MCP Server. They cannot view or
          update any existing MCP Server in the Account through this permission. User
          can add themself as `MCP Server Manager` using collaborator to update and
          access the MCP Server while creating it first time.
        </Info>
      </Step>

      <Step title="Assing the Role to User">
        Navigate to `Access > Users` and click `Edit` for users to allow creating MCP Servers. Assign the created `Role`.

        <Frame>
          <img src="https://mintcdn.com/truefoundry/iWPat58JYfgUGHrb/images/access-roles-custom-role-mcp-server-create-assing-user.png?fit=max&auto=format&n=iWPat58JYfgUGHrb&q=85&s=240cf9848fb1fb302b8b244ae5f07927" width="3024" height="1354" data-path="images/access-roles-custom-role-mcp-server-create-assing-user.png" />
        </Frame>
      </Step>
    </Steps>
  </Accordion>

  <Accordion title="How to allow all Users to view Gateway Controls like Rate limit, Routing config, etc without making them Admin?">
    You can achieve this by editing the default `Member` role to allow them to view `Gateway Controls`.

    <Steps>
      <Step title="Edit default Team Manager Role">
        Navigate to `Access > Roles > Default Roles` and click on `Edit` button on `Member` role.

        <Frame>
          <img src="https://mintcdn.com/truefoundry/iWPat58JYfgUGHrb/images/access-roles-default-role-member-edit.png?fit=max&auto=format&n=iWPat58JYfgUGHrb&q=85&s=113a662e35c325d3bde88e2e33b06361" width="1766" height="800" data-path="images/access-roles-default-role-member-edit.png" />
        </Frame>
      </Step>

      <Step title="Update Team Manager role">
        Update the permissions to add `List Gateway Controls` and click `Update` to allow all the Users in the Account view all the Gateway Configs.

        <Frame>
          <img src="https://mintcdn.com/truefoundry/iWPat58JYfgUGHrb/images/access-roles-default-role-member-form-list-gw-control.png?fit=max&auto=format&n=iWPat58JYfgUGHrb&q=85&s=4304e59a9ab714ba457cbc02d89e0feb" width="1510" height="854" data-path="images/access-roles-default-role-member-form-list-gw-control.png" />
        </Frame>

        <Warning>
          This allows all the Users access to view all the Gateway Controls by default.
        </Warning>
      </Step>
    </Steps>
  </Accordion>

  <Accordion title="What are the permissions given to a Member role by default?">
    Default `Member` role comes with following permissions:

    * settings:ListSettings
    * repository:CreateRepository
    * secret-group:CreateSecretGroup
    * user:ListUsers
  </Accordion>
</AccordionGroup>

***
