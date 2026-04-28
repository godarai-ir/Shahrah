> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Dockerize Your Code

> Learn how to dockerize your code for deployment on TrueFoundry.

Its a good practice to always dockerize your code while deploying so that it is guaranteed to run anywhere. TrueFoundry helps you deploy your code on Kubernetes for which you will first need to create a docker image for your codebase.

<Note>
  For any deployment to work in TrueFoundry, its essential to have a docker registry integrated with TrueFoundry. This should have been already done by the Infra team during the [setup step](/docs/infrastructure/deploy-compute-plane).
</Note>

TrueFoundry can help you deploy in all the three use cases:

<Accordion title="You already have a docker image built and pushed to your registry.">
  To deploy a already present docker image, choose the option of `Docker Image` in the Service Deployment form.

  <iframe href="https://app.supademo.com/embed/VVySYiv-rweC-QNkEJkLX" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/VVySYiv-rweC-QNkEJkLX" style={{ border: "none", display: "flex", margin: "auto" }} />

  Choosing the docker registry is optional - TrueFoundry will automatically try to check if the image is present in any of the integrated registries.
</Accordion>

<Accordion title="Your code is in a git repository and the Git repository is integrated to TrueFoundry. You have written a dockerfile to build your code">
  1. Select the Git Repo option in the Service Deployment form

  <img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/docs/select-git-repo-option.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=6cbcfe15e266b6d68410a84091b63438" width="3024" height="1714" data-path="images/docs/select-git-repo-option.png" />

  2. Choose your repository in the Repo URL box and the option `Dockerfile (I already have DockerFile)` in the Build section.

  You will need to fill up 2 field here:

  1. **Path to Build Context**: This is the path in the Git repository where the code to be deployed is present. Usually this is the root directory of the Git repo and the value is `"./"`, however, if you have multiple services in the same repo, you can specify the path to the directory containing the service you want to deploy - for e.g. `"my-service/"`.
  2. **Path to Dockerfile**: This is the path to the Dockerfile in the Git repository - relative to the `Path to Build Context`. For e.g. if the Dockerfile is present in the `my-service` directory, and you have provided the build context as `"my-service/"`, then the path to the Dockerfile is `"./Dockerfile"`. However, if the `Path to Build Context` is `"./"`, then the path to the Dockerfile should be `"./my-service/Dockerfile"`.

  <img src="https://mintcdn.com/truefoundry/5CkapnZ7CyjQJ4bx/images/docs/fill-git-repo-and-docker-file.png?fit=max&auto=format&n=5CkapnZ7CyjQJ4bx&q=85&s=20528104b5567c03d7527e46aebfc65a" width="3024" height="1712" data-path="images/docs/fill-git-repo-and-docker-file.png" />
</Accordion>

<Accordion title="Your code is in a git repository and the git repository is integrated with TrueFoundry. You don't have a Dockerfile in the Git repository">
  <Note> This option only works for Python code that uses requirements.txt to manage dependencies. For other frameworks and languages, you will need to write a Dockerfile and follow the option 2 above. </Note>

  1. Select the Git Repo option in the Service Deployment form

  <img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/docs/select-git-repo-option.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=6cbcfe15e266b6d68410a84091b63438" width="3024" height="1714" data-path="images/docs/select-git-repo-option.png" />

  2. Choose your repository in the Repo URL box and the option `Python Code (I don't have Dockerfile)` in the Build section.

  You will need to fill up the following fields here:

  1. **Path to Build Context**: This is the path in the Git repository where the code to be deployed is present. Usually this is the root directory of the Git repo and the value is `"./"`, however, if you have multiple services in the same repo, you can specify the path to the directory containing the service you want to deploy - for e.g. `"my-service/"`.
  2. **Path to requirements.txt**: This is the path to requirements.txt in the Git repository - relative to the `Path to Build Context`. For e.g. if the requirements.txt is present in the `my-service` directory, and you have provided the build context as `"my-service/"`, then the path to the requirements.txt is `"./requirements.txt"`. However, if the `Path to Build Context` is `"./"`, then the path to the requirements.txt should be `"./my-service/requirements.txt"`.
  3. **Python Version**: This is the Python version to be used to build the image. This should be the same Python version using which the code has been written and tested.
  4. **Command**: This is the command to be used to run the code. This should be the same command which you are using to run the code locally.

  <img src="https://mintcdn.com/truefoundry/5CkapnZ7CyjQJ4bx/images/docs/fill-git-repo-python-code.png?fit=max&auto=format&n=5CkapnZ7CyjQJ4bx&q=85&s=c4558dd98b50683c5d8c9f88a4a7a3e8" width="2876" height="1714" data-path="images/docs/fill-git-repo-python-code.png" />
</Accordion>

<Accordion title="Your code is only on your laptop and you want to deploy it">
  1. Select the option `Code From Laptop` in the Service Deployment form.

  <img src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/choose-deploy-code-option.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=0040fdb7f6139f99f23549b6b49d01fc" width="3024" height="1716" data-path="images/docs/choose-deploy-code-option.png" />

  2. Follow the steps to install truefoundry cli and login to truefoundry.

  <img src="https://mintcdn.com/truefoundry/5CkapnZ7CyjQJ4bx/images/docs/install-cli-login.png?fit=max&auto=format&n=5CkapnZ7CyjQJ4bx&q=85&s=fc074baa8699c2e695194f87d7f562c0" width="2876" height="1688" data-path="images/docs/install-cli-login.png" />

  3. Fill up the deployment form with the following key fields:

  * **Service Name**: You can give any name which will later be used to identify the service in the TrueFoundry platform.
  * **Source Code (Build and deploy Source Code)**: If you have already written a Dockerfile locally, choose the option `Dockerfile (I already have DockerFile)`, else choose the option `Python Code (I don't have Dockerfile)`. The second option will work if you have Python code locally and a requriements.txt file containing the dependencies. The description of the fields can be found in the previous sections.
  * **Ports**: This is the port on which the service will be exposed. It will be the same port on which the service is running locally.
  * **Resources**: You can start with the default values and later change after the first deployment.

  4. Download the deploy.py file, place it in the root of your project. Run it from the same directory where deploy.py is present. The command to run it will be:

  ```
  python deploy.py
  ```

  The TrueFoundry CLI will build the image locally if Docker is installed in the environment, else it will build the image remotely on the TrueFoundry platform. It will then push the image to the docker registry and deploy the service on Kubernetes.
</Accordion>
