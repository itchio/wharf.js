
deploy:
	gopherjs build -v -m
	gsutil cp -Z -a public-read hopla.js jszip.min.js wharf.js.js wharf.js.js.map gs://dl.itch.ovh/wharf.js/
