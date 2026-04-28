> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# GCP

> Integrate GCP services into your TrueFoundry environment.

TrueFoundry supports integrating with multiple GCP services like GCS, GKE, GCR, GSM, GCP Models etc.

To integrate any of the above services, you simply need to add your GCP account as a provider account and add integrations for the same as shown below:

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/clya760cr0d8fthf7bc1r1lfl" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/clya760cr0d8fthf7bc1r1lfl" style={{ border: "none", display: "flex", margin: "auto" }} />

**Share access with users, teams or everyone in your TrueFoundry account**

As shown in the previous slides, you can share access of each integration with users, teams or everyone in your TrueFoundry account. This would allow them to view and use the integration. Only tenant-admins can edit the integrations.

### Create a custom serviceaccount

1. Create an [IAM serviceaccount](https://cloud.google.com/iam/docs/service-accounts-create) named `tfy-<short-region-name>-<name>-platform` (name can be anything but add a prefix `tfy` to differentiate it with others), if not created before.
2. Once the *IAM serviceaccount* is created, make sure to [create a key](https://cloud.google.com/iam/docs/keys-create-delete#creating) in JSON format.
3. If you using the workflow propeller, you need to add the following permission to the service account:

```shell Shell lines theme={"dark"}
gcloud iam service-accounts add-iam-policy-binding <GCP_SERVICE_ACCOUNT_EMAIL> \
    --role roles/iam.workloadIdentityUser \
    --member "serviceAccount:<PROJECT_ID>.svc.id.goog[tfy-workflow-propeller/flytepropeller]" \
    --project <PROJECT_ID>
```

### Google Cloud Storage Integration

Follow the steps below to connect GCS storage to TrueFoundry:

1. Create a [GCP bucket](https://cloud.google.com/storage/docs/creating-buckets).

   * Make sure to add the lifecycle configurations on the bucket to delete multipart upload after 7 days.
   * For this go to GCP bucket -> Lifecycle -> Add a rule
   * Select `Delete multi-part upload` for 7 days

2. We also need to add the CORS policy to the GCP bucket. Right now adding the CORS policy to the GCP bucket is not possible through the console so for this, we will use *gsutil*

   1. Create a file called `cors.json` using the below command

   <CodeGroup>
     ```powershell Shell lines theme={"dark"}
     cat > cors.json <<EOF
     [
         {
           "origin": ["*"],
           "method": ["GET", "POST", "PUT"],
           "maxAgeSeconds": 3600
         }
     ]
     EOF
     ```
   </CodeGroup>

   2. Attach the above CORS policy to the service account by running the following command using [gsutils](https://cloud.google.com/storage/docs/gsutil_install)

   <CodeGroup>
     ```shell Shell lines theme={"dark"}
     gsutil cors set cors.json  gs://BUCKET_NAME
     ```
   </CodeGroup>

3. Create a [custom IAM role](https://cloud.google.com/iam/docs/creating-custom-roles#creating) with the following permissions and add to the `serviceaccount` created above:

   <CodeGroup>
     ```javascript JSON lines theme={"dark"}
     [
       "storage.objects.create",
       "storage.objects.delete",
       "storage.objects.get",
       "storage.objects.list",
       "storage.objects.update",
       "storage.buckets.create",
       "storage.buckets.get",
       "storage.buckets.list",
       "storage.buckets.create",
       "storage.buckets.update",
       "storage.multipartUploads.create",
       "storage.multipartUploads.list",
       "storage.multipartUploads.listParts",
       "storage.multipartUploads.abort",
       "resourcemanager.projects.get",
     ];
     ```
   </CodeGroup>

4. Add the following [IAM condition](https://cloud.google.com/iam/docs/conditions-overview) - `resource.name.startsWith('projects/\_/buckets/<bucket name>}')`

5. Navigate to **Integrations** tab and follow the steps shown the previous demo to integrate your storage.

### Google Artifact registry Integration

1. Create a [custom IAM role](https://cloud.google.com/iam/docs/creating-custom-roles#creating) with the following permissions and add to the `serviceaccount` created above:

   <CodeGroup>
     ```javascript JSON lines theme={"dark"}
     [
       "artifactregistry.dockerimages.get",
       "artifactregistry.dockerimages.list",
       "artifactregistry.locations.get",
       "artifactregistry.locations.list",
       "artifactregistry.repositories.get",
       "artifactregistry.repositories.list",
       "artifactregistry.repositories.create",
       "artifactregistry.repositories.createTagBinding",
       "artifactregistry.repositories.delete",
       "artifactregistry.repositories.deleteArtifacts",
       "artifactregistry.repositories.deleteTagBinding",
       "artifactregistry.repositories.downloadArtifacts",
       "artifactregistry.repositories.get",
       "artifactregistry.repositories.getIamPolicy",
       "artifactregistry.repositories.list",
       "artifactregistry.repositories.listEffectiveTags",
       "artifactregistry.repositories.listTagBindings",
       "artifactregistry.repositories.update",
       "artifactregistry.repositories.uploadArtifacts",
       "artifactregistry.tags.get",
       "artifactregistry.tags.list",
       "artifactregistry.tags.create",
       "artifactregistry.tags.update",
       "artifactregistry.versions.get",
       "artifactregistry.versions.list",
       "artifactregistry.versions.delete",
     ];
     ```
   </CodeGroup>

2. Navigate to **Integrations** tab and follow the steps shown the previous demo to integrate your Artifact registry.

### Google Secrets Manager Integration

1. Create a [custom IAM role](https://cloud.google.com/iam/docs/creating-custom-roles#creating) with the following permissions and add to the `serviceaccount` created above:

   <CodeGroup>
     ```javascript JSON lines theme={"dark"}
     [
       "secretmanager.secrets.get",
       "secretmanager.secrets.list",
       "secretmanager.secrets.create",
       "secretmanager.secrets.delete",
       "secretmanager.secrets.update",
       "secretmanager.versions.access",
       "secretmanager.versions.list",
       "secretmanager.versions.get",
       "secretmanager.versions.add",
       "secretmanager.versions.destroy",
       "resourcemanager.projects.get",
     ];
     ```
   </CodeGroup>

2. Add the following IAM condition- `resource.name.startsWith('projects/<GCP Project Number>/secrets/tfy')`

3. Navigate to **Integrations** tab and follow the steps shown the previous demo to integrate your secret manager.

### Google GKE cluster Integration

1. Create a [custom IAM role](https://cloud.google.com/iam/docs/creating-custom-roles#creating) with the following permissions and add to the `serviceaccount` created above:

   <CodeGroup>
     ```javascript JSON lines theme={"dark"}
     [
       "container.clusters.get",
       "container.clusters.list",
       "container.clusters.update",
       "container.nodes.delete",
       "container.nodes.list",
       "container.nodes.get",
       "container.nodes.getStatus",
       "container.nodes.list",
       "container.operations.get",
       "resourcemanager.projects.get",
     ];
     ```
   </CodeGroup>

2. Navigate to **Integrations** tab and follow the steps shown the previous demo to integrate your secret manager.

### Google Vertex Model Integration

1. Create the GCP Provider Account as described in the demo at the top of this document.

2. Create a [custom IAM role](https://cloud.google.com/iam/docs/creating-custom-roles#creating) with the following permission and add to the `serviceaccount` created above:

   <CodeGroup>
     ```javascript JSON lines theme={"dark"}
     ["aiplatform.endpoints.predict"];
     ```
   </CodeGroup>

3. Navigate to **Integrations** tab and edit the GCP Provider Account previously created and add the required models using their `model id` and they should start showing up in the AI Gateway. Here's an example of adding `gemini-1.5-flash-001`.

   <Frame caption="">
     <img src="https://mintcdn.com/truefoundry/s4Aj2_qGCrSP-zc8/images/992568b1-f6ef1af721286cf73dcb01a03f9913cf3ac9708924d72ba71a2775fa221e122c-Screenshot_2024-09-24_at_7.22.37_PM.png?fit=max&auto=format&n=s4Aj2_qGCrSP-zc8&q=85&s=78a85686e272ce45d8c27de79878c139" width="1646" height="1088" data-path="images/992568b1-f6ef1af721286cf73dcb01a03f9913cf3ac9708924d72ba71a2775fa221e122c-Screenshot_2024-09-24_at_7.22.37_PM.png" />
   </Frame>

***
