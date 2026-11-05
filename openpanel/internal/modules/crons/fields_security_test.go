package crons

import "testing"

func TestCheckCronJobFieldsRejectsInjection(t *testing.T) {
	bad := []struct{ schedule, command, container, comment string }{
		{"@every 1h", "true\n[job-run \"x\"]\nimage = alpine", "php-fpm-8.3", "backup"},
		{"@every 1h", "true\r\nimage = alpine", "php-fpm-8.3", "backup"},
		{"@every 1h", "echo hi", "--env-file=/home/victim/.env", "backup"},
		{"@every 1h", "echo hi", "-it", "backup"},
		{"@every 1h", "job-local x", "php", "backup"},
	}
	for _, b := range bad {
		if checkCronJobFields(b.schedule, b.command, b.container, b.comment) == "" {
			t.Errorf("accepted %+v", b)
		}
	}
	if msg := checkCronJobFields("@every 1h", "php /var/www/html/cron.php", "php-fpm-8.3", "backup"); msg != "" {
		t.Errorf("rejected a normal job: %s", msg)
	}
}

func TestEditedFileWithInjectedSectionFailsPolicy(t *testing.T) {
	content := cronBlock(CronJob{Comment: "backup", Schedule: "@every 1h", Container: "php", Command: "true\n[job-run \"x\"]\nimage = alpine"})
	if CheckCronsPolicy(content) == "" {
		t.Fatal("policy accepted an injected job-run section")
	}
}

func TestValidateCronFileFormatRejectsOptionContainer(t *testing.T) {
	content := "[job-exec \"a\"]\nschedule = @every 1h\ncontainer = --env-file=/home/v/.env\ncommand = id\n"
	if ValidateCronFileFormat(content) == "" {
		t.Fatal("raw editor accepted an option-shaped container")
	}
}
