# Postmortem Utility
This section contains post-mortems whose purpose is to document incidents, their impact, root causes, resolutions, and lessons learned. The goal is to improve system reliability by learning from failures and preventing their recurrence. Most importantly, it encourages us to learn from our mistakes rather than relying heavily on memory or AI to solve recurring problems. By documenting incidents, we can identify how frequently they occur, understand their causes, and discover ways to prevent them in the future.

Post-mortems differ from runbooks:
* **Post-mortems** are reports documenting what happened, why it happened, how it was resolved, and what can be done to prevent recurrence.
* **Runbooks** are operational procedures containing instructions to follow when diagnosing, handling, or recovering from an incident.

## How Severity Is Determined
Severity is determined by the incident's impact, scope, and urgency, rather than by the technical complexity of the problem.
* **SEV1 - Critical:** Causes a critical system failure, major service outage, significant data loss, or security breach. Requires immediate attention and resolution.
* **SEV2 - High:** Causes significant disruption to major system functionality, substantial service degradation, or partial outages. Requires prompt attention and resolution.
* **SEV3 - Moderate:** Causes limited disruption, such as minor service failures, bugs, or configuration issues. Requires attention but is not immediately critical.
* **SEV4 - Low:** Causes minimal disruption, such as minor development issues, non-critical bugs, or configuration problems with little or no user impact. Can generally be addressed during normal development.