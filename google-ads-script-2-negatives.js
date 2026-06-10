// ============================================
// SCRIPT 2: ADD NEGATIVE KEYWORDS
// Run after Script 1 completes
// ============================================

function main() {
  var campaignName = "V1-GMB";

  var negativeKeywords = [
    // Price filters
    "free", "free room", "without rent", "zero rent", "zero deposit",
    "cheapest", "very cheap", "below 3000", "below 5000",
    "under 3000", "under 5000",
    // Short stay
    "hourly", "daily rent", "one day", "short stay", "one night",
    // Wrong property types
    "commercial", "office space", "shop", "showroom", "godown", "warehouse",
    "buy", "purchase", "for sale", "plot",
    // Jobs
    "hostel job", "hostel warden", "hostel manager",
    "staff", "vacancy", "recruitment",
    // Wrong property
    "flat for sale", "apartment for sale", "villa",
    "independent house", "kothi", "farm house",
    "agricultural land",
    // Venues
    "marriage hall", "banquet", "event space",
    // Hotels (not PG)
    "hotel", "resort", "guest house", "lodge", "dharamshala",
    // Demographics
    "married couple", "family flat", "pet friendly", "dog", "cat",
    "senior citizen", "old age home", "retirement",
    "couple friendly", "unmarried couple",
    // Wrong cities
    "LPU Mumbai", "LPU Bangalore", "LPU Chennai", "LPU Hyderabad",
    // Info-only
    "images", "photos", "pictures", "reviews",
    "contact number", "phone number",
    "address", "directions", "map", "location"
  ];

  var campaign = AdsApp.campaigns()
    .withCondition("Name = '" + campaignName + "'")
    .get()
    .next();

  // Remove existing campaign-level negative keywords first
  var existingNegKwIterator = AdsApp.negativeKeywords()
    .withCondition("CampaignName = '" + campaignName + "'")
    .get();

  var removedCount = 0;
  while (existingNegKwIterator.hasNext()) {
    var existingNegKw = existingNegKwIterator.next();
    existingNegKw.remove();
    removedCount++;
  }
  Logger.log("Removed " + removedCount + " old negative keywords");

  // Add new negative keywords
  var addedCount = 0;
  for (var i = 0; i < negativeKeywords.length; i++) {
    campaign.addNegativeKeyword("[" + negativeKeywords[i] + "]");
    addedCount++;
  }

  Logger.log("=== DONE ===");
  Logger.log("Added " + addedCount + " negative keywords (exact match)");
}
