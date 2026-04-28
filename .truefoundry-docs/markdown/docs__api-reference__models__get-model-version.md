> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Get model version

> Get a model version by its ID.



## OpenAPI

````yaml /openapi.json get /api/ml/v1/model-versions/{id}
openapi: 3.1.0
info:
  title: TrueFoundry API
  description: API for TrueFoundry
  version: 0.1.0
  contact: {}
servers:
  - url: https://{controlPlaneURL}
    variables:
      controlPlaneURL:
        default: app.truefoundry.com
        description: Control Plane URL
security: []
tags:
  - name: Agent
  - name: Applications
  - name: Apply
  - name: Artifacts
  - name: Audit Logs
  - name: Clusters
  - name: Jobs
  - name: Logs
  - name: MCP Servers
  - name: MLRepos
  - name: Model Deployments
  - name: Models
  - name: Personal Access Tokens
  - name: Prompts
  - name: Provider Integrations
  - name: SCIM v2
  - name: Secret Groups
  - name: Secrets
  - name: Teams
  - name: Traces
  - name: Users
  - name: Virtual Accounts
  - name: Workspaces
paths:
  /api/ml/v1/model-versions/{id}:
    get:
      tags:
        - Models
      summary: Get model version
      description: Get a model version by its ID.
      operationId: get_model_version_v1_model_versions__id__get
      parameters:
        - required: true
          schema:
            type: string
            title: Id
          name: id
          in: path
      responses:
        '200':
          description: The model version data
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetModelVersionResponse'
        '422':
          description: Validation Error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HTTPValidationError'
      security:
        - jwt: []
components:
  schemas:
    GetModelVersionResponse:
      properties:
        data:
          allOf:
            - $ref: '#/components/schemas/ModelVersion'
          title: Data
          description: The model version data
      type: object
      required:
        - data
      title: GetModelVersionResponse
      x-fern-type-name: GetModelVersionResponse
    HTTPValidationError:
      properties:
        detail:
          items:
            $ref: '#/components/schemas/ValidationError'
          type: array
          title: Detail
      type: object
      title: HTTPValidationError
      x-fern-type-name: HTTPValidationError
    ModelVersion:
      properties:
        id:
          type: string
          title: Id
          description: Unique identifier for the artifact version
        fqn:
          type: string
          title: Fqn
          description: >-
            Fully qualified name of the artifact version in the format
            '{artifact_type}:{tenant_name}/{ml_repo_name}/{artifact_name}:{version}'
        created_by_subject:
          allOf:
            - $ref: '#/components/schemas/Subject'
          title: Created By Subject
          description: >-
            Subject (user, team, or service account) that created this artifact
            version
        created_at:
          type: string
          format: date-time
          title: Created At
          description: Timestamp when the artifact version was created
        updated_at:
          type: string
          format: date-time
          title: Updated At
          description: Timestamp when the artifact version was last updated
        manifest:
          allOf:
            - $ref: '#/components/schemas/ModelManifest'
          title: Manifest
          description: Manifest containing metadata specific to the model version
        ml_repo_id:
          type: string
          title: Ml Repo Id
          description: ID of the ML Repo that this artifact version belongs to
        tags:
          items:
            type: string
          type: array
          title: Tags
          description: >-
            List of tags associated with this artifact version for filtering and
            organization
        version_alias:
          type: string
          title: Version Alias
          description: Optional human-readable version alias (e.g. v1.0.0)
        usage_code_snippet:
          type: string
          title: Usage Code Snippet
          description: Code snippet demonstrating how to use this artifact version
        model_id:
          type: string
          title: Model Id
          description: ID of the parent model that this version belongs to
        metrics:
          items:
            $ref: '#/components/schemas/Metric'
          type: array
          title: Metrics
          description: List of metrics associated with this model version
        deployable:
          type: boolean
          title: Deployable
          description: Whether this model version is ready for deployment
          default: false
      type: object
      required:
        - id
        - fqn
        - created_by_subject
        - manifest
        - ml_repo_id
        - model_id
      title: ModelVersion
      description: >-
        Tags, optional version alias, and SDK usage snippet (models, prompts,
        generic artifacts).
      x-fern-type-name: ModelVersion
    ValidationError:
      properties:
        loc:
          items:
            anyOf:
              - type: string
              - type: integer
          type: array
          title: Location
        msg:
          type: string
          title: Message
        type:
          type: string
          title: Error Type
      type: object
      required:
        - loc
        - msg
        - type
      title: ValidationError
      x-fern-type-name: ValidationError
    Subject:
      type: object
      properties:
        subjectId:
          type: string
          description: Subject ID
        subjectType:
          description: Subject type
          allOf:
            - $ref: '#/components/schemas/SubjectType'
        subjectSlug:
          type: string
          description: Subject slug
        subjectDisplayName:
          type: string
          description: Subject display name
        subjectPatName:
          type: string
          description: Subject PAT name
        subjectControllerName:
          type: string
        subjectExternalIdentitySlug:
          type: string
          description: External identity slug (external_identity_id:sub:email)
      required:
        - subjectId
        - subjectType
      title: Subject
      x-fern-type-name: Subject
    ModelManifest:
      properties:
        name:
          type: string
          maxLength: 256
          pattern: ^[A-Za-z0-9_\-]+$
          title: Name
          description: >-
            Name of the model (alphanumeric characters, hyphens, and underscores
            only, max 256 characters)
        metadata:
          type: object
          title: Metadata
          description: >-
            Key value metadata. Should be valid JSON. For e.g.
            `{"business-unit": "sales", "quality": "good", "rating": 4.5}`
        ml_repo:
          type: string
          pattern: ^[a-zA-Z][a-zA-Z0-9\-]{1,98}[a-zA-Z0-9]$
          title: Ml Repo
          description: >-
            Name of the ML Repo that this model belongs to (must start and end
            with alphanumeric, 2-100 characters)
        version:
          type: integer
          minimum: 1
          title: Version
          description: Version of the entity
        type:
          type: string
          enum:
            - model-version
          title: Type
          default: model-version
        description:
          type: string
          maxLength: 512
          title: Description
        version_alias:
          type: string
          maxLength: 128
          pattern: ^v[a-zA-Z0-9.-]*([a-zA-Z0-9]+)$
          title: Version Alias
          description: >-
            Version alias is alternate, ideally human readable, version string
            to reference an artifact version. It should start with `v` followed
            by alphanumeric and it can include `.` and `-` in between (e.g.
            `v1.0.0`, `v1-prod`, `v3-dev`, etc)
        source:
          oneOf:
            - $ref: '#/components/schemas/TrueFoundryManagedSource'
            - $ref: '#/components/schemas/ExternalBlobStorageSource'
            - $ref: '#/components/schemas/LocalModelSource'
          title: Source
          discriminator:
            propertyName: type
            mapping:
              truefoundry:
                $ref: '#/components/schemas/TrueFoundryManagedSource'
              external:
                $ref: '#/components/schemas/ExternalBlobStorageSource'
              local:
                $ref: '#/components/schemas/LocalModelSource'
        framework:
          oneOf:
            - $ref: '#/components/schemas/TransformersFramework'
            - $ref: '#/components/schemas/TensorFlowFramework'
            - $ref: '#/components/schemas/SklearnFramework'
            - $ref: '#/components/schemas/PyTorchFramework'
            - $ref: '#/components/schemas/KerasFramework'
            - $ref: '#/components/schemas/XGBoostFramework'
            - $ref: '#/components/schemas/LightGBMFramework'
            - $ref: '#/components/schemas/FastAIFramework'
            - $ref: '#/components/schemas/H2OFramework'
            - $ref: '#/components/schemas/ONNXFramework'
            - $ref: '#/components/schemas/SpaCyFramework'
            - $ref: '#/components/schemas/StatsModelsFramework'
            - $ref: '#/components/schemas/GluonFramework'
            - $ref: '#/components/schemas/PaddleFramework'
          title: Framework
          description: >-
            Framework for the model version like Transformers, PyTorch, Sklearn,
            Xgboost etc with framework specific metadata. This will be used to
            infer model deployment configuration
          discriminator:
            propertyName: type
            mapping:
              transformers:
                $ref: '#/components/schemas/TransformersFramework'
              tensorflow:
                $ref: '#/components/schemas/TensorFlowFramework'
              sklearn:
                $ref: '#/components/schemas/SklearnFramework'
              pytorch:
                $ref: '#/components/schemas/PyTorchFramework'
              keras:
                $ref: '#/components/schemas/KerasFramework'
              xgboost:
                $ref: '#/components/schemas/XGBoostFramework'
              lightgbm:
                $ref: '#/components/schemas/LightGBMFramework'
              fastai:
                $ref: '#/components/schemas/FastAIFramework'
              h2o:
                $ref: '#/components/schemas/H2OFramework'
              onnx:
                $ref: '#/components/schemas/ONNXFramework'
              spacy:
                $ref: '#/components/schemas/SpaCyFramework'
              statsmodels:
                $ref: '#/components/schemas/StatsModelsFramework'
              gluon:
                $ref: '#/components/schemas/GluonFramework'
              paddle:
                $ref: '#/components/schemas/PaddleFramework'
        environment:
          $ref: '#/components/schemas/ModelVersionEnvironment'
        step:
          type: integer
          minimum: 0
          title: Step
          description: >-
            Step/Epoch number in an iterative training loop the model version
            was created. Generally useful when logging a model version from a
            MLRepo Run
          default: 0
        run_id:
          type: string
          title: Run Id
          description: ID of the MLRepo Run that generated the model version
      type: object
      required:
        - name
        - metadata
        - ml_repo
        - source
      title: ModelManifest
      description: Model Version artifact.
      x-fern-type-name: ModelManifest
    Metric:
      properties:
        key:
          type: string
          title: Key
          description: Name of the metric
        value:
          type: number
          title: Value
          description: Value of the metric
        timestamp:
          type: integer
          title: Timestamp
          description: Timestamp when the metric was recorded (epoch milliseconds)
        step:
          type: integer
          title: Step
          description: Training step number when the metric was recorded
          default: 0
      type: object
      required:
        - key
      title: Metric
      x-fern-type-name: Metric
    SubjectType:
      type: string
      enum:
        - user
        - team
        - serviceaccount
        - virtualaccount
        - external-identity
        - role
      title: SubjectType
      x-fern-type-name: SubjectType
    TrueFoundryManagedSource:
      properties:
        type:
          type: string
          enum:
            - truefoundry
          title: Type
          description: Type of the source
        uri:
          type: string
          pattern: ^(s3|gs|wasbs)://.*
          title: Uri
      type: object
      required:
        - type
      title: TrueFoundryManagedSource
      x-fern-type-name: TrueFoundryManagedSource
    ExternalBlobStorageSource:
      properties:
        type:
          type: string
          enum:
            - external
          title: Type
          description: Type of the source
        uri:
          type: string
          pattern: ^(s3|gs|wasbs)://.*
          title: Uri
          description: >-
            URI referencing a path in the blob storage bucket linked to the
            MLRepo
      type: object
      required:
        - type
        - uri
      title: ExternalBlobStorageSource
      x-fern-type-name: ExternalBlobStorageSource
    LocalModelSource:
      properties:
        type:
          type: string
          enum:
            - local
          title: Type
          description: Type of the source
        file_or_folder:
          type: string
          title: File Or Folder
          description: Path to the model file or folder
      type: object
      required:
        - type
        - file_or_folder
      title: LocalModelSource
      x-fern-type-name: LocalModelSource
    TransformersFramework:
      properties:
        type:
          type: string
          enum:
            - transformers
          title: Type
          description: Type of the framework
        library_name:
          allOf:
            - $ref: '#/components/schemas/LibraryName'
          description: Name of the library for the framework
          default: transformers
        pipeline_tag:
          type: string
          title: Pipeline Tag
          description: >-
            The `pipeline()` task this model can be used with e.g.
            `text-generation`. See [huggingface
            docs](https://huggingface.co/docs/transformers/main/en/main_classes/pipelines#transformers.pipeline.task)
            for all possible values
        base_model:
          type: string
          title: Base Model
          description: >-
            Base model Id from HuggingFace Hub. If this is a finetuned model,
            this points to the base model id used for finetuning.
      type: object
      required:
        - type
      title: TransformersFramework
      x-fern-type-name: TransformersFramework
    TensorFlowFramework:
      properties:
        type:
          type: string
          enum:
            - tensorflow
          title: Type
          description: Type of the framework
      type: object
      required:
        - type
      title: TensorFlowFramework
      x-fern-type-name: TensorFlowFramework
    SklearnFramework:
      properties:
        type:
          type: string
          enum:
            - sklearn
          title: Type
          description: Type of the framework
        model_filepath:
          type: string
          title: Model Filepath
          description: Relative path to the model file in the model version contents
        serialization_format:
          $ref: '#/components/schemas/SklearnSerializationFormat'
        model_schema:
          $ref: '#/components/schemas/SklearnModelSchema'
      type: object
      required:
        - type
      title: SklearnFramework
      x-fern-type-name: SklearnFramework
    PyTorchFramework:
      properties:
        type:
          type: string
          enum:
            - pytorch
          title: Type
          description: Type of the framework
      type: object
      required:
        - type
      title: PyTorchFramework
      x-fern-type-name: PyTorchFramework
    KerasFramework:
      properties:
        type:
          type: string
          enum:
            - keras
          title: Type
          description: Type of the framework
      type: object
      required:
        - type
      title: KerasFramework
      x-fern-type-name: KerasFramework
    XGBoostFramework:
      properties:
        type:
          type: string
          enum:
            - xgboost
          title: Type
          description: Type of the framework
        serialization_format:
          $ref: '#/components/schemas/XGBoostSerializationFormat'
        model_filepath:
          type: string
          title: Model Filepath
          description: Relative path to the model file in the model version contents
        model_schema:
          $ref: '#/components/schemas/XGBoostModelSchema'
      type: object
      required:
        - type
      title: XGBoostFramework
      x-fern-type-name: XGBoostFramework
    LightGBMFramework:
      properties:
        type:
          type: string
          enum:
            - lightgbm
          title: Type
          description: Type of the framework
      type: object
      required:
        - type
      title: LightGBMFramework
      x-fern-type-name: LightGBMFramework
    FastAIFramework:
      properties:
        type:
          type: string
          enum:
            - fastai
          title: Type
          description: Type of the framework
      type: object
      required:
        - type
      title: FastAIFramework
      x-fern-type-name: FastAIFramework
    H2OFramework:
      properties:
        type:
          type: string
          enum:
            - h2o
          title: Type
          description: Type of the framework
      type: object
      required:
        - type
      title: H2OFramework
      x-fern-type-name: H2OFramework
    ONNXFramework:
      properties:
        type:
          type: string
          enum:
            - onnx
          title: Type
          description: Type of the framework
      type: object
      required:
        - type
      title: ONNXFramework
      x-fern-type-name: ONNXFramework
    SpaCyFramework:
      properties:
        type:
          type: string
          enum:
            - spacy
          title: Type
          description: Type of the framework
      type: object
      required:
        - type
      title: SpaCyFramework
      x-fern-type-name: SpaCyFramework
    StatsModelsFramework:
      properties:
        type:
          type: string
          enum:
            - statsmodels
          title: Type
          description: Type of the framework
      type: object
      required:
        - type
      title: StatsModelsFramework
      x-fern-type-name: StatsModelsFramework
    GluonFramework:
      properties:
        type:
          type: string
          enum:
            - gluon
          title: Type
          description: Type of the framework
      type: object
      required:
        - type
      title: GluonFramework
      x-fern-type-name: GluonFramework
    PaddleFramework:
      properties:
        type:
          type: string
          enum:
            - paddle
          title: Type
          description: Type of the framework
      type: object
      required:
        - type
      title: PaddleFramework
      x-fern-type-name: PaddleFramework
    ModelVersionEnvironment:
      properties:
        python_version:
          type: string
          pattern: ^\d+(\.\d+){1,2}([\-\.a-z0-9]+)?$
          title: Python Version
          description: Python version to use when running the model
        pip_packages:
          items:
            type: string
          type: array
          title: Pip Packages
          description: pip dependencies needed for running the model
      type: object
      title: ModelVersionEnvironment
      description: >-
        Runtime environment describing python version and dependencies for
        running the model
      x-fern-type-name: ModelVersionEnvironment
    LibraryName:
      type: string
      enum:
        - transformers
        - sentence-transformers
        - diffusers
      title: LibraryName
      description: Name of the library for the framework
      x-fern-type-name: LibraryName
    SklearnSerializationFormat:
      type: string
      enum:
        - cloudpickle
        - joblib
        - pickle
      title: SklearnSerializationFormat
      description: Serialization format used to save the sklearn model
      x-fern-type-name: SklearnSerializationFormat
    SklearnModelSchema:
      properties:
        infer_method_name:
          allOf:
            - $ref: '#/components/schemas/InferMethodName'
          description: Name of the method used for inference
        inputs:
          items:
            type: object
          type: array
          title: Inputs
          description: Schema of the input
        outputs:
          items:
            type: object
          type: array
          title: Outputs
          description: Schema of the output
      type: object
      required:
        - infer_method_name
        - inputs
        - outputs
      title: SklearnModelSchema
      x-fern-type-name: SklearnModelSchema
    XGBoostSerializationFormat:
      type: string
      enum:
        - cloudpickle
        - joblib
        - pickle
        - json
      title: XGBoostSerializationFormat
      description: Serialization format used to save the xgboost model
      x-fern-type-name: XGBoostSerializationFormat
    XGBoostModelSchema:
      properties:
        infer_method_name:
          type: string
          enum:
            - predict
          title: Infer Method Name
          description: Name of the method used for inference
        inputs:
          items:
            type: object
          type: array
          title: Inputs
          description: Schema of the input
        outputs:
          items:
            type: object
          type: array
          title: Outputs
          description: Schema of the output
      type: object
      required:
        - infer_method_name
        - inputs
        - outputs
      title: XGBoostModelSchema
      description: Schema of the XGBoost model
      x-fern-type-name: XGBoostModelSchema
    InferMethodName:
      type: string
      enum:
        - predict
        - predict_proba
      title: InferMethodName
      description: Name of the method used for inference
      x-fern-type-name: InferMethodName
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````