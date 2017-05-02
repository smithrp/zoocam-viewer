# zoocam-viewer
Raspberry Pi Zoo Camera viewer

This is a simple zoo camera viewer for the raspberry pi.  Once installed you can configure the app to view any number of cameras at once (limited by your connection and hardware) and "zoom in" to a single view through the configured "remote" (which is just a basic website).

Note: this was done VERY quickly and is a complete hackjob, but does the job for my purposes.  No judging!

## running
go run main.go

## building
go build -o zoocam main.go

## installing to pi
sudo cp startzoo /etc/init.d/
sudo update-rc.d /etc/init.d/startzoo defaults
