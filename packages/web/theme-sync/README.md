# Shared theme synchronization

This package defines the browser theme contract shared by the public website,
tenant admin, model catalog, documentation, and blog.

- Preference values: `light`, `dark`, or `system`.
- Canonical cookie: `godarai-theme`, scoped to `.godarai.ir` in production.
- Per-origin mirror: `localStorage["godarai-theme"]`.
- Precedence: shared cookie, shared local storage, app-specific legacy storage,
  then `system`.
- The bootstrap script applies the resolved theme before hydration and migrates
  legacy values without deleting the old app settings.
- Storage/custom events update same-origin tabs. Focus, visibility, and
  `pageshow` checks pick up cookie changes made on another GodarAI subdomain.

The cookie contains only a display preference. It is intentionally readable by
client-side code and uses `Path=/`, `SameSite=Lax`, a one-year max age, and
`Secure` on HTTPS.
