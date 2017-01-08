#!/bin/sh

echo "Adding user..."
sudo useradd --groups dialout --no-create-home --system opentrvtrack

echo "Building and installing..."
go install github.com/cylindric/opentrvtrack

echo "Moving executable..."
sudo mv $GOPATH/bin/opentrvtrack /usr/bin

echo "Installing service..."
sudo cp opentrvtrack.service /etc/systemd/system
sudo systemctl daemon-reload
sudo systemctl enable opentrvtrack.service
