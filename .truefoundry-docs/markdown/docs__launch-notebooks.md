> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Launch Jupyter Notebook

> Launch notebooks for experimentation and development.

Jupyter Notebooks are probably the first tool to be used in any ML Project. TrueFoundry enables you to run Jupyter Notebooks on Kubernetes on any hardware you need, and auto-shut down when it's not being used hence saving costs. You can read more about TrueFoundry Notebooks vs other notebook solutions [here](https://blog.truefoundry.com/hosted-jupyter-notebooks-vs-code-on-kubernetes/).

To get started with launching a Jupyter Notebook, you can follow the steps outlined below (You can find details on the configuration options below)

<iframe href="https://app.supademo.com/embed/cmb88nai65o0rppkpwe7cs8m7?embed_v=2" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/cmb88nai65o0rppkpwe7cs8m7?embed_v=2" style={{ border:"none",display:"flex",margin:"auto" }} />

The notebook will go live in a few minutes. It takes a few minutes since the image will need to be downloaded on the machine first. Once the notebook becomes live, you will see the status turn to Running and an endpoint will start showing up. Clicking on the endpoint will take you to the Jupyter Notebook.

<Frame caption>
  <img src="https://mintcdn.com/truefoundry/DdP_2rhue4AQQlob/images/33169a82-1de8483-Screenshot_2023-11-22_at_11.17.52_AM.png?fit=max&auto=format&n=DdP_2rhue4AQQlob&q=85&s=8b78c13a5936bf3e394535167682271a" alt="" width="2940" height="519" data-path="images/33169a82-1de8483-Screenshot_2023-11-22_at_11.17.52_AM.png" />
</Frame>

## Stop and resume your notebook

When you're done working with a notebook instance, you can stop it to conserve resources and reduce costs. Clicking on the **Stop** button shuts down the notebook instance environment and releases associated resources.

<Frame caption>
  <img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/f5c49fe5-a4db272-Screenshot_2023-11-21_at_3.41.26_PM.png?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=d0dadc64dad5c56a417c41b06164cdd1" alt="" width="2940" height="559" data-path="images/f5c49fe5-a4db272-Screenshot_2023-11-21_at_3.41.26_PM.png" />
</Frame>

Your notebook instance data (apart from the apt packages installed as a root user) persists, and you can easily restart it later by clicking the **Resume** button.

## Configure your Notebook

In the notebook creation form, you can configure the following options:

### Image

When creating a Jupyter Notebook, you can choose between three image options:

* **Minimal Image**: A lightweight image with no pre-installed packages, providing a clean slate for customization and package installation.
* **Cuda 12.9 Image**: A pre-configured image with Cuda 12.0 pre-installed
* **Custom Extended Image**: You can extend an Image created by truefoundry and add your custom packages to it. You can read about extended images [here.](/docs/launch-notebooks#launching-jupyter-notebook-with-custom-images)

<Frame caption>
  <img src="https://mintcdn.com/truefoundry/4MAaF__cLD4iud16/images/5343d339-4242d3ab1a2c7d7218f07876ddd3c620ade4688606c2f47f7e9752f7224a863c-image.png?fit=max&auto=format&n=4MAaF__cLD4iud16&q=85&s=0d2f22b8992530bd818c2d6870160c4e" alt="" width="1812" height="440" data-path="images/5343d339-4242d3ab1a2c7d7218f07876ddd3c620ade4688606c2f47f7e9752f7224a863c-image.png" />
</Frame>

### Auto Shutdown on Inactivity

By default, Notebook instances are configured to automatically stop after 30 minutes of inactivity. This helps prevent unnecessary resource consumption when Notebook instances are left idle. You can change the **Stop After (minutes of inactivity)** setting within the deployment form.

### Install Apt packages in the notebook

By default anything you install outside of home directory will NOT be persistent across notebook restarts. This means all apt installs (done directly from notebook) will NOT be persistent across restarts of notebook.

To make these apt packages persistent across restarts, you can add a "Build Script" in the notebook as follows

<Frame caption>
  <img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/f973a62f-673df34057fda6a212b06cff88fbe7e6355e124e99c1a230c08601aeda24817b-image.png?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=620d3615ec8a528decb96942870e05ed" alt="" width="1806" height="966" data-path="images/f973a62f-673df34057fda6a212b06cff88fbe7e6355e124e99c1a230c08601aeda24817b-image.png" />
</Frame>

For example, here is a sample build script to install ffmpeg package:

```bash lines theme={"dark"}
sudo apt update
sudo apt install ffmpeg
```

### Storage Size

Specify the amount of storage space you require for your notebook instance. This is persistent storage that will be used to store your notebook files, data, and any other artifacts generated during your work.

### Set resources for your Notebook

Define the computational resources allocated to your notebook instance. You can adjust the CPU and memory allocation to meet the requirements of your data science tasks.

#### Running a Notebook instance with GPU

In case you want to use GPU in your Notebook Instance, you can follow these steps:

<iframe href="https://app.supademo.com/embed/Q7kAJHEL8Z28STINIDCtO" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/Q7kAJHEL8Z28STINIDCtO" style={{ border:"none",display:"flex",margin:"auto" }} />

Upon successful deployment, your GPU Notebook instance instance will be provisioned with the specified GPU type. You can then utilize the GPU resources to accelerate your computations, such as training deep learning models or running GPU-intensive workloads.

### Access data from S3 or other clouds

In some instances, your Jupyter Notebooks may need to access data stored in S3 or other cloud storage platforms. To facilitate this access, you can employ one of two approaches:

#### Credential-Based Access through environment variables

This approach involves defining specific environment variables that contain the necessary credentials for accessing the cloud storage platform. For instance, to access S3, you would set environment variables for the AWS access key ID and secret access key, the environment variables being: `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`

<iframe href="https://app.supademo.com/embed/ZTKqUvw8CaR2TzI6CjQ4w" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/ZTKqUvw8CaR2TzI6CjQ4w" style={{ border:"none",display:"flex",margin:"auto" }} />

#### IAM Role-Based Access through Service Account

The second approach is to provide your Notebook with a Role with the necessary permission through Service Accounts.

Service Accounts provide a streamlined approach to managing access without the need for tokens or complex authentication methods. This approach involves creating a Principal IAM Role within your cloud platform, granting it the necessary permissions for your project's requirements. Here are detailed guides for creating Principal IAM Roles in your respective cloud platforms and integrating them as Service Accounts within the workspace:

* AWS: [Authenticate to AWS services using IAM service account](/docs/use-aws-services-using-iam-serviceaccount)
* GCP: [Authenticate to GCP using IAM serviceaccount](/docs/authenticate-to-gcp-using-iam-serviceaccount)

Once you've configured the Service Account within the workspace, you can simply toggle the **`Show Advanced Fields`** option at the bottom of the form. This will reveal an expanded set of options, from which you can select the desired Service Account using the provided dropdown menu.

<Frame caption>
  <img src="https://mintcdn.com/truefoundry/DdP_2rhue4AQQlob/images/387e68cd-b561def-Screenshot_2023-11-28_at_12.36.50_AM.png?fit=max&auto=format&n=DdP_2rhue4AQQlob&q=85&s=5e70e3e2084e10a4c542b9fdf42645d9" alt="" width="2350" height="284" data-path="images/387e68cd-b561def-Screenshot_2023-11-28_at_12.36.50_AM.png" />
</Frame>

### Access data in volume from the notebook

Mounting a volume to a notebook allows you to access data stored on the volume from within the notebook environment. This can be useful for a variety of tasks, such as loading data for analysis, training machine learning models, and deploying applications.

You can follow these steps for mounting a volume to the notebook:

<iframe href="https://app.supademo.com/embed/9vViV6VlnuN44e9Wj5fL2" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/9vViV6VlnuN44e9Wj5fL2" style={{ border:"none",display:"flex",margin:"auto" }} />

Once mounted, in your Notebook you will be able to access the data in your Volume from within the Notebook

<Frame caption>
  <img src="https://mintcdn.com/truefoundry/4MAaF__cLD4iud16/images/52ecfdd1-94eedaf-Screenshot_2023-11-28_at_9.17.38_AM.png?fit=max&auto=format&n=4MAaF__cLD4iud16&q=85&s=660722df268d64d9286f760313d26ba3" alt="" width="2929" height="539" data-path="images/52ecfdd1-94eedaf-Screenshot_2023-11-28_at_9.17.38_AM.png" />
</Frame>

## Configuring Python Environments Inside a Notebook

Truefoundry's Deployed Notebooks by default start with a conda environment with Python Version = 3.11.

In case you are working on several projects simultaneously, and you want to maintain multiple Python environments for different Python versions / different sets of tasks. You can do so by following these steps:

<iframe href="https://app.supademo.com/embed/ugLR9DpLla9iA5KmMS3c-" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/ugLR9DpLla9iA5KmMS3c-" style={{ border:"none",display:"flex",margin:"auto" }} />

Command to be executed:

<CodeGroup>
  ```shell Shell lines theme={"dark"}
  # Create a new conda environment named "py39" with Python 3.9
  # It will take around 2 minutes for the changes to sync
  # After 2 minutes, You need to hard refresh after executing this command for kernel to show up in the listing page
  conda create -y -n py39 python=3.9
  ```
</CodeGroup>

## Launching Jupyter Notebook with Custom Images

Instead of using the pre-built Jupyter Lab Images provided by TrueFoundry, you have the flexibility to create and deploy your custom images. This allows you to pre-install specific libraries, tools, and configurations within your notebook environment, tailoring it to your specific needs and project requirements.

To create a custom image, start by using the TrueFoundry SSH Server Image as a base. Below are the latest images at the time of writing

| Image URI                                                            | Size     | Python Env | CUDA 12.9 Toolkit |
| -------------------------------------------------------------------- | -------- | ---------- | ----------------- |
| `public.ecr.aws/truefoundrycloud/jupyter:0.4.6-py3.12.12-sudo`       | \~0.9 GB | ✅          |                   |
| `public.ecr.aws/truefoundrycloud/jupyter:0.4.6-py3.11.14-sudo`       | \~0.9 GB | ✅          |                   |
| `public.ecr.aws/truefoundrycloud/jupyter:0.4.6-py3.10.19-sudo`       | \~0.9 GB | ✅          |                   |
| `public.ecr.aws/truefoundrycloud/jupyter:0.4.6-cu129-py3.12.12-sudo` | \~7 GB   | ✅          | ✅                 |
| `public.ecr.aws/truefoundrycloud/jupyter:0.4.6-cu129-py3.11.14-sudo` | \~7 GB   | ✅          | ✅                 |
| `public.ecr.aws/truefoundrycloud/jupyter:0.4.6-cu129-py3.10.19-sudo` | \~7 GB   | ✅          | ✅                 |

<Info>
  ### All Image Tags

  Jupyter Images are available at [https://gallery.ecr.aws/truefoundrycloud/jupyter](https://gallery.ecr.aws/truefoundrycloud/jupyter)
</Info>

You can use one of these images as a base for creating your custom image while keeping some invariants unchanged.

Let's take an example where we want to customize the image by including a few apt packages like `ffmpeg` and a few pip packages like `gradio`. We will create a new `Dockerfile` using these existing images as the base and push to a registry for later use.

<CodeGroup>
  ```python docker lines theme={"dark"}
  FROM public.ecr.aws/truefoundrycloud/jupyter:0.4.6-py3.11.14-sudo

  # Install apt packages
  USER root
  RUN DEBIAN_FRONTEND=noninteractive apt install -y --no-install-recommends ffmpeg
  USER jovyan

  # Install pip packages
  RUN python3 -m pip install --use-pep517 --no-cache-dir gradio
  ```
</CodeGroup>

<Warning>
  ### Do not overwrite the ENTRYPOINT or CMD instructions. These are built into the base images and are critical for things to work correctly
</Warning>

Other possible customizations can be found [here](https://docs.docker.com/engine/reference/builder/)

Build and push the image to a registry for it to be used in TrueFoundry Notebook. Make sure the destination registry is already integrated with the TrueFoundry platform. Detailed instructions are available [here](/docs/integrations-overview)

### Example: Installing CUDA 11.8 and cuDNN 8

<CodeGroup>
  ```dockerfile Dockerfile lines theme={"dark"}
  FROM public.ecr.aws/truefoundrycloud/jupyter:0.4.6-py3.11.14-sudo
  ENV TORCH_CUDA_ARCH_LIST="7.0 7.5 8.0 8.6 9.0+PTX"
  ENV DEBIAN_FRONTEND=noninteractive
  USER root
  # Install CUDA 11.8
  RUN apt update && \
      apt install -y --no-install-recommends git curl wget htop && \
      wget https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2004/x86_64/cuda-keyring_1.1-1_all.deb -O /tmp/cuda-keyring_1.1-1_all.deb && \
      dpkg -i /tmp/cuda-keyring_1.1-1_all.deb && \
      apt update && \
      apt install -y --no-install-recommends cuda-toolkit-11-8 libcudnn8=8.9.7.29-1+cuda11.8 libcudnn8-dev=8.9.7.29-1+cuda11.8
  USER jovyan
  ```
</CodeGroup>

## Git Login Instructions for JupyterLab Users

Git integration is already installed for you via the JupyterLab Git extension. This guide outlines how to configure Git and authenticate using two recommended methods.

You can access the extention from the left panel or menu bar from the top as shown below. You can clone your repository directly from this extention or via terminal.

<Frame caption>
  <img src="https://mintcdn.com/truefoundry/DdP_2rhue4AQQlob/images/2daea71a-7fa1a05db1bc2b3f3c773ec5f59ae384632bbdc07ee4acb690207be4dd040992-image.png?fit=max&auto=format&n=DdP_2rhue4AQQlob&q=85&s=d88ced40aeadbfd2e68806abfaab5caf" alt="" width="1470" height="664" data-path="images/2daea71a-7fa1a05db1bc2b3f3c773ec5f59ae384632bbdc07ee4acb690207be4dd040992-image.png" />
</Frame>

Before using Git, set your identity:

```javascript lines theme={"dark"}
git config --global user.name "Your Name"\
git config --global user.email "[your.email@example.com](mailto:your.email@example.com)"
```

#### **Option 1: Personal Access Token (Recommended for GitHub/GitLab)**

GitHub and GitLab have deprecated password authentication. Personal Access Tokens (PATs) are the most secure and recommended way.

Steps:

1. Generate a Token:
   1. For GitHub: [https://github.com/settings/tokens](https://github.com/settings/tokens)
   2. For GitLab: [https://gitlab.com/-/profile/personal\_access\_tokens](https://docs.gitlab.com/user/profile/personal_access_tokens/)
2. Scope: Select at least repo or read/write permissions
3. Clone or Pull a Repository: Use the Git extension in JupyterLab or the terminal:

```shell lines theme={"dark"}
git clone <your private repository url>
```

4. Authenticate: When prompted for a username/password, paste your username in the username section and "token" in password section

#### **Option 2: SSH Key Authentication**

Use SSH for password-less authentication after one-time setup.

Follow platform-specific instructions to generate and add your SSH key:

1. GitHub SSH Help: [https://docs.github.com/en/authentication/connecting-to-github-with-ssh](https://docs.github.com/en/authentication/connecting-to-github-with-ssh)
2. GitLab SSH Guide: [https://docs.gitlab.com/user/ssh/](https://docs.gitlab.com/user/ssh/)

After setup, clone your repositories using the SSH URL:

```bash lines theme={"dark"}
git clone git@github.com:sample-org/sample-repo.git
```
