> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Deploy Programatically

> Deploy jobs programmatically using APIs. Automate job creation and execution.

Truefoundry UI makes it really easy to configure all the settings for your job by providing descriptions of all the fields in an intuitive way. However, there are times when you might want to deploy your job programatically. You can do this by using the Truefoundry CLI or Python SDK.

<Tip>
  Truefoundry UI itself shows the instructions to deploy the job using the CLI or UI.

  **So we recommend to first deploy the job using UI and then copy the YAML or Python code generated to deploy programatically**. You can also copy the spec / Python code from an existing job and make the changes to deploy your job.
</Tip>

<Tabs>
  <Tab title="Deploy using YAML">
    <Steps>
      <Step title="Click on Edit Button on a job">
        <img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/edit-button-job.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=2f8bc23c5311a33c2781928e8c64e0ee" alt="Deploy using YAML" width="3024" height="1714" data-path="images/edit-button-job.png" />
      </Step>

      <Step title="Click on Deploy Using YAML button">
        <img src="https://mintcdn.com/truefoundry/PSBc0bX31_cIC7pm/images/deploy-programatically-yaml-button-job.png?fit=max&auto=format&n=PSBc0bX31_cIC7pm&q=85&s=189bfcd62f920c0b068c0dcc62f00164" alt="Deploy using YAML" width="3020" height="1708" data-path="images/deploy-programatically-yaml-button-job.png" />
      </Step>

      <Step title="Copy the YAML code and follow the instructions">
        You will need to install the Truefoundry CLI, login and then execute the `tfy apply` command to deploy the job.

        <img src="https://mintcdn.com/truefoundry/PSBc0bX31_cIC7pm/images/deploy-programatically-yaml-job.png?fit=max&auto=format&n=PSBc0bX31_cIC7pm&q=85&s=02feb9dcb8ebd6d17380922dde792147" alt="Deploy using YAML" width="3022" height="1714" data-path="images/deploy-programatically-yaml-job.png" />
      </Step>
    </Steps>
  </Tab>

  <Tab title="Deploy using Python SDK">
    <Steps>
      <Step title="Click on Edit button on a job">
        <img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/edit-button-job.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=2f8bc23c5311a33c2781928e8c64e0ee" alt="Deploy using YAML" width="3024" height="1714" data-path="images/edit-button-job.png" />
      </Step>

      <Step title="Click on Deploy Using Python button">
        <img src="https://mintcdn.com/truefoundry/PSBc0bX31_cIC7pm/images/deploy-programatically-python-button-job.png?fit=max&auto=format&n=PSBc0bX31_cIC7pm&q=85&s=20363ca68a877d2daf60fb79429b632e" alt="Deploy using YAML" width="3022" height="1714" data-path="images/deploy-programatically-python-button-job.png" />
      </Step>

      <Step title="Copy the Python code and follow the instructions">
        You will need to install the Truefoundry CLI, login, put the Python code in `deploy.py` file and then execute the `python deploy.py` command to deploy the job.

        <img src="https://mintcdn.com/truefoundry/PSBc0bX31_cIC7pm/images/deploy-programatically-python-job.png?fit=max&auto=format&n=PSBc0bX31_cIC7pm&q=85&s=69a55ca3e4ab8a3c94f1ab2a18b7cd6f" alt="Deploy using YAML" width="3024" height="1712" data-path="images/deploy-programatically-python-job.png" />
      </Step>
    </Steps>
  </Tab>
</Tabs>
