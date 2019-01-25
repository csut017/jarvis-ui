To deploy the monitor software as a background process on a Raspberry Pi.

sudo cp monitor.service /lib/systemd/system/
sudo chmod 644 /lib/systemd/system/monitor.service
sudo systemctl daemon-reload
sudo systemctl enable monitor.service
sudo systemctl start monitor.service

To check the status:

sudo systemctl status monitor.service

To stop the service:

sudo systemctl stop monitor.service

