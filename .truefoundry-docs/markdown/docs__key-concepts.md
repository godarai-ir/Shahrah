> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Key Concepts

> Key concepts you need to understand before using TrueFoundry.

Before we get started with deploying on TrueFoundry, it's important to understand a few key concepts - **Cluster**, **Workspace**, **Repository**, **Environment**.

## Cluster

A Cluster in TrueFoundry refers to an actual Kubernetes cluster in one of your cloud accounts or on-prem.

<Info>
  A Kubernetes cluster is a group of computers that work together to automatically run, scale, and manage your applications without you having to worry about the underlying infrastructure.
</Info>

<Accordion title="Creating / Attaching a cluster">
  You can create a new Kubernetes cluster or attach an existing one from the Platform -> Cluster tab in the portal.

  <Frame caption="">
    <img src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/cluster-create-attach.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=e99b404fd7bc4bad6a184e644e56a5dc" width="3024" height="1712" data-path="images/docs/cluster-create-attach.png" />
  </Frame>
</Accordion>

## Workspace

**A TrueFoundry workspace refers to namespace in Kubernetes**. A cluster can have multiple workspaces. Each workspace is a logical separation within a cluster so that different teams, applications, and environments can sit within a cluster.

For e.g, let's say there are three teams within a company. Team1 manages application1, application2, team2 manages application3, and team3 manages application4 and application5. Each of the applications further has three environments - dev, staging, and production. One way to organize the workspaces can be:

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/4MAaF__cLD4iud16/images/58e29dd1-97157ba-image.png?fit=max&auto=format&n=4MAaF__cLD4iud16&q=85&s=eeb3bd240348e2a3bf00f0852082fa14" width="1771" height="957" data-path="images/58e29dd1-97157ba-image.png" />
</Frame>

We can also have the dev and prod workspaces in one cluster if you want to structure them like that.

<Accordion title="Creating a Workspace">
  You can create a workspace from the Workspace tab in the platform. Once you create, you can get the FQN of the workspace from the FQN button.

  <iframe provider="app.supademo.com" href="https://app.supademo.com/embed/cmb7ksf634r9appkppw4reut5?embed_v=2" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/cmb7ksf634r9appkppw4reut5?embed_v=2" style={{ border: "none", display: "flex", margin: "auto" }} />
</Accordion>

## Repositories (also called MLRepo)

Repositories in TrueFoundry are repositories to store models, data, artifacts and prompts. They are backed by a blob storage like S3, GCS, or Minio. Like in Git repositories, you can configure access to read, or write to different users, you can similarily configure access control in TrueFoundry repositories.

You can also grant access of repositories to certain workspaces. Once a workspace has access to an ML repo, all applications inside the workspace can then read or write to the MLRepo depending on whether the workspace has viewer or editor access. This way, we don't need to inject any keys and this makes the entire system much more secure.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/OHzlp6GY5G-JfKle/images/b935c856-7e11e87-image.png?fit=max&auto=format&n=OHzlp6GY5G-JfKle&q=85&s=8b911dcc3d0856f382ef0453e50ed47a" width="1846" height="1366" data-path="images/b935c856-7e11e87-image.png" />
</Frame>

In the case above, workspace1 has access to MLRepo1 and hence team1 can access the assets in MLrepo1. This way, you can also divide datasets and model access across teams and workspaces.

<Accordion title="Creating a Repository">
  <Info>
    ### Prerequiste - Blob Storage Integration

    Before you can create a Repository, you'd need to connect one or more Blob Storages (S3, GCS, Azure Blob, MinIO, etc) to store artifacts and models associated with a ML Repo. If this one time setup is already done, you can skip to next section

    You can refer to one of the following pages to connect your blob storage to TrueFoundry

    * [AWS S3](doc:integration-provider-aws)
    * [Google Cloud Storage](doc:integration-provider-gcp)
    * [Azure Blob Storage](doc:integration-provider-azure)
    * Any S3 API Compatible Storage
  </Info>

  You can then create a repository from the Repositories tab in the Platform.

  <iframe provider="app.supademo.com" href="https://app.supademo.com/embed/cmb7laubs4rckppkpfmfenq9i?embed_v=2" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/cmb7laubs4rckppkpfmfenq9i?embed_v=2" style={{ border: "none", display: "flex", margin: "auto" }} />
</Accordion>

<Accordion title="Grant access of Repository to Workspace">
  Providing access to a certain ML Repos to a workspace ensures that every application in the Workspace gains access to that ML Repo. You can Grant access to an ML Repo to a Workspace while creating or editing a Workspace.

  <Info>
    This method does not grant access to interactive applications like Workbenches (Jupyter Notebooks, SSH Servers, etc).
    For workbenches, you must follow steps at [Setup CLI](/docs/setup-cli)
  </Info>

  <iframe provider="app.supademo.com" href="https://app.supademo.com/embed/cm3mwh0ej0mooahr7tqkjhe8c?embed_v=2" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/cm3mwh0ej0mooahr7tqkjhe8c?embed_v=2" style={{ border: "none", display: "flex", margin: "auto" }} />
</Accordion>

## Environment

An environment is a tag applied to workspaces to categorize them based on factors like development stage - for example: dev, staging, production. We create three environments initially dev, staging and prod - the names of which you can change at any point of time.
A cluster can be associated to multiple environments but a workspace can only be associated to one environment. For each environment, you will need to define its for production and if it needs to be optimized for cost or availability. These settings are used to power the recommendations to improve your applications.

<Accordion title="Creating an environment">
  To create an environment in TrueFoundry, follow these steps:

  <iframe provider="app.supademo.com" href="https://app.supademo.com/embed/cmb7lr5r44s7appkppgk2ke4u?embed_v=2" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/cmb7lr5r44s7appkppgk2ke4u?embed_v=2" style={{ border: "none", display: "flex", margin: "auto" }} />
</Accordion>

<Accordion title="Attach an environment to a cluster and workspace">
  To tag a workspace with an environment, first, the Cluster where the workspace resides needs to have those environments added.

  <iframe provider="app.supademo.com" href="https://app.supademo.com/embed/_lxNNVAU_Po5eXicIXIHF" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/_lxNNVAU_Po5eXicIXIHF" style={{ border: "none", display: "flex", margin: "auto" }} />

  For this, you will have to add all environments relevant to your cluster (one cluster can have multiple environments) in the cluster, using the instructions provided below.

  <iframe provider="app.supademo.com" href="https://app.supademo.com/embed/WqCoUneHoVpxGdDuIMBVQ" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/WqCoUneHoVpxGdDuIMBVQ" style={{ border: "none", display: "flex", margin: "auto" }} />

  Now all your applications deployed within that specific workspace will have the environment of the workspace show up beside them.
</Accordion>

## FQN

FQN stands for Fully Qualified Name and is a more human-readable version of the id in the TrueFoundry platform. Every entity - be it a cluster, workspace, respository, secret, model has a FQN. You can copy the FQN from the UI for every entity.

<Accordion title="Get Workspace FQN">
  In the Workspace section, locate your workspace and then click the FQN button on the right side to copy the FQN to your clipboard

  <Frame caption="">
    <img src="https://mintcdn.com/truefoundry/s4Aj2_qGCrSP-zc8/images/9cf3922e-4d56ff5cc0312af8be61ec7912fe26adce50ac0495cf3dad7fac512017b1d244-image.png?fit=max&auto=format&n=s4Aj2_qGCrSP-zc8&q=85&s=e907c3eb99d60b3788071a26d46f8d6c" width="3010" height="1158" data-path="images/9cf3922e-4d56ff5cc0312af8be61ec7912fe26adce50ac0495cf3dad7fac512017b1d244-image.png" />
  </Frame>
</Accordion>

***
