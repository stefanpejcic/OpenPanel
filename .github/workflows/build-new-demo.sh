name: Trigger Rebuild on Version Update

on:
  push:
    paths:
      - 'version/latest'

env:
  DROPLET_IP: ${{ secrets.DROPLET_IP }}
  DROPLET_ID: ${{ secrets.DROPLET_ID }}
  DIGITALOCEAN_TOKEN: ${{ secrets.DIGITALOCEAN_TOKEN }}
  SSH_KEY_PATH: ${{ secrets.SSH_KEY_PATH }}

jobs:
  rebuild:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Set up SSH key
        run: |
          mkdir -p ~/.ssh
          echo "${{ secrets.SSH_PRIVATE_KEY }}" > ~/.ssh/id_rsa
          chmod 600 ~/.ssh/id_rsa

      # Run the rebuild script
      - name: Rebuilding the demo site for new OpenPanel version
        run: |
          chmod +x demo/rebuild.sh
          ./demo/rebuild.sh
