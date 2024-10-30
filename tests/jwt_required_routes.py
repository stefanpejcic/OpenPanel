import os
import ast

app_path = '/usr/local/panel'

def find_api_routes_without_jwt_required():
    api_routes_without_jwt_required = []

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
                            api_route_decorator = False
                            jwt_required_decorator = False

                            for decorator in node.decorator_list:
                                if isinstance(decorator, ast.Call) and hasattr(decorator.func, 'attr') and decorator.func.attr == 'route' and hasattr(decorator.func.value, 'id') and decorator.func.value.id == 'api':
                                    api_route_decorator = True
                                elif isinstance(decorator, ast.Name) and decorator.id == 'jwt_required':
                                    jwt_required_decorator = True

                            if api_route_decorator and not jwt_required_decorator:
                                api_routes_without_jwt_required.append({
                                    'file': file_path,
                                    'route': node.name,
                                    'line_number': node.lineno
                                })

    return api_routes_without_jwt_required

api_routes = find_api_routes_without_jwt_required()

if api_routes:
    print("API routes without @jwt_required decorator:")
    for route in api_routes:
        print(f"{route['file']} -> API Route: {route['route']} (line {route['line_number']})")
else:
    print("All API routes have the @jwt_required decorator!")
  
