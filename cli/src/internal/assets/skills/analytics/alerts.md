---
name: alerts
description: Configure Azure Monitor alerts and action groups
agent: analytics
---

# Azure Monitor Alerts

## Purpose

Configure effective alerting for Azure applications.

## Alert Types

| Type | Use Case | Data Source |
|------|----------|-------------|
| **Metric Alerts** | Threshold on metrics | Azure Monitor Metrics |
| **Log Alerts** | KQL query results | Log Analytics |
| **Activity Log Alerts** | Azure operations | Activity Log |
| **Smart Detection** | Anomaly detection | Application Insights |

## Alert Severity Levels

| Severity | Level | Response Time | Example |
|----------|-------|---------------|---------|
| Sev 0 - Critical | 0 | Immediate | Service down |
| Sev 1 - Error | 1 | 15 minutes | High error rate |
| Sev 2 - Warning | 2 | 1 hour | Degraded performance |
| Sev 3 - Informational | 3 | Next business day | Capacity warning |
| Sev 4 - Verbose | 4 | No response | Audit/logging |

## Standard Alert Rules

### Availability Alerts

```bicep
// Service unavailable
resource availabilityAlert 'Microsoft.Insights/metricAlerts@2018-03-01' = {
  name: 'service-unavailable'
  location: 'global'
  properties: {
    severity: 0
    enabled: true
    scopes: [appInsights.id]
    evaluationFrequency: 'PT1M'
    windowSize: 'PT5M'
    criteria: {
      'odata.type': 'Microsoft.Azure.Monitor.SingleResourceMultipleMetricCriteria'
      allOf: [{
        name: 'NoRequests'
        metricNamespace: 'microsoft.insights/components'
        metricName: 'requests/count'
        operator: 'LessThanOrEqual'
        threshold: 0
        timeAggregation: 'Total'
      }]
    }
    actions: [{ actionGroupId: criticalActionGroup.id }]
  }
}
```

### Error Rate Alerts

```bicep
// High error rate (> 5%)
resource errorRateAlert 'Microsoft.Insights/scheduledQueryRules@2022-06-15' = {
  name: 'high-error-rate'
  location: location
  properties: {
    severity: 1
    enabled: true
    scopes: [logAnalyticsWorkspace.id]
    evaluationFrequency: 'PT5M'
    windowSize: 'PT15M'
    criteria: {
      allOf: [{
        query: '''
          requests
          | where timestamp > ago(15m)
          | summarize total = count(), errors = countif(success == false)
          | extend errorRate = errors * 100.0 / total
          | where errorRate > 5
        '''
        timeAggregation: 'Count'
        operator: 'GreaterThan'
        threshold: 0
      }]
    }
    actions: {
      actionGroups: [errorActionGroup.id]
    }
  }
}
```

### Performance Alerts

```bicep
// Slow response time (p95 > 2s)
resource slowResponseAlert 'Microsoft.Insights/scheduledQueryRules@2022-06-15' = {
  name: 'slow-response-time'
  location: location
  properties: {
    severity: 2
    enabled: true
    scopes: [logAnalyticsWorkspace.id]
    evaluationFrequency: 'PT5M'
    windowSize: 'PT15M'
    criteria: {
      allOf: [{
        query: '''
          requests
          | where timestamp > ago(15m)
          | summarize p95 = percentile(duration, 95)
          | where p95 > 2000
        '''
        timeAggregation: 'Count'
        operator: 'GreaterThan'
        threshold: 0
      }]
    }
    actions: {
      actionGroups: [warningActionGroup.id]
    }
  }
}
```

### Resource Alerts

```bicep
// High CPU usage
resource highCpuAlert 'Microsoft.Insights/metricAlerts@2018-03-01' = {
  name: 'high-cpu-usage'
  location: 'global'
  properties: {
    severity: 2
    enabled: true
    scopes: [containerApp.id]
    evaluationFrequency: 'PT5M'
    windowSize: 'PT15M'
    criteria: {
      'odata.type': 'Microsoft.Azure.Monitor.SingleResourceMultipleMetricCriteria'
      allOf: [{
        name: 'HighCPU'
        metricNamespace: 'microsoft.app/containerapps'
        metricName: 'CpuPercentage'
        operator: 'GreaterThan'
        threshold: 80
        timeAggregation: 'Average'
      }]
    }
    actions: [{ actionGroupId: warningActionGroup.id }]
  }
}
```

## Action Groups

### Bicep Template

```bicep
resource criticalActionGroup 'Microsoft.Insights/actionGroups@2023-01-01' = {
  name: 'critical-alerts'
  location: 'global'
  properties: {
    groupShortName: 'Critical'
    enabled: true
    emailReceivers: [{
      name: 'oncall-email'
      emailAddress: 'oncall@company.com'
      useCommonAlertSchema: true
    }]
    smsReceivers: [{
      name: 'oncall-sms'
      countryCode: '1'
      phoneNumber: '5551234567'
    }]
    webhookReceivers: [{
      name: 'pagerduty'
      serviceUri: 'https://events.pagerduty.com/integration/{key}/enqueue'
      useCommonAlertSchema: true
    }]
  }
}

resource warningActionGroup 'Microsoft.Insights/actionGroups@2023-01-01' = {
  name: 'warning-alerts'
  location: 'global'
  properties: {
    groupShortName: 'Warning'
    enabled: true
    emailReceivers: [{
      name: 'team-email'
      emailAddress: 'team@company.com'
      useCommonAlertSchema: true
    }]
  }
}
```

## Alert Best Practices

### 1. Avoid Alert Fatigue
- Only alert on actionable conditions
- Use appropriate thresholds (not too sensitive)
- Suppress during maintenance windows

### 2. Provide Context
- Include relevant data in alert body
- Link to runbooks
- Include recent changes info

### 3. Escalation Path
```
Sev 0: Immediate → PagerDuty → Phone call
Sev 1: 5 min delay → Email + Slack
Sev 2: 15 min delay → Email only
Sev 3: Daily digest → Team channel
```

### 4. Test Alerts
- Verify alerts fire correctly
- Test action group delivery
- Validate runbook links

## Alert Runbook Template

```markdown
## Alert: High Error Rate

### Summary
Error rate exceeded 5% over 15 minutes.

### Investigation Steps
1. Check Application Insights for error details
2. Review recent deployments
3. Check dependent services (DB, cache, external APIs)
4. Review Container Apps logs

### Quick Links
- [Application Map](https://portal.azure.com/...)
- [Live Metrics](https://portal.azure.com/...)
- [Recent Deployments](https://github.com/.../actions)

### Escalation
If not resolved in 30 minutes, escalate to on-call architect.

### Resolution
1. If deployment-related: Rollback
2. If dependency-related: Failover or circuit break
3. If traffic-related: Scale up
```

## Suppression Rules

```bicep
// Suppress alerts during maintenance
resource suppressionRule 'Microsoft.AlertsManagement/actionRules@2021-08-08' = {
  name: 'maintenance-window'
  location: 'global'
  properties: {
    scopes: [resourceGroup().id]
    conditions: [{
      field: 'Severity'
      operator: 'Equals'
      values: ['Sev2', 'Sev3', 'Sev4']
    }]
    actions: [{
      actionType: 'RemoveAllActionGroups'
    }]
    schedule: {
      effectiveFrom: '2024-01-20T02:00:00Z'
      effectiveUntil: '2024-01-20T04:00:00Z'
      recurrence: {
        recurrenceType: 'Weekly'
        daysOfWeek: ['Saturday']
      }
    }
  }
}
```
