
def copy_framework(framework: str):
    with open(f'../api_analytics/{framework}.py', 'r') as f:
        code = f.read()
    with open(f'../../{framework}/api_analytics/{framework}.py', 'w') as f:
        f.write(code)
        
    with open(f'../api_analytics/core.py', 'r') as f:
        code = f.read()
    with open(f'../../{framework}/api_analytics/core.py', 'w') as f:
        f.write(code)

def copy_example(framework: str):
    with open(f'../example/{framework}_ex.py', 'r') as f:
        code = f.read()
    with open(f'../../{framework}/example/{framework}_ex.py', 'w') as f:
        f.write(code)

def copy_test(framework: str):
    with open(f'../tests/{framework}_ex.py', 'r') as f:
        code = f.read()
    with open(f'../../{framework}/tests/{framework}_ex.py', 'w') as f:
        f.write(code)
        
def update_package(framework: str):
    copy_framework(framework)
    copy_example(framework)
    copy_test(framework)

update_package('fastapi')
update_package('tornado')