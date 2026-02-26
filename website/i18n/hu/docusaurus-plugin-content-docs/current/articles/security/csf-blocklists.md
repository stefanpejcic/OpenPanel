# CSF tiltólista

A CSF/LFD támogatja az IP-k és CIDR-ek tiltólistáinak letöltését és alkalmazását nyilvános forrásokból.

Az OpenPanel alapértelmezés szerint **nem** engedélyez semmilyen IP-blokkolót a telepítéskor.

## az OpenAdminból

Ha engedélyezni szeretne egy tiltólistát az OpenAdmin felületről, lépjen a **Biztonság > Tűzfal** elemre, majd görgessen le, és kattintson az „LFD blokkollisták” elemre:

[![2025-07-22-17-06.png](https://i.postimg.cc/LsdV3g1b/2025-07-22-17-06.png)](https://postimg.cc/CRNDF1JG)

**Vonja ki a megjegyzést** a kívánt tiltólistával kezdődő sorból, távolítsa el előtte a „#” jelet, majd kattintson a „Módosítás” gombra:

[![2025-07-22-17-06-1.png](https://i.postimg.cc/jjz4gq6d/2025-07-22-17-06-1.png)](https://postimg.cc/sBgW1r0t)

Végül kattintson a „Csf+lfd újraindítása” gombra:

[![2025-07-22-17-07.png](https://i.postimg.cc/KYrt7qF1/2025-07-22-17-07.png)](https://postimg.cc/nsrsp1Xx)


## a terminálról

Egy adott tiltólista engedélyezése:

1. Nyissa meg a `/etc/csf/csf.blocklists` fájlt
2. **Távolítsa el a megjegyzést** a kívánt tiltólistával kezdődő sort.
3. Mentse el a fájlt.
4. **Indítsa újra a CSF-et**, majd **indítsa újra az LFD-t**: `csf -ra && service lfd restart`





## Formátum

| Paraméter | Leírás |
|-----------|-------------|
| **NÉV** | A lista neve csupa nagybetűs alfabetikus karakterrel, szóközök nélkül és legfeljebb 25 karakterből állhat. Ez lesz az iptables lánc neve. |
| **intervallum** | Frissítési időköz (másodpercben) a lista letöltéséhez. Legalább „3600” (1 óra), de általában elegendő „86400” (1 nap). |
| **MAX** | A listából használható IP-címek maximális száma. A „0” érték azt jelenti, hogy az összes IP-cím szerepelni fog. |
| **URL** | Az IP-lista forrásának URL-je. |


## Blokklisták


| Név | Kategória | Karbantartó | Leírás | Alapértelmezetten engedélyezve |
|------|----------|------------|--------------|--------------|
| [ABUSEIPDB](https://api.abuseipdb.com/api/v2/blacklist?key=YOUR_API_KEY&plaintext=1) | hírneve | [abuseipdb.com](https://www.abuseipdb.com/) | Hackelési kísérletben vagy más rosszindulatú viselkedésben részt vevő visszaélésszerű IP-címek IP-reputációs adatbázisa (**Ingyenes API-kulcsért [regisztrálnia kell](https://www.abuseipdb.com/register?plan=free) a webhelyükre, majd a forrás URL-ben le kell cserélnie a YOUR_API_KEY-t arra**). |  |
| [UNLIMITED_RS](https://blacklist.unl.rs/) | hírneve | [unlimited.rs](https://blacklist.unl.rs/) | UNLIMITED.RS támadó IP-címek (minden). |  |
| [BDE](https://api.blocklist.de/getlast.php?time=3600) | támadások | [blocklist.de](https://www.blocklist.de) | A Blocklist.de IP-címeket támad meg (elmúlt órában). |  |
| [BDEALL](https://lists.blocklist.de/lists/all.txt) | támadások | [blocklist.de](https://www.blocklist.de) | Blocklist.de támadó IP-címek (minden). |  |
| [BDS_ATIF](https://www.binarydefense.com/banlist.txt) | hírneve | [binarydefense.com](https://www.binarydefense.com/) | Tüzérségi fenyegetés hírszerzési hírcsatornája és kitiltási hírfolyama. |  |
| [BFB](https://danger.rulez.sk/projects/bruteforceblocker/blist.php) | támadások | [Daniel Gerzo](https://danger.rulez.sk/index.php/bruteforceblocker/) | BruteForceBlocker IP-lista. |  |
| [BLOCKLIST_NET_UA](https://blocklist.net.ua/blocklist.csv) | visszaélés | [blocklist.net.ua](https://blocklist.net.ua) | Segít megállítani a kéretlen leveleket és a kétes forrásból származó brutális erőszakos támadásokat. |  |
| [BOGON](https://www.team-cymru.org/Services/Bogons/bogon-bn-agg.txt) | irányíthatatlan | [team-cymru.org](https://www.team-cymru.org/Services/Bogons/) | Privát/lefoglalt IP-címek és ki nem osztott hálózati blokkok. |  |
| [BOTSCOUT](https://botscout.com/last_caught_cache.htm) | visszaélés | [botscout.com](https://botscout.com/) | Megakadályozza, hogy a robotok visszaéljenek az űrlapokkal, spammeljenek stb. |  |
| [CIARMY](https://cinsscore.com/list/ci-badguys.txt) | hírneve | [cinsscore.com](https://cinsscore.com/) | Rossz szélhámos csomagok IP-címei a CINS hadsereg listájáról. |  |
| [DARKLIST_DE](https://www.darklist.de/raw.php) | támadások | [darklist.de](https://www.darklist.de/) | SSH fail2ban jelentés. |  |
| [DSHIELD](https://feeds.dshield.org/block.txt) | támadások | [dShield.org](https://dshield.org/) | A legjobb 20 támadó C osztályú (/24) alhálózat 3 nap alatt. |  |
| [ET_BLOCK](https://rules.emergingthreats.net/fwrules/emerging-Block-IPs.txt) | támadások | [emergingthreats.net](https://www.emergingthreats.net/) | Alapértelmezett feketelista; jobb az egyes ipseteket használni. |  |
| [ET_COMPROMISED](https://rules.emergingthreats.net/blockrules/compromised-ips.txt) | támadások | [emergingthreats.net](https://www.emergingthreats.net/) | Kompromittált házigazdák. |  |
| [ET_TOR](https://rules.emergingthreats.net/blockrules/emerging-tor.rules) | anonimizálók | [emergingthreats.net](https://www.emergingthreats.net/) | TOR hálózati IP-címek. |  |
| [FEODO](https://feodotracker.abuse.ch/blocklist/?download=ipblocklist) | malware | [abuse.ch](https://feodotracker.abuse.ch/) | Feodo (Cridex/Bugat) trójai IP-címek. |  |
| [GREENSNOW](https://blocklist.greensnow.co/greensnow.txt) | támadások | [greenSnow.co](https://greensnow.co/) | Figyeli a nyers erőt, az FTP-t, az SMTP-t, az SSH-t stb. |  |
| [HONEYPOT](https://www.projecthoneypot.org/list_of_ips.php?t=d&rss=1) | támadások | [projecthoneypot.org](https://www.projecthoneypot.org) | A szótár támadói IP-címei. |  |
| [INTERSERVER_2D](https://sigs.interserver.net/ipslim.txt) | támadások | [interserver.net](https://sigs.interserver.net/) | Brute force/spam/rosszindulatú IP-címek (az elmúlt 2 napban). |  |
| [INTERSERVER_7D](https://sigs.interserver.net/ip.txt) | támadások | [interserver.net](https://sigs.interserver.net/) | Ugyanaz, mint fent (az elmúlt 7 napban). |  |
| [INTERSERVER_ALL](https://sigs.interserver.net/iprbl.txt) | támadások | [interserver.net](https://sigs.interserver.net/) | Minden ismert rosszindulatú IP-cím. |  |
| [SBLAM](https://sblam.com/blacklist.txt) | visszaélés | [sblam.com](https://sblam.com/) | Webes űrlapok spammerei. |  |
| [SPAMDROP](https://www.spamhaus.org/drop/drop.lasso) | spam | [spamhaus.org](https://www.spamhaus.org/drop/) | DROP – Útvonal vagy társlista ne legyen. |  |
| [SPAMDROPV6](https://www.spamhaus.org/drop/dropv6.txt) | spam | [spamhaus.org](https://www.spamhaus.org/drop/) | DROPv6 IPv6-hoz. |  |
| [SPAMEDROP](https://www.spamhaus.org/drop/edrop.lasso) | spam | [spamhaus.org](https://www.spamhaus.org/drop/) | Kibővített DROP lista (EDROP). |  |
| [SSLBL](https://sslbl.abuse.ch/blacklist/sslipblacklist.csv) | malware | [abuse.ch](https://sslbl.abuse.ch/) | Rosszindulatú programokkal/botnetekkel kapcsolatos SSL-forgalom. |  |
| [SSLBL_AGGRESSIVE](https://sslbl.abuse.ch/blacklist/sslipblacklist_aggressive.csv) | malware | [abuse.ch](https://sslbl.abuse.ch/) | Agresszív SSL feketelista (téves pozitív eredményeket okozhat). |  |
| [STOPFORUMSPAM](https://www.stopforumspam.com/downloads/bannedips.zip) | visszaélés | [stopforumspam.com](https://www.stopforumspam.com/) | Fórum spammer IP-címei. |  |
| [STOPFORUMSPAM_180D](https://www.stopforumspam.com/downloads/listed_ip_180.zip) | visszaélés | [stopforumspam.com](https://www.stopforumspam.com/) | Az elmúlt 180 nap. |  |
| [STOPFORUMSPAM_1D](https://www.stopforumspam.com/downloads/listed_ip_1.zip) | visszaélés | [stopforumspam.com](https://www.stopforumspam.com/) | Az elmúlt 24 óra. |  |
| [STOPFORUMSPAM_30D](https://www.stopforumspam.com/downloads/listed_ip_30.zip) | visszaélés | [stopforumspam.com](https://www.stopforumspam.com/) | Az elmúlt 30 nap. |  |
| [STOPFORUMSPAM_365D](https://www.stopforumspam.com/downloads/listed_ip_365.zip) | visszaélés | [stopforumspam.com](https://www.stopforumspam.com/) | Az elmúlt 365 nap. |  |
| [STOPFORUMSPAM_7D](https://www.stopforumspam.com/downloads/listed_ip_7.zip) | visszaélés | [stopforumspam.com](https://www.stopforumspam.com/) | Az elmúlt 7 nap. |  |
| [STOPFORUMSPAM_90D](https://www.stopforumspam.com/downloads/listed_ip_90.zip) | visszaélés | [stopforumspam.com](https://www.stopforumspam.com/) | Az elmúlt 90 nap. |  |
| [STOPFORUMSPAM_TOXIC](https://www.stopforumspam.com/downloads/toxic_ip_cidr.txt) | visszaélés | [stopforumspam.com](https://www.stopforumspam.com/) | Súlyos botaktivitású hálózatok. |  |
| [TOR](https://check.torproject.org/cgi-bin/TorBulkExitList.py?ip=1.2.3.4) | anonimizálók | [torproject.org](https://trac.torproject.org/projects/tor/wiki/doc/TorDNSExitList) | TOR kilépési csomópontok listája. |  |



**MEGJEGYZÉS: Ezek a listák nem tartoznak az OpenPanel felügyelete alá, és hamis pozitív eredményeket tartalmazhatnak.**
