---
sidebar_position: 2
---

# Tűzfal

Az OpenPanel támogatja a [Sentinel Firewall](https://sentinelfirewall.org/) -t – a *ConfigServer Security and Firewall* (CSF) elágazását.

## CSF

A tűzfal (CSF) felhasználói felülete az **OpenAdmin > Firewall** oldalon jelenik meg.

A CSF UI használatára vonatkozó utasításokért tekintse meg a [Sentinel Firewall hivatalos dokumentációját] (https://sentinelfirewall.org/docs/usage/introduction/).

![csf tűzfal](/img/admin/firewall_csf.png)


## Külső tűzfal

Egyes felhőszolgáltatók, például a [Hetzner](https://docs.hetzner.com/robot/dedicated-server/firewall/) saját külső tűzfalakat kínálnak. Ha külső tűzfalat használ, győződjön meg arról, hogy a következő portok nyitva vannak az OpenPanel szolgáltatások eléréséhez: `53` `80` `443` `2083` `2087` `32768:60999`

Ha [egyéni portot használ az OpenPanelhez az alapértelmezett 2083 helyett] (/docs/admin/settings/general/#change-openpanel-port), győződjön meg arról, hogy a port is nyitva van.

