---
name: google-ads
description: Manage Google Ads campaigns, keywords, negative keywords, conversion actions, and budgets via the Google Ads REST API (v20). Use when adding/removing keywords, managing campaigns, creating conversion actions, uploading bulk changes, or querying ad performance for Vats Apartment or any Google Ads account.
---

# Google Ads API Skill

## Prerequisites

This skill requires 4 credentials stored in environment variables or a `.env` file:

```
GOOGLE_ADS_CLIENT_ID=xxxx.apps.googleusercontent.com
GOOGLE_ADS_CLIENT_SECRET=GOCSPX-xxxx
GOOGLE_ADS_REFRESH_TOKEN=1//xxxx
GOOGLE_ADS_DEVELOPER_TOKEN=xxxx
GOOGLE_ADS_CUSTOMER_ID=8306148154
```

### How to obtain credentials

1. **Google Cloud Project + OAuth**: https://console.cloud.google.com
   - Create project → Enable "Google Ads API" → Create OAuth Desktop credentials → copy Client ID + Secret
2. **Refresh Token**: https://developers.google.com/oauthplayground
   - Gear icon → "Use your own OAuth credentials" → paste Client ID/Secret
   - Scope: `https://www.googleapis.com/auth/adwords` → Authorize → Exchange → copy Refresh Token
3. **Developer Token**: https://ads.google.com/aw/apicenter
   - Tools → API Center → Apply for Basic Access (instant for own-account use)
4. **Customer ID**: Found in Google Ads URL (`ocid=XXXXXXXXXX`) or top-right corner

## Authentication

Get a fresh access token before any API call:

```bash
ACCESS_TOKEN=$(curl -s -X POST https://oauth2.googleapis.com/token \
  -d client_id="$GOOGLE_ADS_CLIENT_ID" \
  -d client_secret="$GOOGLE_ADS_CLIENT_SECRET" \
  -d refresh_token="$GOOGLE_ADS_REFRESH_TOKEN" \
  -d grant_type=refresh_token | python3 -c "import sys,json; print(json.load(sys.stdin)['access_token'])")
```

## API Base URL

```
https://googleads.googleapis.com/v20/customers/{customer_id}/
```

Headers for every request:
```
Authorization: Bearer {ACCESS_TOKEN}
developer-token: {DEVELOPER_TOKEN}
login-customer-id: {CUSTOMER_ID}
Content-Type: application/json
```

## Common Operations

### List all campaigns

```bash
curl -s -X POST "https://googleads.googleapis.com/v20/customers/{CID}/googleAds:searchStream" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "developer-token: $DEV_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"query": "SELECT campaign.id, campaign.name, campaign.status, campaign.advertising_channel_type, campaign_budget.amount_micros FROM campaign"}'
```

### List keywords in a campaign

```bash
QUERY="SELECT ad_group_criterion.keyword.text, ad_group_criterion.keyword.match_type, ad_group_criterion.status, ad_group.name, campaign.name FROM keyword_view WHERE campaign.status = 'ENABLED'"
```

### Add keywords to an ad group

```bash
curl -s -X POST "https://googleads.googleapis.com/v20/customers/{CID}/adGroupCriteria:mutate" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "developer-token: $DEV_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "operations": [{
      "create": {
        "adGroup": "customers/{CID}/adGroups/{AD_GROUP_ID}",
        "keyword": {
          "text": "PG near LPU",
          "matchType": "EXACT"
        }
      }
    }]
  }'
```

### Remove a keyword

```bash
curl -s -X POST "https://googleads.googleapis.com/v20/customers/{CID}/adGroupCriteria:mutate" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "developer-token: $DEV_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "operations": [{
      "remove": "customers/{CID}/adGroupCriteria/{CRITERION_ID}"
    }]
  }'
```

### Add negative keywords (campaign-level)

```bash
curl -s -X POST "https://googleads.googleapis.com/v20/customers/{CID}/campaignCriteria:mutate" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "developer-token: $DEV_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "operations": [{
      "create": {
        "campaign": "customers/{CID}/campaigns/{CAMPAIGN_ID}",
        "negative": true,
        "keyword": {
          "text": "free",
          "matchType": "EXACT"
        }
      }
    }]
  }'
```

### Update campaign budget

```bash
curl -s -X POST "https://googleads.googleapis.com/v20/customers/{CID}/campaignBudgets:mutate" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "developer-token: $DEV_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "operations": [{
      "update": {
        "resourceName": "customers/{CID}/campaignBudgets/{BUDGET_ID}",
        "amountMicros": "500000"
      },
      "updateMask": "amountMicros"
    }]
  }'
```

### Create conversion action

```bash
curl -s -X POST "https://googleads.googleapis.com/v20/customers/{CID}/conversionActions:mutate" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "developer-token: $DEV_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "operations": [{
      "create": {
        "name": "WhatsApp Inquiry",
        "category": "CONTACT",
        "type": "WEBPAGE",
        "status": "ENABLED",
        "valueSettings": {
          "value": 9000,
          "alwaysUseDefaultValue": true
        },
        "countingType": "EVERY_CONVERSION",
        "clickThroughLookbackWindowDays": "30",
        "viewThroughLookbackWindowDays": "7"
      }
    }]
  }'
```

## Bulk Operations Pattern

For adding/removing many keywords at once, batch up to 1000 operations in a single mutate call:

```bash
# Build operations array from CSV
python3 -c "
import csv, json
ops = []
with open('google-ads-keywords-exact.csv') as f:
    reader = csv.DictReader(f)
    for row in reader:
        kw = row['Keyword'].strip('[]')
        ops.append({
            'create': {
                'adGroup': 'customers/{CID}/adGroups/{AD_GROUP_ID}',
                'keyword': {'text': kw, 'matchType': 'EXACT'}
            }
        })
print(json.dumps({'operations': ops}))
" > /tmp/keywords_payload.json

curl -s -X POST "https://googleads.googleapis.com/v20/customers/{CID}/adGroupCriteria:mutate" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "developer-token: $DEV_TOKEN" \
  -H "Content-Type: application/json" \
  -d @/tmp/keywords_payload.json
```

## GAQL Query Reference

Common queries for reporting:

```sql
-- Campaign performance
SELECT campaign.name, metrics.impressions, metrics.clicks, metrics.cost_micros, metrics.conversions FROM campaign WHERE segments.date DURING LAST_7_DAYS

-- Keyword performance
SELECT ad_group_criterion.keyword.text, metrics.impressions, metrics.clicks, metrics.cost_micros FROM keyword_view WHERE campaign.status = 'ENABLED'

-- Search term report
SELECT search_term_view.search_term, metrics.impressions, metrics.clicks, metrics.cost_micros FROM search_term_view WHERE campaign.status = 'ENABLED'
```

## Important Notes

- Amounts are in **micros**: ₹500 = `500000000` micros (multiply by 1,000,000)
- Customer ID in URLs: no hyphens (use `8306148154` not `830-614-8154`)
- Resource names follow format: `customers/{CID}/campaigns/{ID}`
- Max 1000 operations per mutate call
- Partial failure mode: set `partialFailure=true` query param to continue on individual errors
- Always verify with a GAQL query after mutations
