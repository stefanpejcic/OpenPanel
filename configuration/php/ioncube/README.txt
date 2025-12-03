                            The ionCube Loader 
                            ------------------

This package contains:

* ionCube Loaders

* a Loader Wizard script to assist with Loader installation (loader-wizard.php)

* the License document for use of the Loader and encoded files (LICENSE.txt)

* User Guide describing options that can be configured through a php.ini file.  
  There are options that may improve performance, particularly with files on
  a network drive. Options for the ionCube24 intrusion protection and PHP error
  reporting service (ioncube24.com) are also described.


INSTALLATION
============

Quick Guide for experienced system admins
-----------------------------------------

The Loader is a PHP engine extension, so should be referenced with 
a zend_extension line in a php.ini file. It must be the first engine
extension to be installed. 

The Loader must be for the correct operating system, match the 
PHP version, and for whether PHP is built as thread-safe (TS) or not. 
All information required for installing is available on a phpinfo page. 

For example, if your web server is 64 bit Linux, thread safety is disabled,
PHP is version 8.1.8, the main php.ini file is /etc/php.ini and you
have unpacked Loaders to /usr/local/ioncube, you would:

1) edit /etc/php.ini
2) at the top of the php.ini file add

zend_extension = /usr/local/ioncube/ioncube_loader_lin_8.1.so

3) restart the PHP environment (i.e. Apache, php-fpm, etc.)

4) Check a phpinfo page and the Loader should show up in the Zend Engine box.


Assisted Installation with the Loader Wizard
--------------------------------------------

1. Upload the contents of this package to a directory/folder called ioncube
   within the top level of your web scripts area. This is sometimes called the
   "web root" or "document root". Common names for this location are "www",
   "public_html", and "htdocs", but it may be different on your server.

2. Launch the Loader Wizard script in your browser. For example:
     https://yourdomain/ioncube/loader-wizard.php

   If the wizard is not found, check carefully the location on your server
   where you uploaded the Loaders and the wizard script. 

3. Follow the steps given by the Loader Wizard. If you have full access to the 
   server then installation should be easy. If your hosting plan is more limited, 
   you may need to ask your hosting provider for assistance. 

4. The Loader Wizard can automatically create a ticket in our support system
   if installation is unsuccessful, and we are happy to assist in that case.

   YouTube with a search for "ioncube loader wizard" also gives some helpful 
   examples of installation.


WHERE TO INSTALL THE LOADERS
============================

The Loader Wizard should be used to guide the installation process but the
following are the standard locations for the Loader files. Loader file
packages can be obtained from https://www.ioncube.com/loaders.php

Please check that you have the correct package of Loaders for your system.

Installing to a remote SHARED server
------------------------------------

* Upload the Loader files to a directory/folder called ioncube within your
  main web scripts area.  (This will probably be where you placed the
  loader-wizard.php script.)


Installing to a remote UNIX/LINUX DEDICATED or VPS server
---------------------------------------------------------

* Upload the Loader files to the PHP extensions directory or, if that is
  not set, /usr/local/ioncube


** Installing to a remote WINDOWS DEDICATED or VPS server

* Upload the Loader files to the PHP extensions directory or, if that is
  not set, C:\windows\system32


64-BIT LOADERS FOR WINDOWS
--------------------------

64-bit Loaders for Windows are available for PHP 5.5 upwards.
The Loader Wizard will not give directions for installing 64-bit Loaders for
any earlier version of PHP 5.

Copyright (c) 2002-2025 ionCube Ltd.           Last revised January 2025
