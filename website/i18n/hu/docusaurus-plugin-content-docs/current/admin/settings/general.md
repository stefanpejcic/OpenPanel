---
sidebar_position: 1
---

# Általános beállítások

Ezen az oldalon a rendszergazdák konfigurálhatják a tartományt az OpenPanel és az OpenAdmin felületek, valamint a szolgáltatások portjainak elérésére.

![openadmin általános panelbeállítások](/img/admin/adminpanel_general_settings.png)

## Domain

Az OpenAdmin és az OpenPanel domain néven (például srv.your-domain.com:2083) keresztüli elérésének engedélyezéséhez kövesse az alábbi három lépést:

1. Állítsa be a gazdagép nevét:
Állítsa be a kívánt aldomaint szerver hosztnévként:
   ```
hostnameectl set-hostname server.example.net
   ```
2. Konfigurálja a DNS-t:
Irányítsa az aldomaint a kiszolgáló nyilvános IP-címére.
Használjon olyan eszközt, mint például a https://www.whatsmydns.net/ annak ellenőrzésére, hogy a tartomány a szerver ip-jére mutat-e.
   
3. Állítsa be az Általános beállításokban
Váltson IP-ről a domain névre az *OpenAdmin > Beállítások > Általános beállítások* menüpontban.

## Állítsa be az OpenPanel IP-címét

Az OpenPanel és az OpenAdmin szerver nyilvános IP-címén keresztüli eléréséhez válassza az „IP-cím” lehetőséget, és kattintson a Mentés gombra. A módosítás azonnali, mentéskor átirányítja az adminisztrációs panel IP:2087 címére.

## Portok

Az OpenAdmin és az OpenPanel interfészek portkonfigurációi módosíthatók az alapértelmezett beállításokról (2087 az OpenAdmin és 2083 az OpenPanel esetében).

- Ha az OpenPanel portját az alapértelmezett 2083-as értékről egy másik értékre szeretné módosítani, egyszerűen beállíthatja a kívánt portot az "OpenPanel Port" mezőben. Fontos megjegyezni, hogy a portnak 1000-33000 tartományba kell esnie.
- Az adminisztrátori port módosítása 2087-ről jelenleg nem lehetséges, mivel a külső eszközöknek, például a WHMCS-nek nincs lehetőségük egyéni port beállítására.

## Átirányítás

Alapértelmezés szerint, amikor a felhasználók tartományt adnak hozzá, az „/openpanel” karakterlánc hozzáadása a tartomány URL-jéhez átirányítja őket az OpenPanel felületére. Ezt azonban rugalmasan személyre szabhatja, például „/awesome”-ra változtatva, lehetővé téve a felhasználók számára, hogy a „saját-domain.com/wesome” címen keresztül hozzáférjenek OpenPanel-jükhöz.

Ha az "/openpanel"-t valami másra szeretné módosítani, egyszerűen állítsa be az "OpenPanel is elérhető:" mező értékét, majd kattintson a Mentés gombra. A változtatások azonnal, a szolgáltatás megszakítása nélkül lépnek életbe.

