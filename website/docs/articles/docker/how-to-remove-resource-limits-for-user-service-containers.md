# How to remove resource limits for user service containers

Unlimited resources for all user services can be achieved by setting 0 for both CPU and RAM on that plan:

![userlimits.png](/img/panel/v2/userlimits-guide.png)

Then set the maximum of the server's available cpu and ram for each service in the /home/USERNAME/.env file:

![userlimits2.png](/img/panel/v2/userlimits-guide2.png)

or you can run the following commands through the servers terminal:

```
cd /home/USERNAME/

sed -i '/^[^#]*_CPU=/ !b; /TOTAL_CPU=/ b; s/=.*/="2.0"/' .env

sed -i '/^[^#]*_RAM=/ !b; /TOTAL_RAM=/ b; s/=.*/="4.0G"/' .env
```

simply replace USERNAME with the desired username, change 2.0 in the example with the desired/total cpu count of your server and 4.0G with ram in GB.

Afterwards, any services that are started will use ALL the available resources and the user will be notified by a warning stating that CPU/RAM are unlimited:

![userlimits3.png](/img/panel/v2/userlimits-guide3.png)

To set unlimited CPU/RAM for ALL new users, run these sed commands on the .env template in /etc/openpanel/docker/compose/1.0/ instead .
