> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Okta

> Set up Okta for SCIM provisioning with Portkey.

Portkey supports provisioning Users & Groups with Okta SAML Apps.

<Warning>
  Okta does not support SCIM Provisioning with OIDC apps; only SAML apps are supported.
</Warning>

To set up SCIM provisioning between Portkey and Okta, you must first create a SAML App on Okta.

***

### Setting up SCIM Provisioning

1. Navigate to the app settings. Under general settings, enable the SCIM provisioning checkbox.

   <img src="https://mintcdn.com/portkey-docs/_Cb_bj7tVjxcfwsu/images/scim/okta/enable.png?fit=max&auto=format&n=_Cb_bj7tVjxcfwsu&q=85&s=a40f4218510a9356a9bc8c1181a718bf" alt="Enable SCIM Provisioning" width="1482" height="936" data-path="images/scim/okta/enable.png" />

   <Info>
     The `Provisioning` tab should be visible after enabling SCIM provisioning. Navigate to that page.
   </Info>

2. Obtain the Tenant URL and Secret Token from the Portkey Admin Settings page (if SCIM is enabled for your organization).

   * [Portkey Settings Page](https://app.portkey.ai/settings/organisation/sso)

   <img src="https://mintcdn.com/portkey-docs/_Cb_bj7tVjxcfwsu/images/scim/portkey.png?fit=max&auto=format&n=_Cb_bj7tVjxcfwsu&q=85&s=d0b627c85923c759a2e107972ad352af" alt="Portkey Admin Settings" width="1600" height="1052" data-path="images/scim/portkey.png" />

3. Fill in the values from the Portkey dashboard into Okta's provisioning settings and click **`Test Connection`**. If successful, click **`Save`**.
   <Note>
     Ensure you choose the Authentication Mode as `HTTP Header`.
   </Note>

4. Check all the boxes as specified in the image below for full support of SCIM provisioning operations.

   <img src="https://mintcdn.com/portkey-docs/_Cb_bj7tVjxcfwsu/images/scim/okta/operations.png?fit=max&auto=format&n=_Cb_bj7tVjxcfwsu&q=85&s=ff197ebd7a9b781d7f3ab72aa70a63e4" alt="Provisioning Settings" width="1518" height="1114" data-path="images/scim/okta/operations.png" />

5. Once the details are saved, you will see two more options along with integration, namely `To App` and `To Okta`.

   Select `To App` to configure provisioning from Okta to Portkey.

   Enable the following checkboxes:

   * Create Users
   * Update User Attributes
   * Deactivate Users

   <img src="https://mintcdn.com/portkey-docs/_Cb_bj7tVjxcfwsu/images/scim/okta/user-settings.png?fit=max&auto=format&n=_Cb_bj7tVjxcfwsu&q=85&s=8a68a9a6940ab9db854713c9f8b24c7c" alt="Provisioning Settings" width="1600" height="1156" data-path="images/scim/okta/user-settings.png" />

   After saving the settings, the application header should resemble the following image.

   <img src="https://mintcdn.com/portkey-docs/_Cb_bj7tVjxcfwsu/images/scim/okta/status.png?fit=max&auto=format&n=_Cb_bj7tVjxcfwsu&q=85&s=b4b9036830190237883c6b5a7decd097" alt="Okta App Provisioning Status" width="1064" height="232" data-path="images/scim/okta/status.png" />

This completes the SCIM provisioning settings between Okta and Portkey.

Whenever you assign a `User` or `Group` to the application, Okta automatically pushes the updates to Portkey.

***

### Organisation role support

Portkey supports the following organisation roles:

* **`owner`**  (Organization Owner)
* **`admin`**  (Organization Admin)
* **`member`**  (Organization Member)

Users assigned any other role will default to the **member** role.

#### Editing Attributes

Okta by default doesn't support role attributes. To support role attributes, you need to edit the attributes in Okta.

1. Navigate to the app settings. Under general settings, click on the `Provisioning` tab.

2. Click on the `Go to Profile Editor` button, found under **Attribute Mappings** section.

3. Click on the `Add Attribute` button.

4. Fill the form with the following details:

   <img src="https://mintcdn.com/portkey-docs/_Cb_bj7tVjxcfwsu/images/scim/okta/add-attribute.png?fit=max&auto=format&n=_Cb_bj7tVjxcfwsu&q=85&s=cc3081783b9e09b3e4cd7a8c9d61116a" alt="Edit Attribute List" width="1378" height="1894" data-path="images/scim/okta/add-attribute.png" />

5. Click on the `Save` button.

#### Verifying the changes

To verify the changes, you can assign a user to the application with the desired role (e.g., **`owner`**, **`member`**, or **`admin`**) for the organization.

<img src="https://mintcdn.com/portkey-docs/_Cb_bj7tVjxcfwsu/images/scim/okta/select-roles.png?fit=max&auto=format&n=_Cb_bj7tVjxcfwsu&q=85&s=227700e88827e364ab6abbb7ca721089" alt="Assigning Roles" width="1404" height="254" data-path="images/scim/okta/select-roles.png" />

<Warning>
  Make sure to select only **one** role for a user, if multiple selected user will be assigned to highest qualified role.
</Warning>

***

### Group Provisioning with Okta

Portkey supports RBAC (Role-Based Access Control) for workspaces mapped to groups in Okta. You have two options for mapping groups to workspaces:

**Option 1: Flexible Group Mapping (Recommended)**

Provision groups with any naming convention from Okta, then map them to workspaces and assign roles directly from Portkey Control Plane. This eliminates the need for specific naming formats.

1. Push your groups to Portkey from Okta (see steps below)
2. Navigate to **Admin Settings > Authentication Settings > SCIM Provisioning** in Portkey
3. Use the **SCIM Mappings List** section to map groups to workspaces and assign roles

For detailed instructions, see the [SCIM Group Management](/product/enterprise-offering/org-management/scim/group-management) guide.

<Info>
  This is the recommended approach as it provides flexibility in group naming and easier management of group-to-workspace mappings.
</Info>

**Option 2: Naming Convention (Legacy)**

Alternatively, you can use the following naming convention for automatic mapping:

* **Format:** `ws-{group}-role-{role}`
  * **Role:** One of `admin`, `member`, or `manager`
* A user should belong to only one group per `{group}`.

**Example:**
For a `Sales` workspace:

* `ws-Sales-role-admin`
* `ws-Sales-role-manager`
* `ws-Sales-role-member`

Users assigned to these groups will inherit the corresponding role in Portkey.

**Custom Prefix and Separator Configuration:**

You can configure your own prefix and separator to match your organization's group naming conventions. Navigate to **Admin Settings > Authentication Settings > SCIM Provisioning** in Portkey Control Plane and configure the **Pattern Based SCIM Grouping** section with your preferred prefix and separator.

For example, if you configure:

* **Prefix:** `ws-`
* **Role Separator:** `-role-`

Then your groups should follow: `ws-{group}{role_separator}{admin,manager,member}`

For detailed instructions, see the [SCIM Group Management](/product/enterprise-offering/org-management/scim/group-management#configuring-group-naming-format-optional) guide.

<img src="https://mintcdn.com/portkey-docs/_Cb_bj7tVjxcfwsu/images/scim/portkey-group-roles.png?fit=max&auto=format&n=_Cb_bj7tVjxcfwsu&q=85&s=f77de0a8e5e770ed1fa78daf7f3a86e5" alt="Group Role Mapping" width="1600" height="425" data-path="images/scim/portkey-group-roles.png" />

<Warning>
  Automatic provisioning with Okta works for `Users`, but it does not automatically work for `Groups`.
</Warning>

To support automatic provisioning for groups, you must first push the groups to the App (Portkey). Then, Okta will automatically provision updates.

To push the groups to Portkey, navigate to the `Push Groups` tab. If it is not found, ensure you have followed all the steps correctly and enabled all the fields mentioned in the Provisioning steps.

1. Click on **Push Groups**.

   <img src="https://mintcdn.com/portkey-docs/_Cb_bj7tVjxcfwsu/images/scim/okta/push-groups.png?fit=max&auto=format&n=_Cb_bj7tVjxcfwsu&q=85&s=927b9f1ffc23f1dd3c5c299171f85172" alt="Push Groups" width="1650" height="558" data-path="images/scim/okta/push-groups.png" />

2. Select **Find group by name**.

3. Enter the name of the group, select the group from the list, and click **Save** or **Save & Add Another** to assign a new group.

<Tip>
  You can also use `Find groups by rule` to push multiple groups using a filter.
</Tip>

<Note>
  If there is any discrepancy or issue with group provisioning, you can retry provisioning by clicking the `Push Now` option. This can be found under the `Push Status` column in the groups list.
</Note>

***

### Support

If you encounter any issues with group provisioning, please reach out to us [here](mailto:support@portkey.ai).
