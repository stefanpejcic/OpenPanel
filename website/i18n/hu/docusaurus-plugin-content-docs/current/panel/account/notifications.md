---
sidebar_position: 3
---

# Értesítések

A felhasználók beállíthatnak olyan műveleteket, amelyekről e-mailes értesítést kapnak:

![notifications.png](/img/panel/v2/notifications.png)

:::info
Ha nem látja az Értesítések oldalt, kérje meg szolgáltatóját, hogy engedélyezze az [értesítési modult] (/docs/admin/settings/openpanel/#notifications).
:::

| Érték | Leírás | Alapértelmezett |
|--------------------------------------------------------------------------------------------------|
| notify_account_login | Értesítés az új bejelentkezésekről | 1 |
| notify_account_login_for_known_netblock | Értesítés az ismert netblockból származó új bejelentkezéskor | 1 |
| notify_account_login_notification_disabled | Értesítés, ha a bejelentkezési értesítések le vannak tiltva | 1 |
| notify_autossl_expiry | Értesítés, ha az SSL-tanúsítvány hamarosan lejár | 0 |
| notify_autossl_expiry_coverage | Értesítés, ha az SSL-tanúsítvány lejárt | 0 |
| notify_autossl_renewal_coverage | Értesítés, ha a tanúsítvány megújítása sikertelen | 0 |
| notify_autossl_renewal_coverage_reduced | Értesítés, ha a tanúsítvány megújítása sikertelen | 0 |
| notify_autossl_renewal_uncovered_domains | Értesítés, ha a tanúsítvány megújítása sikertelen az egyéni SSL| esetén 0 |
| notify_contact_address_change | Értesítés, ha az e-mail cím megváltozik | 1 |
| notify_contact_address_change_notification_disabled | Értesítés, ha az e-mail cím módosítására vonatkozó értesítések le vannak tiltva | 1 |
| notify_disk_limit | Értesítés, ha a fiók elérte a lemezkorlátot | 1 |
| notify_email_quota_limit | Értesítés, ha az e-mail fiókok elérik a kvótakorlátjukat | 1 |
| notify_password_change | Értesítés a jelszó megváltoztatásakor | 1 |
| notify_password_change_notification_disabled | Értesítés, ha a jelszómódosításra vonatkozó értesítések le vannak tiltva | 1 |
| notify_ssl_expiry | Értesítés, ha az SSL-tanúsítvány lejárt | 1 |
| notify_twofactorauth_change | Értesítés, ha a kéttényezős hitelesítés le van tiltva/engedélyezve | 1 |
| notify_twofactorauth_change_notification_disabled | Értesítés, ha a kéttényezős hitelesítés módosítására vonatkozó értesítések le vannak tiltva | 1 |


A felhasználó nem kap értesítést az OpenAdmin felületen keresztül végrehajtott műveletekről (például felhasználó megszemélyesítése, jelszó megváltoztatása, 2FA letiltása stb.) vagy API-hívásokról (pl. WHMCS, Blesta, FOSSBilling).


## E-mail értesítések

![new_login.png](/img/panel/v1/account/new_login.png)

![2fa_disabled.png](/img/panel/v1/account/2fa_disabled.png)

![2fa_enabled.png](/img/panel/v1/account/2fa_enabled.png)

![email_changed.png](/img/panel/v1/account/email_changed.png)

![pass_changed.png](/img/panel/v1/account/pass_changed.png)

![preferences_changed.png](/img/panel/v1/account/preferences_changed.png)
