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

      - name: Run Rebuild Script and Capture Snapshot ID
        id: rebuild
        run: |
          chmod +x demo/rebuild.sh
          output=$(./demo/rebuild.sh)
          echo "$output"
          snapshot_id=$(echo "$output" | grep -oP 'Snapshot ID: \K.*')
          echo "Snapshot ID captured: $snapshot_id"
          echo "snapshot_id=$snapshot_id" >> $GITHUB_ENV

      - name: Update SNAPSHOT_ID.env
        run: |
          echo "SNAPSHOT_ID=${{ env.snapshot_id }}" > demo/SNAPSHOT_ID.env

      - name: Commit and Push Updates
        run: |
          git config --global user.name "github-actions[bot]"
          git config --global user.email "github-actions[bot]@users.noreply.github.com"
          git add demo/SNAPSHOT_ID.env
          git commit -m "Demo rebuilt with new version - snapshot id: ${{ env.snapshot_id }}"
          git push









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
