# CSF Blocklists

CSF/LFD supports downloading and applying blocklists of IPs and CIDRs from public sources. 

OpenPanel does **not** enable any IP blocklists by default upon installation.

## from OpenAdmin

To enable a blocklist from OpenAdmin interface, navigate to **Security > Firewall** then scroll down and click on the 'LFD Blocklists':

[![2025-07-22-17-06.png](https://i.postimg.cc/LsdV3g1b/2025-07-22-17-06.png)](https://postimg.cc/CRNDF1JG)

**Uncomment** the line that starts with the desired blocklist, by removing the `#` before it, then click on 'Change':

[![2025-07-22-17-06-1.png](https://i.postimg.cc/jjz4gq6d/2025-07-22-17-06-1.png)](https://postimg.cc/sBgW1r0t)

Finally click on 'Restart csf+lfd':

[![2025-07-22-17-07.png](https://i.postimg.cc/KYrt7qF1/2025-07-22-17-07.png)](https://postimg.cc/nsrsp1Xx)


## from Terminal

To enable a specific blocklist:

1. Open file `/etc/csf/csf.blocklists`
2. **Uncomment** the line that starts with the desired blocklist.
3. Save the file.
4. **Restart CSF**, then **restart LFD**: `csf -ra && service lfd restart`





## Format

| Parameter | Description |
|-----------|-------------|
| **NAME**     | List name with all uppercase alphabetic characters, no spaces, and a maximum of 25 characters. This will be used as the iptables chain name. |
| **INTERVAL** | Refresh interval (in seconds) to download the list. Must be at least `3600` (1 hour), but `86400` (1 day) is generally sufficient. |
| **MAX**      | Maximum number of IP addresses to use from the list. A value of `0` means all IPs will be included. |
| **URL**      | URL of the IP list source. |


## Blocklists


| Name | Category | Maintainer | Description | Enabled by Default |
|------|----------|------------|-------------|-------------|
| [ABUSEIPDB](https://api.abuseipdb.com/api/v2/blacklist?key=YOUR_API_KEY&plaintext=1) | reputation | [abuseipdb.com](https://www.abuseipdb.com/) | IP reputation database of abusive IPs engaging in hacking attempts or other malicious behavior (**You must [sign up](https://www.abuseipdb.com/register?plan=free) to their website for a free API key then replace YOUR_API_KEY with it in the source URL**). |  |
| [UNLIMITED_RS](https://blacklist.unl.rs/) | reputation | [unlimited.rs](https://blacklist.unl.rs/) | UNLIMITED.RS attacking IP addresses (all). |  |
| [BDE](https://api.blocklist.de/getlast.php?time=3600) | attacks | [blocklist.de](https://www.blocklist.de) | Blocklist.de attacking IP addresses (last hour). |  |
| [BDEALL](https://lists.blocklist.de/lists/all.txt) | attacks | [blocklist.de](https://www.blocklist.de) | Blocklist.de attacking IP addresses (all). |  |
| [BDS_ATIF](https://www.binarydefense.com/banlist.txt) | reputation | [binarydefense.com](https://www.binarydefense.com/) | Artillery Threat Intelligence feed and banlist feed. |  |
| [BFB](https://danger.rulez.sk/projects/bruteforceblocker/blist.php) | attacks | [Daniel Gerzo](https://danger.rulez.sk/index.php/bruteforceblocker/) | BruteForceBlocker IP List. |  |
| [BLOCKLIST_NET_UA](https://blocklist.net.ua/blocklist.csv) | abuse | [blocklist.net.ua](https://blocklist.net.ua) | Helps stop spam and brute force attacks from dubious sources. |  |
| [BOGON](https://www.team-cymru.org/Services/Bogons/bogon-bn-agg.txt) | unroutable | [team-cymru.org](https://www.team-cymru.org/Services/Bogons/) | Private/reserved IPs and unallocated netblocks. |  |
| [BOTSCOUT](https://botscout.com/last_caught_cache.htm) | abuse | [botscout.com](https://botscout.com/) | Prevents bots from abusing forms, spamming, etc. |  |
| [CIARMY](https://cinsscore.com/list/ci-badguys.txt) | reputation | [cinsscore.com](https://cinsscore.com/) | Poor rogue packet score IPs from the CINS Army list. |  |
| [DARKLIST_DE](https://www.darklist.de/raw.php) | attacks | [darklist.de](https://www.darklist.de/) | SSH fail2ban reporting. |  |
| [DSHIELD](https://feeds.dshield.org/block.txt) | attacks | [dShield.org](https://dshield.org/) | Top 20 attacking class C (/24) subnets over 3 days. |  |
| [ET_BLOCK](https://rules.emergingthreats.net/fwrules/emerging-Block-IPs.txt) | attacks | [emergingthreats.net](https://www.emergingthreats.net/) | Default blacklist; better to use individual ipsets. |  |
| [ET_COMPROMISED](https://rules.emergingthreats.net/blockrules/compromised-ips.txt) | attacks | [emergingthreats.net](https://www.emergingthreats.net/) | Compromised hosts. |  |
| [ET_TOR](https://rules.emergingthreats.net/blockrules/emerging-tor.rules) | anonymizers | [emergingthreats.net](https://www.emergingthreats.net/) | TOR network IPs. |  |
| [FEODO](https://feodotracker.abuse.ch/blocklist/?download=ipblocklist) | malware | [abuse.ch](https://feodotracker.abuse.ch/) | Feodo (Cridex/Bugat) trojan IPs. |  |
| [GREENSNOW](https://blocklist.greensnow.co/greensnow.txt) | attacks | [greenSnow.co](https://greensnow.co/) | Monitors brute force, FTP, SMTP, SSH, etc. |  |
| [HONEYPOT](https://www.projecthoneypot.org/list_of_ips.php?t=d&rss=1) | attacks | [projecthoneypot.org](https://www.projecthoneypot.org) | Dictionary attacker IPs. |  |
| [INTERSERVER_2D](https://sigs.interserver.net/ipslim.txt) | attacks | [interserver.net](https://sigs.interserver.net/) | Brute force/spam/malicious IPs (last 2 days). |  |
| [INTERSERVER_7D](https://sigs.interserver.net/ip.txt) | attacks | [interserver.net](https://sigs.interserver.net/) | Same as above (last 7 days). |  |
| [INTERSERVER_ALL](https://sigs.interserver.net/iprbl.txt) | attacks | [interserver.net](https://sigs.interserver.net/) | All known malicious IPs. |  |
| [SBLAM](https://sblam.com/blacklist.txt) | abuse | [sblam.com](https://sblam.com/) | Web form spammers. |  |
| [SPAMDROP](https://www.spamhaus.org/drop/drop.lasso) | spam | [spamhaus.org](https://www.spamhaus.org/drop/) | DROP - Do not Route Or Peer List. |  |
| [SPAMDROPV6](https://www.spamhaus.org/drop/dropv6.txt) | spam | [spamhaus.org](https://www.spamhaus.org/drop/) | DROPv6 for IPv6. |  |
| [SPAMEDROP](https://www.spamhaus.org/drop/edrop.lasso) | spam | [spamhaus.org](https://www.spamhaus.org/drop/) | Extended DROP List (EDROP). |  |
| [SSLBL](https://sslbl.abuse.ch/blacklist/sslipblacklist.csv) | malware | [abuse.ch](https://sslbl.abuse.ch/) | SSL traffic related to malware/botnets. |  |
| [SSLBL_AGGRESSIVE](https://sslbl.abuse.ch/blacklist/sslipblacklist_aggressive.csv) | malware | [abuse.ch](https://sslbl.abuse.ch/) | Aggressive SSL blacklist (may cause false positives). |  |
| [STOPFORUMSPAM](https://www.stopforumspam.com/downloads/bannedips.zip) | abuse | [stopforumspam.com](https://www.stopforumspam.com/) | Forum spammer IPs. |  |
| [STOPFORUMSPAM_180D](https://www.stopforumspam.com/downloads/listed_ip_180.zip) | abuse | [stopforumspam.com](https://www.stopforumspam.com/) | Last 180 days. |  |
| [STOPFORUMSPAM_1D](https://www.stopforumspam.com/downloads/listed_ip_1.zip) | abuse | [stopforumspam.com](https://www.stopforumspam.com/) | Last 24 hours. |  |
| [STOPFORUMSPAM_30D](https://www.stopforumspam.com/downloads/listed_ip_30.zip) | abuse | [stopforumspam.com](https://www.stopforumspam.com/) | Last 30 days. |  |
| [STOPFORUMSPAM_365D](https://www.stopforumspam.com/downloads/listed_ip_365.zip) | abuse | [stopforumspam.com](https://www.stopforumspam.com/) | Last 365 days. |  |
| [STOPFORUMSPAM_7D](https://www.stopforumspam.com/downloads/listed_ip_7.zip) | abuse | [stopforumspam.com](https://www.stopforumspam.com/) | Last 7 days. |  |
| [STOPFORUMSPAM_90D](https://www.stopforumspam.com/downloads/listed_ip_90.zip) | abuse | [stopforumspam.com](https://www.stopforumspam.com/) | Last 90 days. |  |
| [STOPFORUMSPAM_TOXIC](https://www.stopforumspam.com/downloads/toxic_ip_cidr.txt) | abuse | [stopforumspam.com](https://www.stopforumspam.com/) | Networks with heavy bot activity. |  |
| [TOR](https://check.torproject.org/cgi-bin/TorBulkExitList.py?ip=1.2.3.4) | anonymizers | [torproject.org](https://trac.torproject.org/projects/tor/wiki/doc/TorDNSExitList) | TOR exit node list. |  |



**NOTE: These lists are not under the control of OpenPanel and could have false positives.**
