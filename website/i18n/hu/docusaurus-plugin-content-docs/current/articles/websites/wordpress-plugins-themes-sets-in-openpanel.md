# WordPress témák és beépülő modulok

OpenPanel-felhasználóként létrehozhat **témakészleteket** és **bővítménykészleteket** a fiókjához. Az ezekből a készletekből származó beépülő modulok és témák automatikusan telepítésre kerülnek minden új WordPress-webhelyen, amelyet az OpenPanel WP Manager használatával hoz létre. Ez különösen akkor hasznos, ha gyakran használja ugyanazokat a beépülő modulokat és témákat több webhelyen is, például az Elementor témát gyermektémával, az Elementor beépülő modult és így tovább.

:::info
A készletek használatához engedélyezni kell fiókjában a "fájlkezelő" és a "wordpress" funkciókat.
:::
---

## Témák beállítása

Témakészlet létrehozása:

1. Lépjen az **OpenPanel > WP Manager** elemre, és kattintson a **Témák** elemre.
2. Megnyílik egy új oldal egy fájlszerkesztővel, amely lehetővé teszi az automatikusan telepíteni kívánt témák konfigurálását.

**Utasítás:**

* Adjon hozzá **egy témát soronként**.
* Ha ingyenes témákat szeretne telepíteni a WordPress.org webhelyről, egyszerűen használja a téma slug-ját:
![theme slug](https://i.postimg.cc/JnRzksmy/theme-slug.png)

  ```
startup-business-elementor
  ```
* Prémium vagy harmadik féltől származó témák telepítéséhez adja meg a közvetlen letöltési URL-t:

  ```
http://s3.amazonaws.com/bucketname/my-theme.zip?AWSAccessKeyId=123&Expires=456&Signature=abcdef
  ```
* A „#” karakterrel kezdődő sorokat a rendszer megjegyzésként kezeli, és figyelmen kívül hagyja.

**Példa témakészlet fájl:**

```bash
# Free themes from wordpress.org
hello-elementor
startup-business-elementor

# Premium theme from a remote URL
http://s3.amazonaws.com/bucketname/my-theme.zip?AWSAccessKeyId=123&Expires=456&Signature=abcdef
```

Tetszőleges számú témát adhat hozzá. Ha elkészült, **kattintson a „Mentés” gombra**.

---

## Beépülő modulok beállítása

Beépülő modul készlet létrehozása:

1. Lépjen az **OpenPanel > WP Manager** elemre, és kattintson a **Plugins** elemre.
2. Megnyílik egy új oldal egy fájlszerkesztővel, amely lehetővé teszi az automatikusan telepíteni kívánt bővítmények konfigurálását.

**Utasítás:**

* Adjon hozzá **egy bővítményt soronként**.
* Ingyenes beépülő modulok telepítéséhez a WordPress.org webhelyről használja a beépülő modult:
![plugin slug](https://i.postimg.cc/9fdf0d3G/plugin-slug.png)

  ```
enable-wp-debug-toggle
  ```
* Prémium vagy harmadik féltől származó bővítmények telepítéséhez adja meg a közvetlen letöltési URL-t:

  ```
https://github.com/envato/wp-envato-market/archive/master.zip
  ```
* A „#” karakterrel kezdődő sorokat a rendszer megjegyzésként kezeli, és figyelmen kívül hagyja.

**Példa bővítménykészlet fájl:**

```bash
# Free plugins from wordpress.org
akismet
bbpress
enable-wp-debug-toggle

# Premium plugin from GitHub
https://github.com/envato/wp-envato-market/archive/master.zip

# Plugin from S3 with access key
http://s3.amazonaws.com/bucketname/my-plugin.zip?AWSAccessKeyId=123&Expires=456&Signature=abcdef
```

Annyi bővítményt adhat hozzá, amennyit csak akar. Ha elkészült, **kattintson a „Mentés” gombra**.

