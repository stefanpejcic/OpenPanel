---
sidebar_position: 3
---

# Címimportőr

A **Címimportáló** oldalon egyszerre több e-mail fiókot is létrehozhat egy `.xls`/`xlsx` vagy `.csv` fájl feltöltésével.

### E-mail címek importálása

1. Lépjen az **OpenPanel > E-mailek > Címimportőr** elemre.
2. Válassza ki a `.xls`, `.xlsx` vagy `.csv` fájlt eszközéről, majd kattintson a **Feltöltés** gombra.

### Fájlformátumra vonatkozó követelmények

* Támogatott elválasztójelek: "," (vessző) vagy ";" (pontosvessző)
* Kötelező oszlopok: "felhasználónév", "jelszó".
* Nem kötelező oszlop: "kvóta".

* Ha szerepel, a kvótának egy mértékegységet kell megadnia: "K", "M", "G" vagy "T".
* Ha nincs megadva mértékegység, az alapértelmezett **MB**.

### Példafájl

```csv
email,password,quota
alice@example.com,pass123,500M
bob@mydomain.com,s3cr3t,1G
charlie@other.com,charliepwd,
diana@example.com,d1@n@pwd,0
eve@unauthorized.com,evepass,100T
new@example.com,ssaaasa2,10M
```

### Felülvizsgálat és megerősítés

A fájl feltöltése után:

* Megjelenik a létrehozandó e-mail fiókok előnézete, valamint jelszavaik és kvótája (ha van).
* A rendszer megjelöli a bejegyzéseket, ha azok kimaradnak az importálás során:

* Már létező e-mail címek
* Olyan domainek, amelyek nincsenek társítva a fiókjával

### Utolsó lépés

Kattintson a **Feltöltés indítása** gombra az importálási folyamat elindításához. A táblázat alatt megjelenik egy napló, amely megmutatja a művelet előrehaladását és állapotát.

A befejezést követően az összes érvényes e-mail fiók sikeresen importálva lesz.
