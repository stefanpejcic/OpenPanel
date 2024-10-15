import os
import json # will use for get to return data
import socket
from flask import Flask, Response, abort, render_template, request, send_file, g, jsonify, session, url_for, flash, redirect, get_flashed_messages
import subprocess
from app import app, is_license_valid, login_required_route
from modules.helpers import get_all_plans, is_username_unique

@app.route('/import/cpanel', methods=['GET', 'POST'])
@login_required
def import_cpanel_whm_account():
    if request.method == 'POST':
        path = request.form.get('path')
        plan_name = request.form.get('plan_name')

        if not path or not plan_name:
            flash('Both path to the cPanel backup file (.tar.gz) and plan name are required!', 'error')
            return redirect('/import/cpanel')
        try:
            file_name = os.path.basename(path)
            log_file_name = f"cpanel_import_log_{os.path.splitext(file_name)[0]}"
            log_file_path = f"/var/log/openpanel/admin/{log_file_name}.log"

            # Run the subprocess command and redirect stdout and stderr to the log file
            with open(log_file_path, 'w') as log_file:
                subprocess.Popen(['opencli', 'user-import', 'cpanel', path, plan_name], stdout=log_file, stderr=log_file)
            
            flash(f'Import started! To track the progress open the log file: {log_file_path}', 'success')
        except Exception as e:
            flash(f'An error occurred: {str(e)}', 'error')

        return redirect('/import/cpanel')
    else:
        # on GET we will list the sessions in progress..
        return render_template('cpanel-import.html', title='Import cPanel account')
