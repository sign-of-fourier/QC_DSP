import boto3
import json
from datetime import datetime as dt
import random
import string
from decimal import Decimal

bucket = 'sagemaker-us-east-2-344400919253'


from boto3.dynamodb.conditions import Key



class click_db(db):
    def __init__(self):
        self.db = dynamodb.Table('dynamic_ad_click_counts')
        self.keys = ['ad_id', 'click_timestamp']

    def get_clicks(self, ad_id):
        response = self.db.query(
                KeyConditionExpression=Key('ad_id').eq(ad_id)
                )
        return response.get('Items')

    def update(self, template_id, ad_id):

        X = {'click_timestamp':  str(dt.now()),
             'ad_id': ad_id}

        return self.db.put_item(Item = X)






class active_ad_db:
    def __init__(self):
        dynamodb = boto3.resource('dynamodb', region_name='us-east-2')
        self.db = dynamodb.Table('active_campaigns')
        self.keys = ['template_id', 'ad_id', 'active']

    def is_active(self, template_id, ad_id):
        response = self.db.query(
                KeyConditionExpression=Key('ad_id').eq(ad_id) &
                KeyConditionExpression=('template_id')
                )
        return response.get('Items')

    def update(self, template_id, ad_id, active):

        X = {'timestamp':  str(dt.now()),
             'template_id': template_id,
             'ad_id': ad_id,
             'active': active}

        return self.db.put_item(Item = X)






class template_db:
    def __init__(self):
        dynamodb = boto3.resource('dynamodb', region_name='us-east-2')
        self.db = dynamodb.Table('template_mapping')

        # values is a dictionary of unique value_ids to values such as text or urls
        self.keys = ['template_id', 'component_id', 'values']

    def get_ads(self, template_id, component_id=None):
        if component_id:
            response = self.db.query(
                KeyConditionExpression=Key('template_id').eq(template_id) &
                Key('component_id').eq(component_id)
                )
        else:
            response = self.db.query(
                KeyConditionExpression=Key('template_id').eq(template_id) 
                )

        return response.get('Items')

    def update(self, template_id, component_id, possible_values):

        X = {'timestamp':  str(dt.now()),
             'templet_id': template_id,
             'component_id': component_id,
             'possible_values': possible_values}

        return self.db.put_item(Item = X)








