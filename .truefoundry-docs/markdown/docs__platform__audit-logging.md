> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Audit Logging

> Track and monitor all platform activities with comprehensive audit logs for security, compliance, and troubleshooting

TrueFoundry provides comprehensive audit logging that tracks all activities performed on the platform. Audit logs help you monitor user actions, track changes to resources, maintain compliance, and troubleshoot issues by providing a complete history of platform activities.

You can find the audit logs in the `Platform > Access > Audit Logs` section of the TrueFoundry dashboard.

<Frame caption="Audit Logs dashboard">
  <img src="https://mintcdn.com/truefoundry/L6_ximWisGPdTTa9/images/docs/platform/audit-logging-dashboard.png?fit=max&auto=format&n=L6_ximWisGPdTTa9&q=85&s=916e900039f60366e2ff1dc2a6ec9229" alt="TrueFoundry Audit Logs dashboard interface" width="3840" height="1872" data-path="images/docs/platform/audit-logging-dashboard.png" />
</Frame>

Audit logs capture detailed information about every action performed on the TrueFoundry platform, including:

* **User actions** - Who performed what action and when
* **Resource changes** - Creation, updates, and deletions of deployments, workspaces, integrations, and more
* **API activity** - All API calls made to the platform
* **System actions** - Automated actions performed by Autopilot and other system components

<Info>
  Audit logs are only visible to Tenant Admins.
</Info>

The audit logs interface provides a graph of the count of activities over time. You can select the time range to view the activity count. You can also group the activities by user, action type or filter it to autopilot actions.

The detailed log table provides granular information about each activity. Each log entry includes:

* **ACTION** - The type of action performed (e.g., "Apply Provider Integration", "Create Deployment")
* **RESOURCE** - The resource that was affected, including its name and type
* **API ROUTE** - The API endpoint that was called
* **USER** - The user or system component that performed the action
* **TIMESTAMP** - When the action occurred (displayed as relative time, e.g., "8m ago", "16m ago")

You can also view the details of each activity by clicking on the activity in the timeline.

<img src="https://mintcdn.com/truefoundry/L6_ximWisGPdTTa9/images/docs/platform/audit-logs-details.png?fit=max&auto=format&n=L6_ximWisGPdTTa9&q=85&s=cfb6ccea61cf6ef12166683e9cfafd29" alt="TrueFoundry Audit Logs details interface" width="3840" height="1868" data-path="images/docs/platform/audit-logs-details.png" />

## Exporting Audit Logs

The audit logs are printed in application logs in the following format:

```json wrap theme={"dark"}
time="2025-11-24 08:35:39" level=info context=AuditLogInterceptor module=ProviderAccountController tenant=truefoundry user=user:abc@example.com path=/v1/provider-accounts action=APPLY_PROVIDER_INTEGRATION resourceType=provider-integration resourceId=cmi8vpe170ib701rehy456uq0 resourceFqn=truefoundry:aws-bedrock:harshil-test-21-nov:model:anthropic-claude-3-haiku-20240307-v1-0 resourceName=anthropic-claude-3-haiku-20240307-v1-0 metaData="{\"providerAccountFqn\":\"truefoundry:aws-bedrock:harshil-test-21-nov\",\"manifest\":{\"cost\":{\"metric\":\"public_cost\"},\"name\":\"anthropic-claude-3-haiku-20240307-v1-0\",\"type\":\"integration/model/bedrock\",\"region\":\"us-east-2\",\"model_id\":\"anthropic.claude-3-haiku-20240307-v1:0\",\"model_types\":[\"chat\"]}}" trace_id=4e488dc171953c90073c92717f5e1813 span_id=dc4895726fbb0e4e trace_flags=01
```

You can export this to any logging platform of your choice and filter for entries containing `context=AuditLogInterceptor`.

## API Access

For programmatic access to audit logs, you can use the TrueFoundry API. Refer to the [API Reference](http://localhost:3000/api-reference/audit-logs/get-audit-logs) for detailed endpoint documentation.

<Note>
  Access to audit logs via API requires appropriate permissions. Ensure your API keys or authentication tokens have the necessary access rights.
</Note>
