> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Update User Profile Picture

> Update the profile picture URL for the authenticated user



## OpenAPI

````yaml /openapi.json patch /api/svc/v1/users/profile-picture
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
  /api/svc/v1/users/profile-picture:
    patch:
      tags:
        - Users
      summary: Update User Profile Picture
      description: Update the profile picture URL for the authenticated user
      operationId: User.updateUserProfilePicture
      parameters: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UpdateUserProfilePictureRequestDto'
      responses:
        '200':
          description: User profile picture has been successfully updated.
        '400':
          description: Bad Request. Profile picture can only be set for user
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '404':
          description: Not Found. User with the given email not found.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
      security:
        - jwt: []
        - jwt: []
components:
  schemas:
    UpdateUserProfilePictureRequestDto:
      type: object
      properties:
        downloadPath:
          type: string
          description: Path of the profile picture to be saved
      required:
        - downloadPath
      title: UpdateUserProfilePictureRequestDto
      x-fern-type-name: UpdateUserProfilePictureRequestDto
    HttpError:
      type: object
      properties:
        statusCode:
          type: integer
          description: HTTP Status Code
        message:
          type: string
          description: Error Message
        code:
          oneOf:
            - type: integer
            - type: string
          description: Error code
        details:
          type: array
          description: Error details
          items:
            type: object
      required:
        - statusCode
        - message
      title: HttpError
      x-fern-type-name: HttpError
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````