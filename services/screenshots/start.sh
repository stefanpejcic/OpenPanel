#!/bin/bash
RUN_COMMAND="venv/bin/python -m flask run --host=0.0.0.0"
HEALTHCHECK_URL="http://localhost:5000/screenshot/pcx3.com"
#LOGFILE="flask_app.log"
LOGFILE="/dev/null"

start_flask() {
    echo "Starting Flask app..."
    cd /home/screenshot/
    nohup $RUN_COMMAND > "$LOGFILE" 2>&1 &
    FLASK_PID=$!
    echo "Flask app started with PID: $FLASK_PID"
}

check_health() {
    STATUS_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$HEALTHCHECK_URL")
    echo "Health check status code: $STATUS_CODE"
    if [ "$STATUS_CODE" -ne 200 ]; then
        echo "Health check failed. Restarting Flask app..."
        restart_flask
    else
        echo "Flask app is running fine."
    fi
}


restart_flask() {
    kill "$FLASK_PID"
    echo "Flask app stopped."
    start_flask
}


start_flask


while true; do
    sleep 60
    check_health
done
