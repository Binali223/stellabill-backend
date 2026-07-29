# SLO Latency Runbook

## Overview
This runbook addresses alerts triggered by the `SLOBurnRateHigh_API_Latency` alerting rule.
The API Latency SLO specifies that 99.0% of API requests must complete in under 200ms.

## Symptoms
- A high burn rate is detected over the short (fast) window and long (slow) window.
- Customers may experience degraded performance or timeouts.

## Mitigation / Troubleshooting
1. **Check Dashboards**: Review the latency distribution (histograms) in the primary API monitoring dashboard.
2. **Identify Slow Endpoints**: Check if the latency degradation is uniform across the API or isolated to specific endpoints.
3. **Database Performance**: Look for slow queries, locking issues, or increased query volume on the database.
4. **Third-Party Services**: Determine if downstream APIs or integrations are taking longer than usual to respond.
5. **Resource Constraints**: Verify if there is CPU throttling, memory starvation, or network congestion on the API servers.
