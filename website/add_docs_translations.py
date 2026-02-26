import json
import os

code_json_path = "/Users/inak/.gemini/antigravity/scratch/OpenPanel/website/i18n/hu/code.json"

translations = {
    "docs.index.title.docFor": "Dokumentáció a ",
    "docs.index.title.and": "-hez, és az ",
    "docs.section.panel.desc1": "Az OpenPanel a végfelhasználói felület, amelyből az ügyfelek kezelhetik domaineiket, weboldalaikat, szolgáltatásaikat, és az e-mail fiókjaikat.",
    "docs.section.panel.desc2": "A dokumentáció ezen része olyan végfelhasználóknak szól, akik kizárólag az OpenPanel felületre léphetnek be (alapmezett port: 2083).",
    "docs.section.panel.btn": "OpenPanel Doksik",
    "docs.section.admin.desc1": "Az OpenAdmin a legfelsőbb szintű (root) vezérlőpult a szerveren, amiből az adminisztrátorok menedzselhetik a felhasználókat.",
    "docs.section.admin.desc2": "A dokumentáció ezen része adminisztrátoroknak és a supportnak szól, akik hozzáférnek az OpenAdminhoz is (alapértelmezett port: 2087).",
    "docs.section.admin.btn": "OpenAdmin Doksik"
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

print("Added docs index keys to code.json successfully")
