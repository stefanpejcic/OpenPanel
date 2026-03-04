# Az ImunifyAV beállítása

Az ImunifyAV rendszeresen ellenőrzi a felhasználói fájlokat, és figyelmezteti Önt, ha rosszindulatú programokat észlel.

---

Az ImunifyAV konfigurálása a vezérlőpulton keresztül:

1. Nyissa meg az **OpenAdmin > Biztonság > ImunifyAV** menüpontot.
2. Kattintson a **fogaskerék ikonra** a beállítások eléréséhez.
3. Állítsa be a vizsgálatok erőforrás-korlátait, és állítsa be a vizsgálat ütemezését – alapértelmezés szerint a vizsgálatok havonta futnak.

[![imunifyav.png](https://i.postimg.cc/PqmmF0JV/imunifyav.png)](https://postimg.cc/f3RtV2xY)


Alternatív megoldásként ezeket a konfigurációkat közvetlenül a terminálról is elvégezheti:

**Az ImunifyAV erőforrás-korlátainak beállítása**:

```bash
imunify-antivirus config update '{"MALWARE_SCAN_INTENSITY": {"cpu": 2}}'
imunify-antivirus config update '{"MALWARE_SCAN_INTENSITY": {"io": 2}}'
imunify-antivirus config update '{"MALWARE_SCAN_INTENSITY": {"ram": 1024}}'
imunify-antivirus config update '{"MALWARE_SCAN_INTENSITY": {"user_scan_cpu": 2}}'
imunify-antivirus config update '{"MALWARE_SCAN_INTENSITY": {"user_scan_io": 2}}'
imunify-antivirus config update '{"MALWARE_SCAN_INTENSITY": {"user_scan_ram": 1024}}'
imunify-antivirus config update '{"RESOURCE_MANAGEMENT": {"cpu_limit": 1}}'
imunify-antivirus config update '{"RESOURCE_MANAGEMENT": {"io_limit": 1}}'
imunify-antivirus config update '{"RESOURCE_MANAGEMENT": {"ram_limit": 500}}'
```

**A szkennelési ütemezés beállítása**:

```bash
imunify-antivirus config update '{"MALWARE_SCAN_SCHEDULE": {"day_of_month": 1}}'
imunify-antivirus config update '{"MALWARE_SCAN_SCHEDULE": {"hour": 3}}'
imunify-antivirus config update '{"MALWARE_SCAN_SCHEDULE": {"interval": "none"}}'
```

---

## Frissítsen ImunifyAV+-ra

Az Imunify AV+ licenc aktiválásához hajtsa végre a következő parancsot a kiszolgálón.  **Ha az Imunify AV ingyenes verzióját tervezi használni, kihagyhatja ezt a lépést**.

**Aktiválás aktiváló kulccsal**:
Az ImunifyAV+ licenc aktiválásához az aktiváló kulccsal futtassa a következő parancsokat:

```
imunify-antivirus unregister
imunify-antivirus register YOUR_KEY
```

Ahol a „YOUR_KEY” a licenckulcs, cserélje ki a tényleges kulcsra – próbaverzió vagy megvásárolt. A kulcs formátuma: "IMAVPXXXXXXXXXXXXXXX".

**Aktiválás IP-alapú licenccel**:
Ha rendelkezik IP-alapú licenccel, futtassa a következő parancsokat:

```
imunify-antivirus unregister
imunify-antivirus register IPL
```

----


További információkért tekintse meg az ImunifyAV dokumentációját
