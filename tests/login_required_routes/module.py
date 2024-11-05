@app.route('/dev/login_required')
def login_required_view():
    routes = []
    
    for rule in app.url_map.iter_rules():
        route_info = {
            'endpoint': rule.endpoint,
            'methods': ', '.join(rule.methods),
            'login_required': False,
        }
        
        view_func = app.view_functions[rule.endpoint]
        if hasattr(view_func, '__login_required_route__'):
            route_info['login_required'] = True
        
        routes.append(route_info)

    return render_template('login_required.html', routes=routes)

