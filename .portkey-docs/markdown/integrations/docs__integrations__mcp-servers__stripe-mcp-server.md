> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Stripe MCP server

> The Stripe MCP server exposes Stripe's payments and billing operations to AI agents through the Model Context Protocol, providing secure access to customers, products, pricing, invoices, and more.

## When should you use this server

Use the Stripe MCP server to:

* Retrieve account and balance information
* Manage customers, products, and prices
* Create and list invoices, payment links, and payment intents
* Handle subscriptions, disputes, and refunds
* Search Stripe resources and documentation

## Authentication

* **Method:** OAuth2 via Stripe Connect
* **Modes:** Test and Live (most teams start in test mode)
* **Notes:** Use least-privilege scopes; actions like refunds and subscription changes may require elevated scopes

## Endpoint

`https://api.stripe.com`

## Tools provided

### Account

**get\_stripe\_account\_info — Retrieve account**\
Returns high-level information about the connected Stripe account.

### Balance

**retrieve\_balance — Retrieve balance**\
Shows the account's current and pending balances by currency.

### Coupon

**create\_coupon — Create coupon**\
Creates a coupon (percentage/fixed amount, duration, redemption limits).

**list\_coupons — List coupons**\
Lists existing coupons with metadata.

### Customer

**create\_customer — Create customer**\
Creates a customer record (email/name/metadata).

**list\_customers — List customers**\
Lists customers with optional filters.

### Dispute

**list\_disputes — List disputes**\
Retrieves disputes with status filters.

**update\_dispute — Update dispute**\
Submits evidence or updates dispute metadata.

### Invoice

**create\_invoice — Create invoice**\
Creates an invoice for a customer.

**create\_invoice\_item — Create invoice item**\
Adds a line item to an upcoming invoice.

**finalise\_invoice — Finalise invoice**\
Finalizes a draft invoice.

**list\_invoices — List invoices**\
Lists invoices across the account or for a customer.

### Payment Link

**create\_payment\_link — Create payment link**\
Generates a hosted checkout link for one-time or recurring items.

### Payment Intent

**list\_payment\_intents — List payment intents**\
Retrieves payment intents with status filters.

### Price

**create\_price — Create price**\
Creates a price for a product.

**list\_prices — List prices**\
Lists prices, optionally by product or active status.

### Product

**create\_product — Create product**\
Creates a product (name, description, tax code).

**list\_products — List products**\
Lists products, with filters for active/archive.

### Refund

**create\_refund — Create refund**\
Issues a refund for a charge or payment intent.

### Subscription

**cancel\_subscription — Cancel subscription**\
Cancels an existing subscription.

**list\_subscriptions — List subscriptions**\
Lists subscriptions by customer or status.

**update\_subscription — Update subscription**\
Changes a subscription (price, proration, trial).

### Others

**search\_stripe\_resources — Search Stripe resources**\
Searches across objects (customers, invoices, products, etc.).

**fetch\_stripe\_resources — Fetch Stripe object**\
Retrieves a specific object by type and ID.

**search\_stripe\_documentation — Search Stripe knowledge**\
Searches Stripe's docs for guides, API references, and best practices.

## Rate limits

* Enforced by Stripe per API family
* High-volume operations should be paginated and time-bounded

## Notes

* Prefer **test mode** for development and demos
* Use **OAuth scopes** to control access levels
* Sensitive actions (refunds, subscription changes) should be logged and reviewed in the Stripe Dashboard use this server

Use the Stripe MCP server to:

* Retrieve account and balance information
* Manage customers, products, and prices
* Create and list invoices, payment links, and payment intents
* Handle subscriptions, disputes, and refunds
* Search Stripe resources and documentation

## Authentication

* **Method:** OAuth2 via Stripe Connect
* **Modes:** Test and Live (most teams start in test mode)
* **Notes:** Use least-privilege scopes; actions like refunds and subscription changes may require elevated scopes

## Endpoint

`https://api.stripe.com`

## Tools provided

### Account

**get\_stripe\_account\_info — Retrieve account**\
Returns high-level information about the connected Stripe account.

### Balance

**retrieve\_balance — Retrieve balance**\
Shows the account's current and pending balances by currency.

### Coupon

**create\_coupon — Create coupon**\
Creates a coupon (percentage/fixed amount, duration, redemption limits).

**list\_coupons — List coupons**\
Lists existing coupons with metadata.

### Customer

**create\_customer — Create customer**\
Creates a customer record (email/name/metadata).

**list\_customers — List customers**\
Lists customers with optional filters.

### Dispute

**list\_disputes — List disputes**\
Retrieves disputes with status filters.

**update\_dispute — Update dispute**\
Submits evidence or updates dispute metadata.

### Invoice

**create\_invoice — Create invoice**\
Creates an invoice for a customer.

**create\_invoice\_item — Create invoice item**\
Adds a line item to an upcoming invoice.

**finalise\_invoice — Finalise invoice**\
Finalizes a draft invoice.

**list\_invoices — List invoices**\
Lists invoices across the account or for a customer.

### Payment Link

**create\_payment\_link — Create payment link**\
Generates a hosted checkout link for one-time or recurring items.

### Payment Intent

**list\_payment\_intents — List payment intents**\
Retrieves payment intents with status filters.

### Price

**create\_price — Create price**\
Creates a price for a product.

**list\_prices — List prices**\
Lists prices, optionally by product or active status.

### Product

**create\_product — Create product**\
Creates a product (name, description, tax code).

**list\_products — List products**\
Lists products, with filters for active/archive.

### Refund

**create\_refund — Create refund**\
Issues a refund for a charge or payment intent.

### Subscription

**cancel\_subscription — Cancel subscription**\
Cancels an existing subscription.

**list\_subscriptions — List subscriptions**\
Lists subscriptions by customer or status.

**update\_subscription — Update subscription**\
Changes a subscription (price, proration, trial).

### Others

**search\_stripe\_resources — Search Stripe resources**\
Searches across objects (customers, invoices, products, etc.).

**fetch\_stripe\_resources — Fetch Stripe object**\
Retrieves a specific object by type and ID.

**search\_stripe\_documentation — Search Stripe knowledge**\
Searches Stripe's docs for guides, API references, and best practices.
