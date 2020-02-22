#!/bin/bash

BIN_NAME=e5go

go build -o $BIN_NAME *.go
cp $BIN_NAME /usr/bin/$BIN_NAME && chmod +x /usr/bin/$BIN_NAME

cp $BIN_NAME /etc/systemd/system/$BIN_NAME.service

systemctl daemon_reload
systemctl enable $BIN_NAME
systemctl start $BIN_NAME