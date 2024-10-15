## CHANGELOG

### Remove fabric and add support for Ubuntu 20.04

> Branches Affected: master  
> Tags Affected:   
> Date: Fri, 8 Mar 2021 09:50:00 +0000
> Author: Mauricio Vargas (@pachamaltese)  

- add support to image test script for Ubuntu 20.04
- removes fabric because of problems with Python 3
- updated documentation


### Add Support for Debian 9

> Branches Affected: master  
> Tags Affected:   
> Date: Fri, 29 Mar 2019 18:06:36 +0000  
> Author: root (root@debian-s-1vcpu-1gb-nyc1-01.localdomain)  
> Committer: root (root@debian-s-1vcpu-1gb-nyc1-01.localdomain)  

- add support to image test script for Debian
- update README's
- for more info on changelog hooks please ref: https://github.com/dOpensource/dsiprouter/tree/dev/resources/git


---


### Update build-an-image-fabric.md

> Branches Affected: master  
> Tags Affected:   
> Date: Mon, 25 Mar 2019 15:41:52 +0100  
> Author: Divyendu Singh (divyendu.z@gmail.com)  
> Committer: GitHub (noreply@github.com)  



---


### IMG-535 - Ensure validation script exits with proper status codes

> Branches Affected: master  
> Tags Affected:   
> Date: Thu, 21 Mar 2019 13:19:37 -0500  
> Author: rquinn (ryan@digitalocean.com)  
> Committer: rquinn (ryan@digitalocean.com)  



---


### Add step to cleanup to clear root mailbox if any mail is present

> Branches Affected: master  
> Tags Affected:   
> Date: Thu, 21 Mar 2019 13:13:49 -0500  
> Author: rquinn (ryan@digitalocean.com)  
> Committer: rquinn (ryan@digitalocean.com)  



---


### also allow CentOs convention for locked password on root

> Branches Affected: master  
> Tags Affected:   
> Date: Mon, 18 Mar 2019 18:40:25 -0400  
> Author: Ken Bingham (w@qrk.us)  
> Committer: Ken Bingham (w@qrk.us)  



---


### img_check.sh: Print date that script is run

> Branches Affected: master  
> Tags Affected:   
> Date: Wed, 13 Mar 2019 16:03:30 -0400  
> Author: John Gannon (jgannon@digitalocean.com)  
> Committer: GitHub (noreply@github.com)  

Assists in debugging efforts that may come up between DO and Vendors, as well as in DO's listing submission process.


---


### build-an-image.md: doctl instructions added

> Branches Affected: master  
> Tags Affected:   
> Date: Wed, 13 Mar 2019 14:40:54 -0400  
> Author: John Gannon (jgannon@digitalocean.com)  
> Committer: GitHub (noreply@github.com)  



---


### removed .backup file

> Branches Affected: master  
> Tags Affected:   
> Date: Mon, 11 Mar 2019 11:17:59 -0500  
> Author: rquinn (ryan@digitalocean.com)  
> Committer: rquinn (ryan@digitalocean.com)  



---


### Added Packer example and reorganized template folders

> Branches Affected: master  
> Tags Affected:   
> Date: Mon, 11 Mar 2019 11:16:37 -0500  
> Author: rquinn (ryan@digitalocean.com)  
> Committer: rquinn (ryan@digitalocean.com)  



---


### update img_check.sh exclude lfd.log log

> Branches Affected: master  
> Tags Affected:   
> Date: Wed, 6 Mar 2019 23:25:48 +1000  
> Author: George Liu (eva2000@centminmod.com)  
> Committer: George Liu (eva2000@centminmod.com)  



---


### img_check.sh centos 7 & csf firewall compatibility

> Branches Affected: master  
> Tags Affected:   
> Date: Wed, 6 Mar 2019 22:00:27 +1000  
> Author: George Liu (eva2000@centminmod.com)  
> Committer: George Liu (eva2000@centminmod.com)  



---


### README.md: Link out to snapshots documentation

> Branches Affected: master  
> Tags Affected:   
> Date: Mon, 4 Mar 2019 16:38:06 -0500  
> Author: John Gannon (jgannon@digitalocean.com)  
> Committer: GitHub (noreply@github.com)  



---


### Update build-an-image-fabric.md

> Branches Affected: master  
> Tags Affected:   
> Date: Thu, 28 Feb 2019 12:03:14 -0500  
> Author: Rijk van Zanten (rijkvanzanten@me.com)  
> Committer: GitHub (noreply@github.com)  

Hey there! I ran into some versioning issues while working with the example project. I hope this note will make sure other people don't spend 30min messing around as I did 🙂


---


### Update 001_onboot

> Branches Affected: master  
> Tags Affected:   
> Date: Fri, 22 Feb 2019 14:29:46 -0600  
> Author: Ryan Quinn (ryan@digitalocean.com)  
> Committer: GitHub (noreply@github.com)  

Fix for typo reported by @kspearrin


---


### Fix for IMG-559 / Issue 16 - Fail if private key found

> Branches Affected: master  
> Tags Affected:   
> Date: Fri, 22 Feb 2019 08:16:47 -0600  
> Author: rquinn (ryan@digitalocean.com)  
> Committer: rquinn (ryan@digitalocean.com)  



---


### remove shutdown func add exit to build func and add copy for user

> Branches Affected: master  
> Tags Affected:   
> Date: Fri, 15 Feb 2019 10:30:09 -0600  
> Author: rquinn (ryan@digitalocean.com)  
> Committer: rquinn (ryan@digitalocean.com)  



---


### stop apt-get from prompting about config files when upgrading packages

> Branches Affected: master  
> Tags Affected:   
> Date: Fri, 15 Feb 2019 10:04:01 -0600  
> Author: rquinn (ryan@digitalocean.com)  
> Committer: rquinn (ryan@digitalocean.com)  



---


### replace per-once with per-instance

> Branches Affected: master  
> Tags Affected:   
> Date: Fri, 15 Feb 2019 09:57:35 -0600  
> Author: rquinn (ryan@digitalocean.com)  
> Committer: rquinn (ryan@digitalocean.com)  



---


### Fix to change recommendation from per-once to per-instance on firstboot scripts

> Branches Affected: master  
> Tags Affected:   
> Date: Wed, 13 Feb 2019 13:14:51 -0600  
> Author: rquinn (ryan@digitalocean.com)  
> Committer: rquinn (ryan@digitalocean.com)  



---


### Ensure instance folder is removed on build system

> Branches Affected: master  
> Tags Affected:   
> Date: Tue, 12 Feb 2019 15:05:55 -0600  
> Author: rquinn (ryan@digitalocean.com)  
> Committer: rquinn (ryan@digitalocean.com)  



---


### Fix for permissions not being copied and adding noninteractive env var

> Branches Affected: master  
> Tags Affected:   
> Date: Tue, 12 Feb 2019 15:04:53 -0600  
> Author: rquinn (ryan@digitalocean.com)  
> Committer: rquinn (ryan@digitalocean.com)  



---


### moved postfix install fix to run if uncommented before apt installs

> Branches Affected: master  
> Tags Affected:   
> Date: Tue, 12 Feb 2019 12:41:02 -0600  
> Author: rquinn (ryan@digitalocean.com)  
> Committer: rquinn (ryan@digitalocean.com)  



---


### build-an-image.md: typo

> Branches Affected: master  
> Tags Affected:   
> Date: Mon, 4 Feb 2019 17:31:02 -0500  
> Author: John Gannon (jgannon@digitalocean.com)  
> Committer: GitHub (noreply@github.com)  



---


### Couple of extra fixes, emphasis on MOTD

> Branches Affected: master  
> Tags Affected:   
> Date: Fri, 1 Feb 2019 15:28:23 -0800  
> Author: Nick Wade (9014043+wadenick@users.noreply.github.com)  
> Committer: Nick Wade (9014043+wadenick@users.noreply.github.com)  



---


### Add updates and upgrades to cleanup example script for manual build + remove byoi_validation leftover

> Branches Affected: master  
> Tags Affected:   
> Date: Tue, 29 Jan 2019 14:28:20 -0800  
> Author: Nick Wade (9014043+wadenick@users.noreply.github.com)  
> Committer: Nick Wade (9014043+wadenick@users.noreply.github.com)  



---


### Delete BYOI version of the script

> Branches Affected: master  
> Tags Affected:   
> Date: Tue, 22 Jan 2019 16:49:31 -0500  
> Author: John Gannon (jgannon@digitalocean.com)  
> Committer: GitHub (noreply@github.com)  

Will be confusing for users and we don't plan to push anyone to BYOI in the near term. Can always pull it back into master later on.


---


### Initial release

> Branches Affected: master  
> Tags Affected: v1.0  
> Date: Wed, 16 Jan 2019 15:46:02 -0500  
> Author: Nick Wade (9014043+wadenick@users.noreply.github.com)  
> Committer: Nick Wade (9014043+wadenick@users.noreply.github.com)  



---


### removed .vscode directory

> Branches Affected: master  
> Tags Affected:   
> Date: Wed, 16 Jan 2019 14:38:55 -0600  
> Author: rquinn (ryan@digitalocean.com)  
> Committer: rquinn (ryan@digitalocean.com)  



---


### typos *facepalm*

> Branches Affected: master  
> Tags Affected:   
> Date: Wed, 16 Jan 2019 15:35:56 -0500  
> Author: Nick Wade (9014043+wadenick@users.noreply.github.com)  
> Committer: Nick Wade (9014043+wadenick@users.noreply.github.com)  



---


### repo README.md updates to just point to the new docs – gonna merge the branch shortly

> Branches Affected: master  
> Tags Affected:   
> Date: Wed, 16 Jan 2019 15:34:26 -0500  
> Author: Nick Wade (9014043+wadenick@users.noreply.github.com)  
> Committer: Nick Wade (9014043+wadenick@users.noreply.github.com)  



---


### formatting typos

> Branches Affected: master  
> Tags Affected:   
> Date: Wed, 16 Jan 2019 15:28:42 -0500  
> Author: Nick Wade (9014043+wadenick@users.noreply.github.com)  
> Committer: Nick Wade (9014043+wadenick@users.noreply.github.com)  



---


### tweaks to fabric doc

> Branches Affected: master  
> Tags Affected:   
> Date: Wed, 16 Jan 2019 15:26:53 -0500  
> Author: Nick Wade (9014043+wadenick@users.noreply.github.com)  
> Committer: Nick Wade (9014043+wadenick@users.noreply.github.com)  



---


### revision of the manual image build documentation

> Branches Affected: master  
> Tags Affected:   
> Date: Wed, 16 Jan 2019 14:13:08 -0600  
> Author: rquinn (ryan@digitalocean.com)  
> Committer: rquinn (ryan@digitalocean.com)  



---


### adding some simple img_check and snapshot instructions

> Branches Affected: master  
> Tags Affected:   
> Date: Wed, 16 Jan 2019 13:52:45 -0500  
> Author: Nick Wade (9014043+wadenick@users.noreply.github.com)  
> Committer: Nick Wade (9014043+wadenick@users.noreply.github.com)  



---


### just getting my various draft changes into order for Ryan

> Branches Affected: master  
> Tags Affected:   
> Date: Tue, 15 Jan 2019 15:25:41 -0500  
> Author: Nick Wade (9014043+wadenick@users.noreply.github.com)  
> Committer: Nick Wade (9014043+wadenick@users.noreply.github.com)  



---


### img_check.sh: list the uninstalled security packages

> Branches Affected: master  
> Tags Affected:   
> Date: Thu, 10 Jan 2019 16:15:45 -0500  
> Author: John Gannon (jgannon@digitalocean.com)  
> Committer: GitHub (noreply@github.com)  



---


### removed backup file

> Branches Affected: master  
> Tags Affected:   
> Date: Mon, 7 Jan 2019 12:44:24 -0600  
> Author: rquinn (ryan@digitalocean.com)  
> Committer: rquinn (ryan@digitalocean.com)  



---


### changed name from "example" to "template"

> Branches Affected: master  
> Tags Affected:   
> Date: Mon, 7 Jan 2019 12:43:21 -0600  
> Author: rquinn (ryan@digitalocean.com)  
> Committer: rquinn (ryan@digitalocean.com)  



---


### Adding simple fabric example

> Branches Affected: master  
> Tags Affected:   
> Date: Mon, 7 Jan 2019 12:40:08 -0600  
> Author: rquinn (ryan@digitalocean.com)  
> Committer: rquinn (ryan@digitalocean.com)  



---


### fix a typo

> Branches Affected: master  
> Tags Affected:   
> Date: Thu, 3 Jan 2019 12:50:59 -0800  
> Author: Nick Wade (9014043+wadenick@users.noreply.github.com)  
> Committer: Nick Wade (9014043+wadenick@users.noreply.github.com)  



---


### language pass, samples subfolder for various scripts

> Branches Affected: master  
> Tags Affected:   
> Date: Thu, 3 Jan 2019 12:49:42 -0800  
> Author: Nick Wade (9014043+wadenick@users.noreply.github.com)  
> Committer: Nick Wade (9014043+wadenick@users.noreply.github.com)  



---


### tweaker

> Branches Affected: master  
> Tags Affected:   
> Date: Thu, 3 Jan 2019 12:42:34 -0800  
> Author: Nick Wade (9014043+wadenick@users.noreply.github.com)  
> Committer: Nick Wade (9014043+wadenick@users.noreply.github.com)  



---


### prefix tweak

> Branches Affected: master  
> Tags Affected:   
> Date: Thu, 3 Jan 2019 12:37:58 -0800  
> Author: Nick Wade (9014043+wadenick@users.noreply.github.com)  
> Committer: Nick Wade (9014043+wadenick@users.noreply.github.com)  



---


### tweaks

> Branches Affected: master  
> Tags Affected:   
> Date: Thu, 3 Jan 2019 11:57:44 -0800  
> Author: Nick Wade (9014043+wadenick@users.noreply.github.com)  
> Committer: Nick Wade (9014043+wadenick@users.noreply.github.com)  



---


### moving docs around a bit

> Branches Affected: master  
> Tags Affected:   
> Date: Thu, 3 Jan 2019 11:17:55 -0800  
> Author: Nick Wade (9014043+wadenick@users.noreply.github.com)  
> Committer: Nick Wade (9014043+wadenick@users.noreply.github.com)  



---


### Fix for Issue 8 - script crashes if apt-cache is not present on Ubuntu

> Branches Affected: master  
> Tags Affected:   
> Date: Tue, 4 Dec 2018 09:02:39 -0600  
> Author: rquinn (ryan@digitalocean.com)  
> Committer: rquinn (ryan@digitalocean.com)  



---


### Starter doc

> Branches Affected: master  
> Tags Affected:   
> Date: Fri, 9 Nov 2018 10:19:00 -0800  
> Author: Nick Wade (nwade@digitalocean.com)  
> Committer: Nick Wade (nwade@digitalocean.com)  



---


### update for IMG-533 (check for do-agent)

> Branches Affected: master  
> Tags Affected:   
> Date: Fri, 9 Nov 2018 11:02:09 -0600  
> Author: rquinn (ryan@digitalocean.com)  
> Committer: rquinn (ryan@digitalocean.com)  



---


### Print authorized keys file

> Branches Affected: master  
> Tags Affected:   
> Date: Fri, 9 Nov 2018 10:21:30 -0600  
> Author: rquinn (ryan@digitalocean.com)  
> Committer: rquinn (ryan@digitalocean.com)  



---


### print contents of authorized keys file if detected

> Branches Affected: master  
> Tags Affected:   
> Date: Fri, 9 Nov 2018 10:16:53 -0600  
> Author: rquinn (ryan@digitalocean.com)  
> Committer: rquinn (ryan@digitalocean.com)  



---


### Spelling mistakes correction

> Branches Affected: master  
> Tags Affected:   
> Date: Sun, 21 Oct 2018 15:50:28 +0530  
> Author: akshaybengani789 (33260831+akshaybengani789@users.noreply.github.com)  
> Committer: GitHub (noreply@github.com)  



---


### remove final undefined BYOI references

> Branches Affected: master  
> Tags Affected:   
> Date: Thu, 18 Oct 2018 15:11:05 -0700  
> Author: Nick Wade (9014043+wadenick@users.noreply.github.com)  
> Committer: Nick Wade (9014043+wadenick@users.noreply.github.com)  



---


### update to version 0.1 and release tarball

> Branches Affected: master  
> Tags Affected:   
> Date: Thu, 18 Oct 2018 15:03:09 -0700  
> Author: Nick Wade (9014043+wadenick@users.noreply.github.com)  
> Committer: Nick Wade (9014043+wadenick@users.noreply.github.com)  



---


### no message

> Branches Affected: master  
> Tags Affected:   
> Date: Thu, 18 Oct 2018 14:59:34 -0700  
> Author: Nick Wade (9014043+wadenick@users.noreply.github.com)  
> Committer: Nick Wade (9014043+wadenick@users.noreply.github.com)  



---


### last heading fix

> Branches Affected: master  
> Tags Affected:   
> Date: Thu, 18 Oct 2018 14:45:56 -0700  
> Author: Nick Wade (9014043+wadenick@users.noreply.github.com)  
> Committer: Nick Wade (9014043+wadenick@users.noreply.github.com)  



---


### headings skipped

> Branches Affected: master  
> Tags Affected:   
> Date: Thu, 18 Oct 2018 14:44:54 -0700  
> Author: Nick Wade (9014043+wadenick@users.noreply.github.com)  
> Committer: Nick Wade (9014043+wadenick@users.noreply.github.com)  



---


### woops, heading space required

> Branches Affected: master  
> Tags Affected:   
> Date: Thu, 18 Oct 2018 14:43:18 -0700  
> Author: Nick Wade (9014043+wadenick@users.noreply.github.com)  
> Committer: Nick Wade (9014043+wadenick@users.noreply.github.com)  



---


### A few typo and consistency changes, along with fixing the image asset path

> Branches Affected: master  
> Tags Affected:   
> Date: Thu, 18 Oct 2018 14:41:29 -0700  
> Author: Nick Wade (9014043+wadenick@users.noreply.github.com)  
> Committer: Nick Wade (9014043+wadenick@users.noreply.github.com)  



---


### Initial open sourcing

> Branches Affected: master  
> Tags Affected:   
> Date: Thu, 18 Oct 2018 11:11:17 -0700  
> Author: Nick Wade (nwade@digitalocean.com)  
> Committer: Nick Wade (nwade@digitalocean.com)  



---


### Update README.md

> Branches Affected: master  
> Tags Affected:   
> Date: Tue, 16 Oct 2018 13:39:36 -0700  
> Author: wadenick (9014043+wadenick@users.noreply.github.com)  
> Committer: GitHub (noreply@github.com)  



---


### Initial commit

> Branches Affected: master  
> Tags Affected:   
> Date: Tue, 16 Oct 2018 13:20:47 -0700  
> Author: wadenick (9014043+wadenick@users.noreply.github.com)  
> Committer: GitHub (noreply@github.com)  



---
