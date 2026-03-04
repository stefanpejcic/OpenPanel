---
sidebar_position: 9
---

# WAF

A Web Application Firewall (WAF) felület lehetővé teszi, hogy az OpenPanel felhasználók külön-külön be- és kikapcsolhassák a Coraza WAF-ot minden tartományukban.

![waf.png](/img/panel/v2/waf.png)

Ha a listában egy domain névre kattint, letilthatja az adott tartomány egyedi szabályait.

![waf.png](/img/panel/v2/waf_rules.png)

A szabályok letilthatók azonosítóval (SecRuleRemoveByID)

![waf.png](/img/panel/v2/waf_rulesID.png)

és/vagy címkenév alapján (SecRuleRemoveByTag).

![waf.png](/img/panel/v2/waf_rulesTag.png)

Az OWASP Core szabálykészlet alapértelmezés szerint engedélyezve van az újonnan hozzáadott tartományokhoz, a WAF naplófájlok ezen a felületen is megtekinthetők a jobb felső naplók megtekintése gombra kattintva.

![waf.png](/img/panel/v2/waf_logs.png)
