#!/bin/bash
# ======================================================================
# Shared password-strength scoring, sourced by scripts that set/change
# passwords (ftp/add.sh, ftp/password.sh, user/add.sh, user/password.sh).
#
# Keep this 6-check rubric in sync with:
#   - modules/core/validators.py (password_strength_score)
#   - static/js/password-strength.js (passwordStrengthScore)
#
# ======================================================================

PASSWORD_STRENGTH_CONFIG_FILE="/etc/openpanel/openpanel/conf/openpanel.config"

# Echoes a 0-100 score for the given password.
password_strength_score() {
    local password="$1"
    local score=0

    [[ -z "$password" ]] && { echo 0; return; }

    (( ${#password} >= 8 )) && ((score++))
    (( ${#password} >= 12 )) && ((score++))
    [[ "$password" =~ [a-z] ]] && ((score++))
    [[ "$password" =~ [A-Z] ]] && ((score++))
    [[ "$password" =~ [0-9] ]] && ((score++))
    [[ "$password" =~ [[:punct:]] ]] && ((score++))

    echo $(( (score * 100 + 3) / 6 ))
}

# Echoes the admin-configured threshold (1-100), default is 50.
get_password_strength_threshold() {
    local _pw_raw _pw_val
    _pw_raw=$(grep "^password_strength=" "$PASSWORD_STRENGTH_CONFIG_FILE" 2>/dev/null | cut -d= -f2- | tr -d '"'"'"'')

    if [[ "$_pw_raw" =~ ^[0-9]+$ ]]; then
        _pw_val="$_pw_raw"
    else
        _pw_val=50
    fi

    (( _pw_val < 1 )) && _pw_val=1
    (( _pw_val > 100 )) && _pw_val=100

    echo "$_pw_val"
}

# Exits the calling script with an error if the password doesn't meet the threshold.
require_password_strength() {
    local password="$1"
    local score threshold

    score=$(password_strength_score "$password")
    threshold=$(get_password_strength_threshold)

    if (( score < threshold )); then
        echo "ERROR: Password is too weak (strength $score/100, minimum required is $threshold/100)."
        exit 1
    fi
}
