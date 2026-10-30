package docker

import "testing"

// trimmed from `podman image ls --format json` on the test VM
const sampleImageList = `[{"Id": "9214e73b07d3", "Size": 1161306292, "Containers": 1, "Names": ["docker.io/library/node:24.22.2"]}, {"Id": "9d924a42f59b", "Size": 1253889773, "Containers": 1, "Names": ["docker.io/n8nio/n8n:3.0.1"]}, {"Id": "aeb870beb76a", "Size": 51827769, "Containers": 0, "Names": ["docker.io/offen/docker-volume-backup:v2"]}, {"Id": "b92841763c5c", "Size": 8721505, "Containers": 0, "Names": ["docker.io/library/alpine:latest"]}, {"Id": "7a316d927aca", "Size": 99073070, "Containers": 0, "Names": ["docker.io/ubuntu/bind9:latest"]}, {"Id": "0150e3200277", "Size": 281041623, "Containers": 1, "Names": ["docker.io/library/postgres:17.4-alpine3.21"]}]`

func TestParseImageList(t *testing.T) {
	cases := []struct {
		name, in string
		want     ImageSummary
		wantErr  bool
	}{
		{"real output", sampleImageList, ImageSummary{Count: 6, Unused: 3, Bytes: 2855860032, Size: "2.9 GB"}, false},
		{"no images", `[]`, ImageSummary{Size: "0 B"}, false},
		{"one small", `[{"Size": 13435823, "Containers": 2}]`, ImageSummary{Count: 1, Bytes: 13435823, Size: "13.4 MB"}, false},
		{"not json", `Error: no socket`, ImageSummary{}, true},
	}
	for _, c := range cases {
		got, err := parseImageList([]byte(c.in))
		if (err != nil) != c.wantErr {
			t.Errorf("%s: err = %v", c.name, err)
			continue
		}
		if !c.wantErr && got != c.want {
			t.Errorf("%s: got %+v, want %+v", c.name, got, c.want)
		}
	}
}

// from `podman system df --format json` on the test VM: 9.55 GB, where summing image ls gives 10.69 GB
const sampleSystemDF = `[
    {"Type":"Images","Total":22,"Active":19,"RawSize":9550892919,"RawReclaimable":1277030088,"TotalCount":22,"Size":"9.551GB","Reclaimable":"1.277GB (13%)"},
    {"Type":"Containers","Total":19,"Active":19,"RawSize":119082262,"RawReclaimable":0,"TotalCount":19,"Size":"119.1MB","Reclaimable":"0B (0%)"},
    {"Type":"Local Volumes","Total":12,"Active":22,"RawSize":2024518978,"RawReclaimable":0,"TotalCount":12,"Size":"2.025GB","Reclaimable":"0B (0%)"}
]`

func TestParseSystemDF(t *testing.T) {
	got, ok := parseSystemDF([]byte(sampleSystemDF))
	if want := (ImageSummary{Count: 22, Unused: 3, Bytes: 9550892919, Size: "9.6 GB"}); !ok || got != want {
		t.Errorf("got %+v ok=%v, want %+v", got, ok, want)
	}
	for name, in := range map[string]string{
		"no RawSize (older podman)": `[{"Type":"Images","Total":2,"Active":1,"Size":"1GB"}]`,
		"no images row":             `[{"Type":"Containers","Total":1,"Active":1,"RawSize":5}]`,
		"not json":                  `Error: unrecognized command`,
	} {
		if _, ok := parseSystemDF([]byte(in)); ok {
			t.Errorf("%s: expected ok=false so image ls is used instead", name)
		}
	}
}
