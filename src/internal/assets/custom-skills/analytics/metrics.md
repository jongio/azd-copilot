---
name: metrics
description: Define and implement custom metrics in Azure
agent: analytics
---

# Custom Metrics

## Purpose

Define and track custom business and technical metrics in Azure Monitor.

## Metric Types

| Type | Use Case | Example |
|------|----------|---------|
| **Counter** | Cumulative total | Total orders placed |
| **Gauge** | Point-in-time value | Queue depth |
| **Histogram** | Distribution | Request duration buckets |
| **Summary** | Percentiles | Response time p50/p95/p99 |

## Application Insights Custom Metrics

### Tracking in Code

```typescript
// Node.js with Application Insights SDK
import { TelemetryClient } from 'applicationinsights';

const client = new TelemetryClient();

// Track metric
client.trackMetric({
  name: 'OrderValue',
  value: order.total,
  properties: {
    customerId: order.customerId,
    region: order.region
  }
});

// Track event with metrics
client.trackEvent({
  name: 'OrderPlaced',
  properties: { orderId: order.id },
  measurements: {
    itemCount: order.items.length,
    orderValue: order.total
  }
});
```

```csharp
// .NET with Application Insights SDK
var telemetry = new TelemetryClient();

// Track metric
telemetry.TrackMetric("OrderValue", order.Total);

// Track with dimensions
var metric = telemetry.GetMetric("OrderValue", "Region");
metric.TrackValue(order.Total, order.Region);
```

### Querying Custom Metrics

```kql
// Custom metrics table
customMetrics
| where timestamp > ago(24h)
| where name == "OrderValue"
| summarize 
    totalOrders = count(),
    totalValue = sum(value),
    avgValue = avg(value)
    by bin(timestamp, 1h)
| render timechart

// Metrics from custom events
customEvents
| where timestamp > ago(24h)
| where name == "OrderPlaced"
| extend orderValue = todouble(customMeasurements.orderValue)
| summarize 
    orders = count(),
    revenue = sum(orderValue)
    by bin(timestamp, 1h)
```

## Azure Monitor Custom Metrics

### Emitting via REST API

```bash
# POST to Azure Monitor ingestion endpoint
curl -X POST "https://{region}.monitoring.azure.com/{resourceId}/metrics" \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "time": "2024-01-15T12:00:00Z",
    "data": {
      "baseData": {
        "metric": "QueueDepth",
        "namespace": "MyApp/Queues",
        "dimNames": ["QueueName"],
        "series": [{
          "dimValues": ["orders"],
          "count": 1,
          "sum": 150
        }]
      }
    }
  }'
```

### Bicep for Metric Alert on Custom Metric

```bicep
resource metricAlert 'Microsoft.Insights/metricAlerts@2018-03-01' = {
  name: 'high-queue-depth'
  location: 'global'
  properties: {
    severity: 2
    enabled: true
    scopes: [appInsights.id]
    evaluationFrequency: 'PT5M'
    windowSize: 'PT15M'
    criteria: {
      'odata.type': 'Microsoft.Azure.Monitor.SingleResourceMultipleMetricCriteria'
      allOf: [{
        name: 'QueueDepthHigh'
        metricNamespace: 'Azure.ApplicationInsights'
        metricName: 'QueueDepth'
        operator: 'GreaterThan'
        threshold: 1000
        timeAggregation: 'Average'
      }]
    }
    actions: [{
      actionGroupId: actionGroup.id
    }]
  }
}
```

## Standard Metrics to Track

### Business Metrics

| Metric | Type | Dimensions |
|--------|------|------------|
| `OrdersPlaced` | Counter | region, plan_type |
| `OrderValue` | Summary | region, plan_type |
| `ActiveUsers` | Gauge | plan_type |
| `FeatureUsage` | Counter | feature_name |

### Technical Metrics

| Metric | Type | Dimensions |
|--------|------|------------|
| `QueueDepth` | Gauge | queue_name |
| `CacheHitRate` | Gauge | cache_name |
| `BackgroundJobDuration` | Histogram | job_type |
| `ExternalAPILatency` | Histogram | api_name |

## Aggregation Intervals

| Interval | Use Case |
|----------|----------|
| 1 minute | Real-time dashboards |
| 5 minutes | Standard monitoring |
| 15 minutes | Trend analysis |
| 1 hour | Daily reports |
| 1 day | Monthly reports |

## Best Practices

1. **Use dimensions wisely**: Max 10 dimensions per metric
2. **Pre-aggregate**: Send aggregated values, not individual events
3. **Namespace logically**: `MyApp/Orders`, `MyApp/Users`
4. **Document metrics**: Maintain a metric catalog
5. **Set retention**: Configure appropriate retention period
6. **Alert on meaningful thresholds**: Avoid alert fatigue

## Metric Catalog Template

```markdown
## Metric: OrdersPlaced

- **Description**: Number of orders successfully placed
- **Type**: Counter
- **Unit**: Count
- **Dimensions**: region, plan_type, payment_method
- **Collection Frequency**: Per event
- **Aggregation**: Sum
- **Alert Thresholds**: 
  - Low: < 10/hour (warning)
  - None in 30 min (critical)
- **Dashboard**: Business Overview
- **Owner**: Product team
```
