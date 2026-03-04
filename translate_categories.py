import os
import json
import glob
import time
from deep_translator import GoogleTranslator

docs_dir = "/Users/inak/.gemini/antigravity/scratch/OpenPanel/website/docs"
i18n_dir = "/Users/inak/.gemini/antigravity/scratch/OpenPanel/website/i18n/hu/docusaurus-plugin-content-docs/current"

category_files = glob.glob(f"{docs_dir}/**/_category_.json", recursive=True)
translator = GoogleTranslator(source='en', target='hu')

print(f"Found {len(category_files)} _category_.json files to translate.")

trans_cache = {
    "Accounts": "Fiókok",
    "Domains": "Domainek",
    "Emails": "Emailek",
    "Plans": "Csomagok",
    "Security": "Biztonság",
    "Server": "Szerver",
    "Services": "Szolgáltatások",
    "Settings": "Beállítások",
    "Advanced": "Haladó",
    "Backups": "Biztonsági mentések",
    "Databases": "Adatbázisok",
    "Developer Experience": "Fejlesztői élmény",
    "Docker": "Docker",
    "Extensions": "Kiegészítők",
    "Files": "Fájlok",
    "Install & Update": "Telepítés és frissítés",
    "License": "Licenc",
    "Operating Systems": "Operációs rendszerek",
    "Support": "Támogatás",
    "Transfers": "Átvitelek",
    "User Experience": "Felhasználói élmény",
    "Web Servers": "Webszerverek",
    "Websites": "Weboldalak",
}

for filepath in category_files:
    rel_path = os.path.relpath(filepath, docs_dir)
    target_path = os.path.join(i18n_dir, rel_path)
    
    os.makedirs(os.path.dirname(target_path), exist_ok=True)
    
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            data = json.load(f)
            
        if "label" in data:
            english_label = data["label"]
            translated_label = trans_cache.get(english_label)
            
            if not translated_label:
                retries = 3
                for _ in range(retries):
                    try:
                        translated_label = translator.translate(english_label)
                        if translated_label:
                            break
                        time.sleep(1)
                    except:
                        time.sleep(1)
            
            if translated_label:
                data["label"] = translated_label
                print(f"Translated: {english_label} -> {translated_label}")
                
        with open(target_path, 'w', encoding='utf-8') as f:
            json.dump(data, f, indent=2, ensure_ascii=False)
            
    except Exception as e:
        print(f"Failed to process {filepath}: {e}")

print("Done translating categories.")
