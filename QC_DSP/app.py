from flask import Flask, Response,send_file,redirect, send_from_directory, request, make_response
import pandas as pd
import re
import time
import base64
import httpx
from importlib import import_module
import boto3
import json
import datetime
import random
import string
import openai
import itertools
import os
#from matplotlib import pyplot as plt
import numpy as np
import requests
#from sklearn.gaussian_process import GaussianProcessRegressor
#from sklearn.gaussian_process.kernels import WhiteKernel, Matern, DotProduct
#from scipy.stats import ecdf, lognorm
#from multiprocessing import Pool
#from scipy.stats import norm
import db

app = Flask(__name__, static_folder='data')

@app.route('/quante_carlo')
def quante_carlo():
    return "<html>Quante Carlo</html>"

def get_component_values(campaign_id, template_id, segment_id):
    template = db.templates().get(template_id)
    campaign = db.campaigns().get(campaign_id, template_id)
    ad_id = random.choice(campaign[0]['segments'][segment_id])
    #print(template)
    #print(campaign)
    V = {'headline': 'Wool Socks',
         'description': 'Describe This!',
         'cta': 'buy now', 'price': 'Free',
         'image': 'https://www.nasa.gov/wp-content/uploads/2026/02/moon.jpg?resize=300,200'}

    for c, t in zip(ad_id.split('-'), template):
        #print(c, t['component_id'], c)
        V[t['component_id']] = t['possible_values'][c]
        #print(c, t['component_id'], t['possible_values'][c])
    return {'ad_id': ad_id,
            'component_values': V}

@app.route('/favicon.ico')
def favicon():
    return 'smiley face'


@app.route("/test")
def test():
    with open('ad_library/test/test/test/100.template') as f:
        return f.read()



url_stem = 'http://3.15.22.204:8000'



@app.route("/template")
def serve_dynamic_ad():

    campaign = request.args.get('campaign')
    template = request.args.get('template')
    customer = request.args.get('customer', None)
    with open(f'ad_library/{customer}/{campaign}/{template}') as f:
        return f.read().format(url_stem, url_stem)


@app.route("/ad_server/<page>")
def ad_server(page):
    
    campaign = request.args.get('campaign')
    template = request.args.get('template')
    segment = request.args.get('segment', None)
    if page == 'click_counter':
        ad_id = request.args.get('ad_id')
        print(template, campaign, segment, ad_id)
        print (db.clicks().update(campaign, template, ad_id, segment))
        return redirect('https://quantecarlo.com')
    elif page == 'dco':

        component_values = get_component_values(campaign, template, segment)
        return component_values
    else:
        return f"no match for {page}"






@app.route("/ad_inventory")
def add_inventory():
    campaign = 'test7'
    page = """<html>
<h2>ad inventory</h2>
<table border=1>
    <tr>
        <td> Campaign </td> <td> Segment </td> <td> Template </td>
    </tr>
"""
    for segment, template in zip(
            ['news_general', 'fashion_trendy'],
            ['T_X1', 'T_X2']
            ):
        page += f"""<tr>
  <td>{campaign}</td>
  <td>{segment}</td>
  <td><a href=\"{url_stem}/template?customer=test&campaign={campaign}&&template={template}&segment={segment}\">{template}</a>
  </td>
</tr>
"""
    return page + '</table>'






if __name__ == '__main__':
    app.run()

