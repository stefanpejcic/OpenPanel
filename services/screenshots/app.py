from flask import Flask, jsonify, send_file
import os
from playwright.sync_api import sync_playwright
from hashlib import md5
import time
from apscheduler.schedulers.background import BackgroundScheduler

app = Flask(__name__)

CACHE_DIR = 'screenshots_cache'

def capture_screenshot(domain):
    url = f'http://{domain}'
    screenshot_filename = f'{md5(url.encode()).hexdigest()}.png'
    screenshot_path = os.path.join(CACHE_DIR, screenshot_filename)

    if not os.path.exists(screenshot_path):
        try:
            print(f"Starting playwright for {url}")
            with sync_playwright() as p:
                browser = p.chromium.launch()
                page = browser.new_page()
                print(f"Navigating to {url}")
                page.goto(url)
                time.sleep(2)
                print(f"Taking screenshot and saving to {screenshot_path}")
                page.screenshot(path=screenshot_path)
                browser.close()
                print(f"Browser closed")

            if os.path.exists(screenshot_path):
                print(f"Screenshot successfully saved at {screenshot_path}")
            else:
                print(f"Screenshot file not found at {screenshot_path} after saving")
                return None
        except Exception as e:
            print(f"Error capturing screenshot: {e}")
            return None
    else:
        print(f"Using cached screenshot at {screenshot_path}")

    return screenshot_path

def delete_old_screenshots():
    for filename in os.listdir(CACHE_DIR):
        filepath = os.path.join(CACHE_DIR, filename)
        # Check if the file hasn't been accessed in 24 hours (86400 seconds)
        if os.path.getatime(filepath) < time.time() - 86400:
            os.remove(filepath)
            print(f"Deleted old screenshot: {filepath}")

# Create a scheduler that runs in the background
scheduler = BackgroundScheduler()
scheduler.add_job(delete_old_screenshots, 'interval', days=1, start_date='2023-09-30 01:00:00')  # Runs daily at 1 AM
scheduler.start()

@app.route('/screenshot/<path:domain>')
def get_screenshot(domain):
    screenshot_path = capture_screenshot(domain)
    if screenshot_path and os.path.exists(screenshot_path):
        try:
            os.utime(screenshot_path)
            print(f"Sending screenshot file: {screenshot_path}")
            return send_file(screenshot_path, mimetype='image/png')
        except Exception as e:
            print(f"Error sending file: {e}")
            return jsonify({'error': 'Failed to send screenshot', 'details': str(e)}), 500
    else:
        print(f"Screenshot path does not exist: {screenshot_path}")
        return jsonify({'error': 'Failed to capture screenshot'}), 500

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=80, debug=True)

