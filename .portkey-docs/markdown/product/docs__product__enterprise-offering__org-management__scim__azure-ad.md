> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Azure Entra

> Setup Azure Entra for SCIM provisioning with Portkey.

#### Azure Active Directory (Azure AD)

[Reference](https://learn.microsoft.com/en-us/azure/active-directory/app-provisioning/use-scim-to-provision-users-and-groups)

Setting up Azure Entra for SCIM provisioning consists of the following steps:

* **New Entra Application & SCIM Provisioning**
* **Application Roles**
* **SCIM Attribute Mapping Update**

***

##### New Entra Application

First, create a new Azure Entra application to set up SCIM provisioning with Portkey.

1. Navigate to the [Entra Applications Page](https://entra.microsoft.com/?culture=en-in\&country=in#view/Microsoft_AAD_IAM/AppGalleryBladeV2) and click **`Create your own application`**.

   <img src="https://mintcdn.com/portkey-docs/_Cb_bj7tVjxcfwsu/images/scim/azure/create.png?fit=max&auto=format&n=_Cb_bj7tVjxcfwsu&q=85&s=f5206a75481e372422aca79b379a8c49" alt="Application Creation" width="1600" height="899" data-path="images/scim/azure/create.png" />

2. Complete the required fields to create a new application.

3. Once the application is created, navigate to the application's **Provisioning** page under the **Manage** section.

4. Click **`New Configuration`** to go to the provisioning settings page.

   <img src="https://mintcdn.com/portkey-docs/_Cb_bj7tVjxcfwsu/images/scim/azure/settings.png?fit=max&auto=format&n=_Cb_bj7tVjxcfwsu&q=85&s=8611a17a4031e5ce633ae90c63255d57" alt="Provisioning Settings" width="1586" height="1344" data-path="images/scim/azure/settings.png" />

5. Obtain the **Tenant URL** and **Secret Token** from the Portkey Admin Settings page (if SCIM is enabled for your organization).

   * [Portkey Settings Page](https://app.portkey.ai/settings/organisation/sso)

   <img src="https://mintcdn.com/portkey-docs/_Cb_bj7tVjxcfwsu/images/scim/portkey.png?fit=max&auto=format&n=_Cb_bj7tVjxcfwsu&q=85&s=d0b627c85923c759a2e107972ad352af" alt="Portkey Admin Settings" width="1600" height="1052" data-path="images/scim/portkey.png" />

6. Fill in the values from the Portkey dashboard in Entra's provisioning settings and click **`Test Connection`**. If successful, click **`Create`**.

> If the test connection returns any errors, please contact us at [support@portkey.ai](mailto:support@portkey.ai).

***

##### Application Roles

Portkey supported roles should match Entra's application roles.

1. Navigate to **App Registrations** under **Enterprise Applications**, click **All Applications**, and select the application created earlier.
2. Go to the **App Roles** page and click **`Create app role`**.
   > Portkey supports two application-level roles:
   >
   > * **`member`**  (Organization Member)
   > * **`admin`**  (Organization Admin)
   > * **`owner`**  (Organization Owner)

<img src="https://mintcdn.com/portkey-docs/_Cb_bj7tVjxcfwsu/images/scim/azure/roles.png?fit=max&auto=format&n=_Cb_bj7tVjxcfwsu&q=85&s=caf900102f97513a39e2590e6ca4fac0" alt="App Roles" width="1600" height="482" data-path="images/scim/azure/roles.png" />

> Users assigned any other role will default to the **member** role.

3. To support group roles, create a role with the value **`group`** and a name in title-case (e.g., `Group` for the value `group`).

   <img src="https://mintcdn.com/portkey-docs/_Cb_bj7tVjxcfwsu/images/scim/azure/create-roles.png?fit=max&auto=format&n=_Cb_bj7tVjxcfwsu&q=85&s=f8da24ac64707441c8ae31a01d2a39d6" alt="Creating App Roles" width="1600" height="1015" data-path="images/scim/azure/create-roles.png" />

4. Assign users to the application with the desired role (e.g., **`owner`**, **`member`**, or **`admin`**) for the organization.

   <img src="https://mintcdn.com/portkey-docs/_Cb_bj7tVjxcfwsu/images/scim/azure/select-role.png?fit=max&auto=format&n=_Cb_bj7tVjxcfwsu&q=85&s=14aeffd9436871c85b6356f6cee18076" alt="Assigning Roles" width="602" height="686" data-path="images/scim/azure/select-role.png" />

***

#### Attribute Mapping

###### Adding a New Attribute

1. Go to the **Provisioning** page and click **Attribute Mapping (Preview)** to access the attributes page.

2. Enable advanced options and click **`Edit attribute list for customappsso`**.

   <img src="https://mintcdn.com/portkey-docs/_Cb_bj7tVjxcfwsu/images/scim/azure/new-mapping.png?fit=max&auto=format&n=_Cb_bj7tVjxcfwsu&q=85&s=1e04afa8230640c147aacc1dc8201fcf" alt="Edit Attribute List" width="778" height="502" data-path="images/scim/azure/new-mapping.png" />

3. Add a new attribute called **`roles`** with the following properties:

   * **Multi-valued:** Enabled
   * **Type:** String

   <img src="https://mintcdn.com/portkey-docs/_Cb_bj7tVjxcfwsu/images/scim/azure/mapping-properties.png?fit=max&auto=format&n=_Cb_bj7tVjxcfwsu&q=85&s=dc2f184c7aefc5fb247c49e339f87900" alt="Roles Attribute Properties" width="1544" height="214" data-path="images/scim/azure/mapping-properties.png" />

###### Adding a new mapping

1. Click on the **`Add new mapping`** link to add a new mapping. (refer to the above images).
2. Follow the values from the below image to add a new mapping.

<img src="https://mintcdn.com/portkey-docs/_Cb_bj7tVjxcfwsu/images/scim/new-mapping.png?fit=max&auto=format&n=_Cb_bj7tVjxcfwsu&q=85&s=b7e3fc86d59388e209b8593b0c2af41e" alt="New Mapping Attributes" width="1488" height="1388" data-path="images/scim/new-mapping.png" />

3. Once done, save the changes.

###### Removing Unnecessary Attributes

Delete the following unsupported attributes:

* **preferredLanguage**
* **addresses (all fields)**
* **phoneNumbers**

***

#### Updating Attributes

**Update `displayName`**

1. Edit the **`displayName`** field to concatenate `firstName + lastName` instead of using the default `displayName` value from Entra records.

   <img src="https://mintcdn.com/portkey-docs/_Cb_bj7tVjxcfwsu/images/scim/azure/update-displayname.png?fit=max&auto=format&n=_Cb_bj7tVjxcfwsu&q=85&s=5177af2838df60c80814b92d5c8d2521" alt="Update displayName Expression" width="1500" height="526" data-path="images/scim/azure/update-displayname.png" />

2. Save the changes and enable provisioning on the **Overview** page of the provisioning settings.

***

<Tip>Looking for Workspace Management? Check out our [Workspace Management](/product/enterprise-offering/org-management/scim/group-management) page.</Tip>

##### Group (Workspace) Provisioning

Portkey supports RBAC (Role-Based Access Control) for workspaces mapped to groups in Entra. You have two options for mapping groups to workspaces:

**Option 1: Flexible Group Mapping (Recommended)**

Provision groups with any naming convention from Azure Entra, then map them to workspaces and assign roles directly from Portkey Control Plane. This eliminates the need for specific naming formats.

1. Provision your groups from Azure Entra (assign groups to the application)
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

<img src="https://mintcdn.com/portkey-docs/_Cb_bj7tVjxcfwsu/images/scim/portkey-group-roles.png?fit=max&auto=format&n=_Cb_bj7tVjxcfwsu&q=85&s=f77de0a8e5e770ed1fa78daf7f3a86e5" alt="Entra Group Role Mapping" width="1600" height="425" data-path="images/scim/portkey-group-roles.png" />

***

### Support

If you face any issues with the group provisioning, please reach out to us at [here](mailto:support@portkey.ai).
