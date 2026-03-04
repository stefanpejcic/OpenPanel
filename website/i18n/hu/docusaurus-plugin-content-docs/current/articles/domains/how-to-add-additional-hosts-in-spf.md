
# További gazdagépek hozzáadása az SPF rekordokhoz

Ha egy adott gazdagépet szeretne felvenni az összes tartomány SPF-rekordjaiba, kövesse az alábbi utasításokat.

---

## Meglévő domainekhez

A következő "sed" paranccsal hozzáfűzhet egy új gazdagépet (pl. "include:example.com") a szerver zónafájljaiban lévő összes SPF rekordhoz:

```
sed -i 's/\(v=spf1\)/\1 include:example.com/' /etc/bind/zones/*.zone
```

---

## Új domainekhez

Annak biztosítása érdekében, hogy az új domainek automatikusan felvegyék ezt a gazdagépet az SPF-rekordjukba:

1. Menjen ide:
"OpenAdmin > Domains > DNS Zone Templates".

2. Szerkessze a zóna sablont, és adja meg a kívánt gazdagépeket az SPF bejegyzésben:

   ```
"v=spf1 include:example.com ~all"
   ```

Ez biztosítja, hogy minden újonnan létrehozott tartomány a frissített SPF-házirendet használja.
