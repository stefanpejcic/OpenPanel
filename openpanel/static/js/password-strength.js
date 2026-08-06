/*
Shared password-strength meter + admin-configurable strength gate.

Used by every password <input> that opts in via:
  x-data="passwordStrength({weak: '...', medium: '...', strong: '...', tooWeak: '...'}, {{ password_strength }})"
  x-init="checkStrength()"
with the input itself carrying x-ref="input", x-model="password" and @input="checkStrength()".

`{{ password_strength }}` is the admin-configured 1-100 threshold, injected
into every template via app.py's GLOBAL_DATA context processor.

Keep the scoring rubric in sync with:
  - /home/stefan/2083/modules/core/validators.py (password_strength_score)
  - /home/stefan/opencli/lib/password_strength.sh (password_strength_score)
*/

function passwordStrengthScore(value) {
    if (!value) return 0;
    let score = 0;
    if (value.length >= 8) score++;
    if (value.length >= 12) score++;
    if (/[a-z]/.test(value)) score++;
    if (/[A-Z]/.test(value)) score++;
    if (/\d/.test(value)) score++;
    if (/[^a-zA-Z0-9]/.test(value)) score++;
    return Math.round(score / 6 * 100);
}

function passwordStrength(translations, threshold) {
    return {
        password: '',
        strength: '',
        strengthMessage: '',
        translations,
        threshold: threshold || 50,

        checkStrength() {
            const val = this.password;
            const score = passwordStrengthScore(val);

            if (!val) {
                this.strength = '';
                this.strengthMessage = '';
            } else if (score <= 33) {
                this.strength = 'Weak';
                this.strengthMessage = this.translations.weak;
            } else if (score <= 67) {
                this.strength = 'Medium';
                this.strengthMessage = this.translations.medium;
            } else {
                this.strength = 'Strong';
                this.strengthMessage = this.translations.strong;
            }

            // Reuse native constraint validation (same mechanism as
            // required/pattern/minlength) so weak passwords block submit
            // without needing per-template submit-button wiring.
            if (this.$refs.input) {
                this.$refs.input.setCustomValidity(
                    val && score < this.threshold ? (this.translations.tooWeak || '') : ''
                );
            }
        }
    };
}
