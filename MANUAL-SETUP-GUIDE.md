# Vats Apartment — Manual Setup Guide

## Items that MUST be done manually (BrowserMCP can't interact with GTM/Ads SPAs)

---

## 1. GTM CONTAINER IMPORT (2 minutes)

1. Open: https://tagmanager.google.com/#/admin/accounts/6360055644/containers/255041451/import
2. Click **"Choose container file"**
3. Select file: `gtm-container-import.json` (in project root)
4. Keep **"Overwrite"** selected (container is empty)
5. Click **Confirm**
6. This imports:
   - 3 Triggers: `CE - WhatsApp Click`, `CE - Phone Call Click`, `CE - Booking Form Submit`
   - 4 Variables: `DLV - Event Source`, `DLV - Room ID`, `DLV - Room Type`, `DLV - Room Price`
   - 3 Tags: `Conversion - WhatsApp Click`, `Conversion - Phone Call Click`, `Conversion - Booking Form Submit`
7. Click **Submit** (top right) → **Publish** → name it "Initial setup"

---

## 2. GOOGLE ADS CONVERSION ACTION (3 minutes)

1. Open: https://ads.google.com/aw/conversions?ocid=8306148154&authuser=0
2. Click **"+ New Conversion Action"**
3. Select **"Website"**
4. Enter `vatsapartment.com`
5. Click **Scan**
6. If found, select it. If not, click **"Create conversion action manually"**
7. Settings:
   - Name: `WhatsApp/Phone Inquiry`
   - Category: **Contact** (or Submit Lead Form)
   - Value: Don't use a value (or use ₹9,000)
   - Count: Every
   - Click-through conversion window: 30 days
   - View-through conversion window: 7 days
8. Click **Done**
9. Note the **Conversion ID** (AW-XXXXXXXXX) and **Conversion Label**
10. Go back to GTM and update the 3 conversion tags:
    - Replace `AW-XXXXXXXXX` with your actual Conversion ID in each tag's HTML
11. Publish GTM again

---

## 3. GOOGLE ADS KEYWORD FIX — MOST IMPORTANT (5 minutes)

**Problem**: All 85 existing keywords are long exact-match phrases like `[furnished pg kapurthala lpu students]` showing "Not eligible - Low search volume". Nobody searches these exact phrases.

**Solution**: Replace with short, high-volume exact-match keywords.

### Step 3a: Remove ALL 85 existing keywords

1. Go to: https://ads.google.com/aw/keywords?ocid=8306148154
2. Click the **checkbox in the header row** to select all keywords on the page
3. Click **More options** (3 dots) → **Remove**
4. Go to next page (there are 9 pages of 10) and repeat until ALL 85 are removed
5. Alternatively: select all → change status to **Paused** (safer, keeps data)

### Step 3b: Add new short exact-match keywords

1. On the Keywords page, click **"+ Create keywords"** (blue button)
2. Select **V1-GMB** campaign → **Ad group 1** (or create new ad groups if desired)
3. Paste these keywords ONE GROUP AT A TIME:

**PG Near LPU (10 keywords):**
```
[PG near LPU]
[PG LPU]
[LPU PG]
[PG near LPU university]
[student PG LPU]
[boys PG near LPU]
[PG Jalandhar]
[PG Phagwara]
[budget PG LPU]
[best PG near LPU]
```

**Room Near LPU (10 keywords):**
```
[room near LPU]
[room LPU]
[LPU room rent]
[room rent near LPU]
[single room LPU]
[furnished room LPU]
[room for rent Jalandhar]
[1 BHK near LPU]
[flat near LPU]
[room near university Jalandhar]
```

**Hostel Near LPU (5 keywords):**
```
[hostel near LPU]
[LPU hostel]
[hostel near LPU university]
[LPU hostel alternative]
[student hostel LPU]
```

**LPU Accommodation (8 keywords):**
```
[LPU accommodation]
[LPU student housing]
[LPU off campus]
[LPU admission room]
[where to stay near LPU]
[student accommodation LPU]
[LPU room booking]
[Lovely Professional University PG]
```

**Jalandhar Broad (5 keywords):**
```
[room rent Jalandhar]
[paying guest Jalandhar]
[flat rent Jalandhar]
[student room Jalandhar]
[hostel Jalandhar]
```

4. Set match type to **Exact** for all
5. Click **Save**

### Step 3c: Add Negative Keywords

1. Go to Keywords tab → **Negative keywords** sub-tab
2. Click **"+ Add negative keywords"**
3. Paste from `google-ads-negative-keywords.csv` or enter manually:
```
free, cheapest, very cheap, below 3000, below 5000, under 3000, under 5000, hourly, daily rent, one day, short stay, one night, commercial, office space, shop, buy, purchase, for sale, plot, villa, hotel, resort, guest house, lodge, dharamshala, married couple, family flat, pet friendly, old age home, retirement, couple friendly, images, photos, reviews
```

---

## 4. VERIFICATION CHECKLIST

After all setup is complete:

- [x] Website loads at https://vatsapartment.com (confirmed HTTP 200)
- [x] GTM script (`GTM-P6TDL438`) loads in page source
- [x] WhatsApp floating button with tracking code is live
- [x] Phone call click tracking is live
- [ ] GTM container published with 3 tags + 3 triggers + 4 variables
- [ ] Open browser DevTools > Console on website, click WhatsApp button, see dataLayer push
- [ ] Google Ads campaign V1-GMB has 38 short exact-match keywords
- [ ] Google Ads campaign has 75+ negative keywords
- [ ] Google Ads conversion action shows as "Recording conversions"
- [ ] All old 85 low-volume keywords removed/paused

---

## FILES PREPARED

| File | Purpose | Count |
|---|---|---|
| `gtm-container-import.json` | Import into GTM Admin > Import Container | 3 tags, 3 triggers, 4 variables |
| `google-ads-keywords-exact.csv` | NEW: Short exact-match keywords for bulk upload | 38 keywords |
| `google-ads-keywords.csv` | OLD: Phrase/broad match keywords (backup) | 50 keywords |
| `google-ads-negative-keywords.csv` | Negative keywords for bulk upload | 75 negative keywords |
| `GOOGLE-ADS-STRATEGY.md` | Complete strategy document | Full 12-section plan |
| `GOOGLE-ADS-STAGED-PLAN.md` | Staged action plan | Stages 0-11 |

## CODE CHANGES (already committed and pushed)

- All Go templates in `templates/*.html` — GTM + floating WhatsApp + click tracking
