# WEMS Grafana Data Source Plugin

A comprehensive Grafana data source plugin that seamlessly integrates **WAGO Energy Management System (WEMS)** with Grafana dashboards, enabling real-time visualization and analysis of energy data.

## Overview

The WEMS Grafana Plugin bridges the gap between WAGO's industrial IoT energy management platform and Grafana's powerful visualization capabilities. Monitor energy consumption, production efficiency, and system performance metrics across your industrial infrastructure with ease.

## Key Features

### **Data Visualization**
- Support for multiple data types: numeric, boolean, and string values
- Interactive query builder with intuitive cascading selections

### **User-Friendly Interface**
- **Hierarchical Data Selection**: Endpoint → Appliance → Service → Data Point
- **Filtering**: Contextual options based on your selections
- **Model Information**: Device model names displayed for easy identification

### **Advanced Analytics**
- **9 Aggregation Functions**: `mean`, `median`, `min`, `max`, `sum`, `count`, `first`, `last`, `derivative`
- **Gap Filling**: Optional empty value creation for continuous time series

## Requirements

| Component | Version | Notes |
|-----------|---------|-------|
| **Grafana** | ≥ 10.4.0 | Required for plugin compatibility |
| **WEMS API Access** | Latest | Valid client credentials required |
| **Network Access** | HTTPS | To WEMS API endpoints |

## Configuration

### 1. Add Data Source
1. Go to **Configuration** → **Data Sources**
2. Click **Add data source**
3. Select **WEMS** from the list

### 2. Configure Authentication
| Field | Description | Required |
|-------|-------------|----------|
| **Client ID** | Your WEMS application client identifier | ✅ |
| **Client Secret** | Your WEMS application secret key | ✅ |
| **Base URL** | WEMS API endpoint (default: `https://c1.api.wago.com/wems`) | ❌ |

### 3. Test Connection
Click **Save & Test** to verify your configuration. A successful connection will display available endpoints.

## Creating Queries

### Step-by-Step Query Builder

1. **Select Endpoint**: Choose your site or system endpoint
2. **Choose Appliance**: Pick the energy management device
3. **Select Service**: Choose the data category or measurement channel
4. **Pick Data Point**: Select the specific metric to visualize
5. **Configure Aggregation**: Choose how to aggregate data over time intervals

### Query Parameters

| Parameter | Description | Example |
|-----------|-------------|---------|
| **Endpoint ID** | Site or system identifier | `site_001` |
| **Appliance ID** | Device identifier with model info | `inverter_01 (Fronius Gen24)` |
| **Service URI** | Data category or channel | `sgr.ActivePowerAC` |
| **Data Point** | Specific measurement | `ActivePowerACTot` |
| **Aggregate Function** | Time-series aggregation method | `mean` |
| **Create Empty Values** | Fill gaps in time series | `false` |

## Important Notes

### Boolean Data Points
> **Critical**: For boolean data points, always use the `last` aggregate function. Statistical functions like `mean` or `sum` are not meaningful for boolean values and will produce incorrect results.

### Data Type Handling
- **Numeric Values**: All aggregation functions supported
- **Boolean Values**: Use `last`, `first`, or `count` only
- **String Values**: Use `first`, `last`, or `count` only
