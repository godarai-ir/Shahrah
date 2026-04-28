> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Use blob storage with your Spark job

> Run Spark jobs with blob storage. Configure data access, storage, and execution.

You may want to use blob storage for eg., AWS S3 with your Spark job for purposes including but not limited to

* Your main application file is present in blob storage
* You have your training data in blob storage
* You want to write your output to blob storage

To use blob storage with your Spark job

* Add the packages required to interact with blob storage. For eg., with AWS S3 you could set the spark config property `spark.jars.packages` as `org.apache.hadoop:hadoop-aws:3.3.4,com.amazonaws:aws-java-sdk-bundle:1.12.262` and Spark will download the packages on its own. Please chose the versions as per your requirement.
* Either use a kubernetes service account that has access to the bucket/container you want to read to or write from OR add your credentials as environment variables and use them in your application. Its recommended to use [secrets](/docs/manage-secrets) to add the credentials as environment variables.
* Use corresponding file uri for eg., for AWS S3 you would use something like `s3a://my-bucket-name/path/to/file`. The prefix varies with the blob store.
