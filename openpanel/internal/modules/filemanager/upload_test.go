package filemanager

import "testing"

func TestUploadSucceeded(t *testing.T) {
	cases := []struct {
		name       string
		files      int
		allSuccess bool
		errs       []string
		want       bool
	}{
		{"all saved", 2, true, nil, true},
		// the form page said "uploaded successfully" next to "No files were uploaded"
		{"no files", 0, true, nil, false},
		// an invalid name adds an error without clearing allSuccess, the form page still said success
		{"rejected name", 1, true, []string{"Invalid filename: ..."}, false},
		{"failed save", 1, false, []string{"Failed to save a.txt"}, false},
	}
	for _, c := range cases {
		if got := uploadSucceeded(c.files, c.allSuccess, c.errs); got != c.want {
			t.Errorf("%s: uploadSucceeded = %v, want %v", c.name, got, c.want)
		}
	}
}
