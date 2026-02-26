# Az OpenAdmin korlátozása

Ha az OpenAdmin hozzáférést csak a csapatára szeretné korlátozni, tegye engedélyezőlistára a kiszolgáló IP-címeit a tűzfalon, majd tiltsa le a 2087-es portot.

## Tegye engedélyezőlistára csapata IP-címeit

Cserélje ki a *YOUR_TEAM_IP* címet a tényleges IP-címeivel. Ismételje meg minden csapat IP-jére.

```
csf -a YOUR_TEAM_IP
```

## Minden más hozzáférés letiltása a 2087-es porthoz

Szerkessze az `/etc/csf/csf.conf` fájlt, és távolítsa el a *2087*-et az engedélyezett *TCP_IN* listáról, vagy futtassa:

```
csf -d 0.0.0.0/0 2087
```

Ezután indítsa újra a CSF-et:

```
csf -r
```
