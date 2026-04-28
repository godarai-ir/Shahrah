> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Cluster migration with Velero

> Move a TrueFoundry-connected Kubernetes cluster to a new cluster using Velero backups.

This guide describes how to migrate workloads and Kubernetes resources from an existing compute-plane cluster (source) to a new cluster (target) when both clusters live in your cloud account, using [Velero](https://velero.io/). It is aimed at teams who currently have **one** cluster connected to TrueFoundry and want to **disconnect** that cluster and **connect a second cluster** instead, without losing namespace-scoped state that Velero can carry over.

<Warning>
  This only works on truefoundry control plane version 0.135.0 or newer, if tried on an older version, it will delete everything from truefoundry, so proceed with caution.
</Warning>

<Note>
  Velero migrates Kubernetes API objects and (depending on configuration) volume
  data. It does **not** replace a full disaster-recovery design. Always run a
  dry run or non-production rehearsal before production cutovers.
</Note>

## Why stop TFY Agent before backup and restore

[TFY Agent](/docs/tfy-agent) and **TFY Agent Proxy** connect the compute plane to the TrueFoundry control plane. The proxy can **create, update, and delete** resources in the cluster based on control-plane instructions.

During migration, the target cluster may be empty or only partially provisioned for a period of time. If TFY Agent Proxy is running there **before** namespaces and resources restored from Velero exist, the control plane can reconcile against a cluster whose state does not yet match what you expect. That can lead to **unintended deletions or conflicting writes** (for example, pruning or recreating resources in namespaces that are not ready yet).

**Recommended practice:** scale **both** `tfy-agent` and `tfy-agent-proxy` to zero on whichever cluster should not be actively reconciling during the migration window, then bring them up only on the target cluster **after** Velero restore (and your TrueFoundry integration for the new cluster) is in place.

## Prerequisites

* **Velero** on the source cluster with a **BackupStorageLocation** (BSL) to object storage—complete bucket creation, IAM, and install using the first step below and the linked Velero docs.
* **Network and IAM** on the target so Velero there can read the same backup bucket. For migration, Velero’s documentation recommends configuring the target cluster’s BSL as **read-only** so restores do not accidentally delete shared backup data ([Velero cluster migration](https://velero.io/docs/main/migration-case/)).
* **Kubernetes version:** Velero does not support restoring into a cluster whose Kubernetes version is **lower** than the cluster where the backup was taken. Prefer the same or newer minor version on the target.
* **Persistent volumes:** Snapshot-based migration **across different cloud providers** is not natively supported the way same-provider same-region flows are. For cross-provider moves or when snapshots are not portable, consider Velero **File System Backup** or the **snapshot data mover**—see the [Velero cluster migration](https://velero.io/docs/main/migration-case/) “Before migrating your cluster” section.

Provider-specific plugin references:

* [Velero plugin for AWS](https://github.com/vmware-tanzu/velero-plugin-for-aws)
* [Velero plugin for GCP](https://github.com/vmware-tanzu/velero-plugin-for-gcp)
* [Velero plugin for Microsoft Azure](https://github.com/vmware-tanzu/velero-plugin-for-microsoft-azure)

## Migration workflow

<Steps>
  <Step title="Configure object storage, IAM, and Velero on the source cluster">
    Before you can back up to a shared location, create and secure the **backup bucket** (or equivalent) in your cloud and grant Velero the **IAM / workload identity** permissions it needs to read and write backup metadata and objects. You also need a **Velero install** on the source cluster with a **BackupStorageLocation** pointed at that bucket.

    Velero and each provider plugin document the exact policies, roles, service accounts, and install flags—follow those rather than copying partial snippets here:

    * [Velero basic install](https://velero.io/docs/main/basic-install/) and [supported providers](https://velero.io/docs/main/supported-providers/)
    * [Cluster migration](https://velero.io/docs/main/migration-case/) (shared bucket across clusters)
    * Provider plugins: [AWS](https://github.com/vmware-tanzu/velero-plugin-for-aws), [GCP](https://github.com/vmware-tanzu/velero-plugin-for-gcp), [Azure](https://github.com/vmware-tanzu/velero-plugin-for-microsoft-azure)

    Ensure the target cluster’s principals will be allowed to **read** the same bucket when you install Velero there (and use a **read-only** BSL on the target if you follow Velero’s migration guidance).
  </Step>

  <Step title="Scale down TFY Agent on the source cluster">
    Stop the components in the `tfy-agent` namespace (adjust names if your release uses different Deployment names):

    ```bash theme={"dark"}
    kubectl scale deployment tfy-agent tfy-agent-proxy -n tfy-agent --replicas=0
    ```

    Verify they are scaled down:

    ```bash theme={"dark"}
    kubectl get deployment -n tfy-agent
    ```

    <Note>
      If you use GitOps (for example Argo CD) to manage these Deployments, use your
      usual workflow to set replicas to `0` so they are not immediately scaled back
      up by a sync.
    </Note>
  </Step>

  <Step title="Create a Velero backup on the source cluster">
    Create a backup that includes the namespaces and cluster-scoped resources you need. Example (replace `<BACKUP_NAME>` and adjust namespace list or use [resource filters](https://velero.io/docs/main/resource-filtering/) as required):

    ```bash theme={"dark"}
    velero backup create <BACKUP_NAME> --wait
    ```

    To include volume data, follow Velero’s documentation for your storage class (CSI snapshots, File System Backup, etc.). Wait until the backup phase is **Completed**:

    ```bash theme={"dark"}
    velero backup describe <BACKUP_NAME>
    ```
  </Step>

  <Step title="Install Velero on the target cluster and point it at the same backup location">
    On the **target** cluster, install Velero and configure a **BackupStorageLocation** that references the **same** bucket (or prefix) as the source. For migration, Velero recommends a **read-only** BSL on the cluster where you only restore, to avoid accidental deletion of backup objects in object storage—see [Cluster migration](https://velero.io/docs/main/migration-case/).

    Use the install or `velero backup-location create` flow that matches your cloud; the official plugin repositories above contain the correct flags for AWS, GCP, and Azure.

    After install, confirm the backup appears on the target (Velero syncs Backup metadata from object storage; default sync interval is on the order of one minute):

    ```bash theme={"dark"}
    velero backup describe <BACKUP_NAME>
    ```
  </Step>

  <Step title="Restore on the target cluster">
    Create a restore from the backup:

    ```bash theme={"dark"}
    velero restore create --from-backup <BACKUP_NAME> --wait
    ```

    Inspect the restore for warnings or partial failures:

    ```bash theme={"dark"}
    velero restore describe <RESTORE_NAME>
    ```
  </Step>

  <Step title="Validate the target cluster">
    Confirm namespaces, workloads, ConfigMaps, Secrets, ingress, and data volumes behave as expected. Reconcile any cloud-specific resources (DNS, load balancers, IAM, node pools) that are **not** carried by Velero.

    <Warning>
      **Existing resources on the target:** If a Kubernetes resource with the same
      **group, kind, namespace, and name** already exists in the target cluster,
      Velero’s **restore** may **skip** that object instead of overwriting it, so
      you can end up with a **partial** migration. Prefer a **clean** target cluster
      (or a clearly scoped restore) for predictable results. If you intentionally
      keep overlapping resources, review Velero restore options (such as
      existing-resource policies in your Velero version) in the [restore
      reference](https://velero.io/docs/main/restore-reference/).
    </Warning>
  </Step>

  <Step title="Attach the new cluster in TrueFoundry and configure TFY Agent">
    1. Complete [compute plane](/docs/infrastructure/deploy-compute-plane) setup on the **target** cluster if it is not already installed (Istio, Argo CD, TrueFoundry components, etc.), **or** rely on what Velero restored—depending on what you included in the backup.
    2. Ensure the **cluster token / secret** used by TFY Agent matches the **new** cluster registration in the TrueFoundry control plane. If Velero restored an old `tfy-agent` Secret, update it to the credentials issued for the **new** cluster before scaling the agent back up.
    3. Scale TFY Agent back up on the **target** cluster only:

    ```bash theme={"dark"}
    kubectl scale deployment tfy-agent tfy-agent-proxy -n tfy-agent --replicas=1
    ```

    Adjust replica counts to match your high-availability requirements.

    4. Confirm in the control plane that the new cluster is **healthy** and that applications and workspaces appear as expected.
  </Step>
</Steps>

## Summary

1. **Configure** the backup bucket and IAM (per Velero and your cloud provider’s plugin docs), then install Velero on the source cluster with a BackupStorageLocation.
2. **Scale down** `tfy-agent` and `tfy-agent-proxy` on the source cluster (and avoid GitOps immediately reverting that).
3. **Backup** with Velero to shared object storage.
4. **Restore** on the target cluster using the same backup location (prefer read-only BSL on the target).
5. **Validate** data and APIs; watch for **skipped** resources when names already exist on the target.
6. **Connect** the target cluster to TrueFoundry with the correct integration credentials, then **scale up** TFY Agent on the target only.

For the authoritative Velero sequence and caveats (regions, API group versions, volume strategies), keep [Velero’s cluster migration guide](https://velero.io/docs/main/migration-case/) open alongside this page.
