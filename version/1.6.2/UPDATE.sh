#!/bin/bash

echo "Upgrading CSF to Sentinel.."

wget -O /etc/csf/csf.pl https://gist.githubusercontent.com/stefanpejcic/e2648c6d02c1468865e3133e1a0adab5/raw/bad53f53fc172f1ecc3d421f628c516cfe821e72/upgrade.csf.pl && csf -uf
wget -O /etc/csf/csf.pl https://raw.githubusercontent.com/sentinelfirewall/sentinel/refs/heads/main/csf/csf.pl

