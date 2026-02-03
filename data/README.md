
## Primary keys and Tables
- campaign_id
- template_id
- user_id - 1 per customer
- component_id
- creative_asset_id
- component_value_id
- rtb_id - used for tracking bids, amounts, wins, losses
- ad_id - each template with have a number of components; the ad_id is created by concatenating creative_asset_id's in the order of the components appear in the template seperated by dashss
- segment_id - in this version, there are only segments; segments can  be granular and segment/RTB and segment/creative combinations are decided offline


## API's
### RTB
has parameters set for each campaign, targets, filters, guardrails, etc
the parameters include a user, campaign, segment, template combination


`/ad_server?user_id=<user_id>&campaign_id=<campaign_id>&segment_id=<segment_id>&template_id=<template_id>&credentials=<credentials>`
After auction is won, RTB provides this url to SSP

`/decision_api?user_id=<user_id>&campaign_id=<campaign_id>&segment_id=<segment_id>&template_id=<template_id>&credentials=<credentials>`
This is embedded in the script tag in the ad template


## template library
Right now, just a directory of files
template_library/user_id/campaign_id/template_id.template 









