> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Introduction to Volume

> Understand how volumes work for persistent storage, shared data, and model checkpointing in TrueFoundry deployments.

Volumes are used to provide persistent storage to containers so that read and write data to a central disk across multiple pods. They are especially useful in the context of machine learning when you need to store and access data, models, and other artifacts required for training, serving, and inference tasks.

A few usecases where volumes turn out to be very useful:

1. **Share training data:** Its possible multiple datascientists are training on the same data or we are running multiple experiments in parallel on the same dataset. The naive way will be to duplicate the data for multiple data scientists - however this will end up costing us much more. A more efficient way here will be to store the training data in a volume and mount the volume to the notebooks of different datascientists.
2. **Model Storage**: If we are hosting models as realtime APIs, there will be multiple replicas of the api server to handle traffic. Here, every replica will need to download the model from the model registry (let's say S3) to local disk. If every replica does this repeatedly, it will take more time for starting up and also incur more S3 access cost. By using volumes, you can store your trained models externally and mount them onto the inference server.
3. **Checkpointing**: During the training of machine learning models, it's common to save checkpoints periodically to resume training in case of failure or to fine-tune models. Volumes can be used to store these checkpoint files, ensuring that training progress is not lost when a job restarts from failure. This also enables you to run training on spot instances, hence saving a lot of cost.

### When to use Volume vs Blob storage like S3 / GCS / Azure Container?

You can store data in volumes or in blob storage like S3/GCS/Azure container. The key differences between the two are:

| Feature            | Volume                                             | Blob Storage                                                          |
| ------------------ | -------------------------------------------------- | --------------------------------------------------------------------- |
| Access in Code     | Mount volume and use standard file read/write APIs | Use Client SDK like boto3, gcs, azure-storage-blob to read/write data |
| Performance        | Faster read/write speeds                           | Slower read/write speeds                                              |
| Durability         | High                                               | Extremely high ((11 9's))                                             |
| Storage Cost       | Higher cost                                        | Lower cost                                                            |
| Access Cost        | Performance-based (throughput & IOPS)              | Request-based (GET, PUT, LIST, etc.)                                  |
| Access Constraints | Limited to same region and cluster                 | Global access                                                         |

Choose volume like EFS if:

* You need a shared file system accessible by multiple instances.
* Your application requires file system semantics (e.g., file locking, renaming).
* You need low-latency access to files, especially for frequently accessed data.

<Warning>
  We should be careful about not writing to the volume at the same path from multiple pods since it can cause data corruption.
</Warning>

Choose S3 if:

* You need to store large amounts of unstructured data.
* You don't need file system semantics.
* You need to archive data for long-term storage.
* Cost is a primary concern, especially for infrequently accessed data.
