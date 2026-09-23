---
sidebar_position: 9
---

# WAF

The Web Application Firewall (WAF) interface allows OpenPanel users to toggle Coraza WAF on or off for each domain individually.

![WAF page listing domains with a toggle to enable the firewall and Manage Rules and View Logs buttons](/img/openpanel-screenshots/advanced/waf-list.png#gh-light-mode-only)
![WAF page listing domains with a toggle to enable the firewall and Manage Rules and View Logs buttons](/img/openpanel-screenshots/advanced/waf-list_dark.png#gh-dark-mode-only)

## Manage domain

Clicking the **Manage Rules** button next to a domain opens the rule management page where you can disable individual rules for that domain.

![WAF settings for a domain with the status toggle and fields for disabled rule IDs and tags](/img/openpanel-screenshots/advanced/waf_domain-page.png#gh-light-mode-only)
![WAF settings for a domain with the status toggle and fields for disabled rule IDs and tags](/img/openpanel-screenshots/advanced/waf_domain-page_dark.png#gh-dark-mode-only)

Rules can be disabled by ID, in the **Disabled IDs** field (`SecRuleRemoveById`):

![Disabled IDs field of the WAF settings with rule IDs 942100 and 920350](/img/openpanel-screenshots/advanced/waf_domain-ids.png#gh-light-mode-only)
![Disabled IDs field of the WAF settings with rule IDs 942100 and 920350](/img/openpanel-screenshots/advanced/waf_domain-ids_dark.png#gh-dark-mode-only)

Or by tag name, in the **Disabled Tags** field (`SecRuleRemoveByTag`):

![Disabled Tags field of the WAF settings with the attack-sqli tag](/img/openpanel-screenshots/advanced/waf_domain-tags.png#gh-light-mode-only)
![Disabled Tags field of the WAF settings with the attack-sqli tag](/img/openpanel-screenshots/advanced/waf_domain-tags_dark.png#gh-dark-mode-only)

The OWASP Core Rule Set is enabled by default for all newly added domains.


## WAF Logs

WAF logs can be viewed by clicking the **View Logs** button for domain.

![WAF Logs page with the dropdown to choose the domain whose firewall log to view](/img/openpanel-screenshots/advanced/waf_logs-page.png#gh-light-mode-only)
![WAF Logs page with the dropdown to choose the domain whose firewall log to view](/img/openpanel-screenshots/advanced/waf_logs-page_dark.png#gh-dark-mode-only)

### Reading the Logs

Coraza logs all checked requests, not just blocked ones. The key columns are:

**`is_interrupted`**: whether the request was blocked. `true` means blocked.
**`messages.rule_ids`**: the IDs of rules that fired, for example: 
  ```
  932370, 941100, 941160, 949110, 980170
  ```
  These are the IDs you can paste into the *Disabled IDs* field for that domain.
**`messages.tags`** — the tags associated with triggered rules, for example:
  ```
  application-multi, language-shell, platform-windows, attack-rce, paranoia-level/1, OWASP_CRS, OWASP_CRS/ATTACK-RCE, capec/1000/152/248/88
  ```
  When disabling by tag, use the attack category tags such as `attack-rce` or `attack-xss`. Avoid broad tags like `OWASP_CRS` or `paranoia-level/1` as they would disable large portions of the ruleset.
  Disabling rules by tag removes them globally for the entire domain. For false positives on specific pages (e.g. `/wp-admin`), disabling by ID is the safer approach as it can be scoped more precisely.

**`messages.msg`** — the human-readable reason the rule fired, for example:
  ```
  Remote Command Execution: Windows Command Injection
  ```



