name: Restore OpenPanel demo site

on:
  schedule:
    - cron: '0 * * * *'
  workflow_dispatch:

jobs:
  hourly-tasks:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout repository
        uses: actions/checkout@v3

      - name: Load SNAPSHOT_ID from environment file
        run: |
          echo "SNAPSHOT_ID=$(cat demo/SNAPSHOT_ID.env)" >> $GITHUB_ENV

      - name: Restore demo site
        env:
          DIGITALOCEAN_TOKEN: ${{ secrets.DIGITALOCEAN_TOKEN }}
          DROPLET_ID: ${{ secrets.DROPLET_ID }}
          SNAPSHOT_ID: ${{ env.SNAPSHOT_ID }}
          
        run: |
          chmod +x ./demo/restore-snapshot.sh
          ./demo/restore-snapshot.sh
