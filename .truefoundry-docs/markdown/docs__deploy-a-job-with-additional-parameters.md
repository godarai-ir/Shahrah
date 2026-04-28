> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Parameterize A Job

> Deploy jobs with custom parameters. Learn flags, overrides, and runtime configuration.

When you develop a Python script, there are often instances where you want to alter the script's behaviour without having to rewrite the code. This is where arguments come in handy. Arguments allow you to pass values to a script when it is executed, and these values can then be used to control the script's behaviour.

For instance, let's consider a script that trains a Support Vector Machine (SVM) model. Arguments can be used to pass the values of the hyper-parameters to your script as command-line arguments. This enables you to easily modify the values of the hyper-parameters without having to modify your code.

Here's an example of how to use argparse to pass hyper-parameters to a Python script that trains an SVM model:

<CodeGroup>
  ```python train.py lines theme={"dark"}
  import argparse

  parser = argparse.ArgumentParser()
  parser.add_argument("--kernel", type=str, default="rbf", help="The kernel type")
  parser.add_argument("--C", type=float, default=1.0, help="The regularization parameter")

  args = parser.parse_args()

  kernel = args.kernel
  C = args.C

  # Train the model
  train_model(kernel, C)

  # Evaluate the model
  evaluate_model()
  ```
</CodeGroup>

You can now run this script by passing the following command:

<CodeGroup>
  ```shell Terminal lines theme={"dark"}
  python train.py --kernel linear --C 2.0
  ```
</CodeGroup>

This will train the SVM model with a linear kernel and a regularisation parameter of 2.0. You can easily set the kernel type and C parameter without having to change the code repeatedly.

## Configuring a Job with Params

To set up Params, follow these steps:

* **Setup the params in job form**. Here, you'll provide details regarding the arguments you want to make params. You'll need to provide the name, description, and default value for each parameter.
* **In the command you specify while configuring the Job, you need to specify a command template**. It will be like the following:

<CodeGroup>
  ```shell PowerShell lines theme={"dark"}
  python script_name --argument_1_name {{param_1_name}} --argument_2_name {{param_2_name}}
  ```
</CodeGroup>

For example, `python train.py --C {{c_val}} --kernel {{kernel}}.`

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/1jWFbZn1Iu9fQN0kvl5U0" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/1jWFbZn1Iu9fQN0kvl5U0" style={{ border: "none", display: "flex", margin: "auto" }} />

## Running a Job with params

Once you configure the job with params and deploy it, you will start seeing a form rendered on clicking run job.

TrueFoundry automatically generates this form based on the parameters defined in the job's configuration. This dynamic form simplifies parameter handling, where you just need to change the values for the parameters and the command updates accordingly.

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/zF-TI6wFFY2niB4Gvk3S0" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/zF-TI6wFFY2niB4Gvk3S0" style={{ border: "none", display: "flex", margin: "auto" }} />

***
