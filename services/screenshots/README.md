# Screenshots API service for OpenPanel server

Host your own screenshots api for your 'OpenPanel > SiteManager'.


## Installation

We recommend setting at least 2 servers for screenshots and one balancer before it.

Applications run on port 5000 on each server, and proxy handles incomming connections from ports 80 to 5000 on those servers.
```bash
mkdir /home/screenshots
cd /home/screenshots
git init
git remote add -f origin https://github.com/stefanpejcic/OpenPanel
git config core.sparseCheckout true
echo "services/screenshots/" >> .git/info/sparse-checkout
git pull origin master
```


then run INSTALL.sh in that directory.


## Example Droplet and Balancer on DO

![balancer](https://i.postimg.cc/y8CbnpfH/2024-11-13-13-56.png)


![servers](https://i.postimg.cc/t4dmcdfV/2024-11-13-13-55.png)

