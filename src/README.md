<!-- This README file is going to be the one displayed on the Grafana.com website for your plugin. Uncomment and replace the content here before publishing.

Remove any remaining comments before publishing as these may be displayed on Grafana.com -->

# WEMS Grafana Plugin

A Grafana data source plugin for visualizing WEMS (WAGO Energy Management System) history data via the WEMS API.

## Features
- Query WEMS history data directly from Grafana dashboards
- Secure authentication using WEMS supertoken (client_id, client_secret)
- Configurable API base URL
- Interactive query editor with cascading dropdowns:
  - Endpoint ID
  - Appliance ID (with model name)
  - Service URI
  - Data Point
  - Aggregate Function (mean, median, min, max, sum, count, first, last, derivative)
- Support for boolean, numeric, and string data points

## Requirements
- Grafana 9.0 or later
- Access to the WAGO WEMS API with valid credentials (client_id, client_secret)

## Getting Started
1. **Install the plugin** in your Grafana instance (see Grafana plugin catalog or your administrator).
2. **Configure the data source**:
   - Enter your WEMS `client_id` and `client_secret`.
   - Optionally set a custom `base_url` for the WEMS API (defaults to `https://c1.api.wago.com/wems`).
3. **Create a new panel** and select the WEMS data source.
4. **Build your query** using the interactive editor:
   - Select an **Endpoint ID** (site or system).
   - Select an **Appliance ID** (device, with model name shown in brackets).
   - Select a **Service URI** (data category or channel).
   - Select a **Data Point** (measurement or status).
   - Choose an **Aggregate Function** (defaults to `mean`).
   - (Optional) Enable **Create Empty Values** to fill missing data points.

> **Warning:**
> If the expected value type for your data point is boolean, you **must** set the aggregate function to `last`. Other aggregate functions (mean, sum, etc.) are not meaningful for boolean values and may produce incorrect results.

## Query Editor Details
- **Dropdowns are dependent:** Each dropdown is enabled and populated only after the previous selection is made.
- **Appliance ID dropdown** displays both the appliance name and its model for easier identification.
- **Aggregate Function** options:
  - `mean`, `median`, `min`, `max`, `sum`, `count`, `first`, `last`, `derivative`
  - Default is `mean` (except for boolean values; see warning above)

## Troubleshooting
- If you see errors loading dropdowns, check your WEMS credentials and network access.
- If no data appears, verify your query selections and time range.
- For boolean data points, always use the `last` aggregate function.

## Support
For questions, issues, or feature requests, please contact your system administrator or the plugin maintainer.
