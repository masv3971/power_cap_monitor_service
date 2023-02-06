.PHONY: update clean build build-all run package deploy test authors dist

BIN_NAME 				:= power_cap_monitor_service
SERVICE_NAME 			:= power_cap_monitor.service
VERSION                 := $(shell cat VERSION)
LDFLAGS                 := -ldflags "-w -s --extldflags '-static'"

default: linux

build:
		$(info build a static local binary)
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ./bin/${BIN_NAME} ${LDFLAGS} ./main.go

install:
		$(info Installing service)
		$(info Copying binary to /lib/systemd/system)
		sudo cp ${SERVICE_NAME} /lib/systemd/system/.

		$(info Change permissions on service file)
		sudo chmod 755 /lib/systemd/system/${SERVICE_NAME}

		$(info Reloading systemd)
		sudo systemctl daemon-reload

		$(info Enabling service)
		sudo systemctl enable ${SERVICE_NAME}

start:
		$(info Starting service)
		sudo systemctl start ${SERVICE_NAME}

status:
		$(info Checking service status)
		sudo systemctl status ${SERVICE_NAME}

restart:
		$(info Restarting service)
		sudo systemctl restart ${SERVICE_NAME}

stop:
		$(info Stopping service)
		sudo systemctl stop ${SERVICE_NAME}

uninstall:stop
		sudo systemctl disable ${SERVICE_NAME}
		sudo rm /lib/systemd/system/${SERVICE_NAME}
		sudo systemctl daemon-reload
		sudo systemctl reset-failed

reinstall:uninstall install