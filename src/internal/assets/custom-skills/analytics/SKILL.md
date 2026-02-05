---
name: analytics
description: Analytics and observability skills for Azure monitoring
---

# Analytics Skills

Skills for implementing observability, dashboards, and analytics on Azure.

## Available Skills

| Skill | Description | When to Use |
|-------|-------------|-------------|
| [kql](kql.md) | Generate KQL queries for Log Analytics | Querying logs and metrics |
| [dashboards](dashboards.md) | Create Azure Workbooks | Building monitoring dashboards |
| [metrics](metrics.md) | Define custom metrics | Tracking business KPIs |
| [alerts](alerts.md) | Configure Azure Monitor alerts | Setting up alerting |

## Azure Specialization

The analytics role is expert in Azure Monitor ecosystem:

- **Log Analytics**: KQL queries, workspace management
- **Application Insights**: APM, distributed tracing, availability tests
- **Azure Workbooks**: Interactive dashboards
- **Azure Monitor Alerts**: Action groups, smart detection
- **Azure Data Explorer**: Large-scale analytics

## Quick Reference

### KQL Basics

```kql
// Filter and project
requests
| where timestamp > ago(1h)
| where success == false
| project timestamp, name, resultCode, duration
| order by timestamp desc

// Aggregation
requests
| summarize count(), avg(duration) by bin(timestamp, 5m)
| render timechart
```

### Dashboard Structure

```
Dashboard
├── Overview Section
│   ├── Key metrics summary
│   └── Health status
├── Performance Section
│   ├── Response times
│   └── Throughput
├── Errors Section
│   ├── Error rate
│   └── Top errors
└── Resources Section
    ├── CPU/Memory
    └── Scaling events
```
