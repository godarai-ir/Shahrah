> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Firebase MCP server

> The Firebase MCP server gives AI agents programmatic access to Firebase projects, including Auth user management, Firestore data access, Storage rules, and GraphQL features via the Firebase CLI.

## When should you use it

* Initialize/configure Firebase features in a repo
* Inspect or tweak Auth users and custom claims
* Read/query Firestore data and collections (with rule checks/validation)
* Retrieve Data Connect schemas and execute GraphQL queries/mutations
* Validate and fetch Storage rules / object URLs
* Send FCM test messages

## Auth & transport

* **Auth:** uses the same credentials as the **Firebase CLI** (logged-in user or Application Default Credentials). You must be signed in with `firebase-tools` before using the server.
* **Transport:** **stdio** (configure in clients like Claude Desktop, Cursor, VS Code Copilot, etc.)

## Tools provided (official list)

### Core / Environment / Project

* **firebase\_get\_project** — Get the active Firebase project
* **firebase\_list\_apps** — List registered apps
* **firebase\_get\_admin\_sdk\_config** — Admin SDK config for the current project
* **firebase\_list\_projects** — List Firebase projects (limited count)
* **firebase\_get\_sdk\_config** — Client SDK config for a platform or app ID
* **firebase\_create\_project** — Create a new project
* **firebase\_create\_app** — Create a Web/iOS/Android app in the project
* **firebase\_create\_android\_sha** — Add an Android SHA cert hash
* **firebase\_consult\_assistant** — Ask an AI assistant about Firebase products
* **firebase\_get\_environment** — Show current env (user, project dir, active project)
* **firebase\_update\_environment** — Update env settings (dir, active project, user)
* **firebase\_init** — Initialize selected features (Firestore, Data Connect, Realtime DB). Re-init may overwrite; deploy with `firebase deploy`

### Firestore

* **firestore\_delete\_document** — Delete document(s) by full path
* **firestore\_get\_documents** — Get document(s) by full path
* **firestore\_list\_collections** — List collections in a database
* **firestore\_query\_collection** — Query documents in a collection with a filter
* **firestore\_get\_rules** — Retrieve active Firestore Rules
* **firestore\_validate\_rules** — Validate Firestore Rules source or file path

### Authentication

* **auth\_get\_user** — Fetch a user by email/phone/UID
* **auth\_disable\_user** — Disable/enable a user by UID
* **auth\_list\_users** — List users (limit)
* **auth\_set\_claim** — Set a custom claim (string or JSON value)
* **auth\_set\_sms\_region\_policy** — Set ALLOW/DENY list for SMS regions

### Data Connect (GraphQL)

* **dataconnect\_list\_services** — List Data Connect services
* **dataconnect\_generate\_schema** — Generate a schema from a description (uses Gemini in Firebase)
* **dataconnect\_generate\_operation** — Generate a query/mutation from schema (uses Gemini)
* **dataconnect\_get\_schema** — Get schema (Cloud SQL sources, GraphQL SDL)
* **dataconnect\_get\_connectors** — Get connectors & predefined queries
* **dataconnect\_execute\_graphql** — Execute arbitrary GraphQL (read/write)
* **dataconnect\_execute\_graphql\_read** — Execute read-only GraphQL
* **dataconnect\_execute\_mutation** — Execute a deployed mutation
* **dataconnect\_execute\_query** — Execute a deployed query

### Storage

* **storage\_get\_rules** — Retrieve Storage Rules
* **storage\_validate\_rules** — Validate Storage Rules source or file path
* **storage\_get\_object\_download\_url** — Get a download URL for an object

### Messaging

* **messaging\_send\_message** — Send an FCM message to a token or topic (one of `registration_token` or `topic`)
