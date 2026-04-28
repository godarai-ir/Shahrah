> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Get filtered spans data with detailed attributes



## OpenAPI

````yaml /openapi.json post /api/svc/v1/spans/query
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
  /api/svc/v1/spans/query:
    post:
      tags:
        - Traces
      summary: Get filtered spans data with detailed attributes
      operationId: Traces.querySpans
      parameters: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/QuerySpansRequest'
      responses:
        '200':
          description: Returns all the spans matching the query.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/QuerySpansResponse'
      security:
        - jwt: []
components:
  schemas:
    QuerySpansRequest:
      type: object
      properties:
        startTime:
          type: string
          description: Start time in ISO 8601 format (e.g., 2025-03-12T00:00:09.872Z)
        endTime:
          type: string
          description: >-
            End time in ISO 8601 format (e.g., 2025-03-12T00:10:00.000Z).
            Defaults to current time if not provided.
        traceIds:
          description: Array of trace IDs to filter by
          example:
            - traceid1
            - traceid2
          type: array
          items:
            type: string
        spanIds:
          description: Array of span IDs to filter by
          example:
            - spanId1
            - spanId2
          type: array
          items:
            type: string
        parentSpanIds:
          description: Array of parent span IDs to filter by
          example:
            - ''
            - parentSpanId1
            - parentSpanId2
          type: array
          items:
            type: string
        createdBySubjectTypes:
          type: array
          description: Array of subject types to filter by
          example:
            - user
            - virtualaccount
          items:
            $ref: '#/components/schemas/SubjectType'
        createdBySubjectSlugs:
          description: Array of subject slugs to filter by
          example:
            - user@example.com
            - service-account
          type: array
          items:
            type: string
        applicationNames:
          description: Array of application names to filter by
          example:
            - tfy-llm-gateway
            - order-service
          type: array
          items:
            type: string
        limit:
          type: integer
          description: >-
            The maximum number of spans to return per page. Defaults to 200 if
            not provided.
          example: 200
        sortDirection:
          description: >-
            Sort direction for results based on time. Defaults to descending
            (latest first)
          example: desc
          allOf:
            - $ref: '#/components/schemas/SortDirection'
        pageToken:
          type: string
          description: >-
            An opaque string that should be passed as-is from previous response
            for fetching the next page. Pass
            `$response.pagination.nextPageToken` from previous response for
            fetching the next page.
        tracingProjectFqn:
          type: string
          description: Tracing project FQN (e.g., truefoundry:tracing-project:tfy-default)
        dataRoutingDestination:
          type: string
          description: >-
            Data Routing Destination. One of tracingProjectFqn or
            dataRoutingDestination is required.
        filters:
          type: array
          items:
            oneOf:
              - $ref: '#/components/schemas/SpanFieldFilter'
              - $ref: '#/components/schemas/SpanAttributeFilter'
              - $ref: '#/components/schemas/GatewayRequestMetadataFilter'
          description: Array of filters
        includeFeedbacks:
          type: boolean
          description: >-
            When true, feedback data is included in the response. When false,
            feedback data is excluded (returns empty array).
          default: false
      required:
        - startTime
      title: QuerySpansRequest
      x-fern-type-name: QuerySpansRequest
    QuerySpansResponse:
      type: object
      properties:
        data:
          description: Array of flat spans
          type: array
          items:
            $ref: '#/components/schemas/TraceSpan'
        pagination:
          description: Pagination information
          allOf:
            - $ref: '#/components/schemas/TokenPagination'
      required:
        - data
        - pagination
      title: QuerySpansResponse
      x-fern-type-name: QuerySpansResponse
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
    SortDirection:
      type: string
      enum:
        - asc
        - desc
      title: SortDirection
      x-fern-type-name: SortDirection
    SpanFieldFilter:
      type: object
      properties:
        spanFieldName:
          type: string
          enum:
            - spanName
            - spanId
            - traceId
            - parentSpanId
            - applicationName
            - createdBySubjectSlug
            - createdBySubjectType
            - durationInNs
            - durationInMs
            - spanStatusCode
            - tag
            - scopeName
            - environment
            - feedbacks
        operator:
          type: string
          enum:
            - EQUAL
            - GREATER_THAN
            - LESS_THAN
            - GREATER_THAN_EQUAL
            - LESS_THAN_EQUAL
            - BETWEEN
            - IN
            - NOT_IN
            - STRING_CONTAINS
            - STRING_STARTS_WITH
            - STRING_ENDS_WITH
            - ARRAY_HAS_ANY
            - ARRAY_HAS_NONE
            - PRESENT
            - IS_NULL
        value:
          anyOf:
            - type: string
            - type: number
            - type: boolean
            - type: array
              items:
                type: string
            - type: array
              items:
                type: number
      required:
        - spanFieldName
        - operator
        - value
      title: SpanFieldFilter
      x-fern-type-name: SpanFieldFilter
    SpanAttributeFilter:
      type: object
      properties:
        spanAttributeKey:
          type: string
        operator:
          type: string
          enum:
            - EQUAL
            - GREATER_THAN
            - LESS_THAN
            - GREATER_THAN_EQUAL
            - LESS_THAN_EQUAL
            - BETWEEN
            - IN
            - NOT_IN
            - STRING_CONTAINS
            - STRING_STARTS_WITH
            - STRING_ENDS_WITH
            - ARRAY_HAS_ANY
            - ARRAY_HAS_NONE
            - PRESENT
            - IS_NULL
        value:
          anyOf:
            - type: array
              items:
                type: string
            - type: array
              items:
                type: number
            - type: array
              items:
                type: boolean
            - type: string
            - type: number
            - type: boolean
      required:
        - spanAttributeKey
        - operator
        - value
      title: SpanAttributeFilter
      x-fern-type-name: SpanAttributeFilter
    GatewayRequestMetadataFilter:
      type: object
      properties:
        gatewayRequestMetadataKey:
          type: string
        operator:
          type: string
          enum:
            - EQUAL
            - GREATER_THAN
            - LESS_THAN
            - GREATER_THAN_EQUAL
            - LESS_THAN_EQUAL
            - BETWEEN
            - IN
            - NOT_IN
            - STRING_CONTAINS
            - STRING_STARTS_WITH
            - STRING_ENDS_WITH
            - ARRAY_HAS_ANY
            - ARRAY_HAS_NONE
            - PRESENT
            - IS_NULL
        value:
          anyOf:
            - type: string
            - type: array
              items:
                type: string
      required:
        - gatewayRequestMetadataKey
        - operator
        - value
      title: GatewayRequestMetadataFilter
      x-fern-type-name: GatewayRequestMetadataFilter
    TraceSpan:
      type: object
      properties:
        spanId:
          type: string
          description: Unique identifier for the span within the trace.
        traceId:
          type: string
          description: Unique identifier for the trace that contains this span.
        parentSpanId:
          type: string
          description: Identifier of the parent span in the trace hierarchy.
        serviceName:
          type: string
          description: Name of the service that generated this span.
        spanName:
          type: string
          description: Name of the span
        spanKind:
          type: string
          description: Type of span (e.g., CLIENT, SERVER, INTERNAL, PRODUCER, CONSUMER).
        scopeName:
          type: string
          description: Name of the instrumentation scope that created this span.
        scopeVersion:
          type: string
          description: Version of the instrumentation scope that created this span.
        timestamp:
          type: string
          description: Timestamp in ISO 8601 format (e.g., 2025-03-12T00:00:09.872Z).
        durationNs:
          type: number
          description: Duration of the span in nanoseconds.
        statusCode:
          type: string
          description: Status code of the span (e.g., OK, ERROR, UNSET).
        statusMessage:
          type: string
          description: Human-readable status message describing the span result.
        spanAttributes:
          type: object
          additionalProperties: true
          description: Key-value pairs containing additional metadata about the span.
        events:
          additionalProperties: true
          description: Array of events that occurred during the span execution.
          type: array
          items:
            type: object
        createdBySubject:
          description: Subject (user or virtualaccount) that created this span.
          allOf:
            - $ref: '#/components/schemas/Subject'
        feedbacks:
          description: Gateway feedbacks associated with this span (if any).
          type: array
          items:
            type: array
      required:
        - spanId
        - traceId
        - parentSpanId
        - serviceName
        - spanName
        - spanKind
        - scopeName
        - scopeVersion
        - timestamp
        - durationNs
        - statusCode
        - statusMessage
        - spanAttributes
        - events
        - createdBySubject
      title: TraceSpan
      x-fern-type-name: TraceSpan
    TokenPagination:
      type: object
      properties:
        limit:
          type: integer
          description: Number of items per page
          example: 10
        nextPageToken:
          type: string
          description: Base64 encoded token for the next page
        previousPageToken:
          type: string
          description: Base64 encoded token for the previous page
      title: TokenPagination
      x-fern-type-name: TokenPagination
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
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````