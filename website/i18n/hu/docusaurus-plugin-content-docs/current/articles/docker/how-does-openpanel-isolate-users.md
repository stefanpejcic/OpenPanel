# Felhasználó elkülönítése

Az OpenPanel **security-first architektúrával** épül fel, több szintű elszigetelést biztosítva a felhasználói környezet védelme érdekében.

Elszigetelő rétegek:

**Rendszer leválasztás**
Minden OpenPanel felhasználó megfelel egy rendszerfelhasználónak a gazdagépen. Ezek a felhasználók nem rendelkeznek bejelentkezési hozzáféréssel vagy jelszavakkal, és szigorúan a **tárhelykvóták és a tulajdonjog** érvényesítésére használják őket.

**Felhasználói elkülönítés**
Minden OpenPanel-fiók egy **dedikált Docker-környezetben** fut. Ez biztosítja, hogy egy felhasználó ne férhessen hozzá más felhasználókhoz vagy tárolóikhoz, ne zavarja őket, vagy akár észlelje is a létezését.

* **Szolgáltatás elkülönítése**
A felhasználói környezet minden szolgáltatása (például PHP, MySQL, Redis) a saját **saját tárolójában** fut. A szolgáltatások még egymástól is homokozóba vannak helyezve, így ha egy kompromittálódott (pl. MySQL), az **nem tud hatással más tárolókra**, még ugyanazon a felhasználón belül sem.

**Hálózati leválasztás**
A felhasználói szolgáltatások **belső Docker-hálózatokra** vannak felosztva, például:

* "www" a webes komponensekhez (PHP, Nginx, fájlkezelő)
* `db` adatbázisokhoz (MySQL, MariaDB, PostgreSQL)
* "nincs" az elszigetelt szolgáltatásokhoz (Redis, Memcached)

Ez a kialakítás lehetővé teszi a finom vezérlést afelől, hogy mely szolgáltatások kommunikálhatnak, és támogatja a **felhasználónkénti sávszélesség-szabályozást** is.


```
╔════════════════════════════════════════════════════════════════╗
║                     🖥️  OPENPANEL SERVER                       ║
╠════════════════════════════════════════════════════════════════╣
║  • 🎛️ OpenPanel - user control panel                           ║
║  • ⚙️ OpenAdmin - administration panel                         ║
║  • 🌐 Caddy – Reverse Proxy & SSL                              ║
║  • 🔍 BIND9 – DNS Server                                       ║
║  • 🗄️ MySQL – User Management & Metadata                       ║
║  • 🐳 Docker Engine – Container Orchestration                  ║
╚════════════════════════════════════════════════════════════════╝
                                                   │   
        ┌──────────────────────────────────────────┼──────────────────────────────────────────┐
        │                                          │                                          │
        ▼                                          ▼                                          ▼
┌─────────────────────────────────┐ ┌─────────────────────────────────┐ ┌─────────────────────────────────┐
│           👤 USER 1             │ │           👤 USER 2             │ │           👤 USER 3             │ 
├─────────────────────────────────┤ ├─────────────────────────────────┤ ├─────────────────────────────────┤
│  🌐 Web Server:                 │ │  🌐 Web Server:                 │ │  🌐 Web Server:                 │
│  • Nginx + Varnish              │ │  • Apache                       │ │  • OpenResty + Varnish          │
│                                 │ │                                 │ │                                 │
│  ⚡ Applications:               │ │  ⚡ Applications:               │ │  ⚡ Applications:               │
│  • site1.com → PHP 8.4          │ │  • api.site.com → Node.js 20.1  │ │  • classic.com → PHP 7.0        │
│  • site2.com → PHP 8.2          │ │  • main.site.com → PHP 7.4      │ │  • modern.com → PHP 8.1         │
│  • legacy.com → PHP 7.0         │ │                                 │ │  • vintage.com → PHP 5.6        │
│                                 │ │                                 │ │  • api.site.com → Python 3.11   │
│                                 │ │                                 │ │                                 │
│  🗄️  Databases:                 │ │  🗄️  Databases:                 │ │  🗄️  Databases:                 │
│  • MySQL 8.0                    │ │  • MariaDB 10.11                │ │  • PostgreSQL                   │
│  • phpMyAdmin                   │ │  • phpMyAdmin                   │ │                                 │
├─────────────────────────────────┤ ├─────────────────────────────────┤ ├─────────────────────────────────┤
│  📊 Resource Limits:            │ │  📊 Resource Limits:            │ │  📊 Resource Limits:            │
│  • CPU: 2 cores                 │ │  • CPU: 4 cores                 │ │  • CPU: 1 core                  │
│  • RAM: 4 GB                    │ │  • RAM: 8 GB                    │ │  • RAM: 2 GB                    │
│  • Storage: 50 GB SSD           │ │  • Storage: 100 GB SSD          │ │  • Storage: 25 GB SSD           │
└─────────────────────────────────┘ └─────────────────────────────────┘ └─────────────────────────────────┘
```
