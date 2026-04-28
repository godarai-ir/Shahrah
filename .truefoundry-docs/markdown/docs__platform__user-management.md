> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Manage Users

> Detailed guide to how to add, manage and delete users in Truefoundry

Users can be added manually to a tenant or in an automated way by setting up Single Sign-On (SSO) with your Identity Provider (IdP).

<Note>
  We recommend setting up SSO since its more secure and also makes it easier to
  manage users. You can setup SSO with your Identity Provider (IdP) by reading
  [here](/docs/sso).
</Note>

## Add users

If you have setup SSO, users will automatically be added to the platform the first time they log in. They will see a button like `Login wih Google|Azure|Okta|Keycloak` depending on the IdP you have setup.

<img src="https://mintcdn.com/truefoundry/CPOiCulkYtSxMDPK/images/docs/platform/login-button-users.png?fit=max&auto=format&n=CPOiCulkYtSxMDPK&q=85&s=f148b5e2fc48ea43672c781203340a8d" alt="Login with IdP" width="3024" height="1634" data-path="images/docs/platform/login-button-users.png" />

In case you have not setup SSO or you want to invite users who are not a part of your IdP, you can use the Invite User button in Access -> Users page. You will need to enter the email of the user you want to add.

<img src="https://mintcdn.com/truefoundry/5CkapnZ7CyjQJ4bx/images/docs/invite-user-dialog.png?fit=max&auto=format&n=5CkapnZ7CyjQJ4bx&q=85&s=7351d56f7ec83eebb0a910e43c184e58" alt="Invite User" width="3840" height="1870" data-path="images/docs/invite-user-dialog.png" />

<Warning>
  While inviting users manually, there is a checkbox for `Send email to set
      password`. If you check this checkbox, the user will receive an email with a
  link to set their password. You should not check this box if you want the user
  to sign in via SSO - since no password needs to be set in that case.
</Warning>

## Assign or modify user role

Users can be assigned to different roles in the platform. The roles are:

* **Admin** - Admin has the highest level of access and is responsible for managing the overall TrueFoundry platform. They have full control over all resources, including users, clusters, and workspaces. **Usually there should be only a few admins in an organization.**
* **Member** - These are general users of the platform. Members, by default, don't have access to any resources and need to be explicitly granted access to resources.

You can assign or modify user roles in the Access -> Users page.

<iframe href="https://app.supademo.com/embed/M834wawtkNsFANPaoRW8E" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/M834wawtkNsFANPaoRW8E" style={{ border: "none", display: "flex", margin: "auto" }} />

**In case you want the role to be assigned via a SSO group**, you can set that in the SSO settings via the Role Mapping section. In the image below, if the user belongs to the tenant-admin-group in your IdP, they will be assigned the Admin role in Truefoundry when they join.

<img src="https://mintcdn.com/truefoundry/CPOiCulkYtSxMDPK/images/docs/platform/sso-user-role-mapping.png?fit=max&auto=format&n=CPOiCulkYtSxMDPK&q=85&s=b44040f67f330aef4619c061554dc0b6" alt="Assign Role via SSO group" width="3840" height="1866" data-path="images/docs/platform/sso-user-role-mapping.png" />

<Note>
  If once a user's role is updated manually to tenant admin, it will take the
  precedence and will not be converted to tenant member even in absence of the
  SSO groups.
</Note>

## Deactivate user

Admins can deactivate a user's account. This will prevent the user from logging in to the platform. This can be useful if you do not want to delete the user's account, but just want to deactivate them temporarily.

{" "}

<iframe href="https://app.supademo.com/embed/mgzGvinYK6OQa2491Uo7Z" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/mgzGvinYK6OQa2491Uo7Z" style={{ border: "none", display: "flex", margin: "auto" }} />

## Delete user

Admins can delete a user's account that are no longer part of your organization. This will remove the user account from the list of users and will not effect any resources created by the user.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/9lYm7n0BF3i5yINM/images/platform-user-management-delete-user-2.png?fit=max&auto=format&n=9lYm7n0BF3i5yINM&q=85&s=052d2dd5e0c89b5f85cad229554e135b" width="1510" height="740" data-path="images/platform-user-management-delete-user-2.png" />
</Frame>

<Note>
  Before deleting the user, you MUST remove the user explicitily from all
  resources and teams, otherwise System will not allow to delete

  <img src="https://mintcdn.com/truefoundry/9lYm7n0BF3i5yINM/images/platform-user-management-delete-user-1.png?fit=max&auto=format&n=9lYm7n0BF3i5yINM&q=85&s=e84e110d6c9d97897d78d7d0c8669c6f" width="3016" height="1712" data-path="images/platform-user-management-delete-user-1.png" />
</Note>

## Reset Password for a user

This is needed only if you are managing users manually and have not setup SSO.

**Admins** can initiate a password reset process for a user. This will send an email to the user with a link to reset their password.

<iframe href="https://app.supademo.com/embed/_5X8B-50MA-Oom0qUMZA2" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/_5X8B-50MA-Oom0qUMZA2" style={{ border: "none", display: "flex", margin: "auto" }} />

## Personal Access Tokens (PATs)

Personal access tokens are long lived tokens that can assume all the permissions of the user and be used to access the platform programmatically.

### Management of PATs

Users can create and manage their own personal access tokens under `Access > Personal Access Token`

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/clzisp8sh1r6r9x7718o17qy6?embed_v=2" typeofembed="iframe" height="450px" width="100%" src="https://app.supademo.com/embed/clzisp8sh1r6r9x7718o17qy6?embed_v=2" style={{ border: "none", display: "flex", margin: "auto" }} />

### Configuring limits

Admins can limit the age of these tokens and number of these tokens that user can create under Security settings.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/iMid4yIOHvzf4Z4V/images/platform-settings-security-2.png?fit=max&auto=format&n=iMid4yIOHvzf4Z4V&q=85&s=980636eb4386c5eb5fa0bbd71308d4db" width="3335" height="1023" data-path="images/platform-settings-security-2.png" />
</Frame>

### Revoking PATs

Admins can also revoke all the active personal access tokens of a particular user account.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/jw406UAsc7ErYUq8/images/access-user-revoke-all-pat-1.png?fit=max&auto=format&n=jw406UAsc7ErYUq8&q=85&s=c82fe5a9673fcb2c6873bef0833c298f" width="3343" height="983" data-path="images/access-user-revoke-all-pat-1.png" />
</Frame>

## Manage Users programmatically

Admins can manage users programmatically using APIs. Please find the complete API reference [here](https://docs.truefoundry.com/api-reference/users/list-users).
