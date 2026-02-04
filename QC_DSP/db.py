import boto3
import json
from datetime import datetime as dt
import random
import string
from decimal import Decimal

bucket = 'sagemaker-us-east-2-344400919253'

from boto3.dynamodb.conditions import Key


class DB:
    def __init__(self, table_name, partition_key, sort_key):
        dynamodb = boto3.resource('dynamodb', region_name='us-east-2')
        self.db = dynamodb.Table(table_name)
        self.partition_key = partition_key
        self.sort_key = sort_key

    def get(self, partition_key, sort_key=None):
        if sort_key:
            response = self.db.query(
                    KeyConditionExpression=Key(self.partition_key).eq(partition_key) &
                    Key(self.sort_key).eq(sort_key)
                    )
        else:
            response = self.db.query(
                    KeyConditionExpression=Key(self.partition_key).eq(partition_key) 
            )
        return response.get('Items')


class campaigns(DB):
    def __init__(self):
        super().__init__('campaigns', 'campaign_id', 'template_id')
        
    def update(self, campaign_id, template_id, active, segments):
        
        # should check to see that the template_id is available for this campaign
        # this is listed in the customers db
        
        X = {'timestamp':  str(dt.now()),
             'campaign_id': campaign_id,
             'template_id': template_id,
             'active': active,
             'segments': segments}

        return self.db.put_item(Item = X)

    def set(self, status, campaign_id, template_id=None):
        all_campaign_templates = self.get(campaign_id, template_id)
        for ct in all_campaign_templates:
            ct['active'] = status
            ct['timestamp'] = str(dt.now())
            self.db.put_item(Item = ct)


class customers(DB):
    def __init__(self):
        super().__init__('customers', 'customer_id', 'campaign_id')


    def update(self, customer_id, campaign_id, template_ids):

        X = {'timestamp':  str(dt.now()),
             'campaign_id': campaign_id,
             'customer_id': customer_id,
             'template_ids': template_ids}

        return self.db.put_item(Item = X)


class templates(DB):
    def __init__(self):
        super().__init__('template_mapping', 'template_id', 'component_id')
        
    def update(self, template_id, component_id, possible_values):

        X = {'timestamp':  str(dt.now()),
             'template_id': template_id,
             'component_id': component_id,
             'possible_values': possible_values}

        return self.db.put_item(Item = X)


class clicks(DB):
    def __init__(self):
        super().__init__('clicks', 'template_id', 'timestamp')

    def update(self, campaign_id, template_id, ad_id, segment):

        X = {'campaign_id': campaign_id, 
             'template_id': template_id,
             'ad_id': ad_id,
             'segment': segment,
            'timestamp':  str(dt.now())}

        return self.db.put_item(Item = X)





