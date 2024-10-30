import os
import re
import ast

app_path = '/usr/local/panel'

def find_routes_without_login_required():
    routes_without_login_required = []

    for root, dirs, files in os.walk(app_path):
        for file in files:
            if file.endswith('.py'):
                file_path = os.path.join(root, file)

                with open(file_path, 'r', encoding='utf-8') as f:
                    file_content = f.read()
                    try:
                        parsed_content = ast.parse(file_content)
                    except SyntaxError as e:
                        print(f"Skipping {file_path} due to SyntaxError: {e}")
                        continue

                    for node in ast.walk(parsed_content):
                        if isinstance(node, ast.FunctionDef):
                            route_decorator = False
                            login_required_decorator = False

                            for decorator in node.decorator_list:
                                if isinstance(decorator, ast.Call) and hasattr(decorator.func, 'id') and decorator.func.id == 'route':
                                    route_decorator = True
                                elif isinstance(decorator, ast.Name) and decorator.id == 'login_required':
                                    login_required_decorator = True
                            if route_decorator and not login_required_decorator:
                                routes_without_login_required.append({
                                    'file': file_path,
                                    'route': node.name,
                                    'line_number': node.lineno
                                })

    return routes_without_login_required

routes = find_routes_without_login_required()

if routes:
    print("Routes without @login_required decorator:")
    for route in routes:
        print(f"{route['file']} -> Route: {route['route']} (line {route['line_number']})")
else:
    print("All routes have the @login_required decorator!")
  
