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

## 3. GOOGLE ADS CAMPAIGN FIXES (5 minutes)

### 3a. Change Budget (₹500/day is fine for now — ₹15K/month matches plan)

1. Go to: https://ads.google.com/aw/campaigns?ocid=8306148154&authuser=0
2. Click on **V1-GMB** campaign name
3. Click **Settings** tab
4. Change Budget to **₹200/day** (₹6,000/month — safer for testing)
5. Click **Save**

### 3b. Add Keywords (Bulk Upload)

1. Click on V1-GMB campaign → **Ad Groups** → select the ad group
2. Go to **Keywords** section
3. Click **"+" → "Add keywords"**
4. Or use **Tools > Bulk edits > Upload file**
5. Upload: `google-ads-keywords.csv`
6. Remove the 5 existing exact-match keywords (too narrow, low volume)

### 3c. Add Negative Keywords

1. In campaign, go to **Audiences, keywords, and content** → **Negative keywords**
2. Click **"+" → "Add negative keywords"**
3. Or upload: `google-ads-negative-keywords.csv`

---

## 4. FIX WEBSITE DEPLOYMENT

The site returns 530 (origin unreachable via Cloudflare). Check:
1. Dokploy dashboard — is the container running?
2. DNS settings — does `www.vatsapartment.com` point to the Dokploy server IP?
3. Rebuild and redeploy from Dokploy dashboard (latest code is pushed to `main`)

---

## 5. VERIFICATION CHECKLIST

After all setup is complete:

- [ ] Website loads at https://www.vatsapartment.com
- [ ] GTM container published with 3 tags + 3 triggers + 4 variables
- [ ] Open browser DevTools > Console on website, click WhatsApp button, see `[GTM] WhatsApp conversion tracked`
- [ ] Open browser DevTools > Network, click WhatsApp button, see request to `googleads.g.doubleclick.net`
- [ ] Google Ads campaign V1-GMB shows 50+ keywords
- [ ] Google Ads campaign has 70+ negative keywords
- [ ] Google Ads conversion action shows as "Recording conversions"
- [ ] Google Ads tag status shows "Active" (not "Inactive")

---

## FILES PREPARED

| File | Purpose | Lines |
|---|---|---|
| `gtm-container-import.json` | Import into GTM Admin > Import Container | 3 tags, 3 triggers, 4 variables |
| `google-ads-keywords.csv` | Bulk upload keywords to Google Ads | 50 keywords |
| `google-ads-negative-keywords.csv` | Bulk upload negative keywords | 70 negative keywords |
| `GOOGLE-ADS-STRATEGY.md` | Complete strategy document | Full 12-section plan |

## CODE CHANGES (already committed and pushed)

- `app/layout.tsx` — GTM with dataLayer init + GTM_ID constant
- `components/FooterWithWhatsApp.tsx` — WhatsApp + phone click tracking
- `app/page.tsx` — Check Availability + Call Now tracking
- `app/book-now/page.tsx` — Booking form submission tracking
- `types/global.d.ts` — Window.dataLayer TypeScript type
