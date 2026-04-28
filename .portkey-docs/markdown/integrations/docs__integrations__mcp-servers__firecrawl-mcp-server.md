> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Firecrawl MCP server

> The Firecrawl MCP server enables AI agents to perform web scraping, crawling, search, extraction, and research through MCP. Built for content discovery and information retrieval workflows.

## When you should use this server

* Scrape and extract content from web pages or entire domains.
* Perform batch scraping jobs with built-in rate limiting.
* Search the web and optionally extract structured content from results.
* Conduct deep research using a combination of crawling, search, and LLM-powered summarization.
* Generate standardized **llms.txt** and **llms-full.txt** files to guide LLM interactions with your site.
* Integrate **content discovery, crawling, and extraction** into AI-driven workflows.

## Key features

* Web scraping with custom wait times and content filters
* Batch scraping with parallel execution
* Deep web crawling with configurable depth and limits
* Structured data extraction from web pages
* LLM-powered research and summarization
* Generation of standardized LLM guidance files

## Authentication

* **Method:** API Key (FIRECRAWL\_API\_KEY)
* **Notes:** Store securely in environment variables; available from the Firecrawl dashboard

## Endpoints

* **Remote hosted URL:** `https://mcp.firecrawl.dev/{FIRECRAWL_API_KEY}/v2/sse`
* **Local via npx:** `env FIRECRAWL_API_KEY=fc-YOUR_API_KEY npx -y firecrawl-mcp`
* **Manual installation:** `npm install -g firecrawl-mcp`

## Setup & usage

### Environment variables

* FIRECRAWL\_API\_KEY=fc-YOUR\_API\_KEY  # Required
* FIRECRAWL\_API\_URL=[https://custom-endpoint.example.com](https://custom-endpoint.example.com)  # Optional, for self-hosted deployments
* FIRECRAWL\_RETRY\_MAX\_ATTEMPTS=5  # Optional, default retry configuration
* FIRECRAWL\_RETRY\_INITIAL\_DELAY=1000  # Optional, in milliseconds
* FIRECRAWL\_RETRY\_MAX\_DELAY=30000  # Optional, in milliseconds
* FIRECRAWL\_RETRY\_BACKOFF\_FACTOR=2  # Optional, exponential backoff multiplier
* FIRECRAWL\_CREDIT\_WARNING\_THRESHOLD=10  # Optional, percentage threshold for credit warnings
* FIRECRAWL\_CREDIT\_CRITICAL\_THRESHOLD=5  # Optional, percentage threshold for critical credit alerts

## Tools provided

### firecrawl\_scrape

Scrape content from a single URL with options for format, wait time, and content filtering.

### firecrawl\_batch\_scrape

Scrape multiple URLs in parallel; returns a batch operation ID for status tracking.

### firecrawl\_check\_batch\_status

Check the progress or results of a batch scrape operation.

### firecrawl\_search

Run a web search; optionally scrape and extract structured content from search results.

### firecrawl\_crawl

Launch an asynchronous crawl with configurable depth, limits, and external link rules.

### firecrawl\_extract

Extract structured information from pages using an LLM and a defined schema (supports both cloud and self-hosted LLMs).

### firecrawl\_deep\_research

Perform deep research on a query using crawling, search, and LLM analysis with constraints (depth, time, max URLs).

### firecrawl\_generate\_llmstxt

Generate standardized **llms.txt** and **llms-full.txt** files to define how LLMs should interact with a site.

## Notes

* Supports both **cloud-hosted** and **self-hosted** deployments.
* Includes **robust logging** for operation status, rate limits, and credit usage.
* Automatic retries and backoff are built-in for error handling.
* Returned data respects configured API quotas and credit limits.
