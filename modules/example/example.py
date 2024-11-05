# example.py
###################################

# for localization
from flask_babel import Babel, _   

# from flask only what we need for module
from flask import Flask, Response, render_template

# from openpanel app only what we need for templates
from app import login_required_route, query_email_by_id, log_user_action, query_username_by_id, get_user_services_and_domains, gravatar_url, avatar_type

# from available python modules
import os
import string
import requests
import random



# example route
@app.route('/settings/something', methods=['GET', 'POST'])
@login_required_route
def account_notifications():
    user_id = session['user_id']
    title=_('My OpenPanel Module')                               # used for page titles in template
    current_route = "/settings/notifications"                    # used for page breadcrumbs in template
    services, domains = get_user_services_and_domains(user_id)   # only if we need user domains on this route
    current_username = query_username_by_id(user_id)             # for template
    current_email = query_email_by_id(user_id)                   # for template
    gravatar_image_url = gravatar_url(current_email)             # for template
 
    return render_template('my_template.html', title=title, username=current_username, email=current_email, gravatar_image_url=gravatar_image_url, avatar_type=avatar_type, domains=domains, current_route=current_route, current_username=current_username)
