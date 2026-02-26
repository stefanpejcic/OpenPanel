# Az OpenPanel biztosítása

Ez a rész azokat a bevált módszereket és beállításokat ismerteti, amelyek növelhetik OpenPanel-kiszolgálójának biztonságát, és ezáltal megvédhetik azt a különféle típusú támadásoktól és az érzékeny adatok elvesztése ellen:

Szerveradminisztrátoroknak (OpenAdmin):

- [Alapszintű hozzáférési hitelesítés engedélyezése az OpenAdmin számára](/docs/admin/security/basic_auth/)
- [A CorazaWAF-szabályok konfigurálása](/docs/admin/security/waf/)
- [Az OpenPanel vagy az OpenAdmin portok módosítása] (/docs/admin/settings/general/#ports)
- [Az OpenAdmin hozzáférésének korlátozása](/docs/articles/dev-experience/limit_access_to_openadmin/)
- [Ellenőrizze a jelszavakat a gyengepass.com listákon](/docs/admin/settings/openpanel/#display)
- [Az OpenPanel rendszergazdai felhasználónév módosítása](/docs/admin/accounts/administrators/#create-new-admin)
- [Az OpenPanel és a levelezőszerver biztonságossá tétele SSL/TLS-tanúsítványokkal](/docs/admin/settings/general/#domain)
- [A funkciókhoz való hozzáférés korlátozása tárhelytervek alapján](/docs/admin/plans/feature-manager/#use-cases)
- [A Fail2ban beállítása e-mailhez](/docs/articles/email/how-to-setup-fail2ban-mailserver-openpanel/)
- [DKIM beállítása e-mailekhez](/docs/articles/email/how-to-setup-dkim-for-mailserver/)
- [ImunifyAV beállítása](/docs/articles/security/setup-imunifyav/)
- [CSF blokklisták engedélyezése](/docs/articles/security/csf-blocklists/)
- [Kapcsolatok korlátozása a CSF segítségével](/docs/articles/security/how-to-limit-connections-csf/)
- [Rate-limiting sikertelen Openpanel bejelentkezés](/docs/articles/dev-experience/rate_limiting_for_openpanel/)
- [SSH root felhasználói jelszó módosítása](/docs/admin/advanced/root-password/)
- [SSH-port módosítása](/docs/admin/advanced/ssh/#basic-ssh-settings/)
- [SSH-kulcsok konfigurálása](/docs/admin/advanced/ssh/#authorized-ssh-keys)
- [Terminál letiltása az OpenAdminban](docs/articles/dev-experience/disable-openadmin-web-terminal/)
- [Szerver újraindításának letiltása az OpenAdminban](docs/articles/dev-experience/disable-openadmin-server-reboot/)
- [Az OpenAdmin hozzáférés letiltása](/docs/admin/security/disable-admin/)

Webhely-adminisztrátoroknak (OpenPanel):
- [A WAF-védelem engedélyezése domainenként](/docs/panel/advanced/waf/)
- [Automatikus biztonsági mentések konfigurálása](/docs/panel/files/backups/)
- [Fix Files Permissions](/docs/panel/files/fix_permissions/)
- [A webhely fájljainak ellenőrzése rosszindulatú programokra] (/docs/panel/files/malware-scanner/)
- [Feltört webhely felfüggesztése](/docs/panel/domains/suspend/)
- [A többtényezős hitelesítés beállítása az OpenPanelben](/docs/panel/account/2fa/)
- [Automatikus SSL konfigurálása](/docs/panel/domains/ssl/#autossl)
- [Fióktevékenységi napló áttekintése](/docs/panel/account/account_activity/)
- [Az aktív munkamenetek áttekintése](/docs/panel/account/active_sessions/)
- [A bejelentkezési előzmények megtekintése](/docs/panel/account/login_history/)
