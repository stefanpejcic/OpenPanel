# Több fiók törlése

Az `opencli user-delete` paranccsal tömegesen törölheti az OpenPanel felhasználói fiókokat. **Legyen óvatos** – könnyű véletlenül nem a megfelelő felhasználókat törölni.

### Egyetlen felhasználó törlése

```bash
opencli user-delete USERNAME
```

A rendszer megerősítést kér. A megerősítés kihagyásához használja az "-y" jelzőt:

```bash
opencli user-delete USERNAME -y
```

### Minden felhasználó törlése

**összes** felhasználó törlése a szerveren:

```bash
opencli user-delete -all
```

Alapértelmezés szerint a rendszer minden törlés megerősítésére kéri. Az összes megerősítés kihagyásához használja az "-y" jelzőt:

```bash
opencli user-delete -all -y
```
