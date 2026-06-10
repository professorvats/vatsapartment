# Vats Apartment — Google Ads Setup: Stage-by-Stage Action Plan

## Instructions
- Mark each task `[x]` when done
- Follow stages in ORDER — each stage depends on the previous
- Notes column: add date completed and any issues

---

## STAGE 0: Browser Connection Setup
> **Goal:** Get Brave browser connected via MCP so we can access Google Ads together

- [ ] **0.1** Launch Brave with remote debugging enabled
- [ ] **0.2** Verify browser MCP connection working
- [ ] **0.3** Navigate to Google Ads and verify login

**Status:** ⏳ Pending
**Notes:**

---

## STAGE 1: Google Ads Account Foundation
> **Goal:** Get account ready for campaigns — billing, settings, linked accounts

- [ ] **1.1** Log into Google Ads at ads.google.com
- [ ] **1.2** Verify/set up billing (credit card or UPI)
- [ ] **1.3** Set account timezone to Asia/Kolkata (IST)
- [ ] **1.4** Set currency to INR (₹)
- [ ] **1.5** Link Google Business Profile to Google Ads
- [ ] **1.6** Enable location extensions (from GMB link)
- [ ] **1.7** Set up auto-tagging (Account Settings → Preferences → Auto-tagging = ON)

**Status:** ⏳ Pending
**Notes:**

---

## STAGE 2: Conversion Tracking Setup
> **Goal:** Track every call, WhatsApp click, and form submission — THE most critical step

- [ ] **2.1** Create Google Tag Manager (GTM) container
- [ ] **2.2** Install GTM code in website (templates/*.html <head> and <body>)
- [ ] **2.3** Create GA4 property and connect via GTM
- [ ] **2.4** In Google Ads → Tools → Conversions → Create:
  - [ ] **2.4a** Phone Call conversion (calls to +919992937447, count >60s)
  - [ ] **2.4b** WhatsApp Click conversion (custom event via GTM)
  - [ ] **2.4c** Form Submission conversion (book-now page)
- [ ] **2.5** Set up GTM trigger for WhatsApp link clicks (wa.me/*)
- [ ] **2.6** Set up GTM trigger for tel: link clicks
- [ ] **2.7** Test all conversions with GTM Preview mode
- [ ] **2.8** Publish GTM container

**Status:** ⏳ Pending
**Notes:**

---

## STAGE 3: GMB Profile Optimization
> **Goal:** Maximize local visibility and get location extensions showing in ads

- [ ] **3.1** Update GMB business name to: "Vats Apartment — Premium Rooms Near LPU"
- [ ] **3.2** Verify categories: Primary: "Paying Guest Accommodation"
- [ ] **3.3** Add secondary categories: "Apartment Building", "Serviced Accommodation"
- [ ] **3.4** Write optimized description (750 chars) with keywords
- [ ] **3.5** Upload 25+ photos (rooms, kitchen, bathroom, exterior, building, nearby)
- [ ] **3.6** Add all amenities in GMB attributes
- [ ] **3.7** Pre-populate 10 Q&A pairs
- [ ] **3.8** Create first GMB Offer Post ("Room Near LPU — Starting ₹9,000/month")
- [ ] **3.9** Create first GMB Update Post (availability update)
- [ ] **3.10** Set business hours to "Open 24/7 for inquiries"
- [ ] **3.11** Request Google reviews from 5 current tenants
- [ ] **3.12** Verify GMB shows in Google Maps for "PG near LPU"

**Status:** ⏳ Pending
**Notes:**

---

## STAGE 4: Campaign 1 — Search Ads (Tier A: Punjab, Haryana, Delhi, HP, Rajasthan)
> **Goal:** Primary revenue campaign targeting nearby states — 60% of budget

- [ ] **4.1** Create new campaign → Search → Website traffic or Conversions
- [ ] **4.2** Campaign name: "Search — Tier A (Punjab-Haryana-Delhi-HP-Raj)"
- [ ] **4.3** Set bidding: Maximize Clicks (initial phase)
- [ ] **4.4** Set daily budget: ₹200/day
- [ ] **4.5** Set target locations:
  - [ ] Punjab (state)
  - [ ] Haryana (state)
  - [ ] Delhi (NCT)
  - [ ] Himachal Pradesh (state)
  - [ ] Rajasthan (state)
- [ ] **4.6** Set audience: Age 17-55, All genders
- [ ] **4.7** Create Ad Group 1A "Core Room" with keywords:
  - [ ] [room near LPU], "room near LPU"
  - [ ] [rooms for rent near LPU], "rooms for rent near LPU"
  - [ ] [1 BHK near LPU], "1 BHK near LPU"
  - [ ] [room rent near LPU], "room rent near LPU"
  - [ ] [flat near LPU], "flat near LPU"
- [ ] **4.8** Create Ad Group 1B "Core PG" with keywords:
  - [ ] [PG near LPU], "PG near LPU"
  - [ ] [PG near LPU University], "PG near LPU University"
  - [ ] [PG in Jalandhar near LPU], "PG in Jalandhar near LPU"
  - [ ] [student PG near LPU], "student PG near LPU"
  - [ ] [boys PG near LPU], "boys PG near LPU"
  - [ ] [girls PG near LPU], "girls PG near LPU"
- [ ] **4.9** Create Ad Group 1C "Price-Focused" with keywords:
  - [ ] [budget PG near LPU], "budget PG near LPU"
  - [ ] [room near LPU under 10000], "room near LPU under 10000"
  - [ ] [cheap PG near LPU for students]
  - [ ] [rooms under 10000 near LPU]
  - [ ] [affordable room near LPU]
- [ ] **4.10** Create Ad Group 1D "Location-Specific" with keywords:
  - [ ] [PG accommodation Phagwara]
  - [ ] [PG near LPU Lawgate]
  - [ ] [room near Apna Chai Wala LPU]
  - [ ] [PG near LPU Phagwara road]
  - [ ] [furnished room near LPU Jalandhar]
- [ ] **4.11** Write Responsive Search Ads for each ad group (see GOOGLE-ADS-STRATEGY.md Section 9)
- [ ] **4.12** Add ALL ad extensions:
  - [ ] Sitelinks (6: Room Prices, See Rooms, 3D Tour, Location, Reviews, Contact)
  - [ ] Callouts (10: Fully Furnished, No Brokerage, 10 Mins to LPU, etc.)
  - [ ] Structured Snippets (Amenities, Room Types)
  - [ ] Call Extension (+91 99929 37447, with tracking)
  - [ ] Location Extension (from GMB link)
  - [ ] Price Extension (₹9,000 / ₹10,000 / ₹15,000)
- [ ] **4.13** Add ALL negative keywords (see GOOGLE-ADS-STRATEGY.md Appendix A)
- [ ] **4.14** Set ad schedule bid adjustments (6PM-10PM +25%, weekends +10-15%)
- [ ] **4.15** Set device adjustments (Mobile +20%, Desktop -10%)
- [ ] **4.16** Review campaign settings → Launch/Pause (we'll launch together)

**Status:** ⏳ Pending
**Notes:**

---

## STAGE 5: Campaign 2 — Search Ads (Tier B: UP, Bihar, Jharkhand, J&K, MP)
> **Goal:** Pan-India reach targeting top student source states — 30% of budget

- [ ] **5.1** Create new campaign → Search
- [ ] **5.2** Campaign name: "Search — Tier B (UP-Bihar-Jharkhand-JK-MP)"
- [ ] **5.3** Set bidding: Maximize Clicks
- [ ] **5.4** Set daily budget: ₹80/day
- [ ] **5.5** Set target locations:
  - [ ] Uttar Pradesh (state)
  - [ ] Bihar (state)
  - [ ] Jharkhand (state)
  - [ ] Jammu & Kashmir (UT)
  - [ ] Madhya Pradesh (state)
- [ ] **5.6** Create Ad Group 2A "Core Keywords — Tier B" (same keywords as AG1A+1B)
- [ ] **5.7** Create Ad Group 2B "Parent-Focused" with keywords:
  - [ ] [safe student housing near LPU]
  - [ ] [LPU accommodation]
  - [ ] [where to stay near LPU]
  - [ ] [LPU Jalandhar room rent]
  - [ ] [Lovely Professional University PG]
  - [ ] [LPU off campus housing]
- [ ] **5.8** Write Parent-focused ad copy (safety, security, trust messaging)
- [ ] **5.9** Add same ad extensions as Campaign 1
- [ ] **5.10** Add same negative keywords
- [ ] **5.11** Review and save (pause until Stage 4 is running)

**Status:** ⏳ Pending
**Notes:**

---

## STAGE 6: Campaign 3 — Search Ads (Tier C: Rest of India)
> **Goal:** Presence maintenance across remaining states — 10% of budget

- [ ] **6.1** Create campaign → Search
- [ ] **6.2** Name: "Search — Tier C (Rest of India)"
- [ ] **6.3** Budget: ₹20/day
- [ ] **6.4** Locations: All India EXCEPT Punjab, Haryana, Delhi, HP, Rajasthan, UP, Bihar, Jharkhand, J&K, MP
- [ ] **6.5** Keywords: Broad match versions of top keywords
- [ ] **6.6** Same ad copy as Tier A
- [ ] **6.7** Same extensions and negatives
- [ ] **6.8** Pause initially — launch after Tier A+B have data

**Status:** ⏳ Pending
**Notes:**

---

## STAGE 7: Campaign 4 — Call-Only (Mobile, Pan-India)
> **Goal:** Drive direct phone calls from mobile users

- [ ] **7.1** Create campaign → Search → Calls (call-only)
- [ ] **7.2** Name: "Call-Only — Pan-India Mobile"
- [ ] **7.3** Budget: ₹50/day
- [ ] **7.4** Target: All India, Mobile only
- [ ] **7.5** Keywords: Core keywords (room near LPU, PG near LPU, etc.)
- [ ] **7.6** Ad copy: Call-focused headlines
- [ ] **7.7** Phone: +91 99929 37447 with Google forwarding
- [ ] **7.8** Pause — launch after Search campaigns show data

**Status:** ⏳ Pending
**Notes:**

---

## STAGE 8: Website Optimization for Conversions
> **Goal:** Improve landing pages so ads convert better

- [ ] **8.1** Create dedicated landing page /ads/room-near-lpu (headless, no nav distractions)
- [ ] **8.2** Add urgency element: "Only [X] rooms available" banner
- [ ] **8.3** Add real room photos (replace placeholders on /rooms page)
- [ ] **8.4** Add WhatsApp floating button with tracking
- [ ] **8.5** Add phone call button with tracking
- [ ] **8.6** Add social proof section (reviews, student count, rating)
- [ ] **8.7** Add FAQ section for common objections
- [ ] **8.8** Mobile speed optimization (compress images, lazy load)
- [ ] **8.9** Add "Book via WhatsApp" as primary CTA on landing page
- [ ] **8.10** Test full user journey: Ad click → Landing page → WhatsApp → Booking

**Status:** ⏳ Pending
**Notes:**

---

## STAGE 9: Launch & First Week Monitoring
> **Goal:** Go live with Campaign 1 (Tier A), monitor daily, optimize

- [ ] **9.1** UNPAUSE Campaign 1 (Tier A) — LAUNCH!
- [ ] **9.2** Day 1: Check impressions, clicks, CPC, search terms
- [ ] **9.3** Day 2: Review search terms report → add negatives
- [ ] **9.4** Day 3: Check CTR by ad group → pause low performers
- [ ] **9.5** Day 4: Review which ads are showing → check headline combos
- [ ] **9.6** Day 5: UNPAUSE Campaign 2 (Tier B) if Tier A looks good
- [ ] **9.7** Day 6: Check conversion tracking is firing
- [ ] **9.8** Day 7: Week 1 performance report — adjust bids, keywords, budgets

**Status:** ⏳ Pending
**Notes:**

---

## STAGE 10: Optimization & Scaling (Week 2-4)
> **Goal:** Improve CPA, scale winning campaigns, cut losers

- [ ] **10.1** Switch bidding to "Maximize Conversions" (if 5+ conversions recorded)
- [ ] **10.2** Pause keywords with CPC > ₹15 and no conversions
- [ ] **10.3** Increase budget on best-performing ad groups by 20%
- [ ] **10.4** Add new keywords from Search Terms report
- [ ] **10.5** Test new ad variations (try different headlines/descriptions)
- [ ] **10.6** Launch Campaign 3 (Tier C) if Tier A+B performing well
- [ ] **10.7** Launch Campaign 4 (Call-Only)
- [ ] **10.8** Set up Display Remarketing campaign
- [ ] **10.9** Get 10+ more Google reviews
- [ ] **10.10** Monthly performance review and report

**Status:** ⏳ Pending
**Notes:**

---

## STAGE 11: Ongoing Monthly Optimization
> **Goal:** Continuous improvement, seasonal adjustments, scaling

- [ ] **11.1** Weekly: Check search terms → add keywords + negatives
- [ ] **11.2** Weekly: Update GMB posts (new offer, availability update)
- [ ] **11.3** Bi-weekly: Test new ad copy variations
- [ ] **11.4** Monthly: Adjust budgets (increase peak season, decrease off-season)
- [ ] **11.5** Monthly: Add seasonal keywords
- [ ] **11.6** Monthly: Competitor analysis (check what competitors are running)
- [ ] **11.7** Monthly: Review and optimize landing pages
- [ ] **11.8** Ongoing: Request Google reviews from new tenants
- [ ] **11.9** Quarterly: Full account audit and restructure if needed
- [ ] **11.10** May-August: DOUBLE budgets for admission season

**Status:** ⏳ Pending
**Notes:**

---

## Quick Reference

| Stage | What | Time Needed | Depends On |
|---|---|---|---|
| 0 | Browser connection | 10 min | — |
| 1 | Account foundation | 30 min | Stage 0 |
| 2 | Conversion tracking | 2-3 hrs | Stage 1 |
| 3 | GMB optimization | 2-3 hrs | — (parallel with Stage 1-2) |
| 4 | Campaign 1 (Tier A) | 2-3 hrs | Stage 1, 2 |
| 5 | Campaign 2 (Tier B) | 1-2 hrs | Stage 4 |
| 6 | Campaign 3 (Tier C) | 1 hr | Stage 5 |
| 7 | Call-Only campaign | 30 min | Stage 4 |
| 8 | Website optimization | 3-4 hrs | Stage 2 (can start in parallel) |
| 9 | Launch + monitoring | 7 days | Stage 4, 8 |
| 10 | Optimize + scale | 3-4 weeks | Stage 9 |
| 11 | Ongoing | Monthly | Stage 10 |

**Total setup time: ~8-12 hours over 1-2 weeks**
**First ad live: End of Stage 4 (~4-6 hours of work)**
