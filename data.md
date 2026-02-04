
## Primary keys and Tables
- campaign_id
- template_id
- customer_id - 1 per customer
- components
- creative_asset_id
- value_id
- rtb_id - used for tracking bids, amounts, wins, losses
- ad_id - each template with have a number of components; the ad_id is created by concatenating creative_asset_id's in the order of the components appear in the template seperated by dashss
- segment_id - in this version, there are only segments; segments can  be granular and segment/RTB and segment/creative combinations are decided offline

## API's
**1** RTB
has parameters set for each campaign, targets, filters, guardrails, etc
the parameters include a user, campaign, segment, template combination

**2** Lookup and serve template
`/ad_server/template?customer_id=<user_id>&campaign_id=<campaign_id>&template_id=<template_id>&credentials=<credentials>`
- After auction is won, RTB provides this url to SSP.
- java script is called to determine which creative values are to be populated

**2** javascript decision api fetch
`/ad_server/decision_api?campaign=<campaign_id>&segment_id=<segment_id>&template_id=<template_id>&credentials=<credentials>`
- This is embedded in the script tag in the ad template

**3** Record click and redirect to landing page
`/ad_server/ad_server/click_counter?campaign_id=<campaign_id>&segment_id=<segment_id>&template=<template_id>&ad_id&credentials=<credentials>`
- the ad_id encodes the values and is included in the click url

# Dynamo DBs
- customers - associtated with campaigns
- campaigns - associated with templates and batch of available ad confirations (segment -> list of ad_id's)
- templates - associated with compents; for each component (creative_id -> creative text/url/css value)
- clicks - template_id, timestamp associated with campaign_id, ad_id and segment
- impressions - same as clicks
  
## Template Library
Right now, just a directory of files
template_library/customer_id/campaign_id/template_id.template 

## Ingestion
Every possible combination is cycled through and an embedding is created. Embeddings are stored in s3.

## Bayesian Optimization
1. Asyncrhonously checks if impressions have reached a defined threshold
2. Once threshold is reached, triggers retraining
3. CTR and embeddings are used to fit a Gaussian Process Regressor
4. qEI is calculated for large number of batches
5. Best batch is used to update batches 








