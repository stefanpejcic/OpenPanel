# Error 

`opencli error` command allows Administrators to trace OpenPanel error IDs.

:::warning
Error is logged only if [dev_mode](/cli/config.html#dev-mode) is enabled.
:::

```bash
opencli error ID_HERE
```

Example output:

```bash
root@stefan:~# opencli error 9RIdD2X6aGaIzCzAEh0YL57U
[2024-11-06 14:14:47,298] ERROR in app: Exception on /emails/mozda@stefan.openpanel.org [GET]
Traceback (most recent call last):
File "/usr/local/lib/python3.12/site-packages/flask/app.py", line 1473, in wsgi_app
response = self.full_dispatch_request()
^^^^^^^^^^^^^^^^^^^^^^^^^^^^
File "/usr/local/lib/python3.12/site-packages/flask/app.py", line 882, in full_dispatch_request
rv = self.handle_user_exception(e)
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
File "/usr/local/lib/python3.12/site-packages/flask/app.py", line 880, in full_dispatch_request
rv = self.dispatch_request()
^^^^^^^^^^^^^^^^^^^^^^^
File "/usr/local/lib/python3.12/site-packages/flask/app.py", line 865, in dispatch_request
return self.ensure_sync(self.view_functions[rule.endpoint])(**view_args)  # type: ignore[no-any-return]
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
File "/usr/local/panel/app.py", line 478, in wrapper
return func(*args, **kwargs)
^^^^^^^^^^^^^^^^^^^^^
File "/usr/local/panel/modules/emails.py", line 240, in emails
send_output = check_output(f'opencli email-setup email restrict list send {emaifdfdl}', shell=True, stderr=subprocess.STDOUT).decode('utf-8').strip()
^^^^^^^^^
NameError: name 'emaifdfdl' is not defined
ERROR:root:Error Code: 9RIdD2X6aGaIzCzAEh0YL57U - 500 Internal Server Error: The server encountered an internal error and was unable to complete your request. Either the server is overloaded or there is an error in the application.

```
