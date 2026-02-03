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


app = Flask(__name__, static_folder='data')

@app.route('/quante_carlo')
def quante_carlo():
    return "<html>Quante Carlo</html>"

def get_component_values(campaign_id, template_id, segment_id):
    return {'component_values': {'headline': 'Wool Socks',
                                 'description': 'Describe This!',
                                 'cta': 'buy now', 'price': 'Free',
                                 'image': 'https://www.nasa.gov/wp-content/uploads/2026/02/moon.jpg?resize=300,200'},
            'ad_id': '100-100-100-50'}

@app.route('/favicon.ico')
def favicon():
    return 'smiley face'


@app.route("/test")
def test():
    with open('ad_library/test/test/test/100.template') as f:
        return f.read()




@app.route("/ad_server")
def ad_server():

    campaign = request.args.get('campaign')
    template = request.args.get('template')
    segment = request.args.get('segment', None)
    
    component_values = get_component_values(campaign, template, segment)
    print(component_values['component_values']['image'])
    return component_values



@app.route("/click_counter")
def click_counter():

    ad_id = request.args.get('ad_id')
    print(ad_id)
    return redirect('https://quantecarlo.com')



if __name__ == '__main__':
    app.run()

