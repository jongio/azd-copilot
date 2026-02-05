---
name: dashboards
description: Create Azure Workbooks and dashboards
agent: analytics
---

# Azure Dashboards & Workbooks

## Purpose

Create interactive monitoring dashboards using Azure Workbooks.

## Workbook vs Dashboard

| Feature | Azure Workbook | Azure Dashboard |
|---------|----------------|-----------------|
| Interactivity | High (parameters, drill-down) | Low (static tiles) |
| Customization | Full control | Limited templates |
| Sharing | Saved to resource group | Portal-level |
| Use Case | Deep analysis | Quick overview |

**Recommendation**: Use Workbooks for most monitoring needs.

## Workbook Structure

```json
{
  "version": "Notebook/1.0",
  "items": [
    {
      "type": "parameters",
      "parameters": [
        {
          "name": "TimeRange",
          "type": "timerange",
          "defaultValue": "P1D"
        }
      ]
    },
    {
      "type": "query",
      "queryType": "kusto",
      "query": "requests | summarize count() by bin(timestamp, 5m)"
    }
  ]
}
```

## Standard Dashboard Layout

### Overview Section

```kql
// Health summary tile
let errors = requests | where success == false | count;
let total = requests | count;
print 
    Status = iff(toscalar(errors) * 100.0 / toscalar(total) < 1, "Healthy", "Degraded"),
    ErrorRate = strcat(round(toscalar(errors) * 100.0 / toscalar(total), 2), "%"),
    TotalRequests = toscalar(total)
```

### Key Metrics Row

```kql
// Requests per minute
requests
| where timestamp > ago(1h)
| summarize rpm = count() / 60.0
| project Metric = "Requests/min", Value = round(rpm, 1)

// Average response time
requests
| where timestamp > ago(1h)
| summarize avg_ms = avg(duration)
| project Metric = "Avg Response", Value = strcat(round(avg_ms, 0), " ms")

// Error rate
requests
| where timestamp > ago(1h)
| summarize total = count(), errors = countif(success == false)
| project Metric = "Error Rate", Value = strcat(round(errors * 100.0 / total, 2), "%")
```

### Performance Section

```kql
// Response time trend
requests
| where timestamp > ago(24h)
| summarize 
    p50 = percentile(duration, 50),
    p95 = percentile(duration, 95)
    by bin(timestamp, 15m)
| render timechart with (title="Response Time Percentiles")

// Throughput
requests
| where timestamp > ago(24h)
| summarize count() by bin(timestamp, 5m)
| render timechart with (title="Request Volume")
```

### Error Section

```kql
// Error rate over time
requests
| where timestamp > ago(24h)
| summarize 
    total = count(),
    errors = countif(success == false)
    by bin(timestamp, 15m)
| extend errorRate = errors * 100.0 / total
| project timestamp, errorRate
| render timechart with (title="Error Rate %")

// Top errors table
exceptions
| where timestamp > ago(24h)
| summarize count() by type, outerMessage
| order by count_ desc
| take 10
```

### Dependencies Section

```kql
// Dependency health
dependencies
| where timestamp > ago(1h)
| summarize 
    calls = count(),
    failures = countif(success == false),
    avgDuration = avg(duration)
    by target, type
| extend failureRate = round(failures * 100.0 / calls, 2)
| project target, type, calls, failureRate, avgDuration = round(avgDuration, 0)
| order by failureRate desc
```

## Workbook Parameters

### Time Range Parameter

```json
{
  "type": "parameters",
  "items": [{
    "name": "TimeRange",
    "type": "4",  // timerange
    "label": "Time Range",
    "defaultValue": "PT1H",
    "options": ["PT1H", "PT4H", "P1D", "P7D"]
  }]
}
```

### Environment Parameter

```json
{
  "type": "parameters",
  "items": [{
    "name": "Environment",
    "type": "2",  // dropdown
    "label": "Environment",
    "query": "resources | where type == 'microsoft.app/containerapps' | distinct name",
    "defaultValue": "production"
  }]
}
```

## Bicep Template for Workbook

```bicep
resource workbook 'Microsoft.Insights/workbooks@2022-04-01' = {
  name: guid('app-dashboard', resourceGroup().id)
  location: location
  kind: 'shared'
  properties: {
    displayName: 'Application Dashboard'
    category: 'workbook'
    serializedData: loadTextContent('workbook.json')
    sourceId: appInsights.id
  }
}
```

## Best Practices

1. **Use parameters**: Make time range and filters interactive
2. **Group logically**: Overview → Details → Deep dive
3. **Show context**: Include totals alongside percentages
4. **Color code**: Red for errors, green for healthy
5. **Add thresholds**: Show when metrics exceed limits
6. **Include drill-down**: Link to detailed views

## Export Options

- **Share link**: Anyone with access to the resource group
- **Export to PDF**: For reporting
- **Pin to Dashboard**: For portal landing page
- **API access**: Programmatic workbook updates
