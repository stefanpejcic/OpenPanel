import json
import os

code_json_path = "/Users/inak/.gemini/antigravity/scratch/OpenPanel/website/i18n/hu/code.json"

translations = {
    "homepage.sweetSpot.title": "A {sweetSpot} a megosztott tárhely és a VPS között.",
    "homepage.sweetSpot.title.highlight": "tökéletes egyensúly",
    "homepage.sweetSpot.desc": "Minden OpenPanel felhasználó elszigetelt, dedikált környezetet élvezhet, ahol futtathat szolgáltatásokat, meghatározhat PHP/MySQL korlátokat, és konfigurálhat külső elemeket, mint például a Varnish, Redis és Elasticsearch.",
    "homepage.sweetSpot.list.1.title": "Nginx, Apache, Varnish",
    "homepage.sweetSpot.list.1.desc": "Mindegyik felhasználóhoz külön-külön kiválaszthatod az Nginx, OpenLitespeed, Apache, OpenResty vagy Varnish használatát.",
    "homepage.sweetSpot.list.1.icon": "Nginx & OpenLitespeed",
    "homepage.sweetSpot.list.2.title": "Elszigetelt Docker szolgáltatások",
    "homepage.sweetSpot.list.2.desc": "Minden szolgáltatás egy elszigetelt Docker konténerben fut, ami a felhasználó által teljesen testreszabható.",
    "homepage.sweetSpot.list.2.icon": "Konténeres szolgáltatások",
    "homepage.sweetSpot.list.3.title": "Saját adatbázis szerver",
    "homepage.sweetSpot.list.3.desc": "Minden felhasználó saját dedikált MySQL vagy MariaDB példánnyal rendelkezik, és módosíthatja az adatbázis konfigurációját.",
    "homepage.sweetSpot.list.3.icon": "MySQL, MariaDB, Percona",
    "homepage.sweetSpot.list.4.title": "Gyorsítótár",
    "homepage.sweetSpot.list.4.desc": "A REDIS, a Memcached és az Elasticsearch mindössze egy kattintással telepíthető bármely fiók alatt.",
    "homepage.sweetSpot.list.4.icon": "REDIS & Varnish",
    "homepage.sweetSpot.list.5.title": "PHP, Python, NodeJS",
    "homepage.sweetSpot.list.5.desc": "Mindegyik domain név alatt más PHP és NodeJS verziókat hozhatnak létre, szerkeszthetik a .ini kiterjesztésű fájljaikat.",
    "homepage.sweetSpot.list.5.icon": "Multiverzió támogatás",
    "homepage.sweetSpot.list.6.title": "CSF, CorazaWAF és ImunifyAV",
    "homepage.sweetSpot.list.6.desc": "Mind a ConfigServer Sentinel Firewall (CSF), mind pedig a modsecurity (CorazaWAF) integrálva van a rendszerben.",
    "homepage.sweetSpot.list.6.icon": "CorazaWAF és ImunifyAV",
    "homepage.alreadyInvented.list.1": "Erőforrás korlátok",
    "homepage.alreadyInvented.list.2": "Távoli mentések",
    "homepage.alreadyInvented.list.3": "Szolgáltatáskezelő",
    "homepage.alreadyInvented.list.4": "Saját arculat",
    "homepage.alreadyInvented.list.5": "Több felhasználó",
    "homepage.alreadyInvented.title": "Spanyolviasz? Már feltalálták.",
    "homepage.alreadyInvented.desc": "A webtárhelyek üzemeltetéséhez szükséges összes eszköz be van építve, csökkentve az amúgy nehézkes harmadik féltől származó szolgáltatások telepítésének idejét. Nem kell többet külön vásárolnod Backup Pro, izolációs réteget, vagy éppen WordPress menedzsment eszközt.",
    "homepage.alreadyInvented.cta": "Funkciók megismerése",
    "homepage.trustedBy.title": "Jelentős tárhelyszolgáltatók bíznak benne",
    "homepage.heroAnimation.section.1": "SZOFTVEREK",
    "homepage.heroAnimation.section.2": "TECHNOLÓGIA",
    "homepage.heroAnimation.section.3": "WEB SZERVEREK",
    "homepage.heroAnimation.section.4": "ESZKÖZÖK",
    "homepage.packages.title": "Startolj {faster}, karbantarts {easier}, növekedj {indefinitely}.",
    "homepage.packages.title.faster": "gyorsabban",
    "homepage.packages.title.easier": "könnyebben",
    "homepage.packages.title.indefinitely": "végtelenül",
    "homepage.packages.subtitle": "Több mint 100 parancssoros utasítás",
    "homepage.packages.desc": "Az OpenCLI egy központi parancssori eszköz amellyel egy kattintással felépíthetők a funkciók elhagyva a vizuális kezelőfelületet.",
    "homepage.packages.cta": "Összes Parancs",
    "homepage.pureReact.title": "100% Kontroll",
    "homepage.pureReact.desc": "Ne érd be a drága és elavult hoszting rendszerekkel. Az OpenPanel-el 100%-os kontrollt kapsz a szerveradataid és a működése felett bármikor.",
    "homepage.pureReact.cta": "Dokumentáció"
}

if os.path.exists(code_json_path):
    with open(code_json_path, 'r', encoding='utf-8') as f:
        data = json.load(f)
else:
    data = {}

for key, value in translations.items():
    data[key] = {"message": value, "description": "Translated for Hungarian locale"}

with open(code_json_path, 'w', encoding='utf-8') as f:
    json.dump(data, f, ensure_ascii=False, indent=2)

print("Added keys to code.json successfully")
