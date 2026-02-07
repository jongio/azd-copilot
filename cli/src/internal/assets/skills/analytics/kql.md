---
name: kql
description: Generate KQL queries for Azure Log Analytics and Application Insights
agent: analytics
---

# KQL (Kusto Query Language)

## Purpose

Write effective KQL queries for Azure Log Analytics and Application Insights.

## Common Query Patterns

### Application Insights Tables

| Table | Contains |
|-------|----------|
| `requests` | HTTP requests to your app |
| `dependencies` | Outbound calls (DB, HTTP, etc.) |
| `exceptions` | Unhandled exceptions |
| `traces` | Custom trace logs |
| `customEvents` | Custom events |
| `customMetrics` | Custom metrics |
| `performanceCounters` | System metrics |
| `availabilityResults` | Availability test results |

### Basic Queries

```kql
// Recent failed requests
requests
| where timestamp > ago(1h)
| where success == false
| project timestamp, name, resultCode, duration, url
| order by timestamp desc
| take 100

// Slow requests (> 1 second)
requests
| where timestamp > ago(24h)
| where duration > 1000
| summarize count() by name
| order by count_ desc

// Exception breakdown
exceptions
| where timestamp > ago(24h)
| summarize count() by type, outerMessage
| order by count_ desc
| take 20
```

### Performance Analysis

```kql
// Request duration percentiles
requests
| where timestamp > ago(1h)
| summarize 
    p50 = percentile(duration, 50),
    p90 = percentile(duration, 90),
    p95 = percentile(duration, 95),
    p99 = percentile(duration, 99)
    by name
| order by p95 desc

// Dependency performance
dependencies
| where timestamp > ago(1h)
| summarize 
    avg(duration), 
    count(),
    failureCount = countif(success == false)
    by target, type
| order by avg_duration desc

// End-to-end transaction trace
let operationId = "abc123";
union requests, dependencies, exceptions, traces
| where operation_Id == operationId
| project timestamp, itemType, name, duration, success, message
| order by timestamp asc
```

### Error Analysis

```kql
// Error rate over time
requests
| where timestamp > ago(24h)
| summarize 
    total = count(),
    errors = countif(success == false)
    by bin(timestamp, 1h)
| extend errorRate = errors * 100.0 / total
| project timestamp, errorRate
| render timechart

// Top errors with details
exceptions
| where timestamp > ago(24h)
| summarize 
    count(),
    lastSeen = max(timestamp),
    example = any(outerMessage)
    by type, problemId
| order by count_ desc
| take 10

// Failed dependencies
dependencies
| where timestamp > ago(1h)
| where success == false
| summarize count() by target, type, resultCode
| order by count_ desc
```

### Azure Resource Logs

```kql
// Container Apps logs
ContainerAppConsoleLogs
| where TimeGenerated > ago(1h)
| where Log_s contains "error" or Log_s contains "Error"
| project TimeGenerated, ContainerName_s, Log_s
| order by TimeGenerated desc

// PostgreSQL slow queries
AzureDiagnostics
| where ResourceType == "FLEXIBLESERVERS"
| where Category == "PostgreSQLLogs"
| where Message contains "duration:"
| parse Message with * "duration: " duration:real " ms" *
| where duration > 1000
| project TimeGenerated, Message, duration
| order by duration desc
```

### Business Metrics

```kql
// Custom events aggregation
customEvents
| where timestamp > ago(24h)
| where name == "OrderPlaced"
| extend orderValue = todouble(customDimensions.orderValue)
| summarize 
    orderCount = count(),
    totalValue = sum(orderValue),
    avgValue = avg(orderValue)
    by bin(timestamp, 1h)
| render timechart

// User sessions
requests
| where timestamp > ago(24h)
| summarize 
    requestCount = count(),
    uniqueUsers = dcount(user_Id)
    by bin(timestamp, 1h)
| render timechart
```

## Query Optimization Tips

1. **Filter early**: Put `where` clauses before `summarize`
2. **Use time filters**: Always filter by `timestamp` first
3. **Limit results**: Use `take` or `limit` for exploration
4. **Avoid `*`**: Project only needed columns
5. **Use hints**: `hint.shufflekey` for large joins

## Output Formats

```kql
// Table (default)
requests | take 10

// Time chart
requests | summarize count() by bin(timestamp, 5m) | render timechart

// Bar chart
requests | summarize count() by name | render barchart

// Pie chart
requests | summarize count() by resultCode | render piechart
```

## Integration with Workbooks

KQL queries can be embedded in Azure Workbooks for interactive dashboards. See [dashboards.md](dashboards.md) for workbook creation.
