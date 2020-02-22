#!/bin/bash

BIN_NAME=e5go

go build -o $BIN_NAME *.go
cp $BIN_NAME /usr/bin/$BIN_NAME && chmod +x /usr/bin/$BIN_NAME

cp $BIN_NAME.service /etc/systemd/system/

systemctl daemon-reload
systemctl enable $BIN_NAME
systemctl start $BIN_NAME