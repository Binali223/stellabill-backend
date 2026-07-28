# SLO Availability Runbook

## Overview
This runbook addresses alerts triggered by the `SLOBurnRateHigh_API_Availability` alerting rule.
The API Availability SLO specifies that 99.9% of API requests must be successful.

## Symptoms
- A high burn rate is detected over the short (fast) window (e.g., 5m or 1h) and long (slow) window (e.g., 30m or 6h).
- Customers may be experiencing widespread `5xx` errors.

## Mitigation / Troubleshooting
1. **Check Dashboards**: Review the primary API metrics dashboard to confirm the error rate spike.
2. **Review Logs**: Search for `status >= 500` in the centralized logging system. Look for specific endpoints or services throwing errors.
3. **Database Health**: Check if the primary database is experiencing high latency or connection pool exhaustion.
4. **Recent Deployments**: Investigate if a recent deployment or configuration change correlates with the drop in availability. Roll back if necessary.
5. **Dependencies**: Determine if upstream or downstream dependencies are failing.
