import os
import glob
import time
import concurrent.futures
from deep_translator import GoogleTranslator

def translate_markdown_file(filepath):
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()

        lines = content.split('\n')
        translated_lines = []
        
        in_code_block = False
        in_frontmatter = False

        translator = GoogleTranslator(source='en', target='hu')

        for i, line in enumerate(lines):
            # Ignore frontmatter lines entirely
            if i == 0 and line.strip() == '---':
                in_frontmatter = True
                translated_lines.append(line)
                continue
                
            if in_frontmatter and line.strip() == '---':
                in_frontmatter = False
                translated_lines.append(line)
                continue
                
            # Ignore code blocks entirely
            if line.startswith('```'):
                in_code_block = not in_code_block
                translated_lines.append(line)
                continue
                
            # Add skipped lines as-is safely
            if in_code_block or in_frontmatter or not line.strip():
                translated_lines.append(line)
                continue

            # Try API translation, fallback to original if it fails
            retries = 3
            translated_line = None
            for attempt in range(retries):
                try:
                    translated_line = translator.translate(line)
                    if translated_line is not None:
                        break
                    time.sleep(2)
                except Exception as e:
                    time.sleep(2)
            
            if translated_line is None:
                translated_line = line # Fallback to original
                
            translated_lines.append(translated_line)
            
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write('\n'.join(translated_lines))
        return f"Success: {filepath}"
    except Exception as e:
        return f"Failed {filepath}: {e}"

if __name__ == "__main__":
    docs_dir = "/Users/inak/.gemini/antigravity/scratch/OpenPanel/website/docs"
    i18n_dir = "/Users/inak/.gemini/antigravity/scratch/OpenPanel/website/i18n/hu/docusaurus-plugin-content-docs/current"
    
    md_files = glob.glob(f"{i18n_dir}/**/*.md", recursive=True)
    md_files += glob.glob(f"{i18n_dir}/**/*.mdx", recursive=True)
    
    untranslated = []
    for filepath in md_files:
        rel_path = os.path.relpath(filepath, i18n_dir)
        orig_path = os.path.join(docs_dir, rel_path)
        if os.path.exists(orig_path):
            with open(filepath, 'r') as f:
                hu_content = f.read()
            with open(orig_path, 'r') as f:
                en_content = f.read()
            if hu_content == en_content:
                untranslated.append(filepath)
                
    print(f"Found {len(untranslated)} files left to translate.")
    
    with concurrent.futures.ThreadPoolExecutor(max_workers=3) as executor:
        futures = {executor.submit(translate_markdown_file, filepath): filepath for filepath in untranslated}
        
        completed = 0
        for future in concurrent.futures.as_completed(futures):
            completed += 1
            print(f"Translated {completed}/{len(untranslated)} files...")
            
    print("Translation completed.")
