import re
import json

ts_file_path = "/Users/inak/.gemini/antigravity/scratch/OpenPanel/website/src/assets/integrations.ts"

with open(ts_file_path, "r", encoding="utf-8") as file:
    content = file.read()

# Replace .ts with .tsx if needed, or import the translate function
if "import { translate } from '@docusaurus/Translate';" not in content:
    content = "import { translate } from '@docusaurus/Translate';\n" + content

# Function to wrap strings in translate()
def wrap_translate(match):
    key = match.group(1)
    value_string = match.group(2)
    # Remove surrounding quotes (both single and double) but keep internal structure
    clean_val = value_string.strip('\'"').replace('"', '\\"')
    
    # We create a simple dynamic id based on the sanitized content
    safe_id = "".join(c if c.isalnum() else "_" for c in clean_val[:20]).strip("_").lower()
    
    return f"{key}: translate({{ message: \"{clean_val}\", id: \"features.item.{safe_id}\" }})"

# Regex to match `name: "..."` or `description: "..."`
content = re.sub(r'(name|description):\s*("[^"\\]*(?:\\.[^"\\]*)*"|\'[^\'\\]*(?:\\.[^\'\\]*)*\')', wrap_translate, content)

with open(ts_file_path, "w", encoding="utf-8") as file:
    file.write(content)

print("Translation wrapping applied.")
