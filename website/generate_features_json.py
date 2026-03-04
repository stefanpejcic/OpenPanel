import json
import re
import os

ts_file_path = "/Users/inak/.gemini/antigravity/scratch/OpenPanel/website/src/assets/integrations.ts"
code_json_path = "/Users/inak/.gemini/antigravity/scratch/OpenPanel/website/i18n/hu/code.json"

with open(ts_file_path, "r", encoding="utf-8") as file:
    content = file.read()

# Extract the message and id from the translate() calls
matches = re.findall(r'translate\(\{\s*message:\s*"([^"\\]*(?:\\.[^"\\]*)*)"\s*,\s*id:\s*"([^"\\]*(?:\\.[^"\\]*)*)"\s*\}\)', content)

# Also define the static keys we added to index.tsx
static_translations = {
    "features.hero.title": "Nem az, amit egy Vezérlőpulttól várnál",
    "features.hero.subtitle.1": "A nyílt forráskódú OpenPanel az egyik legjobban testreszabható web hosting vezérlőpult a piacon.",
    "features.hero.subtitle.2": "Rendelj felhasználónként különböző MySQL verziókat, webkiszolgálókat és limiteket fiókonként, engedélyezd nekik a Docker konténerek futtatását, kezeld a biztonsági másolatokat, ütemezz cron feladatokat másodpercek alatt, finomhangold a konfigurációkat, és még sok mást.",
    "features.section.1": "Saját webkiszolgáló minden felhasználónak",
    "features.section.2": "Minden klasszikus funkció együtt",
    "features.section.3": "Okosabb szerver menedzsment",
    "features.section.4": "Egyszerűsített kezelőfelület",
    "features.section.5": "Menedzseld felhasználóid egy gombnyomásra",
    "features.section.6": "Beépített izoláció és biztonság",
    "features.section.7": "Integráld a számlázó rendszeredbe"
}

from deep_translator import GoogleTranslator
translator = GoogleTranslator(source='en', target='hu')

# Load the current json file
if os.path.exists(code_json_path):
    with open(code_json_path, 'r', encoding='utf-8') as f:
        data = json.load(f)
else:
    data = {}

# Populate the JSON
print(f"Found {len(matches)} tags to translate...")

for msg, key in matches:
    # unescape string
    msg = msg.replace('\\"', '"')
    try:
        translated = translator.translate(msg)
    except Exception as e:
        print(f"Failed to translate '{msg[:20]}': {e}")
        translated = msg # fallback

    data[key] = {"message": translated, "description": "Translated automatically"}

for key, value in static_translations.items():
    data[key] = {"message": value, "description": "Translated static texts"}

with open(code_json_path, 'w', encoding='utf-8') as f:
    json.dump(data, f, ensure_ascii=False, indent=2)

print("Add features translation keys to json!")
