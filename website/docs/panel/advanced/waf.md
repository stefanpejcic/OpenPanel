---
sidebar_position: 9
---

# WAF

The Web Application Firewall (WAF) interface allows OpenPanel users to toggle on/off Coraza WAF for each of their domains individually.

![waf.png](/img/panel/v2/waf.png)

By clicking on a domain name from the list you can disable individual rules for that domain.

![waf.png](/img/panel/v2/waf_rules.png)

Rules can be disabled by ID (SecRuleRemoveByID) 

![waf.png](/img/panel/v2/waf_rulesID.png)

and/or by Tag name (SecRuleRemoveByTag).

![waf.png](/img/panel/v2/waf_rulesTag.png)

The OWASP Core ruleset is enabled by default for newly added domains, WAF log files can also be viewed from this interface by clicking on the View Logs button top right.

![waf.png](/img/panel/v2/waf_logs.png)
