// ============================================
// SCRIPT 1: REMOVE OLD KEYWORDS & ADD NEW ONES
// Run this in Google Ads > Tools > Scripts
// ============================================

function main() {
  var campaignName = "V1-GMB";

  // STEP 1: Pause all existing keywords in the campaign
  var keywordIterator = AdsApp.keywords()
    .withCondition("CampaignName = '" + campaignName + "'")
    .get();

  var pausedCount = 0;
  while (keywordIterator.hasNext()) {
    var keyword = keywordIterator.next();
    keyword.pause();
    pausedCount++;
  }
  Logger.log("Paused " + pausedCount + " old keywords");

  // STEP 2: Define new keywords by ad group
  var newKeywords = {
    "PG Near LPU": [
      "PG near LPU", "PG LPU", "LPU PG", "PG near LPU university",
      "student PG LPU", "boys PG near LPU", "PG Jalandhar",
      "PG Phagwara", "budget PG LPU", "best PG near LPU"
    ],
    "Room Near LPU": [
      "room near LPU", "room LPU", "LPU room rent", "room rent near LPU",
      "single room LPU", "furnished room LPU", "room for rent Jalandhar",
      "1 BHK near LPU", "flat near LPU", "room near university Jalandhar"
    ],
    "Hostel Near LPU": [
      "hostel near LPU", "LPU hostel", "hostel near LPU university",
      "LPU hostel alternative", "student hostel LPU"
    ],
    "LPU Accommodation": [
      "LPU accommodation", "LPU student housing", "LPU off campus",
      "LPU admission room", "where to stay near LPU",
      "student accommodation LPU", "LPU room booking",
      "Lovely Professional University PG"
    ],
    "Jalandhar Broad": [
      "room rent Jalandhar", "paying guest Jalandhar",
      "flat rent Jalandhar", "student room Jalandhar",
      "hostel Jalandhar"
    ]
  };

  // STEP 3: Create ad groups if they don't exist, then add keywords
  var campaign = AdsApp.campaigns()
    .withCondition("Name = '" + campaignName + "'")
    .get()
    .next();

  var addedCount = 0;
  for (var adGroupName in newKeywords) {
    var keywords = newKeywords[adGroupName];

    // Find or create ad group
    var adGroupIterator = AdsApp.adGroups()
      .withCondition("CampaignName = '" + campaignName + "'")
      .withCondition("Name = '" + adGroupName + "'")
      .get();

    var adGroup;
    if (adGroupIterator.hasNext()) {
      adGroup = adGroupIterator.next();
      Logger.log("Found existing ad group: " + adGroupName);
    } else {
      adGroup = campaign.newAdGroupBuilder()
        .withName(adGroupName)
        .build()
        .getResult();
      Logger.log("Created new ad group: " + adGroupName);
    }

    // Enable the ad group if paused
    adGroup.enable();

    // Add keywords as EXACT match
    for (var i = 0; i < keywords.length; i++) {
      var kw = keywords[i];
      adGroup.newKeywordBuilder()
        .withText("[" + kw + "]")
        .withCpc(10)
        .build();
      addedCount++;
    }
    Logger.log("Added " + keywords.length + " keywords to " + adGroupName);
  }

  Logger.log("=== DONE ===");
  Logger.log("Paused: " + pausedCount + " old keywords");
  Logger.log("Added: " + addedCount + " new exact-match keywords");
}
