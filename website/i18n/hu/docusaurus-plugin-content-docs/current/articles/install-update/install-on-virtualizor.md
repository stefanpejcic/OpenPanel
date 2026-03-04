# Az OpenPanel telepítése a Virtualizoron

Az OpenPanel hozzáadása a Virtualizor-kiszolgálóhoz további vezérlőpultként, amelyet a felhasználók telepíthetnek:

1. Jelentkezzen be a szerverére:
* rootként SSH-n keresztül, vagy
* sudo jogosultságokkal rendelkező felhasználóként, és írja be a `sudo -i` parancsot.

2. Másolja ki és illessze be a következő parancsot a terminálba:
```bash
bash <(curl -sSL https://raw.githubusercontent.com/stefanpejcic/OpenPanel-Virtualizor/refs/heads/main/INSTALL.sh)
```
