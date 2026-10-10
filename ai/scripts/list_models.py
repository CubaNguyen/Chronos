"""List tất cả Gemini models available với API key hiện tại."""

import os
import requests
from dotenv import load_dotenv

load_dotenv()

API_KEY = os.getenv("GEMINI_API_KEY")
if not API_KEY:
    raise SystemExit("❌ GEMINI_API_KEY not set in .env")

url = f"https://generativelanguage.googleapis.com/v1beta/models?key={API_KEY}"
r = requests.get(url, timeout=10)
r.raise_for_status()

data = r.json()
print(f"📋 Found {len(data.get('models', []))} models:\n")

for m in data.get("models", []):
    name = m["name"].replace("models/", "")
    methods = m.get("supportedGenerationMethods", [])
    if "generateContent" in methods:
        print(f"  ✅ {name}")