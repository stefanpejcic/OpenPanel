# Usage

## OS Install tests
Run on playwright server, edit the info on beginning of the file and run the test.
```
bash opencli/os_install.sh
```

---

## API
Run on openpanel test server, not on playwright server:
```
RUN_WRITE_TESTS=1 ADMIN_USER=stefan ADMIN_PASSWORD=stefan ./api.sh
```
