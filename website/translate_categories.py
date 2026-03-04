import os
import json
import glob
from deep_translator import GoogleTranslator

translator = GoogleTranslator(source='en', target='hu')

# Define paths
base_dir = "/Users/inak/.gemini/antigravity/scratch/OpenPanel/website/docs"
hu_dir = "/Users/inak/.gemini/antigravity/scratch/OpenPanel/website/i18n/hu/docusaurus-plugin-content-docs/current"

# Find all _category_.json files in the original docs directory
category_files = glob.glob(os.path.join(base_dir, "**", "_category_.json"), recursive=True)
print(f"Found {len(category_files)} category files.")

for original_path in category_files:
    # Determine corresponding path in hu_dir
    rel_path = os.path.relpath(original_path, base_dir)
    target_path = os.path.join(hu_dir, rel_path)
    
    # Ensure directory exists
    os.makedirs(os.path.dirname(target_path), exist_ok=True)
    
    # Read original json
    with open(original_path, 'r', encoding='utf-8') as f:
        data = json.load(f)
        
    # Translate the label if it exists
    if "label" in data:
        original_label = data["label"]
        try:
            translated_label = translator.translate(original_label)
            # Create a separate object for the translation overrides if we want, or just overwrite the localized one
            data["label"] = translated_label
            print(f"Translated '{original_label}' -> '{translated_label}'")
        except Exception as e:
            print(f"Failed to translate {original_label}: {e}")
            
    # Write to target path
    with open(target_path, 'w', encoding='utf-8') as f:
        json.dump(data, f, ensure_ascii=False, indent=2)

print("Finished translating category files.")
