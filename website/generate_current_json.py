import os
import json
import glob
from deep_translator import GoogleTranslator

translator = GoogleTranslator(source='en', target='hu')

base_dir = "/Users/inak/.gemini/antigravity/scratch/OpenPanel/website/docs"
current_json_path = "/Users/inak/.gemini/antigravity/scratch/OpenPanel/website/i18n/hu/docusaurus-plugin-content-docs/current.json"

category_files = glob.glob(os.path.join(base_dir, "**", "_category_.json"), recursive=True)

result = {}

# Keep existing translations if the file exists
if os.path.exists(current_json_path):
    try:
        with open(current_json_path, 'r', encoding='utf-8') as f:
            result = json.load(f)
    except:
        pass

for original_path in category_files:
    rel_path = os.path.relpath(original_path, base_dir)
    parts = rel_path.split(os.sep)
    sidebar_name = parts[0]
    
    with open(original_path, 'r', encoding='utf-8') as f:
        data = json.load(f)
        
    if "label" in data:
        original_label = data["label"]
        translated_label = translator.translate(original_label)
        
        # Override specific translations based on earlier feedback
        if original_label.lower() == "docker":
            translated_label = "Docker"
        elif original_label.lower() == "transfers":
            translated_label = "Költöztetések"
            
        key = f"sidebar.{sidebar_name}.category.{original_label}"
        result[key] = {
            "message": translated_label,
            "description": f"The label for category {original_label} in sidebar {sidebar_name}"
        }

with open(current_json_path, 'w', encoding='utf-8') as f:
    json.dump(result, f, ensure_ascii=False, indent=2)

print("Generated current.json with translated category keys!")
