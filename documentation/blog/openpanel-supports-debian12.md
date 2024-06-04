---
title: OpenPanel supports Debian
description: OpenPanel now supports Debian12 🎉
slug: openpanel-supports-debian12
authors: stefanpejcic
tags: [OpenPanel, debian, news]
image: https://openpanel.co/img/blog/openpanel_supports_debian12.png
hide_table_of_contents: true
is_featured: false
---

OpenPanel 0.2.0 supports Debian 🚀

<!--truncate-->

Ever since our launch in late February, I've received questions like: "Why Ubuntu and not Debian?" The truth is, I'm more accustomed to Ubuntu, and haven't worked extensively with Debian to feel as comfortable managing and maintaining it as I do with Ubuntu.

However, with the overwhelming demand from our users on GitHub, forums, and Discord to add Debian support, I've accepted the challenge and began working on the integration.

After a few months of development and some adjustments to our existing logic, I'm thrilled to announce that OpenPanel 0.2.0 now fully supports Debian! 🎉

We will support all current LTS versions of Debian, at this moment those are Debian11 and Debian12:

![debian eol](https://i.postimg.cc/6qYhDh6k/image.png)


While there are still a few minor issues remaining—ones that 99.9% of users will likely never encounter—I'm committed to resolving them as swiftly as possible.

**Bugs on debian:**
- ~Pyarmor decoding fails for a specific python version: 3.11~ - Fixed in 0.2.0
- phpMyAdmin version in docker images is not compatible with PHP 8.2+

Please [give it a try](https://openpanel.co/install) and don't hesitate to reach out if you encounter any bugs. Your feedback is invaluable in improving OpenPanel!

---

**Note:** As OpenPanel 0.2.0 is the first version to support Debian, attempting to install older versions of OpenPanel on Debian, such as `--version=0.1.8`, will fail.
