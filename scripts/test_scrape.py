import cloudscraper
import json

scraper = cloudscraper.create_scraper()
resp = scraper.get("https://warthunder.com/en/community/userinfo/?nick=Dark%23598", timeout=30)
print("Status:", resp.status_code)
if resp.status_code == 200:
    print("Body length:", len(resp.text))
    import re
    matches = re.findall(r"window\.mainData\s*=\s*({.*?});", resp.text, re.DOTALL)
    if matches:
        print("Found mainData JSON")
        data = json.loads(matches[0])
        print(json.dumps(data, indent=2, ensure_ascii=False)[:2000])
    else:
        print("No mainData found, first 1000 chars:")
        print(resp.text[:1000])
else:
    print("Error body:", resp.text[:500])
