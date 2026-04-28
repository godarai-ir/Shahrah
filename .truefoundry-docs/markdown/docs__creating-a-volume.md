> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Creating And Using Volumes

> Step-by-step guide for creating a volume, explaining configuration, best practices, and real-world usage on TrueFoundry.

To create a volume, click on `New Deployment` and select `Volume`.

<img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/docs/new-volume-button.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=6e0dc39af98eb4c5811176ee85a4b564" width="3840" height="1938" data-path="images/docs/new-volume-button.png" />

<Info>
  To create a volume, you will need to select a workspace.
  **Volumes created within a specific workspace in TrueFoundry can only be used by applications deployed within the same workspace.**

  This ensures that the volume's data is accessible only to the applications that are part of the same environment, enhancing security and organization.
</Info>

The volume form will look something like this:

<img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/docs/new-volume-form.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=3793475fe804f2abe2617dbadb81cedc" width="3840" height="1934" data-path="images/docs/new-volume-form.png" />

The key fields to fill up in the form are:

1. <u>**Name**</u>: The name of the volume.
2. <u>**Volume Config**</u>: You can select either of `Create New Volume` or
   `Use Existing Volume`.

<Tip>
  In most cases, you will want to create a new volume. The Use Existing Volume
  option is useful when you want to mount a pre-existing volume or bucket like
  S3, GCS, etc.
</Tip>

<Tabs>
  <Tab title="Create New Volume">
    You will need to fill up two fields here:

    1. **Size**: The size of the volume in Gi.

    <Warning>
      In most cases, you can increase the size of the volume after it is created.
      But you cannot decrease the size of the volume once created. So if you are not
      sure about the final size, you can start with a smaller size and increase it
      later.
    </Warning>

    2. **Storage Class**: The Kubernetes storage class to use for the volume. Storage classes provide a way to specify the type of storage that should be provisioned for a Volume. These storage classes differ in their characteristics, such as performance, durability, and cost. The specific storage classes available will depend on the cloud provider you are using and what is preconfigured by the Infra team.

    <Note>
      In case you want to whitelist only certain storage classes in the cluster to
      be used for provisioning volumes, you can add the **`truefoundry.com/enabled:
              'true'`** label to the StorageClass resources. The dropdown will then show
      only the storage classes with this label.
    </Note>

    The most common storage classes available in different cloud providers are:

    <AccordionGroup>
      <Accordion title="AWS Storage Classes">
        | Storage Class | Cloud Provider Storage Name | Storage Class   | Description                                                                                                                                                                                                           |
        | ------------- | --------------------------- | --------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
        | efs-sc        | Elastic File System (EFS)   | efs.csi.aws.com | A fully managed, scalable, and highly durable elastic file system that offers high availability, automatic scaling, and cost-effective general file sharing. It's suitable for workloads with varying capacity needs. |
      </Accordion>

      <Accordion title="GCP Storage Classes">
        | TrueFoundry Storage Name | Cloud Provider Storage Name | Storage Class                | Description                                                                                                                                                                                                                                        |
        | ------------------------ | --------------------------- | ---------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
        | standard-rwx             | Google Basic HDD Filestore  | filestore.csi.storage.gke.io | A cost-effective and scalable file storage solution ideal for general-purpose file storage and cost-sensitive workloads. It offers lower cost but also lower performance due to its HDD-based nature.                                              |
        | premium-rwx              | Google Premium Filestore    | filestore.csi.storage.gke.io | Provides higher performance and throughput compared to Basic HDD, making it suitable for I/O-intensive file operations and demanding workloads. It's SSD-based, offering higher performance at a higher cost.                                      |
        | enterprise-rwx           | Google Enterprise Filestore | filestore.csi.storage.gke.io | Delivers the highest performance, throughput, advanced features, multi-zone support, and high availability, making it ideal for mission-critical workloads and applications with strict availability requirements. It comes with the highest cost. |
      </Accordion>

      <Accordion title="Azure Storage Classes">
        | TrueFoundry Storage Name | Cloud Provider Storage Name       | Storage Class      | Description                                                                                                                                                           |
        | ------------------------ | --------------------------------- | ------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
        | azurefile                | Azure File Storage (Standard)     | file.csi.azure.com | Uses Azure Standard storage to create file shares for general file sharing across VMs or containers, including Windows apps. It offers cost-effective performance.    |
        | azurefile-premium        | Azure File Storage (Premium)      | file.csi.azure.com | Uses Azure Premium storage for higher performance, making it suitable for I/O-intensive file operations.                                                              |
        | azurefile-csi            | Azure File Storage (StandardCSI)  | file.csi.azure.com | Leverages Azure Standard storage with CSI for dynamic provisioning, potentially offering better performance and CSI features.                                         |
        | azurefile-csi-premium    | Azure File Storage (PremiumCSI)   | file.csi.azure.com | Combines Azure Premium storage with CSI for dynamic provisioning and high-performance file operations.                                                                |
        | azureblob-nfs-premium    | Azure Blob Storage (NFS Premium)  | blob.csi.azure.com | Uses Azure Premium storage with NFS v3 protocol for accessing large amounts of unstructured data and object storage, catering to demanding workloads with NFS access. |
        | azureblob-fuse-premium   | Azure Blob Storage (Fuse Premium) | blob.csi.azure.com | Uses Azure Premium storage with BlobFuse for accessing large amounts of unstructured data and object storage, suitable for workloads that require BlobFuse access.    |
      </Accordion>
    </AccordionGroup>
  </Tab>

  <Tab title="Use Existing Volume">
    In this case, you will first create a PersistentVolume object in the Kubernetes cluster and then provide the name of the object in the `PersistentVolume` field. This option is useful if you want to mount the data that is present in an existing S3, EFS or GCS bucket.

    <Note>
      Azure has default support for mounting Azure container buckets as volumes. So,
      you can just follow the route of creating a volume in truefoundry.
    </Note>

    <Accordion title="Create PersistentVolume for S3 Bucket">
      To mount an S3 bucket as a volume on TrueFoundry, you need to follow the following steps:

      ### Setting up IAM Policies and Relevant Roles

      Please follow this [document](https://docs.aws.amazon.com/eks/latest/userguide/s3-csi.html) of AWS to set up mount point of S3 in an EKS cluster.

      This will guide you to do the following things:

      * Create an IAM policy to give permissions for mount point to access the s3 bucket
      * Create an IAM role.
      * Install the mountpoint for Amazon S3 CSI driver and attach the role that was created above.

      ### Creating a Persistent Volume on the Kubernetes Cluster

      Create a PV with the following spec (by doing a `kubectl apply`):

      <CodeGroup>
        ```yaml YAML lines theme={"dark"}
        apiVersion: v1
        kind: PersistentVolume
        metadata:
          name: <Enter persistent volume name here>
        spec:
          capacity:
            storage: 100Gi
          csi:
            driver: s3.csi.aws.com
            volumeHandle: s3-csi-driver-volume # must be unique
            volumeAttributes:
              bucketName: <Enter s3 bucket name here>
          accessModes:
            - ReadWriteMany
          persistentVolumeReclaimPolicy: Retain
          storageClassName: s3-test # put any value here
          mountOptions:
            - allow-delete
            - region <enter aws region here>
            - allow-other
            - uid=1000
          volumeMode: Filesystem
        ```
      </CodeGroup>

      ### Create a Volume on TrueFoundry:

      Please follow this [section](/docs/creating-a-volume#statically-provisioned-volumes) to create volume on TrueFoundry
    </Accordion>

    <Accordion title="Create PersistentVolume for existing EFS volume">
      To mount a EFS as a volume on truefoundry, you need to follow the following steps:

      ### Install EFS CSI driver on your cluster

      * To install EFS CSI driver on your cluster, go to TrueFoundry UI -> `Clusters`-> `Installed Applications`-> `Manage`

              <Frame caption="">
                <img src="https://mintcdn.com/truefoundry/FrY4JbiyZud2He3p/images/17c3f47e-d9cb5b9-image.png?fit=max&auto=format&n=FrY4JbiyZud2He3p&q=85&s=46a94bbbe3e66a0f9327bde360d6f05c" width="978" height="1204" data-path="images/17c3f47e-d9cb5b9-image.png" />
              </Frame>

      * From the `Volumes`section click on install AWS EFS CSI driver and click on Install.

      <Frame caption="">
        <img src="https://mintcdn.com/truefoundry/FrY4JbiyZud2He3p/images/24d5b5e4-2151783-image.png?fit=max&auto=format&n=FrY4JbiyZud2He3p&q=85&s=09e5bfb1a22d09591509b16fad2cedea" width="1904" height="1272" data-path="images/24d5b5e4-2151783-image.png" />
      </Frame>

      ### Create an Access Point for your EFS

      * Locate your EFS in the AWS console and open it. Ensure that the EFS and the K8S cluster are in the same VPC. Click on "Create access point"

              <Frame caption="">
                <img src="https://mintcdn.com/truefoundry/DdP_2rhue4AQQlob/images/29883581-76befc7-image.png?fit=max&auto=format&n=DdP_2rhue4AQQlob&q=85&s=e34eca6442a194fab970242f2a29d46c" width="1920" height="720" data-path="images/29883581-76befc7-image.png" />
              </Frame>

      * Enter details like name, `Root directory path`(please make sure you fill the `Root directory creation permissions`section, can fill in with UID:1000, GID:1000 if you want to attach it to notebook)

              <Frame caption="">
                <img src="https://mintcdn.com/truefoundry/s4Aj2_qGCrSP-zc8/images/8d7eeedc-a1020b7-image.png?fit=max&auto=format&n=s4Aj2_qGCrSP-zc8&q=85&s=3a8c1b75fdfa312a69764f8599017a1a" width="2354" height="2448" data-path="images/8d7eeedc-a1020b7-image.png" />
              </Frame>

      * Click on create.

      ### Create a PersistentVolume on the cluster

      Create a PV with the following spec (by doing a `kubectl apply`):

      <CodeGroup>
        ```yaml YAML lines theme={"dark"}
        apiVersion: v1
        kind: PersistentVolume
        metadata:
          name: <Enter your PV name here>
        spec:
          capacity:
            storage: 5Gi # this number doesn't matter for EFS, any number will work
          csi:
            driver: efs.csi.aws.com
            volumeHandle: <file_system_id>::<access_point_id> # e.g. fs-036e93cbb1fabcdef::fsap-0923ac354cqwerty
          accessModes:
            - ReadWriteMany
          persistentVolumeReclaimPolicy: Retain
          storageClassName: efs-sc
          volumeMode: Filesystem
        ```
      </CodeGroup>

      ### Create a Volume on TrueFoundry:

      Please follow this [section](/docs/creating-a-volume#statically-provisioned-volumes) to create volume on TrueFoundry
    </Accordion>

    <Accordion title="Create PersistentVolume for existing GCS Bucket">
      To mount a GCS bucket as a volume on truefoundry, you need to follow the following steps. You can refer to this [document](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/cloud-storage-fuse-csi-driver) for more details:

      ### Create a GCS bucket

      Create a GCS bucket and ensure the following:

      * Should be single region ( multi-region will work but speed will be slower and costs will be higher )
      * Region should be the same as that of your Kubernetes Cluster

      ### Create Serviceaccount and Grant relevant permissions

      You need to run the following script. This does the following:

      * Enables GCS Fuse Driver on the cluster
      * Create IAM Policy to access your bucket
      * Create K8s service-account and add policy to this service-account
      * Enables role-binding of service account to the desired K8s namespace.

      <CodeGroup>
        ```bash Shell lines theme={"dark"}
        #!/bin/bash

        PROJECT=<Enter your Project Name>
        CLUSTER_NAME=<Enter your Cluster Name>
        REGION=<Region of your Cluster>
        BUCKET_NAME=<Name of GCS Bucket>
        TARGET_NAMESPACE=<Namespace where you want to mount your GCS bucket>

        # Clip bucket name to 20 characters
        CLIPPED_BUCKET_NAME="${BUCKET_NAME:0:20}"

        K8S_SA_NAME="$CLIPPED_BUCKET_NAME-bucket-sa"
        GCP_SA_NAME="$CLIPPED_BUCKET_NAME-gcp-sa"

        GCP_SA_ID="$GCP_SA_NAME@$PROJECT.iam.gserviceaccount.com"


        gcloud container clusters update $CLUSTER_NAME \
            --update-addons GcsFuseCsiDriver=ENABLED \
            --region=$REGION \
            --project=$PROJECT
        gcloud iam service-accounts create $GCP_SA_NAME  --project=$PROJECT
        gcloud storage buckets add-iam-policy-binding gs://$BUCKET_NAME \
            --member "serviceAccount:$GCP_SA_ID" \
            --role "roles/storage.objectAdmin" \
            --project $PROJECT

        gcloud iam service-accounts add-iam-policy-binding $GCP_SA_ID \
            --role roles/iam.workloadIdentityUser \
            --member "serviceAccount:$PROJECT.svc.id.goog[$TARGET_NAMESPACE/$K8S_SA_NAME]" \
            --project $PROJECT
        ```
      </CodeGroup>

      ### Create Service-Account in the Namespace

      We now need to create a serviceaccount on truefoundry in the same namespace with name: `TARGET_NAMESPACE`and the serviceaccount must have the name `K8S_SA_NAME`.

      ```yaml lines theme={"dark"}
      apiVersion: v1
      kind: ServiceAccount
      metadata:
        name: <Enter K8S_SA_NAME same from above step>
        namespace: <NAMESPACE>
        annotations:
          iam.gke.io/gcp-service-account: <Enter GCP IAM Principal Here>
      ```

      <Info>
        ### Important

        The service account name and workspace should be exactly same as the previous step.
      </Info>

      ### Create a PersistentVolume object as shown below:

      Create a persistent volume object with the following step. (by doing a `kubectl apply`)

      <CodeGroup>
        ```yaml YAML lines theme={"dark"}
        apiVersion: v1
        kind: PersistentVolume
        metadata:
          name: <Enter a unique PV name here>
        spec:
          capacity:
            storage: 30Gi
          csi:
            driver: gcsfuse.csi.storage.gke.io
            volumeHandle: <Enter your Bucket Name here>
          accessModes:
            - ReadWriteMany
          persistentVolumeReclaimPolicy: Retain
          storageClassName: np-fuse-test
          mountOptions:
            - implicit-dirs
          volumeMode: Filesystem
        ```
      </CodeGroup>

      ### Create a Volume on TrueFoundry:

      Please follow this [section](/docs/creating-a-volume#statically-provisioned-volumes) to create volume on TrueFoundry
    </Accordion>
  </Tab>
</Tabs>

3. <u>**Volume Browser**</u>: Enabling the volume browser will provide you a
   user-friendly interface that allows you to effortlessly upload, download, and
   view files stored within the Volumes. The interface will need to be
   configured with a password so that only authorized users can view the data in
   the volume. The password will need to be first created as a Secret and then
   entered in the form below.

<Accordion title="Create the secret for the volume browser password">
  <iframe provider="app.supademo.com" href="https://app.supademo.com/embed/Nt_KV4H3mPWiJbL8h_EWt" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/Nt_KV4H3mPWiJbL8h_EWt" style={{ border: "none", display: "flex", margin: "auto" }} />
</Accordion>

You can enable the volume browser by following the demo below:

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/0OVnHAXW7eiFfAIyg5EOx" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/0OVnHAXW7eiFfAIyg5EOx" style={{ border: "none", display: "flex", margin: "auto" }} />

## Using the Volume

Once you press the submit button in the deployment form, the volume will be created within a few seconds. The Volume Dashboard will transform to appear as follows:

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/FrY4JbiyZud2He3p/images/28ca68d8-365997e-Screenshot_2023-08-24_at_10.56.38_AM.png?fit=max&auto=format&n=FrY4JbiyZud2He3p&q=85&s=fc8d583d4d1932546f5ad87e3c1cb3d0" width="2940" height="451" data-path="images/28ca68d8-365997e-Screenshot_2023-08-24_at_10.56.38_AM.png" />
</Frame>

You can use the volume in the following ways:

<AccordionGroup>
  <Accordion title="Browse / Upload / Download data using Volume Browser">
    Once set-up, you can use Volume Browser to effortlessly upload, download, and view files stored within the Volumes.

    <iframe provider="app.supademo.com" href="https://app.supademo.com/embed/4l8sjWeBHeW5a5UbeMUh4" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/4l8sjWeBHeW5a5UbeMUh4" style={{ border: "none", display: "flex", margin: "auto" }} />
  </Accordion>

  <Accordion title="Use the volume to store / read data from your deployments">
    You can attach Volumes to various types of deployments, including Services and Jobs. You can then read and write data to the volume like you would normally read and write data to a file system.

    You can read this [guide](/docs/mounting-volumes-service) to learn more about how to use volumes in deployments. This guide is for Service, but the same can be applied to Jobs as well.
  </Accordion>
</AccordionGroup>
