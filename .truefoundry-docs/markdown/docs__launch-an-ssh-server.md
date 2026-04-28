> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Launch a SSH Server

> Launch an SSH server for debugging and secure access.

If you like controlling your coding setup, consider using your own local VS Code instead of the hosted version. By setting up an SSH Server, you can run your code directly on your server, tapping into its robust cloud resources. This way, you maintain full control over your development environment, customizing VS Code with your favourite extensions and settings.

<iframe href="https://app.supademo.com/embed/aJOSWs5z8lqkcviLScT7Q" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/aJOSWs5z8lqkcviLScT7Q" style={{ border:"none",display:"flex",margin:"auto" }} />

## How to find your SSH Public Key?

Your SSH public key is typically stored in the `~/.ssh/` directory on your local machine. Look for a file ending in `.pub`, which contains the public key. For example:

* `id_rsa.pub`(default key)

You can find your public key using the following command:

**For Windows**: `type $home\.ssh\id_rsa.pub`( windows powershell)

**For Mac/Linux**: `cat ~/.ssh/id_rsa.pub`

## Creating a New SSH Key Pair

If you don't have an SSH Key Pair. You can generate one using the command:

```shell lines theme={"dark"}
ssh-keygen -t rsa
```

You can refer to this [Link](https://docs.oracle.com/en/cloud/cloud-at-customer/occ-get-started/generate-ssh-key-pair.html#GUID-8B9E7FCB-CEA3-4FB3-BF1A-FD3406A2432F) for generating an SSH Key Pair.

## Installing ProxyTunnel

To install [`proxytunnel`](https://github.com/proxytunnel/proxytunnel/releases) , follow the steps below for your Operating System:

### Windows

Open PowerShell and run the following command:

<CodeGroup>
  ```powershell bash lines theme={"dark"}
  Invoke-Expression (Invoke-WebRequest -Uri "https://raw.githubusercontent.com/truefoundry/infra-charts/refs/heads/main/scripts/windows_proxytunnel_install.ps1" -UseBasicParsing).Content
  ```
</CodeGroup>

### MacOS

<CodeGroup>
  ```shell bash lines theme={"dark"}
  brew install proxytunnel
  ```
</CodeGroup>

### Ubuntu

<CodeGroup>
  ```shell bash lines theme={"dark"}
  sudo apt-get update
  sudo apt-get install proxy-tunnel
  ```
</CodeGroup>

### Other Linux Distros

Find `proxytunnel` in your package manager respositories\
Or Build from Source: [https://github.com/proxytunnel/proxytunnel](https://github.com/proxytunnel/proxytunnel)

<Tip>
  Cannot install `proxytunnel`? Use `nc` on Linux or Mac

  <Accordion title="Using `nc` to proxy instead of `proxytunnel`">
    <Steps>
      <Step title="Add `~/ssh/nc-proxy.sh`">
        ```shell theme={"dark"}
        #!/bin/bash
        PROXY="$1"
        DEST="$2"
        PROXY_HOST="${PROXY%%:*}"
        PROXY_PORT="${PROXY##*:}"
        DEST_HOST="${DEST%%:*}"
        DEST_PORT="${DEST##*:}"

        # Send headers, then keep stdin open for SSH
        {
            printf "CONNECT %s:%s HTTP/1.1\r\n" "$DEST_HOST" "$DEST_PORT"
            printf "Host: %s\r\n" "$PROXY_HOST"
            printf "Proxy-Connection: Keep-Alive\r\n"
            printf "\r\n"
            # keep stdin open indefinitely
            cat
        } | exec nc "$PROXY_HOST" "$PROXY_PORT"
        ```
      </Step>

      <Step title="Modify the `~/ssh/config` to use the above script">
        Here we replace the `proxytunnel` invocation with `~/.ssh/nc-proxy.sh`

        ```diff lines theme={"dark"}
        Host your-ssh-server
          ...
        - ProxyCommand proxytunnel -v -p <your-ssh-host>:80 -d %h:%p
        + ProxyCommand ~/.ssh/nc-proxy.sh <your-ssh-host>:80 %h:%p
        ```
      </Step>
    </Steps>
  </Accordion>
</Tip>

## Connect Local VS Code with SSH

To connect Visual Studio Code to a remote SSH server, follow these steps:

1. Install the extension: [Remote-SSH](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-ssh)
2. After installing the extension, reload VS Code to ensure all changes are applied.
3. Once VS Code reloads, you’ll see a new icon labeled `Remote Explorer` in the activity bar on the left side. Click on this icon to manage your SSH connections.
4. Enter your password or passphrase to complete the connection.

<Frame caption>
  <img src="https://mintcdn.com/truefoundry/JFTbQOWMkMfvFjDC/images/b2a70643-ed115dc856a458a845283261330e4c3679e12c7e5503038b12d41681b51705d0-image.png?fit=max&auto=format&n=JFTbQOWMkMfvFjDC&q=85&s=f34c697267ed8017a13641948a7146ab" alt="" width="1365" height="613" data-path="images/b2a70643-ed115dc856a458a845283261330e4c3679e12c7e5503038b12d41681b51705d0-image.png" />
</Frame>

## Adding more users to SSH Server

To add more users to the SSH Server. You need to add the SSH keys to the users to the `mkdir -p /home/jovyan/.ssh/authorized_keys`

<CodeGroup>
  ```bash bash lines theme={"dark"}
  mkdir -p /home/jovyan/.ssh

  echo "ENTER_PUBLIC_KEY" >> /home/jovyan/.ssh/authorized_keys
  ```
</CodeGroup>

You can add multiple users with this. You can also remove users from this file as required.

## Copying Files from SSH Server to Local Machine

To copy files from an SSH server to your local machine, use the following command in PowerShell. Make sure to update the placeholders with the correct values for your environment.

<CodeGroup>
  ```shell bash lines theme={"dark"}
  scp -r <deploymentName>:<remote-file-path> <local-file-path>
  ```
</CodeGroup>

`<deploymentName>`: The name of the SSH server on the platform\
`<remote-file-path>`: The path to the file or directory on the SSH server that you want to copy. `<local-file-path>`: The destination path on your local machine where you want to save the copied file.

## Copying Files from Local Machine to SSH Server

To copy files from your local machine to an SSH server, use the following command in PowerShell. Make sure to update the placeholders with the correct values for your environment.

<CodeGroup>
  ```shell bash lines theme={"dark"}
  scp -r <local-file-path> {{deploymentName}}:<remote-file-path>
  ```
</CodeGroup>

`<deploymentName>`: The name of the SSH server on the platform\
`<remote-file-path>`: The destination path on the SSH server where you want to save the copied file. `<local-file-path>`: The path to the file or directory on your local machine that you want to copy.

### Using `rsync` to Copy Files Between Local Machine and SSH Server

`rsync` is a tool for synchronizing files and directories between a local machine and a remote server via SSH. It only transfers the changes made to files, making it efficient for backups and file synchronization.

## Copying Files from Local Machine to SSH Server

To copy files from your local machine to an SSH server using `rsync`, use the following command:

<CodeGroup>
  ```shell bash lines theme={"dark"}
  rsync -avz <local-file-path> <deploymentName>:<remote-file-path>
  ```
</CodeGroup>

`<deploymentName>`: The name of the SSH server on the platform\
`<remote-file-path>`: The path to the file or directory on the SSH server that you want to copy. `<local-file-path>`: The destination path on your local machine where you want to save the copied file.

1. The `-a` flag enables archive mode, preserving file permissions and timestamps.
2. The `-v` flag enables verbose output, so you can see the progress of the transfer.
3. The `-z` flag enables compression during transfer, which can speed up the process for larger files.

## Copying Files from SSH Server to Local Machine

To copy files from an SSH server to your local machine using `rsync`, use the following command:

```shell lines theme={"dark"}
rsync -avz {{deploymentName}}:<remote-file-path>:<remote-file-path> <local-file-path>
```

`<deploymentName>`: The name of the SSH server on the platform\
`<remote-file-path>`: The path to the file or directory on the SSH server that you want to copy. `<local-file-path>`: The destination path on your local machine where you want to save the copied file.

1. The `-a` flag enables archive mode, preserving file permissions and timestamps.
2. The `-v` flag enables verbose output, so you can see the progress of the transfer.
3. The `-z` flag enables compression during transfer, which can speed up the process for larger files.

## Create More Python Environments

You can create python version with the following commands:

<CodeGroup>
  ```shell she lines theme={"dark"}
  conda create -n my-new-env

  # run the following command to create python environment with different python version
  conda create -n py39-env python=3.9
  ```
</CodeGroup>

## Install apt packages in the SSH Server

By default anything you install outside of home directory will NOT be persistent across restarts. This means all apt installs (done directly from ssh server) will NOT be persistent across restarts.

To make these apt packages persistent across restarts, you can add a "Build Script" in the SSH Server as follows

<Frame caption>
  <img src="https://mintcdn.com/truefoundry/DdP_2rhue4AQQlob/images/44dee676-02724d71acaa04fb736578e31f6a26656edfd31b32fee3d19000212d6b5ea3be-image.png?fit=max&auto=format&n=DdP_2rhue4AQQlob&q=85&s=7c63e473777c0293c9fbbde746c4fe27" alt="" width="1810" height="956" data-path="images/44dee676-02724d71acaa04fb736578e31f6a26656edfd31b32fee3d19000212d6b5ea3be-image.png" />
</Frame>

For example, here is a sample build script to install ffmpeg package:

```bash lines theme={"dark"}
sudo apt update
sudo apt install ffmpeg
```

## Setting up the scale to zero

* You can now set up the scale to zero feature for the SSH server to stop the server after X minutes of inactivity.
* You can add this by setting the value of Stop after field which is the minutes of inactivity after which you want to stop the server.

<Warning>
  ### Scale to zero is only available for ssh server images with tag >= v0.3.10
</Warning>

<Frame caption>
  <img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/f12c58a1-6aa4f5d1ddf1b3edb5b33ba7ecd42f463406c6e7fc940970e8ae0949d89c961b-image.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=9b441a84dae6af246b5ca62cdbb67874" alt="" width="1138" height="178" data-path="images/f12c58a1-6aa4f5d1ddf1b3edb5b33ba7ecd42f463406c6e7fc940970e8ae0949d89c961b-image.png" />
</Frame>

**Following conditions will be considered as an activity while checking the inactivity status**

* An active SSH connection(including connection to vscode).
* Any application or service running in the foreground.
* The applications running(both background and foreground) in tmux or screen session.

**Following conditions will not be considered as activity**

* Applications running in the background in SSH server.

for example if you run the below command, then your service will be running in background and it will be considered as an inactivity.

```shell lines theme={"dark"}
python main.py&
```

## Launching SSH Server Custom Images

Instead of using the pre-built SSH Server Images provided by TrueFoundry, you have the flexibility to create and deploy your custom images. This allows you to pre-install specific libraries, tools, and configurations within your SSH Server environment, tailoring it to your specific needs and project requirements.

To create a custom image, start by using the TrueFoundry SSH Server Image as a base. Below are the latest images at the time of writing

| Image URI                                                          | Size     | Python Env | CUDA 12.9 Toolkit |
| ------------------------------------------------------------------ | -------- | ---------- | ----------------- |
| `public.ecr.aws/truefoundrycloud/ssh-server:0.4.6-py3.12.12`       | \~0.8 GB | ✅          |                   |
| `public.ecr.aws/truefoundrycloud/ssh-server:0.4.6-py3.11.14`       | \~0.8 GB | ✅          |                   |
| `public.ecr.aws/truefoundrycloud/ssh-server:0.4.6-py3.10.19`       | \~0.8 GB | ✅          |                   |
| `public.ecr.aws/truefoundrycloud/ssh-server:0.4.6-cu129-py3.12.12` | \~7 GB   | ✅          | ✅                 |
| `public.ecr.aws/truefoundrycloud/ssh-server:0.4.6-cu129-py3.11.14` | \~7 GB   | ✅          | ✅                 |
| `public.ecr.aws/truefoundrycloud/ssh-server:0.4.6-cu129-py3.10.19` | \~7 GB   | ✅          | ✅                 |

<Info>
  ### All Image Tags

  SSH Server Images are available at [https://gallery.ecr.aws/truefoundrycloud/ssh-server](https://gallery.ecr.aws/truefoundrycloud/ssh-server)
</Info>

You can use one of these images as a base for creating your custom image while keeping some invariants unchanged.

Let's take an example where we want to customize the image by including a few apt packages like `ffmpeg` and a few pip packages like `gradio`. We will create a new `Dockerfile` using these existing images as the base and push to a registry for later use.

<CodeGroup>
  ```python docker lines theme={"dark"}
  FROM public.ecr.aws/truefoundrycloud/ssh-server:0.4.6-py3.11.14

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

Build and push the image to a registry for it to be used in TrueFoundry SSH Server. Make sure the destination registry is already integrated with the TrueFoundry platform. Detailed instructions are available [here](/docs/integrations-overview)

### Example: Installing CUDA 11.8 and cuDNN 8

<CodeGroup>
  ```dockerfile Dockerfile lines theme={"dark"}
  FROM public.ecr.aws/truefoundrycloud/ssh-server:0.4.6-py3.11.14
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
